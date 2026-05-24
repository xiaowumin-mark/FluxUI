<!-- fluxui-doc-meta
{
  "id": "decoration",
  "title": "Decoration 装饰器",
  "category": "样式系统",
  "order": 500,
  "summary": "Decoration 是 FluxUI 的统一样式对象，用于背景、边距、圆角、边框、阴影、状态样式等。",
  "example": { "id": "decoration_basic" },
  "apis": [
    "type Decoration = style.Decoration",
    "Bg(c color.NRGBA) Decoration",
    "Pad(p Insets) Decoration",
    "Margin(m Insets) Decoration",
    "Rad(r float32) Decoration",
    "BorderDeco(width float32, col color.NRGBA) Decoration",
    "Opacity(v float32) Decoration",
    "Circle() Decoration",
    "LinearGrad(start, end image.Point, from, to color.NRGBA) Decoration",
    "Shadow(offX, offY, blur float32, col color.NRGBA) Decoration",
    "Elevation(level int) Decoration",
    "Hover(d Decoration) Decoration",
    "Pressed(d Decoration) Decoration",
    "Focused(d Decoration) Decoration",
    "DisabledDeco(d Decoration) Decoration",
    "HoverBg(c color.NRGBA) Decoration",
    "PressedBg(c color.NRGBA) Decoration",
    "ImageBg(src image.Image, fit ImageFillFit) Decoration",
    "TransformDeco(rotateDeg, scaleX, scaleY, transX, transY float32, origin TransformOrigin) Decoration",
    "(Decoration).WithBg(c color.NRGBA) Decoration",
    "(Decoration).WithGradient(g LinearGradient) Decoration",
    "(Decoration).WithPad(p Insets) Decoration",
    "(Decoration).WithMargin(m Insets) Decoration",
    "(Decoration).WithRad(r float32) Decoration",
    "(Decoration).WithBorder(b style.Border) Decoration",
    "(Decoration).WithOpacity(v float32) Decoration",
    "(Decoration).WithCircleClip() Decoration",
    "(Decoration).WithShadow(s style.BoxShadow) Decoration",
    "(Decoration).WithHover(h Decoration) Decoration",
    "(Decoration).WithPressed(p Decoration) Decoration",
    "(Decoration).WithFocused(f Decoration) Decoration",
    "(Decoration).WithDisabled(d Decoration) Decoration",
    "(Decoration).WithImage(f ImageFill) Decoration",
    "(Decoration).WithTransform(t Transform2D) Decoration",
    "(Decoration).Merge(other Decoration) Decoration"
  ]
}
-->

# Decoration 装饰器

## API 说明
Decoration 是 FluxUI 推荐的统一样式入口。它用可选字段描述视觉外观：背景、渐变、内边距、外边距、圆角、边框、透明度、圆形裁切、阴影、背景图片、2D 变换，以及 hover / pressed / focused / disabled 状态装饰。

所有字段都是“可选覆盖”。字段未设置时，组件会继续使用自身默认样式或主题回退值。

## 使用方法
- 使用 `ui.Bg(...)`、`ui.Pad(...)`、`ui.Rad(...)` 等快捷函数创建单项装饰。
- 使用链式 `WithXxx` 组合多个样式。
- 使用 `Merge` 合并装饰，后者的非空字段覆盖前者。
- 交互组件可通过 `WithHover` / `WithPressed` / `WithDisabled` 定义状态样式。
- 优先使用 `*Decoration(...)` 组件选项，而不是同时散落多个颜色、边距、圆角选项。

## 基础示例
```go
card := ui.Bg(ui.NRGBA(248, 250, 252, 255)).
    WithPad(ui.All(16)).
    WithRad(14).
    WithBorder(ui.Border{Width: 1, Color: ui.NRGBA(203, 213, 225, 255)}).
    Merge(ui.Elevation(2))

ui.ContainerDecoration(card, ui.Text("统一样式卡片"))
```

## 状态样式
```go
primary := ui.Bg(ui.NRGBA(37, 99, 235, 255)).
    WithPad(ui.Symmetric(10, 14)).
    WithRad(10).
    WithHover(ui.Bg(ui.NRGBA(29, 78, 216, 255))).
    WithPressed(ui.Bg(ui.NRGBA(30, 64, 175, 255))).
    WithDisabled(ui.Bg(ui.NRGBA(148, 163, 184, 255)))

ui.Button(
    ui.Text("提交", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
    ui.ButtonDecoration(primary),
)
```

## Resolve 与默认值
组件内部通常使用 `ResolveBg`、`ResolvePad`、`ResolveRad` 等方法读取 Decoration：

```go
bg := deco.ResolveBg(ctx.Theme().Surface)
pad := deco.ResolvePad(ui.All(12))
radius := deco.ResolveRad(8)
```

这意味着 Decoration 不需要设置所有字段，只需要设置你要覆盖的部分。
