package widget

import (
	"image/color"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/style"
	"github.com/xiaowumin-mark/FluxUI/theme"
)

func TestButtonMD3VariantDefaults(t *testing.T) {
	th := theme.New(theme.LightColors())

	filled := resolveButtonDefaults(buttonVariantFilled, th)
	if filled.background != th.Colors.Primary || filled.foreground != th.Colors.OnPrimary {
		t.Fatalf("filled button defaults = %#v, want primary/onPrimary", filled)
	}
	if filled.radius != th.Shapes.Full {
		t.Fatalf("filled button radius = %v, want %v", filled.radius, th.Shapes.Full)
	}

	outlined := resolveButtonDefaults(buttonVariantOutlined, th)
	if outlined.background != (color.NRGBA{}) {
		t.Fatalf("outlined button background = %#v, want transparent", outlined.background)
	}
	if outlined.border.Color != th.Colors.Outline || outlined.border.Width != 1 {
		t.Fatalf("outlined button border = %#v, want outline width 1", outlined.border)
	}
	if outlined.disabledContainer {
		t.Fatalf("outlined button disabled container should remain transparent")
	}

	elevated := resolveButtonDefaults(buttonVariantElevated, th)
	if elevated.background != style.SurfaceAtElevation(th.Colors, 1) {
		t.Fatalf("elevated button background = %#v, want tonal elevation", elevated.background)
	}
	if elevated.shadow.IsZero() {
		t.Fatalf("elevated button shadow should be set")
	}
}

func TestButtonRadiusAllowsExplicitZero(t *testing.T) {
	var cfg buttonConfig
	ButtonRadius(0)(&cfg)
	if !cfg.hasRadius {
		t.Fatalf("ButtonRadius should mark radius as explicitly configured")
	}
	if cfg.radius != 0 {
		t.Fatalf("ButtonRadius(0) radius = %v, want 0", cfg.radius)
	}
}

func TestInputMD3VariantDefaults(t *testing.T) {
	th := theme.New(theme.LightColors())

	outlined := resolveInputDefaults(inputVariantOutlined, th)
	if outlined.background != th.Colors.Surface {
		t.Fatalf("outlined input background = %#v, want surface", outlined.background)
	}
	if outlined.border != th.Colors.Outline || outlined.borderWidth != 1 {
		t.Fatalf("outlined input border = %#v width %v, want outline width 1", outlined.border, outlined.borderWidth)
	}
	if outlined.placeholder != th.Colors.OnSurfaceVariant {
		t.Fatalf("outlined input placeholder = %#v, want onSurfaceVariant", outlined.placeholder)
	}
	if outlined.text != th.Types.BodyLarge {
		t.Fatalf("outlined input text = %#v, want BodyLarge", outlined.text)
	}

	filled := resolveInputDefaults(inputVariantFilled, th)
	if filled.background != th.Colors.SurfaceContainerHighest {
		t.Fatalf("filled input background = %#v, want surfaceContainerHighest", filled.background)
	}
	if filled.border.A != 0 || filled.borderWidth != 0 {
		t.Fatalf("filled input border = %#v width %v, want none", filled.border, filled.borderWidth)
	}
}

func TestCardMD3VariantDefaults(t *testing.T) {
	th := theme.New(theme.LightColors())

	filled := resolveCardDefaults(cardVariantFilled, th)
	if filled.background != th.Colors.SurfaceContainerHighest {
		t.Fatalf("filled card background = %#v, want surfaceContainerHighest", filled.background)
	}
	if filled.radius != th.Shapes.Medium {
		t.Fatalf("filled card radius = %v, want %v", filled.radius, th.Shapes.Medium)
	}

	outlined := resolveCardDefaults(cardVariantOutlined, th)
	if outlined.background != th.Colors.Surface {
		t.Fatalf("outlined card background = %#v, want surface", outlined.background)
	}
	if outlined.border.Color != th.Colors.OutlineVariant || outlined.border.Width != 1 {
		t.Fatalf("outlined card border = %#v, want outlineVariant width 1", outlined.border)
	}

	elevated := resolveCardDefaults(cardVariantElevated, th)
	if elevated.background != style.SurfaceAtElevation(th.Colors, 1) {
		t.Fatalf("elevated card background = %#v, want tonal elevation", elevated.background)
	}
	if elevated.shadow.IsZero() {
		t.Fatalf("elevated card shadow should be set")
	}
}
