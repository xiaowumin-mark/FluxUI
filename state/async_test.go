package state_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/state"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func newAsyncTestCtx() (*internal.Runtime, *internal.Context) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	return rt, ctx
}

func TestAsyncInitialState(t *testing.T) {
	_, ctx := newAsyncTestCtx()
	h := state.UseAsync[int](ctx)

	if h.Status() != state.AsyncIdle {
		t.Fatalf("expected AsyncIdle, got %d", h.Status())
	}
	if h.Loading() {
		t.Fatal("expected Loading() == false")
	}
	if h.Data() != 0 {
		t.Fatalf("expected zero data, got %d", h.Data())
	}
	if h.Error() != nil {
		t.Fatalf("expected nil error, got %v", h.Error())
	}
}

func TestAsyncRunSuccess(t *testing.T) {
	_, ctx := newAsyncTestCtx()
	h := state.UseAsync[string](ctx)

	done := make(chan struct{})
	h.Run(func() (string, error) {
		defer close(done)
		return "hello", nil
	})

	if !h.Loading() {
		t.Fatal("expected Loading() == true immediately after Run")
	}

	<-done
	time.Sleep(5 * time.Millisecond)

	if h.Status() != state.AsyncSuccess {
		t.Fatalf("expected AsyncSuccess, got %d", h.Status())
	}
	if h.Data() != "hello" {
		t.Fatalf("expected \"hello\", got %q", h.Data())
	}
	if h.Error() != nil {
		t.Fatalf("expected nil error, got %v", h.Error())
	}
}

func TestAsyncRunError(t *testing.T) {
	_, ctx := newAsyncTestCtx()
	h := state.UseAsync[int](ctx)

	done := make(chan struct{})
	h.Run(func() (int, error) {
		defer close(done)
		return 0, errors.New("network failure")
	})

	<-done
	time.Sleep(5 * time.Millisecond)

	if h.Status() != state.AsyncError {
		t.Fatalf("expected AsyncError, got %d", h.Status())
	}
	if h.Error() == nil || h.Error().Error() != "network failure" {
		t.Fatalf("expected 'network failure' error, got %v", h.Error())
	}
}

func TestAsyncRunCancelsStale(t *testing.T) {
	_, ctx := newAsyncTestCtx()
	h := state.UseAsync[string](ctx)

	firstStarted := make(chan struct{})
	firstContinue := make(chan struct{})
	secondDone := make(chan struct{})

	// First Run — will be superseded
	h.Run(func() (string, error) {
		close(firstStarted)
		<-firstContinue
		return "stale", nil
	})

	<-firstStarted

	// Second Run — supersedes the first
	h.Run(func() (string, error) {
		defer close(secondDone)
		return "fresh", nil
	})

	// Let the first goroutine finish
	close(firstContinue)
	<-secondDone
	time.Sleep(10 * time.Millisecond)

	if h.Status() != state.AsyncSuccess {
		t.Fatalf("expected AsyncSuccess, got %d", h.Status())
	}
	if h.Data() != "fresh" {
		t.Fatalf("expected \"fresh\", got %q — stale result leaked", h.Data())
	}
}

func TestAsyncReset(t *testing.T) {
	_, ctx := newAsyncTestCtx()
	h := state.UseAsync[int](ctx)

	done := make(chan struct{})
	h.Run(func() (int, error) {
		defer close(done)
		return 42, nil
	})
	<-done
	time.Sleep(5 * time.Millisecond)

	if h.Data() != 42 {
		t.Fatalf("expected 42 before reset, got %d", h.Data())
	}

	h.Reset()

	if h.Status() != state.AsyncIdle {
		t.Fatalf("expected AsyncIdle after reset, got %d", h.Status())
	}
	if h.Data() != 0 {
		t.Fatalf("expected zero data after reset, got %d", h.Data())
	}
}

func TestAsyncNilSafety(t *testing.T) {
	var h *state.AsyncHandle[int]

	if h.Status() != state.AsyncIdle {
		t.Fatal("nil handle Status() should return AsyncIdle")
	}
	if h.Loading() {
		t.Fatal("nil handle Loading() should return false")
	}
	if h.Data() != 0 {
		t.Fatal("nil handle Data() should return zero")
	}
	if h.Error() != nil {
		t.Fatal("nil handle Error() should return nil")
	}
	h.Run(func() (int, error) { return 1, nil })
	h.Reset()
}

func TestAsyncConcurrent(t *testing.T) {
	_, ctx := newAsyncTestCtx()
	h := state.UseAsync[int](ctx)

	const goroutines = 50
	const iterations = 100
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	for range goroutines {
		go func() {
			defer wg.Done()
			for range iterations {
				h.Run(func() (int, error) {
					return 1, nil
				})
			}
		}()
	}
	for range goroutines {
		go func() {
			defer wg.Done()
			for range iterations {
				_ = h.Status()
				_ = h.Loading()
				_ = h.Data()
				_ = h.Error()
			}
		}()
	}
	wg.Wait()
}
