//go:build visual

package main

import (
	"image"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"
	ui "github.com/xiaowumin-mark/FluxUI/ui"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestComponentLabSettlesWithoutIdleRedraw(t *testing.T) {
	runtime := internal.NewRuntime(ui.NewTheme(ui.LightColors()))
	defer runtime.Dispose()

	root := ui.VisualRootBuilder(App)
	baseTime := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

	var ops op.Ops
	for frame := 0; frame < 70; frame++ {
		redraws := 0
		runtime.SetInvalidator(func() {
			redraws++
		})

		ops.Reset()
		gtx := gioLayout.Context{
			Constraints: gioLayout.Exact(image.Pt(1180, 920)),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         baseTime.Add(time.Duration(frame) * 16 * time.Millisecond),
			Ops:         &ops,
		}

		runtime.BeginFrame()
		ctx := runtime.Frame(gtx)
		if widget := root(ctx.Scope("build")); widget != nil {
			widget.Layout(ctx.Scope("tree"))
		}
		runtime.EndFrame()

		if frame >= 40 && redraws != 0 {
			t.Fatalf("frame %d requested %d redraw(s) while component lab was idle", frame, redraws)
		}
	}
}
