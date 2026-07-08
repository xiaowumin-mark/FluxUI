package widget

import (
	"image"
	"testing"
	"time"

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

type scrollViewBenchHarness struct {
	rt     *internal.Runtime
	router input.Router
	ops    op.Ops
	now    time.Time
	size   image.Point
}

func newScrollViewBenchHarness(size image.Point) *scrollViewBenchHarness {
	rt := internal.NewRuntime(nil)
	rt.SetPerfDiagnostics(internal.PerfDiagnostics{
		Enabled:          true,
		MeasureDurations: true,
	})
	return &scrollViewBenchHarness{
		rt:   rt,
		now:  time.Unix(0, 0),
		size: size,
	}
}

func (h *scrollViewBenchHarness) layout(w Widget, events ...gioEvent.Event) internal.FrameStats {
	for _, ev := range events {
		h.router.Queue(ev)
	}
	h.ops.Reset()
	gtx := gioLayout.Context{
		Constraints: gioLayout.Exact(h.size),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Now:         h.now,
		Source:      h.router.Source(),
		Ops:         &h.ops,
	}
	h.rt.BeginFrame()
	ctx := internal.NewContext(gtx, h.rt).Scope("scroll-view-bench")
	layoutDone := h.rt.StartFrameSection(internal.PerfLayout, 1)
	if w != nil {
		w.Layout(ctx)
	}
	if layoutDone != nil {
		layoutDone()
	}
	h.rt.EndFrame()
	h.router.Frame(&h.ops)
	h.now = h.now.Add(16 * time.Millisecond)
	return h.rt.LastFrameStats()
}

func BenchmarkWheelScrollViewVertical(b *testing.B) {
	var changes int64
	var lastY float32
	w := verticalScrollBenchWidget(func(_ *internal.Context, _, y float32) {
		changes++
		lastY = y
	})

	if !scrollBenchScrollsVertical() {
		b.Fatal("vertical ScrollView benchmark fixture did not scroll")
	}

	h := newScrollViewBenchHarness(image.Pt(220, 72))
	h.layout(w)
	var totals scrollBenchTotals
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		delta := float32(48)
		if i%2 == 1 {
			delta = -48
		}
		totals.add(h.layout(w, pointer.Event{
			Kind:     pointer.Scroll,
			Source:   pointer.Mouse,
			Position: f32.Pt(20, 20),
			Scroll:   f32.Pt(0, delta),
		}))
	}
	b.StopTimer()
	totals.report(b)
	if b.N > 0 {
		b.ReportMetric(float64(changes)/float64(b.N), "scroll-callbacks/op")
	}
	b.ReportMetric(float64(lastY), "last-y")
}

func BenchmarkHorizontalWheelDelta(b *testing.B) {
	b.Run("vertical-ignored", func(b *testing.B) {
		var changes int64
		w := horizontalScrollBenchWidget(func(*internal.Context, float32, float32) {
			changes++
		})
		h := newScrollViewBenchHarness(image.Pt(160, 48))
		h.layout(w)
		var totals scrollBenchTotals
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			totals.add(h.layout(w, pointer.Event{
				Kind:     pointer.Scroll,
				Source:   pointer.Mouse,
				Position: f32.Pt(20, 20),
				Scroll:   f32.Pt(0, 48),
			}))
		}
		b.StopTimer()
		totals.report(b)
		if b.N > 0 {
			b.ReportMetric(float64(changes)/float64(b.N), "scroll-callbacks/op")
		}
	})

	b.Run("horizontal-consumed", func(b *testing.B) {
		var changes int64
		var lastX float32
		w := horizontalScrollBenchWidget(func(_ *internal.Context, x, _ float32) {
			changes++
			lastX = x
		})
		if !scrollBenchScrollsHorizontal() {
			b.Fatal("horizontal ScrollView benchmark fixture did not scroll")
		}

		h := newScrollViewBenchHarness(image.Pt(160, 48))
		h.layout(w)
		var totals scrollBenchTotals
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			delta := float32(48)
			if i%2 == 1 {
				delta = -48
			}
			totals.add(h.layout(w, pointer.Event{
				Kind:     pointer.Scroll,
				Source:   pointer.Mouse,
				Position: f32.Pt(20, 20),
				Scroll:   f32.Pt(delta, 0),
			}))
		}
		b.StopTimer()
		totals.report(b)
		if b.N > 0 {
			b.ReportMetric(float64(changes)/float64(b.N), "scroll-callbacks/op")
		}
		b.ReportMetric(float64(lastX), "last-x")
	})
}

func verticalScrollBenchWidget(onChange func(*internal.Context, float32, float32)) Widget {
	rows := make([]Widget, 64)
	for i := range rows {
		rows[i] = layoutWidgetFunc(func(*internal.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(200, 24)}
		})
	}
	return ScrollView(
		Column(rows...),
		ScrollBarVisible(false),
		ScrollOnChange(onChange),
	)
}

func horizontalScrollBenchWidget(onChange func(*internal.Context, float32, float32)) Widget {
	items := make([]Widget, 16)
	for i := range items {
		items[i] = layoutWidgetFunc(func(*internal.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(80, 40)}
		})
	}
	return ScrollView(
		Row(items...),
		ScrollVertical(false),
		ScrollHorizontal(true),
		ScrollBarVisible(false),
		ScrollOnChange(onChange),
	)
}

func scrollBenchScrollsVertical() bool {
	var changed bool
	w := verticalScrollBenchWidget(func(*internal.Context, float32, float32) {
		changed = true
	})
	h := newScrollViewBenchHarness(image.Pt(220, 72))
	h.layout(w)
	h.layout(w, pointer.Event{
		Kind:     pointer.Scroll,
		Source:   pointer.Mouse,
		Position: f32.Pt(20, 20),
		Scroll:   f32.Pt(0, 48),
	})
	return changed
}

func scrollBenchScrollsHorizontal() bool {
	var changed bool
	w := horizontalScrollBenchWidget(func(*internal.Context, float32, float32) {
		changed = true
	})
	h := newScrollViewBenchHarness(image.Pt(160, 48))
	h.layout(w)
	h.layout(w, pointer.Event{
		Kind:     pointer.Scroll,
		Source:   pointer.Mouse,
		Position: f32.Pt(20, 20),
		Scroll:   f32.Pt(48, 0),
	})
	return changed
}

type scrollBenchTotals struct {
	frames    int64
	layoutOps int64
	inputOps  int64
	layoutNS  int64
	inputNS   int64
}

func (t *scrollBenchTotals) add(stats internal.FrameStats) {
	t.frames++
	t.layoutOps += stats.Layout.Count
	t.inputOps += stats.Input.Count
	t.layoutNS += stats.Layout.Duration.Nanoseconds()
	t.inputNS += stats.Input.Duration.Nanoseconds()
}

func (t scrollBenchTotals) report(b *testing.B) {
	b.Helper()
	if t.frames == 0 {
		return
	}
	frames := float64(t.frames)
	b.ReportMetric(float64(t.layoutOps)/frames, "layout-ops/frame")
	b.ReportMetric(float64(t.inputOps)/frames, "input-ops/frame")
	b.ReportMetric(float64(t.layoutNS)/frames, "layout-ns/frame")
	b.ReportMetric(float64(t.inputNS)/frames, "input-ns/frame")
}
