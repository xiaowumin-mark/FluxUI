package widget

import (
	"image"
	"testing"
	"time"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"

	"gioui.org/f32"
	gioEvent "gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestPointerAreaDispatchesPointerWheelAndSyntheticEvents(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	baseTime := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)

	var target fluxevent.TargetID
	var down fluxevent.PointerEvent
	var wheel fluxevent.WheelEvent
	clicks := 0
	dblClicks := 0
	auxClicks := 0
	contextMenus := 0

	render := func(frame int, events ...gioEvent.Event) {
		for _, ev := range events {
			router.Queue(ev)
		}

		ops.Reset()
		gtx := gioLayout.Context{
			Constraints: gioLayout.Exact(image.Pt(160, 120)),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         baseTime.Add(time.Duration(frame) * 16 * time.Millisecond),
			Source:      router.Source(),
			Ops:         &ops,
		}

		rt.BeginFrame()
		ctx := internal.NewContext(gtx, rt)
		PointerArea(
			Spacer(120, 80),
			PointerCaptureOnPress(true),
			PointerOnDown(func(ctx *internal.Context, ev *fluxevent.PointerEvent) {
				target = ctx.PathID()
				down = *ev
			}),
			PointerOnClick(func(ctx *internal.Context, ev *fluxevent.PointerEvent) {
				clicks++
			}),
			PointerOnDoubleClick(func(ctx *internal.Context, ev *fluxevent.PointerEvent) {
				dblClicks++
			}),
			PointerOnAuxClick(func(ctx *internal.Context, ev *fluxevent.PointerEvent) {
				auxClicks++
			}),
			PointerOnContextMenu(func(ctx *internal.Context, ev *fluxevent.PointerEvent) {
				contextMenus++
			}),
			PointerOnWheel(func(ctx *internal.Context, ev *fluxevent.WheelEvent) {
				wheel = *ev
			}),
		).Layout(ctx.Scope("pointer-area"))
		rt.EndFrame()
		router.Frame(&ops)
	}

	render(0)
	render(1, pointerAreaPointer(pointer.Press, pointer.ButtonSecondary, 7, 18, 24, 16*time.Millisecond, key.ModCtrl|key.ModShift))

	if down.PointerID != 7 || down.PointerType != fluxevent.PointerMouse {
		t.Fatalf("down pointer identity = %d/%s, want 7/mouse", down.PointerID, down.PointerType)
	}
	if down.Position != f32.Pt(18, 24) {
		t.Fatalf("down position = %v, want (18,24)", down.Position)
	}
	if down.Button != fluxevent.ButtonSecondary || !down.Buttons.Contain(fluxevent.ButtonsSecondary) {
		t.Fatalf("down button/buttons = %v/%v, want secondary", down.Button, down.Buttons)
	}
	if !down.Modifiers.Ctrl || !down.Modifiers.Shift {
		t.Fatalf("down modifiers = %+v, want ctrl+shift", down.Modifiers)
	}
	if captured, ok := rt.PointerCaptureTarget(7); !ok || captured != target {
		t.Fatalf("pointer capture = %v/%v, want target %v", captured, ok, target)
	}

	render(2, pointerAreaPointer(pointer.Release, 0, 7, 18, 24, 32*time.Millisecond, 0))
	if _, ok := rt.PointerCaptureTarget(7); ok {
		t.Fatal("pointer capture still active after release")
	}
	if auxClicks != 1 || contextMenus != 1 {
		t.Fatalf("secondary release aux/context = %d/%d, want 1/1", auxClicks, contextMenus)
	}

	render(3, pointerAreaPointer(pointer.Press, pointer.ButtonTertiary, 7, 22, 24, 48*time.Millisecond, 0))
	render(4, pointerAreaPointer(pointer.Release, 0, 7, 22, 24, 64*time.Millisecond, 0))
	if auxClicks != 2 || contextMenus != 1 {
		t.Fatalf("middle release aux/context = %d/%d, want 2/1", auxClicks, contextMenus)
	}

	render(5, pointerAreaPointer(pointer.Press, pointer.ButtonPrimary, 7, 20, 24, 80*time.Millisecond, 0))
	render(6, pointerAreaPointer(pointer.Release, 0, 7, 20, 24, 96*time.Millisecond, 0))
	render(7, pointerAreaPointer(pointer.Press, pointer.ButtonPrimary, 7, 20, 24, 112*time.Millisecond, 0))
	render(8, pointerAreaPointer(pointer.Release, 0, 7, 20, 24, 128*time.Millisecond, 0))
	if clicks != 2 || dblClicks != 1 {
		t.Fatalf("primary click/dblclick = %d/%d, want 2/1", clicks, dblClicks)
	}

	render(9, pointer.Event{
		Kind:      pointer.Scroll,
		Source:    pointer.Mouse,
		PointerID: 7,
		Time:      144 * time.Millisecond,
		Position:  f32.Pt(30, 40),
		Scroll:    f32.Pt(4, -12),
		Modifiers: key.ModAlt,
	})
	if wheel.DeltaX != 4 || wheel.DeltaY != -12 || wheel.Position != f32.Pt(30, 40) || !wheel.Modifiers.Alt {
		t.Fatalf("wheel = %+v, want delta(4,-12), pos(30,40), alt", wheel)
	}
}

func TestPointerAreaCoalescesMovesToLatestSample(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	baseTime := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	var moves []fluxevent.PointerEvent

	render := func(frame int, events ...gioEvent.Event) {
		for _, ev := range events {
			router.Queue(ev)
		}

		ops.Reset()
		gtx := gioLayout.Context{
			Constraints: gioLayout.Exact(image.Pt(160, 120)),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         baseTime.Add(time.Duration(frame) * 16 * time.Millisecond),
			Source:      router.Source(),
			Ops:         &ops,
		}

		rt.BeginFrame()
		ctx := internal.NewContext(gtx, rt)
		PointerArea(
			Spacer(120, 80),
			PointerOnMove(func(ctx *internal.Context, ev *fluxevent.PointerEvent) {
				moves = append(moves, *ev)
			}),
		).Layout(ctx.Scope("pointer-area"))
		rt.EndFrame()
		router.Frame(&ops)
	}

	render(0)
	render(1,
		pointerAreaPointer(pointer.Move, pointer.ButtonPrimary, 3, 10, 20, 16*time.Millisecond, 0),
		pointerAreaPointer(pointer.Move, pointer.ButtonPrimary, 3, 30, 40, 17*time.Millisecond, 0),
	)

	if len(moves) != 1 {
		t.Fatalf("move dispatch count = %d, want 1", len(moves))
	}
	if moves[0].Position != f32.Pt(30, 40) {
		t.Fatalf("move position = %v, want latest (30,40)", moves[0].Position)
	}
	if got := moves[0].CoalescedSamples(); len(got) != 2 || got[0].Position != f32.Pt(10, 20) || got[1].Position != f32.Pt(30, 40) {
		t.Fatalf("coalesced samples = %+v, want two samples ending at (30,40)", got)
	}
}

func pointerAreaPointer(kind pointer.Kind, buttons pointer.Buttons, pointerID pointer.ID, x, y float32, at time.Duration, modifiers key.Modifiers) pointer.Event {
	return pointer.Event{
		Kind:      kind,
		Source:    pointer.Mouse,
		PointerID: pointerID,
		Time:      at,
		Buttons:   buttons,
		Position:  f32.Pt(x, y),
		Modifiers: modifiers,
	}
}
