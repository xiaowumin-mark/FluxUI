package widget

import (
	"image/color"
	"math"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
	layout "github.com/xiaowumin-mark/FluxUI/layout"

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
}

type sliderWidget struct {
	config sliderConfig
}

func Slider(value float32, opts ...SliderOption) Widget {
	cfg := sliderConfig{
		min:           0,
		max:           100,
		value:         value,
		step:          1,
		trackColor:    color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		thumbColor:    color.NRGBA{R: 66, G: 133, B: 244, A: 255},
		progressColor: color.NRGBA{R: 66, G: 133, B: 244, A: 255},
		width:         200,
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

func (s *sliderWidget) Layout(ctx *internal.Context) layout.Dimensions {
	sliderState := sliderStateFor(ctx)
	currentValue := applySliderStep(s.config.value, s.config.min, s.config.max, s.config.step)
	if s.config.ref != nil {
		s.config.ref.bindInvalidator(ctx.Runtime().RequestRedraw)
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
	sliderState.Value = progress
	before := sliderState.Value

	trackColor := ctx.Theme().SurfaceMuted
	if s.config.hasTrackColor {
		trackColor = s.config.trackColor
	}
	thumbColor := ctx.Theme().Primary
	if s.config.hasThumbColor {
		thumbColor = s.config.thumbColor
	}
	progressColor := ctx.Theme().Primary
	if s.config.hasProgressColor {
		progressColor = s.config.progressColor
	}

	size := ctx.LayoutSlider(sliderState, internal.SliderSpec{
		Width:         s.config.width,
		TrackColor:    trackColor,
		ThumbColor:    thumbColor,
		ProgressColor: progressColor,
		Disabled:      s.config.disabled,
	})

	if !s.config.disabled && math.Abs(float64(sliderState.Value-before)) > 0.0001 {
		next := s.config.min + sliderState.Value*(s.config.max-s.config.min)
		next = applySliderStep(next, s.config.min, s.config.max, s.config.step)
		sliderState.Value = toSliderProgress(next, s.config.min, s.config.max)
		if s.config.onChange != nil {
			s.config.onChange(ctx, next)
		}
	}

	return layout.Dimensions{Size: size}
}

func sliderStateFor(ctx *internal.Context) *gioWidget.Float {
	value := ctx.Memo("slider", func() any {
		return &gioWidget.Float{}
	})
	state, ok := value.(*gioWidget.Float)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUIwidget: slider state type mismatch")
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
