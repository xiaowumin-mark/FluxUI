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
	Key         string
	Name        string
	Title       string
	Meta        map[string]any
	BeforeEnter func(ctx *Context, from, to string) bool
}

// RouteKey sets an explicit identity key for a route declaration.
func RouteKey(key string) RouteElementOption {
	return func(cfg *RouteElementConfig) {
		cfg.Key = key
	}
}

// RouteName sets a semantic name for a route declaration.
func RouteName(name string) RouteElementOption {
	return func(cfg *RouteElementConfig) {
		cfg.Name = name
	}
}

// RouteTitle sets a display title for a route declaration.
func RouteTitle(title string) RouteElementOption {
	return func(cfg *RouteElementConfig) {
		cfg.Title = title
	}
}

// RouteMeta attaches arbitrary metadata to a route declaration.
func RouteMeta(key string, value any) RouteElementOption {
	return func(cfg *RouteElementConfig) {
		if key == "" {
			return
		}
		if cfg.Meta == nil {
			cfg.Meta = map[string]any{}
		}
		cfg.Meta[key] = value
	}
}

// RouteMetaMap attaches a metadata map to a route declaration.
func RouteMetaMap(meta map[string]any) RouteElementOption {
	return func(cfg *RouteElementConfig) {
		if len(meta) == 0 {
			return
		}
		if cfg.Meta == nil {
			cfg.Meta = map[string]any{}
		}
		for key, value := range meta {
			if key != "" {
				cfg.Meta[key] = value
			}
		}
	}
}

// RouteBeforeEnter sets a route-level guard for a route declaration.
func RouteBeforeEnter(fn func(ctx *Context, from, to string) bool) RouteElementOption {
	return func(cfg *RouteElementConfig) {
		cfg.BeforeEnter = fn
	}
}

// RouteElementSpec describes a componentized route.
type RouteElementSpec struct {
	Path        string
	Key         string
	Name        string
	Title       string
	Meta        map[string]any
	BeforeEnter func(ctx *Context, from, to string) bool
	Component   Component
}

// RouteElement declares a componentized route.
func RouteElement(path string, component Component, opts ...RouteElementOption) RouteElementSpec {
	cfg := RouteElementConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return RouteElementSpec{
		Path:        path,
		Key:         cfg.Key,
		Name:        cfg.Name,
		Title:       cfg.Title,
		Meta:        cloneElementRouteMeta(cfg.Meta),
		BeforeEnter: cfg.BeforeEnter,
		Component:   component,
	}
}

// RouterElementSpec describes a componentized router Element.
type RouterElementSpec struct {
	routes []RouteElementSpec
	opts   []RouterOption
}

// RouterElement renders componentized routes while preserving legacy router behavior.
func RouterElement(routes ...RouteElementSpec) *RouterElementSpec {
	if len(routes) == 0 {
		return nil
	}
	return &RouterElementSpec{routes: append([]RouteElementSpec(nil), routes...)}
}

// With attaches router-level options to a RouterElement.
func (e *RouterElementSpec) With(opts ...RouterOption) *RouterElementSpec {
	if e == nil {
		return nil
	}
	next := &RouterElementSpec{
		routes: append([]RouteElementSpec(nil), e.routes...),
		opts:   append([]RouterOption(nil), e.opts...),
	}
	next.opts = append(next.opts, opts...)
	return next
}

func (e *RouterElementSpec) render() widget.Widget {
	return nil
}

func (e *RouterElementSpec) identity() ElementIdentity {
	if e == nil {
		return ElementIdentity{}
	}
	return ElementIdentity{Kind: "router-element", ChildCount: len(e.routes)}
}

func (e *RouterElementSpec) renderWithContext(ctx *Context) widget.Widget {
	if e == nil {
		return nil
	}
	legacyRoutes := make([]router.Route, 0, len(e.routes))
	for _, routeSpec := range e.routes {
		routeSpec := routeSpec
		legacyRoutes = append(legacyRoutes, router.Route{
			Path:  routeSpec.Path,
			Name:  routeSpec.Name,
			Title: routeSpec.Title,
			Meta:  cloneElementRouteMeta(routeSpec.Meta),
			BeforeEnter: func(routeCtx *Context, from, to string) bool {
				if routeSpec.BeforeEnter == nil {
					return true
				}
				return routeSpec.BeforeEnter(routeCtx, from, to)
			},
			Builder: func(routeCtx *Context) widget.Widget {
				return renderRouteComponent(routeCtx, routeSpec)
			},
		})
	}
	return Router(ctx, legacyRoutes, e.opts...)
}

func renderRouteComponent(ctx *Context, routeSpec RouteElementSpec) widget.Widget {
	if ctx == nil || routeSpec.Component == nil {
		return nil
	}
	return renderRouterElementComponentWithIdentity(ctx, routeParentID(routeSpec), "route:"+routeSpec.Path, routeSpec.Key, routeSpec.Component)
}

func renderRouterElementComponent(ctx *Context, parentID string, component Component) widget.Widget {
	return renderRouterElementComponentWithIdentity(ctx, parentID, componentTypeID(component), "", component)
}

func renderRouterElementComponentWithIdentity(ctx *Context, parentID string, typeID string, key string, component Component) widget.Widget {
	if ctx == nil || component == nil {
		return nil
	}
	identity := internal.ComponentIdentity{
		ParentID: parentID,
		TypeID:   typeID,
		Key:      key,
		Position: 0,
	}
	inst := beginComponentInstance(ctx, identity)
	componentCtx := ctx
	if inst != nil {
		componentCtx = ctx.WithComponentInstance(inst)
	}
	el := component(componentCtx)
	return renderElementWithContext(componentCtx, el)
}

func cloneElementRouteMeta(meta map[string]any) map[string]any {
	if len(meta) == 0 {
		return nil
	}
	out := make(map[string]any, len(meta))
	for key, value := range meta {
		out[key] = value
	}
	return out
}

func routeParentID(routeSpec RouteElementSpec) string {
	if routeSpec.Key == "" {
		return "route-element:" + routeSpec.Path
	}
	return "route-element:" + routeSpec.Path + "#" + routeSpec.Key
}
