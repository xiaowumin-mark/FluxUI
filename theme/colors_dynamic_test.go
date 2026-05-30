package theme

import (
	"image/color"
	"math"
	"testing"
)

func TestColorSchemeFromSeedProvidesCompleteLightRoles(t *testing.T) {
	cs := LightColorSchemeFromSeed(color.NRGBA{R: 0x00, G: 0x57, B: 0xd9, A: 255})

	assertCompleteColorScheme(t, cs)
	assertReadablePair(t, "primary", cs.Primary, cs.OnPrimary, 4.5)
	assertReadablePair(t, "primary container", cs.PrimaryContainer, cs.OnPrimaryContainer, 4.5)
	assertReadablePair(t, "secondary", cs.Secondary, cs.OnSecondary, 4.5)
	assertReadablePair(t, "tertiary", cs.Tertiary, cs.OnTertiary, 4.5)
	assertReadablePair(t, "error", cs.Error, cs.OnError, 4.5)
	assertReadablePair(t, "surface", cs.Surface, cs.OnSurface, 4.5)

	if cs.SurfaceMuted != cs.SurfaceVariant {
		t.Fatalf("SurfaceMuted compatibility alias should follow SurfaceVariant")
	}
}

func TestColorSchemeFromSeedProvidesCompleteDarkRoles(t *testing.T) {
	seed := color.NRGBA{R: 0x00, G: 0x57, B: 0xd9, A: 255}
	light := ColorSchemeFromSeed(seed)
	dark := DarkColorSchemeFromSeed(seed)
	darkViaOption := ColorSchemeFromSeed(seed, WithColorSchemeDark(true))

	assertCompleteColorScheme(t, dark)
	assertReadablePair(t, "dark primary", dark.Primary, dark.OnPrimary, 4.5)
	assertReadablePair(t, "dark primary container", dark.PrimaryContainer, dark.OnPrimaryContainer, 4.5)
	assertReadablePair(t, "dark secondary", dark.Secondary, dark.OnSecondary, 4.5)
	assertReadablePair(t, "dark tertiary", dark.Tertiary, dark.OnTertiary, 4.5)
	assertReadablePair(t, "dark error", dark.Error, dark.OnError, 4.5)
	assertReadablePair(t, "dark surface", dark.Surface, dark.OnSurface, 4.5)

	if light.Primary == dark.Primary {
		t.Fatalf("light and dark primary roles should differ")
	}
	if light.Surface == dark.Surface {
		t.Fatalf("light and dark surface roles should differ")
	}
	if darkViaOption != dark {
		t.Fatalf("WithColorSchemeDark(true) should match DarkColorSchemeFromSeed")
	}
}

func TestColorSchemeFromSeedRespondsToSeedAndOverrides(t *testing.T) {
	blue := color.NRGBA{R: 0x00, G: 0x57, B: 0xd9, A: 255}
	green := color.NRGBA{R: 0x1b, G: 0x80, B: 0x4d, A: 255}
	base := LightColorSchemeFromSeed(blue)
	greenScheme := LightColorSchemeFromSeed(green)
	custom := LightColorSchemeFromSeed(
		blue,
		WithSecondarySeed(color.NRGBA{R: 0xa6, G: 0x33, B: 0xff, A: 255}),
		WithTertiarySeed(color.NRGBA{R: 0xff, G: 0x6f, B: 0x00, A: 255}),
		WithErrorSeed(color.NRGBA{R: 0xbd, G: 0x00, B: 0x3c, A: 255}),
		WithSuccessSeed(color.NRGBA{R: 0x00, G: 0x78, B: 0x4d, A: 255}),
		WithWarningSeed(color.NRGBA{R: 0xe0, G: 0x96, B: 0x00, A: 255}),
	)

	if base.Primary == greenScheme.Primary {
		t.Fatalf("different primary seeds should produce different primary colors")
	}
	if custom.Primary != base.Primary {
		t.Fatalf("secondary/tertiary/business overrides should not change primary")
	}
	if custom.Secondary == base.Secondary {
		t.Fatalf("WithSecondarySeed should change secondary role")
	}
	if custom.Tertiary == base.Tertiary {
		t.Fatalf("WithTertiarySeed should change tertiary role")
	}
	if custom.Error == base.Error {
		t.Fatalf("WithErrorSeed should change error role")
	}
	if custom.Success == base.Success {
		t.Fatalf("WithSuccessSeed should change success role")
	}
	if custom.Warning == base.Warning {
		t.Fatalf("WithWarningSeed should change warning role")
	}
}

func TestThemeFromSeedSyncsCompatibilityFields(t *testing.T) {
	th := ThemeFromSeed(color.NRGBA{R: 0x67, G: 0x50, B: 0xa4, A: 255})

	if th.Primary != th.Colors.Primary {
		t.Fatalf("Primary compatibility field should sync from seed scheme")
	}
	if th.Surface != th.Colors.Surface {
		t.Fatalf("Surface compatibility field should sync from seed scheme")
	}
	if th.SurfaceMuted != th.Colors.SurfaceVariant {
		t.Fatalf("SurfaceMuted compatibility field should sync from seed scheme")
	}
	if th.TextColor != th.Colors.OnSurface {
		t.Fatalf("TextColor compatibility field should sync from seed scheme")
	}
	if th.TextOnPrimary != th.Colors.OnPrimary {
		t.Fatalf("TextOnPrimary compatibility field should sync from seed scheme")
	}
	if th.Disabled != th.Colors.Disabled {
		t.Fatalf("Disabled compatibility field should sync from seed scheme")
	}
}

func assertCompleteColorScheme(t *testing.T, cs ColorScheme) {
	t.Helper()
	fields := []struct {
		name string
		col  color.NRGBA
	}{
		{"Primary", cs.Primary},
		{"OnPrimary", cs.OnPrimary},
		{"PrimaryContainer", cs.PrimaryContainer},
		{"OnPrimaryContainer", cs.OnPrimaryContainer},
		{"Secondary", cs.Secondary},
		{"OnSecondary", cs.OnSecondary},
		{"SecondaryContainer", cs.SecondaryContainer},
		{"OnSecondaryContainer", cs.OnSecondaryContainer},
		{"Tertiary", cs.Tertiary},
		{"OnTertiary", cs.OnTertiary},
		{"TertiaryContainer", cs.TertiaryContainer},
		{"OnTertiaryContainer", cs.OnTertiaryContainer},
		{"Error", cs.Error},
		{"OnError", cs.OnError},
		{"ErrorContainer", cs.ErrorContainer},
		{"OnErrorContainer", cs.OnErrorContainer},
		{"Background", cs.Background},
		{"OnBackground", cs.OnBackground},
		{"Surface", cs.Surface},
		{"OnSurface", cs.OnSurface},
		{"SurfaceVariant", cs.SurfaceVariant},
		{"OnSurfaceVariant", cs.OnSurfaceVariant},
		{"SurfaceContainerLowest", cs.SurfaceContainerLowest},
		{"SurfaceContainerLow", cs.SurfaceContainerLow},
		{"SurfaceContainer", cs.SurfaceContainer},
		{"SurfaceContainerHigh", cs.SurfaceContainerHigh},
		{"SurfaceContainerHighest", cs.SurfaceContainerHighest},
		{"Outline", cs.Outline},
		{"OutlineVariant", cs.OutlineVariant},
		{"InverseSurface", cs.InverseSurface},
		{"InverseOnSurface", cs.InverseOnSurface},
		{"InversePrimary", cs.InversePrimary},
		{"Scrim", cs.Scrim},
		{"Shadow", cs.Shadow},
		{"Success", cs.Success},
		{"OnSuccess", cs.OnSuccess},
		{"Warning", cs.Warning},
		{"OnWarning", cs.OnWarning},
		{"Disabled", cs.Disabled},
		{"SurfaceMuted", cs.SurfaceMuted},
	}
	for _, field := range fields {
		assertColorSet(t, field.name, field.col)
	}
}

func assertReadablePair(t *testing.T, name string, bg, fg color.NRGBA, min float64) {
	t.Helper()
	if got := contrastRatio(bg, fg); got < min {
		t.Fatalf("%s contrast = %.2f, want at least %.2f", name, got, min)
	}
}

func contrastRatio(a, b color.NRGBA) float64 {
	l1 := relativeLuminance(a)
	l2 := relativeLuminance(b)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

func relativeLuminance(c color.NRGBA) float64 {
	r := linearRGB(c.R)
	g := linearRGB(c.G)
	b := linearRGB(c.B)
	return 0.2126*r + 0.7152*g + 0.0722*b
}

func linearRGB(v uint8) float64 {
	x := float64(v) / 255
	if x <= 0.03928 {
		return x / 12.92
	}
	return math.Pow((x+0.055)/1.055, 2.4)
}
