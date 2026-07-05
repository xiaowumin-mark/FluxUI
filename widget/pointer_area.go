package widget

import (
	"image"
	"math"
	"time"

	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"

	"gioui.org/f32"
	gioEvent "gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/op/clip"
)

type PointerAreaOption func(*pointerAreaConfig)

type pointerAreaConfig struct {
	disabled       bool
	passThrough    bool
	captureOnPress bool
	listeners      []pointerAreaListener
}

type pointerAreaListener struct {
	eventType fluxevent.Type
	pointer   fluxevent.PointerHandler
	wheel     fluxevent.WheelHandler
	options   []fluxevent.ListenerOption
}

type pointerAreaWidget struct {
	child  Widget
	config pointerAreaConfig
}

type pointerAreaState struct {
	tag           any
	pressed       bool
	pointerID     uint64
	button        fluxevent.Button
	pressPos      f32.Point
	pressTime     time.Time
	lastClickPos  f32.Point
	lastClickTime time.Time
	lastButton    fluxevent.Button
	clickCount    int
}

// PointerArea creates a low-level pointer event target around child.
func PointerArea(child Widget, opts ...PointerAreaOption) Widget {
	cfg := pointerAreaConfig{
		passThrough: true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return &pointerAreaWidget{
		child:  child,
		config: cfg,
	}
}

func PointerAreaDisabled(disabled bool) PointerAreaOption {
	return func(cfg *pointerAreaConfig) {
		cfg.disabled = disabled
	}
}

func PointerAreaPassThrough(passThrough bool) PointerAreaOption {
	return func(cfg *pointerAreaConfig) {
		cfg.passThrough = passThrough
	}
}

func PointerCaptureOnPress(capture bool) PointerAreaOption {
	return func(cfg *pointerAreaConfig) {
		cfg.captureOnPress = capture
	}
}

func PointerOn(eventType fluxevent.Type, fn fluxevent.PointerHandler, opts ...fluxevent.ListenerOption) PointerAreaOption {
	return func(cfg *pointerAreaConfig) {
		cfg.listeners = append(cfg.listeners, pointerAreaListener{
			eventType: eventType,
			pointer:   fn,
			options:   append([]fluxevent.ListenerOption(nil), opts...),
		})
	}
}

func PointerOnDown(fn fluxevent.PointerHandler, opts ...fluxevent.ListenerOption) PointerAreaOption {
	return PointerOn(fluxevent.PointerDown, fn, opts...)
}

func PointerOnUp(fn fluxevent.PointerHandler, opts ...fluxevent.ListenerOption) PointerAreaOption {
	return PointerOn(fluxevent.PointerUp, fn, opts...)
}

func PointerOnMove(fn fluxevent.PointerHandler, opts ...fluxevent.ListenerOption) PointerAreaOption {
	return PointerOn(fluxevent.PointerMove, fn, opts...)
}

func PointerOnEnter(fn fluxevent.PointerHandler, opts ...fluxevent.ListenerOption) PointerAreaOption {
	return PointerOn(fluxevent.PointerEnter, fn, opts...)
}

func PointerOnLeave(fn fluxevent.PointerHandler, opts ...fluxevent.ListenerOption) PointerAreaOption {
	return PointerOn(fluxevent.PointerLeave, fn, opts...)
}

func PointerOnOver(fn fluxevent.PointerHandler, opts ...fluxevent.ListenerOption) PointerAreaOption {
	return PointerOn(fluxevent.PointerOver, fn, opts...)
}

func PointerOnOut(fn fluxevent.PointerHandler, opts ...fluxevent.ListenerOption) PointerAreaOption {
	return PointerOn(fluxevent.PointerOut, fn, opts...)
}

func PointerOnCancel(fn fluxevent.PointerHandler, opts ...fluxevent.ListenerOption) PointerAreaOption {
	return PointerOn(fluxevent.PointerCancel, fn, opts...)
}

func PointerOnClick(fn fluxevent.PointerHandler, opts ...fluxevent.ListenerOption) PointerAreaOption {
	return PointerOn(fluxevent.Click, fn, opts...)
}

func PointerOnDoubleClick(fn fluxevent.PointerHandler, opts ...fluxevent.ListenerOption) PointerAreaOption {
	return PointerOn(fluxevent.DblClick, fn, opts...)
}

func PointerOnAuxClick(fn fluxevent.PointerHandler, opts ...fluxevent.ListenerOption) PointerAreaOption {
	return PointerOn(fluxevent.AuxClick, fn, opts...)
}

func PointerOnContextMenu(fn fluxevent.PointerHandler, opts ...fluxevent.ListenerOption) PointerAreaOption {
	return PointerOn(fluxevent.ContextMenu, fn, opts...)
}

func PointerOnWheel(fn fluxevent.WheelHandler, opts ...fluxevent.ListenerOption) PointerAreaOption {
	return func(cfg *pointerAreaConfig) {
		cfg.listeners = append(cfg.listeners, pointerAreaListener{
			eventType: fluxevent.Wheel,
			wheel:     fn,
			options:   append([]fluxevent.ListenerOption(nil), opts...),
		})
	}
}

func (w *pointerAreaWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if w.child == nil {
		return layout.Dimensions{}
	}
	if w.config.disabled {
		return w.child.Layout(ctx.Child(0))
	}

	state := pointerAreaStateFor(ctx)
	registerPointerAreaListeners(ctx, w.config)

	childDims := w.child.Layout(ctx.Child(0))
	registerPointerArea(ctx, state, childDims.Size, w.config)
	processPointerAreaEvents(ctx, state, w.config)

	return childDims
}

func pointerAreaStateFor(ctx *internal.Context) *pointerAreaState {
	value := ctx.Memo("pointer-area", func() any {
		state := &pointerAreaState{}
		state.tag = state
		return state
	})
	state, ok := value.(*pointerAreaState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUI/widget: pointer area state type mismatch")
	}
	if state.tag == nil {
		state.tag = state
	}
	return state
}

func registerPointerAreaListeners(ctx *internal.Context, cfg pointerAreaConfig) {
	for _, listener := range cfg.listeners {
		switch {
		case listener.pointer != nil:
			fluxevent.OnPointer(ctx, listener.eventType, listener.pointer, listener.options...)
		case listener.wheel != nil:
			fluxevent.OnWheel(ctx, listener.wheel, listener.options...)
		}
	}
}

func registerPointerArea(ctx *internal.Context, state *pointerAreaState, size image.Point, cfg pointerAreaConfig) {
	if state == nil || state.tag == nil || size.X <= 0 || size.Y <= 0 {
		return
	}
	var pass pointer.PassStack
	if cfg.passThrough {
		pass = pointer.PassOp{}.Push(ctx.Gtx.Ops)
		defer pass.Pop()
	}
	area := clip.Rect(image.Rectangle{Max: size}).Push(ctx.Gtx.Ops)
	gioEvent.Op(ctx.Gtx.Ops, state.tag)
	area.Pop()
}

func processPointerAreaEvents(ctx *internal.Context, state *pointerAreaState, cfg pointerAreaConfig) {
	if state == nil || state.tag == nil {
		return
	}
	events := drainPointerAreaEvents(ctx, state.tag)
	for i := 0; i < len(events); {
		ev := events[i]
		if isPointerMoveKind(ev.Kind) {
			j := i + 1
			coalesced := []fluxevent.PointerSample{fluxevent.PointerSampleFromGio(ev)}
			for j < len(events) && isPointerMoveKind(events[j].Kind) && events[j].PointerID == ev.PointerID {
				coalesced = append(coalesced, fluxevent.PointerSampleFromGio(events[j]))
				ev = events[j]
				j++
			}
			dispatchPointerAreaPointerEvent(ctx, fluxevent.PointerMove, ev, fluxevent.ButtonNone, coalesced)
			i = j
			continue
		}

		switch ev.Kind {
		case pointer.Press:
			pointerEvent := dispatchPointerAreaPointerEvent(ctx, fluxevent.PointerDown, ev, fluxevent.ButtonNone, nil)
			state.pressed = true
			state.pointerID = uint64(ev.PointerID)
			state.button = pointerEvent.Button
			state.pressPos = ev.Position
			state.pressTime = pointerEvent.Time
			if cfg.captureOnPress && ctx.Runtime() != nil {
				ctx.Runtime().SetPointerCapture(pointerEvent.PointerID, ctx.PathID())
			}
		case pointer.Release:
			override := state.button
			pointerEvent := dispatchPointerAreaPointerEvent(ctx, fluxevent.PointerUp, ev, override, nil)
			if state.pressed && state.pointerID == uint64(ev.PointerID) {
				dispatchPointerAreaClickLike(ctx, state, pointerEvent)
			}
			if ctx.Runtime() != nil {
				ctx.Runtime().ReleasePointerCapture(uint64(ev.PointerID), ctx.PathID())
			}
			state.pressed = false
			state.pointerID = 0
			state.button = fluxevent.ButtonNone
		case pointer.Enter:
			dispatchPointerAreaPointerEvent(ctx, fluxevent.PointerOver, ev, fluxevent.ButtonNone, nil)
			dispatchPointerAreaPointerEvent(ctx, fluxevent.PointerEnter, ev, fluxevent.ButtonNone, nil)
		case pointer.Leave:
			dispatchPointerAreaPointerEvent(ctx, fluxevent.PointerOut, ev, fluxevent.ButtonNone, nil)
			dispatchPointerAreaPointerEvent(ctx, fluxevent.PointerLeave, ev, fluxevent.ButtonNone, nil)
		case pointer.Cancel:
			dispatchPointerAreaPointerEvent(ctx, fluxevent.PointerCancel, ev, state.button, nil)
			if ctx.Runtime() != nil {
				ctx.Runtime().ReleasePointerCapture(uint64(ev.PointerID), ctx.PathID())
			}
			state.pressed = false
			state.pointerID = 0
			state.button = fluxevent.ButtonNone
		case pointer.Scroll:
			wheel := fluxevent.WheelEventFromGio(ctx, ev)
			fluxevent.DispatchWheelEvent(ctx, pointerAreaDispatchTarget(ctx, ev), &wheel)
		}
		i++
	}
}

func drainPointerAreaEvents(ctx *internal.Context, tag any) []pointer.Event {
	const scrollLimit = 1 << 30
	filter := pointer.Filter{
		Target:  tag,
		Kinds:   pointer.Press | pointer.Release | pointer.Move | pointer.Drag | pointer.Enter | pointer.Leave | pointer.Cancel | pointer.Scroll,
		ScrollX: pointer.ScrollRange{Min: -scrollLimit, Max: scrollLimit},
		ScrollY: pointer.ScrollRange{Min: -scrollLimit, Max: scrollLimit},
	}
	var out []pointer.Event
	for {
		ev, ok := ctx.Gtx.Event(filter)
		if !ok {
			break
		}
		if pe, ok := ev.(pointer.Event); ok {
			out = append(out, pe)
		}
	}
	return out
}

func dispatchPointerAreaPointerEvent(ctx *internal.Context, eventType fluxevent.Type, ev pointer.Event, overrideButton fluxevent.Button, coalesced []fluxevent.PointerSample) *fluxevent.PointerEvent {
	pe := fluxevent.PointerEventFromGio(ctx, eventType, ev)
	if overrideButton != fluxevent.ButtonNone {
		pe.Button = overrideButton
	}
	if len(coalesced) > 0 {
		pe.Coalesced = append([]fluxevent.PointerSample(nil), coalesced...)
	}
	fluxevent.DispatchPointerEvent(ctx, pointerAreaDispatchTarget(ctx, ev), &pe)
	return &pe
}

func pointerAreaDispatchTarget(ctx *internal.Context, ev pointer.Event) fluxevent.TargetID {
	if ctx == nil {
		return 0
	}
	target := ctx.PathID()
	if rt := ctx.Runtime(); rt != nil {
		if captured, ok := rt.PointerCaptureTarget(uint64(ev.PointerID)); ok {
			target = captured
		}
	}
	return target
}

func dispatchPointerAreaClickLike(ctx *internal.Context, state *pointerAreaState, source *fluxevent.PointerEvent) {
	if source == nil {
		return
	}
	switch source.Button {
	case fluxevent.ButtonPrimary:
		click := clonePointerEventForType(source, fluxevent.Click)
		if state.lastButton == source.Button &&
			!state.lastClickTime.IsZero() &&
			source.Time.Sub(state.lastClickTime) <= 500*time.Millisecond &&
			pointerDistance(source.Position, state.lastClickPos) <= 5 {
			state.clickCount++
		} else {
			state.clickCount = 1
		}
		state.lastButton = source.Button
		state.lastClickTime = source.Time
		state.lastClickPos = source.Position
		click.ClickCount = state.clickCount
		click.Event.Detail = nil
		click.Event.DefaultPrevented = false
		click.Event.Target = 0
		click.Event.CurrentTarget = 0
		click.Event.Phase = fluxevent.PhaseNone
		fluxevent.DispatchPointerEvent(ctx, source.Target, &click)
		if state.clickCount == 2 {
			dblClick := clonePointerEventForType(source, fluxevent.DblClick)
			dblClick.ClickCount = state.clickCount
			fluxevent.DispatchPointerEvent(ctx, source.Target, &dblClick)
			state.clickCount = 0
		}
	case fluxevent.ButtonAuxiliary, fluxevent.ButtonSecondary, fluxevent.ButtonBack, fluxevent.ButtonForward:
		auxClick := clonePointerEventForType(source, fluxevent.AuxClick)
		fluxevent.DispatchPointerEvent(ctx, source.Target, &auxClick)
		if source.Button == fluxevent.ButtonSecondary {
			contextMenu := clonePointerEventForType(source, fluxevent.ContextMenu)
			fluxevent.DispatchPointerEvent(ctx, source.Target, &contextMenu)
		}
	}
}

func clonePointerEventForType(source *fluxevent.PointerEvent, eventType fluxevent.Type) fluxevent.PointerEvent {
	return fluxevent.PointerEvent{
		Event: fluxevent.Event{
			Type:       eventType,
			Time:       source.Time,
			Bubbles:    true,
			Cancelable: true,
			Trusted:    source.Trusted,
		},
		PointerID:          source.PointerID,
		PointerType:        source.PointerType,
		IsPrimary:          source.IsPrimary,
		Button:             source.Button,
		Buttons:            source.Buttons,
		Position:           source.Position,
		Modifiers:          source.Modifiers,
		TimeOffset:         source.TimeOffset,
		ClickCount:         source.ClickCount,
		Width:              source.Width,
		Height:             source.Height,
		Pressure:           source.Pressure,
		TangentialPressure: source.TangentialPressure,
		TiltX:              source.TiltX,
		TiltY:              source.TiltY,
		Twist:              source.Twist,
	}
}

func isPointerMoveKind(kind pointer.Kind) bool {
	return kind == pointer.Move || kind == pointer.Drag
}

func pointerDistance(a, b f32.Point) float32 {
	dx := float64(a.X - b.X)
	dy := float64(a.Y - b.Y)
	return float32(math.Sqrt(dx*dx + dy*dy))
}
