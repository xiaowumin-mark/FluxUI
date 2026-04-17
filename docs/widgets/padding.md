<!-- fluxui-doc-meta
{
  "id": "padding",
  "title": "Padding 内边距",
  "category": "布局系统",
  "order": 60,
  "summary": "Padding 只负责边距，不引入额外背景样式。",
  "example": { "id": "padding_basic" },
  "apis": [
    "Padding(insets Insets, child Widget) Widget",
    "All(value float32) Insets",
    "Symmetric(vertical, horizontal float32) Insets"
  ]
}
-->

# Padding 内边距

## 组件说明
Padding 用于明确声明内容间距，避免在业务组件内部手工硬编码位置偏移。

## 使用方法
- 统一边距可用 `All`。
- 对称边距可用 `Symmetric`。
- 局部边距可直接使用 `ui.Insets{Top: 8, Left: 12}`。

## 使用示例
```go
ui.Padding(
    ui.Symmetric(12, 16),
    ui.Text("这段文本有上下 12、左右 16 的内边距"),
)
```
