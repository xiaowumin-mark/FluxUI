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
		width:  52,
		height: 32,
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
		s.config.ref.bindInvalidator(redrawInvalidator(ctx))
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

	cs := ctx.Theme().Colors
	trackColor := cs.SurfaceContainerHighest
	trackBorderColor := cs.Outline
	trackBorderWidth := float32(1)
	if s.config.hasTrackColor {
		trackColor = s.config.trackColor
		trackBorderColor = color.NRGBA{}
		trackBorderWidth = 0
	}
	if s.value {
		if s.config.hasColor {
			trackColor = s.config.color
		} else {
			trackColor = cs.Primary
		}
		trackBorderColor = color.NRGBA{}
		trackBorderWidth = 0
	}

	thumbColor := cs.Outline
	if s.value {
		thumbColor = cs.OnPrimary
	}
	if s.config.hasThumbColor {
		thumbColor = s.config.thumbColor
	}
	if s.config.disabled {
		trackColor = style.DisabledContainer(cs.OnSurface)
		trackBorderColor = color.NRGBA{}
		trackBorderWidth = 0
		thumbColor = style.DisabledContent(cs.OnSurface)
	}
	checkedProgress := md3SelectionProgress(ctx, s.value)
	trackColor = md3AnimateColor(ctx, "switch-track", trackColor, style.InteractionSelectedDuration, style.InteractionStandardEasing)
	thumbColor = md3AnimateColor(ctx, "switch-thumb", thumbColor, style.InteractionSelectedDuration, style.InteractionStandardEasing)
	content := func(contentCtx *internal.Context) image.Point {
		return contentCtx.LayoutSwitch(nil, s.value, internal.SwitchSpec{
			Width:            s.config.width,
			Height:           s.config.height,
			TrackColor:       trackColor,
			TrackBorderColor: trackBorderColor,
			TrackBorderWidth: trackBorderWidth,
			ThumbColor:       thumbColor,
			CheckedProgress:  checkedProgress,
			Disabled:         s.config.disabled,
			Hovered:          clickable.Hovered(),
			Pressed:          clickable.Pressed(),
		})
	}

	deco := withDefaultStates(s.config.decoration,
		style.Decoration{}.WithBg(style.StateLayer(color.NRGBA{}, trackColor, style.StateLayerHoverOpacity)).WithRad(s.config.height/2),
		style.Decoration{}.WithBg(style.StateLayer(color.NRGBA{}, trackColor, style.StateLayerPressedOpacity)).WithRad(s.config.height/2),
		style.Decoration{}.WithBg(style.DisabledContainer(cs.OnSurface)).WithRad(s.config.height/2),
	)
	target := func(targetCtx *internal.Context) image.Point {
		if hasAnyDecoration(s.config.decoration) {
			active := resolveDecorationState(deco, clickable.Hovered(), clickable.Pressed(), s.config.disabled)
			duration, easing := md3InteractionTiming(clickable.Hovered(), clickable.Pressed(), false, s.config.disabled)
			visual := md3AnimateDecoration(targetCtx, "switch-decoration", stripStateDecoration(active), duration, easing)
			return layoutDecorationShell(targetCtx.Child(0), visual, content).Size
		}
		return content(targetCtx.Child(0))
	}
	stateLayerColor := trackColor
	if !s.value {
		stateLayerColor = thumbColor
	}
	return layoutStateLayerTouchTarget(ctx, clickable.Handle(), s.config.disabled, internal.RippleSpec{
		Color:   stateLayerColor,
		Radius:  s.config.height / 2,
		Opacity: style.StateLayerPressedOpacity,
	}, 48, 40, func(size image.Point) image.Point {
		trackHeight := size.Y
		if trackHeight <= 0 {
			trackHeight = ctx.Gtx.Dp(safeDp(s.config.height))
		}
		thumbSize := trackHeight / 2
		if s.value {
			thumbSize = trackHeight - ctx.Gtx.Dp(safeDp(8))
		}
		if clickable.Pressed() {
			pressedSize := ctx.Gtx.Dp(safeDp(28))
			if pressedSize > thumbSize {
				thumbSize = pressedSize
			}
		}
		maxThumb := trackHeight - ctx.Gtx.Dp(safeDp(4))
		if thumbSize > maxThumb {
			thumbSize = maxThumb
		}
		left := ctx.Gtx.Dp(safeDp(4))
		x := left + thumbSize/2
		if s.value {
			x = size.X - left - thumbSize/2
		}
		return image.Pt(x, size.Y/2)
	}, func() float32 {
		return materialStateLayerOpacity(clickable.Hovered(), clickable.Pressed())
	}, target)
}
