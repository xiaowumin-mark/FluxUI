<!-- fluxui-doc-meta
{
  "id": "color_scheme",
  "title": "ColorScheme 语义色板",
  "category": "主题系统",
  "order": 610,
  "summary": "ColorScheme 定义 Primary、Surface、Error、Warning 等语义颜色及其前景色。",
  "example": { "id": "color_scheme_basic" },
  "apis": [
    "type ColorScheme = theme.ColorScheme",
    "LightColors() ColorScheme",
    "DarkColors() ColorScheme",
    "ColorSchemeFromSeed(seed color.NRGBA, opts ...ColorOption) ColorScheme",
    "LightColorSchemeFromSeed(seed color.NRGBA, opts ...ColorOption) ColorScheme",
    "DarkColorSchemeFromSeed(seed color.NRGBA, opts ...ColorOption) ColorScheme",
    "ThemeFromSeed(seed color.NRGBA, opts ...ColorOption) *Theme",
    "LightThemeFromSeed(seed color.NRGBA, opts ...ColorOption) *Theme",
    "DarkThemeFromSeed(seed color.NRGBA, opts ...ColorOption) *Theme",
    "WithColorSchemeDark(dark bool) ColorOption",
    "WithSecondarySeed(seed color.NRGBA) ColorOption",
    "WithTertiarySeed(seed color.NRGBA) ColorOption",
    "WithErrorSeed(seed color.NRGBA) ColorOption",
    "WithSuccessSeed(seed color.NRGBA) ColorOption",
    "WithWarningSeed(seed color.NRGBA) ColorOption",
    "NewTheme(cs ColorScheme) *Theme",
    "NRGBA(r, g, b, a uint8) color.NRGBA"
  ]
}
-->

# ColorScheme 语义色板

## API 说明
ColorScheme 是 Theme 的语义色核心。FluxUI 默认色板已按 Material Design 3 roles 组织：主色、容器色、surface 层级、outline、inverse surface、scrim、shadow 以及项目保留的 success/warning 业务色。

```go
type ColorScheme struct {
    Primary            color.NRGBA
    OnPrimary          color.NRGBA
    PrimaryContainer   color.NRGBA
    OnPrimaryContainer color.NRGBA
    Secondary          color.NRGBA
    OnSecondary        color.NRGBA
    SecondaryContainer color.NRGBA
    OnSecondaryContainer color.NRGBA
    Tertiary           color.NRGBA
    OnTertiary         color.NRGBA
    Error              color.NRGBA
    OnError            color.NRGBA
    ErrorContainer     color.NRGBA
    OnErrorContainer   color.NRGBA
    Background         color.NRGBA
    OnBackground       color.NRGBA
    Surface            color.NRGBA
    OnSurface          color.NRGBA
    SurfaceVariant     color.NRGBA
    OnSurfaceVariant   color.NRGBA
    SurfaceContainerLowest  color.NRGBA
    SurfaceContainerLow     color.NRGBA
    SurfaceContainer        color.NRGBA
    SurfaceContainerHigh    color.NRGBA
    SurfaceContainerHighest color.NRGBA
    Outline           color.NRGBA
    OutlineVariant    color.NRGBA
    InverseSurface    color.NRGBA
    InverseOnSurface  color.NRGBA
    InversePrimary    color.NRGBA
    Scrim             color.NRGBA
    Shadow            color.NRGBA
    Success           color.NRGBA
    OnSuccess         color.NRGBA
    Warning           color.NRGBA
    OnWarning         color.NRGBA
    Disabled          color.NRGBA
    SurfaceMuted      color.NRGBA
}
```

## 预设色板
- `LightColors()`：浅色主题预设。
- `DarkColors()`：深色主题预设。

## 自定义色板
```go
cs := ui.LightColors()
cs.Primary = ui.NRGBA(20, 184, 166, 255)
cs.OnPrimary = ui.NRGBA(255, 255, 255, 255)
cs.Warning = ui.NRGBA(245, 158, 11, 255)
cs.OnWarning = ui.NRGBA(17, 24, 39, 255)

th := ui.NewTheme(cs)
```

## Seed 色板
Material 3 seed API 可从主色自动生成 light/dark 语义色板，并可为 secondary、tertiary、error、success、warning 提供独立 seed。

```go
seed := ui.NRGBA(56, 106, 232, 255)
light := ui.LightColorSchemeFromSeed(
    seed,
    ui.WithSecondarySeed(ui.NRGBA(0, 143, 104, 255)),
    ui.WithWarningSeed(ui.NRGBA(183, 100, 0, 255)),
)
dark := ui.DarkColorSchemeFromSeed(seed)

_ = ui.NewTheme(light)
_ = ui.DarkThemeFromSeed(seed)
_ = ui.ColorSchemeFromSeed(seed, ui.WithColorSchemeDark(true))
_ = dark
```

## MD3 使用建议
- `OnXxx` 应始终能在对应 `Xxx` 背景上清晰可读。
- Button、TextField、Card、Selection、Navigation、Overlay 默认样式会优先读取 MD3 roles。
- `SurfaceContainer*` 用于卡片、导航栏、弹层等容器层级。
- `Outline` 用于交互控件边框；`OutlineVariant` 用于较弱的信息容器边框。
- `SurfaceMuted` 是兼容字段，新代码优先使用 `SurfaceVariant` 或 `SurfaceContainer*`。
