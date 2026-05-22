<!-- fluxui-doc-meta
{
  "id": "toast",
  "title": "Toast 轻提示",
  "category": "反馈组件",
  "order": 330,
  "summary": "Toast 用于短时非阻塞提示。",
  "example": { "id": "toast_basic" },
  "apis": [
    "Toast(message string, opts ...ToastOption) Widget",
    "ToastTypeOf(kind ToastType) ToastOption",
    "ToastDuration(duration time.Duration) ToastOption",
    "ToastPositionOf(position ToastPosition) ToastOption",
    "ToastOnClose(fn func(ctx *Context)) ToastOption"
  ]
}
-->

# Toast 轻提示

## 组件说明
Toast 用于完成提示、错误提示、状态提醒等短时反馈，不打断主流程。

## 使用方法
- 通过状态控制是否渲染 Toast。
- `ToastOnClose` 中清理消息状态，避免重复展示。
- 类型和位置分别用 `ToastTypeOf`、`ToastPositionOf`。
- `ToastDuration` 控制自动关闭时间；传入非正值时不会走自动关闭计时。
- 同一消息重复展示时建议变更 message 或外部状态 key，避免复用旧的计时状态。

## Lifecycle / React-style 说明

- 当前仍以 legacy `Widget` 作为兼容实现。
- `Toast` 内部保存 `lastMessage`、`startAt` 和 `closed`，用于管理短时通知的 timer / auto-close 生命周期。
- `ToastOnClose` 会在 duration 到期后触发，通常需要在回调中清理外部消息状态，避免下一帧继续渲染同一 Toast。
- Toast 的 timer、close 状态和 redraw 请求都属于 overlay/notification host state，不应混入 component HookSlot。
- 后续若引入 Element wrapper，需要先明确重复 message、手动移除、unmount cleanup 和多个 toast stacking 的所有权规则，本批次不冻结 `ToastElement` API 名称。
- 在 Batch 5 期间，文档保持 legacy-first，只记录通知生命周期边界，不新增 React-style snippet，也不修改 docs_browser 示例映射。

## 使用示例
```go
msg := ui.State[string](ctx)
if msg.Value() != "" {
    ui.Toast(
        msg.Value(),
        ui.ToastTypeOf(ui.ToastSuccess),
        ui.ToastDuration(1500*time.Millisecond),
        ui.ToastOnClose(func(ctx *ui.Context) {
            msg.Set("")
        }),
    )
}
```
