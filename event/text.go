package event

import "github.com/xiaowumin-mark/FluxUI/internal"

const (
	BeforeInput       Type = internal.EventTypeBeforeInput
	Input             Type = internal.EventTypeInput
	Change            Type = internal.EventTypeChange
	Submit            Type = internal.EventTypeSubmit
	CompositionStart  Type = internal.EventTypeCompositionStart
	CompositionUpdate Type = internal.EventTypeCompositionUpdate
	CompositionEnd    Type = internal.EventTypeCompositionEnd
)

// InputSource identifies the best-known origin of a text edit.
type InputSource string

const (
	InputSourceUnknown      InputSource = ""
	InputSourceUser         InputSource = "user"
	InputSourceProgrammatic InputSource = "programmatic"
	InputSourcePaste        InputSource = "paste"
	InputSourceDelete       InputSource = "delete"
	InputSourceUndo         InputSource = "undo"
	InputSourceRedo         InputSource = "redo"
)

const (
	InputTypeInsertText            = "insertText"
	InputTypeInsertFromPaste       = "insertFromPaste"
	InputTypeInsertReplacement     = "insertReplacementText"
	InputTypeDeleteContent         = "deleteContent"
	InputTypeDeleteContentBackward = "deleteContentBackward"
	InputTypeHistoryUndo           = "historyUndo"
	InputTypeHistoryRedo           = "historyRedo"
	InputTypeInsertLineBreak       = "insertLineBreak"
	InputTypeProgrammaticSetText   = "programmaticSetText"
	InputTypeProgrammaticAppend    = "programmaticAppend"
	InputTypeProgrammaticClear     = "programmaticClear"
)

// InputEvent is FluxUI's browser-style text editing event.
//
// Gio currently exposes most editor changes after mutation. When BestEffort is
// true, fields such as InputType and Source were inferred from the before/after
// text values and beforeinput cancellation is implemented by rolling back.
type InputEvent struct {
	Event

	Data          string
	InputType     string
	IsComposing   bool
	Source        InputSource
	Value         string
	PreviousValue string
	BestEffort    bool
	Native        any
}

// CompositionEvent describes an IME composition lifecycle event.
type CompositionEvent struct {
	Event

	Data   string
	Native any
}

// InputHandler handles a typed text input event.
type InputHandler func(ctx *internal.Context, event *InputEvent)

// CompositionHandler handles a typed composition event.
type CompositionHandler func(ctx *internal.Context, event *CompositionEvent)

// OnInput registers a typed text input listener for the current context target.
func OnInput(ctx *internal.Context, eventType Type, handler InputHandler, opts ...ListenerOption) {
	if handler == nil {
		return
	}
	On(ctx, eventType, func(ctx *internal.Context, ev *Event) {
		if ev == nil {
			return
		}
		if inputEvent, ok := ev.Detail.(*InputEvent); ok {
			handler(ctx, inputEvent)
		}
	}, opts...)
}

// OnComposition registers a typed composition listener for the current context target.
func OnComposition(ctx *internal.Context, eventType Type, handler CompositionHandler, opts ...ListenerOption) {
	if handler == nil {
		return
	}
	On(ctx, eventType, func(ctx *internal.Context, ev *Event) {
		if ev == nil {
			return
		}
		if compositionEvent, ok := ev.Detail.(*CompositionEvent); ok {
			handler(ctx, compositionEvent)
		}
	}, opts...)
}

// DispatchInputEvent dispatches a typed text input event.
func DispatchInputEvent(ctx *internal.Context, target TargetID, ev *InputEvent) bool {
	if ev == nil {
		return true
	}
	applyInputDefaults(ev)
	ev.Event.Detail = ev
	return DispatchEvent(ctx, target, &ev.Event)
}

// DispatchCompositionEvent dispatches a typed composition event.
func DispatchCompositionEvent(ctx *internal.Context, target TargetID, ev *CompositionEvent) bool {
	if ev == nil {
		return true
	}
	applyCompositionDefaults(ev)
	ev.Event.Detail = ev
	return DispatchEvent(ctx, target, &ev.Event)
}

func applyInputDefaults(ev *InputEvent) {
	if ev.Type == "" {
		ev.Type = Input
	}
	ev.Bubbles = true
	switch ev.Type {
	case BeforeInput, Submit:
		ev.Cancelable = true
	default:
		ev.Cancelable = false
	}
}

func applyCompositionDefaults(ev *CompositionEvent) {
	if ev.Type == "" {
		ev.Type = CompositionUpdate
	}
	ev.Bubbles = true
	ev.Cancelable = false
}
