package event

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"

	gioLayout "gioui.org/layout"
)

func newEventTestTree(t *testing.T) (*internal.Runtime, *internal.Context, *internal.Context, *internal.Context) {
	t.Helper()
	rt := internal.NewRuntime(nil)
	rt.BeginFrame()
	root := internal.NewContext(gioLayout.Context{}, rt)
	parent := root.Child(0)
	child := parent.Scope("child")
	t.Cleanup(rt.EndFrame)
	return rt, root, parent, child
}

func TestEventDiagnosticsRecordPathDurationAndCancellation(t *testing.T) {
	var logs bytes.Buffer
	rt := internal.NewRuntime(nil)
	rt.SetPerfDiagnostics(internal.PerfDiagnostics{
		Enabled:          true,
		MeasureDurations: true,
		LogEvents:        true,
		Writer:           &logs,
	})
	rt.BeginFrame()
	root := internal.NewContext(gioLayout.Context{}, rt)
	parent := root.Child(0)
	child := parent.Scope("child")

	On(parent, Click, func(ctx *internal.Context, ev *Event) {
		time.Sleep(time.Microsecond)
	}, Capture())
	On(child, Click, func(ctx *internal.Context, ev *Event) {
		ev.PreventDefault()
		ev.StopPropagation()
	})

	ev := Event{Type: Click, Bubbles: true, Cancelable: true}
	if DispatchEvent(root, child.PathID(), &ev) {
		t.Fatal("dispatch returned true after preventDefault")
	}
	rt.EndFrame()

	stats := rt.LastFrameStats()
	if stats.Events.Dispatches != 1 || stats.Events.ListenerCalls != 2 {
		t.Fatalf("event diagnostics dispatch/listeners = %d/%d, want 1/2", stats.Events.Dispatches, stats.Events.ListenerCalls)
	}
	if stats.Events.PointerEvents != 1 || stats.Interaction.PointerEvents != 1 {
		t.Fatalf("pointer event diagnostics = events:%d interaction:%d, want 1/1", stats.Events.PointerEvents, stats.Interaction.PointerEvents)
	}
	if stats.Events.DefaultPrevented != 1 || stats.Events.PropagationStopped != 1 {
		t.Fatalf("cancel diagnostics = prevented:%d stopped:%d, want 1/1", stats.Events.DefaultPrevented, stats.Events.PropagationStopped)
	}
	if stats.Events.ListenerDuration <= 0 {
		t.Fatalf("listener duration = %s, want > 0", stats.Events.ListenerDuration)
	}
	if stats.Events.TargetsRegistered != 3 || stats.Events.ListenersRegistered != 2 {
		t.Fatalf("registry diagnostics targets/listeners = %d/%d, want 3/2", stats.Events.TargetsRegistered, stats.Events.ListenersRegistered)
	}
	if stats.Events.LastPreventDefaultTarget != "root/0/child" || stats.Events.LastPreventDefaultPhase != "target" {
		t.Fatalf("preventDefault diagnostics = %q/%q, want child/target", stats.Events.LastPreventDefaultTarget, stats.Events.LastPreventDefaultPhase)
	}
	if stats.Events.LastStopTarget != "root/0/child" || stats.Events.LastStopPhase != "target" {
		t.Fatalf("stop diagnostics = %q/%q, want child/target", stats.Events.LastStopTarget, stats.Events.LastStopPhase)
	}
	if !stats.Events.LastDefaultPrevented || !stats.Events.LastPropagationStopped || stats.Events.LastDefaultAllowed {
		t.Fatalf("last dispatch flags = prevented:%t stopped:%t allowed:%t, want true/true/false", stats.Events.LastDefaultPrevented, stats.Events.LastPropagationStopped, stats.Events.LastDefaultAllowed)
	}

	line := logs.String()
	for _, want := range []string{
		`event type="click"`,
		`target="root/0/child"`,
		`path="root/0/child>root/0>root"`,
		`listeners=2`,
		`default_prevented=true`,
		`propagation_stopped=true`,
		`default_allowed=false`,
		`prevent_default_target="root/0/child"`,
		`prevent_default_phase="target"`,
		`stop_target="root/0/child"`,
		`stop_phase="target"`,
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("event log missing %q in %s", want, line)
		}
	}

	formatted := internal.FormatFrameStats(stats)
	for _, want := range []string{
		"event_dispatches=1",
		"event_listeners=2",
		"event_targets=3",
		"event_registered_listeners=2",
		"event_default_prevented=1",
		`event_last_type="click"`,
		`event_last_prevent_default_target="root/0/child"`,
		`event_last_stop_target="root/0/child"`,
	} {
		if !strings.Contains(formatted, want) {
			t.Fatalf("frame stats missing %q in %s", want, formatted)
		}
	}
}

func TestEventDiagnosticsClassifyPointerWheelKeyboardFocus(t *testing.T) {
	rt := internal.NewRuntime(nil)
	rt.SetPerfDiagnostics(internal.PerfDiagnostics{
		Enabled:          true,
		MeasureDurations: true,
	})
	rt.BeginFrame()
	root := internal.NewContext(gioLayout.Context{}, rt)
	child := root.Scope("child")

	RegisterFocusTarget(child)
	DispatchPointerEvent(root, child.PathID(), &PointerEvent{Event: Event{Type: PointerMove}})
	DispatchWheelEvent(root, child.PathID(), &WheelEvent{})
	DispatchKeyboardEvent(root, child.PathID(), &KeyboardEvent{
		Event: Event{Type: KeyDown},
		Key:   "A",
		Code:  "KeyA",
	})
	if !RequestFocus(child) {
		t.Fatal("RequestFocus returned false")
	}
	rt.EndFrame()

	stats := rt.LastFrameStats()
	if stats.Events.PointerEvents != 1 || stats.Interaction.PointerEvents != 1 {
		t.Fatalf("pointer diagnostics = events:%d interaction:%d, want 1/1", stats.Events.PointerEvents, stats.Interaction.PointerEvents)
	}
	if stats.Events.WheelEvents != 1 || stats.Interaction.WheelEvents != 1 {
		t.Fatalf("wheel diagnostics = events:%d interaction:%d, want 1/1", stats.Events.WheelEvents, stats.Interaction.WheelEvents)
	}
	if stats.Events.KeyboardEvents != 1 || stats.Interaction.KeyboardEvents != 1 {
		t.Fatalf("keyboard diagnostics = events:%d interaction:%d, want 1/1", stats.Events.KeyboardEvents, stats.Interaction.KeyboardEvents)
	}
	if stats.Events.FocusEvents != 2 || stats.Interaction.FocusEvents != 2 {
		t.Fatalf("focus diagnostics = events:%d interaction:%d, want 2/2", stats.Events.FocusEvents, stats.Interaction.FocusEvents)
	}
	if stats.Events.FocusTargetsRegistered != 1 {
		t.Fatalf("focus target diagnostics = %d, want 1", stats.Events.FocusTargetsRegistered)
	}
	if stats.Events.LastType == "" || stats.Events.LastTarget == "" || stats.Events.LastPath == "" {
		t.Fatalf("last dispatch diagnostics missing type/target/path: %#v", stats.Events)
	}
}

func TestEventDiagnosticsPathRewriteSummary(t *testing.T) {
	var logs bytes.Buffer
	rt := internal.NewRuntime(nil)
	rt.SetPerfDiagnostics(internal.PerfDiagnostics{
		Enabled:   true,
		LogEvents: true,
		Writer:    &logs,
	})
	rt.BeginFrame()
	root := internal.NewContext(gioLayout.Context{}, rt)
	owner := root.Child(0)
	eventType := Type("app:rewrite")

	portal := root.Child(1)
	RegisterPortal(portal, owner.PathID())
	insidePortal := portal.Child(0)
	portalEvent := NewCustomEvent(eventType, nil)
	DispatchEvent(root, insidePortal.PathID(), &portalEvent)

	boundary := root.Child(2)
	RegisterBoundary(boundary)
	insideBoundary := boundary.Child(0)
	stopEvent := NewCustomEvent(eventType, nil)
	DispatchEvent(root, insideBoundary.PathID(), &stopEvent)

	redirectTarget := root.Child(3)
	redirectBoundary := root.Child(4)
	RegisterBoundary(redirectBoundary, BoundaryRedirectTo(redirectTarget.PathID()))
	insideRedirect := redirectBoundary.Child(0)
	redirectEvent := NewCustomEvent(eventType, nil)
	DispatchEvent(root, insideRedirect.PathID(), &redirectEvent)
	rt.EndFrame()

	line := logs.String()
	for _, want := range []string{
		`path_rewrite="portal:root/1->root/0"`,
		`path_rewrite="boundary-stop:root/2"`,
		`path_rewrite="boundary-redirect:root/4->root/3"`,
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("event rewrite log missing %q in %s", want, line)
		}
	}

	stats := rt.LastFrameStats()
	if stats.Events.LastPathRewrite != "boundary-redirect:root/4->root/3" {
		t.Fatalf("last path rewrite = %q, want redirect summary", stats.Events.LastPathRewrite)
	}
}

func TestRedrawDiagnosticsRecordSourceAndOwnerPath(t *testing.T) {
	rt := internal.NewRuntime(nil)
	rt.SetPerfDiagnostics(internal.PerfDiagnostics{Enabled: true})
	rt.BeginFrame()
	root := internal.NewContext(gioLayout.Context{}, rt)
	owner := root.Child(0)

	owner.RequestRedrawReason("test.redraw")
	rt.EndFrame()

	stats := rt.LastFrameStats()
	if got := stats.ReasonCounts["test.redraw"]; got != 1 {
		t.Fatalf("reason count = %d, want 1", got)
	}
	if len(stats.RedrawReasons) != 1 {
		t.Fatalf("redraw reasons = %#v, want one record", stats.RedrawReasons)
	}
	record := stats.RedrawReasons[0]
	if record.Reason != "test.redraw" || record.Source != "context" || record.OwnerPath != "root/0" || record.Count != 1 {
		t.Fatalf("redraw record = %#v, want context root/0 test.redraw", record)
	}
	formatted := internal.FormatFrameStats(stats)
	if !strings.Contains(formatted, `redraw=test.redraw@context@root/0`) {
		t.Fatalf("formatted redraw diagnostics missing owner in %s", formatted)
	}
}

func TestDispatchOrderCaptureTargetBubble(t *testing.T) {
	_, root, parent, child := newEventTestTree(t)
	eventType := Type("custom")
	var calls []string

	On(root, eventType, func(ctx *internal.Context, ev *Event) {
		calls = append(calls, "root-capture")
		if ctx.PathID() != root.PathID() || ev.CurrentTarget != root.PathID() || ev.Phase != PhaseCapture {
			t.Fatalf("root capture ctx/path/phase mismatch: ctx=%v current=%v phase=%v", ctx.PathID(), ev.CurrentTarget, ev.Phase)
		}
	}, Capture())
	On(parent, eventType, func(ctx *internal.Context, ev *Event) {
		calls = append(calls, "parent-capture")
		if ctx.PathID() != parent.PathID() || ev.CurrentTarget != parent.PathID() || ev.Phase != PhaseCapture {
			t.Fatalf("parent capture ctx/path/phase mismatch")
		}
	}, Capture())
	On(child, eventType, func(ctx *internal.Context, ev *Event) {
		calls = append(calls, "child-target")
		if ctx.PathID() != child.PathID() || ev.CurrentTarget != child.PathID() || ev.Phase != PhaseTarget {
			t.Fatalf("child target ctx/path/phase mismatch")
		}
	})
	On(parent, eventType, func(ctx *internal.Context, ev *Event) {
		calls = append(calls, "parent-bubble")
		if ctx.PathID() != parent.PathID() || ev.CurrentTarget != parent.PathID() || ev.Phase != PhaseBubble {
			t.Fatalf("parent bubble ctx/path/phase mismatch")
		}
	})
	On(root, eventType, func(ctx *internal.Context, ev *Event) {
		calls = append(calls, "root-bubble")
		if ctx.PathID() != root.PathID() || ev.CurrentTarget != root.PathID() || ev.Phase != PhaseBubble {
			t.Fatalf("root bubble ctx/path/phase mismatch")
		}
	})

	ev := Event{Type: eventType, Bubbles: true, Cancelable: true}
	if !DispatchEvent(root, child.PathID(), &ev) {
		t.Fatal("dispatch returned false without preventDefault")
	}

	want := []string{"root-capture", "parent-capture", "child-target", "parent-bubble", "root-bubble"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
	if got, want := ev.ComposedPath(), []TargetID{child.PathID(), parent.PathID(), root.PathID()}; !reflect.DeepEqual(got, want) {
		t.Fatalf("composed path = %v, want %v", got, want)
	}
}

func TestStopPropagationPreventsLaterTargets(t *testing.T) {
	_, root, parent, child := newEventTestTree(t)
	eventType := Type("custom")
	var calls []string

	On(root, eventType, func(ctx *internal.Context, ev *Event) {
		calls = append(calls, "root-capture")
		ev.StopPropagation()
	}, Capture())
	On(parent, eventType, func(ctx *internal.Context, ev *Event) {
		calls = append(calls, "parent-capture")
	}, Capture())
	On(child, eventType, func(ctx *internal.Context, ev *Event) {
		calls = append(calls, "child-target")
	})

	Dispatch(root, child.PathID(), Event{Type: eventType, Bubbles: true})

	want := []string{"root-capture"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestStopImmediatePropagationSkipsSameTargetListeners(t *testing.T) {
	_, root, _, child := newEventTestTree(t)
	eventType := Type("custom")
	var calls []string

	On(root, eventType, func(ctx *internal.Context, ev *Event) {
		calls = append(calls, "first")
		ev.StopImmediatePropagation()
	}, Capture())
	On(root, eventType, func(ctx *internal.Context, ev *Event) {
		calls = append(calls, "second")
	}, Capture())
	On(child, eventType, func(ctx *internal.Context, ev *Event) {
		calls = append(calls, "child")
	})

	Dispatch(root, child.PathID(), Event{Type: eventType, Bubbles: true})

	want := []string{"first"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestOnceListenerIsRemovedForCurrentFrame(t *testing.T) {
	_, root, _, child := newEventTestTree(t)
	eventType := Type("custom")
	calls := 0

	On(child, eventType, func(ctx *internal.Context, ev *Event) {
		calls++
	}, Once())

	Dispatch(root, child.PathID(), Event{Type: eventType})
	Dispatch(root, child.PathID(), Event{Type: eventType})

	if calls != 1 {
		t.Fatalf("once listener calls = %d, want 1", calls)
	}
}

func TestPreventDefaultAndPassiveListener(t *testing.T) {
	_, root, _, child := newEventTestTree(t)
	eventType := Type("custom")

	On(child, eventType, func(ctx *internal.Context, ev *Event) {
		if !ev.PreventDefault() {
			t.Fatal("PreventDefault returned false for cancelable non-passive listener")
		}
	})
	ev := Event{Type: eventType, Cancelable: true}
	if DispatchEvent(root, child.PathID(), &ev) {
		t.Fatal("dispatch returned true after preventDefault on cancelable event")
	}
	if !ev.DefaultPrevented {
		t.Fatal("DefaultPrevented = false, want true")
	}

	rt, root, _, child := newEventTestTree(t)
	_ = rt
	On(child, eventType, func(ctx *internal.Context, ev *Event) {
		if ev.PreventDefault() {
			t.Fatal("passive listener should not be able to preventDefault")
		}
	}, Passive())
	passiveEvent := Event{Type: eventType, Cancelable: true}
	if !DispatchEvent(root, child.PathID(), &passiveEvent) {
		t.Fatal("passive dispatch returned false")
	}
	if passiveEvent.DefaultPrevented {
		t.Fatal("passive listener changed DefaultPrevented")
	}
}

func TestDispatchClickBridgesLegacyHandler(t *testing.T) {
	_, root, _, child := newEventTestTree(t)
	legacyCalls := 0
	eventCalls := 0

	On(child, Click, func(ctx *internal.Context, ev *Event) {
		eventCalls++
		if ev.Target != child.PathID() || ev.Phase != PhaseTarget {
			t.Fatalf("click event target/phase = %v/%v, want %v/%v", ev.Target, ev.Phase, child.PathID(), PhaseTarget)
		}
	})
	Dispatcher{Click: func(ctx *internal.Context) {
		legacyCalls++
	}}.DispatchClick(child)

	if eventCalls != 1 || legacyCalls != 1 {
		t.Fatalf("eventCalls=%d legacyCalls=%d, want 1/1", eventCalls, legacyCalls)
	}

	rt, root, _, child := newEventTestTree(t)
	_ = rt
	legacyCalls = 0
	On(child, Click, func(ctx *internal.Context, ev *Event) {
		ev.PreventDefault()
	})
	Dispatcher{Click: func(ctx *internal.Context) {
		legacyCalls++
	}}.DispatchClick(child)

	if legacyCalls != 0 {
		t.Fatalf("legacy click calls after preventDefault = %d, want 0", legacyCalls)
	}
	_ = root
}

func TestDispatchDragEventDefaultsAndPreventDefault(t *testing.T) {
	_, root, parent, child := newEventTestTree(t)
	var calls []string
	var received DragEvent

	OnDrag(parent, DragOver, func(ctx *internal.Context, ev *DragEvent) {
		calls = append(calls, "parent-capture")
		if ev.Phase != PhaseCapture || ev.CurrentTarget != parent.PathID() {
			t.Fatalf("capture phase/current = %v/%v, want %v/%v", ev.Phase, ev.CurrentTarget, PhaseCapture, parent.PathID())
		}
	}, Capture())
	OnDrag(child, DragOver, func(ctx *internal.Context, ev *DragEvent) {
		calls = append(calls, "child-target")
		received = *ev
		ev.PreventDefault()
	})
	OnDrag(parent, DragOver, func(ctx *internal.Context, ev *DragEvent) {
		calls = append(calls, "parent-bubble")
	})

	ev := &DragEvent{
		Event: Event{Type: DragOver},
		Types: []string{"text/plain"},
	}
	if DispatchDragEvent(root, child.PathID(), ev) {
		t.Fatal("dragover dispatch returned true after preventDefault")
	}
	if !ev.Bubbles || !ev.Cancelable || !ev.DefaultPrevented || ev.Detail != ev {
		t.Fatalf("dragover defaults = bubbles:%v cancelable:%v prevented:%v detail:%T", ev.Bubbles, ev.Cancelable, ev.DefaultPrevented, ev.Detail)
	}
	if !reflect.DeepEqual(calls, []string{"parent-capture", "child-target", "parent-bubble"}) {
		t.Fatalf("dragover calls = %v", calls)
	}
	if !reflect.DeepEqual(received.Types, []string{"text/plain"}) {
		t.Fatalf("received drag types = %v", received.Types)
	}
}

func TestDragLeaveAndDragEndAreNotCancelable(t *testing.T) {
	_, root, _, child := newEventTestTree(t)
	for _, eventType := range []Type{DragLeave, DragEnd} {
		OnDrag(child, eventType, func(ctx *internal.Context, ev *DragEvent) {
			if ev.PreventDefault() {
				t.Fatalf("%s should not be cancelable", eventType)
			}
		})
		ev := &DragEvent{Event: Event{Type: eventType}}
		if !DispatchDragEvent(root, child.PathID(), ev) {
			t.Fatalf("%s dispatch returned false", eventType)
		}
		if ev.Cancelable || ev.DefaultPrevented {
			t.Fatalf("%s cancelable/defaultPrevented = %v/%v, want false/false", eventType, ev.Cancelable, ev.DefaultPrevented)
		}
	}
}

func TestCustomEventDetailAndActivation(t *testing.T) {
	_, root, parent, child := newEventTestTree(t)
	customType := Type("app:selection-change")
	payload := map[string]string{"id": "42"}
	var gotDetail any
	var gotPhase Phase

	On(child, customType, func(ctx *internal.Context, ev *Event) {
		ev.PreventDefault()
	})
	On(parent, customType, func(ctx *internal.Context, ev *Event) {
		gotDetail = ev.Detail
		gotPhase = ev.Phase
	})

	if DispatchCustomEvent(root, child.PathID(), customType, payload, CustomCancelable(true)) {
		t.Fatal("custom dispatch returned true after preventDefault")
	}
	if !reflect.DeepEqual(gotDetail, payload) || gotPhase != PhaseBubble {
		t.Fatalf("custom detail/phase = %v/%v, want %v/%v", gotDetail, gotPhase, payload, PhaseBubble)
	}

	var activation ActivationEvent
	OnActivate(child, func(ctx *internal.Context, ev *ActivationEvent) {
		activation = *ev
		ev.PreventDefault()
	})
	if DispatchActivationEvent(root, child.PathID(), &ActivationEvent{
		Source:             ActivationSourceKeyboard,
		KeyboardEquivalent: "Enter",
	}) {
		t.Fatal("activation dispatch returned true after preventDefault")
	}
	if activation.Source != ActivationSourceKeyboard || activation.KeyboardEquivalent != "Enter" || activation.Type != Activate {
		t.Fatalf("activation event = source:%q key:%q type:%q", activation.Source, activation.KeyboardEquivalent, activation.Type)
	}
}

func TestPortalAndBoundaryEventPathRules(t *testing.T) {
	_, root, owner, _ := newEventTestTree(t)
	eventType := Type("app:portal-click")
	portal := root.Child(1)
	RegisterPortal(portal, owner.PathID())
	insidePortal := portal.Child(0)
	var portalCalls []string

	On(owner, eventType, func(ctx *internal.Context, ev *Event) {
		portalCalls = append(portalCalls, "owner")
	})
	On(root, eventType, func(ctx *internal.Context, ev *Event) {
		portalCalls = append(portalCalls, "root")
	})
	ev := NewCustomEvent(eventType, nil)
	DispatchEvent(root, insidePortal.PathID(), &ev)

	if !reflect.DeepEqual(portalCalls, []string{"owner", "root"}) {
		t.Fatalf("portal calls = %v, want owner/root", portalCalls)
	}
	if got, want := ev.ComposedPath(), []TargetID{insidePortal.PathID(), portal.PathID(), owner.PathID(), root.PathID()}; !reflect.DeepEqual(got, want) {
		t.Fatalf("portal path = %v, want %v", got, want)
	}

	boundary := root.Child(2)
	RegisterBoundary(boundary)
	insideBoundary := boundary.Child(0)
	var stopCalls []string
	On(boundary, eventType, func(ctx *internal.Context, ev *Event) {
		stopCalls = append(stopCalls, "boundary")
	})
	On(root, eventType, func(ctx *internal.Context, ev *Event) {
		stopCalls = append(stopCalls, "root")
	})
	stopEvent := NewCustomEvent(eventType, nil)
	DispatchEvent(root, insideBoundary.PathID(), &stopEvent)

	if !reflect.DeepEqual(stopCalls, []string{"boundary"}) {
		t.Fatalf("stop boundary calls = %v, want boundary", stopCalls)
	}
	if got, want := stopEvent.ComposedPath(), []TargetID{insideBoundary.PathID(), boundary.PathID()}; !reflect.DeepEqual(got, want) {
		t.Fatalf("stop boundary path = %v, want %v", got, want)
	}

	redirectTarget := root.Child(3)
	redirectBoundary := root.Child(4)
	RegisterBoundary(redirectBoundary, BoundaryRedirectTo(redirectTarget.PathID()))
	insideRedirect := redirectBoundary.Child(0)
	var redirectCalls []string
	On(redirectTarget, eventType, func(ctx *internal.Context, ev *Event) {
		redirectCalls = append(redirectCalls, "redirect")
	})
	On(root, eventType, func(ctx *internal.Context, ev *Event) {
		redirectCalls = append(redirectCalls, "root")
	})
	redirectEvent := NewCustomEvent(eventType, nil)
	DispatchEvent(root, insideRedirect.PathID(), &redirectEvent)

	if !reflect.DeepEqual(redirectCalls, []string{"redirect", "root"}) {
		t.Fatalf("redirect boundary calls = %v, want redirect/root", redirectCalls)
	}
	if got, want := redirectEvent.ComposedPath(), []TargetID{insideRedirect.PathID(), redirectBoundary.PathID(), redirectTarget.PathID(), root.PathID()}; !reflect.DeepEqual(got, want) {
		t.Fatalf("redirect boundary path = %v, want %v", got, want)
	}
}
