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
	"gioui.org/io/pointer"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

type pointerAreaBenchHarness struct {
	rt     *internal.Runtime
	router input.Router
	ops    op.Ops
	now    time.Time
	size   image.Point
}

func newPointerAreaBenchHarness(size image.Point) *pointerAreaBenchHarness {
	return &pointerAreaBenchHarness{
		rt:   internal.NewRuntime(nil),
		now:  time.Unix(0, 0),
		size: size,
	}
}

func (h *pointerAreaBenchHarness) layout(w Widget, events ...gioEvent.Event) {
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
	_ = w.Layout(internal.NewContext(gtx, h.rt).Scope("pointer-area-bench"))
	h.rt.EndFrame()
	h.router.Frame(&h.ops)
	h.now = h.now.Add(16 * time.Millisecond)
}

func BenchmarkPointerAreaHighFrequencyMove(b *testing.B) {
	h := newPointerAreaBenchHarness(image.Pt(320, 240))
	moves := 0
	w := PointerArea(
		Spacer(300, 200),
		PointerOnMove(func(ctx *internal.Context, ev *fluxevent.PointerEvent) {
			moves += len(ev.Coalesced)
		}),
	)
	h.layout(w)

	events := make([]gioEvent.Event, 8)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range events {
			events[j] = pointer.Event{
				Kind:      pointer.Move,
				Source:    pointer.Mouse,
				PointerID: 1,
				Time:      time.Duration(i*8+j) * time.Millisecond,
				Position:  f32.Pt(20+float32(j), 40),
			}
		}
		h.layout(w, events...)
	}
	_ = moves
}

func BenchmarkPointerAreaHighFrequencyWheel(b *testing.B) {
	h := newPointerAreaBenchHarness(image.Pt(320, 240))
	wheels := 0
	w := PointerArea(
		Spacer(300, 200),
		PointerOnWheel(func(ctx *internal.Context, ev *fluxevent.WheelEvent) {
			if ev.DeltaY != 0 {
				wheels++
			}
		}),
	)
	h.layout(w)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.layout(w, pointer.Event{
			Kind:      pointer.Scroll,
			Source:    pointer.Mouse,
			PointerID: 1,
			Time:      time.Duration(i) * time.Millisecond,
			Position:  f32.Pt(80, 64),
			Scroll:    f32.Pt(0, -12),
		})
	}
	_ = wheels
}
