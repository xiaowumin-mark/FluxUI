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
    "ButtonElement(child Element, opts ...ButtonOption) Element",
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
Button 是最基础交互组件。点击、悬停、禁用状态全部通过 Option 声明配置。

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
