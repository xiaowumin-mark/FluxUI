package widget

import (
	internal "github.com/xiaowumin-mark/FluxUI/internal"
	layout "github.com/xiaowumin-mark/FluxUI/layout"
)

type flexWidget struct {
	axis     layout.Axis
	children []Widget
}

type flexedWidget struct {
	weight float32
	child  Widget
}

// Row 创建横向布局。
func Row(children ...Widget) Widget {
	return &flexWidget{
		axis:     layout.Horizontal,
		children: append([]Widget(nil), children...),
	}
}

// Column 创建纵向布局。
func Column(children ...Widget) Widget {
	return &flexWidget{
		axis:     layout.Vertical,
		children: append([]Widget(nil), children...),
	}
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

func (f *flexWidget) Layout(ctx *internal.Context) layout.Dimensions {
	items := make([]layout.FlexChild, 0, len(f.children))
	for index, child := range f.children {
		idx := index
		current := child
		if flexed, ok := current.(*flexedWidget); ok {
			weight := flexed.weight
			inner := flexed.child
			items = append(items, layout.Flexed(weight, func(childCtx *internal.Context) layout.Dimensions {
				if inner == nil {
					return layout.Dimensions{}
				}
				return inner.Layout(childCtx.Child(idx))
			}))
			continue
		}

		items = append(items, layout.Rigid(func(childCtx *internal.Context) layout.Dimensions {
			if current == nil {
				return layout.Dimensions{}
			}
			return current.Layout(childCtx.Child(idx))
		}))
	}
	return layout.Flex(ctx, f.axis, items...)
}

func (f *flexedWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if f.child == nil {
		return layout.Dimensions{}
	}
	return f.child.Layout(ctx.Child(0))
}
