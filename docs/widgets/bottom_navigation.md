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
    "FromWidget(w Widget) Element",
    "BottomNavOnChange(fn func(ctx *Context, key string)) BottomNavOption",
    "BottomNavBackground(col color.NRGBA) BottomNavOption",
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

## Host-state / React-style 状态

- 当前 `BottomNavigation` 仍以 legacy `Widget` + `BottomNavRef` 作为稳定实现。
- React-style root 中可先使用 `FromWidget(ui.BottomNavigation(...))` 桥接到底层 legacy navigation。
- `active` key、`BottomNavOnChange`、`BottomNavAttachRef` 和页面切换状态仍由调用方/legacy path 管理，不自动迁入 component HookSlot。
- 本阶段不冻结 `BottomNavigationElement` API 名称；后续如引入 Element wrapper，需要先明确 nav item identity、ref 命令、页面 subtree key/remount 和 host-state 归属。
