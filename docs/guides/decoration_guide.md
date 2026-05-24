<!-- fluxui-doc-meta
{
  "id": "decoration_guide",
  "title": "Decoration 使用指南",
  "category": "使用指南",
  "order": 710,
  "summary": "通过实战模式理解 Decoration 的组合、状态、主题回退和组件接入方式。",
  "example": { "id": "decoration_guide_basic" },
  "apis": [
    "ContainerDecoration(d Decoration, child Widget, opts ...ContainerDecorationOption) Widget",
    "ContainerDecorationElement(d Decoration, child Element, opts ...ContainerDecorationOption) Element",
    "ButtonDecoration(d Decoration) ButtonOption",
    "InputDecoration(d Decoration) InputOption",
    "CardDecoration(d Decoration) CardOption",
    "DialogDecoration(d Decoration) DialogOption",
    "PopupDecoration(d Decoration) PopupOption",
    "ToastDecoration(d Decoration) ToastOption",
    "SwitchDecoration(d Decoration) SwitchOption",
    "SliderDecoration(d Decoration) SliderOption",
    "TabsDecoration(d Decoration) TabsOption",
    "TabsTabDecoration(d Decoration) TabsOption"
  ]
}
-->

# Decoration 使用指南

## 推荐模式
Decoration 最适合作为“视觉语义”的封装，而不是把颜色、边距、圆角写散在业务组件中。

```go
func primaryButton(ctx *ui.Context) ui.ButtonOption {
    th := ui.UseTheme(ctx)
    d := ui.Bg(th.Primary).
        WithPad(ui.Symmetric(10, 14)).
        WithRad(10).
        WithHover(ui.Bg(th.Colors.Secondary)).
        WithPressed(ui.Bg(ui.NRGBA(30, 64, 175, 255))).
        WithDisabled(ui.Bg(th.Disabled))
    return ui.ButtonDecoration(d)
}
```

## 组件接入
支持 Decoration 的组件一般使用 `XxxDecoration(d)` 选项：

```go
ui.ButtonElement(
    ui.TextElement("保存", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
    ui.ButtonDecoration(ui.Bg(th.Primary).WithPad(ui.Symmetric(10, 16)).WithRad(10)),
)

ui.TextFieldElement(
    value.Value(),
    ui.InputDecoration(ui.Bg(th.Surface).WithRad(10).WithPad(ui.Symmetric(8, 12))),
)
```

## 主题回退
Decoration 不必设置所有字段：

```go
// 只覆盖圆角和边距，背景仍由组件默认值或主题决定。
ui.ButtonDecoration(ui.Rad(12).WithPad(ui.Symmetric(8, 14)))
```

## Merge 组合
```go
baseCard := ui.Bg(th.Surface).WithPad(ui.All(16)).WithRad(16)
outlined := baseCard.Merge(ui.BorderDeco(1, th.Colors.Outline))
raised := baseCard.Merge(ui.Elevation(2))
```

## 动效组合
Decoration 可以和 `UseAnimatedDecoration` 配合：

```go
target := ui.Bg(th.Surface).WithRad(12).WithPad(ui.All(16))
if selected {
    target = target.Merge(ui.Elevation(2)).WithTransform(ui.Transform2D{ScaleX: 1.03, ScaleY: 1.03})
}
animated := ui.UseAnimatedDecoration(ctx, target, 180*time.Millisecond, ui.EaseOut)

return ui.ContainerDecorationElement(animated, child)
```

## 使用建议
- 定义一组可复用的 style helper，例如 `surfaceCard(ctx)`、`primaryAction(ctx)`。
- 状态样式放进 Decoration，交互逻辑放进组件 Option。
- 对外暴露组件时，优先提供 Decoration 参数，而不是多个颜色参数。
