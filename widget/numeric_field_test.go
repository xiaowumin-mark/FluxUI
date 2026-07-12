package widget

import (
	"image"
	"reflect"
	"testing"
	"time"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	internal "github.com/xiaowumin-mark/FluxUI/internal"

	gioEvent "gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/key"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestParseNumericValuePreservesRawTextAndCanonicalDecimal(t *testing.T) {
	cases := []struct {
		text      string
		valid     bool
		canonical string
		errorText string
	}{
		{text: "001.2300", valid: true, canonical: "1.23"},
		{text: "-0.00", valid: true, canonical: "0"},
		{text: ".125", valid: true, canonical: "0.125"},
		{text: "1e-3", valid: true, canonical: "0.001"},
		{text: "", valid: false},
		{text: "-", valid: false, errorText: "Enter a valid number"},
		{text: "1/3", valid: false, errorText: "Enter a valid number"},
	}

	for _, tc := range cases {
		t.Run(tc.text, func(t *testing.T) {
			got := ParseNumericValue(tc.text)
			if got.Text != tc.text {
				t.Fatalf("raw text = %q, want %q", got.Text, tc.text)
			}
			if got.Valid != tc.valid || got.Value != tc.canonical || got.Error != tc.errorText {
				t.Fatalf("ParseNumericValue(%q) = %#v, want valid=%t canonical=%q error=%q", tc.text, got, tc.valid, tc.canonical, tc.errorText)
			}
		})
	}
}

func TestNumericValueAppliesIntegerAndBoundsWithoutLosingCanonicalValue(t *testing.T) {
	cfg := numericFieldConfig{}
	NumericFieldInteger(true)(&cfg)
	NumericFieldMin("0")(&cfg)
	NumericFieldMax("10")(&cfg)

	if got := parseNumericValue("2.5", cfg); got.Valid || got.Value != "2.5" || got.Error != "Enter a whole number" {
		t.Fatalf("fractional integer value = %#v", got)
	}
	if got := parseNumericValue("-1", cfg); got.Valid || got.Value != "-1" || got.Error != "Value must be at least 0" {
		t.Fatalf("below-min value = %#v", got)
	}
	if got := parseNumericValue("11", cfg); got.Valid || got.Value != "11" || got.Error != "Value must be at most 10" {
		t.Fatalf("above-max value = %#v", got)
	}
	if got := parseNumericValue("010", cfg); !got.Valid || got.Value != "10" || got.Error != "" {
		t.Fatalf("valid bounded integer = %#v", got)
	}
}

func TestSpinBoxSteppingUsesExactDecimalsAndBounds(t *testing.T) {
	cfg := numericFieldConfig{}
	NumericFieldMin("0")(&cfg)
	NumericFieldMax("0.3")(&cfg)
	NumericFieldStep("0.1")(&cfg)

	if got, changed := numericSteppedText("0.2", cfg, 1); !changed || got != "0.3" {
		t.Fatalf("0.2 + 0.1 = %q changed=%t, want 0.3/true", got, changed)
	}
	if got, changed := numericSteppedText("0.3", cfg, 1); changed || got != "0.3" {
		t.Fatalf("upper bound step = %q changed=%t, want 0.3/false", got, changed)
	}
	if got, changed := numericSteppedText("0.1", cfg, -1); !changed || got != "0" {
		t.Fatalf("0.1 - 0.1 = %q changed=%t, want 0/true", got, changed)
	}
	if got, changed := numericSteppedText("", cfg, 1); !changed || got != "0" {
		t.Fatalf("empty increment with min = %q changed=%t, want 0/true", got, changed)
	}
}

func TestNumericFieldRefForwardsCommandsAndHonorsDisabled(t *testing.T) {
	ref := NewNumericFieldRef()
	inputRef := NewInputRef()
	field := &numericFieldWidget{config: numericFieldConfig{ref: ref}}

	ref.SetValue("12.50")
	field.consumeRef(inputRef)
	commands := inputRef.drainCommands()
	if len(commands) != 1 || commands[0].kind != inputCmdSetText || commands[0].text != "12.50" {
		t.Fatalf("forwarded commands = %#v", commands)
	}

	ref.SetText("7")
	field.config.disabled = true
	field.consumeRef(inputRef)
	if commands := inputRef.drainCommands(); len(commands) != 0 {
		t.Fatalf("disabled NumericField forwarded commands = %#v", commands)
	}
}

func TestNumericFieldCallbacksDeliverRawThenParsedValue(t *testing.T) {
	var order []string
	field := &numericFieldWidget{config: numericFieldConfig{
		onChange: func(_ *internal.Context, value string) {
			order = append(order, "raw:"+value)
		},
		onParsedChange: func(_ *internal.Context, value NumericValue) {
			order = append(order, "parsed:"+value.Value)
		},
	}}
	field.dispatchChange(nil, "01.20")
	want := []string{"raw:01.20", "parsed:1.2"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("callback order = %#v, want %#v", order, want)
	}
}

func TestNumericFieldForwardsCancelableBeforeInputPasteAndIMECommit(t *testing.T) {
	runtime := internal.NewRuntime(nil)
	defer runtime.Dispose()
	harness := newNumericFieldInputHarness(runtime)

	value := ""
	var beforeEvents []fluxevent.InputEvent
	var inputEvents []fluxevent.InputEvent
	var changes []string
	var parsed []NumericValue
	opts := []NumericFieldOption{
		NumericFieldInputOptions(
			InputOnBeforeInput(func(_ *internal.Context, event *fluxevent.InputEvent) {
				if event.Data == "x" {
					event.PreventDefault()
				}
				beforeEvents = append(beforeEvents, *event)
			}),
			InputOnInputEvent(func(_ *internal.Context, event *fluxevent.InputEvent) {
				inputEvents = append(inputEvents, *event)
			}),
		),
		NumericFieldOnChange(func(_ *internal.Context, next string) {
			changes = append(changes, next)
			value = next
		}),
		NumericFieldOnParsedChange(func(_ *internal.Context, next NumericValue) {
			parsed = append(parsed, next)
		}),
	}

	state := harness.render(t, 0, value, opts)
	harness.focus(state)
	harness.render(t, 1, value, opts)

	// Multi-rune edits follow the backing Input's best-effort paste path.
	state = harness.render(t, 2, value, opts, key.EditEvent{
		Range: key.Range{Start: 0, End: 0},
		Text:  "12.50",
	})
	if got := state.editor.Text(); got != "12.50" {
		t.Fatalf("text after paste = %q, want 12.50", got)
	}
	if len(beforeEvents) != 1 || beforeEvents[0].Source != fluxevent.InputSourcePaste || beforeEvents[0].InputType != fluxevent.InputTypeInsertFromPaste || beforeEvents[0].Data != "12.50" {
		t.Fatalf("beforeinput after paste = %#v, want inferred paste event", beforeEvents)
	}
	if len(inputEvents) != 1 || inputEvents[0].Source != fluxevent.InputSourcePaste || inputEvents[0].InputType != fluxevent.InputTypeInsertFromPaste || inputEvents[0].Value != "12.50" {
		t.Fatalf("input after paste = %#v, want inferred paste event", inputEvents)
	}
	if !reflect.DeepEqual(changes, []string{"12.50"}) {
		t.Fatalf("raw changes after paste = %#v, want [12.50]", changes)
	}
	if len(parsed) != 1 || parsed[0].Text != "12.50" || !parsed[0].Valid || parsed[0].Value != "12.5" {
		t.Fatalf("parsed changes after paste = %#v, want valid canonical 12.5", parsed)
	}

	// Gio delivers committed IME text through key.EditEvent. NumericField must
	// preserve that raw text even when it is not a valid numeric value.
	state = harness.render(t, 3, value, opts, key.EditEvent{
		Range: key.Range{Start: 0, End: len("12.50")},
		Text:  "３",
	})
	if got := state.editor.Text(); got != "３" {
		t.Fatalf("text after IME commit = %q, want full-width 3", got)
	}
	if len(beforeEvents) != 2 || beforeEvents[1].Data != "３" || beforeEvents[1].Source != fluxevent.InputSourceUser || beforeEvents[1].IsComposing {
		t.Fatalf("beforeinput after IME commit = %#v, want committed user edit", beforeEvents)
	}
	if len(inputEvents) != 2 || inputEvents[1].Data != "３" || inputEvents[1].Value != "３" || inputEvents[1].IsComposing {
		t.Fatalf("input after IME commit = %#v, want committed raw text", inputEvents)
	}
	if !reflect.DeepEqual(changes, []string{"12.50", "３"}) {
		t.Fatalf("raw changes after IME commit = %#v", changes)
	}
	if len(parsed) != 2 || parsed[1].Text != "３" || parsed[1].Valid || parsed[1].Error != "Enter a valid number" {
		t.Fatalf("parsed IME commit = %#v, want invalid raw numeric value", parsed)
	}

	state = harness.render(t, 4, value, opts, key.EditEvent{
		Range: key.Range{Start: 1, End: 1},
		Text:  "x",
	})
	if got := state.editor.Text(); got != "３" {
		t.Fatalf("text after canceled beforeinput = %q, want preserved IME text", got)
	}
	if len(beforeEvents) != 3 || !beforeEvents[2].DefaultPrevented || beforeEvents[2].Data != "x" {
		t.Fatalf("canceled beforeinput = %#v, want canceled x event", beforeEvents)
	}
	if len(inputEvents) != 2 || !reflect.DeepEqual(changes, []string{"12.50", "３"}) || len(parsed) != 2 {
		t.Fatalf("canceled beforeinput dispatched downstream callbacks: input=%#v changes=%#v parsed=%#v", inputEvents, changes, parsed)
	}
}

func TestNumericFieldForwardsUndoAndRedo(t *testing.T) {
	runtime := internal.NewRuntime(nil)
	defer runtime.Dispose()
	harness := newNumericFieldInputHarness(runtime)

	value := ""
	var beforeEvents []fluxevent.InputEvent
	var inputEvents []fluxevent.InputEvent
	var changes []string
	var parsed []NumericValue
	opts := []NumericFieldOption{
		NumericFieldInputOptions(
			InputOnBeforeInput(func(_ *internal.Context, event *fluxevent.InputEvent) {
				beforeEvents = append(beforeEvents, *event)
			}),
			InputOnInputEvent(func(_ *internal.Context, event *fluxevent.InputEvent) {
				inputEvents = append(inputEvents, *event)
			}),
		),
		NumericFieldOnChange(func(_ *internal.Context, next string) {
			changes = append(changes, next)
			value = next
		}),
		NumericFieldOnParsedChange(func(_ *internal.Context, next NumericValue) {
			parsed = append(parsed, next)
		}),
	}

	state := harness.render(t, 0, value, opts)
	harness.focus(state)
	harness.render(t, 1, value, opts)
	state = harness.render(t, 2, value, opts, key.EditEvent{
		Range: key.Range{Start: 0, End: 0},
		Text:  "1",
	})
	if got := state.editor.Text(); got != "1" {
		t.Fatalf("text after first edit = %q, want 1", got)
	}
	state = harness.render(t, 3, value, opts, key.EditEvent{
		Range: key.Range{Start: 1, End: 1},
		Text:  "2",
	})
	if got := state.editor.Text(); got != "12" {
		t.Fatalf("text after second edit = %q, want 12", got)
	}

	state = harness.render(t, 4, value, opts, key.Event{
		Name:      "Z",
		Modifiers: key.ModShortcut,
		State:     key.Press,
	})
	if got := state.editor.Text(); got != "1" || value != "1" {
		t.Fatalf("undo text/value = %q/%q, want 1/1", got, value)
	}
	state = harness.render(t, 5, value, opts, key.Event{
		Name:      "Z",
		Modifiers: key.ModShortcut | key.ModShift,
		State:     key.Press,
	})
	if got := state.editor.Text(); got != "12" || value != "12" {
		t.Fatalf("redo text/value = %q/%q, want 12/12", got, value)
	}

	if got, want := inputEventKinds(inputEvents), []inputEventKind{
		{source: fluxevent.InputSourceUser, inputType: fluxevent.InputTypeInsertText, value: "1"},
		{source: fluxevent.InputSourceUser, inputType: fluxevent.InputTypeInsertText, value: "12"},
		{source: fluxevent.InputSourceUndo, inputType: fluxevent.InputTypeHistoryUndo, value: "1"},
		{source: fluxevent.InputSourceRedo, inputType: fluxevent.InputTypeHistoryRedo, value: "12"},
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("input event kinds = %#v, want %#v", got, want)
	}
	if got, want := inputEventKinds(beforeEvents), inputEventKinds(inputEvents); !reflect.DeepEqual(got, want) {
		t.Fatalf("beforeinput kinds = %#v, want %#v", got, want)
	}
	if got, want := changes, []string{"1", "12", "1", "12"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("raw changes = %#v, want %#v", got, want)
	}
	if len(parsed) != 4 || parsed[2].Text != "1" || parsed[3].Text != "12" || !parsed[2].Valid || !parsed[3].Valid {
		t.Fatalf("parsed undo/redo callbacks = %#v, want valid 1 then 12", parsed)
	}
}

func TestNumericFieldRefCommandsPrecedeSubmitCallbacks(t *testing.T) {
	runtime := internal.NewRuntime(nil)
	defer runtime.Dispose()
	harness := newNumericFieldInputHarness(runtime)

	ref := NewNumericFieldRef()
	value := ""
	var order []string
	opts := []NumericFieldOption{
		NumericFieldAttachRef(ref),
		NumericFieldInputOptions(
			InputOnBeforeInput(func(_ *internal.Context, event *fluxevent.InputEvent) {
				order = append(order, "before:"+event.InputType+":"+event.Value)
			}),
			InputOnInputEvent(func(_ *internal.Context, event *fluxevent.InputEvent) {
				order = append(order, "input:"+event.InputType+":"+event.Value)
			}),
			InputOnSubmit(func(_ *internal.Context, event *fluxevent.InputEvent) {
				order = append(order, "submit:"+event.Value)
			}),
		),
		NumericFieldOnChange(func(_ *internal.Context, next string) {
			order = append(order, "raw:"+next)
			value = next
		}),
		NumericFieldOnParsedChange(func(_ *internal.Context, next NumericValue) {
			order = append(order, "parsed:"+next.Value)
		}),
	}

	state := harness.render(t, 0, value, opts)
	ref.Focus()
	state = harness.render(t, 1, value, opts)
	if !harness.router.Source().Focused(state.editor) {
		t.Fatal("NumericFieldRef.Focus did not focus the backing input")
	}

	// The ref commands and submit event are delivered in one frame. Input's
	// programmatic mutations must finish before the later submit callback.
	ref.SetValue("12")
	ref.Append(".5")
	state = harness.render(t, 2, value, opts, key.EditEvent{
		Range: key.Range{Start: 4, End: 4},
		Text:  "\n",
	})
	if got := state.editor.Text(); got != "12.5" || value != "12.5" {
		t.Fatalf("ref text/value = %q/%q, want 12.5/12.5", got, value)
	}
	wantOrder := []string{
		"before:" + fluxevent.InputTypeProgrammaticSetText + ":12",
		"input:" + fluxevent.InputTypeProgrammaticSetText + ":12",
		"raw:12",
		"parsed:12",
		"before:" + fluxevent.InputTypeProgrammaticAppend + ":12.5",
		"input:" + fluxevent.InputTypeProgrammaticAppend + ":12.5",
		"raw:12.5",
		"parsed:12.5",
		"submit:12.5",
	}
	if !reflect.DeepEqual(order, wantOrder) {
		t.Fatalf("callback order = %#v, want %#v", order, wantOrder)
	}
}

func TestNumericFieldInputSemanticsHonorDisabled(t *testing.T) {
	runtime := internal.NewRuntime(nil)
	defer runtime.Dispose()
	harness := newNumericFieldInputHarness(runtime)

	ref := NewNumericFieldRef()
	value := "7"
	var beforeEvents []fluxevent.InputEvent
	var inputEvents []fluxevent.InputEvent
	var changes []string
	disabledOpts := []NumericFieldOption{
		NumericFieldDisabled(true),
		NumericFieldAttachRef(ref),
		// NumericFieldDisabled is authoritative over forwarded Input options.
		NumericFieldInputOptions(
			InputDisabled(false),
			InputOnBeforeInput(func(_ *internal.Context, event *fluxevent.InputEvent) {
				beforeEvents = append(beforeEvents, *event)
			}),
			InputOnInputEvent(func(_ *internal.Context, event *fluxevent.InputEvent) {
				inputEvents = append(inputEvents, *event)
			}),
		),
		NumericFieldOnChange(func(_ *internal.Context, next string) {
			changes = append(changes, next)
			value = next
		}),
	}

	state := harness.render(t, 0, value, disabledOpts)
	if !state.editor.ReadOnly {
		t.Fatal("disabled NumericField left backing editor writable")
	}
	harness.focus(state)
	harness.render(t, 1, value, disabledOpts)
	ref.SetValue("12")
	ref.Append(".5")
	ref.Clear()
	state = harness.render(t, 2, value, disabledOpts, key.EditEvent{
		Range: key.Range{Start: 1, End: 1},
		Text:  "8",
	})
	if got := state.editor.Text(); got != "7" || value != "7" {
		t.Fatalf("disabled text/value = %q/%q, want 7/7", got, value)
	}
	if len(beforeEvents) != 0 || len(inputEvents) != 0 || len(changes) != 0 {
		t.Fatalf("disabled field dispatched callbacks: before=%#v input=%#v changes=%#v", beforeEvents, inputEvents, changes)
	}

	// Commands discarded while disabled must not become stale work if the host
	// reenables the controlled field in a later frame.
	enabledOpts := append([]NumericFieldOption(nil), disabledOpts...)
	enabledOpts[0] = NumericFieldDisabled(false)
	state = harness.render(t, 3, value, enabledOpts)
	if got := state.editor.Text(); got != "7" || value != "7" {
		t.Fatalf("reenabled field replayed disabled ref commands: text/value = %q/%q", got, value)
	}
}

type inputEventKind struct {
	source    fluxevent.InputSource
	inputType string
	value     string
}

func inputEventKinds(events []fluxevent.InputEvent) []inputEventKind {
	kinds := make([]inputEventKind, len(events))
	for index, event := range events {
		kinds[index] = inputEventKind{
			source:    event.Source,
			inputType: event.InputType,
			value:     event.Value,
		}
	}
	return kinds
}

type numericFieldInputHarness struct {
	runtime *internal.Runtime
	router  input.Router
	ops     op.Ops
}

func newNumericFieldInputHarness(runtime *internal.Runtime) *numericFieldInputHarness {
	return &numericFieldInputHarness{runtime: runtime}
}

func (h *numericFieldInputHarness) render(t *testing.T, frame int, value string, options []NumericFieldOption, events ...gioEvent.Event) *inputState {
	t.Helper()
	for _, event := range events {
		h.router.Queue(event)
	}
	h.ops.Reset()
	gtx := gioLayout.Context{
		Constraints: gioLayout.Exact(image.Pt(320, 72)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Now:         time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC).Add(time.Duration(frame) * 16 * time.Millisecond),
		Source:      h.router.Source(),
		Ops:         &h.ops,
	}
	h.runtime.BeginFrame()
	ctx := internal.NewContext(gtx, h.runtime).Scope("numeric-field-input")
	field := NumericField(value, options...).(*numericFieldWidget)
	field.Layout(ctx)
	stateCtx := internal.NewContext(gtx, h.runtime).Scope("numeric-field-input")
	if field.config.ref != nil {
		// NumericField's private backing InputRef is memoized immediately before
		// TextField obtains its state, so consume the same memory slot here.
		numericFieldInputRefFor(stateCtx)
	}
	state := inputStateFor(stateCtx)
	h.runtime.EndFrame()
	h.router.Frame(&h.ops)
	return state
}

func (h *numericFieldInputHarness) focus(state *inputState) {
	if h == nil || state == nil || state.editor == nil {
		return
	}
	h.router.Source().Execute(key.FocusCmd{Tag: state.editor})
}
