<!-- fluxui-doc-meta
{
  "id": "gradient",
  "title": "LinearGradient 渐变",
  "category": "样式系统",
  "order": 540,
  "summary": "LinearGradient 用于 Decoration 的线性渐变背景。",
  "example": { "id": "gradient_basic" },
  "apis": [
    "type LinearGradient = style.LinearGradient",
    "type LinearGradient struct { Start image.Point; End image.Point; From color.NRGBA; To color.NRGBA }",
    "LinearGrad(start, end image.Point, from, to color.NRGBA) Decoration",
    "(Decoration).WithGradient(g LinearGradient) Decoration",
    "(Decoration).ResolveGradient() *LinearGradient"
  ]
}
-->

# LinearGradient 渐变

## API 说明
`LinearGradient` 描述从 `Start` 到 `End` 的两色线性渐变，坐标位于组件本地坐标系。

```go
type LinearGradient struct {
    Start image.Point
    End   image.Point
    From  color.NRGBA
    To    color.NRGBA
}
```

当 Decoration 同时设置 `Gradient` 和 `Background` 时，渲染优先使用渐变。

## 使用示例
```go
hero := ui.LinearGrad(
    image.Point{X: 0, Y: 0},
    image.Point{X: 240, Y: 120},
    ui.NRGBA(14, 165, 233, 255),
    ui.NRGBA(34, 197, 94, 255),
).WithPad(ui.All(18)).WithRad(18)

ui.ContainerDecoration(
    hero,
    ui.Text("渐变背景", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
)
```

## 使用建议
- `Start` / `End` 不需要等于最终组件尺寸，但应能表达方向。
- 大面积渐变建议搭配白色或深色高对比文字。
- 渐变与阴影搭配时，先确认文字对比度。
