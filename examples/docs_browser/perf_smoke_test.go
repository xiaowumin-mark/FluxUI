package main

import (
	"image"
	"io"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"
	ui "github.com/xiaowumin-mark/FluxUI/ui"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestPerfScenario(t *testing.T) {
	docs, err := loadWidgetDocs()
	if err != nil {
		t.Fatalf("load docs: %v", err)
	}

	runtime := internal.NewRuntime(ui.NewTheme(ui.LightColors()))
	defer runtime.Dispose()
	runtime.SetPerfDiagnostics(internal.PerfDiagnostics{
		Enabled:          true,
		MeasureDurations: true,
		LogRedrawReasons: true,
		Writer:           io.Discard,
	})

	root := ui.ElementRootBuilder(func(ctx *ui.Context) ui.Element {
		return docsBrowserApp(ctx, &docsRuntimeState{
			Docs:   docs,
			Source: "local",
		})
	})

	viewport := image.Pt(1360, 880)
	baseTime := time.Date(2026, 7, 1, 9, 30, 0, 0, time.UTC)

	var ops op.Ops
	var router input.Router
	moveSeq := 0
	for frame := 0; frame < 36; frame++ {
		if frame > 0 {
			runtime.QueuePointerMoveF32(f32.Pt(42+float32((frame%5)*180), 96+float32((frame%12)*48)))
		}
		if pos, ok := runtime.CoalescedPointerMove(); ok {
			moveSeq++
			router.Queue(pointer.Event{
				Kind:      pointer.Move,
				Source:    pointer.Mouse,
				PointerID: 1,
				Time:      time.Duration(moveSeq) * 16 * time.Millisecond,
				Position:  f32.Pt(float32(pos.X), float32(pos.Y)),
			})
		}

		ops.Reset()
		gtx := gioLayout.Context{
			Constraints: gioLayout.Exact(viewport),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         baseTime.Add(time.Duration(frame) * 16 * time.Millisecond),
			Source:      router.Source(),
			Ops:         &ops,
		}

		runtime.BeginFrame()
		ctx := runtime.Frame(gtx)
		layoutDone := runtime.StartFrameSection(internal.PerfLayout, 1)
		if widget := root(ctx.Scope("build")); widget != nil {
			widget.Layout(ctx.Scope("layout"))
		}
		if layoutDone != nil {
			layoutDone()
		}
		runtime.EndFrame()
		router.Frame(&ops)
	}

	stats := runtime.LastFrameStats()
	assertDocsPerfStats(t, stats)
	t.Log(internal.FormatFrameStats(stats))
}

func assertDocsPerfStats(t *testing.T, stats internal.FrameStats) {
	t.Helper()
	if stats.Frame == 0 {
		t.Fatal("expected recorded frame stats")
	}
	if stats.Layout.Count == 0 {
		t.Fatalf("expected layout stats, got %s", internal.FormatFrameStats(stats))
	}
	if stats.Text.Count == 0 && stats.Cache.TextHits == 0 {
		t.Fatalf("expected text layout or text cache stats, got %s", internal.FormatFrameStats(stats))
	}
	if stats.Input.Count == 0 {
		t.Fatalf("expected input stats, got %s", internal.FormatFrameStats(stats))
	}
	if len(stats.Reasons) == 0 {
		t.Fatalf("expected redraw reasons, got %s", internal.FormatFrameStats(stats))
	}
	if stats.Virtualization.TotalItems == 0 || stats.Virtualization.VisibleItems == 0 {
		t.Fatalf("expected virtualized docs browser stats, got %s", internal.FormatFrameStats(stats))
	}
	if stats.Virtualization.VisibleItems >= stats.Virtualization.TotalItems {
		t.Fatalf("expected docs browser to build only visible list items, got %s", internal.FormatFrameStats(stats))
	}
	if stats.Virtualization.CulledItems == 0 {
		t.Fatalf("expected docs browser to cull offscreen list items, got %s", internal.FormatFrameStats(stats))
	}
	if stats.Cache.TextHits == 0 {
		t.Fatalf("expected docs browser smoke to hit text cache, got %s", internal.FormatFrameStats(stats))
	}
}
