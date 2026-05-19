package widget

import (
	"image"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/state"
	"github.com/xiaowumin-mark/FluxUI/style"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func newListGridMismatchTestContext() (*internal.Runtime, *internal.Context) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{
		Ops:         &ops,
		Constraints: gioLayout.Exact(image.Pt(800, 600)),
	}
	rt.BeginFrame()
	return rt, internal.NewContext(gtx, rt)
}

func TestListViewStateFollowsPositionOnInsertAndDelete(t *testing.T) {
	rt, ctx := newListGridMismatchTestContext()

	items1 := []string{"A", "B", "C"}
	observed1 := renderListViewStates(ctx, items1)
	rt.EndFrame()

	rt.BeginFrame()
	ctx2 := internal.NewContext(ctx.Gtx, rt)
	items2 := []string{"X", "A", "B", "C"}
	observed2 := renderListViewStates(ctx2, items2)
	rt.EndFrame()

	if got := observed2["X"]; got != "A" {
		t.Fatalf("expected inserted item to inherit prior index state A, got %q", got)
	}
	if got := observed2["A"]; got != "B" {
		t.Fatalf("expected shifted item A to inherit prior index state B, got %q", got)
	}
	if got := observed2["B"]; got != "C" {
		t.Fatalf("expected shifted item B to inherit prior index state C, got %q", got)
	}
	if got := observed1["A"]; got != "A" {
		t.Fatalf("expected initial state to bind to first position, got %q", got)
	}

	rt.BeginFrame()
	ctx3 := internal.NewContext(ctx.Gtx, rt)
	items3 := []string{"B", "C"}
	observed3 := renderListViewStates(ctx3, items3)
	rt.EndFrame()

	if got := observed3["B"]; got != "A" {
		t.Fatalf("expected deleted-head item B to inherit prior index state A, got %q", got)
	}
	if got := observed3["C"]; got != "B" {
		t.Fatalf("expected deleted-head item C to inherit prior index state B, got %q", got)
	}
}

func TestGridViewStateFollowsPositionOnInsertAndDelete(t *testing.T) {
	rt, ctx := newListGridMismatchTestContext()

	items1 := []string{"A", "B", "C", "D"}
	observed1 := renderGridViewStates(ctx, items1, 2)
	rt.EndFrame()

	rt.BeginFrame()
	ctx2 := internal.NewContext(ctx.Gtx, rt)
	items2 := []string{"X", "A", "B", "C", "D"}
	observed2 := renderGridViewStates(ctx2, items2, 2)
	rt.EndFrame()

	if got := observed2["X"]; got != "A" {
		t.Fatalf("expected inserted grid item X to inherit prior slot state A, got %q", got)
	}
	if got := observed2["A"]; got != "B" {
		t.Fatalf("expected shifted grid item A to inherit prior slot state B, got %q", got)
	}
	if got := observed2["B"]; got != "C" {
		t.Fatalf("expected shifted grid item B to inherit prior slot state C, got %q", got)
	}
	if got := observed2["C"]; got != "D" {
		t.Fatalf("expected shifted grid item C to inherit prior slot state D, got %q", got)
	}
	if got := observed1["A"]; got != "A" {
		t.Fatalf("expected initial grid state to bind to first slot, got %q", got)
	}

	rt.BeginFrame()
	ctx3 := internal.NewContext(ctx.Gtx, rt)
	items3 := []string{"B", "C", "D"}
	observed3 := renderGridViewStates(ctx3, items3, 2)
	rt.EndFrame()

	if got := observed3["B"]; got != "A" {
		t.Fatalf("expected deleted-head grid item B to inherit prior slot state A, got %q", got)
	}
	if got := observed3["C"]; got != "B" {
		t.Fatalf("expected deleted-head grid item C to inherit prior slot state B, got %q", got)
	}
	if got := observed3["D"]; got != "C" {
		t.Fatalf("expected deleted-head grid item D to inherit prior slot state C, got %q", got)
	}
}

func renderListViewStates(ctx *internal.Context, items []string) map[string]string {
	observed := make(map[string]string, len(items))
	view := ListView(len(items), func(listCtx *internal.Context, index int) Widget {
		id := items[index]
		return layoutWidgetFunc(func(itemCtx *internal.Context) layout.Dimensions {
			s := state.Use[string](itemCtx)
			if s.Value() == "" {
				s.Set(id)
			}
			observed[id] = s.Value()
			return layout.Dimensions{Size: image.Pt(24, 24)}
		})
	})
	_ = view.Layout(ctx)
	return observed
}

func renderGridViewStates(ctx *internal.Context, items []string, columns int) map[string]string {
	observed := make(map[string]string, len(items))
	view := GridView(len(items), columns, func(gridCtx *internal.Context, index int) Widget {
		id := items[index]
		return layoutWidgetFunc(func(itemCtx *internal.Context) layout.Dimensions {
			s := state.Use[string](itemCtx)
			if s.Value() == "" {
				s.Set(id)
			}
			observed[id] = s.Value()
			return layout.Dimensions{Size: image.Pt(24, 24)}
		})
	}, GridPadding(style.Insets{}))
	_ = view.Layout(ctx)
	return observed
}
