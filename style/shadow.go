package style

import "image/color"

// BoxShadow 描述盒阴影。用多层半透明偏移矩形模拟模糊，
// 近似于 Material Design 的阴影规范。
type BoxShadow struct {
	OffsetX float32
	OffsetY float32
	Blur    float32
	Color   color.NRGBA
}

func (s BoxShadow) IsZero() bool {
	return s.Blur <= 0 || s.Color.A == 0
}

// ElevationBoxShadow 返回 Material Design 高度等级对应的阴影预设。
// level [1,5] 映射：1=按钮hover, 2=卡片, 3=浮卡/FAB,
// 4=对话框, 5=模态/抽屉。
func ElevationBoxShadow(level int) BoxShadow {
	switch level {
	case 1:
		return BoxShadow{OffsetY: 1, Blur: 4, Color: color.NRGBA{R: 0, G: 0, B: 0, A: 40}}
	case 2:
		return BoxShadow{OffsetY: 2, Blur: 8, Color: color.NRGBA{R: 0, G: 0, B: 0, A: 52}}
	case 3:
		return BoxShadow{OffsetY: 5, Blur: 16, Color: color.NRGBA{R: 0, G: 0, B: 0, A: 62}}
	case 4:
		return BoxShadow{OffsetY: 10, Blur: 28, Color: color.NRGBA{R: 0, G: 0, B: 0, A: 72}}
	case 5:
		return BoxShadow{OffsetY: 16, Blur: 40, Color: color.NRGBA{R: 0, G: 0, B: 0, A: 82}}
	default:
		if level <= 0 {
			return BoxShadow{}
		}
		s := ElevationBoxShadow(5)
		s.Blur = s.Blur * float32(level) / 5.0
		return s
	}
}
