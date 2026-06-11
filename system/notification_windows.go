//go:build windows

package system

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
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

	defaultToastAppID = "FluxUI"

	windowsToastReadyLine    = "FLUXUI_TOAST_READY"
	windowsToastEventPrefix  = "FLUXUI_TOAST_EVENT|"
	windowsToastActionPrefix = "fluxui:action:"
	windowsToastClickPrefix  = "fluxui:click:"
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
	group     string
	onClick   func(NotificationEvent)
	onDismiss func(NotificationEvent)
}

type windowsToastCallbacks struct {
	group     string
	actionIDs map[string]bool
	onClick   func(NotificationEvent)
	onDismiss func(NotificationEvent)
	onAction  func(NotificationEvent)
}

type windowsToastEntry struct {
	appID string
	stop  func()
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
	groups  map[string]uint32
	toasts  map[string]windowsToastEntry
}

func newWindowsNotificationState() *windowsNotificationWindowState {
	return &windowsNotificationWindowState{
		ready:   make(chan struct{}),
		entries: make(map[uint32]*windowsNotificationEntry),
		groups:  make(map[string]uint32),
		toasts:  make(map[string]windowsToastEntry),
	}
}

func (windowsDriver) notify(ctx context.Context, opts notificationOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateWindowsNotificationBackend(opts.backend); err != nil {
		return err
	}
	if windowsNotificationUsesToast(opts) {
		return showWindowsToastNotification(ctx, opts)
	}
	if len(opts.actions) > 0 {
		return fmt.Errorf("system: %s: notification actions require Toast: %w", CapabilityNotification, ErrUnsupported)
	}
	if err := windowsNotificationState.replaceGroup(opts.group); err != nil {
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
			group:     opts.group,
			onClick:   opts.onClick,
			onDismiss: opts.onDismiss,
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

func (windowsDriver) cancelNotificationGroup(group string) error {
	windowsNotificationState.cleanupGroup(group)
	return windowsNotificationState.cleanupToastGroup(group)
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
	if entry != nil && entry.callback.group != "" {
		s.groups[entry.callback.group] = id
	}
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
	if entry != nil && entry.callback.group != "" && s.groups[entry.callback.group] == id {
		delete(s.groups, entry.callback.group)
	}
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

func (s *windowsNotificationWindowState) dismiss(id uint32) {
	entry := s.entry(id)
	if entry == nil || entry.callback.onDismiss == nil {
		return
	}

	go entry.callback.onDismiss(NotificationEvent{
		Kind:  NotificationEventDismissed,
		Group: entry.callback.group,
	})
}

func (s *windowsNotificationWindowState) cleanupGroup(group string) {
	if group == "" {
		return
	}

	s.mu.Lock()
	id := s.groups[group]
	s.mu.Unlock()
	if id != 0 {
		s.cleanup(id)
	}
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

func (s *windowsNotificationWindowState) replaceGroup(group string) error {
	if group == "" {
		return nil
	}
	s.cleanupGroup(group)
	return s.cleanupToastGroup(group)
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
		if eventKind == NotificationEventDismissed {
			windowsNotificationState.dismiss(id)
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
		return NotificationEventDismissed, true
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

var shellNotifyIcon = defaultShellNotifyIcon

func defaultShellNotifyIcon(message uint32, data *notifyIconData) error {
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

func windowsNotificationUsesToast(opts notificationOptions) bool {
	switch opts.backend {
	case NotificationBackendToast:
		return true
	case NotificationBackendBalloon:
		return false
	default:
		return len(opts.actions) > 0
	}
}

func validateWindowsNotificationBackend(backend NotificationBackend) error {
	switch backend {
	case NotificationBackendAuto, NotificationBackendBalloon, NotificationBackendToast:
		return nil
	default:
		return fmt.Errorf("system: unknown notification backend %q", backend)
	}
}

func showWindowsToastNotification(ctx context.Context, opts notificationOptions) error {
	if opts.backend == NotificationBackendBalloon && len(opts.actions) > 0 {
		return fmt.Errorf("system: %s: balloon notifications do not support actions: %w", CapabilityNotification, ErrUnsupported)
	}
	if err := validateWindowsToastOptions(opts); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := windowsNotificationState.replaceGroup(opts.group); err != nil {
		return err
	}

	appID := windowsToastAppID(opts.appID)
	xmlText, err := windowsToastXML(opts)
	if err != nil {
		return err
	}
	listenFor := windowsToastEventListenDuration(opts)
	script := windowsToastPowerShellScript(appID, opts.group, xmlText, listenFor)

	if listenFor > 0 {
		listener, err := startWindowsToastEventProcess(ctx, script, windowsToastCallbacksFromOptions(opts))
		if err != nil {
			return err
		}
		windowsNotificationState.rememberToastGroup(opts.group, appID, listener.stop)
		return nil
	}

	cmd := exec.CommandContext(ctx,
		"powershell.exe",
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy",
		"Bypass",
		"-EncodedCommand",
		powershellEncodedCommand(script),
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return fmt.Errorf("system: %s: show toast notification: %w: %s", CapabilityNotification, ErrUnavailable, strings.TrimSpace(string(output)))
	}
	windowsNotificationState.rememberToastGroup(opts.group, appID, nil)
	return nil
}

func windowsToastCallbacksFromOptions(opts notificationOptions) windowsToastCallbacks {
	actionIDs := make(map[string]bool, len(opts.actions))
	for _, action := range opts.actions {
		if action.URI == "" {
			actionIDs[action.ID] = true
		}
	}
	return windowsToastCallbacks{
		group:     opts.group,
		actionIDs: actionIDs,
		onClick:   opts.onClick,
		onDismiss: opts.onDismiss,
		onAction:  opts.onAction,
	}
}

func windowsToastEventListenDuration(opts notificationOptions) time.Duration {
	if opts.onClick == nil && opts.onDismiss == nil && opts.onAction == nil {
		return 0
	}
	timeout := opts.timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	listenFor := timeout + 30*time.Second
	if listenFor < 30*time.Second {
		listenFor = 30 * time.Second
	}
	if listenFor > 5*time.Minute {
		listenFor = 5 * time.Minute
	}
	return listenFor
}

type windowsToastListener struct {
	stop func()
}

func startWindowsToastEventProcess(ctx context.Context, script string, callbacks windowsToastCallbacks) (windowsToastListener, error) {
	cmd := exec.CommandContext(ctx,
		"powershell.exe",
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy",
		"Bypass",
		"-EncodedCommand",
		powershellEncodedCommand(script),
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return windowsToastListener{}, fmt.Errorf("system: %s: prepare toast listener stdout: %w", CapabilityNotification, ErrUnavailable)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return windowsToastListener{}, fmt.Errorf("system: %s: prepare toast listener stderr: %w", CapabilityNotification, ErrUnavailable)
	}
	if err := cmd.Start(); err != nil {
		return windowsToastListener{}, fmt.Errorf("system: %s: start toast listener: %w: %v", CapabilityNotification, ErrUnavailable, err)
	}

	ready := make(chan struct{})
	scanErr := make(chan error, 1)
	waitErr := make(chan error, 1)
	stderrText := make(chan string, 1)
	var readyOnce sync.Once

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == windowsToastReadyLine {
				readyOnce.Do(func() { close(ready) })
				continue
			}
			if strings.HasPrefix(line, windowsToastEventPrefix) {
				dispatchWindowsToastEventLine(line, callbacks)
			}
		}
		scanErr <- scanner.Err()
	}()
	go func() {
		data, _ := io.ReadAll(stderr)
		stderrText <- strings.TrimSpace(string(data))
	}()
	go func() {
		waitErr <- cmd.Wait()
	}()

	select {
	case <-ready:
		return windowsToastListener{stop: func() {
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
		}}, nil
	case err := <-waitErr:
		stderrOutput := ""
		select {
		case stderrOutput = <-stderrText:
		default:
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return windowsToastListener{}, ctxErr
		}
		if err != nil {
			return windowsToastListener{}, fmt.Errorf("system: %s: show toast notification: %w: %s", CapabilityNotification, ErrUnavailable, strings.TrimSpace(stderrOutput))
		}
		return windowsToastListener{}, fmt.Errorf("system: %s: toast listener exited before ready: %w: %s", CapabilityNotification, ErrUnavailable, strings.TrimSpace(stderrOutput))
	case err := <-scanErr:
		if err != nil {
			return windowsToastListener{}, fmt.Errorf("system: %s: read toast listener output: %w: %v", CapabilityNotification, ErrUnavailable, err)
		}
		return windowsToastListener{}, fmt.Errorf("system: %s: toast listener closed before ready: %w", CapabilityNotification, ErrUnavailable)
	case <-ctx.Done():
		return windowsToastListener{}, ctx.Err()
	case <-time.After(10 * time.Second):
		_ = cmd.Process.Kill()
		return windowsToastListener{}, fmt.Errorf("system: %s: toast listener did not report ready: %w", CapabilityNotification, ErrUnavailable)
	}
}

func dispatchWindowsToastEventLine(line string, callbacks windowsToastCallbacks) {
	kind, arg, ok := parseWindowsToastEventLine(line)
	if !ok {
		return
	}
	switch kind {
	case "activated":
		if strings.HasPrefix(arg, windowsToastActionPrefix) {
			action := strings.TrimPrefix(arg, windowsToastActionPrefix)
			if callbacks.actionIDs[action] && callbacks.onAction != nil {
				go callbacks.onAction(NotificationEvent{
					Kind:   NotificationEventAction,
					Group:  callbacks.group,
					Action: action,
				})
			}
			return
		}
		if callbacks.onClick != nil {
			go callbacks.onClick(NotificationEvent{
				Kind:  NotificationEventClicked,
				Group: callbacks.group,
			})
		}
	case "dismissed":
		if callbacks.onDismiss != nil {
			go callbacks.onDismiss(NotificationEvent{
				Kind:  NotificationEventDismissed,
				Group: callbacks.group,
			})
		}
	}
}

func parseWindowsToastEventLine(line string) (kind string, arg string, ok bool) {
	if !strings.HasPrefix(line, windowsToastEventPrefix) {
		return "", "", false
	}
	parts := strings.SplitN(strings.TrimPrefix(line, windowsToastEventPrefix), "|", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "", "", false
	}
	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", false
	}
	return parts[0], string(decoded), true
}

func validateWindowsToastOptions(opts notificationOptions) error {
	if opts.launchURI != "" && !windowsToastLooksLikeURI(opts.launchURI) {
		return fmt.Errorf("system: notification launch URI %q is invalid", opts.launchURI)
	}
	for i, action := range opts.actions {
		actionID := strings.TrimSpace(action.ID)
		actionURI := strings.TrimSpace(action.URI)
		if actionID == "" && actionURI == "" {
			return fmt.Errorf("system: notification action %d has empty ID", i)
		}
		if strings.TrimSpace(action.Label) == "" {
			return fmt.Errorf("system: notification action %q has empty label", action.ID)
		}
		if actionURI != "" && !windowsToastLooksLikeURI(actionURI) {
			return fmt.Errorf("system: notification action %q has invalid URI %q", action.ID, action.URI)
		}
	}
	return nil
}

func windowsToastAppID(appID string) string {
	appID = strings.TrimSpace(appID)
	if appID == "" {
		return defaultToastAppID
	}
	return appID
}

func windowsToastXML(opts notificationOptions) (string, error) {
	title := opts.title
	if title == "" {
		title = "FluxUI"
	}

	var b strings.Builder
	b.WriteString(`<toast`)
	if opts.launchURI != "" {
		b.WriteString(` activationType="protocol" launch="`)
		b.WriteString(escapeWindowsToastXMLAttr(opts.launchURI))
		b.WriteString(`"`)
	} else {
		b.WriteString(` launch="`)
		b.WriteString(escapeWindowsToastXMLAttr(windowsToastClickPrefix + opts.group))
		b.WriteString(`"`)
	}
	b.WriteString(`><visual><binding template="ToastGeneric">`)
	b.WriteString(`<text>`)
	b.WriteString(escapeWindowsToastXMLText(title))
	b.WriteString(`</text>`)
	if opts.body != "" {
		b.WriteString(`<text>`)
		b.WriteString(escapeWindowsToastXMLText(opts.body))
		b.WriteString(`</text>`)
	}
	if opts.icon != "" {
		b.WriteString(`<image placement="appLogoOverride" src="`)
		b.WriteString(escapeWindowsToastXMLAttr(opts.icon))
		b.WriteString(`"/>`)
	}
	b.WriteString(`</binding></visual>`)
	if len(opts.actions) > 0 {
		b.WriteString(`<actions>`)
		for _, action := range opts.actions {
			b.WriteString(`<action activationType="`)
			if action.URI != "" {
				b.WriteString(`protocol" arguments="`)
				b.WriteString(escapeWindowsToastXMLAttr(action.URI))
			} else {
				b.WriteString(`foreground" arguments="`)
				b.WriteString(escapeWindowsToastXMLAttr(windowsToastActionPrefix + action.ID))
			}
			b.WriteString(`" content="`)
			b.WriteString(escapeWindowsToastXMLAttr(action.Label))
			b.WriteString(`"/>`)
		}
		b.WriteString(`</actions>`)
	}
	b.WriteString(`</toast>`)
	return b.String(), nil
}

func windowsToastLooksLikeURI(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	schemeEnd := strings.IndexByte(value, ':')
	if schemeEnd <= 0 {
		return false
	}
	for i, r := range value[:schemeEnd] {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			continue
		}
		if i > 0 && ((r >= '0' && r <= '9') || r == '+' || r == '-' || r == '.') {
			continue
		}
		return false
	}
	return true
}

func escapeWindowsToastXMLText(value string) string {
	value = strings.ReplaceAll(value, "&", "&amp;")
	value = strings.ReplaceAll(value, "<", "&lt;")
	value = strings.ReplaceAll(value, ">", "&gt;")
	return value
}

func escapeWindowsToastXMLAttr(value string) string {
	value = escapeWindowsToastXMLText(value)
	value = strings.ReplaceAll(value, `"`, "&quot;")
	value = strings.ReplaceAll(value, `'`, "&apos;")
	return value
}

func windowsToastPowerShellScript(appID, group, xmlText string, listenFor time.Duration) string {
	xmlBase64 := base64.StdEncoding.EncodeToString([]byte(xmlText))
	var b strings.Builder
	b.WriteString(`$ErrorActionPreference='Stop';`)
	b.WriteString(`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null;`)
	b.WriteString(`[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null;`)
	b.WriteString(`$xml=[Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('`)
	b.WriteString(xmlBase64)
	b.WriteString(`'));`)
	b.WriteString(`$doc=[Windows.Data.Xml.Dom.XmlDocument]::new();`)
	b.WriteString(`$doc.LoadXml($xml);`)
	b.WriteString(`$toast=[Windows.UI.Notifications.ToastNotification]::new($doc);`)
	if group != "" {
		b.WriteString(`$toast.Group='`)
		b.WriteString(escapePowerShellSingleQuoted(group))
		b.WriteString(`';`)
		b.WriteString(`$toast.Tag='`)
		b.WriteString(escapePowerShellSingleQuoted(group))
		b.WriteString(`';`)
	}
	b.WriteString(`$notifier=[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('`)
	b.WriteString(escapePowerShellSingleQuoted(appID))
	b.WriteString(`');`)
	if listenFor > 0 {
		listenMilliseconds := int64(listenFor / time.Millisecond)
		b.WriteString(`$sid='FluxUIToast'+[Guid]::NewGuid().ToString('N');`)
		b.WriteString(`$activatedId=$sid+'Activated';$dismissedId=$sid+'Dismissed';`)
		b.WriteString(`$activatedSub=Register-ObjectEvent -InputObject $toast -EventName Activated -SourceIdentifier $activatedId;`)
		b.WriteString(`$dismissedSub=Register-ObjectEvent -InputObject $toast -EventName Dismissed -SourceIdentifier $dismissedId;`)
		b.WriteString(`try{`)
		b.WriteString(`$notifier.Show($toast);`)
		b.WriteString(`Write-Output '`)
		b.WriteString(windowsToastReadyLine)
		b.WriteString(`';`)
		b.WriteString(`$deadline=[DateTime]::UtcNow.AddMilliseconds(`)
		b.WriteString(fmt.Sprintf("%d", listenMilliseconds))
		b.WriteString(`);`)
		b.WriteString(`while([DateTime]::UtcNow -lt $deadline){`)
		b.WriteString(`$event=Wait-Event -Timeout 1;`)
		b.WriteString(`if($null -eq $event){continue;}`)
		b.WriteString(`try{`)
		b.WriteString(`if($event.SourceIdentifier -eq $activatedId){`)
		b.WriteString(`$argsText='';if($null -ne $event.SourceEventArgs -and $null -ne $event.SourceEventArgs.Arguments){$argsText=[string]$event.SourceEventArgs.Arguments};`)
		b.WriteString(`$encoded=[Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($argsText));`)
		b.WriteString(`Write-Output ('`)
		b.WriteString(windowsToastEventPrefix)
		b.WriteString(`activated|' + $encoded);break;`)
		b.WriteString(`}`)
		b.WriteString(`elseif($event.SourceIdentifier -eq $dismissedId){`)
		b.WriteString(`$reason='';if($null -ne $event.SourceEventArgs -and $null -ne $event.SourceEventArgs.Reason){$reason=[string]$event.SourceEventArgs.Reason};`)
		b.WriteString(`$encoded=[Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($reason));`)
		b.WriteString(`Write-Output ('`)
		b.WriteString(windowsToastEventPrefix)
		b.WriteString(`dismissed|' + $encoded);break;`)
		b.WriteString(`}`)
		b.WriteString(`}finally{Remove-Event -EventIdentifier $event.EventIdentifier -ErrorAction SilentlyContinue}`)
		b.WriteString(`}`)
		b.WriteString(`}finally{`)
		b.WriteString(`Unregister-Event -SourceIdentifier $activatedId -ErrorAction SilentlyContinue;`)
		b.WriteString(`Unregister-Event -SourceIdentifier $dismissedId -ErrorAction SilentlyContinue;`)
		b.WriteString(`Remove-Job -Id $activatedSub.Id -Force -ErrorAction SilentlyContinue;`)
		b.WriteString(`Remove-Job -Id $dismissedSub.Id -Force -ErrorAction SilentlyContinue;`)
		b.WriteString(`}`)
		return b.String()
	}
	b.WriteString(`$notifier.Show($toast);`)
	return b.String()
}

func powershellEncodedCommand(script string) string {
	encoded := make([]byte, 0, len(script)*2)
	for _, r := range script {
		if r > 0xffff {
			r = '?'
		}
		encoded = append(encoded, byte(r), byte(r>>8))
	}
	return base64.StdEncoding.EncodeToString(encoded)
}

func escapePowerShellSingleQuoted(value string) string {
	return strings.ReplaceAll(value, `'`, `''`)
}

func (s *windowsNotificationWindowState) rememberToastGroup(group, appID string, stop func()) {
	if group == "" {
		return
	}
	s.mu.Lock()
	s.toasts[group] = windowsToastEntry{appID: appID, stop: stop}
	s.mu.Unlock()
}

func (s *windowsNotificationWindowState) takeToastGroup(group string) (windowsToastEntry, bool) {
	if group == "" {
		return windowsToastEntry{}, false
	}
	s.mu.Lock()
	entry, ok := s.toasts[group]
	if ok {
		delete(s.toasts, group)
	}
	s.mu.Unlock()
	return entry, ok
}

func stopWindowsToastEntry(entry windowsToastEntry) {
	if entry.stop != nil {
		entry.stop()
	}
}

func (s *windowsNotificationWindowState) cleanupToastGroup(group string) error {
	entry, ok := s.takeToastGroup(group)
	if !ok {
		return nil
	}
	stopWindowsToastEntry(entry)
	script := windowsToastCancelPowerShellScript(entry.appID, group)
	cmd := exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy",
		"Bypass",
		"-EncodedCommand",
		powershellEncodedCommand(script),
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("system: %s: cancel toast notification group: %w: %s", CapabilityNotification, ErrUnavailable, strings.TrimSpace(string(output)))
	}
	return nil
}

func windowsToastCancelPowerShellScript(appID, group string) string {
	var b strings.Builder
	b.WriteString(`$ErrorActionPreference='Stop';`)
	b.WriteString(`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null;`)
	b.WriteString(`[Windows.UI.Notifications.ToastNotificationManager]::History.RemoveGroup('`)
	b.WriteString(escapePowerShellSingleQuoted(group))
	b.WriteString(`','`)
	b.WriteString(escapePowerShellSingleQuoted(appID))
	b.WriteString(`');`)
	return b.String()
}

func windowsToastProbePowerShellScript(appID string) string {
	var b strings.Builder
	b.WriteString(`$ErrorActionPreference='Stop';`)
	b.WriteString(`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null;`)
	b.WriteString(`$notifier=[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('`)
	b.WriteString(escapePowerShellSingleQuoted(appID))
	b.WriteString(`');`)
	b.WriteString(`if($null -eq $notifier){throw 'CreateToastNotifier returned null'};`)
	return b.String()
}
