package internal

import (
	"image"

	"gioui.org/io/key"
	gioWidget "gioui.org/widget"
)

type ClickableState struct {
	button  gioWidget.Clickable
	runtime *Runtime
}

type ClickableSnapshot struct {
	Hovered bool
	Pressed bool
	Focused bool
}

// ClickData is the best-effort payload for a Clickable activation.
type ClickData struct {
	Modifiers   key.Modifiers
	NumClicks   int
	Position    image.Point
	HasPosition bool
}

func NewClickableState() *ClickableState {
	return &ClickableState{}
}

func (c *ClickableState) BindRuntime(rt *Runtime) {
	if c == nil {
		return
	}
	c.runtime = rt
}

func (c *ClickableState) Clicked(ctx *Context) bool {
	_, clicked := c.ClickedEvent(ctx)
	return clicked
}

func (c *ClickableState) ClickedEvent(ctx *Context) (ClickData, bool) {
	if c == nil || ctx == nil {
		return ClickData{}, false
	}
	if ctx.runtime != nil {
		c.runtime = ctx.runtime
	}
	done := c.startInputSection()
	if done != nil {
		defer done()
	}
	click, clicked := c.button.Update(ctx.Gtx)
	if !clicked {
		return ClickData{}, false
	}
	data := ClickData{
		Modifiers: click.Modifiers,
		NumClicks: click.NumClicks,
	}
	history := c.button.History()
	if len(history) > 0 {
		last := history[len(history)-1]
		if !last.Cancelled {
			data.Position = last.Position
			data.HasPosition = true
		}
	}
	return data, true
}

func (c *ClickableState) Hovered() bool {
	if c == nil {
		return false
	}
	done := c.startInputSection()
	if done != nil {
		defer done()
	}
	return c.button.Hovered()
}

func (c *ClickableState) Pressed() bool {
	if c == nil {
		return false
	}
	done := c.startInputSection()
	if done != nil {
		defer done()
	}
	return c.button.Pressed()
}

func (c *ClickableState) Focused(ctx *Context) bool {
	if c == nil || ctx == nil {
		return false
	}
	if ctx.runtime != nil {
		c.runtime = ctx.runtime
	}
	done := c.startInputSection()
	if done != nil {
		defer done()
	}
	return ctx.Gtx.Focused(&c.button)
}

func (c *ClickableState) Snapshot(ctx *Context, includeFocus bool) ClickableSnapshot {
	if c == nil {
		return ClickableSnapshot{}
	}
	if ctx != nil && ctx.runtime != nil {
		c.runtime = ctx.runtime
	}
	done := c.startInputSection()
	if done != nil {
		defer done()
	}
	snapshot := ClickableSnapshot{
		Hovered: c.button.Hovered(),
		Pressed: c.button.Pressed(),
	}
	if includeFocus && ctx != nil {
		snapshot.Focused = ctx.Gtx.Focused(&c.button)
	}
	return snapshot
}

func (c *ClickableState) History() []gioWidget.Press {
	if c == nil {
		return nil
	}
	return c.button.History()
}

func (c *ClickableState) raw() *gioWidget.Clickable {
	return &c.button
}

func (c *ClickableState) startInputSection() func() {
	if c == nil || c.runtime == nil {
		return nil
	}
	return c.runtime.StartFrameSection(PerfInput, 1)
}
