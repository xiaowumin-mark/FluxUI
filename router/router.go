package router

import (
	"image"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/widget"

	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// Route 定义一条路由规则。
type Route struct {
	Path    string
	Builder func(ctx *internal.Context) widget.Widget
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

	// 启动过渡
	if trans != TransitionNone && currentPath != fullPath {
		s.transition = transitionState{
			active:     true,
			from:       currentPath,
			to:         fullPath,
			transition: trans,
			startTime:  ctx.Now(),
			duration:   s.config.transitionDuration,
			progress:   0,
		}
	}

	switch action {
	case navPush:
		s.stack = append(s.stack, entry)
	case navReplace:
		if len(s.stack) > 0 {
			s.stack[len(s.stack)-1] = entry
		} else {
			s.stack = []stackEntry{entry}
		}
	case navPop:
		if len(s.stack) > 1 {
			s.stack = s.stack[:len(s.stack)-1]
		}
	}

	ctx.RequestRedraw()
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
		if elapsed >= st.transition.duration {
			st.transition.active = false
			st.transition.progress = 1
		} else {
			st.transition.progress = float32(elapsed) / float32(st.transition.duration)
			// EaseOut 缓动
			p := st.transition.progress
			st.transition.progress = 1 - (1-p)*(1-p)
			next.RequestRedraw()
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
		var page widget.Widget
		withRouteView(ctx, view, func() {
			scopeCtx := ctx.Scope(fullPath)
			page = routes[idx].Builder(scopeCtx)
		})
		if page != nil {
			return page, view, true
		}
	}

	if notFound != nil {
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

func routeLayoutScope(path string) string {
	if path == "" {
		return "route:__empty__"
	}
	return "route:" + path
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
				scopeName = routeLayoutScope(view.path)
			}
			pageCtx = ctx.Scope(scopeName)
		}
		dims = page.Layout(pageCtx)
	})
	return dims
}

// layoutWithTransition 在过渡期间布局页面。
// 旧页面透明度从 1 递减到 0，新页面按过渡类型入场。
func layoutWithTransition(
	ctx *internal.Context,
	fromPage widget.Widget,
	fromView *routeView,
	toPage widget.Widget,
	toView *routeView,
	ts transitionState,
) layout.Dimensions {
	var merged layout.Dimensions

	oldOpacity := 1 - ts.progress
	if oldOpacity < 0 {
		oldOpacity = 0
	}
	if oldOpacity > 1 {
		oldOpacity = 1
	}

	if fromPage != nil && oldOpacity > 0 {
		opacityLayer := paint.PushOpacity(ctx.Gtx.Ops, oldOpacity)
		merged = layoutPageWithRouteView(ctx, fromPage, fromView)
		opacityLayer.Pop()
	}

	incoming := layoutIncomingPage(ctx, toPage, toView, ts)
	if incoming.Size.X > merged.Size.X {
		merged.Size.X = incoming.Size.X
	}
	if incoming.Size.Y > merged.Size.Y {
		merged.Size.Y = incoming.Size.Y
	}

	return merged
}

func layoutIncomingPage(ctx *internal.Context, page widget.Widget, view *routeView, ts transitionState) layout.Dimensions {
	gtx := ctx.Gtx
	maxWidth := float32(gtx.Constraints.Max.X)

	switch ts.transition {
	case TransitionFade:
		opacity := ts.progress
		if opacity < 0 {
			opacity = 0
		}
		if opacity > 1 {
			opacity = 1
		}
		layer := paint.PushOpacity(gtx.Ops, opacity)
		dims := layoutPageWithRouteView(ctx, page, view)
		layer.Pop()
		return dims

	case TransitionSlideLeft:
		offset := int(maxWidth * (1 - ts.progress))
		return layoutWithOffset(ctx, view, page, offset, 0)

	case TransitionSlideRight:
		offset := int(-maxWidth * (1 - ts.progress))
		return layoutWithOffset(ctx, view, page, offset, 0)

	default:
		return layoutPageWithRouteView(ctx, page, view)
	}
}

// layoutWithOffset 带偏移量渲染。
func layoutWithOffset(ctx *internal.Context, view *routeView, page widget.Widget, dx, dy int) layout.Dimensions {
	gtx := ctx.Gtx
	constraints := gtx.Constraints

	defer op.Offset(image.Pt(dx, dy)).Push(gtx.Ops).Pop()
	// 裁剪到原始约束范围
	defer clip.Rect{Max: constraints.Max}.Push(gtx.Ops).Pop()

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
	params     Params
	canGoBack  bool
	stackDepth int
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
	if st == nil {
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
	ctx.RequestRedraw()
}

// NavigateReplace 替换当前路径（不增加栈深度）。
func NavigateReplace(ctx *internal.Context, path string, opts ...NavigateOption) {
	st := getRouterState(ctx)
	if st == nil {
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
	ctx.RequestRedraw()
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
	ctx.RequestRedraw()
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

// RouteParams 返回当前路由的参数。
func RouteParams(ctx *internal.Context) *Params {
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
