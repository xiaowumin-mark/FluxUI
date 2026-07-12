package widget

import (
	"image"
	"reflect"
	"testing"
	"time"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/internal/collection"

	gioEvent "gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/key"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

type advancedInputHarness struct {
	runtime *internal.Runtime
	router  input.Router
	ops     op.Ops
	now     time.Time
}

func newAdvancedInputHarness() *advancedInputHarness {
	return &advancedInputHarness{
		runtime: internal.NewRuntime(nil),
		now:     time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC),
	}
}

func (h *advancedInputHarness) render(t *testing.T, widget Widget, inputContext func(*internal.Context) *internal.Context, events ...gioEvent.Event) (*inputState, *internal.Context) {
	t.Helper()
	for _, event := range events {
		h.router.Queue(event)
	}
	h.ops.Reset()
	gtx := gioLayout.Context{
		Constraints: gioLayout.Exact(image.Pt(360, 260)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Now:         h.now,
		Ops:         &h.ops,
		Source:      h.router.Source(),
	}
	h.runtime.BeginFrame()
	ctx := internal.NewContext(gtx, h.runtime)
	widget.Layout(ctx)
	var state *inputState
	if inputContext != nil {
		state = inputStateFor(inputContext(internal.NewContext(gtx, h.runtime)))
	}
	h.runtime.EndFrame()
	h.router.Frame(&h.ops)
	h.now = h.now.Add(16 * time.Millisecond)
	return state, ctx
}

func advancedChoiceInputContext(ctx *internal.Context) *internal.Context {
	return ctx.Child(0).Scope("keyboard-scope").Child(0).Child(0)
}

func focusAdvancedInput(t *testing.T, h *advancedInputHarness, root *internal.Context, state *inputState) {
	t.Helper()
	if h == nil || root == nil || state == nil || state.editor == nil {
		t.Fatal("advanced input focus setup is incomplete")
	}
	scope := internal.NewContext(root.Gtx, h.runtime).Child(0).Scope("keyboard-scope")
	if !h.runtime.RequestFocus(root, scope.PathID()) {
		t.Fatal("failed to focus advanced picker keyboard scope")
	}
	h.router.Source().Execute(key.FocusCmd{Tag: state.editor})
}

func TestChoiceFilteringKeepsHostKeysAndSourceUntouched(t *testing.T) {
	items := []ChoiceItem[string]{
		{Key: "alpha", Label: "Alpha", Value: "a"},
		{Key: "bravo", Label: "Bravo", Value: "b", TypeaheadText: "second"},
	}
	got := filteredChoiceItems(items, "SECOND", true)
	if len(got) != 1 || got[0].Key != "bravo" {
		t.Fatalf("filtered items = %#v, want bravo", got)
	}
	if items[1].Key != "bravo" || items[1].Label != "Bravo" {
		t.Fatalf("source items were mutated: %#v", items)
	}

	missing := filteredChoiceItems([]ChoiceItem[int]{{Label: "One", Value: 1}}, "", true)
	if _, ok := choiceCollectionModel(missing); ok {
		t.Fatalf("missing stable key must not create an index/value identity fallback: %#v", missing)
	}
}

func TestChoiceCollectionRejectsDuplicateKeys(t *testing.T) {
	items := filteredChoiceItems([]ChoiceItem[int]{
		{Key: "duplicate", Label: "One", Value: 1},
		{Key: "duplicate", Label: "Two", Value: 2},
	}, "", false)
	if _, ok := choiceCollectionModel(items); ok {
		t.Fatal("duplicate keyed options must not share roving state")
	}
}

func TestSuggestionRovingFocusUsesKeysAcrossReorder(t *testing.T) {
	items := filteredChoiceItems([]ChoiceItem[string]{
		{Key: "a", Label: "Alpha", Value: "a"},
		{Key: "b", Label: "Bravo", Value: "b", Disabled: true},
		{Key: "c", Label: "Charlie", Value: "c"},
	}, "", false)
	model, ok := choiceCollectionModel(items)
	if !ok {
		t.Fatal("initial model should be valid")
	}
	state := &suggestionState{active: collection.RovingFocus{}}
	widget := &suggestionWidget[string]{config: suggestionConfig[string]{}}
	widget.handleKey(nil, state, items, model, true, &fluxevent.KeyboardEvent{Key: "ArrowDown"})
	if got := state.active.Key; got != "a" {
		t.Fatalf("first ArrowDown active = %q, want a", got)
	}
	widget.handleKey(nil, state, items, model, true, &fluxevent.KeyboardEvent{Key: "ArrowDown"})
	if got := state.active.Key; got != "c" {
		t.Fatalf("second ArrowDown active = %q, want c (skip disabled b)", got)
	}

	reordered := filteredChoiceItems([]ChoiceItem[string]{
		{Key: "c", Label: "Charlie", Value: "c"},
		{Key: "a", Label: "Alpha", Value: "a"},
		{Key: "b", Label: "Bravo", Value: "b", Disabled: true},
	}, "", false)
	reorderedModel, ok := choiceCollectionModel(reordered)
	if !ok {
		t.Fatal("reordered model should be valid")
	}
	state.active = state.active.Reconcile(reorderedModel)
	if got := state.active.Key; got != "c" {
		t.Fatalf("reorder transferred active state: got %q, want c", got)
	}
}

func TestSearchSelectCallbacksRemainSelectionOnly(t *testing.T) {
	var got string
	widget := SearchSelect("old", "", []SearchSelectItem[string]{
		{Key: "one", Label: "One", Value: "one"},
	}, SearchSelectOnChange(func(_ *internal.Context, value string) { got = value })).(*suggestionWidget[string])
	widget.selectItem(nil, &suggestionState{}, ChoiceItem[string]{Key: "one", Label: "One", Value: "one"})
	if got != "one" {
		t.Fatalf("selection callback = %q, want one", got)
	}
	if widget.config.allowCustom || widget.config.onCustomValue != nil {
		t.Fatalf("search select unexpectedly has a custom-value path: %#v", widget.config)
	}
}

func TestSearchSelectCanKeepDuplicateValuesBySelectedKey(t *testing.T) {
	widget := SearchSelect(1, "", []SearchSelectItem[int]{
		{Key: "first", Label: "First", Value: 1},
		{Key: "second", Label: "Second", Value: 1},
	}, SearchSelectSelectedKey[int]("second")).(*suggestionWidget[int])
	if widget.config.selectedKey != "second" {
		t.Fatalf("selected key = %q, want second", widget.config.selectedKey)
	}
}

func TestSuggestionRefOpenToggleComposesInOneFrame(t *testing.T) {
	requests := make([]bool, 0, 2)
	widget := &suggestionWidget[string]{config: suggestionConfig[string]{
		onOpenChange: func(_ *internal.Context, opened bool) { requests = append(requests, opened) },
	}}
	state := &suggestionState{}
	widget.requestOpen(nil, state, true)
	widget.requestOpen(nil, state, !widget.effectiveOpen(state))
	if want := []bool{true, false}; !reflect.DeepEqual(requests, want) {
		t.Fatalf("open/toggle requests = %#v, want %#v", requests, want)
	}
	if widget.effectiveOpen(state) {
		t.Fatal("Open followed by Toggle in the same frame should end closed")
	}
}

func TestComboboxEditorKeyboardIntegration(t *testing.T) {
	h := newAdvancedInputHarness()
	defer h.runtime.Dispose()

	opened := true
	query := ""
	items := []ComboboxItem[string]{
		{Key: "alpha", Label: "Alpha", Value: "alpha"},
		{Key: "bravo", Label: "Bravo", Value: "bravo"},
		{Key: "charlie", Label: "Charlie", Value: "charlie"},
	}
	var active []string
	var selected []string
	var sequence []string
	var beforeSpacePrevented []bool

	render := func() Widget {
		return Combobox(query, items,
			ComboboxOpened[string](opened),
			ComboboxOnOpenChange[string](func(_ *internal.Context, next bool) { opened = next }),
			ComboboxOnQueryChange[string](func(_ *internal.Context, next string) { query = next }),
			ComboboxOnActiveChange[string](func(_ *internal.Context, key string) { active = append(active, key) }),
			ComboboxOnSubmit[string](func(_ *internal.Context, _ *fluxevent.InputEvent) { sequence = append(sequence, "submit") }),
			ComboboxOnSelect[string](func(_ *internal.Context, item ComboboxItem[string]) {
				sequence = append(sequence, "select:"+item.Key)
				selected = append(selected, item.Key)
			}),
			ComboboxOnBeforeInput[string](func(_ *internal.Context, event *fluxevent.InputEvent) {
				if event != nil && event.Data == " " {
					beforeSpacePrevented = append(beforeSpacePrevented, event.DefaultPrevented)
				}
			}),
		)
	}

	state, root := h.render(t, render(), advancedChoiceInputContext)
	focusAdvancedInput(t, h, root, state)
	state, _ = h.render(t, render(), advancedChoiceInputContext)
	if !h.router.Source().Focused(state.editor) {
		t.Fatal("advanced Combobox editor did not retain Gio focus")
	}

	state, _ = h.render(t, render(), advancedChoiceInputContext, key.Event{Name: key.NameDownArrow, State: key.Press})
	if got, want := active, []string{"bravo"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ArrowDown active changes = %#v, want %#v", got, want)
	}
	_, _ = h.render(t, render(), advancedChoiceInputContext, key.Event{Name: key.NameEnd, State: key.Press})
	if got := active[len(active)-1]; got != "charlie" {
		t.Fatalf("End active = %q, want charlie", got)
	}
	_, _ = h.render(t, render(), advancedChoiceInputContext, key.Event{Name: key.NameHome, State: key.Press})
	if got := active[len(active)-1]; got != "alpha" {
		t.Fatalf("Home active = %q, want alpha", got)
	}

	_, _ = h.render(t, render(), advancedChoiceInputContext, key.Event{Name: key.NameEnter, State: key.Press})
	if got, want := sequence, []string{"submit", "select:alpha"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Enter callback sequence = %#v, want %#v", got, want)
	}
	if got, want := selected, []string{"alpha"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Enter selected = %#v, want %#v", got, want)
	}
	if opened {
		t.Fatal("Enter selection did not request close")
	}

	opened = true
	h.render(t, render(), advancedChoiceInputContext)
	_, _ = h.render(t, render(), advancedChoiceInputContext,
		key.Event{Name: key.NameSpace, State: key.Press},
		key.EditEvent{Range: key.Range{Start: 0, End: 0}, Text: " "},
	)
	if got, want := selected, []string{"alpha", "alpha"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Space selected = %#v, want %#v", got, want)
	}
	if query != "" {
		t.Fatalf("Space selection wrote query = %q, want empty", query)
	}
	if got, want := beforeSpacePrevented, []bool{false}; !reflect.DeepEqual(got, want) {
		t.Fatalf("beforeinput space observations = %#v, want %#v before component cancellation", got, want)
	}

	opened = true
	h.render(t, render(), advancedChoiceInputContext)
	_, _ = h.render(t, render(), advancedChoiceInputContext, key.EditEvent{Range: key.Range{Start: 0, End: 0}, Text: "b"})
	if query != "b" {
		t.Fatalf("typeahead query = %q, want b", query)
	}
	state, _ = h.render(t, render(), advancedChoiceInputContext, key.Event{Name: key.NameEscape, State: key.Press})
	if opened {
		t.Fatal("Escape did not request close")
	}
	state, _ = h.render(t, render(), advancedChoiceInputContext)
	if !h.router.Source().Focused(state.editor) {
		t.Fatal("Escape did not restore focus to the Combobox input")
	}

	opened = true
	h.render(t, render(), advancedChoiceInputContext)
	_, _ = h.render(t, render(), advancedChoiceInputContext, key.Event{Name: key.NameTab, State: key.Press})
	if opened {
		t.Fatal("Tab did not request close")
	}
	opened = true
	h.render(t, render(), advancedChoiceInputContext)
	_, _ = h.render(t, render(), advancedChoiceInputContext, key.Event{Name: key.NameTab, State: key.Press, Modifiers: key.ModShift})
	if opened {
		t.Fatal("Shift+Tab did not request close")
	}
}

func TestComboboxDisabledBlocksMenuAndClosesControlledOverlay(t *testing.T) {
	h := newAdvancedInputHarness()
	defer h.runtime.Dispose()

	opened := true
	disabled := false
	var requests []bool
	render := func() Widget {
		return Combobox("", []ComboboxItem[string]{{Key: "one", Label: "One", Value: "one"}},
			ComboboxOpened[string](opened),
			ComboboxDisabled[string](disabled),
			ComboboxOnOpenChange[string](func(_ *internal.Context, next bool) {
				requests = append(requests, next)
				opened = next
			}),
		)
	}

	h.render(t, render(), advancedChoiceInputContext)
	disabled = true
	h.render(t, render(), nil)
	if got, want := requests, []bool{false}; !reflect.DeepEqual(got, want) {
		t.Fatalf("disabled overlay-close requests = %#v, want %#v", got, want)
	}
	if opened {
		t.Fatal("disabled transition left controlled overlay open")
	}

	requests = nil
	opened = false
	h.render(t, render(), nil)
	// A disabled advanced picker renders its Input directly, never a clickable
	// DropdownMenu trigger, so a pointer cannot request or flash an overlay.
	h.router.Queue(key.Event{Name: key.NameDownArrow, State: key.Press})
	h.render(t, render(), nil)
	if len(requests) != 0 || opened {
		t.Fatalf("disabled picker reacted to keyboard: requests=%#v opened=%v", requests, opened)
	}
}
