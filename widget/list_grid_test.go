package widget

import (
	"image"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func newWidgetTestContext() (*internal.Runtime, *internal.Context) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	rt.BeginFrame()
	return rt, internal.NewContext(gtx, rt)
}

func newWidgetPerfTestContext(size image.Point) (*internal.Runtime, *internal.Context) {
	rt := internal.NewRuntime(nil)
	rt.SetPerfDiagnostics(internal.PerfDiagnostics{
		Enabled:          true,
		MeasureDurations: true,
	})
	return rt, newWidgetPerfFrameContext(rt, size, time.Time{})
}

func newWidgetPerfFrameContext(rt *internal.Runtime, size image.Point, now time.Time) *internal.Context {
	var ops op.Ops
	gtx := gioLayout.Context{
		Ops:         &ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: gioLayout.Exact(size),
		Now:         now,
	}
	rt.BeginFrame()
	return internal.NewContext(gtx, rt)
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

func TestListViewVirtualizationBuildsVisibleItemsOnly(t *testing.T) {
	rt, ctx := newWidgetPerfTestContext(image.Pt(320, 160))
	built := 0
	view := ListView(10000, func(_ *internal.Context, index int) Widget {
		built++
		return layoutWidgetFunc(func(*internal.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(300, 24)}
		})
	}, ListItemSpacing(2))

	_ = view.Layout(ctx)
	rt.EndFrame()

	stats := rt.LastFrameStats()
	if built <= 0 || built >= 10000 {
		t.Fatalf("list built %d items, want visible subset", built)
	}
	if stats.Virtualization.TotalItems < 10000 || stats.Virtualization.VisibleItems <= 0 || stats.Virtualization.VisibleItems >= 10000 {
		t.Fatalf("unexpected list virtualization stats: %+v", stats.Virtualization)
	}
	if stats.Virtualization.CulledItems <= 0 {
		t.Fatalf("expected culled list items, got %+v", stats.Virtualization)
	}
}

func TestGridViewVirtualizationBuildsVisibleCellsOnly(t *testing.T) {
	rt, ctx := newWidgetPerfTestContext(image.Pt(360, 160))
	built := 0
	view := GridView(10000, 4, func(_ *internal.Context, index int) Widget {
		built++
		return layoutWidgetFunc(func(*internal.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(80, 24)}
		})
	}, GridGap(2, 2))

	_ = view.Layout(ctx)
	rt.EndFrame()

	stats := rt.LastFrameStats()
	if built <= 0 || built >= 10000 {
		t.Fatalf("grid built %d cells, want visible subset", built)
	}
	if stats.Virtualization.TotalItems < 10000 || stats.Virtualization.VisibleItems <= 0 || stats.Virtualization.VisibleItems >= 10000 {
		t.Fatalf("unexpected grid virtualization stats: %+v", stats.Virtualization)
	}
}

func TestLargeMenuAndSelectUseVirtualizedOptionLists(t *testing.T) {
	menuItems := make([]MenuItem, 1000)
	selectItems := make([]SelectOptionItem[int], 1000)
	for i := range menuItems {
		menuItems[i] = MenuItem{Key: "item", Label: "Menu item"}
		selectItems[i] = SelectOptionItem[int]{Label: "Option", Value: i}
	}

	rt, ctx := newWidgetPerfTestContext(image.Pt(420, 240))
	_ = Menu(menuItems, MenuMaxHeight(160)).Layout(ctx)
	rt.EndFrame()
	menuStats := rt.LastFrameStats().Virtualization
	if menuStats.TotalItems < 1000 || menuStats.VisibleItems <= 0 || menuStats.VisibleItems >= 1000 {
		t.Fatalf("menu should virtualize large options, got %+v", menuStats)
	}

	ref := NewSelectRef[int]()
	ref.Open()
	rt = internal.NewRuntime(nil)
	rt.SetPerfDiagnostics(internal.PerfDiagnostics{
		Enabled:          true,
		MeasureDurations: true,
	})
	baseTime := time.Unix(0, 0)
	ctx = newWidgetPerfFrameContext(rt, image.Pt(420, 240), baseTime)
	_ = Select(0, selectItems, SelectAttachRef(ref), SelectMaxHeight[int](160)).Layout(ctx)
	rt.EndFrame()
	ctx = newWidgetPerfFrameContext(rt, image.Pt(420, 240), baseTime.Add(50*time.Millisecond))
	_ = Select(0, selectItems, SelectAttachRef(ref), SelectMaxHeight[int](160)).Layout(ctx)
	rt.EndFrame()
	selectStats := rt.LastFrameStats().Virtualization
	if selectStats.TotalItems < 1000 || selectStats.VisibleItems <= 0 || selectStats.VisibleItems >= 1000 {
		t.Fatalf("select should virtualize large options, got %+v", selectStats)
	}
}

func TestNonVirtualizedLargeListReportsWarning(t *testing.T) {
	rt, ctx := newWidgetPerfTestContext(image.Pt(320, 160))
	view := ListView(300, func(_ *internal.Context, index int) Widget {
		return layoutWidgetFunc(func(*internal.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(300, 24)}
		})
	}, ListVirtualized(false))

	_ = view.Layout(ctx)
	rt.EndFrame()

	if warnings := rt.LastFrameStats().Virtualization.NonVirtualizedWarnings; warnings == 0 {
		t.Fatalf("expected non-virtualized warning, got %+v", rt.LastFrameStats().Virtualization)
	}
}
