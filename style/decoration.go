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
	Border     *Border
	Opacity    *float32
	CircleClip bool
	Shadow     *BoxShadow
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
