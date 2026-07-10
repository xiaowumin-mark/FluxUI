package widget

import (
	"image"
	"image/color"

	event "github.com/xiaowumin-mark/FluxUI/event"
	internal "github.com/xiaowumin-mark/FluxUI/internal"
	layout "github.com/xiaowumin-mark/FluxUI/layout"
	style "github.com/xiaowumin-mark/FluxUI/style"
	theme "github.com/xiaowumin-mark/FluxUI/theme"
)

// ButtonOption 定义按钮配置项。
type ButtonOption func(*buttonConfig)

type buttonVariant int

const (
	buttonVariantFilled buttonVariant = iota
	buttonVariantFilledTonal
	buttonVariantOutlined
	buttonVariantText
	buttonVariantElevated
)

type buttonConfig struct {
	dispatcher    event.Dispatcher
	variant       buttonVariant
	disabled      bool
	padding       style.Insets
	hasPadding    bool
	radius        float32
	hasRadius     bool
	background    color.NRGBA
	foreground    color.NRGBA
	hasBackground bool
	hasForeground bool
	ref           *ButtonRef
	decoration    style.Decoration
	loading       bool
	loadingIcon   Widget
}

type buttonWidget struct {
	child  Widget
	config buttonConfig
}

// Button 创建按钮组件。
func Button(child Widget, opts ...ButtonOption) Widget {
	return newButton(buttonVariantFilled, child, opts...)
}

func FilledButton(child Widget, opts ...ButtonOption) Widget {
	return newButton(buttonVariantFilled, child, opts...)
}

func FilledTonalButton(child Widget, opts ...ButtonOption) Widget {
	return newButton(buttonVariantFilledTonal, child, opts...)
}

func OutlinedButton(child Widget, opts ...ButtonOption) Widget {
	return newButton(buttonVariantOutlined, child, opts...)
}

func TextButton(child Widget, opts ...ButtonOption) Widget {
	return newButton(buttonVariantText, child, opts...)
}

func ElevatedButton(child Widget, opts ...ButtonOption) Widget {
	return newButton(buttonVariantElevated, child, opts...)
}

func newButton(variant buttonVariant, child Widget, opts ...ButtonOption) Widget {
	cfg := buttonConfig{
		variant: variant,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &buttonWidget{
		child:  child,
		config: cfg,
	}
}

// OnClick 绑定点击回调。
func OnClick(fn event.ClickHandler) ButtonOption {
	return func(cfg *buttonConfig) {
		cfg.dispatcher.Click = fn
	}
}

// OnHover 绑定悬浮变化回调。
func OnHover(fn event.HoverHandler) ButtonOption {
	return func(cfg *buttonConfig) {
		cfg.dispatcher.Hover = fn
	}
}

// Disabled 设置按钮禁用状态。
func Disabled(disabled bool) ButtonOption {
	return func(cfg *buttonConfig) {
		cfg.disabled = disabled
	}
}

// ButtonPadding 设置按钮内边距。
func ButtonPadding(insets style.Insets) ButtonOption {
	return func(cfg *buttonConfig) {
		cfg.padding = insets
		cfg.hasPadding = true
	}
}

// ButtonRadius 设置按钮圆角。
func ButtonRadius(radius float32) ButtonOption {
	return func(cfg *buttonConfig) {
		cfg.radius = radius
		cfg.hasRadius = true
	}
}

// ButtonBackground 设置按钮背景色。
func ButtonBackground(value color.NRGBA) ButtonOption {
	return func(cfg *buttonConfig) {
		cfg.background = value
		cfg.hasBackground = true
	}
}

// ButtonForeground 设置按钮前景色。
func ButtonForeground(value color.NRGBA) ButtonOption {
	return func(cfg *buttonConfig) {
		cfg.foreground = value
		cfg.hasForeground = true
	}
}

// ButtonDecoration 通过 Decoration 统一设置背景、内边距和圆角。
func ButtonDecoration(d style.Decoration) ButtonOption {
	return func(cfg *buttonConfig) {
		cfg.decoration = d
	}
}

// ButtonAttachRef 绑定命令型引用，用于外部主动触发点击。
func ButtonAttachRef(ref *ButtonRef) ButtonOption {
	return func(cfg *buttonConfig) {
		cfg.ref = ref
	}
}

// ButtonLoading 显示按钮内 loading indicator，并暂时禁用点击分发。
func ButtonLoading(loading bool) ButtonOption {
	return func(cfg *buttonConfig) {
		cfg.loading = loading
	}
}

// ButtonLoadingIndicator 设置 loading 状态下显示的指示器。
func ButtonLoadingIndicator(indicator Widget) ButtonOption {
	return func(cfg *buttonConfig) {
		cfg.loadingIcon = indicator
	}
}

func (b *buttonWidget) Layout(ctx *internal.Context) layout.Dimensions {
	clickable := event.UseClickable(ctx)
	disabled := b.config.disabled || b.config.loading
	registerClickableFocusAction(ctx, disabled, b.config.dispatcher.Click)
	if b.config.ref != nil {
		bindCommandRef(ctx, "button", b.config.ref, &b.config.ref.queue)
		commands := b.config.ref.drainCommands()
		if !disabled {
			for range commands {
				runClickableDefaultAction(ctx, nil, false, b.config.dispatcher.Click)
			}
		}
	}
	interaction := event.InteractionSnapshot{}
	if !disabled {
		drainClickableDefaultAction(ctx, clickable, false, b.config.dispatcher.Click)
		interaction = clickableInteractionSnapshot(ctx, clickable, false, true)
		if changed, hovering := clickable.HoverChangedWithSnapshot(ctx, interaction.Hovered); changed {
			b.config.dispatcher.DispatchHover(ctx, hovering)
		}
	}

	th := ctx.Theme()
	cs := th.Colors
	spec := resolveButtonDefaults(b.config.variant, th)
	hovered := interaction.Hovered
	pressed := interaction.Pressed
	focused := interaction.Focused
	activeDecoration := resolveDecorationState(b.config.decoration, hovered, pressed, b.config.disabled)

	backgroundDefault := spec.background
	if b.config.hasBackground {
		backgroundDefault = b.config.background
	}
	background := activeDecoration.ResolveBg(backgroundDefault)
	foreground := spec.foreground
	if b.config.hasForeground {
		foreground = b.config.foreground
	}

	duration, easing := md3InteractionTiming(ctx, hovered, pressed, focused, b.config.disabled)
	if !b.config.disabled && (hovered || pressed) {
		opacity := style.StateLayerHoverOpacity
		if pressed {
			opacity = style.StateLayerPressedOpacity
		}
		opacity = md3AnimateStateLayerOpacity(ctx, "button-state-layer", opacity)
		background = style.StateLayer(background, foreground, opacity)
	}
	if b.config.disabled {
		if b.config.decoration.Disabled == nil || b.config.decoration.Disabled.Background == nil {
			if spec.disabledContainer {
				background = style.DisabledContainer(cs.OnSurface)
			} else {
				background = color.NRGBA{}
			}
		}
		if !b.config.hasForeground {
			foreground = style.DisabledContent(cs.OnSurface)
		}
	}
	animateInteractionColors := !b.config.disabled && (hovered || pressed || focused)
	if !b.config.disabled {
		background = md3AnimateColorIfActive(ctx, "button-background", background, animateInteractionColors, duration, easing)
		foreground = md3AnimateColorIfActive(ctx, "button-foreground", foreground, animateInteractionColors, duration, easing)
	}
	radiusDefault := spec.radius
	if b.config.hasRadius {
		radiusDefault = b.config.radius
	}
	radius := activeDecoration.ResolveRad(radiusDefault)
	paddingDefault := densityInsets(ctx, style.Symmetric(10, 24), style.Symmetric(6, 16))
	if b.config.hasPadding {
		paddingDefault = b.config.padding
	}
	padding := activeDecoration.ResolvePad(paddingDefault)
	borderColor := spec.border.Color
	borderWidth := spec.border.Width
	focusOpacity := float32(0)
	if !b.config.disabled {
		borderColor = md3AnimateColorIfActive(ctx, "button-border-color", borderColor, animateInteractionColors, duration, easing)
		borderWidth = md3AnimateFloat(ctx, "button-border-width", borderWidth, duration, easing)
		focusOpacity = md3FocusProgress(ctx, focused, false)
	}

	size := ctx.LayoutButton(clickable.Handle(), internal.ButtonSpec{
		Background:  background,
		Foreground:  foreground,
		Radius:      radius,
		CornerShape: md3PillCornerShape(activeDecoration),
		Padding:     toInternalInsets(padding),
		TextStyle:   spec.text,
		BorderColor: borderColor,
		BorderWidth: borderWidth,
		HasShadow:   !spec.shadow.IsZero(),
		Shadow: internal.ShadowSpec{
			OffsetX: spec.shadow.OffsetX,
			OffsetY: spec.shadow.OffsetY,
			Blur:    spec.shadow.Blur,
			Spread:  spec.shadow.Spread,
			Color:   spec.shadow.Color,
			Layers:  spec.shadow.EffectiveLayers(),
		},
		FocusOpacity: focusOpacity,
		Disabled:     b.config.disabled,
	}, func(childCtx *internal.Context) image.Point {
		child := b.child
		if b.config.loading {
			indicator := b.config.loadingIcon
			if indicator == nil {
				indicator = LoadingIndicator(ProgressSize(18), ProgressThickness(3), ProgressFillColor(foreground), ProgressTrackColor(color.NRGBA{}))
			}
			child = Row(indicator, Padding(style.Insets{Left: 8}, child))
		}
		if child == nil {
			return image.Point{}
		}
		return child.Layout(childCtx.Child(0)).Size
	})

	return layout.Dimensions{Size: size}
}

type buttonDefaults struct {
	background        color.NRGBA
	foreground        color.NRGBA
	radius            float32
	border            style.Border
	shadow            style.BoxShadow
	text              theme.TextStyle
	disabledContainer bool
}

func resolveButtonDefaults(variant buttonVariant, th *theme.Theme) buttonDefaults {
	if th == nil {
		th = theme.Default()
	}
	cs := th.Colors
	defaults := buttonDefaults{
		background:        cs.Primary,
		foreground:        cs.OnPrimary,
		radius:            th.Shapes.Full,
		text:              th.Types.LabelLarge,
		disabledContainer: true,
	}
	switch variant {
	case buttonVariantFilledTonal:
		defaults.background = cs.SecondaryContainer
		defaults.foreground = cs.OnSecondaryContainer
	case buttonVariantOutlined, buttonVariantText:
		defaults.background = color.NRGBA{}
		defaults.foreground = cs.Primary
		defaults.disabledContainer = false
		if variant == buttonVariantOutlined {
			defaults.border = style.Border{Width: 1, Color: cs.Outline}
		}
	case buttonVariantElevated:
		defaults.background = style.SurfaceAtElevation(cs, 1)
		defaults.foreground = cs.Primary
		defaults.shadow = style.ElevationShadow(cs, 1)
	}
	return defaults
}
