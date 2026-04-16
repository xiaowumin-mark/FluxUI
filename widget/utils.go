package widget

import (
	"image"

	"fluxui/internal"
	"fluxui/layout"

	"gioui.org/unit"
)

func safeDp(v float32) unit.Dp {
	if v < 0 {
		return 0
	}
	return unit.Dp(v)
}

func clampFloat32(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func clampPointToConstraints(size image.Point, min, max image.Point) image.Point {
	if size.X < min.X {
		size.X = min.X
	}
	if size.Y < min.Y {
		size.Y = min.Y
	}
	if size.X > max.X {
		size.X = max.X
	}
	if size.Y > max.Y {
		size.Y = max.Y
	}
	return size
}

type expandWidthWidget struct {
	child Widget
}

func expandWidth(child Widget) Widget {
	return &expandWidthWidget{child: child}
}

func (e *expandWidthWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if e.child == nil {
		return layout.Dimensions{}
	}

	gtx := ctx.Gtx
	width := gtx.Constraints.Max.X
	if width <= 0 {
		return e.child.Layout(ctx.Child(0))
	}

	stretch := gtx
	if stretch.Constraints.Min.X < width {
		stretch.Constraints.Min.X = width
	}
	stretch.Constraints.Max.X = width
	if stretch.Constraints.Min.X > stretch.Constraints.Max.X {
		stretch.Constraints.Min.X = stretch.Constraints.Max.X
	}

	next := *ctx
	next.Gtx = stretch
	dims := e.child.Layout(next.Child(0))
	dims.Size.X = width
	return dims
}
