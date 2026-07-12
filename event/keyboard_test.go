package event

import (
	"reflect"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"

	"gioui.org/io/key"
	gioLayout "gioui.org/layout"
)

func TestKeyboardEventFromGioUsesStableNavigationNames(t *testing.T) {
	cases := []struct {
		name string
		raw  key.Name
		want string
	}{
		{name: "left", raw: key.NameLeftArrow, want: "ArrowLeft"},
		{name: "right", raw: key.NameRightArrow, want: "ArrowRight"},
		{name: "up", raw: key.NameUpArrow, want: "ArrowUp"},
		{name: "down", raw: key.NameDownArrow, want: "ArrowDown"},
		{name: "home", raw: key.NameHome, want: "Home"},
		{name: "end", raw: key.NameEnd, want: "End"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := KeyboardEventFromGio(nil, key.Event{Name: test.raw, State: key.Press})
			if got.Key != test.want {
				t.Fatalf("key = %q, want %q", got.Key, test.want)
			}
		})
	}
}

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

func TestMoveFocusWithinScopeDoesNotLeaveScope(t *testing.T) {
	rt := internal.NewRuntime(nil)
	rt.BeginFrame()
	root := internal.NewContext(gioLayout.Context{}, rt)
	outside := root.Scope("outside")
	scope := root.Scope("dialog")
	first := scope.Scope("first")
	second := scope.Scope("second")
	t.Cleanup(rt.EndFrame)

	RegisterFocusTarget(outside)
	RegisterFocusTarget(first)
	RegisterFocusTarget(second)

	if !RequestFocus(first) {
		t.Fatal("focus request failed")
	}
	if !rt.MoveFocusWithin(root, scope.PathID(), FocusForward) {
		t.Fatal("MoveFocusWithin forward returned false")
	}
	if FocusedTarget(root) != second.PathID() {
		t.Fatalf("focus after scoped forward = %v, want %v", FocusedTarget(root), second.PathID())
	}
	if !rt.MoveFocusWithin(root, scope.PathID(), FocusForward) {
		t.Fatal("MoveFocusWithin wrap returned false")
	}
	if FocusedTarget(root) != first.PathID() {
		t.Fatalf("focus after scoped wrap = %v, want %v", FocusedTarget(root), first.PathID())
	}
	if !rt.MoveFocusWithin(root, scope.PathID(), FocusBackward) {
		t.Fatal("MoveFocusWithin backward returned false")
	}
	if FocusedTarget(root) != second.PathID() {
		t.Fatalf("focus after scoped backward = %v, want %v", FocusedTarget(root), second.PathID())
	}

	if !RequestFocus(outside) {
		t.Fatal("outside focus request failed")
	}
	if !rt.MoveFocusWithin(root, scope.PathID(), FocusForward) {
		t.Fatal("MoveFocusWithin from outside returned false")
	}
	if FocusedTarget(root) != first.PathID() {
		t.Fatalf("focus entering scope = %v, want %v", FocusedTarget(root), first.PathID())
	}
	if !RequestFocus(outside) {
		t.Fatal("outside refocus request failed")
	}
	if !rt.MoveFocusWithin(root, scope.PathID(), FocusBackward) {
		t.Fatal("MoveFocusWithin backward from outside returned false")
	}
	if FocusedTarget(root) != second.PathID() {
		t.Fatalf("focus backward entering scope = %v, want %v", FocusedTarget(root), second.PathID())
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
