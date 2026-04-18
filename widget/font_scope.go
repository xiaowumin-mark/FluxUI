package widget

import (
	"fluxui/internal"
	"fluxui/layout"
	"fluxui/theme"
)

type fontScopeWidget struct {
	font  theme.FontSpec
	child Widget
}

// WithFont 在当前子树范围内设置默认字体。
func WithFont(font theme.FontSpec, child Widget) Widget {
	return &fontScopeWidget{
		font:  font.Normalize(),
		child: child,
	}
}

func (w *fontScopeWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if w.child == nil {
		return layout.Dimensions{}
	}
	next := ctx.WithFont(w.font)
	return w.child.Layout(next)
}
