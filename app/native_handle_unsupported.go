//go:build !windows

package app

import gioApp "gioui.org/app"

func (entry *windowEntry) updateNativeHandle(event gioApp.ViewEvent) {
	entry.setNativeWindowHandle(0)
}
