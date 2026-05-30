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
    "IconElement(name string, opts ...IconOption) Element",
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

### React-style Element

`IconElement` 已可在 `RunElement` root 下直接使用。`IconOnClick` / `IconAttachRef` 仍复用底层 button-style 事件与 ref 行为。

```go
func StatusIcon(ctx *ui.Context) ui.Element {
    return ui.IconElement("H", ui.IconSize(20), ui.IconColor(ui.NRGBA(30, 136, 229, 255)))
}
```

### Legacy Widget
旧 `ui.Icon` / `Widget` 写法继续可用：

```go
ui.Icon(
    "H",
    ui.IconSize(20),
    ui.IconColor(ui.NRGBA(30, 136, 229, 255)),
)
```
