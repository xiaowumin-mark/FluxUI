<!-- fluxui-doc-meta
{
  "id": "insets",
  "title": "Insets 边距",
  "category": "样式系统",
  "order": 510,
  "summary": "Insets 描述上下左右四个方向的 dp 边距，用于 Padding、Margin 和组件内边距选项。",
  "example": { "id": "insets_basic" },
  "apis": [
    "type Insets = style.Insets",
    "All(value float32) Insets",
    "Symmetric(vertical, horizontal float32) Insets",
    "Only(top, right, bottom, left float32) Insets",
    "LeftRight(v float32) Insets",
    "TopBottom(v float32) Insets",
    "style.Horizontal(v float32) Insets",
    "style.Vertical(v float32) Insets",
    "(Insets).IsZero() bool"
  ]
}
-->

# Insets 边距

## API 说明
`Insets` 用于描述四个方向的距离，单位是 dp。它既可以作为布局组件的参数，也可以放入 Decoration 的 padding / margin 字段。

```go
type Insets struct {
    Top    float32
    Right  float32
    Bottom float32
    Left   float32
}
```

## 快捷构造
- `All(12)`：四边都是 12。
- `Symmetric(8, 16)`：上下 8，左右 16。
- `Only(4, 8, 12, 16)`：依次设置 top/right/bottom/left。
- `LeftRight(16)`：左右 16。
- `TopBottom(10)`：上下 10。

## 使用示例
```go
ui.Padding(
    ui.Symmetric(8, 16),
    ui.Text("左右更宽，上下更窄"),
)

ui.ContainerDecoration(
    ui.Bg(ui.NRGBA(241, 245, 249, 255)).
        WithPad(ui.All(12)).
        WithMargin(ui.TopBottom(8)).
        WithRad(8),
    ui.Text("带 margin 和 padding 的装饰容器"),
)
```

## 使用建议
- 组件内部留白使用 padding。
- 组件之间距离使用 margin 或外层 `Padding`。
- 同一页面建议固定一组间距尺度，例如 4 / 8 / 12 / 16 / 24。
