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

// InputOption 是输入框配置项。
type InputOption = widget.InputOption

// CheckboxOption 是复选框配置项。
type CheckboxOption = widget.CheckboxOption

// SwitchOption 是开关配置项。
type SwitchOption = widget.SwitchOption

// SliderOption 是滑块配置项。
type SliderOption = widget.SliderOption

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

// Stack 创建堆叠布局。
func Stack(children ...Widget) Widget {
	return widget.Stack(children...)
}

// Center 创建居中布局。
func Center(child Widget) Widget {
	return widget.Center(child)
}

// Flexed 创建带权重的弹性子项。
func Flexed(weight float32, child Widget) Widget {
	return widget.Flexed(weight, child)
}

// Expanded 创建权重为 1 的弹性子项。
func Expanded(child Widget) Widget {
	return widget.Expanded(child)
}

// Text 创建文本组件。
func Text(content string, opts ...TextOption) Widget {
	return widget.Text(content, opts...)
}

// Button 创建按钮组件。
func Button(child Widget, opts ...ButtonOption) Widget {
	return widget.Button(child, opts...)
}

// TextField 创建输入框组件。
func TextField(value string, opts ...InputOption) Widget {
	return widget.TextField(value, opts...)
}

// Checkbox 创建复选框组件。
func Checkbox(label string, checked bool, opts ...CheckboxOption) Widget {
	return widget.Checkbox(label, checked, opts...)
}

// Switch 创建开关组件。
func Switch(checked bool, opts ...SwitchOption) Widget {
	return widget.Switch(checked, opts...)
}

// Slider 创建滑块组件。
func Slider(value float32, opts ...SliderOption) Widget {
	return widget.Slider(value, opts...)
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

// InputPlaceholder 设置输入框占位符。
func InputPlaceholder(text string) InputOption {
	return widget.InputPlaceholder(text)
}

// InputPadding 设置输入框内边距。
func InputPadding(insets Insets) InputOption {
	return widget.InputPadding(insets)
}

// InputRadius 设置输入框圆角。
func InputRadius(radius float32) InputOption {
	return widget.InputRadius(radius)
}

// InputBorder 设置输入框边框颜色。
func InputBorder(color color.NRGBA) InputOption {
	return widget.InputBorder(color)
}

// InputBorderFocus 设置输入框聚焦时边框颜色。
func InputBorderFocus(color color.NRGBA) InputOption {
	return widget.InputBorderFocus(color)
}

// InputBackground 设置输入框背景色。
func InputBackground(color color.NRGBA) InputOption {
	return widget.InputBackground(color)
}

// InputForeground 设置输入框前景色。
func InputForeground(color color.NRGBA) InputOption {
	return widget.InputForeground(color)
}

// InputTextSize 设置输入框字号。
func InputTextSize(size float32) InputOption {
	return widget.InputTextSize(size)
}

// InputMaxLen 设置输入框最大长度。
func InputMaxLen(maxLen int) InputOption {
	return widget.InputMaxLen(maxLen)
}

// InputPassword 设置密码模式。
func InputPassword(password bool) InputOption {
	return widget.InputPassword(password)
}

// InputSingleLine 设置单行模式。
func InputSingleLine(singleLine bool) InputOption {
	return widget.InputSingleLine(singleLine)
}

// InputDisabled 设置输入框禁用状态。
func InputDisabled(disabled bool) InputOption {
	return widget.InputDisabled(disabled)
}

// InputOnChange 绑定输入框内容变化事件。
func InputOnChange(fn func(ctx *Context, value string)) InputOption {
	return widget.InputOnChange(fn)
}

// InputOnFocus 绑定输入框焦点变化事件。
func InputOnFocus(fn func(ctx *Context, focused bool)) InputOption {
	return widget.InputOnFocus(fn)
}

// CheckboxOnChange 绑定复选框变化事件。
func CheckboxOnChange(fn func(ctx *Context, checked bool)) CheckboxOption {
	return widget.CheckboxOnChange(fn)
}

// CheckboxDisabled 设置复选框禁用状态。
func CheckboxDisabled(disabled bool) CheckboxOption {
	return widget.CheckboxDisabled(disabled)
}

// CheckboxSize 设置复选框大小。
func CheckboxSize(size float32) CheckboxOption {
	return widget.CheckboxSize(size)
}

// CheckboxColor 设置复选框颜色。
func CheckboxColor(color color.NRGBA) CheckboxOption {
	return widget.CheckboxColor(color)
}

// SwitchDisabled 设置开关禁用状态。
func SwitchDisabled(disabled bool) SwitchOption {
	return widget.SwitchDisabled(disabled)
}

// SwitchWidth 设置开关宽度。
func SwitchWidth(width float32) SwitchOption {
	return widget.SwitchWidth(width)
}

// SwitchHeight 设置开关高度。
func SwitchHeight(height float32) SwitchOption {
	return widget.SwitchHeight(height)
}

// SwitchColor 设置开关颜色。
func SwitchColor(color color.NRGBA) SwitchOption {
	return widget.SwitchColor(color)
}

// SwitchTrackColor 设置开关轨道颜色。
func SwitchTrackColor(color color.NRGBA) SwitchOption {
	return widget.SwitchTrackColor(color)
}

// SwitchThumbColor 设置开关拇指颜色。
func SwitchThumbColor(color color.NRGBA) SwitchOption {
	return widget.SwitchThumbColor(color)
}

// SwitchOnChange 绑定开关变化事件。
func SwitchOnChange(fn func(ctx *Context, checked bool)) SwitchOption {
	return widget.SwitchOnChange(fn)
}

// SliderDisabled 设置滑块禁用状态。
func SliderDisabled(disabled bool) SliderOption {
	return widget.SliderDisabled(disabled)
}

// SliderMin 设置滑块最小值。
func SliderMin(min float32) SliderOption {
	return widget.SliderMin(min)
}

// SliderMax 设置滑块最大值。
func SliderMax(max float32) SliderOption {
	return widget.SliderMax(max)
}

// SliderStep 设置滑块步进值。
func SliderStep(step float32) SliderOption {
	return widget.SliderStep(step)
}

// SliderWidth 设置滑块宽度。
func SliderWidth(width float32) SliderOption {
	return widget.SliderWidth(width)
}

// SliderTrackColor 设置滑块轨道颜色。
func SliderTrackColor(color color.NRGBA) SliderOption {
	return widget.SliderTrackColor(color)
}

// SliderThumbColor 设置滑块拇指颜色。
func SliderThumbColor(color color.NRGBA) SliderOption {
	return widget.SliderThumbColor(color)
}

// SliderProgressColor 设置滑块进度颜色。
func SliderProgressColor(color color.NRGBA) SliderOption {
	return widget.SliderProgressColor(color)
}

// SliderOnChange 绑定滑块变化事件。
func SliderOnChange(fn func(ctx *Context, value float32)) SliderOption {
	return widget.SliderOnChange(fn)
}
