<!-- fluxui-doc-meta
{
  "id": "app_bar",
  "title": "AppBar 顶部栏",
  "category": "导航组件",
  "order": 410,
  "summary": "AppBar 用于页面顶部导航与全局操作。",
  "example": { "id": "app_bar_basic" },
  "apis": [
    "AppBar(title Widget, opts ...AppBarOption) Widget",
    "AppBarLeading(leading Widget) AppBarOption",
    "AppBarActions(actions ...Widget) AppBarOption",
    "AppBarHeight(height float32) AppBarOption",
    "AppBarBackground(col color.NRGBA) AppBarOption"
  ]
}
-->

# AppBar 顶部栏

## 组件说明
AppBar 是页面级顶部导航组件，适合放页面标题、返回入口和全局操作按钮。

## 使用方法
- 标题作为第一个参数传入。
- 左侧入口用 `AppBarLeading`。
- 右侧操作组用 `AppBarActions`。

## 使用示例
```go
ui.AppBar(
    ui.Text("文档中心"),
    ui.AppBarActions(
        ui.Button(ui.Text("刷新")),
    ),
)
```
