package widget

import (
	"image"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
)

type staticWidget struct {
	child    Widget
	depsHash uint64
}

// Static caches layout and draw ops for a subtree that has no pointer/key
// handlers. Pass values that affect the subtree output as deps.
func Static(child Widget, deps ...any) Widget {
	return &staticWidget{child: child, depsHash: internal.HashStaticDeps(deps...)}
}

func (s *staticWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if s == nil || s.child == nil {
		return layout.Dimensions{}
	}
	size := ctx.LayoutStaticSubtree(s.depsHash, func(next *internal.Context) image.Point {
		return s.child.Layout(next.Child(0)).Size
	})
	return layout.Dimensions{Size: size}
}
