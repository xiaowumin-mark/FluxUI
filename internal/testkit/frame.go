package testkit

import (
	"image"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

// FrameHarness drives deterministic layout-only frames for internal component
// behavior tests. It has no dependency on widget or ui.
type FrameHarness struct {
	Runtime *internal.Runtime
	Size    image.Point
	Now     time.Time
	ops     op.Ops
}

// NewFrameHarness creates a harness with a 1x density metric and 16 ms frame
// cadence, suitable for deterministic interaction and animation tests.
func NewFrameHarness(size image.Point) *FrameHarness {
	if size.X < 1 {
		size.X = 1
	}
	if size.Y < 1 {
		size.Y = 1
	}
	runtime := internal.NewRuntime(nil)
	runtime.SetInvalidator(func() {})
	return &FrameHarness{
		Runtime: runtime,
		Size:    size,
		Now:     time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
}

// Frame runs one layout frame and advances time by 16 ms.
func (h *FrameHarness) Frame(layout func(*internal.Context)) {
	if h == nil || h.Runtime == nil {
		return
	}
	h.ops.Reset()
	h.Runtime.BeginFrame()
	ctx := h.Runtime.Frame(gioLayout.Context{
		Constraints: gioLayout.Exact(h.Size),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Now:         h.Now,
		Ops:         &h.ops,
	})
	if layout != nil {
		layout(ctx)
	}
	h.Runtime.EndFrame()
	h.Now = h.Now.Add(16 * time.Millisecond)
}

// Close releases runtime resources held by the harness.
func (h *FrameHarness) Close() {
	if h != nil && h.Runtime != nil {
		h.Runtime.Dispose()
		h.Runtime = nil
	}
}
