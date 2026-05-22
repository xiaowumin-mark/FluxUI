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
    "FromWidget(w Widget) Element",
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

## React-style 状态

- 当前 `AppBar` 仍以 legacy `Widget` composition 作为稳定实现。
- React-style root 中可先使用 `FromWidget(ui.AppBar(...))` 桥接完整 AppBar 子树。
- `title`、`leading`、`actions` 当前接收 legacy `Widget`，不是 `Element` child；不要在文档中假定 `AppBarElement` 已稳定。
- 本阶段不冻结 `AppBarElement` API 名称；后续如引入 Element wrapper，需要先明确 action 子树 identity、布局约束和事件归属。
