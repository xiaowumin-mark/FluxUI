//go:build !windows

package app

func setNativeWindowAlwaysOnTop(nativeWindowTarget, bool) bool {
	return false
}

func setNativeWindowVisible(nativeWindowTarget, bool) bool {
	return false
}

func requestNativeWindowFocus(nativeWindowTarget) bool {
	return false
}

func startNativeWindowDragMove(nativeWindowTarget) bool {
	return false
}

func setNativeWindowFrameStyle(nativeWindowTarget, WindowsFrameStyle, bool, bool) bool {
	return false
}

func probeNativeWindowsChrome() WindowsChromeAvailability {
	return WindowsChromeAvailability{}
}

func setNativeWindowPosition(nativeWindowTarget, int, int) bool {
	return false
}

func setNativeWindowResizable(nativeWindowTarget, bool) bool {
	return false
}

func setNativeWindowMaximizeEnabled(nativeWindowTarget, bool) bool {
	return false
}

func installNativeWindowCloseHook(*windowEntry) bool {
	return false
}

func uninstallNativeWindowCloseHook(*windowEntry) {
}

func forgetNativeWindowCloseHook(*windowEntry) {
}
