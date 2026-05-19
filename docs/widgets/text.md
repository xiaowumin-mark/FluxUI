<!-- fluxui-doc-meta
{
  "id": "text",
  "title": "Text 文本",
  "category": "基础显示",
  "order": 110,
  "summary": "Text 用于展示静态或动态文本内容。",
  "example": { "id": "text_basic" },
  "apis": [
    "Text(content string, opts ...TextOption) Widget",
    "TextElement(content string, opts ...TextOption) Element",
    "TextSize(size float32) TextOption",
    "TextColor(value color.NRGBA) TextOption",
    "TextAlign(alignment TextAlignment) TextOption"
  ]
}
-->

# Text 文本

## 组件说明
Text 是最基础展示组件，支持字号、颜色和对齐控制。所有文案都建议走统一文本层，避免在业务组件里散落样式。

## 使用方法
- 标题、正文、说明分别定义字号规范。
- 强调文本建议结合主题色，而不是硬编码随机颜色。

## 使用示例

### Legacy Widget
旧 `ui.Text` / `Widget` 写法继续可用：

```go
ui.Text(
    "Hello FluxUI",
    ui.TextSize(18),
    ui.TextColor(ui.NRGBA(30, 41, 59, 255)),
)
```

### React-style Element
新代码可在 `RunElement` root 下返回 `TextElement`：

```go
func App(ctx *ui.Context) ui.Element {
    return ui.TextElement(
        "Hello FluxUI",
        ui.TextSize(18),
        ui.TextColor(ui.NRGBA(30, 41, 59, 255)),
    )
}
```
