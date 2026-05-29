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
    "type ShapeScale = theme.ShapeScale",
    "type TypeScale = theme.TypeScale",
    "type TextStyle = theme.TextStyle",
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
Theme 是 FluxUI 的全局视觉上下文，包含 MD3 语义色板、shape scale、type scale、兼容字段、默认字号、默认字体和系统字体回退配置。

```go
type Theme struct {
    Colors ColorScheme
    Shapes ShapeScale
    Types  TypeScale

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

## MD3 Token

- `Theme.Colors`: Material Design 3 color roles。
- `Theme.Shapes`: `None`、`ExtraSmall`、`Small`、`Medium`、`Large`、`ExtraLarge`、`Full`。
- `Theme.Types`: Display、Headline、Title、Body、Label 字体层级。
- 旧字段如 `Primary`、`Surface`、`TextColor` 继续保留，创建主题时从 `Colors` 同步。

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
- 新默认样式优先使用 `Colors` / `Shapes` / `Types`；兼容字段只作为旧代码过渡。
