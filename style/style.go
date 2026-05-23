package style

import "image/color"

// Deprecated: Style 已被 Decoration 取代。新代码请使用 ContainerDecoration +
// ui.Bg/Pad/Rad/Margin/Border 构造函数。旧 Container(Style) 仍可工作但不再推荐。
type Style struct {
	Background color.NRGBA
	Padding    Insets
	Margin     Insets
	Radius     float32
}
