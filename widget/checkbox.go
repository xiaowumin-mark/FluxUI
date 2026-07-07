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
		size: 18,
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
	activate := func(actionCtx *internal.Context) {
		next := !c.value
		if c.config.onChange != nil {
			c.config.onChange(actionCtx, next)
		}
	}
	registerClickableFocusAction(ctx, c.config.disabled, activate)
	if c.config.ref != nil {
		c.config.ref.bindInvalidator(redrawInvalidator(ctx))
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
	interaction := event.InteractionSnapshot{}
	if !c.config.disabled {
		drainClickableDefaultAction(ctx, clickable, false, activate)
		interaction = clickableInteractionSnapshot(ctx, clickable, false, true)
	}

	cs := ctx.Theme().Colors
	checkColor := cs.Primary
	if c.config.hasColor {
		checkColor = c.config.color
	}
	if c.config.disabled {
		checkColor = style.DisabledContent(cs.OnSurface)
	}
	duration, easing := md3InteractionTiming(ctx, interaction.Hovered, interaction.Pressed, false, c.config.disabled)
	if c.config.disabled {
		duration = 0
	}
	checkColor = md3AnimateColor(ctx, "checkbox-color", checkColor, duration, easing)

	boxWidget := layoutWidgetFunc(func(childCtx *internal.Context) layout.Dimensions {
		checkedProgress := md3SelectionProgress(childCtx, c.value)
		size := childCtx.LayoutCheckbox(nil, c.value, internal.CheckboxSpec{
			Size:            c.config.size,
			Color:           checkColor,
			CheckedProgress: checkedProgress,
			Disabled:        c.config.disabled,
			Hovered:         interaction.Hovered,
			Pressed:         interaction.Pressed,
		})
		return layout.Dimensions{Size: size}
	})

	labelColor := cs.OnSurface
	if c.config.disabled {
		labelColor = style.DisabledContent(cs.OnSurface)
	} else if interaction.Hovered {
		labelColor = style.StateLayer(labelColor, checkColor, style.StateLayerHoverOpacity)
	}
	labelColor = md3AnimateColor(ctx, "checkbox-label", labelColor, duration, easing)

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
					return Text(c.label, TextType(labelCtx.Theme().Types.BodyMedium), TextColor(labelColor)).Layout(labelCtx.Child(0)).Size
				})
				return gioLayout.Dimensions{Size: size}
			}),
		)
		return dims.Size
	}

	baseDeco := withDefaultStates(c.config.decoration,
		style.Decoration{}.WithBg(style.StateLayer(color.NRGBA{}, checkColor, style.StateLayerHoverOpacity)),
		style.Decoration{}.WithBg(style.StateLayer(color.NRGBA{}, checkColor, style.StateLayerPressedOpacity)),
		style.Decoration{}.WithBg(style.DisabledContainer(cs.OnSurface)),
	)
	target := func(targetCtx *internal.Context) image.Point {
		if hasAnyDecoration(c.config.decoration) {
			active := resolveDecorationState(baseDeco, interaction.Hovered, interaction.Pressed, c.config.disabled)
			visual := md3AnimateDecoration(targetCtx, "checkbox-decoration", stripStateDecoration(active), duration, easing)
			return layoutDecorationShell(targetCtx.Child(0), visual, content).Size
		}
		return content(targetCtx.Child(0))
	}
	return layoutStateLayerTouchTarget(ctx, clickable.Handle(), c.config.disabled, internal.RippleSpec{
		Color:   checkColor,
		Radius:  20,
		Opacity: style.StateLayerPressedOpacity,
	}, 48, 40, func(size image.Point) image.Point {
		controlSize := selectionControlSizePx(ctx, c.config.size, 18)
		return image.Pt(controlSize/2, size.Y/2)
	}, func() float32 {
		return materialStateLayerOpacity(ctx, interaction.Hovered, interaction.Pressed)
	}, target)
}
