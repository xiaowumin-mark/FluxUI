package widget

import (
	"image"

	"fluxui/event"
	"fluxui/internal"
	"fluxui/layout"
)

type ClickAreaOption func(*clickAreaConfig)

type clickAreaConfig struct {
	ref *ClickAreaRef
}

// ClickArea 创建无视觉反馈的可点击区域。
func ClickArea(child Widget, onClick func(ctx *internal.Context), opts ...ClickAreaOption) Widget {
	cfg := clickAreaConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &clickAreaWidget{
		child:   child,
		onClick: onClick,
		config:  cfg,
	}
}

func ClickAreaAttachRef(ref *ClickAreaRef) ClickAreaOption {
	return func(cfg *clickAreaConfig) {
		cfg.ref = ref
	}
}

type clickAreaWidget struct {
	child   Widget
	onClick func(ctx *internal.Context)
	config  clickAreaConfig
}

func (c *clickAreaWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if c.child == nil {
		return layout.Dimensions{}
	}

	clickable := event.UseClickable(ctx)
	if c.config.ref != nil {
		c.config.ref.bindInvalidator(ctx.Runtime().RequestRedraw)
		for range c.config.ref.drainCommands() {
			if c.onClick != nil {
				c.onClick(ctx)
			}
		}
	}
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
