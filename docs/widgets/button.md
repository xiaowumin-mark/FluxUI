<!-- fluxui-doc-meta
{
  "id": "button",
  "title": "Button 按钮",
  "category": "输入交互",
  "order": 200,
  "summary": "Button 用于触发点击行为和命令。",
  "example": { "id": "button_basic" },
  "apis": [
    "Button(child Widget, opts ...ButtonOption) Widget",
    "FilledButton(child Widget, opts ...ButtonOption) Widget",
    "FilledTonalButton(child Widget, opts ...ButtonOption) Widget",
    "OutlinedButton(child Widget, opts ...ButtonOption) Widget",
    "TextButton(child Widget, opts ...ButtonOption) Widget",
    "ElevatedButton(child Widget, opts ...ButtonOption) Widget",
    "ButtonElement(child Element, opts ...ButtonOption) Element",
    "FilledButtonElement(child Element, opts ...ButtonOption) Element",
    "FilledTonalButtonElement(child Element, opts ...ButtonOption) Element",
    "OutlinedButtonElement(child Element, opts ...ButtonOption) Element",
    "TextButtonElement(child Element, opts ...ButtonOption) Element",
    "ElevatedButtonElement(child Element, opts ...ButtonOption) Element",
    "OnClick(fn func(ctx *Context)) ButtonOption",
    "OnHover(fn func(ctx *Context, hovering bool)) ButtonOption",
    "Disabled(disabled bool) ButtonOption",
    "ButtonPadding(insets Insets) ButtonOption",
    "ButtonRadius(radius float32) ButtonOption",
    "ButtonBackground(value color.NRGBA) ButtonOption",
    "ButtonForeground(value color.NRGBA) ButtonOption",
    "ButtonDecoration(d Decoration) ButtonOption",
    "NewButtonRef() *ButtonRef",
    "ButtonAttachRef(ref *ButtonRef) ButtonOption",
    "(*ButtonRef).Click()"
  ]
}
-->

# Button 按钮

## 组件说明
Button 是最基础交互组件。默认样式已按 Material Design 3 对齐，`Button` 兼容旧 API 并映射到 Filled Button。

## MD3 变体

- `FilledButton`: `Primary` / `OnPrimary`。
- `FilledTonalButton`: `SecondaryContainer` / `OnSecondaryContainer`。
- `OutlinedButton`: 透明背景、`Outline` 边框、`Primary` 内容色。
- `TextButton`: 透明背景、`Primary` 内容色。
- `ElevatedButton`: tonal elevation level 1 surface、`Primary` 内容色和轻量阴影。

所有变体默认使用 `Theme.Shapes.Full`，hover/pressed 使用统一 state layer，disabled 使用 `OnSurface` 12% / 38%。

## 使用方法
- 点击逻辑放在 `OnClick` 中。
- 禁用状态统一用 `Disabled(true)`，避免仅靠颜色模拟禁用。
- 样式统一在 Option 层声明，不要在业务中散落魔法数。
- 需要外部主动触发时，使用 `ButtonAttachRef` 并调用 `ref.Click()`。

## 使用示例

### Legacy Widget
旧 `ui.Button` / `Widget` 写法继续可用：

```go
count := ui.State[int](ctx)
ui.Button(
    ui.Text("点击 +1"),
    ui.OnClick(func(ctx *ui.Context) {
        count.Set(count.Value() + 1)
    }),
)
```

```go
ui.Row(
    ui.FilledButton(ui.Text("Filled")),
    ui.FilledTonalButton(ui.Text("Tonal")),
    ui.OutlinedButton(ui.Text("Outlined")),
    ui.TextButton(ui.Text("Text")),
    ui.ElevatedButton(ui.Text("Elevated")),
)
```

### React-style Element
新代码可在 `RunElement` root 下返回 `ButtonElement`：

```go
func CounterButton(ctx *ui.Context) ui.Element {
    count := ui.UseState(ctx, 0)
    return ui.ColumnElement(
        ui.ButtonElement(
            ui.TextElement("点击 +1"),
            ui.OnClick(func(ctx *ui.Context) {
                count.Set(count.Value() + 1)
            }),
        ),
        ui.PaddingElement(
            ui.Insets{Top: 8},
            ui.TextElement(fmt.Sprintf("count = %d", count.Value())),
        ),
    )
}
```
