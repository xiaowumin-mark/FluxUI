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
- `ListItemSpacing` 和 `ListPadding` 用于控制项间距与列表内边距。
- 大列表和分页场景要让 item 内容尽量由稳定数据源驱动，避免只依赖 index 保存业务状态。

## Lifecycle / React-style 说明

- 当前仍以 legacy `Widget` + Gio list state 作为兼容实现。
- `ListView` 的滚动位置、虚拟窗口、触底回调去重状态和 itemBuilder 调用时机都属于 host/list state，不应放入 component HookSlot。
- `itemBuilder` 当前按 index 生成 child context；动态插入、删除或重排时，业务状态需要由外部稳定 id 或后续 Element `Key` 规则承担。
- `ListOnReachEnd` 依赖滚动位置和 viewport 判断，后续若引入 Element wrapper，需要先明确触底回调的去重、重置和 unmount cleanup 语义。
- 在 Batch 5 期间，文档保持 legacy-first，只记录虚拟列表与 item identity 边界，不冻结 `ListViewElement` API 名称，也不新增 React-style snippet。

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
