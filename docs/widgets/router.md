<!-- fluxui-doc-meta
{
  "id": "router",
  "title": "Router 路由",
  "category": "导航组件",
  "order": 430,
  "summary": "Router 提供页面级路由、动态参数、查询参数、守卫与过渡动画。",
  "example": { "id": "router_basic" },
  "apis": [
    "Router(ctx *Context, routes []Route, opts ...RouterOption) Widget",
    "type Route struct { Path string; Builder func(ctx *Context) Widget }",
    "RouterTransition(t Transition) RouterOption",
    "RouterTransitionDuration(d time.Duration) RouterOption",
    "RouterBeforeEach(fn func(ctx *Context, from, to string) bool) RouterOption",
    "RouterNotFound(builder func(ctx *Context) Widget) RouterOption",
    "Navigate(ctx *Context, path string, opts ...NavigateOption)",
    "NavigateReplace(ctx *Context, path string, opts ...NavigateOption)",
    "NavigateBack(ctx *Context, opts ...NavigateOption)",
    "WithNavTransition(t Transition) NavigateOption",
    "CurrentPath(ctx *Context) string",
    "RouteParams(ctx *Context) *RouteParamsType",
    "CanGoBack(ctx *Context) bool",
    "StackDepth(ctx *Context) int",
    "(*RouteParamsType).Path(name string) string",
    "(*RouteParamsType).Query(name string) string"
  ]
}
-->

# Router 路由

## 组件说明
Router 是 FluxUI 的页面级导航能力，支持以下核心场景：

- 静态路由：`/`、`/settings`
- 动态路由：`/users/:id`
- 查询参数：`/users/u1001?tab=profile`
- 栈式导航：前进、替换、返回
- 路由守卫：导航前统一拦截
- 404 兜底页面：未匹配路径处理
- 页面过渡：淡入淡出、左右滑动

## 基础用法

```go
routes := []ui.Route{
    {
        Path: "/",
        Builder: func(ctx *ui.Context) ui.Widget {
            return ui.Text("首页")
        },
    },
    {
        Path: "/users/:id",
        Builder: func(ctx *ui.Context) ui.Widget {
            params := ui.RouteParams(ctx)
            return ui.Text("用户ID: " + params.Path("id"))
        },
    },
}

router := ui.Router(
    ctx,
    routes,
    ui.RouterTransition(ui.TransitionSlideLeft),
    ui.RouterNotFound(func(ctx *ui.Context) ui.Widget {
        return ui.Text("404 Not Found")
    }),
)
```

## 导航操作

- `Navigate`: 入栈跳转，保留当前页作为返回栈
- `NavigateReplace`: 替换当前栈顶页，不增加栈深
- `NavigateBack`: 返回上一页（栈深 > 1 时有效）

```go
ui.Navigate(ctx, "/users/u1002?tab=profile")
ui.NavigateReplace(ctx, "/users/u1002?tab=activity")
ui.NavigateBack(ctx)
```

## 路由守卫

守卫会在每次导航前调用，返回 `false` 可阻止跳转：

```go
ui.RouterBeforeEach(func(ctx *ui.Context, from, to string) bool {
    if to == "/settings" && !hasPermission {
        return false
    }
    return true
})
```

## 参数读取

- `Path("id")`：读取路径参数
- `Query("tab")`：读取查询参数

```go
params := ui.RouteParams(ctx)
userID := params.Path("id")
tab := params.Query("tab")
```

## 实战示例

- 独立示例：`examples/router/main.go`
- 文档浏览器示例：`router_basic`

该示例覆盖了动态参数、query、守卫、404、`NavigateReplace`、`NavigateBack` 以及过渡动画切换。
