package widget

import (
	"image"
	"reflect"
	"testing"
	"time"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/internal/collection"

	"gioui.org/io/key"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func layoutTagsForTest(rt *internal.Runtime, w Widget, frame int) {
	var ops op.Ops
	gtx := gioLayout.Context{
		Constraints: gioLayout.Exact(image.Pt(320, 180)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Now:         time.Date(2026, 7, 12, 12, 0, frame, 0, time.UTC),
		Ops:         &ops,
	}
	rt.BeginFrame()
	if w != nil {
		w.Layout(internal.NewContext(gtx, rt).Scope("tags-test"))
	}
	rt.EndFrame()
}

func TestMultiSelectStableKeyFocusAndRemovedValuePreserved(t *testing.T) {
	initial := []ChoiceItem[string]{
		{Key: "alpha", Label: "Alpha", Value: "a"},
		{Key: "beta", Label: "Beta", Value: "b"},
	}
	initialModel, ok := choiceCollectionModel(initial)
	if !ok {
		t.Fatal("initial collection model rejected stable keys")
	}
	active := collectionRovingFocusForTest("beta").Reconcile(initialModel)

	reordered := []ChoiceItem[string]{
		{Key: "beta", Label: "Beta", Value: "b"},
		{Key: "alpha", Label: "Alpha", Value: "a"},
	}
	reorderedModel, ok := choiceCollectionModel(reordered)
	if !ok {
		t.Fatal("reordered collection model rejected stable keys")
	}
	active = active.Reconcile(reorderedModel)
	if got := string(active.Key); got != "beta" {
		t.Fatalf("active key after reorder = %q, want beta", got)
	}

	chips := multiSelectSelectedItems([]string{"b", "removed"}, reordered)
	if len(chips) != 2 {
		t.Fatalf("selected chips = %#v, want retained option plus fallback", chips)
	}
	if chips[0].Key != "beta" || chips[0].Value != "b" {
		t.Fatalf("reordered selected chip = %#v, want beta", chips[0])
	}
	if chips[1].Value != "removed" || chips[1].Label != "removed" {
		t.Fatalf("removed selected value was lost: %#v", chips[1])
	}
}

// collection.RovingFocus is intentionally kept behind this small helper so
// this test reads as a stable-key behavior test rather than a collection API
// test.
func collectionRovingFocusForTest(key string) collection.RovingFocus {
	return collection.RovingFocus{Key: collection.Key(key)}
}

func TestMultiSelectOnChangeOnlyForActualToggle(t *testing.T) {
	var received [][]string
	m := &multiSelectWidget[string]{
		selected: []string{"a"},
		config: multiSelectConfig[string]{
			onChange: func(_ *internal.Context, selected []string) {
				received = append(received, multiSelectCopy(selected))
			},
		},
	}
	state := &multiSelectState{inputRef: NewInputRef()}
	item := ChoiceItem[string]{Key: "alpha", Value: "a"}

	if _, _, changed := m.toggleItem(nil, state, item, []string{"a"}, nil, multiSelectSelect); changed {
		t.Fatal("SelectKey semantics changed an already selected value")
	}
	if _, _, changed := m.toggleItem(nil, state, item, nil, nil, multiSelectRemove); changed {
		t.Fatal("RemoveKey semantics changed an absent value")
	}
	next, _, changed := m.toggleItem(nil, state, item, []string{"a"}, nil, multiSelectToggle)
	if !changed || len(next) != 0 {
		t.Fatalf("toggle selected value = %#v, changed=%v; want empty, true", next, changed)
	}
	if want := [][]string{{}}; !reflect.DeepEqual(received, want) {
		t.Fatalf("OnChange calls = %#v, want %#v", received, want)
	}
}

func TestMultiSelectRefUsesWorkingSelectionAndDisabledDrains(t *testing.T) {
	options := []ChoiceItem[string]{
		{Key: "alpha", Label: "Alpha", Value: "a"},
		{Key: "beta", Label: "Beta", Value: "b"},
	}

	t.Run("same frame commands", func(t *testing.T) {
		runtime := internal.NewRuntime(nil)
		defer runtime.Dispose()
		ref := NewMultiSelectRef[string]()
		var calls [][]string
		w := MultiSelect([]string{"a"}, options,
			MultiSelectAttachRef(ref),
			MultiSelectOnChange(func(_ *internal.Context, selected []string) {
				calls = append(calls, multiSelectCopy(selected))
			}),
		)
		layoutTagsForTest(runtime, w, 0)
		ref.RemoveKey("alpha")
		ref.SelectKey("beta")
		layoutTagsForTest(runtime, w, 1)

		if want := [][]string{{}, {"b"}}; !reflect.DeepEqual(calls, want) {
			t.Fatalf("same-frame Ref changes = %#v, want %#v", calls, want)
		}
	})

	t.Run("disabled drains without callback", func(t *testing.T) {
		runtime := internal.NewRuntime(nil)
		defer runtime.Dispose()
		ref := NewMultiSelectRef[string]()
		calls := 0
		w := MultiSelect([]string{"a"}, options,
			MultiSelectDisabled[string](true),
			MultiSelectAttachRef(ref),
			MultiSelectOnChange(func(_ *internal.Context, _ []string) { calls++ }),
		)
		layoutTagsForTest(runtime, w, 0)
		ref.RemoveKey("alpha")
		layoutTagsForTest(runtime, w, 1)

		if calls != 0 {
			t.Fatalf("disabled MultiSelect called OnChange %d times", calls)
		}
		if commands := ref.drainCommands(); len(commands) != 0 {
			t.Fatalf("disabled MultiSelect retained Ref commands: %#v", commands)
		}
	})
}

func TestMultiSelectSelectedKeysDistinguishEqualValues(t *testing.T) {
	options := []ChoiceItem[string]{
		{Key: "first", Label: "First", Value: "same"},
		{Key: "second", Label: "Second", Value: "same"},
	}
	var keyChanges [][]string
	var valueChanges [][]string
	m := &multiSelectWidget[string]{
		selected: []string{"same"},
		options:  options,
		config: multiSelectConfig[string]{
			hasSelectedKeys: true,
			selectedKeys:    []string{"first"},
			onSelectedKeysChange: func(_ *internal.Context, keys []string) {
				keyChanges = append(keyChanges, multiSelectKeyCopy(keys))
			},
			onChange: func(_ *internal.Context, values []string) {
				valueChanges = append(valueChanges, multiSelectCopy(values))
			},
		},
	}
	state := &multiSelectState{inputRef: NewInputRef()}
	nextValues, nextKeys, changed := m.toggleItem(nil, state, options[1], m.selected, m.config.selectedKeys, multiSelectToggle)
	if !changed {
		t.Fatal("key-controlled toggle did not select a distinct equal-value option")
	}
	if want := []string{"first", "second"}; !reflect.DeepEqual(nextKeys, want) {
		t.Fatalf("next key selection = %#v, want %#v", nextKeys, want)
	}
	if want := []string{"same"}; !reflect.DeepEqual(nextValues, want) {
		t.Fatalf("key-controlled toggle rewrote compatibility values = %#v, want %#v", nextValues, want)
	}
	if want := [][]string{{"first", "second"}}; !reflect.DeepEqual(keyChanges, want) {
		t.Fatalf("key callbacks = %#v, want %#v", keyChanges, want)
	}
	if want := [][]string{{"same", "same"}}; !reflect.DeepEqual(valueChanges, want) {
		t.Fatalf("value compatibility callbacks = %#v, want %#v", valueChanges, want)
	}

	// Selection and rendered chips continue to follow keys after option reorder.
	reordered := []ChoiceItem[string]{options[1], options[0]}
	chips := multiSelectSelectedItemsByKeys[string](nextKeys, reordered)
	if len(chips) != 2 || chips[0].Key != "first" || chips[1].Key != "second" {
		t.Fatalf("key-controlled chips lost selection on reorder: %#v", chips)
	}

	picker := TagPicker([]string{"second"}, "", []TagOptionItem{
		{Key: "first", Label: "First"},
		{Key: "second", Label: "Second"},
	})
	pickerWidget, ok := picker.(*multiSelectWidget[string])
	if !ok || !pickerWidget.config.hasSelectedKeys || !reflect.DeepEqual(pickerWidget.config.selectedKeys, []string{"second"}) {
		t.Fatalf("TagPicker did not use stable key selection: %#v", picker)
	}
}

func TestMultiSelectFirstArrowChoosesCollectionBoundary(t *testing.T) {
	items := []ChoiceItem[string]{
		{Key: "first", Value: "a"},
		{Key: "last", Value: "b"},
	}
	model, ok := choiceCollectionModel(items)
	if !ok {
		t.Fatal("test model rejected stable keys")
	}
	m := &multiSelectWidget[string]{config: defaultMultiSelectConfig[string]()}

	downState := &multiSelectState{}
	m.handleKey(nil, downState, items, model, true, &fluxevent.KeyboardEvent{Key: "ArrowDown"})
	if got := string(downState.active.Key); got != "first" {
		t.Fatalf("first ArrowDown active key = %q, want first", got)
	}

	upState := &multiSelectState{}
	m.handleKey(nil, upState, items, model, true, &fluxevent.KeyboardEvent{Key: "ArrowUp"})
	if got := string(upState.active.Key); got != "last" {
		t.Fatalf("first ArrowUp active key = %q, want last", got)
	}

	menu := multiSelectMenuItems(items, nil, nil, false, "last")
	if !menu[1].Active || menu[0].Active {
		t.Fatalf("active key was not forwarded to menu rows: %#v", menu)
	}
}

func TestMultiSelectRefOpenThenToggleComposesPendingState(t *testing.T) {
	runtime := internal.NewRuntime(nil)
	defer runtime.Dispose()
	ref := NewMultiSelectRef[string]()
	var opened []bool
	w := MultiSelect([]string(nil), []ChoiceItem[string]{{Key: "one", Value: "one"}},
		MultiSelectAttachRef(ref),
		MultiSelectOnOpenChange[string](func(_ *internal.Context, open bool) { opened = append(opened, open) }),
	)
	layoutTagsForTest(runtime, w, 0)
	ref.Open()
	ref.Toggle()
	layoutTagsForTest(runtime, w, 1)
	if want := []bool{true, false}; !reflect.DeepEqual(opened, want) {
		t.Fatalf("Open then Toggle callbacks = %#v, want %#v", opened, want)
	}
}

func TestMultiSelectInvalidKeyModelDoesNotActivateRows(t *testing.T) {
	runtime := internal.NewRuntime(nil)
	defer runtime.Dispose()
	ref := NewMultiSelectRef[string]()
	changes := 0
	options := []ChoiceItem[string]{
		{Key: "duplicate", Value: "first"},
		{Key: "duplicate", Value: "second"},
	}
	w := MultiSelect([]string(nil), options,
		MultiSelectOpened[string](true),
		MultiSelectAttachRef(ref),
		MultiSelectOnChange(func(_ *internal.Context, _ []string) { changes++ }),
	)
	layoutTagsForTest(runtime, w, 0)
	ref.SelectKey("duplicate")
	layoutTagsForTest(runtime, w, 1)
	if changes != 0 {
		t.Fatalf("invalid collection activated a Ref selection %d times", changes)
	}
}

func TestTagInputOpenedIntentsAreNoOpWithoutSuggestions(t *testing.T) {
	runtime := internal.NewRuntime(nil)
	defer runtime.Dispose()
	ref := NewTagInputRef()
	opened := 0
	w := TagInput(nil, "",
		TagInputAttachRef(ref),
		TagInputOnOpenChange(func(_ *internal.Context, _ bool) { opened++ }),
	)
	layoutTagsForTest(runtime, w, 0)
	ref.Open()
	ref.Toggle()
	ref.Close()
	layoutTagsForTest(runtime, w, 1)
	if opened != 0 {
		t.Fatalf("TagInput reported %d open changes without a popup", opened)
	}
}

func TestTagInputChoiceItemsAreUniqueStableKeys(t *testing.T) {
	items := tagInputChoiceItems([]string{"go", "rust", "go", ""})
	if len(items) != 2 {
		t.Fatalf("tag input items = %#v, want two unique non-empty tags", items)
	}
	if items[0].Key != "go" || items[0].Value != "go" || items[1].Key != "rust" {
		t.Fatalf("tag input keys are not stable values: %#v", items)
	}
}

func TestMultiSelectEditorKeyboardIntegration(t *testing.T) {
	h := newAdvancedInputHarness()
	defer h.runtime.Dispose()

	opened := true
	query := ""
	selectedKeys := []string(nil)
	options := []ChoiceItem[string]{
		{Key: "alpha", Label: "Alpha", Value: "alpha"},
		{Key: "bravo", Label: "Bravo", Value: "bravo"},
		{Key: "charlie", Label: "Charlie", Value: "charlie"},
	}
	var active []string
	var toggles []string
	var submits int
	var beforeSpacePrevented []bool

	render := func() Widget {
		return MultiSelect([]string(nil), options,
			MultiSelectSelectedKeys[string](selectedKeys),
			MultiSelectQuery[string](query),
			MultiSelectOpened[string](opened),
			MultiSelectOnOpenChange[string](func(_ *internal.Context, next bool) { opened = next }),
			MultiSelectOnQueryChange[string](func(_ *internal.Context, next string) { query = next }),
			MultiSelectOnSelectedKeysChange[string](func(_ *internal.Context, next []string) {
				selectedKeys = append([]string(nil), next...)
			}),
			MultiSelectOnActiveChange[string](func(_ *internal.Context, next string) { active = append(active, next) }),
			MultiSelectOnSubmit[string](func(_ *internal.Context, _ *fluxevent.InputEvent) { submits++ }),
			MultiSelectOnToggle[string](func(_ *internal.Context, item ChoiceItem[string], selected bool) {
				toggles = append(toggles, item.Key+":"+map[bool]string{true: "on", false: "off"}[selected])
			}),
			MultiSelectOnBeforeInput[string](func(_ *internal.Context, event *fluxevent.InputEvent) {
				if event != nil && event.Data == " " {
					beforeSpacePrevented = append(beforeSpacePrevented, event.DefaultPrevented)
				}
			}),
		)
	}

	_, root := h.render(t, render(), nil)
	multiState := multiSelectStateFor(internal.NewContext(root.Gtx, h.runtime))
	if multiState.inputRef == nil {
		t.Fatal("MultiSelect did not bind its backing InputRef")
	}
	scope := internal.NewContext(root.Gtx, h.runtime).Child(0).Scope("keyboard-scope")
	if !h.runtime.RequestFocus(root, scope.PathID()) {
		t.Fatal("failed to focus MultiSelect keyboard scope")
	}
	multiState.inputRef.Focus()
	h.render(t, render(), nil)

	h.render(t, render(), nil, key.Event{Name: key.NameDownArrow, State: key.Press})
	if got, want := active, []string{"bravo"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ArrowDown active changes = %#v, want %#v", got, want)
	}
	h.render(t, render(), nil, key.Event{Name: key.NameEnd, State: key.Press})
	if got := active[len(active)-1]; got != "charlie" {
		t.Fatalf("End active = %q, want charlie", got)
	}
	h.render(t, render(), nil, key.Event{Name: key.NameHome, State: key.Press})
	if got := active[len(active)-1]; got != "alpha" {
		t.Fatalf("Home active = %q, want alpha", got)
	}

	h.render(t, render(), nil, key.Event{Name: key.NameEnter, State: key.Press})
	if submits != 1 {
		t.Fatalf("Enter submit callbacks = %d, want 1", submits)
	}
	if got, want := selectedKeys, []string{"alpha"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Enter selected keys = %#v, want %#v", got, want)
	}
	if !opened {
		t.Fatal("MultiSelect Enter unexpectedly closed its menu")
	}

	h.render(t, render(), nil,
		key.Event{Name: key.NameSpace, State: key.Press},
		key.EditEvent{Range: key.Range{Start: 0, End: 0}, Text: " "},
	)
	if len(selectedKeys) != 0 {
		t.Fatalf("Space did not toggle active key off: %#v", selectedKeys)
	}
	if query != "" {
		t.Fatalf("Space selection wrote query = %q, want empty", query)
	}
	if got, want := beforeSpacePrevented, []bool{false}; !reflect.DeepEqual(got, want) {
		t.Fatalf("beforeinput space observations = %#v, want %#v before component cancellation", got, want)
	}
	if got, want := toggles, []string{"alpha:on", "alpha:off"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("toggle sequence = %#v, want %#v", got, want)
	}

	h.render(t, render(), nil, key.EditEvent{Range: key.Range{Start: 0, End: 0}, Text: "b"})
	if query != "b" {
		t.Fatalf("typeahead query = %q, want b", query)
	}
	h.render(t, render(), nil, key.Event{Name: key.NameEscape, State: key.Press})
	if opened {
		t.Fatal("Escape did not request MultiSelect close")
	}
}
