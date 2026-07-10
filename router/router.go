package router

import (
	"image"
	"time"

	"github.com/xiaowumin-mark/FluxUI/anim"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/widget"

	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

const (
	maxRouterPathBytes    = 64 << 10
	maxRouterStackEntries = 128
	maxRouterStackBytes   = 1 << 20
)

// NavigateFunc 是实验性路由导航函数签名。
type NavigateFunc func(path string, opts ...NavigateOption)

// Location 描述当前路由位置。
type Location struct {
	Path        string
	Pathname    string
	QueryParams map[string]string
}

// RouteInfo describes the currently matched route declaration.
type RouteInfo struct {
	Path    string
	Name    string
	Title   string
	Meta    map[string]any
	Matched bool
}

// Query 返回查询参数值。
func (l *Location) Query(name string) string {
	if l == nil || l.QueryParams == nil {
		return ""
	}
	return l.QueryParams[name]
}

// AllQueryParams 返回查询参数副本。
func (l *Location) AllQueryParams() map[string]string {
	if l == nil || l.QueryParams == nil {
		return nil
	}
	out := make(map[string]string, len(l.QueryParams))
	for k, v := range l.QueryParams {
		out[k] = v
	}
	return out
}

// Route 定义一条路由规则。
type Route struct {
	Path        string
	Name        string
	Title       string
	Meta        map[string]any
	BeforeEnter BeforeEachFunc
	Builder     func(ctx *internal.Context) widget.Widget
}

// Option 定义路由器配置项。
type Option func(*routerConfig)

// BeforeEachFunc 路由守卫函数。from/to 是路径，返回 false 阻止导航。
type BeforeEachFunc func(ctx *internal.Context, from, to string) bool

type routerConfig struct {
	transition         Transition
	transitionDuration time.Duration
	notFound           func(ctx *internal.Context) widget.Widget
	beforeEach         BeforeEachFunc
}

// RouterTransition 设置全局默认过渡动画。
func RouterTransition(t Transition) Option {
	return func(cfg *routerConfig) {
		cfg.transition = t
	}
}

// RouterTransitionDuration 设置过渡动画时长。
func RouterTransitionDuration(d time.Duration) Option {
	return func(cfg *routerConfig) {
		cfg.transitionDuration = d
	}
}

// RouterNotFound 设置 404 页面。
func RouterNotFound(builder func(ctx *internal.Context) widget.Widget) Option {
	return func(cfg *routerConfig) {
		cfg.notFound = builder
	}
}

// RouterBeforeEach 设置全局路由守卫。
func RouterBeforeEach(fn BeforeEachFunc) Option {
	return func(cfg *routerConfig) {
		cfg.beforeEach = fn
	}
}

// NavigateOption 定义单次导航的配置。
type NavigateOption func(*navigateOpts)

type navigateOpts struct {
	transition    Transition
	hasTransition bool
}

// WithTransition 为单次导航指定过渡动画。
func WithTransition(t Transition) NavigateOption {
	return func(opts *navigateOpts) {
		opts.transition = t
		opts.hasTransition = true
	}
}

// stackEntry 导航栈条目。
type stackEntry struct {
	path       string
	params     Params
	routeIndex int // 对应 routes 数组下标，-1 为未匹配
}

// routerState 路由器持久化状态。
type routerState struct {
	stack      []stackEntry
	routes     []Route
	config     routerConfig
	transition transitionState
	// 挂起的导航操作
	pendingNav *pendingNavigation
}

type pendingNavigation struct {
	path   string
	action navAction
	opts   navigateOpts
}

type navAction int

const (
	navPush navAction = iota
	navReplace
	navPop
)

// routerWidget 是路由器组件。
type routerWidget struct {
	routes []Route
	config routerConfig
}

// New 创建路由器组件。
func New(ctx *internal.Context, routes []Route, opts ...Option) widget.Widget {
	cfg := routerConfig{
		transition:         TransitionNone,
		transitionDuration: defaultTransitionDuration,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return &routerWidget{
		routes: routes,
		config: cfg,
	}
}

// stateFor 获取持久化路由状态。
func stateFor(ctx *internal.Context, routes []Route, config routerConfig) *routerState {
	value := ctx.Memo("router", func() any {
		st := &routerState{
			routes: routes,
			config: config,
		}
		// 初始化：导航到第一个路由
		if len(routes) > 0 {
			path := routes[0].Path
			idx, params := st.matchRoute(path)
			st.stack = []stackEntry{{
				path:       path,
				params:     params,
				routeIndex: idx,
			}}
		}
		return st
	})
	st := value.(*routerState)
	// 每帧更新路由表和配置
	st.routes = routes
	st.config = config
	return st
}

// matchRoute 在路由表中查找匹配的路由。
func (s *routerState) matchRoute(fullPath string) (int, Params) {
	path, query := extractQueryParams(fullPath)
	for i, route := range s.routes {
		result := matchPath(route.Path, path)
		if result.matched {
			return i, Params{
				pathParams:  result.params,
				queryParams: query,
			}
		}
	}
	return -1, Params{queryParams: query}
}

// currentEntry 返回栈顶条目。
func (s *routerState) currentEntry() *stackEntry {
	if len(s.stack) == 0 {
		return nil
	}
	return &s.stack[len(s.stack)-1]
}

// navigate 执行导航操作。
func (s *routerState) navigate(ctx *internal.Context, fullPath string, action navAction, opts navigateOpts) {
	if len(fullPath) > maxRouterPathBytes {
		return
	}
	current := s.currentEntry()
	currentPath := ""
	if current != nil {
		currentPath = current.path
	}

	// 守卫统一使用纯路径，避免 from/to 一个带 query 一个不带 query 的不一致。
	currentCleanPath := normalizePathForGuard(currentPath)
	targetCleanPath := normalizePathForGuard(fullPath)

	// 路由守卫
	if s.config.beforeEach != nil {
		if !s.config.beforeEach(ctx, currentCleanPath, targetCleanPath) {
			return
		}
	}

	idx, params := s.matchRoute(fullPath)
	if idx >= 0 && idx < len(s.routes) {
		if guard := s.routes[idx].BeforeEnter; guard != nil {
			if !guard(ctx, currentCleanPath, targetCleanPath) {
				return
			}
		}
	}

	entry := stackEntry{
		path:       fullPath,
		params:     params,
		routeIndex: idx,
	}

	// 确定过渡动画
	trans := s.config.transition
	if opts.hasTransition {
		trans = opts.transition
	}
	if action == navPop {
		trans = reverseTransition(trans)
	}

	// 同一路由声明内的参数/query 更新复用同一稳定 scope，不同时布局两份共享状态。
	sameRoute := current != nil && current.routeIndex == idx
	if trans != TransitionNone && currentPath != fullPath && !sameRoute {
		s.transition = transitionState{
			active:     true,
			from:       currentPath,
			to:         fullPath,
			transition: trans,
			startTime:  ctx.Now(),
			duration:   s.config.transitionDuration,
			progress:   0,
		}
	} else {
		s.transition = transitionState{}
	}

	switch action {
	case navPush:
		s.stack = append(s.stack, entry)
		s.trimStack()
	case navReplace:
		if len(s.stack) > 0 {
			s.stack[len(s.stack)-1] = entry
		} else {
			s.stack = []stackEntry{entry}
		}
	case navPop:
		if len(s.stack) > 1 {
			last := len(s.stack) - 1
			s.stack[last] = stackEntry{}
			s.stack = s.stack[:last]
		}
	}

	ctx.RequestFrameRedrawReason("router.navigate")
}

func (s *routerState) trimStack() {
	if s == nil || len(s.stack) <= 1 {
		return
	}
	totalBytes := 0
	for i := range s.stack {
		totalBytes += retainedStackEntryBytes(s.stack[i])
	}
	drop := 0
	for len(s.stack)-drop > 1 && (len(s.stack)-drop > maxRouterStackEntries || totalBytes > maxRouterStackBytes) {
		totalBytes -= retainedStackEntryBytes(s.stack[drop])
		s.stack[drop] = stackEntry{}
		drop++
	}
	if drop == 0 {
		return
	}
	copy(s.stack, s.stack[drop:])
	newLen := len(s.stack) - drop
	clear(s.stack[newLen:])
	s.stack = s.stack[:newLen]
}

func retainedStackEntryBytes(entry stackEntry) int {
	total := len(entry.path)
	for key, value := range entry.params.pathParams {
		total += len(key) + len(value)
	}
	for key, value := range entry.params.queryParams {
		total += len(key) + len(value)
	}
	return total
}

func normalizePathForGuard(fullPath string) string {
	path, _ := extractQueryParams(fullPath)
	return path
}

// Layout 实现 Widget 接口。
func (w *routerWidget) Layout(ctx *internal.Context) layout.Dimensions {
	next := ctx.Scope("router")
	st := stateFor(next, w.routes, w.config)

	// 处理挂起导航
	if pn := st.pendingNav; pn != nil {
		st.pendingNav = nil
		if pn.action == navPop {
			st.navigate(next, pn.path, navPop, pn.opts)
		} else {
			st.navigate(next, pn.path, pn.action, pn.opts)
		}
	}

	// 存储 state 到 context 供 Navigate/NavigateBack 使用
	routerCtx := next.Scope("content")
	setRouterState(routerCtx, st)

	// 更新过渡进度
	if st.transition.active {
		elapsed := next.Now().Sub(st.transition.startTime)
		progress, active := transitionProgress(elapsed, st.transition.duration)
		st.transition.progress = progress
		st.transition.active = active
		if active {
			if rt := next.Runtime(); rt != nil {
				rt.RecordFrameSection(internal.PerfAnimation, 1)
			}
			next.RequestFrameRedrawReason("animation.router")
		} else {
			st.transition = transitionState{}
		}
	}

	// 构建当前页面（transition.to）
	toPage, toView, ok := resolvePageForPath(routerCtx, st, w.routes, w.config.notFound, entryPath(st))
	if !ok {
		return layout.Dimensions{}
	}

	// 过渡期：同时渲染旧页面和新页面，旧页面做透明度递减，避免突兀消失。
	if st.transition.active {
		fromPage, fromView, _ := resolvePageForPath(routerCtx, st, w.routes, w.config.notFound, st.transition.from)
		return layoutWithTransition(routerCtx, fromPage, fromView, toPage, toView, st.transition)
	}

	return layoutPageWithRouteView(routerCtx, toPage, toView)
}

func transitionProgress(elapsed, duration time.Duration) (float32, bool) {
	if duration <= 0 || elapsed >= duration {
		return 1, false
	}
	return anim.EaseOut(float32(elapsed) / float32(duration)), true
}

func entryPath(st *routerState) string {
	if st == nil {
		return ""
	}
	entry := st.currentEntry()
	if entry == nil {
		return ""
	}
	return entry.path
}

func resolvePageForPath(
	ctx *internal.Context,
	base *routerState,
	routes []Route,
	notFound func(ctx *internal.Context) widget.Widget,
	fullPath string,
) (widget.Widget, *routeView, bool) {
	if ctx == nil || base == nil {
		return nil, nil, false
	}
	idx, params := base.matchRoute(fullPath)
	view := &routeView{
		path:       fullPath,
		params:     params,
		canGoBack:  len(base.stack) > 1,
		stackDepth: len(base.stack),
	}

	if idx >= 0 && idx < len(routes) {
		view.scopeID = routeScopeID(routes[idx])
		view.route = routeInfoFromRoute(routes[idx], true)
		var page widget.Widget
		withRouteView(ctx, view, func() {
			scopeCtx := ctx.Scope("route-builder:" + view.scopeID)
			page = routes[idx].Builder(scopeCtx)
		})
		if page != nil {
			return page, view, true
		}
	}

	if notFound != nil {
		view.scopeID = "not-found"
		var page widget.Widget
		withRouteView(ctx, view, func() {
			page = notFound(ctx.Scope("not-found"))
		})
		if page != nil {
			return page, view, true
		}
	}

	return nil, nil, false
}

func withRouteView(ctx *internal.Context, view *routeView, fn func()) {
	if fn == nil {
		return
	}
	if ctx == nil {
		fn()
		return
	}
	value := ctx.Persistent(routeViewKey, func() any {
		return &routeViewHolder{}
	})
	holder, ok := value.(*routeViewHolder)
	if !ok {
		fn()
		return
	}
	prev := holder.view
	holder.view = view
	defer func() {
		holder.view = prev
	}()
	fn()
}

func routeScopeID(route Route) string {
	path, _ := extractQueryParams(route.Path)
	if path == "" {
		path = "__empty__"
	}
	if route.Name == "" {
		return "path:" + path
	}
	return "path:" + path + "|name:" + route.Name
}

func routeLayoutScope(scopeID string) string {
	if scopeID == "" {
		return "route:default"
	}
	return "route-layout:" + scopeID
}

func layoutPageWithRouteView(ctx *internal.Context, page widget.Widget, view *routeView) layout.Dimensions {
	if page == nil {
		return layout.Dimensions{}
	}
	var dims layout.Dimensions
	withRouteView(ctx, view, func() {
		pageCtx := ctx
		if ctx != nil {
			scopeName := "route:default"
			if view != nil {
				scopeName = routeLayoutScope(view.scopeID)
			}
			pageCtx = ctx.Scope(scopeName)
		}
		dims = page.Layout(pageCtx)
	})
	return dims
}

// layoutWithTransition 在过渡期间布局页面。
// 滑动过渡按移动端页面栈顺序绘制：前进时旧页在下，返回时新页在下。
func layoutWithTransition(
	ctx *internal.Context,
	fromPage widget.Widget,
	fromView *routeView,
	toPage widget.Widget,
	toView *routeView,
	ts transitionState,
) layout.Dimensions {
	var merged layout.Dimensions
	width := 0
	if ctx != nil {
		width = ctx.Gtx.Constraints.Max.X
	}

	for _, frame := range transitionPageFrames(ts.transition, ts.progress, width) {
		page := fromPage
		view := fromView
		if frame.page == transitionToPage {
			page = toPage
			view = toView
		}

		dims := layoutTransitionPage(ctx, page, view, frame.dx, frame.opacity)
		merged = mergeDimensions(merged, dims)
	}

	return merged
}

type transitionPage int

const (
	transitionFromPage transitionPage = iota
	transitionToPage
)

type transitionPageFrame struct {
	page    transitionPage
	dx      int
	opacity float32
}

func transitionPageFrames(t Transition, progress float32, width int) []transitionPageFrame {
	progress = clampProgress(progress)
	maxWidth := float32(width)

	switch t {
	case TransitionFade:
		return []transitionPageFrame{
			{page: transitionFromPage, opacity: 1 - progress},
			{page: transitionToPage, opacity: progress},
		}

	case TransitionSlideLeft:
		return []transitionPageFrame{
			{page: transitionFromPage, dx: int(-maxWidth * progress / 2), opacity: 1 - progress},
			{page: transitionToPage, dx: int(maxWidth * (1 - progress)), opacity: 1},
		}

	case TransitionSlideRight:
		return []transitionPageFrame{
			{page: transitionToPage, dx: int(-maxWidth * (1 - progress) / 2), opacity: progress},
			{page: transitionFromPage, dx: int(maxWidth * progress), opacity: 1},
		}

	default:
		return []transitionPageFrame{{page: transitionToPage, opacity: 1}}
	}
}

func clampProgress(progress float32) float32 {
	if progress < 0 {
		return 0
	}
	if progress > 1 {
		return 1
	}
	return progress
}

func mergeDimensions(a, b layout.Dimensions) layout.Dimensions {
	if b.Size.X > a.Size.X {
		a.Size.X = b.Size.X
	}
	if b.Size.Y > a.Size.Y {
		a.Size.Y = b.Size.Y
	}
	return a
}

func layoutTransitionPage(ctx *internal.Context, page widget.Widget, view *routeView, dx int, opacity float32) layout.Dimensions {
	if page == nil {
		return layout.Dimensions{}
	}
	if ctx == nil {
		return layoutPageWithRouteView(ctx, page, view)
	}

	gtx := ctx.Gtx
	clipStack := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	defer clipStack.Pop()

	offsetStack := op.Offset(image.Pt(dx, 0)).Push(gtx.Ops)
	defer offsetStack.Pop()

	opacity = clampProgress(opacity)
	if opacity < 1 {
		opacityLayer := paint.PushOpacity(gtx.Ops, opacity)
		defer opacityLayer.Pop()
	}

	return layoutPageWithRouteView(ctx, page, view)
}

// ──────────────────────────────
// 全局路由操作（通过 context 存储 state）
// ──────────────────────────────

const routerStateKey = "__fluxui_router_state__"

// setRouterState 将路由状态存储到 context 供子组件使用。
func setRouterState(ctx *internal.Context, st *routerState) {
	ctx.Persistent(routerStateKey, func() any {
		return &routerStateHolder{}
	}).(*routerStateHolder).state = st
}

// getRouterState 从 context 获取路由状态。
func getRouterState(ctx *internal.Context) *routerState {
	if ctx == nil {
		return nil
	}
	value := ctx.Persistent(routerStateKey, func() any {
		return &routerStateHolder{}
	})
	holder, ok := value.(*routerStateHolder)
	if !ok {
		return nil
	}
	return holder.state
}

type routerStateHolder struct {
	state *routerState
}

type routeView struct {
	path       string
	scopeID    string
	params     Params
	canGoBack  bool
	stackDepth int
	route      RouteInfo
}

const routeViewKey = "__fluxui_router_view__"

type routeViewHolder struct {
	view *routeView
}

func setRouteView(ctx *internal.Context, view *routeView) {
	if ctx == nil {
		return
	}
	ctx.Persistent(routeViewKey, func() any {
		return &routeViewHolder{}
	}).(*routeViewHolder).view = view
}

func getRouteView(ctx *internal.Context) *routeView {
	if ctx == nil {
		return nil
	}
	value := ctx.Persistent(routeViewKey, func() any {
		return &routeViewHolder{}
	})
	holder, ok := value.(*routeViewHolder)
	if !ok {
		return nil
	}
	return holder.view
}

// Navigate 导航到指定路径（push 到栈）。
func Navigate(ctx *internal.Context, path string, opts ...NavigateOption) {
	st := getRouterState(ctx)
	if st == nil || len(path) > maxRouterPathBytes {
		return
	}
	var navOpts navigateOpts
	for _, opt := range opts {
		opt(&navOpts)
	}
	st.pendingNav = &pendingNavigation{
		path:   path,
		action: navPush,
		opts:   navOpts,
	}
	ctx.RequestRedrawReason("router.navigate")
}

// NavigateReplace 替换当前路径（不增加栈深度）。
func NavigateReplace(ctx *internal.Context, path string, opts ...NavigateOption) {
	st := getRouterState(ctx)
	if st == nil || len(path) > maxRouterPathBytes {
		return
	}
	var navOpts navigateOpts
	for _, opt := range opts {
		opt(&navOpts)
	}
	st.pendingNav = &pendingNavigation{
		path:   path,
		action: navReplace,
		opts:   navOpts,
	}
	ctx.RequestRedrawReason("router.navigate")
}

// NavigateBack 返回上一页。
func NavigateBack(ctx *internal.Context, opts ...NavigateOption) {
	st := getRouterState(ctx)
	if st == nil || len(st.stack) <= 1 {
		return
	}
	var navOpts navigateOpts
	for _, opt := range opts {
		opt(&navOpts)
	}
	// 目标是栈的倒数第二个
	target := st.stack[len(st.stack)-2].path
	st.pendingNav = &pendingNavigation{
		path:   target,
		action: navPop,
		opts:   navOpts,
	}
	ctx.RequestRedrawReason("router.navigate")
}

// UseNavigate 返回绑定当前上下文的导航函数。
func UseNavigate(ctx *internal.Context) NavigateFunc {
	return func(path string, opts ...NavigateOption) {
		Navigate(ctx, path, opts...)
	}
}

// UseLocation 返回当前路由位置。
func UseLocation(ctx *internal.Context) *Location {
	if view := getRouteView(ctx); view != nil {
		return locationFromPath(view.path)
	}
	st := getRouterState(ctx)
	if st == nil {
		return &Location{}
	}
	entry := st.currentEntry()
	if entry == nil {
		return &Location{}
	}
	return locationFromPath(entry.path)
}

// UseParams 返回当前路由参数。
func UseParams(ctx *internal.Context) *Params {
	if view := getRouteView(ctx); view != nil {
		params := view.params
		return &params
	}
	st := getRouterState(ctx)
	if st == nil {
		return &Params{}
	}
	entry := st.currentEntry()
	if entry == nil {
		return &Params{}
	}
	return &entry.params
}

func locationFromPath(fullPath string) *Location {
	path, query := extractQueryParams(fullPath)
	loc := &Location{
		Path:     fullPath,
		Pathname: path,
	}
	if len(query) > 0 {
		loc.QueryParams = make(map[string]string, len(query))
		for k, v := range query {
			loc.QueryParams[k] = v
		}
	}
	return loc
}

// UseRoute returns metadata for the currently matched route declaration.
func UseRoute(ctx *internal.Context) *RouteInfo {
	if view := getRouteView(ctx); view != nil {
		info := cloneRouteInfo(view.route)
		return &info
	}
	st := getRouterState(ctx)
	if st == nil {
		return &RouteInfo{}
	}
	entry := st.currentEntry()
	if entry == nil || entry.routeIndex < 0 || entry.routeIndex >= len(st.routes) {
		return &RouteInfo{}
	}
	info := routeInfoFromRoute(st.routes[entry.routeIndex], true)
	return &info
}

func routeInfoFromRoute(route Route, matched bool) RouteInfo {
	return RouteInfo{
		Path:    route.Path,
		Name:    route.Name,
		Title:   route.Title,
		Meta:    cloneRouteMeta(route.Meta),
		Matched: matched,
	}
}

func cloneRouteInfo(info RouteInfo) RouteInfo {
	info.Meta = cloneRouteMeta(info.Meta)
	return info
}

func cloneRouteMeta(meta map[string]any) map[string]any {
	if len(meta) == 0 {
		return nil
	}
	out := make(map[string]any, len(meta))
	for key, value := range meta {
		out[key] = value
	}
	return out
}

// CurrentPath 返回当前路由路径。
func CurrentPath(ctx *internal.Context) string {
	if view := getRouteView(ctx); view != nil {
		return view.path
	}
	st := getRouterState(ctx)
	if st == nil {
		return ""
	}
	entry := st.currentEntry()
	if entry == nil {
		return ""
	}
	return entry.path
}

// CanGoBack 返回是否可以返回上一页。
func CanGoBack(ctx *internal.Context) bool {
	if view := getRouteView(ctx); view != nil {
		return view.canGoBack
	}
	st := getRouterState(ctx)
	if st == nil {
		return false
	}
	return len(st.stack) > 1
}

// StackDepth 返回导航栈深度。
func StackDepth(ctx *internal.Context) int {
	if view := getRouteView(ctx); view != nil {
		return view.stackDepth
	}
	st := getRouterState(ctx)
	if st == nil {
		return 0
	}
	return len(st.stack)
}
