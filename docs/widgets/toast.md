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
    "ToastElement(message string, opts ...ToastOption) Element",
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

## Lifecycle / React-style Element

- `ToastElement` 已可在 `RunElement` root 下直接使用。
- `lastMessage`、`startAt`、`closed`、auto-close timer 和 redraw 请求仍属于底层 toast host state。
- `ToastOnClose` 通常仍需要清理外部消息状态，避免下一帧继续渲染同一 Toast。

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

### React-style Element

```go
func SaveToast(ctx *ui.Context) ui.Element {
    msg := ui.UseState(ctx, "已保存")
    if msg.Value() == "" {
        return nil
    }
    return ui.ToastElement(
        msg.Value(),
        ui.ToastTypeOf(ui.ToastSuccess),
        ui.ToastDuration(1500*time.Millisecond),
        ui.ToastOnClose(func(ctx *ui.Context) { msg.Set("") }),
    )
}
```
