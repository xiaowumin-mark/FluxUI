package theme

import (
	"image/color"
	"math"

	"github.com/worldiety/material-color-utilities/hct"
)

// ColorOption configures seed-generated Material Design 3 color schemes.
type ColorOption func(*colorSchemeOptions)

type colorSchemeOptions struct {
	dark bool

	secondarySeed    color.NRGBA
	hasSecondarySeed bool
	tertiarySeed     color.NRGBA
	hasTertiarySeed  bool
	errorSeed        color.NRGBA
	hasErrorSeed     bool
	successSeed      color.NRGBA
	hasSuccessSeed   bool
	warningSeed      color.NRGBA
	hasWarningSeed   bool
}

// WithColorSchemeDark switches ColorSchemeFromSeed between light and dark role mapping.
func WithColorSchemeDark(dark bool) ColorOption {
	return func(cfg *colorSchemeOptions) {
		cfg.dark = dark
	}
}

// WithSecondarySeed overrides the secondary tonal palette seed.
func WithSecondarySeed(seed color.NRGBA) ColorOption {
	return func(cfg *colorSchemeOptions) {
		cfg.secondarySeed = opaque(seed)
		cfg.hasSecondarySeed = true
	}
}

// WithTertiarySeed overrides the tertiary tonal palette seed.
func WithTertiarySeed(seed color.NRGBA) ColorOption {
	return func(cfg *colorSchemeOptions) {
		cfg.tertiarySeed = opaque(seed)
		cfg.hasTertiarySeed = true
	}
}

// WithErrorSeed overrides the error tonal palette seed.
func WithErrorSeed(seed color.NRGBA) ColorOption {
	return func(cfg *colorSchemeOptions) {
		cfg.errorSeed = opaque(seed)
		cfg.hasErrorSeed = true
	}
}

// WithSuccessSeed overrides the success business color seed.
func WithSuccessSeed(seed color.NRGBA) ColorOption {
	return func(cfg *colorSchemeOptions) {
		cfg.successSeed = opaque(seed)
		cfg.hasSuccessSeed = true
	}
}

// WithWarningSeed overrides the warning business color seed.
func WithWarningSeed(seed color.NRGBA) ColorOption {
	return func(cfg *colorSchemeOptions) {
		cfg.warningSeed = opaque(seed)
		cfg.hasWarningSeed = true
	}
}

// ColorSchemeFromSeed creates a complete MD3-like ColorScheme from a brand seed.
func ColorSchemeFromSeed(seed color.NRGBA, opts ...ColorOption) ColorScheme {
	cfg := applyColorOptions(opts...)
	return colorSchemeFromSeed(seed, cfg)
}

// LightColorSchemeFromSeed creates a light ColorScheme from a brand seed.
func LightColorSchemeFromSeed(seed color.NRGBA, opts ...ColorOption) ColorScheme {
	cfg := applyColorOptions(opts...)
	cfg.dark = false
	return colorSchemeFromSeed(seed, cfg)
}

// DarkColorSchemeFromSeed creates a dark ColorScheme from a brand seed.
func DarkColorSchemeFromSeed(seed color.NRGBA, opts ...ColorOption) ColorScheme {
	cfg := applyColorOptions(opts...)
	cfg.dark = true
	return colorSchemeFromSeed(seed, cfg)
}

// ThemeFromSeed creates a Theme from a seed-generated ColorScheme.
func ThemeFromSeed(seed color.NRGBA, opts ...ColorOption) *Theme {
	return New(ColorSchemeFromSeed(seed, opts...))
}

// LightThemeFromSeed creates a light Theme from a seed-generated ColorScheme.
func LightThemeFromSeed(seed color.NRGBA, opts ...ColorOption) *Theme {
	return New(LightColorSchemeFromSeed(seed, opts...))
}

// DarkThemeFromSeed creates a dark Theme from a seed-generated ColorScheme.
func DarkThemeFromSeed(seed color.NRGBA, opts ...ColorOption) *Theme {
	return New(DarkColorSchemeFromSeed(seed, opts...))
}

func applyColorOptions(opts ...ColorOption) colorSchemeOptions {
	cfg := colorSchemeOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

type corePalette struct {
	primary        tonalPalette
	secondary      tonalPalette
	tertiary       tonalPalette
	neutral        tonalPalette
	neutralVariant tonalPalette
	error          tonalPalette
	success        tonalPalette
	warning        tonalPalette
}

type tonalPalette struct {
	hue    float64
	chroma float64
}

func colorSchemeFromSeed(seed color.NRGBA, cfg colorSchemeOptions) ColorScheme {
	palette := corePaletteFromSeed(opaque(seed))
	if cfg.hasSecondarySeed {
		palette.secondary = accentPaletteFromSeed(cfg.secondarySeed)
	}
	if cfg.hasTertiarySeed {
		palette.tertiary = accentPaletteFromSeed(cfg.tertiarySeed)
	}
	if cfg.hasErrorSeed {
		palette.error = accentPaletteFromSeed(cfg.errorSeed)
	}
	if cfg.hasSuccessSeed {
		palette.success = accentPaletteFromSeed(cfg.successSeed)
	}
	if cfg.hasWarningSeed {
		palette.warning = accentPaletteFromSeed(cfg.warningSeed)
	}

	if cfg.dark {
		return darkSchemeFromPalette(palette)
	}
	return lightSchemeFromPalette(palette)
}

func corePaletteFromSeed(seed color.NRGBA) corePalette {
	source := hct.FromInt(argbFromNRGBA(seed))
	hue := source.GetHue()
	chroma := source.GetChroma()
	return corePalette{
		primary:        tonalPalette{hue: hue, chroma: math.Max(48, chroma)},
		secondary:      tonalPalette{hue: hue, chroma: 16},
		tertiary:       tonalPalette{hue: sanitizeHue(hue + 60), chroma: 24},
		neutral:        tonalPalette{hue: hue, chroma: 4},
		neutralVariant: tonalPalette{hue: hue, chroma: 8},
		error:          tonalPalette{hue: 25, chroma: 84},
		success:        accentPaletteFromSeed(color.NRGBA{R: 48, G: 106, B: 71, A: 255}),
		warning:        accentPaletteFromSeed(color.NRGBA{R: 116, G: 86, B: 0, A: 255}),
	}
}

func accentPaletteFromSeed(seed color.NRGBA) tonalPalette {
	source := hct.FromInt(argbFromNRGBA(opaque(seed)))
	return tonalPalette{hue: source.GetHue(), chroma: math.Max(48, source.GetChroma())}
}

func lightSchemeFromPalette(p corePalette) ColorScheme {
	cs := ColorScheme{
		Primary:            p.primary.tone(40),
		OnPrimary:          p.primary.tone(100),
		PrimaryContainer:   p.primary.tone(90),
		OnPrimaryContainer: p.primary.tone(10),

		Secondary:            p.secondary.tone(40),
		OnSecondary:          p.secondary.tone(100),
		SecondaryContainer:   p.secondary.tone(90),
		OnSecondaryContainer: p.secondary.tone(10),

		Tertiary:            p.tertiary.tone(40),
		OnTertiary:          p.tertiary.tone(100),
		TertiaryContainer:   p.tertiary.tone(90),
		OnTertiaryContainer: p.tertiary.tone(10),

		Error:            p.error.tone(40),
		OnError:          p.error.tone(100),
		ErrorContainer:   p.error.tone(90),
		OnErrorContainer: p.error.tone(10),

		Background:   p.neutral.tone(98),
		OnBackground: p.neutral.tone(10),

		Surface:                 p.neutral.tone(98),
		OnSurface:               p.neutral.tone(10),
		SurfaceVariant:          p.neutralVariant.tone(90),
		OnSurfaceVariant:        p.neutralVariant.tone(30),
		SurfaceContainerLowest:  p.neutral.tone(100),
		SurfaceContainerLow:     p.neutral.tone(96),
		SurfaceContainer:        p.neutral.tone(94),
		SurfaceContainerHigh:    p.neutral.tone(92),
		SurfaceContainerHighest: p.neutral.tone(90),

		Outline:        p.neutralVariant.tone(50),
		OutlineVariant: p.neutralVariant.tone(80),

		InverseSurface:   p.neutral.tone(20),
		InverseOnSurface: p.neutral.tone(95),
		InversePrimary:   p.primary.tone(80),

		Scrim:  color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		Shadow: color.NRGBA{R: 0, G: 0, B: 0, A: 255},

		Success:   p.success.tone(40),
		OnSuccess: p.success.tone(100),
		Warning:   p.warning.tone(40),
		OnWarning: p.warning.tone(100),
		Disabled:  p.neutralVariant.tone(50),
	}
	cs.SurfaceMuted = cs.SurfaceVariant
	return normalizeColorScheme(cs)
}

func darkSchemeFromPalette(p corePalette) ColorScheme {
	cs := ColorScheme{
		Primary:            p.primary.tone(80),
		OnPrimary:          p.primary.tone(20),
		PrimaryContainer:   p.primary.tone(30),
		OnPrimaryContainer: p.primary.tone(90),

		Secondary:            p.secondary.tone(80),
		OnSecondary:          p.secondary.tone(20),
		SecondaryContainer:   p.secondary.tone(30),
		OnSecondaryContainer: p.secondary.tone(90),

		Tertiary:            p.tertiary.tone(80),
		OnTertiary:          p.tertiary.tone(20),
		TertiaryContainer:   p.tertiary.tone(30),
		OnTertiaryContainer: p.tertiary.tone(90),

		Error:            p.error.tone(80),
		OnError:          p.error.tone(20),
		ErrorContainer:   p.error.tone(30),
		OnErrorContainer: p.error.tone(90),

		Background:   p.neutral.tone(6),
		OnBackground: p.neutral.tone(90),

		Surface:                 p.neutral.tone(6),
		OnSurface:               p.neutral.tone(90),
		SurfaceVariant:          p.neutralVariant.tone(30),
		OnSurfaceVariant:        p.neutralVariant.tone(80),
		SurfaceContainerLowest:  p.neutral.tone(4),
		SurfaceContainerLow:     p.neutral.tone(10),
		SurfaceContainer:        p.neutral.tone(12),
		SurfaceContainerHigh:    p.neutral.tone(17),
		SurfaceContainerHighest: p.neutral.tone(22),

		Outline:        p.neutralVariant.tone(60),
		OutlineVariant: p.neutralVariant.tone(30),

		InverseSurface:   p.neutral.tone(90),
		InverseOnSurface: p.neutral.tone(20),
		InversePrimary:   p.primary.tone(40),

		Scrim:  color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		Shadow: color.NRGBA{R: 0, G: 0, B: 0, A: 255},

		Success:   p.success.tone(80),
		OnSuccess: p.success.tone(20),
		Warning:   p.warning.tone(80),
		OnWarning: p.warning.tone(20),
		Disabled:  p.neutralVariant.tone(60),
	}
	cs.SurfaceMuted = cs.SurfaceVariant
	return normalizeColorScheme(cs)
}

func (p tonalPalette) tone(tone float64) color.NRGBA {
	col := hct.From(p.hue, p.chroma, tone)
	return nrgbaFromARGB(col.ToInt())
}

func argbFromNRGBA(c color.NRGBA) int {
	return (int(c.A) << 24) | (int(c.R) << 16) | (int(c.G) << 8) | int(c.B)
}

func nrgbaFromARGB(argb int) color.NRGBA {
	return color.NRGBA{
		R: uint8((argb >> 16) & 0xff),
		G: uint8((argb >> 8) & 0xff),
		B: uint8(argb & 0xff),
		A: 255,
	}
}

func opaque(c color.NRGBA) color.NRGBA {
	c.A = 255
	return c
}

func sanitizeHue(hue float64) float64 {
	hue = math.Mod(hue, 360)
	if hue < 0 {
		hue += 360
	}
	return hue
}
