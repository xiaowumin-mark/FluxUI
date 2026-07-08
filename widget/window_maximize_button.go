package widget

import (
	"image"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"

	"gioui.org/io/system"
	"gioui.org/op/clip"
)

type windowMaximizeButtonWidget struct {
	child  Widget
	config windowMaximizeButtonConfig
}

// WindowMaximizeButtonOption configures WindowMaximizeButton.
type WindowMaximizeButtonOption func(*windowMaximizeButtonConfig)

type windowMaximizeButtonConfig struct {
	disabled bool
}

// WindowMaximizeButtonDisabled disables native maximize-button hit testing
// without changing child layout.
func WindowMaximizeButtonDisabled(disabled bool) WindowMaximizeButtonOption {
	return func(cfg *windowMaximizeButtonConfig) {
		cfg.disabled = disabled
	}
}

// WindowMaximizeButton marks the child area as a Windows native maximize
// button hit-test region. On Windows this lets the OS show Snap Layouts when
// hovering the custom chrome button. Unsupported platforms only lay out child.
func WindowMaximizeButton(child Widget, opts ...WindowMaximizeButtonOption) Widget {
	cfg := windowMaximizeButtonConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &windowMaximizeButtonWidget{child: child, config: cfg}
}

func (w *windowMaximizeButtonWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if w == nil || w.child == nil {
		return layout.Dimensions{}
	}

	pos := ctx.Position()
	dims := w.child.Layout(ctx.Child(0))
	if !w.config.disabled && dims.Size.X > 0 && dims.Size.Y > 0 {
		rect := image.Rectangle{
			Min: pos,
			Max: pos.Add(dims.Size),
		}
		ctx.RegisterWindowActionInput()
		ctx.RegisterWindowMaximizeButton(rect)

		actionRect := image.Rectangle{Max: dims.Size}
		if viewport, ok := ctx.Viewport(); ok {
			clipped := rect.Intersect(viewport)
			if !clipped.Empty() {
				actionRect = image.Rectangle{
					Min: clipped.Min.Sub(pos),
					Max: clipped.Max.Sub(pos),
				}
			}
		}
		stack := clip.Rect(actionRect).Push(ctx.Gtx.Ops)
		system.ActionInputOp(system.ActionMaximize).Add(ctx.Gtx.Ops)
		stack.Pop()
	}
	return dims
}
