<!-- fluxui-doc-meta
{
  "id": "badge",
  "title": "Badge 徽标",
  "category": "反馈组件",
  "order": 336,
  "summary": "Badge 用于在控件角落显示数量或状态点。",
  "example": { "id": "badge_basic" },
  "apis": [
    "Badge(child Widget, label string, opts ...BadgeOption) Widget",
    "BadgeElement(child Element, label string, opts ...BadgeOption) Element",
    "BadgeVisible(visible bool) BadgeOption",
    "BadgeBackground(col color.NRGBA) BadgeOption",
    "BadgeForeground(col color.NRGBA) BadgeOption",
    "BadgeDecoration(d Decoration) BadgeOption",
    "BadgeOffset(x, y int) BadgeOption"
  ]
}
-->

# Badge 徽标

Badge 是附着在控件右上角的 Material 3 状态标记。它可以显示短数字，也可以用空文本配合 `BadgeVisible(true)` 显示一个点状徽标。

## 使用方法

- `label` 非空时显示文本徽标。
- `label` 为空且 `BadgeVisible(true)` 时显示点状徽标。
- 可通过 `BadgeOffset` 微调徽标位置。

```go
ui.BadgeElement(
    ui.IconButtonElement(ui.IconElement("mail")),
    "3",
)
```
