//go:build linux

package system

import (
	"strings"
	"testing"
)

func TestLinuxTrayMenuSpecFlattensMenu(t *testing.T) {
	handle := &linuxTrayHandle{
		tempDir:  t.TempDir(),
		fifoPath: "/tmp/fluxui-tray-test.fifo",
	}
	menu := TrayMenu{
		TrayMenuItem{ID: "open", Label: "Open"},
		TrayMenuSeparator(),
		TrayMenuItem{
			ID:    "tools",
			Label: "Tools",
			Children: TrayMenu{
				TrayMenuItem{ID: "settings", Label: "Settings"},
			},
		},
		TrayMenuItem{ID: "disabled", Label: "Disabled", Disabled: true},
	}

	spec, err := handle.menuSpecLocked(menu)
	if err != nil {
		t.Fatalf("unexpected menu spec error: %v", err)
	}
	if !strings.Contains(spec, "Open!") {
		t.Fatalf("expected open menu entry, got %q", spec)
	}
	if !strings.Contains(spec, "Tools / Settings!") {
		t.Fatalf("expected flattened submenu entry, got %q", spec)
	}
	if !strings.Contains(spec, "Disabled!true") {
		t.Fatalf("expected disabled entry to use no-op command, got %q", spec)
	}
	if strings.Contains(spec, "|!") {
		t.Fatalf("unexpected separator command in menu spec: %q", spec)
	}
}

func TestLinuxTrayEscaping(t *testing.T) {
	if got := linuxTrayMenuEscape("A!B|C"); got != "A-B/C" {
		t.Fatalf("unexpected menu escape: %q", got)
	}
	if got := linuxTrayShellQuote("can't"); got != "'can'\"'\"'t'" {
		t.Fatalf("unexpected shell quote: %q", got)
	}
	if got := linuxTraySafeFileName("menu:open/settings"); got != "menu_open_settings" {
		t.Fatalf("unexpected safe filename: %q", got)
	}
}

func TestLinuxTrayResourceIconUnsupported(t *testing.T) {
	handle := &linuxTrayHandle{}
	if err := handle.setIconResource(1); !IsUnsupported(err) {
		t.Fatalf("expected resource icon unsupported, got %v", err)
	}
}
