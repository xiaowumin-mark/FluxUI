package anim

import (
	"image/color"
	"math"
)

func clamp01(v float32) float32 {
	switch {
	case math.IsNaN(float64(v)):
		return 0
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

func lerpNRGBA(from, to color.NRGBA, progress float32) color.NRGBA {
	p := clamp01(progress)
	return color.NRGBA{
		R: uint8(float32(from.R) + (float32(to.R)-float32(from.R))*p),
		G: uint8(float32(from.G) + (float32(to.G)-float32(from.G))*p),
		B: uint8(float32(from.B) + (float32(to.B)-float32(from.B))*p),
		A: uint8(float32(from.A) + (float32(to.A)-float32(from.A))*p),
	}
}
