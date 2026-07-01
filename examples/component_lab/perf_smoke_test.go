package main

import (
	"fmt"
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
	runtime := internal.NewRuntime(ui.NewTheme(ui.LightColors()))
	defer runtime.Dispose()
	runtime.SetPerfDiagnostics(internal.PerfDiagnostics{
		Enabled:          true,
		MeasureDurations: true,
		LogRedrawReasons: true,
		Writer:           io.Discard,
	})

	root := ui.ElementRootBuilder(componentLabPerfScenario)
	viewport := image.Pt(1180, 920)
	baseTime := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)

	var ops op.Ops
	var router input.Router
	for frame := 0; frame < 60; frame++ {
		if frame > 0 {
			y := float32(88 + (frame%20)*38)
			router.Queue(pointer.Event{
				Kind:      pointer.Move,
				Source:    pointer.Mouse,
				PointerID: 1,
				Time:      time.Duration(frame) * 16 * time.Millisecond,
				Position:  f32.Pt(64+float32((frame%4)*230), y),
			})
			runtime.RecordRedrawReason("pointer.move")
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
			widget.Layout(ctx.Scope("tree"))
		}
		if layoutDone != nil {
			layoutDone()
		}
		runtime.EndFrame()
		router.Frame(&ops)
	}

	stats := runtime.LastFrameStats()
	assertPerfStats(t, stats)
	t.Log(internal.FormatFrameStats(stats))
}

func componentLabPerfScenario(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	rows := make([]ui.Element, 0, 160)
	for row := 0; row < 160; row++ {
		cells := make([]ui.Element, 0, 4)
		for col := 0; col < 4; col++ {
			index := row*4 + col
			cells = append(cells,
				ui.FixedWidthElement(220,
					ui.FilledTonalButtonElement(
						ui.TextElement(fmt.Sprintf("Component %03d", index), ui.TextType(th.Types.LabelLarge)),
						ui.OnHover(func(*ui.Context, bool) {}),
					),
				),
			)
			if col < 3 {
				cells = append(cells, ui.HSpacerElement(10))
			}
		}
		rows = append(rows, ui.PaddingElement(ui.Insets{Bottom: 8}, ui.RowElement(cells...)))
	}

	return ui.ThemeProviderElement(th,
		ui.ContainerDecorationElement(
			ui.Bg(th.Colors.Surface).WithPad(ui.All(16)),
			ui.ScrollViewElement(
				ui.ColumnElement(
					ui.TextElement("Component Lab Performance Smoke", ui.TextType(th.Types.TitleLarge), ui.TextColor(th.Colors.OnSurface)),
					ui.VSpacerElement(12),
					ui.ColumnElement(rows...),
				),
				ui.ScrollBarVisible(true),
			),
		),
	)
}

func assertPerfStats(t *testing.T, stats internal.FrameStats) {
	t.Helper()
	if stats.Frame == 0 {
		t.Fatal("expected recorded frame stats")
	}
	if stats.Layout.Count == 0 {
		t.Fatalf("expected layout stats, got %s", internal.FormatFrameStats(stats))
	}
	if stats.Layout.Duration == 0 {
		t.Fatalf("expected layout duration, got %s", internal.FormatFrameStats(stats))
	}
	if stats.Text.Count == 0 {
		t.Fatalf("expected text stats, got %s", internal.FormatFrameStats(stats))
	}
	if stats.Input.Count == 0 {
		t.Fatalf("expected input stats, got %s", internal.FormatFrameStats(stats))
	}
	if len(stats.Reasons) == 0 {
		t.Fatalf("expected redraw reasons, got %s", internal.FormatFrameStats(stats))
	}
}
