//go:build !windows

package app

func setNativeWindowAlwaysOnTop(uintptr, bool) bool {
	return false
}

func setNativeWindowVisible(uintptr, bool) bool {
	return false
}

func setNativeWindowMaximizeEnabled(uintptr, bool) bool {
	return false
}
