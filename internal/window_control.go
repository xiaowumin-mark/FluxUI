package internal

// WindowID 是内部窗口唯一标识。
type WindowID uint64

// WindowHiddenMemoryPolicy 定义窗口隐藏后的渲染内存策略。
type WindowHiddenMemoryPolicy int

// WindowController 定义当前窗口可执行的控制动作。
type WindowController interface {
	WindowID() WindowID
	Close() bool
	Show() bool
	Hide() bool
	SetHiddenMemoryPolicy(policy WindowHiddenMemoryPolicy) bool
	Minimize() bool
	Maximize() bool
	Restore() bool
	Fullscreen() bool
	Raise() bool
	SetAlwaysOnTop(always bool) bool
	Center() bool
	SetTitle(title string) bool
	SetSize(width, height int) bool
	SetMinSize(width, height int) bool
	SetMaxSize(width, height int) bool
	Invalidate() bool
	IsAlive() bool
}
