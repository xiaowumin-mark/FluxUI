package ui

import (
	"reflect"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestRouteElementDefaultIdentityReusesByPattern(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	seen := make([]int, 0, 2)

	routeSpec := RouteElement("/users/:id", func(ctx *Context) Element {
		state := UseState(ctx, 1)
		seen = append(seen, state.Value())
		state.Set(state.Value() + 1)
		return nil
	})

	for range 2 {
		rt.BeginFrame()
		ctx := internal.NewContext(gtx, rt)
		renderRouteComponent(ctx, routeSpec)
		rt.EndFrame()
	}

	if len(seen) != 2 || seen[0] != 1 || seen[1] != 2 {
		t.Fatalf("expected route pattern identity to reuse state, got %v", seen)
	}
}

func TestRouteElementExplicitKeyRemounts(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	seen := make([]int, 0, 2)

	component := func(ctx *Context) Element {
		state := UseState(ctx, 1)
		seen = append(seen, state.Value())
		state.Set(state.Value() + 1)
		return nil
	}

	for _, id := range []string{"42", "99"} {
		rt.BeginFrame()
		ctx := internal.NewContext(gtx, rt)
		renderRouteComponent(ctx, RouteElement("/users/:id", component, RouteKey(id)))
		rt.EndFrame()
	}

	if len(seen) != 2 || seen[0] != 1 || seen[1] != 1 {
		t.Fatalf("expected explicit route key change to remount state, got %v", seen)
	}
}

func TestRouteElementMetadataOptions(t *testing.T) {
	guard := func(ctx *Context, from, to string) bool { return true }
	route := RouteElement(
		"/projects/:id",
		func(ctx *Context) Element { return nil },
		RouteKey("project-route"),
		RouteName("project"),
		RouteTitle("Project"),
		RouteMeta("layout", "workspace"),
		RouteMetaMap(map[string]any{"auth": true}),
		RouteBeforeEnter(guard),
	)

	if route.Key != "project-route" || route.Name != "project" || route.Title != "Project" {
		t.Fatalf("unexpected route metadata: %#v", route)
	}
	if route.Meta["layout"] != "workspace" || route.Meta["auth"] != true {
		t.Fatalf("unexpected route meta map: %#v", route.Meta)
	}
	if route.BeforeEnter == nil || !route.BeforeEnter(nil, "/", "/projects/42") {
		t.Fatal("expected route guard to be stored")
	}
}

func TestRouterElementPassesOptionsAndRouteMetadata(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	var seenInfo *RouteInfo
	var notFoundPath string

	routerWidget := renderElementWithContext(internal.NewContext(gtx, rt), RouterElement(
		RouteElement("/", func(ctx *Context) Element { return TextElement("home") }, RouteName("home")),
	).With(
		RouterTransition(TransitionFade),
		RouterTransitionDuration(time.Second),
		RouterNotFoundElement(func(ctx *Context) Element {
			notFoundPath = CurrentPath(ctx)
			return TextElement("404")
		}),
	))

	rt.BeginFrame()
	ctx := internal.NewContext(gtx, rt)
	if routerWidget == nil {
		t.Fatal("expected router widget")
	}
	routerWidget.Layout(ctx)
	rt.EndFrame()

	rt.BeginFrame()
	ctx = internal.NewContext(gtx, rt)
	Navigate(ctx.Scope("router").Scope("content"), "/missing")
	routerWidget.Layout(ctx)
	rt.EndFrame()

	if notFoundPath != "/missing" {
		t.Fatalf("expected Element not-found to receive route context, got %q", notFoundPath)
	}

	rt = internal.NewRuntime(nil)
	routerWidget = renderElementWithContext(internal.NewContext(gtx, rt), RouterElement(
		RouteElement("/", func(ctx *Context) Element {
			seenInfo = UseRoute(ctx)
			return TextElement("home")
		}, RouteName("home"), RouteMeta("slot", "root")),
	))

	rt.BeginFrame()
	ctx = internal.NewContext(gtx, rt)
	routerWidget.Layout(ctx)
	rt.EndFrame()

	if seenInfo == nil || seenInfo.Name != "home" || seenInfo.Meta["slot"] != "root" {
		t.Fatalf("expected route metadata through RouterElement, got %#v", seenInfo)
	}
}

func TestWindowElementCreatesWindowSpec(t *testing.T) {
	spec := WindowElement(func(ctx *Context) Element {
		return TextElement("window")
	}, Title("Element Window"), Size(320, 240), MinSize(240, 160), MaxSize(960, 720))
	if spec.Root == nil || len(spec.Options) != 4 {
		t.Fatalf("expected WindowElement to create root and clone options, got %#v", spec)
	}
	if reflect.ValueOf(spec.Root).IsNil() {
		t.Fatal("expected WindowElement root to be non-nil")
	}
}
