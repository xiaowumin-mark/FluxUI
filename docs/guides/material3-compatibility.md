<!-- fluxui-doc-meta
{
  "id": "material3_compatibility",
  "title": "Material 3 版本与兼容策略",
  "category": "使用指南",
  "order": 121,
  "summary": "说明 MD3 默认视觉变化、legacy API 定位、ClickArea 迁移、旧 Theme 字段和恢复旧样式的最小方案。",
  "example": { "id": "material3_showcase" },
  "apis": [
    "RunElement(root Component, opts ...AppOption) error",
    "Run(root func(ctx *Context) Widget, opts ...AppOption) error",
    "FromWidget(w Widget) Element",
    "PressableElement(child Element, onClick func(ctx *Context), opts ...PressableOption) Element",
    "ClickAreaElement(child Element, onClick func(ctx *Context), opts ...ClickAreaOption) Element",
    "WithTheme(th *Theme) AppOption",
    "ThemeProviderElement(th *Theme, child Element) Element"
  ]
}
-->

# Material 3 版本与兼容策略

## Release notes 摘要

FluxUI 默认视觉已迁移到 Material Design 3。升级后，默认 Light/Dark 色板、组件圆角、文字层级、surface container、disabled 颜色、state layer、ripple、focus ring、部分组件 spacing 和 compact density 行为都会更接近 MD3。

这属于默认样式变化，不是 API 删除。旧项目通常可以继续编译运行，但如果依赖默认颜色、圆角、阴影、按钮高度或输入框外观，升级后需要做一次视觉验收。

建议在项目 release notes 中标记：

- 默认主题现在以 MD3 color roles、shape scale、type scale 和 density 为基础。
- Button、TextField、Card、Selection controls、Tabs、Navigation、Menu、Dialog、Snackbar、Tooltip、Chip、SearchBar、ProgressIndicators 等默认视觉可能变化。
- 新代码推荐 Element API 与明确的 MD3 variant API。
- legacy Widget API、旧 root API、旧 Router API 和旧 Theme 扁平字段继续保留。
- `ClickArea` 继续兼容，但新代码推荐 `Pressable`。

## 兼容承诺

当前阶段不删除旧 API，也不添加代码级 `Deprecated:` 注释。

| API / 行为 | 当前定位 | 推荐动作 |
| --- | --- | --- |
| `RunElement` / `Element` / `Component` | 新项目推荐入口 | 新页面和新示例优先使用 |
| `Run` / `App` / `Window` / `RunMulti` | 稳定兼容入口 | 旧项目可继续使用 |
| `Widget` API | 稳定兼容层 | 不要求一次性迁移 |
| `FromWidget` | 长期 escape hatch | 用于新旧 API 混用 |
| 旧 Router API | 稳定兼容入口 | 新代码优先使用 hooks 和 `RouterElement` |
| `ClickArea` | 兼容旧名称 | 新代码改用 `Pressable` |
| `Theme.Primary` / `Surface` / `TextColor` 等扁平字段 | 兼容字段 | 新代码优先读取 `Theme.Colors` |

如果未来引入 breaking change，必须同时满足：

- 有明确 release notes。
- 有迁移指南和替代 API。
- 有足够长的版本窗口，不在短期小版本内删除旧入口。
- 旧项目能判断继续使用兼容 API，还是迁移到新 API。

## 应该用新 API 还是兼容 API

新项目或新页面：

```go
func App(ctx *ui.Context) ui.Element {
    return ui.ColumnElement(
        ui.TextElement("Material 3 UI"),
        ui.FilledButtonElement(ui.TextElement("Action")),
    )
}

func main() {
    ui.RunElement(App, ui.Title("MD3 App"))
}
```

旧项目可以保持原入口：

```go
func App(ctx *ui.Context) ui.Widget {
    return ui.Column(
        ui.Text("Legacy root can stay"),
        ui.Button(ui.Text("Action")),
    )
}

func main() {
    ui.Run(App)
}
```

新旧混用时，用 `FromWidget`：

```go
func App(ctx *ui.Context) ui.Element {
    return ui.ColumnElement(
        ui.TextElement("Element page"),
        ui.FromWidget(ui.Text("Legacy widget")),
    )
}
```

## ClickArea 迁移

`ClickArea` 是旧名称，继续保留为兼容 API。新代码优先使用 `Pressable`，因为它更准确表达“无固定视觉样式的可按压区域”。

迁移方式：

| 旧写法 | 新写法 |
| --- | --- |
| `ClickArea(child, onClick)` | `Pressable(child, onClick)` |
| `ClickAreaElement(child, onClick)` | `PressableElement(child, onClick)` |
| `ClickAreaRef` | `PressableRef` |

`ClickArea` 不会自动获得 MD3 button 背景、ripple 或固定 state layer。主操作按钮仍应使用 Button / IconButton / FAB 等 Material 组件。

## 旧 Theme 字段

`Theme.Colors` 是新样式的主入口。旧字段继续同步和保留，主要服务旧代码：

- `Primary` 对应 `Colors.Primary`
- `Surface` 对应 `Colors.Surface`
- `SurfaceMuted` 对应 `Colors.SurfaceVariant`
- `TextColor` 对应 `Colors.OnSurface`
- `TextOnPrimary` 对应 `Colors.OnPrimary`
- `Disabled` 对应 `Colors.Disabled`

新代码建议读取 `th.Colors.Primary`、`th.Colors.SurfaceContainer`、`th.Shapes.Medium`、`th.Types.BodyMedium`。旧代码读取 `th.Primary`、`th.Surface`、`th.TextColor` 仍可运行。

## 恢复旧样式的最小方案

FluxUI 不提供逐像素复原旧默认样式的开关。需要降低 MD3 默认化影响时，推荐自定义 Theme，至少复原主要颜色、surface、文字色和圆角。

```go
package main

import (
    "image/color"

    ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func legacyLikeTheme() *ui.Theme {
    th := ui.NewTheme(ui.LightColors())

    th.Colors.Primary = color.NRGBA{R: 33, G: 150, B: 243, A: 255}
    th.Colors.OnPrimary = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
    th.Colors.Surface = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
    th.Colors.OnSurface = color.NRGBA{R: 33, G: 33, B: 33, A: 255}
    th.Colors.SurfaceVariant = color.NRGBA{R: 245, G: 245, B: 245, A: 255}
    th.Colors.SurfaceContainer = th.Colors.Surface
    th.Colors.SurfaceContainerLow = th.Colors.Surface
    th.Colors.SurfaceContainerHigh = th.Colors.SurfaceVariant
    th.Colors.SurfaceMuted = th.Colors.SurfaceVariant
    th.Colors.Outline = color.NRGBA{R: 189, G: 189, B: 189, A: 255}

    th.Shapes = ui.ShapeScale{
        None:       0,
        ExtraSmall: 4,
        Small:      4,
        Medium:     6,
        Large:      8,
        ExtraLarge: 12,
        Full:       999,
    }

    // 同步旧字段，服务仍在读取 th.Primary / th.Surface / th.TextColor 的旧代码。
    th.Primary = th.Colors.Primary
    th.Surface = th.Colors.Surface
    th.SurfaceMuted = th.Colors.SurfaceMuted
    th.TextColor = th.Colors.OnSurface
    th.TextOnPrimary = th.Colors.OnPrimary
    th.Disabled = th.Colors.Disabled

    return th
}

func main() {
    ui.RunElement(App, ui.WithTheme(legacyLikeTheme()))
}
```

局部恢复某个区域的旧样式时，用 `ThemeProviderElement` 包住子树即可。更细的差异，例如某个按钮的圆角、边框、padding、背景色，应优先通过组件 option 或 `Decoration` 覆盖，不要 fork 默认组件。

## 升级检查清单

- 运行 `go test ./...` 和 `go vet ./...`。
- 涉及默认视觉时运行 `make visual`，或执行等价 visual 测试命令。
- 检查 Light/Dark、按钮、输入框、卡片、导航、弹层和禁用态。
- 检查旧项目是否依赖 `Theme.Primary`、`Theme.Surface`、`Theme.TextColor` 等扁平字段。
- 检查旧 `ClickArea` 是否应迁移为 `Pressable`。
- 如果业务有严格品牌规范，使用 `ThemeFromSeed` 或自定义 `ColorScheme`，不要依赖默认紫色 baseline。
