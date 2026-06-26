<!-- fluxui-doc-meta
{
  "id": "icon_fonts",
  "title": "Icon Fonts 图标字体",
  "category": "使用指南",
  "order": 126,
  "summary": "说明如何使用内置 Material Design 3 图标字体，以及如何追加自定义图标字体。",
  "example": { "id": "icon_fonts" },
  "apis": [
    "Icon(name string, opts ...IconOption) Widget",
    "IconElement(name string, opts ...IconOption) Element",
    "IconUseFont(id string) IconOption",
    "IconFontFamily(family string) IconOption",
    "LoadIconFontFromPath(id string, path string, opts ...IconFontOption) (IconFont, error)",
    "WithIconFonts(fonts ...IconFont) AppOption"
  ]
}
-->

# Icon Fonts 图标字体

FluxUI 的 `Icon` 支持图标字体 ligature。引入内置 MD3 图标字体包后，`Icon("home")`、`Icon("search")`、`Icon("settings")` 等名称会直接渲染为 Material Symbols 图标。

## 内置 MD3 图标

应用只需要空白导入一次：

```go
import (
    ui "github.com/xiaowumin-mark/FluxUI/ui"
    _ "github.com/xiaowumin-mark/FluxUI/icons/md3"
)
```

之后即可直接使用官方 Material Symbols 名称：

```go
ui.IconElement("home", ui.IconSize(24))
ui.IconButtonElement(ui.IconElement("search"))
ui.FloatingActionButtonElement(ui.IconElement("add"))
```

`icons/md3` 包通过 `go:embed` 内置 `MaterialSymbolsOutlined.ttf`，并在 `init` 中注册为默认图标字体。普通 `Text` 不会使用该字体，只有 `Icon` 会自动选择默认图标字体。

## 指定图标字体

如果同时注册了多个图标字体，可以在单个图标上指定字体 ID：

```go
ui.IconElement("github", ui.IconUseFont("brand"))
```

也可以直接指定字体族名，用于系统已安装或外部 shaper 可解析的图标字体：

```go
ui.IconElement("home", ui.IconFontFamily("Material Symbols Outlined"))
```

## 加载自定义字体

开发者可以从文件加载额外图标字体，并通过 `WithIconFonts` 加入当前应用主题：

```go
brand, err := ui.LoadIconFontFromPath(
    "brand",
    "assets/BrandIcons.ttf",
    ui.IconFontFamilyName("Brand Icons"),
)
if err != nil {
    panic(err)
}

_ = ui.RunElement(App,
    ui.WithIconFonts(brand),
)
```

需要把自定义字体设为默认图标字体时：

```go
brand, err := ui.LoadIconFontFromPath(
    "brand",
    "assets/BrandIcons.ttf",
    ui.IconFontFamilyName("Brand Icons"),
    ui.IconFontDefault(true),
)
```

也可以在启动应用时指定默认字体 ID：

```go
_ = ui.RunElement(App,
    ui.WithIconFonts(brand),
    ui.WithDefaultIconFont("brand"),
)
```

如果没有导入内置 MD3 包，也没有注册任何图标字体，`Icon` 会回退为普通文本显示，方便定位缺失配置。
