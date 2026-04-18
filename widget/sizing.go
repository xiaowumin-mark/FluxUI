package widget

import (
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
)

// FixedWidth 固定子组件宽度。
func FixedWidth(width float32, child Widget) Widget {
	return &fixedSizeWidget{
		width: width,
		child: child,
	}
}

// FixedHeight 固定子组件高度。
func FixedHeight(height float32, child Widget) Widget {
	return &fixedSizeWidget{
		height: height,
		child:  child,
	}
}

// FixedSize 固定子组件宽高。
func FixedSize(width, height float32, child Widget) Widget {
	return &fixedSizeWidget{
		width:  width,
		height: height,
		child:  child,
	}
}

// FillWidth 让子组件占满父组件可用宽度。
func FillWidth(child Widget) Widget {
	return expandWidth(child)
}

// FillHeight 让子组件占满父组件可用高度。
func FillHeight(child Widget) Widget {
	return &expandHeightWidget{child: child}
}

// Fill 让子组件占满父组件可用宽高。
func Fill(child Widget) Widget {
	return FillHeight(FillWidth(child))
}

type expandHeightWidget struct {
	child Widget
}

func (e *expandHeightWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if e.child == nil {
		return layout.Dimensions{}
	}

	gtx := ctx.Gtx
	stretch := gtx
	height := gtx.Constraints.Max.Y
	if height < 0 {
		height = 0
	}
	if stretch.Constraints.Min.Y < height {
		stretch.Constraints.Min.Y = height
	}
	stretch.Constraints.Max.Y = height
	if stretch.Constraints.Min.Y > stretch.Constraints.Max.Y {
		stretch.Constraints.Min.Y = stretch.Constraints.Max.Y
	}

	next := *ctx
	next.Gtx = stretch
	dims := e.child.Layout(next.Child(0))
	if dims.Size.Y < height {
		dims.Size.Y = height
	}
	dims.Size = clampPointToConstraints(dims.Size, stretch.Constraints.Min, stretch.Constraints.Max)
	return dims
}
