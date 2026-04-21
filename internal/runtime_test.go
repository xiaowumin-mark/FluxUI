package internal

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestNewRuntimeNilTheme(t *testing.T) {
	rt := NewRuntime(nil)
	if rt == nil {
		t.Fatal("expected non-nil runtime")
	}
	if rt.Theme() == nil {
		t.Fatal("expected default theme")
	}
	if rt.MaterialTheme() == nil {
		t.Fatal("expected material theme")
	}
}

func TestBeginEndFrameNil(t *testing.T) {
	var rt *Runtime
	rt.BeginFrame()
	rt.EndFrame()
}

func TestRequestRedrawCallsInvalidator(t *testing.T) {
	rt := NewRuntime(nil)
	var called atomic.Int32
	rt.SetInvalidator(func() {
		called.Add(1)
	})
	rt.RequestRedraw()
	if called.Load() == 0 {
		t.Fatal("expected invalidator to be called")
	}
}

func TestRequestRedrawNilInvalidator(t *testing.T) {
	rt := NewRuntime(nil)
	rt.RequestRedraw()
}

func TestRequestRedrawConcurrent(t *testing.T) {
	rt := NewRuntime(nil)
	var called atomic.Int64
	rt.SetInvalidator(func() {
		called.Add(1)
	})

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				rt.RequestRedraw()
			}
		}()
	}
	wg.Wait()

	if called.Load() == 0 {
		t.Fatal("expected invalidator to be called")
	}
}

func TestUseEffectMount(t *testing.T) {
	rt := NewRuntime(nil)
	var ran bool

	rt.BeginFrame()
	rt.UseEffect("test-key", false, nil, func() func() {
		ran = true
		return nil
	})
	rt.EndFrame()

	if !ran {
		t.Fatal("expected effect to run on mount")
	}
}

func TestUseEffectCleanup(t *testing.T) {
	rt := NewRuntime(nil)
	var cleanupCalled bool

	// Frame 1: register effect with cleanup
	rt.BeginFrame()
	rt.UseEffect("eff", false, nil, func() func() {
		return func() { cleanupCalled = true }
	})
	rt.EndFrame()

	// Frame 2: same effect re-registered — cleanup from frame 1 should run
	rt.BeginFrame()
	rt.UseEffect("eff", false, nil, func() func() {
		return nil
	})
	rt.EndFrame()

	if !cleanupCalled {
		t.Fatal("expected cleanup from previous effect to be called")
	}
}

func TestUseEffectUnmount(t *testing.T) {
	rt := NewRuntime(nil)
	var cleanupCalled bool

	// Frame 1: register effect
	rt.BeginFrame()
	rt.UseEffect("eff-unmount", false, nil, func() func() {
		return func() { cleanupCalled = true }
	})
	rt.EndFrame()

	// Frame 2: don't register the effect — it's unmounted
	rt.BeginFrame()
	rt.EndFrame()

	if !cleanupCalled {
		t.Fatal("expected cleanup on unmount")
	}
}

func TestUseEffectDepsUnchanged(t *testing.T) {
	rt := NewRuntime(nil)
	runCount := 0

	for range 3 {
		rt.BeginFrame()
		rt.UseEffect("dep-key", true, []any{42}, func() func() {
			runCount++
			return nil
		})
		rt.EndFrame()
	}

	if runCount != 1 {
		t.Fatalf("expected effect to run once (mount only), ran %d times", runCount)
	}
}

func TestUseEffectDepsChanged(t *testing.T) {
	rt := NewRuntime(nil)
	runCount := 0

	for i := range 3 {
		rt.BeginFrame()
		rt.UseEffect("dep-change", true, []any{i}, func() func() {
			runCount++
			return nil
		})
		rt.EndFrame()
	}

	if runCount != 3 {
		t.Fatalf("expected effect to run 3 times (deps changed each frame), ran %d", runCount)
	}
}

func TestUseEffectNoDeps(t *testing.T) {
	rt := NewRuntime(nil)
	runCount := 0

	for range 3 {
		rt.BeginFrame()
		rt.UseEffect("no-deps", false, nil, func() func() {
			runCount++
			return nil
		})
		rt.EndFrame()
	}

	if runCount != 3 {
		t.Fatalf("expected effect to run every frame (no deps), ran %d", runCount)
	}
}

func TestUseEffectSkipNilSetup(t *testing.T) {
	rt := NewRuntime(nil)
	rt.BeginFrame()
	rt.UseEffect("", false, nil, func() func() {
		t.Fatal("should not run with empty key")
		return nil
	})
	rt.UseEffect("valid", false, nil, nil)
	rt.EndFrame()
}

func TestHookCountPanicFormat(t *testing.T) {
	rt := NewRuntime(nil)

	rt.BeginFrame()
	rt.RecordHookCount("root", 2)
	rt.EndFrame()

	rt.BeginFrame()
	rt.RecordHookCount("root", 1)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !containsStr(msg, "root") || !containsStr(msg, "1") || !containsStr(msg, "2") {
			t.Fatalf("panic message missing expected content: %s", msg)
		}
	}()
	rt.EndFrame()
}

func TestDispose(t *testing.T) {
	rt := NewRuntime(nil)
	var cleanupCalled bool

	rt.BeginFrame()
	rt.UseEffect("dispose-eff", false, nil, func() func() {
		return func() { cleanupCalled = true }
	})
	rt.EndFrame()

	rt.Dispose()

	if !cleanupCalled {
		t.Fatal("expected cleanup on dispose")
	}
}

func TestDisposeTwice(t *testing.T) {
	rt := NewRuntime(nil)
	rt.Dispose()
	rt.Dispose()
}

func TestRememberCachesValue(t *testing.T) {
	rt := NewRuntime(nil)
	callCount := 0
	factory := func() any {
		callCount++
		return 42
	}

	v1 := rt.remember("key1", factory)
	v2 := rt.remember("key1", factory)

	if callCount != 1 {
		t.Fatalf("expected factory to be called once, called %d", callCount)
	}
	if v1 != v2 {
		t.Fatal("expected same value from cache")
	}
}

func TestDepsEqual(t *testing.T) {
	if !depsEqual(nil, nil) {
		t.Fatal("nil == nil")
	}
	// nil and empty both have len 0, so they are equal by design
	if !depsEqual(nil, []any{}) {
		t.Fatal("nil and empty should be equal (both len 0)")
	}
	if !depsEqual([]any{1, "a"}, []any{1, "a"}) {
		t.Fatal("same values should be equal")
	}
	if depsEqual([]any{1}, []any{2}) {
		t.Fatal("different values should not be equal")
	}
	if depsEqual([]any{1}, []any{1, 2}) {
		t.Fatal("different lengths should not be equal")
	}
}

func TestCloneDeps(t *testing.T) {
	if cloneDeps(nil) != nil {
		t.Fatal("nil should return nil")
	}

	orig := []any{1, "a"}
	cloned := cloneDeps(orig)
	if len(cloned) != 2 || cloned[0] != 1 || cloned[1] != "a" {
		t.Fatal("cloned values don't match")
	}

	cloned[0] = 99
	if orig[0] == 99 {
		t.Fatal("clone should be independent")
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
