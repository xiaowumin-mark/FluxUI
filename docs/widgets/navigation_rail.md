<!-- fluxui-doc-meta
{
  "id": "navigation_rail",
  "title": "NavigationRail 导航栏",
  "category": "导航组件",
  "order": 430,
  "summary": "NavigationRail 用于中宽屏的一组主导航目的地。",
  "example": { "id": "navigation_rail_basic" },
  "apis": [
    "NavigationRail(active string, items []NavItem, opts ...NavigationRailOption) Widget",
    "NavigationRailElement(active string, items []ElementNavItem, opts ...NavigationRailOption) Element",
    "NavigationRailOnChange(fn func(ctx *Context, key string)) NavigationRailOption",
    "NavigationRailWidth(width float32) NavigationRailOption",
    "NavigationRailHeader(header Widget) NavigationRailOption",
    "NavigationRailFooter(footer Widget) NavigationRailOption",
    "NavigationRailActiveColor(col color.NRGBA) NavigationRailOption",
    "NavigationRailInactiveColor(col color.NRGBA) NavigationRailOption",
    "NavigationRailDecoration(d Decoration) NavigationRailOption"
  ]
}
-->

# NavigationRail 导航栏

NavigationRail 默认宽度 80dp，容器使用 `SurfaceContainer`。选中项显示 56x32dp pill indicator，内容色使用 `OnSecondaryContainer`。

```go
ui.NavigationRailElement(
    active.Value(),
    []ui.ElementNavItem{
        {Key: "home", Label: "Home", Icon: ui.IconElement("H")},
        {Key: "search", Label: "Search", Icon: ui.IconElement("S")},
    },
    ui.NavigationRailActiveColor(ui.NRGBA(30, 64, 175, 255)),
    ui.NavigationRailInactiveColor(ui.NRGBA(100, 116, 139, 255)),
    ui.NavigationRailOnChange(func(ctx *ui.Context, key string) { active.Set(key) }),
)
```
