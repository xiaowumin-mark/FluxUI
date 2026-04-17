<!-- fluxui-doc-meta
{
  "id": "center",
  "title": "Center 居中布局",
  "category": "布局系统",
  "order": 40,
  "summary": "Center 将子组件在可用区域中居中显示。",
  "example": { "id": "center_basic" },
  "apis": [
    "Center(child Widget) Widget"
  ]
}
-->

# Center 居中布局

## 组件说明
Center 用于快速做居中布局，不负责背景和边距。常用于空状态、加载中、局部提示等场景。

## 使用方法
- 仅做对齐时使用 `Center`。
- 与 `Container`、`FixedHeight` 组合可形成稳定的展示区域。

## 使用示例
```go
ui.FixedHeight(
    120,
    ui.Fill(
        ui.Container(
            ui.Style{Background: ui.NRGBA(240, 244, 248, 255), Radius: 8},
            ui.Center(ui.Text("居中内容")),
        ),
    ),
)
```
