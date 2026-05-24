<!-- fluxui-doc-meta
{
  "id": "box_shadow",
  "title": "BoxShadow 阴影",
  "category": "样式系统",
  "order": 530,
  "summary": "BoxShadow 为装饰容器提供偏移和模糊阴影，Elevation 提供常用预设。",
  "example": { "id": "shadow_basic" },
  "apis": [
    "type BoxShadow = style.BoxShadow",
    "type BoxShadow struct { OffsetX float32; OffsetY float32; Blur float32; Color color.NRGBA }",
    "Shadow(offX, offY, blur float32, col color.NRGBA) Decoration",
    "Elevation(level int) Decoration",
    "style.ElevationBoxShadow(level int) BoxShadow",
    "(Decoration).WithShadow(s style.BoxShadow) Decoration"
  ]
}
-->

# BoxShadow 阴影

## API 说明
`BoxShadow` 描述盒阴影，单位为 dp。FluxUI 使用多层半透明偏移矩形近似模糊效果。

```go
type BoxShadow struct {
    OffsetX float32
    OffsetY float32
    Blur    float32
    Color   color.NRGBA
}
```

## Elevation 预设
`Elevation(level)` 是推荐入口，level 越大层级越高：

- `1`：按钮 hover、轻浮起。
- `2`：卡片。
- `3`：浮卡或 FAB。
- `4`：对话框。
- `5`：模态、抽屉。

## 使用示例
```go
card := ui.Bg(ui.NRGBA(255, 255, 255, 255)).
    WithPad(ui.All(16)).
    WithRad(16).
    Merge(ui.Elevation(2))

ui.ContainerDecoration(card, ui.Text("带阴影的卡片"))
```

自定义阴影：

```go
deco := ui.Shadow(0, 8, 24, ui.NRGBA(15, 23, 42, 48)).
    WithBg(ui.NRGBA(255, 255, 255, 255)).
    WithRad(18).
    WithPad(ui.All(16))
```

## 使用建议
- 阴影适合强调浮层，不适合给所有列表项都加。
- 暗色主题中应降低黑色阴影透明度，或使用边框替代。
- 弹窗、Popup、浮动卡片可以使用 level 3~5。
