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
    "DialogRadius(radius float32) DialogOption",
    "DialogMaskClosable(maskClosable bool) DialogOption",
    "DialogOnOpenChange(fn func(ctx *Context, open bool)) DialogOption",
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
- mask 使用 `Scrim`。
- title 推荐使用 `HeadlineSmall`。

## 使用方法
- `open` 控制显示与隐藏（受控模式）。
- 推荐统一处理 `DialogOnOpenChange`，保证遮罩点击关闭和外部状态同步。
- 确认/取消逻辑分别使用 `DialogOnConfirm` / `DialogOnCancel`。
- 外部程序可通过 `DialogAttachRef` 命令式控制开关。
- 如果 `DialogMaskClosable(false)`，关闭逻辑应由按钮回调、外部状态或 `DialogRef` 命令触发。

## Lifecycle / React-style Element

- `DialogElement` 已可在 `RunElement` root 下直接包裹 Element 子树。
- `open` 是受控输入；`DialogRef` 的 `Open` / `Close` / `Toggle` 命令仍在布局时消费，并通过 `DialogOnOpenChange` 同步外部状态。
- overlay/mask/panel、open-change 去重、焦点归属和 ref command queue 仍属于底层 dialog host lifecycle，不迁入 component HookSlot。

## 使用示例

### React-style Element

```go
func ConfirmDialog(ctx *ui.Context) ui.Element {
    open := ui.UseState(ctx, false)
    return ui.DialogElement(
        open.Value(),
        ui.TextElement("确认执行该操作吗？"),
        ui.DialogTitle("操作确认"),
        ui.DialogOnOpenChange(func(ctx *ui.Context, v bool) { open.Set(v) }),
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
