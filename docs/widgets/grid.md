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
    "GridMinItemWidth(width float32) GridOption"
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

## 使用示例
```go
ui.Grid(
    3,
    ui.Text("A"),
    ui.Text("B"),
    ui.Text("C"),
)
```
