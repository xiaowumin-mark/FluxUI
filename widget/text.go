package widget

import (
	"image/color"
	"strings"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
	layout "github.com/xiaowumin-mark/FluxUI/layout"
	theme "github.com/xiaowumin-mark/FluxUI/theme"
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
	size      float32
	color     color.NRGBA
	hasColor  bool
	align     TextAlignment
	font      theme.FontSpec
	hasFamily bool
	hasStyle  bool
	hasWeight bool
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

// TextFont 设置文本字体（局部覆盖）。
func TextFont(spec theme.FontSpec) TextOption {
	return func(cfg *textConfig) {
		cfg.font = spec
		cfg.hasStyle = true
		cfg.hasWeight = true
		if strings.TrimSpace(spec.Family) != "" {
			cfg.hasFamily = true
		}
	}
}

// TextFontFamily 设置文本字体族（局部覆盖）。
func TextFontFamily(family string) TextOption {
	return func(cfg *textConfig) {
		cfg.font.Family = strings.TrimSpace(family)
		cfg.hasFamily = true
	}
}

// TextFontStyle 设置文本字体样式（局部覆盖）。
func TextFontStyle(style theme.FontStyle) TextOption {
	return func(cfg *textConfig) {
		cfg.font.Style = style
		cfg.hasStyle = true
	}
}

// TextFontWeight 设置文本字体字重（局部覆盖）。
func TextFontWeight(weight theme.FontWeight) TextOption {
	return func(cfg *textConfig) {
		cfg.font.Weight = weight
		cfg.hasWeight = true
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
	font := ctx.Font()
	if t.config.hasFamily && strings.TrimSpace(t.config.font.Family) != "" {
		font.Family = strings.TrimSpace(t.config.font.Family)
	}
	if t.config.hasStyle {
		font.Style = t.config.font.Style
	}
	if t.config.hasWeight {
		font.Weight = t.config.font.Weight
	}
	font = font.Normalize()

	return layout.Dimensions{
		Size: ctx.LayoutText(internal.TextSpec{
			Content:   t.content,
			Size:      size,
			Color:     color,
			Alignment: toInternalAlignment(t.config.align),
			Font:      font,
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
