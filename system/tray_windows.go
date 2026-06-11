//go:build windows

package system

import (
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	wmFluxUITray = wmApp + 0x466

	wmLButtonUp     = 0x0202
	wmLButtonDblClk = 0x0203
	wmRButtonUp     = 0x0205
	wmContextMenu   = 0x007B
	wmNull          = 0x0000

	mfString    = 0x0000
	mfGrayed    = 0x0001
	mfSeparator = 0x0800
	mfChecked   = 0x0008
	mfPopup     = 0x0010
	mfDefault   = 0x1000

	tpmRightButton = 0x0002
	tpmReturnCmd   = 0x0100

	rtIconVersion = 0x00030000
)

var (
	procCreatePopupMenu       = user32.NewProc("CreatePopupMenu")
	procAppendMenu            = user32.NewProc("AppendMenuW")
	procDestroyMenu           = user32.NewProc("DestroyMenu")
	procTrackPopupMenu        = user32.NewProc("TrackPopupMenu")
	procGetCursorPos          = user32.NewProc("GetCursorPos")
	procSetForegroundWindow   = user32.NewProc("SetForegroundWindow")
	procPostMessage           = user32.NewProc("PostMessageW")
	procRegisterWindowMessage = user32.NewProc("RegisterWindowMessageW")
	procCreateIconFromResourceEx = user32.NewProc("CreateIconFromResourceEx")

	windowsTrayState             = newWindowsTrayState()
	windowsTrayProc              = syscall.NewCallback(windowsTrayWindowProc)
	windowsTaskbarCreatedMessage = registerWindowsMessage("TaskbarCreated")
	trackWindowsTrayMenuFunc     = trackWindowsTrayMenu
)

type windowsPoint struct {
	x int32
	y int32
}

type windowsTrayWindowState struct {
	once    sync.Once
	ready   chan struct{}
	mu      sync.Mutex
	hWnd    uintptr
	err     error
	nextID  uint32
	entries map[uint32]*windowsTrayHandle
}

type windowsTrayHandle struct {
	mu           sync.Mutex
	id           uint32
	data         notifyIconData
	icon         uintptr
	destroyIcon  bool
	menu         TrayMenu
	menuProvider func() TrayMenu
	onClick      func(TrayEvent)
	onDblClick   func(TrayEvent)
	visible      bool
	closed       bool
}

func newWindowsTrayState() *windowsTrayWindowState {
	return &windowsTrayWindowState{
		ready:   make(chan struct{}),
		entries: make(map[uint32]*windowsTrayHandle),
	}
}

func (windowsDriver) newTray(opts trayOptions) (trayHandle, error) {
	hWnd, err := windowsTrayState.window()
	if err != nil {
		return nil, fmt.Errorf("system: %s: initialize tray window: %w", CapabilityTray, err)
	}

	icon, destroyIcon, err := windowsTrayIcon(opts)
	if err != nil {
		return nil, fmt.Errorf("system: %s: load tray icon: %w", CapabilityTray, err)
	}

	id := windowsTrayState.nextIconID()
	data, err := newWindowsTrayData(hWnd, id, icon, opts.tooltip)
	if err != nil {
		if destroyIcon {
			destroyWindowsIcon(icon)
		}
		return nil, err
	}

	handle := &windowsTrayHandle{
		id:           id,
		data:         data,
		icon:         icon,
		destroyIcon:  destroyIcon,
		menu:         cloneTrayMenu(opts.menu),
		menuProvider: opts.menuProvider,
		onClick:      opts.onClick,
		onDblClick:   opts.onDoubleClick,
	}
	runtime.SetFinalizer(handle, func(h *windowsTrayHandle) {
		_ = h.close()
	})
	windowsTrayState.setEntry(id, handle)
	return handle, nil
}

func (s *windowsTrayWindowState) window() (uintptr, error) {
	s.once.Do(func() {
		go s.run()
	})
	<-s.ready

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hWnd, s.err
}

func (s *windowsTrayWindowState) run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hWnd, err := createWindowsTrayWindow()

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

func (s *windowsTrayWindowState) nextIconID() uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	if s.nextID == 0 {
		s.nextID = 1
	}
	return s.nextID
}

func (s *windowsTrayWindowState) setEntry(id uint32, entry *windowsTrayHandle) {
	s.mu.Lock()
	s.entries[id] = entry
	s.mu.Unlock()
}

func (s *windowsTrayWindowState) entry(id uint32) *windowsTrayHandle {
	s.mu.Lock()
	entry := s.entries[id]
	s.mu.Unlock()
	return entry
}

func (s *windowsTrayWindowState) removeEntry(id uint32) *windowsTrayHandle {
	s.mu.Lock()
	entry := s.entries[id]
	delete(s.entries, id)
	s.mu.Unlock()
	return entry
}

func (s *windowsTrayWindowState) restoreVisibleIcons() {
	s.mu.Lock()
	entries := make([]*windowsTrayHandle, 0, len(s.entries))
	for _, entry := range s.entries {
		entries = append(entries, entry)
	}
	s.mu.Unlock()

	for _, entry := range entries {
		entry.restoreAfterTaskbarCreated()
	}
}

func createWindowsTrayWindow() (uintptr, error) {
	hInstance, err := windowsModuleHandle()
	if err != nil {
		return 0, err
	}

	className, err := windows.UTF16PtrFromString(windowsTrayClassName())
	if err != nil {
		return 0, err
	}

	wc := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   windowsTrayProc,
		hInstance:     hInstance,
		lpszClassName: className,
	}

	r, _, callErr := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if r == 0 && callErr != syscall.Errno(1410) {
		if callErr != syscall.Errno(0) {
			return 0, fmt.Errorf("register tray window class: %w", callErr)
		}
		return 0, fmt.Errorf("register tray window class: %w", ErrUnavailable)
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
			return 0, fmt.Errorf("create tray window: %w", callErr)
		}
		return 0, fmt.Errorf("create tray window: %w", ErrUnavailable)
	}

	return hWnd, nil
}

func windowsTrayWindowProc(hWnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	if msg == wmFluxUITray {
		id := uint32(wParam)
		if entry := windowsTrayState.entry(id); entry != nil {
			entry.handleMessage(uint32(lParam))
		}
		return 0
	}
	if windowsTaskbarCreatedMessage != 0 && msg == windowsTaskbarCreatedMessage {
		windowsTrayState.restoreVisibleIcons()
		return 0
	}

	r, _, _ := procDefWindowProc.Call(hWnd, uintptr(msg), wParam, lParam)
	return r
}

func windowsTrayClassName() string {
	return fmt.Sprintf("FluxUITrayWindow-%d-%x", os.Getpid(), uintptr(windowsTrayProc))
}

func newWindowsTrayData(hWnd uintptr, id uint32, icon uintptr, tooltip string) (notifyIconData, error) {
	if tooltip == "" {
		tooltip = "FluxUI"
	}
	data := notifyIconData{
		cbSize:           uint32(unsafe.Sizeof(notifyIconData{})),
		hWnd:             hWnd,
		uID:              id,
		uFlags:           nifMessage | nifIcon | nifTip,
		uCallbackMessage: wmFluxUITray,
		hIcon:            icon,
	}
	if err := copyWindowsUTF16Fixed(data.szTip[:], tooltip); err != nil {
		return notifyIconData{}, fmt.Errorf("system: configure tray tooltip: %w", err)
	}
	return data, nil
}

func (h *windowsTrayHandle) setIcon(path string) error {
	icon, destroyIcon, err := windowsTrayIcon(trayOptions{icon: path})
	if err != nil {
		return fmt.Errorf("system: %s: load tray icon: %w", CapabilityTray, err)
	}
	return h.replaceIcon(icon, destroyIcon)
}

func (h *windowsTrayHandle) setIconData(data []byte) error {
	icon, destroyIcon, err := windowsTrayIcon(trayOptions{iconData: data})
	if err != nil {
		return fmt.Errorf("system: %s: load tray icon: %w", CapabilityTray, err)
	}
	return h.replaceIcon(icon, destroyIcon)
}

func (h *windowsTrayHandle) setIconResource(id uint16) error {
	icon, destroyIcon, err := windowsTrayIcon(trayOptions{iconResource: id})
	if err != nil {
		return fmt.Errorf("system: %s: load tray icon: %w", CapabilityTray, err)
	}
	return h.replaceIcon(icon, destroyIcon)
}

func (h *windowsTrayHandle) replaceIcon(icon uintptr, destroyIcon bool) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		if destroyIcon {
			destroyWindowsIcon(icon)
		}
		return trayClosedError()
	}
	nextData := h.data
	nextData.hIcon = icon
	nextData.uFlags = nifIcon
	if h.visible {
		if err := shellNotifyIcon(nimModify, &nextData); err != nil {
			if destroyIcon {
				destroyWindowsIcon(icon)
			}
			return fmt.Errorf("system: %s: update icon: %w", CapabilityTray, err)
		}
	}

	oldIcon := h.icon
	oldDestroyIcon := h.destroyIcon
	h.data.hIcon = icon
	h.icon = icon
	h.destroyIcon = destroyIcon
	if oldDestroyIcon {
		destroyWindowsIcon(oldIcon)
	}
	return nil
}

func windowsTrayIcon(opts trayOptions) (uintptr, bool, error) {
	if len(opts.iconData) > 0 {
		return windowsIconFromICO(opts.iconData)
	}
	if opts.iconResource != 0 {
		return windowsIconFromResource(opts.iconResource)
	}
	return windowsNotificationIcon(opts.icon)
}

func windowsIconFromResource(id uint16) (uintptr, bool, error) {
	hInstance, err := windowsModuleHandle()
	if err != nil {
		return 0, false, err
	}
	icon, _, callErr := procLoadIcon.Call(hInstance, uintptr(id))
	if icon == 0 {
		if callErr != syscall.Errno(0) {
			return 0, false, callErr
		}
		return 0, false, ErrUnavailable
	}
	return icon, false, nil
}

func windowsIconFromICO(data []byte) (uintptr, bool, error) {
	resource, err := windowsIconResourceDataFromICO(data)
	if err != nil {
		return 0, false, err
	}
	icon, _, callErr := procCreateIconFromResourceEx.Call(
		uintptr(unsafe.Pointer(&resource[0])),
		uintptr(len(resource)),
		1,
		rtIconVersion,
		0,
		0,
		lrDefaultSize,
	)
	if icon == 0 {
		if callErr != syscall.Errno(0) {
			return 0, false, callErr
		}
		return 0, false, ErrUnavailable
	}
	return icon, true, nil
}

func windowsIconResourceDataFromICO(data []byte) ([]byte, error) {
	if len(data) < 22 {
		return nil, fmt.Errorf("ico data too short: %w", ErrUnavailable)
	}
	if binary.LittleEndian.Uint16(data[0:2]) != 0 || binary.LittleEndian.Uint16(data[2:4]) != 1 {
		return nil, fmt.Errorf("ico header invalid: %w", ErrUnavailable)
	}
	count := int(binary.LittleEndian.Uint16(data[4:6]))
	if count <= 0 {
		return nil, fmt.Errorf("ico contains no images: %w", ErrUnavailable)
	}
	entryOffset := 6
	if len(data) < entryOffset+16 {
		return nil, fmt.Errorf("ico directory missing: %w", ErrUnavailable)
	}
	size := int(binary.LittleEndian.Uint32(data[entryOffset+8 : entryOffset+12]))
	offset := int(binary.LittleEndian.Uint32(data[entryOffset+12 : entryOffset+16]))
	if size <= 0 || offset < 0 || offset+size > len(data) {
		return nil, fmt.Errorf("ico image out of range: %w", ErrUnavailable)
	}
	return data[offset : offset+size], nil
}

func (h *windowsTrayHandle) setTooltip(text string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return trayClosedError()
	}

	nextData := h.data
	nextData.uFlags = nifTip
	if err := copyWindowsUTF16Fixed(nextData.szTip[:], text); err != nil {
		return fmt.Errorf("system: %s: update tooltip: %w", CapabilityTray, err)
	}

	if h.visible {
		if err := shellNotifyIcon(nimModify, &nextData); err != nil {
			return fmt.Errorf("system: %s: update tooltip: %w", CapabilityTray, err)
		}
	}
	h.data.szTip = nextData.szTip
	return nil
}

func (h *windowsTrayHandle) setMenu(menu TrayMenu) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return trayClosedError()
	}
	h.menu = cloneTrayMenu(menu)
	return nil
}

func (h *windowsTrayHandle) setMenuProvider(fn func() TrayMenu) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return trayClosedError()
	}
	h.menuProvider = fn
	return nil
}

func (h *windowsTrayHandle) show() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return trayClosedError()
	}
	if h.visible {
		return nil
	}
	data := h.data
	data.uFlags = nifMessage | nifIcon | nifTip
	if err := shellNotifyIcon(nimAdd, &data); err != nil {
		return fmt.Errorf("system: %s: show: %w", CapabilityTray, err)
	}
	h.visible = true
	return nil
}

func (h *windowsTrayHandle) hide() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return trayClosedError()
	}
	if !h.visible {
		return nil
	}
	if err := shellNotifyIcon(nimDelete, &h.data); err != nil {
		return fmt.Errorf("system: %s: hide: %w", CapabilityTray, err)
	}
	h.visible = false
	return nil
}

func (h *windowsTrayHandle) close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return trayClosedError()
	}
	h.closed = true
	runtime.SetFinalizer(h, nil)
	windowsTrayState.removeEntry(h.id)

	var firstErr error
	if h.visible {
		if err := shellNotifyIcon(nimDelete, &h.data); err != nil {
			firstErr = fmt.Errorf("system: %s: close: %w", CapabilityTray, err)
		}
		h.visible = false
	}
	if h.destroyIcon {
		destroyWindowsIcon(h.icon)
		h.destroyIcon = false
	}
	if firstErr != nil {
		return firstErr
	}
	return nil
}

func (h *windowsTrayHandle) restoreAfterTaskbarCreated() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed || !h.visible {
		return
	}
	data := h.data
	data.uFlags = nifMessage | nifIcon | nifTip
	_ = shellNotifyIcon(nimAdd, &data)
}

func (h *windowsTrayHandle) handleMessage(message uint32) {
	switch message {
	case wmLButtonUp:
		h.dispatch(TrayEvent{Kind: TrayEventClicked})
	case wmLButtonDblClk:
		h.dispatch(TrayEvent{Kind: TrayEventDoubleClick})
	case wmRButtonUp, wmContextMenu:
		h.showContextMenu()
	}
}

func (h *windowsTrayHandle) dispatch(event TrayEvent) {
	h.mu.Lock()
	onClick := h.onClick
	onDblClick := h.onDblClick
	h.mu.Unlock()

	switch event.Kind {
	case TrayEventClicked:
		if onClick != nil {
			onClick(event)
		}
	case TrayEventDoubleClick:
		if onDblClick != nil {
			onDblClick(event)
		}
	}
}

func (h *windowsTrayHandle) showContextMenu() {
	h.mu.Lock()
	menu := cloneTrayMenu(h.menu)
	provider := h.menuProvider
	hWnd := h.data.hWnd
	h.mu.Unlock()
	if provider != nil {
		menu = provider()
	}
	if len(menu) == 0 || hWnd == 0 {
		return
	}

	command, items, err := trackWindowsTrayMenuFunc(hWnd, menu)
	if err != nil || command == 0 {
		return
	}
	item, ok := items[command]
	if !ok || item.Disabled {
		return
	}
	if item.OnClick != nil {
		item.OnClick(TrayEvent{Kind: TrayEventMenuItem, ItemID: item.ID})
	}
}

func trackWindowsTrayMenu(hWnd uintptr, menu TrayMenu) (uint32, map[uint32]TrayMenuItem, error) {
	hMenu, _, callErr := procCreatePopupMenu.Call()
	if hMenu == 0 {
		if callErr != syscall.Errno(0) {
			return 0, nil, callErr
		}
		return 0, nil, ErrUnavailable
	}
	defer procDestroyMenu.Call(hMenu)

	items := make(map[uint32]TrayMenuItem)
	nextCommand, err := appendWindowsTrayMenuItems(hMenu, menu, 1, items)
	if err != nil {
		return 0, nil, err
	}
	if nextCommand == 1 {
		return 0, nil, nil
	}

	var pt windowsPoint
	if r, _, callErr := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt))); r == 0 {
		if callErr != syscall.Errno(0) {
			return 0, nil, callErr
		}
		return 0, nil, ErrUnavailable
	}

	procSetForegroundWindow.Call(hWnd)
	command, _, _ := procTrackPopupMenu.Call(
		hMenu,
		tpmRightButton|tpmReturnCmd,
		uintptr(pt.x),
		uintptr(pt.y),
		0,
		hWnd,
		0,
	)
	procPostMessage.Call(hWnd, wmNull, 0, 0)
	return uint32(command), items, nil
}

func appendWindowsTrayMenuItems(hMenu uintptr, menu TrayMenu, nextCommand uint32, items map[uint32]TrayMenuItem) (uint32, error) {
	for _, item := range menu {
		if item.Separator {
			if err := appendWindowsTrayMenuSeparator(hMenu); err != nil {
				return nextCommand, err
			}
			continue
		}

		if len(item.Children) > 0 {
			hSubMenu, _, callErr := procCreatePopupMenu.Call()
			if hSubMenu == 0 {
				if callErr != syscall.Errno(0) {
					return nextCommand, callErr
				}
				return nextCommand, ErrUnavailable
			}
			var err error
			nextCommand, err = appendWindowsTrayMenuItems(hSubMenu, item.Children, nextCommand, items)
			if err != nil {
				procDestroyMenu.Call(hSubMenu)
				return nextCommand, err
			}
			if err := appendWindowsTraySubMenu(hMenu, hSubMenu, item); err != nil {
				procDestroyMenu.Call(hSubMenu)
				return nextCommand, err
			}
			continue
		}

		command := nextCommand
		nextCommand++
		if err := appendWindowsTrayMenuItem(hMenu, command, item); err != nil {
			return nextCommand, err
		}
		items[command] = item
	}
	return nextCommand, nil
}

func appendWindowsTrayMenuSeparator(hMenu uintptr) error {
	r, _, callErr := procAppendMenu.Call(hMenu, mfSeparator, 0, 0)
	if r == 0 {
		if callErr != syscall.Errno(0) {
			return callErr
		}
		return ErrUnavailable
	}
	return nil
}

func appendWindowsTrayMenuItem(hMenu uintptr, command uint32, item TrayMenuItem) error {
	label := item.Label
	if label == "" {
		label = item.ID
	}
	if label == "" {
		label = strconv.FormatUint(uint64(command), 10)
	}
	label16, err := windows.UTF16PtrFromString(label)
	if err != nil {
		return err
	}

	flags := uintptr(mfString)
	if item.Disabled {
		flags |= mfGrayed
	}
	if item.Checked {
		flags |= mfChecked
	}
	if item.Default {
		flags |= mfDefault
	}

	r, _, callErr := procAppendMenu.Call(
		hMenu,
		flags,
		uintptr(command),
		uintptr(unsafe.Pointer(label16)),
	)
	if r == 0 {
		if callErr != syscall.Errno(0) {
			return callErr
		}
		return ErrUnavailable
	}
	return nil
}

func appendWindowsTraySubMenu(hMenu uintptr, hSubMenu uintptr, item TrayMenuItem) error {
	label := item.Label
	if label == "" {
		label = item.ID
	}
	if label == "" {
		label = "Menu"
	}
	label16, err := windows.UTF16PtrFromString(label)
	if err != nil {
		return err
	}

	flags := uintptr(mfString | mfPopup)
	if item.Disabled {
		flags |= mfGrayed
	}
	if item.Default {
		flags |= mfDefault
	}

	r, _, callErr := procAppendMenu.Call(
		hMenu,
		flags,
		hSubMenu,
		uintptr(unsafe.Pointer(label16)),
	)
	if r == 0 {
		if callErr != syscall.Errno(0) {
			return callErr
		}
		return ErrUnavailable
	}
	return nil
}

func registerWindowsMessage(name string) uint32 {
	name16, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return 0
	}
	r, _, _ := procRegisterWindowMessage.Call(uintptr(unsafe.Pointer(name16)))
	return uint32(r)
}
