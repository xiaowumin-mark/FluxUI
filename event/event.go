package event

import "github.com/xiaowumin-mark/FluxUI/internal"

// TargetID identifies a FluxUI event target in the current runtime tree.
type TargetID = internal.PathID

// Type identifies an event kind.
type Type = internal.EventType

const (
	// Click is the semantic activation event used by existing clickable widgets.
	Click Type = "click"
	// Activate is the semantic activation event reserved for accessible default
	// actions and keyboard-equivalent activation.
	Activate Type = internal.EventTypeActivate
)

// Phase identifies the current event dispatch phase.
type Phase = internal.EventPhase

const (
	PhaseNone    Phase = internal.EventPhaseNone
	PhaseCapture Phase = internal.EventPhaseCapture
	PhaseTarget  Phase = internal.EventPhaseTarget
	PhaseBubble  Phase = internal.EventPhaseBubble
)

// Event is the base event object dispatched through the FluxUI event system.
type Event = internal.Event

// ListenerOptions configures listener dispatch behavior.
type ListenerOptions = internal.EventListenerOptions

// Handler handles an event for the listener's target context.
type Handler = internal.EventHandler

// ListenerOption mutates listener options.
type ListenerOption func(*ListenerOptions)

// Capture registers a listener for the capture phase.
func Capture() ListenerOption {
	return func(opts *ListenerOptions) {
		opts.Capture = true
	}
}

// Once removes the listener after its first dispatch in the current frame.
func Once() ListenerOption {
	return func(opts *ListenerOptions) {
		opts.Once = true
	}
}

// Passive marks a listener as unable to cancel the event default action.
func Passive() ListenerOption {
	return func(opts *ListenerOptions) {
		opts.Passive = true
	}
}

// Priority sets listener ordering for listeners on the same target and phase.
func Priority(priority int) ListenerOption {
	return func(opts *ListenerOptions) {
		opts.Priority = priority
	}
}

// BoundaryMode controls how propagation continues after an event boundary.
type BoundaryMode = internal.EventBoundaryMode

const (
	BoundaryNone     BoundaryMode = internal.EventBoundaryNone
	BoundaryStop     BoundaryMode = internal.EventBoundaryStop
	BoundaryRedirect BoundaryMode = internal.EventBoundaryRedirect
)

// BoundaryPolicy describes propagation behavior at a boundary target.
type BoundaryPolicy = internal.EventBoundaryPolicy

// BoundaryOption mutates boundary registration options.
type BoundaryOption func(*BoundaryPolicy)

// BoundaryStopPropagation stops propagation at the registered boundary target.
func BoundaryStopPropagation() BoundaryOption {
	return func(policy *BoundaryPolicy) {
		policy.Mode = BoundaryStop
		policy.Redirect = 0
	}
}

// BoundaryRedirectTo redirects propagation from the boundary target to target.
func BoundaryRedirectTo(target TargetID) BoundaryOption {
	return func(policy *BoundaryPolicy) {
		policy.Mode = BoundaryRedirect
		policy.Redirect = target
	}
}

// On registers an event listener for the current context target.
func On(ctx *internal.Context, eventType Type, handler Handler, opts ...ListenerOption) {
	if ctx == nil || ctx.Runtime() == nil {
		return
	}
	options := ListenerOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	ctx.Runtime().RegisterEventListener(ctx, eventType, handler, options)
}

// RegisterBoundary marks ctx as an event boundary for the current frame.
func RegisterBoundary(ctx *internal.Context, opts ...BoundaryOption) {
	if ctx == nil || ctx.Runtime() == nil {
		return
	}
	policy := BoundaryPolicy{Mode: BoundaryStop}
	for _, opt := range opts {
		if opt != nil {
			opt(&policy)
		}
	}
	ctx.Runtime().RegisterEventBoundary(ctx, policy)
}

// RegisterPortal makes ctx's logical event parent owner for the current frame.
// This is used by overlay and portal roots whose layout parent differs from the
// component owner.
func RegisterPortal(ctx *internal.Context, owner TargetID) {
	if ctx == nil || ctx.Runtime() == nil {
		return
	}
	ctx.Runtime().RegisterEventPortal(ctx, owner)
}

// Dispatch dispatches an event value to target. It returns false when the event
// is cancelable and a listener prevented the default action.
func Dispatch(ctx *internal.Context, target TargetID, ev Event) bool {
	return DispatchEvent(ctx, target, &ev)
}

// DispatchEvent dispatches an event pointer to target and preserves event
// mutations such as DefaultPrevented and ComposedPath for the caller.
func DispatchEvent(ctx *internal.Context, target TargetID, ev *Event) bool {
	if ev == nil {
		return true
	}
	if ctx == nil || ctx.Runtime() == nil {
		return !(ev.Cancelable && ev.DefaultPrevented)
	}
	if target == 0 {
		target = ctx.PathID()
	}
	return ctx.Runtime().DispatchEvent(ctx, target, ev)
}
