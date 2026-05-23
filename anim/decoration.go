package anim

import (
	"image/color"

	style "github.com/xiaowumin-mark/FluxUI/style"
)

// LerpDecoration 逐字段插值两个 Decoration。
//
// 平滑插值：Background(颜色), Opacity, Radius, Transform(全部5个float),
// Border(Width+Color), Shadow(OffsetX/Y+Blur+Color), Gradient(两端颜色)。
// 瞬切：Padding, Margin, CircleClip, Image, Border.Style, Transform.Origin,
// Hover/Pressed/Focused/Disabled（交互态）。
func LerpDecoration(from, to style.Decoration, t float32) style.Decoration {
	if t >= 1 {
		return to
	}
	if t <= 0 {
		return from
	}

	result := from

	if from.Background != nil && to.Background != nil {
		c := lerpNRGBA(*from.Background, *to.Background, t)
		result.Background = &c
	} else if to.Background != nil {
		c := *to.Background
		result.Background = &c
	}

	if from.Gradient != nil && to.Gradient != nil {
		g := lerpGradient(*from.Gradient, *to.Gradient, t)
		result.Gradient = &g
	} else if to.Gradient != nil {
		g := *to.Gradient
		result.Gradient = &g
	}

	if from.Radius != nil && to.Radius != nil {
		v := lerp(*from.Radius, *to.Radius, t)
		result.Radius = &v
	} else if to.Radius != nil {
		v := *to.Radius
		result.Radius = &v
	}

	if from.Opacity != nil && to.Opacity != nil {
		v := lerp(*from.Opacity, *to.Opacity, t)
		result.Opacity = &v
	} else if to.Opacity != nil {
		v := *to.Opacity
		result.Opacity = &v
	}

	if from.Border != nil && to.Border != nil {
		b := lerpBorder(*from.Border, *to.Border, t)
		result.Border = &b
	} else if to.Border != nil {
		b := *to.Border
		result.Border = &b
	}

	if from.Shadow != nil && to.Shadow != nil {
		s := lerpShadow(*from.Shadow, *to.Shadow, t)
		result.Shadow = &s
	} else if to.Shadow != nil {
		s := *to.Shadow
		result.Shadow = &s
	}

	if from.Transform != nil && to.Transform != nil {
		tr := lerpTransform(*from.Transform, *to.Transform, t)
		result.Transform = &tr
	} else if to.Transform != nil {
		tr := *to.Transform
		result.Transform = &tr
	}

	if to.Padding != nil {
		result.Padding = to.Padding
	}
	if to.Margin != nil {
		result.Margin = to.Margin
	}
	if to.CircleClip {
		result.CircleClip = true
	}
	if to.Image != nil {
		result.Image = to.Image
	}
	if to.Hover != nil {
		result.Hover = to.Hover
	}
	if to.Pressed != nil {
		result.Pressed = to.Pressed
	}
	if to.Focused != nil {
		result.Focused = to.Focused
	}
	if to.Disabled != nil {
		result.Disabled = to.Disabled
	}

	return result
}

func lerpBorder(from, to style.Border, t float32) style.Border {
	return style.Border{
		Width: lerp(from.Width, to.Width, t),
		Color: lerpNRGBA(from.Color, to.Color, t),
	}
}

func lerpShadow(from, to style.BoxShadow, t float32) style.BoxShadow {
	return style.BoxShadow{
		OffsetX: lerp(from.OffsetX, to.OffsetX, t),
		OffsetY: lerp(from.OffsetY, to.OffsetY, t),
		Blur:    lerp(from.Blur, to.Blur, t),
		Color:   lerpNRGBA(from.Color, to.Color, t),
	}
}

func lerpTransform(from, to style.Transform2D, t float32) style.Transform2D {
	result := style.Transform2D{
		RotateDeg:  lerp(from.RotateDeg, to.RotateDeg, t),
		ScaleX:     lerp(from.ScaleX, to.ScaleX, t),
		ScaleY:     lerp(from.ScaleY, to.ScaleY, t),
		TranslateX: lerp(from.TranslateX, to.TranslateX, t),
		TranslateY: lerp(from.TranslateY, to.TranslateY, t),
		Origin:     to.Origin,
	}
	return result
}

func lerpGradient(from, to style.LinearGradient, t float32) style.LinearGradient {
	return style.LinearGradient{
		Start: from.Start,
		End:   from.End,
		From:  lerpNRGBA(from.From, to.From, t),
		To:    lerpNRGBA(from.To, to.To, t),
	}
}

// DecorationEqual 深度比较两个 Decoration 的所有字段是否相等。
func DecorationEqual(a, b style.Decoration) bool {
	return colorNRGBAPtrEqual(a.Background, b.Background) &&
		gradientEqual(a.Gradient, b.Gradient) &&
		insetsPtrEqual(a.Padding, b.Padding) &&
		insetsPtrEqual(a.Margin, b.Margin) &&
		float32PtrEqual(a.Radius, b.Radius) &&
		borderPtrEqual(a.Border, b.Border) &&
		float32PtrEqual(a.Opacity, b.Opacity) &&
		a.CircleClip == b.CircleClip &&
		shadowPtrEqual(a.Shadow, b.Shadow) &&
		transformPtrEqual(a.Transform, b.Transform) &&
		imageFillPtrEqual(a.Image, b.Image)
}

func colorNRGBAPtrEqual(a, b *color.NRGBA) bool { //nolint:unused // false-positive
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func float32PtrEqual(a, b *float32) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func borderPtrEqual(a, b *style.Border) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Width == b.Width && a.Color == b.Color
}

func shadowPtrEqual(a, b *style.BoxShadow) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.OffsetX == b.OffsetX && a.OffsetY == b.OffsetY && a.Blur == b.Blur && a.Color == b.Color
}

func transformPtrEqual(a, b *style.Transform2D) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.RotateDeg == b.RotateDeg &&
		a.ScaleX == b.ScaleX &&
		a.ScaleY == b.ScaleY &&
		a.TranslateX == b.TranslateX &&
		a.TranslateY == b.TranslateY &&
		a.Origin == b.Origin
}

func gradientEqual(a, b *style.LinearGradient) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Start == b.Start && a.End == b.End && a.From == b.From && a.To == b.To
}

func imageFillPtrEqual(a, b *style.ImageFill) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Src == b.Src && a.Fit == b.Fit
}

func insetsPtrEqual(a, b *style.Insets) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Top == b.Top && a.Right == b.Right && a.Bottom == b.Bottom && a.Left == b.Left
}
