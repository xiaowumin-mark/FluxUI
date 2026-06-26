<!-- fluxui-doc-meta
{
  "id": "floating_action_button",
  "title": "FloatingActionButton 浮动操作按钮",
  "category": "输入交互",
  "order": 215,
  "summary": "FloatingActionButton 提供 regular、small、large 和 extended MD3 variants。",
  "example": { "id": "floating_action_button_basic" },
  "apis": [
    "FloatingActionButton(icon Widget, opts ...FloatingActionButtonOption) Widget",
    "SmallFloatingActionButton(icon Widget, opts ...FloatingActionButtonOption) Widget",
    "LargeFloatingActionButton(icon Widget, opts ...FloatingActionButtonOption) Widget",
    "ExtendedFloatingActionButton(icon, label Widget, opts ...FloatingActionButtonOption) Widget",
    "FloatingActionButtonElement(icon Element, opts ...FloatingActionButtonOption) Element",
    "SmallFloatingActionButtonElement(icon Element, opts ...FloatingActionButtonOption) Element",
    "LargeFloatingActionButtonElement(icon Element, opts ...FloatingActionButtonOption) Element",
    "ExtendedFloatingActionButtonElement(icon, label Element, opts ...FloatingActionButtonOption) Element",
    "FloatingActionButtonOnClick(fn func(ctx *Context)) FloatingActionButtonOption",
    "FloatingActionButtonDisabled(disabled bool) FloatingActionButtonOption",
    "FloatingActionButtonBackground(col color.NRGBA) FloatingActionButtonOption",
    "FloatingActionButtonForeground(col color.NRGBA) FloatingActionButtonOption",
    "FloatingActionButtonDecoration(d Decoration) FloatingActionButtonOption"
  ]
}
-->

# FloatingActionButton 浮动操作按钮

FAB 使用 `PrimaryContainer` / `OnPrimaryContainer`、level 3 elevation 和 MD3 尺寸：regular 56dp，small 40dp，large 96dp。

```go
ui.RowElement(
    ui.SmallFloatingActionButtonElement(ui.IconElement("add")),
    ui.FloatingActionButtonElement(ui.IconElement("add")),
    ui.LargeFloatingActionButtonElement(ui.IconElement("add")),
    ui.ExtendedFloatingActionButtonElement(ui.IconElement("add"), ui.TextElement("Create")),
)
```

```go
ui.LargeFloatingActionButtonElement(
    ui.IconElement("add"),
    ui.FloatingActionButtonBackground(ui.NRGBA(37, 99, 235, 255)),
    ui.FloatingActionButtonForeground(ui.NRGBA(255, 255, 255, 255)),
    ui.FloatingActionButtonDecoration(ui.Bg(ui.NRGBA(37, 99, 235, 255)).WithRad(28)),
)
```
