package widget

import (
	"image"
	"image/color"

	event "github.com/xiaowumin-mark/FluxUI/event"
	internal "github.com/xiaowumin-mark/FluxUI/internal"
	layout "github.com/xiaowumin-mark/FluxUI/layout"
	style "github.com/xiaowumin-mark/FluxUI/style"

	gioLayout "gioui.org/layout"
)

type CheckboxOption func(*checkboxConfig)

type checkboxConfig struct {
	disabled   bool
	size       float32
	color      color.NRGBA
	hasColor   bool
	onChange   func(ctx *internal.Context, checked bool)
	ref        *CheckboxRef
	decoration style.Decoration
}

type checkboxWidget struct {
	label  string
	value  bool
	config checkboxConfig
}

func Checkbox(label string, checked bool, opts ...CheckboxOption) Widget {
	cfg := checkboxConfig{
		size: 24,
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

// CheckboxDecoration 通过 Decoration 统一设置背景、内边距和圆角。
func CheckboxDecoration(d style.Decoration) CheckboxOption {
	return func(cfg *checkboxConfig) {
		cfg.decoration = d
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

	cs := ctx.Theme().Colors
	checkColor := cs.Primary
	if c.config.hasColor {
		checkColor = c.config.color
	}
	if c.config.disabled {
		checkColor = style.DisabledContent(cs.OnSurface)
	}

	boxWidget := layoutWidgetFunc(func(childCtx *internal.Context) layout.Dimensions {
		size := childCtx.LayoutCheckbox(clickable.Handle(), c.value, internal.CheckboxSpec{
			Size:     c.config.size,
			Color:    checkColor,
			Disabled: c.config.disabled,
			Hovered:  clickable.Hovered(),
			Pressed:  clickable.Pressed(),
		})
		return layout.Dimensions{Size: size}
	})

	labelColor := cs.OnSurface
	if c.config.disabled {
		labelColor = style.DisabledContent(cs.OnSurface)
	} else if clickable.Hovered() {
		labelColor = style.StateLayer(labelColor, checkColor, style.StateLayerHoverOpacity)
	}

	content := func(contentCtx *internal.Context) image.Point {
		if c.label == "" {
			return boxWidget.Layout(contentCtx.Child(0)).Size
		}
		dims := gioLayout.Flex{Axis: gioLayout.Horizontal, Alignment: gioLayout.Middle}.Layout(contentCtx.Gtx,
			gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
				next := *contentCtx
				next.Gtx = gtx
				return gioLayout.Dimensions{Size: boxWidget.Layout(next.Child(0)).Size}
			}),
			gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
				next := *contentCtx
				next.Gtx = gtx
				next.Gtx.Constraints.Min = image.Point{}
				size := next.LayoutInset(internal.Insets{Left: 8}, func(labelCtx *internal.Context) image.Point {
					return labelCtx.LayoutText(internal.TextSpec{
						Content:   c.label,
						Size:      labelCtx.Theme().Types.BodyMedium.Size,
						Color:     labelColor,
						Alignment: internal.AlignStart,
					})
				})
				return gioLayout.Dimensions{Size: size}
			}),
		)
		return dims.Size
	}

	if hasAnyDecoration(c.config.decoration) {
		baseDeco := withDefaultStates(c.config.decoration,
			style.Decoration{}.WithBg(style.StateLayer(color.NRGBA{}, checkColor, style.StateLayerHoverOpacity)),
			style.Decoration{}.WithBg(style.StateLayer(color.NRGBA{}, checkColor, style.StateLayerPressedOpacity)),
			style.Decoration{}.WithBg(style.DisabledContainer(cs.OnSurface)),
		)
		return layoutDecoratedClickTarget(ctx.Child(0), clickable.Handle(), clickable.Hovered(), clickable.Pressed(), baseDeco, c.config.disabled, content)
	}

	return layout.Dimensions{Size: content(ctx.Child(0))}
}
