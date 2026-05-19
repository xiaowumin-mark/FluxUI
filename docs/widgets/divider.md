<!-- fluxui-doc-meta
{
  "id": "divider",
  "title": "Divider 分割线",
  "category": "布局系统",
  "order": 80,
  "summary": "Divider 支持横向和纵向分割，适合内容分区。",
  "example": { "id": "divider_basic" },
  "apis": [
    "Divider(opts ...DividerOption) Widget",
    "DividerElement(opts ...DividerOption) Element",
    "DividerVertical(vertical bool) DividerOption",
    "DividerThickness(thickness float32) DividerOption",
    "DividerColor(col color.NRGBA) DividerOption",
    "DividerLength(length float32) DividerOption",
    "DividerMargin(insets Insets) DividerOption"
  ]
}
-->

# Divider 分割线

## 组件说明
Divider 用于视觉分组，帮助用户快速理解区域边界。可用于卡片内容分段、菜单项分组和双栏结构隔断。

## 使用方法
- 默认是横线。
- 设置 `DividerVertical(true)` 可切换为竖线。
- 厚度、颜色、边距都通过 Option 控制。

## 使用示例

### Legacy Widget
旧 `ui.Divider` / `Widget` 写法继续可用：

```go
ui.Column(
    ui.Text("Section A"),
    ui.Divider(
        ui.DividerThickness(1),
        ui.DividerColor(ui.NRGBA(203, 213, 225, 255)),
    ),
    ui.Text("Section B"),
)
```

### React-style Element
新代码可在 `RunElement` root 下组合 `ColumnElement` 和 `DividerElement`：

```go
func App(ctx *ui.Context) ui.Element {
    return ui.ColumnElement(
        ui.TextElement("Section A"),
        ui.DividerElement(
            ui.DividerThickness(1),
            ui.DividerColor(ui.NRGBA(203, 213, 225, 255)),
        ),
        ui.TextElement("Section B"),
    )
}
```
