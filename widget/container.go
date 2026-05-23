package widget

import (
	"image"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
	layout "github.com/xiaowumin-mark/FluxUI/layout"
	style "github.com/xiaowumin-mark/FluxUI/style"

	"gioui.org/f32"
)

type containerWidget struct {
	style style.Style
	child Widget
}

type decorationContainerWidget struct {
	decoration style.Decoration
	child      Widget
}

type paddingWidget struct {
	insets style.Insets
	child  Widget
}

// Deprecated: Container 已被 ContainerDecoration 取代。请使用 ContainerDecoration + Decoration 构造器链。
func Container(st style.Style, child Widget) Widget {
	return &containerWidget{
		style: st,
		child: child,
	}
}

// ContainerDecoration 基于 Decoration 创建带背景、内/外边距、圆角和边框装饰的容器。
func ContainerDecoration(d style.Decoration, child Widget) Widget {
	return &decorationContainerWidget{
		decoration: d,
		child:      child,
	}
}

// Padding 创建仅带内边距的容器。
func Padding(insets style.Insets, child Widget) Widget {
	return &paddingWidget{
		insets: insets,
		child:  child,
	}
}

func (c *containerWidget) Layout(ctx *internal.Context) layout.Dimensions {
	size := ctx.LayoutInset(toInternalInsets(c.style.Margin), func(marginCtx *internal.Context) image.Point {
		return marginCtx.LayoutSurface(internal.SurfaceSpec{
			Background: c.style.Background,
			Radius:     c.style.Radius,
			Padding:    toInternalInsets(c.style.Padding),
		}, func(contentCtx *internal.Context) image.Point {
			if c.child == nil {
				return image.Point{}
			}
			return c.child.Layout(contentCtx.Child(0)).Size
		})
	})
	return layout.Dimensions{Size: size}
}

func (c *decorationContainerWidget) Layout(ctx *internal.Context) layout.Dimensions {
	bg := c.decoration.ResolveBg(ctx.Theme().Surface)
	pad := c.decoration.ResolvePad(style.Insets{})
	rad := c.decoration.ResolveRad(0)
	margin := c.decoration.ResolveMargin(style.Insets{})
	border := c.decoration.ResolveBorder(style.Border{})
	opacity := c.decoration.ResolveOpacity()
	grad := c.decoration.ResolveGradient()
	shadow := c.decoration.ResolveShadow()

	spec := internal.SurfaceSpec{
		Background:  bg,
		Radius:      rad,
		Padding:     toInternalInsets(pad),
		BorderColor: border.Color,
		BorderWidth: border.Width,
		Opacity:     opacity,
		CircleClip:  c.decoration.CircleClip,
	}

	if grad != nil {
		spec.HasGradient = true
		spec.GradientStart = f32.Pt(float32(grad.Start.X), float32(grad.Start.Y))
		spec.GradientEnd = f32.Pt(float32(grad.End.X), float32(grad.End.Y))
		spec.GradientFrom = grad.From
		spec.GradientTo = grad.To
	}

	if shadow != nil {
		spec.HasShadow = true
		spec.ShadowOffsetX = shadow.OffsetX
		spec.ShadowOffsetY = shadow.OffsetY
		spec.ShadowBlur = shadow.Blur
		spec.ShadowColor = shadow.Color
	}

	size := ctx.LayoutInset(toInternalInsets(margin), func(marginCtx *internal.Context) image.Point {
		return marginCtx.LayoutSurface(spec, func(contentCtx *internal.Context) image.Point {
			if c.child == nil {
				return image.Point{}
			}
			return c.child.Layout(contentCtx.Child(0)).Size
		})
	})
	return layout.Dimensions{Size: size}
}

func (p *paddingWidget) Layout(ctx *internal.Context) layout.Dimensions {
	size := ctx.LayoutInset(toInternalInsets(p.insets), func(contentCtx *internal.Context) image.Point {
		if p.child == nil {
			return image.Point{}
		}
		return p.child.Layout(contentCtx.Child(0)).Size
	})
	return layout.Dimensions{Size: size}
}
