package widget

import (
	"image"
	"image/color"
	"strings"

	event "github.com/xiaowumin-mark/FluxUI/event"
	internal "github.com/xiaowumin-mark/FluxUI/internal"
	layout "github.com/xiaowumin-mark/FluxUI/layout"
	style "github.com/xiaowumin-mark/FluxUI/style"
)

type SwitchOption func(*switchConfig)

type switchConfig struct {
	disabled       bool
	width          float32
	height         float32
	color          color.NRGBA
	trackColor     color.NRGBA
	thumbColor     color.NRGBA
	hasColor       bool
	hasTrackColor  bool
	hasThumbColor  bool
	onChange       func(ctx *internal.Context, checked bool)
	ref            *SwitchRef
	decoration     style.Decoration
	checkedIcon    string
	uncheckedIcon  string
	iconFontID     string
	iconFontFamily string
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

// SwitchIcons sets optional thumb icons for checked and unchecked states.
func SwitchIcons(checkedIcon, uncheckedIcon string) SwitchOption {
	return func(cfg *switchConfig) {
		cfg.checkedIcon = strings.TrimSpace(checkedIcon)
		cfg.uncheckedIcon = strings.TrimSpace(uncheckedIcon)
	}
}

// SwitchCheckedIcon sets the thumb icon used when the switch is checked.
func SwitchCheckedIcon(name string) SwitchOption {
	return func(cfg *switchConfig) {
		cfg.checkedIcon = strings.TrimSpace(name)
	}
}

// SwitchUncheckedIcon sets the thumb icon used when the switch is unchecked.
func SwitchUncheckedIcon(name string) SwitchOption {
	return func(cfg *switchConfig) {
		cfg.uncheckedIcon = strings.TrimSpace(name)
	}
}

// SwitchIconUseFont renders switch thumb icons through a registered icon font ID.
func SwitchIconUseFont(id string) SwitchOption {
	return func(cfg *switchConfig) {
		cfg.iconFontID = strings.TrimSpace(id)
	}
}

// SwitchIconFontFamily directly selects the icon font family for switch thumb icons.
func SwitchIconFontFamily(family string) SwitchOption {
	return func(cfg *switchConfig) {
		cfg.iconFontFamily = strings.TrimSpace(family)
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
	interaction := event.InteractionSnapshot{}
	if !s.config.disabled {
		for clickable.Clicked(ctx) {
			next := !s.value
			if s.config.onChange != nil {
				s.config.onChange(ctx, next)
			}
		}
		interaction = clickable.Snapshot(ctx, false)
	}

	cs := ctx.Theme().Colors
	trackColor := cs.SurfaceContainerHighest
	trackBorderColor := cs.Outline
	trackBorderWidth := float32(2)
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
	duration, easing := md3InteractionTiming(ctx, interaction.Hovered, interaction.Pressed, false, s.config.disabled)
	if s.config.disabled {
		duration = 0
	}
	checkedProgress := boolProgress(s.value)
	pressedProgress := float32(0)
	if !s.config.disabled {
		checkedProgress = md3SwitchSelectionProgress(ctx, s.value)
		pressedProgress = md3SwitchPressedProgress(ctx, interaction.Pressed)
	}
	iconText, iconFamily := s.resolveThumbIcon(ctx)
	iconColor := cs.OnPrimaryContainer
	if s.value {
		iconColor = cs.Primary
	}
	if s.config.disabled {
		iconColor = style.DisabledContent(cs.OnSurface)
	}
	colorDuration := style.InteractionSelectedDuration
	colorEasing := style.InteractionStandardEasing
	if s.config.disabled {
		colorDuration = 0
	}
	if !s.config.disabled && !s.config.hasThumbColor {
		if interaction.Hovered || interaction.Pressed {
			if s.value {
				thumbColor = cs.PrimaryContainer
			} else {
				thumbColor = cs.OnSurfaceVariant
			}
		}
	}
	trackColor = md3AnimateColorRetained(ctx, "switch-track", trackColor, colorDuration, colorEasing)
	thumbColor = md3AnimateColorRetained(ctx, "switch-thumb", thumbColor, colorDuration, colorEasing)
	content := func(contentCtx *internal.Context) image.Point {
		return contentCtx.LayoutSwitch(nil, s.value, internal.SwitchSpec{
			Width:               s.config.width,
			Height:              s.config.height,
			TrackColor:          trackColor,
			TrackBorderColor:    trackBorderColor,
			TrackBorderWidth:    trackBorderWidth,
			ThumbColor:          thumbColor,
			CheckedProgress:     checkedProgress,
			PositionProgress:    checkedProgress,
			PressedProgress:     pressedProgress,
			ThumbIcon:           iconText,
			ThumbIconFontFamily: iconFamily,
			ThumbIconColor:      iconColor,
			Disabled:            s.config.disabled,
			Hovered:             interaction.Hovered,
			Pressed:             interaction.Pressed,
		})
	}

	deco := withDefaultStates(s.config.decoration,
		style.Decoration{}.WithBg(style.StateLayer(color.NRGBA{}, trackColor, style.StateLayerHoverOpacity)).WithRad(s.config.height/2),
		style.Decoration{}.WithBg(style.StateLayer(color.NRGBA{}, trackColor, style.StateLayerPressedOpacity)).WithRad(s.config.height/2),
		style.Decoration{}.WithBg(style.DisabledContainer(cs.OnSurface)).WithRad(s.config.height/2),
	)
	target := func(targetCtx *internal.Context) image.Point {
		if hasAnyDecoration(s.config.decoration) {
			active := resolveDecorationState(deco, interaction.Hovered, interaction.Pressed, s.config.disabled)
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
		return switchThumbCenter(ctx, size, checkedProgress)
	}, func() float32 {
		return materialStateLayerOpacity(ctx, interaction.Hovered, interaction.Pressed)
	}, target)
}

func (s *switchWidget) resolveThumbIcon(ctx *internal.Context) (text string, family string) {
	name := s.config.uncheckedIcon
	if s.value {
		name = s.config.checkedIcon
	}
	name = strings.TrimSpace(name)
	if name == "" || ctx == nil {
		return "", ""
	}
	name = iconLigatureName(name)
	if family := strings.TrimSpace(s.config.iconFontFamily); family != "" {
		return name, family
	}
	th := ctx.Theme()
	var fontOK bool
	var fontFamily string
	if id := strings.TrimSpace(s.config.iconFontID); id != "" {
		if font, ok := th.ResolveIconFont(id); ok {
			fontOK = true
			fontFamily = font.Family
			if resolved, ok := font.ResolveIconText(name); ok {
				name = resolved
			}
		}
	} else if font, ok := th.DefaultIconFont(); ok {
		fontOK = true
		fontFamily = font.Family
		if resolved, ok := font.ResolveIconText(name); ok {
			name = resolved
		}
	}
	if !fontOK {
		return "", ""
	}
	return name, fontFamily
}
