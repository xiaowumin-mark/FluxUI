package widget

import (
	"fmt"
	"image"
	"strings"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/internal/collection"
	"github.com/xiaowumin-mark/FluxUI/layout"

	"gioui.org/io/key"
	gioLayout "gioui.org/layout"
)

// MultiSelectOption configures a controlled MultiSelect.
//
// selected, query, and opened are host-owned snapshots. User interaction and
// Ref commands report an intention through callbacks; they never mutate a
// caller-owned slice or start asynchronous suggestion work in Layout.
type MultiSelectOption[T comparable] func(*multiSelectConfig[T])

type multiSelectConfig[T comparable] struct {
	selectedKeys    []string
	hasSelectedKeys bool
	query           string
	opened          bool
	pending         bool
	error           bool
	errorText       string
	disabled        bool
	filter          bool
	showMenu        bool
	allowCustom     bool
	placeholder     string
	label           string
	supportingText  string
	required        bool
	maxHeight       float32
	inputOptions    []InputOption

	onChange             func(ctx *internal.Context, selected []T)
	onSelectedKeysChange func(ctx *internal.Context, selectedKeys []string)
	onToggle             func(ctx *internal.Context, item ChoiceItem[T], selected bool)
	onQueryChange        func(ctx *internal.Context, query string)
	onOpenChange         func(ctx *internal.Context, opened bool)
	onActiveChange       func(ctx *internal.Context, key string)
	onCustomValue        func(ctx *internal.Context, value string)
	onBeforeInput        fluxevent.InputHandler
	onInputEvent         fluxevent.InputHandler
	onSubmit             fluxevent.InputHandler
	ref                  *MultiSelectRef[T]
}

func defaultMultiSelectConfig[T comparable]() multiSelectConfig[T] {
	return multiSelectConfig[T]{
		filter:      true,
		showMenu:    true,
		placeholder: "Search…",
		maxHeight:   280,
	}
}

// MultiSelect renders a controlled searchable multi-picker. ChoiceItem.Key
// owns option identity and roving focus. selected is a value set supplied by
// the host; option filtering, reordering, and removal never rewrite it.
func MultiSelect[T comparable](selected []T, options []ChoiceItem[T], opts ...MultiSelectOption[T]) Widget {
	cfg := defaultMultiSelectConfig[T]()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return &multiSelectWidget[T]{
		selected: append([]T(nil), selected...),
		options:  append([]ChoiceItem[T](nil), options...),
		config:   cfg,
	}
}

// MultiSelectQuery supplies the controlled filter query. MultiSelect has no
// query constructor argument so that selected remains its concise primary API.
func MultiSelectQuery[T comparable](query string) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.query = query }
}

// MultiSelectOpened supplies the controlled overlay state.
func MultiSelectOpened[T comparable](opened bool) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.opened = opened }
}

// MultiSelectPending presents a host-controlled loading snapshot.
func MultiSelectPending[T comparable](pending bool) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.pending = pending }
}

// MultiSelectError presents a host-controlled invalid state.
func MultiSelectError[T comparable](invalid bool) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.error = invalid }
}

// MultiSelectErrorText presents a host-controlled validation message.
func MultiSelectErrorText[T comparable](text string) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.errorText = text }
}

func MultiSelectDisabled[T comparable](disabled bool) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.disabled = disabled }
}

// MultiSelectFilterOptions controls local case-insensitive matching over
// TypeaheadText (or Label). It never starts a host query.
func MultiSelectFilterOptions[T comparable](filter bool) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.filter = filter }
}

func MultiSelectPlaceholder[T comparable](text string) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.placeholder = text }
}

func MultiSelectLabel[T comparable](text string) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.label = text }
}

func MultiSelectSupportingText[T comparable](text string) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.supportingText = text }
}

func MultiSelectRequired[T comparable](required bool) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.required = required }
}

func MultiSelectMaxHeight[T comparable](height float32) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.maxHeight = height }
}

// MultiSelectInputOptions forwards presentation-only Input options. The
// multi-picker remains authoritative for value, disabled, and event bridges.
func MultiSelectInputOptions[T comparable](opts ...InputOption) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) {
		cfg.inputOptions = append(cfg.inputOptions, opts...)
	}
}

// MultiSelectSelectedKeys supplies an explicit, stable-key selection
// snapshot. When present (including an empty slice), keys are authoritative
// for chips, menu rows, keyboard actions, and Ref commands. This is the
// preferred selection API when distinct options may carry equal values.
func MultiSelectSelectedKeys[T comparable](keys []string) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) {
		cfg.selectedKeys = append([]string(nil), keys...)
		cfg.hasSelectedKeys = true
	}
}

// MultiSelectOnChange is called exactly once for an actual option toggle. The
// passed slice is a copy in visual selection order and can be retained by the
// host.
func MultiSelectOnChange[T comparable](fn func(ctx *internal.Context, selected []T)) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.onChange = fn }
}

// MultiSelectOnSelectedKeysChange receives the next stable-key selection for
// an actual toggle. It is the matching callback for MultiSelectSelectedKeys.
func MultiSelectOnSelectedKeysChange[T comparable](fn func(ctx *internal.Context, selectedKeys []string)) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.onSelectedKeysChange = fn }
}

// MultiSelectOnToggle receives the stable option identity and its resulting
// selected state. Multiple observers compose in option order.
func MultiSelectOnToggle[T comparable](fn func(ctx *internal.Context, item ChoiceItem[T], selected bool)) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) {
		previous := cfg.onToggle
		cfg.onToggle = func(ctx *internal.Context, item ChoiceItem[T], selected bool) {
			if previous != nil {
				previous(ctx, item, selected)
			}
			if fn != nil {
				fn(ctx, item, selected)
			}
		}
	}
}

// MultiSelectOnSelect observes only transitions into the selected state.
func MultiSelectOnSelect[T comparable](fn func(ctx *internal.Context, item ChoiceItem[T])) MultiSelectOption[T] {
	return MultiSelectOnToggle(func(ctx *internal.Context, item ChoiceItem[T], selected bool) {
		if selected && fn != nil {
			fn(ctx, item)
		}
	})
}

// MultiSelectOnRemove observes chip or menu transitions out of the selected
// state.
func MultiSelectOnRemove[T comparable](fn func(ctx *internal.Context, item ChoiceItem[T])) MultiSelectOption[T] {
	return MultiSelectOnToggle(func(ctx *internal.Context, item ChoiceItem[T], selected bool) {
		if !selected && fn != nil {
			fn(ctx, item)
		}
	})
}

func MultiSelectOnQueryChange[T comparable](fn func(ctx *internal.Context, query string)) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.onQueryChange = fn }
}

func MultiSelectOnOpenChange[T comparable](fn func(ctx *internal.Context, opened bool)) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.onOpenChange = fn }
}

func MultiSelectOnActiveChange[T comparable](fn func(ctx *internal.Context, key string)) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.onActiveChange = fn }
}

func MultiSelectOnBeforeInput[T comparable](fn fluxevent.InputHandler) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.onBeforeInput = fn }
}

func MultiSelectOnInputEvent[T comparable](fn fluxevent.InputHandler) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.onInputEvent = fn }
}

func MultiSelectOnSubmit[T comparable](fn fluxevent.InputHandler) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.onSubmit = fn }
}

type multiSelectWidget[T comparable] struct {
	selected []T
	options  []ChoiceItem[T]
	config   multiSelectConfig[T]
}

type multiSelectState struct {
	active        collection.RovingFocus
	inputRef      *InputRef
	requestedOpen *bool
	customValue   string
}

func (m *multiSelectWidget[T]) Layout(ctx *internal.Context) layout.Dimensions {
	if ctx == nil {
		return layout.Dimensions{}
	}
	state := multiSelectStateFor(ctx)
	if state.inputRef == nil {
		state.inputRef = NewInputRef()
	}
	// opened is controlled: preserve request composition only within the current
	// frame, then let the host snapshot win on the following layout.
	state.requestedOpen = nil
	if m.config.disabled {
		m.closeForDisabled(ctx, state)
	}
	if state.customValue != "" && state.customValue != strings.TrimSpace(m.config.query) {
		state.customValue = ""
	}

	// Validate the complete option snapshot before rendering any filtered rows.
	// A duplicate or blank collection key must not become a silently indexed
	// menu merely because a query happens to hide one of the offending rows.
	allItems := filteredChoiceItems(m.options, "", false)
	_, optionsValid := choiceCollectionModel(allItems)
	items := filteredChoiceItems(m.options, m.config.query, m.config.filter)
	model, visibleModelOK := choiceCollectionModel(items)
	modelOK := optionsValid && visibleModelOK
	if modelOK {
		state.active = state.active.Reconcile(model)
	} else {
		state.active = collection.RovingFocus{}
	}

	m.drainRef(ctx, state, allItems, modelOK)

	errorText := m.config.errorText
	if errorText == "" && !modelOK && len(m.options) > 0 {
		errorText = "Options require unique, non-empty keys"
	}
	inputOptions := append([]InputOption(nil), m.config.inputOptions...)
	if captureKeys := m.captureKeys(state, items, modelOK); len(captureKeys) > 0 {
		inputOptions = append(inputOptions, inputCaptureKeys(captureKeys...))
	}
	inputOptions = append(inputOptions,
		InputPlaceholder(m.config.placeholder),
		InputLabel(choiceRequiredLabel(m.config.label, m.config.required)),
		InputSupportingText(m.config.supportingText),
		InputError(m.config.error || errorText != ""),
		InputErrorText(errorText),
		InputDisabled(m.config.disabled),
		InputAttachRef(state.inputRef),
	)
	if m.config.pending {
		inputOptions = append(inputOptions, InputTrailing(Text("Loading…")))
	}
	inputOptions = append(inputOptions, InputOnBeforeInput(func(actionCtx *internal.Context, event *fluxevent.InputEvent) {
		if m.config.onBeforeInput != nil {
			m.config.onBeforeInput(actionCtx, event)
		}
	}))
	if m.config.onInputEvent != nil {
		inputOptions = append(inputOptions, InputOnInputEvent(m.config.onInputEvent))
	}
	inputOptions = append(inputOptions,
		InputOnFocus(func(actionCtx *internal.Context, focused bool) {
			if focused && !m.config.disabled {
				m.requestOpen(actionCtx, state, true)
			}
		}),
		InputOnChange(func(actionCtx *internal.Context, query string) {
			if m.config.disabled {
				return
			}
			if m.config.onQueryChange != nil && query != m.config.query {
				m.config.onQueryChange(actionCtx, query)
			}
			m.requestOpen(actionCtx, state, true)
		}),
		InputOnSubmit(func(actionCtx *internal.Context, event *fluxevent.InputEvent) {
			if m.config.onSubmit != nil {
				m.config.onSubmit(actionCtx, event)
			}
			if event == nil || event.DefaultPrevented || m.config.disabled {
				return
			}
			if m.config.showMenu && m.effectiveOpened(state) && modelOK {
				active := state.active.Reconcile(model)
				state.active = active
				if item, ok := choiceItemByKey(items, string(active.Key)); ok && !item.Disabled {
					m.toggleItem(actionCtx, state, item, m.selected, m.config.selectedKeys, multiSelectToggle)
					event.PreventDefault()
					return
				}
			}
			if m.config.allowCustom {
				if m.requestCustomValue(actionCtx, state, event.Value) {
					event.PreventDefault()
				}
			}
		}),
	)

	trigger := TextField(m.config.query, inputOptions...)
	menuItems := []MenuItem(nil)
	if modelOK {
		menuItems = multiSelectMenuItems(items, m.selected, m.config.selectedKeys, m.config.hasSelectedKeys, string(state.active.Key))
	}
	picker := DropdownMenu(m.config.opened && m.config.showMenu && len(menuItems) > 0, trigger, menuItems,
		DropdownMenuMaxHeight(m.config.maxHeight),
		dropdownMenuDisabled(m.config.disabled || !m.config.showMenu),
		DropdownMenuOnOpenChange(func(actionCtx *internal.Context, opened bool) {
			m.requestOpen(actionCtx, state, opened)
		}),
		DropdownMenuOnSelect(func(actionCtx *internal.Context, key string) {
			item, ok := choiceItemByKey(items, key)
			if !ok || item.Disabled || m.config.disabled {
				return
			}
			m.toggleItem(actionCtx, state, item, m.selected, m.config.selectedKeys, multiSelectToggle)
		}),
	)

	// The chips slot is always present. Keeping the input in the second fixed
	// child slot prevents a new first chip from moving its path/focus target.
	chips := ScrollView(multiSelectChipList[T]{
		selected:        m.selected,
		selectedKeys:    m.config.selectedKeys,
		useSelectedKeys: m.config.hasSelectedKeys,
		options:         allItems,
		disabled:        m.config.disabled,
		onRemove: func(actionCtx *internal.Context, item ChoiceItem[T]) {
			m.toggleItem(actionCtx, state, item, m.selected, m.config.selectedKeys, multiSelectRemove)
		},
	}, ScrollVertical(false), ScrollHorizontal(true), ScrollBarVisible(false))
	body := Column(chips, picker)
	return KeyboardScope(body,
		KeyboardScopeFocusable(!m.config.disabled),
		KeyOnDown(func(actionCtx *internal.Context, event *fluxevent.KeyboardEvent) {
			m.handleKey(actionCtx, state, items, model, modelOK, event)
		}),
	).Layout(ctx.Child(0))
}

func (m *multiSelectWidget[T]) closeForDisabled(ctx *internal.Context, state *multiSelectState) {
	if state == nil {
		return
	}
	wasOpen := m.config.showMenu && m.effectiveOpened(state)
	if state.inputRef != nil {
		state.inputRef.Blur()
	}
	md3ClearFocusIfInside(ctx, ctx.PathID())
	if !wasOpen {
		return
	}
	closed := false
	state.requestedOpen = &closed
	if m.config.onOpenChange != nil {
		m.config.onOpenChange(ctx, false)
	}
}

func (m *multiSelectWidget[T]) captureKeys(state *multiSelectState, items []ChoiceItem[T], modelOK bool) []key.Name {
	if m.config.disabled || !m.config.showMenu || !modelOK || len(items) == 0 {
		return nil
	}
	keys := []key.Name{key.NameUpArrow, key.NameDownArrow, key.NameHome, key.NameEnd}
	if m.config.opened && state != nil {
		if item, ok := choiceItemByKey(items, string(state.active.Key)); ok && !item.Disabled {
			keys = append(keys, key.NameSpace)
		}
	}
	return keys
}

func (m *multiSelectWidget[T]) requestOpen(ctx *internal.Context, state *multiSelectState, opened bool) {
	if state == nil || m.config.disabled || !m.config.showMenu {
		return
	}
	if m.effectiveOpened(state) == opened {
		return
	}
	next := opened
	state.requestedOpen = &next
	if m.config.onOpenChange != nil {
		m.config.onOpenChange(ctx, opened)
	}
}

// effectiveOpened is the controlled snapshot plus a same-frame request. Ref
// commands can therefore compose: Open followed by Toggle reports true, then
// false, instead of evaluating both operations against the stale prop.
func (m *multiSelectWidget[T]) effectiveOpened(state *multiSelectState) bool {
	if state != nil && state.requestedOpen != nil {
		return *state.requestedOpen
	}
	return m.config.opened
}

type multiSelectIntent uint8

const (
	multiSelectToggle multiSelectIntent = iota
	multiSelectSelect
	multiSelectRemove
)

// toggleItem reports an actual state transition and returns the resulting
// value/key selections. Both current snapshots are explicit so a same-frame
// Ref command sequence can be evaluated without pretending that controlled
// props already changed.
func (m *multiSelectWidget[T]) toggleItem(ctx *internal.Context, state *multiSelectState, item ChoiceItem[T], current []T, currentKeys []string, intent multiSelectIntent) ([]T, []string, bool) {
	if m.config.disabled || item.Disabled {
		return multiSelectCopy(current), multiSelectKeyCopy(currentKeys), false
	}
	item.Key = choiceItemKey(item)
	if m.config.hasSelectedKeys {
		wasSelected := multiSelectContainsKey(currentKeys, item.Key)
		if intent == multiSelectSelect && wasSelected {
			return multiSelectCopy(current), multiSelectKeyCopy(currentKeys), false
		}
		if intent == multiSelectRemove && !wasSelected {
			return multiSelectCopy(current), multiSelectKeyCopy(currentKeys), false
		}

		var nextKeys []string
		selected := !wasSelected
		if wasSelected {
			nextKeys = multiSelectWithoutKey(currentKeys, item.Key)
			selected = false
		} else {
			nextKeys = append(multiSelectKeyCopy(currentKeys), item.Key)
		}
		if m.config.onToggle != nil {
			m.config.onToggle(ctx, item, selected)
		}
		if m.config.onSelectedKeysChange != nil {
			m.config.onSelectedKeysChange(ctx, multiSelectKeyCopy(nextKeys))
		}
		if m.config.onChange != nil {
			m.config.onChange(ctx, multiSelectValuesForKeys(nextKeys, m.options, current))
		}
		if state != nil && state.inputRef != nil {
			state.inputRef.Focus()
		}
		return multiSelectCopy(current), nextKeys, true
	}
	wasSelected := multiSelectContains(current, item.Value)
	if intent == multiSelectSelect && wasSelected {
		return multiSelectCopy(current), multiSelectKeyCopy(currentKeys), false
	}
	if intent == multiSelectRemove && !wasSelected {
		return multiSelectCopy(current), multiSelectKeyCopy(currentKeys), false
	}

	var next []T
	selected := !wasSelected
	if wasSelected {
		next = multiSelectWithout(current, item.Value)
		selected = false
	} else {
		next = append(append([]T(nil), current...), item.Value)
	}
	if m.config.onToggle != nil {
		m.config.onToggle(ctx, item, selected)
	}
	if m.config.onChange != nil {
		m.config.onChange(ctx, multiSelectCopy(next))
	}
	if state != nil && state.inputRef != nil {
		// Menu rows may have become the active focus target. Restore typing focus
		// through the regular Input Ref on the next input layout.
		state.inputRef.Focus()
	}
	return next, multiSelectKeyCopy(currentKeys), true
}

func (m *multiSelectWidget[T]) handleKey(ctx *internal.Context, state *multiSelectState, items []ChoiceItem[T], model collection.Model, modelOK bool, event *fluxevent.KeyboardEvent) {
	if event == nil || event.DefaultPrevented || m.config.disabled {
		return
	}
	switch event.Key {
	case "ArrowDown", "ArrowUp", "Home", "End":
		if !m.config.showMenu || !modelOK || len(items) == 0 {
			return
		}
		m.requestOpen(ctx, state, true)
		previous := state.active.Key
		next := state.active.Reconcile(model)
		var changed bool
		switch event.Key {
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
		if (changed || previous != next.Key) && m.config.onActiveChange != nil {
			m.config.onActiveChange(ctx, string(next.Key))
		}
		event.PreventDefault()
	case "Space":
		if m.config.showMenu && m.effectiveOpened(state) && modelOK {
			if state.active.Key == "" {
				return
			}
			active := state.active.Reconcile(model)
			state.active = active
			if item, ok := choiceItemByKey(items, string(active.Key)); ok && !item.Disabled {
				if _, _, changed := m.toggleItem(ctx, state, item, m.selected, m.config.selectedKeys, multiSelectToggle); changed {
					event.PreventDefault()
					return
				}
			}
		}
	case "Escape":
		m.requestOpen(ctx, state, false)
	case "Tab":
		// Preserve normal Tab / Shift+Tab focus movement.
		m.requestOpen(ctx, state, false)
	}
}

func (m *multiSelectWidget[T]) requestCustomValue(ctx *internal.Context, state *multiSelectState, value string) bool {
	if m.config.disabled || m.config.onCustomValue == nil {
		return false
	}
	value = strings.TrimSpace(value)
	if value == "" || state == nil {
		return false
	}
	// A KeyDown and an Input submit can describe the same Enter. Retain this
	// request until the controlled query changes so the host sees it once.
	if state.customValue == value {
		return false
	}
	state.customValue = value
	m.config.onCustomValue(ctx, value)
	return true
}

func (m *multiSelectWidget[T]) drainRef(ctx *internal.Context, state *multiSelectState, allItems []ChoiceItem[T], modelOK bool) {
	if m.config.ref == nil {
		return
	}
	bindCommandRef(ctx, "multi-select", m.config.ref, &m.config.ref.queue)
	commands := m.config.ref.drainCommands()
	if m.config.disabled {
		return
	}
	working := multiSelectCopy(m.selected)
	workingKeys := multiSelectKeyCopy(m.config.selectedKeys)
	for _, command := range commands {
		switch command.kind {
		case multiSelectCmdSetQuery:
			if command.query != m.config.query && m.config.onQueryChange != nil {
				m.config.onQueryChange(ctx, command.query)
			}
		case multiSelectCmdOpen:
			m.requestOpen(ctx, state, true)
		case multiSelectCmdClose:
			m.requestOpen(ctx, state, false)
		case multiSelectCmdToggle:
			m.requestOpen(ctx, state, !m.effectiveOpened(state))
		case multiSelectCmdToggleKey, multiSelectCmdSelectKey, multiSelectCmdRemoveKey:
			if !modelOK {
				continue
			}
			item, ok := choiceItemByKey(allItems, command.key)
			if !ok && m.config.hasSelectedKeys && command.kind != multiSelectCmdSelectKey && multiSelectContainsKey(workingKeys, command.key) {
				item = ChoiceItem[T]{Key: command.key, Label: command.key}
				ok = true
			}
			if !ok || item.Disabled {
				continue
			}
			intent := multiSelectToggle
			switch command.kind {
			case multiSelectCmdSelectKey:
				intent = multiSelectSelect
			case multiSelectCmdRemoveKey:
				intent = multiSelectRemove
			}
			working, workingKeys, _ = m.toggleItem(ctx, state, item, working, workingKeys, intent)
		case multiSelectCmdFocus:
			if state.inputRef != nil {
				state.inputRef.Focus()
			}
		case multiSelectCmdBlur:
			if state.inputRef != nil {
				state.inputRef.Blur()
			}
		}
	}
}

func multiSelectStateFor(ctx *internal.Context) *multiSelectState {
	value := ctx.Memo("advanced-multi-select", func() any { return &multiSelectState{} })
	state, ok := value.(*multiSelectState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUI/widget: multi-select state type mismatch")
	}
	return state
}

func multiSelectContains[T comparable](selected []T, value T) bool {
	for _, current := range selected {
		if current == value {
			return true
		}
	}
	return false
}

func multiSelectContainsKey(selected []string, key string) bool {
	for _, current := range selected {
		if current == key {
			return true
		}
	}
	return false
}

func multiSelectWithout[T comparable](selected []T, value T) []T {
	result := make([]T, 0, len(selected))
	for _, current := range selected {
		if current != value {
			result = append(result, current)
		}
	}
	return result
}

func multiSelectWithoutKey(selected []string, key string) []string {
	result := make([]string, 0, len(selected))
	for _, current := range selected {
		if current != key {
			result = append(result, current)
		}
	}
	return result
}

func multiSelectCopy[T any](values []T) []T {
	if values == nil {
		return nil
	}
	result := make([]T, len(values))
	copy(result, values)
	return result
}

func multiSelectKeyCopy(values []string) []string {
	if values == nil {
		return nil
	}
	result := make([]string, len(values))
	copy(result, values)
	return result
}

// multiSelectValuesForKeys preserves the legacy value callback when the
// stable-key API is active. A known key resolves against the current complete
// option snapshot; a missing key keeps its same-position legacy value when one
// exists, so deleting/reordering async results cannot fabricate a value.
func multiSelectValuesForKeys[T comparable](keys []string, options []ChoiceItem[T], fallback []T) []T {
	values := make([]T, 0, len(keys))
	for index, key := range keys {
		if item, ok := multiSelectItemByKey(options, key); ok {
			values = append(values, item.Value)
			continue
		}
		if index < len(fallback) {
			values = append(values, fallback[index])
		}
	}
	return values
}

func multiSelectMenuItems[T comparable](items []ChoiceItem[T], selected []T, selectedKeys []string, useSelectedKeys bool, activeKey string) []MenuItem {
	menuItems := make([]MenuItem, 0, len(items))
	for _, item := range items {
		isSelected := multiSelectContains(selected, item.Value)
		if useSelectedKeys {
			isSelected = multiSelectContainsKey(selectedKeys, item.Key)
		}
		menuItems = append(menuItems, MenuItem{
			Key:           item.Key,
			Label:         multiSelectChoiceLabel(item),
			Leading:       item.Leading,
			Trailing:      item.Trailing,
			Disabled:      item.Disabled,
			Selected:      isSelected,
			Active:        item.Key == activeKey,
			KeepOpen:      true,
			TypeaheadText: item.TypeaheadText,
		})
	}
	return menuItems
}

func multiSelectChoiceLabel[T comparable](item ChoiceItem[T]) string {
	if item.Label != "" {
		return item.Label
	}
	return fmt.Sprintf("%v", item.Value)
}

type multiSelectChipList[T comparable] struct {
	selected        []T
	selectedKeys    []string
	useSelectedKeys bool
	options         []ChoiceItem[T]
	disabled        bool
	onRemove        func(ctx *internal.Context, item ChoiceItem[T])
}

func (l multiSelectChipList[T]) Layout(ctx *internal.Context) layout.Dimensions {
	if ctx == nil {
		return layout.Dimensions{}
	}
	items := multiSelectSelectedItems(l.selected, l.options)
	if l.useSelectedKeys {
		items = multiSelectSelectedItemsByKeys[T](l.selectedKeys, l.options)
	}
	if len(items) == 0 {
		return layout.Dimensions{}
	}
	children := make([]gioLayout.FlexChild, 0, len(items)*2-1)
	for index := range items {
		item := items[index]
		key := choiceItemKey(item)
		chipOptions := []ChipOption{
			ChipSelected(true),
			ChipDisabled(l.disabled || item.Disabled),
		}
		if item.Leading != nil {
			chipOptions = append(chipOptions, ChipLeading(item.Leading))
		}
		if l.onRemove != nil && !l.disabled && !item.Disabled {
			selectedItem := item
			chipOptions = append(chipOptions, ChipOnRemove(func(actionCtx *internal.Context) {
				l.onRemove(actionCtx, selectedItem)
			}))
		}
		chip := InputChip(multiSelectChoiceLabel(item), chipOptions...)
		children = append(children, gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
			next := ctx.Scope("multi-select-chip:" + key)
			next.Gtx = gtx
			next.Gtx.Constraints.Min = image.Point{}
			dims := chip.Layout(next)
			return gioLayout.Dimensions{Size: dims.Size}
		}))
		if index < len(items)-1 {
			children = append(children, gioLayout.Rigid(func(gtx gioLayout.Context) gioLayout.Dimensions {
				return gioLayout.Dimensions{Size: image.Point{X: gtx.Dp(4)}}
			}))
		}
	}
	dims := gioLayout.Flex{Axis: gioLayout.Horizontal, Alignment: gioLayout.Middle}.Layout(ctx.Gtx, children...)
	return layout.Dimensions{Size: dims.Size}
}

func multiSelectSelectedItems[T comparable](selected []T, options []ChoiceItem[T]) []ChoiceItem[T] {
	items := make([]ChoiceItem[T], 0, len(selected))
	seen := make(map[T]struct{}, len(selected))
	for _, value := range selected {
		if _, duplicate := seen[value]; duplicate {
			continue
		}
		seen[value] = struct{}{}
		item, ok := multiSelectItemByValue(options, value)
		if !ok {
			// A host can legitimately retain selected values after an async result
			// snapshot drops an option. Keep a removable fallback chip instead of
			// silently losing the controlled selection.
			item = ChoiceItem[T]{
				Key:   fmt.Sprintf("selected:%T:%v", value, value),
				Label: fmt.Sprintf("%v", value),
				Value: value,
			}
		}
		item.Key = choiceItemKey(item)
		items = append(items, item)
	}
	return items
}

func multiSelectSelectedItemsByKeys[T comparable](selectedKeys []string, options []ChoiceItem[T]) []ChoiceItem[T] {
	items := make([]ChoiceItem[T], 0, len(selectedKeys))
	seen := make(map[string]struct{}, len(selectedKeys))
	for _, key := range selectedKeys {
		if key == "" {
			continue
		}
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key] = struct{}{}
		item, ok := multiSelectItemByKey(options, key)
		if !ok {
			// Key-controlled selection remains visible even when a host removes
			// the corresponding async option snapshot.
			item = ChoiceItem[T]{Key: key, Label: key}
		}
		item.Key = choiceItemKey(item)
		items = append(items, item)
	}
	return items
}

func multiSelectItemByValue[T comparable](items []ChoiceItem[T], value T) (ChoiceItem[T], bool) {
	for _, source := range items {
		if source.Value != value {
			continue
		}
		item := source
		item.Key = choiceItemKey(item)
		return item, true
	}
	return ChoiceItem[T]{}, false
}

func multiSelectItemByKey[T comparable](items []ChoiceItem[T], key string) (ChoiceItem[T], bool) {
	for _, source := range items {
		item := source
		item.Key = choiceItemKey(item)
		if item.Key == key {
			return item, true
		}
	}
	return ChoiceItem[T]{}, false
}

type multiSelectCommandKind uint8

const (
	multiSelectCmdSetQuery multiSelectCommandKind = iota + 1
	multiSelectCmdOpen
	multiSelectCmdClose
	multiSelectCmdToggle
	multiSelectCmdToggleKey
	multiSelectCmdSelectKey
	multiSelectCmdRemoveKey
	multiSelectCmdFocus
	multiSelectCmdBlur
)

type multiSelectCommand[T comparable] struct {
	kind  multiSelectCommandKind
	query string
	key   string
}

func (c multiSelectCommand[T]) pendingRefBytes() int {
	return len(c.query) + len(c.key) + 1
}

// MultiSelectRef is a bounded command queue for a controlled MultiSelect.
// Commands are consumed during the next mounted layout and are discarded while
// disabled, on unmount, or when rebound to another owner.
type MultiSelectRef[T comparable] struct {
	queue commandQueue[multiSelectCommand[T]]
}

func NewMultiSelectRef[T comparable]() *MultiSelectRef[T] { return &MultiSelectRef[T]{} }

func (r *MultiSelectRef[T]) SetQuery(query string) {
	if r != nil {
		r.queue.enqueue(multiSelectCommand[T]{kind: multiSelectCmdSetQuery, query: query})
	}
}

func (r *MultiSelectRef[T]) Open() {
	if r != nil {
		r.queue.enqueue(multiSelectCommand[T]{kind: multiSelectCmdOpen})
	}
}

func (r *MultiSelectRef[T]) Close() {
	if r != nil {
		r.queue.enqueue(multiSelectCommand[T]{kind: multiSelectCmdClose})
	}
}

func (r *MultiSelectRef[T]) Toggle() {
	if r != nil {
		r.queue.enqueue(multiSelectCommand[T]{kind: multiSelectCmdToggle})
	}
}

// ToggleKey asks the host to toggle the option identified by its stable key.
func (r *MultiSelectRef[T]) ToggleKey(key string) {
	if r != nil && key != "" {
		r.queue.enqueue(multiSelectCommand[T]{kind: multiSelectCmdToggleKey, key: key})
	}
}

// SelectKey asks the host to select a currently unselected option. It is a
// no-op for an already selected option and therefore never emits a duplicate
// change callback.
func (r *MultiSelectRef[T]) SelectKey(key string) {
	if r != nil && key != "" {
		r.queue.enqueue(multiSelectCommand[T]{kind: multiSelectCmdSelectKey, key: key})
	}
}

// RemoveKey asks the host to remove a currently selected option by stable key.
func (r *MultiSelectRef[T]) RemoveKey(key string) {
	if r != nil && key != "" {
		r.queue.enqueue(multiSelectCommand[T]{kind: multiSelectCmdRemoveKey, key: key})
	}
}

func (r *MultiSelectRef[T]) Focus() {
	if r != nil {
		r.queue.enqueue(multiSelectCommand[T]{kind: multiSelectCmdFocus})
	}
}

func (r *MultiSelectRef[T]) Blur() {
	if r != nil {
		r.queue.enqueue(multiSelectCommand[T]{kind: multiSelectCmdBlur})
	}
}

func (r *MultiSelectRef[T]) drainCommands() []multiSelectCommand[T] {
	if r == nil {
		return nil
	}
	return r.queue.drainCommands()
}

// MultiSelectAttachRef binds a command reference for the current frame.
func MultiSelectAttachRef[T comparable](ref *MultiSelectRef[T]) MultiSelectOption[T] {
	return func(cfg *multiSelectConfig[T]) { cfg.ref = ref }
}

// TagOptionItem is one stable TagPicker option. Key is both the option's
// identity and the value returned in TagPicker's selectedKeys slice.
type TagOptionItem struct {
	Key      string
	Label    string
	Disabled bool
	Leading  Widget
	Trailing Widget
}

// TagPickerOption shares MultiSelect's controlled picker contract.
type TagPickerOption = MultiSelectOption[string]

// TagPickerRef shares MultiSelect's bounded command contract.
type TagPickerRef = MultiSelectRef[string]

// TagPicker renders a searchable controlled picker for stable string tag keys.
func TagPicker(selectedKeys []string, query string, options []TagOptionItem, opts ...TagPickerOption) Widget {
	cfg := defaultMultiSelectConfig[string]()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	// query is a required controlled input to TagPicker and wins over a stale
	// generic MultiSelectQuery option accidentally reused by a caller.
	cfg.query = query
	// TagPicker's public value is explicitly a stable key slice, never the
	// coincidental string Value field of a converted ChoiceItem.
	cfg.selectedKeys = multiSelectKeyCopy(selectedKeys)
	cfg.hasSelectedKeys = true
	return &multiSelectWidget[string]{
		selected: append([]string(nil), selectedKeys...),
		options:  tagPickerChoiceItems(options),
		config:   cfg,
	}
}

func TagPickerOpened(opened bool) TagPickerOption     { return MultiSelectOpened[string](opened) }
func TagPickerPending(pending bool) TagPickerOption   { return MultiSelectPending[string](pending) }
func TagPickerError(invalid bool) TagPickerOption     { return MultiSelectError[string](invalid) }
func TagPickerErrorText(text string) TagPickerOption  { return MultiSelectErrorText[string](text) }
func TagPickerDisabled(disabled bool) TagPickerOption { return MultiSelectDisabled[string](disabled) }
func TagPickerFilterOptions(filter bool) TagPickerOption {
	return MultiSelectFilterOptions[string](filter)
}
func TagPickerPlaceholder(text string) TagPickerOption { return MultiSelectPlaceholder[string](text) }
func TagPickerLabel(text string) TagPickerOption       { return MultiSelectLabel[string](text) }
func TagPickerSupportingText(text string) TagPickerOption {
	return MultiSelectSupportingText[string](text)
}
func TagPickerRequired(required bool) TagPickerOption { return MultiSelectRequired[string](required) }
func TagPickerMaxHeight(height float32) TagPickerOption {
	return MultiSelectMaxHeight[string](height)
}
func TagPickerInputOptions(opts ...InputOption) TagPickerOption {
	return MultiSelectInputOptions[string](opts...)
}
func TagPickerOnChange(fn func(ctx *internal.Context, selectedKeys []string)) TagPickerOption {
	return MultiSelectOnSelectedKeysChange[string](fn)
}

// TagPickerOnSelectedKeysChange is the explicit stable-key spelling of
// TagPickerOnChange.
func TagPickerOnSelectedKeysChange(fn func(ctx *internal.Context, selectedKeys []string)) TagPickerOption {
	return MultiSelectOnSelectedKeysChange[string](fn)
}
func TagPickerOnToggle(fn func(ctx *internal.Context, item TagOptionItem, selected bool)) TagPickerOption {
	return MultiSelectOnToggle(func(ctx *internal.Context, item ChoiceItem[string], selected bool) {
		if fn != nil {
			fn(ctx, tagOptionItemFromChoice(item), selected)
		}
	})
}
func TagPickerOnSelect(fn func(ctx *internal.Context, item TagOptionItem)) TagPickerOption {
	return TagPickerOnToggle(func(ctx *internal.Context, item TagOptionItem, selected bool) {
		if selected && fn != nil {
			fn(ctx, item)
		}
	})
}
func TagPickerOnRemove(fn func(ctx *internal.Context, key string)) TagPickerOption {
	return TagPickerOnToggle(func(ctx *internal.Context, item TagOptionItem, selected bool) {
		if !selected && fn != nil {
			fn(ctx, item.Key)
		}
	})
}
func TagPickerOnQueryChange(fn func(ctx *internal.Context, query string)) TagPickerOption {
	return MultiSelectOnQueryChange[string](fn)
}
func TagPickerOnOpenChange(fn func(ctx *internal.Context, opened bool)) TagPickerOption {
	return MultiSelectOnOpenChange[string](fn)
}
func TagPickerOnActiveChange(fn func(ctx *internal.Context, key string)) TagPickerOption {
	return MultiSelectOnActiveChange[string](fn)
}
func TagPickerOnBeforeInput(fn fluxevent.InputHandler) TagPickerOption {
	return MultiSelectOnBeforeInput[string](fn)
}
func TagPickerOnInputEvent(fn fluxevent.InputHandler) TagPickerOption {
	return MultiSelectOnInputEvent[string](fn)
}
func TagPickerOnSubmit(fn fluxevent.InputHandler) TagPickerOption {
	return MultiSelectOnSubmit[string](fn)
}
func NewTagPickerRef() *TagPickerRef { return NewMultiSelectRef[string]() }
func TagPickerAttachRef(ref *TagPickerRef) TagPickerOption {
	return MultiSelectAttachRef[string](ref)
}

func tagPickerChoiceItems(options []TagOptionItem) []ChoiceItem[string] {
	items := make([]ChoiceItem[string], 0, len(options))
	for _, option := range options {
		items = append(items, ChoiceItem[string]{
			Key:      option.Key,
			Label:    option.Label,
			Value:    option.Key,
			Disabled: option.Disabled,
			Leading:  option.Leading,
			Trailing: option.Trailing,
		})
	}
	return items
}

func tagOptionItemFromChoice(item ChoiceItem[string]) TagOptionItem {
	return TagOptionItem{
		Key:      item.Key,
		Label:    item.Label,
		Disabled: item.Disabled,
		Leading:  item.Leading,
		Trailing: item.Trailing,
	}
}

// TagInputOption configures a controlled free-form TagInput. It has the same
// Input event ordering as MultiSelect, but Enter submits one non-empty,
// previously absent tag to the host instead of selecting a suggestion.
type TagInputOption func(*tagInputConfig)

type tagInputConfig struct {
	opened         bool
	pending        bool
	error          bool
	errorText      string
	disabled       bool
	placeholder    string
	label          string
	supportingText string
	required       bool
	maxHeight      float32
	inputOptions   []InputOption

	onChange       func(ctx *internal.Context, tags []string)
	onAdd          func(ctx *internal.Context, tag string)
	onRemove       func(ctx *internal.Context, tag string)
	onQueryChange  func(ctx *internal.Context, query string)
	onOpenChange   func(ctx *internal.Context, opened bool)
	onActiveChange func(ctx *internal.Context, key string)
	onBeforeInput  fluxevent.InputHandler
	onInputEvent   fluxevent.InputHandler
	onSubmit       fluxevent.InputHandler
	ref            *TagInputRef
}

func defaultTagInputConfig() tagInputConfig {
	return tagInputConfig{
		placeholder: "Add tag…",
		maxHeight:   280,
	}
}

// TagInputRef exposes the query, focus, open, and stable-key removal commands
// of its backing controlled picker. Tag creation remains a host-visible Enter
// intention so it cannot bypass the controlled tags snapshot.
type TagInputRef = MultiSelectRef[string]

// TagInput renders controlled free-form tags and a controlled text query.
// Duplicate text values are ignored so each rendered chip has a stable key.
func TagInput(tags []string, query string, opts ...TagInputOption) Widget {
	cfg := defaultTagInputConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return &tagInputWidget{
		tags:   append([]string(nil), tags...),
		query:  query,
		config: cfg,
	}
}

func TagInputOpened(opened bool) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.opened = opened }
}
func TagInputPending(pending bool) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.pending = pending }
}
func TagInputError(invalid bool) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.error = invalid }
}
func TagInputErrorText(text string) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.errorText = text }
}
func TagInputDisabled(disabled bool) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.disabled = disabled }
}
func TagInputPlaceholder(text string) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.placeholder = text }
}
func TagInputLabel(text string) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.label = text }
}
func TagInputSupportingText(text string) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.supportingText = text }
}
func TagInputRequired(required bool) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.required = required }
}
func TagInputMaxHeight(height float32) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.maxHeight = height }
}
func TagInputInputOptions(opts ...InputOption) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.inputOptions = append(cfg.inputOptions, opts...) }
}

// TagInputOnChange is called only for an actual add or removal.
func TagInputOnChange(fn func(ctx *internal.Context, tags []string)) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.onChange = fn }
}
func TagInputOnAdd(fn func(ctx *internal.Context, tag string)) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.onAdd = fn }
}

// TagInputOnTagAdd is an explicit spelling of TagInputOnAdd.
func TagInputOnTagAdd(fn func(ctx *internal.Context, tag string)) TagInputOption {
	return TagInputOnAdd(fn)
}
func TagInputOnRemove(fn func(ctx *internal.Context, tag string)) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.onRemove = fn }
}

// TagInputOnTagRemove is an explicit spelling of TagInputOnRemove.
func TagInputOnTagRemove(fn func(ctx *internal.Context, tag string)) TagInputOption {
	return TagInputOnRemove(fn)
}
func TagInputOnQueryChange(fn func(ctx *internal.Context, query string)) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.onQueryChange = fn }
}
func TagInputOnOpenChange(fn func(ctx *internal.Context, opened bool)) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.onOpenChange = fn }
}
func TagInputOnActiveChange(fn func(ctx *internal.Context, key string)) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.onActiveChange = fn }
}
func TagInputOnBeforeInput(fn fluxevent.InputHandler) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.onBeforeInput = fn }
}
func TagInputOnInputEvent(fn fluxevent.InputHandler) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.onInputEvent = fn }
}
func TagInputOnSubmit(fn fluxevent.InputHandler) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.onSubmit = fn }
}
func NewTagInputRef() *TagInputRef { return NewMultiSelectRef[string]() }
func TagInputAttachRef(ref *TagInputRef) TagInputOption {
	return func(cfg *tagInputConfig) { cfg.ref = ref }
}

type tagInputWidget struct {
	tags   []string
	query  string
	config tagInputConfig
}

func (t *tagInputWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if ctx == nil {
		return layout.Dimensions{}
	}
	cfg := defaultMultiSelectConfig[string]()
	cfg.query = t.query
	cfg.opened = t.config.opened
	cfg.pending = t.config.pending
	cfg.error = t.config.error
	cfg.errorText = t.config.errorText
	cfg.disabled = t.config.disabled
	cfg.filter = false
	cfg.showMenu = false
	cfg.allowCustom = true
	cfg.placeholder = t.config.placeholder
	cfg.label = t.config.label
	cfg.supportingText = t.config.supportingText
	cfg.required = t.config.required
	cfg.maxHeight = t.config.maxHeight
	cfg.inputOptions = append([]InputOption(nil), t.config.inputOptions...)
	cfg.onChange = t.config.onChange
	cfg.onQueryChange = t.config.onQueryChange
	cfg.onOpenChange = t.config.onOpenChange
	cfg.onActiveChange = t.config.onActiveChange
	cfg.onBeforeInput = t.config.onBeforeInput
	cfg.onInputEvent = t.config.onInputEvent
	cfg.onSubmit = t.config.onSubmit
	cfg.ref = t.config.ref
	cfg.onToggle = func(actionCtx *internal.Context, item ChoiceItem[string], selected bool) {
		if !selected && t.config.onRemove != nil {
			t.config.onRemove(actionCtx, item.Value)
		}
	}
	cfg.onCustomValue = func(actionCtx *internal.Context, value string) {
		if multiSelectContains(t.tags, value) {
			return
		}
		next := append(append([]string(nil), t.tags...), value)
		item := tagInputChoiceItem(value)
		if cfg.onToggle != nil {
			cfg.onToggle(actionCtx, item, true)
		}
		if t.config.onChange != nil {
			t.config.onChange(actionCtx, append([]string(nil), next...))
		}
		if t.config.onAdd != nil {
			t.config.onAdd(actionCtx, value)
		}
		if t.config.onQueryChange != nil && t.query != "" {
			t.config.onQueryChange(actionCtx, "")
		}
	}
	return (&multiSelectWidget[string]{
		selected: append([]string(nil), t.tags...),
		options:  tagInputChoiceItems(t.tags),
		config:   cfg,
	}).Layout(ctx)
}

func tagInputChoiceItem(tag string) ChoiceItem[string] {
	return ChoiceItem[string]{Key: tag, Label: tag, Value: tag}
}

func tagInputChoiceItems(tags []string) []ChoiceItem[string] {
	items := make([]ChoiceItem[string], 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		if _, duplicate := seen[tag]; duplicate {
			continue
		}
		seen[tag] = struct{}{}
		items = append(items, tagInputChoiceItem(tag))
	}
	return items
}
