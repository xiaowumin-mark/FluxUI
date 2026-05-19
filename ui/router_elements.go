package ui

import (
	internal "github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/router"
	widget "github.com/xiaowumin-mark/FluxUI/widget"
)

// RouteElementOption configures a componentized route declaration.
type RouteElementOption func(*RouteElementConfig)

// RouteElementConfig stores optional route metadata.
type RouteElementConfig struct {
	Key string
}

// RouteKey sets an explicit identity key for a route declaration.
func RouteKey(key string) RouteElementOption {
	return func(cfg *RouteElementConfig) {
		cfg.Key = key
	}
}

// RouteElementSpec describes a componentized route.
type RouteElementSpec struct {
	Path      string
	Key       string
	Component Component
}

// RouteElement declares a componentized route.
func RouteElement(path string, component Component, opts ...RouteElementOption) RouteElementSpec {
	cfg := RouteElementConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return RouteElementSpec{
		Path:      path,
		Key:       cfg.Key,
		Component: component,
	}
}

type routerElement struct {
	routes []RouteElementSpec
}

// RouterElement renders componentized routes while preserving legacy router behavior.
func RouterElement(routes ...RouteElementSpec) Element {
	if len(routes) == 0 {
		return nil
	}
	return routerElement{routes: append([]RouteElementSpec(nil), routes...)}
}

func (e routerElement) render() widget.Widget {
	return nil
}

func (e routerElement) identity() ElementIdentity {
	return ElementIdentity{Kind: "router-element", ChildCount: len(e.routes)}
}

func (e routerElement) renderWithContext(ctx *Context) widget.Widget {
	legacyRoutes := make([]router.Route, 0, len(e.routes))
	for _, routeSpec := range e.routes {
		routeSpec := routeSpec
		legacyRoutes = append(legacyRoutes, router.Route{
			Path: routeSpec.Path,
			Builder: func(routeCtx *Context) widget.Widget {
				return renderRouteComponent(routeCtx, routeSpec)
			},
		})
	}
	return Router(ctx, legacyRoutes)
}

func renderRouteComponent(ctx *Context, routeSpec RouteElementSpec) widget.Widget {
	if ctx == nil || routeSpec.Component == nil {
		return nil
	}
	identity := internal.ComponentIdentity{
		ParentID: routeParentID(routeSpec),
		TypeID:   "route:" + routeSpec.Path,
		Key:      routeSpec.Key,
		Position: 0,
	}
	inst := beginComponentInstance(ctx, identity)
	componentCtx := ctx
	if inst != nil {
		componentCtx = ctx.WithComponentInstance(inst)
	}
	el := routeSpec.Component(componentCtx)
	return renderElementWithContext(componentCtx, el)
}

func routeParentID(routeSpec RouteElementSpec) string {
	if routeSpec.Key == "" {
		return "route-element:" + routeSpec.Path
	}
	return "route-element:" + routeSpec.Path + "#" + routeSpec.Key
}
