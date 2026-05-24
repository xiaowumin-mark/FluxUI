<!-- fluxui-doc-meta
{
  "id": "theme",
  "title": "Theme 主题",
  "category": "主题系统",
  "order": 600,
  "summary": "Theme 统一管理语义色、默认字号、默认字体和系统字体回退。",
  "example": { "id": "theme_basic" },
  "apis": [
    "type Theme = theme.Theme",
    "NewTheme(cs ColorScheme) *Theme",
    "WithTheme(th *Theme) AppOption",
    "UseTheme(ctx *Context) *Theme",
    "ThemeProviderElement(th *Theme, child Element) Element",
    "UseFont(ctx *Context) FontSpec",
    "WithDefaultFont(spec FontSpec) AppOption",
    "WithFonts(faces ...FontFace) AppOption",
    "WithSystemFonts(enabled bool) AppOption"
  ]
}
-->

# Theme 主题

## API 说明
Theme 是 FluxUI 的全局视觉上下文，包含语义色板、兼容字段、默认字号、默认字体和系统字体回退配置。

```go
type Theme struct {
    Colors ColorScheme

    Primary        color.NRGBA
    Surface        color.NRGBA
    SurfaceMuted   color.NRGBA
    TextColor      color.NRGBA
    TextOnPrimary  color.NRGBA
    Disabled       color.NRGBA
    TextSize       float32
    DefaultFont    FontSpec
    UseSystemFonts bool
    Fonts          []FontFace
}
```

## 全局主题
```go
func main() {
    th := ui.NewTheme(ui.LightColors())
    th.TextSize = 15

    ui.RunElement(App, ui.Title("Theme Demo"), ui.WithTheme(th))
}
```

## 组件中读取主题
```go
func App(ctx *ui.Context) ui.Element {
    th := ui.UseTheme(ctx)
    return ui.TextElement("主题色文本", ui.TextColor(th.Primary))
}
```

## 局部主题
`ThemeProviderElement` 可以只覆盖某棵 Element 子树：

```go
dark := ui.NewTheme(ui.DarkColors())

return ui.ThemeProviderElement(
    dark,
    ui.ContainerDecorationElement(
        ui.Bg(dark.Surface).WithPad(ui.All(16)).WithRad(12),
        ui.TextElement("局部深色主题", ui.TextColor(dark.TextColor)),
    ),
)
```

## 使用建议
- 组件样式优先读取 `ctx.Theme()`，避免硬编码黑白色。
- 新主题应先定义 `ColorScheme`，再通过 `NewTheme` 创建。
- 局部主题适合预览卡片、嵌入式面板、主题切换示例。
