//go:build !windows

package app

import gioApp "gioui.org/app"

func (entry *windowEntry) updateNativeHandle(event gioApp.ViewEvent) {
	entry.mu.Lock()
	entry.nativeHandle = 0
	entry.mu.Unlock()
}
