<!-- fluxui-doc-meta
{
  "id": "navigation_drawer",
  "title": "NavigationDrawer 导航抽屉",
  "category": "导航组件",
  "order": 440,
  "summary": "NavigationDrawer 用于宽屏侧栏导航，提供 MD3 drawer item 选中态和 ripple。",
  "example": { "id": "navigation_drawer_basic" },
  "apis": [
    "NavigationDrawer(active string, items []NavItem, opts ...NavigationDrawerOption) Widget",
    "NavigationDrawerElement(active string, items []ElementNavItem, opts ...NavigationDrawerOption) Element",
    "NavigationDrawerOnChange(fn func(ctx *Context, key string)) NavigationDrawerOption",
    "NavigationDrawerWidth(width float32) NavigationDrawerOption",
    "NavigationDrawerHeader(header Widget) NavigationDrawerOption",
    "NavigationDrawerFooter(footer Widget) NavigationDrawerOption",
    "NavigationDrawerDecoration(d Decoration) NavigationDrawerOption"
  ]
}
-->

# NavigationDrawer 导航抽屉

NavigationDrawer 默认宽度 360dp，容器使用 `SurfaceContainerLow`。选中项使用 `SecondaryContainer` 胶囊背景，item 高度为 56dp。

```go
ui.NavigationDrawerElement(
    active.Value(),
    []ui.ElementNavItem{
        {Key: "inbox", Label: "Inbox", Icon: ui.IconElement("I")},
        {Key: "sent", Label: "Sent", Icon: ui.IconElement("S")},
    },
    ui.NavigationDrawerOnChange(func(ctx *ui.Context, key string) { active.Set(key) }),
)
```
