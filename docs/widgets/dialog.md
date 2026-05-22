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
- 如果 `DialogMaskClosable(false)`，关闭逻辑应由按钮回调、外部状态或 `DialogRef` 命令触发。

## Lifecycle / React-style 说明

- 当前仍以 legacy `Widget` + `DialogRef` 作为兼容实现。
- `open` 是受控输入；`DialogRef` 的 `Open` / `Close` / `Toggle` 命令会在布局时消费，并通过 `DialogOnOpenChange` 让外部状态同步。
- `Dialog` 内部用 overlay/mask + panel 组合渲染，关闭时子内容不布局；打开、关闭和 mask click 都属于 overlay host lifecycle，不应混入 component HookSlot。
- `DialogOnOpenChange` 会基于内部 `wasOpen` 状态做变化通知，后续 Element API 需要先明确通知去重、ref 命令释放和 host unmount 的处理规则。
- overlay stacking、焦点归属、输入事件拦截和子树 mount/unmount 语义尚未冻结，本批次不推荐 `DialogElement` 作为稳定公开 API。
- 在 Batch 5 期间，文档保持 legacy-first，只记录受控 open、ref 和 overlay 生命周期边界，不新增 React-style snippet，也不修改 docs_browser 示例映射。

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
