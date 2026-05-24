<!-- fluxui-doc-meta
{
  "id": "border",
  "title": "Border 边框",
  "category": "样式系统",
  "order": 520,
  "summary": "Border 描述边框宽度和颜色，通常通过 Decoration.WithBorder 或 BorderDeco 使用。",
  "example": { "id": "border_basic" },
  "apis": [
    "type Border = style.Border",
    "type Border struct { Width float32; Color color.NRGBA }",
    "BorderDeco(width float32, col color.NRGBA) Decoration",
    "(Decoration).WithBorder(b style.Border) Decoration",
    "(Border).IsZero() bool"
  ]
}
-->

# Border 边框

## API 说明
Border 是 Decoration 的边框字段，包含宽度和颜色。宽度单位为 dp，颜色使用 `color.NRGBA`。

```go
type Border struct {
    Width float32
    Color color.NRGBA
}
```

`Width <= 0` 或 `Color.A == 0` 时，边框视为不可见。

## 使用示例
```go
outline := ui.Bg(ui.NRGBA(255, 255, 255, 255)).
    WithPad(ui.All(14)).
    WithRad(12).
    WithBorder(ui.Border{
        Width: 1,
        Color: ui.NRGBA(203, 213, 225, 255),
    })

ui.ContainerDecoration(outline, ui.Text("带边框的内容块"))
```

也可以使用快捷函数：

```go
deco := ui.Bg(ui.NRGBA(255, 255, 255, 255)).
    Merge(ui.BorderDeco(1, ui.NRGBA(203, 213, 225, 255))).
    WithRad(10)
```

## 使用建议
- 表单控件边框应优先使用主题色：`ctx.Theme().Colors.Outline`。
- 激活态边框建议使用 `ctx.Theme().Primary`。
- 高密度列表中，优先使用浅边框或 `Divider`，避免视觉过重。
