package router

import (
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
	if got := RouteParams(ctx); got.Get("id") != "42" || got.Query("tab") != "posts" {
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
		params := RouteParams(ctx)
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
	if got := RouteParams(ctx); got.Get("id") != "42" || got.Query("tab") != "posts" {
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
