<!-- fluxui-doc-meta
{
  "id": "popup",
  "title": "Popup 弹窗",
  "category": "反馈组件",
  "order": 325,
  "summary": "Popup 提供纯净的弹窗容器，内部内容完全由用户自定义。",
  "example": { "id": "popup_basic" },
  "apis": [
    "Popup(open bool, child Widget, opts ...PopupOption) Widget",
    "PopupElement(open bool, child Element, opts ...PopupOption) Element",
    "PopupWidth(width float32) PopupOption",
    "PopupRadius(radius float32) PopupOption",
    "PopupMaskClosable(maskClosable bool) PopupOption",
    "PopupBackground(bg color.NRGBA) PopupOption",
    "PopupPadding(insets Insets) PopupOption",
    "PopupOnOpenChange(fn func(ctx *Context, open bool)) PopupOption",
    "PopupDecoration(d Decoration) PopupOption",
    "PopupMaskColor(col color.NRGBA) PopupOption",
    "PopupMaskAlpha(alpha uint8) PopupOption",
    "NewPopupRef() *PopupRef",
    "PopupAttachRef(ref *PopupRef) PopupOption"
  ]
}
-->

# Popup 弹窗

## 组件说明
Popup 是一个纯净的弹窗容器，只提供遮罩层和居中面板，不自带标题栏或操作按钮。内部内容完全由用户定义，适合需要高度自定义弹窗布局的场景。

与 Dialog 的区别：Dialog 自带标题、确认/取消按钮等结构化布局；Popup 则是一个空壳，类似 Web 中的 Modal 组件。

## MD3 默认样式

- container 使用 `SurfaceContainer`。
- shape 使用 `Small`。
- elevation 使用 level 2。
- mask 使用 `Scrim`。

## 使用方法
- `open` 控制显示与隐藏（受控模式）。
- 通过 `PopupOnOpenChange` 同步遮罩点击关闭和外部状态。
- 使用 `PopupPadding` 设置内边距，`PopupBackground` 设置背景色。
- 可通过 `PopupAttachRef` 绑定 `PopupRef` 实现命令式控制。
- Popup 不自带确认/取消按钮，内部关闭行为需要由用户内容、外部状态或 `PopupRef` 负责。

## Lifecycle / React-style Element

- `PopupElement` 已可在 `RunElement` root 下直接包裹 Element 子树。
- `open` 是受控输入；`PopupAttachRef` 绑定的 `PopupRef` 命令仍在布局时消费，并通过 `PopupOnOpenChange` 同步外部状态。
- mask click、overlay stacking、焦点归属和 ref command queue 仍属于底层 popup host lifecycle，不迁入 component HookSlot。

## 使用示例
```go
open := ui.State[bool](ctx)
ui.Popup(
    open.Value(),
    ui.Column(
        ui.Text("自定义标题", ui.TextSize(18)),
        ui.VSpacer(8),
        ui.Text("这里可以放任意组件。"),
        ui.VSpacer(12),
        ui.Button(
            ui.Text("关闭"),
            ui.OnClick(func(ctx *ui.Context) {
                open.Set(false)
            }),
        ),
    ),
    ui.PopupWidth(320),
    ui.PopupPadding(ui.All(16)),
    ui.PopupOnOpenChange(func(ctx *ui.Context, v bool) {
        open.Set(v)
    }),
)
```

### React-style Element

```go
func CustomPopup(ctx *ui.Context) ui.Element {
    open := ui.UseState(ctx, true)
    return ui.PopupElement(
        open.Value(),
        ui.ColumnElement(
            ui.TextElement("自定义标题", ui.TextSize(18)),
            ui.SpacerElement(0, 8),
            ui.TextElement("这里可以放任意 Element。"),
        ),
        ui.PopupWidth(320),
        ui.PopupPadding(ui.All(16)),
        ui.PopupOnOpenChange(func(ctx *ui.Context, v bool) { open.Set(v) }),
    )
}
```
