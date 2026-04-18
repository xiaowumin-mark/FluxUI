package theme

import "image/color"

// Theme 定义 FluxUI 的基础主题令牌。
type Theme struct {
	Primary        color.NRGBA
	Surface        color.NRGBA
	SurfaceMuted   color.NRGBA
	TextColor      color.NRGBA
	TextOnPrimary  color.NRGBA
	Disabled       color.NRGBA
	TextSize       float32
	DefaultFont    FontSpec
	UseSystemFonts bool
	Fonts          []FontFace
}

// Default 返回默认主题。
func Default() *Theme {
	return &Theme{
		Primary:        color.NRGBA{R: 49, G: 107, B: 255, A: 255},
		Surface:        color.NRGBA{R: 248, G: 250, B: 252, A: 255},
		SurfaceMuted:   color.NRGBA{R: 226, G: 232, B: 240, A: 255},
		TextColor:      color.NRGBA{R: 15, G: 23, B: 42, A: 255},
		TextOnPrimary:  color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Disabled:       color.NRGBA{R: 148, G: 163, B: 184, A: 255},
		TextSize:       16,
		DefaultFont:    DefaultFontSpec(),
		UseSystemFonts: true,
	}
}
