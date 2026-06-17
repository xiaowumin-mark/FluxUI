package system

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

const (
	singleInstancePortBase  = 49152
	singleInstancePortRange = 10000
	singleInstanceIOTimeout = 1500 * time.Millisecond
	singleInstanceMaxBytes  = 64 * 1024
)

// SingleInstanceEvent is delivered to the primary instance when a secondary
// launch forwards its startup data.
type SingleInstanceEvent struct {
	ID      string
	Args    []string
	Payload string
}

// SingleInstanceOption configures single-instance acquisition.
type SingleInstanceOption func(*singleInstanceOptions)

type singleInstanceOptions struct {
	id             string
	args           []string
	payload        string
	port           int
	onSecondLaunch func(SingleInstanceEvent)
}

type singleInstanceDriver interface {
	acquireSingleInstance(ctx context.Context, opts singleInstanceOptions) (singleInstanceHandle, error)
}

type singleInstanceHandle interface {
	events() <-chan SingleInstanceEvent
	close() error
}

// SingleInstance owns the primary instance coordination channel.
type SingleInstance struct {
	handle singleInstanceHandle
}

// SingleInstanceID sets the stable application identity used for coordination.
//
// Applications should pass an explicit reverse-DNS style value, for example
// "com.example.myapp". If omitted, FluxUI derives a best-effort ID from the
// current executable name.
func SingleInstanceID(id string) SingleInstanceOption {
	return func(opts *singleInstanceOptions) {
		opts.id = strings.TrimSpace(id)
	}
}

// SingleInstanceArgs sets the startup arguments forwarded by secondary launches.
//
// When omitted, os.Args[1:] is used.
func SingleInstanceArgs(args ...string) SingleInstanceOption {
	return func(opts *singleInstanceOptions) {
		opts.args = append([]string(nil), args...)
	}
}

// SingleInstancePayload sets an optional string payload forwarded by secondary
// launches. Protocol activation handlers can use this for the raw activation URI.
func SingleInstancePayload(payload string) SingleInstanceOption {
	return func(opts *singleInstanceOptions) {
		opts.payload = payload
	}
}

// SingleInstanceOnSecondLaunch registers a callback for forwarded secondary
// launches. Events are also available through (*SingleInstance).Events().
func SingleInstanceOnSecondLaunch(fn func(SingleInstanceEvent)) SingleInstanceOption {
	return func(opts *singleInstanceOptions) {
		opts.onSecondLaunch = fn
	}
}

// SingleInstancePort overrides the deterministic loopback port derived from ID.
//
// This is mainly useful when two applications would otherwise collide on the
// same derived port. Values must be in the TCP user port range.
func SingleInstancePort(port int) SingleInstanceOption {
	return func(opts *singleInstanceOptions) {
		opts.port = port
	}
}

// AcquireSingleInstance becomes the primary instance or forwards startup data to
// the current primary instance.
//
// On success, the caller owns the primary instance and should close the returned
// handle during shutdown. If another instance is already running, the startup
// data is forwarded and the function returns an error wrapping ErrAlreadyRunning.
func AcquireSingleInstance(ctx context.Context, options ...SingleInstanceOption) (*SingleInstance, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	opts := defaultSingleInstanceOptions()
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}
	if err := normalizeSingleInstanceOptions(&opts); err != nil {
		return nil, err
	}

	d, supported := currentDriverFor(CapabilitySingleInstance)
	sd, ok := d.(singleInstanceDriver)
	if !ok || !supported {
		return nil, fmt.Errorf("system: %s: %w", CapabilitySingleInstance, ErrUnsupported)
	}

	handle, err := sd.acquireSingleInstance(ctx, opts)
	if err != nil {
		return nil, err
	}
	if handle == nil {
		return nil, fmt.Errorf("system: %s: acquire: %w", CapabilitySingleInstance, ErrUnavailable)
	}
	return &SingleInstance{handle: handle}, nil
}

// Events returns secondary launch events delivered to the primary instance.
func (s *SingleInstance) Events() <-chan SingleInstanceEvent {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.events()
}

// Close releases the primary instance coordination channel.
func (s *SingleInstance) Close() error {
	if s == nil || s.handle == nil {
		return singleInstanceClosedError()
	}
	return s.handle.close()
}

func defaultSingleInstanceOptions() singleInstanceOptions {
	return singleInstanceOptions{
		id:   defaultSingleInstanceID(),
		args: append([]string(nil), os.Args[1:]...),
	}
}

func normalizeSingleInstanceOptions(opts *singleInstanceOptions) error {
	if opts == nil {
		return fmt.Errorf("system: %s: missing options: %w", CapabilitySingleInstance, ErrUnavailable)
	}
	opts.id = strings.TrimSpace(opts.id)
	if opts.id == "" {
		return fmt.Errorf("system: %s: id is empty", CapabilitySingleInstance)
	}
	opts.args = append([]string(nil), opts.args...)
	if opts.port == 0 {
		opts.port = singleInstancePortForID(opts.id)
	}
	if opts.port < 1024 || opts.port > 65535 {
		return fmt.Errorf("system: %s: invalid port %d", CapabilitySingleInstance, opts.port)
	}
	return nil
}

func defaultSingleInstanceID() string {
	exe, err := os.Executable()
	if err == nil && exe != "" {
		return "fluxui." + filepath.Base(exe)
	}
	if len(os.Args) > 0 && os.Args[0] != "" {
		return "fluxui." + filepath.Base(os.Args[0])
	}
	return "fluxui.app"
}

func singleInstancePortForID(id string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(id))
	return singleInstancePortBase + int(h.Sum32()%singleInstancePortRange)
}

func acquireLoopbackSingleInstance(ctx context.Context, opts singleInstanceOptions) (singleInstanceHandle, error) {
	addr := singleInstanceAddress(opts.port)
	listener, listenErr := new(net.ListenConfig).Listen(ctx, "tcp", addr)
	if listenErr == nil {
		handle := newLoopbackSingleInstanceHandle(listener, opts)
		go handle.acceptLoop()
		return handle, nil
	}
	if err := forwardLoopbackSingleInstance(ctx, addr, opts); err != nil {
		if IsAlreadyRunning(err) {
			return nil, err
		}
		return nil, fmt.Errorf("system: %s: acquire %q on %s: %w: listen=%v forward=%v", CapabilitySingleInstance, opts.id, addr, ErrUnavailable, listenErr, err)
	}
	return nil, fmt.Errorf("system: %s: %w", CapabilitySingleInstance, ErrAlreadyRunning)
}

func singleInstanceAddress(port int) string {
	return fmt.Sprintf("127.0.0.1:%d", port)
}

type loopbackSingleInstanceMessage struct {
	Protocol string   `json:"protocol"`
	ID       string   `json:"id"`
	Args     []string `json:"args,omitempty"`
	Payload  string   `json:"payload,omitempty"`
}

type loopbackSingleInstanceResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type loopbackSingleInstanceHandle struct {
	listener net.Listener
	eventsCh chan SingleInstanceEvent
	callback func(SingleInstanceEvent)
	id       string
	closed   atomic.Bool
	done     chan struct{}
}

func newLoopbackSingleInstanceHandle(listener net.Listener, opts singleInstanceOptions) *loopbackSingleInstanceHandle {
	return &loopbackSingleInstanceHandle{
		listener: listener,
		eventsCh: make(chan SingleInstanceEvent, 16),
		callback: opts.onSecondLaunch,
		id:       opts.id,
		done:     make(chan struct{}),
	}
}

func (h *loopbackSingleInstanceHandle) events() <-chan SingleInstanceEvent {
	if h == nil {
		return nil
	}
	return h.eventsCh
}

func (h *loopbackSingleInstanceHandle) close() error {
	if h == nil || !h.closed.CompareAndSwap(false, true) {
		return singleInstanceClosedError()
	}
	err := h.listener.Close()
	<-h.done
	if err != nil {
		return fmt.Errorf("system: %s: close: %w", CapabilitySingleInstance, err)
	}
	return nil
}

func (h *loopbackSingleInstanceHandle) acceptLoop() {
	defer close(h.eventsCh)
	defer close(h.done)

	for {
		conn, err := h.listener.Accept()
		if err != nil {
			return
		}
		go h.handleConn(conn)
	}
}

func (h *loopbackSingleInstanceHandle) handleConn(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(singleInstanceIOTimeout))

	var msg loopbackSingleInstanceMessage
	if err := json.NewDecoder(io.LimitReader(conn, singleInstanceMaxBytes)).Decode(&msg); err != nil {
		writeSingleInstanceResponse(conn, "error", "decode")
		return
	}
	if msg.Protocol != singleInstanceProtocol(h.id) || msg.ID != h.id {
		writeSingleInstanceResponse(conn, "mismatch", "id mismatch")
		return
	}

	event := SingleInstanceEvent{
		ID:      msg.ID,
		Args:    append([]string(nil), msg.Args...),
		Payload: msg.Payload,
	}
	h.dispatch(event)
	writeSingleInstanceResponse(conn, "ok", "")
}

func (h *loopbackSingleInstanceHandle) dispatch(event SingleInstanceEvent) {
	if h.callback != nil {
		go h.callback(event)
	}
	select {
	case h.eventsCh <- event:
	default:
	}
}

func forwardLoopbackSingleInstance(ctx context.Context, addr string, opts singleInstanceOptions) error {
	deadline := time.Now().Add(singleInstanceIOTimeout)
	var lastErr error
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		conn, err := new(net.Dialer).DialContext(ctx, "tcp", addr)
		if err == nil {
			return writeLoopbackSingleInstanceMessage(conn, opts)
		}
		lastErr = err
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	return fmt.Errorf("connect primary: %w", lastErr)
}

func writeLoopbackSingleInstanceMessage(conn net.Conn, opts singleInstanceOptions) error {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(singleInstanceIOTimeout))

	msg := loopbackSingleInstanceMessage{
		Protocol: singleInstanceProtocol(opts.id),
		ID:       opts.id,
		Args:     append([]string(nil), opts.args...),
		Payload:  opts.payload,
	}
	if err := json.NewEncoder(conn).Encode(msg); err != nil {
		return fmt.Errorf("write payload: %w", err)
	}
	var response loopbackSingleInstanceResponse
	if err := json.NewDecoder(io.LimitReader(conn, singleInstanceMaxBytes)).Decode(&response); err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if response.Status != "ok" {
		if response.Error == "" {
			response.Error = response.Status
		}
		return fmt.Errorf("primary rejected payload: %s", response.Error)
	}
	return fmt.Errorf("system: %s: %w", CapabilitySingleInstance, ErrAlreadyRunning)
}

func writeSingleInstanceResponse(conn net.Conn, status, detail string) {
	_ = json.NewEncoder(conn).Encode(loopbackSingleInstanceResponse{
		Status: status,
		Error:  detail,
	})
}

func singleInstanceProtocol(id string) string {
	sum := sha1.Sum([]byte(id))
	return fmt.Sprintf("fluxui-single-instance-v1-%08x", binary.BigEndian.Uint32(sum[:4]))
}

func singleInstanceClosedError() error {
	return fmt.Errorf("system: %s: %w", CapabilitySingleInstance, ErrClosed)
}
