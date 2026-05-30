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
	hovering = c.Hovered()
	changed = !c.initialized || c.hovered != hovering
	c.hovered = hovering
	c.initialized = true
	return changed, hovering
}

func (c *Clickable) Handle() *internal.ClickableState {
	if c == nil {
		return nil
	}
	return c.handle
}
