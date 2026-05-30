<!-- fluxui-doc-meta
{
  "id": "chip",
  "title": "Chip 纸片",
  "category": "输入交互",
  "order": 276,
  "summary": "Chip 用于紧凑展示辅助操作、筛选项、输入项和建议项。",
  "example": { "id": "chip_basic" },
  "apis": [
    "AssistChip(label string, opts ...ChipOption) Widget",
    "FilterChip(label string, opts ...ChipOption) Widget",
    "InputChip(label string, opts ...ChipOption) Widget",
    "SuggestionChip(label string, opts ...ChipOption) Widget",
    "AssistChipElement(label string, opts ...ChipOption) Element",
    "FilterChipElement(label string, opts ...ChipOption) Element",
    "InputChipElement(label string, opts ...ChipOption) Element",
    "SuggestionChipElement(label string, opts ...ChipOption) Element",
    "ChipSelected(selected bool) ChipOption",
    "ChipDisabled(disabled bool) ChipOption",
    "ChipOnClick(fn func(ctx *Context)) ChipOption",
    "ChipLeading(leading Widget) ChipOption",
    "ChipTrailing(trailing Widget) ChipOption"
  ]
}
-->

# Chip 纸片

Chip 是 Material 3 中的紧凑型交互控件，适合辅助操作、筛选条件、输入标签和建议项。FluxUI 提供 Assist、Filter、Input、Suggestion 四种常用变体。

## 使用方法

- `AssistChipElement` 用于普通辅助操作。
- `FilterChipElement` 可通过 `ChipSelected` 展示选中状态。
- `InputChipElement` 常配合 `ChipTrailing` 显示删除或关闭按钮。
- `SuggestionChipElement` 用于推荐动作或快捷输入。

```go
ui.FilterChipElement(
    "筛选",
    ui.ChipSelected(selected.Value()),
    ui.ChipOnClick(func(ctx *ui.Context) {
        selected.Set(!selected.Value())
    }),
)
```
