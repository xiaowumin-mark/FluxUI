package event

import (
	"fmt"

	"github.com/xiaowumin-mark/FluxUI/internal"
)

type ClickHandler func(ctx *internal.Context)

type HoverHandler func(ctx *internal.Context, hovering bool)

type Clickable struct {
	handle      *internal.ClickableState
	hovered     bool
	initialized bool
}

func UseClickable(ctx *internal.Context) *Clickable {
	value := ctx.Memo("clickable", func() any {
		return &Clickable{handle: internal.NewClickableState()}
	})

	clickable, ok := value.(*Clickable)
	if !ok {
		panic(fmt.Sprintf("github.com/xiaowumin-mark/FluxUI/event: key %q clickable state type mismatch", ctx.TreePath()))
	}
	if clickable.handle != nil {
		clickable.handle.BindRuntime(ctx.Runtime())
	}

	return clickable
}

func (c *Clickable) Clicked(ctx *internal.Context) bool {
	if c == nil || c.handle == nil {
		return false
	}
	return c.handle.Clicked(ctx)
}

func (c *Clickable) Hovered() bool {
	if c == nil || c.handle == nil {
		return false
	}
	return c.handle.Hovered()
}

func (c *Clickable) Pressed() bool {
	if c == nil || c.handle == nil {
		return false
	}
	return c.handle.Pressed()
}

func (c *Clickable) Focused(ctx *internal.Context) bool {
	if c == nil || c.handle == nil {
		return false
	}
	return c.handle.Focused(ctx)
}

func (c *Clickable) HoverChanged() (changed bool, hovering bool) {
	return c.hoverChanged(nil)
}

func (c *Clickable) HoverChangedWithContext(ctx *internal.Context) (changed bool, hovering bool) {
	return c.hoverChanged(ctx)
}

func (c *Clickable) hoverChanged(ctx *internal.Context) (changed bool, hovering bool) {
	if c == nil {
		return false, false
	}
	wasInitialized := c != nil && c.initialized
	hovering = c.Hovered()
	changed = !c.initialized || c.hovered != hovering
	c.hovered = hovering
	c.initialized = true
	if changed && wasInitialized && ctx != nil && ctx.Runtime() != nil {
		ctx.Runtime().RecordRedrawReason("pointer.hover_changed")
	}
	return changed, hovering
}

func (c *Clickable) Handle() *internal.ClickableState {
	if c == nil {
		return nil
	}
	return c.handle
}
