package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func App(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	input := ui.UseState(ctx, "FluxUI 字体能力示例")

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(16)),
		ui.ColumnElement(
			ui.TextElement("字体能力示例", ui.TextSize(24), ui.TextFontWeight(ui.FontWeightSemiBold)),
			ui.SpacerElement(0, 8),
			ui.TextElement(
				fmt.Sprintf("系统字体族数量: %s, 全局默认字体: %s", familiesInfo(), globalFontInfo()),
				ui.TextColor(th.SurfaceMuted),
			),
			ui.SpacerElement(0, 12),
			ui.TextElement("这一段使用全局字体。"),
			ui.SpacerElement(0, 8),
			ui.WithFontElement(
				ui.FontFamily(localFontFamily()),
				ui.ColumnElement(
					ui.TextElement(fmt.Sprintf("这段使用局部字体作用域: %s", localFontFamily())),
					ui.SpacerElement(0, 6),
					ui.TextElement(
						"局部文本字重覆盖为 Bold。",
						ui.TextFontWeight(ui.FontWeightBold),
					),
				),
			),
			ui.SpacerElement(0, 12),
			ui.TextFieldElement(
				input.Value(),
				ui.InputOnChange(func(ctx *ui.Context, value string) {
					input.Set(value)
				}),
				ui.InputFontFamily(localFontFamily()),
			),
		),
	)
}

var (
	localFamily string
	globalFont  ui.FontSpec
)

func familiesInfo() string {
	families, _ := ui.DiscoverSystemFontFamilies()
	return fmt.Sprintf("%d", len(families))
}

func globalFontInfo() string {
	return globalFont.Family
}

func localFontFamily() string {
	return localFamily
}

func main() {
	families, _ := ui.DiscoverSystemFontFamilies()

	global := ui.DefaultFontSpec()
	if len(families) > 0 {
		global = ui.FontFamily(families[0]).WithStyle(ui.FontStyleRegular).WithWeight(ui.FontWeightNormal)
	}
	globalFont = global

	local := "serif"
	if len(families) > 1 {
		local = families[1]
	}
	localFamily = local

	_ = ui.RunElement(App,
		ui.Title("FluxUI Fonts"),
		ui.Size(640, 420),
		ui.WithSystemFonts(true),
		ui.WithDefaultFont(global),
	)
}
