package internal

import (
	"image"
	"image/color"
	"strconv"
	"time"

	theme "github.com/xiaowumin-mark/FluxUI/theme"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
)

// Context 是每一帧传递给组件树的执行上下文。
type Context struct {
	Gtx        gioLayout.Context
	runtime    *Runtime
	path       string
	hookIndex  int
	foreground color.NRGBA
	font       theme.FontSpec
	hasFont    bool
}

// NewContext 创建 frame 级上下文。
func NewContext(gtx gioLayout.Context, runtime *Runtime) *Context {
	return &Context{
		Gtx:        gtx,
		runtime:    runtime,
		path:       "root",
		foreground: runtime.Theme().TextColor,
	}
}

// Runtime 返回运行时实例。
func (c *Context) Runtime() *Runtime {
	return c.runtime
}

// Theme 返回当前主题。
func (c *Context) Theme() *theme.Theme {
	return c.runtime.Theme()
}

// MaterialTheme 返回内部 Gio 主题。
func (c *Context) MaterialTheme() *material.Theme {
	return c.runtime.MaterialTheme()
}

// Foreground 返回当前默认前景色。
func (c *Context) Foreground() color.NRGBA {
	return c.foreground
}

// Font 返回当前默认字体。
func (c *Context) Font() theme.FontSpec {
	if c.hasFont {
		return c.font.Normalize()
	}
	th := c.Theme()
	if th == nil {
		return theme.DefaultFontSpec()
	}
	return th.DefaultFont.Normalize()
}

// Now 返回当前 frame 时间。
func (c *Context) Now() time.Time {
	return c.Gtx.Now
}

// MinConstraints 返回当前最小约束。
func (c *Context) MinConstraints() image.Point {
	return c.Gtx.Constraints.Min
}

// MaxConstraints 返回当前最大约束。
func (c *Context) MaxConstraints() image.Point {
	return c.Gtx.Constraints.Max
}

// RequestRedraw 请求下一帧刷新。
func (c *Context) RequestRedraw() {
	c.Gtx.Execute(op.InvalidateCmd{})
	c.runtime.RequestRedraw()
}

// WindowID 返回当前窗口 ID。
func (c *Context) WindowID() WindowID {
	ctrl := c.runtime.WindowController()
	if ctrl == nil {
		return 0
	}
	return ctrl.WindowID()
}

// WindowClose 请求关闭当前窗口。
func (c *Context) WindowClose() bool {
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Close()
}

// WindowMinimize 请求最小化当前窗口。
func (c *Context) WindowMinimize() bool {
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Minimize()
}

// WindowMaximize 请求最大化当前窗口。
func (c *Context) WindowMaximize() bool {
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Maximize()
}

// WindowRestore 请求还原当前窗口。
func (c *Context) WindowRestore() bool {
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Restore()
}

// WindowFullscreen 请求全屏当前窗口。
func (c *Context) WindowFullscreen() bool {
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Fullscreen()
}

// WindowRaise 请求将当前窗口置顶。
func (c *Context) WindowRaise() bool {
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Raise()
}

// WindowCenter 请求将当前窗口居中。
func (c *Context) WindowCenter() bool {
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Center()
}

// WindowSetTitle 更新当前窗口标题。
func (c *Context) WindowSetTitle(title string) bool {
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.SetTitle(title)
}

// WindowSetSize 更新当前窗口尺寸（单位 dp）。
func (c *Context) WindowSetSize(width, height int) bool {
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.SetSize(width, height)
}

// WindowInvalidate 请求当前窗口立即重绘。
func (c *Context) WindowInvalidate() bool {
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Invalidate()
}

// WindowIsAlive 返回当前窗口是否仍然存活。
func (c *Context) WindowIsAlive() bool {
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.IsAlive()
}

// NextKey 生成当前作用域下稳定的 hook key。
func (c *Context) NextKey(namespace string) string {
	key := c.path + "/" + namespace + ":" + strconv.Itoa(c.hookIndex)
	c.hookIndex++
	if c.runtime != nil {
		c.runtime.RecordHookCount(c.path, c.hookIndex)
	}
	return key
}

// Persistent 读取或创建稳定对象。
func (c *Context) Persistent(key string, factory func() any) any {
	return c.runtime.remember(key, factory)
}

// Memo 使用稳定 hook key 读取或创建对象。
func (c *Context) Memo(namespace string, factory func() any) any {
	return c.Persistent(c.NextKey(namespace), factory)
}

// Child 为子组件创建独立作用域。
func (c *Context) Child(index int) *Context {
	return c.childWithGtx(c.Gtx, strconv.Itoa(index))
}

// Scope 创建命名作用域。
func (c *Context) Scope(name string) *Context {
	return c.childWithGtx(c.Gtx, name)
}

// WithForeground 覆盖当前默认前景色。
func (c *Context) WithForeground(col color.NRGBA) *Context {
	next := c.sameScope(c.Gtx)
	next.foreground = col
	return next
}

// WithFont 覆盖当前作用域默认字体。
func (c *Context) WithFont(spec theme.FontSpec) *Context {
	next := c.sameScope(c.Gtx)
	next.font = spec.Normalize()
	next.hasFont = true
	return next
}

func (c *Context) sameScope(gtx gioLayout.Context) *Context {
	next := *c
	next.Gtx = gtx
	return &next
}

func (c *Context) childWithGtx(gtx gioLayout.Context, segment string) *Context {
	next := c.sameScope(gtx)
	next.path = c.path + "/" + segment
	next.hookIndex = 0
	return next
}
