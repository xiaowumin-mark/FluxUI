package layout

import (
	"image"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
)

// Axis 表示 Flex 主轴。
type Axis int

const (
	Horizontal Axis = iota
	Vertical
)

// FlexChild 描述 Flex 子项。
type FlexChild struct {
	flexed bool
	weight float32
	layout func(*internal.Context) Dimensions
}

// Rigid 创建固定尺寸子项。
func Rigid(layout func(*internal.Context) Dimensions) FlexChild {
	return FlexChild{layout: layout}
}

// Flexed 创建带权重子项。
func Flexed(weight float32, layout func(*internal.Context) Dimensions) FlexChild {
	return FlexChild{
		flexed: true,
		weight: weight,
		layout: layout,
	}
}

// Flex 执行 Row/Column 布局。
func Flex(ctx *internal.Context, axis Axis, children ...FlexChild) Dimensions {
	internalChildren := make([]internal.FlexChild, 0, len(children))
	for _, child := range children {
		current := child
		internalChildren = append(internalChildren, internal.FlexChild{
			Flexed: current.flexed,
			Weight: current.weight,
			Layout: func(childCtx *internal.Context) image.Point {
				if current.layout == nil {
					return image.Point{}
				}
				return current.layout(childCtx).Size
			},
		})
	}

	size := ctx.LayoutFlex(toInternalAxis(axis), internalChildren...)
	return Dimensions{Size: size}
}

func toInternalAxis(axis Axis) internal.Axis {
	if axis == Vertical {
		return internal.Vertical
	}
	return internal.Horizontal
}
