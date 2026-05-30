<!-- fluxui-doc-meta
{
  "id": "stack",
  "title": "Stack 层叠布局",
  "category": "布局系统",
  "order": 30,
  "summary": "Stack 用于在同一区域叠放多个组件。",
  "example": { "id": "stack_basic" },
  "apis": [
    "Stack(children ...Widget) Widget",
    "StackElement(children ...Element) Element"
  ]
}
-->

# Stack 层叠布局

## 组件说明
Stack 适合遮罩、角标、浮层提示和局部叠加场景。第一个子组件通常作为底层内容，后续组件按顺序覆盖在上方。

## 使用方法
- 基础内容放在第一个子组件中。
- 覆盖层（提示、角标、浮窗）作为后续子组件。
- 建议外层设置固定高度或填充约束，避免叠层区域尺寸不稳定。

## 使用示例

### React-style Element
新代码可在 `RunElement` root 下返回 `StackElement`：

```go
func App(ctx *ui.Context) ui.Element {
    return ui.StackElement(
        ui.ContainerElement(
            ui.Style{Background: ui.NRGBA(240, 244, 248, 255), Radius: 8},
            ui.SpacerElement(240, 120),
        ),
        ui.CenterElement(ui.TextElement("Center Layer")),
    )
}
```

### Legacy Widget
旧 `ui.Stack` / `Widget` 写法继续可用：

```go
ui.Stack(
    ui.Fill(ui.Container(ui.Style{Background: ui.NRGBA(240, 244, 248, 255)}, ui.Spacer(0, 0))),
    ui.Center(ui.Text("Center Layer")),
)
```
