<!-- fluxui-doc-meta
{
  "id": "tabs",
  "title": "Tabs 标签栏",
  "category": "导航组件",
  "order": 400,
  "summary": "Tabs 用于同级视图切换。",
  "example": { "id": "tabs_basic" },
  "apis": [
    "Tabs(active string, items []TabItem, opts ...TabsOption) Widget",
    "FromWidget(w Widget) Element",
    "TabsOnChange(fn func(ctx *Context, key string)) TabsOption",
    "TabsScrollable(scrollable bool) TabsOption",
    "TabsIndicatorColor(col color.NRGBA) TabsOption",
    "TabsTextColor(col color.NRGBA) TabsOption",
    "TabsActiveTextColor(col color.NRGBA) TabsOption",
    "NewTabsRef() *TabsRef",
    "TabsAttachRef(ref *TabsRef) TabsOption",
    "(*TabsRef).SetActive(key string)"
  ]
}
-->

# Tabs 标签栏

## 组件说明
Tabs 用于在同一层级内容中切换不同子页面，适合文档页、设置页、详情页多标签场景。

## 使用方法
- `active` 标识当前选中标签 key。
- `items` 定义可切换标签集合。
- 通过 `TabsOnChange` 回写状态。
- 需要外部命令式切换标签时，使用 `TabsAttachRef` + `SetActive`。

## 使用示例
```go
active := ui.State[string](ctx)
ui.Tabs(
    active.Value(),
    []ui.TabItem{
        {Key: "overview", Label: "Overview"},
        {Key: "api", Label: "API"},
        {Key: "example", Label: "Example"},
    },
    ui.TabsOnChange(func(ctx *ui.Context, key string) {
        active.Set(key)
    }),
)
```

## Host-state / React-style 状态

- 当前 `Tabs` 仍以 legacy `Widget` + `TabsRef` 作为稳定实现。
- React-style root 中可先使用 `FromWidget(ui.Tabs(...))` 桥接到 Element 树。
- `active` key、滚动标签栏状态、`TabsOnChange` 和 `TabsAttachRef` 仍由 legacy path / 调用方状态管理。
- 本阶段不冻结 `TabsElement` API 名称；后续如引入 Element wrapper，需要先明确 tab item identity、ref 命令、切换页面 key/remount 和 scroll host-state 归属。
