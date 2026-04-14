package layout

import (
	"image"

	"fluxui/internal"
)

// StackChild 描述 Stack 子项。
type StackChild struct {
	expanded bool
	layout   func(*internal.Context) Dimensions
}

// Stacked 创建普通堆叠子项。
func Stacked(layout func(*internal.Context) Dimensions) StackChild {
	return StackChild{layout: layout}
}

// Expanded 创建扩展子项。
func Expanded(layout func(*internal.Context) Dimensions) StackChild {
	return StackChild{
		expanded: true,
		layout:   layout,
	}
}

// Stack 执行堆叠布局。
func Stack(ctx *internal.Context, children ...StackChild) Dimensions {
	internalChildren := make([]internal.StackChild, 0, len(children))
	for _, child := range children {
		current := child
		internalChildren = append(internalChildren, internal.StackChild{
			Expanded: current.expanded,
			Layout: func(childCtx *internal.Context) image.Point {
				if current.layout == nil {
					return image.Point{}
				}
				return current.layout(childCtx).Size
			},
		})
	}

	size := ctx.LayoutStack(internalChildren...)
	return Dimensions{Size: size}
}
