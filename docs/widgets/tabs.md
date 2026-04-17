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
    "TabsOnChange(fn func(ctx *Context, key string)) TabsOption",
    "TabsScrollable(scrollable bool) TabsOption",
    "TabsIndicatorColor(col color.NRGBA) TabsOption",
    "TabsTextColor(col color.NRGBA) TabsOption",
    "TabsActiveTextColor(col color.NRGBA) TabsOption"
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
