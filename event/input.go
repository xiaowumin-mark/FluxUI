package event

import (
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
)

const (
	PointerDown   Type = "pointerdown"
	PointerUp     Type = "pointerup"
	PointerMove   Type = "pointermove"
	PointerEnter  Type = "pointerenter"
	PointerLeave  Type = "pointerleave"
	PointerOver   Type = "pointerover"
	PointerOut    Type = "pointerout"
	PointerCancel Type = "pointercancel"
	DblClick      Type = "dblclick"
	AuxClick      Type = "auxclick"
	ContextMenu   Type = "contextmenu"
	Wheel         Type = "wheel"
)

// PointerType identifies the physical pointer source.
type PointerType string

const (
	PointerMouse PointerType = "mouse"
	PointerTouch PointerType = "touch"
	PointerPen   PointerType = "pen"
	PointerOther PointerType = "other"
)

// Button follows the DOM PointerEvent button values.
type Button int

const (
	ButtonNone      Button = -1
	ButtonPrimary   Button = 0
	ButtonAuxiliary Button = 1
	ButtonSecondary Button = 2
	ButtonBack      Button = 3
	ButtonForward   Button = 4
)

// Buttons is a bitset of currently pressed buttons.
type Buttons uint16

const (
	ButtonsPrimary Buttons = 1 << iota
	ButtonsSecondary
	ButtonsAuxiliary
	ButtonsBack
	ButtonsForward
)

// Contain reports whether all buttons are present.
func (b Buttons) Contain(buttons Buttons) bool {
	return b&buttons == buttons
}

// Contains reports whether all buttons are present.
func (b Buttons) Contains(buttons Buttons) bool {
	return b.Contain(buttons)
}

// Modifiers records the keyboard modifiers active for an input event.
type Modifiers = internal.Modifiers

// PointerSample records one raw pointer sample. PointerEvent.CoalescedSamples
// exposes samples that were folded into the dispatched pointermove event.
type PointerSample struct {
	PointerID   uint64
	PointerType PointerType
	Button      Button
	Buttons     Buttons
	Position    f32.Point
	Modifiers   Modifiers
	TimeOffset  time.Duration
}

// PointerEvent is FluxUI's browser-style pointer event.
type PointerEvent struct {
	Event

	PointerID   uint64
	PointerType PointerType
	IsPrimary   bool
	Button      Button
	Buttons     Buttons
	Position    f32.Point
	Modifiers   Modifiers
	TimeOffset  time.Duration
	ClickCount  int

	Width              float32
	Height             float32
	Pressure           float32
	TangentialPressure float32
	TiltX              float32
	TiltY              float32
	Twist              float32

	Coalesced []PointerSample
}

// CoalescedSamples returns the raw samples folded into this event.
func (e *PointerEvent) CoalescedSamples() []PointerSample {
	if e == nil || len(e.Coalesced) == 0 {
		return nil
	}
	out := make([]PointerSample, len(e.Coalesced))
	copy(out, e.Coalesced)
	return out
}

// SetPointerCapture captures later events for this pointer to ctx's target.
func (e *PointerEvent) SetPointerCapture(ctx *internal.Context) {
	if e == nil || ctx == nil || ctx.Runtime() == nil {
		return
	}
	ctx.Runtime().SetPointerCapture(e.PointerID, ctx.PathID())
}

// ReleasePointerCapture releases capture for this pointer if ctx owns it.
func (e *PointerEvent) ReleasePointerCapture(ctx *internal.Context) bool {
	if e == nil || ctx == nil || ctx.Runtime() == nil {
		return false
	}
	return ctx.Runtime().ReleasePointerCapture(e.PointerID, ctx.PathID())
}

// HasPointerCapture reports whether ctx currently owns this pointer capture.
func (e *PointerEvent) HasPointerCapture(ctx *internal.Context) bool {
	if e == nil || ctx == nil || ctx.Runtime() == nil {
		return false
	}
	return ctx.Runtime().HasPointerCapture(e.PointerID, ctx.PathID())
}

// WheelDeltaMode follows the DOM WheelEvent deltaMode values.
type WheelDeltaMode int

const (
	WheelDeltaPixel WheelDeltaMode = iota
	WheelDeltaLine
	WheelDeltaPage
)

// WheelEvent exposes raw wheel/scroll delta and pointer position.
type WheelEvent struct {
	Event

	DeltaX     float32
	DeltaY     float32
	DeltaZ     float32
	DeltaMode  WheelDeltaMode
	Position   f32.Point
	Modifiers  Modifiers
	TimeOffset time.Duration
}

// PointerHandler handles a typed pointer event.
type PointerHandler func(ctx *internal.Context, event *PointerEvent)

// WheelHandler handles a typed wheel event.
type WheelHandler func(ctx *internal.Context, event *WheelEvent)

// OnPointer registers a typed pointer listener for the current context target.
func OnPointer(ctx *internal.Context, eventType Type, handler PointerHandler, opts ...ListenerOption) {
	if handler == nil {
		return
	}
	On(ctx, eventType, func(ctx *internal.Context, ev *Event) {
		if ev == nil {
			return
		}
		if pointerEvent, ok := ev.Detail.(*PointerEvent); ok {
			handler(ctx, pointerEvent)
		}
	}, opts...)
}

// OnWheel registers a typed wheel listener for the current context target.
func OnWheel(ctx *internal.Context, handler WheelHandler, opts ...ListenerOption) {
	if handler == nil {
		return
	}
	On(ctx, Wheel, func(ctx *internal.Context, ev *Event) {
		if ev == nil {
			return
		}
		if wheelEvent, ok := ev.Detail.(*WheelEvent); ok {
			handler(ctx, wheelEvent)
		}
	}, opts...)
}

// DispatchPointerEvent dispatches a typed pointer event through the FluxUI
// capture/target/bubble event system.
func DispatchPointerEvent(ctx *internal.Context, target TargetID, ev *PointerEvent) bool {
	if ev == nil {
		return true
	}
	applyPointerDefaults(ev)
	ev.Event.Detail = ev
	return DispatchEvent(ctx, target, &ev.Event)
}

// DispatchWheelEvent dispatches a typed wheel event.
func DispatchWheelEvent(ctx *internal.Context, target TargetID, ev *WheelEvent) bool {
	if ev == nil {
		return true
	}
	if ev.Type == "" {
		ev.Type = Wheel
	}
	ev.Bubbles = true
	ev.Cancelable = true
	ev.Event.Detail = ev
	return DispatchEvent(ctx, target, &ev.Event)
}

// PointerEventFromGio converts a Gio pointer event to a FluxUI PointerEvent.
func PointerEventFromGio(ctx *internal.Context, eventType Type, ev pointer.Event) PointerEvent {
	return PointerEvent{
		Event: Event{
			Type:       eventType,
			Time:       eventTime(ctx),
			Bubbles:    pointerEventBubbles(eventType),
			Cancelable: pointerEventCancelable(eventType),
			Trusted:    true,
		},
		PointerID:   uint64(ev.PointerID),
		PointerType: PointerTypeFromGio(ev.Source),
		IsPrimary:   true,
		Button:      ButtonFromGio(ev),
		Buttons:     ButtonsFromGio(ev.Buttons),
		Position:    ev.Position,
		Modifiers:   ModifiersFromGio(ev.Modifiers),
		TimeOffset:  ev.Time,
	}
}

// WheelEventFromGio converts a Gio scroll event to a FluxUI WheelEvent.
func WheelEventFromGio(ctx *internal.Context, ev pointer.Event) WheelEvent {
	return WheelEvent{
		Event: Event{
			Type:       Wheel,
			Time:       eventTime(ctx),
			Bubbles:    true,
			Cancelable: true,
			Trusted:    true,
		},
		DeltaX:     ev.Scroll.X,
		DeltaY:     ev.Scroll.Y,
		DeltaMode:  WheelDeltaPixel,
		Position:   ev.Position,
		Modifiers:  ModifiersFromGio(ev.Modifiers),
		TimeOffset: ev.Time,
	}
}

// PointerTypeFromGio maps Gio pointer sources to FluxUI pointer types.
func PointerTypeFromGio(source pointer.Source) PointerType {
	switch source {
	case pointer.Mouse:
		return PointerMouse
	case pointer.Touch:
		return PointerTouch
	default:
		return PointerOther
	}
}

// TypeFromGioPointerKind maps Gio pointer kinds to FluxUI pointer event types.
func TypeFromGioPointerKind(kind pointer.Kind) (Type, bool) {
	switch kind {
	case pointer.Press:
		return PointerDown, true
	case pointer.Release:
		return PointerUp, true
	case pointer.Move, pointer.Drag:
		return PointerMove, true
	case pointer.Enter:
		return PointerEnter, true
	case pointer.Leave:
		return PointerLeave, true
	case pointer.Cancel:
		return PointerCancel, true
	case pointer.Scroll:
		return Wheel, true
	default:
		return "", false
	}
}

// ButtonFromGio maps Gio's current button set to the DOM-style changed button.
func ButtonFromGio(ev pointer.Event) Button {
	switch {
	case ev.Buttons.Contain(pointer.ButtonPrimary):
		return ButtonPrimary
	case ev.Buttons.Contain(pointer.ButtonTertiary):
		return ButtonAuxiliary
	case ev.Buttons.Contain(pointer.ButtonSecondary):
		return ButtonSecondary
	case ev.Buttons.Contain(pointer.ButtonQuaternary):
		return ButtonBack
	case ev.Buttons.Contain(pointer.ButtonQuinary):
		return ButtonForward
	default:
		return ButtonNone
	}
}

// ButtonsFromGio maps Gio button bitsets to DOM-style button bitsets.
func ButtonsFromGio(buttons pointer.Buttons) Buttons {
	var out Buttons
	if buttons.Contain(pointer.ButtonPrimary) {
		out |= ButtonsPrimary
	}
	if buttons.Contain(pointer.ButtonSecondary) {
		out |= ButtonsSecondary
	}
	if buttons.Contain(pointer.ButtonTertiary) {
		out |= ButtonsAuxiliary
	}
	if buttons.Contain(pointer.ButtonQuaternary) {
		out |= ButtonsBack
	}
	if buttons.Contain(pointer.ButtonQuinary) {
		out |= ButtonsForward
	}
	return out
}

// ModifiersFromGio maps Gio key modifiers to FluxUI modifiers.
func ModifiersFromGio(mods key.Modifiers) Modifiers {
	return Modifiers{
		Ctrl:     mods.Contain(key.ModCtrl),
		Shift:    mods.Contain(key.ModShift),
		Alt:      mods.Contain(key.ModAlt),
		Meta:     mods.Contain(key.ModCommand) || mods.Contain(key.ModSuper),
		Shortcut: mods.Contain(key.ModShortcut),
	}
}

func PointerSampleFromGio(ev pointer.Event) PointerSample {
	return PointerSample{
		PointerID:   uint64(ev.PointerID),
		PointerType: PointerTypeFromGio(ev.Source),
		Button:      ButtonFromGio(ev),
		Buttons:     ButtonsFromGio(ev.Buttons),
		Position:    ev.Position,
		Modifiers:   ModifiersFromGio(ev.Modifiers),
		TimeOffset:  ev.Time,
	}
}

func applyPointerDefaults(ev *PointerEvent) {
	if ev.Type == "" {
		ev.Type = PointerMove
	}
	ev.Bubbles = pointerEventBubbles(ev.Type)
	ev.Cancelable = pointerEventCancelable(ev.Type)
}

func pointerEventBubbles(eventType Type) bool {
	switch eventType {
	case PointerEnter, PointerLeave:
		return false
	default:
		return true
	}
}

func pointerEventCancelable(eventType Type) bool {
	switch eventType {
	case PointerEnter, PointerLeave, PointerCancel:
		return false
	default:
		return true
	}
}

func eventTime(ctx *internal.Context) time.Time {
	if ctx == nil || ctx.Now().IsZero() {
		return time.Now()
	}
	return ctx.Now()
}
