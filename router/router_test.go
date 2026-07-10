package router

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/widget"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func newRouterTestContext() (*internal.Runtime, *internal.Context) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	rt.BeginFrame()
	return rt, internal.NewContext(gtx, rt)
}

func TestMatchPathStatic(t *testing.T) {
	result := matchPath("/home", "/home")
	if !result.matched {
		t.Fatal("expected /home to match /home")
	}
}

func TestMatchPathStaticMismatch(t *testing.T) {
	result := matchPath("/home", "/about")
	if result.matched {
		t.Fatal("expected /home not to match /about")
	}
}

func TestMatchPathParam(t *testing.T) {
	result := matchPath("/users/:id", "/users/42")
	if !result.matched {
		t.Fatal("expected match")
	}
	if result.params["id"] != "42" {
		t.Fatalf("expected id=42, got %s", result.params["id"])
	}
}

func TestMatchPathMultipleParams(t *testing.T) {
	result := matchPath("/users/:uid/posts/:pid", "/users/10/posts/99")
	if !result.matched {
		t.Fatal("expected match")
	}
	if result.params["uid"] != "10" {
		t.Fatalf("expected uid=10, got %s", result.params["uid"])
	}
	if result.params["pid"] != "99" {
		t.Fatalf("expected pid=99, got %s", result.params["pid"])
	}
}

func TestMatchPathTooShort(t *testing.T) {
	result := matchPath("/users/:id/posts", "/users/42")
	if result.matched {
		t.Fatal("expected no match for shorter path")
	}
}

func TestMatchPathTooLong(t *testing.T) {
	result := matchPath("/users", "/users/42")
	if result.matched {
		t.Fatal("expected no match for longer path")
	}
}

func TestMatchPathRoot(t *testing.T) {
	result := matchPath("/", "/")
	if !result.matched {
		t.Fatal("expected root match")
	}
}

func TestMatchPathWildcard(t *testing.T) {
	result := matchPath("/*", "/anything/here")
	if !result.matched {
		t.Fatal("expected wildcard match")
	}
}

func TestExtractQueryParams(t *testing.T) {
	path, query := extractQueryParams("/users/42?tab=posts&sort=asc")
	if path != "/users/42" {
		t.Fatalf("expected path=/users/42, got %s", path)
	}
	if query["tab"] != "posts" {
		t.Fatalf("expected tab=posts, got %s", query["tab"])
	}
	if query["sort"] != "asc" {
		t.Fatalf("expected sort=asc, got %s", query["sort"])
	}
}

func TestExtractQueryParamsDecodesURLValues(t *testing.T) {
	path, query := extractQueryParams("/search?q=hello+world&path=a%2Fb&tag=old&tag=new")
	if path != "/search" {
		t.Fatalf("expected path=/search, got %s", path)
	}
	if query["q"] != "hello world" {
		t.Fatalf("expected decoded q, got %q", query["q"])
	}
	if query["path"] != "a/b" {
		t.Fatalf("expected decoded path query, got %q", query["path"])
	}
	if query["tag"] != "new" {
		t.Fatalf("expected last repeated tag value, got %q", query["tag"])
	}
}

func TestExtractQueryParamsNone(t *testing.T) {
	path, query := extractQueryParams("/users/42")
	if path != "/users/42" {
		t.Fatalf("expected path=/users/42, got %s", path)
	}
	if query != nil {
		t.Fatal("expected nil query")
	}
}

func TestParamsGet(t *testing.T) {
	p := &Params{
		pathParams:  map[string]string{"id": "42"},
		queryParams: map[string]string{"tab": "posts"},
	}
	if p.Get("id") != "42" {
		t.Fatal("expected id=42")
	}
	if p.Query("tab") != "posts" {
		t.Fatal("expected tab=posts")
	}
	if p.Get("missing") != "" {
		t.Fatal("expected empty for missing param")
	}
	if p.Query("missing") != "" {
		t.Fatal("expected empty for missing query")
	}
}

func TestParamsNil(t *testing.T) {
	var p *Params
	if p.Get("id") != "" {
		t.Fatal("expected empty from nil params")
	}
	if p.Query("tab") != "" {
		t.Fatal("expected empty from nil params")
	}
}

func TestSplitPath(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{"/", nil},
		{"/home", []string{"home"}},
		{"/users/42/posts", []string{"users", "42", "posts"}},
		{"users/42", []string{"users", "42"}},
		{"/a//b/", []string{"a", "b"}},
	}

	for _, tc := range cases {
		got := splitPath(tc.input)
		if len(got) == 0 && len(tc.want) == 0 {
			continue
		}
		if len(got) != len(tc.want) {
			t.Fatalf("splitPath(%q): want %v, got %v", tc.input, tc.want, got)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("splitPath(%q)[%d]: want %q, got %q", tc.input, i, tc.want[i], got[i])
			}
		}
	}
}

func TestReverseTransition(t *testing.T) {
	if reverseTransition(TransitionSlideLeft) != TransitionSlideRight {
		t.Fatal("SlideLeft should reverse to SlideRight")
	}
	if reverseTransition(TransitionSlideRight) != TransitionSlideLeft {
		t.Fatal("SlideRight should reverse to SlideLeft")
	}
	if reverseTransition(TransitionFade) != TransitionFade {
		t.Fatal("Fade should reverse to Fade")
	}
	if reverseTransition(TransitionNone) != TransitionNone {
		t.Fatal("None should reverse to None")
	}
}

func TestTransitionPageFramesSlideLeft(t *testing.T) {
	frames := transitionPageFrames(TransitionSlideLeft, 0.5, 400)
	if len(frames) != 2 {
		t.Fatalf("expected two transition frames, got %d", len(frames))
	}

	if frames[0].page != transitionFromPage {
		t.Fatalf("expected old page to render below new page, got %v", frames[0].page)
	}
	if frames[0].dx != -100 {
		t.Fatalf("expected old page to move left by half speed, got dx=%d", frames[0].dx)
	}
	if frames[0].opacity != 0.5 {
		t.Fatalf("expected old page opacity 0.5, got %v", frames[0].opacity)
	}

	if frames[1].page != transitionToPage {
		t.Fatalf("expected new page to render above old page, got %v", frames[1].page)
	}
	if frames[1].dx != 200 {
		t.Fatalf("expected new page to move in from right, got dx=%d", frames[1].dx)
	}
	if frames[1].opacity != 1 {
		t.Fatalf("expected new page opacity 1, got %v", frames[1].opacity)
	}
}

func TestTransitionPageFramesSlideRight(t *testing.T) {
	frames := transitionPageFrames(TransitionSlideRight, 0.5, 400)
	if len(frames) != 2 {
		t.Fatalf("expected two transition frames, got %d", len(frames))
	}

	if frames[0].page != transitionToPage {
		t.Fatalf("expected new page to render below old page, got %v", frames[0].page)
	}
	if frames[0].dx != -100 {
		t.Fatalf("expected new page to move from left by half speed, got dx=%d", frames[0].dx)
	}
	if frames[0].opacity != 0.5 {
		t.Fatalf("expected new page opacity 0.5, got %v", frames[0].opacity)
	}

	if frames[1].page != transitionFromPage {
		t.Fatalf("expected old page to render above new page, got %v", frames[1].page)
	}
	if frames[1].dx != 200 {
		t.Fatalf("expected old page to move out to right, got dx=%d", frames[1].dx)
	}
	if frames[1].opacity != 1 {
		t.Fatalf("expected old page opacity 1, got %v", frames[1].opacity)
	}
}

func TestTransitionProgressUsesEaseOut(t *testing.T) {
	progress, active := transitionProgress(500*time.Millisecond, time.Second)
	if !active {
		t.Fatal("expected transition to remain active before duration ends")
	}
	if progress != 0.75 {
		t.Fatalf("expected ease-out progress 0.75 at half duration, got %v", progress)
	}

	progress, active = transitionProgress(time.Second, time.Second)
	if active {
		t.Fatal("expected transition to stop at duration end")
	}
	if progress != 1 {
		t.Fatalf("expected completed progress 1, got %v", progress)
	}
}

func TestSameRouteQueryNavigationDoesNotDuplicateStableScopeTransition(t *testing.T) {
	_, ctx := newRouterTestContext()
	state := &routerState{
		routes: []Route{{Path: "/search"}},
		config: routerConfig{
			transition:         TransitionFade,
			transitionDuration: time.Second,
		},
		stack: []stackEntry{{path: "/search?q=first", routeIndex: 0}},
		transition: transitionState{
			active: true,
			from:   "/old?token=secret",
			to:     "/search?q=first",
		},
	}

	state.navigate(ctx, "/search?q=second", navPush, navigateOpts{})
	if state.transition.active {
		t.Fatal("same-route query update must not render two pages with the same stable scope")
	}
	if state.transition != (transitionState{}) {
		t.Fatalf("same-route navigation retained stale transition data: %#v", state.transition)
	}
}

func TestRouteScopeIDSeparatesDuplicateNames(t *testing.T) {
	first := routeScopeID(Route{Path: "/first/:id", Name: "duplicate"})
	second := routeScopeID(Route{Path: "/second/:id", Name: "duplicate"})
	if first == second {
		t.Fatalf("duplicate route names shared scope %q", first)
	}
	if first != routeScopeID(Route{Path: "/first/:id", Name: "duplicate"}) {
		t.Fatal("route scope ID must remain stable for the same declaration")
	}
}

func TestNavigationWithoutTransitionClearsRetainedQueryPaths(t *testing.T) {
	_, ctx := newRouterTestContext()
	state := &routerState{
		routes: []Route{
			{Path: "/from"},
			{Path: "/to"},
		},
		config: routerConfig{transition: TransitionFade},
		stack:  []stackEntry{{path: "/from?token=stack-secret", routeIndex: 0}},
		transition: transitionState{
			active:     true,
			from:       "/before?token=from-secret",
			to:         "/from?token=to-secret",
			transition: TransitionFade,
			duration:   time.Second,
		},
	}

	state.navigate(ctx, "/to?token=next-secret", navPush, navigateOpts{
		hasTransition: true,
		transition:    TransitionNone,
	})
	if state.transition != (transitionState{}) {
		t.Fatalf("canceled transition retained path data: %#v", state.transition)
	}
}

func TestCompletedTransitionClearsRetainedQueryPaths(t *testing.T) {
	runtime := internal.NewRuntime(nil)
	var ops op.Ops
	now := time.Unix(100, 0)
	routes := []Route{
		{Path: "/from", Builder: func(*internal.Context) widget.Widget { return countLayoutWidget{} }},
		{Path: "/to", Builder: func(*internal.Context) widget.Widget { return countLayoutWidget{} }},
	}
	routerWidget := New(nil, routes)

	runtime.BeginFrame()
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops, Now: now}, runtime)
	routerWidget.Layout(ctx)
	state := getRouterState(ctx.Scope("router").Scope("content"))
	if state == nil {
		t.Fatal("expected router state")
	}
	state.stack = []stackEntry{{path: "/to?token=current-secret", routeIndex: 1}}
	state.transition = transitionState{
		active:     true,
		from:       "/from?token=from-secret",
		to:         "/to?token=to-secret",
		transition: TransitionFade,
		startTime:  now.Add(-2 * time.Second),
		duration:   time.Second,
	}
	runtime.EndFrame()

	ops.Reset()
	runtime.BeginFrame()
	ctx = internal.NewContext(gioLayout.Context{Ops: &ops, Now: now}, runtime)
	routerWidget.Layout(ctx)
	runtime.EndFrame()
	if state.transition != (transitionState{}) {
		t.Fatalf("completed transition retained path data: %#v", state.transition)
	}
}

func TestNormalizePathForGuard(t *testing.T) {
	if got := normalizePathForGuard("/users/42?tab=posts&sort=asc"); got != "/users/42" {
		t.Fatalf("expected /users/42, got %s", got)
	}
	if got := normalizePathForGuard("/settings"); got != "/settings" {
		t.Fatalf("expected /settings, got %s", got)
	}
	if got := normalizePathForGuard(""); got != "" {
		t.Fatalf("expected empty path, got %s", got)
	}
}

func TestCurrentPathPrefersRouteView(t *testing.T) {
	_, ctx := newRouterTestContext()
	st := &routerState{
		stack: []stackEntry{{
			path:       "/home?tab=posts",
			routeIndex: 0,
			params: Params{
				pathParams:  map[string]string{"id": "42"},
				queryParams: map[string]string{"tab": "posts"},
			},
		}},
	}
	setRouterState(ctx, st)

	if got := CurrentPath(ctx); got != "/home?tab=posts" {
		t.Fatalf("expected current path from router state, got %s", got)
	}
	if got := UseParams(ctx); got.Get("id") != "42" || got.Query("tab") != "posts" {
		t.Fatalf("unexpected route params from state: %#v", got)
	}
	if CanGoBack(ctx) {
		t.Fatal("expected CanGoBack to be false with a single stack entry")
	}
	if StackDepth(ctx) != 1 {
		t.Fatalf("expected stack depth 1, got %d", StackDepth(ctx))
	}

	view := &routeView{
		path:       "/settings",
		canGoBack:  true,
		stackDepth: 3,
		params: Params{
			pathParams:  map[string]string{"section": "general"},
			queryParams: map[string]string{"tab": "profile"},
		},
	}

	withRouteView(ctx, view, func() {
		if got := CurrentPath(ctx); got != "/settings" {
			t.Fatalf("expected route view path, got %s", got)
		}
		params := UseParams(ctx)
		if params.Get("section") != "general" || params.Query("tab") != "profile" {
			t.Fatalf("unexpected route view params: %#v", params)
		}
		if !CanGoBack(ctx) {
			t.Fatal("expected CanGoBack to reflect route view")
		}
		if StackDepth(ctx) != 3 {
			t.Fatalf("expected stack depth 3 from route view, got %d", StackDepth(ctx))
		}
	})

	if got := CurrentPath(ctx); got != "/home?tab=posts" {
		t.Fatalf("expected route view to be restored, got %s", got)
	}
	if got := UseParams(ctx); got.Get("id") != "42" || got.Query("tab") != "posts" {
		t.Fatalf("unexpected restored route params: %#v", got)
	}
}

func TestNavigateQueuesPendingNavigation(t *testing.T) {
	_, ctx := newRouterTestContext()
	st := &routerState{
		stack: []stackEntry{{path: "/home", routeIndex: 0}},
	}
	setRouterState(ctx, st)

	Navigate(ctx, "/settings?tab=profile")
	if st.pendingNav == nil {
		t.Fatal("expected pending navigation to be queued")
	}
	if st.pendingNav.action != navPush {
		t.Fatalf("expected navPush, got %v", st.pendingNav.action)
	}
	if st.pendingNav.path != "/settings?tab=profile" {
		t.Fatalf("unexpected pending path: %s", st.pendingNav.path)
	}
	if got := CurrentPath(ctx); got != "/home" {
		t.Fatalf("expected current path to remain unchanged before layout, got %s", got)
	}
}

func TestUseNavigateUsesCurrentContext(t *testing.T) {
	_, ctx := newRouterTestContext()
	st := &routerState{
		stack: []stackEntry{{path: "/home", routeIndex: 0}},
	}
	setRouterState(ctx, st)

	nav := UseNavigate(ctx)
	nav("/settings?tab=profile")

	if st.pendingNav == nil {
		t.Fatal("expected pending navigation from UseNavigate")
	}
	if st.pendingNav.path != "/settings?tab=profile" {
		t.Fatalf("unexpected pending path: %s", st.pendingNav.path)
	}
}

func TestUseLocationAndParams(t *testing.T) {
	_, ctx := newRouterTestContext()
	st := &routerState{
		stack: []stackEntry{{
			path:       "/users/42?tab=posts",
			routeIndex: 0,
			params: Params{
				pathParams:  map[string]string{"id": "42"},
				queryParams: map[string]string{"tab": "posts"},
			},
		}},
	}
	setRouterState(ctx, st)

	loc := UseLocation(ctx)
	if loc == nil {
		t.Fatal("expected location")
	}
	if loc.Path != "/users/42?tab=posts" {
		t.Fatalf("unexpected location path: %s", loc.Path)
	}
	if loc.Pathname != "/users/42" {
		t.Fatalf("unexpected pathname: %s", loc.Pathname)
	}
	if loc.Query("tab") != "posts" {
		t.Fatalf("unexpected query tab: %s", loc.Query("tab"))
	}

	params := UseParams(ctx)
	if params.Get("id") != "42" || params.Query("tab") != "posts" {
		t.Fatalf("unexpected params: %#v", params)
	}
}

func TestUseLocationInsideRouteView(t *testing.T) {
	_, ctx := newRouterTestContext()
	view := &routeView{
		path: "/settings?tab=profile",
		params: Params{
			pathParams:  map[string]string{"section": "general"},
			queryParams: map[string]string{"tab": "profile"},
		},
	}
	withRouteView(ctx, view, func() {
		loc := UseLocation(ctx)
		if loc.Path != "/settings?tab=profile" {
			t.Fatalf("unexpected location path in route view: %s", loc.Path)
		}
		if loc.Pathname != "/settings" {
			t.Fatalf("unexpected pathname in route view: %s", loc.Pathname)
		}
		if UseParams(ctx).Get("section") != "general" {
			t.Fatal("expected route view params to be available")
		}
	})
}

func TestTransitionLayoutsFromAndToUntilDurationEnds(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	now := time.Unix(0, 0)
	var homeLayouts, settingsLayouts int
	var homePath, settingsPath string

	routes := []Route{
		{
			Path: "/home",
			Builder: func(ctx *internal.Context) widget.Widget {
				homePath = CurrentPath(ctx)
				return countLayoutWidget{count: &homeLayouts}
			},
		},
		{
			Path: "/settings",
			Builder: func(ctx *internal.Context) widget.Widget {
				settingsPath = CurrentPath(ctx)
				return countLayoutWidget{count: &settingsLayouts}
			},
		},
	}
	routerWidget := New(nil, routes, RouterTransition(TransitionFade), RouterTransitionDuration(time.Second))

	rt.BeginFrame()
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops, Now: now}, rt)
	routerWidget.Layout(ctx)
	rt.EndFrame()
	if homeLayouts != 1 || settingsLayouts != 0 {
		t.Fatalf("expected initial frame to layout only home, got home=%d settings=%d", homeLayouts, settingsLayouts)
	}

	rt.BeginFrame()
	ctx = internal.NewContext(gioLayout.Context{Ops: &ops, Now: now}, rt)
	Navigate(ctx.Scope("router").Scope("content"), "/settings")
	routerWidget.Layout(ctx)
	rt.EndFrame()
	if homeLayouts != 2 || settingsLayouts != 1 {
		t.Fatalf("expected transition start to layout from and to pages, got home=%d settings=%d", homeLayouts, settingsLayouts)
	}
	if homePath != "/home" || settingsPath != "/settings" {
		t.Fatalf("expected route views for from/to pages, got home=%q settings=%q", homePath, settingsPath)
	}

	rt.BeginFrame()
	ctx = internal.NewContext(gioLayout.Context{Ops: &ops, Now: now.Add(500 * time.Millisecond)}, rt)
	routerWidget.Layout(ctx)
	rt.EndFrame()
	if homeLayouts != 3 || settingsLayouts != 2 {
		t.Fatalf("expected active transition to keep both pages mounted, got home=%d settings=%d", homeLayouts, settingsLayouts)
	}

	rt.BeginFrame()
	ctx = internal.NewContext(gioLayout.Context{Ops: &ops, Now: now.Add(time.Second)}, rt)
	routerWidget.Layout(ctx)
	rt.EndFrame()
	if homeLayouts != 3 || settingsLayouts != 3 {
		t.Fatalf("expected completed transition to layout only destination, got home=%d settings=%d", homeLayouts, settingsLayouts)
	}
	state := getRouterState(ctx.Scope("router").Scope("content"))
	if state == nil || state.transition != (transitionState{}) {
		t.Fatalf("completed transition retained route data: %#v", state)
	}
}

func TestRouteBeforeEnterBlocksNavigation(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	var blocked bool
	var homeLayouts, adminLayouts int

	routes := []Route{
		{Path: "/", Builder: func(ctx *internal.Context) widget.Widget { return countLayoutWidget{count: &homeLayouts} }},
		{
			Path: "/admin",
			BeforeEnter: func(ctx *internal.Context, from, to string) bool {
				blocked = from == "/" && to == "/admin"
				return false
			},
			Builder: func(ctx *internal.Context) widget.Widget { return countLayoutWidget{count: &adminLayouts} },
		},
	}
	routerWidget := New(nil, routes)

	rt.BeginFrame()
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops}, rt)
	routerWidget.Layout(ctx)
	rt.EndFrame()

	rt.BeginFrame()
	ctx = internal.NewContext(gioLayout.Context{Ops: &ops}, rt)
	Navigate(ctx.Scope("router").Scope("content"), "/admin")
	routerWidget.Layout(ctx)
	rt.EndFrame()

	if !blocked {
		t.Fatal("expected route-level guard to run")
	}
	if homeLayouts != 2 || adminLayouts != 0 {
		t.Fatalf("expected navigation to stay on home, got home=%d admin=%d", homeLayouts, adminLayouts)
	}
}

func TestUseRouteReturnsMatchedMetadata(t *testing.T) {
	rt := internal.NewRuntime(nil)
	var ops op.Ops
	var info *RouteInfo
	var detailLayouts int

	routes := []Route{
		{Path: "/", Builder: func(ctx *internal.Context) widget.Widget { return countLayoutWidget{} }},
		{
			Path:  "/projects/:id",
			Name:  "project",
			Title: "Project",
			Meta:  map[string]any{"layout": "workspace"},
			Builder: func(ctx *internal.Context) widget.Widget {
				info = UseRoute(ctx)
				return countLayoutWidget{count: &detailLayouts}
			},
		},
	}
	routerWidget := New(nil, routes)

	rt.BeginFrame()
	ctx := internal.NewContext(gioLayout.Context{Ops: &ops}, rt)
	routerWidget.Layout(ctx)
	rt.EndFrame()

	rt.BeginFrame()
	ctx = internal.NewContext(gioLayout.Context{Ops: &ops}, rt)
	Navigate(ctx.Scope("router").Scope("content"), "/projects/42")
	routerWidget.Layout(ctx)
	rt.EndFrame()

	if detailLayouts != 1 {
		t.Fatalf("expected project route to layout once, got %d", detailLayouts)
	}

	if info == nil || !info.Matched || info.Path != "/projects/:id" || info.Name != "project" || info.Title != "Project" {
		t.Fatalf("unexpected route info: %#v", info)
	}
	if info.Meta["layout"] != "workspace" {
		t.Fatalf("expected metadata to be available, got %#v", info.Meta)
	}
	info.Meta["layout"] = "changed"
	if routes[1].Meta["layout"] != "workspace" {
		t.Fatal("expected UseRoute metadata to be cloned")
	}
}

func TestRouteScopesExcludeDynamicPathAndQueryData(t *testing.T) {
	runtime := internal.NewRuntime(nil)
	var ops op.Ops
	var builderPaths []internal.PathID
	var layoutPaths []internal.PathID
	var locations []string

	routes := []Route{{
		Path: "/search/:category",
		Name: "search",
		Builder: func(ctx *internal.Context) widget.Widget {
			builderPaths = append(builderPaths, ctx.PathID())
			locations = append(locations, UseLocation(ctx).Path)
			return routePathWidget{paths: &layoutPaths}
		},
	}}
	routerWidget := New(nil, routes)

	render := func(target string) {
		runtime.BeginFrame()
		ctx := internal.NewContext(gioLayout.Context{Ops: &ops}, runtime)
		if target != "" {
			Navigate(ctx.Scope("router").Scope("content"), target)
		}
		routerWidget.Layout(ctx)
		runtime.EndFrame()
	}

	render("")
	render("/search/private?token=first-secret")
	render("/search/other?token=second-secret")

	if len(builderPaths) != 3 || len(layoutPaths) != 3 {
		t.Fatalf("expected three route renders, got builders=%d layouts=%d", len(builderPaths), len(layoutPaths))
	}
	for _, id := range builderPaths[1:] {
		if id != builderPaths[0] {
			t.Fatalf("builder scope changed with route data: %#v", builderPaths)
		}
	}
	for _, id := range layoutPaths[1:] {
		if id != layoutPaths[0] {
			t.Fatalf("layout scope changed with route data: %#v", layoutPaths)
		}
	}
	for _, id := range []internal.PathID{builderPaths[0], layoutPaths[0]} {
		debugPath := runtime.DebugPath(id)
		if strings.Contains(debugPath, "private") || strings.Contains(debugPath, "token") || strings.Contains(debugPath, "secret") {
			t.Fatalf("route debug path retained dynamic or query data: %q", debugPath)
		}
	}
	if locations[1] != "/search/private?token=first-secret" || locations[2] != "/search/other?token=second-secret" {
		t.Fatalf("route location data was not preserved for consumers: %#v", locations)
	}
}

func TestDuplicateRouteNamesUseDistinctStableScopes(t *testing.T) {
	runtime := internal.NewRuntime(nil)
	var ops op.Ops
	var firstBuilderPaths, secondBuilderPaths []internal.PathID
	var firstLayoutPaths, secondLayoutPaths []internal.PathID
	routes := []Route{
		{
			Path: "/first",
			Name: "duplicate",
			Builder: func(ctx *internal.Context) widget.Widget {
				firstBuilderPaths = append(firstBuilderPaths, ctx.PathID())
				return routePathWidget{paths: &firstLayoutPaths}
			},
		},
		{
			Path: "/second",
			Name: "duplicate",
			Builder: func(ctx *internal.Context) widget.Widget {
				secondBuilderPaths = append(secondBuilderPaths, ctx.PathID())
				return routePathWidget{paths: &secondLayoutPaths}
			},
		},
	}
	routerWidget := New(nil, routes)

	render := func(target string) {
		runtime.BeginFrame()
		ctx := internal.NewContext(gioLayout.Context{Ops: &ops}, runtime)
		if target != "" {
			Navigate(ctx.Scope("router").Scope("content"), target)
		}
		routerWidget.Layout(ctx)
		runtime.EndFrame()
		ops.Reset()
	}
	render("")
	render("/second")

	if len(firstBuilderPaths) != 1 || len(secondBuilderPaths) != 1 ||
		len(firstLayoutPaths) != 1 || len(secondLayoutPaths) != 1 {
		t.Fatalf("unexpected duplicate-name render counts: builders=(%d,%d) layouts=(%d,%d)",
			len(firstBuilderPaths), len(secondBuilderPaths), len(firstLayoutPaths), len(secondLayoutPaths))
	}
	if firstBuilderPaths[0] == secondBuilderPaths[0] {
		t.Fatalf("duplicate route names shared builder scope %v", firstBuilderPaths[0])
	}
	if firstLayoutPaths[0] == secondLayoutPaths[0] {
		t.Fatalf("duplicate route names shared layout scope %v", firstLayoutPaths[0])
	}
}

func TestRouterStackBoundsUniqueQueryHistory(t *testing.T) {
	_, ctx := newRouterTestContext()
	state := &routerState{
		routes: []Route{{Path: "/search"}},
		stack:  []stackEntry{{path: "/search", routeIndex: 0}},
	}

	for i := 0; i < maxRouterStackEntries*2; i++ {
		path := "/search?token=" + strconv.Itoa(i) + strings.Repeat("x", 8<<10)
		state.navigate(ctx, path, navPush, navigateOpts{})
	}

	if len(state.stack) > maxRouterStackEntries {
		t.Fatalf("router stack entry limit exceeded: %d", len(state.stack))
	}
	totalBytes := 0
	for _, entry := range state.stack {
		totalBytes += retainedStackEntryBytes(entry)
	}
	if totalBytes > maxRouterStackBytes {
		t.Fatalf("router stack byte limit exceeded: %d", totalBytes)
	}
	latest := state.stack[len(state.stack)-1].path
	if !strings.Contains(latest, "token="+strconv.Itoa(maxRouterStackEntries*2-1)) {
		t.Fatalf("expected newest navigation to be retained, got %q", latest)
	}
	for _, entry := range state.stack {
		if strings.Contains(entry.path, "token=0x") {
			t.Fatalf("oldest sensitive query was retained: %q", entry.path)
		}
	}

	before := len(state.stack)
	state.navigate(ctx, "/search?token="+strings.Repeat("z", maxRouterPathBytes), navPush, navigateOpts{})
	if len(state.stack) != before {
		t.Fatal("oversized route path should be ignored")
	}
}

type routePathWidget struct {
	paths *[]internal.PathID
}

func (w routePathWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if w.paths != nil {
		*w.paths = append(*w.paths, ctx.PathID())
	}
	return layout.Dimensions{}
}

type countLayoutWidget struct {
	count *int
}

func (w countLayoutWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if w.count != nil {
		(*w.count)++
	}
	return layout.Dimensions{}
}
