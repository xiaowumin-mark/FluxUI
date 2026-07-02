package widget

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/style"

	"gioui.org/f32"
	gioEvent "gioui.org/io/event"
	"gioui.org/io/pointer"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

type SliderOption func(*sliderConfig)

type sliderConfig struct {
	disabled         bool
	rangeSlider      bool
	min              float32
	max              float32
	value            float32
	valueStart       float32
	valueEnd         float32
	step             float32
	width            float32
	labeled          bool
	ticks            bool
	valueLabel       string
	valueLabelStart  string
	valueLabelEnd    string
	trackColor       color.NRGBA
	thumbColor       color.NRGBA
	progressColor    color.NRGBA
	labelColor       color.NRGBA
	labelTextColor   color.NRGBA
	hasTrackColor    bool
	hasThumbColor    bool
	hasProgressColor bool
	hasLabelColor    bool
	hasLabelText     bool
	onChange         func(ctx *internal.Context, value float32)
	onRangeChange    func(ctx *internal.Context, start, end float32)
	ref              *SliderRef
	decoration       style.Decoration
}

type sliderWidget struct {
	config sliderConfig
}

type sliderState struct {
	tag        any
	start      float32
	end        float32
	active     sliderHandle
	pressed    bool
	hoverStart bool
	hoverEnd   bool
	pointerID  pointer.ID
	actionFrom float32
	actionTo   float32
	canFlip    bool
}

type sliderHandle uint8

const (
	sliderHandleNone sliderHandle = iota
	sliderHandleStart
	sliderHandleEnd
)

func Slider(value float32, opts ...SliderOption) Widget {
	cfg := defaultSliderConfig()
	cfg.value = value
	cfg.valueStart = cfg.min
	cfg.valueEnd = value
	for _, opt := range opts {
		opt(&cfg)
	}
	return &sliderWidget{config: normalizeSliderConfig(cfg)}
}

// RangeSlider creates a Material 3 range slider with start and end handles.
func RangeSlider(start, end float32, opts ...SliderOption) Widget {
	cfg := defaultSliderConfig()
	cfg.rangeSlider = true
	cfg.value = end
	cfg.valueStart = start
	cfg.valueEnd = end
	for _, opt := range opts {
		opt(&cfg)
	}
	return &sliderWidget{config: normalizeSliderConfig(cfg)}
}

func defaultSliderConfig() sliderConfig {
	return sliderConfig{
		min:   0,
		max:   100,
		step:  1,
		width: 200,
	}
}

func normalizeSliderConfig(cfg sliderConfig) sliderConfig {
	if cfg.max <= cfg.min {
		cfg.max = cfg.min + 1
	}
	if cfg.width <= 0 {
		cfg.width = 200
	}
	cfg.value = applySliderStep(cfg.value, cfg.min, cfg.max, cfg.step)
	cfg.valueStart = applySliderStep(cfg.valueStart, cfg.min, cfg.max, cfg.step)
	cfg.valueEnd = applySliderStep(cfg.valueEnd, cfg.min, cfg.max, cfg.step)
	if cfg.valueStart > cfg.valueEnd {
		cfg.valueStart, cfg.valueEnd = cfg.valueEnd, cfg.valueStart
	}
	return cfg
}

func SliderOnChange(fn func(ctx *internal.Context, value float32)) SliderOption {
	return func(cfg *sliderConfig) { cfg.onChange = fn }
}

func SliderOnRangeChange(fn func(ctx *internal.Context, start, end float32)) SliderOption {
	return func(cfg *sliderConfig) { cfg.onRangeChange = fn }
}

func SliderAttachRef(ref *SliderRef) SliderOption {
	return func(cfg *sliderConfig) { cfg.ref = ref }
}

func SliderDisabled(disabled bool) SliderOption {
	return func(cfg *sliderConfig) { cfg.disabled = disabled }
}

func SliderMin(min float32) SliderOption {
	return func(cfg *sliderConfig) { cfg.min = min }
}

func SliderMax(max float32) SliderOption {
	return func(cfg *sliderConfig) { cfg.max = max }
}

func SliderStep(step float32) SliderOption {
	return func(cfg *sliderConfig) { cfg.step = step }
}

func SliderWidth(width float32) SliderOption {
	return func(cfg *sliderConfig) { cfg.width = width }
}

func SliderLabeled(labeled bool) SliderOption {
	return func(cfg *sliderConfig) { cfg.labeled = labeled }
}

func SliderTicks(ticks bool) SliderOption {
	return func(cfg *sliderConfig) { cfg.ticks = ticks }
}

func SliderValueLabel(label string) SliderOption {
	return func(cfg *sliderConfig) { cfg.valueLabel = label }
}

func SliderValueLabels(start, end string) SliderOption {
	return func(cfg *sliderConfig) {
		cfg.valueLabelStart = start
		cfg.valueLabelEnd = end
	}
}

func SliderRange(start, end float32) SliderOption {
	return func(cfg *sliderConfig) {
		cfg.rangeSlider = true
		cfg.valueStart = start
		cfg.valueEnd = end
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

func SliderLabelColor(color color.NRGBA) SliderOption {
	return func(cfg *sliderConfig) {
		cfg.labelColor = color
		cfg.hasLabelColor = true
	}
}

func SliderLabelTextColor(color color.NRGBA) SliderOption {
	return func(cfg *sliderConfig) {
		cfg.labelTextColor = color
		cfg.hasLabelText = true
	}
}

func SliderDecoration(d style.Decoration) SliderOption {
	return func(cfg *sliderConfig) { cfg.decoration = d }
}

func (s *sliderWidget) Layout(ctx *internal.Context) layout.Dimensions {
	state := sliderStateFor(ctx)
	cfg := normalizeSliderConfig(s.config)
	if cfg.ref != nil {
		cfg.ref.bindInvalidator(redrawInvalidator(ctx))
		current := cfg.value
		if cfg.rangeSlider {
			current = cfg.valueEnd
		}
		for _, cmd := range cfg.ref.drainCommands() {
			switch cmd.kind {
			case sliderCmdSet:
				current = applySliderStep(cmd.value, cfg.min, cfg.max, cfg.step)
			case sliderCmdStep:
				current = applySliderStep(current+cmd.delta, cfg.min, cfg.max, cfg.step)
			}
		}
		if cfg.rangeSlider {
			cfg.valueEnd = current
			if cfg.valueStart > cfg.valueEnd {
				cfg.valueStart = cfg.valueEnd
			}
		} else {
			cfg.value = current
		}
	}

	if !state.pressed {
		state.start = toSliderProgress(cfg.valueStart, cfg.min, cfg.max)
		if cfg.rangeSlider {
			state.end = toSliderProgress(cfg.valueEnd, cfg.min, cfg.max)
		} else {
			state.end = toSliderProgress(cfg.value, cfg.min, cfg.max)
		}
	}
	if !cfg.rangeSlider {
		state.start = 0
	}
	if state.end < state.start {
		state.end = state.start
	}

	content := func(contentCtx *internal.Context) image.Point {
		return s.layoutMaterialSlider(contentCtx, state, cfg)
	}
	target := content
	if hasAnyDecoration(cfg.decoration) {
		target = func(targetCtx *internal.Context) image.Point {
			hovered := state.hoverStart || state.hoverEnd
			pressed := state.pressed
			active := resolveDecorationState(cfg.decoration, hovered, pressed, cfg.disabled)
			duration, easing := md3InteractionTiming(targetCtx, hovered, pressed, false, cfg.disabled)
			visual := md3AnimateDecoration(targetCtx, "slider-decoration", stripStateDecoration(active), duration, easing)
			return layoutDecorationShell(targetCtx.Child(0), visual, content).Size
		}
	}
	dims := layout.Dimensions{Size: target(ctx.Child(0))}

	startValue := applySliderStep(cfg.min+state.start*(cfg.max-cfg.min), cfg.min, cfg.max, cfg.step)
	endValue := applySliderStep(cfg.min+state.end*(cfg.max-cfg.min), cfg.min, cfg.max, cfg.step)
	if !cfg.disabled {
		if cfg.rangeSlider {
			if (math.Abs(float64(startValue-cfg.valueStart)) > 0.0001 || math.Abs(float64(endValue-cfg.valueEnd)) > 0.0001) && cfg.onRangeChange != nil {
				cfg.onRangeChange(ctx, startValue, endValue)
			}
		} else if math.Abs(float64(endValue-cfg.value)) > 0.0001 && cfg.onChange != nil {
			cfg.onChange(ctx, endValue)
		}
	}
	return dims
}

func (s *sliderWidget) layoutMaterialSlider(ctx *internal.Context, state *sliderState, cfg sliderConfig) image.Point {
	gtx := ctx.Gtx
	width := gtx.Dp(unit.Dp(cfg.width))
	if width < gtx.Dp(unit.Dp(200)) {
		width = gtx.Dp(unit.Dp(200))
	}
	stateLayer := gtx.Dp(unit.Dp(40))
	labelSpace := 0
	if cfg.labeled {
		labelSpace = gtx.Dp(unit.Dp(34))
	}
	size := image.Pt(width, stateLayer+labelSpace)
	trackTop := labelSpace
	trackRect := image.Rect(0, trackTop, width, trackTop+stateLayer)

	registerSliderPointer(ctx, state, cfg.disabled, trackRect)
	processSliderPointerEvents(ctx, state, cfg, trackRect)

	colors := resolveSliderColors(ctx, cfg)
	startProgress := clampFloat32(state.start, 0, 1)
	endProgress := clampFloat32(state.end, 0, 1)
	if !cfg.rangeSlider {
		startProgress = 0
	}
	if endProgress < startProgress {
		endProgress = startProgress
	}

	trackHeight := gtx.Dp(unit.Dp(4))
	tickSize := gtx.Dp(unit.Dp(2))
	thumbSize := gtx.Dp(unit.Dp(20))
	trackStart := stateLayer / 2
	trackEnd := width - stateLayer/2
	if trackEnd < trackStart {
		trackEnd = trackStart
	}
	trackWidth := trackEnd - trackStart
	centerY := trackTop + stateLayer/2
	startX := trackStart + int(float32(trackWidth)*startProgress+0.5)
	endX := trackStart + int(float32(trackWidth)*endProgress+0.5)

	drawSliderTrack(gtx, trackStart, trackEnd, centerY, trackHeight, startX, endX, colors.track, colors.active)
	if cfg.ticks {
		drawSliderTicks(gtx, cfg, trackStart, trackEnd, centerY, tickSize, startProgress, endProgress, colors.tickInactive, colors.tickActive)
	}

	activeOpacity := materialStateLayerOpacity(ctx, state.hoverEnd || state.hoverStart, state.pressed)
	if state.pressed {
		activeOpacity = style.StateLayerPressedOpacity
	}
	activeOpacity = md3AnimateStateLayerOpacity(ctx, "slider-state-layer", activeOpacity)
	if activeOpacity > 0 && !cfg.disabled {
		cx := endX
		if state.active == sliderHandleStart {
			cx = startX
		}
		ctx.DrawStateLayerCircle(nil, image.Pt(cx, centerY), 40, internal.RippleSpec{Color: colors.active, Radius: 20, Opacity: style.StateLayerPressedOpacity}, activeOpacity)
	}

	if cfg.rangeSlider && state.active == sliderHandleStart {
		drawSliderHandle(ctx, image.Pt(endX, centerY), thumbSize, colors.thumb)
		drawSliderHandle(ctx, image.Pt(startX, centerY), thumbSize, colors.thumb)
	} else {
		if cfg.rangeSlider {
			drawSliderHandle(ctx, image.Pt(startX, centerY), thumbSize, colors.thumb)
		}
		drawSliderHandle(ctx, image.Pt(endX, centerY), thumbSize, colors.thumb)
	}

	if cfg.labeled {
		showStart := cfg.rangeSlider && (state.hoverStart || state.pressed && state.active == sliderHandleStart)
		showEnd := state.hoverEnd || state.pressed && state.active == sliderHandleEnd || (!cfg.rangeSlider && state.pressed)
		startLabelProgress := md3SliderLabelProgress(ctx.Child(20), showStart)
		endLabelProgress := md3SliderLabelProgress(ctx.Child(21), showEnd || (!cfg.rangeSlider && state.hoverEnd))
		if startLabelProgress > 0.001 {
			drawSliderLabel(ctx, startX, labelSpace, sliderLabelText(cfg.valueLabelStart, cfg.min+startProgress*(cfg.max-cfg.min)), colors.label, colors.labelText, startLabelProgress)
		}
		if endLabelProgress > 0.001 {
			label := cfg.valueLabel
			if cfg.rangeSlider {
				label = cfg.valueLabelEnd
			}
			drawSliderLabel(ctx, endX, labelSpace, sliderLabelText(label, cfg.min+endProgress*(cfg.max-cfg.min)), colors.label, colors.labelText, endLabelProgress)
		}
	}

	return size
}

type sliderColors struct {
	track        color.NRGBA
	active       color.NRGBA
	thumb        color.NRGBA
	label        color.NRGBA
	labelText    color.NRGBA
	tickActive   color.NRGBA
	tickInactive color.NRGBA
}

func resolveSliderColors(ctx *internal.Context, cfg sliderConfig) sliderColors {
	cs := ctx.Theme().Colors
	track := cs.SurfaceContainerHighest
	if track.A == 0 {
		track = cs.SurfaceVariant
	}
	active := cs.Primary
	thumb := cs.Primary
	label := cs.Primary
	labelText := cs.OnPrimary
	if cfg.hasTrackColor {
		track = cfg.trackColor
	}
	if cfg.hasProgressColor {
		active = cfg.progressColor
	}
	if cfg.hasThumbColor {
		thumb = cfg.thumbColor
	}
	if cfg.hasLabelColor {
		label = cfg.labelColor
	}
	if cfg.hasLabelText {
		labelText = cfg.labelTextColor
	}
	if cfg.disabled {
		track = withAlphaFactor(cs.OnSurface, 0.12)
		active = withAlphaFactor(cs.OnSurface, 0.38)
		thumb = withAlphaFactor(cs.OnSurface, 0.38)
	}
	return sliderColors{
		track:        track,
		active:       active,
		thumb:        thumb,
		label:        label,
		labelText:    labelText,
		tickActive:   withAlphaFactor(cs.OnPrimary, 0.38),
		tickInactive: withAlphaFactor(cs.OnSurfaceVariant, 0.38),
	}
}

func registerSliderPointer(ctx *internal.Context, state *sliderState, disabled bool, rect image.Rectangle) {
	if disabled || state == nil || state.tag == nil {
		return
	}
	defer pointer.PassOp{}.Push(ctx.Gtx.Ops).Pop()
	area := clip.Rect(rect).Push(ctx.Gtx.Ops)
	gioEvent.Op(ctx.Gtx.Ops, state.tag)
	area.Pop()
	pointer.CursorPointer.Add(ctx.Gtx.Ops)
}

func processSliderPointerEvents(ctx *internal.Context, state *sliderState, cfg sliderConfig, rect image.Rectangle) {
	if cfg.disabled || state == nil || state.tag == nil {
		state.hoverStart = false
		state.hoverEnd = false
		state.pressed = false
		state.active = sliderHandleNone
		return
	}
	for {
		ev, ok := ctx.Gtx.Event(pointer.Filter{Target: state.tag, Kinds: pointer.Press | pointer.Release | pointer.Move | pointer.Drag | pointer.Enter | pointer.Leave | pointer.Cancel})
		if !ok {
			break
		}
		pe, ok := ev.(pointer.Event)
		if !ok {
			continue
		}
		switch pe.Kind {
		case pointer.Press:
			state.pressed = true
			state.pointerID = pe.PointerID
			state.active = nearestSliderHandle(state, cfg.rangeSlider, pe.Position, rect)
			state.actionFrom = state.start
			state.actionTo = state.end
			state.canFlip = cfg.rangeSlider
			updateSliderValueFromPosition(state, cfg, pe.Position, rect)
		case pointer.Move, pointer.Drag:
			updateSliderHover(state, cfg.rangeSlider, pe.Position, rect)
			if state.pressed && (pe.Buttons.Contain(pointer.ButtonPrimary) || pe.Kind == pointer.Drag) {
				updateSliderValueFromPosition(state, cfg, pe.Position, rect)
			}
		case pointer.Enter:
			updateSliderHover(state, cfg.rangeSlider, pe.Position, rect)
		case pointer.Leave:
			state.hoverStart = false
			state.hoverEnd = false
		case pointer.Release, pointer.Cancel:
			if state.pressed {
				updateSliderValueFromPosition(state, cfg, pe.Position, rect)
			}
			state.pressed = false
			state.active = sliderHandleNone
			state.canFlip = false
		}
	}
}

func nearestSliderHandle(state *sliderState, ranged bool, pos f32.Point, rect image.Rectangle) sliderHandle {
	if !ranged {
		return sliderHandleEnd
	}
	startX := sliderProgressToX(state.start, rect)
	endX := sliderProgressToX(state.end, rect)
	x := int(pos.X + 0.5)
	if absInt(x-startX) <= absInt(x-endX) {
		return sliderHandleStart
	}
	return sliderHandleEnd
}

func updateSliderHover(state *sliderState, ranged bool, pos f32.Point, rect image.Rectangle) {
	const hoverRadius = 20
	x := int(pos.X + 0.5)
	y := int(pos.Y + 0.5)
	centerY := rect.Min.Y + rect.Dy()/2
	state.hoverEnd = absInt(x-sliderProgressToX(state.end, rect)) <= hoverRadius && absInt(y-centerY) <= hoverRadius
	if ranged {
		state.hoverStart = absInt(x-sliderProgressToX(state.start, rect)) <= hoverRadius && absInt(y-centerY) <= hoverRadius
	} else {
		state.hoverStart = false
	}
}

func updateSliderValueFromPosition(state *sliderState, cfg sliderConfig, pos f32.Point, rect image.Rectangle) {
	progress := sliderXToProgress(pos.X, rect)
	if cfg.step > 0 {
		value := applySliderStep(cfg.min+progress*(cfg.max-cfg.min), cfg.min, cfg.max, cfg.step)
		progress = toSliderProgress(value, cfg.min, cfg.max)
	}
	if cfg.rangeSlider && state.active == sliderHandleStart {
		if progress > state.end {
			if state.canFlip && sliderProgressEqual(state.actionFrom, state.actionTo) {
				state.active = sliderHandleEnd
				state.start = state.actionFrom
				state.end = progress
				state.canFlip = false
				return
			}
			progress = state.end
		}
		state.start = progress
		return
	}
	if cfg.rangeSlider && progress < state.start {
		if state.canFlip && sliderProgressEqual(state.actionFrom, state.actionTo) {
			state.active = sliderHandleStart
			state.end = state.actionTo
			state.start = progress
			state.canFlip = false
			return
		}
		progress = state.start
	}
	state.end = progress
}

func sliderProgressEqual(a, b float32) bool {
	return math.Abs(float64(a-b)) <= 0.0001
}

func sliderProgressToX(progress float32, rect image.Rectangle) int {
	trackStart := rect.Min.X + rect.Dy()/2
	trackEnd := rect.Max.X - rect.Dy()/2
	if trackEnd < trackStart {
		trackEnd = trackStart
	}
	return trackStart + int(float32(trackEnd-trackStart)*clampFloat32(progress, 0, 1)+0.5)
}

func sliderXToProgress(x float32, rect image.Rectangle) float32 {
	trackStart := float32(rect.Min.X + rect.Dy()/2)
	trackEnd := float32(rect.Max.X - rect.Dy()/2)
	if trackEnd <= trackStart {
		return 0
	}
	return clampFloat32((x-trackStart)/(trackEnd-trackStart), 0, 1)
}

func drawSliderTrack(gtx gioLayout.Context, start, end, centerY, height, activeStart, activeEnd int, inactive, active color.NRGBA) {
	if end <= start || height <= 0 {
		return
	}
	rr := height / 2
	track := image.Rect(start, centerY-height/2, end, centerY+height/2)
	paint.FillShape(gtx.Ops, inactive, clip.UniformRRect(track, rr).Op(gtx.Ops))
	if activeEnd > activeStart {
		activeRect := image.Rect(activeStart, centerY-height/2, activeEnd, centerY+height/2)
		paint.FillShape(gtx.Ops, active, clip.UniformRRect(activeRect, rr).Op(gtx.Ops))
	}
}

func drawSliderTicks(gtx gioLayout.Context, cfg sliderConfig, start, end, centerY, size int, startProgress, endProgress float32, inactive, active color.NRGBA) {
	if cfg.step <= 0 || end <= start || size <= 0 {
		return
	}
	count := int(math.Round(float64((cfg.max - cfg.min) / cfg.step)))
	if count <= 0 || count > 200 {
		return
	}
	for i := 0; i <= count; i++ {
		p := float32(i) / float32(count)
		x := start + int(float32(end-start)*p+0.5)
		col := inactive
		if p >= startProgress-0.0001 && p <= endProgress+0.0001 {
			col = active
		}
		r := image.Rect(x-size/2, centerY-size/2, x+size/2, centerY+size/2)
		paint.FillShape(gtx.Ops, col, clip.Ellipse(r).Op(gtx.Ops))
	}
}

func drawSliderHandle(ctx *internal.Context, center image.Point, size int, col color.NRGBA) {
	if size <= 0 {
		return
	}
	rect := image.Rect(center.X-size/2, center.Y-size/2, center.X+size/2, center.Y+size/2)
	paint.FillShape(ctx.Gtx.Ops, col, clip.Ellipse(rect).Op(ctx.Gtx.Ops))
}

func md3SliderLabelProgress(ctx *internal.Context, visible bool) float32 {
	return md3AnimateFloatDirectionalRetained(ctx, "slider-label", boolProgress(visible), 150*time.Millisecond, 100*time.Millisecond, style.InteractionEmphasizedEasing, style.InteractionStandardAccelerateEasing)
}

func drawSliderLabel(ctx *internal.Context, centerX, bottomY int, text string, bg, fg color.NRGBA, progress float32) {
	if text == "" {
		return
	}
	progress = clampFloat32(progress, 0, 1)
	labelH := ctx.Gtx.Dp(unit.Dp(28))
	padX := ctx.Gtx.Dp(unit.Dp(8))
	measureCtx := *ctx
	measureCtx.Gtx.Constraints = gioLayout.Constraints{Max: image.Pt(10000, 10000)}
	measureOps := new(op.Ops)
	measureCtx.Gtx.Ops = measureOps
	textSize := measureCtx.LayoutText(internal.TextSpec{Content: text, Size: 12, LineHeight: 16, Color: color.NRGBA{}, Alignment: internal.AlignStart, Font: ctx.Font(), FontReady: true})
	labelW := textSize.X + padX*2
	if labelW < labelH {
		labelW = labelH
	}
	x := centerX - labelW/2
	y := bottomY - labelH - ctx.Gtx.Dp(unit.Dp(4))
	if x < 0 {
		x = 0
	}
	if max := ctx.Gtx.Constraints.Max.X; max > 0 && x+labelW > max {
		x = max - labelW
	}
	if y < 0 {
		y = 0
	}
	visibleH := int(float32(labelH)*progress + 0.5)
	if visibleH < 1 {
		visibleH = 1
	}
	visibleY := y + labelH - visibleH
	rect := image.Rect(x, visibleY, x+labelW, y+labelH)
	if progress < 0.999 {
		defer paint.PushOpacity(ctx.Gtx.Ops, progress).Pop()
	}
	paint.FillShape(ctx.Gtx.Ops, bg, clip.UniformRRect(rect, labelH/2).Op(ctx.Gtx.Ops))
	tri := ctx.Gtx.Dp(unit.Dp(8))
	var path clip.Path
	path.Begin(ctx.Gtx.Ops)
	path.MoveTo(f32.Pt(float32(centerX), float32(bottomY)))
	path.LineTo(f32.Pt(float32(centerX-tri/2), float32(bottomY-tri)))
	path.LineTo(f32.Pt(float32(centerX+tri/2), float32(bottomY-tri)))
	path.Close()
	paint.FillShape(ctx.Gtx.Ops, bg, clip.Outline{Path: path.End()}.Op())
	labelCtx := *ctx
	labelCtx.Gtx.Constraints = gioLayout.Exact(image.Pt(labelW, textSize.Y))
	stack := op.Offset(image.Pt(x, y+(labelH-textSize.Y)/2)).Push(ctx.Gtx.Ops)
	labelCtx.LayoutText(internal.TextSpec{Content: text, Size: 12, LineHeight: 16, Color: fg, Alignment: internal.AlignCenter, Font: ctx.Font(), FontReady: true})
	stack.Pop()
}

func sliderLabelText(label string, value float32) string {
	if label != "" {
		return label
	}
	if math.Abs(float64(value-float32(int(value)))) < 0.0001 {
		return strconv.Itoa(int(value))
	}
	return fmt.Sprintf("%.2f", value)
}

func sliderStateFor(ctx *internal.Context) *sliderState {
	value := ctx.Memo("slider", func() any {
		state := &sliderState{}
		state.tag = state
		return state
	})
	state, ok := value.(*sliderState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUI/widget: slider state type mismatch")
	}
	if state.tag == nil {
		state.tag = state
	}
	return state
}

func toSliderProgress(value, min, max float32) float32 {
	if max <= min {
		return 0
	}
	p := (value - min) / (max - min)
	return clampFloat32(p, 0, 1)
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

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
