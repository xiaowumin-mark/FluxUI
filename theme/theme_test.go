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

func assertColorSet(t *testing.T, name string, col color.NRGBA) {
	t.Helper()
	if col.A == 0 {
		t.Fatalf("%s alpha should be set", name)
	}
}
