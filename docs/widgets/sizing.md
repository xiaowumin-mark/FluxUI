<!-- fluxui-doc-meta
{
  "id": "sizing",
  "title": "Fixed / Fill 尺寸控制",
  "category": "布局系统",
  "order": 90,
  "summary": "Fixed 和 Fill 系列用于控制组件占位行为。",
  "example": { "id": "sizing_basic" },
  "apis": [
    "FixedWidth(width float32, child Widget) Widget",
    "FixedHeight(height float32, child Widget) Widget",
    "FixedSize(width, height float32, child Widget) Widget",
    "FillWidth(child Widget) Widget",
    "FillHeight(child Widget) Widget",
    "Fill(child Widget) Widget",
    "FromWidget(w Widget) Element"
  ]
}
-->

# Fixed / Fill 尺寸控制

## 组件说明
该组组件用于显式控制控件尺寸和占满行为，常用于边栏、状态栏、主内容区等复杂布局。

## 使用方法
- 固定边栏：`FixedWidth`
- 固定高度工具栏：`FixedHeight`
- 主区域填充：`Fill` / `Expanded`

## 使用示例

### Legacy Widget
旧 `Fixed` / `Fill` / `Widget` 写法继续可用：

```go
ui.Row(
    ui.FixedWidth(120, ui.Text("固定宽度")),
    ui.Padding(
        ui.Insets{Left: 8},
        ui.Expanded(ui.Text("剩余空间")),
    ),
)
```

### React-style Element
当前阶段 `Fixed` / `Fill` / `Expanded` 的原生 Element wrapper 尚未冻结；React-style root 中可先用 `FromWidget` 桥接旧尺寸 helper：

```go
func App(ctx *ui.Context) ui.Element {
    return ui.FromWidget(ui.Row(
        ui.FixedWidth(120, ui.Text("固定宽度")),
        ui.Padding(
            ui.Insets{Left: 8},
            ui.Expanded(ui.Text("剩余空间")),
        ),
    ))
}
```

也可以先用基础 Element wrapper 表达固定空白和简单布局：

```go
func App(ctx *ui.Context) ui.Element {
    return ui.RowElement(
        ui.ContainerElement(
            ui.Style{Padding: ui.All(12), Background: ui.NRGBA(240, 244, 248, 255)},
            ui.TextElement("固定区域"),
        ),
        ui.SpacerElement(8, 0),
        ui.TextElement("后续内容"),
    )
}
```
