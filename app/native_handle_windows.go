//go:build windows

package app

import gioApp "gioui.org/app"

func (entry *windowEntry) updateNativeHandle(event gioApp.ViewEvent) {
	handle := uintptr(0)
	if event.Valid() {
		if win32, ok := event.(gioApp.Win32ViewEvent); ok {
			handle = uintptr(win32.HWND)
		}
	}

	entry.mu.Lock()
	entry.nativeHandle = handle
	entry.mu.Unlock()
	entry.syncNativeMaximizeAvailability()
	entry.syncNativeResizable()
	entry.syncNativeChrome()
	entry.syncNativeCloseHook()
}
