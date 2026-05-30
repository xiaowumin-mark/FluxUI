package style

import "time"

const (
	InteractionHoverDuration    = 120 * time.Millisecond
	InteractionPressedDuration  = 80 * time.Millisecond
	InteractionSelectedDuration = 160 * time.Millisecond
	InteractionMenuDuration     = 140 * time.Millisecond
	InteractionToastDuration    = 180 * time.Millisecond
	InteractionRippleExpand     = 450 * time.Millisecond
	InteractionRippleFade       = 550 * time.Millisecond
)

func InteractionEasing(v float32) float32 {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return v * v * (3 - 2*v)
}
