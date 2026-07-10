//go:build windows

package app

import (
	"image/color"
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

	nativeWMClose                 = 0x0010
	nativeWMNCHitTest             = 0x0084
	nativeWMNCCalcSize            = 0x0083
	nativeWMNCPaint               = 0x0085
	nativeWMNCActivate            = 0x0086
	nativeWMMouseMove             = 0x0200
	nativeWMLButtonDown           = 0x0201
	nativeWMLButtonUp             = 0x0202
	nativeWMNCMouseMove           = 0x00A0
	nativeWMNCLButtonDown         = 0x00A1
	nativeWMNCLButtonUp           = 0x00A2
	nativeWMNCLDblClk             = 0x00A3
	nativeWMNCUAHDrawCaption      = 0x00AE
	nativeWMNCUAHDrawFrame        = 0x00AF
	nativeWMNCDestroy             = 0x0082
	nativeWMNCMouseLeave          = 0x02A2
	nativeWMMouseLeave            = 0x02A3
	nativeWMNCPointerUpdate       = 0x0241
	nativeWMNCPointerDown         = 0x0242
	nativeWMNCPointerUp           = 0x0243
	nativeWMPointerUpdate         = 0x0245
	nativeWMPointerDown           = 0x0246
	nativeWMPointerUp             = 0x0247
	nativeWMPointerCaptureChanged = 0x024C
	nativeHTClient                = 0x0001
	nativeHTCaption               = 0x0002
	nativeHTMaxButton             = 0x0009
	nativeMKLButton               = 0x0001

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

	nativeWMCancelMode = 0x001F
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
	nativeGetCapture       = nativeUser32.NewProc("GetCapture")
	nativeGetWindowLong    = nativeUser32.NewProc("GetWindowLongW")
	nativeIsWindow         = nativeUser32.NewProc("IsWindow")
	nativeIsWindowVisible  = nativeUser32.NewProc("IsWindowVisible")
	nativeSetWindowPos     = nativeUser32.NewProc("SetWindowPos")
	nativeSetWindowLong    = nativeUser32.NewProc("SetWindowLongW")
	nativeSetWindowLongPtr = nativeUser32.NewProc("SetWindowLongPtrW")
	nativeSetForeground    = nativeUser32.NewProc("SetForegroundWindow")
	nativeEnableNCDpi      = nativeUser32.NewProc("EnableNonClientDpiScaling")
	nativePostMessage      = nativeUser32.NewProc("PostMessageW")
	nativeReleaseCapture   = nativeUser32.NewProc("ReleaseCapture")
	nativeRedrawWindow     = nativeUser32.NewProc("RedrawWindow")
	nativeClientToScreen   = nativeUser32.NewProc("ClientToScreen")
	nativeScreenToClient   = nativeUser32.NewProc("ScreenToClient")
	nativeShowWindow       = nativeUser32.NewProc("ShowWindow")

	nativeKernel32           = syscall.NewLazyDLL("kernel32.dll")
	nativeGetCurrentThreadID = nativeKernel32.NewProc("GetCurrentThreadId")
	nativeSetLastError       = nativeKernel32.NewProc("SetLastError")

	nativeDwm              = syscall.NewLazyDLL("dwmapi.dll")
	nativeDwmGetWindowAttr = nativeDwm.NewProc("DwmGetWindowAttribute")
	nativeDwmSetWindowAttr = nativeDwm.NewProc("DwmSetWindowAttribute")
	nativeDwmExtendFrame   = nativeDwm.NewProc("DwmExtendFrameIntoClientArea")
	nativeDwmFlush         = nativeDwm.NewProc("DwmFlush")

	nativeCloseHookCallback = syscall.NewCallback(nativeWindowProc)
	nativeCloseHookMu       sync.RWMutex
	nativeCloseHookEntries  = make(map[uintptr]nativeCloseHookBinding)
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

type nativeCloseHookBinding struct {
	target  nativeWindowTarget
	oldProc uintptr
}

func runOnNativeWindowThread(entry *windowEntry, run func()) bool {
	if entry == nil || entry.win == nil || run == nil || !entry.alive.Load() {
		return false
	}
	current, _, _ := nativeGetCurrentThreadID.Call()
	if owner := entry.nativeWindowThreadID.Load(); owner != 0 && owner == uint32(current) {
		run()
		return true
	}

	completed := false
	entry.win.Run(func() {
		threadID, _, _ := nativeGetCurrentThreadID.Call()
		entry.nativeWindowThreadID.Store(uint32(threadID))
		run()
		completed = true
	})
	return completed
}

func runNativeWindowCommand(target nativeWindowTarget, run func(uintptr) bool) bool {
	if run == nil || !target.valid() {
		return false
	}
	succeeded := false
	if !runOnNativeWindowThread(target.entry, func() {
		if !target.valid() {
			return
		}
		result, _, _ := nativeIsWindow.Call(target.handle)
		if result == 0 {
			return
		}
		succeeded = run(target.handle) && target.valid()
	}) {
		return false
	}
	return succeeded
}

func nativeBoolCall(proc *syscall.LazyProc, args ...uintptr) bool {
	result, _, _ := proc.Call(args...)
	return result != 0
}

func nativeZeroValueCall(proc *syscall.LazyProc, args ...uintptr) (uintptr, bool) {
	// GetWindowLong/SetWindowLong use zero both as a valid value and as their
	// failure sentinel, so clear and inspect the thread's last-error value.
	nativeSetLastError.Call(0)
	result, _, callErr := proc.Call(args...)
	return result, result != 0 || callErr == syscall.Errno(0)
}

func nativeHRESULTCall(proc *syscall.LazyProc, args ...uintptr) bool {
	result, _, _ := proc.Call(args...)
	return int32(result) == 0
}

func setNativeWindowAlwaysOnTop(target nativeWindowTarget, always bool) bool {
	insertAfter := nativeHWNDNotopmost
	if always {
		insertAfter = nativeHWNDTopmost
	}
	return runNativeWindowCommand(target, func(handle uintptr) bool {
		return nativeBoolCall(
			nativeSetWindowPos,
			handle,
			insertAfter,
			0,
			0,
			0,
			0,
			nativeSWPNoMove|nativeSWPNoSize|nativeSWPNoActivate,
		)
	})
}

func setNativeWindowVisible(target nativeWindowTarget, visible bool) bool {
	cmd := uintptr(nativeSWHide)
	if visible {
		cmd = nativeSWShowNA
	}
	return runNativeWindowCommand(target, func(handle uintptr) bool {
		// ShowWindow reports the previous visibility, not success. Verify the
		// resulting state after the synchronous call instead.
		nativeShowWindow.Call(handle, cmd)
		shown, _, _ := nativeIsWindowVisible.Call(handle)
		return (shown != 0) == visible
	})
}

func requestNativeWindowFocus(target nativeWindowTarget) bool {
	return runNativeWindowCommand(target, func(handle uintptr) bool {
		return nativeBoolCall(nativeSetForeground, handle)
	})
}

func startNativeWindowDragMove(target nativeWindowTarget) bool {
	return runNativeWindowCommand(target, func(handle uintptr) bool {
		capture, _, _ := nativeGetCapture.Call()
		if capture != 0 && !nativeBoolCall(nativeReleaseCapture) {
			return false
		}
		return nativeBoolCall(nativePostMessage, handle, nativeWMNCLButtonDown, nativeHTCaption, 0)
	})
}

func setNativeWindowFrameStyle(target nativeWindowTarget, style WindowsFrameStyle, resizable, maximizeEnabled bool) bool {
	style = normalizeWindowsFrameStyle(style)
	return runNativeWindowCommand(target, func(handle uintptr) bool {
		styleValue, ok := nativeZeroValueCall(nativeGetWindowLong, handle, nativeGWLStyle)
		if !ok {
			return false
		}
		nextStyle := nativeWindowChromeStyle(styleValue, style.Mode, resizable, maximizeEnabled)
		oldMaximizeEnabled := styleValue&nativeWSMaximizeBox != 0
		type dwmSnapshot struct {
			attr  uintptr
			value uint32
		}
		var dwmSnapshots []dwmSnapshot
		styleChanged := false
		menuChanged := false
		committed := false
		defer func() {
			if committed {
				return
			}
			if styleChanged {
				_, _ = nativeZeroValueCall(nativeSetWindowLong, handle, nativeGWLStyle, styleValue)
			}
			if menuChanged {
				_ = setNativeMaximizeMenuEnabledHandle(handle, oldMaximizeEnabled)
			}
			for i := len(dwmSnapshots) - 1; i >= 0; i-- {
				snapshot := dwmSnapshots[i]
				_ = setDwmUint32Attribute(handle, snapshot.attr, snapshot.value)
			}
			if styleChanged || menuChanged {
				_ = nativeBoolCall(
					nativeSetWindowPos,
					handle,
					0,
					0,
					0,
					0,
					0,
					nativeSWPNoMove|nativeSWPNoSize|nativeSWPNoZOrder|nativeSWPNoActivate|nativeSWPFrame,
				)
				_ = nativeBoolCall(nativeDrawMenuBar, handle)
				_ = nativeBoolCall(nativeRedrawWindow, handle, 0, 0, nativeRDWInvalidate|nativeRDWFrame|nativeRDWAllChildren)
			}
		}()

		applyDwm := func(attr uintptr, value uint32, required bool) bool {
			previous, available := getDwmUint32Attribute(handle, attr)
			if !available {
				return !required
			}
			if previous == value {
				return true
			}
			if !setDwmUint32Attribute(handle, attr, value) {
				return !required
			}
			dwmSnapshots = append(dwmSnapshots, dwmSnapshot{attr: attr, value: previous})
			return true
		}
		_ = enableNativeNonClientDPI(handle)
		if !applyDwm(nativeDWMWANCRenderingPolicy, nativeDWMNCRenderingPolicyEnabled, false) {
			return false
		}
		if !applyDwm(nativeDWMWAWindowCornerPreference, uint32(windowsCornerPreferenceValue(style.Corner)), style.Corner != WindowsCornerDefault) {
			return false
		}
		if !applyDwm(nativeDWMWABorderColor, windowsFrameBorderColor(style), style.Border != WindowsFrameBorderDefault) {
			return false
		}
		if !applyDwm(nativeDWMWACaptionColor, windowsFrameCaptionColor(style), style.CaptionColor.A > 0) {
			return false
		}
		if !applyDwm(nativeDWMWATextColor, windowsFrameTextColor(style), style.TextColor.A > 0) {
			return false
		}
		if !setNativeMaximizeMenuEnabledHandle(handle, maximizeEnabled) {
			return false
		}
		menuChanged = true
		if nextStyle != styleValue {
			if _, ok := nativeZeroValueCall(nativeSetWindowLong, handle, nativeGWLStyle, nextStyle); !ok {
				return false
			}
			styleChanged = true
		}
		if !nativeBoolCall(
			nativeSetWindowPos,
			handle,
			0,
			0,
			0,
			0,
			0,
			nativeSWPNoMove|nativeSWPNoSize|nativeSWPNoZOrder|nativeSWPNoActivate|nativeSWPFrame,
		) {
			return false
		}
		if !nativeBoolCall(nativeDrawMenuBar, handle) {
			return false
		}
		if !nativeBoolCall(nativeRedrawWindow, handle, 0, 0, nativeRDWInvalidate|nativeRDWFrame|nativeRDWAllChildren) {
			return false
		}
		frameApplied := extendNativeFrame(handle, nativeFrameMargins(style))
		if style.Mode == WindowsFrameHidden && style.Shadow && !frameApplied {
			return false
		}
		_ = flushNativeDwm()
		committed = true
		return true
	})
}

func probeNativeWindowsChrome() WindowsChromeAvailability {
	return WindowsChromeAvailability{
		Supported:  true,
		FrameStyle: true,
		DragMove:   true,
	}
}

func setNativeWindowPosition(target nativeWindowTarget, x, y int) bool {
	return runNativeWindowCommand(target, func(handle uintptr) bool {
		return nativeBoolCall(
			nativeSetWindowPos,
			handle,
			0,
			uintptr(x),
			uintptr(y),
			0,
			0,
			nativeSWPNoSize|nativeSWPNoZOrder|nativeSWPNoActivate,
		)
	})
}

func setNativeWindowResizable(target nativeWindowTarget, resizable bool) bool {
	return runNativeWindowCommand(target, func(handle uintptr) bool {
		style, ok := nativeZeroValueCall(nativeGetWindowLong, handle, nativeGWLStyle)
		if !ok {
			return false
		}
		nextStyle := style
		if resizable {
			nextStyle |= nativeWSSizeBox
		} else {
			nextStyle &^= nativeWSSizeBox
		}
		if nextStyle == style {
			return true
		}
		if _, ok := nativeZeroValueCall(nativeSetWindowLong, handle, nativeGWLStyle, nextStyle); !ok {
			return false
		}
		committed := false
		defer func() {
			if committed {
				return
			}
			_, _ = nativeZeroValueCall(nativeSetWindowLong, handle, nativeGWLStyle, style)
			_ = nativeBoolCall(
				nativeSetWindowPos,
				handle,
				0,
				0,
				0,
				0,
				0,
				nativeSWPNoMove|nativeSWPNoSize|nativeSWPNoZOrder|nativeSWPNoActivate|nativeSWPFrame,
			)
			_ = nativeBoolCall(nativeDrawMenuBar, handle)
		}()
		if !nativeBoolCall(
			nativeSetWindowPos,
			handle,
			0,
			0,
			0,
			0,
			0,
			nativeSWPNoMove|nativeSWPNoSize|nativeSWPNoZOrder|nativeSWPNoActivate|nativeSWPFrame,
		) {
			return false
		}
		if !nativeBoolCall(nativeDrawMenuBar, handle) {
			return false
		}
		committed = true
		return true
	})
}

func setNativeWindowMaximizeEnabled(target nativeWindowTarget, enabled bool) bool {
	return runNativeWindowCommand(target, func(handle uintptr) bool {
		style, ok := nativeZeroValueCall(nativeGetWindowLong, handle, nativeGWLStyle)
		if !ok {
			return false
		}
		nextStyle := style
		oldEnabled := style&nativeWSMaximizeBox != 0
		if enabled {
			nextStyle |= nativeWSMaximizeBox
		} else {
			nextStyle &^= nativeWSMaximizeBox
		}
		if nextStyle != style {
			if _, ok := nativeZeroValueCall(nativeSetWindowLong, handle, nativeGWLStyle, nextStyle); !ok {
				return false
			}
		}
		committed := false
		defer func() {
			if committed {
				return
			}
			if nextStyle != style {
				_, _ = nativeZeroValueCall(nativeSetWindowLong, handle, nativeGWLStyle, style)
			}
			_ = setNativeMaximizeMenuEnabledHandle(handle, oldEnabled)
			_ = nativeBoolCall(
				nativeSetWindowPos,
				handle,
				0,
				0,
				0,
				0,
				0,
				nativeSWPNoMove|nativeSWPNoSize|nativeSWPNoZOrder|nativeSWPNoActivate|nativeSWPFrame,
			)
			_ = nativeBoolCall(nativeDrawMenuBar, handle)
		}()
		if !setNativeMaximizeMenuEnabledHandle(handle, enabled) {
			return false
		}
		if !nativeBoolCall(
			nativeSetWindowPos,
			handle,
			0,
			0,
			0,
			0,
			0,
			nativeSWPNoMove|nativeSWPNoSize|nativeSWPNoZOrder|nativeSWPNoActivate|nativeSWPFrame,
		) {
			return false
		}
		if !nativeBoolCall(nativeDrawMenuBar, handle) {
			return false
		}
		committed = true
		return true
	})
}

func setNativeMaximizeMenuEnabledHandle(handle uintptr, enabled bool) bool {
	if handle == 0 {
		return false
	}
	menu, _, _ := nativeGetSystemMenu.Call(handle, 0)
	if menu == 0 {
		return false
	}
	flags := uintptr(nativeMFByCommand | nativeMFGrayed)
	if enabled {
		flags = nativeMFByCommand | nativeMFEnabled
	}
	result, _, _ := nativeEnableMenuItem.Call(menu, nativeSCMaximize, flags)
	return uint32(result) != ^uint32(0)
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
	return nativeBoolCall(nativeEnableNCDpi, handle)
}

func getDwmUint32Attribute(handle uintptr, attr uintptr) (uint32, bool) {
	if handle == 0 || nativeDwmGetWindowAttr.Find() != nil {
		return 0, false
	}
	var value uint32
	if !nativeHRESULTCall(
		nativeDwmGetWindowAttr,
		handle,
		attr,
		uintptr(unsafe.Pointer(&value)),
		unsafe.Sizeof(value),
	) {
		return 0, false
	}
	return value, true
}

func setDwmUint32Attribute(handle uintptr, attr uintptr, value uint32) bool {
	if handle == 0 || nativeDwmSetWindowAttr.Find() != nil {
		return false
	}
	return nativeHRESULTCall(
		nativeDwmSetWindowAttr,
		handle,
		attr,
		uintptr(unsafe.Pointer(&value)),
		unsafe.Sizeof(value),
	)
}

func extendNativeFrame(handle uintptr, margins nativeMargins) bool {
	if handle == 0 || nativeDwmExtendFrame.Find() != nil {
		return false
	}
	return nativeHRESULTCall(nativeDwmExtendFrame, handle, uintptr(unsafe.Pointer(&margins)))
}

func flushNativeDwm() bool {
	if nativeDwmFlush.Find() != nil {
		return false
	}
	return nativeHRESULTCall(nativeDwmFlush)
}

func installNativeWindowCloseHook(entry *windowEntry) bool {
	if entry == nil {
		return false
	}

	target := entry.nativeRawWindowTargetSnapshot()
	if !target.valid() {
		return false
	}
	entry.mu.Lock()
	alreadyHooked := entry.nativeCloseHooked &&
		entry.nativeCloseHookHandle == target.handle &&
		entry.nativeCloseHookGeneration == target.generation
	oldTarget := nativeWindowTarget{
		entry:      entry,
		handle:     entry.nativeCloseHookHandle,
		generation: entry.nativeCloseHookGeneration,
	}
	oldHookProc := entry.nativeCloseOldProc
	oldHooked := entry.nativeCloseHooked
	entry.mu.Unlock()
	if alreadyHooked {
		if binding, ok := nativeCloseHookBindingForHandle(target.handle); ok && binding.target == target {
			return true
		}
		if !restoreNativeWindowCloseHook(target, oldHookProc) {
			return false
		}
		oldHooked = false
	}

	if oldHooked && oldTarget != target {
		restoreNativeWindowCloseHook(oldTarget, oldHookProc)
	}

	return runNativeWindowCommand(target, func(handle uintptr) bool {
		nativeCloseHookMu.Lock()
		if current, ok := nativeCloseHookEntries[handle]; ok {
			if current.target != target {
				nativeCloseHookMu.Unlock()
				return false
			}
			entry.mu.Lock()
			if target.matchesEntryLocked() {
				entry.nativeCloseHookHandle = handle
				entry.nativeCloseHookGeneration = target.generation
				entry.nativeCloseHooked = true
				entry.nativeCloseOldProc = current.oldProc
			}
			entry.mu.Unlock()
			nativeCloseHookMu.Unlock()
			return true
		}

		oldProc, ok := nativeZeroValueCall(nativeSetWindowLongPtr, handle, nativeGWLPWndProc, nativeCloseHookCallback)
		if !ok {
			nativeCloseHookMu.Unlock()
			return false
		}
		binding := nativeCloseHookBinding{target: target, oldProc: oldProc}

		entry.mu.Lock()
		if !target.matchesEntryLocked() {
			entry.mu.Unlock()
			if _, restored := nativeZeroValueCall(nativeSetWindowLongPtr, handle, nativeGWLPWndProc, oldProc); !restored {
				nativeCloseHookEntries[handle] = binding
			}
			nativeCloseHookMu.Unlock()
			return false
		}
		entry.nativeCloseHookHandle = handle
		entry.nativeCloseHookGeneration = target.generation
		entry.nativeCloseHooked = true
		entry.nativeCloseOldProc = oldProc
		entry.mu.Unlock()

		nativeCloseHookEntries[handle] = binding
		nativeCloseHookMu.Unlock()
		return true
	})
}

func uninstallNativeWindowCloseHook(entry *windowEntry) {
	if entry == nil {
		return
	}

	entry.mu.Lock()
	target := nativeWindowTarget{
		entry:      entry,
		handle:     entry.nativeCloseHookHandle,
		generation: entry.nativeCloseHookGeneration,
	}
	oldProc := entry.nativeCloseOldProc
	hooked := entry.nativeCloseHooked
	entry.mu.Unlock()

	if !hooked || target.handle == 0 {
		clearNativeWindowCloseHookTarget(target)
		return
	}
	restoreNativeWindowCloseHook(target, oldProc)
}

func forgetNativeWindowCloseHook(entry *windowEntry) {
	if entry == nil {
		return
	}

	entry.mu.RLock()
	target := nativeWindowTarget{
		entry:      entry,
		handle:     entry.nativeCloseHookHandle,
		generation: entry.nativeCloseHookGeneration,
	}
	entry.mu.RUnlock()
	clearNativeWindowCloseHookTarget(target)

	if target.handle == 0 {
		return
	}
	nativeCloseHookMu.Lock()
	if current, ok := nativeCloseHookEntries[target.handle]; ok && current.target == target {
		isWindow, _, _ := nativeIsWindow.Call(target.handle)
		if isWindow == 0 {
			delete(nativeCloseHookEntries, target.handle)
		}
	}
	nativeCloseHookMu.Unlock()
}

func forgetNativeWindowCloseHookTarget(target nativeWindowTarget) {
	if target.handle == 0 {
		return
	}
	nativeCloseHookMu.Lock()
	if current, ok := nativeCloseHookEntries[target.handle]; ok && current.target == target {
		delete(nativeCloseHookEntries, target.handle)
	}
	nativeCloseHookMu.Unlock()
	clearNativeWindowCloseHookTarget(target)
}

func clearNativeWindowCloseHookTarget(target nativeWindowTarget) {
	if target.entry == nil || target.handle == 0 {
		return
	}
	target.entry.mu.Lock()
	if target.entry.nativeCloseHookHandle == target.handle &&
		target.entry.nativeCloseHookGeneration == target.generation {
		target.entry.nativeCloseHookHandle = 0
		target.entry.nativeCloseHookGeneration = 0
		target.entry.nativeCloseOldProc = 0
		target.entry.nativeCloseHooked = false
	}
	target.entry.mu.Unlock()
}

func restoreNativeWindowCloseHook(target nativeWindowTarget, fallbackOldProc uintptr) bool {
	if target.entry == nil || target.handle == 0 {
		return true
	}
	restored := false
	if !runOnNativeWindowThread(target.entry, func() {
		nativeCloseHookMu.Lock()
		binding, ok := nativeCloseHookEntries[target.handle]
		if ok && binding.target != target {
			nativeCloseHookMu.Unlock()
			return
		}
		if !ok {
			if fallbackOldProc == 0 {
				restored = true
				nativeCloseHookMu.Unlock()
				return
			}
			binding = nativeCloseHookBinding{target: target, oldProc: fallbackOldProc}
			nativeCloseHookEntries[target.handle] = binding
		}

		isWindow, _, _ := nativeIsWindow.Call(target.handle)
		if isWindow == 0 {
			delete(nativeCloseHookEntries, target.handle)
			restored = true
			nativeCloseHookMu.Unlock()
			return
		}
		_, restored = nativeZeroValueCall(nativeSetWindowLongPtr, target.handle, nativeGWLPWndProc, binding.oldProc)
		if restored {
			delete(nativeCloseHookEntries, target.handle)
		}
		nativeCloseHookMu.Unlock()
	}) {
		return false
	}
	if restored {
		clearNativeWindowCloseHookTarget(target)
	}
	return restored
}

func nativeWindowProc(hwnd, msg, wparam, lparam uintptr) uintptr {
	binding, hooked := nativeCloseHookBindingForHandle(hwnd)
	target := binding.target
	oldProc := binding.oldProc
	var entry *windowEntry
	if hooked && target.valid() {
		entry = target.entry
		entry.nativeCallbackDepth.Add(1)
		defer entry.nativeCallbackDepth.Add(-1)
	}

	if msg == nativeWMNCDestroy {
		if target.entry != nil {
			target.entry.invalidateNativeWindowTarget(target)
		}
		result := nativeCallDefaultWindowProc(oldProc, hwnd, msg, wparam, lparam)
		if hooked {
			forgetNativeWindowCloseHookTarget(target)
		}
		return result
	}

	if msg == nativeWMClose && entry != nil {
		if !entry.handleCloseRequested() {
			return 0
		}
	}

	if msg == nativeWMNCHitTest {
		result := nativeCallDefaultWindowProc(oldProc, hwnd, msg, wparam, lparam)
		if nativeMaximizeButtonHitTest(entry, hwnd, lparam) {
			return nativeHTMaxButton
		}
		if result == nativeHTClient {
			if nativeDragActionHitTest(entry, hwnd, lparam) {
				return nativeHTCaption
			}
		}
		return result
	}

	if msg == nativeWMNCPointerUpdate || msg == nativeWMNCMouseMove {
		inside := nativeMaximizeButtonHitTest(entry, hwnd, lparam)
		hit := wparam
		if inside && msg == nativeWMNCMouseMove {
			hit = nativeHTMaxButton
		}
		nativeForwardNonClientInputToGio(oldProc, hwnd, msg, hit, lparam)
		result := nativeCallDefaultWindowProc(oldProc, hwnd, msg, hit, lparam)
		if hit == nativeHTMaxButton || inside {
			nativeInvalidateHiddenFrame(entry, hwnd)
		}
		return result
	}

	if msg == nativeWMMouseMove {
		result := nativeCallDefaultWindowProc(oldProc, hwnd, msg, wparam, lparam)
		if nativeMaximizeButtonClientHitTest(entry, hwnd, lparam) {
			nativeForwardMaximizeClientHoverToNonClient(oldProc, hwnd, lparam)
			nativeInvalidateHiddenFrame(entry, hwnd)
		}
		return result
	}

	if msg == nativeWMNCPointerDown || msg == nativeWMNCPointerUp {
		inside := nativeMaximizeButtonHitTest(entry, hwnd, lparam)
		active := nativeMaximizePointerDown(entry)
		if inside || active {
			if msg == nativeWMNCPointerDown {
				nativeSetMaximizePointerDown(entry, true)
				nativeSetMaximizeMouseDown(entry, false)
			}
			nativeForwardNonClientInputToGio(oldProc, hwnd, msg, wparam, lparam)
			if msg == nativeWMNCPointerUp {
				nativeSetMaximizePointerDown(entry, false)
			}
			nativeInvalidateHiddenFrame(entry, hwnd)
			return 0
		}
	}

	if wparam == nativeHTMaxButton {
		switch msg {
		case nativeWMNCLButtonDown:
			nativeForwardNonClientInputToGio(oldProc, hwnd, msg, wparam, lparam)
			if !nativeMaximizePointerDown(entry) {
				nativeSetMaximizeMouseDown(entry, true)
			}
			return 0
		case nativeWMNCLButtonUp:
			nativeForwardNonClientInputToGio(oldProc, hwnd, msg, wparam, lparam)
			activate := nativeMaximizeMouseDown(entry) && !nativeMaximizePointerDown(entry)
			nativeSetMaximizeMouseDown(entry, false)
			if activate && entry != nil {
				entry.activateNativeMaximizeButton()
			}
			return 0
		case nativeWMNCLDblClk:
			return 0
		}
	}

	if msg == nativeWMNCMouseLeave {
		nativeForwardNonClientInputToGio(oldProc, hwnd, msg, wparam, lparam)
		result := nativeCallDefaultWindowProc(oldProc, hwnd, msg, wparam, lparam)
		nativeInvalidateHiddenFrame(entry, hwnd)
		return result
	}

	if msg == nativeWMPointerUp || msg == nativeWMPointerCaptureChanged || msg == nativeWMCancelMode {
		nativeSetMaximizePointerDown(entry, false)
		nativeSetMaximizeMouseDown(entry, false)
	}

	if entry != nil && nativeHiddenFrame(entry) {
		if nativeSuppressHiddenFrameNonClient(msg) {
			return nativeHiddenFrameNonClientResult(msg)
		}
	}

	return nativeCallDefaultWindowProc(oldProc, hwnd, msg, wparam, lparam)
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

func nativeMaximizeButtonHitTest(entry *windowEntry, hwnd, lparam uintptr) bool {
	if entry == nil || hwnd == 0 {
		return false
	}
	pt := nativePointFromLParam(lparam)
	if ok, _, _ := nativeScreenToClient.Call(hwnd, uintptr(unsafe.Pointer(&pt))); ok == 0 {
		return false
	}
	return entry.nativeMaximizeButtonAt(int(pt.X), int(pt.Y))
}

func nativeMaximizeButtonClientHitTest(entry *windowEntry, hwnd, lparam uintptr) bool {
	if entry == nil || hwnd == 0 {
		return false
	}
	pt := nativePointFromLParam(lparam)
	return entry.nativeMaximizeButtonAt(int(pt.X), int(pt.Y))
}

func nativeForwardMaximizeClientHoverToNonClient(oldProc, hwnd, lparam uintptr) bool {
	if oldProc == 0 || hwnd == 0 {
		return false
	}
	pt := nativePointFromLParam(lparam)
	if ok, _, _ := nativeClientToScreen.Call(hwnd, uintptr(unsafe.Pointer(&pt))); ok == 0 {
		return false
	}
	nativeCallDefaultWindowProc(oldProc, hwnd, nativeWMNCMouseMove, nativeHTMaxButton, nativeLParamFromPoint(pt))
	return true
}

func nativeForwardNonClientInputToGio(oldProc, hwnd, msg, wparam, lparam uintptr) bool {
	if oldProc == 0 || hwnd == 0 {
		return false
	}
	clientMsg, clientWParam, clientCoords, ok := nativeClientInputMessage(msg, wparam)
	if !ok {
		return false
	}
	clientLParam := lparam
	if clientCoords {
		pt := nativePointFromLParam(lparam)
		if ok, _, _ := nativeScreenToClient.Call(hwnd, uintptr(unsafe.Pointer(&pt))); ok == 0 {
			return false
		}
		clientLParam = nativeLParamFromPoint(pt)
	}
	nativeCallDefaultWindowProc(oldProc, hwnd, clientMsg, clientWParam, clientLParam)
	return true
}

func nativeClientInputMessage(msg, wparam uintptr) (clientMsg, clientWParam uintptr, clientCoords bool, ok bool) {
	switch msg {
	case nativeWMNCPointerUpdate:
		return nativeWMPointerUpdate, wparam, false, true
	case nativeWMNCPointerDown:
		return nativeWMPointerDown, wparam, false, true
	case nativeWMNCPointerUp:
		return nativeWMPointerUp, wparam, false, true
	case nativeWMNCMouseMove:
		return nativeWMMouseMove, 0, true, true
	case nativeWMNCLButtonDown, nativeWMNCLDblClk:
		return nativeWMLButtonDown, nativeMKLButton, true, true
	case nativeWMNCLButtonUp:
		return nativeWMLButtonUp, 0, true, true
	case nativeWMNCMouseLeave:
		return nativeWMMouseLeave, 0, true, true
	default:
		return 0, 0, false, false
	}
}

func nativeMaximizePointerDown(entry *windowEntry) bool {
	if entry == nil {
		return false
	}
	entry.mu.RLock()
	down := entry.nativeMaximizePointerDown
	entry.mu.RUnlock()
	return down
}

func nativeSetMaximizePointerDown(entry *windowEntry, down bool) {
	if entry == nil {
		return
	}
	entry.mu.Lock()
	entry.nativeMaximizePointerDown = down
	entry.mu.Unlock()
}

func nativeMaximizeMouseDown(entry *windowEntry) bool {
	if entry == nil {
		return false
	}
	entry.mu.RLock()
	down := entry.nativeMaximizeMouseDown
	entry.mu.RUnlock()
	return down
}

func nativeSetMaximizeMouseDown(entry *windowEntry, down bool) {
	if entry == nil {
		return
	}
	entry.mu.Lock()
	entry.nativeMaximizeMouseDown = down
	entry.mu.Unlock()
}

func nativeLParamFromPoint(pt nativePoint) uintptr {
	return uintptr(uint16(pt.X)) | uintptr(uint16(pt.Y))<<16
}

func nativeInvalidateHiddenFrame(entry *windowEntry, hwnd uintptr) {
	if hwnd == 0 || entry == nil || !nativeHiddenFrame(entry) {
		return
	}
	nativeRedrawWindow.Call(hwnd, 0, 0, nativeRDWInvalidate|nativeRDWFrame|nativeRDWAllChildren)
	if entry.win != nil {
		entry.win.Invalidate()
	}
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

func nativeCloseHookBindingForHandle(hwnd uintptr) (nativeCloseHookBinding, bool) {
	nativeCloseHookMu.RLock()
	binding, ok := nativeCloseHookEntries[hwnd]
	nativeCloseHookMu.RUnlock()
	if !ok || binding.target.handle != hwnd {
		return nativeCloseHookBinding{}, false
	}
	return binding, true
}

func nativeCloseHookTarget(hwnd uintptr) (nativeWindowTarget, bool) {
	binding, ok := nativeCloseHookBindingForHandle(hwnd)
	if !ok || !binding.target.valid() {
		return nativeWindowTarget{}, false
	}
	return binding.target, true
}
