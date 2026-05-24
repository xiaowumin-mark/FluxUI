package widget

import (
	"image"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func newInteractiveLayoutTestContext(rt *internal.Runtime) *internal.Context {
	var ops op.Ops
	gtx := gioLayout.Context{
		Ops:         &ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: gioLayout.Exact(image.Pt(800, 600)),
	}
	return internal.NewContext(gtx, rt)
}

func TestSliderAndSwitchCanReusePathWithoutHookCountPanic(t *testing.T) {
	rt := internal.NewRuntime(nil)

	rt.BeginFrame()
	_ = Slider(40).Layout(newInteractiveLayoutTestContext(rt))
	rt.EndFrame()

	rt.BeginFrame()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("expected switching from Slider to Switch at same path not to panic, got %v", r)
		}
	}()
	_ = Switch(true).Layout(newInteractiveLayoutTestContext(rt))
	rt.EndFrame()
}

func TestDefaultInteractiveWidgetsDoNotAddDecorationPadding(t *testing.T) {
	rt := internal.NewRuntime(nil)
	rt.BeginFrame()
	ctx := newInteractiveLayoutTestContext(rt)

	if got, want := Switch(true).Layout(ctx).Size, image.Pt(50, 26); got != want {
		t.Fatalf("default switch size = %v, want %v", got, want)
	}

	rt.EndFrame()
	rt.BeginFrame()
	ctx = newInteractiveLayoutTestContext(rt)

	if got, want := Slider(40).Layout(ctx).Size, image.Pt(200, 20); got != want {
		t.Fatalf("default slider size = %v, want %v", got, want)
	}

	rt.EndFrame()
}
