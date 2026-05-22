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
    "FromWidget(w Widget) Element",
    "CardPadding(insets Insets) CardOption",
    "CardRadius(radius float32) CardOption",
    "CardBackground(col color.NRGBA) CardOption",
    "CardBorder(col color.NRGBA, width float32) CardOption",
    "CardShadow(level int) CardOption",
    "CardOnClick(fn func(ctx *Context)) CardOption",
    "CardAttachRef(ref *ButtonRef) CardOption"
  ]
}
-->

# Card 卡片

## 组件说明
Card 适合信息摘要、列表条目、统计块等场景。它是“有视觉结构”的内容容器。

## 使用方法
- 内容区域直接作为 child 传入。
- 可按场景配置圆角、边框和背景。
- 有交互时使用 `CardOnClick`。
- 需要业务层外部触发卡片点击时，使用 `CardAttachRef`。

## 使用示例
```go
ui.Card(
    ui.Column(
        ui.Text("卡片标题"),
        ui.Padding(ui.Insets{Top: 6}, ui.Text("卡片内容")),
    ),
)
```

## React-style 状态

- 当前 `Card` 仍以 legacy `Widget` 作为稳定实现。
- React-style root 中可先使用 `FromWidget(ui.Card(...))` 桥接到 Element 树。
- `CardOnClick` / `CardAttachRef` 仍走 legacy 事件与 ref 路径，不归入 component HookSlot。
- 本阶段不冻结 `CardElement` API 名称；后续如引入 Element wrapper，需要先明确点击事件、ref 命令和 host-state 归属。
