package widget

import (
	"image"
	"testing"
	"time"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"

	"gioui.org/f32"
	gioEvent "gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
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

func TestScrollViewWheelTargetDoesNotBlockChildClick(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	now := time.Unix(0, 0)
	clicks := 0
	w := ScrollView(
		Button(Text("Tap"), OnClick(func(*internal.Context) {
			clicks++
		})),
		ScrollBarVisible(false),
	)

	render := func(frame int, events ...gioEvent.Event) {
		for _, ev := range events {
			router.Queue(ev)
		}
		ops.Reset()
		gtx := gioLayout.Context{
			Ops:         &ops,
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         now.Add(time.Duration(frame) * 16 * time.Millisecond),
			Constraints: gioLayout.Exact(image.Pt(240, 120)),
			Source:      router.Source(),
		}
		rt.BeginFrame()
		w.Layout(internal.NewContext(gtx, rt))
		rt.EndFrame()
		router.Frame(&ops)
	}

	render(0)
	render(1, pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(20, 20),
	})
	render(2, pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  f32.Pt(20, 20),
	})

	if clicks != 1 {
		t.Fatalf("scrollview child clicks = %d, want 1", clicks)
	}
}

func TestScrollViewWheelScrollsContent(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	now := time.Unix(0, 0)
	changes := 0
	lastY := float32(0)
	rows := make([]Widget, 20)
	for i := range rows {
		rows[i] = layoutWidgetFunc(func(*internal.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(200, 24)}
		})
	}
	w := ScrollView(
		Column(rows...),
		ScrollBarVisible(false),
		ScrollOnChange(func(_ *internal.Context, _, y float32) {
			changes++
			lastY = y
		}),
	)

	render := func(frame int, events ...gioEvent.Event) {
		for _, ev := range events {
			router.Queue(ev)
		}
		ops.Reset()
		gtx := gioLayout.Context{
			Ops:         &ops,
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         now.Add(time.Duration(frame) * 16 * time.Millisecond),
			Constraints: gioLayout.Exact(image.Pt(220, 72)),
			Source:      router.Source(),
		}
		rt.BeginFrame()
		w.Layout(internal.NewContext(gtx, rt))
		rt.EndFrame()
		router.Frame(&ops)
	}

	render(0)
	render(1, pointer.Event{
		Kind:     pointer.Scroll,
		Source:   pointer.Mouse,
		Position: f32.Pt(20, 20),
		Scroll:   f32.Pt(0, 48),
	})

	if changes == 0 || lastY <= 0 {
		t.Fatalf("scrollview did not report wheel scroll, changes=%d y=%0.3f", changes, lastY)
	}
}

func TestScrollViewWheelPreventDefaultBlocksOnlyWheelDefaultAction(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	now := time.Unix(0, 0)
	changes := 0
	lastY := float32(0)
	ref := NewScrollRef()
	rows := make([]Widget, 20)
	for i := range rows {
		rows[i] = layoutWidgetFunc(func(*internal.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(200, 24)}
		})
	}
	scroll := ScrollView(
		Column(rows...),
		ScrollBarVisible(false),
		ScrollAttachRef(ref),
		ScrollOnChange(func(_ *internal.Context, _, y float32) {
			changes++
			lastY = y
		}),
	)
	w := layoutWidgetFunc(func(ctx *internal.Context) layout.Dimensions {
		fluxevent.OnWheel(ctx, func(_ *internal.Context, ev *fluxevent.WheelEvent) {
			ev.PreventDefault()
		})
		return scroll.Layout(ctx)
	})

	render := func(frame int, events ...gioEvent.Event) {
		for _, ev := range events {
			router.Queue(ev)
		}
		ops.Reset()
		gtx := gioLayout.Context{
			Ops:         &ops,
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         now.Add(time.Duration(frame) * 16 * time.Millisecond),
			Constraints: gioLayout.Exact(image.Pt(220, 72)),
			Source:      router.Source(),
		}
		rt.BeginFrame()
		w.Layout(internal.NewContext(gtx, rt))
		rt.EndFrame()
		router.Frame(&ops)
	}

	render(0)
	render(1, pointer.Event{
		Kind:     pointer.Scroll,
		Source:   pointer.Mouse,
		Position: f32.Pt(20, 20),
		Scroll:   f32.Pt(0, 48),
	})
	if changes != 0 || lastY != 0 {
		t.Fatalf("prevented wheel changed scroll position, changes=%d y=%0.3f", changes, lastY)
	}

	ref.ScrollToOffset(96)
	render(2)
	if changes == 0 || lastY <= 0 {
		t.Fatalf("preventDefault blocked non-wheel ref scroll, changes=%d y=%0.3f", changes, lastY)
	}
}

func TestHorizontalScrollViewRequiresHorizontalWheelDelta(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	now := time.Unix(0, 0)
	changes := 0
	lastX := float32(0)
	items := make([]Widget, 8)
	for i := range items {
		items[i] = layoutWidgetFunc(func(*internal.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(80, 40)}
		})
	}
	w := ScrollView(
		Row(items...),
		ScrollVertical(false),
		ScrollHorizontal(true),
		ScrollBarVisible(false),
		ScrollOnChange(func(_ *internal.Context, x, _ float32) {
			changes++
			lastX = x
		}),
	)

	render := func(frame int, events ...gioEvent.Event) {
		for _, ev := range events {
			router.Queue(ev)
		}
		ops.Reset()
		gtx := gioLayout.Context{
			Ops:         &ops,
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         now.Add(time.Duration(frame) * 16 * time.Millisecond),
			Constraints: gioLayout.Exact(image.Pt(160, 48)),
			Source:      router.Source(),
		}
		rt.BeginFrame()
		w.Layout(internal.NewContext(gtx, rt))
		rt.EndFrame()
		router.Frame(&ops)
	}

	render(0)
	render(1, pointer.Event{
		Kind:     pointer.Scroll,
		Source:   pointer.Mouse,
		Position: f32.Pt(20, 20),
		Scroll:   f32.Pt(0, 48),
	})
	if changes != 0 || lastX != 0 {
		t.Fatalf("vertical wheel scrolled horizontal view, changes=%d x=%0.3f", changes, lastX)
	}

	render(2, pointer.Event{
		Kind:     pointer.Scroll,
		Source:   pointer.Mouse,
		Position: f32.Pt(20, 20),
		Scroll:   f32.Pt(48, 0),
	})
	if changes == 0 || lastX <= 0 {
		t.Fatalf("horizontal wheel did not scroll horizontal view, changes=%d x=%0.3f", changes, lastX)
	}
}

func TestHorizontalScrollViewScrollsOverClickableChild(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	now := time.Unix(0, 0)
	changes := 0
	items := make([]Widget, 8)
	for i := range items {
		items[i] = Button(Text("Item"), OnClick(func(*internal.Context) {}))
	}
	w := ScrollView(
		Row(items...),
		ScrollVertical(false),
		ScrollHorizontal(true),
		ScrollBarVisible(false),
		ScrollOnChange(func(_ *internal.Context, _ float32, _ float32) {
			changes++
		}),
	)

	render := func(frame int, events ...gioEvent.Event) {
		for _, ev := range events {
			router.Queue(ev)
		}
		ops.Reset()
		gtx := gioLayout.Context{
			Ops:         &ops,
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         now.Add(time.Duration(frame) * 16 * time.Millisecond),
			Constraints: gioLayout.Exact(image.Pt(160, 48)),
			Source:      router.Source(),
		}
		rt.BeginFrame()
		w.Layout(internal.NewContext(gtx, rt))
		rt.EndFrame()
		router.Frame(&ops)
	}

	render(0)
	render(1, pointer.Event{
		Kind:     pointer.Scroll,
		Source:   pointer.Mouse,
		Position: f32.Pt(20, 20),
		Scroll:   f32.Pt(48, 0),
	})

	if changes == 0 {
		t.Fatal("horizontal scrollview did not scroll when wheel was over clickable child")
	}
}

func TestVerticalScrollViewClickAfterScrollUsesUpdatedVisualPosition(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	now := time.Unix(0, 0)
	clicked := -1
	items := make([]Widget, 5)
	for i := range items {
		index := i
		items[i] = ClickArea(Spacer(80, 40), func(*internal.Context) {
			clicked = index
		})
	}
	w := ScrollView(
		Column(items...),
		ScrollBarVisible(false),
	)

	render := func(frame int, events ...gioEvent.Event) {
		for _, ev := range events {
			router.Queue(ev)
		}
		ops.Reset()
		gtx := gioLayout.Context{
			Ops:         &ops,
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         now.Add(time.Duration(frame) * 16 * time.Millisecond),
			Constraints: gioLayout.Exact(image.Pt(120, 48)),
			Source:      router.Source(),
		}
		rt.BeginFrame()
		w.Layout(internal.NewContext(gtx, rt))
		rt.EndFrame()
		router.Frame(&ops)
	}

	render(0)
	render(1, pointer.Event{
		Kind:     pointer.Scroll,
		Source:   pointer.Mouse,
		Position: f32.Pt(20, 20),
		Scroll:   f32.Pt(0, 90),
	})
	render(2, pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(20, 20),
	})
	render(3, pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  f32.Pt(20, 20),
	})

	if clicked != 2 {
		t.Fatalf("clicked item after vertical scroll = %d, want 2", clicked)
	}
}

func TestVerticalWheelOverHorizontalScrollViewScrollsOuter(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	now := time.Unix(0, 0)
	outerChanges := 0
	innerChanges := 0
	items := make([]Widget, 8)
	for i := range items {
		items[i] = layoutWidgetFunc(func(*internal.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(80, 40)}
		})
	}
	inner := FixedHeight(48, ScrollView(
		Row(items...),
		ScrollVertical(false),
		ScrollHorizontal(true),
		ScrollBarVisible(false),
		ScrollOnChange(func(_ *internal.Context, _ float32, _ float32) {
			innerChanges++
		}),
	))
	w := ScrollView(
		Column(
			inner,
			layoutWidgetFunc(func(*internal.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Pt(200, 240)}
			}),
		),
		ScrollBarVisible(false),
		ScrollOnChange(func(_ *internal.Context, _ float32, _ float32) {
			outerChanges++
		}),
	)

	render := func(frame int, events ...gioEvent.Event) {
		for _, ev := range events {
			router.Queue(ev)
		}
		ops.Reset()
		gtx := gioLayout.Context{
			Ops:         &ops,
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         now.Add(time.Duration(frame) * 16 * time.Millisecond),
			Constraints: gioLayout.Exact(image.Pt(220, 120)),
			Source:      router.Source(),
		}
		rt.BeginFrame()
		w.Layout(internal.NewContext(gtx, rt))
		rt.EndFrame()
		router.Frame(&ops)
	}

	render(0)
	render(1, pointer.Event{
		Kind:     pointer.Scroll,
		Source:   pointer.Mouse,
		Position: f32.Pt(20, 20),
		Scroll:   f32.Pt(0, 48),
	})

	if outerChanges == 0 {
		t.Fatal("outer vertical scrollview did not receive vertical wheel over horizontal child")
	}
	if innerChanges != 0 {
		t.Fatalf("horizontal inner scrollview consumed vertical wheel, changes=%d", innerChanges)
	}
}

func TestHorizontalScrollViewClickAfterScrollUsesUpdatedVisualPosition(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	now := time.Unix(0, 0)
	clicked := -1
	items := make([]Widget, 5)
	for i := range items {
		index := i
		items[i] = ClickArea(Spacer(80, 40), func(*internal.Context) {
			clicked = index
		})
	}
	w := ScrollView(
		Row(items...),
		ScrollVertical(false),
		ScrollHorizontal(true),
		ScrollBarVisible(false),
	)

	render := func(frame int, events ...gioEvent.Event) {
		for _, ev := range events {
			router.Queue(ev)
		}
		ops.Reset()
		gtx := gioLayout.Context{
			Ops:         &ops,
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         now.Add(time.Duration(frame) * 16 * time.Millisecond),
			Constraints: gioLayout.Exact(image.Pt(120, 48)),
			Source:      router.Source(),
		}
		rt.BeginFrame()
		w.Layout(internal.NewContext(gtx, rt))
		rt.EndFrame()
		router.Frame(&ops)
	}

	render(0)
	render(1, pointer.Event{
		Kind:     pointer.Scroll,
		Source:   pointer.Mouse,
		Position: f32.Pt(20, 20),
		Scroll:   f32.Pt(90, 0),
	})
	render(2, pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(20, 20),
	})
	render(3, pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  f32.Pt(20, 20),
	})

	if clicked != 1 {
		t.Fatalf("clicked item after horizontal scroll = %d, want 1", clicked)
	}
}

func TestListViewWheelScrollsContent(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	now := time.Unix(0, 0)
	reachedEnd := false
	w := ListView(
		40,
		func(_ *internal.Context, index int) Widget {
			return layoutWidgetFunc(func(*internal.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Pt(200, 24)}
			})
		},
		ListOnReachEnd(func(*internal.Context) {
			reachedEnd = true
		}),
	)

	render := func(frame int, events ...gioEvent.Event) {
		for _, ev := range events {
			router.Queue(ev)
		}
		ops.Reset()
		gtx := gioLayout.Context{
			Ops:         &ops,
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         now.Add(time.Duration(frame) * 16 * time.Millisecond),
			Constraints: gioLayout.Exact(image.Pt(220, 72)),
			Source:      router.Source(),
		}
		rt.BeginFrame()
		w.Layout(internal.NewContext(gtx, rt))
		rt.EndFrame()
		router.Frame(&ops)
	}

	render(0)
	if reachedEnd {
		t.Fatal("listview reached end before scrolling")
	}
	render(1, pointer.Event{
		Kind:     pointer.Scroll,
		Source:   pointer.Mouse,
		Position: f32.Pt(20, 20),
		Scroll:   f32.Pt(0, 1000),
	})

	if !reachedEnd {
		t.Fatal("listview did not reach end after wheel scroll")
	}
}

func TestListViewClickAfterScrollUsesUpdatedVisibleItem(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	now := time.Unix(0, 0)
	clicked := -1
	w := ListView(
		40,
		func(_ *internal.Context, index int) Widget {
			return ClickArea(Spacer(200, 40), func(*internal.Context) {
				clicked = index
			})
		},
	)

	render := func(frame int, events ...gioEvent.Event) {
		for _, ev := range events {
			router.Queue(ev)
		}
		ops.Reset()
		gtx := gioLayout.Context{
			Ops:         &ops,
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         now.Add(time.Duration(frame) * 16 * time.Millisecond),
			Constraints: gioLayout.Exact(image.Pt(220, 80)),
			Source:      router.Source(),
		}
		rt.BeginFrame()
		w.Layout(internal.NewContext(gtx, rt))
		rt.EndFrame()
		router.Frame(&ops)
	}

	render(0)
	render(1, pointer.Event{
		Kind:     pointer.Scroll,
		Source:   pointer.Mouse,
		Position: f32.Pt(20, 20),
		Scroll:   f32.Pt(0, 90),
	})
	render(2, pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(20, 20),
	})
	render(3, pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  f32.Pt(20, 20),
	})

	if clicked != 2 {
		t.Fatalf("clicked list item after scroll = %d, want 2", clicked)
	}
}

func TestNestedScrollPrefersInnerListViewUnderPointer(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	now := time.Unix(0, 0)
	outerChanges := 0
	innerReachedEnd := false
	inner := FixedHeight(72, ListView(
		40,
		func(_ *internal.Context, index int) Widget {
			return layoutWidgetFunc(func(*internal.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Pt(200, 24)}
			})
		},
		ListOnReachEnd(func(*internal.Context) {
			innerReachedEnd = true
		}),
	))
	w := ScrollView(
		Column(
			inner,
			layoutWidgetFunc(func(*internal.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Pt(200, 240)}
			}),
		),
		ScrollBarVisible(false),
		ScrollOnChange(func(_ *internal.Context, _ float32, _ float32) {
			outerChanges++
		}),
	)

	render := func(frame int, events ...gioEvent.Event) {
		for _, ev := range events {
			router.Queue(ev)
		}
		ops.Reset()
		gtx := gioLayout.Context{
			Ops:         &ops,
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         now.Add(time.Duration(frame) * 16 * time.Millisecond),
			Constraints: gioLayout.Exact(image.Pt(220, 120)),
			Source:      router.Source(),
		}
		rt.BeginFrame()
		w.Layout(internal.NewContext(gtx, rt))
		rt.EndFrame()
		router.Frame(&ops)
	}

	render(0)
	render(1, pointer.Event{
		Kind:     pointer.Scroll,
		Source:   pointer.Mouse,
		Position: f32.Pt(20, 20),
		Scroll:   f32.Pt(0, 1000),
	})

	if !innerReachedEnd {
		t.Fatal("inner listview did not receive wheel scroll")
	}
	if outerChanges != 0 {
		t.Fatalf("outer scrollview also handled inner wheel scroll, changes=%d", outerChanges)
	}
}

func TestNestedScrollPrefersInnerScrollViewUnderPointer(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	now := time.Unix(0, 0)
	outerChanges := 0
	innerChanges := 0
	rows := make([]Widget, 20)
	for i := range rows {
		rows[i] = layoutWidgetFunc(func(*internal.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(200, 24)}
		})
	}
	inner := FixedHeight(72, ScrollView(
		Column(rows...),
		ScrollBarVisible(false),
		ScrollOnChange(func(_ *internal.Context, _ float32, _ float32) {
			innerChanges++
		}),
	))
	w := ScrollView(
		Column(
			inner,
			layoutWidgetFunc(func(*internal.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Pt(200, 240)}
			}),
		),
		ScrollBarVisible(false),
		ScrollOnChange(func(_ *internal.Context, _ float32, _ float32) {
			outerChanges++
		}),
	)

	render := func(frame int, events ...gioEvent.Event) {
		for _, ev := range events {
			router.Queue(ev)
		}
		ops.Reset()
		gtx := gioLayout.Context{
			Ops:         &ops,
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         now.Add(time.Duration(frame) * 16 * time.Millisecond),
			Constraints: gioLayout.Exact(image.Pt(220, 120)),
			Source:      router.Source(),
		}
		rt.BeginFrame()
		w.Layout(internal.NewContext(gtx, rt))
		rt.EndFrame()
		router.Frame(&ops)
	}

	render(0)
	render(1, pointer.Event{
		Kind:     pointer.Scroll,
		Source:   pointer.Mouse,
		Position: f32.Pt(20, 20),
		Scroll:   f32.Pt(0, 48),
	})

	if innerChanges == 0 {
		t.Fatal("inner scrollview did not receive wheel scroll")
	}
	if outerChanges != 0 {
		t.Fatalf("outer scrollview also handled inner wheel scroll, changes=%d", outerChanges)
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
