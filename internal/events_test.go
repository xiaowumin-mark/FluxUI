package internal

import (
	"testing"
	"time"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func newEventTestTree(t *testing.T) (*Runtime, *Context, *Context, *Context) {
	t.Helper()

	var ops op.Ops
	runtime := NewRuntime(nil)
	runtime.BeginFrame()
	root := NewContext(gioLayout.Context{
		Ops: &ops,
		Now: time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC),
	}, runtime)
	parent := root.Scope("parent")
	child := parent.Child(0)
	return runtime, root, parent, child
}

func assertPathIDs(t *testing.T, got, want []PathID) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("path length = %d, want %d: got %#v want %#v", len(got), len(want), got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("path[%d] = %d, want %d: got %#v want %#v", index, got[index], want[index], got, want)
		}
	}
}

func assertStrings(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("entry count = %d, want %d: got %#v want %#v", len(got), len(want), got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("entry[%d] = %q, want %q: got %#v want %#v", index, got[index], want[index], got, want)
		}
	}
}

func TestDispatchEventCapturesTargetBubblesAndPriority(t *testing.T) {
	runtime, root, parent, child := newEventTestTree(t)
	const eventType EventType = "test"

	var seen []string
	register := func(ctx *Context, name string, opts EventListenerOptions) {
		runtime.RegisterEventListener(ctx, eventType, func(_ *Context, event *Event) {
			if event.Target != child.PathID() {
				t.Errorf("target = %d, want %d", event.Target, child.PathID())
			}
			seen = append(seen, name)
		}, opts)
	}
	register(root, "root capture", EventListenerOptions{Capture: true})
	register(parent, "parent capture", EventListenerOptions{Capture: true})
	register(child, "child capture", EventListenerOptions{Capture: true})
	register(child, "child low", EventListenerOptions{})
	register(child, "child high", EventListenerOptions{Priority: 10})
	register(parent, "parent bubble", EventListenerOptions{})
	register(root, "root bubble", EventListenerOptions{})

	event := &Event{Type: eventType, Bubbles: true, Cancelable: true}
	if allowed := runtime.DispatchEvent(root, child.PathID(), event); !allowed {
		t.Fatal("expected event default action to be allowed")
	}
	assertStrings(t, seen, []string{
		"root capture",
		"parent capture",
		"child capture",
		"child high",
		"child low",
		"parent bubble",
		"root bubble",
	})
	assertPathIDs(t, event.ComposedPath(), []PathID{child.PathID(), parent.PathID(), root.PathID()})
	if event.Phase != EventPhaseNone || event.CurrentTarget != 0 {
		t.Fatalf("dispatch did not reset event state: phase=%v current=%d", event.Phase, event.CurrentTarget)
	}
	if !event.Time.Equal(root.Now()) {
		t.Fatalf("event time = %v, want frame time %v", event.Time, root.Now())
	}

	path := event.ComposedPath()
	path[0] = 0
	if event.ComposedPath()[0] != child.PathID() {
		t.Fatal("ComposedPath returned the event's internal path")
	}
}

func TestDispatchEventHonorsCancellationOnceAndPropagation(t *testing.T) {
	t.Run("passive and cancelable listeners", func(t *testing.T) {
		runtime, root, _, child := newEventTestTree(t)

		passiveCalled := false
		runtime.RegisterEventListener(child, "passive", func(_ *Context, event *Event) {
			passiveCalled = true
			if event.PreventDefault() {
				t.Fatal("passive listener prevented the default action")
			}
		}, EventListenerOptions{Passive: true})
		passive := &Event{Type: "passive", Cancelable: true}
		if allowed := runtime.DispatchEvent(root, child.PathID(), passive); !allowed {
			t.Fatal("passive listener should not cancel the event")
		}
		if !passiveCalled || passive.DefaultPrevented {
			t.Fatalf("unexpected passive result: called=%t prevented=%t", passiveCalled, passive.DefaultPrevented)
		}
		if passive.passivePreventDefaultTarget != child.PathID() || passive.passivePreventDefaultPhase != EventPhaseTarget {
			t.Fatalf("passive prevention provenance = target %d phase %v", passive.passivePreventDefaultTarget, passive.passivePreventDefaultPhase)
		}

		runtime.RegisterEventListener(child, "cancel", func(_ *Context, event *Event) {
			if !event.PreventDefault() {
				t.Error("cancelable listener did not prevent the default action")
			}
		}, EventListenerOptions{})
		cancel := &Event{Type: "cancel", Cancelable: true}
		if allowed := runtime.DispatchEvent(root, child.PathID(), cancel); allowed {
			t.Fatal("expected prevented event to reject its default action")
		}
		if !cancel.DefaultPrevented || cancel.preventDefaultTarget != child.PathID() || cancel.preventDefaultPhase != EventPhaseTarget {
			t.Fatalf("unexpected prevention provenance: %#v", cancel)
		}
	})

	t.Run("once listener is removed after dispatch", func(t *testing.T) {
		runtime, root, _, child := newEventTestTree(t)
		calls := 0
		runtime.RegisterEventListener(child, "once", func(_ *Context, _ *Event) {
			calls++
		}, EventListenerOptions{Once: true})

		runtime.DispatchEvent(root, child.PathID(), &Event{Type: "once"})
		runtime.DispatchEvent(root, child.PathID(), &Event{Type: "once"})
		if calls != 1 {
			t.Fatalf("once listener called %d times, want 1", calls)
		}
		if _, ok := runtime.events.listeners[child.PathID()]; ok {
			t.Fatal("removed once listener remained in the registry")
		}
	})

	t.Run("stop propagation keeps current target listeners", func(t *testing.T) {
		runtime, root, _, child := newEventTestTree(t)
		var seen []string
		runtime.RegisterEventListener(child, "stop", func(_ *Context, event *Event) {
			seen = append(seen, "first")
			event.StopPropagation()
		}, EventListenerOptions{})
		runtime.RegisterEventListener(child, "stop", func(_ *Context, _ *Event) {
			seen = append(seen, "second")
		}, EventListenerOptions{})
		runtime.RegisterEventListener(root, "stop", func(_ *Context, _ *Event) {
			seen = append(seen, "root")
		}, EventListenerOptions{})

		event := &Event{Type: "stop", Bubbles: true}
		runtime.DispatchEvent(root, child.PathID(), event)
		assertStrings(t, seen, []string{"first", "second"})
		if !event.PropagationStopped() || event.propagationStopTarget != child.PathID() {
			t.Fatalf("unexpected propagation state: %#v", event)
		}
	})

	t.Run("immediate stop skips later target listeners", func(t *testing.T) {
		runtime, root, _, child := newEventTestTree(t)
		var seen []string
		runtime.RegisterEventListener(child, "immediate", func(_ *Context, event *Event) {
			seen = append(seen, "first")
			event.StopImmediatePropagation()
		}, EventListenerOptions{})
		runtime.RegisterEventListener(child, "immediate", func(_ *Context, _ *Event) {
			seen = append(seen, "second")
		}, EventListenerOptions{})
		runtime.RegisterEventListener(root, "immediate", func(_ *Context, _ *Event) {
			seen = append(seen, "root")
		}, EventListenerOptions{})

		event := &Event{Type: "immediate", Bubbles: true}
		runtime.DispatchEvent(root, child.PathID(), event)
		assertStrings(t, seen, []string{"first"})
		if !event.ImmediatePropagationStopped() || event.immediateStopTarget != child.PathID() {
			t.Fatalf("unexpected immediate propagation state: %#v", event)
		}
	})

	var nilEvent *Event
	nilEvent.StopPropagation()
	nilEvent.StopImmediatePropagation()
	if nilEvent.PreventDefault() || nilEvent.PropagationStopped() || nilEvent.ImmediatePropagationStopped() || nilEvent.ComposedPath() != nil {
		t.Fatal("nil event helpers should be inert")
	}
}

func TestEventPathsRespectBoundariesAndPortals(t *testing.T) {
	t.Run("stop boundary excludes ancestors", func(t *testing.T) {
		runtime, root, parent, child := newEventTestTree(t)
		rootCalls := 0
		runtime.RegisterEventListener(root, "boundary", func(_ *Context, _ *Event) {
			rootCalls++
		}, EventListenerOptions{})
		runtime.RegisterEventBoundary(parent, EventBoundaryPolicy{
			Mode:     EventBoundaryStop,
			Redirect: root.PathID(),
		})

		event := &Event{Type: "boundary", Bubbles: true}
		runtime.DispatchEvent(root, child.PathID(), event)
		assertPathIDs(t, event.ComposedPath(), []PathID{child.PathID(), parent.PathID()})
		if rootCalls != 0 {
			t.Fatalf("root listener called %d times through stop boundary", rootCalls)
		}
		if got := runtime.events.targets[parent.PathID()].Boundary; got.Redirect != 0 {
			t.Fatalf("stop boundary retained redirect: %#v", got)
		}
	})

	t.Run("redirect boundary routes to its configured owner", func(t *testing.T) {
		runtime, root, _, _ := newEventTestTree(t)
		owner := root.Scope("owner")
		boundary := root.Scope("boundary")
		child := boundary.Child(0)
		runtime.RegisterEventBoundary(boundary, EventBoundaryPolicy{
			Mode:     EventBoundaryRedirect,
			Redirect: owner.PathID(),
		})

		event := &Event{Type: "redirect"}
		runtime.DispatchEvent(root, child.PathID(), event)
		assertPathIDs(t, event.ComposedPath(), []PathID{child.PathID(), boundary.PathID(), owner.PathID(), root.PathID()})
		if !runtime.EventPathContains(child.PathID(), owner.PathID()) {
			t.Fatal("redirect owner was not included in the logical path")
		}
	})

	t.Run("portal replaces its layout parent in the event path", func(t *testing.T) {
		runtime, root, _, _ := newEventTestTree(t)
		owner := root.Scope("owner")
		layoutParent := root.Scope("layout")
		portal := layoutParent.Scope("portal")
		child := portal.Child(0)
		runtime.RegisterEventPortal(portal, owner.PathID())

		event := &Event{Type: "portal"}
		runtime.DispatchEvent(root, child.PathID(), event)
		assertPathIDs(t, event.ComposedPath(), []PathID{child.PathID(), portal.PathID(), owner.PathID(), root.PathID()})
		if runtime.EventPathContains(child.PathID(), layoutParent.PathID()) {
			t.Fatal("portal retained its layout parent in the event path")
		}
	})
}

func TestFocusNavigationEventsAndFrameCleanup(t *testing.T) {
	runtime, root, parent, first := newEventTestTree(t)
	second := parent.Child(1)
	positive := parent.Child(2)
	disabled := parent.Child(3)
	hidden := parent.Child(4)
	programmaticOnly := parent.Child(5)

	var events []string
	runtime.RegisterEventListener(first, EventTypeFocus, func(_ *Context, event *Event) {
		events = append(events, "first focus")
		if _, ok := event.Detail.(*FocusEvent); !ok {
			t.Errorf("unexpected focus detail: %#v", event.Detail)
		}
	}, EventListenerOptions{})
	runtime.RegisterEventListener(first, EventTypeBlur, func(_ *Context, event *Event) {
		events = append(events, "first blur")
		if _, ok := event.Detail.(*FocusEvent); !ok {
			t.Errorf("unexpected blur detail: %#v", event.Detail)
		}
	}, EventListenerOptions{})
	runtime.RegisterEventListener(parent, EventTypeFocusOut, func(_ *Context, _ *Event) {
		events = append(events, "parent focusout")
	}, EventListenerOptions{})
	runtime.RegisterEventListener(parent, EventTypeFocusIn, func(_ *Context, _ *Event) {
		events = append(events, "parent focusin")
	}, EventListenerOptions{})

	runtime.RegisterFocusTarget(first, FocusTargetOptions{TabIndex: 0})
	runtime.RegisterFocusTarget(second, FocusTargetOptions{TabIndex: 0})
	runtime.RegisterFocusTarget(positive, FocusTargetOptions{TabIndex: 1})
	runtime.RegisterFocusTarget(disabled, FocusTargetOptions{TabIndex: 0, Disabled: true})
	runtime.RegisterFocusTarget(hidden, FocusTargetOptions{TabIndex: 0, Hidden: true})
	runtime.RegisterFocusTarget(programmaticOnly, FocusTargetOptions{TabIndex: -1})

	assertPathIDs(t, runtime.sortedFocusOrder(), []PathID{positive.PathID(), first.PathID(), second.PathID()})
	if runtime.RequestFocus(root, disabled.PathID()) || runtime.RequestFocus(root, hidden.PathID()) {
		t.Fatal("disabled or hidden target accepted focus")
	}
	if !runtime.RequestFocus(root, first.PathID()) || runtime.FocusedTarget() != first.PathID() {
		t.Fatalf("failed to focus first target: %d", runtime.FocusedTarget())
	}
	if !runtime.RequestFocus(root, second.PathID()) || runtime.FocusedTarget() != second.PathID() {
		t.Fatalf("failed to focus second target: %d", runtime.FocusedTarget())
	}
	assertStrings(t, events, []string{"first focus", "parent focusin", "first blur", "parent focusout", "parent focusin"})

	if !runtime.MoveFocus(root, FocusBackward) || runtime.FocusedTarget() != first.PathID() {
		t.Fatalf("backward focus target = %d, want %d", runtime.FocusedTarget(), first.PathID())
	}
	if !runtime.MoveFocusWithin(root, parent.PathID(), FocusBackward) || runtime.FocusedTarget() != positive.PathID() {
		t.Fatalf("scoped backward focus target = %d, want %d", runtime.FocusedTarget(), positive.PathID())
	}
	if !runtime.RequestFocus(root, programmaticOnly.PathID()) || runtime.FocusedTarget() != programmaticOnly.PathID() {
		t.Fatalf("programmatic-only target did not receive focus: %d", runtime.FocusedTarget())
	}
	if !runtime.MoveFocus(root, FocusForward) || runtime.FocusedTarget() != positive.PathID() {
		t.Fatalf("tab navigation from programmatic-only target = %d, want %d", runtime.FocusedTarget(), positive.PathID())
	}
	if runtime.BlurFocus(root, first.PathID()) {
		t.Fatal("wrong target cleared focus")
	}
	if !runtime.BlurFocus(root, 0) || runtime.FocusedTarget() != 0 {
		t.Fatalf("blur did not clear focus: %d", runtime.FocusedTarget())
	}

	if !runtime.RequestFocus(root, second.PathID()) {
		t.Fatal("failed to restore focus before frame cleanup")
	}
	runtime.EndFrame()
	runtime.BeginFrame()
	runtime.EndFrame()
	if runtime.FocusedTarget() != 0 {
		t.Fatalf("unregistered focus target survived frame cleanup: %d", runtime.FocusedTarget())
	}
}

func TestKeyboardDispatchTracksRepeatShortcutsAndDefaults(t *testing.T) {
	t.Run("default actions and repeat tracking", func(t *testing.T) {
		runtime, root, parent, first := newEventTestTree(t)
		second := parent.Child(1)
		activations := 0
		runtime.RegisterFocusTarget(first, FocusTargetOptions{
			TabIndex: 0,
			Activate: func(ctx *Context) {
				if ctx != first {
					t.Errorf("activation context = %p, want %p", ctx, first)
				}
				activations++
			},
		})
		runtime.RegisterFocusTarget(second, FocusTargetOptions{TabIndex: 0})
		if !runtime.RequestFocus(root, first.PathID()) {
			t.Fatal("failed to focus activation target")
		}

		activate := &KeyboardEvent{Key: "Enter"}
		if allowed := runtime.DispatchKeyboardEvent(root, 0, activate); !allowed {
			t.Fatal("activation event was unexpectedly canceled")
		}
		if activations != 1 || activate.Type != EventTypeKeyDown || !activate.Bubbles || !activate.Cancelable {
			t.Fatalf("unexpected activation result: activations=%d event=%#v", activations, activate)
		}
		if activate.Detail != activate || !activate.Time.Equal(root.Now()) {
			t.Fatalf("keyboard event detail/time were not initialized: %#v", activate)
		}

		runtime.DispatchKeyboardEvent(root, 0, &KeyboardEvent{Key: "Tab"})
		if runtime.FocusedTarget() != second.PathID() {
			t.Fatalf("tab target = %d, want %d", runtime.FocusedTarget(), second.PathID())
		}
		runtime.DispatchKeyboardEvent(root, 0, &KeyboardEvent{Key: "Tab", Modifiers: Modifiers{Shift: true}})
		if runtime.FocusedTarget() != first.PathID() {
			t.Fatalf("shift-tab target = %d, want %d", runtime.FocusedTarget(), first.PathID())
		}

		firstDown := &KeyboardEvent{Key: "a", Code: "KeyA"}
		secondDown := &KeyboardEvent{Key: "a", Code: "KeyA"}
		runtime.DispatchKeyboardEvent(root, first.PathID(), firstDown)
		runtime.DispatchKeyboardEvent(root, first.PathID(), secondDown)
		if firstDown.Repeat || !secondDown.Repeat {
			t.Fatalf("unexpected key repeat state: first=%t second=%t", firstDown.Repeat, secondDown.Repeat)
		}
		runtime.DispatchKeyboardEvent(root, first.PathID(), &KeyboardEvent{Event: Event{Type: EventTypeKeyUp}, Key: "a", Code: "KeyA"})
		thirdDown := &KeyboardEvent{Key: "a", Code: "KeyA"}
		runtime.DispatchKeyboardEvent(root, first.PathID(), thirdDown)
		if thirdDown.Repeat {
			t.Fatal("key remained pressed after keyup")
		}

		runtime.RegisterEventListener(first, EventTypeKeyDown, func(_ *Context, event *Event) {
			event.PreventDefault()
		}, EventListenerOptions{})
		if allowed := runtime.DispatchKeyboardEvent(root, 0, &KeyboardEvent{Key: "Enter"}); allowed {
			t.Fatal("cancelable keydown was not canceled")
		}
		if activations != 1 {
			t.Fatalf("canceled keydown ran activation default: %d", activations)
		}
	})

	t.Run("shortcuts use nearest scope, priority, and once semantics", func(t *testing.T) {
		runtime, root, parent, child := newEventTestTree(t)
		runtime.RegisterFocusTarget(child, FocusTargetOptions{TabIndex: 0})
		if !runtime.RequestFocus(root, child.PathID()) {
			t.Fatal("failed to focus shortcut target")
		}

		var seen []string
		spec := Shortcut{Key: "k", Modifiers: Modifiers{Ctrl: true}}
		runtime.RegisterShortcut(child, spec, func(_ *Context, _ *KeyboardEvent) {
			seen = append(seen, "child")
		}, EventListenerOptions{})
		runtime.RegisterShortcut(child, spec, func(_ *Context, _ *KeyboardEvent) {
			seen = append(seen, "child once")
		}, EventListenerOptions{Priority: 5, Once: true})
		runtime.RegisterShortcut(parent, spec, func(_ *Context, _ *KeyboardEvent) {
			seen = append(seen, "parent")
		}, EventListenerOptions{Priority: 10})

		runtime.DispatchKeyboardEvent(root, 0, &KeyboardEvent{Key: "k", Modifiers: Modifiers{Ctrl: true}})
		assertStrings(t, seen, []string{"child once", "child", "parent"})
		seen = nil
		runtime.DispatchKeyboardEvent(root, 0, &KeyboardEvent{Key: "k", Modifiers: Modifiers{Ctrl: true}})
		assertStrings(t, seen, []string{"child", "parent"})
	})
}

func TestEventHelpersAndPointerCapture(t *testing.T) {
	runtime, root, parent, child := newEventTestTree(t)

	runtime.SetPointerCapture(7, child.PathID())
	if target, ok := runtime.PointerCaptureTarget(7); !ok || target != child.PathID() {
		t.Fatalf("pointer capture = %d, %t; want %d, true", target, ok, child.PathID())
	}
	if !runtime.HasPointerCapture(7, child.PathID()) || runtime.HasPointerCapture(7, parent.PathID()) {
		t.Fatal("pointer capture ownership was not reported correctly")
	}
	if runtime.ReleasePointerCapture(7, parent.PathID()) {
		t.Fatal("non-owner released pointer capture")
	}
	if !runtime.ReleasePointerCapture(7, child.PathID()) || runtime.ReleasePointerCapture(7, child.PathID()) {
		t.Fatal("pointer capture release semantics were incorrect")
	}
	runtime.SetPointerCapture(8, child.PathID())
	if !runtime.ReleasePointerCapture(8, 0) {
		t.Fatal("unconditional pointer capture release failed")
	}

	runtime.RegisterEventTargetOptions(child.PathID(), parent.PathID(), EventTargetOptions{
		Boundary: EventBoundaryPolicy{Mode: EventBoundaryRedirect, Redirect: root.PathID()},
	})
	entry := runtime.events.targets[child.PathID()]
	if entry.Parent != parent.PathID() || entry.Boundary.Mode != EventBoundaryRedirect || entry.Boundary.Redirect != root.PathID() {
		t.Fatalf("event target options were not applied: %#v", entry)
	}
	runtime.RegisterEventTargetOptions(child.PathID(), root.PathID(), EventTargetOptions{})
	entry = runtime.events.targets[child.PathID()]
	if entry.Parent != root.PathID() || entry.Boundary.Mode != EventBoundaryRedirect {
		t.Fatalf("event target update lost its boundary: %#v", entry)
	}

	if targets, listeners, focusTargets, shortcuts := runtime.eventRegistryCounts(); targets < 3 || listeners != 0 || focusTargets != 0 || shortcuts != 0 {
		t.Fatalf("unexpected registry counts: targets=%d listeners=%d focus=%d shortcuts=%d", targets, listeners, focusTargets, shortcuts)
	}

	if keyboardIdentity(nil) != "" || keyboardIdentity(&KeyboardEvent{Key: "k"}) != "k" || keyboardIdentity(&KeyboardEvent{Key: "k", Code: "KeyK"}) != "KeyK" {
		t.Fatal("keyboard identities were not normalized")
	}
	if shortcutMatches(Shortcut{}, nil) {
		t.Fatal("nil keyboard event matched a shortcut")
	}
	if !shortcutMatches(Shortcut{Key: "k", Modifiers: Modifiers{Ctrl: true}}, &KeyboardEvent{Key: "k", Modifiers: Modifiers{Ctrl: true, Shift: true}}) {
		t.Fatal("shortcut did not allow additional modifiers")
	}
	if shortcutMatches(Shortcut{Key: "k", Modifiers: Modifiers{Ctrl: true}, ExactModifiers: true}, &KeyboardEvent{Key: "k", Modifiers: Modifiers{Ctrl: true, Shift: true}}) {
		t.Fatal("exact shortcut accepted additional modifier")
	}
	if shortcutMatches(Shortcut{Code: "KeyK"}, &KeyboardEvent{Key: "k", Code: "KeyJ"}) {
		t.Fatal("shortcut accepted incorrect key code")
	}
	if modifiersContain(Modifiers{}, Modifiers{Ctrl: true}) || modifiersContain(Modifiers{}, Modifiers{Shift: true}) || modifiersContain(Modifiers{}, Modifiers{Alt: true}) || modifiersContain(Modifiers{}, Modifiers{Meta: true}) || modifiersContain(Modifiers{}, Modifiers{Shortcut: true}) {
		t.Fatal("missing modifier was accepted")
	}
	if !modifiersEqual(Modifiers{Ctrl: true}, Modifiers{Ctrl: true}) || modifiersEqual(Modifiers{Ctrl: true}, Modifiers{Shift: true}) {
		t.Fatal("modifier equality was incorrect")
	}
	if pathIndex([]PathID{child.PathID(), root.PathID()}, 0) != 1 || pathIndex([]PathID{child.PathID()}, parent.PathID()) != -1 {
		t.Fatal("path index normalization was incorrect")
	}
	if normalizeOptionalPathID(0) != 0 || normalizeOptionalPathID(child.PathID()) != child.PathID() {
		t.Fatal("optional path normalization was incorrect")
	}
	if got := normalizeEventBoundary(EventBoundaryPolicy{Mode: EventBoundaryStop, Redirect: child.PathID()}); got.Redirect != 0 {
		t.Fatalf("stop boundary retained a redirect: %#v", got)
	}
	if got := normalizeEventBoundary(EventBoundaryPolicy{Mode: EventBoundaryNone, Redirect: child.PathID()}); got != (EventBoundaryPolicy{}) {
		t.Fatalf("none boundary was not reset: %#v", got)
	}

	var nilRuntime *Runtime
	nilRuntime.SetPointerCapture(1, child.PathID())
	if nilRuntime.ReleasePointerCapture(1, child.PathID()) || nilRuntime.Focused(child.PathID()) || nilRuntime.EventPathContains(child.PathID(), root.PathID()) {
		t.Fatal("nil runtime event helpers should be inert")
	}
	if nilRuntime.DispatchEvent(root, child.PathID(), &Event{Type: "x", Cancelable: true, DefaultPrevented: true}) {
		t.Fatal("nil runtime accepted a pre-canceled event")
	}
	if nilRuntime.DispatchKeyboardEvent(root, child.PathID(), &KeyboardEvent{Event: Event{Cancelable: true, DefaultPrevented: true}}) {
		t.Fatal("nil runtime accepted a pre-canceled keyboard event")
	}
}
