package style

import "image/color"

// Decoration 为任何 widget 提供可选的视觉表面属性。所有指针字段为 nil 时
// 表示"使用组件自身默认值"。
//
// 推荐通过 ui 包全局构造函数链式组合：
//
//	ui.Bg(primary).Pad(ui.All(12)).Rad(8).Margin(ui.LeftRight(16)).BorderDeco(2, gray)
type Decoration struct {
	Background *color.NRGBA
	Gradient   *LinearGradient
	Padding    *Insets
	Margin     *Insets
	Radius     *float32
	Corner     *CornerShapes
	Border     *Border
	Opacity    *float32
	CircleClip bool
	Shadow     *BoxShadow

	Hover    *Decoration
	Pressed  *Decoration
	Focused  *Decoration
	Disabled *Decoration

	Image *ImageFill

	Transform *Transform2D
}

// WithBg 设置纯色背景。
func (d Decoration) WithBg(c color.NRGBA) Decoration {
	d.Background = &c
	return d
}

// WithGradient 设置渐变背景。设置后 Background 被忽略。
func (d Decoration) WithGradient(g LinearGradient) Decoration {
	d.Gradient = &g
	return d
}

// WithPad 设置内边距。
func (d Decoration) WithPad(p Insets) Decoration {
	d.Padding = &p
	return d
}

// WithMargin 设置外边距。
func (d Decoration) WithMargin(m Insets) Decoration {
	d.Margin = &m
	return d
}

// WithRad 设置圆角半径。
func (d Decoration) WithRad(r float32) Decoration {
	d.Radius = &r
	return d
}

// WithCornerShape sets the shape for every rounded corner.
func (d Decoration) WithCornerShape(shape CornerShape) Decoration {
	shapes := UniformCornerShape(shape)
	d.Corner = &shapes
	return d
}

// WithCornerShapes sets per-corner shapes in CSS order:
// top-left, top-right, bottom-right, bottom-left.
func (d Decoration) WithCornerShapes(topLeft, topRight, bottomRight, bottomLeft CornerShape) Decoration {
	shapes := CornerShapes{
		TopLeft:     topLeft,
		TopRight:    topRight,
		BottomRight: bottomRight,
		BottomLeft:  bottomLeft,
	}
	d.Corner = &shapes
	return d
}

// WithBorder 设置边框。
func (d Decoration) WithBorder(b Border) Decoration {
	d.Border = &b
	return d
}

// WithOpacity 设置不透明度（0 完全透明 ~ 1 完全不透明）。
func (d Decoration) WithOpacity(v float32) Decoration {
	d.Opacity = &v
	return d
}

// WithCircleClip 启用圆形裁切（自动取 min(w,h) 为直径的内切圆）。
func (d Decoration) WithCircleClip() Decoration {
	d.CircleClip = true
	return d
}

// WithShadow 设置阴影。
func (d Decoration) WithShadow(s BoxShadow) Decoration {
	d.Shadow = &s
	return d
}

// WithHover 设置悬浮态装饰。
func (d Decoration) WithHover(h Decoration) Decoration {
	d.Hover = &h
	return d
}

// WithPressed 设置按下态装饰。
func (d Decoration) WithPressed(p Decoration) Decoration {
	d.Pressed = &p
	return d
}

// WithFocused 设置聚焦态装饰。
func (d Decoration) WithFocused(f Decoration) Decoration {
	d.Focused = &f
	return d
}

// WithDisabled 设置禁用态装饰。
func (d Decoration) WithDisabled(dd Decoration) Decoration {
	d.Disabled = &dd
	return d
}

// WithImage 设置背景图片。
func (d Decoration) WithImage(f ImageFill) Decoration {
	d.Image = &f
	return d
}

// WithTransform 设置 2D 仿射变换。
func (d Decoration) WithTransform(t Transform2D) Decoration {
	d.Transform = &t
	return d
}

// Merge 返回新的 Decoration，其中 other 的非 nil 字段覆盖当前值。
func (d Decoration) Merge(other Decoration) Decoration {
	if other.Background != nil {
		d.Background = other.Background
	}
	if other.Gradient != nil {
		d.Gradient = other.Gradient
	}
	if other.Padding != nil {
		d.Padding = other.Padding
	}
	if other.Margin != nil {
		d.Margin = other.Margin
	}
	if other.Radius != nil {
		d.Radius = other.Radius
	}
	if other.Corner != nil {
		d.Corner = other.Corner
	}
	if other.Border != nil {
		d.Border = other.Border
	}
	if other.Opacity != nil {
		d.Opacity = other.Opacity
	}
	if other.CircleClip {
		d.CircleClip = true
	}
	if other.Shadow != nil {
		d.Shadow = other.Shadow
	}
	if other.Hover != nil {
		d.Hover = other.Hover
	}
	if other.Pressed != nil {
		d.Pressed = other.Pressed
	}
	if other.Focused != nil {
		d.Focused = other.Focused
	}
	if other.Disabled != nil {
		d.Disabled = other.Disabled
	}
	if other.Image != nil {
		d.Image = other.Image
	}
	if other.Transform != nil {
		d.Transform = other.Transform
	}
	return d
}

// ResolveBg 若 Background 已设置则返回其值，否则返回 defaultBg。
func (d Decoration) ResolveBg(defaultBg color.NRGBA) color.NRGBA {
	if d.Background != nil {
		return *d.Background
	}
	return defaultBg
}

// ResolveGradient 返回渐变设置（未设置返回 nil）。
func (d Decoration) ResolveGradient() *LinearGradient {
	return d.Gradient
}

// ResolvePad 若 Padding 已设置则返回其值，否则返回 defaultPad。
func (d Decoration) ResolvePad(defaultPad Insets) Insets {
	if d.Padding != nil {
		return *d.Padding
	}
	return defaultPad
}

// ResolveMargin 若 Margin 已设置则返回其值，否则返回 defaultMargin。
func (d Decoration) ResolveMargin(defaultMargin Insets) Insets {
	if d.Margin != nil {
		return *d.Margin
	}
	return defaultMargin
}

// ResolveRad 若 Radius 已设置则返回其值，否则返回 defaultRad。
func (d Decoration) ResolveRad(defaultRad float32) float32 {
	if d.Radius != nil {
		return *d.Radius
	}
	return defaultRad
}

// ResolveCornerShape returns configured corner shapes or all-round by default.
func (d Decoration) ResolveCornerShape() CornerShapes {
	if d.Corner != nil {
		return *d.Corner
	}
	return CornerShapes{}
}

// ResolveBorder 若 Border 已设置则返回其值，否则返回 defaultBorder。
func (d Decoration) ResolveBorder(defaultBorder Border) Border {
	if d.Border != nil {
		return *d.Border
	}
	return defaultBorder
}

// ResolveOpacity 若 Opacity 已设置则返回其值，否则返回 1.0（完全不透明）。
func (d Decoration) ResolveOpacity() float32 {
	if d.Opacity != nil {
		return *d.Opacity
	}
	return 1.0
}

// ResolveShadow 若 Shadow 已设置则返回其值，否则返回 nil。
func (d Decoration) ResolveShadow() *BoxShadow {
	return d.Shadow
}

// ResolveHover 返回悬浮态装饰。
func (d Decoration) ResolveHover() *Decoration {
	return d.Hover
}

// ResolvePressed 返回按下态装饰。
func (d Decoration) ResolvePressed() *Decoration {
	return d.Pressed
}

// ResolveFocused 返回聚焦态装饰。
func (d Decoration) ResolveFocused() *Decoration {
	return d.Focused
}

// ResolveDisabled 返回禁用态装饰。
func (d Decoration) ResolveDisabled() *Decoration {
	return d.Disabled
}

// ResolveImage 返回背景图片（未设置返回 nil）。
func (d Decoration) ResolveImage() *ImageFill {
	return d.Image
}

// ResolveTransform 返回变换（未设置返回 nil）。
func (d Decoration) ResolveTransform() *Transform2D {
	return d.Transform
}
