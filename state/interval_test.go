package state_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/state"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestUseIntervalMountsOnceAndInvalidates(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	var ticks atomic.Int64
	var invalidates atomic.Int64
	rt.SetInvalidator(func() {
		invalidates.Add(1)
	})

	for range 3 {
		rt.BeginFrame()
		ctx := internal.NewContext(gioLayout.Context{Ops: &ops}, rt)
		state.UseInterval(ctx, time.Millisecond, func() {
			ticks.Add(1)
		})
		rt.EndFrame()
	}

	waitFor(t, time.Second, func() bool {
		return ticks.Load() > 0 && invalidates.Load() > 0
	})

	if got := ticks.Load(); got > 20 {
		t.Fatalf("expected a single mounted interval, got %d ticks", got)
	}

	rt.BeginFrame()
	rt.EndFrame()
	afterUnmount := ticks.Load()
	time.Sleep(20 * time.Millisecond)
	if got := ticks.Load(); got != afterUnmount {
		t.Fatalf("expected interval to stop after unmount, ticks changed from %d to %d", afterUnmount, got)
	}
}

func waitFor(t *testing.T, timeout time.Duration, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition was not met before timeout")
}
