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
    "AssistChipWithSlots(label Widget, opts ...ChipOption) Widget",
    "FilterChipWithSlots(label Widget, opts ...ChipOption) Widget",
    "InputChipWithSlots(label Widget, opts ...ChipOption) Widget",
    "SuggestionChipWithSlots(label Widget, opts ...ChipOption) Widget",
    "AssistChipElement(label string, opts ...ChipOption) Element",
    "FilterChipElement(label string, opts ...ChipOption) Element",
    "InputChipElement(label string, opts ...ChipOption) Element",
    "SuggestionChipElement(label string, opts ...ChipOption) Element",
    "AssistChipElementWithSlots(label Element, opts ...ChipOption) Element",
    "FilterChipElementWithSlots(label Element, opts ...ChipOption) Element",
    "InputChipElementWithSlots(label Element, opts ...ChipOption) Element",
    "SuggestionChipElementWithSlots(label Element, opts ...ChipOption) Element",
    "ChipWithSlots(label Widget, opts ...ChipOption) Widget",
    "ChipElementWithSlots(label Element, opts ...ChipOption) Element",
    "ChipSelected(selected bool) ChipOption",
    "ChipDisabled(disabled bool) ChipOption",
    "ChipSoftDisabled(disabled bool) ChipOption",
    "ChipElevated(elevated bool) ChipOption",
    "ChipRemovable(removable bool) ChipOption",
    "ChipOnClick(fn func(ctx *Context)) ChipOption",
    "ChipOnRemove(fn func(ctx *Context)) ChipOption",
    "ChipLeading(leading Widget) ChipOption",
    "ChipTrailing(trailing Widget) ChipOption",
    "ChipBackground(col color.NRGBA) ChipOption",
    "ChipForeground(col color.NRGBA) ChipOption",
    "ChipDecoration(d Decoration) ChipOption"
  ]
}
-->

# Chip 纸片

Chip 是 Material 3 中的紧凑型交互控件，适合辅助操作、筛选条件、输入标签和建议项。FluxUI 提供 Assist、Filter、Input、Suggestion 四种常用变体。

## 使用方法

- `AssistChipElement` 用于普通辅助操作，可用 `ChipElevated(true)` 切换 elevated 样式。
- `FilterChipElement` 可通过 `ChipSelected` 展示选中状态，默认会显示选中 check 图标。
- `InputChipElement` 默认带 remove affordance；`ChipOnRemove` 绑定删除回调。
- `SuggestionChipElement` 用于推荐动作或快捷输入，也支持 elevated 和 leading icon。
- `ChipSoftDisabled(true)` 使用 disabled 视觉但仍保持可聚焦，适合表单中需要键盘可达的软禁用项。
- `AssistChipWithSlots` / `FilterChipWithSlots` / `InputChipWithSlots` / `SuggestionChipWithSlots` 允许 label 自己组织图标、文本和布局。
- `ChipBackground`、`ChipForeground`、`ChipDecoration` 用于局部覆盖 chip 视觉样式。

```go
ui.RowElement(
    ui.AssistChipElement("Assist chip with icon", ui.ChipLeading(ui.Icon("info", ui.IconSize(18)))),
    ui.FilterChipElement(
        "Filter chip",
        ui.ChipSelected(selected.Value()),
        ui.ChipOnClick(func(ctx *ui.Context) {
            selected.Set(!selected.Value())
        }),
    ),
    ui.InputChipElement("Input chip", ui.ChipOnRemove(func(ctx *ui.Context) {
        removeTag()
    })),
    ui.SuggestionChipElement("Suggestion chip", ui.ChipElevated(true)),
)
```

```go
ui.FilterChipElementWithSlots(
    ui.TextElement("Slot filter"),
    ui.ChipLeading(ui.Icon("label", ui.IconSize(18))),
    ui.ChipSelected(true),
    ui.ChipRemovable(true),
)
```
