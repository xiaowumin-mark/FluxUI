//go:build windows

package system

import "testing"

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
