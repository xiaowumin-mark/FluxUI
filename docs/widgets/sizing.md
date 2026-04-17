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
    "Fill(child Widget) Widget"
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
```go
ui.Row(
    ui.FixedWidth(120, ui.Text("固定宽度")),
    ui.Padding(
        ui.Insets{Left: 8},
        ui.Expanded(ui.Text("剩余空间")),
    ),
)
```
