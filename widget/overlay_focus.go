package widget

import (
	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"

	"gioui.org/io/key"
)

var md3OverlayKeyFilter = key.Filter{
	Optional: key.ModCtrl | key.ModCommand | key.ModShift | key.ModAlt | key.ModSuper,
}

func md3ProcessOverlayKeyboardEvents(ctx *internal.Context) {
	if ctx == nil || ctx.Runtime() == nil {
		return
	}
	focused := ctx.Runtime().FocusedTarget()
	if focused == 0 || !ctx.Runtime().EventPathContains(focused, ctx.PathID()) {
		return
	}
	for {
		raw, ok := ctx.Gtx.Event(md3OverlayKeyFilter)
		if !ok {
			break
		}
		if keyEvent, ok := raw.(key.Event); ok {
			ev := fluxevent.KeyboardEventFromGio(ctx, keyEvent)
			fluxevent.DispatchKeyboardEvent(ctx, focused, &ev)
		}
	}
}

func md3RegisterEscapeClose(ctx *internal.Context, onEscape func(*internal.Context)) {
	if ctx == nil || ctx.Runtime() == nil || onEscape == nil {
		return
	}
	fluxevent.OnKeyDown(ctx, func(eventCtx *internal.Context, ev *fluxevent.KeyboardEvent) {
		if ev == nil || ev.Key != "Escape" || ev.DefaultPrevented {
			return
		}
		ev.PreventDefault()
		onEscape(eventCtx)
	})
}

func md3FocusWithin(ctx *internal.Context) bool {
	if ctx == nil || ctx.Runtime() == nil {
		return false
	}
	focused := ctx.Runtime().FocusedTarget()
	return focused != 0 && ctx.Runtime().EventPathContains(focused, ctx.PathID())
}

func md3EnsureFocusWithin(ctx *internal.Context) bool {
	if md3FocusWithin(ctx) {
		return true
	}
	if ctx == nil || ctx.Runtime() == nil {
		return false
	}
	return ctx.Runtime().MoveFocusWithin(ctx, ctx.PathID(), internal.FocusForward)
}

func md3ClearFocusIfInside(ctx *internal.Context, scope internal.PathID) {
	if ctx == nil || ctx.Runtime() == nil {
		return
	}
	focused := ctx.Runtime().FocusedTarget()
	if focused != 0 && ctx.Runtime().EventPathContains(focused, scope) {
		ctx.Runtime().BlurFocus(ctx, 0)
	}
}
