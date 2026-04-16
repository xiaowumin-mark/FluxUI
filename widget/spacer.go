package widget

import (
	"image"

	"fluxui/internal"
	"fluxui/layout"
)

type spacerWidget struct {
	width  float32
	height float32
}

// Spacer 创建固定尺寸空白。
func Spacer(width, height float32) Widget {
	return &spacerWidget{
		width:  width,
		height: height,
	}
}

// HSpacer 创建水平空白。
func HSpacer(width float32) Widget {
	return Spacer(width, 0)
}

// VSpacer 创建垂直空白。
func VSpacer(height float32) Widget {
	return Spacer(0, height)
}

func (s *spacerWidget) Layout(ctx *internal.Context) layout.Dimensions {
	size := image.Point{
		X: ctx.Gtx.Dp(safeDp(s.width)),
		Y: ctx.Gtx.Dp(safeDp(s.height)),
	}
	size = ctx.Gtx.Constraints.Constrain(size)
	return layout.Dimensions{Size: size}
}
