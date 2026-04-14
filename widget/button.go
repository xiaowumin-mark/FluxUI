package widget

import (
	"image"
	"image/color"

	"fluxui/event"
	"fluxui/internal"
	"fluxui/layout"
	"fluxui/style"
)

// ButtonOption 定义按钮配置项。
type ButtonOption func(*buttonConfig)

type buttonConfig struct {
	dispatcher    event.Dispatcher
	disabled      bool
	padding       style.Insets
	radius        float32
	background    color.NRGBA
	foreground    color.NRGBA
	hasBackground bool
	hasForeground bool
}

type buttonWidget struct {
	child  Widget
	config buttonConfig
}

// Button 创建按钮组件。
func Button(child Widget, opts ...ButtonOption) Widget {
	cfg := buttonConfig{
		padding: style.Symmetric(10, 14),
		radius:  8,
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
	}
}

// ButtonRadius 设置按钮圆角。
func ButtonRadius(radius float32) ButtonOption {
	return func(cfg *buttonConfig) {
		cfg.radius = radius
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

func (b *buttonWidget) Layout(ctx *internal.Context) layout.Dimensions {
	clickable := event.UseClickable(ctx)
	if !b.config.disabled {
		for clickable.Clicked(ctx) {
			b.config.dispatcher.DispatchClick(ctx)
		}
		if changed, hovering := clickable.HoverChanged(); changed {
			b.config.dispatcher.DispatchHover(ctx, hovering)
		}
	}

	background := ctx.Theme().Primary
	if b.config.hasBackground {
		background = b.config.background
	}
	if b.config.disabled {
		background = ctx.Theme().Disabled
	}

	foreground := ctx.Theme().TextOnPrimary
	if b.config.hasForeground {
		foreground = b.config.foreground
	}

	size := ctx.LayoutButton(clickable.Handle(), internal.ButtonSpec{
		Background: background,
		Foreground: foreground,
		Radius:     b.config.radius,
		Padding:    toInternalInsets(b.config.padding),
		Disabled:   b.config.disabled,
	}, func(childCtx *internal.Context) image.Point {
		if b.child == nil {
			return image.Point{}
		}
		return b.child.Layout(childCtx.Child(0)).Size
	})

	return layout.Dimensions{Size: size}
}

func toInternalInsets(insets style.Insets) internal.Insets {
	return internal.Insets{
		Top:    insets.Top,
		Right:  insets.Right,
		Bottom: insets.Bottom,
		Left:   insets.Left,
	}
}
