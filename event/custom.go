package event

import "github.com/xiaowumin-mark/FluxUI/internal"

// CustomEventConfig configures a synthetic application-defined event.
type CustomEventConfig struct {
	Bubbles    bool
	Cancelable bool
	Trusted    bool
}

// CustomEventOption mutates custom event creation.
type CustomEventOption func(*CustomEventConfig)

// CustomBubbles sets whether the custom event bubbles. FluxUI defaults custom
// events to bubbling so parent components can observe component communication.
func CustomBubbles(bubbles bool) CustomEventOption {
	return func(cfg *CustomEventConfig) {
		cfg.Bubbles = bubbles
	}
}

// CustomCancelable sets whether PreventDefault can cancel the event.
func CustomCancelable(cancelable bool) CustomEventOption {
	return func(cfg *CustomEventConfig) {
		cfg.Cancelable = cancelable
	}
}

// CustomTrusted marks a custom event as trusted. Application-created events
// should normally keep the default false value.
func CustomTrusted(trusted bool) CustomEventOption {
	return func(cfg *CustomEventConfig) {
		cfg.Trusted = trusted
	}
}

// NewCustomEvent creates a synthetic application-defined event. The payload is
// exposed through Event.Detail and is not interpreted by FluxUI.
func NewCustomEvent(eventType Type, detail any, opts ...CustomEventOption) Event {
	cfg := CustomEventConfig{
		Bubbles: true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return Event{
		Type:       eventType,
		Bubbles:    cfg.Bubbles,
		Cancelable: cfg.Cancelable,
		Trusted:    cfg.Trusted,
		Detail:     detail,
	}
}

// DispatchCustomEvent creates and dispatches a synthetic application-defined
// event to target.
func DispatchCustomEvent(ctx *internal.Context, target TargetID, eventType Type, detail any, opts ...CustomEventOption) bool {
	ev := NewCustomEvent(eventType, detail, opts...)
	return DispatchEvent(ctx, target, &ev)
}

// ActivationSource describes what caused an activation event.
type ActivationSource string

const (
	ActivationSourceUnknown      ActivationSource = ""
	ActivationSourcePointer      ActivationSource = "pointer"
	ActivationSourceKeyboard     ActivationSource = "keyboard"
	ActivationSourceProgrammatic ActivationSource = "programmatic"
)

// ActivationEvent reserves a semantic activation event for accessibility and
// keyboard-equivalent default actions.
type ActivationEvent struct {
	Event

	Source             ActivationSource
	KeyboardEquivalent string
	Native             any
}

// ActivationHandler handles a typed activation event.
type ActivationHandler func(ctx *internal.Context, event *ActivationEvent)

// OnActivate registers a typed activation listener for the current target.
func OnActivate(ctx *internal.Context, handler ActivationHandler, opts ...ListenerOption) {
	if handler == nil {
		return
	}
	On(ctx, Activate, func(ctx *internal.Context, ev *Event) {
		if activationEvent, ok := ev.Detail.(*ActivationEvent); ok {
			handler(ctx, activationEvent)
		}
	}, opts...)
}

// DispatchActivationEvent dispatches a typed activation event.
func DispatchActivationEvent(ctx *internal.Context, target TargetID, ev *ActivationEvent) bool {
	if ev == nil {
		return true
	}
	if ev.Type == "" {
		ev.Type = Activate
	}
	ev.Bubbles = true
	ev.Cancelable = true
	ev.Event.Detail = ev
	return DispatchEvent(ctx, target, &ev.Event)
}
