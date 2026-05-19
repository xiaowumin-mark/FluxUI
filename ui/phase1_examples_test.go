package ui

import (
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestPhase1CounterExample(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}

	root := func(ctx *Context) Element {
		count := UseState(ctx, 0)
		if count.Value() == 0 {
			count.Set(1)
		}
		return Fragment(
			FromWidget(Text("Count")),
			FromWidget(Text("Ready")),
		)
	}

	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	el := root(ctx)
	if info := ElementInfo(el); info.Kind != "fragment" {
		t.Fatalf("expected fragment root, got %q", info.Kind)
	}
	if RenderElement(el) == nil {
		t.Fatal("expected counter example to render")
	}
	rt.EndFrame()

	rt.BeginFrame()
	ctx2 := internal.NewContext(gtx, rt)
	count := UseState(ctx2, 0)
	if count.Value() != 1 {
		t.Fatalf("expected persisted count 1, got %d", count.Value())
	}
	rt.EndFrame()
}

func TestPhase1KeyedListExample(t *testing.T) {
	items := []string{"a", "b", "c"}
	root := func(ctx *Context) Element {
		children := make([]Element, 0, len(items))
		for _, item := range items {
			children = append(children, Key(item, FromWidget(Text(item))))
		}
		return Fragment(children...)
	}

	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	el := root(ctx)
	if ElementInfo(el).Kind != "fragment" {
		t.Fatal("expected keyed list root to be a fragment")
	}
	if RenderElement(el) == nil {
		t.Fatal("expected keyed list example to render")
	}
	rt.EndFrame()
}

func TestPhase1FragmentExample(t *testing.T) {
	root := func(ctx *Context) Element {
		_ = ctx
		return Fragment(
			FromWidget(Text("Left")),
			FromWidget(Text("Right")),
		)
	}

	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	el := root(ctx)
	info := ElementInfo(el)
	if info.Kind != "fragment" || info.ChildCount != 2 {
		t.Fatalf("expected fragment with 2 children, got %#v", info)
	}
	if RenderElement(el) == nil {
		t.Fatal("expected fragment example to render")
	}
	rt.EndFrame()
}
