<!-- fluxui-doc-meta
{
  "id": "card",
  "title": "Card 卡片",
  "category": "基础显示",
  "order": 140,
  "summary": "Card 是信息块容器，支持背景、圆角、边框与点击。",
  "example": { "id": "card_basic" },
  "apis": [
    "Card(child Widget, opts ...CardOption) Widget",
    "FilledCard(child Widget, opts ...CardOption) Widget",
    "ElevatedCard(child Widget, opts ...CardOption) Widget",
    "OutlinedCard(child Widget, opts ...CardOption) Widget",
    "CardElement(child Element, opts ...CardOption) Element",
    "FilledCardElement(child Element, opts ...CardOption) Element",
    "ElevatedCardElement(child Element, opts ...CardOption) Element",
    "OutlinedCardElement(child Element, opts ...CardOption) Element",
    "CardPadding(insets Insets) CardOption",
    "CardRadius(radius float32) CardOption",
    "CardBackground(col color.NRGBA) CardOption",
    "CardBorder(col color.NRGBA, width float32) CardOption",
    "CardShadow(level int) CardOption",
    "CardOnClick(fn func(ctx *Context)) CardOption",
    "CardAttachRef(ref *ButtonRef) CardOption",
    "CardDecoration(d Decoration) CardOption"
  ]
}
-->

# Card 卡片

## 组件说明
Card 适合信息摘要、列表条目、统计块等场景。默认 `Card` 映射到 MD3 Filled Card。

## MD3 变体

- `FilledCard`: `SurfaceContainerHighest`，无阴影。
- `ElevatedCard`: tonal elevation level 1 surface，并带轻量阴影。
- `OutlinedCard`: `Surface` 背景，`OutlineVariant` 边框。
- 所有默认变体使用 `Theme.Shapes.Medium`。

## 使用方法
- 内容区域直接作为 child 传入。
- 可按场景配置圆角、边框和背景。
- 有交互时使用 `CardOnClick`。
- 需要业务层外部触发卡片点击时，使用 `CardAttachRef`。

## 使用示例

### React-style Element

`CardElement` 已可在 `RunElement` root 下直接包裹 Element 子树。`CardOnClick` / `CardAttachRef` 仍走底层 widget 事件与 ref 路径，不归入 component HookSlot。

```go
func SummaryCard(ctx *ui.Context) ui.Element {
    return ui.CardElement(
        ui.ColumnElement(
            ui.TextElement("卡片标题"),
            ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("卡片内容")),
        ),
        ui.CardRadius(12),
    )
}
```

```go
ui.RowElement(
    ui.FilledCardElement(ui.TextElement("Filled")),
    ui.ElevatedCardElement(ui.TextElement("Elevated")),
    ui.OutlinedCardElement(ui.TextElement("Outlined")),
)
```

### Legacy Widget
旧 `ui.Card` / `Widget` 写法继续可用：

```go
ui.Card(
    ui.Column(
        ui.Text("卡片标题"),
        ui.Padding(ui.Insets{Top: 6}, ui.Text("卡片内容")),
    ),
)
```
