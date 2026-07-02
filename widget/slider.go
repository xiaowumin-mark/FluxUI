package widget

import (
	"image"
	"image/color"
	"math"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
	layout "github.com/xiaowumin-mark/FluxUI/layout"
	style "github.com/xiaowumin-mark/FluxUI/style"

	"gioui.org/io/pointer"
	"gioui.org/op"
	"gioui.org/unit"
	gioWidget "gioui.org/widget"
)

type SliderOption func(*sliderConfig)

type sliderConfig struct {
	disabled         bool
	min              float32
	max              float32
	value            float32
	step             float32
	trackColor       color.NRGBA
	thumbColor       color.NRGBA
	progressColor    color.NRGBA
	width            float32
	hasTrackColor    bool
	hasThumbColor    bool
	hasProgressColor bool
	onChange         func(ctx *internal.Context, value float32)
	ref              *SliderRef
	decoration       style.Decoration
}

type sliderWidget struct {
	config sliderConfig
}

type sliderState struct {
	value *gioWidget.Float
	hover *internal.ClickableState
}

func Slider(value float32, opts ...SliderOption) Widget {
	cfg := sliderConfig{
		min:   0,
		max:   100,
		value: value,
		step:  1,
		width: 200,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &sliderWidget{
		config: cfg,
	}
}

func SliderOnChange(fn func(ctx *internal.Context, value float32)) SliderOption {
	return func(cfg *sliderConfig) {
		cfg.onChange = fn
	}
}

// SliderAttachRef 绑定命令型引用，用于外部主动设置值。
func SliderAttachRef(ref *SliderRef) SliderOption {
	return func(cfg *sliderConfig) {
		cfg.ref = ref
	}
}

func SliderDisabled(disabled bool) SliderOption {
	return func(cfg *sliderConfig) {
		cfg.disabled = disabled
	}
}

func SliderMin(min float32) SliderOption {
	return func(cfg *sliderConfig) {
		cfg.min = min
	}
}

func SliderMax(max float32) SliderOption {
	return func(cfg *sliderConfig) {
		cfg.max = max
	}
}

func SliderStep(step float32) SliderOption {
	return func(cfg *sliderConfig) {
		cfg.step = step
	}
}

func SliderWidth(width float32) SliderOption {
	return func(cfg *sliderConfig) {
		cfg.width = width
	}
}

func SliderTrackColor(color color.NRGBA) SliderOption {
	return func(cfg *sliderConfig) {
		cfg.trackColor = color
		cfg.hasTrackColor = true
	}
}

func SliderThumbColor(color color.NRGBA) SliderOption {
	return func(cfg *sliderConfig) {
		cfg.thumbColor = color
		cfg.hasThumbColor = true
	}
}

func SliderProgressColor(color color.NRGBA) SliderOption {
	return func(cfg *sliderConfig) {
		cfg.progressColor = color
		cfg.hasProgressColor = true
	}
}

// SliderDecoration 通过 Decoration 统一设置滑块外层装饰和交互态。
func SliderDecoration(d style.Decoration) SliderOption {
	return func(cfg *sliderConfig) {
		cfg.decoration = d
	}
}

func (s *sliderWidget) Layout(ctx *internal.Context) layout.Dimensions {
	state := sliderStateFor(ctx)
	sliderValue := state.value
	clickable := state.hover
	currentValue := applySliderStep(s.config.value, s.config.min, s.config.max, s.config.step)
	if sliderValue.Dragging() {
		currentValue = s.config.min + sliderValue.Value*(s.config.max-s.config.min)
		currentValue = applySliderStep(currentValue, s.config.min, s.config.max, s.config.step)
	}
	if s.config.ref != nil {
		s.config.ref.bindInvalidator(redrawInvalidator(ctx))
		for _, cmd := range s.config.ref.drainCommands() {
			switch cmd.kind {
			case sliderCmdSet:
				currentValue = applySliderStep(cmd.value, s.config.min, s.config.max, s.config.step)
			case sliderCmdStep:
				currentValue = applySliderStep(currentValue+cmd.delta, s.config.min, s.config.max, s.config.step)
			}
		}
		if currentValue != s.config.value && s.config.onChange != nil {
			s.config.onChange(ctx, currentValue)
		}
	}
	progress := toSliderProgress(currentValue, s.config.min, s.config.max)
	if !sliderValue.Dragging() {
		sliderValue.Value = progress
	}
	before := sliderValue.Value

	cs := ctx.Theme().Colors
	trackColor := cs.SurfaceVariant
	if s.config.hasTrackColor {
		trackColor = s.config.trackColor
	}
	thumbColor := cs.Primary
	if s.config.hasThumbColor {
		thumbColor = s.config.thumbColor
	}
	progressColor := cs.Primary
	if s.config.hasProgressColor {
		progressColor = s.config.progressColor
	}
	if s.config.disabled {
		trackColor = style.DisabledContainer(cs.OnSurface)
		thumbColor = style.DisabledContent(cs.OnSurface)
		progressColor = style.DisabledContent(cs.OnSurface)
	}
	content := func(contentCtx *internal.Context) image.Point {
		return contentCtx.LayoutSlider(sliderValue, internal.SliderSpec{
			Width:         s.config.width,
			TrackColor:    trackColor,
			ThumbColor:    thumbColor,
			ProgressColor: progressColor,
			Disabled:      s.config.disabled,
			Hovered:       clickable != nil && clickable.Hovered(),
			Pressed:       sliderValue.Dragging() || clickable != nil && clickable.Pressed(),
			Dragged:       sliderValue.Dragging(),
		})
	}

	target := content
	if hasAnyDecoration(s.config.decoration) {
		target = func(targetCtx *internal.Context) image.Point {
			hovered := clickable != nil && clickable.Hovered()
			pressed := sliderValue.Dragging() || clickable != nil && clickable.Pressed()
			active := resolveDecorationState(s.config.decoration, hovered, pressed, s.config.disabled)
			duration, easing := md3InteractionTiming(targetCtx, hovered, pressed, false, s.config.disabled)
			visual := md3AnimateDecoration(targetCtx, "slider-decoration", stripStateDecoration(active), duration, easing)
			return layoutDecorationShell(targetCtx.Child(0), visual, content).Size
		}
	}

	dims := layoutSliderStateLayer(ctx.Child(0), clickable, s.config.disabled, internal.RippleSpec{
		Color:   thumbColor,
		Radius:  20,
		Opacity: style.StateLayerPressedOpacity,
	}, 48, 40, func(size image.Point) image.Point {
		thumbSize := size.Y
		if thumbSize <= 0 {
			thumbSize = ctx.Gtx.Dp(safeDp(20))
		}
		travel := size.X - thumbSize
		if travel < 0 {
			travel = 0
		}
		x := thumbSize/2 + int(float32(travel)*sliderValue.Value+0.5)
		return image.Pt(x, size.Y/2)
	}, func() float32 {
		if sliderValue.Dragging() {
			return style.StateLayerDraggedOpacity
		}
		return materialStateLayerOpacity(ctx, clickable != nil && clickable.Hovered(), clickable != nil && clickable.Pressed())
	}, target)

	if !s.config.disabled && math.Abs(float64(sliderValue.Value-before)) > 0.0001 {
		next := s.config.min + sliderValue.Value*(s.config.max-s.config.min)
		next = applySliderStep(next, s.config.min, s.config.max, s.config.step)
		sliderValue.Value = toSliderProgress(next, s.config.min, s.config.max)
		if s.config.onChange != nil {
			s.config.onChange(ctx, next)
		}
	}

	return dims
}

func layoutSliderStateLayer(ctx *internal.Context, clickable *internal.ClickableState, disabled bool, spec internal.RippleSpec, minTargetDp, layerDiameterDp float32, layerCenter func(image.Point) image.Point, stateOpacity func() float32, child func(*internal.Context) image.Point) layout.Dimensions {
	if child == nil {
		return layout.Dimensions{}
	}

	macro := op.Record(ctx.Gtx.Ops)
	childSize := child(ctx.Child(0))
	childCall := macro.Stop()

	if disabled || clickable == nil {
		childCall.Add(ctx.Gtx.Ops)
		return layout.Dimensions{Size: childSize}
	}

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

	targetOffset := image.Point{
		X: -(targetSize.X - childSize.X) / 2,
		Y: -(targetSize.Y - childSize.Y) / 2,
	}
	targetMacro := op.Record(ctx.Gtx.Ops)
	pass := pointer.PassOp{}.Push(ctx.Gtx.Ops)
	stack := op.Offset(targetOffset).Push(ctx.Gtx.Ops)
	ctx.LayoutClickArea(clickable, func(*internal.Context) image.Point {
		return targetSize
	})
	stack.Pop()
	pass.Pop()
	targetCall := targetMacro.Stop()

	center := image.Point{}
	if layerCenter != nil {
		center = layerCenter(childSize)
	}
	opacity := float32(0)
	if stateOpacity != nil {
		opacity = stateOpacity()
	}
	opacity = md3AnimateStateLayerOpacity(ctx, "slider-state-layer", opacity)

	layerMacro := op.Record(ctx.Gtx.Ops)
	ctx.DrawStateLayerCircle(nil, center, layerDiameterDp, spec, opacity)
	layerCall := layerMacro.Stop()

	childCall.Add(ctx.Gtx.Ops)
	targetCall.Add(ctx.Gtx.Ops)
	layerCall.Add(ctx.Gtx.Ops)
	return layout.Dimensions{Size: childSize}
}

func sliderStateFor(ctx *internal.Context) *sliderState {
	value := ctx.Memo("slider", func() any {
		return &sliderState{
			value: &gioWidget.Float{},
			hover: internal.NewClickableState(),
		}
	})
	state, ok := value.(*sliderState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUIwidget: slider state type mismatch")
	}
	if state.value == nil {
		state.value = &gioWidget.Float{}
	}
	if state.hover == nil {
		state.hover = internal.NewClickableState()
	}
	return state
}

func toSliderProgress(value, min, max float32) float32 {
	if max <= min {
		return 0
	}
	p := (value - min) / (max - min)
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return p
}

func applySliderStep(value, min, max, step float32) float32 {
	if value < min {
		value = min
	}
	if value > max {
		value = max
	}
	if step <= 0 {
		return value
	}

	steps := float64((value - min) / step)
	rounded := math.Round(steps)
	next := min + float32(rounded)*step
	if next < min {
		next = min
	}
	if next > max {
		next = max
	}
	return next
}
