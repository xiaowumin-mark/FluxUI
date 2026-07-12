package event

import (
	"github.com/xiaowumin-mark/FluxUI/internal"

	"gioui.org/io/key"
)

const (
	Focus    Type = internal.EventTypeFocus
	Blur     Type = internal.EventTypeBlur
	FocusIn  Type = internal.EventTypeFocusIn
	FocusOut Type = internal.EventTypeFocusOut
	KeyDown  Type = internal.EventTypeKeyDown
	KeyUp    Type = internal.EventTypeKeyUp
)

// KeyLocation follows the DOM KeyboardEvent location constants.
type KeyLocation = internal.KeyLocation

const (
	KeyLocationStandard KeyLocation = internal.KeyLocationStandard
	KeyLocationLeft     KeyLocation = internal.KeyLocationLeft
	KeyLocationRight    KeyLocation = internal.KeyLocationRight
	KeyLocationNumpad   KeyLocation = internal.KeyLocationNumpad
)

// FocusDirection identifies tab-order focus movement.
type FocusDirection = internal.FocusDirection

const (
	FocusForward  FocusDirection = internal.FocusForward
	FocusBackward FocusDirection = internal.FocusBackward
)

// FocusEvent is the browser-style focus event payload.
type FocusEvent = internal.FocusEvent

// KeyboardEvent is the browser-style keyboard event payload.
type KeyboardEvent = internal.KeyboardEvent

// FocusHandler handles a typed focus event.
type FocusHandler func(ctx *internal.Context, event *FocusEvent)

// KeyboardHandler handles a typed keyboard event.
type KeyboardHandler = internal.KeyboardHandler

// FocusActivation is the default activation action for focused targets.
type FocusActivation = internal.FocusActivation

// FocusTargetOptions configures a focusable target in the current frame.
type FocusTargetOptions = internal.FocusTargetOptions

// FocusTargetOption mutates focus target options.
type FocusTargetOption func(*FocusTargetOptions)

// FocusTabIndex sets the target's tab order index. Negative values keep the
// target programmatically focusable but remove it from tab navigation.
func FocusTabIndex(index int) FocusTargetOption {
	return func(opts *FocusTargetOptions) {
		opts.TabIndex = index
	}
}

// FocusDisabled marks the target disabled for focus purposes.
func FocusDisabled(disabled bool) FocusTargetOption {
	return func(opts *FocusTargetOptions) {
		opts.Disabled = disabled
	}
}

// FocusHidden marks the target hidden for focus purposes.
func FocusHidden(hidden bool) FocusTargetOption {
	return func(opts *FocusTargetOptions) {
		opts.Hidden = hidden
	}
}

// FocusActivate registers the target's keyboard default activation action.
func FocusActivate(fn FocusActivation) FocusTargetOption {
	return func(opts *FocusTargetOptions) {
		opts.Activate = fn
	}
}

// RegisterFocusTarget records ctx as a focus target for the current frame.
func RegisterFocusTarget(ctx *internal.Context, opts ...FocusTargetOption) {
	if ctx == nil || ctx.Runtime() == nil {
		return
	}
	options := FocusTargetOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	ctx.Runtime().RegisterFocusTarget(ctx, options)
}

// FocusManager is a small facade over the runtime focus state.
type FocusManager struct {
	ctx *internal.Context
}

// FocusManagerFor returns a component-tree focus manager bound to ctx.
func FocusManagerFor(ctx *internal.Context) FocusManager {
	return FocusManager{ctx: ctx}
}

// Target returns the current focused target.
func (m FocusManager) Target() TargetID {
	if m.ctx == nil || m.ctx.Runtime() == nil {
		return 0
	}
	return m.ctx.Runtime().FocusedTarget()
}

// Request moves focus to target.
func (m FocusManager) Request(target TargetID) bool {
	if m.ctx == nil || m.ctx.Runtime() == nil {
		return false
	}
	return m.ctx.Runtime().RequestFocus(m.ctx, target)
}

// Blur clears the current focus target.
func (m FocusManager) Blur() bool {
	if m.ctx == nil || m.ctx.Runtime() == nil {
		return false
	}
	return m.ctx.Runtime().BlurFocus(m.ctx, 0)
}

// Move advances focus through the current frame's tab order.
func (m FocusManager) Move(direction FocusDirection) bool {
	if m.ctx == nil || m.ctx.Runtime() == nil {
		return false
	}
	return m.ctx.Runtime().MoveFocus(m.ctx, direction)
}

// RequestFocus moves component-tree focus to ctx's target.
func RequestFocus(ctx *internal.Context) bool {
	if ctx == nil || ctx.Runtime() == nil {
		return false
	}
	return ctx.Runtime().RequestFocus(ctx, ctx.PathID())
}

// BlurFocus clears focus when ctx owns it.
func BlurFocus(ctx *internal.Context) bool {
	if ctx == nil || ctx.Runtime() == nil {
		return false
	}
	return ctx.Runtime().BlurFocus(ctx, ctx.PathID())
}

// Focused reports whether ctx is the current component-tree focus target.
func Focused(ctx *internal.Context) bool {
	if ctx == nil || ctx.Runtime() == nil {
		return false
	}
	return ctx.Runtime().Focused(ctx.PathID())
}

// FocusedTarget returns the current component-tree focus target.
func FocusedTarget(ctx *internal.Context) TargetID {
	if ctx == nil || ctx.Runtime() == nil {
		return 0
	}
	return ctx.Runtime().FocusedTarget()
}

// MoveFocus moves focus through the current frame's tab order.
func MoveFocus(ctx *internal.Context, direction FocusDirection) bool {
	if ctx == nil || ctx.Runtime() == nil {
		return false
	}
	return ctx.Runtime().MoveFocus(ctx, direction)
}

// OnFocus registers a typed focus listener for the current context target.
func OnFocus(ctx *internal.Context, eventType Type, handler FocusHandler, opts ...ListenerOption) {
	if handler == nil {
		return
	}
	On(ctx, eventType, func(ctx *internal.Context, ev *Event) {
		if ev == nil {
			return
		}
		if focusEvent, ok := ev.Detail.(*FocusEvent); ok {
			handler(ctx, focusEvent)
		}
	}, opts...)
}

// OnKeyboard registers a typed keyboard listener for the current context target.
func OnKeyboard(ctx *internal.Context, eventType Type, handler KeyboardHandler, opts ...ListenerOption) {
	if handler == nil {
		return
	}
	On(ctx, eventType, func(ctx *internal.Context, ev *Event) {
		if ev == nil {
			return
		}
		if keyboardEvent, ok := ev.Detail.(*KeyboardEvent); ok {
			handler(ctx, keyboardEvent)
		}
	}, opts...)
}

// OnKeyDown registers a keydown listener.
func OnKeyDown(ctx *internal.Context, handler KeyboardHandler, opts ...ListenerOption) {
	OnKeyboard(ctx, KeyDown, handler, opts...)
}

// OnKeyUp registers a keyup listener.
func OnKeyUp(ctx *internal.Context, handler KeyboardHandler, opts ...ListenerOption) {
	OnKeyboard(ctx, KeyUp, handler, opts...)
}

// Shortcut describes a local component-tree shortcut.
type Shortcut = internal.Shortcut

// ShortcutKey creates a shortcut that matches a browser-style key value.
func ShortcutKey(key string, modifiers Modifiers) Shortcut {
	return Shortcut{Key: key, Modifiers: modifiers}
}

// ShortcutCode creates a shortcut that matches a browser-style code value.
func ShortcutCode(code string, modifiers Modifiers) Shortcut {
	return Shortcut{Code: code, Modifiers: modifiers}
}

// ShortcutExactModifiers requires no additional modifiers beyond those in spec.
func ShortcutExactModifiers(spec Shortcut) Shortcut {
	spec.ExactModifiers = true
	return spec
}

// ShortcutScope restricts a shortcut to a specific focus scope.
func ShortcutScope(spec Shortcut, scope TargetID) Shortcut {
	spec.Scope = scope
	return spec
}

// OnShortcut registers a local shortcut listener. By default the listener is
// scoped to ctx's target and fires only when the focused target is inside that
// scope. Use ShortcutScope to bind it to an explicit scope.
func OnShortcut(ctx *internal.Context, spec Shortcut, handler KeyboardHandler, opts ...ListenerOption) {
	if ctx == nil || ctx.Runtime() == nil || handler == nil {
		return
	}
	options := ListenerOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	ctx.Runtime().RegisterShortcut(ctx, spec, handler, options)
}

// DispatchKeyboardEvent dispatches a typed keyboard event through the FluxUI
// capture/target/bubble event system and runs cancelable keydown defaults.
func DispatchKeyboardEvent(ctx *internal.Context, target TargetID, ev *KeyboardEvent) bool {
	if ev == nil {
		return true
	}
	if ctx == nil || ctx.Runtime() == nil {
		return !(ev.Cancelable && ev.DefaultPrevented)
	}
	if target == 0 {
		target = ctx.Runtime().FocusedTarget()
	}
	return ctx.Runtime().DispatchKeyboardEvent(ctx, target, ev)
}

// KeyboardEventFromGio converts a Gio key event to a FluxUI KeyboardEvent.
func KeyboardEventFromGio(ctx *internal.Context, ev key.Event) KeyboardEvent {
	eventType := KeyDown
	if ev.State == key.Release {
		eventType = KeyUp
	}
	return KeyboardEvent{
		Event: Event{
			Type:       eventType,
			Time:       eventTime(ctx),
			Bubbles:    true,
			Cancelable: true,
			Trusted:    true,
		},
		Key:       keyNameForEvent(ev),
		Code:      string(ev.Name),
		Modifiers: ModifiersFromGio(ev.Modifiers),
		Native:    ev,
	}
}

func keyNameForEvent(ev key.Event) string {
	switch ev.Name {
	case key.NameReturn, key.NameEnter:
		return "Enter"
	case key.NameLeftArrow:
		return "ArrowLeft"
	case key.NameRightArrow:
		return "ArrowRight"
	case key.NameUpArrow:
		return "ArrowUp"
	case key.NameDownArrow:
		return "ArrowDown"
	case key.NameHome:
		return "Home"
	case key.NameEnd:
		return "End"
	case key.NameSpace:
		return "Space"
	case key.NameEscape:
		return "Escape"
	case key.NameTab:
		return "Tab"
	default:
		return string(ev.Name)
	}
}
