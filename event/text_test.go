package event

import (
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"
)

func TestInputEventDispatchCanBeCanceled(t *testing.T) {
	_, root, _, child := newEventTestTree(t)

	var got *InputEvent
	var gotTarget TargetID
	var gotPhase Phase
	OnInput(child, BeforeInput, func(ctx *internal.Context, ev *InputEvent) {
		got = ev
		gotTarget = ev.Target
		gotPhase = ev.Phase
		if ev.Data == "x" {
			ev.PreventDefault()
		}
	})

	ev := &InputEvent{
		Event:         Event{Type: BeforeInput},
		Data:          "x",
		InputType:     InputTypeInsertText,
		Source:        InputSourceUser,
		Value:         "x",
		PreviousValue: "",
	}
	if DispatchInputEvent(root, child.PathID(), ev) {
		t.Fatal("beforeinput dispatch returned true after preventDefault")
	}
	if got == nil || got.Type != BeforeInput || gotTarget != child.PathID() || gotPhase != PhaseTarget {
		t.Fatalf("unexpected beforeinput event: %#v", got)
	}
	if !ev.DefaultPrevented {
		t.Fatal("beforeinput DefaultPrevented = false, want true")
	}
}

func TestCompositionEventSyntheticDispatch(t *testing.T) {
	_, root, _, child := newEventTestTree(t)

	var got *CompositionEvent
	OnComposition(child, CompositionUpdate, func(ctx *internal.Context, ev *CompositionEvent) {
		got = ev
	})

	ev := &CompositionEvent{
		Event: Event{Type: CompositionUpdate},
		Data:  "pin",
	}
	if !DispatchCompositionEvent(root, child.PathID(), ev) {
		t.Fatal("composition dispatch returned false")
	}
	if got == nil || got.Data != "pin" || got.Type != CompositionUpdate || got.Target != child.PathID() {
		t.Fatalf("unexpected composition event: %#v", got)
	}
}
