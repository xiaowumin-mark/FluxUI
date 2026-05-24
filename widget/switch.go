package widget

import (
	"image"
	"image/color"

	event "github.com/xiaowumin-mark/FluxUI/event"
	internal "github.com/xiaowumin-mark/FluxUI/internal"
	layout "github.com/xiaowumin-mark/FluxUI/layout"
	style "github.com/xiaowumin-mark/FluxUI/style"
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
	ref           *SwitchRef
	decoration    style.Decoration
}

type switchWidget struct {
	value  bool
	config switchConfig
}

func Switch(checked bool, opts ...SwitchOption) Widget {
	cfg := switchConfig{
		width:  50,
		height: 26,
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

// SwitchAttachRef 绑定命令型引用，用于外部主动设置值。
func SwitchAttachRef(ref *SwitchRef) SwitchOption {
	return func(cfg *switchConfig) {
		cfg.ref = ref
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

// SwitchDecoration 通过 Decoration 统一设置开关外层装饰和交互态。
func SwitchDecoration(d style.Decoration) SwitchOption {
	return func(cfg *switchConfig) {
		cfg.decoration = d
	}
}

func (s *switchWidget) Layout(ctx *internal.Context) layout.Dimensions {
	clickable := event.UseClickable(ctx)
	if s.config.ref != nil {
		s.config.ref.bindInvalidator(ctx.Runtime().RequestRedraw)
		for _, cmd := range s.config.ref.drainCommands() {
			if s.config.disabled {
				continue
			}
			next := s.value
			switch cmd.kind {
			case boolCmdSet:
				next = cmd.value
			case boolCmdToggle:
				next = !s.value
			}
			if next != s.value && s.config.onChange != nil {
				s.config.onChange(ctx, next)
			}
		}
	}
	if !s.config.disabled {
		for clickable.Clicked(ctx) {
			next := !s.value
			if s.config.onChange != nil {
				s.config.onChange(ctx, next)
			}
		}
	}

	trackColor := ctx.Theme().SurfaceMuted
	if s.config.hasTrackColor {
		trackColor = s.config.trackColor
	}
	if s.value {
		if s.config.hasColor {
			trackColor = s.config.color
		} else {
			trackColor = ctx.Theme().Primary
		}
	}

	thumbColor := ctx.Theme().Surface
	if s.config.hasThumbColor {
		thumbColor = s.config.thumbColor
	}
	content := func(contentCtx *internal.Context) image.Point {
		return contentCtx.LayoutSwitch(clickable.Handle(), s.value, internal.SwitchSpec{
			Width:      s.config.width,
			Height:     s.config.height,
			TrackColor: trackColor,
			ThumbColor: thumbColor,
			Disabled:   s.config.disabled,
			Hovered:    clickable.Hovered(),
			Pressed:    clickable.Pressed(),
		})
	}

	if hasAnyDecoration(s.config.decoration) {
		deco := withDefaultStates(s.config.decoration,
			style.Decoration{}.WithBg(withAlpha(trackColor, 18)).WithRad(s.config.height/2),
			style.Decoration{}.WithBg(withAlpha(trackColor, 28)).WithRad(s.config.height/2),
			style.Decoration{}.WithBg(withAlpha(ctx.Theme().Disabled, 14)).WithRad(s.config.height/2),
		)
		return layoutDecoratedClickTarget(ctx.Child(0), clickable.Handle(), clickable.Hovered(), clickable.Pressed(), deco, s.config.disabled, content)
	}

	return layout.Dimensions{Size: content(ctx.Child(0))}
}
