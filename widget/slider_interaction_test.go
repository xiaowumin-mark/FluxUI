package widget

import (
	"image"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"

	"gioui.org/f32"
	gioEvent "gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestSliderDragDispatchesOnChange(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	value := float32(25)
	var changes []float32
	baseTime := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

	render := func(frame int, events ...gioEvent.Event) {
		for _, ev := range events {
			router.Queue(ev)
		}

		ops.Reset()
		gtx := gioLayout.Context{
			Constraints: gioLayout.Exact(image.Pt(240, 80)),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         baseTime.Add(time.Duration(frame) * 16 * time.Millisecond),
			Source:      router.Source(),
			Ops:         &ops,
		}

		rt.BeginFrame()
		ctx := internal.NewContext(gtx, rt)
		Slider(
			value,
			SliderMin(0),
			SliderMax(100),
			SliderStep(0),
			SliderWidth(200),
			SliderOnChange(func(ctx *internal.Context, next float32) {
				value = next
				changes = append(changes, next)
			}),
		).Layout(ctx.Scope("slider"))
		rt.EndFrame()
		router.Frame(&ops)
	}

	render(0)
	render(1, sliderPointer(pointer.Press, 40, 10, 16*time.Millisecond))
	render(2, sliderPointer(pointer.Move, 170, 10, 32*time.Millisecond))
	render(3, sliderPointer(pointer.Release, 170, 10, 48*time.Millisecond))

	if len(changes) == 0 {
		t.Fatal("expected SliderOnChange to be called after dragging")
	}
	if value < 70 {
		t.Fatalf("expected slider value to follow drag, got %.2f after changes %v", value, changes)
	}
}

func sliderPointer(kind pointer.Kind, x, y float32, at time.Duration) pointer.Event {
	event := pointer.Event{
		Kind:      kind,
		Source:    pointer.Mouse,
		PointerID: 1,
		Time:      at,
		Position:  f32.Pt(x, y),
	}
	if kind == pointer.Press || kind == pointer.Move {
		event.Buttons = pointer.ButtonPrimary
	}
	return event
}
