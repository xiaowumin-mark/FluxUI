package ui

import (
	"time"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
	router "github.com/xiaowumin-mark/FluxUI/router"
	widget "github.com/xiaowumin-mark/FluxUI/widget"
)

// ──────────────────────────────
// 路由类型
// ──────────────────────────────

// Route 定义一条路由规则。
type Route = router.Route

// RouterOption 定义路由器配置项。
type RouterOption = router.Option

// NavigateOption 定义单次导航配置项。
type NavigateOption = router.NavigateOption

// NavigateFunc 定义实验性导航函数。
type NavigateFunc = router.NavigateFunc

// Location 描述当前路由位置。
type Location = router.Location

// Transition 定义页面切换动画类型。
type Transition = router.Transition

// RouteParamsType 路由参数类型。
type RouteParamsType = router.Params

// BeforeEachFunc 路由守卫函数类型。
type BeforeEachFunc = router.BeforeEachFunc

// ──────────────────────────────
// 过渡动画常量
// ──────────────────────────────

const (
	// TransitionNone 无动画，直接切换。
	TransitionNone = router.TransitionNone
	// TransitionFade 淡入淡出。
	TransitionFade = router.TransitionFade
	// TransitionSlideLeft 从右向左滑入（前进）。
	TransitionSlideLeft = router.TransitionSlideLeft
	// TransitionSlideRight 从左向右滑入（后退）。
	TransitionSlideRight = router.TransitionSlideRight
)

// ──────────────────────────────
// 路由器构建
// ──────────────────────────────

// Router 创建路由器组件。
//
//	router := ui.Router(ctx, []ui.Route{
//	    {Path: "/", Builder: homePage},
//	    {Path: "/users/:id", Builder: userPage},
//	}, ui.RouterTransition(ui.TransitionSlideLeft))
func Router(ctx *Context, routes []Route, opts ...RouterOption) Widget {
	// 转换 Builder 签名：用户传入 func(*ui.Context) ui.Widget
	// 内部需要 func(*internal.Context) widget.Widget（类型别名所以直接传即可）
	return router.New(ctx, routes, opts...)
}

// ──────────────────────────────
// 路由器配置选项
// ──────────────────────────────

// RouterTransition 设置全局默认过渡动画。
func RouterTransition(t Transition) RouterOption {
	return router.RouterTransition(t)
}

// RouterTransitionDuration 设置过渡动画时长。
func RouterTransitionDuration(d time.Duration) RouterOption {
	return router.RouterTransitionDuration(d)
}

// RouterNotFound 设置 404 页面。
func RouterNotFound(builder func(ctx *Context) Widget) RouterOption {
	return router.RouterNotFound(func(ctx *internal.Context) widget.Widget {
		return builder(ctx)
	})
}

// RouterBeforeEach 设置全局路由守卫。
func RouterBeforeEach(fn func(ctx *Context, from, to string) bool) RouterOption {
	return router.RouterBeforeEach(func(ctx *internal.Context, from, to string) bool {
		return fn(ctx, from, to)
	})
}

// WithTransition 为单次导航指定过渡动画。
func WithNavTransition(t Transition) NavigateOption {
	return router.WithTransition(t)
}

// ──────────────────────────────
// 导航操作
// ──────────────────────────────

// Navigate 导航到指定路径（push 到栈）。
func Navigate(ctx *Context, path string, opts ...NavigateOption) {
	router.Navigate(ctx, path, opts...)
}

// NavigateReplace 替换当前路径（不增加栈深度）。
func NavigateReplace(ctx *Context, path string, opts ...NavigateOption) {
	router.NavigateReplace(ctx, path, opts...)
}

// NavigateBack 返回上一页。
func NavigateBack(ctx *Context, opts ...NavigateOption) {
	router.NavigateBack(ctx, opts...)
}

// UseNavigate 返回绑定当前上下文的导航函数。
func UseNavigate(ctx *Context) NavigateFunc {
	return router.UseNavigate(ctx)
}

// UseLocation 返回当前路由位置。
func UseLocation(ctx *Context) *Location {
	return router.UseLocation(ctx)
}

// UseParams 返回当前路由参数。
func UseParams(ctx *Context) *RouteParamsType {
	return router.UseParams(ctx)
}

// ──────────────────────────────
// 路由状态查询
// ──────────────────────────────

// CurrentPath 返回当前路由路径。
func CurrentPath(ctx *Context) string {
	return router.CurrentPath(ctx)
}

// CanGoBack 返回是否可以返回上一页。
func CanGoBack(ctx *Context) bool {
	return router.CanGoBack(ctx)
}

// StackDepth 返回导航栈深度。
func StackDepth(ctx *Context) int {
	return router.StackDepth(ctx)
}
