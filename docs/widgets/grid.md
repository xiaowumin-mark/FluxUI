<!-- fluxui-doc-meta
{
  "id": "grid",
  "title": "Grid 网格布局",
  "category": "布局系统",
  "order": 120,
  "summary": "Grid 用于多列网格排布。",
  "example": { "id": "grid_basic" },
  "apis": [
    "Grid(columns int, children ...Widget) Widget",
    "GridView(count int, columns int, itemBuilder func(ctx *Context, index int) Widget, opts ...GridOption) Widget",
    "GridGap(rowGap, colGap float32) GridOption",
    "GridPadding(insets Insets) GridOption",
    "GridMinItemWidth(width float32) GridOption",
    "GridOnReachEnd(fn func(ctx *Context)) GridOption"
  ]
}
-->

# Grid 网格布局

## 组件说明
Grid 适合图库、卡片矩阵、能力入口面板等“多列同级内容”场景。

## 使用方法
- 固定内容可用 `Grid(columns, children...)`。
- 大量动态内容可用 `GridView`。
- 使用 `GridGap` 控制行列间距。
- 使用 `GridMinItemWidth` 可以在宽度变化时调整实际列数。
- `GridView` 的分页加载可使用 `GridOnReachEnd`。

## Lifecycle / React-style 说明

- 当前仍以 legacy `Widget` 路径作为稳定实现。
- `Grid` 是固定 children 的布局包装，生命周期风险低；本批次仍不冻结 `GridElement` API 名称。
- `GridView` 内部按行使用 Gio list state 承载动态网格，滚动位置、viewport、触底回调去重状态都属于 host/list state，不应放入 component HookSlot。
- `GridView` 的 `itemBuilder` 仍按 index 构建 cell；动态插入、删除或重排时，业务状态需要由外部稳定 id 或后续 Element `Key` 规则承担。
- `GridMinItemWidth` 会根据约束改变实际列数，可能让同一个 index 映射到不同 row/column；后续 Element API 需要先明确 cell identity 与 host-state 复用规则。
- 在 Batch 5 期间，文档保持 legacy-first，只记录固定布局与动态 GridView 的边界，不新增 React-style snippet，也不修改 docs_browser 示例映射。

## 使用示例
```go
ui.Grid(
    3,
    ui.Text("A"),
    ui.Text("B"),
    ui.Text("C"),
)
```
