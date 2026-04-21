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
    "DialogTitle(title string) DialogOption",
    "DialogWidth(width float32) DialogOption",
    "DialogRadius(radius float32) DialogOption",
    "DialogMaskClosable(maskClosable bool) DialogOption",
    "DialogOnOpenChange(fn func(ctx *Context, open bool)) DialogOption",
    "DialogOnConfirm(fn func(ctx *Context)) DialogOption",
    "DialogOnCancel(fn func(ctx *Context)) DialogOption",
    "DialogConfirmText(text string) DialogOption",
    "DialogCancelText(text string) DialogOption",
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

## 使用方法
- `open` 控制显示与隐藏（受控模式）。
- 推荐统一处理 `DialogOnOpenChange`，保证遮罩点击关闭和外部状态同步。
- 确认/取消逻辑分别使用 `DialogOnConfirm` / `DialogOnCancel`。
- 外部程序可通过 `DialogAttachRef` 命令式控制开关。

## 使用示例
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
