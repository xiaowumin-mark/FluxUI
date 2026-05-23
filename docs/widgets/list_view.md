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
    "ListViewElement(count int, itemBuilder func(ctx *Context, index int) Element, opts ...ListOption) Element",
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
- `ListItemSpacing` 和 `ListPadding` 用于控制项间距与列表内边距。
- 大列表和分页场景要让 item 内容尽量由稳定数据源驱动，避免只依赖 index 保存业务状态。

## Lifecycle / React-style Element

- `ListViewElement` 已可在 `RunElement` root 下直接使用。
- 滚动位置、虚拟窗口和触底回调去重状态仍属于底层 list host state，不放入 component HookSlot。
- `itemBuilder` 仍按 index 构建；动态插入、删除或重排时，业务状态应由外部稳定 id 或 `Key` 显式表达。

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

### React-style Element

```go
func TodoList(ctx *ui.Context) ui.Element {
    items := []string{"设计", "实现", "测试"}
    return ui.ListViewElement(
        len(items),
        func(ctx *ui.Context, index int) ui.Element {
            return ui.Key(items[index], ui.TextElement(items[index]))
        },
        ui.ListItemSpacing(6),
    )
}
```
