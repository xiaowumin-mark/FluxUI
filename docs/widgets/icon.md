<!-- fluxui-doc-meta
{
  "id": "icon",
  "title": "Icon 图标",
  "category": "基础显示",
  "order": 130,
  "summary": "Icon 提供图标语义展示能力。",
  "example": { "id": "icon_basic" },
  "apis": [
    "Icon(name string, opts ...IconOption) Widget",
    "FromWidget(w Widget) Element",
    "IconSize(size float32) IconOption",
    "IconColor(col color.NRGBA) IconOption",
    "IconOnClick(fn func(ctx *Context)) IconOption",
    "IconAttachRef(ref *ButtonRef) IconOption"
  ]
}
-->

# Icon 图标

## 组件说明
Icon 用于表达操作和状态语义。当前实现为轻量占位图标，后续可扩展到矢量图标体系。

## 使用方法
- 通过 `name` 传入图标语义标识。
- 调整大小和颜色时优先使用 Option，不要直接包额外文本样式。
- 有外部触发需求时，可通过 `IconAttachRef` 复用按钮 ref 行为。

## 使用示例
```go
ui.Icon(
    "H",
    ui.IconSize(20),
    ui.IconColor(ui.NRGBA(30, 136, 229, 255)),
)
```

## React-style 状态

- 当前 `Icon` 仍以 legacy `Widget` 作为稳定实现。
- React-style root 中可先使用 `FromWidget(ui.Icon(...))` 桥接到 Element 树。
- `IconOnClick` / `IconAttachRef` 仍复用 legacy button-style 事件与 ref 行为。
- 本阶段不冻结 `IconElement` API 名称；后续如引入 Element wrapper，需要先明确图标资源体系、点击命中和 ref 命令归属。
