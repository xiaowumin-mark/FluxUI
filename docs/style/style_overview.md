<!-- fluxui-doc-meta
{
  "id": "style_overview",
  "title": "Style 样式总览",
  "category": "样式系统",
  "order": 495,
  "summary": "集中说明 FluxUI 样式库的颜色、边距、圆角、corner-shape、边框、阴影、渐变、图片填充、状态样式和 transform。",
  "example": { "id": "style_showcase" },
  "apis": [
    "type Decoration = style.Decoration",
    "type Insets = style.Insets",
    "type Border = style.Border",
    "type BoxShadow = style.BoxShadow",
    "type CornerShape = style.CornerShape",
    "const CornerRound, CornerSquare, CornerBevel, CornerNotch, CornerScoop, CornerSquircle",
    "Bg(c color.NRGBA) Decoration",
    "Pad(p Insets) Decoration",
    "Margin(m Insets) Decoration",
    "Rad(r float32) Decoration",
    "CornerShapeDeco(shape CornerShape) Decoration",
    "CornerShapesDeco(topLeft, topRight, bottomRight, bottomLeft CornerShape) Decoration",
    "BorderDeco(width float32, col color.NRGBA) Decoration",
    "Opacity(v float32) Decoration",
    "Circle() Decoration",
    "Shadow(offX, offY, blur float32, col color.NRGBA) Decoration",
    "Elevation(level int) Decoration",
    "LinearGrad(start, end image.Point, from, to color.NRGBA) Decoration",
    "ImageBg(src image.Image, fit ImageFillFit) Decoration",
    "TransformDeco(rotateDeg, scaleX, scaleY, transX, transY float32, origin TransformOrigin) Decoration",
    "Rotate(deg float32) Decoration",
    "ScaleDeco(sx, sy float32) Decoration",
    "TranslateDeco(tx, ty float32) Decoration",
    "Hover(d Decoration) Decoration",
    "Pressed(d Decoration) Decoration",
    "Focused(d Decoration) Decoration",
    "DisabledDeco(d Decoration) Decoration",
    "HoverBg(c color.NRGBA) Decoration",
    "PressedBg(c color.NRGBA) Decoration"
  ]
}
-->

# Style 样式总览

FluxUI 的样式库以 `Decoration` 为核心。它把背景、渐变、内边距、外边距、圆角、角形状、边框、阴影、透明度、图片填充、状态样式和 transform 放在同一个可组合对象里。

## 基础对象

- `Insets`: 描述 `Top / Right / Bottom / Left`，通过 `All`、`Symmetric`、`Only`、`LeftRight`、`TopBottom` 创建。
- `Border`: 描述边框宽度和颜色。
- `BoxShadow`: 描述 CSS box-shadow 风格阴影，`Elevation` 提供 Material Design 层级预设。
- `CornerShape`: 描述 CSS `corner-shape` 风格的角形状。
- `LinearGradient`: 描述线性渐变。
- `ImageFill`: 描述背景图和填充方式。
- `Transform2D`: 描述旋转、缩放、平移和变换原点。

`Decoration.Merge` 对同一个字段使用覆盖语义。需要同时旋转、缩放和平移时，使用一次 `TransformDeco` 或 `WithTransform(Transform2D{...})` 设置完整 transform，不要把 `Rotate`、`ScaleDeco`、`TranslateDeco` 连续 `Merge` 后期待自动叠加。

## Corner Shape

`corner-shape` 需要配合非零圆角半径使用。FluxUI 支持以下关键词：

- `CornerRound`: 默认圆角。
- `CornerSquare`: 方角。
- `CornerBevel`: 斜切角。
- `CornerNotch`: 内凹折角。
- `CornerScoop`: 内凹曲线角。
- `CornerSquircle`: 对齐 CSS `corner-shape: squircle`，按 `superellipse(2)` 计算，也就是传统超椭圆指数 4。

```go
deco := ui.Bg(th.Colors.PrimaryContainer).
    WithPad(ui.All(16)).
    WithRad(24).
    WithCornerShape(ui.CornerSquircle)
```

也可以按 CSS 四值顺序分别设置四个角：

```go
deco := ui.Bg(th.Colors.SurfaceContainer).
    WithRad(24).
    WithCornerShapes(
        ui.CornerSquircle,
        ui.CornerBevel,
        ui.CornerScoop,
        ui.CornerNotch,
    )
```

## 集中示例

```go
ui.ContainerDecorationElement(
    ui.Bg(th.Colors.SurfaceContainer).
        WithPad(ui.All(16)).
        WithMargin(ui.All(8)).
        WithRad(20).
        WithCornerShape(ui.CornerSquircle).
        WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}).
        WithHover(ui.Bg(th.Colors.PrimaryContainer)).
        WithPressed(ui.Bg(th.Colors.SecondaryContainer)).
        Merge(ui.Elevation(2)),
    ui.TextElement("Styled surface"),
)
```

## 使用建议

- 普通组件优先使用主题语义色，例如 `SurfaceContainer`、`PrimaryContainer`、`OutlineVariant`。
- 需要完整视觉外观时优先使用 `ContainerDecorationElement` 或组件的 `*Decoration` option。
- `CornerShape` 主要影响 `Decoration` 容器；默认 `CornerRound` 仍走 Gio 圆角快路径。
- 图片背景、阴影和复杂 transform 同时使用时，要优先验证目标平台的视觉效果。
