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
    "InputLabel(text string) InputOption",
    "InputSupportingText(text string) InputOption",
    "InputErrorText(text string) InputOption",
    "InputError(error bool) InputOption",
    "InputRequired(required bool) InputOption",
    "InputNoAsterisk(noAsterisk bool) InputOption",
    "InputPrefixText(text string) InputOption",
    "InputSuffixText(text string) InputOption",
    "InputLeading(leading Widget) InputOption",
    "InputTrailing(trailing Widget) InputOption",
    "InputLeadingElement(leading Element) InputOption",
    "InputTrailingElement(trailing Element) InputOption",
    "InputCounter(visible bool) InputOption",
    "InputRows(rows int) InputOption",
    "InputMinRows(rows int) InputOption",
    "InputMaxRows(rows int) InputOption",
    "InputPadding(insets Insets) InputOption",
    "InputRadius(radius float32) InputOption",
    "InputBorder(color color.NRGBA) InputOption",
    "InputBorderFocus(color color.NRGBA) InputOption",
    "InputBackground(color color.NRGBA) InputOption",
    "InputForeground(color color.NRGBA) InputOption",
    "InputDecoration(d Decoration) InputOption",
    "InputTextSize(size float32) InputOption",
    "InputFontFamily(family string) InputOption",
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

- `OutlinedTextField`: `Surface` 背景，`Outline` 边框，focus/error 边框加粗，label 浮到 outline notch。
- `FilledTextField`: `SurfaceContainerHighest` 背景，底部 active indicator 在 focus/error 时加粗。
- 输入文字默认使用 `Theme.Types.BodyLarge`。
- placeholder 使用 `OnSurfaceVariant`。
- disabled 使用统一 disabled opacity。
- `InputLabel` 支持 Material floating label 动画；空值未聚焦时 label 位于容器中，聚焦或有值时上浮。
- `InputSupportingText`、`InputErrorText` 和 `InputCounter` 会在 field 下方形成 supporting row。
- `InputLeading` / `InputTrailing`、`InputPrefixText` / `InputSuffixText` 支持常见 Material Web text-field 插槽。

## 使用方法
- 受控绑定：`value -> TextField(value)`，`InputOnChange -> state.Set(value)`。
- 密码场景使用 `InputPassword(true)`。
- 长文本场景建议关闭单行模式。
- 多行输入使用 `InputSingleLine(false)` 配合 `InputRows` / `InputMinRows` / `InputMaxRows`。
- 必填字段使用 `InputRequired(true)`；如不显示星号可加 `InputNoAsterisk(true)`。
- 命令式控制（聚焦/清空/追加）建议通过 `InputAttachRef` + `InputRef` 方法完成。

## React-style Element

- `TextFieldElement` 已可在 `RunElement` root 下直接使用。
- 文本值仍是受控输入：`value` 来自调用方状态，`InputOnChange` 回写状态。
- 编辑器焦点、光标、选择区和 `InputRef` 命令队列仍由底层 input host state 管理，不迁入 component HookSlot。

## 使用示例

### React-style Element

```go
func NameField(ctx *ui.Context) ui.Element {
    name := ui.UseState(ctx, "")
    return ui.TextFieldElement(
        name.Value(),
        ui.InputLabel("Username"),
        ui.InputPlaceholder("name@example.com"),
        ui.InputLeading(ui.Icon("person")),
        ui.InputSupportingText("This field is controlled by host state."),
        ui.InputFontFamily("Segoe UI"),
        ui.InputOnChange(func(ctx *ui.Context, value string) {
            name.Set(value)
        }),
    )
}
```

```go
ui.RowElement(
    ui.OutlinedTextFieldElement("Outlined", ui.InputLabel("Outlined"), ui.InputCounter(true), ui.InputMaxLen(20)),
    ui.FilledTextFieldElement("Filled", ui.InputLabel("Filled"), ui.InputPrefixText("$"), ui.InputSuffixText("USD")),
)
```

### Error 和多行

```go
ui.ColumnElement(
    ui.OutlinedTextFieldElement(
        "Draft",
        ui.InputLabel("Title"),
        ui.InputRequired(true),
        ui.InputMaxLen(10),
        ui.InputCounter(true),
        ui.InputError(true),
        ui.InputErrorText("Use 10 characters or less"),
    ),
    ui.OutlinedTextFieldElement(
        notes.Value(),
        ui.InputLabel("Notes"),
        ui.InputSingleLine(false),
        ui.InputRows(4),
        ui.InputSupportingText("Textarea mode uses the same MD3 chrome."),
    ),
)
```

### Legacy Widget
旧 `ui.TextField` / `Widget` 写法继续可用：

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
