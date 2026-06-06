//go:build windows

package system

import "testing"

func TestWindowsTrayData(t *testing.T) {
	data, err := newWindowsTrayData(123, 7, 456, "FluxUI Tray")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.cbSize == 0 {
		t.Fatal("expected notify icon data size to be set")
	}
	if data.hWnd != 123 {
		t.Fatalf("expected hwnd 123, got %d", data.hWnd)
	}
	if data.uID != 7 {
		t.Fatalf("expected icon id 7, got %d", data.uID)
	}
	if data.hIcon != 456 {
		t.Fatalf("expected icon 456, got %d", data.hIcon)
	}
	if data.uFlags != nifMessage|nifIcon|nifTip {
		t.Fatalf("unexpected tray flags 0x%x", data.uFlags)
	}
	if data.uCallbackMessage != wmFluxUITray {
		t.Fatalf("unexpected callback message 0x%x", data.uCallbackMessage)
	}
	if fixedUTF16String(data.szTip[:]) != "FluxUI Tray" {
		t.Fatalf("unexpected tooltip %q", fixedUTF16String(data.szTip[:]))
	}
}

func TestWindowsTrayDataDefaultTooltip(t *testing.T) {
	data, err := newWindowsTrayData(1, 2, 3, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fixedUTF16String(data.szTip[:]) != "FluxUI" {
		t.Fatalf("expected default tooltip FluxUI, got %q", fixedUTF16String(data.szTip[:]))
	}
}

func TestWindowsTrayClassNameIsProcessScoped(t *testing.T) {
	name := windowsTrayClassName()
	if name == "" {
		t.Fatal("expected non-empty class name")
	}
	if name == "FluxUITrayWindow" {
		t.Fatal("expected class name to be process scoped")
	}
}

func TestWindowsTaskbarCreatedMessageRegistered(t *testing.T) {
	if windowsTaskbarCreatedMessage == 0 {
		t.Fatal("expected TaskbarCreated message to be registered")
	}
}

func TestWindowsTrayMenuFallbackLabels(t *testing.T) {
	if err := appendWindowsTrayMenuItemForTest(TrayMenuItem{ID: "open"}); err != nil {
		t.Fatalf("expected id-only item to be appendable: %v", err)
	}
	if err := appendWindowsTrayMenuItemForTest(TrayMenuItem{}); err != nil {
		t.Fatalf("expected empty item to use command fallback label: %v", err)
	}
}

func appendWindowsTrayMenuItemForTest(item TrayMenuItem) error {
	hMenu, _, _ := procCreatePopupMenu.Call()
	if hMenu == 0 {
		return ErrUnavailable
	}
	defer procDestroyMenu.Call(hMenu)
	return appendWindowsTrayMenuItem(hMenu, 1, item)
}
