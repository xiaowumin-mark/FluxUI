package event

import "github.com/xiaowumin-mark/FluxUI/internal"

// Dispatcher 统一管理组件事件分发。
type Dispatcher struct {
	Click ClickHandler
	Hover HoverHandler
}

// DispatchClick 分发点击事件。
func (d Dispatcher) DispatchClick(ctx *internal.Context) bool {
	return d.DispatchClickEvent(ctx, nil)
}

// DispatchClickEvent dispatches a cancelable click event before running the
// legacy click handler.
func (d Dispatcher) DispatchClickEvent(ctx *internal.Context, source *PointerEvent) bool {
	if ctx == nil {
		return false
	}
	if source == nil {
		source = &PointerEvent{
			Event: Event{
				Type:       Click,
				Target:     ctx.PathID(),
				Bubbles:    true,
				Cancelable: true,
				Trusted:    true,
			},
			PointerType: PointerOther,
			Button:      ButtonNone,
		}
	}
	if source.Type == "" {
		source.Type = Click
	}
	if source.Target == 0 {
		source.Target = ctx.PathID()
	}
	if DispatchPointerEvent(ctx, ctx.PathID(), source) && d.Click != nil {
		d.Click(ctx)
	}
	return !(source.Cancelable && source.DefaultPrevented)
}

// DispatchHover 分发悬浮事件。
func (d Dispatcher) DispatchHover(ctx *internal.Context, hovering bool) {
	if d.Hover != nil {
		d.Hover(ctx, hovering)
	}
}
