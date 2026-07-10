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

	ready := entry.withNativeWindowOperation(func() bool {
		entry.stageNativeWindowHandle(handle)
		return entry.syncNativeCloseHookLocked()
	})
	if !ready || handle == 0 {
		return
	}
	entry.syncNativeMaximizeAvailability()
	entry.syncNativeResizable()
	entry.syncNativeChrome()
}
