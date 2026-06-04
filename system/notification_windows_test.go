//go:build windows

package system

import (
	"testing"
	"time"
)

func TestWindowsNotificationData(t *testing.T) {
	data, err := newWindowsNotificationData(123, 7, 456, notificationOptions{
		title:   "Done",
		body:    "Export finished.",
		kind:    NotificationSuccess,
		group:   "exports",
		timeout: 5 * time.Second,
	})
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
	if data.uFlags != nifMessage|nifIcon|nifTip|nifInfo {
		t.Fatalf("unexpected flags 0x%x", data.uFlags)
	}
	if data.uCallbackMessage != wmFluxUINotification {
		t.Fatalf("unexpected callback message 0x%x", data.uCallbackMessage)
	}
	if data.uTimeout != 5000 {
		t.Fatalf("expected timeout 5000ms, got %d", data.uTimeout)
	}
	if data.dwInfoFlags != niifInfo {
		t.Fatalf("expected success to map to info icon, got 0x%x", data.dwInfoFlags)
	}
	if fixedUTF16String(data.szInfoTitle[:]) != "Done" {
		t.Fatalf("unexpected title %q", fixedUTF16String(data.szInfoTitle[:]))
	}
	if fixedUTF16String(data.szInfo[:]) != "Export finished." {
		t.Fatalf("unexpected body %q", fixedUTF16String(data.szInfo[:]))
	}
}

func TestWindowsNotificationInfoFlags(t *testing.T) {
	tests := []struct {
		kind NotificationKind
		want uint32
	}{
		{kind: NotificationInfo, want: niifInfo},
		{kind: NotificationSuccess, want: niifInfo},
		{kind: NotificationWarning, want: niifWarning},
		{kind: NotificationError, want: niifError},
		{kind: NotificationKind("custom"), want: niifInfo},
	}

	for _, tt := range tests {
		if got := windowsNotificationInfoFlags(tt.kind); got != tt.want {
			t.Fatalf("expected kind %q to map to 0x%x, got 0x%x", tt.kind, tt.want, got)
		}
	}
}

func TestWindowsNotificationTimeoutMilliseconds(t *testing.T) {
	tests := []struct {
		timeout time.Duration
		want    uint32
	}{
		{timeout: 0, want: 10000},
		{timeout: 100 * time.Millisecond, want: 1000},
		{timeout: 5 * time.Second, want: 5000},
		{timeout: 2 * time.Minute, want: 60000},
	}

	for _, tt := range tests {
		if got := windowsNotificationTimeoutMilliseconds(tt.timeout); got != tt.want {
			t.Fatalf("expected timeout %s to map to %d, got %d", tt.timeout, tt.want, got)
		}
	}
}

func TestWindowsNotificationDefaultTitle(t *testing.T) {
	data, err := newWindowsNotificationData(1, 2, 3, notificationOptions{
		body: "Body only",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fixedUTF16String(data.szInfoTitle[:]) != "FluxUI" {
		t.Fatalf("expected default title FluxUI, got %q", fixedUTF16String(data.szInfoTitle[:]))
	}
}

func fixedUTF16String(value []uint16) string {
	for i, c := range value {
		if c == 0 {
			return stringFromUTF16(value[:i])
		}
	}
	return stringFromUTF16(value)
}

func stringFromUTF16(value []uint16) string {
	runes := make([]rune, 0, len(value))
	for _, c := range value {
		runes = append(runes, rune(c))
	}
	return string(runes)
}
