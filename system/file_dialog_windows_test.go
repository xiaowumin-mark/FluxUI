//go:build windows

package system

import (
	"context"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func TestWindowsFileDialogCancelWatcherClosesDialog(t *testing.T) {
	var closed atomic.Int32
	var badHR atomic.Uint32
	closeProc := syscall.NewCallback(func(_ uintptr, hr uintptr) uintptr {
		if hr != hresultCancelled {
			badHR.Store(uint32(hr))
		}
		closed.Add(1)
		return 0
	})
	dialog := &iFileDialog{lpVtbl: &iFileDialogVtbl{Close: closeProc}}

	ctx, cancel := context.WithCancel(context.Background())
	stop := watchWindowsFileDialogContext(ctx, dialog)
	cancel()

	deadline := time.Now().Add(time.Second)
	for closed.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	stop()
	if closed.Load() != 1 {
		t.Fatalf("expected dialog to be closed once, got %d", closed.Load())
	}
	if got := badHR.Load(); got != 0 {
		t.Fatalf("expected cancel HRESULT 0x%08x, got 0x%08x", uint32(hresultCancelled), got)
	}
}

func TestWindowsFileDialogCancelWatcherStopPreventsClose(t *testing.T) {
	var closed atomic.Int32
	closeProc := syscall.NewCallback(func(uintptr, uintptr) uintptr {
		closed.Add(1)
		return 0
	})
	dialog := &iFileDialog{lpVtbl: &iFileDialogVtbl{Close: closeProc}}

	ctx, cancel := context.WithCancel(context.Background())
	stop := watchWindowsFileDialogContext(ctx, dialog)
	stop()
	cancel()
	time.Sleep(10 * time.Millisecond)
	if closed.Load() != 0 {
		t.Fatalf("expected stopped watcher not to close dialog, got %d", closed.Load())
	}
	stop()
}

func TestWindowsFilterSpecsNormalizePatterns(t *testing.T) {
	specs, keepAlive, err := windowsFilterSpecs([]FileFilter{
		{Name: "Images", Patterns: []string{"png", ".jpg", "*.webp", ""}},
		{Patterns: []string{"*.*"}},
	})
	if err != nil {
		t.Fatalf("unexpected filter spec error: %v", err)
	}
	if len(specs) != 2 || len(keepAlive) != 4 {
		t.Fatalf("unexpected filter spec count: specs=%d keepAlive=%d", len(specs), len(keepAlive))
	}
	if got := stringFromUTF16Ptr(specs[0].Spec); got != "*.png;*.jpg;*.webp" {
		t.Fatalf("unexpected normalized image patterns: %q", got)
	}
	if got := stringFromUTF16Ptr(specs[1].Name); got != "*.*" {
		t.Fatalf("expected unnamed filter to use patterns as name, got %q", got)
	}
}

func TestWindowsFilterSpecsRejectEmptyPatterns(t *testing.T) {
	if _, _, err := windowsFilterSpecs([]FileFilter{{Name: "Empty"}}); err == nil {
		t.Fatal("expected empty filter patterns to fail")
	}
}

func TestWindowsHRESULTHelpers(t *testing.T) {
	if hresultError(0) != nil {
		t.Fatal("expected successful HRESULT to be nil")
	}
	err := hresultError(hresultCancelled)
	if err == nil {
		t.Fatal("expected failed HRESULT to return error")
	}
	if !isHRESULT(err, hresultCancelled) {
		t.Fatalf("expected HRESULT helper to recognize cancel error, got %v", err)
	}
	if isHRESULT(err, 0x80070005) {
		t.Fatal("expected HRESULT helper not to match a different HRESULT")
	}
}

func stringFromUTF16Ptr(ptr *uint16) string {
	if ptr == nil {
		return ""
	}
	return windows.UTF16PtrToString(ptr)
}
