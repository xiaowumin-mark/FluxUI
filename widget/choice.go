package widget

import (
	"fmt"
	"strings"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/internal/collection"
	"github.com/xiaowumin-mark/FluxUI/layout"

	"gioui.org/io/key"
)

// ChoiceItem is one stable, host-owned suggestion or picker option.
//
// Key must be non-empty and remain stable while the logical option exists. In
// particular, do not derive it from a filtered or sorted index. Value is the
// host value that is returned by selection callbacks; Key owns focus and
// selection identity.
type ChoiceItem[T comparable] struct {
	Key           string
	Label         string
	Value         T
	Disabled      bool
	Leading       Widget
	Trailing      Widget
	TypeaheadText string
}

// ComboboxItem is an alias kept to make Combobox call sites self-documenting.
type ComboboxItem[T comparable] = ChoiceItem[T]

// AutocompleteItem is an alias kept to make Autocomplete call sites
// self-documenting.
type AutocompleteItem[T comparable] = ChoiceItem[T]

// SearchSelectItem is an alias kept to make searchable Select call sites
// self-documenting.
type SearchSelectItem[T comparable] = ChoiceItem[T]

type suggestionConfig[T comparable] struct {
	opened         bool
	selectedKey    string
	pending        bool
	errorText      string
	disabled       bool
	allowCustom    bool
	filter         bool
	placeholder    string
	label          string
	supportingText string
	required       bool
	maxHeight      float32
	inputOptions   []InputOption
	onQueryChange  func(ctx *internal.Context, query string)
	onOpenChange   func(ctx *internal.Context, opened bool)
	onSelect       func(ctx *internal.Context, item ChoiceItem[T])
	onValueChange  func(ctx *internal.Context, value T)
	onCustomValue  func(ctx *internal.Context, value string)
	onActiveChange func(ctx *internal.Context, key string)
	onBeforeInput  fluxevent.InputHandler
	onInputEvent   fluxevent.InputHandler
	onSubmit       fluxevent.InputHandler
	ref            *ComboboxRef[T]
}

func defaultSuggestionConfig[T comparable]() suggestionConfig[T] {
	return suggestionConfig[T]{
		allowCustom: true,
		filter:      true,
		placeholder: "Search…",
		maxHeight:   280,
	}
}

// ComboboxOption configures a controlled Combobox.
//
// query and opened are controlled snapshots. User edits and pointer/keyboard
// actions only call the corresponding callbacks; the component never starts a
// query or commits a business value by itself.
type ComboboxOption[T comparable] func(*suggestionConfig[T])

// Combobox renders a controlled editable picker. The host owns query, opened,
// selected key, suggestions and pending/error snapshots. When custom values
// are enabled (the default), Enter with no active suggestion calls
// ComboboxOnCustomValue; disabling it makes the control selection-only.
func Combobox[T comparable](query string, items []ComboboxItem[T], opts ...ComboboxOption[T]) Widget {
	cfg := defaultSuggestionConfig[T]()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return &suggestionWidget[T]{
		query:  query,
		items:  append([]ChoiceItem[T](nil), items...),
		config: cfg,
	}
}

func ComboboxOpened[T comparable](opened bool) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.opened = opened }
}

func ComboboxSelectedKey[T comparable](key string) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.selectedKey = key }
}

func ComboboxPending[T comparable](pending bool) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.pending = pending }
}

func ComboboxErrorText[T comparable](text string) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.errorText = text }
}

func ComboboxDisabled[T comparable](disabled bool) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.disabled = disabled }
}

func ComboboxAllowCustomValue[T comparable](allow bool) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.allowCustom = allow }
}

// ComboboxFilterOptions controls local case-insensitive label filtering. It
// does not perform a business query; hosts may disable it when items are
// already an asynchronously supplied suggestion snapshot.
func ComboboxFilterOptions[T comparable](filter bool) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.filter = filter }
}

func ComboboxPlaceholder[T comparable](text string) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.placeholder = text }
}

func ComboboxLabel[T comparable](text string) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.label = text }
}

func ComboboxSupportingText[T comparable](text string) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.supportingText = text }
}

func ComboboxRequired[T comparable](required bool) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.required = required }
}

func ComboboxMaxHeight[T comparable](height float32) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.maxHeight = height }
}

// ComboboxInputOptions appends presentation-only Input options before the
// Combobox's own event bridge. InputOnChange is intentionally owned by the
// Combobox; use the typed Combobox input callbacks below for event observation.
func ComboboxInputOptions[T comparable](opts ...InputOption) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) {
		cfg.inputOptions = append(cfg.inputOptions, opts...)
	}
}

func ComboboxOnQueryChange[T comparable](fn func(ctx *internal.Context, query string)) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onQueryChange = fn }
}

func ComboboxOnOpenChange[T comparable](fn func(ctx *internal.Context, opened bool)) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onOpenChange = fn }
}

func ComboboxOnSelect[T comparable](fn func(ctx *internal.Context, item ComboboxItem[T])) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onSelect = fn }
}

func ComboboxOnCustomValue[T comparable](fn func(ctx *internal.Context, value string)) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onCustomValue = fn }
}

func ComboboxOnActiveChange[T comparable](fn func(ctx *internal.Context, key string)) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onActiveChange = fn }
}

func ComboboxOnBeforeInput[T comparable](fn fluxevent.InputHandler) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onBeforeInput = fn }
}

func ComboboxOnInputEvent[T comparable](fn fluxevent.InputHandler) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onInputEvent = fn }
}

func ComboboxOnSubmit[T comparable](fn fluxevent.InputHandler) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onSubmit = fn }
}

func ComboboxAttachRef[T comparable](ref *ComboboxRef[T]) ComboboxOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.ref = ref }
}

// AutocompleteOption configures a controlled Autocomplete. It has the same
// query/opened/suggestion contract as Combobox, but never accepts a free value.
type AutocompleteOption[T comparable] func(*suggestionConfig[T])

// Autocomplete renders a controlled suggestion field. Selecting a suggestion
// calls AutocompleteOnSelect. Typing only reports a query intent and never
// implicitly submits a business search.
func Autocomplete[T comparable](query string, items []AutocompleteItem[T], opts ...AutocompleteOption[T]) Widget {
	cfg := defaultSuggestionConfig[T]()
	cfg.allowCustom = false
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return &suggestionWidget[T]{
		query:  query,
		items:  append([]ChoiceItem[T](nil), items...),
		config: cfg,
	}
}

func AutocompleteOpened[T comparable](opened bool) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.opened = opened }
}

func AutocompleteSelectedKey[T comparable](key string) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.selectedKey = key }
}

func AutocompletePending[T comparable](pending bool) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.pending = pending }
}

func AutocompleteErrorText[T comparable](text string) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.errorText = text }
}

func AutocompleteDisabled[T comparable](disabled bool) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.disabled = disabled }
}

func AutocompleteFilterOptions[T comparable](filter bool) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.filter = filter }
}

func AutocompletePlaceholder[T comparable](text string) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.placeholder = text }
}

func AutocompleteLabel[T comparable](text string) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.label = text }
}

func AutocompleteSupportingText[T comparable](text string) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.supportingText = text }
}

func AutocompleteRequired[T comparable](required bool) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.required = required }
}

func AutocompleteMaxHeight[T comparable](height float32) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.maxHeight = height }
}

func AutocompleteInputOptions[T comparable](opts ...InputOption) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.inputOptions = append(cfg.inputOptions, opts...) }
}

func AutocompleteOnQueryChange[T comparable](fn func(ctx *internal.Context, query string)) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onQueryChange = fn }
}

func AutocompleteOnOpenChange[T comparable](fn func(ctx *internal.Context, opened bool)) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onOpenChange = fn }
}

func AutocompleteOnSelect[T comparable](fn func(ctx *internal.Context, item AutocompleteItem[T])) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onSelect = fn }
}

func AutocompleteOnActiveChange[T comparable](fn func(ctx *internal.Context, key string)) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onActiveChange = fn }
}

func AutocompleteOnBeforeInput[T comparable](fn fluxevent.InputHandler) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onBeforeInput = fn }
}

func AutocompleteOnInputEvent[T comparable](fn fluxevent.InputHandler) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onInputEvent = fn }
}

func AutocompleteOnSubmit[T comparable](fn fluxevent.InputHandler) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onSubmit = fn }
}

func AutocompleteAttachRef[T comparable](ref *AutocompleteRef[T]) AutocompleteOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.ref = ref }
}

// SearchSelectOption configures a searchable selection-only field. Unlike the
// deprecated SelectSearchable compatibility option, SearchSelect has a real
// controlled query and result list.
type SearchSelectOption[T comparable] func(*suggestionConfig[T])

// SearchSelect creates a controlled, searchable Select. Enter and pointer
// selection can only choose an item from items; arbitrary query text is never
// committed as a value.
func SearchSelect[T comparable](value T, query string, items []SearchSelectItem[T], opts ...SearchSelectOption[T]) Widget {
	cfg := defaultSuggestionConfig[T]()
	cfg.allowCustom = false
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.selectedKey == "" {
		for _, item := range items {
			if item.Value == value {
				cfg.selectedKey = choiceItemKey(item)
				break
			}
		}
	}
	selectItem := cfg.onSelect
	valueChange := cfg.onValueChange
	cfg.onSelect = func(ctx *internal.Context, item ChoiceItem[T]) {
		if selectItem != nil {
			selectItem(ctx, item)
		}
		if valueChange != nil && item.Value != value {
			valueChange(ctx, item.Value)
		}
	}
	return &suggestionWidget[T]{
		query:  query,
		items:  append([]ChoiceItem[T](nil), items...),
		config: cfg,
	}
}

// SearchableSelect is the descriptive spelling of SearchSelect.
func SearchableSelect[T comparable](value T, query string, items []SearchSelectItem[T], opts ...SearchSelectOption[T]) Widget {
	return SearchSelect(value, query, items, opts...)
}

func SearchSelectOpened[T comparable](opened bool) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.opened = opened }
}

// SearchSelectSelectedKey supplies the controlled stable key for the selected
// option. It is necessary when distinct options share the same comparable
// Value; otherwise SearchSelect derives a key from value for convenience.
func SearchSelectSelectedKey[T comparable](key string) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.selectedKey = key }
}

func SearchableSelectSelectedKey[T comparable](key string) SearchSelectOption[T] {
	return SearchSelectSelectedKey[T](key)
}

func SearchSelectPending[T comparable](pending bool) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.pending = pending }
}

func SearchSelectErrorText[T comparable](text string) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.errorText = text }
}

func SearchSelectDisabled[T comparable](disabled bool) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.disabled = disabled }
}

func SearchSelectFilterOptions[T comparable](filter bool) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.filter = filter }
}

func SearchSelectPlaceholder[T comparable](text string) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.placeholder = text }
}

func SearchSelectLabel[T comparable](text string) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.label = text }
}

func SearchSelectSupportingText[T comparable](text string) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.supportingText = text }
}

func SearchSelectRequired[T comparable](required bool) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.required = required }
}

func SearchSelectMaxHeight[T comparable](height float32) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.maxHeight = height }
}

func SearchSelectInputOptions[T comparable](opts ...InputOption) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.inputOptions = append(cfg.inputOptions, opts...) }
}

func SearchSelectOnQueryChange[T comparable](fn func(ctx *internal.Context, query string)) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onQueryChange = fn }
}

func SearchSelectOnOpenChange[T comparable](fn func(ctx *internal.Context, opened bool)) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onOpenChange = fn }
}

func SearchSelectOnChange[T comparable](fn func(ctx *internal.Context, value T)) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onValueChange = fn }
}

func SearchSelectOnSelect[T comparable](fn func(ctx *internal.Context, item SearchSelectItem[T])) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onSelect = fn }
}

func SearchSelectOnActiveChange[T comparable](fn func(ctx *internal.Context, key string)) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onActiveChange = fn }
}

func SearchSelectOnBeforeInput[T comparable](fn fluxevent.InputHandler) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onBeforeInput = fn }
}

func SearchSelectOnInputEvent[T comparable](fn fluxevent.InputHandler) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onInputEvent = fn }
}

func SearchSelectOnSubmit[T comparable](fn fluxevent.InputHandler) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.onSubmit = fn }
}

func SearchSelectAttachRef[T comparable](ref *SearchSelectRef[T]) SearchSelectOption[T] {
	return func(cfg *suggestionConfig[T]) { cfg.ref = ref }
}

type suggestionWidget[T comparable] struct {
	query  string
	items  []ChoiceItem[T]
	config suggestionConfig[T]
}

type suggestionState struct {
	active        collection.RovingFocus
	inputRef      *InputRef
	requestedOpen *bool
}

func (s *suggestionWidget[T]) Layout(ctx *internal.Context) layout.Dimensions {
	if ctx == nil {
		return layout.Dimensions{}
	}
	state := suggestionStateFor(ctx)
	if state.inputRef == nil {
		state.inputRef = NewInputRef()
	}
	// opened is controlled: a previous-frame request is only a same-frame
	// reconciliation aid. The host snapshot wins on the next layout even when it
	// intentionally rejects or reverses the requested value.
	state.requestedOpen = nil
	if s.config.disabled {
		s.closeForDisabled(ctx, state)
	}

	// Validate the complete host snapshot before considering a filtered result.
	// A query must not hide a duplicate/blank key and accidentally turn the
	// remaining row into an index-addressed interactive fallback.
	allItems := filteredChoiceItems(s.items, "", false)
	_, optionsValid := choiceCollectionModel(allItems)
	items := filteredChoiceItems(s.items, s.query, s.config.filter)
	model, visibleModelOK := choiceCollectionModel(items)
	modelOK := optionsValid && visibleModelOK
	if modelOK {
		state.active = state.active.Reconcile(model)
	} else {
		state.active = collection.RovingFocus{}
		// Empty or duplicate keys are a host configuration error. Do not render
		// an index-addressed fallback list that could transfer active state to a
		// different option after a reorder.
		items = nil
	}

	s.drainRef(ctx, state, items)

	inputOptions := append([]InputOption(nil), s.config.inputOptions...)
	if captureKeys := s.captureKeys(state, items, modelOK); len(captureKeys) > 0 {
		inputOptions = append(inputOptions, inputCaptureKeys(captureKeys...))
	}
	errorText := s.config.errorText
	if errorText == "" && !modelOK && len(s.items) > 0 {
		errorText = "Options require unique, non-empty keys"
	}
	inputOptions = append(inputOptions,
		InputPlaceholder(s.config.placeholder),
		InputLabel(choiceRequiredLabel(s.config.label, s.config.required)),
		InputSupportingText(s.config.supportingText),
		InputError(errorText != ""),
		InputErrorText(errorText),
		InputDisabled(s.config.disabled),
		InputAttachRef(state.inputRef),
	)
	if s.config.pending {
		inputOptions = append(inputOptions, InputTrailing(Text("Loading…")))
	}
	inputOptions = append(inputOptions, InputOnBeforeInput(func(actionCtx *internal.Context, ev *fluxevent.InputEvent) {
		if s.config.onBeforeInput != nil {
			s.config.onBeforeInput(actionCtx, ev)
		}
	}))
	if s.config.onInputEvent != nil {
		inputOptions = append(inputOptions, InputOnInputEvent(s.config.onInputEvent))
	}
	inputOptions = append(inputOptions,
		InputOnFocus(func(actionCtx *internal.Context, focused bool) {
			if focused && !s.config.disabled {
				s.requestOpen(actionCtx, state, true)
			}
		}),
		InputOnChange(func(actionCtx *internal.Context, query string) {
			if s.config.disabled {
				return
			}
			if s.config.onQueryChange != nil && query != s.query {
				s.config.onQueryChange(actionCtx, query)
			}
			s.requestOpen(actionCtx, state, true)
		}),
		InputOnSubmit(func(actionCtx *internal.Context, ev *fluxevent.InputEvent) {
			if s.config.onSubmit != nil {
				s.config.onSubmit(actionCtx, ev)
			}
			if ev == nil || ev.DefaultPrevented || s.config.disabled {
				return
			}
			if s.effectiveOpen(state) && modelOK {
				active := state.active.Reconcile(model)
				state.active = active
				if item, ok := choiceItemByKey(items, string(active.Key)); ok && !item.Disabled {
					s.selectItem(actionCtx, state, item)
					s.requestOpen(actionCtx, state, false)
					ev.PreventDefault()
					return
				}
			}
			if s.config.allowCustom && s.config.onCustomValue != nil && strings.TrimSpace(ev.Value) != "" {
				ev.PreventDefault()
				s.config.onCustomValue(actionCtx, ev.Value)
				s.requestOpen(actionCtx, state, false)
			}
		}),
	)

	trigger := TextField(s.query, inputOptions...)
	menuItems := suggestionMenuItems(items, s.config.selectedKey, string(state.active.Key))
	dropdown := DropdownMenu(s.config.opened && len(menuItems) > 0, trigger, menuItems,
		DropdownMenuMaxHeight(s.config.maxHeight),
		dropdownMenuDisabled(s.config.disabled),
		DropdownMenuOnOpenChange(func(actionCtx *internal.Context, opened bool) {
			s.requestOpen(actionCtx, state, opened)
		}),
		DropdownMenuOnSelect(func(actionCtx *internal.Context, key string) {
			item, ok := choiceItemByKey(items, key)
			if !ok || item.Disabled || s.config.disabled {
				return
			}
			s.selectItem(actionCtx, state, item)
		}),
	)
	return KeyboardScope(dropdown,
		KeyboardScopeFocusable(!s.config.disabled),
		KeyOnDown(func(actionCtx *internal.Context, ev *fluxevent.KeyboardEvent) {
			s.handleKey(actionCtx, state, items, model, modelOK, ev)
		}),
	).Layout(ctx.Child(0))
}

func (s *suggestionWidget[T]) closeForDisabled(ctx *internal.Context, state *suggestionState) {
	if state == nil {
		return
	}
	wasOpen := s.effectiveOpen(state)
	if state.inputRef != nil {
		state.inputRef.Blur()
	}
	md3ClearFocusIfInside(ctx, ctx.PathID())
	if !wasOpen {
		return
	}
	closed := false
	state.requestedOpen = &closed
	if s.config.onOpenChange != nil {
		s.config.onOpenChange(ctx, false)
	}
}

func (s *suggestionWidget[T]) captureKeys(state *suggestionState, items []ChoiceItem[T], modelOK bool) []key.Name {
	if s.config.disabled || !modelOK || len(items) == 0 {
		return nil
	}
	keys := []key.Name{key.NameUpArrow, key.NameDownArrow, key.NameHome, key.NameEnd}
	if s.config.opened && state != nil {
		if item, ok := choiceItemByKey(items, string(state.active.Key)); ok && !item.Disabled {
			keys = append(keys, key.NameSpace)
		}
	}
	return keys
}

func (s *suggestionWidget[T]) requestOpen(ctx *internal.Context, state *suggestionState, opened bool) {
	if state == nil || s.config.disabled {
		return
	}
	current := s.effectiveOpen(state)
	if current == opened {
		return
	}
	if s.config.opened == opened {
		state.requestedOpen = nil
	} else {
		next := opened
		state.requestedOpen = &next
	}
	if s.config.onOpenChange != nil {
		s.config.onOpenChange(ctx, opened)
	}
}

func (s *suggestionWidget[T]) effectiveOpen(state *suggestionState) bool {
	if state != nil && state.requestedOpen != nil {
		return *state.requestedOpen
	}
	return s.config.opened
}

func (s *suggestionWidget[T]) selectItem(ctx *internal.Context, state *suggestionState, item ChoiceItem[T]) {
	if s.config.disabled || item.Disabled {
		return
	}
	if s.config.onSelect != nil {
		s.config.onSelect(ctx, item)
	}
	if state != nil && state.inputRef != nil {
		// The menu selection happens after the trigger was laid out. Queue focus
		// through the existing Input Ref so it is consumed on the next frame.
		state.inputRef.Focus()
	}
}

func (s *suggestionWidget[T]) handleKey(ctx *internal.Context, state *suggestionState, items []ChoiceItem[T], model collection.Model, modelOK bool, ev *fluxevent.KeyboardEvent) {
	if ev == nil || ev.DefaultPrevented || s.config.disabled {
		return
	}
	switch ev.Key {
	case "ArrowDown", "ArrowUp", "Home", "End":
		if !modelOK || len(items) == 0 {
			return
		}
		s.requestOpen(ctx, state, true)
		previous := state.active.Key
		next := state.active.Reconcile(model)
		var changed bool
		switch ev.Key {
		case "ArrowDown":
			if previous != "" {
				next, changed = next.Move(model, 1)
			}
		case "ArrowUp":
			if previous == "" {
				next, changed = next.End(model)
			} else {
				next, changed = next.Move(model, -1)
			}
		case "Home":
			next, changed = next.Home(model)
		case "End":
			next, changed = next.End(model)
		}
		state.active = next
		if (changed || previous != next.Key) && s.config.onActiveChange != nil {
			s.config.onActiveChange(ctx, string(next.Key))
		}
		ev.PreventDefault()
	case "Space":
		if s.effectiveOpen(state) && modelOK {
			if state.active.Key == "" {
				return
			}
			active := state.active.Reconcile(model)
			state.active = active
			if item, ok := choiceItemByKey(items, string(active.Key)); ok && !item.Disabled {
				s.selectItem(ctx, state, item)
				s.requestOpen(ctx, state, false)
				ev.PreventDefault()
				return
			}
		}
	case "Escape":
		s.requestOpen(ctx, state, false)
		if state.inputRef != nil {
			state.inputRef.Focus()
		}
		ev.PreventDefault()
	case "Tab":
		// Do not prevent Tab: the runtime must retain its ordinary focus movement.
		s.requestOpen(ctx, state, false)
	}
}

func (s *suggestionWidget[T]) drainRef(ctx *internal.Context, state *suggestionState, items []ChoiceItem[T]) {
	if s.config.ref == nil {
		return
	}
	bindCommandRef(ctx, "combobox", s.config.ref, &s.config.ref.queue)
	if s.config.disabled {
		// Draining preserves the normal Ref rule: commands do not survive a
		// disabled frame and cannot bypass disabled input handling.
		_ = s.config.ref.drainCommands()
		return
	}
	for _, command := range s.config.ref.drainCommands() {
		switch command.kind {
		case comboboxCmdSetQuery:
			if command.query != s.query && s.config.onQueryChange != nil {
				s.config.onQueryChange(ctx, command.query)
			}
		case comboboxCmdOpen:
			s.requestOpen(ctx, state, true)
		case comboboxCmdClose:
			s.requestOpen(ctx, state, false)
		case comboboxCmdToggle:
			s.requestOpen(ctx, state, !s.effectiveOpen(state))
		case comboboxCmdSelectKey:
			if item, ok := choiceItemByKey(items, command.key); ok && !item.Disabled {
				s.selectItem(ctx, state, item)
				s.requestOpen(ctx, state, false)
			}
		case comboboxCmdFocus:
			if state.inputRef != nil {
				state.inputRef.Focus()
			}
		case comboboxCmdBlur:
			if state.inputRef != nil {
				state.inputRef.Blur()
			}
		}
	}
}

func suggestionStateFor(ctx *internal.Context) *suggestionState {
	value := ctx.Memo("advanced-suggestion", func() any { return &suggestionState{} })
	state, ok := value.(*suggestionState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUI/widget: suggestion state type mismatch")
	}
	return state
}

func filteredChoiceItems[T comparable](items []ChoiceItem[T], query string, filter bool) []ChoiceItem[T] {
	result := make([]ChoiceItem[T], 0, len(items))
	needle := strings.ToLower(strings.TrimSpace(query))
	for _, source := range items {
		item := source
		item.Key = choiceItemKey(item)
		if filter && needle != "" {
			haystack := item.TypeaheadText
			if haystack == "" {
				haystack = item.Label
			}
			if !strings.Contains(strings.ToLower(haystack), needle) {
				continue
			}
		}
		result = append(result, item)
	}
	return result
}

func choiceItemKey[T comparable](item ChoiceItem[T]) string {
	return strings.TrimSpace(item.Key)
}

func choiceCollectionModel[T comparable](items []ChoiceItem[T]) (collection.Model, bool) {
	modelItems := make([]collection.Item, 0, len(items))
	for _, item := range items {
		modelItems = append(modelItems, collection.Item{Key: collection.Key(item.Key), Disabled: item.Disabled})
	}
	model, err := collection.New(modelItems)
	if err != nil {
		return collection.Model{}, false
	}
	return model, true
}

func choiceItemByKey[T comparable](items []ChoiceItem[T], key string) (ChoiceItem[T], bool) {
	for _, item := range items {
		if item.Key == key {
			return item, true
		}
	}
	return ChoiceItem[T]{}, false
}

func suggestionMenuItems[T comparable](items []ChoiceItem[T], selectedKey, activeKey string) []MenuItem {
	menuItems := make([]MenuItem, 0, len(items))
	for _, item := range items {
		label := item.Label
		if label == "" {
			label = fmt.Sprintf("%v", item.Value)
		}
		menuItems = append(menuItems, MenuItem{
			Key:           item.Key,
			Label:         label,
			Leading:       item.Leading,
			Trailing:      item.Trailing,
			Disabled:      item.Disabled,
			Selected:      item.Key == selectedKey,
			Active:        item.Key == activeKey,
			TypeaheadText: item.TypeaheadText,
		})
	}
	return menuItems
}

func choiceRequiredLabel(label string, required bool) string {
	if label == "" || !required {
		return label
	}
	return label + " *"
}

type comboboxCommandKind uint8

const (
	comboboxCmdSetQuery comboboxCommandKind = iota + 1
	comboboxCmdOpen
	comboboxCmdClose
	comboboxCmdToggle
	comboboxCmdSelectKey
	comboboxCmdFocus
	comboboxCmdBlur
)

type comboboxCommand[T comparable] struct {
	kind  comboboxCommandKind
	query string
	key   string
}

func (c comboboxCommand[T]) pendingRefBytes() int {
	return len(c.query) + len(c.key) + 1
}

// ComboboxRef is a bounded command queue for a controlled Combobox. Commands
// report intentions through callbacks; they never bypass disabled or controlled
// state.
type ComboboxRef[T comparable] struct {
	queue commandQueue[comboboxCommand[T]]
}

func NewComboboxRef[T comparable]() *ComboboxRef[T] { return &ComboboxRef[T]{} }

func (r *ComboboxRef[T]) SetQuery(query string) {
	if r != nil {
		r.queue.enqueue(comboboxCommand[T]{kind: comboboxCmdSetQuery, query: query})
	}
}

func (r *ComboboxRef[T]) Open() {
	if r != nil {
		r.queue.enqueue(comboboxCommand[T]{kind: comboboxCmdOpen})
	}
}

func (r *ComboboxRef[T]) Close() {
	if r != nil {
		r.queue.enqueue(comboboxCommand[T]{kind: comboboxCmdClose})
	}
}

func (r *ComboboxRef[T]) Toggle() {
	if r != nil {
		r.queue.enqueue(comboboxCommand[T]{kind: comboboxCmdToggle})
	}
}

func (r *ComboboxRef[T]) SelectKey(key string) {
	if r != nil && key != "" {
		r.queue.enqueue(comboboxCommand[T]{kind: comboboxCmdSelectKey, key: key})
	}
}

func (r *ComboboxRef[T]) Focus() {
	if r != nil {
		r.queue.enqueue(comboboxCommand[T]{kind: comboboxCmdFocus})
	}
}

func (r *ComboboxRef[T]) Blur() {
	if r != nil {
		r.queue.enqueue(comboboxCommand[T]{kind: comboboxCmdBlur})
	}
}

func (r *ComboboxRef[T]) drainCommands() []comboboxCommand[T] {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}

// AutocompleteRef and SearchSelectRef share the Combobox command contract.
type AutocompleteRef[T comparable] = ComboboxRef[T]
type SearchSelectRef[T comparable] = ComboboxRef[T]

func NewAutocompleteRef[T comparable]() *AutocompleteRef[T] { return NewComboboxRef[T]() }
func NewSearchSelectRef[T comparable]() *SearchSelectRef[T] { return NewComboboxRef[T]() }
