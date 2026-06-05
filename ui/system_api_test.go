package ui

import (
	"context"
	"errors"
	"testing"
)

func TestCurrentWindowNativeHandleNilContext(t *testing.T) {
	if handle, ok := CurrentWindowNativeHandle(nil); ok || handle != 0 {
		t.Fatalf("expected missing native handle, got %d ok=%v", handle, ok)
	}
}

func TestOpenFileDialogContextChecksCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := OpenFileDialogContext(nil, ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestShowMessageBoxContextChecksCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ShowMessageBoxContext(nil, ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
