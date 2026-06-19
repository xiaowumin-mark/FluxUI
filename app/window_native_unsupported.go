//go:build !windows

package app

func setNativeWindowAlwaysOnTop(uintptr, bool) bool {
	return false
}

func setNativeWindowVisible(uintptr, bool) bool {
	return false
}

func requestNativeWindowFocus(uintptr) bool {
	return false
}

func startNativeWindowDragMove(uintptr) bool {
	return false
}

func setNativeWindowFrameStyle(uintptr, WindowsFrameStyle, bool, bool) bool {
	return false
}

func probeNativeWindowsChrome() WindowsChromeAvailability {
	return WindowsChromeAvailability{}
}

func setNativeWindowPosition(uintptr, int, int) bool {
	return false
}

func setNativeWindowResizable(uintptr, bool) bool {
	return false
}

func setNativeWindowMaximizeEnabled(uintptr, bool) bool {
	return false
}

func installNativeWindowCloseHook(*windowEntry) bool {
	return false
}

func uninstallNativeWindowCloseHook(*windowEntry) {
}

func forgetNativeWindowCloseHook(*windowEntry) {
}
