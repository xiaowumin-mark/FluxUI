//go:build windows

package app

import (
	"runtime"
	"sync"
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
	nativeWSSizeBox     = 0x00040000
	nativeSCMaximize    = 0xF030
	nativeMFByCommand   = 0x0000
	nativeMFEnabled     = 0x0000
	nativeMFGrayed      = 0x0001

	nativeWMClose     = 0x0010
	nativeWMNCDestroy = 0x0082
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
	nativeShowWindowAsync  = nativeUser32.NewProc("ShowWindowAsync")

	nativeCloseHookCallback = syscall.NewCallback(nativeWindowProc)
	nativeCloseHookMu       sync.RWMutex
	nativeCloseHookEntries  = make(map[uintptr]*windowEntry)
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

	result := uintptr(0)
	if oldProc != 0 {
		result, _, _ = nativeCallWindowProc.Call(oldProc, hwnd, msg, wparam, lparam)
	} else {
		result, _, _ = nativeDefWindowProc.Call(hwnd, msg, wparam, lparam)
	}

	if msg == nativeWMNCDestroy && entry != nil {
		forgetNativeWindowCloseHook(entry)
	}
	return result
}

func nativeCloseHookEntry(hwnd uintptr) *windowEntry {
	nativeCloseHookMu.RLock()
	entry := nativeCloseHookEntries[hwnd]
	nativeCloseHookMu.RUnlock()
	return entry
}
