package internal

// WindowID 是内部窗口唯一标识。
type WindowID uint64

// WindowController 定义当前窗口可执行的控制动作。
type WindowController interface {
	WindowID() WindowID
	Close() bool
	Minimize() bool
	Maximize() bool
	Restore() bool
	Fullscreen() bool
	Raise() bool
	Center() bool
	SetTitle(title string) bool
	SetSize(width, height int) bool
	Invalidate() bool
	IsAlive() bool
}
