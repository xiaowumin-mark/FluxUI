<!-- fluxui-doc-meta
{
  "id": "tooltip",
  "title": "Tooltip 工具提示",
  "category": "反馈组件",
  "order": 334,
  "summary": "Tooltip 在控件悬停或聚焦时显示简短提示。",
  "example": { "id": "tooltip_basic" },
  "apis": [
    "Tooltip(label string, child Widget, opts ...TooltipOption) Widget",
    "TooltipElement(label string, child Element, opts ...TooltipOption) Element",
    "TooltipDisabled(disabled bool) TooltipOption",
    "TooltipOffset(offset float32) TooltipOption",
    "TooltipDecoration(d Decoration) TooltipOption",
    "TooltipTextColor(col color.NRGBA) TooltipOption"
  ]
}
-->

# Tooltip 工具提示

Tooltip 用于包裹目标控件，在悬停或聚焦时显示一个紧凑的 Material 3 提示标签。适合图标按钮、工具栏按钮等缺少可见文字的控件。

## 使用方法

- 第一个参数是提示文本。
- 第二个参数是被包裹的目标控件。
- 可通过 `TooltipOffset`、`TooltipDecoration` 和 `TooltipTextColor` 调整距离和样式。

```go
ui.TooltipElement(
    "刷新",
    ui.IconButtonElement(ui.IconElement("R")),
)
```
