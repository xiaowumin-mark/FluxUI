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

func TestWindowsIconResourceDataFromICO(t *testing.T) {
	data := []byte{
		0, 0, // reserved
		1, 0, // icon
		1, 0, // count
		16, 16, 0, 0, // width/height/colors/reserved
		1, 0, 32, 0, // planes/bitcount
		4, 0, 0, 0, // bytes in resource
		22, 0, 0, 0, // image offset
		0xde, 0xad, 0xbe, 0xef,
	}
	resource, err := windowsIconResourceDataFromICO(data)
	if err != nil {
		t.Fatalf("unexpected icon parse error: %v", err)
	}
	if len(resource) != 4 || resource[0] != 0xde || resource[3] != 0xef {
		t.Fatalf("unexpected icon resource data: %#v", resource)
	}
}

func TestWindowsIconResourceDataFromICORejectsInvalidData(t *testing.T) {
	if _, err := windowsIconResourceDataFromICO([]byte{1, 2, 3}); err == nil {
		t.Fatal("expected invalid ico data to fail")
	}
}

func TestWindowsTrayRestoreAfterTaskbarCreatedOnlyRestoresVisibleIcons(t *testing.T) {
	oldShellNotifyIcon := shellNotifyIcon
	defer func() {
		shellNotifyIcon = oldShellNotifyIcon
	}()

	var calls []uint32
	shellNotifyIcon = func(message uint32, data *notifyIconData) error {
		calls = append(calls, message)
		if message != nimAdd {
			t.Fatalf("expected restore to use NIM_ADD, got %d", message)
		}
		if data == nil || data.uFlags != nifMessage|nifIcon|nifTip {
			t.Fatalf("unexpected restore data: %#v", data)
		}
		return nil
	}

	visible := &windowsTrayHandle{
		visible: true,
		data: notifyIconData{
			uFlags: nifTip,
		},
	}
	hidden := &windowsTrayHandle{visible: false}
	closed := &windowsTrayHandle{visible: true, closed: true}

	visible.restoreAfterTaskbarCreated()
	hidden.restoreAfterTaskbarCreated()
	closed.restoreAfterTaskbarCreated()

	if len(calls) != 1 {
		t.Fatalf("expected one restore call for visible tray, got %d", len(calls))
	}
}

func TestWindowsTrayMenuProviderEvaluatesWhenMenuOpens(t *testing.T) {
	oldTrackMenu := trackWindowsTrayMenuFunc
	defer func() {
		trackWindowsTrayMenuFunc = oldTrackMenu
	}()

	providerCalls := 0
	clicked := make([]string, 0, 2)
	handle := &windowsTrayHandle{
		data: notifyIconData{hWnd: 123},
		menuProvider: func() TrayMenu {
			providerCalls++
			id := "first"
			if providerCalls == 2 {
				id = "second"
			}
			return TrayMenu{
				TrayMenuItem{
					ID:    id,
					Label: id,
					OnClick: func(event TrayEvent) {
						clicked = append(clicked, event.ItemID)
					},
				},
			}
		},
	}

	trackWindowsTrayMenuFunc = func(hWnd uintptr, menu TrayMenu) (uint32, map[uint32]TrayMenuItem, error) {
		if hWnd != 123 {
			t.Fatalf("unexpected tray hwnd %d", hWnd)
		}
		if len(menu) != 1 {
			t.Fatalf("expected one provided menu item, got %d", len(menu))
		}
		return 1, map[uint32]TrayMenuItem{1: menu[0]}, nil
	}

	handle.showContextMenu()
	handle.showContextMenu()

	if providerCalls != 2 {
		t.Fatalf("expected provider to run for each menu open, got %d", providerCalls)
	}
	if len(clicked) != 2 || clicked[0] != "first" || clicked[1] != "second" {
		t.Fatalf("unexpected clicked items: %#v", clicked)
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
