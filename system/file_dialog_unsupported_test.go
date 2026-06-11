//go:build !windows

package system

import (
	"context"
	"runtime"
	"testing"
)

func TestPlatformFileDialogUnsupported(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		t.Skip("file dialog has a minimal platform driver on macOS/Linux")
	}
	if Supports(CapabilityFileDialog) {
		t.Fatal("unsupported platform should not expose file dialog capability")
	}

	_, err := OpenFileDialog(context.Background())
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestPlatformMessageBoxUnsupported(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		t.Skip("message box has a minimal platform driver on macOS/Linux")
	}
	if Supports(CapabilityMessageBox) {
		t.Fatal("unsupported platform should not expose message box capability")
	}

	_, err := ShowMessageBox(context.Background())
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestPlatformNotificationUnsupported(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		t.Skip("notification has a minimal platform driver on macOS/Linux")
	}
	if Supports(CapabilityNotification) {
		t.Fatal("unsupported platform should not expose notification capability")
	}

	err := Notify(context.Background())
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestPlatformTrayUnsupported(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		t.Skip("tray has a minimal platform driver on macOS/Linux")
	}
	if Supports(CapabilityTray) {
		t.Fatal("unsupported platform should not expose tray capability")
	}

	_, err := NewTray()
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestPlatformSystemEventsUnsupported(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		t.Skip("system events has a minimal platform driver on macOS/Linux")
	}
	if Supports(CapabilitySystemEvents) {
		t.Fatal("unsupported platform should not expose system events capability")
	}

	_, err := SubscribeSystemEvents(context.Background())
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported system events error, got %v", err)
	}
}

func TestPlatformClipboardUnsupported(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		t.Skip("clipboard has a minimal platform driver on macOS/Linux")
	}
	if Supports(CapabilityClipboard) {
		t.Fatal("unsupported platform should not expose clipboard capability")
	}

	if _, err := ReadClipboardText(context.Background()); !IsUnsupported(err) {
		t.Fatalf("expected unsupported read error, got %v", err)
	}
	if err := WriteClipboardText(context.Background(), "text"); !IsUnsupported(err) {
		t.Fatalf("expected unsupported write error, got %v", err)
	}
	if _, err := ReadClipboardImagePNG(context.Background()); !IsUnsupported(err) {
		t.Fatalf("expected unsupported image read error, got %v", err)
	}
}

func TestPlatformShellUnsupported(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		t.Skip("shell has a minimal platform driver on macOS/Linux")
	}
	if Supports(CapabilityShell) {
		t.Fatal("unsupported platform should not expose shell capability")
	}

	if err := OpenURL(context.Background(), "https://example.com"); !IsUnsupported(err) {
		t.Fatalf("expected unsupported OpenURL error, got %v", err)
	}
	if err := OpenPath(context.Background(), "/tmp"); !IsUnsupported(err) {
		t.Fatalf("expected unsupported OpenPath error, got %v", err)
	}
	if err := RevealPath(context.Background(), "/tmp"); !IsUnsupported(err) {
		t.Fatalf("expected unsupported RevealPath error, got %v", err)
	}
}

func TestPlatformSingleInstanceUnsupported(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		t.Skip("single instance has a platform driver on macOS/Linux")
	}
	if Supports(CapabilitySingleInstance) {
		t.Fatal("unsupported platform should not expose single instance capability")
	}

	if _, err := AcquireSingleInstance(context.Background(), SingleInstanceID("com.example.test")); !IsUnsupported(err) {
		t.Fatalf("expected unsupported single instance error, got %v", err)
	}
}

func TestPlatformSystemRegistrationUnsupported(t *testing.T) {
	if Supports(CapabilitySystemRegistration) {
		t.Fatal("non-Windows platform should not expose system registration capability")
	}

	if err := RegisterProtocolHandler(context.Background(), "fluxui", "/tmp/app"); !IsUnsupported(err) {
		t.Fatalf("expected unsupported protocol registration error, got %v", err)
	}
	if err := RegisterStartupTask(context.Background(), "FluxUI", "/tmp/app"); !IsUnsupported(err) {
		t.Fatalf("expected unsupported startup registration error, got %v", err)
	}
}

func TestPlatformGlobalShortcutUnsupported(t *testing.T) {
	if Supports(CapabilityGlobalShortcut) {
		t.Fatal("non-Windows platform should not expose global shortcut capability")
	}

	_, err := RegisterGlobalShortcut(context.Background(), GlobalShortcutSpec{
		ID:        "open",
		Key:       "F9",
		Modifiers: GlobalShortcutControl,
	}, nil)
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported global shortcut error, got %v", err)
	}
}
