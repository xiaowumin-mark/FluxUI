<!-- fluxui-doc-meta
{
  "id": "row",
  "title": "Row 横向布局",
  "category": "布局系统",
  "order": 10,
  "summary": "Row 按从左到右顺序布局子组件。",
  "example": { "id": "row_basic" },
  "apis": [
    "Row(children ...Widget) Widget",
    "Flexed(weight float32, child Widget) Widget",
    "Expanded(child Widget) Widget"
  ]
}
-->

# Row 横向布局

## 组件说明
Row 是 FluxUI 最常用的横向布局容器。你可以把标题、工具按钮、状态信息放在同一行中，并通过 `Expanded` 把剩余空间分配给主内容。

## 使用方法
- 固定内容直接放在 `Row(...)` 中。
- 需要弹性拉伸的内容使用 `Expanded(...)` 或 `Flexed(...)` 包裹。
- 组件间距推荐通过 `Padding` 或 `HSpacer` 明确声明。

## 使用示例
```go
ui.Row(
    ui.Text("标题"),
    ui.Padding(ui.Insets{Left: 8}, ui.Text("副标题")),
    ui.Expanded(ui.Spacer(0, 0)),
    ui.Button(ui.Text("操作")),
)
```
