package widget

import (
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func newWidgetTestContext() (*internal.Runtime, *internal.Context) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	rt.BeginFrame()
	return rt, internal.NewContext(gtx, rt)
}

func TestScrollStatePersistsAcrossFrames(t *testing.T) {
	rt, ctx := newWidgetTestContext()

	state1 := scrollStateFor(ctx)
	state1.lastFirst = 7
	state1.lastOff = 12
	rt.EndFrame()

	rt.BeginFrame()
	ctx2 := internal.NewContext(ctx.Gtx, rt)
	state2 := scrollStateFor(ctx2)
	rt.EndFrame()

	if state2.lastFirst != 7 || state2.lastOff != 12 {
		t.Fatalf("expected scroll state to persist, got first=%d off=%d", state2.lastFirst, state2.lastOff)
	}
}

func TestScrollStateIsolatedByScope(t *testing.T) {
	rt, ctx := newWidgetTestContext()

	rootState := scrollStateFor(ctx)
	childState := scrollStateFor(ctx.Child(0))
	rootState.lastFirst = 1
	childState.lastFirst = 9
	rt.EndFrame()

	rt.BeginFrame()
	ctx2 := internal.NewContext(ctx.Gtx, rt)
	rootState2 := scrollStateFor(ctx2)
	childState2 := scrollStateFor(ctx2.Child(0))
	rt.EndFrame()

	if rootState2.lastFirst != 1 {
		t.Fatalf("expected root scroll state 1, got %d", rootState2.lastFirst)
	}
	if childState2.lastFirst != 9 {
		t.Fatalf("expected child scroll state 9, got %d", childState2.lastFirst)
	}
	if rootState2 == childState2 {
		t.Fatal("expected scroll states to be distinct per scope")
	}
}

func TestListAndGridStatePersistAcrossFrames(t *testing.T) {
	rt, ctx := newWidgetTestContext()

	listState1 := listViewStateFor(ctx)
	gridState1 := gridViewStateFor(ctx)
	listState1.reachCalled = true
	gridState1.reachCalled = true
	rt.EndFrame()

	rt.BeginFrame()
	ctx2 := internal.NewContext(ctx.Gtx, rt)
	listState2 := listViewStateFor(ctx2)
	gridState2 := gridViewStateFor(ctx2)
	rt.EndFrame()

	if !listState2.reachCalled {
		t.Fatal("expected list view state to persist across frames")
	}
	if !gridState2.reachCalled {
		t.Fatal("expected grid view state to persist across frames")
	}
}
