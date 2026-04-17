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
    "IconSize(size float32) IconOption",
    "IconColor(col color.NRGBA) IconOption",
    "IconOnClick(fn func(ctx *Context)) IconOption"
  ]
}
-->

# Icon 图标

## 组件说明
Icon 用于表达操作和状态语义。当前实现为轻量占位图标，后续可扩展到矢量图标体系。

## 使用方法
- 通过 `name` 传入图标语义标识。
- 调整大小和颜色时优先使用 Option，不要直接包额外文本样式。

## 使用示例
```go
ui.Icon(
    "H",
    ui.IconSize(20),
    ui.IconColor(ui.NRGBA(30, 136, 229, 255)),
)
```
