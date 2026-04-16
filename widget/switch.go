package widget

import (
	"image/color"

	"fluxui/event"
	"fluxui/internal"
	"fluxui/layout"
)

type SwitchOption func(*switchConfig)

type switchConfig struct {
	disabled      bool
	width         float32
	height        float32
	color         color.NRGBA
	trackColor    color.NRGBA
	thumbColor    color.NRGBA
	hasColor      bool
	hasTrackColor bool
	hasThumbColor bool
	onChange      func(ctx *internal.Context, checked bool)
}

type switchWidget struct {
	value  bool
	config switchConfig
}

func Switch(checked bool, opts ...SwitchOption) Widget {
	cfg := switchConfig{
		width:      50,
		height:     26,
		color:      color.NRGBA{R: 66, G: 133, B: 244, A: 255},
		trackColor: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		thumbColor: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &switchWidget{
		value:  checked,
		config: cfg,
	}
}

func SwitchOnChange(fn func(ctx *internal.Context, checked bool)) SwitchOption {
	return func(cfg *switchConfig) {
		cfg.onChange = fn
	}
}

func SwitchDisabled(disabled bool) SwitchOption {
	return func(cfg *switchConfig) {
		cfg.disabled = disabled
	}
}

func SwitchWidth(width float32) SwitchOption {
	return func(cfg *switchConfig) {
		cfg.width = width
	}
}

func SwitchHeight(height float32) SwitchOption {
	return func(cfg *switchConfig) {
		cfg.height = height
	}
}

func SwitchColor(color color.NRGBA) SwitchOption {
	return func(cfg *switchConfig) {
		cfg.color = color
		cfg.hasColor = true
	}
}

func SwitchTrackColor(color color.NRGBA) SwitchOption {
	return func(cfg *switchConfig) {
		cfg.trackColor = color
		cfg.hasTrackColor = true
	}
}

func SwitchThumbColor(color color.NRGBA) SwitchOption {
	return func(cfg *switchConfig) {
		cfg.thumbColor = color
		cfg.hasThumbColor = true
	}
}

func (s *switchWidget) Layout(ctx *internal.Context) layout.Dimensions {
	clickable := event.UseClickable(ctx)
	if !s.config.disabled {
		for clickable.Clicked(ctx) {
			next := !s.value
			if s.config.onChange != nil {
				s.config.onChange(ctx, next)
			}
		}
	}

	trackColor := s.config.trackColor
	if s.value {
		trackColor = s.config.color
	}

	thumbColor := s.config.thumbColor

	size := ctx.LayoutSwitch(clickable.Handle(), s.value, internal.SwitchSpec{
		Width:      s.config.width,
		Height:     s.config.height,
		TrackColor: trackColor,
		ThumbColor: thumbColor,
		Disabled:   s.config.disabled,
	})

	return layout.Dimensions{Size: size}
}
