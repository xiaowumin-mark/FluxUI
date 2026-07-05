package widget

import (
	"errors"
	"fmt"
	"image"
	"io"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"

	gioEvent "gioui.org/io/event"
	"gioui.org/io/transfer"
	"gioui.org/op/clip"
)

const defaultDropTargetMaxBytes int64 = 32 << 20

var defaultDropTargetTypes = []string{
	"text/uri-list",
	"text/plain;charset=utf-8",
	"text/plain;charset=utf8",
	"text/plain",
	"application/text",
}

// DropEvent describes data dropped onto a DropTarget.
type DropEvent struct {
	Type      string
	Data      []byte
	Text      string
	Paths     []string
	Operation DragOperation
	Err       error
}

// DropTargetStateEvent describes drag activity over a DropTarget.
type DropTargetStateEvent struct {
	Active bool
	Types  []string
}

// DropTargetOption configures a DropTarget.
type DropTargetOption func(*dropTargetConfig)

type dropTargetConfig struct {
	types          []string
	maxBytes       int64
	operation      DragOperation
	disabled       bool
	onActiveChange func(ctx *internal.Context, event DropTargetStateEvent)
	onError        func(ctx *internal.Context, event DropEvent)
}

type dropTargetState struct {
	active bool
}

// DropTarget accepts drag-and-drop transfer data over the child area.
//
// The first version is receive-only. It listens for URI lists and common text
// MIME types; file drops are reported through DropEvent.Paths when the platform
// backend supplies file URIs or absolute paths.
func DropTarget(child Widget, onDrop func(ctx *internal.Context, event DropEvent), opts ...DropTargetOption) Widget {
	cfg := dropTargetConfig{
		types:     append([]string(nil), defaultDropTargetTypes...),
		maxBytes:  defaultDropTargetMaxBytes,
		operation: DragOperationCopy,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	cfg.types = normalizeDropTypes(cfg.types)
	if cfg.maxBytes <= 0 {
		cfg.maxBytes = defaultDropTargetMaxBytes
	}
	return &dropTargetWidget{
		child:  child,
		onDrop: onDrop,
		config: cfg,
	}
}

// DropTargetTypes replaces the MIME types accepted by a DropTarget.
func DropTargetTypes(types ...string) DropTargetOption {
	return func(cfg *dropTargetConfig) {
		cfg.types = normalizeDropTypes(types)
	}
}

// DropTargetMaxBytes limits the number of bytes read from one drop payload.
func DropTargetMaxBytes(maxBytes int64) DropTargetOption {
	return func(cfg *dropTargetConfig) {
		cfg.maxBytes = maxBytes
	}
}

// DropTargetOperation sets the application-level operation reported with drops.
func DropTargetOperation(operation DragOperation) DropTargetOption {
	return func(cfg *dropTargetConfig) {
		cfg.operation = firstDragOperation([]DragOperation{operation})
	}
}

// DropTargetDisabled disables transfer handling without changing child layout.
func DropTargetDisabled(disabled bool) DropTargetOption {
	return func(cfg *dropTargetConfig) {
		cfg.disabled = disabled
	}
}

// DropTargetOnActiveChange observes potential drag enter/leave activity.
func DropTargetOnActiveChange(fn func(ctx *internal.Context, event DropTargetStateEvent)) DropTargetOption {
	return func(cfg *dropTargetConfig) {
		cfg.onActiveChange = fn
	}
}

// DropTargetOnError observes payload read errors in addition to onDrop.
func DropTargetOnError(fn func(ctx *internal.Context, event DropEvent)) DropTargetOption {
	return func(cfg *dropTargetConfig) {
		cfg.onError = fn
	}
}

type dropTargetWidget struct {
	child  Widget
	onDrop func(ctx *internal.Context, event DropEvent)
	config dropTargetConfig
}

func (d *dropTargetWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if d.child == nil || ctx == nil {
		return layout.Dimensions{}
	}

	state := dropTargetStateFor(ctx)
	childDims := d.child.Layout(ctx.Child(0))

	if childDims.Size.X > 0 && childDims.Size.Y > 0 && !d.config.disabled && d.hasHandlers() {
		stack := clip.Rect(image.Rectangle{Max: childDims.Size}).Push(ctx.Gtx.Ops)
		gioEvent.Op(ctx.Gtx.Ops, state)
		stack.Pop()
	}
	d.dispatchDrops(ctx, state)
	return childDims
}

func (d *dropTargetWidget) hasHandlers() bool {
	return d != nil && (d.onDrop != nil || d.config.onError != nil || d.config.onActiveChange != nil)
}

func (d *dropTargetWidget) dispatchDrops(ctx *internal.Context, state *dropTargetState) {
	if ctx == nil || state == nil {
		return
	}
	if d.config.disabled {
		d.setActive(ctx, state, false)
		return
	}
	if d.onDrop == nil && d.config.onError == nil && d.config.onActiveChange == nil {
		return
	}
	for _, typ := range d.config.types {
		for {
			ev, ok := ctx.Gtx.Event(transfer.TargetFilter{Target: state, Type: typ})
			if !ok {
				break
			}
			switch ev := ev.(type) {
			case transfer.InitiateEvent:
				if d.dispatchDropTargetEvent(ctx, fluxevent.DragEnter, DropEvent{}, true, ev) &&
					d.dispatchDropTargetEvent(ctx, fluxevent.DragOver, DropEvent{}, true, ev) {
					d.setActive(ctx, state, true)
				}
			case transfer.CancelEvent:
				d.dispatchDropTargetEvent(ctx, fluxevent.DragLeave, DropEvent{}, false, ev)
				d.setActive(ctx, state, false)
			case transfer.DataEvent:
				event := dropEventFromTransfer(ev, d.config.maxBytes, d.config.operation)
				allowed := d.dispatchDropTargetEvent(ctx, fluxevent.Drop, event, false, ev)
				if allowed && event.Err != nil && d.config.onError != nil {
					d.config.onError(ctx, event)
				}
				if allowed && d.onDrop != nil {
					d.onDrop(ctx, event)
				}
				d.setActive(ctx, state, false)
			}
		}
	}
}

func (d *dropTargetWidget) setActive(ctx *internal.Context, state *dropTargetState, active bool) {
	if state.active == active {
		return
	}
	state.active = active
	if d.config.onActiveChange == nil {
		return
	}
	d.config.onActiveChange(ctx, DropTargetStateEvent{
		Active: active,
		Types:  append([]string(nil), d.config.types...),
	})
}

func (d *dropTargetWidget) dispatchDropTargetEvent(ctx *internal.Context, eventType fluxevent.Type, event DropEvent, active bool, native any) bool {
	if ctx == nil {
		return false
	}
	data := append([]byte(nil), event.Data...)
	paths := append([]string(nil), event.Paths...)
	operation := event.Operation
	if operation == "" {
		operation = d.config.operation
	}
	return fluxevent.DispatchDragEvent(ctx, ctx.PathID(), &fluxevent.DragEvent{
		Event: fluxevent.Event{
			Type:    eventType,
			Trusted: true,
		},
		MIMEType:  event.Type,
		Data:      data,
		Text:      event.Text,
		Paths:     paths,
		Operation: string(firstDragOperation([]DragOperation{operation})),
		Types:     append([]string(nil), d.config.types...),
		Active:    active,
		Err:       event.Err,
		Native:    native,
	})
}

func dropTargetStateFor(ctx *internal.Context) *dropTargetState {
	value := ctx.Memo("drop-target", func() any {
		return &dropTargetState{}
	})
	state, ok := value.(*dropTargetState)
	if !ok {
		panic(fmt.Sprintf("github.com/xiaowumin-mark/FluxUI/widget: key %q drop target state type mismatch", ctx.TreePath()))
	}
	return state
}

func dropEventFromTransfer(event transfer.DataEvent, maxBytes int64, operation DragOperation) DropEvent {
	result := DropEvent{Type: event.Type, Operation: firstDragOperation([]DragOperation{operation})}
	if event.Open == nil {
		result.Err = errors.New("drop target: transfer data is unavailable")
		return result
	}
	reader := event.Open()
	if reader == nil {
		result.Err = errors.New("drop target: transfer data is unavailable")
		return result
	}
	defer reader.Close()

	data, err := readDropData(reader, maxBytes)
	if err != nil {
		result.Err = err
		return result
	}
	result.Data = data
	result.Text = string(data)
	result.Paths = parseDropPaths(event.Type, data)
	return result
}

func readDropData(reader io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		maxBytes = defaultDropTargetMaxBytes
	}
	data, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("drop target: transfer data exceeds %d bytes", maxBytes)
	}
	return data, nil
}

func normalizeDropTypes(types []string) []string {
	if len(types) == 0 {
		return append([]string(nil), defaultDropTargetTypes...)
	}
	seen := map[string]bool{}
	normalized := make([]string, 0, len(types))
	for _, typ := range types {
		typ = strings.TrimSpace(strings.ToLower(typ))
		if typ == "" || seen[typ] {
			continue
		}
		seen[typ] = true
		normalized = append(normalized, typ)
	}
	if len(normalized) == 0 {
		return append([]string(nil), defaultDropTargetTypes...)
	}
	return normalized
}

func parseDropPaths(mime string, data []byte) []string {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	paths := make([]string, 0, len(lines))
	uriList := strings.EqualFold(strings.TrimSpace(mime), "text/uri-list")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(strings.ToLower(line), "file:") {
			if path, ok := fileURIToPath(line); ok {
				paths = append(paths, path)
			}
			continue
		}
		if !uriList && filepath.IsAbs(line) {
			paths = append(paths, filepath.Clean(line))
		}
	}
	return uniqueDropPaths(paths)
}

func fileURIToPath(value string) (string, bool) {
	parsed, err := url.Parse(value)
	if err != nil || !strings.EqualFold(parsed.Scheme, "file") {
		return "", false
	}
	path := parsed.Opaque
	if path == "" {
		path, err = url.PathUnescape(parsed.EscapedPath())
		if err != nil {
			return "", false
		}
	}
	if path == "" {
		return "", false
	}
	if parsed.Host != "" && !strings.EqualFold(parsed.Host, "localhost") {
		path = `\\` + parsed.Host + filepath.FromSlash(path)
	} else if runtime.GOOS == "windows" && strings.HasPrefix(path, "/") && len(path) >= 3 && path[2] == ':' {
		path = path[1:]
	}
	return filepath.Clean(filepath.FromSlash(path)), true
}

func uniqueDropPaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(paths))
	unique := make([]string, 0, len(paths))
	for _, path := range paths {
		key := path
		if runtime.GOOS == "windows" {
			key = strings.ToLower(key)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		unique = append(unique, path)
	}
	return unique
}
