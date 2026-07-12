package widget

import (
	"image"
	"testing"
	"time"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	internal "github.com/xiaowumin-mark/FluxUI/internal"
	layout "github.com/xiaowumin-mark/FluxUI/layout"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

type formSubmitProbe struct {
	value   string
	allowed bool
}

func (p *formSubmitProbe) Layout(ctx *internal.Context) layout.Dimensions {
	input := &fluxevent.InputEvent{
		Event: fluxevent.Event{Type: fluxevent.Submit, Target: ctx.PathID()},
		Value: p.value,
	}
	p.allowed = fluxevent.DispatchInputEvent(ctx, ctx.PathID(), input)
	return layout.Dimensions{}
}

func layoutFormForTest(rt *internal.Runtime, w Widget, frame int) {
	var ops op.Ops
	gtx := gioLayout.Context{
		Constraints: gioLayout.Exact(image.Pt(320, 180)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Now:         time.Date(2026, 7, 12, 12, 0, frame, 0, time.UTC),
		Ops:         &ops,
	}
	rt.BeginFrame()
	if w != nil {
		w.Layout(internal.NewContext(gtx, rt).Scope("form-test"))
	}
	rt.EndFrame()
}

func TestFormBubbledSubmitCanBeCanceled(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	probe := &formSubmitProbe{value: "Ada"}
	var received *FormSubmitEvent
	form := Form(probe, FormOnSubmit(func(_ *internal.Context, event *FormSubmitEvent) {
		received = event
		event.PreventDefault()
	}))

	layoutFormForTest(rt, form, 0)

	if received == nil {
		t.Fatal("FormOnSubmit was not called for bubbling Input submit")
	}
	if received.Input == nil || received.Input.Value != "Ada" {
		t.Fatalf("submit input = %#v, want Input value Ada", received.Input)
	}
	if received.FromRef {
		t.Fatal("bubbling Input submit reported FromRef")
	}
	if !received.DefaultPrevented {
		t.Fatal("PreventDefault did not mark FormSubmitEvent")
	}
	if probe.allowed {
		t.Fatal("PreventDefault did not cancel the bubbling Input submit")
	}
}

func TestFormDisabledPreventsBubbledSubmit(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	probe := &formSubmitProbe{value: "Ada"}
	called := false
	form := Form(probe,
		FormDisabled(true),
		FormOnSubmit(func(_ *internal.Context, _ *FormSubmitEvent) {
			called = true
		}),
	)

	layoutFormForTest(rt, form, 0)

	if called {
		t.Fatal("disabled Form called FormOnSubmit")
	}
	if probe.allowed {
		t.Fatal("disabled Form did not prevent bubbling Input submit")
	}
}

func TestFormRefSubmitIsHostIntent(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	ref := NewFormRef()
	calls := 0
	form := Form(nil,
		FormAttachRef(ref),
		FormOnSubmit(func(_ *internal.Context, event *FormSubmitEvent) {
			calls++
			if !event.FromRef {
				t.Error("FormRef submit did not report FromRef")
			}
			if event.Input != nil {
				t.Error("FormRef submit unexpectedly carried an Input event")
			}
		}),
	)

	layoutFormForTest(rt, form, 0)
	ref.Submit()
	layoutFormForTest(rt, form, 1)

	if calls != 1 {
		t.Fatalf("FormRef submit calls = %d, want 1", calls)
	}
}

func TestFieldStateMessagePrecedence(t *testing.T) {
	invalid := FieldState{
		SupportingText: "normal",
		ErrorText:      "bad value",
		PendingText:    "checking",
		Status:         FieldInvalid,
	}
	if got := invalid.Message(); got != "bad value" {
		t.Fatalf("invalid message = %q, want bad value", got)
	}
	pending := invalid
	pending.Status = FieldPending
	if got := pending.Message(); got != "checking" {
		t.Fatalf("pending message = %q, want checking", got)
	}
	if !pending.Normalized().Pending {
		t.Fatal("StatusPending compatibility did not normalize Pending=true")
	}
}

func TestFieldStateRetainsErrorAndPendingMessage(t *testing.T) {
	field := FieldState{
		ErrorText:   "Email is invalid",
		PendingText: "Checking availability…",
		Pending:     true,
		Status:      FieldInvalid,
	}

	if !field.IsInvalid() {
		t.Fatal("invalid field reported valid")
	}
	if !field.IsPending() {
		t.Fatal("pending field reported not pending")
	}
	if got := field.Message(); got != "Email is invalid" {
		t.Fatalf("primary message = %q, want error text", got)
	}
	if got := field.PendingMessage(); got != "Checking availability…" {
		t.Fatalf("pending message = %q, want pending text", got)
	}
	message, pendingMessage := formFieldMessages(field)
	if message != "Email is invalid" || pendingMessage != "Checking availability…" {
		t.Fatalf("FormField messages = (%q, %q), want separate error and pending text", message, pendingMessage)
	}
}

func TestFormFieldConstructorKeyWinsOverStateSnapshot(t *testing.T) {
	field := FormField("email", nil,
		FormFieldState(FieldState{Key: "other", Label: "Email"}),
		FormFieldPending(true),
	)
	formField, ok := field.(*formFieldWidget)
	if !ok {
		t.Fatalf("FormField type = %T, want *formFieldWidget", field)
	}
	if got := formField.config.state.Key; got != "email" {
		t.Fatalf("FormField key = %q, want email", got)
	}
	if !formField.config.state.Pending {
		t.Fatal("FormFieldPending did not set the host snapshot")
	}
}

func TestValidationSummaryFiltersOnlyInvalidMessages(t *testing.T) {
	fields := validationSummaryInvalidFields([]FieldState{
		{Key: "valid", ErrorText: "ignored", Status: FieldValid},
		{Key: "empty", Status: FieldInvalid},
		{
			Key:         "email",
			ErrorText:   "invalid email",
			PendingText: "checking",
			Pending:     true,
			Status:      FieldInvalid,
		},
	})
	if len(fields) != 1 || fields[0].Key != "email" {
		t.Fatalf("invalid fields = %#v, want only email", fields)
	}
	if got := fields[0].Message(); got != "invalid email" {
		t.Fatalf("summary message = %q, want error text", got)
	}
}
