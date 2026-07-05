package widget

import (
	"bytes"
	"fmt"
	"image"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"

	"gioui.org/f32"
	"gioui.org/gesture"
	gioEvent "gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/transfer"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// DragPayload is one MIME payload offered by a DragSource.
type DragPayload struct {
	Type string
	Data []byte
}

// DragOperation identifies the application-level operation associated with a
// drag-and-drop transfer.
type DragOperation string

const (
	DragOperationCopy DragOperation = "copy"
	DragOperationMove DragOperation = "move"
	DragOperationLink DragOperation = "link"
)

// DragSourceEventKind identifies a DragSource lifecycle event.
type DragSourceEventKind string

const (
	DragSourceEventStarted   DragSourceEventKind = "started"
	DragSourceEventRequested DragSourceEventKind = "requested"
	DragSourceEventCompleted DragSourceEventKind = "completed"
	DragSourceEventCancelled DragSourceEventKind = "cancelled"
)

// DragSourceEvent describes a transfer request handled by a DragSource.
type DragSourceEvent struct {
	Kind      DragSourceEventKind
	Type      string
	Data      []byte
	Operation DragOperation
	Err       error
}

// DragSourceOption configures a DragSource.
type DragSourceOption func(*dragSourceConfig)

type dragSourceConfig struct {
	payloads   []DragPayload
	preview    Widget
	operations []DragOperation
	disabled   bool
	onEvent    func(ctx *internal.Context, event DragSourceEvent)
	onRequest  func(ctx *internal.Context, event DragSourceEvent)
}

type dragSourceState struct {
	drag   gesture.Drag
	click  f32.Point
	pos    f32.Point
	active bool
}

type dragSourcePointerUpdate struct {
	started   bool
	ended     bool
	cancelled bool
}

// DragSource makes the child area draggable and offers transfer data to drop
// targets.
//
// Use DragSourceFiles for file drag-out and DragSourceText or DragSourceData
// for text/custom payloads. The first version depends on Gio's transfer
// backend; platforms that do not expose native drag-out still keep the API
// compileable and usable between FluxUI widgets.
func DragSource(child Widget, opts ...DragSourceOption) Widget {
	cfg := dragSourceConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	cfg.payloads = normalizeDragPayloads(cfg.payloads)
	cfg.operations = normalizeDragOperations(cfg.operations)
	return &dragSourceWidget{
		child:  child,
		config: cfg,
	}
}

// DragSourcePayloads replaces the payloads offered by a DragSource.
func DragSourcePayloads(payloads ...DragPayload) DragSourceOption {
	return func(cfg *dragSourceConfig) {
		cfg.payloads = append([]DragPayload(nil), payloads...)
	}
}

// DragSourceData adds one MIME payload offered by a DragSource.
func DragSourceData(mime string, data []byte) DragSourceOption {
	return func(cfg *dragSourceConfig) {
		cfg.payloads = append(cfg.payloads, DragPayload{Type: mime, Data: append([]byte(nil), data...)})
	}
}

// DragSourceText offers text payloads.
func DragSourceText(text string) DragSourceOption {
	return func(cfg *dragSourceConfig) {
		data := []byte(text)
		cfg.payloads = append(cfg.payloads,
			DragPayload{Type: "text/plain;charset=utf-8", Data: append([]byte(nil), data...)},
			DragPayload{Type: "text/plain", Data: append([]byte(nil), data...)},
			DragPayload{Type: "application/text", Data: append([]byte(nil), data...)},
		)
	}
}

// DragSourceFiles offers file paths as text/uri-list and plain text payloads.
func DragSourceFiles(paths ...string) DragSourceOption {
	return func(cfg *dragSourceConfig) {
		normalized := normalizeDragFilePaths(paths)
		if len(normalized) == 0 {
			return
		}
		cfg.payloads = append(cfg.payloads,
			DragPayload{Type: "text/uri-list", Data: []byte(dragFileURIList(normalized))},
			DragPayload{Type: "text/plain", Data: []byte(strings.Join(normalized, "\n"))},
		)
	}
}

// DragSourcePreview sets the widget drawn under the pointer while dragging.
func DragSourcePreview(preview Widget) DragSourceOption {
	return func(cfg *dragSourceConfig) {
		cfg.preview = preview
	}
}

// DragSourceOperations declares the operations the drag represents.
//
// Gio's transfer API does not expose OS-level operation negotiation. FluxUI
// records the normalized operation on DragSourceEvent for app logic and docs.
func DragSourceOperations(operations ...DragOperation) DragSourceOption {
	return func(cfg *dragSourceConfig) {
		cfg.operations = normalizeDragOperations(operations)
	}
}

// DragSourceDisabled disables the source without changing its child layout.
func DragSourceDisabled(disabled bool) DragSourceOption {
	return func(cfg *dragSourceConfig) {
		cfg.disabled = disabled
	}
}

// DragSourceOnEvent observes start/request/cancel lifecycle events.
func DragSourceOnEvent(fn func(ctx *internal.Context, event DragSourceEvent)) DragSourceOption {
	return func(cfg *dragSourceConfig) {
		cfg.onEvent = fn
	}
}

// DragSourceOnRequest observes transfer requests served by a DragSource.
func DragSourceOnRequest(fn func(ctx *internal.Context, event DragSourceEvent)) DragSourceOption {
	return func(cfg *dragSourceConfig) {
		cfg.onRequest = fn
	}
}

type dragSourceWidget struct {
	child  Widget
	config dragSourceConfig
}

func (d *dragSourceWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if d.child == nil || ctx == nil {
		return layout.Dimensions{}
	}

	state := dragSourceStateFor(ctx)
	dims := d.layoutSource(ctx, state)
	d.dispatchRequests(ctx, state)
	return dims
}

func (d *dragSourceWidget) layoutSource(ctx *internal.Context, state *dragSourceState) layout.Dimensions {
	childDims := d.child.Layout(ctx.Child(0))
	if childDims.Size.X <= 0 || childDims.Size.Y <= 0 || d.config.disabled || len(d.config.payloads) == 0 {
		return childDims
	}

	stack := clip.Rect(image.Rectangle{Max: childDims.Size}).Push(ctx.Gtx.Ops)
	state.drag.Add(ctx.Gtx.Ops)
	gioEvent.Op(ctx.Gtx.Ops, state)
	stack.Pop()

	preview := d.config.preview
	if preview == nil {
		preview = d.child
	}
	if state.drag.Pressed() && preview != nil {
		macro := op.Record(ctx.Gtx.Ops)
		op.Offset(state.pos.Round()).Add(ctx.Gtx.Ops)
		preview.Layout(ctx.Child(1))
		op.Defer(ctx.Gtx.Ops, macro.Stop())
	}

	return childDims
}

func (d *dragSourceWidget) dispatchRequests(ctx *internal.Context, state *dragSourceState) {
	if ctx == nil || state == nil {
		return
	}
	if d.config.disabled || len(d.config.payloads) == 0 {
		d.setActive(ctx, state, false)
		return
	}
	pointerUpdate := state.update(ctx)
	if pointerUpdate.started {
		d.setActive(ctx, state, true)
	}
	requested := false
	cancelled := pointerUpdate.cancelled
	ended := pointerUpdate.ended
	for _, payload := range d.config.payloads {
		for {
			ev, ok := ctx.Gtx.Event(transfer.SourceFilter{Target: state, Type: payload.Type})
			if !ok {
				break
			}
			switch ev := ev.(type) {
			case transfer.InitiateEvent:
				d.setActive(ctx, state, true)
			case transfer.CancelEvent:
				cancelled = true
			case transfer.RequestEvent:
				if ev.Type != payload.Type {
					continue
				}
				requested = true
				data := append([]byte(nil), payload.Data...)
				ctx.Gtx.Execute(transfer.OfferCmd{
					Tag:  state,
					Type: ev.Type,
					Data: ioNopReadCloser(data),
				})
				event := DragSourceEvent{
					Kind:      DragSourceEventRequested,
					Type:      ev.Type,
					Data:      data,
					Operation: firstDragOperation(d.config.operations),
				}
				allowed := d.dispatchSourceEvent(ctx, event)
				if allowed && d.config.onRequest != nil {
					d.config.onRequest(ctx, event)
				}
				d.finish(ctx, state, DragSourceEventCompleted)
			}
		}
	}
	if !requested && (cancelled || ended) {
		d.finish(ctx, state, DragSourceEventCancelled)
	}
}

func (d *dragSourceWidget) setActive(ctx *internal.Context, state *dragSourceState, active bool) {
	if state.active == active {
		return
	}
	state.active = active
	kind := DragSourceEventStarted
	if !active {
		kind = DragSourceEventCancelled
	}
	d.dispatchSourceEvent(ctx, DragSourceEvent{
		Kind:      kind,
		Operation: firstDragOperation(d.config.operations),
	})
}

func (d *dragSourceWidget) finish(ctx *internal.Context, state *dragSourceState, kind DragSourceEventKind) {
	if !state.active && kind != DragSourceEventCompleted {
		return
	}
	state.active = false
	d.dispatchSourceEvent(ctx, DragSourceEvent{
		Kind:      kind,
		Operation: firstDragOperation(d.config.operations),
	})
}

func (d *dragSourceWidget) dispatchSourceEvent(ctx *internal.Context, event DragSourceEvent) bool {
	if ctx == nil {
		return false
	}
	drag := dragEventFromSourceEvent(event)
	allowed := fluxevent.DispatchDragEvent(ctx, ctx.PathID(), drag)
	if d.config.onEvent == nil {
		return allowed
	}
	if allowed {
		d.config.onEvent(ctx, event)
	}
	return allowed
}

func dragEventFromSourceEvent(event DragSourceEvent) *fluxevent.DragEvent {
	data := append([]byte(nil), event.Data...)
	return &fluxevent.DragEvent{
		Event: fluxevent.Event{
			Type:    dragSourceFluxEventType(event.Kind),
			Trusted: true,
		},
		MIMEType:  event.Type,
		Data:      data,
		Text:      string(data),
		Operation: string(event.Operation),
		Err:       event.Err,
	}
}

func dragSourceFluxEventType(kind DragSourceEventKind) fluxevent.Type {
	switch kind {
	case DragSourceEventStarted:
		return fluxevent.DragStart
	case DragSourceEventRequested:
		return fluxevent.Drag
	case DragSourceEventCompleted, DragSourceEventCancelled:
		return fluxevent.DragEnd
	default:
		return fluxevent.Drag
	}
}

func (s *dragSourceState) update(ctx *internal.Context) dragSourcePointerUpdate {
	pos := s.pos
	update := dragSourcePointerUpdate{}
	for {
		ev, ok := s.drag.Update(ctx.Gtx.Metric, ctx.Gtx.Source, gesture.Both)
		if !ok {
			break
		}
		switch ev.Kind {
		case pointer.Press:
			s.click = ev.Position
			pos = f32.Point{}
		case pointer.Drag:
			pos = ev.Position.Sub(s.click)
			update.started = true
		case pointer.Release:
			pos = ev.Position.Sub(s.click)
			update.ended = true
		case pointer.Cancel:
			update.cancelled = true
		}
	}
	s.pos = pos
	return update
}

func dragSourceStateFor(ctx *internal.Context) *dragSourceState {
	value := ctx.Memo("drag-source", func() any {
		return &dragSourceState{}
	})
	state, ok := value.(*dragSourceState)
	if !ok {
		panic(fmt.Sprintf("github.com/xiaowumin-mark/FluxUI/widget: key %q drag source state type mismatch", ctx.TreePath()))
	}
	return state
}

func normalizeDragPayloads(payloads []DragPayload) []DragPayload {
	if len(payloads) == 0 {
		return nil
	}
	seen := map[string]bool{}
	normalized := make([]DragPayload, 0, len(payloads))
	for _, payload := range payloads {
		typ := strings.TrimSpace(strings.ToLower(payload.Type))
		if typ == "" || len(payload.Data) == 0 || seen[typ] {
			continue
		}
		seen[typ] = true
		normalized = append(normalized, DragPayload{
			Type: typ,
			Data: append([]byte(nil), payload.Data...),
		})
	}
	return normalized
}

func normalizeDragOperations(operations []DragOperation) []DragOperation {
	if len(operations) == 0 {
		return []DragOperation{DragOperationCopy}
	}
	seen := map[DragOperation]bool{}
	normalized := make([]DragOperation, 0, len(operations))
	for _, op := range operations {
		switch op {
		case "", DragOperationCopy:
			op = DragOperationCopy
		case DragOperationMove, DragOperationLink:
		default:
			continue
		}
		if seen[op] {
			continue
		}
		seen[op] = true
		normalized = append(normalized, op)
	}
	if len(normalized) == 0 {
		return []DragOperation{DragOperationCopy}
	}
	return normalized
}

func firstDragOperation(operations []DragOperation) DragOperation {
	operations = normalizeDragOperations(operations)
	if len(operations) == 0 {
		return DragOperationCopy
	}
	return operations[0]
}

func normalizeDragFilePaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	out := make([]string, 0, len(paths))
	seen := map[string]bool{}
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		path = filepath.Clean(path)
		key := path
		if runtime.GOOS == "windows" {
			key = strings.ToLower(key)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, path)
	}
	return out
}

func dragFileURIList(paths []string) string {
	uris := make([]string, 0, len(paths))
	for _, path := range paths {
		uris = append(uris, pathToFileURI(path))
	}
	return strings.Join(uris, "\r\n")
}

func pathToFileURI(path string) string {
	slash := filepath.ToSlash(filepath.Clean(path))
	if strings.HasPrefix(slash, "//") {
		rest := strings.TrimPrefix(slash, "//")
		host, tail, ok := strings.Cut(rest, "/")
		if ok && host != "" {
			return (&url.URL{Scheme: "file", Host: host, Path: "/" + tail}).String()
		}
	}
	if runtime.GOOS == "windows" && len(slash) >= 2 && slash[1] == ':' {
		slash = "/" + slash
	}
	return (&url.URL{Scheme: "file", Path: slash}).String()
}

func ioNopReadCloser(data []byte) *nopReadCloser {
	return &nopReadCloser{Reader: bytes.NewReader(data)}
}

type nopReadCloser struct {
	*bytes.Reader
}

func (n *nopReadCloser) Close() error {
	return nil
}
