package internal

import (
	"image"
	"image/color"
	"math"
	"time"

	fluxstyle "github.com/xiaowumin-mark/FluxUI/style"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

type RippleSpec struct {
	Color   color.NRGBA
	Radius  float32
	Opacity float32
}

type rippleFrame struct {
	Diameter   int
	AlphaScale float32
	Invalidate bool
	Expired    bool
}

const (
	rippleExpandDuration = fluxstyle.InteractionRippleExpand
	rippleFadeDuration   = fluxstyle.InteractionRippleFade
)

func (c *Context) DrawRipple(clickable *ClickableState, size image.Point, spec RippleSpec) {
	if clickable == nil || size.X <= 0 || size.Y <= 0 || spec.Color.A == 0 {
		return
	}
	if done := c.startFrameSection(PerfDraw, 1); done != nil {
		defer done()
	}
	opacity := spec.Opacity
	if opacity <= 0 {
		opacity = 0.16
	}

	rr := clampRoundedRadiusPx(size, c.Gtx.Dp(unit.Dp(spec.Radius)))
	clipStack := clip.UniformRRect(image.Rectangle{Max: size}, rr).Push(c.Gtx.Ops)
	for _, press := range clickable.History() {
		c.drawRipplePress(press, size, press.Position, spec.Color, opacity)
	}
	clipStack.Pop()
}

// DrawStateLayerCircle draws a Material state layer centered on a control
// handle without affecting layout or using the pointer position as origin.
func (c *Context) DrawStateLayerCircle(clickable *ClickableState, center image.Point, diameterDp float32, spec RippleSpec, stateOpacity float32) {
	if center.X < 0 || center.Y < 0 || spec.Color.A == 0 {
		return
	}
	if done := c.startFrameSection(PerfDraw, 1); done != nil {
		defer done()
	}
	diameter := c.Gtx.Dp(unit.Dp(diameterDp))
	if diameter <= 0 {
		diameter = c.Gtx.Dp(unit.Dp(40))
	}
	if diameter <= 0 {
		diameter = 40
	}

	origin := image.Point{
		X: center.X - diameter/2,
		Y: center.Y - diameter/2,
	}
	offset := op.Offset(origin).Push(c.Gtx.Ops)
	clipStack := clip.Ellipse(image.Rectangle{Max: image.Pt(diameter, diameter)}).Push(c.Gtx.Ops)

	if stateOpacity > 0 {
		col := spec.Color
		if stateOpacity > 1 {
			stateOpacity = 1
		}
		col.A = uint8(float32(col.A)*stateOpacity + 0.5)
		if col.A > 0 {
			paint.Fill(c.Gtx.Ops, col)
		}
	}

	if clickable != nil {
		opacity := spec.Opacity
		if opacity <= 0 {
			opacity = 0.16
		}
		localCenter := image.Pt(diameter/2, diameter/2)
		bounds := image.Pt(diameter, diameter)
		for _, press := range clickable.History() {
			c.drawRipplePress(press, bounds, localCenter, spec.Color, opacity)
		}
	}

	clipStack.Pop()
	offset.Pop()
}

// LayoutRippleArea registers a click target and draws bounded ripple feedback
// behind the child without changing the child's measured size.
func (c *Context) LayoutRippleArea(clickable *ClickableState, spec RippleSpec, child func(*Context) image.Point) image.Point {
	c.recordFrameSection(PerfLayout, 1)
	if child == nil {
		return image.Point{}
	}
	originalConstraints := c.Gtx.Constraints
	inputGtx := c.Gtx
	inputGtx.Constraints.Min.Y = 0
	if clickable == nil {
		return originalConstraints.Constrain(child(c.sameScope(inputGtx)))
	}
	clickable.BindRuntime(c.Runtime())

	inputDone := c.startFrameSection(PerfInput, 1)
	popPass := c.pushPointerPassThrough(c.Gtx.Ops)
	dims := clickable.raw().Layout(inputGtx, func(gtx gioLayout.Context) gioLayout.Dimensions {
		next := c.sameScope(gtx)
		recorded := op.Record(gtx.Ops)
		size := child(next)
		call := recorded.Stop()
		next.DrawRipple(clickable, size, spec)
		call.Add(gtx.Ops)
		return gioLayout.Dimensions{Size: size}
	})
	if popPass != nil {
		popPass()
	}
	if inputDone != nil {
		inputDone()
	}
	return originalConstraints.Constrain(dims.Size)
}

// LayoutRippleOverlayArea registers a click target and draws bounded ripple
// feedback above the child. This is useful for surfaces whose child draws an
// opaque container background before its content.
func (c *Context) LayoutRippleOverlayArea(clickable *ClickableState, spec RippleSpec, child func(*Context) image.Point) image.Point {
	c.recordFrameSection(PerfLayout, 1)
	if child == nil {
		return image.Point{}
	}
	originalConstraints := c.Gtx.Constraints
	inputGtx := c.Gtx
	inputGtx.Constraints.Min.Y = 0
	if clickable == nil {
		return originalConstraints.Constrain(child(c.sameScope(inputGtx)))
	}
	clickable.BindRuntime(c.Runtime())

	inputDone := c.startFrameSection(PerfInput, 1)
	popPass := c.pushPointerPassThrough(c.Gtx.Ops)
	dims := clickable.raw().Layout(inputGtx, func(gtx gioLayout.Context) gioLayout.Dimensions {
		next := c.sameScope(gtx)
		recorded := op.Record(gtx.Ops)
		size := child(next)
		call := recorded.Stop()
		call.Add(gtx.Ops)
		next.DrawRipple(clickable, size, spec)
		return gioLayout.Dimensions{Size: size}
	})
	if popPass != nil {
		popPass()
	}
	if inputDone != nil {
		inputDone()
	}
	return originalConstraints.Constrain(dims.Size)
}

func (c *Context) drawRipplePress(press widget.Press, bounds, center image.Point, col color.NRGBA, opacity float32) {
	frame := calculateRippleFrame(press, c.Gtx.Now, bounds)
	if frame.Expired {
		return
	}
	if frame.Invalidate {
		c.recordFrameSection(PerfAnimation, 1)
		c.RequestFrameRedrawReason("animation.ripple")
	}
	if frame.Diameter <= 0 {
		return
	}

	col.A = uint8(float32(col.A)*opacity*frame.AlphaScale + 0.5)
	if col.A == 0 {
		return
	}

	radius := frame.Diameter / 2
	offset := op.Offset(center.Add(image.Point{X: -radius, Y: -radius})).Push(c.Gtx.Ops)
	paint.FillShape(c.Gtx.Ops, col, clip.Ellipse(image.Rectangle{Max: image.Pt(frame.Diameter, frame.Diameter)}).Op(c.Gtx.Ops))
	offset.Pop()
}

func calculateRippleFrame(press widget.Press, now time.Time, bounds image.Point) rippleFrame {
	if bounds.X <= 0 || bounds.Y <= 0 || now.Before(press.Start) {
		return rippleFrame{Expired: true}
	}

	t := float32(now.Sub(press.Start).Seconds())
	end := press.End
	if end.IsZero() {
		end = now
	}
	endt := float32(end.Sub(press.Start).Seconds())

	var alphat float32
	var haste float32
	fadeSeconds := float32(rippleFadeDuration.Seconds())
	if press.Cancelled {
		if h := 0.5 - endt/fadeSeconds; h > 0 {
			haste = h
		}
	}
	half1 := t/fadeSeconds + haste
	if half1 > 0.5 {
		half1 = 0.5
	}
	half2 := float32(now.Sub(end).Seconds())/fadeSeconds + haste
	if half2 > 0.5 {
		return rippleFrame{Expired: true}
	}
	alphat = half1 + half2

	sizet := t
	if press.Cancelled {
		sizet = endt
	}
	sizet /= float32(rippleExpandDuration.Seconds())
	invalidate := !press.End.IsZero() || sizet <= 1
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
		return rippleFrame{
			AlphaScale: alphaBezier,
			Invalidate: invalidate,
		}
	}

	return rippleFrame{
		Diameter:   diameter,
		AlphaScale: alphaBezier,
		Invalidate: invalidate,
	}
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
