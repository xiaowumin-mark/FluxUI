//go:build windows

package app

import (
	"image/color"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

const (
	nativeSWPNoSize     = 0x0001
	nativeSWPNoMove     = 0x0002
	nativeSWPNoZOrder   = 0x0004
	nativeSWPNoActivate = 0x0010
	nativeSWPFrame      = 0x0020
	nativeSWPAsync      = 0x4000

	nativeSWHide   = 0
	nativeSWShowNA = 8

	nativeWSVisible     = 0x10000000
	nativeWSDisabled    = 0x08000000
	nativeWSMinimize    = 0x20000000
	nativeWSMaximize    = 0x01000000
	nativeWSCaption     = 0x00C00000
	nativeWSSysMenu     = 0x00080000
	nativeWSMinimizeBox = 0x00020000
	nativeWSMaximizeBox = 0x00010000
	nativeWSSizeBox     = 0x00040000

	nativeSCMaximize  = 0xF030
	nativeMFByCommand = 0x0000
	nativeMFEnabled   = 0x0000
	nativeMFGrayed    = 0x0001

	nativeWMClose            = 0x0010
	nativeWMNCHitTest        = 0x0084
	nativeWMNCCalcSize       = 0x0083
	nativeWMNCPaint          = 0x0085
	nativeWMNCActivate       = 0x0086
	nativeWMNCLButtonDown    = 0x00A1
	nativeWMNCLDblClk        = 0x00A3
	nativeWMNCUAHDrawCaption = 0x00AE
	nativeWMNCUAHDrawFrame   = 0x00AF
	nativeWMNCDestroy        = 0x0082
	nativeHTClient           = 0x0001
	nativeHTCaption          = 0x0002

	nativeRDWInvalidate  = 0x0001
	nativeRDWAllChildren = 0x0080
	nativeRDWFrame       = 0x0400

	nativeDWMWANCRenderingPolicy      = 2
	nativeDWMNCRenderingPolicyEnabled = 2
	nativeDWMWAWindowCornerPreference = 33
	nativeDWMWABorderColor            = 34
	nativeDWMWACaptionColor           = 35
	nativeDWMWATextColor              = 36
	nativeDWMWAColorDefault           = 0xFFFFFFFF
	nativeDWMWAColorNone              = 0xFFFFFFFE
)

var (
	nativeHWNDTopmost   = ^uintptr(0)
	nativeHWNDNotopmost = ^uintptr(0) - 1
	nativeGWLStyle      = ^uintptr(15)
	nativeGWLPWndProc   = ^uintptr(3)

	nativeUser32           = syscall.NewLazyDLL("user32.dll")
	nativeCallWindowProc   = nativeUser32.NewProc("CallWindowProcW")
	nativeDefWindowProc    = nativeUser32.NewProc("DefWindowProcW")
	nativeDrawMenuBar      = nativeUser32.NewProc("DrawMenuBar")
	nativeEnableMenuItem   = nativeUser32.NewProc("EnableMenuItem")
	nativeGetSystemMenu    = nativeUser32.NewProc("GetSystemMenu")
	nativeGetWindowLong    = nativeUser32.NewProc("GetWindowLongW")
	nativeSetWindowPos     = nativeUser32.NewProc("SetWindowPos")
	nativeSetWindowLong    = nativeUser32.NewProc("SetWindowLongW")
	nativeSetWindowLongPtr = nativeUser32.NewProc("SetWindowLongPtrW")
	nativeSetForeground    = nativeUser32.NewProc("SetForegroundWindow")
	nativeEnableNCDpi      = nativeUser32.NewProc("EnableNonClientDpiScaling")
	nativePostMessage      = nativeUser32.NewProc("PostMessageW")
	nativeReleaseCapture   = nativeUser32.NewProc("ReleaseCapture")
	nativeRedrawWindow     = nativeUser32.NewProc("RedrawWindow")
	nativeScreenToClient   = nativeUser32.NewProc("ScreenToClient")
	nativeShowWindowAsync  = nativeUser32.NewProc("ShowWindowAsync")

	nativeDwm              = syscall.NewLazyDLL("dwmapi.dll")
	nativeDwmSetWindowAttr = nativeDwm.NewProc("DwmSetWindowAttribute")
	nativeDwmExtendFrame   = nativeDwm.NewProc("DwmExtendFrameIntoClientArea")

	nativeCloseHookCallback = syscall.NewCallback(nativeWindowProc)
	nativeCloseHookMu       sync.RWMutex
	nativeCloseHookEntries  = make(map[uintptr]*windowEntry)
)

type nativeMargins struct {
	Left   int32
	Right  int32
	Top    int32
	Bottom int32
}

type nativePoint struct {
	X int32
	Y int32
}

func setNativeWindowAlwaysOnTop(handle uintptr, always bool) bool {
	if handle == 0 {
		return false
	}
	insertAfter := nativeHWNDNotopmost
	if always {
		insertAfter = nativeHWNDTopmost
	}
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		nativeSetWindowPos.Call(
			handle,
			insertAfter,
			0,
			0,
			0,
			0,
			nativeSWPNoMove|nativeSWPNoSize|nativeSWPNoActivate|nativeSWPAsync,
		)
	}()
	return true
}

func setNativeWindowVisible(handle uintptr, visible bool) bool {
	if handle == 0 {
		return false
	}
	cmd := uintptr(nativeSWHide)
	if visible {
		cmd = nativeSWShowNA
	}
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		nativeShowWindowAsync.Call(handle, cmd)
	}()
	return true
}

func requestNativeWindowFocus(handle uintptr) bool {
	if handle == 0 {
		return false
	}
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		nativeSetForeground.Call(handle)
	}()
	return true
}

func startNativeWindowDragMove(handle uintptr) bool {
	if handle == 0 {
		return false
	}
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		nativeReleaseCapture.Call()
		nativePostMessage.Call(handle, nativeWMNCLButtonDown, nativeHTCaption, 0)
	}()
	return true
}

func setNativeWindowFrameStyle(handle uintptr, style WindowsFrameStyle, resizable, maximizeEnabled bool) bool {
	if handle == 0 {
		return false
	}
	style = normalizeWindowsFrameStyle(style)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		styleChanged := false
		styleValue, _, err := nativeGetWindowLong.Call(handle, nativeGWLStyle)
		if styleValue != 0 || err == syscall.Errno(0) {
			nextStyle := nativeWindowChromeStyle(styleValue, style.Mode, resizable, maximizeEnabled)
			if nextStyle != styleValue {
				nativeSetWindowLong.Call(handle, nativeGWLStyle, nextStyle)
				styleChanged = true
			}
		}

		enableNativeNonClientDPI(handle)
		setDwmIntAttribute(handle, nativeDWMWANCRenderingPolicy, nativeDWMNCRenderingPolicyEnabled)
		setDwmIntAttribute(handle, nativeDWMWAWindowCornerPreference, int32(windowsCornerPreferenceValue(style.Corner)))
		setDwmColorAttribute(handle, nativeDWMWABorderColor, windowsFrameBorderColor(style))
		setDwmColorAttribute(handle, nativeDWMWACaptionColor, windowsFrameCaptionColor(style))
		setDwmColorAttribute(handle, nativeDWMWATextColor, windowsFrameTextColor(style))
		extendNativeFrame(handle, nativeFrameMargins(style))
		setNativeMaximizeMenuEnabled(handle, maximizeEnabled)
		if styleChanged {
			nativeSetWindowPos.Call(
				handle,
				0,
				0,
				0,
				0,
				0,
				nativeSWPNoMove|nativeSWPNoSize|nativeSWPNoZOrder|nativeSWPNoActivate|nativeSWPFrame|nativeSWPAsync,
			)
			nativeDrawMenuBar.Call(handle)
			nativeRedrawWindow.Call(handle, 0, 0, nativeRDWInvalidate|nativeRDWFrame|nativeRDWAllChildren)
		}
	}()
	return true
}

func probeNativeWindowsChrome() WindowsChromeAvailability {
	return WindowsChromeAvailability{
		Supported:  true,
		FrameStyle: true,
		DragMove:   true,
	}
}

func setNativeWindowPosition(handle uintptr, x, y int) bool {
	if handle == 0 {
		return false
	}
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		nativeSetWindowPos.Call(
			handle,
			0,
			uintptr(x),
			uintptr(y),
			0,
			0,
			nativeSWPNoSize|nativeSWPNoZOrder|nativeSWPNoActivate|nativeSWPAsync,
		)
	}()
	return true
}

func setNativeWindowResizable(handle uintptr, resizable bool) bool {
	if handle == 0 {
		return false
	}
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		style, _, err := nativeGetWindowLong.Call(handle, nativeGWLStyle)
		if style == 0 && err != syscall.Errno(0) {
			return
		}
		nextStyle := style
		if resizable {
			nextStyle |= nativeWSSizeBox
		} else {
			nextStyle &^= nativeWSSizeBox
		}
		if nextStyle == style {
			return
		}
		nativeSetWindowLong.Call(handle, nativeGWLStyle, nextStyle)
		nativeSetWindowPos.Call(
			handle,
			0,
			0,
			0,
			0,
			0,
			nativeSWPNoMove|nativeSWPNoSize|nativeSWPNoZOrder|nativeSWPNoActivate|nativeSWPFrame|nativeSWPAsync,
		)
		nativeDrawMenuBar.Call(handle)
	}()
	return true
}

func setNativeWindowMaximizeEnabled(handle uintptr, enabled bool) bool {
	if handle == 0 {
		return false
	}
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		style, _, err := nativeGetWindowLong.Call(handle, nativeGWLStyle)
		if style == 0 && err != syscall.Errno(0) {
			return
		}
		nextStyle := style
		if enabled {
			nextStyle |= nativeWSMaximizeBox
		} else {
			nextStyle &^= nativeWSMaximizeBox
		}
		if nextStyle == style {
			return
		}
		nativeSetWindowLong.Call(handle, nativeGWLStyle, nextStyle)

		setNativeMaximizeMenuEnabled(handle, enabled)

		nativeSetWindowPos.Call(
			handle,
			0,
			0,
			0,
			0,
			0,
			nativeSWPNoMove|nativeSWPNoSize|nativeSWPNoZOrder|nativeSWPNoActivate|nativeSWPFrame|nativeSWPAsync,
		)
		nativeDrawMenuBar.Call(handle)
	}()
	return true
}

func setNativeMaximizeMenuEnabled(handle uintptr, enabled bool) {
	if handle == 0 {
		return
	}
	if menu, _, _ := nativeGetSystemMenu.Call(handle, 0); menu != 0 {
		flags := uintptr(nativeMFByCommand | nativeMFGrayed)
		if enabled {
			flags = nativeMFByCommand | nativeMFEnabled
		}
		nativeEnableMenuItem.Call(menu, nativeSCMaximize, flags)
	}
}

func nativeWindowChromeStyle(current uintptr, mode WindowsFrameMode, resizable, maximizeEnabled bool) uintptr {
	const preserve = nativeWSVisible | nativeWSDisabled | nativeWSMaximize | nativeWSMinimize
	next := (current & preserve) | nativeWSSysMenu | nativeWSMinimizeBox
	if mode == WindowsFrameDefault {
		next |= nativeWSCaption
	}
	if resizable {
		next |= nativeWSSizeBox
	}
	if maximizeEnabled {
		next |= nativeWSMaximizeBox
	}
	return next
}

func nativeFrameMargins(style WindowsFrameStyle) nativeMargins {
	if style.Mode != WindowsFrameHidden || !style.Shadow {
		return nativeMargins{}
	}
	return nativeMargins{Left: -1, Right: -1, Top: -1, Bottom: -1}
}

func windowsCornerPreferenceValue(corner WindowsCornerPreference) int {
	switch corner {
	case WindowsCornerDoNotRound:
		return 1
	case WindowsCornerRound:
		return 2
	case WindowsCornerRoundSmall:
		return 3
	default:
		return 0
	}
}

func windowsFrameBorderColor(style WindowsFrameStyle) uint32 {
	switch style.Border {
	case WindowsFrameBorderHidden:
		return nativeDWMWAColorNone
	case WindowsFrameBorderColor:
		if style.BorderColor.A > 0 {
			return colorRef(style.BorderColor)
		}
	}
	return nativeDWMWAColorDefault
}

func windowsFrameCaptionColor(style WindowsFrameStyle) uint32 {
	if style.CaptionColor.A > 0 {
		return colorRef(style.CaptionColor)
	}
	return nativeDWMWAColorDefault
}

func windowsFrameTextColor(style WindowsFrameStyle) uint32 {
	if style.TextColor.A > 0 {
		return colorRef(style.TextColor)
	}
	return nativeDWMWAColorDefault
}

func colorRef(col color.NRGBA) uint32 {
	return uint32(col.R) | uint32(col.G)<<8 | uint32(col.B)<<16
}

func enableNativeNonClientDPI(handle uintptr) bool {
	if handle == 0 || nativeEnableNCDpi.Find() != nil {
		return false
	}
	result, _, _ := nativeEnableNCDpi.Call(handle)
	return result != 0
}

func setDwmIntAttribute(handle uintptr, attr uintptr, value int32) bool {
	if handle == 0 || nativeDwmSetWindowAttr.Find() != nil {
		return false
	}
	nativeDwmSetWindowAttr.Call(
		handle,
		attr,
		uintptr(unsafe.Pointer(&value)),
		unsafe.Sizeof(value),
	)
	return true
}

func setDwmColorAttribute(handle uintptr, attr uintptr, value uint32) bool {
	if handle == 0 || nativeDwmSetWindowAttr.Find() != nil {
		return false
	}
	nativeDwmSetWindowAttr.Call(
		handle,
		attr,
		uintptr(unsafe.Pointer(&value)),
		unsafe.Sizeof(value),
	)
	return true
}

func extendNativeFrame(handle uintptr, margins nativeMargins) bool {
	if handle == 0 || nativeDwmExtendFrame.Find() != nil {
		return false
	}
	nativeDwmExtendFrame.Call(handle, uintptr(unsafe.Pointer(&margins)))
	return true
}

func installNativeWindowCloseHook(entry *windowEntry) bool {
	if entry == nil {
		return false
	}

	entry.mu.Lock()
	handle := entry.nativeHandle
	if handle == 0 {
		entry.mu.Unlock()
		return false
	}
	if entry.nativeCloseHooked && entry.nativeCloseHookHandle == handle {
		entry.mu.Unlock()
		return true
	}
	oldHookHandle := entry.nativeCloseHookHandle
	oldHookProc := entry.nativeCloseOldProc
	oldHooked := entry.nativeCloseHooked
	entry.mu.Unlock()

	if oldHooked && oldHookHandle != 0 && oldHookHandle != handle {
		forgetNativeWindowCloseHookHandle(oldHookHandle, entry)
		if oldHookProc != 0 {
			nativeSetWindowLongPtr.Call(oldHookHandle, nativeGWLPWndProc, oldHookProc)
		}
	}

	oldProc, _, err := nativeSetWindowLongPtr.Call(handle, nativeGWLPWndProc, nativeCloseHookCallback)
	if oldProc == 0 && err != syscall.Errno(0) {
		return false
	}

	entry.mu.Lock()
	entry.nativeCloseHookHandle = handle
	entry.nativeCloseHooked = true
	entry.nativeCloseOldProc = oldProc
	entry.mu.Unlock()

	nativeCloseHookMu.Lock()
	nativeCloseHookEntries[handle] = entry
	nativeCloseHookMu.Unlock()
	return true
}

func uninstallNativeWindowCloseHook(entry *windowEntry) {
	if entry == nil {
		return
	}

	entry.mu.Lock()
	handle := entry.nativeCloseHookHandle
	oldProc := entry.nativeCloseOldProc
	hooked := entry.nativeCloseHooked
	entry.nativeCloseHookHandle = 0
	entry.nativeCloseOldProc = 0
	entry.nativeCloseHooked = false
	entry.mu.Unlock()

	if handle != 0 {
		forgetNativeWindowCloseHookHandle(handle, entry)
	}
	if hooked && handle != 0 && oldProc != 0 {
		nativeSetWindowLongPtr.Call(handle, nativeGWLPWndProc, oldProc)
	}
}

func forgetNativeWindowCloseHook(entry *windowEntry) {
	if entry == nil {
		return
	}

	entry.mu.Lock()
	handle := entry.nativeCloseHookHandle
	entry.nativeCloseHookHandle = 0
	entry.nativeCloseOldProc = 0
	entry.nativeCloseHooked = false
	entry.mu.Unlock()

	if handle != 0 {
		forgetNativeWindowCloseHookHandle(handle, entry)
	}
}

func forgetNativeWindowCloseHookHandle(handle uintptr, entry *windowEntry) {
	nativeCloseHookMu.Lock()
	if current := nativeCloseHookEntries[handle]; current == entry {
		delete(nativeCloseHookEntries, handle)
	}
	nativeCloseHookMu.Unlock()
}

func nativeWindowProc(hwnd, msg, wparam, lparam uintptr) uintptr {
	entry := nativeCloseHookEntry(hwnd)
	if msg == nativeWMClose && entry != nil {
		if !entry.handleCloseRequested() {
			return 0
		}
	}

	oldProc := uintptr(0)
	if entry != nil {
		entry.mu.RLock()
		if entry.nativeCloseHooked && entry.nativeCloseHookHandle == hwnd {
			oldProc = entry.nativeCloseOldProc
		}
		entry.mu.RUnlock()
	}

	if msg == nativeWMNCHitTest {
		result := nativeCallDefaultWindowProc(oldProc, hwnd, msg, wparam, lparam)
		if result == nativeHTClient && nativeDragActionHitTest(entry, hwnd, lparam) {
			return nativeHTCaption
		}
		return result
	}

	if entry != nil && nativeHiddenFrame(entry) {
		if nativeSuppressHiddenFrameNonClient(msg) {
			return nativeHiddenFrameNonClientResult(msg)
		}
	}

	result := uintptr(0)
	result = nativeCallDefaultWindowProc(oldProc, hwnd, msg, wparam, lparam)

	if msg == nativeWMNCDestroy && entry != nil {
		forgetNativeWindowCloseHook(entry)
	}
	return result
}

func nativeCallDefaultWindowProc(oldProc, hwnd, msg, wparam, lparam uintptr) uintptr {
	if oldProc != 0 {
		result, _, _ := nativeCallWindowProc.Call(oldProc, hwnd, msg, wparam, lparam)
		return result
	}
	result, _, _ := nativeDefWindowProc.Call(hwnd, msg, wparam, lparam)
	return result
}

func nativeDragActionHitTest(entry *windowEntry, hwnd, lparam uintptr) bool {
	if entry == nil || hwnd == 0 {
		return false
	}
	pt := nativePointFromLParam(lparam)
	if ok, _, _ := nativeScreenToClient.Call(hwnd, uintptr(unsafe.Pointer(&pt))); ok == 0 {
		return false
	}
	return entry.nativeActionMoveAt(int(pt.X), int(pt.Y))
}

func nativePointFromLParam(lparam uintptr) nativePoint {
	return nativePoint{
		X: int32(int16(lparam & 0xffff)),
		Y: int32(int16((lparam >> 16) & 0xffff)),
	}
}

func nativeHiddenFrame(entry *windowEntry) bool {
	if entry == nil {
		return false
	}
	entry.mu.RLock()
	hidden := entry.state.WindowsFrameStyle.Mode == WindowsFrameHidden
	entry.mu.RUnlock()
	return hidden
}

func nativeSuppressHiddenFrameNonClient(msg uintptr) bool {
	switch msg {
	case nativeWMNCPaint,
		nativeWMNCActivate,
		nativeWMNCUAHDrawCaption,
		nativeWMNCUAHDrawFrame:
		return true
	default:
		return false
	}
}

func nativeHiddenFrameNonClientResult(msg uintptr) uintptr {
	if msg == nativeWMNCActivate {
		return 1
	}
	return 0
}

func nativeCloseHookEntry(hwnd uintptr) *windowEntry {
	nativeCloseHookMu.RLock()
	entry := nativeCloseHookEntries[hwnd]
	nativeCloseHookMu.RUnlock()
	return entry
}
