<!-- fluxui-doc-meta
{
  "id": "icon_button",
  "title": "IconButton 图标按钮",
  "category": "输入交互",
  "order": 205,
  "summary": "IconButton 提供 MD3 standard、filled、filled tonal 和 outlined variants。",
  "example": { "id": "icon_button_basic" },
  "apis": [
    "IconButton(child Widget, opts ...IconButtonOption) Widget",
    "FilledIconButton(child Widget, opts ...IconButtonOption) Widget",
    "FilledTonalIconButton(child Widget, opts ...IconButtonOption) Widget",
    "OutlinedIconButton(child Widget, opts ...IconButtonOption) Widget",
    "IconButtonElement(child Element, opts ...IconButtonOption) Element",
    "FilledIconButtonElement(child Element, opts ...IconButtonOption) Element",
    "FilledTonalIconButtonElement(child Element, opts ...IconButtonOption) Element",
    "OutlinedIconButtonElement(child Element, opts ...IconButtonOption) Element",
    "IconButtonSelected(selected bool) IconButtonOption",
    "IconButtonDisabled(disabled bool) IconButtonOption",
    "IconButtonOnClick(fn func(ctx *Context)) IconButtonOption",
    "IconButtonSize(size float32) IconButtonOption",
    "IconButtonBackground(col color.NRGBA) IconButtonOption",
    "IconButtonForeground(col color.NRGBA) IconButtonOption",
    "IconButtonDecoration(d Decoration) IconButtonOption"
  ]
}
-->

# IconButton 图标按钮

IconButton 默认是 40dp 圆形 touch target，ripple 和 state layer 被裁剪在按钮内部，不改变布局尺寸。

```go
ui.RowElement(
    ui.IconButtonElement(ui.IconElement("S")),
    ui.FilledIconButtonElement(ui.IconElement("E"), ui.IconButtonSelected(true)),
    ui.FilledTonalIconButtonElement(ui.IconElement("T")),
    ui.OutlinedIconButtonElement(ui.IconElement("M")),
)
```

```go
ui.FilledTonalIconButtonElement(
    ui.IconElement("T"),
    ui.IconButtonSize(42),
    ui.IconButtonBackground(ui.NRGBA(219, 234, 254, 255)),
    ui.IconButtonForeground(ui.NRGBA(30, 64, 175, 255)),
)
```
