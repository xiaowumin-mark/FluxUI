package internal

import gioWidget "gioui.org/widget"

type ClickableState struct {
	button gioWidget.Clickable
}

func NewClickableState() *ClickableState {
	return &ClickableState{}
}

func (c *ClickableState) Clicked(ctx *Context) bool {
	return c != nil && ctx != nil && c.button.Clicked(ctx.Gtx)
}

func (c *ClickableState) Hovered() bool {
	return c != nil && c.button.Hovered()
}

func (c *ClickableState) Pressed() bool {
	return c != nil && c.button.Pressed()
}

func (c *ClickableState) Focused(ctx *Context) bool {
	return c != nil && ctx != nil && ctx.Gtx.Focused(&c.button)
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
