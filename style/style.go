package style

import "image/color"

// Style 描述通用容器样式。
type Style struct {
	Background color.NRGBA
	Padding    Insets
	Margin     Insets
	Radius     float32
}
