<!-- fluxui-doc-meta
{
  "id": "list_view",
  "title": "ListView 列表",
  "category": "布局系统",
  "order": 110,
  "summary": "ListView 支持高效列表渲染与触底回调。",
  "example": { "id": "list_view_basic" },
  "apis": [
    "ListView(count int, itemBuilder func(ctx *Context, index int) Widget, opts ...ListOption) Widget",
    "ListAxis(axis Axis) ListOption",
    "ListVirtualized(virtualized bool) ListOption",
    "ListItemSpacing(spacing float32) ListOption",
    "ListPadding(insets Insets) ListOption",
    "ListOnReachEnd(fn func(ctx *Context)) ListOption"
  ]
}
-->

# ListView 列表

## 组件说明
ListView 用于长列表展示，适合日志、消息流、任务清单等场景。

## 使用方法
- `count` 指定列表项数量。
- `itemBuilder` 按 index 构建每一项。
- `ListOnReachEnd` 可做分页加载。

## 使用示例
```go
ui.ListView(
    100,
    func(ctx *ui.Context, index int) ui.Widget {
        return ui.Text(fmt.Sprintf("Item %d", index))
    },
    ui.ListItemSpacing(6),
)
```
