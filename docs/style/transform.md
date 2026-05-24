<!-- fluxui-doc-meta
{
  "id": "transform",
  "title": "Transform2D 变换",
  "category": "样式系统",
  "order": 550,
  "summary": "Transform2D 让 Decoration 支持旋转、缩放、平移和变换原点。",
  "example": { "id": "transform_basic" },
  "apis": [
    "type Transform2D = style.Transform2D",
    "type TransformOrigin = style.TransformOrigin",
    "const TransformCenter, TransformTopLeft, TransformTopRight, TransformBottomLeft, TransformBottomRight",
    "TransformDeco(rotateDeg, scaleX, scaleY, transX, transY float32, origin TransformOrigin) Decoration",
    "Rotate(deg float32) Decoration",
    "ScaleDeco(sx, sy float32) Decoration",
    "TranslateDeco(tx, ty float32) Decoration",
    "(Decoration).WithTransform(t Transform2D) Decoration",
    "(Transform2D).Affine(size image.Point) f32.Affine2D"
  ]
}
-->

# Transform2D 变换

## API 说明
Transform2D 描述 2D 仿射变换，可用于装饰容器的旋转、缩放和平移。

```go
type Transform2D struct {
    RotateDeg  float32
    ScaleX     float32
    ScaleY     float32
    TranslateX float32
    TranslateY float32
    Origin     TransformOrigin
}
```

`ScaleX` 或 `ScaleY` 为 0 时会按 1 处理，避免组件被意外压扁。

## 使用示例
```go
deco := ui.Bg(ui.NRGBA(239, 246, 255, 255)).
    WithPad(ui.All(14)).
    WithRad(12).
    Merge(ui.Rotate(-4))

ui.ContainerDecoration(deco, ui.Text("轻微旋转"))
```

组合变换：

```go
deco := ui.TransformDeco(
    8,       // rotateDeg
    1.05,    // scaleX
    1.05,    // scaleY
    4, 0,    // translateX, translateY
    ui.TransformCenter,
)
```

## 使用建议
- 小幅旋转适合强调标签或徽章。
- 缩放和位移适合配合 `UseAnimatedDecoration` 做动效。
- 变换可能影响视觉位置，但布局占位仍以原尺寸为主。
