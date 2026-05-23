package ui

import (
	"image"
	"image/color"
	"time"

	anim "github.com/xiaowumin-mark/FluxUI/anim"
	fluxapp "github.com/xiaowumin-mark/FluxUI/app"
	internal "github.com/xiaowumin-mark/FluxUI/internal"
	state "github.com/xiaowumin-mark/FluxUI/state"
	style "github.com/xiaowumin-mark/FluxUI/style"
	theme "github.com/xiaowumin-mark/FluxUI/theme"
	widget "github.com/xiaowumin-mark/FluxUI/widget"
)

// Widget 是对外暴露的统一组件接口。
//
// Deprecated: Widget 已不再维护，推荐使用 React-style Element API (RunElement + Element)。
type Widget = widget.Widget

// Context 是对外暴露的 frame 上下文。
type Context = internal.Context

// AppOption 是应用配置项。
type AppOption = fluxapp.Option

// WindowSpec 是多窗口启动配置。
type WindowSpec = fluxapp.WindowSpec

// WindowID 是窗口唯一标识。
type WindowID = fluxapp.WindowID

// WindowHandle 是运行中的窗口句柄。
type WindowHandle = fluxapp.WindowHandle

// Insets 是公开的边距类型。
type Insets = style.Insets

// Style 是公开的容器样式。
type Style = style.Style

// Theme 是公开主题类型。
type Theme = theme.Theme

// ColorScheme 定义语义化色板。
type ColorScheme = theme.ColorScheme

// FontSpec 是公开字体规格。
type FontSpec = theme.FontSpec

// FontFace 是公开字体面。
type FontFace = theme.FontFace

// FontStyle 是字体样式枚举。
type FontStyle = theme.FontStyle

// FontWeight 是字体字重枚举。
type FontWeight = theme.FontWeight

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
type ButtonRef = widget.ButtonRef
type ClickAreaRef = widget.ClickAreaRef
type InputRef = widget.InputRef
type CheckboxRef = widget.CheckboxRef
type SwitchRef = widget.SwitchRef
type SliderRef = widget.SliderRef

// TextAlignment 是文本对齐枚举。
type TextAlignment = widget.TextAlignment

const (
	AlignStart  = widget.AlignStart
	AlignCenter = widget.AlignCenter
	AlignEnd    = widget.AlignEnd
)

var (
	Linear        anim.Easing = anim.Linear
	EaseOut       anim.Easing = anim.EaseOut
	EaseInOut     anim.Easing = anim.EaseInOut
	EaseIn        anim.Easing = anim.EaseInQuad
	EaseInBack    anim.Easing = anim.EaseInBack
	EaseOutBack   anim.Easing = anim.EaseOutBack
	EaseInOutBack anim.Easing = anim.EaseInOutBack
	EaseOutBounce anim.Easing = anim.EaseOutBounce
)

const (
	FontStyleRegular = theme.FontStyleRegular
	FontStyleItalic  = theme.FontStyleItalic
)

const (
	FontWeightThin       = theme.FontWeightThin
	FontWeightExtraLight = theme.FontWeightExtraLight
	FontWeightLight      = theme.FontWeightLight
	FontWeightNormal     = theme.FontWeightNormal
	FontWeightMedium     = theme.FontWeightMedium
	FontWeightSemiBold   = theme.FontWeightSemiBold
	FontWeightBold       = theme.FontWeightBold
	FontWeightExtraBold  = theme.FontWeightExtraBold
	FontWeightBlack      = theme.FontWeightBlack
)

// App 创建应用对象。
//
// Deprecated: App 已不再维护。请使用 RunElement 配合组件函数 (Component)。
func App(root func(ctx *Context) Widget, opts ...AppOption) *fluxapp.Application {
	return fluxapp.New(func(ctx *internal.Context) widget.Widget {
		return root(ctx)
	}, opts...)
}

// Run 启动应用。
//
// Deprecated: Run 已不再维护。请使用 RunElement 配合组件函数 (Component)。
func Run(root func(ctx *Context) Widget, opts ...AppOption) error {
	return fluxapp.Run(func(ctx *internal.Context) widget.Widget {
		return root(ctx)
	}, opts...)
}

// Window 创建多窗口启动中的单个窗口定义。
//
// Deprecated: Window 已不再维护。请使用 RunElement。
func Window(root func(ctx *Context) Widget, opts ...AppOption) WindowSpec {
	return fluxapp.Window(func(ctx *internal.Context) widget.Widget {
		return root(ctx)
	}, opts...)
}

// RunMulti 同时启动多个窗口（桌面端）。
//
// Deprecated: RunMulti 已不再维护。请使用 RunElement。
func RunMulti(windows ...WindowSpec) error {
	return fluxapp.RunMulti(windows...)
}

// ListWindows 返回当前所有存活窗口。
func ListWindows() []WindowHandle {
	return fluxapp.ListWindows()
}

// GetWindow 按 ID 查询窗口句柄。
func GetWindow(id WindowID) (WindowHandle, bool) {
	return fluxapp.GetWindow(id)
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

// NewTheme 从色板创建主题。
func NewTheme(cs ColorScheme) *Theme {
	return theme.New(cs)
}

// LightColors 返回浅色色板。
func LightColors() ColorScheme {
	return theme.LightColors()
}

// DarkColors 返回深色色板。
func DarkColors() ColorScheme {
	return theme.DarkColors()
}

// WithFonts 追加全局字体集合。
func WithFonts(faces ...FontFace) AppOption {
	return fluxapp.WithFonts(faces...)
}

// WithDefaultFont 设置全局默认字体。
func WithDefaultFont(spec FontSpec) AppOption {
	return fluxapp.WithDefaultFont(spec)
}

// WithSystemFonts 控制是否启用系统字体回退。
func WithSystemFonts(enabled bool) AppOption {
	return fluxapp.WithSystemFonts(enabled)
}

// FontFamily 创建字体规格。
func FontFamily(family string) FontSpec {
	return theme.FontFamily(family)
}

// ListFontFamilies 返回去重后的字体族名称。
func ListFontFamilies(faces []FontFace) []string {
	return theme.ListFontFamilies(faces)
}

// DefaultFontSpec 返回默认字体规格。
func DefaultFontSpec() FontSpec {
	return theme.DefaultFontSpec()
}

// ParseFontFile 解析单个字体文件。
func ParseFontFile(path string) ([]FontFace, error) {
	return theme.ParseFontFile(path)
}

// LoadFontsFromPaths 加载多个字体文件。
func LoadFontsFromPaths(paths ...string) ([]FontFace, error) {
	return theme.LoadFontsFromPaths(paths...)
}

// LoadFontsFromDir 递归加载目录下字体文件。
func LoadFontsFromDir(dir string) ([]FontFace, error) {
	return theme.LoadFontsFromDir(dir)
}

// DiscoverSystemFonts 扫描系统字体。
func DiscoverSystemFonts() ([]FontFace, error) {
	return theme.DiscoverSystemFonts()
}

// DiscoverSystemFontFamilies 扫描系统字体族名。
func DiscoverSystemFontFamilies() ([]string, error) {
	return theme.DiscoverSystemFontFamilies()
}

// SystemFontDirs 返回常见系统字体目录。
func SystemFontDirs() []string {
	return theme.SystemFontDirs()
}

// UseTheme 返回当前主题。
func UseTheme(ctx *Context) *Theme {
	return ctx.Theme()
}

// UseFont 返回当前作用域默认字体。
func UseFont(ctx *Context) FontSpec {
	return ctx.Font()
}

// CurrentWindowID 返回当前窗口 ID。
func CurrentWindowID(ctx *Context) WindowID {
	return WindowID(ctx.WindowID())
}

// WindowClose 请求关闭当前窗口。
func WindowClose(ctx *Context) bool {
	return ctx.WindowClose()
}

// WindowMinimize 请求最小化当前窗口。
func WindowMinimize(ctx *Context) bool {
	return ctx.WindowMinimize()
}

// WindowMaximize 请求最大化当前窗口。
func WindowMaximize(ctx *Context) bool {
	return ctx.WindowMaximize()
}

// WindowRestore 请求还原当前窗口。
func WindowRestore(ctx *Context) bool {
	return ctx.WindowRestore()
}

// WindowFullscreen 请求全屏当前窗口。
func WindowFullscreen(ctx *Context) bool {
	return ctx.WindowFullscreen()
}

// WindowRaise 请求将当前窗口置顶。
func WindowRaise(ctx *Context) bool {
	return ctx.WindowRaise()
}

// WindowCenter 请求将当前窗口居中。
func WindowCenter(ctx *Context) bool {
	return ctx.WindowCenter()
}

// WindowSetTitle 更新当前窗口标题。
func WindowSetTitle(ctx *Context, title string) bool {
	return ctx.WindowSetTitle(title)
}

// WindowSetSize 更新当前窗口尺寸（单位 dp）。
func WindowSetSize(ctx *Context, width, height int) bool {
	return ctx.WindowSetSize(width, height)
}

// WindowInvalidate 请求当前窗口立即重绘。
func WindowInvalidate(ctx *Context) bool {
	return ctx.WindowInvalidate()
}

// WindowIsAlive 返回当前窗口是否仍然存活。
func WindowIsAlive(ctx *Context) bool {
	return ctx.WindowIsAlive()
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
//
// Deprecated: Container 已被 ContainerDecoration 取代。
func Container(st Style, child Widget) Widget {
	return widget.Container(st, child)
}

// ContainerDecorationOption 是装饰容器的可选配置项。
type ContainerDecorationOption = widget.ContainerDecorationOption

// ContainerDecoration 基于 Decoration 创建装饰容器。
// 可选 opts 支持交互状态和事件回调。
func ContainerDecoration(d Decoration, child Widget, opts ...ContainerDecorationOption) Widget {
	return widget.ContainerDecoration(d, child, opts...)
}

// WithFont 在子树中覆盖默认字体。
func WithFont(font FontSpec, child Widget) Widget {
	return widget.WithFont(font, child)
}

// Padding 创建带边距的容器。
func Padding(insets Insets, child Widget) Widget {
	return widget.Padding(insets, child)
}

// State 返回当前作用域的稳定状态。
func State[T any](ctx *Context) *state.State[T] {
	return state.Use[T](ctx)
}

// Easing 定义插值曲线类型。
type Easing = anim.Easing

// Animate 创建动画定义。
//
// Deprecated: Animate 是旧 Widget API 的动画构造器。新代码请使用 UseAnimatedValue 或 UseAnimatedDecoration。
func Animate(opts ...anim.Option) *anim.Animation {
	return anim.New(opts...)
}

// Duration 配置动画时长。
//
// Deprecated: 新代码请直接传 time.Duration 给 UseAnimatedValue / UseAnimatedDecoration。
func Duration(duration time.Duration) anim.Option {
	return anim.Duration(duration)
}

// From 配置动画起始值。
//
// Deprecated: 新代码请直接传 target 给 UseAnimatedValue / UseAnimatedDecoration。
func From(value float32) anim.Option {
	return anim.From(value)
}

// To 配置动画结束值。
//
// Deprecated: 新代码请直接传 target 给 UseAnimatedValue / UseAnimatedDecoration。
func To(value float32) anim.Option {
	return anim.To(value)
}

// Ease 配置动画缓动函数。
//
// Deprecated: 新代码请直接传 anim.Easing 给 UseAnimatedValue / UseAnimatedDecoration。
func Ease(easing anim.Easing) anim.Option {
	return anim.Ease(easing)
}

// CubicBezier 创建自定义三次贝塞尔缓动曲线 (x1,y1,x2,y2 均在 [0,1])。
func CubicBezier(x1, y1, x2, y2 float32) anim.Easing {
	return anim.CubicBezier(x1, y1, x2, y2)
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

// TextFont 设置文本字体（局部覆盖）。
func TextFont(font FontSpec) TextOption {
	return widget.TextFont(font)
}

// TextFontWeight 设置文本字体字重（局部覆盖）。
func TextFontWeight(weight FontWeight) TextOption {
	return widget.TextFontWeight(weight)
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

// ButtonDecoration 通过 Decoration 统一设置按钮背景、内边距和圆角。
func ButtonDecoration(d Decoration) ButtonOption {
	return widget.ButtonDecoration(d)
}

func NewButtonRef() *ButtonRef {
	return widget.NewButtonRef()
}

func ButtonAttachRef(ref *ButtonRef) ButtonOption {
	return widget.ButtonAttachRef(ref)
}

// All 创建统一边距。
func All(value float32) Insets {
	return style.All(value)
}

// Symmetric 创建对称边距。
func Symmetric(vertical, horizontal float32) Insets {
	return style.Symmetric(vertical, horizontal)
}

// Only 创建四周独立的边距。
func Only(top, right, bottom, left float32) Insets {
	return style.Only(top, right, bottom, left)
}

// LeftRight 创建左右水平边距。
func LeftRight(v float32) Insets {
	return style.Horizontal(v)
}

// TopBottom 创建上下垂直边距。
func TopBottom(v float32) Insets {
	return style.Vertical(v)
}

// NRGBA 创建颜色。
func NRGBA(r, g, b, a uint8) color.NRGBA {
	return style.NRGBA(r, g, b, a)
}

// Decoration 提供可选的可视装饰属性（背景 / 内边距 / 外边距 / 圆角 / 边框 / 渐变 / 透明度 / 圆形裁切）。
type Decoration = style.Decoration

// Border 定义边框样式。
type Border = style.Border

// LinearGradient 定义线性渐变。
type LinearGradient = style.LinearGradient

// ImageFill 定义背景图片及其缩放模式。
type ImageFill = style.ImageFill

// ImageFillFit 定义背景图片缩放模式。
type ImageFillFit = style.ImageFillFit

// Transform2D 定义 2D 仿射变换。
type Transform2D = style.Transform2D

// TransformOrigin 定义变换原点。
type TransformOrigin = style.TransformOrigin

const (
	ImageFillContain = style.ImageFillContain
	ImageFillCover   = style.ImageFillCover
	ImageFillFill    = style.ImageFillFill
	ImageFillNone    = style.ImageFillNone
)

const (
	TransformCenter      = style.TransformCenter
	TransformTopLeft     = style.TransformTopLeft
	TransformTopRight    = style.TransformTopRight
	TransformBottomLeft  = style.TransformBottomLeft
	TransformBottomRight = style.TransformBottomRight
)

// Bg 创建仅设置背景色的装饰。
func Bg(c color.NRGBA) style.Decoration {
	return style.Decoration{}.WithBg(c)
}

// Pad 创建仅设置内边距的装饰。
func Pad(p Insets) style.Decoration {
	return style.Decoration{}.WithPad(p)
}

// Margin 创建仅设置外边距的装饰。
func Margin(m Insets) style.Decoration {
	return style.Decoration{}.WithMargin(m)
}

// Rad 创建仅设置圆角的装饰。
func Rad(r float32) style.Decoration {
	return style.Decoration{}.WithRad(r)
}

// BorderDeco 创建仅设置边框的装饰。
func BorderDeco(width float32, col color.NRGBA) style.Decoration {
	return style.Decoration{}.WithBorder(style.Border{Width: width, Color: col})
}

// Opacity 创建仅设置不透明度的装饰（0 完全透明 ~ 1 完全不透明）。
func Opacity(v float32) style.Decoration {
	return style.Decoration{}.WithOpacity(v)
}

// LinearGrad 创建仅设置线性渐变的装饰。
// start 和 end 为组件本地坐标系内的渐变方向点。
func LinearGrad(start, end image.Point, from, to color.NRGBA) style.Decoration {
	return style.Decoration{}.WithGradient(style.LinearGradient{
		Start: start,
		End:   end,
		From:  from,
		To:    to,
	})
}

// Circle 创建仅启用圆形裁切的装饰。
func Circle() style.Decoration {
	return style.Decoration{}.WithCircleClip()
}

// Shadow 创建仅设置阴影的装饰。offset/blur 单位为 dp。
func Shadow(offX, offY, blur float32, col color.NRGBA) style.Decoration {
	return style.Decoration{}.WithShadow(style.BoxShadow{
		OffsetX: offX,
		OffsetY: offY,
		Blur:    blur,
		Color:   col,
	})
}

// Elevation 创建 Material Design 高度等级对应的阴影装饰（1~5）。
// 1=按钮hover 2=卡片 3=浮卡/FAB 4=对话框 5=模态。
func Elevation(level int) style.Decoration {
	s := style.ElevationBoxShadow(level)
	return style.Decoration{}.WithShadow(s)
}

// Hover 创建仅设置悬浮态装饰的 Decoration。
func Hover(d Decoration) style.Decoration {
	return style.Decoration{}.WithHover(d)
}

// Pressed 创建仅设置按下态装饰的 Decoration。
func Pressed(d Decoration) style.Decoration {
	return style.Decoration{}.WithPressed(d)
}

// Focused 创建仅设置聚焦态装饰的 Decoration。
func Focused(d Decoration) style.Decoration {
	return style.Decoration{}.WithFocused(d)
}

// DisabledDeco 创建仅设置禁用态装饰的 Decoration。
func DisabledDeco(d Decoration) style.Decoration {
	return style.Decoration{}.WithDisabled(d)
}

// HoverBg 创建悬浮时背景色变化的 Decoration 快捷方式。
func HoverBg(c color.NRGBA) style.Decoration {
	return Hover(Bg(c))
}

// PressedBg 创建按下时背景色变化的 Decoration 快捷方式。
func PressedBg(c color.NRGBA) style.Decoration {
	return Pressed(Bg(c))
}

// ImageBg 创建仅设置背景图片的 Decoration。
func ImageBg(src image.Image, fit ImageFillFit) style.Decoration {
	return style.Decoration{}.WithImage(style.ImageFill{Src: src, Fit: fit})
}

// LoadImage 自动检测 URL 或文件路径，解码图片。
func LoadImage(src string) (image.Image, error) {
	return style.LoadImage(src)
}

// DecodeImageURL 通过 HTTP(S) 下载并解码图片。
func DecodeImageURL(url string) (image.Image, error) {
	return style.DecodeImageURL(url)
}

// DecodeImageFile 从本地文件路径解码图片。
func DecodeImageFile(path string) (image.Image, error) {
	return style.DecodeImageFile(path)
}

// TransformDeco 创建仅设置变换的 Decoration。
func TransformDeco(rotateDeg, scaleX, scaleY, transX, transY float32, origin TransformOrigin) style.Decoration {
	return style.Decoration{}.WithTransform(style.Transform2D{
		RotateDeg: rotateDeg, ScaleX: scaleX, ScaleY: scaleY,
		TranslateX: transX, TranslateY: transY, Origin: origin,
	})
}

// Rotate 创建仅设置旋转的 Decoration。
func Rotate(deg float32) style.Decoration {
	return TransformDeco(deg, 1, 1, 0, 0, TransformCenter)
}

// ScaleDeco 创建仅设置缩放的 Decoration。
func ScaleDeco(sx, sy float32) style.Decoration {
	return TransformDeco(0, sx, sy, 0, 0, TransformCenter)
}

// TranslateDeco 创建仅设置平移的 Decoration。
func TranslateDeco(tx, ty float32) style.Decoration {
	return TransformDeco(0, 1, 1, tx, ty, TransformCenter)
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

// InputFontFamily 设置输入框字体族（局部覆盖）。
func InputFontFamily(family string) InputOption {
	return widget.InputFontFamily(family)
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

func NewInputRef() *InputRef {
	return widget.NewInputRef()
}

func InputAttachRef(ref *InputRef) InputOption {
	return widget.InputAttachRef(ref)
}

// InputDecoration 通过 Decoration 统一设置输入框背景、内边距和圆角。
func InputDecoration(d Decoration) InputOption {
	return widget.InputDecoration(d)
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

func NewCheckboxRef() *CheckboxRef {
	return widget.NewCheckboxRef()
}

func CheckboxAttachRef(ref *CheckboxRef) CheckboxOption {
	return widget.CheckboxAttachRef(ref)
}

// CheckboxDecoration 通过 Decoration 统一设置复选框装饰。
func CheckboxDecoration(d Decoration) CheckboxOption {
	return widget.CheckboxDecoration(d)
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

func NewSwitchRef() *SwitchRef {
	return widget.NewSwitchRef()
}

func SwitchAttachRef(ref *SwitchRef) SwitchOption {
	return widget.SwitchAttachRef(ref)
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

func NewSliderRef() *SliderRef {
	return widget.NewSliderRef()
}

func SliderAttachRef(ref *SliderRef) SliderOption {
	return widget.SliderAttachRef(ref)
}

// ContainerDecorationDisabled 设置容器禁用状态。
func ContainerDecorationDisabled(disabled bool) ContainerDecorationOption {
	return widget.ContainerDecorationDisabled(disabled)
}

// OnDecoClick 设置装饰容器点击回调。
func OnDecoClick(fn func(ctx *Context)) ContainerDecorationOption {
	return widget.ContainerDecorationOnClick(fn)
}

// OnDecoHoverEnter 设置鼠标进入装饰容器回调。
func OnDecoHoverEnter(fn func(ctx *Context)) ContainerDecorationOption {
	return widget.ContainerDecorationOnHoverEnter(fn)
}

// OnDecoHoverLeave 设置鼠标离开装饰容器回调。
func OnDecoHoverLeave(fn func(ctx *Context)) ContainerDecorationOption {
	return widget.ContainerDecorationOnHoverLeave(fn)
}

// OnDecoHover 设置装饰容器悬浮态每帧回调。
func OnDecoHover(fn func(ctx *Context, hovering bool)) ContainerDecorationOption {
	return widget.ContainerDecorationOnHover(fn)
}

// OnDecoPressed 设置装饰容器按下/释放回调。
func OnDecoPressed(fn func(ctx *Context, pressed bool)) ContainerDecorationOption {
	return widget.ContainerDecorationOnPressed(fn)
}
