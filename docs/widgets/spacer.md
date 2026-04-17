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
    "HSpacer(width float32) Widget",
    "VSpacer(height float32) Widget"
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
```go
ui.Row(
    ui.Text("左"),
    ui.HSpacer(16),
    ui.Text("右"),
)
```
