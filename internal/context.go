package internal

import (
	"image"
	"image/color"
	"reflect"
	"strconv"
	"time"

	theme "github.com/xiaowumin-mark/FluxUI/theme"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
)

// Context 是每一帧传递给组件树的执行上下文。
type Context struct {
	Gtx           gioLayout.Context
	runtime       *Runtime
	path          string
	hookIndex     int
	foreground    color.NRGBA
	font          theme.FontSpec
	hasFont       bool
	textStyle     theme.TextStyle
	hasTextStyle  bool
	instance      *ComponentInstance
	providers     map[ProviderKey]any
	themeOverride *theme.Theme
}

// NewContext 创建 frame 级上下文。
func NewContext(gtx gioLayout.Context, runtime *Runtime) *Context {
	foreground := theme.Default().TextColor
	if runtime != nil && runtime.Theme() != nil {
		foreground = runtime.Theme().TextColor
	}
	return &Context{
		Gtx:        gtx,
		runtime:    runtime,
		path:       "root",
		foreground: foreground,
	}
}

// Runtime 返回运行时实例。
func (c *Context) Runtime() *Runtime {
	if c == nil {
		return nil
	}
	return c.runtime
}

// Theme 返回当前作用域的主题。若当前上下文有 theme 覆盖（通过 ThemeProvider），
// 则返回覆盖的主题；否则返回运行时全局主题。
func (c *Context) Theme() *theme.Theme {
	if c == nil {
		return theme.Default()
	}
	if c.themeOverride != nil {
		return c.themeOverride
	}
	if c.runtime == nil {
		return theme.Default()
	}
	return c.runtime.Theme()
}

// SetThemeOverride sets a scoped theme on this context (used by ThemeProvider).
func (c *Context) SetThemeOverride(th *theme.Theme) {
	if c == nil {
		return
	}
	c.themeOverride = th
}

// MaterialTheme 返回内部 Gio 主题。
func (c *Context) MaterialTheme() *material.Theme {
	if c == nil || c.runtime == nil {
		return material.NewTheme()
	}
	return c.runtime.MaterialTheme()
}

// Foreground 返回当前默认前景色。
func (c *Context) Foreground() color.NRGBA {
	if c == nil {
		return theme.Default().TextColor
	}
	return c.foreground
}

// Font 返回当前默认字体。
func (c *Context) Font() theme.FontSpec {
	if c == nil {
		return theme.DefaultFontSpec()
	}
	if c.hasFont {
		return c.font.Normalize()
	}
	th := c.Theme()
	if th == nil {
		return theme.DefaultFontSpec()
	}
	return th.DefaultFont.Normalize()
}

func (c *Context) TextStyle() (theme.TextStyle, bool) {
	if c == nil || !c.hasTextStyle {
		return theme.TextStyle{}, false
	}
	return c.textStyle, true
}

// Now 返回当前 frame 时间。
func (c *Context) Now() time.Time {
	if c == nil {
		return time.Time{}
	}
	return c.Gtx.Now
}

// MinConstraints 返回当前最小约束。
func (c *Context) MinConstraints() image.Point {
	if c == nil {
		return image.Point{}
	}
	return c.Gtx.Constraints.Min
}

// MaxConstraints 返回当前最大约束。
func (c *Context) MaxConstraints() image.Point {
	if c == nil {
		return image.Point{}
	}
	return c.Gtx.Constraints.Max
}

// RequestRedraw 请求窗口重绘。
// 该方法只依赖 Runtime，可安全用于事件回调或 goroutine；不要把 Context 长期保存。
func (c *Context) RequestRedraw() {
	if c == nil || c.runtime == nil {
		return
	}
	c.runtime.RequestRedraw()
}

// RequestFrameRedraw 请求 frame 驱动的下一帧刷新。
// 仅在当前 Layout/Render frame 内部使用；跨 goroutine 请使用 RequestRedraw。
func (c *Context) RequestFrameRedraw() {
	if c == nil {
		return
	}
	c.Gtx.Execute(op.InvalidateCmd{})
	c.RequestRedraw()
}

// WindowID 返回当前窗口 ID。
func (c *Context) WindowID() WindowID {
	if c == nil || c.runtime == nil {
		return 0
	}
	ctrl := c.runtime.WindowController()
	if ctrl == nil {
		return 0
	}
	return ctrl.WindowID()
}

// WindowClose 请求关闭当前窗口。
func (c *Context) WindowClose() bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Close()
}

// WindowShow 请求显示当前窗口。
func (c *Context) WindowShow() bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Show()
}

// WindowHide 请求隐藏当前窗口。
func (c *Context) WindowHide() bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Hide()
}

// WindowSetHiddenMemoryPolicy 设置当前窗口隐藏后的渲染内存策略。
func (c *Context) WindowSetHiddenMemoryPolicy(policy WindowHiddenMemoryPolicy) bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.SetHiddenMemoryPolicy(policy)
}

// WindowMinimize 请求最小化当前窗口。
func (c *Context) WindowMinimize() bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Minimize()
}

// WindowMaximize 请求最大化当前窗口。
func (c *Context) WindowMaximize() bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Maximize()
}

// WindowRestore 请求还原当前窗口。
func (c *Context) WindowRestore() bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Restore()
}

// WindowFullscreen 请求全屏当前窗口。
func (c *Context) WindowFullscreen() bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Fullscreen()
}

// WindowRaise 请求将当前窗口一次性前置。
func (c *Context) WindowRaise() bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Raise()
}

// WindowSetAlwaysOnTop 设置当前窗口是否持续置顶。
func (c *Context) WindowSetAlwaysOnTop(always bool) bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.SetAlwaysOnTop(always)
}

// WindowCenter 请求将当前窗口居中。
func (c *Context) WindowCenter() bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Center()
}

// WindowRequestFocus 请求当前窗口获得焦点。
func (c *Context) WindowRequestFocus() bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.RequestFocus()
}

// WindowSetTitle 更新当前窗口标题。
func (c *Context) WindowSetTitle(title string) bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.SetTitle(title)
}

// WindowSetPosition 更新当前窗口位置（单位 dp）。
func (c *Context) WindowSetPosition(x, y int) bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.SetPosition(x, y)
}

// WindowSetSize 更新当前窗口尺寸（单位 dp）。
func (c *Context) WindowSetSize(width, height int) bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.SetSize(width, height)
}

// WindowSetResizable 设置当前窗口是否可被系统边框调整大小。
func (c *Context) WindowSetResizable(resizable bool) bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.SetResizable(resizable)
}

// WindowSetDecorated 设置当前窗口是否使用系统装饰边框。
func (c *Context) WindowSetDecorated(decorated bool) bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.SetDecorated(decorated)
}

// WindowSetMinSize 更新当前窗口最小尺寸（单位 dp）。
func (c *Context) WindowSetMinSize(width, height int) bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.SetMinSize(width, height)
}

// WindowSetMaxSize 更新当前窗口最大尺寸（单位 dp）。
func (c *Context) WindowSetMaxSize(width, height int) bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.SetMaxSize(width, height)
}

// WindowInvalidate 请求当前窗口立即重绘。
func (c *Context) WindowInvalidate() bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.Invalidate()
}

// WindowIsAlive 返回当前窗口是否仍然存活。
func (c *Context) WindowIsAlive() bool {
	if c == nil || c.runtime == nil {
		return false
	}
	ctrl := c.runtime.WindowController()
	return ctrl != nil && ctrl.IsAlive()
}

// NextKey 生成当前作用域下稳定的 hook key。
func (c *Context) NextKey(namespace string) string {
	if c == nil {
		return namespace + ":0"
	}
	key := c.path + "/" + namespace + ":" + strconv.Itoa(c.hookIndex)
	c.hookIndex++
	if c.runtime != nil {
		c.runtime.RecordHookCount(c.path, c.hookIndex)
	}
	return key
}

// Persistent 读取或创建稳定对象。
func (c *Context) Persistent(key string, factory func() any) any {
	if c == nil || c.runtime == nil {
		if factory == nil {
			return nil
		}
		return factory()
	}
	return c.runtime.remember(key, factory)
}

// Memo 使用稳定 hook key 读取或创建对象。
func (c *Context) Memo(namespace string, factory func() any) any {
	return c.Persistent(c.NextKey(namespace), factory)
}

// WithComponentInstance binds this context to a component hook instance.
func (c *Context) WithComponentInstance(instance *ComponentInstance) *Context {
	if c == nil {
		return nil
	}
	next := c.sameScope(c.Gtx)
	next.instance = instance
	return next
}

// ComponentInstance returns the currently bound experimental component instance.
func (c *Context) ComponentInstance() *ComponentInstance {
	if c == nil {
		return nil
	}
	return c.instance
}

// NextHookSlot returns the next component-owned hook slot, if this context is
// currently rendering inside an experimental component instance.
func (c *Context) NextHookSlot(kind HookKind) *HookSlot {
	if c == nil || c.instance == nil {
		return nil
	}
	return c.instance.NextHook(kind)
}

// ProviderKey identifies a typed context provider slot.
type ProviderKey struct {
	typ  reflect.Type
	id   uint64
	name string
}

// ProviderKeyFor returns a provider key for T. id 0 keeps the legacy type-wide
// provider behavior.
func ProviderKeyFor[T any](id uint64, name string) ProviderKey {
	return ProviderKey{typ: contextKeyType[T](), id: id, name: name}
}

// WithProviderValue returns a context with one type-wide provider value
// overridden. Prefer WithProviderKeyValue when a public ContextKey is
// available.
func WithProviderValue[T any](c *Context, value T) *Context {
	return WithProviderKeyValue(c, ProviderKeyFor[T](0, ""), value)
}

// WithProviderKeyValue returns a context with one provider key overridden.
func WithProviderKeyValue[T any](c *Context, key ProviderKey, value T) *Context {
	if c == nil {
		return nil
	}
	next := c.sameScope(c.Gtx)
	next.providers = cloneProviders(c.providers)
	next.providers[key] = value
	return next
}

// ProviderValue reads a type-wide provider value from context, or fallback
// when absent.
func ProviderValue[T any](c *Context, fallback T) T {
	return ProviderKeyValue(c, ProviderKeyFor[T](0, ""), fallback)
}

// ProviderKeyValue reads a provider key value from context, or fallback when
// absent.
func ProviderKeyValue[T any](c *Context, key ProviderKey, fallback T) T {
	if c == nil || c.providers == nil {
		return fallback
	}
	value, ok := c.providers[key].(T)
	if !ok {
		return fallback
	}
	return value
}

// Child 为子组件创建独立作用域。
func (c *Context) Child(index int) *Context {
	if c == nil {
		return nil
	}
	return c.childWithGtx(c.Gtx, strconv.Itoa(index))
}

// Scope 创建命名作用域。
func (c *Context) Scope(name string) *Context {
	if c == nil {
		return nil
	}
	return c.childWithGtx(c.Gtx, name)
}

// WithForeground 覆盖当前默认前景色。
func (c *Context) WithForeground(col color.NRGBA) *Context {
	if c == nil {
		return nil
	}
	next := c.sameScope(c.Gtx)
	next.foreground = col
	return next
}

// WithFont 覆盖当前作用域默认字体。
func (c *Context) WithFont(spec theme.FontSpec) *Context {
	if c == nil {
		return nil
	}
	next := c.sameScope(c.Gtx)
	next.font = spec.Normalize()
	next.hasFont = true
	return next
}

// WithTheme 覆盖当前作用域主题（返回新上下文，原上下文不变）。
func (c *Context) WithTheme(th *theme.Theme) *Context {
	if c == nil {
		return nil
	}
	next := c.sameScope(c.Gtx)
	next.themeOverride = th
	return next
}

func (c *Context) WithTextStyle(style theme.TextStyle) *Context {
	if c == nil {
		return nil
	}
	next := c.sameScope(c.Gtx)
	next.textStyle = style
	next.hasTextStyle = style.Size > 0 || style.LineHeight > 0
	return next
}

func (c *Context) sameScope(gtx gioLayout.Context) *Context {
	if c == nil {
		return nil
	}
	next := *c
	next.Gtx = gtx
	return &next
}

func cloneProviders(providers map[ProviderKey]any) map[ProviderKey]any {
	next := make(map[ProviderKey]any, len(providers)+1)
	for key, value := range providers {
		next[key] = value
	}
	return next
}

func contextKeyType[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func (c *Context) childWithGtx(gtx gioLayout.Context, segment string) *Context {
	if c == nil {
		return nil
	}
	next := c.sameScope(gtx)
	next.path = c.path + "/" + segment
	next.hookIndex = 0
	return next
}
