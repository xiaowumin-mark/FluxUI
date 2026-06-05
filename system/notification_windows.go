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
	wmApp                = 0x8000
	wmFluxUINotification = wmApp + 0x465

	nimAdd    = 0x00000000
	nimModify = 0x00000001
	nimDelete = 0x00000002

	nifMessage = 0x00000001
	nifIcon    = 0x00000002
	nifTip     = 0x00000004
	nifInfo    = 0x00000010

	niifInfo    = 0x00000001
	niifWarning = 0x00000002
	niifError   = 0x00000003

	ninBalloonHide      = 0x0403
	ninBalloonTimeout   = 0x0404
	ninBalloonUserClick = 0x0405

	imageIcon = 1

	lrLoadFromFile = 0x00000010
	lrDefaultSize  = 0x00000040

	idiApplication = 32512
)

var (
	kernel32                 = windows.NewLazySystemDLL("kernel32.dll")
	procGetModuleHandle      = kernel32.NewProc("GetModuleHandleW")
	procShellNotifyIcon      = shell32.NewProc("Shell_NotifyIconW")
	procRegisterClassEx      = user32.NewProc("RegisterClassExW")
	procCreateWindowEx       = user32.NewProc("CreateWindowExW")
	procDefWindowProc        = user32.NewProc("DefWindowProcW")
	procGetMessage           = user32.NewProc("GetMessageW")
	procTranslateMessage     = user32.NewProc("TranslateMessage")
	procDispatchMessage      = user32.NewProc("DispatchMessageW")
	procLoadIcon             = user32.NewProc("LoadIconW")
	procLoadImage            = user32.NewProc("LoadImageW")
	procDestroyIcon          = user32.NewProc("DestroyIcon")
	windowsNotificationState = newWindowsNotificationState()
	windowsNotificationProc  = syscall.NewCallback(windowsNotificationWindowProc)
)

type notifyIconData struct {
	cbSize           uint32
	hWnd             uintptr
	uID              uint32
	uFlags           uint32
	uCallbackMessage uint32
	hIcon            uintptr
	szTip            [128]uint16
	dwState          uint32
	dwStateMask      uint32
	szInfo           [256]uint16
	uTimeout         uint32
	szInfoTitle      [64]uint16
	dwInfoFlags      uint32
	guidItem         windows.GUID
	hBalloonIcon     uintptr
}

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type windowsMessage struct {
	hWnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      struct {
		x int32
		y int32
	}
}

type windowsNotificationCallback struct {
	group   string
	onClick func(NotificationEvent)
}

type windowsNotificationEntry struct {
	data        notifyIconData
	icon        uintptr
	destroyIcon bool
	callback    windowsNotificationCallback
}

type windowsNotificationWindowState struct {
	once    sync.Once
	ready   chan struct{}
	mu      sync.Mutex
	hWnd    uintptr
	err     error
	nextID  uint32
	entries map[uint32]*windowsNotificationEntry
}

func newWindowsNotificationState() *windowsNotificationWindowState {
	return &windowsNotificationWindowState{
		ready:   make(chan struct{}),
		entries: make(map[uint32]*windowsNotificationEntry),
	}
}

func (windowsDriver) notify(ctx context.Context, opts notificationOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	hWnd, err := windowsNotificationState.window()
	if err != nil {
		return fmt.Errorf("system: %s: initialize notification window: %w", CapabilityNotification, err)
	}

	icon, destroyIcon, err := windowsNotificationIcon(opts.icon)
	if err != nil {
		return fmt.Errorf("system: %s: load notification icon: %w", CapabilityNotification, err)
	}

	id := windowsNotificationState.nextIconID()
	data, err := newWindowsNotificationData(hWnd, id, icon, opts)
	if err != nil {
		if destroyIcon {
			destroyWindowsIcon(icon)
		}
		return err
	}

	entry := &windowsNotificationEntry{
		data:        data,
		icon:        icon,
		destroyIcon: destroyIcon,
		callback: windowsNotificationCallback{
			group:   opts.group,
			onClick: opts.onClick,
		},
	}

	addData := data
	addData.uFlags = nifMessage | nifIcon | nifTip
	if err := shellNotifyIcon(nimAdd, &addData); err != nil {
		if destroyIcon {
			destroyWindowsIcon(icon)
		}
		return fmt.Errorf("system: %s: add notification icon: %w", CapabilityNotification, err)
	}

	windowsNotificationState.setEntry(id, entry)

	data.uFlags = nifInfo
	if err := ctx.Err(); err != nil {
		windowsNotificationState.cleanup(id)
		return err
	}
	if err := shellNotifyIcon(nimModify, &data); err != nil {
		windowsNotificationState.cleanup(id)
		return fmt.Errorf("system: %s: show notification: %w", CapabilityNotification, err)
	}

	go windowsNotificationState.cleanupAfter(id, windowsNotificationFallbackCleanupDelay(opts.timeout))
	return nil
}

func (s *windowsNotificationWindowState) window() (uintptr, error) {
	s.once.Do(func() {
		go s.run()
	})
	<-s.ready

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hWnd, s.err
}

func (s *windowsNotificationWindowState) run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hWnd, err := createWindowsNotificationWindow()

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

func (s *windowsNotificationWindowState) nextIconID() uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	if s.nextID == 0 {
		s.nextID = 1
	}
	return s.nextID
}

func (s *windowsNotificationWindowState) setEntry(id uint32, entry *windowsNotificationEntry) {
	s.mu.Lock()
	s.entries[id] = entry
	s.mu.Unlock()
}

func (s *windowsNotificationWindowState) entry(id uint32) *windowsNotificationEntry {
	s.mu.Lock()
	entry := s.entries[id]
	s.mu.Unlock()
	return entry
}

func (s *windowsNotificationWindowState) removeEntry(id uint32) *windowsNotificationEntry {
	s.mu.Lock()
	entry := s.entries[id]
	delete(s.entries, id)
	s.mu.Unlock()
	return entry
}

func (s *windowsNotificationWindowState) click(id uint32) {
	entry := s.entry(id)
	if entry == nil || entry.callback.onClick == nil {
		return
	}

	go entry.callback.onClick(NotificationEvent{
		Kind:  NotificationEventClicked,
		Group: entry.callback.group,
	})
}

func (s *windowsNotificationWindowState) cleanup(id uint32) {
	entry := s.removeEntry(id)
	if entry == nil {
		return
	}
	_ = shellNotifyIcon(nimDelete, &entry.data)
	if entry.destroyIcon {
		destroyWindowsIcon(entry.icon)
	}
}

func (s *windowsNotificationWindowState) cleanupAfter(id uint32, delay time.Duration) {
	time.Sleep(delay)
	s.cleanup(id)
}

func createWindowsNotificationWindow() (uintptr, error) {
	hInstance, err := windowsModuleHandle()
	if err != nil {
		return 0, err
	}

	className, err := windows.UTF16PtrFromString(windowsNotificationClassName())
	if err != nil {
		return 0, err
	}

	wc := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   windowsNotificationProc,
		hInstance:     hInstance,
		lpszClassName: className,
	}

	r, _, callErr := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if r == 0 && callErr != syscall.Errno(1410) {
		if callErr != syscall.Errno(0) {
			return 0, fmt.Errorf("register notification window class: %w", callErr)
		}
		return 0, fmt.Errorf("register notification window class: %w", ErrUnavailable)
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
			return 0, fmt.Errorf("create notification window: %w", callErr)
		}
		return 0, fmt.Errorf("create notification window: %w", ErrUnavailable)
	}

	return hWnd, nil
}

func windowsNotificationWindowProc(hWnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	if msg == wmFluxUINotification {
		id := uint32(wParam)
		eventKind, shouldCleanup := windowsNotificationEvent(uint32(lParam))
		if eventKind == NotificationEventClicked {
			windowsNotificationState.click(id)
		}
		if shouldCleanup {
			windowsNotificationState.cleanup(id)
		}
		return 0
	}

	r, _, _ := procDefWindowProc.Call(hWnd, uintptr(msg), wParam, lParam)
	return r
}

func windowsNotificationClassName() string {
	return fmt.Sprintf("FluxUINotificationWindow-%d-%x", os.Getpid(), uintptr(windowsNotificationProc))
}

func windowsNotificationEvent(lParam uint32) (NotificationEventKind, bool) {
	switch lParam {
	case ninBalloonUserClick:
		return NotificationEventClicked, true
	case ninBalloonHide, ninBalloonTimeout:
		return "", true
	default:
		return "", false
	}
}

func newWindowsNotificationData(hWnd uintptr, id uint32, icon uintptr, opts notificationOptions) (notifyIconData, error) {
	title := opts.title
	if title == "" {
		title = "FluxUI"
	}

	tip := title
	if tip == "" {
		tip = opts.body
	}

	data := notifyIconData{
		cbSize:           uint32(unsafe.Sizeof(notifyIconData{})),
		hWnd:             hWnd,
		uID:              id,
		uFlags:           nifMessage | nifIcon | nifTip | nifInfo,
		uCallbackMessage: wmFluxUINotification,
		hIcon:            icon,
		uTimeout:         windowsNotificationTimeoutMilliseconds(opts.timeout),
		dwInfoFlags:      windowsNotificationInfoFlags(opts.kind),
	}
	if err := copyWindowsUTF16Fixed(data.szTip[:], tip); err != nil {
		return notifyIconData{}, fmt.Errorf("system: configure notification tooltip: %w", err)
	}
	if err := copyWindowsUTF16Fixed(data.szInfoTitle[:], title); err != nil {
		return notifyIconData{}, fmt.Errorf("system: configure notification title: %w", err)
	}
	if err := copyWindowsUTF16Fixed(data.szInfo[:], opts.body); err != nil {
		return notifyIconData{}, fmt.Errorf("system: configure notification body: %w", err)
	}
	return data, nil
}

func windowsNotificationIcon(path string) (uintptr, bool, error) {
	if path != "" {
		path16, err := windows.UTF16PtrFromString(path)
		if err != nil {
			return 0, false, err
		}
		icon, _, callErr := procLoadImage.Call(
			0,
			uintptr(unsafe.Pointer(path16)),
			imageIcon,
			0,
			0,
			lrLoadFromFile|lrDefaultSize,
		)
		if icon == 0 {
			if callErr != syscall.Errno(0) {
				return 0, false, callErr
			}
			return 0, false, ErrUnavailable
		}
		return icon, true, nil
	}

	icon, _, callErr := procLoadIcon.Call(0, idiApplication)
	if icon == 0 {
		if callErr != syscall.Errno(0) {
			return 0, false, callErr
		}
		return 0, false, ErrUnavailable
	}
	return icon, false, nil
}

func windowsNotificationInfoFlags(kind NotificationKind) uint32 {
	switch kind {
	case NotificationWarning:
		return niifWarning
	case NotificationError:
		return niifError
	default:
		return niifInfo
	}
}

func windowsNotificationTimeoutMilliseconds(timeout time.Duration) uint32 {
	if timeout <= 0 {
		return 10_000
	}
	if timeout < time.Second {
		return 1_000
	}
	if timeout > time.Minute {
		return 60_000
	}
	return uint32(timeout / time.Millisecond)
}

func windowsNotificationFallbackCleanupDelay(timeout time.Duration) time.Duration {
	if timeout > 2*time.Minute {
		return timeout + 30*time.Second
	}
	return 2 * time.Minute
}

func copyWindowsUTF16Fixed(dst []uint16, value string) error {
	if len(dst) == 0 {
		return nil
	}

	encoded, err := windows.UTF16FromString(value)
	if err != nil {
		return err
	}
	n := len(encoded) - 1
	if n > len(dst)-1 {
		n = len(dst) - 1
	}
	copy(dst, encoded[:n])
	dst[n] = 0
	return nil
}

func shellNotifyIcon(message uint32, data *notifyIconData) error {
	r, _, callErr := procShellNotifyIcon.Call(uintptr(message), uintptr(unsafe.Pointer(data)))
	if r == 0 {
		if callErr != syscall.Errno(0) {
			return callErr
		}
		return ErrUnavailable
	}
	return nil
}

func destroyWindowsIcon(icon uintptr) {
	if icon == 0 {
		return
	}
	procDestroyIcon.Call(icon)
}

func windowsModuleHandle() (uintptr, error) {
	hInstance, _, callErr := procGetModuleHandle.Call(0)
	if hInstance == 0 {
		if callErr != syscall.Errno(0) {
			return 0, callErr
		}
		return 0, ErrUnavailable
	}
	return hInstance, nil
}
