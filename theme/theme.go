package theme

import "image/color"

type Theme struct {
	Colors             ColorScheme
	Shapes             ShapeScale
	Types              TypeScale
	Density            DensityScale
	InteractionQuality InteractionQuality

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
func Default(opts ...ThemeOption) *Theme { return New(LightColors(), opts...) }

// DarkTheme returns the dark theme.
func DarkTheme(opts ...ThemeOption) *Theme { return New(DarkColors(), opts...) }

// New creates a Theme from a ColorScheme, syncing the flat backward-compatible fields.
func New(cs ColorScheme, opts ...ThemeOption) *Theme {
	cs = normalizeColorScheme(cs)
	th := &Theme{
		Colors:             cs,
		Shapes:             DefaultShapeScale(),
		Types:              DefaultTypeScale(),
		Density:            DefaultDensityScale(),
		InteractionQuality: defaultInteractionQuality(),
		Primary:            cs.Primary,
		Surface:            cs.Surface,
		SurfaceMuted:       cs.SurfaceVariant,
		TextColor:          cs.OnSurface,
		TextOnPrimary:      cs.OnPrimary,
		Disabled:           cs.Disabled,
		TextSize:           16,
		DefaultFont:        DefaultFontSpec(),
		UseSystemFonts:     true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(th)
		}
	}
	return th
}
