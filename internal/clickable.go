package internal

import gioWidget "gioui.org/widget"

// ClickableState 封装 Gio 的点击状态。
type ClickableState struct {
	button gioWidget.Clickable
}

// NewClickableState 创建点击状态。
func NewClickableState() *ClickableState {
	return &ClickableState{}
}

// Clicked 返回当前 frame 是否有点击。
func (c *ClickableState) Clicked(ctx *Context) bool {
	return c.button.Clicked(ctx.Gtx)
}

// Hovered 返回是否悬浮。
func (c *ClickableState) Hovered() bool {
	return c.button.Hovered()
}

// Pressed 返回是否按下。
func (c *ClickableState) Pressed() bool {
	return c.button.Pressed()
}

// History 返回近期按压历史，用于绘制 ripple 等交互反馈。
func (c *ClickableState) History() []gioWidget.Press {
	return c.button.History()
}

func (c *ClickableState) raw() *gioWidget.Clickable {
	return &c.button
}
