package router

import (
	"image"
	"image/color"
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
	path        string
	params      Params
	routeIndex  int // 对应 routes 数组下标，-1 为未匹配
}

// routerState 路由器持久化状态。
type routerState struct {
	stack      []stackEntry
	routes     []Route
	config     routerConfig
	transition transitionState
	// 挂起的导航操作
	pendingNav    *pendingNavigation
}

type pendingNavigation struct {
	path       string
	action     navAction
	opts       navigateOpts
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

	// 纯路径用于守卫
	cleanPath, _ := extractQueryParams(fullPath)

	// 路由守卫
	if s.config.beforeEach != nil {
		if !s.config.beforeEach(ctx, currentPath, cleanPath) {
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

	// 构建当前页面
	entry := st.currentEntry()
	if entry == nil {
		if w.config.notFound != nil {
			page := w.config.notFound(routerCtx)
			return page.Layout(routerCtx)
		}
		return layout.Dimensions{}
	}

	var page widget.Widget
	if entry.routeIndex >= 0 && entry.routeIndex < len(w.routes) {
		scopeCtx := routerCtx.Scope(entry.path)
		page = w.routes[entry.routeIndex].Builder(scopeCtx)
	} else if w.config.notFound != nil {
		page = w.config.notFound(routerCtx)
	} else {
		return layout.Dimensions{}
	}

	// 应用过渡动画
	if st.transition.active {
		return layoutWithTransition(routerCtx, page, st.transition)
	}

	return page.Layout(routerCtx)
}

// layoutWithTransition 在过渡期间布局页面。
func layoutWithTransition(ctx *internal.Context, page widget.Widget, ts transitionState) layout.Dimensions {
	gtx := ctx.Gtx
	constraints := gtx.Constraints
	maxWidth := float32(constraints.Max.X)

	switch ts.transition {
	case TransitionFade:
		// 淡入效果
		alpha := uint8(ts.progress * 255)
		return layoutWithAlpha(ctx, page, alpha)

	case TransitionSlideLeft:
		// 从右向左滑入
		offset := int(maxWidth * (1 - ts.progress))
		return layoutWithOffset(ctx, page, offset, 0)

	case TransitionSlideRight:
		// 从左向右滑入
		offset := int(-maxWidth * (1 - ts.progress))
		return layoutWithOffset(ctx, page, offset, 0)

	default:
		return page.Layout(ctx)
	}
}

// layoutWithAlpha 带透明度渲染。
// 在目标位置画一层半透明白色遮罩模拟淡入。
func layoutWithAlpha(ctx *internal.Context, page widget.Widget, alpha uint8) layout.Dimensions {
	gtx := ctx.Gtx
	dims := page.Layout(ctx)

	if alpha < 255 {
		// 用白色遮罩模拟淡入
		maskAlpha := 255 - alpha
		sz := dims.Size
		defer clip.Rect{Max: sz}.Push(gtx.Ops).Pop()
		paint.ColorOp{Color: color.NRGBA{R: 255, G: 255, B: 255, A: maskAlpha}}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	}

	return dims
}

// layoutWithOffset 带偏移量渲染。
func layoutWithOffset(ctx *internal.Context, page widget.Widget, dx, dy int) layout.Dimensions {
	gtx := ctx.Gtx
	constraints := gtx.Constraints

	defer op.Offset(image.Pt(dx, dy)).Push(gtx.Ops).Pop()

	// 裁剪到原始约束范围
	defer clip.Rect{Max: constraints.Max}.Push(gtx.Ops).Pop()

	return page.Layout(ctx)
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
	value := ctx.Persistent(routerStateKey, func() any {
		return &routerStateHolder{}
	})
	return value.(*routerStateHolder).state
}

type routerStateHolder struct {
	state *routerState
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
	st := getRouterState(ctx)
	if st == nil {
		return false
	}
	return len(st.stack) > 1
}

// StackDepth 返回导航栈深度。
func StackDepth(ctx *internal.Context) int {
	st := getRouterState(ctx)
	if st == nil {
		return 0
	}
	return len(st.stack)
}
