package internal

import (
	"image"
	"image/color"
	"math"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

// RippleSpec 描述 Material 交互 ripple 的视觉参数。
type RippleSpec struct {
	Color   color.NRGBA
	Radius  float32
	Opacity float32
}

// DrawRipple 在给定区域内绘制可复用的有界 ripple 反馈。
func (c *Context) DrawRipple(clickable *ClickableState, size image.Point, spec RippleSpec) {
	if clickable == nil || size.X <= 0 || size.Y <= 0 || spec.Color.A == 0 {
		return
	}
	opacity := spec.Opacity
	if opacity <= 0 {
		opacity = 0.16
	}

	rr := clampRoundedRadiusPx(size, c.Gtx.Dp(unit.Dp(spec.Radius)))
	clipStack := clip.UniformRRect(image.Rectangle{Max: size}, rr).Push(c.Gtx.Ops)
	for _, press := range clickable.History() {
		drawRipplePress(c.Gtx, press, size, spec.Color, opacity)
	}
	clipStack.Pop()
}

func drawRipplePress(gtx gioLayout.Context, press widget.Press, bounds image.Point, col color.NRGBA, opacity float32) {
	const (
		expandDuration = float32(0.45)
		fadeDuration   = float32(0.55)
	)

	now := gtx.Now
	t := float32(now.Sub(press.Start).Seconds())
	end := press.End
	if end.IsZero() {
		end = now
	}
	endt := float32(end.Sub(press.Start).Seconds())

	var alphat float32
	var haste float32
	if press.Cancelled {
		if h := 0.5 - endt/fadeDuration; h > 0 {
			haste = h
		}
	}
	half1 := t/fadeDuration + haste
	if half1 > 0.5 {
		half1 = 0.5
	}
	half2 := float32(now.Sub(end).Seconds())/fadeDuration + haste
	if half2 > 0.5 {
		return
	}
	alphat = half1 + half2

	sizet := t
	if press.Cancelled {
		sizet = endt
	}
	sizet /= expandDuration
	if !press.End.IsZero() || sizet <= 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	if sizet > 1 {
		sizet = 1
	}
	if alphat > 0.5 {
		alphat = 1 - alphat
	}

	alphaBezier := easeInOut(alphat * 2)
	sizeBezier := easeInOut(sizet)
	diameter := bounds.X
	if bounds.Y > diameter {
		diameter = bounds.Y
	}
	diameter = int(float32(diameter) * 2 * float32(math.Sqrt2) * sizeBezier)
	if diameter <= 0 {
		return
	}

	col.A = uint8(float32(col.A)*opacity*alphaBezier + 0.5)
	if col.A == 0 {
		return
	}

	radius := diameter / 2
	offset := op.Offset(press.Position.Add(image.Point{X: -radius, Y: -radius})).Push(gtx.Ops)
	paint.FillShape(gtx.Ops, col, clip.Ellipse(image.Rectangle{Max: image.Pt(diameter, diameter)}).Op(gtx.Ops))
	offset.Pop()
}

func easeInOut(v float32) float32 {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return v * v * (3 - 2*v)
}
