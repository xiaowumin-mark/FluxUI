package theme

import (
	"image/color"
	"testing"
)

func TestDefaultThemeProvidesMD3Tokens(t *testing.T) {
	th := Default()
	if th == nil {
		t.Fatal("expected default theme")
	}
	assertColorSet(t, "PrimaryContainer", th.Colors.PrimaryContainer)
	assertColorSet(t, "OnPrimaryContainer", th.Colors.OnPrimaryContainer)
	assertColorSet(t, "SecondaryContainer", th.Colors.SecondaryContainer)
	assertColorSet(t, "Tertiary", th.Colors.Tertiary)
	assertColorSet(t, "ErrorContainer", th.Colors.ErrorContainer)
	assertColorSet(t, "SurfaceVariant", th.Colors.SurfaceVariant)
	assertColorSet(t, "SurfaceContainer", th.Colors.SurfaceContainer)
	assertColorSet(t, "OutlineVariant", th.Colors.OutlineVariant)
	assertColorSet(t, "InverseSurface", th.Colors.InverseSurface)
	assertColorSet(t, "Scrim", th.Colors.Scrim)
	assertColorSet(t, "Shadow", th.Colors.Shadow)

	if th.Shapes.Full < 100 {
		t.Fatalf("expected full shape token, got %v", th.Shapes.Full)
	}
	if th.Types.LabelLarge.Size <= 0 || th.Types.BodyLarge.Size <= 0 {
		t.Fatalf("expected non-zero type scale, got %#v", th.Types)
	}
	if th.Density.Level != DensityDefault {
		t.Fatalf("expected default density, got %#v", th.Density)
	}
}

func TestDensityScaleHelpers(t *testing.T) {
	defaultDensity := DefaultDensityScale()
	compactDensity := CompactDensityScale()

	if defaultDensity.IsCompact() {
		t.Fatalf("default density should not be compact")
	}
	if !compactDensity.IsCompact() {
		t.Fatalf("compact density should be compact")
	}
	if got := defaultDensity.ComponentHeight(56, 48); got != 56 {
		t.Fatalf("default component height = %v, want 56", got)
	}
	if got := compactDensity.ComponentHeight(56, 48); got != 48 {
		t.Fatalf("compact component height = %v, want 48", got)
	}
	if got := compactDensity.Metric(16, 12); got != 12 {
		t.Fatalf("compact metric = %v, want 12", got)
	}
}

func TestThemeCompatibilityFieldsSyncFromColorScheme(t *testing.T) {
	cs := LightColors()
	th := New(cs)

	if th.Primary != cs.Primary {
		t.Fatal("Primary should sync from ColorScheme.Primary")
	}
	if th.Surface != cs.Surface {
		t.Fatal("Surface should sync from ColorScheme.Surface")
	}
	if th.SurfaceMuted != cs.SurfaceVariant {
		t.Fatal("SurfaceMuted compatibility field should sync from SurfaceVariant")
	}
	if th.TextColor != cs.OnSurface {
		t.Fatal("TextColor should sync from ColorScheme.OnSurface")
	}
	if th.TextOnPrimary != cs.OnPrimary {
		t.Fatal("TextOnPrimary should sync from ColorScheme.OnPrimary")
	}
	if th.Disabled != cs.Disabled {
		t.Fatal("Disabled should sync from ColorScheme.Disabled")
	}
}

func TestThemeNormalizesLegacyColorScheme(t *testing.T) {
	legacy := ColorScheme{
		Primary:      color.NRGBA{R: 1, A: 255},
		OnPrimary:    color.NRGBA{R: 2, A: 255},
		Secondary:    color.NRGBA{R: 3, A: 255},
		OnSecondary:  color.NRGBA{R: 4, A: 255},
		Surface:      color.NRGBA{R: 5, A: 255},
		OnSurface:    color.NRGBA{R: 6, A: 255},
		SurfaceMuted: color.NRGBA{R: 7, A: 255},
		Error:        color.NRGBA{R: 8, A: 255},
		OnError:      color.NRGBA{R: 9, A: 255},
		Outline:      color.NRGBA{R: 10, A: 255},
		Disabled:     color.NRGBA{R: 11, A: 255},
	}
	th := New(legacy)

	if th.Colors.SurfaceVariant != legacy.SurfaceMuted {
		t.Fatal("legacy SurfaceMuted should backfill SurfaceVariant")
	}
	if th.Colors.PrimaryContainer != legacy.Primary || th.Colors.OnPrimaryContainer != legacy.OnPrimary {
		t.Fatal("legacy primary roles should backfill primary containers")
	}
	if th.Colors.SecondaryContainer != legacy.Secondary || th.Colors.OnSecondaryContainer != legacy.OnSecondary {
		t.Fatal("legacy secondary roles should backfill secondary containers")
	}
	if th.Colors.Tertiary != legacy.Secondary || th.Colors.OnTertiary != legacy.OnSecondary {
		t.Fatal("legacy secondary roles should backfill tertiary roles")
	}
	if th.Colors.OutlineVariant != legacy.Outline {
		t.Fatal("legacy Outline should backfill OutlineVariant")
	}
	if th.SurfaceMuted != legacy.SurfaceMuted {
		t.Fatal("compatibility SurfaceMuted should remain set")
	}
}

func TestThemeNormalizesSparseLegacyColorScheme(t *testing.T) {
	th := New(ColorScheme{
		Primary:   color.NRGBA{R: 1, A: 255},
		OnPrimary: color.NRGBA{R: 2, A: 255},
		Surface:   color.NRGBA{R: 3, A: 255},
		OnSurface: color.NRGBA{R: 4, A: 255},
	})

	assertCompleteColorSchemeForTheme(t, th.Colors)
	if th.Colors.Background != th.Colors.Surface {
		t.Fatal("sparse scheme should backfill Background from Surface")
	}
	if th.Colors.OnBackground != th.Colors.OnSurface {
		t.Fatal("sparse scheme should backfill OnBackground from OnSurface")
	}
}

func assertColorSet(t *testing.T, name string, col color.NRGBA) {
	t.Helper()
	if col.A == 0 {
		t.Fatalf("%s alpha should be set", name)
	}
}

func assertCompleteColorSchemeForTheme(t *testing.T, cs ColorScheme) {
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
