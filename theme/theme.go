package theme

import "image/color"

type Theme struct {
	Colors  ColorScheme
	Shapes  ShapeScale
	Types   TypeScale
	Density DensityScale

	// Backward-compatible flat fields, synced from Colors by Default() / DarkTheme().
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
	IconFonts      []IconFont
}

// Default returns the light theme.
func Default() *Theme { return New(LightColors()) }

// DarkTheme returns the dark theme.
func DarkTheme() *Theme { return New(DarkColors()) }

// New creates a Theme from a ColorScheme, syncing the flat backward-compatible fields.
func New(cs ColorScheme) *Theme {
	cs = normalizeColorScheme(cs)
	return &Theme{
		Colors:         cs,
		Shapes:         DefaultShapeScale(),
		Types:          DefaultTypeScale(),
		Density:        DefaultDensityScale(),
		Primary:        cs.Primary,
		Surface:        cs.Surface,
		SurfaceMuted:   cs.SurfaceVariant,
		TextColor:      cs.OnSurface,
		TextOnPrimary:  cs.OnPrimary,
		Disabled:       cs.Disabled,
		TextSize:       16,
		DefaultFont:    DefaultFontSpec(),
		UseSystemFonts: true,
	}
}
