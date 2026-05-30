package widget

import (
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/theme"
)

type textStyleWidget struct {
	style theme.TextStyle
	child Widget
}

func withTextStyle(style theme.TextStyle, child Widget) Widget {
	if child == nil {
		return nil
	}
	return &textStyleWidget{style: style, child: child}
}

func (t *textStyleWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if t == nil || t.child == nil {
		return layout.Dimensions{}
	}
	return t.child.Layout(ctx.WithTextStyle(t.style).Child(0))
}
