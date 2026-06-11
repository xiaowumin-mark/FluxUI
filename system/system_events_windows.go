//go:build windows

package system

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	wmDisplayChange  = 0x007E
	wmSettingChange  = 0x001A
	wmThemeChanged   = 0x031A
	wmPowerBroadcast = 0x0218
	wmDPIChanged     = 0x02E0
	wmWTSSession     = 0x02B1

	pbtAPMSuspend          = 0x0004
	pbtAPMResumeSuspend    = 0x0007
	pbtAPMResumeAutomatic  = 0x0012
	pbtPowerSettingChange  = 0x8013
	notifyForThisSession   = 0
	systemEventChannelSize = 16
)

var (
	wtsapi32                           = windows.NewLazySystemDLL("wtsapi32.dll")
	procWTSRegisterSessionNotification = wtsapi32.NewProc("WTSRegisterSessionNotification")
	windowsSystemEventState            = newWindowsSystemEventState()
	windowsSystemEventProc             = syscall.NewCallback(windowsSystemEventWindowProc)
)

type windowsSystemEventStateData struct {
	once        sync.Once
	ready       chan struct{}
	mu          sync.Mutex
	hWnd        uintptr
	err         error
	nextID      uint64
	subscribers map[uint64]*windowsSystemEventSubscription
}

type windowsSystemEventSubscription struct {
	id     uint64
	ch     chan SystemEvent
	filter map[SystemEventKind]bool
}

func newWindowsSystemEventState() *windowsSystemEventStateData {
	return &windowsSystemEventStateData{
		ready:       make(chan struct{}),
		subscribers: make(map[uint64]*windowsSystemEventSubscription),
	}
}

func (windowsDriver) subscribeSystemEvents(ctx context.Context, kinds []SystemEventKind) (systemEventHandle, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if _, err := windowsSystemEventState.window(); err != nil {
		return nil, fmt.Errorf("system: %s: initialize system event window: %w", CapabilitySystemEvents, err)
	}
	return windowsSystemEventState.add(kinds), nil
}

func (s *windowsSystemEventStateData) window() (uintptr, error) {
	s.once.Do(func() {
		go s.run()
	})
	<-s.ready

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hWnd, s.err
}

func (s *windowsSystemEventStateData) run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hWnd, err := createWindowsSystemEventWindow()

	s.mu.Lock()
	s.hWnd = hWnd
	s.err = err
	close(s.ready)
	s.mu.Unlock()

	if err != nil {
		return
	}

	registerWindowsSessionNotifications(hWnd)

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

func (s *windowsSystemEventStateData) add(kinds []SystemEventKind) *windowsSystemEventSubscription {
	sub := &windowsSystemEventSubscription{
		ch:     make(chan SystemEvent, systemEventChannelSize),
		filter: systemEventFilter(kinds),
	}

	s.mu.Lock()
	s.nextID++
	sub.id = s.nextID
	s.subscribers[sub.id] = sub
	s.mu.Unlock()
	return sub
}

func (s *windowsSystemEventStateData) remove(id uint64) {
	s.mu.Lock()
	sub := s.subscribers[id]
	delete(s.subscribers, id)
	s.mu.Unlock()

	if sub != nil {
		close(sub.ch)
	}
}

func (s *windowsSystemEventStateData) broadcast(event SystemEvent) {
	if event.Kind == "" {
		return
	}
	if event.Time.IsZero() {
		event.Time = time.Now()
	}

	s.mu.Lock()
	subscribers := make([]*windowsSystemEventSubscription, 0, len(s.subscribers))
	for _, sub := range s.subscribers {
		subscribers = append(subscribers, sub)
	}
	s.mu.Unlock()

	for _, sub := range subscribers {
		if !sub.accepts(event.Kind) {
			continue
		}
		select {
		case sub.ch <- event:
		default:
		}
	}
}

func (s *windowsSystemEventSubscription) events() <-chan SystemEvent {
	if s == nil {
		return nil
	}
	return s.ch
}

func (s *windowsSystemEventSubscription) close() error {
	if s == nil {
		return ErrClosed
	}
	windowsSystemEventState.remove(s.id)
	return nil
}

func (s *windowsSystemEventSubscription) accepts(kind SystemEventKind) bool {
	return s != nil && (len(s.filter) == 0 || s.filter[kind])
}

func systemEventFilter(kinds []SystemEventKind) map[SystemEventKind]bool {
	if len(kinds) == 0 {
		return nil
	}
	filter := make(map[SystemEventKind]bool, len(kinds))
	for _, kind := range kinds {
		if kind != "" {
			filter[kind] = true
		}
	}
	return filter
}

func createWindowsSystemEventWindow() (uintptr, error) {
	hInstance, err := windowsModuleHandle()
	if err != nil {
		return 0, err
	}

	className, err := windows.UTF16PtrFromString(windowsSystemEventClassName())
	if err != nil {
		return 0, err
	}

	wc := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   windowsSystemEventProc,
		hInstance:     hInstance,
		lpszClassName: className,
	}

	r, _, callErr := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if r == 0 && callErr != syscall.Errno(1410) {
		if callErr != syscall.Errno(0) {
			return 0, fmt.Errorf("register system event window class: %w", callErr)
		}
		return 0, fmt.Errorf("register system event window class: %w", ErrUnavailable)
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
			return 0, fmt.Errorf("create system event window: %w", callErr)
		}
		return 0, fmt.Errorf("create system event window: %w", ErrUnavailable)
	}

	return hWnd, nil
}

func windowsSystemEventClassName() string {
	return fmt.Sprintf("FluxUISystemEventWindow-%d-%x", os.Getpid(), uintptr(windowsSystemEventProc))
}

func registerWindowsSessionNotifications(hWnd uintptr) {
	if hWnd == 0 {
		return
	}
	if err := procWTSRegisterSessionNotification.Find(); err != nil {
		return
	}
	procWTSRegisterSessionNotification.Call(hWnd, notifyForThisSession)
}

func windowsSystemEventWindowProc(hWnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	if event, ok := windowsSystemEventFromMessage(msg, wParam, lParam); ok {
		windowsSystemEventState.broadcast(event)
	}

	r, _, _ := procDefWindowProc.Call(hWnd, uintptr(msg), wParam, lParam)
	return r
}

func windowsSystemEventFromMessage(msg uint32, wParam uintptr, lParam uintptr) (SystemEvent, bool) {
	switch msg {
	case wmDisplayChange:
		return SystemEvent{
			Kind:   SystemEventDisplayChanged,
			Detail: windowsDisplayChangeDetail(lParam),
		}, true
	case wmDPIChanged:
		return SystemEvent{
			Kind:   SystemEventDPIChanged,
			Detail: fmt.Sprintf("%d", lowWord(wParam)),
		}, true
	case wmThemeChanged:
		return SystemEvent{Kind: SystemEventThemeChanged}, true
	case wmSettingChange:
		return SystemEvent{Kind: SystemEventSettingsChanged}, true
	case wmPowerBroadcast:
		return SystemEvent{
			Kind:   SystemEventPowerChanged,
			Detail: windowsPowerBroadcastDetail(wParam),
		}, true
	case wmWTSSession:
		return SystemEvent{
			Kind:   SystemEventSessionChanged,
			Detail: fmt.Sprintf("%d", wParam),
		}, true
	default:
		return SystemEvent{}, false
	}
}

func windowsDisplayChangeDetail(lParam uintptr) string {
	width := lowWord(lParam)
	height := highWord(lParam)
	if width == 0 || height == 0 {
		return ""
	}
	return fmt.Sprintf("%dx%d", width, height)
}

func windowsPowerBroadcastDetail(wParam uintptr) string {
	switch uint32(wParam) {
	case pbtAPMSuspend:
		return "suspend"
	case pbtAPMResumeSuspend:
		return "resume_suspend"
	case pbtAPMResumeAutomatic:
		return "resume_automatic"
	case pbtPowerSettingChange:
		return "power_setting_change"
	default:
		return fmt.Sprintf("%d", wParam)
	}
}

func lowWord(value uintptr) uint16 {
	return uint16(value & 0xffff)
}

func highWord(value uintptr) uint16 {
	return uint16((value >> 16) & 0xffff)
}
