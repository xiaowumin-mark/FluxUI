package ui

import (
	"image/color"
	"time"

	"fluxui/anim"
	fluxapp "fluxui/app"
	"fluxui/internal"
	"fluxui/state"
	"fluxui/style"
	"fluxui/theme"
	"fluxui/widget"
)

// Widget 是对外暴露的统一组件接口。
type Widget = widget.Widget

// Context 是对外暴露的 frame 上下文。
type Context = internal.Context

// AppOption 是应用配置项。
type AppOption = fluxapp.Option

// Insets 是公开的边距类型。
type Insets = style.Insets

// Style 是公开的容器样式。
type Style = style.Style

// Theme 是公开主题类型。
type Theme = theme.Theme

// TextOption 是文本配置项。
type TextOption = widget.TextOption

// ButtonOption 是按钮配置项。
type ButtonOption = widget.ButtonOption

// TextAlignment 是文本对齐枚举。
type TextAlignment = widget.TextAlignment

const (
	AlignStart  = widget.AlignStart
	AlignCenter = widget.AlignCenter
	AlignEnd    = widget.AlignEnd
)

var (
	Linear    anim.Easing = anim.Linear
	EaseOut   anim.Easing = anim.EaseOut
	EaseInOut anim.Easing = anim.EaseInOut
)

// App 创建应用对象。
func App(root func(ctx *Context) Widget, opts ...AppOption) *fluxapp.Application {
	return fluxapp.New(func(ctx *internal.Context) widget.Widget {
		return root(ctx)
	}, opts...)
}

// Run 启动应用。
func Run(root func(ctx *Context) Widget, opts ...AppOption) error {
	return fluxapp.Run(func(ctx *internal.Context) widget.Widget {
		return root(ctx)
	}, opts...)
}

// Title 设置窗口标题。
func Title(value string) AppOption {
	return fluxapp.Title(value)
}

// Size 设置窗口尺寸。
func Size(width, height int) AppOption {
	return fluxapp.Size(width, height)
}

// WithTheme 设置应用主题。
func WithTheme(th *Theme) AppOption {
	return fluxapp.WithTheme(th)
}

// UseTheme 返回当前主题。
func UseTheme(ctx *Context) *Theme {
	return ctx.Theme()
}

// Column 创建纵向布局。
func Column(children ...Widget) Widget {
	return widget.Column(children...)
}

// Row 创建横向布局。
func Row(children ...Widget) Widget {
	return widget.Row(children...)
}

// Text 创建文本组件。
func Text(content string, opts ...TextOption) Widget {
	return widget.Text(content, opts...)
}

// Button 创建按钮组件。
func Button(child Widget, opts ...ButtonOption) Widget {
	return widget.Button(child, opts...)
}

// Container 创建容器组件。
func Container(st Style, child Widget) Widget {
	return widget.Container(st, child)
}

// Padding 创建带边距的容器。
func Padding(insets Insets, child Widget) Widget {
	return widget.Padding(insets, child)
}

// State 返回当前作用域的稳定状态。
func State[T any](ctx *Context) *state.State[T] {
	return state.Use[T](ctx)
}

// Animate 创建动画定义。
func Animate(opts ...anim.Option) *anim.Animation {
	return anim.New(opts...)
}

// Duration 配置动画时长。
func Duration(duration time.Duration) anim.Option {
	return anim.Duration(duration)
}

// From 配置动画起始值。
func From(value float32) anim.Option {
	return anim.From(value)
}

// To 配置动画结束值。
func To(value float32) anim.Option {
	return anim.To(value)
}

// Ease 配置动画缓动函数。
func Ease(easing anim.Easing) anim.Option {
	return anim.Ease(easing)
}

// TextSize 设置文本字号。
func TextSize(size float32) TextOption {
	return widget.TextSize(size)
}

// TextColor 设置文本颜色。
func TextColor(value color.NRGBA) TextOption {
	return widget.TextColor(value)
}

// TextAlign 设置文本对齐。
func TextAlign(alignment TextAlignment) TextOption {
	return widget.TextAlign(alignment)
}

// OnClick 绑定按钮点击事件。
func OnClick(fn func(ctx *Context)) ButtonOption {
	return widget.OnClick(fn)
}

// OnHover 绑定按钮悬浮事件。
func OnHover(fn func(ctx *Context, hovering bool)) ButtonOption {
	return widget.OnHover(fn)
}

// Disabled 设置按钮禁用状态。
func Disabled(disabled bool) ButtonOption {
	return widget.Disabled(disabled)
}

// ButtonPadding 设置按钮内边距。
func ButtonPadding(insets Insets) ButtonOption {
	return widget.ButtonPadding(insets)
}

// ButtonRadius 设置按钮圆角。
func ButtonRadius(radius float32) ButtonOption {
	return widget.ButtonRadius(radius)
}

// ButtonBackground 设置按钮背景色。
func ButtonBackground(value color.NRGBA) ButtonOption {
	return widget.ButtonBackground(value)
}

// ButtonForeground 设置按钮前景色。
func ButtonForeground(value color.NRGBA) ButtonOption {
	return widget.ButtonForeground(value)
}

// All 创建统一边距。
func All(value float32) Insets {
	return style.All(value)
}

// Symmetric 创建对称边距。
func Symmetric(vertical, horizontal float32) Insets {
	return style.Symmetric(vertical, horizontal)
}

// NRGBA 创建颜色。
func NRGBA(r, g, b, a uint8) color.NRGBA {
	return style.NRGBA(r, g, b, a)
}
