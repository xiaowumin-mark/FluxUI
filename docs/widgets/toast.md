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
