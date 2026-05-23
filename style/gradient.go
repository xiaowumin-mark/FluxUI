package style

import (
	"image"
	"image/color"
)

// LinearGradient 描述两点之间的线性渐变。Start 和 End 在组件本地坐标系内，
// 渲染时转换为 Gio 的 f32.Point。
type LinearGradient struct {
	Start image.Point
	End   image.Point
	From  color.NRGBA
	To    color.NRGBA
}
