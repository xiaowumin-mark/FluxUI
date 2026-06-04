//go:build !windows

package system

import (
	"context"
	"testing"
)

func TestPlatformFileDialogUnsupported(t *testing.T) {
	if Supports(CapabilityFileDialog) {
		t.Fatal("unsupported platform should not expose file dialog capability")
	}

	_, err := OpenFileDialog(context.Background())
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestPlatformMessageBoxUnsupported(t *testing.T) {
	if Supports(CapabilityMessageBox) {
		t.Fatal("unsupported platform should not expose message box capability")
	}

	_, err := ShowMessageBox(context.Background())
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestPlatformNotificationUnsupported(t *testing.T) {
	if Supports(CapabilityNotification) {
		t.Fatal("unsupported platform should not expose notification capability")
	}

	err := Notify(context.Background())
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}
