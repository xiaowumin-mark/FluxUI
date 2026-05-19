package ui

import (
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestRouteElementDefaultIdentityReusesByPattern(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	seen := make([]int, 0, 2)

	routeSpec := RouteElement("/users/:id", func(ctx *Context) Element {
		state := UseState(ctx, 1)
		seen = append(seen, state.Value())
		state.Set(state.Value() + 1)
		return nil
	})

	for range 2 {
		rt.BeginFrame()
		ctx := internal.NewContext(gtx, rt)
		renderRouteComponent(ctx, routeSpec)
		rt.EndFrame()
	}

	if len(seen) != 2 || seen[0] != 1 || seen[1] != 2 {
		t.Fatalf("expected route pattern identity to reuse state, got %v", seen)
	}
}

func TestRouteElementExplicitKeyRemounts(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	seen := make([]int, 0, 2)

	component := func(ctx *Context) Element {
		state := UseState(ctx, 1)
		seen = append(seen, state.Value())
		state.Set(state.Value() + 1)
		return nil
	}

	for _, id := range []string{"42", "99"} {
		rt.BeginFrame()
		ctx := internal.NewContext(gtx, rt)
		renderRouteComponent(ctx, RouteElement("/users/:id", component, RouteKey(id)))
		rt.EndFrame()
	}

	if len(seen) != 2 || seen[0] != 1 || seen[1] != 1 {
		t.Fatalf("expected explicit route key change to remount state, got %v", seen)
	}
}
