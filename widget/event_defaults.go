package widget

import (
	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
)

func dispatchClickDefault(ctx *internal.Context, source *fluxevent.PointerEvent, action func()) bool {
	return runClickableDefaultAction(ctx, source, false, func(*internal.Context) {
		if action != nil {
			action()
		}
	})
}

func drainClickableDefaultAction(ctx *internal.Context, clickable *fluxevent.Clickable, disabled bool, action func(*internal.Context)) bool {
	if disabled || clickable == nil {
		return false
	}
	ran := false
	for {
		click, ok := clickable.ClickedEvent(ctx)
		if !ok {
			break
		}
		if runClickableDefaultAction(ctx, click, false, action) {
			ran = true
		}
	}
	return ran
}

func registerClickableFocusAction(ctx *internal.Context, disabled bool, action func(*internal.Context)) {
	if disabled {
		fluxevent.RegisterFocusTarget(ctx, fluxevent.FocusDisabled(true))
		return
	}
	fluxevent.RegisterFocusTarget(ctx, fluxevent.FocusActivate(func(actionCtx *internal.Context) {
		runClickableDefaultAction(actionCtx, nil, false, action)
	}))
}

func clickableInteractionSnapshot(ctx *internal.Context, clickable *fluxevent.Clickable, disabled bool, includeFocus bool) fluxevent.InteractionSnapshot {
	if disabled || clickable == nil {
		return fluxevent.InteractionSnapshot{}
	}
	return clickable.Snapshot(ctx, includeFocus)
}

func runClickableDefaultAction(ctx *internal.Context, source *fluxevent.PointerEvent, disabled bool, action func(*internal.Context)) bool {
	if ctx == nil {
		return false
	}
	if disabled {
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
		action(ctx)
	}
	return allowed
}
