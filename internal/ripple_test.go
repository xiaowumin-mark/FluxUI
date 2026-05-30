package internal

import (
	"image"
	"testing"
	"time"

	"gioui.org/widget"
)

func TestCalculateRippleFrameActivePress(t *testing.T) {
	start := time.Unix(100, 0)
	frame := calculateRippleFrame(widget.Press{Start: start}, start.Add(120*time.Millisecond), image.Pt(80, 40))

	if frame.Expired {
		t.Fatal("active press should not be expired")
	}
	if frame.Diameter <= 0 {
		t.Fatalf("active press diameter = %d, want > 0", frame.Diameter)
	}
	if frame.AlphaScale <= 0 || frame.AlphaScale > 1 {
		t.Fatalf("active press alpha scale = %v, want within (0, 1]", frame.AlphaScale)
	}
	if !frame.Invalidate {
		t.Fatal("expanding active press should request invalidation")
	}
}

func TestCalculateRippleFramePressStartRequestsInvalidation(t *testing.T) {
	start := time.Unix(100, 0)
	frame := calculateRippleFrame(widget.Press{Start: start}, start, image.Pt(80, 40))

	if frame.Expired {
		t.Fatal("press start should stay alive so it can schedule the next ripple frame")
	}
	if frame.Diameter != 0 {
		t.Fatalf("press start diameter = %d, want 0 before the first animated frame", frame.Diameter)
	}
	if !frame.Invalidate {
		t.Fatal("press start should request invalidation even before the ripple has visible diameter")
	}
}

func TestCalculateRippleFrameHeldPressStopsInvalidatingAfterExpansion(t *testing.T) {
	start := time.Unix(100, 0)
	frame := calculateRippleFrame(widget.Press{Start: start}, start.Add(rippleExpandDuration+50*time.Millisecond), image.Pt(80, 40))

	if frame.Expired {
		t.Fatal("held press should remain visible")
	}
	if frame.Invalidate {
		t.Fatal("fully expanded held press should not keep invalidating")
	}
	wantMin := int(float32(80) * 2 * 1.4)
	if frame.Diameter < wantMin {
		t.Fatalf("expanded diameter = %d, want at least %d", frame.Diameter, wantMin)
	}
}

func TestCalculateRippleFrameReleasedExpiresAfterFade(t *testing.T) {
	start := time.Unix(100, 0)
	end := start.Add(100 * time.Millisecond)
	frame := calculateRippleFrame(widget.Press{Start: start, End: end}, end.Add(rippleFadeDuration/2+20*time.Millisecond), image.Pt(80, 40))

	if !frame.Expired {
		t.Fatalf("released press should expire after fade window, got %#v", frame)
	}
}

func TestCalculateRippleFrameCancelledFreezesDiameterAtCancelTime(t *testing.T) {
	start := time.Unix(100, 0)
	end := start.Add(80 * time.Millisecond)
	now := start.Add(160 * time.Millisecond)

	cancelled := calculateRippleFrame(widget.Press{Start: start, End: end, Cancelled: true}, now, image.Pt(100, 40))
	active := calculateRippleFrame(widget.Press{Start: start}, now, image.Pt(100, 40))

	if cancelled.Expired {
		t.Fatal("recently cancelled press should still fade out")
	}
	if cancelled.Diameter >= active.Diameter {
		t.Fatalf("cancelled diameter = %d, active diameter = %d; cancel should freeze expansion earlier", cancelled.Diameter, active.Diameter)
	}
}

func TestCalculateRippleFrameRejectsInvalidBoundsAndFuturePress(t *testing.T) {
	start := time.Unix(100, 0)
	cases := []struct {
		name   string
		now    time.Time
		bounds image.Point
	}{
		{name: "zero width", now: start, bounds: image.Pt(0, 40)},
		{name: "zero height", now: start, bounds: image.Pt(40, 0)},
		{name: "future start", now: start.Add(-time.Millisecond), bounds: image.Pt(40, 40)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			frame := calculateRippleFrame(widget.Press{Start: start}, tc.now, tc.bounds)
			if !frame.Expired {
				t.Fatalf("invalid frame should be expired, got %#v", frame)
			}
		})
	}
}
