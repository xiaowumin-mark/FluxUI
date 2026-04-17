<!-- fluxui-doc-meta
{
  "id": "container",
  "title": "Container 容器",
  "category": "布局系统",
  "order": 50,
  "summary": "Container 提供背景、圆角、边距等样式能力。",
  "example": { "id": "container_basic" },
  "apis": [
    "Container(st Style, child Widget) Widget",
    "type Style struct { Background color.NRGBA; Padding Insets; Margin Insets; Radius float32 }"
  ]
}
-->

# Container 容器

## 组件说明
Container 是样式容器，负责背景色、内外边距、圆角等视觉外观。业务内容应该作为 child 放入容器，不应在容器里塞业务逻辑。

## 使用方法
- 设置统一视觉块：背景 + 圆角 + 内边距。
- 卡片、面板、状态块都建议由 `Container` 或 `Card` 承载。

## 使用示例
```go
ui.Container(
    ui.Style{
        Background: ui.NRGBA(30, 136, 229, 255),
        Padding:    ui.All(12),
        Radius:     8,
    },
    ui.Text("容器内容", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
)
```
