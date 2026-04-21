package anim

import (
	"math"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func approxEqual(a, b, epsilon float32) bool {
	return float32(math.Abs(float64(a-b))) < epsilon
}

func TestClamp01(t *testing.T) {
	cases := []struct {
		in, out float32
	}{
		{-1, 0}, {-0.1, 0}, {0, 0}, {0.5, 0.5}, {1, 1}, {1.5, 1}, {100, 1},
	}
	for _, c := range cases {
		got := clamp01(c.in)
		if got != c.out {
			t.Errorf("clamp01(%v) = %v, want %v", c.in, got, c.out)
		}
	}
}

func TestLerp(t *testing.T) {
	if lerp(0, 10, 0) != 0 {
		t.Error("lerp at 0")
	}
	if lerp(0, 10, 1) != 10 {
		t.Error("lerp at 1")
	}
	if lerp(0, 10, 0.5) != 5 {
		t.Error("lerp at 0.5")
	}
	if lerp(10, 0, 0.5) != 5 {
		t.Error("lerp reverse at 0.5")
	}
	if lerp(-5, 5, 0.5) != 0 {
		t.Error("lerp negative range at 0.5")
	}
}

func TestLinear(t *testing.T) {
	if Linear(0) != 0 {
		t.Error("Linear(0)")
	}
	if Linear(0.5) != 0.5 {
		t.Error("Linear(0.5)")
	}
	if Linear(1) != 1 {
		t.Error("Linear(1)")
	}
	if Linear(-0.5) != 0 {
		t.Error("Linear(-0.5) should clamp to 0")
	}
	if Linear(1.5) != 1 {
		t.Error("Linear(1.5) should clamp to 1")
	}
}

func TestEaseOut(t *testing.T) {
	if EaseOut(0) != 0 {
		t.Error("EaseOut(0)")
	}
	if EaseOut(1) != 1 {
		t.Error("EaseOut(1)")
	}
	if !approxEqual(EaseOut(0.5), 0.75, 0.001) {
		t.Errorf("EaseOut(0.5) = %v, want ~0.75", EaseOut(0.5))
	}
}

func TestEaseInOut(t *testing.T) {
	if EaseInOut(0) != 0 {
		t.Error("EaseInOut(0)")
	}
	if EaseInOut(1) != 1 {
		t.Error("EaseInOut(1)")
	}
	// Continuity at 0.5 — both branches should return 0.5
	if !approxEqual(EaseInOut(0.5), 0.5, 0.001) {
		t.Errorf("EaseInOut(0.5) = %v, want 0.5 (continuity)", EaseInOut(0.5))
	}
	// First half should be less than linear
	if EaseInOut(0.25) >= 0.25 {
		t.Error("EaseInOut(0.25) should be < 0.25 (ease-in phase)")
	}
	// Second half should be greater than linear
	if EaseInOut(0.75) <= 0.75 {
		t.Error("EaseInOut(0.75) should be > 0.75 (ease-out phase)")
	}
}

func TestAnimationDefaults(t *testing.T) {
	a := New()
	if a.from != 0 {
		t.Errorf("default from = %v, want 0", a.from)
	}
	if a.to != 1 {
		t.Errorf("default to = %v, want 1", a.to)
	}
	if a.duration != 300*time.Millisecond {
		t.Errorf("default duration = %v, want 300ms", a.duration)
	}
}

func TestAnimationZeroDuration(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops, Now: time.Now()}
	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)

	a := New(Duration(0), From(10), To(20))
	v := a.Value(ctx)
	if v != 20 {
		t.Fatalf("zero duration should return to=%v, got %v", 20, v)
	}
}

func TestAnimationNil(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops, Now: time.Now()}
	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)

	var a *Animation
	if a.Value(ctx) != 0 {
		t.Fatal("nil animation should return 0")
	}
}

func TestAnimationValue(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	a := New(From(0), To(100), Duration(100*time.Millisecond))

	// Frame 1: t=0 — animation starts
	rt.BeginFrame()
	gtx0 := gioLayout.Context{Ops: &ops, Now: now}
	ctx0 := internal.NewContext(gtx0, rt)
	v0 := a.Value(ctx0)
	rt.EndFrame()

	if !approxEqual(v0, 0, 1) {
		t.Fatalf("at t=0, expected ~0, got %v", v0)
	}

	// Frame 2: t=50ms — halfway
	rt.BeginFrame()
	gtx50 := gioLayout.Context{Ops: &ops, Now: now.Add(50 * time.Millisecond)}
	ctx50 := internal.NewContext(gtx50, rt)
	v50 := a.Value(ctx50)
	rt.EndFrame()

	if !approxEqual(v50, 50, 1) {
		t.Fatalf("at t=50ms, expected ~50, got %v", v50)
	}

	// Frame 3: t=100ms — completed
	rt.BeginFrame()
	gtx100 := gioLayout.Context{Ops: &ops, Now: now.Add(100 * time.Millisecond)}
	ctx100 := internal.NewContext(gtx100, rt)
	v100 := a.Value(ctx100)
	rt.EndFrame()

	if v100 != 100 {
		t.Fatalf("at t=100ms, expected 100, got %v", v100)
	}

	// Frame 4: t=200ms — past completion, should stay at 100
	rt.BeginFrame()
	gtx200 := gioLayout.Context{Ops: &ops, Now: now.Add(200 * time.Millisecond)}
	ctx200 := internal.NewContext(gtx200, rt)
	v200 := a.Value(ctx200)
	rt.EndFrame()

	if v200 != 100 {
		t.Fatalf("past completion, expected 100, got %v", v200)
	}
}

func TestAnimationEaseNil(t *testing.T) {
	a := New(Ease(nil))
	if a.easing == nil {
		t.Fatal("Ease(nil) should not override default easing")
	}
}
