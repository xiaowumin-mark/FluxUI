package widget

import (
	"fluxui/internal"
	"fluxui/layout"
)

type flexWidget struct {
	axis     layout.Axis
	children []Widget
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

func (f *flexWidget) Layout(ctx *internal.Context) layout.Dimensions {
	items := make([]layout.FlexChild, 0, len(f.children))
	for index, child := range f.children {
		idx := index
		current := child
		items = append(items, layout.Rigid(func(childCtx *internal.Context) layout.Dimensions {
			if current == nil {
				return layout.Dimensions{}
			}
			return current.Layout(childCtx.Child(idx))
		}))
	}
	return layout.Flex(ctx, f.axis, items...)
}
