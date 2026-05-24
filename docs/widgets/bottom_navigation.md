<!-- fluxui-doc-meta
{
  "id": "bottom_navigation",
  "title": "BottomNavigation 底部导航",
  "category": "导航组件",
  "order": 420,
  "summary": "BottomNavigation 用于底部一级页面切换。",
  "example": { "id": "bottom_navigation_basic" },
  "apis": [
    "BottomNavigation(active string, items []NavItem, opts ...BottomNavOption) Widget",
    "BottomNavigationElement(active string, items []ElementNavItem, opts ...BottomNavOption) Element",
    "type ElementNavItem struct { Key string; Label string; Icon Element }",
    "BottomNavOnChange(fn func(ctx *Context, key string)) BottomNavOption",
    "BottomNavBackground(col color.NRGBA) BottomNavOption",
    "BottomNavDecoration(d Decoration) BottomNavOption",
    "BottomNavActiveColor(col color.NRGBA) BottomNavOption",
    "BottomNavInactiveColor(col color.NRGBA) BottomNavOption",
    "BottomNavAlignmentOf(alignment BottomNavAlignment) BottomNavOption",
    "NewBottomNavRef() *BottomNavRef",
    "BottomNavAttachRef(ref *BottomNavRef) BottomNavOption",
    "(*BottomNavRef).SetActive(key string)"
  ]
}
-->

# BottomNavigation 底部导航

## 组件说明
BottomNavigation 适用于移动端或工具型应用的一级页面切换。

## 使用方法
- `active` 保存当前页面 key。
- `items` 定义导航项。
- 在 `BottomNavOnChange` 中回写状态并切换页面内容。
- 外部程序切页可通过 `BottomNavAttachRef` 绑定后调用 `SetActive`。

## 使用示例
```go
active := ui.State[string](ctx)
ui.BottomNavigation(
    active.Value(),
    []ui.NavItem{
        {Key: "home", Label: "首页", Icon: ui.Text("H")},
        {Key: "docs", Label: "文档", Icon: ui.Text("D")},
    },
    ui.BottomNavOnChange(func(ctx *ui.Context, key string) {
        active.Set(key)
    }),
)
```

## React-style Element

- `BottomNavigationElement` 已可在 `RunElement` root 下直接使用。
- React-style 版本使用 `[]ElementNavItem`，因此每个 item 的 icon 可以是 `Element`。
- `active` key、`BottomNavOnChange`、`BottomNavAttachRef` 和页面切换状态仍由调用方/底层 navigation host 管理。

```go
func BottomTabs(ctx *ui.Context) ui.Element {
    active := ui.UseState(ctx, "home")
    return ui.BottomNavigationElement(
        active.Value(),
        []ui.ElementNavItem{
            {Key: "home", Label: "首页", Icon: ui.TextElement("H")},
            {Key: "docs", Label: "文档", Icon: ui.TextElement("D")},
        },
        ui.BottomNavOnChange(func(ctx *ui.Context, key string) { active.Set(key) }),
    )
}
```
