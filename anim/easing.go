package anim

import "math"

type Easing func(float32) float32

var (
	Linear    Easing = easeLinear
	EaseOut   Easing = easeOutQuad
	EaseInOut Easing = easeInOutQuad

	EaseInQuad    Easing = easeInQuad
	EaseOutQuad   Easing = easeOutQuad
	EaseInOutQuad Easing = easeInOutQuad

	EaseInCubic    Easing = easeInCubic
	EaseOutCubic   Easing = easeOutCubic
	EaseInOutCubic Easing = easeInOutCubic

	EaseInQuart    Easing = easeInQuart
	EaseOutQuart   Easing = easeOutQuart
	EaseInOutQuart Easing = easeInOutQuart

	EaseInQuint    Easing = easeInQuint
	EaseOutQuint   Easing = easeOutQuint
	EaseInOutQuint Easing = easeInOutQuint

	EaseInSine    Easing = easeInSine
	EaseOutSine   Easing = easeOutSine
	EaseInOutSine Easing = easeInOutSine

	EaseInExpo    Easing = easeInExpo
	EaseOutExpo   Easing = easeOutExpo
	EaseInOutExpo Easing = easeInOutExpo

	EaseInCirc    Easing = easeInCirc
	EaseOutCirc   Easing = easeOutCirc
	EaseInOutCirc Easing = easeInOutCirc

	EaseInBack    Easing = easeInBack
	EaseOutBack   Easing = easeOutBack
	EaseInOutBack Easing = easeInOutBack

	EaseInElastic    Easing = easeInElastic
	EaseOutElastic   Easing = easeOutElastic
	EaseInOutElastic Easing = easeInOutElastic

	EaseInBounce    Easing = easeInBounce
	EaseOutBounce   Easing = easeOutBounce
	EaseInOutBounce Easing = easeInOutBounce
)

func easeLinear(v float32) float32 { return Clamp01(v) }

func easeInQuad(v float32) float32  { v = Clamp01(v); return v * v }
func easeOutQuad(v float32) float32 { v = Clamp01(v); return 1 - (1-v)*(1-v) }
func easeInOutQuad(v float32) float32 {
	v = Clamp01(v)
	if v < 0.5 {
		return 2 * v * v
	}
	return 1 - (-2*v+2)*(-2*v+2)/2
}

func easeInCubic(v float32) float32  { v = Clamp01(v); return v * v * v }
func easeOutCubic(v float32) float32 { v = Clamp01(v); return 1 - (1-v)*(1-v)*(1-v) }
func easeInOutCubic(v float32) float32 {
	v = Clamp01(v)
	if v < 0.5 {
		return 4 * v * v * v
	}
	return 1 - (-2*v+2)*(-2*v+2)*(-2*v+2)/2
}

func easeInQuart(v float32) float32  { v = Clamp01(v); return v * v * v * v }
func easeOutQuart(v float32) float32 { v = Clamp01(v); return 1 - (1-v)*(1-v)*(1-v)*(1-v) }
func easeInOutQuart(v float32) float32 {
	v = Clamp01(v)
	if v < 0.5 {
		return 8 * v * v * v * v
	}
	return 1 - (-2*v+2)*(-2*v+2)*(-2*v+2)*(-2*v+2)/2
}

func easeInQuint(v float32) float32  { v = Clamp01(v); return v * v * v * v * v }
func easeOutQuint(v float32) float32 { v = Clamp01(v); return 1 - (1-v)*(1-v)*(1-v)*(1-v)*(1-v) }
func easeInOutQuint(v float32) float32 {
	v = Clamp01(v)
	if v < 0.5 {
		return 16 * v * v * v * v * v
	}
	return 1 - (-2*v+2)*(-2*v+2)*(-2*v+2)*(-2*v+2)*(-2*v+2)/2
}

func easeInSine(v float32) float32 {
	v = Clamp01(v)
	return 1 - float32(math.Cos(float64(v)*math.Pi/2))
}
func easeOutSine(v float32) float32 {
	v = Clamp01(v)
	return float32(math.Sin(float64(v) * math.Pi / 2))
}
func easeInOutSine(v float32) float32 {
	v = Clamp01(v)
	return float32(-(math.Cos(math.Pi*float64(v)) - 1) / 2)
}

func easeInExpo(v float32) float32 {
	v = Clamp01(v)
	if v == 0 {
		return 0
	}
	return float32(math.Pow(2, 10*float64(v)-10))
}
func easeOutExpo(v float32) float32 {
	v = Clamp01(v)
	if v == 1 {
		return 1
	}
	return 1 - float32(math.Pow(2, -10*float64(v)))
}
func easeInOutExpo(v float32) float32 {
	v = Clamp01(v)
	if v == 0 {
		return 0
	}
	if v == 1 {
		return 1
	}
	if v < 0.5 {
		return float32(math.Pow(2, 20*float64(v)-10) / 2)
	}
	return (2 - float32(math.Pow(2, -20*float64(v)+10))) / 2
}

func easeInCirc(v float32) float32 { v = Clamp01(v); return 1 - float32(math.Sqrt(1-float64(v*v))) }
func easeOutCirc(v float32) float32 {
	v = Clamp01(v)
	return float32(math.Sqrt(1 - float64((v-1)*(v-1))))
}
func easeInOutCirc(v float32) float32 {
	v = Clamp01(v)
	if v < 0.5 {
		return (1 - float32(math.Sqrt(1-float64(4*v*v)))) / 2
	}
	return (float32(math.Sqrt(1-float64((-2*v+2)*(-2*v+2)))) + 1) / 2
}

const easeBackS float32 = 1.70158

func easeInBack(v float32) float32  { v = Clamp01(v); return v * v * ((easeBackS+1)*v - easeBackS) }
func easeOutBack(v float32) float32 { v = Clamp01(v); v--; return v*v*((easeBackS+1)*v+easeBackS) + 1 }
func easeInOutBack(v float32) float32 {
	v = Clamp01(v)
	s := easeBackS * 1.525
	if v < 0.5 {
		v = 2 * v
		return (v * v * ((s+1)*v - s)) / 2
	}
	v = 2*v - 2
	return (v*v*((s+1)*v+s) + 2) / 2
}

func easeInElastic(v float32) float32 {
	v = Clamp01(v)
	if v == 0 {
		return 0
	}
	if v == 1 {
		return 1
	}
	return -float32(math.Pow(2, 10*float64(v)-10)) * float32(math.Sin(float64(v*10-10.75)*2*math.Pi/3))
}

func easeOutElastic(v float32) float32 {
	v = Clamp01(v)
	if v == 0 {
		return 0
	}
	if v == 1 {
		return 1
	}
	return float32(math.Pow(2, -10*float64(v)))*float32(math.Sin(float64(v*10-0.75)*2*math.Pi/3)) + 1
}

func easeInOutElastic(v float32) float32 {
	v = Clamp01(v)
	if v == 0 {
		return 0
	}
	if v == 1 {
		return 1
	}
	c := 2 * math.Pi / 4.5
	x := float64(v)
	if v < 0.5 {
		return -0.5 * float32(math.Pow(2, 20*x-10)) * float32(math.Sin((20*x-11.125)*c))
	}
	return float32(math.Pow(2, -20*x+10))*float32(math.Sin((20*x-11.125)*c))*0.5 + 1
}

var bounceN1 float32 = 7.5625
var bounceD1 float32 = 2.75

func easeInBounce(v float32) float32  { return 1 - easeOutBounce(1-Clamp01(v)) }
func easeOutBounce(v float32) float32 { return bounceOut(Clamp01(v)) }
func easeInOutBounce(v float32) float32 {
	v = Clamp01(v)
	if v < 0.5 {
		return (1 - easeOutBounce(1-2*v)) / 2
	}
	return (1 + easeOutBounce(2*v-1)) / 2
}

func bounceOut(x float32) float32 {
	switch {
	case x < 1/bounceD1:
		return bounceN1 * x * x
	case x < 2/bounceD1:
		x -= 1.5 / bounceD1
		return bounceN1*x*x + 0.75
	case x < 2.5/bounceD1:
		x -= 2.25 / bounceD1
		return bounceN1*x*x + 0.9375
	default:
		x -= 2.625 / bounceD1
		return bounceN1*x*x + 0.984375
	}
}
