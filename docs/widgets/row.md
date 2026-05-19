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
    "RowElement(children ...Element) Element",
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

### Legacy Widget
旧 `ui.Row` / `Widget` 写法继续可用：

```go
ui.Row(
    ui.Text("标题"),
    ui.Padding(ui.Insets{Left: 8}, ui.Text("副标题")),
    ui.Expanded(ui.Spacer(0, 0)),
    ui.Button(ui.Text("操作")),
)
```

### React-style Element
新代码可在 `RunElement` root 下返回 `RowElement`。当前阶段 `Expanded` / `Flexed` 的原生 Element wrapper 尚未冻结，基础横向布局优先用 Element wrapper：

```go
func Toolbar(ctx *ui.Context) ui.Element {
    return ui.RowElement(
        ui.TextElement("标题"),
        ui.PaddingElement(ui.Insets{Left: 8}, ui.TextElement("副标题")),
        ui.SpacerElement(16, 0),
        ui.TextElement("操作"),
    )
}
```
