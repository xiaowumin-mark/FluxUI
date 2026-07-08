package internal

import "image"

// NativeWindowAction identifies a native non-client window action region.
type NativeWindowAction int

const (
	// NativeWindowActionMaximizeButton marks a Windows maximize button hit-test
	// area. On Windows this enables the OS Snap Layouts flyout for custom chrome.
	NativeWindowActionMaximizeButton NativeWindowAction = iota + 1
)

// NativeWindowActionRegion describes a native window action hit-test rectangle
// in window coordinates.
type NativeWindowActionRegion struct {
	Action NativeWindowAction
	Rect   image.Rectangle
}

// RegisterNativeWindowActionRegion records a native window action region for
// the current frame.
func (r *Runtime) RegisterNativeWindowActionRegion(action NativeWindowAction, rect image.Rectangle) {
	if r == nil || action == 0 || rect.Empty() {
		return
	}
	r.nativeWindowActions = append(r.nativeWindowActions, NativeWindowActionRegion{
		Action: action,
		Rect:   rect,
	})
}

// NativeWindowActionRegions returns the native window action regions registered
// in the current frame. The returned slice is valid until the next BeginFrame.
func (r *Runtime) NativeWindowActionRegions() []NativeWindowActionRegion {
	if r == nil || len(r.nativeWindowActions) == 0 {
		return nil
	}
	return r.nativeWindowActions
}

// NativeWindowActionRegionsActive reports whether the current frame registered
// any native window action regions.
func (r *Runtime) NativeWindowActionRegionsActive() bool {
	return r != nil && len(r.nativeWindowActions) > 0
}

// RegisterNativeWindowActionRouter marks the current frame as containing a
// Gio system action input region that native hit testing should query.
func (r *Runtime) RegisterNativeWindowActionRouter() {
	if r == nil {
		return
	}
	r.nativeWindowActionRouter = true
}

// NativeWindowActionRouterActive reports whether native hit testing should keep
// a Gio action router for the current frame.
func (r *Runtime) NativeWindowActionRouterActive() bool {
	return r != nil && r.nativeWindowActionRouter
}
