//go:build darwin

package system

import (
	"strings"
	"testing"
)

func TestDarwinTrayScriptBuildsStatusItemMenu(t *testing.T) {
	script := darwinTrayScript(trayOptions{
		tooltip: "FluxUI tray",
		icon:    "/tmp/fluxui.png",
		menu: TrayMenu{
			TrayMenuItem{ID: "open", Label: "Open"},
			TrayMenuSeparator(),
			TrayMenuItem{ID: "sync", Label: "Sync", Checked: true},
			TrayMenuItem{ID: "disabled", Label: "Disabled", Disabled: true},
			TrayMenuItem{
				ID:    "tools",
				Label: "Tools",
				Children: TrayMenu{
					TrayMenuItem{ID: "settings", Label: "Settings"},
				},
			},
		},
	}, "/tmp/fluxui-events.log", "/tmp/fluxui.png")

	for _, want := range []string{
		`NSStatusBar`,
		`NSApplication's sharedApplication()`,
		`NSApplicationActivationPolicyAccessory`,
		`property eventPath : "/tmp/fluxui-events.log"`,
		`buttonRef's setToolTip:"FluxUI tray"`,
		`initWithContentsOfFile:"/tmp/fluxui.png"`,
		`menuRef's setAutoenablesItems:false`,
		`initWithTitle:"Open" action:"menuClicked:"`,
		`setRepresentedObject:"open"`,
		`separatorItem()`,
		`setState:1`,
		`setEnabled:false`,
		`setSubmenu:`,
		`initWithTitle:"Settings" action:"menuClicked:"`,
		`my writeEvent("menu:" & itemID)`,
		`quoted form of (theEvent & linefeed)`,
		`appRef's run()`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("expected script to contain %q, got:\n%s", want, script)
		}
	}
}

func TestDarwinTrayScriptEnablesClickWithoutMenu(t *testing.T) {
	script := darwinTrayScript(trayOptions{
		onClick: func(TrayEvent) {},
	}, "/tmp/fluxui-events.log", "")

	if !strings.Contains(script, `buttonRef's setAction:"statusClicked:"`) {
		t.Fatalf("expected status click action, got:\n%s", script)
	}
	if strings.Contains(script, `set menuRef`) {
		t.Fatalf("unexpected menu for click-only tray script:\n%s", script)
	}
}

func TestDarwinTrayTextSanitizesNewlines(t *testing.T) {
	if got := darwinTrayText("hello\nworld"); got != "hello world" {
		t.Fatalf("unexpected text sanitization: %q", got)
	}
	if got := darwinTrayText(""); got != "FluxUI" {
		t.Fatalf("unexpected empty text fallback: %q", got)
	}
}

func TestDarwinTrayResourceIconUnsupported(t *testing.T) {
	handle := &darwinTrayHandle{}
	if err := handle.setIconResource(1); !IsUnsupported(err) {
		t.Fatalf("expected resource icon unsupported, got %v", err)
	}
}
