<!-- fluxui-doc-meta
{
  "id": "spacer",
  "title": "Spacer 间距占位",
  "category": "布局系统",
  "order": 70,
  "summary": "Spacer/HSpacer/VSpacer 用于显式插入空白间距。",
  "example": { "id": "spacer_basic" },
  "apis": [
    "Spacer(width, height float32) Widget",
    "SpacerElement(width, height float32) Element",
    "HSpacer(width float32) Widget",
    "HSpacerElement(width float32) Element",
    "VSpacer(height float32) Widget",
    "VSpacerElement(height float32) Element"
  ]
}
-->

# Spacer 间距占位

## 组件说明
Spacer 是布局阶段的空白组件，用于控制横向和纵向空隙，适合提升布局可读性。

## 使用方法
- 横向间距优先 `HSpacer`。
- 纵向间距优先 `VSpacer`。
- 双向固定占位可用 `Spacer(width, height)`。

## 使用示例

### React-style Element
React-style 写法使用 `SpacerElement(width, height)` 显式声明空白区域：

```go
func App(ctx *ui.Context) ui.Element {
    return ui.RowElement(
        ui.TextElement("左"),
        ui.HSpacerElement(16),
        ui.TextElement("右"),
    )
}
```

### Legacy Widget
旧 `ui.Spacer` / `HSpacer` / `VSpacer` 写法继续可用：

```go
ui.Row(
    ui.Text("左"),
    ui.HSpacer(16),
    ui.Text("右"),
)
```
