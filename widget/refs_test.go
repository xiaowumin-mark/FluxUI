package widget

import (
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/style"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestButtonRefQueuesCommands(t *testing.T) {
	ref := NewButtonRef()
	ref.Click()
	ref.Click()

	cmds := ref.drainCommands()
	if len(cmds) != 2 {
		t.Fatalf("expected 2 button commands, got %d", len(cmds))
	}
	if next := ref.drainCommands(); len(next) != 0 {
		t.Fatalf("expected queue to drain cleanly, got %d commands", len(next))
	}
}

func TestInputRefQueuesCommands(t *testing.T) {
	ref := NewInputRef()
	ref.SetText("hello")
	ref.Append(" world")
	ref.Clear()
	ref.Focus()
	ref.Blur()
	ref.Append("")

	cmds := ref.drainCommands()
	if len(cmds) != 5 {
		t.Fatalf("expected 5 input commands, got %d", len(cmds))
	}
	if cmds[0].kind != inputCmdSetText || cmds[0].text != "hello" {
		t.Fatalf("unexpected first command: %#v", cmds[0])
	}
	if cmds[1].kind != inputCmdAppend || cmds[1].text != " world" {
		t.Fatalf("unexpected second command: %#v", cmds[1])
	}
	if cmds[2].kind != inputCmdClear || cmds[3].kind != inputCmdFocus || cmds[4].kind != inputCmdBlur {
		t.Fatalf("unexpected command sequence: %#v", cmds)
	}
}

func TestCheckboxAndSwitchRefsQueueCommands(t *testing.T) {
	checkbox := NewCheckboxRef()
	checkbox.SetChecked(true)
	checkbox.Toggle()
	if cmds := checkbox.drainCommands(); len(cmds) != 2 {
		t.Fatalf("expected 2 checkbox commands, got %d", len(cmds))
	}

	switchRef := NewSwitchRef()
	switchRef.SetChecked(true)
	switchRef.Toggle()
	if cmds := switchRef.drainCommands(); len(cmds) != 2 {
		t.Fatalf("expected 2 switch commands, got %d", len(cmds))
	}
}

func TestScrollRefQueuesCommands(t *testing.T) {
	ref := NewScrollRef()
	ref.ScrollToStart()
	ref.ScrollToEnd()
	ref.ScrollToOffset(24)
	ref.ScrollBy(3)

	cmds := ref.drainCommands()
	if len(cmds) != 4 {
		t.Fatalf("expected 4 scroll commands, got %d", len(cmds))
	}
	if cmds[0].kind != scrollCmdToStart || cmds[1].kind != scrollCmdToEnd || cmds[2].kind != scrollCmdToOffset || cmds[3].kind != scrollCmdBy {
		t.Fatalf("unexpected scroll command order: %#v", cmds)
	}
}

func TestDialogTabsAndNavigationRefsQueueCommands(t *testing.T) {
	dialog := NewDialogRef()
	dialog.Open()
	dialog.Toggle()
	if cmds := dialog.drainCommands(); len(cmds) != 2 {
		t.Fatalf("expected 2 dialog commands, got %d", len(cmds))
	}

	popup := NewPopupRef()
	popup.Open()
	popup.Toggle()
	if cmds := popup.drainCommands(); len(cmds) != 2 {
		t.Fatalf("expected 2 popup commands, got %d", len(cmds))
	}

	tabs := NewTabsRef()
	tabs.SetActive("settings")
	if cmds := tabs.drainCommands(); len(cmds) != 1 || cmds[0] != "settings" {
		t.Fatalf("unexpected tabs commands: %#v", cmds)
	}

	bottomNav := NewBottomNavRef()
	bottomNav.SetActive("home")
	if cmds := bottomNav.drainCommands(); len(cmds) != 1 || cmds[0] != "home" {
		t.Fatalf("unexpected bottom nav commands: %#v", cmds)
	}
}

func TestRefsCanBindWithoutRuntime(t *testing.T) {
	var ops op.Ops
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops}, nil)

	cases := []struct {
		name string
		run  func()
	}{
		{"button", func() { Button(Text("go"), ButtonAttachRef(NewButtonRef())).Layout(ctx) }},
		{"checkbox", func() { Checkbox("ok", false, CheckboxAttachRef(NewCheckboxRef())).Layout(ctx) }},
		{"pressable", func() { Pressable(Text("go"), nil, PressableAttachRef(NewPressableRef())).Layout(ctx) }},
		{"input", func() { TextField("", InputAttachRef(NewInputRef())).Layout(ctx) }},
		{"scroll", func() { ScrollView(Text("body"), ScrollAttachRef(NewScrollRef())).Layout(ctx) }},
		{"bottom-nav", func() {
			BottomNavigation("home", []NavItem{{Key: "home", Label: "Home"}}, BottomNavAttachRef(NewBottomNavRef())).Layout(ctx)
		}},
		{"radio", func() {
			RadioGroup("a", []RadioItem{{Label: "A", Value: "a"}}, RadioGroupAttachRef(NewRadioGroupRef())).Layout(ctx)
		}},
		{"select", func() {
			Select("a", []SelectOptionItem[string]{{Label: "A", Value: "a"}}, SelectAttachRef(NewSelectRef[string]())).Layout(ctx)
		}},
		{"slider", func() { Slider(0.5, SliderAttachRef(NewSliderRef())).Layout(ctx) }},
		{"switch", func() { Switch(false, SwitchAttachRef(NewSwitchRef())).Layout(ctx) }},
		{"tabs", func() { Tabs("a", []TabItem{{Key: "a", Label: "A"}}, TabsAttachRef(NewTabsRef())).Layout(ctx) }},
		{"dialog", func() { Dialog(false, Text("body"), DialogAttachRef(NewDialogRef())).Layout(ctx) }},
		{"popup", func() { Popup(false, Text("body"), PopupAttachRef(NewPopupRef())).Layout(ctx) }},
		{"image", func() { Image(ImageSource{}, ImageAttachRef(NewButtonRef())).Layout(ctx) }},
		{"icon", func() { Icon("i", IconAttachRef(NewButtonRef())).Layout(ctx) }},
		{"card", func() {
			Card(Text("body"), CardAttachRef(NewButtonRef()), CardDecoration(style.Decoration{})).Layout(ctx)
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("expected ref binding without runtime not to panic, got %v", r)
				}
			}()
			tc.run()
		})
	}
}
