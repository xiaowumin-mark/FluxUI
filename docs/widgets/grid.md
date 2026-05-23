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
    "GridElement(columns int, children ...Element) Element",
    "GridView(count int, columns int, itemBuilder func(ctx *Context, index int) Widget, opts ...GridOption) Widget",
    "GridViewElement(count int, columns int, itemBuilder func(ctx *Context, index int) Element, opts ...GridOption) Element",
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

## Lifecycle / React-style Element

- `GridElement` 适合固定 children 的 Element 网格布局。
- `GridViewElement` 适合动态网格列表；滚动位置、viewport 和触底回调去重状态仍属于底层 grid/list host state。
- `GridViewElement` 的 `itemBuilder` 仍按 index 构建 cell；动态重排时业务状态应由稳定 id 或 `Key` 显式表达。
- `GridMinItemWidth` 会根据约束改变实际列数，可能让同一个 index 映射到不同 row/column；不要把 row/column 当成业务 identity。

## 使用示例
```go
ui.Grid(
    3,
    ui.Text("A"),
    ui.Text("B"),
    ui.Text("C"),
)
```

### React-style Element

```go
func CapabilityGrid(ctx *ui.Context) ui.Element {
    return ui.GridElement(
        3,
        ui.TextElement("A"),
        ui.TextElement("B"),
        ui.TextElement("C"),
    )
}
```
