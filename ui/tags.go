package ui

import (
	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	widget "github.com/xiaowumin-mark/FluxUI/widget"
)

// MultiSelectOption configures a controlled MultiSelect.
type MultiSelectOption[T comparable] = widget.MultiSelectOption[T]

// MultiSelectRef queues controlled MultiSelect commands for the next layout.
type MultiSelectRef[T comparable] = widget.MultiSelectRef[T]

// MultiSelect renders a searchable controlled multi-picker.
func MultiSelect[T comparable](selected []T, options []ChoiceItem[T], opts ...MultiSelectOption[T]) Widget {
	return widget.MultiSelect(selected, options, opts...)
}

// MultiSelectElement creates a reconciler-compatible controlled multi-picker.
func MultiSelectElement[T comparable](selected []T, options []ChoiceItem[T], opts ...MultiSelectOption[T]) Element {
	return FromWidget(widget.MultiSelect(selected, options, opts...))
}

func MultiSelectQuery[T comparable](query string) MultiSelectOption[T] {
	return widget.MultiSelectQuery[T](query)
}
func MultiSelectOpened[T comparable](opened bool) MultiSelectOption[T] {
	return widget.MultiSelectOpened[T](opened)
}
func MultiSelectPending[T comparable](pending bool) MultiSelectOption[T] {
	return widget.MultiSelectPending[T](pending)
}
func MultiSelectError[T comparable](invalid bool) MultiSelectOption[T] {
	return widget.MultiSelectError[T](invalid)
}
func MultiSelectErrorText[T comparable](text string) MultiSelectOption[T] {
	return widget.MultiSelectErrorText[T](text)
}
func MultiSelectDisabled[T comparable](disabled bool) MultiSelectOption[T] {
	return widget.MultiSelectDisabled[T](disabled)
}
func MultiSelectFilterOptions[T comparable](filter bool) MultiSelectOption[T] {
	return widget.MultiSelectFilterOptions[T](filter)
}
func MultiSelectPlaceholder[T comparable](text string) MultiSelectOption[T] {
	return widget.MultiSelectPlaceholder[T](text)
}
func MultiSelectLabel[T comparable](text string) MultiSelectOption[T] {
	return widget.MultiSelectLabel[T](text)
}
func MultiSelectSupportingText[T comparable](text string) MultiSelectOption[T] {
	return widget.MultiSelectSupportingText[T](text)
}
func MultiSelectRequired[T comparable](required bool) MultiSelectOption[T] {
	return widget.MultiSelectRequired[T](required)
}
func MultiSelectMaxHeight[T comparable](height float32) MultiSelectOption[T] {
	return widget.MultiSelectMaxHeight[T](height)
}
func MultiSelectInputOptions[T comparable](opts ...InputOption) MultiSelectOption[T] {
	return widget.MultiSelectInputOptions[T](opts...)
}
func MultiSelectSelectedKeys[T comparable](keys []string) MultiSelectOption[T] {
	return widget.MultiSelectSelectedKeys[T](keys)
}
func MultiSelectOnChange[T comparable](fn func(ctx *Context, selected []T)) MultiSelectOption[T] {
	return widget.MultiSelectOnChange[T](fn)
}
func MultiSelectOnSelectedKeysChange[T comparable](fn func(ctx *Context, selectedKeys []string)) MultiSelectOption[T] {
	return widget.MultiSelectOnSelectedKeysChange[T](fn)
}
func MultiSelectOnToggle[T comparable](fn func(ctx *Context, item ChoiceItem[T], selected bool)) MultiSelectOption[T] {
	return widget.MultiSelectOnToggle[T](fn)
}
func MultiSelectOnSelect[T comparable](fn func(ctx *Context, item ChoiceItem[T])) MultiSelectOption[T] {
	return widget.MultiSelectOnSelect[T](fn)
}
func MultiSelectOnRemove[T comparable](fn func(ctx *Context, item ChoiceItem[T])) MultiSelectOption[T] {
	return widget.MultiSelectOnRemove[T](fn)
}
func MultiSelectOnQueryChange[T comparable](fn func(ctx *Context, query string)) MultiSelectOption[T] {
	return widget.MultiSelectOnQueryChange[T](fn)
}
func MultiSelectOnOpenChange[T comparable](fn func(ctx *Context, opened bool)) MultiSelectOption[T] {
	return widget.MultiSelectOnOpenChange[T](fn)
}
func MultiSelectOnActiveChange[T comparable](fn func(ctx *Context, key string)) MultiSelectOption[T] {
	return widget.MultiSelectOnActiveChange[T](fn)
}
func MultiSelectOnBeforeInput[T comparable](fn fluxevent.InputHandler) MultiSelectOption[T] {
	return widget.MultiSelectOnBeforeInput[T](fn)
}
func MultiSelectOnInputEvent[T comparable](fn fluxevent.InputHandler) MultiSelectOption[T] {
	return widget.MultiSelectOnInputEvent[T](fn)
}
func MultiSelectOnSubmit[T comparable](fn fluxevent.InputHandler) MultiSelectOption[T] {
	return widget.MultiSelectOnSubmit[T](fn)
}
func NewMultiSelectRef[T comparable]() *MultiSelectRef[T] { return widget.NewMultiSelectRef[T]() }
func MultiSelectAttachRef[T comparable](ref *MultiSelectRef[T]) MultiSelectOption[T] {
	return widget.MultiSelectAttachRef[T](ref)
}

// TagOptionItem is one stable TagPicker option.
type TagOptionItem = widget.TagOptionItem

// TagPickerOption configures a controlled TagPicker.
type TagPickerOption = widget.TagPickerOption

// TagPickerRef queues controlled TagPicker commands for the next layout.
type TagPickerRef = widget.TagPickerRef

// TagPicker renders a searchable controlled picker for string tag keys.
func TagPicker(selectedKeys []string, query string, options []TagOptionItem, opts ...TagPickerOption) Widget {
	return widget.TagPicker(selectedKeys, query, options, opts...)
}

// TagPickerElement creates a reconciler-compatible TagPicker.
func TagPickerElement(selectedKeys []string, query string, options []TagOptionItem, opts ...TagPickerOption) Element {
	return FromWidget(widget.TagPicker(selectedKeys, query, options, opts...))
}

func TagPickerOpened(opened bool) TagPickerOption     { return widget.TagPickerOpened(opened) }
func TagPickerPending(pending bool) TagPickerOption   { return widget.TagPickerPending(pending) }
func TagPickerError(invalid bool) TagPickerOption     { return widget.TagPickerError(invalid) }
func TagPickerErrorText(text string) TagPickerOption  { return widget.TagPickerErrorText(text) }
func TagPickerDisabled(disabled bool) TagPickerOption { return widget.TagPickerDisabled(disabled) }
func TagPickerFilterOptions(filter bool) TagPickerOption {
	return widget.TagPickerFilterOptions(filter)
}
func TagPickerPlaceholder(text string) TagPickerOption { return widget.TagPickerPlaceholder(text) }
func TagPickerLabel(text string) TagPickerOption       { return widget.TagPickerLabel(text) }
func TagPickerSupportingText(text string) TagPickerOption {
	return widget.TagPickerSupportingText(text)
}
func TagPickerRequired(required bool) TagPickerOption   { return widget.TagPickerRequired(required) }
func TagPickerMaxHeight(height float32) TagPickerOption { return widget.TagPickerMaxHeight(height) }
func TagPickerInputOptions(opts ...InputOption) TagPickerOption {
	return widget.TagPickerInputOptions(opts...)
}
func TagPickerOnChange(fn func(ctx *Context, selectedKeys []string)) TagPickerOption {
	return widget.TagPickerOnChange(fn)
}
func TagPickerOnSelectedKeysChange(fn func(ctx *Context, selectedKeys []string)) TagPickerOption {
	return widget.TagPickerOnSelectedKeysChange(fn)
}
func TagPickerOnToggle(fn func(ctx *Context, item TagOptionItem, selected bool)) TagPickerOption {
	return widget.TagPickerOnToggle(fn)
}
func TagPickerOnSelect(fn func(ctx *Context, item TagOptionItem)) TagPickerOption {
	return widget.TagPickerOnSelect(fn)
}
func TagPickerOnRemove(fn func(ctx *Context, key string)) TagPickerOption {
	return widget.TagPickerOnRemove(fn)
}
func TagPickerOnQueryChange(fn func(ctx *Context, query string)) TagPickerOption {
	return widget.TagPickerOnQueryChange(fn)
}
func TagPickerOnOpenChange(fn func(ctx *Context, opened bool)) TagPickerOption {
	return widget.TagPickerOnOpenChange(fn)
}
func TagPickerOnActiveChange(fn func(ctx *Context, key string)) TagPickerOption {
	return widget.TagPickerOnActiveChange(fn)
}
func TagPickerOnBeforeInput(fn fluxevent.InputHandler) TagPickerOption {
	return widget.TagPickerOnBeforeInput(fn)
}
func TagPickerOnInputEvent(fn fluxevent.InputHandler) TagPickerOption {
	return widget.TagPickerOnInputEvent(fn)
}
func TagPickerOnSubmit(fn fluxevent.InputHandler) TagPickerOption {
	return widget.TagPickerOnSubmit(fn)
}
func NewTagPickerRef() *TagPickerRef { return widget.NewTagPickerRef() }
func TagPickerAttachRef(ref *TagPickerRef) TagPickerOption {
	return widget.TagPickerAttachRef(ref)
}

// TagInputOption configures a controlled free-form TagInput.
type TagInputOption = widget.TagInputOption

// TagInputRef queues controlled TagInput commands for the next layout.
type TagInputRef = widget.TagInputRef

// TagInput renders controlled free-form tags and a controlled query.
func TagInput(tags []string, query string, opts ...TagInputOption) Widget {
	return widget.TagInput(tags, query, opts...)
}

// TagInputElement creates a reconciler-compatible TagInput.
func TagInputElement(tags []string, query string, opts ...TagInputOption) Element {
	return FromWidget(widget.TagInput(tags, query, opts...))
}

func TagInputOpened(opened bool) TagInputOption         { return widget.TagInputOpened(opened) }
func TagInputPending(pending bool) TagInputOption       { return widget.TagInputPending(pending) }
func TagInputError(invalid bool) TagInputOption         { return widget.TagInputError(invalid) }
func TagInputErrorText(text string) TagInputOption      { return widget.TagInputErrorText(text) }
func TagInputDisabled(disabled bool) TagInputOption     { return widget.TagInputDisabled(disabled) }
func TagInputPlaceholder(text string) TagInputOption    { return widget.TagInputPlaceholder(text) }
func TagInputLabel(text string) TagInputOption          { return widget.TagInputLabel(text) }
func TagInputSupportingText(text string) TagInputOption { return widget.TagInputSupportingText(text) }
func TagInputRequired(required bool) TagInputOption     { return widget.TagInputRequired(required) }
func TagInputMaxHeight(height float32) TagInputOption   { return widget.TagInputMaxHeight(height) }
func TagInputInputOptions(opts ...InputOption) TagInputOption {
	return widget.TagInputInputOptions(opts...)
}
func TagInputOnChange(fn func(ctx *Context, tags []string)) TagInputOption {
	return widget.TagInputOnChange(fn)
}
func TagInputOnAdd(fn func(ctx *Context, tag string)) TagInputOption {
	return widget.TagInputOnAdd(fn)
}
func TagInputOnTagAdd(fn func(ctx *Context, tag string)) TagInputOption {
	return widget.TagInputOnTagAdd(fn)
}
func TagInputOnRemove(fn func(ctx *Context, tag string)) TagInputOption {
	return widget.TagInputOnRemove(fn)
}
func TagInputOnTagRemove(fn func(ctx *Context, tag string)) TagInputOption {
	return widget.TagInputOnTagRemove(fn)
}
func TagInputOnQueryChange(fn func(ctx *Context, query string)) TagInputOption {
	return widget.TagInputOnQueryChange(fn)
}
func TagInputOnOpenChange(fn func(ctx *Context, opened bool)) TagInputOption {
	return widget.TagInputOnOpenChange(fn)
}
func TagInputOnActiveChange(fn func(ctx *Context, key string)) TagInputOption {
	return widget.TagInputOnActiveChange(fn)
}
func TagInputOnBeforeInput(fn fluxevent.InputHandler) TagInputOption {
	return widget.TagInputOnBeforeInput(fn)
}
func TagInputOnInputEvent(fn fluxevent.InputHandler) TagInputOption {
	return widget.TagInputOnInputEvent(fn)
}
func TagInputOnSubmit(fn fluxevent.InputHandler) TagInputOption {
	return widget.TagInputOnSubmit(fn)
}
func NewTagInputRef() *TagInputRef { return widget.NewTagInputRef() }
func TagInputAttachRef(ref *TagInputRef) TagInputOption {
	return widget.TagInputAttachRef(ref)
}
