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
    "IconUseFont(id string) IconOption",
    "IconFontFamily(family string) IconOption",
    "IconOnClick(fn func(ctx *Context)) IconOption",
    "IconAttachRef(ref *ButtonRef) IconOption"
  ]
}
-->

# Icon 图标

## 组件说明
Icon 用于表达操作和状态语义。导入 `github.com/xiaowumin-mark/FluxUI/icons/md3` 后，Icon 默认使用内置 Material Symbols Outlined 图标字体，并把 `name` 当作 ligature 名称渲染。

## 使用方法
- 通过 `name` 传入图标语义标识，例如 `home`、`search`、`settings`、`add`。
- 调整大小和颜色时优先使用 Option，不要直接包额外文本样式。
- 多图标字体并存时，通过 `IconUseFont(id)` 指定注册字体；需要直接指定字体族名时使用 `IconFontFamily(family)`。
- 有外部触发需求时，可通过 `IconAttachRef` 复用按钮 ref 行为。

## 使用示例

### React-style Element

`IconElement` 已可在 `RunElement` root 下直接使用。`IconOnClick` / `IconAttachRef` 仍复用底层 button-style 事件与 ref 行为。

```go
import _ "github.com/xiaowumin-mark/FluxUI/icons/md3"

func StatusIcon(ctx *ui.Context) ui.Element {
    return ui.IconElement("home", ui.IconSize(20), ui.IconColor(ui.NRGBA(30, 136, 229, 255)))
}
```

### Legacy Widget
旧 `ui.Icon` / `Widget` 写法继续可用：

```go
ui.Icon(
    "home",
    ui.IconSize(20),
    ui.IconColor(ui.NRGBA(30, 136, 229, 255)),
)
```
