package style

import (
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/theme"
)

func TestStateLayerBlendsOnColorOverContainer(t *testing.T) {
	container := NRGBA(100, 100, 100, 255)
	onColor := NRGBA(200, 0, 0, 255)

	got := StateLayer(container, onColor, 0.5)
	if got.R != 150 || got.G != 50 || got.B != 50 || got.A != 255 {
		t.Fatalf("unexpected state layer color: %#v", got)
	}
}

func TestSurfaceAtElevationUsesPrimaryOverlay(t *testing.T) {
	cs := theme.LightColors()

	level0 := SurfaceAtElevation(cs, 0)
	level1 := SurfaceAtElevation(cs, 1)
	level5 := SurfaceAtElevation(cs, 5)

	if level0 != cs.Surface {
		t.Fatal("level 0 should return base surface")
	}
	if level1 == level0 {
		t.Fatal("level 1 should apply tonal overlay")
	}
	if level5 == level1 {
		t.Fatal("higher elevation should use a stronger tonal overlay")
	}
}

func TestDisabledHelpersUseMD3Opacity(t *testing.T) {
	onSurface := theme.LightColors().OnSurface

	if got := DisabledContainer(onSurface); got.A != 31 {
		t.Fatalf("expected disabled container alpha 31, got %d", got.A)
	}
	if got := DisabledContent(onSurface); got.A != 97 {
		t.Fatalf("expected disabled content alpha 97, got %d", got.A)
	}
}

func TestInteractionTokensAreStableAndBounded(t *testing.T) {
	if InteractionHoverDuration != 100*time.Millisecond {
		t.Fatalf("hover duration = %v, want 100ms", InteractionHoverDuration)
	}
	if InteractionPressedDuration <= 0 || InteractionPressedDuration >= InteractionSelectedDuration {
		t.Fatalf("pressed duration should be positive and shorter than selected duration")
	}
	if InteractionFocusDuration != 100*time.Millisecond {
		t.Fatalf("focus duration = %v, want 100ms", InteractionFocusDuration)
	}
	if InteractionSelectedIndicatorDuration <= InteractionSelectedDuration {
		t.Fatalf("selected indicator should be at least as long as selected color transition")
	}
	if InteractionMenuEnterDuration <= InteractionMenuExitDuration {
		t.Fatalf("menu enter should be longer than menu exit")
	}
	if InteractionToastEnterDuration <= InteractionToastExitDuration {
		t.Fatalf("toast enter should be longer than toast exit")
	}
	if InteractionLoadingValueDuration != 250*time.Millisecond {
		t.Fatalf("loading value duration = %v, want 250ms", InteractionLoadingValueDuration)
	}
	if InteractionLoadingLinearCycle != 1200*time.Millisecond || InteractionLoadingCircularCycle != 1400*time.Millisecond {
		t.Fatalf("unexpected loading cycles: linear=%v circular=%v", InteractionLoadingLinearCycle, InteractionLoadingCircularCycle)
	}
	if InteractionRippleExpand != 450*time.Millisecond || InteractionRippleFade != 550*time.Millisecond {
		t.Fatalf("unexpected ripple timing: expand=%v fade=%v", InteractionRippleExpand, InteractionRippleFade)
	}
	if InteractionEasing(-1) != 0 || InteractionEasing(2) != 1 {
		t.Fatalf("interaction easing should clamp outside [0,1]")
	}
	if got := InteractionEasing(0.5); got <= 0 || got >= 1 {
		t.Fatalf("interaction easing midpoint = %v, want inside (0,1)", got)
	}
	if got := InteractionLinearEasing(0.5); got != 0.5 {
		t.Fatalf("linear easing midpoint = %v, want 0.5", got)
	}
	if got := InteractionStandardDecelerateEasing(0.5); got <= InteractionStandardAccelerateEasing(0.5) {
		t.Fatalf("decelerate easing should progress faster than accelerate at midpoint")
	}
	if got := InteractionEmphasizedDecelerateEasing(0.5); got <= InteractionStandardEasing(0.5) {
		t.Fatalf("emphasized decelerate should progress faster than standard at midpoint")
	}
}
