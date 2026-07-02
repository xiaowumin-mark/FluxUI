package widget

import (
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/theme"
)

type fontScopeWidget struct {
	font  theme.FontSpec
	child Widget
}

type themeScopeWidget struct {
	theme *theme.Theme
	child Widget
}

type interactionQualityScopeWidget struct {
	quality theme.InteractionQuality
	child   Widget
}

// WithFont 在当前子树范围内设置默认字体。
func WithFont(font theme.FontSpec, child Widget) Widget {
	return &fontScopeWidget{
		font:  font.Normalize(),
		child: child,
	}
}

// WithTheme scopes the child layout to the provided theme.
func WithTheme(th *theme.Theme, child Widget) Widget {
	if th == nil {
		return child
	}
	return &themeScopeWidget{
		theme: th,
		child: child,
	}
}

func WithInteractionQuality(quality theme.InteractionQuality, child Widget) Widget {
	return &interactionQualityScopeWidget{
		quality: theme.NormalizeInteractionQuality(quality),
		child:   child,
	}
}

func (w *fontScopeWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if w.child == nil {
		return layout.Dimensions{}
	}
	next := ctx.WithFont(w.font)
	return w.child.Layout(next)
}

func (w *themeScopeWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if w.child == nil {
		return layout.Dimensions{}
	}
	next := ctx.WithTheme(w.theme)
	return w.child.Layout(next)
}

func (w *interactionQualityScopeWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if w.child == nil {
		return layout.Dimensions{}
	}
	next := ctx.WithInteractionQuality(w.quality)
	return w.child.Layout(next)
}
