package widget

import (
	"image"
	"image/color"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/style"

	"gioui.org/op"
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

func clampRRectRadiusPx(size image.Point, rr int) int {
	if rr <= 0 || size.X <= 0 || size.Y <= 0 {
		return 0
	}
	limit := size.X
	if size.Y < limit {
		limit = size.Y
	}
	limit /= 2
	if limit < 0 {
		return 0
	}
	if rr > limit {
		return limit
	}
	return rr
}

func toInternalInsets(insets style.Insets) internal.Insets {
	return internal.Insets{
		Top:    insets.Top,
		Right:  insets.Right,
		Bottom: insets.Bottom,
		Left:   insets.Left,
	}
}

func resolveDecorationState(d style.Decoration, hovered, pressed, disabled bool) style.Decoration {
	active := d
	if disabled {
		if d.Disabled != nil {
			active = d.Merge(*d.Disabled)
		}
	} else if pressed {
		if d.Pressed != nil {
			active = d.Merge(*d.Pressed)
		}
	} else if hovered {
		if d.Hover != nil {
			active = d.Merge(*d.Hover)
		}
	}
	return active
}

func stripStateDecoration(d style.Decoration) style.Decoration {
	d.Hover = nil
	d.Pressed = nil
	d.Focused = nil
	d.Disabled = nil
	return d
}

func componentDecoration(d style.Decoration, defaultBg color.NRGBA, defaultPad style.Insets, defaultRad float32) style.Decoration {
	base := style.Decoration{}.WithBg(defaultBg).WithPad(defaultPad).WithRad(defaultRad)
	return stripStateDecoration(base.Merge(d))
}

func withDefaultStates(d style.Decoration, hover, pressed, disabled style.Decoration) style.Decoration {
	if d.Hover == nil {
		d.Hover = &hover
	}
	if d.Pressed == nil {
		d.Pressed = &pressed
	}
	if d.Disabled == nil {
		d.Disabled = &disabled
	}
	return d
}

func densityMetric(ctx *internal.Context, defaultValue, compactValue float32) float32 {
	if ctx == nil {
		return defaultValue
	}
	return ctx.Theme().Density.Metric(defaultValue, compactValue)
}

func densityHeight(ctx *internal.Context, defaultHeight, compactHeight float32) float32 {
	if ctx == nil {
		return defaultHeight
	}
	return ctx.Theme().Density.ComponentHeight(defaultHeight, compactHeight)
}

func densityInsets(ctx *internal.Context, defaultInsets, compactInsets style.Insets) style.Insets {
	if ctx == nil || !ctx.Theme().Density.IsCompact() {
		return defaultInsets
	}
	return compactInsets
}

func withAlpha(col color.NRGBA, alpha uint8) color.NRGBA {
	col.A = alpha
	return col
}

func colorOr(col, fallback color.NRGBA) color.NRGBA {
	if col.A == 0 {
		return fallback
	}
	return col
}

func hasDecorationVisual(d style.Decoration) bool {
	return d.Background != nil || d.Gradient != nil || d.Padding != nil || d.Margin != nil ||
		d.Radius != nil || d.Border != nil || d.Opacity != nil || d.CircleClip ||
		d.Shadow != nil || d.Image != nil || d.Transform != nil
}

func hasDecorationState(d style.Decoration) bool {
	return d.Hover != nil || d.Pressed != nil || d.Focused != nil || d.Disabled != nil
}

func hasAnyDecoration(d style.Decoration) bool {
	return hasDecorationVisual(d) || hasDecorationState(d)
}

func layoutDecoratedClickTarget(ctx *internal.Context, clickable *internal.ClickableState, hovered, pressed bool, d style.Decoration, disabled bool, child func(*internal.Context) image.Point) layout.Dimensions {
	active := resolveDecorationState(d, hovered, pressed, disabled)
	visualDecoration := stripStateDecoration(active)

	macro := op.Record(ctx.Gtx.Ops)
	dims := layoutDecorationShell(ctx.Child(0), visualDecoration, child)
	call := macro.Stop()

	if clickable != nil && !disabled {
		layoutTransformedClickArea(ctx, clickable, visualDecoration.ResolveMargin(style.Insets{}), visualDecoration.ResolveTransform(), dims.Size)
	}
	call.Add(ctx.Gtx.Ops)

	return dims
}

func layoutRippleTouchTarget(ctx *internal.Context, clickable *internal.ClickableState, disabled bool, spec internal.RippleSpec, minTargetDp float32, child func(*internal.Context) image.Point) layout.Dimensions {
	if child == nil {
		return layout.Dimensions{}
	}

	macro := op.Record(ctx.Gtx.Ops)
	childSize := child(ctx.Child(0))
	childCall := macro.Stop()

	minTarget := ctx.Gtx.Dp(unit.Dp(minTargetDp))
	targetSize := childSize
	if targetSize.X < minTarget {
		targetSize.X = minTarget
	}
	if targetSize.Y < minTarget {
		targetSize.Y = minTarget
	}
	if max := ctx.Gtx.Constraints.Max; max.X > 0 || max.Y > 0 {
		if max.X > 0 && targetSize.X > max.X {
			targetSize.X = max.X
		}
		if max.Y > 0 && targetSize.Y > max.Y {
			targetSize.Y = max.Y
		}
	}

	if disabled || clickable == nil {
		childCall.Add(ctx.Gtx.Ops)
		return layout.Dimensions{Size: childSize}
	}

	offset := image.Point{
		X: -(targetSize.X - childSize.X) / 2,
		Y: -(targetSize.Y - childSize.Y) / 2,
	}
	rippleMacro := op.Record(ctx.Gtx.Ops)
	stack := op.Offset(offset).Push(ctx.Gtx.Ops)
	ctx.LayoutRippleArea(clickable, spec, func(*internal.Context) image.Point {
		return targetSize
	})
	stack.Pop()
	rippleCall := rippleMacro.Stop()

	rippleCall.Add(ctx.Gtx.Ops)
	childCall.Add(ctx.Gtx.Ops)
	return layout.Dimensions{Size: childSize}
}

func layoutStateLayerTouchTarget(ctx *internal.Context, clickable *internal.ClickableState, disabled bool, spec internal.RippleSpec, minTargetDp, layerDiameterDp float32, layerCenter func(image.Point) image.Point, stateOpacity func() float32, child func(*internal.Context) image.Point) layout.Dimensions {
	if child == nil {
		return layout.Dimensions{}
	}

	macro := op.Record(ctx.Gtx.Ops)
	childSize := child(ctx.Child(0))
	childCall := macro.Stop()

	minTarget := ctx.Gtx.Dp(unit.Dp(minTargetDp))
	targetSize := childSize
	if targetSize.X < minTarget {
		targetSize.X = minTarget
	}
	if targetSize.Y < minTarget {
		targetSize.Y = minTarget
	}
	if max := ctx.Gtx.Constraints.Max; max.X > 0 || max.Y > 0 {
		if max.X > 0 && targetSize.X > max.X {
			targetSize.X = max.X
		}
		if max.Y > 0 && targetSize.Y > max.Y {
			targetSize.Y = max.Y
		}
	}

	if disabled || clickable == nil {
		childCall.Add(ctx.Gtx.Ops)
		return layout.Dimensions{Size: childSize}
	}

	targetOffset := image.Point{
		X: -(targetSize.X - childSize.X) / 2,
		Y: -(targetSize.Y - childSize.Y) / 2,
	}
	targetMacro := op.Record(ctx.Gtx.Ops)
	stack := op.Offset(targetOffset).Push(ctx.Gtx.Ops)
	ctx.LayoutClickArea(clickable, func(*internal.Context) image.Point {
		return targetSize
	})
	stack.Pop()
	targetCall := targetMacro.Stop()

	center := image.Point{}
	if layerCenter != nil {
		center = layerCenter(childSize)
	}
	opacity := float32(0)
	if stateOpacity != nil {
		opacity = stateOpacity()
	}
	layerMacro := op.Record(ctx.Gtx.Ops)
	ctx.DrawStateLayerCircle(clickable, center, layerDiameterDp, spec, opacity)
	layerCall := layerMacro.Stop()

	targetCall.Add(ctx.Gtx.Ops)
	layerCall.Add(ctx.Gtx.Ops)
	childCall.Add(ctx.Gtx.Ops)
	return layout.Dimensions{Size: childSize}
}

func materialStateLayerOpacity(hovered, pressed bool) float32 {
	if pressed {
		return style.StateLayerPressedOpacity
	}
	if hovered {
		return style.StateLayerHoverOpacity
	}
	return 0
}

func selectionControlSizePx(ctx *internal.Context, sizeDp, fallbackDp float32) int {
	if fallbackDp <= 0 {
		fallbackDp = 18
	}
	size := ctx.Gtx.Dp(unit.Dp(sizeDp))
	if size <= 0 {
		size = ctx.Gtx.Dp(unit.Dp(fallbackDp))
	}
	if size < 14 {
		size = 14
	}
	return size
}

func layoutTransformedClickArea(ctx *internal.Context, clickable *internal.ClickableState, margin style.Insets, transform *style.Transform2D, outerSize image.Point) image.Point {
	if clickable == nil {
		return outerSize
	}

	innerSize := insetContentSize(ctx, outerSize, margin)
	return ctx.LayoutInset(toInternalInsets(margin), func(marginCtx *internal.Context) image.Point {
		if transform == nil {
			return marginCtx.LayoutClickArea(clickable, func(*internal.Context) image.Point {
				return innerSize
			})
		}

		stack := op.Affine(decorationTransformMatrix(marginCtx.Gtx, *transform, innerSize)).Push(marginCtx.Gtx.Ops)
		size := marginCtx.LayoutClickArea(clickable, func(*internal.Context) image.Point {
			return innerSize
		})
		stack.Pop()
		return size
	})
}

func insetContentSize(ctx *internal.Context, outerSize image.Point, margin style.Insets) image.Point {
	if ctx == nil {
		return outerSize
	}
	left := ctx.Gtx.Dp(unit.Dp(margin.Left))
	right := ctx.Gtx.Dp(unit.Dp(margin.Right))
	top := ctx.Gtx.Dp(unit.Dp(margin.Top))
	bottom := ctx.Gtx.Dp(unit.Dp(margin.Bottom))

	inner := image.Point{
		X: outerSize.X - left - right,
		Y: outerSize.Y - top - bottom,
	}
	if inner.X < 0 {
		inner.X = 0
	}
	if inner.Y < 0 {
		inner.Y = 0
	}
	return inner
}

func layoutDecorationShell(ctx *internal.Context, d style.Decoration, child func(*internal.Context) image.Point) layout.Dimensions {
	if !hasDecorationVisual(d) {
		if child == nil {
			return layout.Dimensions{}
		}
		return layout.Dimensions{Size: child(ctx.Child(0))}
	}

	spec := decorationSurfaceSpec(d, ctx, color.NRGBA{})
	margin := d.ResolveMargin(style.Insets{})
	size := ctx.LayoutInset(toInternalInsets(margin), func(marginCtx *internal.Context) image.Point {
		return layoutSurfaceWithTransform(marginCtx, spec, d.ResolveTransform(), func(contentCtx *internal.Context) image.Point {
			if child == nil {
				return image.Point{}
			}
			return child(contentCtx.Child(0))
		})
	})
	return layout.Dimensions{Size: size}
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
