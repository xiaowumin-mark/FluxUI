package event

import (
	"fmt"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
)

// ClickHandler 处理点击事件。
type ClickHandler func(ctx *internal.Context)

// HoverHandler 处理悬浮变化事件。
type HoverHandler func(ctx *internal.Context, hovering bool)

// Clickable 是对点击交互的稳定封装。
type Clickable struct {
	handle      *internal.ClickableState
	hovered     bool
	initialized bool
}

// UseClickable 绑定当前作用域的点击状态。
func UseClickable(ctx *internal.Context) *Clickable {
	value := ctx.Memo("clickable", func() any {
		return &Clickable{handle: internal.NewClickableState()}
	})

	clickable, ok := value.(*Clickable)
	if !ok {
		panic(fmt.Sprintf("github.com/xiaowumin-mark/FluxUIevent: key %q 的点击状态类型错误", ctx.TreePath()))
	}

	return clickable
}

// Clicked 返回当前 frame 是否发生点击。
func (c *Clickable) Clicked(ctx *internal.Context) bool {
	if c == nil || c.handle == nil {
		return false
	}
	return c.handle.Clicked(ctx)
}

// Hovered 返回当前是否悬浮。
func (c *Clickable) Hovered() bool {
	if c == nil || c.handle == nil {
		return false
	}
	return c.handle.Hovered()
}

// Pressed 返回当前是否按下。
func (c *Clickable) Pressed() bool {
	if c == nil || c.handle == nil {
		return false
	}
	return c.handle.Pressed()
}

// HoverChanged 返回悬浮状态是否发生变化。
func (c *Clickable) HoverChanged() (changed bool, hovering bool) {
	hovering = c.Hovered()
	changed = !c.initialized || c.hovered != hovering
	c.hovered = hovering
	c.initialized = true
	return changed, hovering
}

// Handle 返回内部点击句柄，供 widget 层注册事件区域。
func (c *Clickable) Handle() *internal.ClickableState {
	if c == nil {
		return nil
	}
	return c.handle
}
