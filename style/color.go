package style

import "image/color"

// NRGBA 是创建颜色的便捷方法。
func NRGBA(r, g, b, a uint8) color.NRGBA {
	return color.NRGBA{R: r, G: g, B: b, A: a}
}
