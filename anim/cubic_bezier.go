package anim

import "math"

func CubicBezier(x1, y1, x2, y2 float32) Easing {
	x1 = clampDefault(x1, 0, 1)
	x2 = clampDefault(x2, 0, 1)

	const samples = 1024
	var lut [samples]float32

	for i := 0; i < samples; i++ {
		t := float32(i) / float32(samples-1)
		lut[i] = bezierY(t, x1, y1, x2, y2)
	}

	return func(v float32) float32 {
		v = clamp01(v)
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 1
		}

		idx := int(v * float32(samples-1))
		if idx >= samples-1 {
			return lut[samples-1]
		}

		frac := v*float32(samples-1) - float32(idx)
		return lut[idx] + (lut[idx+1]-lut[idx])*frac
	}
}

func bezierY(t, x1, y1, x2, y2 float32) float32 {
	const epsilon = 0.0005
	lo := float32(0)
	hi := float32(1)
	guess := t

	for i := 0; i < 32; i++ {
		bx := cubicBezierComponent(guess, 0, x1, x2, 1)
		if absDiff(bx, t) < epsilon {
			return cubicBezierComponent(guess, 0, y1, y2, 1)
		}
		if bx < t {
			lo = guess
		} else {
			hi = guess
		}
		guess = (lo + hi) / 2
	}
	return cubicBezierComponent(guess, 0, y1, y2, 1)
}

func cubicBezierComponent(t, p0, p1, p2, p3 float32) float32 {
	u := 1 - t
	return u*u*u*p0 + 3*u*u*t*p1 + 3*u*t*t*p2 + t*t*t*p3
}

func absDiff(a, b float32) float32 {
	if a > b {
		return a - b
	}
	return b - a
}

func clampDefault(v, lo, hi float32) float32 {
	if math.IsNaN(float64(v)) {
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
