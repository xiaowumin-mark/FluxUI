package event

import (
	"reflect"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"

	gioLayout "gioui.org/layout"
)

func TestFocusMoveDispatchesFocusEventOrder(t *testing.T) {
	_, _, parent, first := newEventTestTree(t)
	second := parent.Scope("second")

	RegisterFocusTarget(first)
	RegisterFocusTarget(second)

	if !RequestFocus(first) {
		t.Fatal("initial focus request failed")
	}

	var calls []string
	OnFocus(first, Blur, func(ctx *internal.Context, ev *FocusEvent) {
		calls = append(calls, "blur:first")
		if ev.RelatedTarget != second.PathID() {
			t.Fatalf("blur related target = %v, want %v", ev.RelatedTarget, second.PathID())
		}
	})
	OnFocus(first, FocusOut, func(ctx *internal.Context, ev *FocusEvent) {
		calls = append(calls, "focusout:first")
		if ev.RelatedTarget != second.PathID() {
			t.Fatalf("focusout related target = %v, want %v", ev.RelatedTarget, second.PathID())
		}
	})
	OnFocus(second, Focus, func(ctx *internal.Context, ev *FocusEvent) {
		calls = append(calls, "focus:second")
		if ev.RelatedTarget != first.PathID() {
			t.Fatalf("focus related target = %v, want %v", ev.RelatedTarget, first.PathID())
		}
	})
	OnFocus(second, FocusIn, func(ctx *internal.Context, ev *FocusEvent) {
		calls = append(calls, "focusin:second")
		if ev.RelatedTarget != first.PathID() {
			t.Fatalf("focusin related target = %v, want %v", ev.RelatedTarget, first.PathID())
		}
	})

	if !RequestFocus(second) {
		t.Fatal("second focus request failed")
	}

	want := []string{"blur:first", "focusout:first", "focus:second", "focusin:second"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("focus calls = %v, want %v", calls, want)
	}
}

func TestKeyDownCanCancelActivationAndFocusMove(t *testing.T) {
	_, root, _, child := newEventTestTree(t)
	activated := 0

	RegisterFocusTarget(child, FocusActivate(func(ctx *internal.Context) {
		activated++
	}))
	OnKeyDown(child, func(ctx *internal.Context, ev *KeyboardEvent) {
		if ev.Key == "Enter" {
			ev.PreventDefault()
		}
	})
	if !RequestFocus(child) {
		t.Fatal("focus request failed")
	}

	enter := KeyboardEvent{
		Event: Event{Type: KeyDown},
		Key:   "Enter",
		Code:  "Enter",
	}
	if DispatchKeyboardEvent(root, 0, &enter) {
		t.Fatal("keydown dispatch returned true after preventDefault")
	}
	if activated != 0 {
		t.Fatalf("activation count = %d, want 0", activated)
	}

	_, root, parent, first := newEventTestTree(t)
	second := parent.Scope("second")
	RegisterFocusTarget(first)
	RegisterFocusTarget(second)
	OnKeyDown(first, func(ctx *internal.Context, ev *KeyboardEvent) {
		if ev.Key == "Tab" {
			ev.PreventDefault()
		}
	})
	RequestFocus(first)
	tab := KeyboardEvent{Event: Event{Type: KeyDown}, Key: "Tab", Code: "Tab"}
	DispatchKeyboardEvent(root, 0, &tab)
	if FocusedTarget(root) != first.PathID() {
		t.Fatalf("focus moved after canceled Tab, got %v want %v", FocusedTarget(root), first.PathID())
	}

	_, root, parent, first = newEventTestTree(t)
	second = parent.Scope("second")
	RegisterFocusTarget(first)
	RegisterFocusTarget(second)
	RequestFocus(first)
	tab = KeyboardEvent{Event: Event{Type: KeyDown}, Key: "Tab", Code: "Tab"}
	DispatchKeyboardEvent(root, 0, &tab)
	if FocusedTarget(root) != second.PathID() {
		t.Fatalf("focus after Tab = %v, want %v", FocusedTarget(root), second.PathID())
	}
}

func TestLocalShortcutOnlyFiresInsideFocusedScope(t *testing.T) {
	rt := internal.NewRuntime(nil)
	rt.BeginFrame()
	root := internal.NewContext(gioLayout.Context{}, rt)
	scopeA := root.Scope("scope-a")
	targetA := scopeA.Scope("target")
	scopeB := root.Scope("scope-b")
	targetB := scopeB.Scope("target")
	t.Cleanup(rt.EndFrame)

	RegisterFocusTarget(targetA)
	RegisterFocusTarget(targetB)

	spec := ShortcutExactModifiers(ShortcutKey("S", Modifiers{Ctrl: true, Shortcut: true}))
	aCalls := 0
	bCalls := 0
	OnShortcut(scopeA, spec, func(ctx *internal.Context, ev *KeyboardEvent) {
		aCalls++
	})
	OnShortcut(scopeB, spec, func(ctx *internal.Context, ev *KeyboardEvent) {
		bCalls++
	})

	RequestFocus(targetA)
	DispatchKeyboardEvent(root, 0, &KeyboardEvent{
		Event:     Event{Type: KeyDown},
		Key:       "S",
		Code:      "KeyS",
		Modifiers: Modifiers{Ctrl: true, Shortcut: true},
	})
	DispatchKeyboardEvent(root, 0, &KeyboardEvent{Event: Event{Type: KeyUp}, Key: "S", Code: "KeyS"})
	if aCalls != 1 || bCalls != 0 {
		t.Fatalf("shortcut after focusing scope A = %d/%d, want 1/0", aCalls, bCalls)
	}

	RequestFocus(targetB)
	DispatchKeyboardEvent(root, 0, &KeyboardEvent{
		Event:     Event{Type: KeyDown},
		Key:       "S",
		Code:      "KeyS",
		Modifiers: Modifiers{Ctrl: true, Shortcut: true},
	})
	if aCalls != 1 || bCalls != 1 {
		t.Fatalf("shortcut after focusing scope B = %d/%d, want 1/1", aCalls, bCalls)
	}
}
