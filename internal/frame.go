package internal

import gioLayout "gioui.org/layout"

// Frame 为当前 Gio frame 创建根上下文。
func (r *Runtime) Frame(gtx gioLayout.Context) *Context {
	return NewContext(gtx, r)
}
