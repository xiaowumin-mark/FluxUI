package widget

import (
	"image"
	"strings"
	"testing"
	"time"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"

	gioEvent "gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/key"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestInputRefDispatchesProgrammaticInputEvents(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	ref := NewInputRef()
	ref.SetText("42")

	var before fluxevent.InputEvent
	var input fluxevent.InputEvent
	var changes []string

	var ops op.Ops
	gtx := gioLayout.Context{
		Constraints: gioLayout.Exact(image.Pt(180, 56)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Now:         time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC),
		Ops:         &ops,
	}
	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt).Scope("input")
	TextField("",
		InputAttachRef(ref),
		InputOnBeforeInput(func(ctx *internal.Context, ev *fluxevent.InputEvent) {
			before = *ev
		}),
		InputOnInputEvent(func(ctx *internal.Context, ev *fluxevent.InputEvent) {
			input = *ev
		}),
		InputOnChange(func(ctx *internal.Context, value string) {
			changes = append(changes, value)
		}),
	).Layout(ctx)
	rt.EndFrame()

	if before.Source != fluxevent.InputSourceProgrammatic || before.InputType != fluxevent.InputTypeProgrammaticSetText || before.Value != "42" {
		t.Fatalf("beforeinput = %+v, want programmatic setText to 42", before)
	}
	if input.Source != fluxevent.InputSourceProgrammatic || input.Value != "42" {
		t.Fatalf("input = %+v, want programmatic value 42", input)
	}
	if len(changes) != 1 || changes[0] != "42" {
		t.Fatalf("legacy changes = %v, want [42]", changes)
	}
}

func TestInputRefSameValueDoesNotDispatchProgrammaticChange(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	ref := NewInputRef()
	var beforeEvents []fluxevent.InputEvent
	var inputEvents []fluxevent.InputEvent
	var changes []string
	var ops op.Ops

	render := func(frame int) {
		ops.Reset()
		gtx := gioLayout.Context{
			Constraints: gioLayout.Exact(image.Pt(180, 56)),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC).Add(time.Duration(frame) * 16 * time.Millisecond),
			Ops:         &ops,
		}
		rt.BeginFrame()
		ctx := internal.NewContext(gtx, rt).Scope("input")
		TextField("42",
			InputAttachRef(ref),
			InputOnBeforeInput(func(ctx *internal.Context, ev *fluxevent.InputEvent) {
				beforeEvents = append(beforeEvents, *ev)
			}),
			InputOnInputEvent(func(ctx *internal.Context, ev *fluxevent.InputEvent) {
				inputEvents = append(inputEvents, *ev)
			}),
			InputOnChange(func(ctx *internal.Context, value string) {
				changes = append(changes, value)
			}),
		).Layout(ctx)
		rt.EndFrame()
	}

	render(0)
	ref.SetText("42")
	render(1)

	if len(beforeEvents) != 0 {
		t.Fatalf("same-value beforeinput events = %d, want 0", len(beforeEvents))
	}
	if len(inputEvents) != 0 {
		t.Fatalf("same-value input events = %d, want 0", len(inputEvents))
	}
	if len(changes) != 0 {
		t.Fatalf("same-value legacy changes = %v, want none", changes)
	}
}

func TestInputBeforeInputCancelRollsBackUserEdit(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	baseTime := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)

	var beforeEvents []fluxevent.InputEvent
	var inputEvents []fluxevent.InputEvent
	var changes []string

	opts := []InputOption{
		InputOnBeforeInput(func(ctx *internal.Context, ev *fluxevent.InputEvent) {
			if strings.ContainsAny(ev.Data, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
				ev.PreventDefault()
			}
			beforeEvents = append(beforeEvents, *ev)
		}),
		InputOnInputEvent(func(ctx *internal.Context, ev *fluxevent.InputEvent) {
			inputEvents = append(inputEvents, *ev)
		}),
		InputOnChange(func(ctx *internal.Context, value string) {
			changes = append(changes, value)
		}),
	}

	render := func(frame int, events ...gioEvent.Event) *inputState {
		for _, ev := range events {
			router.Queue(ev)
		}
		ops.Reset()
		gtx := gioLayout.Context{
			Constraints: gioLayout.Exact(image.Pt(180, 56)),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         baseTime.Add(time.Duration(frame) * 16 * time.Millisecond),
			Source:      router.Source(),
			Ops:         &ops,
		}
		rt.BeginFrame()
		ctx := internal.NewContext(gtx, rt).Scope("input")
		TextField("", opts...).Layout(ctx)
		state := inputStateFor(internal.NewContext(gtx, rt).Scope("input"))
		rt.EndFrame()
		router.Frame(&ops)
		return state
	}

	state := render(0)
	router.Source().Execute(key.FocusCmd{Tag: state.editor})
	render(1)
	state = render(2, key.EditEvent{Range: key.Range{Start: 0, End: 0}, Text: "a"})

	if len(beforeEvents) != 1 || beforeEvents[0].Data != "a" || !beforeEvents[0].DefaultPrevented {
		t.Fatalf("beforeinput events after rejected edit = %+v, want canceled data a", beforeEvents)
	}
	if got := state.editor.Text(); got != "" {
		t.Fatalf("editor text after canceled beforeinput = %q, want empty", got)
	}
	if len(inputEvents) != 0 {
		t.Fatalf("input events after canceled beforeinput = %d, want 0", len(inputEvents))
	}
	if len(changes) != 0 {
		t.Fatalf("legacy changes after canceled beforeinput = %v, want none", changes)
	}

	state = render(3, key.EditEvent{Range: key.Range{Start: 0, End: 0}, Text: "5"})
	if got := state.editor.Text(); got != "5" {
		t.Fatalf("editor text after accepted digit = %q, want 5", got)
	}
	if len(inputEvents) != 1 || inputEvents[0].Data != "5" || inputEvents[0].InputType != fluxevent.InputTypeInsertText {
		t.Fatalf("input events after accepted digit = %+v, want insertText data 5", inputEvents)
	}
	if len(changes) != 1 || changes[0] != "5" {
		t.Fatalf("legacy changes after accepted digit = %v, want [5]", changes)
	}
}

func TestInputSubmitEvent(t *testing.T) {
	rt := internal.NewRuntime(nil)
	defer rt.Dispose()

	var router input.Router
	var ops op.Ops
	baseTime := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	var submits []fluxevent.InputEvent

	opts := []InputOption{
		InputOnSubmit(func(ctx *internal.Context, ev *fluxevent.InputEvent) {
			submits = append(submits, *ev)
		}),
	}

	render := func(frame int, events ...gioEvent.Event) *inputState {
		for _, ev := range events {
			router.Queue(ev)
		}
		ops.Reset()
		gtx := gioLayout.Context{
			Constraints: gioLayout.Exact(image.Pt(180, 56)),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         baseTime.Add(time.Duration(frame) * 16 * time.Millisecond),
			Source:      router.Source(),
			Ops:         &ops,
		}
		rt.BeginFrame()
		ctx := internal.NewContext(gtx, rt).Scope("input")
		TextField("", opts...).Layout(ctx)
		state := inputStateFor(internal.NewContext(gtx, rt).Scope("input"))
		rt.EndFrame()
		router.Frame(&ops)
		return state
	}

	state := render(0)
	router.Source().Execute(key.FocusCmd{Tag: state.editor})
	render(1)
	state = render(2, key.EditEvent{Range: key.Range{Start: 0, End: 0}, Text: "12\n"})

	if got := state.editor.Text(); got != "12" {
		t.Fatalf("editor text after submit edit = %q, want 12", got)
	}
	if len(submits) != 1 {
		t.Fatalf("submit events = %d, want 1", len(submits))
	}
	if submits[0].Type != fluxevent.Submit || submits[0].Value != "12" || submits[0].InputType != fluxevent.InputTypeInsertLineBreak {
		t.Fatalf("submit event = %+v, want submit value 12", submits[0])
	}
	if submits[0].IsComposing {
		t.Fatal("submit event IsComposing = true, want false for Gio non-composition edit")
	}
}
