<!-- fluxui-doc-meta
{
  "id": "snackbar",
  "title": "Snackbar 消息条",
  "category": "反馈组件",
  "order": 332,
  "summary": "Snackbar 用于显示短消息，并可附带一个操作按钮。",
  "example": { "id": "snackbar_basic" },
  "apis": [
    "Snackbar(message string, opts ...SnackbarOption) Widget",
    "SnackbarElement(message string, opts ...SnackbarOption) Element",
    "SnackbarAction(label string, fn func(ctx *Context)) SnackbarOption",
    "ToastDuration(duration time.Duration) ToastOption",
    "ToastOnClose(fn func(ctx *Context)) ToastOption"
  ]
}
-->

# Snackbar 消息条

Snackbar 是 Material 3 中用于短消息反馈的底部浮层。它复用 Toast 的宿主状态，同时提供 Snackbar 命名和 action API，适合“已保存”“已归档”“撤销”等轻量反馈。

## 使用方法

- `SnackbarElement` 可直接在 Element 树中使用。
- `SnackbarAction` 用于配置右侧操作按钮。
- `ToastDuration(0)` 可以保持显示，直到用户点击 action 或外部状态移除组件。

```go
ui.SnackbarElement(
    "草稿已归档",
    ui.SnackbarAction("撤销", func(ctx *ui.Context) {
        // 执行撤销逻辑
    }),
    ui.ToastDuration(0),
)
```
