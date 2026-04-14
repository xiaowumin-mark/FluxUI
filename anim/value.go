package anim

func clamp01(v float32) float32 {
	switch {
	case v < 0:
		return 0
	case v > 1:
		return 1
	default:
		return v
	}
}

func lerp(from, to, progress float32) float32 {
	return from + (to-from)*progress
}
