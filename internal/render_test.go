package internal

import (
	"image"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/theme"
)

func TestClampRoundedRadiusPxCapsFullRadiusToHalfShortestSide(t *testing.T) {
	got := clampRoundedRadiusPx(image.Pt(120, 40), 999)
	if got != 20 {
		t.Fatalf("clampRoundedRadiusPx() = %d, want 20", got)
	}
}

func TestResolveSliderVisualColorsKeepsStateLayerSeparate(t *testing.T) {
	cs := theme.LightColors()
	spec := SliderSpec{
		TrackColor:    cs.SurfaceVariant,
		ProgressColor: cs.Primary,
		ThumbColor:    cs.Primary,
		Dragged:       true,
		Pressed:       true,
		Hovered:       true,
	}

	track, _, thumb := resolveSliderVisualColors(spec, cs)
	if track != cs.SurfaceVariant {
		t.Fatalf("dragged track = %#v, want %#v", track, cs.SurfaceVariant)
	}
	if thumb != cs.Primary {
		t.Fatalf("dragged thumb = %#v, want %#v", thumb, cs.Primary)
	}
}

func TestDrawFocusIndicatorHandlesEmptyInputs(t *testing.T) {
	_, ctx := newTestContext()
	ctx.DrawFocusIndicator(image.Point{}, FocusIndicatorSpec{Color: theme.LightColors().Primary})
	ctx.DrawFocusIndicator(image.Pt(80, 32), FocusIndicatorSpec{})
}
