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
    "NewTheme(cs ColorScheme) *Theme",
    "NRGBA(r, g, b, a uint8) color.NRGBA"
  ]
}
-->

# ColorScheme 语义色板

## API 说明
ColorScheme 是 Theme 的语义色核心。它不仅包含背景色，还包含每个主色上的前景色，确保文本对比度可控。

```go
type ColorScheme struct {
    Primary      color.NRGBA
    OnPrimary    color.NRGBA
    Secondary    color.NRGBA
    OnSecondary  color.NRGBA
    Surface      color.NRGBA
    OnSurface    color.NRGBA
    SurfaceMuted color.NRGBA
    Background   color.NRGBA
    OnBackground color.NRGBA
    Error        color.NRGBA
    OnError      color.NRGBA
    Success      color.NRGBA
    OnSuccess    color.NRGBA
    Warning      color.NRGBA
    OnWarning    color.NRGBA
    Disabled     color.NRGBA
    Outline      color.NRGBA
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

## 使用建议
- `OnXxx` 应始终能在对应 `Xxx` 背景上清晰可读。
- Toast、按钮、表单焦点态等组件会读取语义色，不建议直接硬编码同一套色值。
- `Outline` 用于边框；`SurfaceMuted` 用于次级背景、轨道、分割区域。
