package state_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/state"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func newTestCtx() (*internal.Runtime, *internal.Context) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	return rt, ctx
}

func TestStateInitialValue(t *testing.T) {
	_, ctx := newTestCtx()
	s := state.Use[int](ctx)
	if s.Value() != 0 {
		t.Fatalf("expected 0, got %d", s.Value())
	}
}

func TestStateSetAndGet(t *testing.T) {
	_, ctx := newTestCtx()
	s := state.Use[int](ctx)
	s.Set(42)
	if s.Value() != 42 {
		t.Fatalf("expected 42, got %d", s.Value())
	}
}

func TestStateStringType(t *testing.T) {
	_, ctx := newTestCtx()
	s := state.Use[string](ctx)
	if s.Value() != "" {
		t.Fatalf("expected empty string, got %q", s.Value())
	}
	s.Set("hello")
	if s.Value() != "hello" {
		t.Fatalf("expected \"hello\", got %q", s.Value())
	}
}

func TestStateKey(t *testing.T) {
	_, ctx := newTestCtx()
	s := state.Use[int](ctx)
	if s.Key() == "" {
		t.Fatal("expected non-empty key")
	}
}

func TestStateNilSafety(t *testing.T) {
	var s *state.State[int]
	if s.Key() != "" {
		t.Fatal("nil State.Key() should return empty string")
	}
	if s.Value() != 0 {
		t.Fatal("nil State.Value() should return zero value")
	}
	s.Set(1) // should not panic
}

func TestMultipleStates(t *testing.T) {
	_, ctx := newTestCtx()
	a := state.Use[int](ctx)
	b := state.Use[int](ctx)

	if a.Key() == b.Key() {
		t.Fatal("two states should have different keys")
	}

	a.Set(10)
	b.Set(20)
	if a.Value() != 10 {
		t.Fatalf("a: expected 10, got %d", a.Value())
	}
	if b.Value() != 20 {
		t.Fatalf("b: expected 20, got %d", b.Value())
	}
}

func TestStatePersistAcrossFrames(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}

	// Frame 1: create state and set value
	rt.BeginFrame()
	ctx1 := internal.NewContext(gtx, rt)
	s1 := state.Use[int](ctx1)
	s1.Set(99)
	rt.EndFrame()

	// Frame 2: retrieve the same state, value should persist
	rt.BeginFrame()
	ctx2 := internal.NewContext(gtx, rt)
	s2 := state.Use[int](ctx2)
	rt.EndFrame()

	if s2.Value() != 99 {
		t.Fatalf("expected 99 across frames, got %d", s2.Value())
	}
}

func TestStateTypeMismatchPanics(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}

	// Frame 1: create State[int]
	rt.BeginFrame()
	ctx1 := internal.NewContext(gtx, rt)
	state.Use[int](ctx1)
	rt.EndFrame()

	// Frame 2: same key but different type — should panic
	rt.BeginFrame()
	ctx2 := internal.NewContext(gtx, rt)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on type mismatch")
		}
	}()
	state.Use[string](ctx2)
}

func TestSetTriggersRedraw(t *testing.T) {
	rt, ctx := newTestCtx()
	var called atomic.Int32
	rt.SetInvalidator(func() {
		called.Add(1)
	})

	s := state.Use[int](ctx)
	s.Set(1)

	if called.Load() == 0 {
		t.Fatal("expected invalidator to be called after Set")
	}
}

func TestSetWithoutInvalidator(t *testing.T) {
	_, ctx := newTestCtx()
	s := state.Use[int](ctx)
	s.Set(1) // should not panic even without invalidator
	if s.Value() != 1 {
		t.Fatalf("expected 1, got %d", s.Value())
	}
}

func TestStateConcurrentSetGet(t *testing.T) {
	_, ctx := newTestCtx()
	s := state.Use[int](ctx)

	const goroutines = 100
	const iterations = 1000
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := range goroutines {
		go func(id int) {
			defer wg.Done()
			for j := range iterations {
				s.Set(id*iterations + j)
				_ = s.Value()
			}
		}(i)
	}
	wg.Wait()
}

func TestStateConcurrentMultipleStates(t *testing.T) {
	_, ctx := newTestCtx()
	s1 := state.Use[int](ctx)
	s2 := state.Use[string](ctx)

	const goroutines = 50
	const iterations = 1000
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	for i := range goroutines {
		go func(id int) {
			defer wg.Done()
			for j := range iterations {
				s1.Set(id*iterations + j)
				_ = s1.Value()
			}
		}(i)
	}
	for range goroutines {
		go func() {
			defer wg.Done()
			for range iterations {
				s2.Set("test")
				_ = s2.Value()
			}
		}()
	}
	wg.Wait()
}

func TestHookCountConsistencyPanics(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}

	// Frame 1: call 2 hooks
	rt.BeginFrame()
	ctx1 := internal.NewContext(gtx, rt)
	state.Use[int](ctx1)
	state.Use[int](ctx1)
	rt.EndFrame()

	// Frame 2: call only 1 hook on the same path — should panic in EndFrame
	rt.BeginFrame()
	ctx2 := internal.NewContext(gtx, rt)
	state.Use[int](ctx2)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on hook count mismatch")
		}
		msg, ok := r.(string)
		if !ok || !contains(msg, "hooks must not be called conditionally") {
			t.Fatalf("unexpected panic message: %v", r)
		}
	}()
	rt.EndFrame()
}

func TestHookUnmountDoesNotPanic(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}

	// Frame 1: root has hooks, child scope also has hooks
	rt.BeginFrame()
	ctx1 := internal.NewContext(gtx, rt)
	state.Use[int](ctx1)
	child1 := ctx1.Child(0)
	state.Use[string](child1)
	rt.EndFrame()

	// Frame 2: root still has hooks, but child scope is gone (unmounted)
	rt.BeginFrame()
	ctx2 := internal.NewContext(gtx, rt)
	state.Use[int](ctx2)
	// no child scope this frame — should NOT panic
	rt.EndFrame()
}

func TestHookCountConsistentAcrossFrames(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}

	for range 5 {
		rt.BeginFrame()
		ctx := internal.NewContext(gtx, rt)
		state.Use[int](ctx)
		state.Use[string](ctx)
		state.Use[bool](ctx)
		rt.EndFrame()
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
