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
    "OutlinedTextField(value string, opts ...InputOption) Widget",
    "FilledTextField(value string, opts ...InputOption) Widget",
    "TextFieldElement(value string, opts ...InputOption) Element",
    "OutlinedTextFieldElement(value string, opts ...InputOption) Element",
    "FilledTextFieldElement(value string, opts ...InputOption) Element",
    "InputPlaceholder(text string) InputOption",
    "InputPadding(insets Insets) InputOption",
    "InputRadius(radius float32) InputOption",
    "InputBorder(color color.NRGBA) InputOption",
    "InputBorderFocus(color color.NRGBA) InputOption",
    "InputBackground(color color.NRGBA) InputOption",
    "InputForeground(color color.NRGBA) InputOption",
    "InputDecoration(d Decoration) InputOption",
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
TextField 是受控输入组件，值由外部状态提供，输入变化通过 `InputOnChange` 回传。默认 `TextField` 映射到 MD3 Outlined Text Field。

## MD3 变体

- `OutlinedTextField`: `Surface` 背景，`Outline` 边框，focus 使用 `Primary`。
- `FilledTextField`: `SurfaceContainerHighest` 背景，无外边框。
- 输入文字默认使用 `Theme.Types.BodyLarge`。
- placeholder 使用 `OnSurfaceVariant`。
- disabled 使用统一 disabled opacity。

## 使用方法
- 受控绑定：`value -> TextField(value)`，`InputOnChange -> state.Set(value)`。
- 密码场景使用 `InputPassword(true)`。
- 长文本场景建议关闭单行模式。
- 命令式控制（聚焦/清空/追加）建议通过 `InputAttachRef` + `InputRef` 方法完成。

## React-style Element

- `TextFieldElement` 已可在 `RunElement` root 下直接使用。
- 文本值仍是受控输入：`value` 来自调用方状态，`InputOnChange` 回写状态。
- 编辑器焦点、光标、选择区和 `InputRef` 命令队列仍由底层 input host state 管理，不迁入 component HookSlot。

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

```go
ui.Row(
    ui.OutlinedTextField("Outlined", ui.InputPlaceholder("Outlined")),
    ui.FilledTextField("Filled", ui.InputPlaceholder("Filled")),
)
```

### React-style Element

```go
func NameField(ctx *ui.Context) ui.Element {
    name := ui.UseState(ctx, "")
    return ui.TextFieldElement(
        name.Value(),
        ui.InputPlaceholder("请输入名称"),
        ui.InputOnChange(func(ctx *ui.Context, value string) {
            name.Set(value)
        }),
    )
}
```
