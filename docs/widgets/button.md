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
    "OnClick(fn func(ctx *Context)) ButtonOption",
    "OnHover(fn func(ctx *Context, hovering bool)) ButtonOption",
    "Disabled(disabled bool) ButtonOption",
    "ButtonPadding(insets Insets) ButtonOption",
    "ButtonRadius(radius float32) ButtonOption",
    "ButtonBackground(value color.NRGBA) ButtonOption",
    "ButtonForeground(value color.NRGBA) ButtonOption"
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

## 使用示例
```go
count := ui.State[int](ctx)
ui.Button(
    ui.Text("点击 +1"),
    ui.OnClick(func(ctx *ui.Context) {
        count.Set(count.Value() + 1)
    }),
)
```
