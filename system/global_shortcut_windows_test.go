//go:build windows

package system

import (
	"errors"
	"sync"
	"testing"
)

func TestWindowsGlobalShortcutModifiers(t *testing.T) {
	got := windowsGlobalShortcutModifiers(GlobalShortcutAlt | GlobalShortcutControl | GlobalShortcutShift | GlobalShortcutMeta)
	want := uint32(windowsModAlt | windowsModControl | windowsModShift | windowsModWin)
	if got != want {
		t.Fatalf("unexpected modifiers: got %#x want %#x", got, want)
	}
}

func TestWindowsGlobalShortcutKey(t *testing.T) {
	tests := map[string]uint32{
		"A":      0x41,
		"9":      0x39,
		"F1":     0x70,
		"F12":    0x7B,
		"ESC":    0x1B,
		"SPACE":  0x20,
		"DELETE": 0x2E,
	}
	for key, want := range tests {
		got, err := windowsGlobalShortcutKey(key)
		if err != nil {
			t.Fatalf("unexpected key %s error: %v", key, err)
		}
		if got != want {
			t.Fatalf("unexpected key %s: got %#x want %#x", key, got, want)
		}
	}
	if _, err := windowsGlobalShortcutKey("mouse"); !IsUnsupported(err) {
		t.Fatalf("expected unsupported key error, got %v", err)
	}
}

func TestWindowsGlobalShortcutClassName(t *testing.T) {
	if windowsGlobalShortcutClassName() == "" {
		t.Fatal("expected non-empty class name")
	}
}

func TestWindowsGlobalShortcutDispatchConcurrentWithClose(t *testing.T) {
	const dispatchers = 8
	handle := &windowsGlobalShortcutHandle{
		spec: GlobalShortcutSpec{
			ID:        "concurrent",
			Key:       "K",
			Modifiers: GlobalShortcutControl,
		},
		ch: make(chan GlobalShortcutEvent, 1),
	}

	start := make(chan struct{})
	var wg sync.WaitGroup
	for range dispatchers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range 1_000 {
				handle.dispatch()
			}
		}()
	}

	close(start)
	handle.mu.Lock()
	handle.closed = true
	close(handle.ch)
	handle.mu.Unlock()
	wg.Wait()
}

func TestWindowsGlobalShortcutCloseClosesEventsWhenUnregisterFails(t *testing.T) {
	wantErr := errors.New("unregister failed")
	previous := unregisterWindowsGlobalShortcut
	unregisterWindowsGlobalShortcut = func(uintptr, int32) error {
		return wantErr
	}
	t.Cleanup(func() {
		unregisterWindowsGlobalShortcut = previous
	})

	handle := &windowsGlobalShortcutHandle{
		id: 17,
		ch: make(chan GlobalShortcutEvent),
	}
	if err := handle.close(); !errors.Is(err, wantErr) {
		t.Fatalf("expected unregister error, got %v", err)
	}
	if _, ok := <-handle.events(); ok {
		t.Fatal("expected event channel to close even when unregister fails")
	}
}
