package ui

import (
	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	widget "github.com/xiaowumin-mark/FluxUI/widget"
)

// ChoiceItem is a stable, host-owned advanced-picker option.
type ChoiceItem[T comparable] = widget.ChoiceItem[T]
type ComboboxItem[T comparable] = widget.ComboboxItem[T]
type AutocompleteItem[T comparable] = widget.AutocompleteItem[T]
type SearchSelectItem[T comparable] = widget.SearchSelectItem[T]

type ComboboxOption[T comparable] = widget.ComboboxOption[T]
type AutocompleteOption[T comparable] = widget.AutocompleteOption[T]
type SearchSelectOption[T comparable] = widget.SearchSelectOption[T]

type ComboboxRef[T comparable] = widget.ComboboxRef[T]
type AutocompleteRef[T comparable] = widget.AutocompleteRef[T]
type SearchSelectRef[T comparable] = widget.SearchSelectRef[T]

func Combobox[T comparable](query string, items []ComboboxItem[T], opts ...ComboboxOption[T]) Widget {
	return widget.Combobox(query, items, opts...)
}

func ComboboxElement[T comparable](query string, items []ComboboxItem[T], opts ...ComboboxOption[T]) Element {
	return FromWidget(widget.Combobox(query, items, opts...))
}

func ComboboxOpened[T comparable](opened bool) ComboboxOption[T] {
	return widget.ComboboxOpened[T](opened)
}
func ComboboxSelectedKey[T comparable](key string) ComboboxOption[T] {
	return widget.ComboboxSelectedKey[T](key)
}
func ComboboxPending[T comparable](pending bool) ComboboxOption[T] {
	return widget.ComboboxPending[T](pending)
}
func ComboboxErrorText[T comparable](text string) ComboboxOption[T] {
	return widget.ComboboxErrorText[T](text)
}
func ComboboxDisabled[T comparable](disabled bool) ComboboxOption[T] {
	return widget.ComboboxDisabled[T](disabled)
}
func ComboboxAllowCustomValue[T comparable](allow bool) ComboboxOption[T] {
	return widget.ComboboxAllowCustomValue[T](allow)
}
func ComboboxFilterOptions[T comparable](filter bool) ComboboxOption[T] {
	return widget.ComboboxFilterOptions[T](filter)
}
func ComboboxPlaceholder[T comparable](text string) ComboboxOption[T] {
	return widget.ComboboxPlaceholder[T](text)
}
func ComboboxLabel[T comparable](text string) ComboboxOption[T] {
	return widget.ComboboxLabel[T](text)
}
func ComboboxSupportingText[T comparable](text string) ComboboxOption[T] {
	return widget.ComboboxSupportingText[T](text)
}
func ComboboxRequired[T comparable](required bool) ComboboxOption[T] {
	return widget.ComboboxRequired[T](required)
}
func ComboboxMaxHeight[T comparable](height float32) ComboboxOption[T] {
	return widget.ComboboxMaxHeight[T](height)
}
func ComboboxInputOptions[T comparable](opts ...InputOption) ComboboxOption[T] {
	return widget.ComboboxInputOptions[T](opts...)
}
func ComboboxOnQueryChange[T comparable](fn func(ctx *Context, query string)) ComboboxOption[T] {
	return widget.ComboboxOnQueryChange[T](fn)
}
func ComboboxOnOpenChange[T comparable](fn func(ctx *Context, opened bool)) ComboboxOption[T] {
	return widget.ComboboxOnOpenChange[T](fn)
}
func ComboboxOnSelect[T comparable](fn func(ctx *Context, item ComboboxItem[T])) ComboboxOption[T] {
	return widget.ComboboxOnSelect[T](fn)
}
func ComboboxOnCustomValue[T comparable](fn func(ctx *Context, value string)) ComboboxOption[T] {
	return widget.ComboboxOnCustomValue[T](fn)
}
func ComboboxOnActiveChange[T comparable](fn func(ctx *Context, key string)) ComboboxOption[T] {
	return widget.ComboboxOnActiveChange[T](fn)
}
func ComboboxOnBeforeInput[T comparable](fn fluxevent.InputHandler) ComboboxOption[T] {
	return widget.ComboboxOnBeforeInput[T](fn)
}
func ComboboxOnInputEvent[T comparable](fn fluxevent.InputHandler) ComboboxOption[T] {
	return widget.ComboboxOnInputEvent[T](fn)
}
func ComboboxOnSubmit[T comparable](fn fluxevent.InputHandler) ComboboxOption[T] {
	return widget.ComboboxOnSubmit[T](fn)
}
func NewComboboxRef[T comparable]() *ComboboxRef[T] { return widget.NewComboboxRef[T]() }
func ComboboxAttachRef[T comparable](ref *ComboboxRef[T]) ComboboxOption[T] {
	return widget.ComboboxAttachRef[T](ref)
}

func Autocomplete[T comparable](query string, items []AutocompleteItem[T], opts ...AutocompleteOption[T]) Widget {
	return widget.Autocomplete(query, items, opts...)
}

func AutocompleteElement[T comparable](query string, items []AutocompleteItem[T], opts ...AutocompleteOption[T]) Element {
	return FromWidget(widget.Autocomplete(query, items, opts...))
}

func AutocompleteOpened[T comparable](opened bool) AutocompleteOption[T] {
	return widget.AutocompleteOpened[T](opened)
}
func AutocompleteSelectedKey[T comparable](key string) AutocompleteOption[T] {
	return widget.AutocompleteSelectedKey[T](key)
}
func AutocompletePending[T comparable](pending bool) AutocompleteOption[T] {
	return widget.AutocompletePending[T](pending)
}
func AutocompleteErrorText[T comparable](text string) AutocompleteOption[T] {
	return widget.AutocompleteErrorText[T](text)
}
func AutocompleteDisabled[T comparable](disabled bool) AutocompleteOption[T] {
	return widget.AutocompleteDisabled[T](disabled)
}
func AutocompleteFilterOptions[T comparable](filter bool) AutocompleteOption[T] {
	return widget.AutocompleteFilterOptions[T](filter)
}
func AutocompletePlaceholder[T comparable](text string) AutocompleteOption[T] {
	return widget.AutocompletePlaceholder[T](text)
}
func AutocompleteLabel[T comparable](text string) AutocompleteOption[T] {
	return widget.AutocompleteLabel[T](text)
}
func AutocompleteSupportingText[T comparable](text string) AutocompleteOption[T] {
	return widget.AutocompleteSupportingText[T](text)
}
func AutocompleteRequired[T comparable](required bool) AutocompleteOption[T] {
	return widget.AutocompleteRequired[T](required)
}
func AutocompleteMaxHeight[T comparable](height float32) AutocompleteOption[T] {
	return widget.AutocompleteMaxHeight[T](height)
}
func AutocompleteInputOptions[T comparable](opts ...InputOption) AutocompleteOption[T] {
	return widget.AutocompleteInputOptions[T](opts...)
}
func AutocompleteOnQueryChange[T comparable](fn func(ctx *Context, query string)) AutocompleteOption[T] {
	return widget.AutocompleteOnQueryChange[T](fn)
}
func AutocompleteOnOpenChange[T comparable](fn func(ctx *Context, opened bool)) AutocompleteOption[T] {
	return widget.AutocompleteOnOpenChange[T](fn)
}
func AutocompleteOnSelect[T comparable](fn func(ctx *Context, item AutocompleteItem[T])) AutocompleteOption[T] {
	return widget.AutocompleteOnSelect[T](fn)
}
func AutocompleteOnActiveChange[T comparable](fn func(ctx *Context, key string)) AutocompleteOption[T] {
	return widget.AutocompleteOnActiveChange[T](fn)
}
func AutocompleteOnBeforeInput[T comparable](fn fluxevent.InputHandler) AutocompleteOption[T] {
	return widget.AutocompleteOnBeforeInput[T](fn)
}
func AutocompleteOnInputEvent[T comparable](fn fluxevent.InputHandler) AutocompleteOption[T] {
	return widget.AutocompleteOnInputEvent[T](fn)
}
func AutocompleteOnSubmit[T comparable](fn fluxevent.InputHandler) AutocompleteOption[T] {
	return widget.AutocompleteOnSubmit[T](fn)
}
func NewAutocompleteRef[T comparable]() *AutocompleteRef[T] { return widget.NewAutocompleteRef[T]() }
func AutocompleteAttachRef[T comparable](ref *AutocompleteRef[T]) AutocompleteOption[T] {
	return widget.AutocompleteAttachRef[T](ref)
}

func SearchSelect[T comparable](value T, query string, items []SearchSelectItem[T], opts ...SearchSelectOption[T]) Widget {
	return widget.SearchSelect(value, query, items, opts...)
}

func SearchableSelect[T comparable](value T, query string, items []SearchSelectItem[T], opts ...SearchSelectOption[T]) Widget {
	return widget.SearchableSelect(value, query, items, opts...)
}

func SearchSelectElement[T comparable](value T, query string, items []SearchSelectItem[T], opts ...SearchSelectOption[T]) Element {
	return FromWidget(widget.SearchSelect(value, query, items, opts...))
}

func SearchableSelectElement[T comparable](value T, query string, items []SearchSelectItem[T], opts ...SearchSelectOption[T]) Element {
	return SearchSelectElement(value, query, items, opts...)
}

func SearchSelectOpened[T comparable](opened bool) SearchSelectOption[T] {
	return widget.SearchSelectOpened[T](opened)
}
func SearchSelectSelectedKey[T comparable](key string) SearchSelectOption[T] {
	return widget.SearchSelectSelectedKey[T](key)
}
func SearchableSelectSelectedKey[T comparable](key string) SearchSelectOption[T] {
	return widget.SearchableSelectSelectedKey[T](key)
}
func SearchSelectPending[T comparable](pending bool) SearchSelectOption[T] {
	return widget.SearchSelectPending[T](pending)
}
func SearchSelectErrorText[T comparable](text string) SearchSelectOption[T] {
	return widget.SearchSelectErrorText[T](text)
}
func SearchSelectDisabled[T comparable](disabled bool) SearchSelectOption[T] {
	return widget.SearchSelectDisabled[T](disabled)
}
func SearchSelectFilterOptions[T comparable](filter bool) SearchSelectOption[T] {
	return widget.SearchSelectFilterOptions[T](filter)
}
func SearchSelectPlaceholder[T comparable](text string) SearchSelectOption[T] {
	return widget.SearchSelectPlaceholder[T](text)
}
func SearchSelectLabel[T comparable](text string) SearchSelectOption[T] {
	return widget.SearchSelectLabel[T](text)
}
func SearchSelectSupportingText[T comparable](text string) SearchSelectOption[T] {
	return widget.SearchSelectSupportingText[T](text)
}
func SearchSelectRequired[T comparable](required bool) SearchSelectOption[T] {
	return widget.SearchSelectRequired[T](required)
}
func SearchSelectMaxHeight[T comparable](height float32) SearchSelectOption[T] {
	return widget.SearchSelectMaxHeight[T](height)
}
func SearchSelectInputOptions[T comparable](opts ...InputOption) SearchSelectOption[T] {
	return widget.SearchSelectInputOptions[T](opts...)
}
func SearchSelectOnQueryChange[T comparable](fn func(ctx *Context, query string)) SearchSelectOption[T] {
	return widget.SearchSelectOnQueryChange[T](fn)
}
func SearchSelectOnOpenChange[T comparable](fn func(ctx *Context, opened bool)) SearchSelectOption[T] {
	return widget.SearchSelectOnOpenChange[T](fn)
}
func SearchSelectOnChange[T comparable](fn func(ctx *Context, value T)) SearchSelectOption[T] {
	return widget.SearchSelectOnChange[T](fn)
}
func SearchSelectOnSelect[T comparable](fn func(ctx *Context, item SearchSelectItem[T])) SearchSelectOption[T] {
	return widget.SearchSelectOnSelect[T](fn)
}
func SearchSelectOnActiveChange[T comparable](fn func(ctx *Context, key string)) SearchSelectOption[T] {
	return widget.SearchSelectOnActiveChange[T](fn)
}
func SearchSelectOnBeforeInput[T comparable](fn fluxevent.InputHandler) SearchSelectOption[T] {
	return widget.SearchSelectOnBeforeInput[T](fn)
}
func SearchSelectOnInputEvent[T comparable](fn fluxevent.InputHandler) SearchSelectOption[T] {
	return widget.SearchSelectOnInputEvent[T](fn)
}
func SearchSelectOnSubmit[T comparable](fn fluxevent.InputHandler) SearchSelectOption[T] {
	return widget.SearchSelectOnSubmit[T](fn)
}
func NewSearchSelectRef[T comparable]() *SearchSelectRef[T] { return widget.NewSearchSelectRef[T]() }
func SearchSelectAttachRef[T comparable](ref *SearchSelectRef[T]) SearchSelectOption[T] {
	return widget.SearchSelectAttachRef[T](ref)
}
