package ui

import "github.com/xiaowumin-mark/FluxUI/widget"

// NumericValue is the exact parse result reported by NumericField.
type NumericValue = widget.NumericValue

// NumericFieldOption configures NumericField and SpinBox.
type NumericFieldOption = widget.NumericFieldOption

// NumericFieldRef provides programmatic commands for NumericField and SpinBox.
type NumericFieldRef = widget.NumericFieldRef

// NumericField creates a controlled exact-decimal text field.
func NumericField(value string, opts ...NumericFieldOption) Widget {
	return widget.NumericField(value, opts...)
}

// SpinBox creates a NumericField with exact increment and decrement controls.
func SpinBox(value string, opts ...NumericFieldOption) Widget {
	return widget.SpinBox(value, opts...)
}

// NumericFieldElement creates a reconciler Element for NumericField.
func NumericFieldElement(value string, opts ...NumericFieldOption) Element {
	return FromWidget(widget.NumericField(value, opts...))
}

// SpinBoxElement creates a reconciler Element for SpinBox.
func SpinBoxElement(value string, opts ...NumericFieldOption) Element {
	return FromWidget(widget.SpinBox(value, opts...))
}

func NumericFieldMin(value string) NumericFieldOption {
	return widget.NumericFieldMin(value)
}

func NumericFieldMax(value string) NumericFieldOption {
	return widget.NumericFieldMax(value)
}

func NumericFieldStep(value string) NumericFieldOption {
	return widget.NumericFieldStep(value)
}

func NumericFieldInteger(integer bool) NumericFieldOption {
	return widget.NumericFieldInteger(integer)
}

func NumericFieldDisabled(disabled bool) NumericFieldOption {
	return widget.NumericFieldDisabled(disabled)
}

func NumericFieldPlaceholder(text string) NumericFieldOption {
	return widget.NumericFieldPlaceholder(text)
}

func NumericFieldLabel(text string) NumericFieldOption {
	return widget.NumericFieldLabel(text)
}

func NumericFieldSupportingText(text string) NumericFieldOption {
	return widget.NumericFieldSupportingText(text)
}

func NumericFieldErrorText(text string) NumericFieldOption {
	return widget.NumericFieldErrorText(text)
}

func NumericFieldError(err bool) NumericFieldOption {
	return widget.NumericFieldError(err)
}

func NumericFieldRequired(required bool) NumericFieldOption {
	return widget.NumericFieldRequired(required)
}

func NumericFieldNoAsterisk(noAsterisk bool) NumericFieldOption {
	return widget.NumericFieldNoAsterisk(noAsterisk)
}

func NumericFieldPending(pending bool) NumericFieldOption {
	return widget.NumericFieldPending(pending)
}

func NumericFieldInputOptions(opts ...InputOption) NumericFieldOption {
	return widget.NumericFieldInputOptions(opts...)
}

func NumericFieldOnChange(fn func(ctx *Context, value string)) NumericFieldOption {
	return widget.NumericFieldOnChange(fn)
}

func NumericFieldOnParsedChange(fn func(ctx *Context, value NumericValue)) NumericFieldOption {
	return widget.NumericFieldOnParsedChange(fn)
}

func NewNumericFieldRef() *NumericFieldRef {
	return widget.NewNumericFieldRef()
}

func NumericFieldAttachRef(ref *NumericFieldRef) NumericFieldOption {
	return widget.NumericFieldAttachRef(ref)
}

// ParseNumericValue uses NumericField's exact-decimal parser without bounds.
func ParseNumericValue(text string) NumericValue {
	return widget.ParseNumericValue(text)
}
