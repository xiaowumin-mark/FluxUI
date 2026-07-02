package style

import "time"

const (
	InteractionHoverEnterDuration = 100 * time.Millisecond
	InteractionHoverExitDuration  = 100 * time.Millisecond
	InteractionHoverDuration      = InteractionHoverEnterDuration

	InteractionPressedEnterDuration = 50 * time.Millisecond
	InteractionPressedExitDuration  = 100 * time.Millisecond
	InteractionPressedDuration      = InteractionPressedEnterDuration

	InteractionFocusEnterDuration = 100 * time.Millisecond
	InteractionFocusExitDuration  = 100 * time.Millisecond
	InteractionFocusDuration      = InteractionFocusEnterDuration

	InteractionSelectedDuration          = 150 * time.Millisecond
	InteractionSelectedIndicatorDuration = 200 * time.Millisecond

	InteractionMenuEnterDuration = 260 * time.Millisecond
	InteractionMenuExitDuration  = 180 * time.Millisecond
	InteractionMenuDuration      = InteractionMenuEnterDuration

	InteractionToastEnterDuration = 200 * time.Millisecond
	InteractionToastExitDuration  = 150 * time.Millisecond
	InteractionToastDuration      = InteractionToastEnterDuration

	InteractionLoadingValueDuration   = 250 * time.Millisecond
	InteractionLoadingLinearCycle     = 1200 * time.Millisecond
	InteractionLoadingCircularCycle   = 1400 * time.Millisecond
	InteractionLoadingShimmerCycle    = 1200 * time.Millisecond
	InteractionLoadingDefaultDuration = InteractionLoadingLinearCycle

	InteractionRippleExpand = 450 * time.Millisecond
	InteractionRippleFade   = 550 * time.Millisecond
)

func InteractionEasing(v float32) float32 {
	return InteractionStandardEasing(v)
}

func InteractionLinearEasing(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func InteractionStandardEasing(v float32) float32 {
	return interactionCubicBezier(v, 0.2, 0, 0, 1)
}

func InteractionStandardDecelerateEasing(v float32) float32 {
	return interactionCubicBezier(v, 0, 0, 0, 1)
}

func InteractionStandardAccelerateEasing(v float32) float32 {
	return interactionCubicBezier(v, 0.3, 0, 1, 1)
}

func InteractionEmphasizedEasing(v float32) float32 {
	return interactionCubicBezier(v, 0.2, 0, 0, 1)
}

func InteractionEmphasizedDecelerateEasing(v float32) float32 {
	return interactionCubicBezier(v, 0.05, 0.7, 0.1, 1)
}

func InteractionEmphasizedAccelerateEasing(v float32) float32 {
	return interactionCubicBezier(v, 0.3, 0, 0.8, 0.15)
}

func interactionCubicBezier(v, x1, y1, x2, y2 float32) float32 {
	v = InteractionLinearEasing(v)
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 1
	}

	lo := float32(0)
	hi := float32(1)
	for range 24 {
		t := (lo + hi) / 2
		x := interactionCubic(t, 0, x1, x2, 1)
		if x < v {
			lo = t
		} else {
			hi = t
		}
	}
	t := (lo + hi) / 2
	return interactionCubic(t, 0, y1, y2, 1)
}

func interactionCubic(t, p0, p1, p2, p3 float32) float32 {
	u := 1 - t
	return u*u*u*p0 + 3*u*u*t*p1 + 3*u*t*t*p2 + t*t*t*p3
}
