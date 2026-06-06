//go:build windows

package app

import (
	"runtime"
	"syscall"
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

	nativeWSMaximizeBox = 0x00010000
	nativeSCMaximize    = 0xF030
	nativeMFByCommand   = 0x0000
	nativeMFEnabled     = 0x0000
	nativeMFGrayed      = 0x0001
)

var (
	nativeHWNDTopmost   = ^uintptr(0)
	nativeHWNDNotopmost = ^uintptr(0) - 1
	nativeGWLStyle      = ^uintptr(15)

	nativeUser32          = syscall.NewLazyDLL("user32.dll")
	nativeDrawMenuBar     = nativeUser32.NewProc("DrawMenuBar")
	nativeEnableMenuItem  = nativeUser32.NewProc("EnableMenuItem")
	nativeGetSystemMenu   = nativeUser32.NewProc("GetSystemMenu")
	nativeGetWindowLong   = nativeUser32.NewProc("GetWindowLongW")
	nativeSetWindowPos    = nativeUser32.NewProc("SetWindowPos")
	nativeSetWindowLong   = nativeUser32.NewProc("SetWindowLongW")
	nativeShowWindowAsync = nativeUser32.NewProc("ShowWindowAsync")
)

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

		if menu, _, _ := nativeGetSystemMenu.Call(handle, 0); menu != 0 {
			flags := uintptr(nativeMFByCommand | nativeMFGrayed)
			if enabled {
				flags = nativeMFByCommand | nativeMFEnabled
			}
			nativeEnableMenuItem.Call(menu, nativeSCMaximize, flags)
		}

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
