package widget

import (
	"image/color"

	"fluxui/internal"
	"fluxui/layout"
)

// TextAlignment 表示文本对齐。
type TextAlignment int

const (
	AlignStart TextAlignment = iota
	AlignCenter
	AlignEnd
)

// TextOption 定义文本配置项。
type TextOption func(*textConfig)

type textConfig struct {
	size     float32
	color    color.NRGBA
	hasColor bool
	align    TextAlignment
}

type textWidget struct {
	content string
	config  textConfig
}

// Text 创建文本组件。
func Text(content string, opts ...TextOption) Widget {
	cfg := textConfig{align: AlignStart}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &textWidget{
		content: content,
		config:  cfg,
	}
}

// TextSize 设置字号。
func TextSize(size float32) TextOption {
	return func(cfg *textConfig) {
		cfg.size = size
	}
}

// TextColor 设置文本颜色。
func TextColor(value color.NRGBA) TextOption {
	return func(cfg *textConfig) {
		cfg.color = value
		cfg.hasColor = true
	}
}

// TextAlign 设置文本对齐。
func TextAlign(alignment TextAlignment) TextOption {
	return func(cfg *textConfig) {
		cfg.align = alignment
	}
}

func (t *textWidget) Layout(ctx *internal.Context) layout.Dimensions {
	color := ctx.Foreground()
	if t.config.hasColor {
		color = t.config.color
	}

	size := t.config.size
	if size <= 0 {
		size = ctx.Theme().TextSize
	}

	return layout.Dimensions{
		Size: ctx.LayoutText(internal.TextSpec{
			Content:   t.content,
			Size:      size,
			Color:     color,
			Alignment: toInternalAlignment(t.config.align),
		}),
	}
}

func toInternalAlignment(alignment TextAlignment) internal.Alignment {
	switch alignment {
	case AlignCenter:
		return internal.AlignCenter
	case AlignEnd:
		return internal.AlignEnd
	default:
		return internal.AlignStart
	}
}
