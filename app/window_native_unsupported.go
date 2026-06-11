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
