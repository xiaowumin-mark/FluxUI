//go:build windows

package system

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	wmHotkey = 0x0312

	windowsModAlt     = 0x0001
	windowsModControl = 0x0002
	windowsModShift   = 0x0004
	windowsModWin     = 0x0008
)

var (
	procRegisterHotKey              = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey            = user32.NewProc("UnregisterHotKey")
	windowsShortcutState            = newWindowsGlobalShortcutState()
	windowsShortcutProc             = syscall.NewCallback(windowsGlobalShortcutWindowProc)
	unregisterWindowsGlobalShortcut = windowsUnregisterHotKey
)

type windowsGlobalShortcutStateData struct {
	once    sync.Once
	ready   chan struct{}
	mu      sync.Mutex
	hWnd    uintptr
	err     error
	nextID  int32
	entries map[int32]*windowsGlobalShortcutHandle
}

type windowsGlobalShortcutHandle struct {
	mu       sync.Mutex
	id       int32
	spec     GlobalShortcutSpec
	callback func(GlobalShortcutEvent)
	ch       chan GlobalShortcutEvent
	closed   bool
}

func newWindowsGlobalShortcutState() *windowsGlobalShortcutStateData {
	return &windowsGlobalShortcutStateData{
		ready:   make(chan struct{}),
		entries: make(map[int32]*windowsGlobalShortcutHandle),
	}
}

func (windowsDriver) registerGlobalShortcut(ctx context.Context, spec GlobalShortcutSpec, fn func(GlobalShortcutEvent)) (globalShortcutHandle, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	hWnd, err := windowsShortcutState.window()
	if err != nil {
		return nil, fmt.Errorf("system: %s: initialize hotkey window: %w", CapabilityGlobalShortcut, err)
	}
	modifiers := windowsGlobalShortcutModifiers(spec.Modifiers)
	vk, err := windowsGlobalShortcutKey(spec.Key)
	if err != nil {
		return nil, err
	}
	id := windowsShortcutState.nextShortcutID()
	if err := windowsRegisterHotKey(hWnd, id, modifiers, vk); err != nil {
		return nil, err
	}
	handle := &windowsGlobalShortcutHandle{
		id:       id,
		spec:     spec,
		callback: fn,
		ch:       make(chan GlobalShortcutEvent, 16),
	}
	windowsShortcutState.setEntry(id, handle)
	return handle, nil
}

func windowsProbeGlobalShortcutCapability(cap Capability) CapabilityAvailability {
	if err := procRegisterHotKey.Find(); err != nil {
		return windowsCapabilityUnavailable(cap, err)
	}
	if err := procUnregisterHotKey.Find(); err != nil {
		return windowsCapabilityUnavailable(cap, err)
	}
	if _, err := windowsShortcutState.window(); err != nil {
		return windowsCapabilityUnavailable(cap, err)
	}
	return windowsCapabilityAvailable(cap)
}

func (s *windowsGlobalShortcutStateData) window() (uintptr, error) {
	s.once.Do(func() {
		go s.run()
	})
	<-s.ready

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hWnd, s.err
}

func (s *windowsGlobalShortcutStateData) run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hWnd, err := createWindowsGlobalShortcutWindow()
	s.mu.Lock()
	s.hWnd = hWnd
	s.err = err
	close(s.ready)
	s.mu.Unlock()
	if err != nil {
		return
	}

	var msg windowsMessage
	for {
		r, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			return
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func createWindowsGlobalShortcutWindow() (uintptr, error) {
	hInstance, err := windowsModuleHandle()
	if err != nil {
		return 0, err
	}
	className, err := windows.UTF16PtrFromString(windowsGlobalShortcutClassName())
	if err != nil {
		return 0, err
	}
	wc := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   windowsShortcutProc,
		hInstance:     hInstance,
		lpszClassName: className,
	}
	r, _, callErr := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if r == 0 && callErr != syscall.Errno(1410) {
		if callErr != syscall.Errno(0) {
			return 0, fmt.Errorf("register hotkey window class: %w", callErr)
		}
		return 0, fmt.Errorf("register hotkey window class: %w", ErrUnavailable)
	}
	hWnd, _, callErr := procCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(className)),
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		hInstance,
		0,
	)
	if hWnd == 0 {
		if callErr != syscall.Errno(0) {
			return 0, fmt.Errorf("create hotkey window: %w", callErr)
		}
		return 0, fmt.Errorf("create hotkey window: %w", ErrUnavailable)
	}
	return hWnd, nil
}

func windowsGlobalShortcutWindowProc(hWnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	if msg == wmHotkey {
		id := int32(wParam)
		if entry := windowsShortcutState.entry(id); entry != nil {
			entry.dispatch()
		}
		return 0
	}
	r, _, _ := procDefWindowProc.Call(hWnd, uintptr(msg), wParam, lParam)
	return r
}

func windowsGlobalShortcutClassName() string {
	return fmt.Sprintf("FluxUIGlobalShortcutWindow-%d-%x", os.Getpid(), uintptr(windowsShortcutProc))
}

func (s *windowsGlobalShortcutStateData) nextShortcutID() int32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	if s.nextID == 0 {
		s.nextID = 1
	}
	return s.nextID
}

func (s *windowsGlobalShortcutStateData) setEntry(id int32, entry *windowsGlobalShortcutHandle) {
	s.mu.Lock()
	s.entries[id] = entry
	s.mu.Unlock()
}

func (s *windowsGlobalShortcutStateData) entry(id int32) *windowsGlobalShortcutHandle {
	s.mu.Lock()
	entry := s.entries[id]
	s.mu.Unlock()
	return entry
}

func (s *windowsGlobalShortcutStateData) removeEntry(id int32) {
	s.mu.Lock()
	delete(s.entries, id)
	s.mu.Unlock()
}

func (h *windowsGlobalShortcutHandle) events() <-chan GlobalShortcutEvent {
	return h.ch
}

func (h *windowsGlobalShortcutHandle) close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return globalShortcutClosedError()
	}
	h.closed = true
	windowsShortcutState.removeEntry(h.id)
	err := unregisterWindowsGlobalShortcut(windowsShortcutState.hWndSnapshot(), h.id)
	close(h.ch)
	return err
}

func (h *windowsGlobalShortcutHandle) dispatch() {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	event := GlobalShortcutEvent{
		ID:        h.spec.ID,
		Key:       h.spec.Key,
		Modifiers: h.spec.Modifiers,
	}
	callback := h.callback
	select {
	case h.ch <- event:
	default:
	}
	h.mu.Unlock()
	if callback != nil {
		go callback(event)
	}
}

func (s *windowsGlobalShortcutStateData) hWndSnapshot() uintptr {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hWnd
}

func windowsRegisterHotKey(hWnd uintptr, id int32, modifiers, key uint32) error {
	r, _, callErr := procRegisterHotKey.Call(hWnd, uintptr(id), uintptr(modifiers), uintptr(key))
	if r != 0 {
		return nil
	}
	if callErr != syscall.Errno(0) {
		return fmt.Errorf("system: %s: register hotkey: %w: %v", CapabilityGlobalShortcut, ErrUnavailable, callErr)
	}
	return fmt.Errorf("system: %s: register hotkey: %w", CapabilityGlobalShortcut, ErrUnavailable)
}

func windowsUnregisterHotKey(hWnd uintptr, id int32) error {
	r, _, callErr := procUnregisterHotKey.Call(hWnd, uintptr(id))
	if r != 0 {
		return nil
	}
	if callErr != syscall.Errno(0) {
		return fmt.Errorf("system: %s: unregister hotkey: %w: %v", CapabilityGlobalShortcut, ErrUnavailable, callErr)
	}
	return fmt.Errorf("system: %s: unregister hotkey: %w", CapabilityGlobalShortcut, ErrUnavailable)
}

func windowsGlobalShortcutModifiers(modifiers GlobalShortcutModifiers) uint32 {
	var result uint32
	if modifiers&GlobalShortcutAlt != 0 {
		result |= windowsModAlt
	}
	if modifiers&GlobalShortcutControl != 0 {
		result |= windowsModControl
	}
	if modifiers&GlobalShortcutShift != 0 {
		result |= windowsModShift
	}
	if modifiers&GlobalShortcutMeta != 0 {
		result |= windowsModWin
	}
	return result
}

func windowsGlobalShortcutKey(key string) (uint32, error) {
	key = strings.ToUpper(strings.TrimSpace(key))
	if len(key) == 1 {
		r := key[0]
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return uint32(r), nil
		}
	}
	if strings.HasPrefix(key, "F") {
		n, err := strconv.Atoi(strings.TrimPrefix(key, "F"))
		if err == nil && n >= 1 && n <= 24 {
			return uint32(0x70 + n - 1), nil
		}
	}
	if vk, ok := windowsNamedGlobalShortcutKeys[key]; ok {
		return vk, nil
	}
	return 0, fmt.Errorf("system: %s: unsupported key %q: %w", CapabilityGlobalShortcut, key, ErrUnsupported)
}

var windowsNamedGlobalShortcutKeys = map[string]uint32{
	"BACKSPACE": 0x08,
	"TAB":       0x09,
	"ENTER":     0x0D,
	"RETURN":    0x0D,
	"ESC":       0x1B,
	"ESCAPE":    0x1B,
	"SPACE":     0x20,
	"PAGEUP":    0x21,
	"PAGEDOWN":  0x22,
	"END":       0x23,
	"HOME":      0x24,
	"LEFT":      0x25,
	"UP":        0x26,
	"RIGHT":     0x27,
	"DOWN":      0x28,
	"INSERT":    0x2D,
	"DELETE":    0x2E,
}
