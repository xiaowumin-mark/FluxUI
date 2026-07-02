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
    "PopupHeight(height float32) PopupOption",
    "PopupRadius(radius float32) PopupOption",
    "PopupMaskClosable(maskClosable bool) PopupOption",
    "PopupBackground(bg color.NRGBA) PopupOption",
    "PopupPadding(insets Insets) PopupOption",
    "PopupQuick(quick bool) PopupOption",
    "PopupTypeOf(kind DialogType) PopupOption",
    "PopupMountMode(mount DialogMount) PopupOption",
    "PopupGlobalOverlay(global bool) PopupOption",
    "PopupOnOpenChange(fn func(ctx *Context, open bool)) PopupOption",
    "PopupOnOpen(fn func(ctx *Context)) PopupOption",
    "PopupOnOpened(fn func(ctx *Context)) PopupOption",
    "PopupOnClose(fn func(ctx *Context)) PopupOption",
    "PopupOnClosed(fn func(ctx *Context)) PopupOption",
    "PopupOnCancelOnly(fn func(ctx *Context)) PopupOption",
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

- container 继承 Dialog 风格，使用 `SurfaceContainerHigh`。
- shape 继承 Dialog 风格，使用 `ExtraLarge`。
- elevation 继承 Dialog 风格，使用 level 3。
- Popup 与 Dialog 共用同一个 modal surface 默认装饰：默认 28dp 圆角和 level 3 柔和阴影；只有显式传入 `PopupRadius` / `PopupDecoration` 时才覆盖。
- mask 使用 `Scrim`，默认 32% opacity。
- 默认尺寸策略继承 Dialog：最小 280dp，最大 560dp，最大高度 560dp，距离窗口边缘至少 24dp。
- `PopupHeight`、`PopupTypeOf(DialogTypeFullscreen)`、`PopupGlobalOverlay` 与 Dialog 使用同一套 modal surface 规则。
- 默认动画继承 Dialog：打开 500ms，关闭 150ms，scrim 线性淡入淡出，面板从上方 50dp 进入，内容按 Dialog content timing 淡入。

## 使用方法
- `open` 控制显示与隐藏（受控模式）。
- 通过 `PopupOnOpenChange` 同步遮罩点击关闭和外部状态。
- 使用 `PopupPadding` 设置内边距，`PopupBackground` 设置背景色。
- 可通过 `PopupAttachRef` 绑定 `PopupRef` 实现命令式控制。
- Popup 不自带确认/取消按钮，内部关闭行为需要由用户内容、外部状态或 `PopupRef` 负责。
- 默认 Popup 是 Dialog 风格 modal surface；如果需要紧凑空壳，使用 `PopupPadding(ui.All(0))`、`PopupRadius(...)` 或 `PopupDecoration(...)` 显式覆盖。
- `PopupQuick(true)` 会跳过出入场动画。
- `PopupTypeOf(ui.DialogTypeFullscreen)` 使用 Dialog fullscreen 的尺寸策略，但仍保留 Popup 的自定义内容结构。
- 默认 `PopupGlobalOverlay(true)`。为了让遮罩真正覆盖全局，应把 `PopupElement` 放在根 `StackElement` 的顶层；需要局部遮罩时可设置 `PopupGlobalOverlay(false)` 或 `PopupMountMode(DialogMountLocal)`。

## Lifecycle / React-style Element

- `PopupElement` 已可在 `RunElement` root 下直接包裹 Element 子树。
- `open` 是受控输入；`PopupAttachRef` 绑定的 `PopupRef` 命令仍在布局时消费，并通过 `PopupOnOpenChange` 同步外部状态。
- `PopupOnOpen` / `PopupOnOpened` / `PopupOnClose` / `PopupOnClosed` 对齐 Dialog 生命周期。
- `PopupOnCancelOnly` 专门处理 scrim/Escape 这类取消语义。
- mask click、overlay stacking、焦点归属和 ref command queue 仍属于底层 popup host lifecycle，不迁入 component HookSlot。

## 使用示例

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
        ui.PopupOnOpenChange(func(ctx *ui.Context, v bool) { open.Set(v) }),
    )
}
```

### Legacy Widget
旧 `ui.Popup` / `Widget` 写法继续可用：

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
