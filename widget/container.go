package widget

import (
	"image"

	"fluxui/internal"
	"fluxui/layout"
	"fluxui/style"
)

type containerWidget struct {
	style style.Style
	child Widget
}

type paddingWidget struct {
	insets style.Insets
	child  Widget
}

// Container 创建带背景与内外边距的容器。
func Container(st style.Style, child Widget) Widget {
	return &containerWidget{
		style: st,
		child: child,
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

func (p *paddingWidget) Layout(ctx *internal.Context) layout.Dimensions {
	size := ctx.LayoutInset(toInternalInsets(p.insets), func(contentCtx *internal.Context) image.Point {
		if p.child == nil {
			return image.Point{}
		}
		return p.child.Layout(contentCtx.Child(0)).Size
	})
	return layout.Dimensions{Size: size}
}
