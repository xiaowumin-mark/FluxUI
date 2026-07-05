package widget

import (
	"image"
	"image/color"
	"time"

	"github.com/xiaowumin-mark/FluxUI/anim"
	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/style"
	"github.com/xiaowumin-mark/FluxUI/theme"

	"gioui.org/f32"
	gioEvent "gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

func safeDp(v float32) unit.Dp {
	if v < 0 {
		return 0
	}
	return unit.Dp(v)
}

func redrawInvalidator(ctx *internal.Context) func() {
	if ctx == nil || ctx.Runtime() == nil {
		return nil
	}
	return func() {
		ctx.Runtime().RequestRedrawReason("widget.redraw")
	}
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

func md3DismissOnOutsidePress(ctx *internal.Context, tag any, protected []image.Rectangle, onDismiss func(*internal.Context)) {
	if ctx == nil || tag == nil || onDismiss == nil {
		return
	}
	viewport, ok := ctx.Viewport()
	if !ok || viewport.Dx() <= 0 || viewport.Dy() <= 0 {
		viewport = image.Rectangle{Max: ctx.Gtx.Constraints.Max}
	}
	if viewport.Dx() <= 0 || viewport.Dy() <= 0 {
		return
	}
	position := ctx.Position()
	pass := pointer.PassOp{}.Push(ctx.Gtx.Ops)
	offset := op.Offset(image.Pt(-position.X, -position.Y)).Push(ctx.Gtx.Ops)
	clipStack := clip.Rect(viewport).Push(ctx.Gtx.Ops)
	gioEvent.Op(ctx.Gtx.Ops, tag)
	clipStack.Pop()
	offset.Pop()
	pass.Pop()

	for {
		ev, ok := ctx.Gtx.Event(pointer.Filter{Target: tag, Kinds: pointer.Press})
		if !ok {
			break
		}
		pe, ok := ev.(pointer.Event)
		if !ok {
			continue
		}
		point := image.Pt(int(pe.Position.X+0.5), int(pe.Position.Y+0.5))
		inside := false
		for _, rect := range protected {
			if point.In(rect) {
				inside = true
				break
			}
		}
		if !inside {
			pointerEvent := fluxevent.PointerEventFromGio(ctx, fluxevent.PointerDown, pe)
			pointerEvent.Position = f32.Pt(float32(point.X), float32(point.Y))
			if fluxevent.DispatchPointerEvent(ctx, ctx.PathID(), &pointerEvent) {
				onDismiss(ctx)
			}
		}
	}
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
		d.Corner != nil || d.Shadow != nil || d.Image != nil || d.Transform != nil
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
	duration, easing := md3InteractionTiming(ctx, hovered, pressed, false, disabled)
	visualDecoration = md3AnimateDecoration(ctx, "decorated-click-target-decoration", visualDecoration, duration, easing)

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
	opacity = md3AnimateStateLayerOpacity(ctx, "touch-state-layer", opacity)
	layerMacro := op.Record(ctx.Gtx.Ops)
	ctx.DrawStateLayerCircle(clickable, center, layerDiameterDp, spec, opacity)
	layerCall := layerMacro.Stop()

	childCall.Add(ctx.Gtx.Ops)
	targetCall.Add(ctx.Gtx.Ops)
	layerCall.Add(ctx.Gtx.Ops)
	return layout.Dimensions{Size: childSize}
}

func materialStateLayerOpacity(ctx *internal.Context, hovered, pressed bool) float32 {
	if pressed {
		return style.StateLayerPressedOpacity
	}
	if hovered && interactionQuality(ctx) != theme.InteractionQualityLowCPU {
		return style.StateLayerHoverOpacity
	}
	return 0
}

func materialAnimatedStateLayerOpacity(ctx *internal.Context, hovered, pressed, disabled bool) float32 {
	if disabled {
		return 0
	}
	return md3AnimateStateLayerOpacity(ctx, "md3-state-layer", materialStateLayerOpacity(ctx, hovered, pressed))
}

func md3AnimateStateLayerOpacity(ctx *internal.Context, namespace string, target float32) float32 {
	if ctx == nil {
		return target
	}
	key := md3MotionKey(ctx, namespace)
	if target > 0 && target <= style.StateLayerHoverOpacity+0.0001 {
		switch interactionQuality(ctx) {
		case theme.InteractionQualityLowCPU:
			ctx.ForgetPersistentKey(key)
			return 0
		case theme.InteractionQualityBalanced:
			ctx.ForgetPersistentKey(key)
			return target
		}
	}
	enterDuration := style.InteractionHoverEnterDuration
	if target >= style.StateLayerPressedOpacity {
		enterDuration = style.InteractionPressedEnterDuration
	}
	if target <= 0 {
		state, ok := md3ExistingFloatState(ctx, key)
		if !ok {
			return 0
		}
		if interactionQuality(ctx) != theme.InteractionQualityFull && state.to <= style.StateLayerHoverOpacity+0.0001 {
			ctx.ForgetPersistentKey(key)
			return 0
		}
		progress := state.advance(ctx, 0, style.InteractionPressedExitDuration, style.InteractionStandardAccelerateEasing)
		if progress <= 0 {
			if _, running := state.currentAndRunning(ctx.Now()); !running && state.to <= 0 {
				ctx.ForgetPersistentKey(key)
			}
		}
		return progress
	}
	state := md3FloatStateForKey(ctx, key, 0)
	current := state.valueAt(ctx.Now())
	duration := style.InteractionPressedExitDuration
	easing := style.InteractionStandardAccelerateEasing
	if target > current {
		duration = enterDuration
		easing = style.InteractionStandardDecelerateEasing
	}
	return state.advance(ctx, target, duration, easing)
}

func md3SelectionProgress(ctx *internal.Context, active bool) float32 {
	target := float32(0)
	if active {
		target = 1
	}
	return md3AnimateFloat(ctx, "md3-selection-progress", target, style.InteractionSelectedDuration, style.InteractionStandardEasing)
}

func md3SwitchSelectionProgress(ctx *internal.Context, active bool) float32 {
	target := float32(0)
	if active {
		target = 1
	}
	return md3AnimateFloatDirectionalRetained(
		ctx,
		"md3-switch-position",
		target,
		250*time.Millisecond,
		250*time.Millisecond,
		md3SwitchBackEasing,
		md3SwitchBackEasing,
	)
}

func md3SwitchPressedProgress(ctx *internal.Context, pressed bool) float32 {
	target := float32(0)
	if pressed {
		target = 1
	}
	return md3AnimateFloatDirectionalRetained(
		ctx,
		"md3-switch-press",
		target,
		100*time.Millisecond,
		250*time.Millisecond,
		style.InteractionLinearEasing,
		style.InteractionStandardEasing,
	)
}

func md3SwitchBackEasing(v float32) float32 {
	v = clampFloat32(v, 0, 1)
	base := style.InteractionEmphasizedDecelerateEasing(v)
	overshoot := float32(0.10) * 4 * v * (1 - v) * v
	if v > 0.82 {
		return base + overshoot
	}
	return base
}

func switchThumbCenter(ctx *internal.Context, size image.Point, progress float32) image.Point {
	if ctx == nil {
		return image.Pt(size.X/2, size.Y/2)
	}
	start := ctx.Gtx.Dp(safeDp(16))
	end := size.X - start
	if end < start {
		end = start
	}
	progress = clampFloat32(progress, -0.15, 1.15)
	x := start + int(float32(end-start)*progress+0.5)
	return image.Pt(x, size.Y/2)
}

func md3AnimateFloatDirectionalRetained(ctx *internal.Context, namespace string, target float32, enterDuration, exitDuration time.Duration, enterEasing, exitEasing func(float32) float32) float32 {
	if ctx == nil {
		return target
	}
	key := md3MotionKey(ctx, namespace)
	state := md3FloatStateForKey(ctx, key, target)
	now := ctx.Now()
	current, running := state.currentAndRunning(now)
	duration := exitDuration
	easing := exitEasing
	if target > current {
		duration = enterDuration
		easing = enterEasing
	}
	if state.to == target {
		state.current = current
		state.easing = easing
		if running {
			if rt := ctx.Runtime(); rt != nil {
				rt.RecordFrameSection(internal.PerfAnimation, 1)
			}
			ctx.RequestFrameRedrawReason("animation.running")
			return current
		}
		state.snap(now, target, duration, easing)
		return target
	}
	if current == target {
		state.snap(now, target, duration, easing)
		return target
	}
	return state.advance(ctx, target, duration, easing)
}

func md3FocusProgress(ctx *internal.Context, focused, disabled bool) float32 {
	target := float32(0)
	if focused && !disabled {
		target = 1
	}
	return md3AnimateFloatDirectional(
		ctx,
		"md3-focus-progress",
		target,
		style.InteractionFocusEnterDuration,
		style.InteractionFocusExitDuration,
		style.InteractionStandardDecelerateEasing,
		style.InteractionStandardAccelerateEasing,
	)
}

func md3DrawFocusIndicator(ctx *internal.Context, size image.Point, spec internal.FocusIndicatorSpec, focused, disabled bool) {
	opacity := md3FocusProgress(ctx, focused, disabled)
	if opacity <= 0 {
		return
	}
	spec.Color = withAlphaFactor(spec.Color, opacity)
	ctx.DrawFocusIndicator(size, spec)
}

func md3InteractionTiming(ctx *internal.Context, hovered, pressed, focused, disabled bool) (time.Duration, func(float32) float32) {
	quality := interactionQuality(ctx)
	switch {
	case disabled:
		return style.InteractionSelectedDuration, style.InteractionStandardEasing
	case pressed:
		return style.InteractionPressedEnterDuration, style.InteractionStandardDecelerateEasing
	case hovered:
		if quality != theme.InteractionQualityFull {
			if focused {
				return style.InteractionFocusEnterDuration, style.InteractionStandardDecelerateEasing
			}
			return 0, style.InteractionLinearEasing
		}
		return style.InteractionHoverEnterDuration, style.InteractionStandardDecelerateEasing
	case focused:
		return style.InteractionFocusEnterDuration, style.InteractionStandardDecelerateEasing
	default:
		if quality != theme.InteractionQualityFull {
			return 0, style.InteractionLinearEasing
		}
		return style.InteractionHoverExitDuration, style.InteractionStandardAccelerateEasing
	}
}

func interactionQuality(ctx *internal.Context) theme.InteractionQuality {
	if ctx == nil {
		return theme.InteractionQualityFull
	}
	return ctx.InteractionQuality()
}

type md3AnimatedFloatState struct {
	startedAt time.Time
	from      float32
	current   float32
	to        float32
	duration  time.Duration
	easing    func(float32) float32
}

func md3AnimateFloat(ctx *internal.Context, namespace string, target float32, duration time.Duration, easing func(float32) float32) float32 {
	if ctx == nil {
		return target
	}
	if target <= 0 {
		key := md3MotionKey(ctx, namespace)
		state, ok := md3ExistingFloatState(ctx, key)
		if !ok {
			return 0
		}
		progress := state.advance(ctx, 0, duration, easing)
		if progress <= 0 {
			forgetIdleFloatState(ctx, key, state)
		}
		return progress
	}
	state := md3FloatStateFor(ctx, namespace, target)
	return state.advance(ctx, target, duration, easing)
}

func md3AnimateFloatDirectional(ctx *internal.Context, namespace string, target float32, enterDuration, exitDuration time.Duration, enterEasing, exitEasing func(float32) float32) float32 {
	if ctx == nil {
		return target
	}
	if target <= 0 {
		key := md3MotionKey(ctx, namespace)
		state, ok := md3ExistingFloatState(ctx, key)
		if !ok {
			return 0
		}
		progress := state.advance(ctx, 0, exitDuration, exitEasing)
		if progress <= 0 {
			forgetIdleFloatState(ctx, key, state)
		}
		return progress
	}
	state := md3FloatStateFor(ctx, namespace, target)
	current := state.valueAt(ctx.Now())
	duration := exitDuration
	easing := exitEasing
	if target > current {
		duration = enterDuration
		easing = enterEasing
	}
	return state.advance(ctx, target, duration, easing)
}

func forgetIdleFloatState(ctx *internal.Context, key internal.MemoryKey, state *md3AnimatedFloatState) {
	if ctx == nil || state == nil {
		return
	}
	if _, running := state.currentAndRunning(ctx.Now()); !running && state.to <= 0 {
		ctx.ForgetPersistentKey(key)
	}
}

func md3FloatStateFor(ctx *internal.Context, namespace string, initial float32) *md3AnimatedFloatState {
	return md3FloatStateForKey(ctx, md3MotionKey(ctx, namespace), initial)
}

func md3FloatStateForKey(ctx *internal.Context, key internal.MemoryKey, initial float32) *md3AnimatedFloatState {
	value := ctx.PersistentKey(key, func() any {
		return &md3AnimatedFloatState{
			startedAt: ctx.Now(),
			from:      initial,
			current:   initial,
			to:        initial,
			easing:    style.InteractionLinearEasing,
		}
	})
	state, ok := value.(*md3AnimatedFloatState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUI/widget: md3 animated float state type mismatch")
	}
	if state.easing == nil {
		state.easing = style.InteractionLinearEasing
	}
	return state
}

func md3ExistingFloatState(ctx *internal.Context, key internal.MemoryKey) (*md3AnimatedFloatState, bool) {
	if ctx == nil {
		return nil, false
	}
	value, ok := ctx.PersistentValueKey(key)
	if !ok {
		return nil, false
	}
	state, ok := value.(*md3AnimatedFloatState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUI/widget: md3 animated float state type mismatch")
	}
	if state.easing == nil {
		state.easing = style.InteractionLinearEasing
	}
	return state, true
}

func (s *md3AnimatedFloatState) advance(ctx *internal.Context, target float32, duration time.Duration, easing func(float32) float32) float32 {
	if ctx == nil {
		return target
	}
	if rt := ctx.Runtime(); rt != nil {
		rt.RecordFrameSection(internal.PerfAnimation, 1)
	}
	if easing == nil {
		easing = style.InteractionLinearEasing
	}
	if duration <= 0 {
		s.snap(ctx.Now(), target, duration, easing)
		return target
	}
	if s.to == target && s.duration == duration {
		current, running := s.currentAndRunning(ctx.Now())
		s.current = current
		s.easing = easing
		if running {
			ctx.RequestFrameRedrawReason("animation.running")
			return current
		}
		s.snap(ctx.Now(), target, duration, easing)
		return target
	}

	now := ctx.Now()
	current := s.valueAt(now)
	if current == target {
		s.snap(now, target, duration, easing)
		return target
	}
	s.startedAt = now
	s.from = current
	s.current = current
	s.to = target
	s.duration = duration
	s.easing = easing
	ctx.RequestFrameRedrawReason("animation.running")
	return current
}

func (s *md3AnimatedFloatState) snap(now time.Time, target float32, duration time.Duration, easing func(float32) float32) {
	s.startedAt = now
	s.from = target
	s.current = target
	s.to = target
	s.duration = duration
	s.easing = easing
}

func (s *md3AnimatedFloatState) valueAt(now time.Time) float32 {
	current, _ := s.currentAndRunning(now)
	return current
}

func (s *md3AnimatedFloatState) currentAndRunning(now time.Time) (float32, bool) {
	if s.duration <= 0 {
		return s.to, false
	}
	if s.from == s.to {
		return s.to, false
	}
	elapsed := now.Sub(s.startedAt)
	if elapsed <= 0 {
		return s.from, true
	}
	if elapsed >= s.duration {
		return s.to, false
	}
	p := float32(elapsed) / float32(s.duration)
	if s.easing != nil {
		p = s.easing(p)
	}
	return s.from + (s.to-s.from)*p, true
}

type md3AnimatedColorState struct {
	startedAt time.Time
	from      color.NRGBA
	current   color.NRGBA
	to        color.NRGBA
	duration  time.Duration
	easing    func(float32) float32
}

func md3AnimateColor(ctx *internal.Context, namespace string, target color.NRGBA, duration time.Duration, easing func(float32) float32) color.NRGBA {
	if ctx == nil || duration <= 0 {
		return target
	}
	key := md3MotionKey(ctx, namespace)
	state := md3ColorStateForKey(ctx, key, target, duration, easing)
	return md3AdvanceColorState(ctx, key, state, target, duration, easing)
}

func md3AnimateColorIfActive(ctx *internal.Context, namespace string, target color.NRGBA, active bool, duration time.Duration, easing func(float32) float32) color.NRGBA {
	if ctx == nil || duration <= 0 {
		return target
	}
	key := md3MotionKey(ctx, namespace)
	if !active {
		state, ok := md3ExistingColorState(ctx, key)
		if !ok {
			return target
		}
		return md3AdvanceColorState(ctx, key, state, target, duration, easing)
	}
	state := md3ColorStateForKey(ctx, key, target, duration, easing)
	return md3AdvanceColorState(ctx, key, state, target, duration, easing)
}

func md3AnimateColorRetained(ctx *internal.Context, namespace string, target color.NRGBA, duration time.Duration, easing func(float32) float32) color.NRGBA {
	if ctx == nil || duration <= 0 {
		return target
	}
	if easing == nil {
		easing = style.InteractionLinearEasing
	}
	key := md3MotionKey(ctx, namespace)
	state := md3ColorStateForKeyRetained(ctx, key, target, duration, easing)
	now := ctx.Now()
	if state.to == target && state.duration == duration {
		current, running := state.currentAndRunning(now)
		state.current = current
		state.easing = easing
		if running {
			if rt := ctx.Runtime(); rt != nil {
				rt.RecordFrameSection(internal.PerfAnimation, 1)
			}
			ctx.RequestFrameRedrawReason("animation.running")
			return current
		}
		state.snap(now, target, duration, easing)
		return target
	}
	current := state.valueAt(now)
	if current == target {
		state.snap(now, target, duration, easing)
		return target
	}
	if rt := ctx.Runtime(); rt != nil {
		rt.RecordFrameSection(internal.PerfAnimation, 1)
	}
	return md3AdvanceColorState(ctx, key, state, target, duration, easing)
}

func md3ColorStateForKeyRetained(ctx *internal.Context, key internal.MemoryKey, target color.NRGBA, duration time.Duration, easing func(float32) float32) *md3AnimatedColorState {
	if easing == nil {
		easing = style.InteractionLinearEasing
	}
	value := ctx.PersistentKey(key, func() any {
		return &md3AnimatedColorState{
			startedAt: ctx.Now(),
			from:      target,
			current:   target,
			to:        target,
			duration:  duration,
			easing:    easing,
		}
	})
	state, ok := value.(*md3AnimatedColorState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUI/widget: md3 animated color state type mismatch")
	}
	if state.easing == nil {
		state.easing = style.InteractionLinearEasing
	}
	return state
}

func md3ColorStateForKey(ctx *internal.Context, key internal.MemoryKey, target color.NRGBA, duration time.Duration, easing func(float32) float32) *md3AnimatedColorState {
	if rt := ctx.Runtime(); rt != nil {
		rt.RecordFrameSection(internal.PerfAnimation, 1)
	}
	if easing == nil {
		easing = style.InteractionLinearEasing
	}
	value := ctx.PersistentKey(key, func() any {
		return &md3AnimatedColorState{
			startedAt: ctx.Now(),
			from:      target,
			current:   target,
			to:        target,
			duration:  duration,
			easing:    easing,
		}
	})
	state, ok := value.(*md3AnimatedColorState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUI/widget: md3 animated color state type mismatch")
	}
	if state.easing == nil {
		state.easing = style.InteractionLinearEasing
	}
	return state
}

func md3ExistingColorState(ctx *internal.Context, key internal.MemoryKey) (*md3AnimatedColorState, bool) {
	if ctx == nil {
		return nil, false
	}
	value, ok := ctx.PersistentValueKey(key)
	if !ok {
		return nil, false
	}
	state, ok := value.(*md3AnimatedColorState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUI/widget: md3 animated color state type mismatch")
	}
	if state.easing == nil {
		state.easing = style.InteractionLinearEasing
	}
	return state, true
}

func md3AdvanceColorState(ctx *internal.Context, key internal.MemoryKey, state *md3AnimatedColorState, target color.NRGBA, duration time.Duration, easing func(float32) float32) color.NRGBA {
	if ctx == nil || state == nil {
		return target
	}
	if easing == nil {
		easing = style.InteractionLinearEasing
	}
	if state.to == target && state.duration == duration {
		current, running := state.currentAndRunning(ctx.Now())
		state.current = current
		state.easing = easing
		if running {
			ctx.RequestFrameRedrawReason("animation.running")
			return current
		}
		state.snap(ctx.Now(), target, duration, easing)
		forgetIdleColorState(ctx, key, state)
		return target
	}

	now := ctx.Now()
	current := state.valueAt(now)
	if current == target {
		state.snap(now, target, duration, easing)
		forgetIdleColorState(ctx, key, state)
		return target
	}
	state.startedAt = now
	state.from = current
	state.current = current
	state.to = target
	state.duration = duration
	state.easing = easing
	ctx.RequestFrameRedrawReason("animation.running")
	return current
}

func forgetIdleColorState(ctx *internal.Context, key internal.MemoryKey, state *md3AnimatedColorState) {
	if ctx == nil || state == nil {
		return
	}
	if state.from == state.to {
		ctx.ForgetPersistentKey(key)
		return
	}
	if _, running := state.currentAndRunning(ctx.Now()); !running {
		ctx.ForgetPersistentKey(key)
	}
}

func (s *md3AnimatedColorState) snap(now time.Time, target color.NRGBA, duration time.Duration, easing func(float32) float32) {
	s.startedAt = now
	s.from = target
	s.current = target
	s.to = target
	s.duration = duration
	s.easing = easing
}

func (s *md3AnimatedColorState) valueAt(now time.Time) color.NRGBA {
	current, _ := s.currentAndRunning(now)
	return current
}

func (s *md3AnimatedColorState) currentAndRunning(now time.Time) (color.NRGBA, bool) {
	if s.duration <= 0 {
		return s.to, false
	}
	if s.from == s.to {
		return s.to, false
	}
	elapsed := now.Sub(s.startedAt)
	if elapsed <= 0 {
		return s.from, true
	}
	if elapsed >= s.duration {
		return s.to, false
	}
	p := float32(elapsed) / float32(s.duration)
	if s.easing != nil {
		p = s.easing(p)
	}
	return lerpNRGBA(s.from, s.to, p), true
}

type md3AnimatedDecorationState struct {
	startedAt time.Time
	from      style.Decoration
	current   style.Decoration
	to        style.Decoration
	duration  time.Duration
	easing    func(float32) float32
}

func md3AnimateDecoration(ctx *internal.Context, namespace string, target style.Decoration, duration time.Duration, easing func(float32) float32) style.Decoration {
	if ctx == nil || duration <= 0 {
		return target
	}
	if rt := ctx.Runtime(); rt != nil {
		rt.RecordFrameSection(internal.PerfAnimation, 1)
	}
	if easing == nil {
		easing = style.InteractionLinearEasing
	}
	value := ctx.PersistentKey(md3MotionKey(ctx, namespace), func() any {
		return &md3AnimatedDecorationState{
			startedAt: ctx.Now(),
			from:      target,
			current:   target,
			to:        target,
			duration:  duration,
			easing:    easing,
		}
	})
	state, ok := value.(*md3AnimatedDecorationState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUI/widget: md3 animated decoration state type mismatch")
	}
	if anim.DecorationEqual(state.to, target) && state.duration == duration {
		current, running := state.currentAndRunning(ctx.Now())
		state.current = current
		state.easing = easing
		if running {
			ctx.RequestFrameRedrawReason("animation.running")
			return current
		}
		state.snap(ctx.Now(), target, duration, easing)
		return target
	}

	now := ctx.Now()
	current := state.valueAt(now)
	if anim.DecorationEqual(current, target) {
		state.snap(now, target, duration, easing)
		return target
	}
	state.startedAt = now
	state.from = current
	state.current = current
	state.to = target
	state.duration = duration
	state.easing = easing
	ctx.RequestFrameRedrawReason("animation.running")
	return current
}

func (s *md3AnimatedDecorationState) snap(now time.Time, target style.Decoration, duration time.Duration, easing func(float32) float32) {
	s.startedAt = now
	s.from = target
	s.current = target
	s.to = target
	s.duration = duration
	s.easing = easing
}

func (s *md3AnimatedDecorationState) valueAt(now time.Time) style.Decoration {
	current, _ := s.currentAndRunning(now)
	return current
}

func (s *md3AnimatedDecorationState) currentAndRunning(now time.Time) (style.Decoration, bool) {
	if s.duration <= 0 {
		return s.to, false
	}
	if anim.DecorationEqual(s.from, s.to) {
		return s.to, false
	}
	elapsed := now.Sub(s.startedAt)
	if elapsed <= 0 {
		return s.from, true
	}
	if elapsed >= s.duration {
		return s.to, false
	}
	p := float32(elapsed) / float32(s.duration)
	if s.easing != nil {
		p = s.easing(p)
	}
	return anim.LerpDecoration(s.from, s.to, p), true
}

func md3OverlayProgress(ctx *internal.Context, namespace string, visible bool, enterDuration, exitDuration time.Duration, enterEasing, exitEasing func(float32) float32) (progress float32, shouldRender bool) {
	if ctx == nil {
		return boolProgress(visible), visible
	}
	target := float32(0)
	if visible {
		target = 1
	}
	state := md3FloatStateFor(ctx, namespace, 0)
	current := state.valueAt(ctx.Now())
	duration := exitDuration
	easing := exitEasing
	if target > current {
		duration = enterDuration
		easing = enterEasing
	}
	progress = state.advance(ctx, target, duration, easing)
	return progress, visible || progress > 0.001
}

func layoutMD3OverlayTransition(ctx *internal.Context, progress float32, offsetDp float32, child func(*internal.Context) image.Point) image.Point {
	if child == nil || progress <= 0 {
		return image.Point{}
	}
	progress = clampFloat32(progress, 0, 1)
	if progress < 1 {
		defer paint.PushOpacity(ctx.Gtx.Ops, progress).Pop()
	}
	offset := int(float32(ctx.Gtx.Dp(unit.Dp(offsetDp))) * (1 - progress))
	stack := op.Offset(image.Point{Y: offset}).Push(ctx.Gtx.Ops)
	size := child(ctx.Child(0))
	stack.Pop()
	return size
}

func layoutMD3RevealTransition(ctx *internal.Context, progress float32, revealFromBottom bool, child func(*internal.Context) image.Point) image.Point {
	if child == nil || progress <= 0 {
		return image.Point{}
	}
	progress = clampFloat32(progress, 0, 1)
	recorded := op.Record(ctx.Gtx.Ops)
	size := child(ctx.Child(0))
	call := recorded.Stop()
	if size.X <= 0 || size.Y <= 0 {
		return size
	}
	if progress >= 0.999 {
		call.Add(ctx.Gtx.Ops)
		return size
	}
	revealHeight := int(float32(size.Y)*progress + 0.5)
	if revealHeight < 1 {
		revealHeight = 1
	}
	if revealHeight > size.Y {
		revealHeight = size.Y
	}
	clipRect := image.Rectangle{Max: image.Point{X: size.X, Y: revealHeight}}
	offsetY := 0
	if revealFromBottom {
		offsetY = size.Y - revealHeight
		clipRect = image.Rectangle{Min: image.Point{Y: offsetY}, Max: size}
	}
	if progress < 1 {
		defer paint.PushOpacity(ctx.Gtx.Ops, progress).Pop()
	}
	stack := clip.Rect(clipRect).Push(ctx.Gtx.Ops)
	call.Add(ctx.Gtx.Ops)
	stack.Pop()
	return size
}

type md3PopupDirection uint8

const (
	md3PopupDown md3PopupDirection = iota
	md3PopupUp
)

type md3PopupPlacement struct {
	Direction        md3PopupDirection
	GapPx            int
	MaxHeightPx      int
	OffsetX          int
	TransitionOffset float32
}

func md3PopupPlacementForAnchor(ctx *internal.Context, anchorSize, popupSize image.Point, preferredHeightPx, gapPx int) md3PopupPlacement {
	if gapPx < 0 {
		gapPx = 0
	}
	if preferredHeightPx <= 0 {
		preferredHeightPx = 1
	}

	placement := md3PopupPlacement{
		Direction:        md3PopupDown,
		GapPx:            gapPx,
		MaxHeightPx:      preferredHeightPx,
		TransitionOffset: 4,
	}
	if ctx == nil || ctx.Gtx.Constraints.Max.Y <= 0 {
		return placement
	}

	anchorPos := ctx.Position()
	viewport, ok := ctx.Viewport()
	if !ok {
		viewport = image.Rectangle{Max: ctx.Gtx.Constraints.Max}
	}

	spaceBelow := viewport.Max.Y - (anchorPos.Y + anchorSize.Y) - gapPx
	spaceAbove := anchorPos.Y - viewport.Min.Y - gapPx
	if spaceBelow >= preferredHeightPx || spaceBelow >= spaceAbove {
		placement.MaxHeightPx = clampPositiveInt(preferredHeightPx, spaceBelow)
	} else {
		placement.Direction = md3PopupUp
		placement.TransitionOffset = -4
		placement.MaxHeightPx = clampPositiveInt(preferredHeightPx, spaceAbove)
	}

	popupWidth := popupSize.X
	if popupWidth <= 0 {
		popupWidth = anchorSize.X
	}
	overflowRight := anchorPos.X + popupWidth - viewport.Max.X
	if overflowRight > 0 {
		placement.OffsetX = -overflowRight
	}
	if anchorPos.X+placement.OffsetX < viewport.Min.X {
		placement.OffsetX = viewport.Min.X - anchorPos.X
	}
	return placement
}

func md3PopupPlacementForMeasuredPopup(ctx *internal.Context, anchorSize, popupSize image.Point, placement md3PopupPlacement, extraYOffsetPx int) md3PopupPlacement {
	if ctx == nil || popupSize.Y <= 0 {
		return placement
	}
	viewport, ok := ctx.Viewport()
	if !ok || viewport.Dx() <= 0 || viewport.Dy() <= 0 {
		viewport = image.Rectangle{Max: ctx.Gtx.Constraints.Max}
	}
	if viewport.Dx() <= 0 || viewport.Dy() <= 0 {
		return placement
	}
	anchorPos := ctx.Position()
	down := image.Rect(
		anchorPos.X+placement.OffsetX,
		anchorPos.Y+anchorSize.Y+placement.GapPx+extraYOffsetPx,
		anchorPos.X+placement.OffsetX+popupSize.X,
		anchorPos.Y+anchorSize.Y+placement.GapPx+extraYOffsetPx+popupSize.Y,
	)
	up := image.Rect(
		anchorPos.X+placement.OffsetX,
		anchorPos.Y-placement.GapPx-popupSize.Y+extraYOffsetPx,
		anchorPos.X+placement.OffsetX+popupSize.X,
		anchorPos.Y-placement.GapPx+extraYOffsetPx,
	)
	downOverflow := rectViewportOverflow(down, viewport)
	upOverflow := rectViewportOverflow(up, viewport)
	spaceBelow := viewport.Max.Y - (anchorPos.Y + anchorSize.Y) - placement.GapPx - extraYOffsetPx
	spaceAbove := anchorPos.Y - viewport.Min.Y - placement.GapPx + extraYOffsetPx
	next := placement
	if placement.Direction == md3PopupDown && downOverflow > 0 && upOverflow < downOverflow {
		next.Direction = md3PopupUp
		next.TransitionOffset = -4
	} else if placement.Direction == md3PopupUp && upOverflow > 0 && downOverflow < upOverflow {
		next.Direction = md3PopupDown
		next.TransitionOffset = 4
	}
	if next.Direction == md3PopupUp {
		next.MaxHeightPx = clampPositiveInt(placement.MaxHeightPx, spaceAbove)
	} else {
		next.MaxHeightPx = clampPositiveInt(placement.MaxHeightPx, spaceBelow)
	}
	return next
}

func rectViewportOverflow(rect, viewport image.Rectangle) int {
	overflow := 0
	if rect.Min.X < viewport.Min.X {
		overflow += viewport.Min.X - rect.Min.X
	}
	if rect.Max.X > viewport.Max.X {
		overflow += rect.Max.X - viewport.Max.X
	}
	if rect.Min.Y < viewport.Min.Y {
		overflow += viewport.Min.Y - rect.Min.Y
	}
	if rect.Max.Y > viewport.Max.Y {
		overflow += rect.Max.Y - viewport.Max.Y
	}
	return overflow
}

func md3PopupVerticalPlacement(ctx *internal.Context, anchorHeightPx, preferredHeightPx, gapPx int) md3PopupPlacement {
	return md3PopupPlacementForAnchor(ctx, image.Point{Y: anchorHeightPx}, image.Point{}, preferredHeightPx, gapPx)
}

func clampPositiveInt(value, max int) int {
	if max <= 0 {
		return 1
	}
	if value <= 0 || value > max {
		return max
	}
	return value
}

func md3PopupOffsetY(anchorHeightPx, popupHeightPx int, placement md3PopupPlacement) int {
	if placement.Direction == md3PopupUp {
		return -placement.GapPx - popupHeightPx
	}
	return anchorHeightPx + placement.GapPx
}

func lerpNRGBA(from, to color.NRGBA, t float32) color.NRGBA {
	t = clampFloat32(t, 0, 1)
	return color.NRGBA{
		R: uint8(float32(from.R) + (float32(to.R)-float32(from.R))*t + 0.5),
		G: uint8(float32(from.G) + (float32(to.G)-float32(from.G))*t + 0.5),
		B: uint8(float32(from.B) + (float32(to.B)-float32(from.B))*t + 0.5),
		A: uint8(float32(from.A) + (float32(to.A)-float32(from.A))*t + 0.5),
	}
}

func withAlphaFactor(col color.NRGBA, opacity float32) color.NRGBA {
	opacity = clampFloat32(opacity, 0, 1)
	col.A = uint8(float32(col.A)*opacity + 0.5)
	return col
}

func md3MotionKey(ctx *internal.Context, namespace string) internal.MemoryKey {
	if ctx == nil {
		return internal.MemoryKey{Opaque: "md3-motion:" + namespace}
	}
	return ctx.ScopeMemoryKey("md3-motion:" + namespace)
}

func boolProgress(active bool) float32 {
	if active {
		return 1
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
