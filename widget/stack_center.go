package widget

import (
	"fluxui/internal"
	"fluxui/layout"

	gioLayout "gioui.org/layout"
)

type centerWidget struct {
	child Widget
}

// Center 将子组件居中布局。
func Center(child Widget) Widget {
	return &centerWidget{child: child}
}

func (c *centerWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if c.child == nil {
		return layout.Dimensions{}
	}
	dims := gioLayout.Center.Layout(ctx.Gtx, func(gtx gioLayout.Context) gioLayout.Dimensions {
		next := *ctx
		next.Gtx = gtx
		childDims := c.child.Layout(next.Child(0))
		return gioLayout.Dimensions{Size: childDims.Size}
	})
	return layout.Dimensions{Size: dims.Size}
}

type stackWidget struct {
	children []Widget
}

// Stack 将多个组件堆叠显示，第一个子项为 Expanded，其余为 Stacked。
func Stack(children ...Widget) Widget {
	return &stackWidget{
		children: append([]Widget(nil), children...),
	}
}

func (s *stackWidget) Layout(ctx *internal.Context) layout.Dimensions {
	baseCtx := ctx
	max := ctx.MaxConstraints()
	if max.X > 0 && max.Y > 0 {
		gtx := ctx.Gtx
		gtx.Constraints = gioLayout.Exact(max)
		next := *ctx
		next.Gtx = gtx
		baseCtx = &next
	}

	stackChildren := make([]layout.StackChild, 0, len(s.children))
	for idx := range s.children {
		child := s.children[idx]
		if child == nil {
			continue
		}
		current := child
		childIndex := idx
		if len(stackChildren) == 0 {
			stackChildren = append(stackChildren, layout.Expanded(func(childCtx *internal.Context) layout.Dimensions {
				return current.Layout(childCtx.Child(childIndex))
			}))
			continue
		}
		stackChildren = append(stackChildren, layout.Stacked(func(childCtx *internal.Context) layout.Dimensions {
			return current.Layout(childCtx.Child(childIndex))
		}))
	}
	if len(stackChildren) == 0 {
		return layout.Dimensions{}
	}
	return layout.Stack(baseCtx, stackChildren...)
}
