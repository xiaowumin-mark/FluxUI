package widget

import (
	"image"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"

	"gioui.org/io/system"
	"gioui.org/op/clip"
)

type windowDragAreaWidget struct {
	child  Widget
	config windowDragAreaConfig
}

type WindowDragAreaOption func(*windowDragAreaConfig)

type windowDragAreaConfig struct {
	disabled bool
}

func WindowDragAreaDisabled(disabled bool) WindowDragAreaOption {
	return func(cfg *windowDragAreaConfig) {
		cfg.disabled = disabled
	}
}

// WindowDragArea marks the child area as a native window move region. On
// unsupported platforms it only lays out the child.
func WindowDragArea(child Widget, opts ...WindowDragAreaOption) Widget {
	cfg := windowDragAreaConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &windowDragAreaWidget{child: child, config: cfg}
}

func (w *windowDragAreaWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if w == nil || w.child == nil {
		return layout.Dimensions{}
	}

	dims := w.child.Layout(ctx.Child(0))
	if !w.config.disabled && dims.Size.X > 0 && dims.Size.Y > 0 {
		ctx.RegisterWindowDragArea()
		stack := clip.Rect(image.Rectangle{Max: dims.Size}).Push(ctx.Gtx.Ops)
		system.ActionInputOp(system.ActionMove).Add(ctx.Gtx.Ops)
		stack.Pop()
	}
	return dims
}
