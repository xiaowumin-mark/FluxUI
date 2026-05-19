<!-- fluxui-doc-meta
{
  "id": "column",
  "title": "Column 纵向布局",
  "category": "布局系统",
  "order": 20,
  "summary": "Column 按从上到下顺序布局子组件。",
  "example": { "id": "column_basic" },
  "apis": [
    "Column(children ...Widget) Widget",
    "ColumnElement(children ...Element) Element",
    "Flexed(weight float32, child Widget) Widget",
    "Expanded(child Widget) Widget"
  ]
}
-->

# Column 纵向布局

## 组件说明
Column 是页面主骨架最常用的布局容器，适合组织页面标题、表单区、列表区和底部操作区。

## 使用方法
- 多段内容按视觉顺序声明在 `Column(...)` 中。
- 需要占满剩余高度的区域使用 `Expanded(...)` 包裹。
- 长内容建议把 `ScrollView` 放到 `Expanded` 区域里。

## 使用示例

### Legacy Widget
旧 `ui.Column` / `Widget` 写法继续可用：

```go
ui.Column(
    ui.Text("页面标题", ui.TextSize(20)),
    ui.Padding(ui.Insets{Top: 8}, ui.Text("说明文本")),
    ui.Expanded(ui.Spacer(0, 0)),
    ui.Button(ui.Text("提交")),
)
```

### React-style Element
新代码可在 `RunElement` root 下返回 `ColumnElement`。当前阶段 `Expanded` / `Flexed` 的原生 Element wrapper 尚未冻结，基础纵向布局优先用 Element wrapper：

```go
func Page(ctx *ui.Context) ui.Element {
    return ui.ColumnElement(
        ui.TextElement("页面标题", ui.TextSize(20)),
        ui.PaddingElement(ui.Insets{Top: 8}, ui.TextElement("说明文本")),
        ui.SpacerElement(0, 24),
        ui.TextElement("提交"),
    )
}
```
