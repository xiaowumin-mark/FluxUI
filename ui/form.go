package ui

import widget "github.com/xiaowumin-mark/FluxUI/widget"

// FormOption configures a Form.
type FormOption = widget.FormOption

// FormRef queues imperative Form commands.
type FormRef = widget.FormRef

// FormSubmitEvent describes a Form submit intention.
type FormSubmitEvent = widget.FormSubmitEvent

// FieldStatus identifies host-supplied field validation progress.
type FieldStatus = widget.FieldStatus

const (
	FieldValid         = widget.FieldValid
	FieldInvalid       = widget.FieldInvalid
	FieldPending       = widget.FieldPending
	FieldStatusValid   = widget.FieldStatusValid
	FieldStatusInvalid = widget.FieldStatusInvalid
	FieldStatusPending = widget.FieldStatusPending
)

// FieldState is the host-controlled presentation snapshot for a form field.
type FieldState = widget.FieldState

// FormFieldOption configures a FormField.
type FormFieldOption = widget.FormFieldOption

// FormFieldRef queues imperative FormField commands.
type FormFieldRef = widget.FormFieldRef

// ValidationSummaryOption configures a ValidationSummary.
type ValidationSummaryOption = widget.ValidationSummaryOption

// Form groups fields and forwards submit intentions to its host callback.
func Form(child Widget, opts ...FormOption) Widget {
	return widget.Form(child, opts...)
}

// FormElement creates a reconciler Element for Form.
func FormElement(child Element, opts ...FormOption) Element {
	return &singleChildElement{kind: "form", child: child, renderFn: func(child Widget) Widget {
		return widget.Form(child, opts...)
	}}
}

// FormOnSubmit observes a synchronous submit intention. The host owns
// validation, pending state, and business submission work.
func FormOnSubmit(fn func(ctx *Context, event *FormSubmitEvent)) FormOption {
	return widget.FormOnSubmit(fn)
}

// FormDisabled blocks Form submit intentions.
func FormDisabled(disabled bool) FormOption {
	return widget.FormDisabled(disabled)
}

// FormPending blocks repeat Form submit intentions while host work is pending.
func FormPending(pending bool) FormOption {
	return widget.FormPending(pending)
}

// NewFormRef creates an unbound FormRef.
func NewFormRef() *FormRef {
	return widget.NewFormRef()
}

// FormAttachRef binds a FormRef for the current frame.
func FormAttachRef(ref *FormRef) FormOption {
	return widget.FormAttachRef(ref)
}

// FormField adds field presentation semantics around child.
func FormField(key string, child Widget, opts ...FormFieldOption) Widget {
	return widget.FormField(key, child, opts...)
}

// FormFieldElement creates a reconciler Element for FormField.
func FormFieldElement(key string, child Element, opts ...FormFieldOption) Element {
	return Key(key, &singleChildElement{kind: "form-field", child: child, renderFn: func(child Widget) Widget {
		return widget.FormField(key, child, opts...)
	}})
}

// FormFieldLabel sets a field label.
func FormFieldLabel(label string) FormFieldOption {
	return widget.FormFieldLabel(label)
}

// FormFieldSupportingText sets normal supporting text.
func FormFieldSupportingText(text string) FormFieldOption {
	return widget.FormFieldSupportingText(text)
}

// FormFieldErrorText sets error text used by an invalid field.
func FormFieldErrorText(text string) FormFieldOption {
	return widget.FormFieldErrorText(text)
}

// FormFieldPendingText sets pending text used by a pending field.
func FormFieldPendingText(text string) FormFieldOption {
	return widget.FormFieldPendingText(text)
}

// FormFieldRequired controls required presentation.
func FormFieldRequired(required bool) FormFieldOption {
	return widget.FormFieldRequired(required)
}

// FormFieldDisabled controls disabled presentation and FormFieldRef focus.
func FormFieldDisabled(disabled bool) FormFieldOption {
	return widget.FormFieldDisabled(disabled)
}

// FormFieldReadOnly controls read-only presentation.
func FormFieldReadOnly(readOnly bool) FormFieldOption {
	return widget.FormFieldReadOnly(readOnly)
}

// FormFieldPending controls pending presentation independently from invalid
// status, allowing hosts to show both error and progress text.
func FormFieldPending(pending bool) FormFieldOption {
	return widget.FormFieldPending(pending)
}

// FormFieldState supplies a complete host-controlled field snapshot.
func FormFieldState(state FieldState) FormFieldOption {
	return widget.FormFieldState(state)
}

// NewFormFieldRef creates an unbound FormFieldRef.
func NewFormFieldRef() *FormFieldRef {
	return widget.NewFormFieldRef()
}

// FormFieldAttachRef binds a FormFieldRef for the current frame.
func FormFieldAttachRef(ref *FormFieldRef) FormFieldOption {
	return widget.FormFieldAttachRef(ref)
}

// ValidationSummary presents invalid fields and lets the host route focus by
// stable field key.
func ValidationSummary(fields []FieldState, opts ...ValidationSummaryOption) Widget {
	return widget.ValidationSummary(fields, opts...)
}

// ValidationSummaryElement creates a reconciler Element for ValidationSummary.
func ValidationSummaryElement(fields []FieldState, opts ...ValidationSummaryOption) Element {
	return FromWidget(widget.ValidationSummary(fields, opts...))
}

// ValidationSummaryOnFocus handles activation of an invalid summary entry.
func ValidationSummaryOnFocus(fn func(ctx *Context, key string)) ValidationSummaryOption {
	return widget.ValidationSummaryOnFocus(fn)
}

// ValidationSummaryTitle sets the invalid-field title.
func ValidationSummaryTitle(title string) ValidationSummaryOption {
	return widget.ValidationSummaryTitle(title)
}

// ValidationSummaryEmptyText sets the no-error text.
func ValidationSummaryEmptyText(text string) ValidationSummaryOption {
	return widget.ValidationSummaryEmptyText(text)
}

// ValidationSummaryDisabled disables summary-entry activation.
func ValidationSummaryDisabled(disabled bool) ValidationSummaryOption {
	return widget.ValidationSummaryDisabled(disabled)
}
