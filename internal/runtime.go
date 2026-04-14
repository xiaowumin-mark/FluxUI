package internal

import (
	"fluxui/theme"

	"gioui.org/font/gofont"
	gioText "gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// Runtime 持有跨 frame 的稳定数据。
type Runtime struct {
	memory     map[string]any
	theme      *theme.Theme
	material   *material.Theme
	invalidate func()
}

// NewRuntime 创建运行时。
func NewRuntime(th *theme.Theme) *Runtime {
	if th == nil {
		th = theme.Default()
	}

	mt := material.NewTheme()
	mt.Shaper = gioText.NewShaper(gioText.WithCollection(gofont.Collection()))
	mt.Fg = th.TextColor
	mt.Bg = th.Surface
	mt.ContrastBg = th.Primary
	mt.ContrastFg = th.TextOnPrimary
	mt.TextSize = unit.Sp(th.TextSize)

	return &Runtime{
		memory:   make(map[string]any),
		theme:    th,
		material: mt,
	}
}

// Theme 返回当前主题。
func (r *Runtime) Theme() *theme.Theme {
	return r.theme
}

// MaterialTheme 返回内部使用的 Gio 主题。
func (r *Runtime) MaterialTheme() *material.Theme {
	return r.material
}

// SetInvalidator 绑定窗口重绘函数。
func (r *Runtime) SetInvalidator(fn func()) {
	r.invalidate = fn
}

// RequestRedraw 请求窗口重绘。
func (r *Runtime) RequestRedraw() {
	if r.invalidate != nil {
		r.invalidate()
	}
}

func (r *Runtime) remember(key string, factory func() any) any {
	if value, ok := r.memory[key]; ok {
		return value
	}
	value := factory()
	r.memory[key] = value
	return value
}
