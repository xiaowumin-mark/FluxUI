package event

import "fluxui/internal"

// Dispatcher 统一管理组件事件分发。
type Dispatcher struct {
	Click ClickHandler
	Hover HoverHandler
	Key   KeyHandler
}

// DispatchClick 分发点击事件。
func (d Dispatcher) DispatchClick(ctx *internal.Context) {
	if d.Click != nil {
		d.Click(ctx)
	}
}

// DispatchHover 分发悬浮事件。
func (d Dispatcher) DispatchHover(ctx *internal.Context, hovering bool) {
	if d.Hover != nil {
		d.Hover(ctx, hovering)
	}
}

// DispatchKey 分发键盘事件。
func (d Dispatcher) DispatchKey(ctx *internal.Context, key string) {
	if d.Key != nil {
		d.Key(ctx, key)
	}
}
