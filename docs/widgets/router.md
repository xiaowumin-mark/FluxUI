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
    "type Location struct { Path string; Pathname string; QueryParams map[string]string }",
    "type BeforeEachFunc func(ctx *Context, from, to string) bool",
    "(*Location).Query(name string) string",
    "(*Location).AllQueryParams() map[string]string",
    "RouterElement(routes ...RouteElementSpec) Element",
    "RouteElement(path string, component Component, opts ...RouteElementOption) RouteElementSpec",
    "RouteKey(key string) RouteElementOption",
    "RouteName(name string) RouteElementOption",
    "RouteTitle(title string) RouteElementOption",
    "RouteMeta(key string, value any) RouteElementOption",
    "RouteMetaMap(meta map[string]any) RouteElementOption",
    "RouteBeforeEnter(fn func(ctx *Context, from, to string) bool) RouteElementOption",
    "UseNavigate(ctx *Context) NavigateFunc",
    "UseLocation(ctx *Context) *Location",
    "UseParams(ctx *Context) *RouteParamsType",
    "UseRoute(ctx *Context) *RouteInfo",
    "RouterTransition(t Transition) RouterOption",
    "RouterTransitionDuration(d time.Duration) RouterOption",
    "RouterBeforeEach(fn func(ctx *Context, from, to string) bool) RouterOption",
    "RouterNotFound(builder func(ctx *Context) Widget) RouterOption",
    "RouterNotFoundElement(component Component) RouterOption",
    "Navigate(ctx *Context, path string, opts ...NavigateOption)",
    "NavigateReplace(ctx *Context, path string, opts ...NavigateOption)",
    "NavigateBack(ctx *Context, opts ...NavigateOption)",
    "WithNavTransition(t Transition) NavigateOption",
    "CurrentPath(ctx *Context) string",
    "CanGoBack(ctx *Context) bool",
    "StackDepth(ctx *Context) int",
    "(*RouteParamsType).Path(name string) string",
    "(*RouteParamsType).Query(name string) string",
    "(*RouteParamsType).HasParam(name string) bool",
    "(*RouteParamsType).HasQuery(name string) bool",
    "(*RouteParamsType).AllPathParams() map[string]string",
    "(*RouteParamsType).AllQueryParams() map[string]string"
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

新代码推荐使用 `RunElement` + `RouterElement` 组合：

```go
func App(ctx *ui.Context) ui.Element {
    return ui.RouterElement(
        ui.RouteElement("/", HomePage),
        ui.RouteElement("/users/:id", UserPage),
        ui.RouteElement("/settings", SettingsPage),
    )
}

func HomePage(ctx *ui.Context) ui.Element {
    navigate := ui.UseNavigate(ctx)
    location := ui.UseLocation(ctx)

    return ui.ColumnElement(
        ui.TextElement("Home"),
        ui.TextElement("Pathname: "+location.Pathname),
        ui.ButtonElement(
            ui.TextElement("Open user u1001"),
            ui.OnClick(func(ctx *ui.Context) {
                navigate("/users/u1001?tab=profile", ui.WithNavTransition(ui.TransitionSlideLeft))
            }),
        ),
    )
}

func UserPage(ctx *ui.Context) ui.Element {
    navigate := ui.UseNavigate(ctx)
    location := ui.UseLocation(ctx)
    params := ui.UseParams(ctx)

    return ui.ColumnElement(
        ui.TextElement("User detail"),
        ui.TextElement("id: "+params.Get("id")),
        ui.TextElement("tab: "+location.Query("tab")),
        ui.TextElement("Route identity follows the pattern unless RouteKey changes."),
        ui.ButtonElement(
            ui.TextElement("Home"),
            ui.OnClick(func(ctx *ui.Context) {
                navigate("/")
            }),
        ),
    )
}
```

## 兼容用法

旧 `Router` / `Navigate` API 仍然保留，参数读取统一使用 `UseParams`：

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
            params := ui.UseParams(ctx)
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

如果需要强制重建某个页面实例，可以给 `RouteElement` 加 `RouteKey`：

```go
ui.RouteElement("/users/:id", UserPage, ui.RouteKey("user-42"))
```

## 路由身份与兼容性

- 默认情况下，`RouteElement` 的组件身份按路由 pattern 复用，而不是按具体参数值重建。
- 这意味着 `/users/1`、`/users/2` 这类同 pattern 路由会复用同一个页面实例，除非显式传入 `RouteKey`。
- `RouteKey` 适合用于需要按业务对象强制重建页面的场景，例如切换不同用户详情页时需要清空局部状态。
- 这个行为与现有路由测试保持一致：pattern identity 复用、显式 key 变化 remount。
- 旧的 `Router` / `Navigate` API 仍然保留，参数读取统一使用 `UseParams`。
- `RouterElement` 目前是 React-style 对照入口，不改变 `examples/router` 和 `router_basic` 的既有行为。

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
params := ui.UseParams(ctx)
userID := params.Path("id")
tab := params.Query("tab")
```

## 实战示例

- 独立示例：`examples/router/main.go`
- React-style 对照示例：`examples/react_workspace/main.go`
- 文档浏览器示例：`router_basic`

该示例覆盖了动态参数、query、守卫、404、`NavigateReplace`、`NavigateBack` 以及过渡动画切换。
