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

func TestShowMessageBoxAsyncContextChecksCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	response := <-ShowMessageBoxAsyncContext(nil, ctx)
	if !errors.Is(response.Err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", response.Err)
	}
}

func TestShowMessageBoxDetailedContextChecksCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ShowMessageBoxDetailedContext(nil, ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestShowMessageBoxDetailedAsyncContextChecksCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	response := <-ShowMessageBoxDetailedAsyncContext(nil, ctx)
	if !errors.Is(response.Err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", response.Err)
	}
}
