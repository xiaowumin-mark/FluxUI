package widget

import (
	"image"

	"fluxui/event"
	"fluxui/internal"
	"fluxui/layout"
)

// ClickArea 创建无视觉反馈的可点击区域。
func ClickArea(child Widget, onClick func(ctx *internal.Context)) Widget {
	return &clickAreaWidget{
		child:   child,
		onClick: onClick,
	}
}

type clickAreaWidget struct {
	child   Widget
	onClick func(ctx *internal.Context)
}

func (c *clickAreaWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if c.child == nil {
		return layout.Dimensions{}
	}

	clickable := event.UseClickable(ctx)
	for clickable.Clicked(ctx) {
		if c.onClick != nil {
			c.onClick(ctx)
		}
	}

	size := ctx.LayoutClickArea(clickable.Handle(), func(childCtx *internal.Context) image.Point {
		childDims := c.child.Layout(childCtx.Child(0))
		return childDims.Size
	})
	return layout.Dimensions{Size: size}
}
