package widget

import (
	"image"
	"image/color"

	"fluxui/event"
	"fluxui/internal"
	"fluxui/layout"
	"fluxui/style"
)

type CheckboxOption func(*checkboxConfig)

type checkboxConfig struct {
	dispatcher    event.Dispatcher
	disabled      bool
	padding       style.Insets
	size          float32
	color         color.NRGBA
	background    color.NRGBA
	hasColor      bool
	hasBackground bool
	onChange      func(ctx *internal.Context, checked bool)
}

type checkboxWidget struct {
	label  string
	value  bool
	config checkboxConfig
}

func Checkbox(label string, checked bool, opts ...CheckboxOption) Widget {
	cfg := checkboxConfig{
		padding: style.Symmetric(8, 8),
		size:    24,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &checkboxWidget{
		label:  label,
		value:  checked,
		config: cfg,
	}
}

func CheckboxOnChange(fn func(ctx *internal.Context, checked bool)) CheckboxOption {
	return func(cfg *checkboxConfig) {
		cfg.onChange = fn
	}
}

func CheckboxDisabled(disabled bool) CheckboxOption {
	return func(cfg *checkboxConfig) {
		cfg.disabled = disabled
	}
}

func CheckboxSize(size float32) CheckboxOption {
	return func(cfg *checkboxConfig) {
		cfg.size = size
	}
}

func CheckboxColor(color color.NRGBA) CheckboxOption {
	return func(cfg *checkboxConfig) {
		cfg.color = color
		cfg.hasColor = true
	}
}

func (c *checkboxWidget) Layout(ctx *internal.Context) layout.Dimensions {
	clickable := event.UseClickable(ctx)
	if !c.config.disabled {
		for clickable.Clicked(ctx) {
			next := !c.value
			if c.config.onChange != nil {
				c.config.onChange(ctx, next)
			}
		}
	}

	checkColor := ctx.Theme().Primary
	if c.config.hasColor {
		checkColor = c.config.color
	}

	box := layout.Rigid(func(childCtx *internal.Context) layout.Dimensions {
		size := childCtx.LayoutCheckbox(clickable.Handle(), c.value, internal.CheckboxSpec{
			Size:     c.config.size,
			Color:    checkColor,
			Disabled: c.config.disabled,
		})
		return layout.Dimensions{Size: size}
	})

	if c.label == "" {
		return layout.Flex(ctx, layout.Horizontal, box)
	}

	labelColor := ctx.Theme().TextColor
	if c.config.disabled {
		labelColor = ctx.Theme().Disabled
	}

	label := layout.Rigid(func(childCtx *internal.Context) layout.Dimensions {
		size := childCtx.LayoutInset(internal.Insets{Left: 8}, func(contentCtx *internal.Context) image.Point {
			return contentCtx.LayoutText(internal.TextSpec{
				Content:   c.label,
				Size:      contentCtx.Theme().TextSize,
				Color:     labelColor,
				Alignment: internal.AlignStart,
			})
		})
		return layout.Dimensions{Size: size}
	})

	return layout.Flex(ctx, layout.Horizontal, box, label)
}
