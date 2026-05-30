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
    "FixedWidthElement(width float32, child Element) Element",
    "FixedHeight(height float32, child Widget) Widget",
    "FixedHeightElement(height float32, child Element) Element",
    "FixedSize(width, height float32, child Widget) Widget",
    "FixedSizeElement(width, height float32, child Element) Element",
    "FillWidth(child Widget) Widget",
    "FillWidthElement(child Element) Element",
    "FillHeight(child Widget) Widget",
    "FillHeightElement(child Element) Element",
    "Fill(child Widget) Widget",
    "FillElement(child Element) Element",
    "FlexedElement(weight float32, child Element) Element",
    "ExpandedElement(child Element) Element"
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

### React-style Element
新代码可在 `RunElement` root 下直接使用 sizing Element wrappers：

```go
func App(ctx *ui.Context) ui.Element {
    return ui.RowElement(
        ui.FixedWidthElement(120, ui.TextElement("固定宽度")),
        ui.PaddingElement(
            ui.Insets{Left: 8},
            ui.ExpandedElement(ui.TextElement("剩余空间")),
        ),
    )
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
