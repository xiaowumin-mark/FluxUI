package widget

import (
	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
)

// EventBoundaryOption configures an EventBoundary.
type EventBoundaryOption func(*eventBoundaryConfig)

type eventBoundaryConfig struct {
	disabled bool
	mode     fluxevent.BoundaryMode
	redirect fluxevent.TargetID
}

type eventBoundaryWidget struct {
	child  Widget
	config eventBoundaryConfig
}

type eventPortalWidget struct {
	child Widget
	owner fluxevent.TargetID
}

// EventBoundary creates a propagation boundary around child. By default it
// stops propagation at the boundary target.
func EventBoundary(child Widget, opts ...EventBoundaryOption) Widget {
	cfg := eventBoundaryConfig{
		mode: fluxevent.BoundaryStop,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return &eventBoundaryWidget{child: child, config: cfg}
}

// EventBoundaryDisabled disables boundary registration and lays out child
// directly under the parent context.
func EventBoundaryDisabled(disabled bool) EventBoundaryOption {
	return func(cfg *eventBoundaryConfig) {
		cfg.disabled = disabled
	}
}

// EventBoundaryStopPropagation stops propagation at the boundary target.
func EventBoundaryStopPropagation() EventBoundaryOption {
	return func(cfg *eventBoundaryConfig) {
		cfg.mode = fluxevent.BoundaryStop
		cfg.redirect = 0
	}
}

// EventBoundaryRedirectTo redirects propagation from the boundary target to
// target. Use this for portal roots that should bubble to an owner target.
func EventBoundaryRedirectTo(target fluxevent.TargetID) EventBoundaryOption {
	return func(cfg *eventBoundaryConfig) {
		cfg.mode = fluxevent.BoundaryRedirect
		cfg.redirect = target
	}
}

// EventPortal creates a logical event portal. Events inside child bubble to
// owner even if the widget is visually mounted elsewhere.
func EventPortal(child Widget, owner fluxevent.TargetID) Widget {
	return &eventPortalWidget{child: child, owner: owner}
}

func (w *eventBoundaryWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if w.child == nil {
		return layout.Dimensions{}
	}
	if w.config.disabled {
		return w.child.Layout(ctx.Child(0))
	}
	boundaryCtx := ctx.Scope("event-boundary")
	switch w.config.mode {
	case fluxevent.BoundaryRedirect:
		fluxevent.RegisterBoundary(boundaryCtx, fluxevent.BoundaryRedirectTo(w.config.redirect))
	default:
		fluxevent.RegisterBoundary(boundaryCtx, fluxevent.BoundaryStopPropagation())
	}
	return w.child.Layout(boundaryCtx.Child(0))
}

func (w *eventPortalWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if w.child == nil {
		return layout.Dimensions{}
	}
	portalCtx := ctx.Scope("event-portal")
	fluxevent.RegisterPortal(portalCtx, w.owner)
	return w.child.Layout(portalCtx.Child(0))
}
