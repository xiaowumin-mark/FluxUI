package perf

import (
	"fmt"
	"image"
	"strconv"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/style"
	fluxwidget "github.com/xiaowumin-mark/FluxUI/widget"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

type frameHarness struct {
	rt     *internal.Runtime
	ops    op.Ops
	router input.Router
	now    time.Time
	size   image.Point
}

func newFrameHarness(size image.Point) *frameHarness {
	rt := internal.NewRuntime(nil)
	rt.SetPerfDiagnostics(internal.PerfDiagnostics{
		Enabled:          true,
		MeasureDurations: true,
	})
	return &frameHarness{
		rt:   rt,
		now:  time.Unix(0, 0),
		size: size,
	}
}

func (h *frameHarness) layout(w fluxwidget.Widget, routeInput bool) internal.FrameStats {
	h.ops.Reset()
	gtx := gioLayout.Context{
		Ops:         &h.ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Now:         h.now,
		Constraints: gioLayout.Exact(h.size),
	}
	if routeInput {
		gtx.Source = h.router.Source()
	}

	h.rt.BeginFrame()
	ctx := internal.NewContext(gtx, h.rt)
	layoutDone := h.rt.StartFrameSection(internal.PerfLayout, 1)
	if w != nil {
		_ = w.Layout(ctx)
	}
	if layoutDone != nil {
		layoutDone()
	}
	h.rt.EndFrame()
	if routeInput {
		h.router.Frame(&h.ops)
	}
	h.now = h.now.Add(16 * time.Millisecond)
	return h.rt.LastFrameStats()
}

func (h *frameHarness) queuePointerMove(i int, x, y float32) {
	h.router.Queue(pointer.Event{
		Kind:      pointer.Move,
		Source:    pointer.Mouse,
		PointerID: 1,
		Time:      time.Duration(i) * time.Millisecond,
		Position:  f32.Pt(x, y),
	})
}

type frameStatsTotal struct {
	frames      int64
	frameNS     int64
	layoutNS    int64
	drawNS      int64
	animationNS int64
	stateNS     int64
	textNS      int64
	inputNS     int64
	layoutOps   int64
	drawOps     int64
	animations  int64
	stateOps    int64
	textOps     int64
	inputOps    int64
}

func (t *frameStatsTotal) add(stats internal.FrameStats) {
	t.frames++
	t.frameNS += stats.Duration.Nanoseconds()
	t.layoutNS += stats.Layout.Duration.Nanoseconds()
	t.drawNS += stats.Draw.Duration.Nanoseconds()
	t.animationNS += stats.Animation.Duration.Nanoseconds()
	t.stateNS += stats.State.Duration.Nanoseconds()
	t.textNS += stats.Text.Duration.Nanoseconds()
	t.inputNS += stats.Input.Duration.Nanoseconds()
	t.layoutOps += stats.Layout.Count
	t.drawOps += stats.Draw.Count
	t.animations += stats.Animation.Count
	t.stateOps += stats.State.Count
	t.textOps += stats.Text.Count
	t.inputOps += stats.Input.Count
}

func (t *frameStatsTotal) report(b *testing.B) {
	b.Helper()
	if t.frames == 0 {
		return
	}
	frames := float64(t.frames)
	b.ReportMetric(float64(t.frameNS)/frames, "frame-ns/frame")
	b.ReportMetric(float64(t.layoutNS)/frames, "layout-ns/frame")
	b.ReportMetric(float64(t.drawNS)/frames, "draw-ns/frame")
	b.ReportMetric(float64(t.animationNS)/frames, "animation-ns/frame")
	b.ReportMetric(float64(t.stateNS)/frames, "state-ns/frame")
	b.ReportMetric(float64(t.textNS)/frames, "text-ns/frame")
	b.ReportMetric(float64(t.inputNS)/frames, "input-ns/frame")
	b.ReportMetric(float64(t.layoutOps)/frames, "layout-ops/frame")
	b.ReportMetric(float64(t.drawOps)/frames, "draw-ops/frame")
	b.ReportMetric(float64(t.animations)/frames, "animations/frame")
	b.ReportMetric(float64(t.stateOps)/frames, "state-ops/frame")
	b.ReportMetric(float64(t.textOps)/frames, "text-ops/frame")
	b.ReportMetric(float64(t.inputOps)/frames, "input-ops/frame")
}

func BenchmarkLayoutStaticTree_1k_5k_10k(b *testing.B) {
	for _, count := range []int{1000, 5000, 10000} {
		b.Run(fmt.Sprintf("%dk", count/1000), func(b *testing.B) {
			w := staticTree(count)
			h := newFrameHarness(image.Pt(1200, 1_000_000))
			h.layout(w, false)
			var total frameStatsTotal
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				total.add(h.layout(w, false))
			}
			b.StopTimer()
			total.report(b)
		})
	}
}

func BenchmarkMouseMoveInteractiveTree_1k_5k(b *testing.B) {
	for _, count := range []int{1000, 5000} {
		b.Run(fmt.Sprintf("%dk", count/1000), func(b *testing.B) {
			w := interactiveTree(count)
			h := newFrameHarness(image.Pt(1200, 1_000_000))
			h.layout(w, true)
			var total frameStatsTotal
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				row := i % count
				h.queuePointerMove(i, 24, float32(row*36+18))
				total.add(h.layout(w, true))
			}
			b.StopTimer()
			total.report(b)
		})
	}
}

func BenchmarkHoverTargetChange(b *testing.B) {
	w := interactiveTree(1024)
	h := newFrameHarness(image.Pt(900, 40_000))
	h.layout(w, true)
	var total frameStatsTotal
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		row := (i * 17) % 1024
		h.queuePointerMove(i, 30, float32(row*36+16))
		total.add(h.layout(w, true))
	}
	b.StopTimer()
	total.report(b)
}

func BenchmarkListVirtualized_10k(b *testing.B) {
	w := fluxwidget.ListView(10000, func(_ *internal.Context, index int) fluxwidget.Widget {
		return fluxwidget.ListItem("Row " + strconv.Itoa(index))
	}, fluxwidget.ListItemSpacing(2))
	h := newFrameHarness(image.Pt(900, 640))
	h.layout(w, true)
	var total frameStatsTotal
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.queuePointerMove(i, 20, float32(24+(i%20)*28))
		total.add(h.layout(w, true))
	}
	b.StopTimer()
	total.report(b)
}

func BenchmarkTextHeavyList(b *testing.B) {
	w := textHeavyTree(750)
	h := newFrameHarness(image.Pt(1000, 1_000_000))
	h.layout(w, false)
	var total frameStatsTotal
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		total.add(h.layout(w, false))
	}
	b.StopTimer()
	total.report(b)
}

func BenchmarkMaterialStateLayerIdle(b *testing.B) {
	w := interactiveTree(2000)
	h := newFrameHarness(image.Pt(1200, 90_000))
	h.layout(w, true)
	var total frameStatsTotal
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		total.add(h.layout(w, true))
	}
	b.StopTimer()
	total.report(b)
}

func BenchmarkSelectMenuOpenLargeOptions(b *testing.B) {
	const count = 1000
	options := make([]fluxwidget.SelectOptionItem[int], count)
	for i := range options {
		options[i] = fluxwidget.SelectOptionItem[int]{
			Label: "Option " + strconv.Itoa(i),
			Value: i,
		}
	}
	ref := fluxwidget.NewSelectRef[int]()
	ref.Open()
	w := fluxwidget.Select(0, options,
		fluxwidget.SelectAttachRef(ref),
		fluxwidget.SelectMaxHeight[int](520),
	)
	h := newFrameHarness(image.Pt(480, 720))
	h.layout(w, true)
	var total frameStatsTotal
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.queuePointerMove(i, 24, float32(72+(i%16)*32))
		total.add(h.layout(w, true))
	}
	b.StopTimer()
	total.report(b)
}

func staticTree(count int) fluxwidget.Widget {
	children := make([]fluxwidget.Widget, 0, count)
	for i := 0; i < count; i++ {
		children = append(children, fluxwidget.Padding(
			style.Insets{Top: 2, Right: 8, Bottom: 2, Left: 8},
			fluxwidget.Row(
				fluxwidget.Text("Static row "+strconv.Itoa(i)),
				fluxwidget.Spacer(12, 1),
				fluxwidget.Text("metadata "+strconv.Itoa(i%17)),
			),
		))
	}
	return fluxwidget.Column(children...)
}

func interactiveTree(count int) fluxwidget.Widget {
	children := make([]fluxwidget.Widget, 0, count)
	for i := 0; i < count; i++ {
		label := "Action " + strconv.Itoa(i)
		children = append(children, fluxwidget.Button(
			fluxwidget.Text(label),
			fluxwidget.ButtonPadding(style.Symmetric(4, 10)),
			fluxwidget.OnClick(func(*internal.Context) {}),
		))
	}
	return fluxwidget.Column(children...)
}

func textHeavyTree(count int) fluxwidget.Widget {
	const paragraph = "FluxUI performance baseline text shaping paragraph with mixed punctuation, numbers 1234567890, and repeated layout content. "
	children := make([]fluxwidget.Widget, 0, count)
	for i := 0; i < count; i++ {
		children = append(children, fluxwidget.Padding(
			style.Insets{Top: 3, Right: 12, Bottom: 3, Left: 12},
			fluxwidget.Text(paragraph+strconv.Itoa(i)),
		))
	}
	return fluxwidget.Column(children...)
}
