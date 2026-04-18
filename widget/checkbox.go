package widget

import (
	"image"
	"image/color"

	"fluxui/event"
	"fluxui/internal"
	"fluxui/layout"
	"fluxui/style"

	gioLayout "gioui.org/layout"
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
	ref           *CheckboxRef
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

// CheckboxAttachRef 绑定命令型引用，用于外部主动设置值。
func CheckboxAttachRef(ref *CheckboxRef) CheckboxOption {
	return func(cfg *checkboxConfig) {
		cfg.ref = ref
	}
}

func (c *checkboxWidget) Layout(ctx *internal.Context) layout.Dimensions {
	clickable := event.UseClickable(ctx)
	if c.config.ref != nil {
		c.config.ref.bindInvalidator(ctx.Runtime().RequestRedraw)
		for _, cmd := range c.config.ref.drainCommands() {
			if c.config.disabled {
				continue
			}
			next := c.value
			switch cmd.kind {
			case boolCmdSet:
				next = cmd.value
			case boolCmdToggle:
				next = !c.value
			}
			if next != c.value && c.config.onChange != nil {
				c.config.onChange(ctx, next)
			}
		}
	}
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

	boxWidget := layoutWidgetFunc(func(childCtx *internal.Context) layout.Dimensions {
		size := childCtx.LayoutCheckbox(clickable.Handle(), c.value, internal.CheckboxSpec{
			Size:     c.config.size,
			Color:    checkColor,
			Disabled: c.config.disabled,
		})
		return layout.Dimensions{Size: size}
	})

	if c.label == "" {
		return boxWidget.Layout(ctx.Child(0))
	}

	labelColor := ctx.Theme().TextColor
	if c.config.disabled {
		labelColor = ctx.Theme().Disabled
	}

	dims := gioLayout.Flex{Axis: gioLayout.Horizontal, Alignment: gioLayout.Middle}.Layout(ctx.Gtx,
		gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
			next := *ctx
			next.Gtx = gtx
			return gioLayout.Dimensions{Size: boxWidget.Layout(next.Child(0)).Size}
		}),
		gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
			next := *ctx
			next.Gtx = gtx
			next.Gtx.Constraints.Min = image.Point{}
			size := next.LayoutInset(internal.Insets{Left: 8}, func(contentCtx *internal.Context) image.Point {
				return contentCtx.LayoutText(internal.TextSpec{
					Content:   c.label,
					Size:      contentCtx.Theme().TextSize,
					Color:     labelColor,
					Alignment: internal.AlignStart,
				})
			})
			return gioLayout.Dimensions{Size: size}
		}),
	)
	return layout.Dimensions{Size: dims.Size}
}
