<!-- fluxui-doc-meta
{
  "id": "textfield",
  "title": "TextField 输入框",
  "category": "输入交互",
  "order": 210,
  "summary": "TextField 支持受控输入、样式定制与焦点回调。",
  "example": { "id": "textfield_basic" },
  "apis": [
    "TextField(value string, opts ...InputOption) Widget",
    "InputPlaceholder(text string) InputOption",
    "InputPadding(insets Insets) InputOption",
    "InputRadius(radius float32) InputOption",
    "InputBorder(color color.NRGBA) InputOption",
    "InputBorderFocus(color color.NRGBA) InputOption",
    "InputBackground(color color.NRGBA) InputOption",
    "InputForeground(color color.NRGBA) InputOption",
    "InputTextSize(size float32) InputOption",
    "InputMaxLen(maxLen int) InputOption",
    "InputPassword(password bool) InputOption",
    "InputSingleLine(singleLine bool) InputOption",
    "InputDisabled(disabled bool) InputOption",
    "InputOnChange(fn func(ctx *Context, value string)) InputOption",
    "InputOnFocus(fn func(ctx *Context, focused bool)) InputOption",
    "NewInputRef() *InputRef",
    "InputAttachRef(ref *InputRef) InputOption",
    "(*InputRef).SetText(value string)",
    "(*InputRef).Append(value string)",
    "(*InputRef).Clear()",
    "(*InputRef).Focus()",
    "(*InputRef).Blur()"
  ]
}
-->

# TextField 输入框

## 组件说明
TextField 是受控输入组件，值由外部状态提供，输入变化通过 `InputOnChange` 回传。

## 使用方法
- 受控绑定：`value -> TextField(value)`，`InputOnChange -> state.Set(value)`。
- 密码场景使用 `InputPassword(true)`。
- 长文本场景建议关闭单行模式。
- 命令式控制（聚焦/清空/追加）建议通过 `InputAttachRef` + `InputRef` 方法完成。

## Host-state / React-style 说明

- 当前仍以 legacy `Widget` + `InputRef` 作为兼容实现。
- `TextField` 的 Element wrapper 名称与 host-state 边界尚未冻结，不在本批次文档中作为稳定公开 API 推荐。
- 在 Batch 4 期间，文档保持 legacy-first，只保留受控输入和命令式 ref 用法，不引入新的 React-style snippet。
- 更完整的表单组合可参考 `examples/form_validation`，它仍然是 legacy 组合示例。

## 使用示例
```go
name := ui.State[string](ctx)
ui.TextField(
    name.Value(),
    ui.InputPlaceholder("请输入名称"),
    ui.InputOnChange(func(ctx *ui.Context, value string) {
        name.Set(value)
    }),
)
```
