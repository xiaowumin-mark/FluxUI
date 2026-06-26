package main

import (
	"image/color"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

const docsOverviewID = "docs_browser_overview"

func buildDocsBrowserOverviewDoc() widgetDoc {
	return widgetDoc{
		Meta: docMeta{
			ID:       docsOverviewID,
			Title:    "FluxUI 功能总览",
			Category: "使用指南",
			Order:    0,
			Summary:  "从这个入口快速了解 FluxUI 的 React-style 编程模型、Material 3 主题、组件示例、System API 和 API 速查能力。",
			Example:  docDemo{ID: "docs_overview"},
			APIs: []string{
				"RunElement(root Component, opts ...AppOption) error",
				"RunMulti(windows ...WindowSpec) error",
				"RenderElement(el Element) Widget",
				"ElementKey(el Element) string",
				"ElementInfo(el Element) ElementIdentity",
				"VisualRootBuilder(root Component) func(ctx *Context) Widget",
				"UseState[T any](ctx *Context, initial T) *state.State[T]",
				"UseAsync[T any](ctx *Context) *AsyncHandle[T]",
				"UseInterval(ctx *Context, interval time.Duration, fn func())",
				"LightThemeFromSeed(seed color.NRGBA) *Theme",
				"DarkThemeFromSeed(seed color.NRGBA) *Theme",
				"ColorSchemeFromSeed(seed color.NRGBA, opts ...ColorOption) ColorScheme",
				"LightColorSchemeFromSeed(seed color.NRGBA, opts ...ColorOption) ColorScheme",
				"DarkColorSchemeFromSeed(seed color.NRGBA, opts ...ColorOption) ColorScheme",
				"ThemeProviderElement(theme *Theme, child Element) Element",
				"WithColorSchemeDark(dark bool) ColorOption",
				"WithSecondarySeed(seed color.NRGBA) ColorOption",
				"WithTertiarySeed(seed color.NRGBA) ColorOption",
				"WithErrorSeed(seed color.NRGBA) ColorOption",
				"WithSuccessSeed(seed color.NRGBA) ColorOption",
				"WithWarningSeed(seed color.NRGBA) ColorOption",
				"WithDensity(scale DensityScale) AppOption",
				"DefaultFontSpec() FontSpec",
				"ParseFontFile(path string) ([]FontFace, error)",
				"ParseFontBytes(name string, data []byte) ([]FontFace, error)",
				"LoadFontsFromPaths(paths ...string) ([]FontFace, error)",
				"LoadFontsFromDir(dir string) ([]FontFace, error)",
				"LoadIconFontFromPath(id string, path string, opts ...IconFontOption) (IconFont, error)",
				"WithIconFonts(fonts ...IconFont) AppOption",
				"DiscoverSystemFonts() ([]FontFace, error)",
				"DiscoverSystemFontFamilies() ([]string, error)",
				"SystemFontDirs() []string",
				"ButtonElement(child Element, opts ...ButtonOption) Element",
				"RouterElement(routes ...RouteDeclaration) *RouterBuilder",
				"CurrentWindowID(ctx *Context) WindowID",
				"OpenFilesDialogContext(ctx *Context, callCtx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
				"SaveFileDialogContext(ctx *Context, callCtx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
				"PickFolderDialogContext(ctx *Context, callCtx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
				"system.Capabilities() system.CapabilitySet",
				"system.Supports(cap system.Capability) bool",
			},
		},
		Content: stringsJoinLines(
			"# FluxUI Docs Browser",
			"",
			"这个示例是 FluxUI 的交互式文档入口，用同一个应用展示组件、布局、主题、动画、运行时 hooks、路由、拖拽和 System API。",
			"",
			"## 推荐探索路径",
			"",
			"1. 使用左侧搜索框按组件名、API 签名或正文关键词搜索。",
			"2. 使用分类筛选切换布局、输入、反馈、导航、系统能力等功能域。",
			"3. 在每篇文档的内嵌示例里直接交互，或点击弹窗查看放大示例。",
			"4. 在 API 索引中复制单个签名或整组签名。",
			"",
			"## React-style 基本形态",
			"",
			"```go",
			"func App(ctx *ui.Context) ui.Element {",
			"    count := ui.UseState(ctx, 0)",
			"    return ui.FilledButtonElement(",
			"        ui.TextElement(fmt.Sprintf(\"count = %d\", count.Value())),",
			"        ui.OnClick(func(ctx *ui.Context) {",
			"            count.Set(count.Value() + 1)",
			"        }),",
			"    )",
			"}",
			"",
			"func main() {",
			"    _ = ui.RunElement(App, ui.Title(\"FluxUI\"))",
			"}",
			"```",
			"",
			"## 能力地图",
			"",
			"| 能力 | 示例入口 |",
			"| --- | --- |",
			"| Material 3 主题 | 主题色和 Dark 开关 |",
			"| 组件与布局 | 左侧分类中的组件文档 |",
			"| API 速查 | 搜索框与每篇文档 API 索引 |",
			"| System API | System API 文档和示例区域 |",
		),
		Path: "generated://docs-browser-overview",
	}
}

func buildDocsOverviewDemo(th *ui.Theme) ui.Element {
	if th == nil {
		th = docsBrowserTheme(defaultDocsThemeSeed, false)
	}
	return ui.ColumnElement(
		ui.TextElement("FluxUI capability map", ui.TextSize(16), ui.TextColor(th.Colors.OnSurface)),
		ui.VSpacerElement(10),
		ui.RowElement(
			ui.ExpandedElement(docsOverviewCard("React-style", "RunElement, ComponentElement, hooks, context, refs", th.Colors.PrimaryContainer, th.Colors.OnPrimaryContainer)),
			ui.HSpacerElement(10),
			ui.ExpandedElement(docsOverviewCard("Material 3", "Seed themes, dark mode, shapes, typography, state layers", th.Colors.SecondaryContainer, th.Colors.OnSecondaryContainer)),
		),
		ui.VSpacerElement(10),
		ui.RowElement(
			ui.ExpandedElement(docsOverviewCard("Component catalog", "Inputs, feedback, navigation, layout, media, drag/drop", th.Colors.TertiaryContainer, th.Colors.OnTertiaryContainer)),
			ui.HSpacerElement(10),
			ui.ExpandedElement(docsOverviewCard("System API", "Windows-first dialogs, tray, notifications, clipboard, window control", th.Colors.SurfaceContainerHigh, th.Colors.OnSurface)),
		),
		ui.VSpacerElement(12),
		ui.ContainerDecorationElement(
			ui.Bg(th.Colors.SurfaceContainerLow).
				WithPad(ui.All(12)).
				WithRad(10).
				WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
			ui.ColumnElement(
				ui.TextElement("How to use this browser", ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
				ui.VSpacerElement(6),
				ui.TextElement("Search API signatures, filter by category, copy code blocks, and open examples in a standalone popup.", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			),
		),
	)
}

func docsOverviewCard(title string, body string, bg color.NRGBA, fg color.NRGBA) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(bg).
			WithPad(ui.All(12)).
			WithRad(10),
		ui.ColumnElement(
			ui.TextElement(title, ui.TextSize(13), ui.TextColor(fg)),
			ui.VSpacerElement(6),
			ui.TextElement(body, ui.TextSize(11), ui.TextColor(fg)),
		),
	)
}

func stringsJoinLines(lines ...string) string {
	if len(lines) == 0 {
		return ""
	}
	out := lines[0]
	for _, line := range lines[1:] {
		out += "\n" + line
	}
	return out
}
