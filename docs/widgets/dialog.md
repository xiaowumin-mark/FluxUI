<!-- fluxui-doc-meta
{
  "id": "dialog",
  "title": "Dialog 对话框",
  "category": "反馈组件",
  "order": 320,
  "summary": "Dialog 用于中断式确认和补充信息输入。",
  "example": { "id": "dialog_basic" },
  "apis": [
    "Dialog(open bool, child Widget, opts ...DialogOption) Widget",
    "DialogElement(open bool, child Element, opts ...DialogOption) Element",
    "DialogTitle(title string) DialogOption",
    "DialogWidth(width float32) DialogOption",
    "DialogHeight(height float32) DialogOption",
    "DialogRadius(radius float32) DialogOption",
    "DialogQuick(quick bool) DialogOption",
    "DialogTypeOf(kind DialogType) DialogOption",
    "DialogMountMode(mount DialogMount) DialogOption",
    "DialogGlobalOverlay(global bool) DialogOption",
    "DialogIcon(icon Widget) DialogOption",
    "DialogIconElement(icon Element) DialogOption",
    "DialogHeadline(headline Widget) DialogOption",
    "DialogHeadlineElement(headline Element) DialogOption",
    "DialogActions(actions ...Widget) DialogOption",
    "DialogActionElements(actions ...Element) DialogOption",
    "DialogMaskClosable(maskClosable bool) DialogOption",
    "DialogOnOpenChange(fn func(ctx *Context, open bool)) DialogOption",
    "DialogOnOpen(fn func(ctx *Context)) DialogOption",
    "DialogOnOpened(fn func(ctx *Context)) DialogOption",
    "DialogOnClose(fn func(ctx *Context)) DialogOption",
    "DialogOnClosed(fn func(ctx *Context)) DialogOption",
    "DialogOnCancelOnly(fn func(ctx *Context)) DialogOption",
    "DialogOnConfirm(fn func(ctx *Context)) DialogOption",
    "DialogOnCancel(fn func(ctx *Context)) DialogOption",
    "DialogConfirmText(text string) DialogOption",
    "DialogCancelText(text string) DialogOption",
    "DialogDecoration(d Decoration) DialogOption",
    "DialogMaskColor(col color.NRGBA) DialogOption",
    "DialogMaskAlpha(alpha uint8) DialogOption",
    "NewDialogRef() *DialogRef",
    "DialogAttachRef(ref *DialogRef) DialogOption",
    "(*DialogRef).Open()",
    "(*DialogRef).Close()",
    "(*DialogRef).Toggle()"
  ]
}
-->

# Dialog 对话框

## 组件说明
Dialog 用于确认、警告、补充输入等高优先级交互，通常以遮罩方式覆盖当前内容。

## MD3 默认样式

- container 使用 `SurfaceContainerHigh`。
- shape 使用 `ExtraLarge`。
- elevation 使用 level 3。
- mask 使用 `Scrim`，默认 32% opacity。
- title 推荐使用 `HeadlineSmall`。
- 默认宽度遵循 Material Web：最小 280dp，最大 560dp，距离窗口边缘至少 24dp。
- 默认动画遵循 Material Web：打开 500ms，关闭 150ms；scrim 线性淡入淡出，dialog 从上方 50dp 进入，container 使用 emphasized 曲线。

## 使用方法
- `open` 控制显示与隐藏（受控模式）。
- 推荐统一处理 `DialogOnOpenChange`，保证遮罩点击关闭和外部状态同步。
- 确认/取消逻辑分别使用 `DialogOnConfirm` / `DialogOnCancel`。
- 外部程序可通过 `DialogAttachRef` 命令式控制开关。
- 如果 `DialogMaskClosable(false)`，关闭逻辑应由按钮回调、外部状态或 `DialogRef` 命令触发。
- 使用 `DialogIcon` / `DialogHeadline` / `DialogActions` 可按 Material Web 的 icon、headline、content、actions 分区组织内容。
- `DialogQuick(true)` 会跳过出入场动画，用于测试或低动效场景。
- `DialogTypeOf(DialogTypeAlert)` 会把对话框标记为 alert 语义；`DialogTypeFullscreen` 使用接近全屏的尺寸策略。
- 默认 `DialogGlobalOverlay(true)`。为了让遮罩真正覆盖全局，应把 `DialogElement` 放在根 `StackElement` 的顶层；需要局部遮罩时可设置 `DialogGlobalOverlay(false)` 或 `DialogMountMode(DialogMountLocal)`。

## Lifecycle / React-style Element

- `DialogElement` 已可在 `RunElement` root 下直接包裹 Element 子树。
- `open` 是受控输入；`DialogRef` 的 `Open` / `Close` / `Toggle` 命令仍在布局时消费，并通过 `DialogOnOpenChange` 同步外部状态。
- `DialogOnOpen` / `DialogOnOpened` / `DialogOnClose` / `DialogOnClosed` 对齐 Material Web 的 open/opened/close/closed 生命周期。
- `DialogOnCancelOnly` 专门处理 scrim/Escape 这类取消语义；按钮取消仍使用 `DialogOnCancel`。
- overlay/mask/panel、open-change 去重、焦点归属和 ref command queue 仍属于底层 dialog host lifecycle，不迁入 component HookSlot。

## 使用示例

### React-style Element

```go
func ConfirmDialog(ctx *ui.Context) ui.Element {
    open := ui.UseState(ctx, false)
    return ui.StackElement(
        ui.FilledButtonElement(ui.TextElement("Open dialog"), ui.OnClick(func(ctx *ui.Context) {
            open.Set(true)
        })),
        ui.DialogElement(
            open.Value(),
            ui.TextElement("确认执行该操作吗？"),
            ui.DialogIconElement(ui.IconElement("warning")),
            ui.DialogHeadlineElement(ui.TextElement("操作确认")),
            ui.DialogTypeOf(ui.DialogTypeAlert),
            ui.DialogGlobalOverlay(true),
            ui.DialogActionElements(
                ui.TextButtonElement(ui.TextElement("取消"), ui.OnClick(func(ctx *ui.Context) { open.Set(false) })),
                ui.TextButtonElement(ui.TextElement("确认"), ui.OnClick(func(ctx *ui.Context) { open.Set(false) })),
            ),
            ui.DialogOnOpenChange(func(ctx *ui.Context, v bool) { open.Set(v) }),
        ),
    )
}
```

### Legacy Widget
旧 `ui.Dialog` / `Widget` 写法继续可用：

```go
open := ui.State[bool](ctx)
ui.Dialog(
    open.Value(),
    ui.Text("确认执行该操作吗？"),
    ui.DialogTitle("操作确认"),
    ui.DialogMaskClosable(true),
    ui.DialogOnOpenChange(func(ctx *ui.Context, v bool) {
        open.Set(v)
    }),
    ui.DialogOnConfirm(func(ctx *ui.Context) {
        open.Set(false)
    }),
)
```
