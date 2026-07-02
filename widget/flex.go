package widget

import (
	"image"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
	layout "github.com/xiaowumin-mark/FluxUI/layout"
)

type flexWidget struct {
	axis         layout.Axis
	internalAxis internal.Axis
	children     []Widget
	items        []internal.FlexChild
}

type flexedWidget struct {
	weight float32
	child  Widget
}

// Row 创建横向布局。
func Row(children ...Widget) Widget {
	return newFlexWidget(layout.Horizontal, children...)
}

// Column 创建纵向布局。
func Column(children ...Widget) Widget {
	return newFlexWidget(layout.Vertical, children...)
}

// Flexed 创建带权重的弹性子项。
func Flexed(weight float32, child Widget) Widget {
	if weight <= 0 {
		weight = 1
	}
	return &flexedWidget{
		weight: weight,
		child:  child,
	}
}

// Expanded 创建权重为 1 的弹性子项。
func Expanded(child Widget) Widget {
	return Flexed(1, child)
}

func newFlexWidget(axis layout.Axis, children ...Widget) *flexWidget {
	f := &flexWidget{
		axis:         axis,
		internalAxis: toInternalFlexAxis(axis),
		children:     append([]Widget(nil), children...),
	}
	f.items = make([]internal.FlexChild, len(f.children))
	for index, child := range f.children {
		idx := index
		if flexed, ok := child.(*flexedWidget); ok {
			f.items[index] = internal.FlexChild{
				Flexed: true,
				Weight: flexed.weight,
				Layout: func(childCtx *internal.Context) image.Point {
					current, ok := f.children[idx].(*flexedWidget)
					if !ok || current == nil || current.child == nil {
						return image.Point{}
					}
					return current.child.Layout(childCtx.Child(idx)).Size
				},
			}
			continue
		}
		f.items[index] = internal.FlexChild{
			Layout: func(childCtx *internal.Context) image.Point {
				current := f.children[idx]
				if current == nil {
					return image.Point{}
				}
				return current.Layout(childCtx.Child(idx)).Size
			},
		}
	}
	return f
}

func (f *flexWidget) Layout(ctx *internal.Context) layout.Dimensions {
	return layout.Dimensions{Size: ctx.LayoutFlex(f.internalAxis, f.items...)}
}

func (f *flexedWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if f.child == nil {
		return layout.Dimensions{}
	}
	return f.child.Layout(ctx.Child(0))
}

func toInternalFlexAxis(axis layout.Axis) internal.Axis {
	if axis == layout.Vertical {
		return internal.Vertical
	}
	return internal.Horizontal
}
