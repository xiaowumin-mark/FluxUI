package internal

import (
	"testing"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestClickableStateInertQueriesAndRuntimeBinding(t *testing.T) {
	state := NewClickableState()
	if state == nil || state.raw() == nil {
		t.Fatal("new clickable state was not initialized")
	}
	state.BindRuntime(NewRuntime(nil))
	if state.Hovered() || state.Pressed() {
		t.Fatal("fresh clickable reported pointer state")
	}
	if snapshot := state.Snapshot(nil, false); snapshot != (ClickableSnapshot{}) {
		t.Fatalf("nil-context snapshot = %#v", snapshot)
	}
	if history := state.History(); len(history) != 0 {
		t.Fatalf("fresh clickable history = %#v", history)
	}

	var ops op.Ops
	context := NewContext(gioLayout.Context{Ops: &ops}, NewRuntime(nil))
	if state.Focused(context) || state.Clicked(context) {
		t.Fatal("fresh clickable accepted an input event")
	}
	if event, ok := state.ClickedEvent(context); ok || event != (ClickData{}) {
		t.Fatalf("fresh clickable event = %#v, %t", event, ok)
	}
	if snapshot := state.Snapshot(context, true); snapshot != (ClickableSnapshot{}) {
		t.Fatalf("fresh clickable focused snapshot = %#v", snapshot)
	}
	state.BindRuntime(nil)

	var nilState *ClickableState
	nilState.BindRuntime(NewRuntime(nil))
	if nilState.Clicked(nil) || nilState.Hovered() || nilState.Pressed() || nilState.Focused(nil) || nilState.Snapshot(nil, true) != (ClickableSnapshot{}) || nilState.History() != nil {
		t.Fatal("nil clickable helpers should be inert")
	}
	if event, ok := nilState.ClickedEvent(nil); ok || event != (ClickData{}) {
		t.Fatalf("nil clickable event = %#v, %t", event, ok)
	}
}
