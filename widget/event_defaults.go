package widget

import (
	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
)

func dispatchClickDefault(ctx *internal.Context, source *fluxevent.PointerEvent, action func()) bool {
	if ctx == nil {
		return false
	}
	if source == nil {
		source = &fluxevent.PointerEvent{
			Event: fluxevent.Event{
				Type:       fluxevent.Click,
				Target:     ctx.PathID(),
				Bubbles:    true,
				Cancelable: true,
				Trusted:    true,
			},
			PointerType: fluxevent.PointerOther,
			Button:      fluxevent.ButtonNone,
		}
	}
	if source.Type == "" {
		source.Type = fluxevent.Click
	}
	if source.Target == 0 {
		source.Target = ctx.PathID()
	}
	allowed := fluxevent.DispatchPointerEvent(ctx, ctx.PathID(), source)
	if allowed && action != nil {
		action()
	}
	return allowed
}
