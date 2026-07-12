package event

import (
	"reflect"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	gioLayout "gioui.org/layout"
)

func TestPointerGioConversionHelpers(t *testing.T) {
	if !(ButtonsPrimary | ButtonsSecondary).Contain(ButtonsPrimary | ButtonsSecondary) {
		t.Fatal("Contain did not find all requested buttons")
	}
	if (ButtonsPrimary | ButtonsSecondary).Contains(ButtonsAuxiliary) {
		t.Fatal("Contains reported a button that is not set")
	}

	for _, tc := range []struct {
		source pointer.Source
		want   PointerType
	}{
		{pointer.Mouse, PointerMouse},
		{pointer.Touch, PointerTouch},
		{pointer.Source(99), PointerOther},
	} {
		if got := PointerTypeFromGio(tc.source); got != tc.want {
			t.Errorf("PointerTypeFromGio(%v) = %q, want %q", tc.source, got, tc.want)
		}
	}

	for _, tc := range []struct {
		kind pointer.Kind
		want Type
		ok   bool
	}{
		{pointer.Press, PointerDown, true},
		{pointer.Release, PointerUp, true},
		{pointer.Move, PointerMove, true},
		{pointer.Drag, PointerMove, true},
		{pointer.Enter, PointerEnter, true},
		{pointer.Leave, PointerLeave, true},
		{pointer.Cancel, PointerCancel, true},
		{pointer.Scroll, Wheel, true},
		{pointer.Kind(1 << 20), "", false},
	} {
		got, ok := TypeFromGioPointerKind(tc.kind)
		if got != tc.want || ok != tc.ok {
			t.Errorf("TypeFromGioPointerKind(%v) = %q, %t; want %q, %t", tc.kind, got, ok, tc.want, tc.ok)
		}
	}

	for _, tc := range []struct {
		buttons pointer.Buttons
		want    Button
	}{
		{pointer.ButtonPrimary, ButtonPrimary},
		{pointer.ButtonTertiary, ButtonAuxiliary},
		{pointer.ButtonSecondary, ButtonSecondary},
		{pointer.ButtonQuaternary, ButtonBack},
		{pointer.ButtonQuinary, ButtonForward},
		{0, ButtonNone},
	} {
		if got := ButtonFromGio(pointer.Event{Buttons: tc.buttons}); got != tc.want {
			t.Errorf("ButtonFromGio(%v) = %d, want %d", tc.buttons, got, tc.want)
		}
	}

	allButtons := pointer.ButtonPrimary | pointer.ButtonSecondary | pointer.ButtonTertiary | pointer.ButtonQuaternary | pointer.ButtonQuinary
	if got, want := ButtonsFromGio(allButtons), ButtonsPrimary|ButtonsSecondary|ButtonsAuxiliary|ButtonsBack|ButtonsForward; got != want {
		t.Fatalf("ButtonsFromGio(all) = %b, want %b", got, want)
	}
	if got := ButtonsFromGio(0); got != 0 {
		t.Fatalf("ButtonsFromGio(0) = %b, want 0", got)
	}

	gioModifiers := key.ModCtrl | key.ModShift | key.ModAlt | key.ModCommand | key.ModSuper
	if got := ModifiersFromGio(gioModifiers); !got.Ctrl || !got.Shift || !got.Alt || !got.Meta || !got.Shortcut {
		t.Fatalf("ModifiersFromGio(%v) = %#v, want every modifier", gioModifiers, got)
	}
	if got := ModifiersFromGio(0); got != (Modifiers{}) {
		t.Fatalf("ModifiersFromGio(0) = %#v, want empty", got)
	}

	now := time.Date(2024, time.April, 5, 6, 7, 8, 9, time.UTC)
	ctx := internal.NewContext(gioLayout.Context{Now: now}, nil)
	raw := pointer.Event{
		Source:    pointer.Mouse,
		PointerID: 17,
		Buttons:   pointer.ButtonPrimary | pointer.ButtonSecondary,
		Position:  f32.Pt(12, 34),
		Scroll:    f32.Pt(-2, 5),
		Modifiers: key.ModCtrl | key.ModShift,
		Time:      47 * time.Millisecond,
	}

	gotPointer := PointerEventFromGio(ctx, PointerDown, raw)
	if gotPointer.Type != PointerDown || !gotPointer.Bubbles || !gotPointer.Cancelable || !gotPointer.Trusted || !gotPointer.Time.Equal(now) {
		t.Fatalf("PointerEventFromGio event = %#v", gotPointer.Event)
	}
	if gotPointer.PointerID != 17 || gotPointer.PointerType != PointerMouse || gotPointer.Button != ButtonPrimary || gotPointer.Buttons != (ButtonsPrimary|ButtonsSecondary) || gotPointer.Position != raw.Position || gotPointer.TimeOffset != raw.Time {
		t.Fatalf("PointerEventFromGio pointer fields = %#v", gotPointer)
	}
	if !gotPointer.Modifiers.Ctrl || !gotPointer.Modifiers.Shift {
		t.Fatalf("PointerEventFromGio modifiers = %#v", gotPointer.Modifiers)
	}

	gotWheel := WheelEventFromGio(ctx, raw)
	if gotWheel.Type != Wheel || !gotWheel.Bubbles || !gotWheel.Cancelable || !gotWheel.Trusted || !gotWheel.Time.Equal(now) || gotWheel.DeltaX != -2 || gotWheel.DeltaY != 5 || gotWheel.DeltaMode != WheelDeltaPixel || gotWheel.Position != raw.Position || gotWheel.TimeOffset != raw.Time {
		t.Fatalf("WheelEventFromGio = %#v", gotWheel)
	}

	sample := PointerSampleFromGio(raw)
	if sample.PointerID != 17 || sample.PointerType != PointerMouse || sample.Button != ButtonPrimary || sample.Buttons != (ButtonsPrimary|ButtonsSecondary) || sample.Position != raw.Position || sample.TimeOffset != raw.Time {
		t.Fatalf("PointerSampleFromGio = %#v", sample)
	}

	for _, tc := range []struct {
		name       string
		eventType  Type
		bubbles    bool
		cancelable bool
	}{
		{"default", "", true, true},
		{"enter", PointerEnter, false, false},
		{"leave", PointerLeave, false, false},
		{"cancel", PointerCancel, true, false},
		{"down", PointerDown, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ev := &PointerEvent{Event: Event{Type: tc.eventType}}
			applyPointerDefaults(ev)
			if ev.Type == "" || ev.Bubbles != tc.bubbles || ev.Cancelable != tc.cancelable {
				t.Fatalf("applyPointerDefaults(%q) = type:%q bubbles:%t cancelable:%t", tc.eventType, ev.Type, ev.Bubbles, ev.Cancelable)
			}
		})
	}

	if fallback := eventTime(nil); fallback.IsZero() {
		t.Fatal("eventTime(nil) returned zero")
	}
	if fallback := eventTime(internal.NewContext(gioLayout.Context{}, nil)); fallback.IsZero() {
		t.Fatal("eventTime with zero context time returned zero")
	}
}

func TestPointerEventsTypedListenersCaptureAndCopies(t *testing.T) {
	_, root, _, child := newEventTestTree(t)

	var nilPointer *PointerEvent
	if nilPointer.CoalescedSamples() != nil || nilPointer.ReleasePointerCapture(child) || nilPointer.HasPointerCapture(child) {
		t.Fatal("nil pointer event did not retain its no-op contract")
	}
	nilPointer.SetPointerCapture(child)
	(&PointerEvent{}).SetPointerCapture(nil)
	if (&PointerEvent{}).ReleasePointerCapture(nil) || (&PointerEvent{}).HasPointerCapture(nil) {
		t.Fatal("nil context unexpectedly owned pointer capture")
	}

	samples := []PointerSample{{PointerID: 1}, {PointerID: 2}}
	eventWithSamples := &PointerEvent{Coalesced: samples}
	copyOfSamples := eventWithSamples.CoalescedSamples()
	if !reflect.DeepEqual(copyOfSamples, samples) {
		t.Fatalf("CoalescedSamples() = %#v, want %#v", copyOfSamples, samples)
	}
	copyOfSamples[0].PointerID = 99
	if eventWithSamples.Coalesced[0].PointerID == 99 {
		t.Fatal("CoalescedSamples returned the backing slice")
	}
	if (&PointerEvent{}).CoalescedSamples() != nil {
		t.Fatal("empty CoalescedSamples should return nil")
	}

	capture := &PointerEvent{PointerID: 42}
	capture.SetPointerCapture(child)
	if !capture.HasPointerCapture(child) {
		t.Fatal("SetPointerCapture did not assign ownership")
	}
	if !capture.ReleasePointerCapture(child) || capture.HasPointerCapture(child) {
		t.Fatal("ReleasePointerCapture did not release ownership")
	}

	if !DispatchPointerEvent(root, child.PathID(), nil) || !DispatchWheelEvent(root, child.PathID(), nil) {
		t.Fatal("nil typed event dispatch should allow the default action")
	}

	pointerCalls := 0
	wheelCalls := 0
	OnPointer(child, PointerMove, nil)
	OnWheel(child, nil)
	OnPointer(child, PointerMove, func(ctx *internal.Context, event *PointerEvent) {
		pointerCalls++
		if ctx.PathID() != child.PathID() || event.Type != PointerMove || event.Detail != event || !event.Bubbles || !event.Cancelable {
			t.Fatalf("typed pointer listener received %#v", event)
		}
	})
	OnWheel(child, func(ctx *internal.Context, event *WheelEvent) {
		wheelCalls++
		if ctx.PathID() != child.PathID() || event.Type != Wheel || event.Detail != event || !event.Bubbles || !event.Cancelable {
			t.Fatalf("typed wheel listener received %#v", event)
		}
	})

	// The typed wrappers must ignore unrelated event details.
	if !DispatchEvent(root, child.PathID(), &Event{Type: PointerMove, Detail: &WheelEvent{}}) {
		t.Fatal("generic pointer event dispatch failed")
	}
	if !DispatchEvent(root, child.PathID(), &Event{Type: Wheel, Detail: &PointerEvent{}}) {
		t.Fatal("generic wheel event dispatch failed")
	}
	if pointerCalls != 0 || wheelCalls != 0 {
		t.Fatalf("typed handlers accepted unrelated details: pointer=%d wheel=%d", pointerCalls, wheelCalls)
	}

	pointerEvent := &PointerEvent{}
	if !DispatchPointerEvent(root, child.PathID(), pointerEvent) {
		t.Fatal("pointer dispatch returned false")
	}
	wheelEvent := &WheelEvent{}
	if !DispatchWheelEvent(root, child.PathID(), wheelEvent) {
		t.Fatal("wheel dispatch returned false")
	}
	if pointerCalls != 1 || wheelCalls != 1 {
		t.Fatalf("typed listener calls = pointer:%d wheel:%d, want 1/1", pointerCalls, wheelCalls)
	}
}

func TestFocusFacadeKeyboardConversionsAndShortcutHelpers(t *testing.T) {
	nilManager := FocusManagerFor(nil)
	if nilManager.Target() != 0 || nilManager.Request(1) || nilManager.Blur() || nilManager.Move(FocusForward) {
		t.Fatal("nil focus manager did not retain its no-op contract")
	}
	if RequestFocus(nil) || BlurFocus(nil) || Focused(nil) || FocusedTarget(nil) != 0 || MoveFocus(nil, FocusForward) {
		t.Fatal("nil focus helper did not retain its no-op contract")
	}
	RegisterFocusTarget(nil)
	noRuntime := internal.NewContext(gioLayout.Context{}, nil)
	if RequestFocus(noRuntime) || BlurFocus(noRuntime) || Focused(noRuntime) || FocusedTarget(noRuntime) != 0 || MoveFocus(noRuntime, FocusForward) {
		t.Fatal("focus helper accepted a context without a runtime")
	}

	_, root, parent, first := newEventTestTree(t)
	second := parent.Scope("second")
	disabled := parent.Scope("disabled")
	hidden := parent.Scope("hidden")
	programmatic := parent.Scope("programmatic")
	RegisterFocusTarget(first, FocusTabIndex(1), nil)
	RegisterFocusTarget(second, FocusTabIndex(0))
	RegisterFocusTarget(disabled, FocusDisabled(true))
	RegisterFocusTarget(hidden, FocusHidden(true))
	RegisterFocusTarget(programmatic, FocusTabIndex(-1))
	if RequestFocus(disabled) || RequestFocus(hidden) {
		t.Fatal("disabled or hidden target accepted focus")
	}
	if !RequestFocus(programmatic) {
		t.Fatal("negative tab index should remain programmatically focusable")
	}

	manager := FocusManagerFor(root)
	if !manager.Request(first.PathID()) || manager.Target() != first.PathID() || !Focused(first) || FocusedTarget(root) != first.PathID() {
		t.Fatal("FocusManager request did not set the focused target")
	}
	if !manager.Move(FocusForward) || manager.Target() != second.PathID() {
		t.Fatalf("FocusManager.Move() target = %v, want %v", manager.Target(), second.PathID())
	}
	if !MoveFocus(root, FocusBackward) || manager.Target() != first.PathID() {
		t.Fatalf("MoveFocus backward target = %v, want %v", manager.Target(), first.PathID())
	}
	if !manager.Blur() || manager.Target() != 0 || Focused(first) {
		t.Fatal("FocusManager.Blur did not clear focus")
	}

	keyUpCalls := 0
	OnKeyboard(first, KeyUp, nil)
	OnKeyUp(first, func(ctx *internal.Context, event *KeyboardEvent) {
		keyUpCalls++
		if ctx.PathID() != first.PathID() || event.Type != KeyUp || event.Detail != event {
			t.Fatalf("OnKeyUp received %#v", event)
		}
	})
	if !DispatchEvent(root, first.PathID(), &Event{Type: KeyUp, Detail: &PointerEvent{}}) {
		t.Fatal("generic keyup dispatch failed")
	}
	if keyUpCalls != 0 {
		t.Fatal("typed keyboard handler accepted unrelated detail")
	}
	if !DispatchKeyboardEvent(root, first.PathID(), &KeyboardEvent{Event: Event{Type: KeyUp}, Key: "x"}) || keyUpCalls != 1 {
		t.Fatalf("typed keyup dispatch calls = %d, want 1", keyUpCalls)
	}

	if !DispatchKeyboardEvent(nil, 0, nil) {
		t.Fatal("nil keyboard event dispatch should allow the default action")
	}
	canceled := &KeyboardEvent{Event: Event{Cancelable: true, DefaultPrevented: true}}
	if DispatchKeyboardEvent(nil, 0, canceled) {
		t.Fatal("canceled keyboard event without a runtime should return false")
	}

	now := time.Date(2024, time.January, 2, 3, 4, 5, 6, time.UTC)
	ctx := internal.NewContext(gioLayout.Context{Now: now}, nil)
	for _, tc := range []struct {
		name    key.Name
		state   key.State
		wantKey string
		wantTyp Type
	}{
		{key.NameReturn, key.Press, "Enter", KeyDown},
		{key.NameEnter, key.Press, "Enter", KeyDown},
		{key.NameSpace, key.Press, "Space", KeyDown},
		{key.NameEscape, key.Press, "Escape", KeyDown},
		{key.NameTab, key.Press, "Tab", KeyDown},
		{key.Name("x"), key.Release, "x", KeyUp},
	} {
		raw := key.Event{Name: tc.name, State: tc.state, Modifiers: key.ModCtrl}
		got := KeyboardEventFromGio(ctx, raw)
		if got.Key != tc.wantKey || got.Type != tc.wantTyp || got.Code != string(tc.name) || !got.Bubbles || !got.Cancelable || !got.Trusted || !got.Time.Equal(now) || !got.Modifiers.Ctrl || !reflect.DeepEqual(got.Native, raw) {
			t.Errorf("KeyboardEventFromGio(%#v) = %#v", raw, got)
		}
	}

	codeSpec := ShortcutCode("KeyK", Modifiers{Ctrl: true})
	if codeSpec.Code != "KeyK" || !codeSpec.Modifiers.Ctrl {
		t.Fatalf("ShortcutCode = %#v", codeSpec)
	}
	scoped := ShortcutScope(codeSpec, first.PathID())
	if scoped.Scope != first.PathID() {
		t.Fatalf("ShortcutScope = %#v, want scope %v", scoped, first.PathID())
	}
	OnShortcut(nil, scoped, nil)
}

func TestPublicOptionAndDefaultHelpers(t *testing.T) {
	options := ListenerOptions{}
	Priority(7)(&options)
	if options.Priority != 7 {
		t.Fatalf("Priority option = %#v", options)
	}

	policy := BoundaryPolicy{Mode: BoundaryRedirect, Redirect: 99}
	BoundaryStopPropagation()(&policy)
	if policy.Mode != BoundaryStop || policy.Redirect != 0 {
		t.Fatalf("BoundaryStopPropagation policy = %#v", policy)
	}

	custom := NewCustomEvent("app:custom", "detail", CustomBubbles(false), CustomCancelable(true), CustomTrusted(true), nil)
	if custom.Bubbles || !custom.Cancelable || !custom.Trusted || custom.Detail != "detail" {
		t.Fatalf("NewCustomEvent options = %#v", custom)
	}

	for _, tc := range []struct {
		name       string
		eventType  Type
		cancelable bool
	}{
		{"input default", "", false},
		{"before input", BeforeInput, true},
		{"submit", Submit, true},
		{"input", Input, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ev := &InputEvent{Event: Event{Type: tc.eventType}}
			applyInputDefaults(ev)
			if ev.Type == "" || !ev.Bubbles || ev.Cancelable != tc.cancelable {
				t.Fatalf("applyInputDefaults(%q) = %#v", tc.eventType, ev.Event)
			}
		})
	}
	composition := &CompositionEvent{}
	applyCompositionDefaults(composition)
	if composition.Type != CompositionUpdate || !composition.Bubbles || composition.Cancelable {
		t.Fatalf("applyCompositionDefaults default = %#v", composition.Event)
	}
	composition = &CompositionEvent{Event: Event{Type: CompositionStart}}
	applyCompositionDefaults(composition)
	if composition.Type != CompositionStart || !composition.Bubbles || composition.Cancelable {
		t.Fatalf("applyCompositionDefaults explicit = %#v", composition.Event)
	}

	if !DispatchDragEvent(nil, 0, nil) || !DispatchInputEvent(nil, 0, nil) || !DispatchCompositionEvent(nil, 0, nil) || !DispatchActivationEvent(nil, 0, nil) {
		t.Fatal("nil typed event dispatch should allow the default action")
	}

	if (Dispatcher{}).DispatchClickEvent(nil, &PointerEvent{}) {
		t.Fatal("DispatchClickEvent with nil context should return false")
	}
	_, _, _, child := newEventTestTree(t)
	hoverCalls := 0
	(Dispatcher{}).DispatchHover(nil, true)
	dispatcher := Dispatcher{
		Click: func(ctx *internal.Context) {
			if ctx != child {
				t.Fatalf("legacy click context = %p, want %p", ctx, child)
			}
		},
		Hover: func(ctx *internal.Context, hovering bool) {
			hoverCalls++
			if ctx != child || !hovering {
				t.Fatalf("hover callback = %p/%t", ctx, hovering)
			}
		},
	}
	dispatcher.DispatchHover(child, true)
	source := &PointerEvent{}
	if !dispatcher.DispatchClickEvent(child, source) || source.Type != Click || source.Target != child.PathID() {
		t.Fatalf("DispatchClickEvent default source = %#v", source)
	}
	if hoverCalls != 1 {
		t.Fatalf("hover calls = %d, want 1", hoverCalls)
	}
}
