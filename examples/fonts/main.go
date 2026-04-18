package main

import (
	"fmt"

	"fluxui/ui"
)

func main() {
	families, _ := ui.DiscoverSystemFontFamilies()

	global := ui.DefaultFontSpec()
	if len(families) > 0 {
		global = ui.FontFamily(families[0]).WithStyle(ui.FontStyleRegular).WithWeight(ui.FontWeightNormal)
	}

	local := "serif"
	if len(families) > 1 {
		local = families[1]
	}

	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)
		input := ui.State[string](ctx)
		if input.Value() == "" {
			input.Set("FluxUI 字体能力示例")
		}

		return ui.Container(
			ui.Style{
				Background: th.Surface,
				Padding:    ui.All(16),
			},
			ui.Column(
				ui.Text("字体能力示例", ui.TextSize(24), ui.TextFontWeight(ui.FontWeightSemiBold)),
				ui.Padding(
					ui.Insets{Top: 8},
					ui.Text(fmt.Sprintf("系统字体族数量: %d", len(families)), ui.TextColor(th.SurfaceMuted)),
				),
				ui.Padding(
					ui.Insets{Top: 4},
					ui.Text(fmt.Sprintf("全局默认字体: %s", global.Family), ui.TextColor(th.SurfaceMuted)),
				),
				ui.Padding(
					ui.Insets{Top: 12},
					ui.Text("这一段使用全局字体。"),
				),
				ui.Padding(
					ui.Insets{Top: 8},
					ui.WithFont(
						ui.FontFamily(local),
						ui.Column(
							ui.Text(fmt.Sprintf("这段使用局部字体作用域: %s", local)),
							ui.Padding(
								ui.Insets{Top: 6},
								ui.Text(
									"局部文本字重覆盖为 Bold。",
									ui.TextFontWeight(ui.FontWeightBold),
								),
							),
						),
					),
				),
				ui.Padding(
					ui.Insets{Top: 12},
					ui.TextField(
						input.Value(),
						ui.InputOnChange(func(ctx *ui.Context, value string) {
							input.Set(value)
						}),
						ui.InputFontFamily(local),
					),
				),
			),
		)
	},
		ui.Title("FluxUI Fonts"),
		ui.Size(640, 420),
		ui.WithSystemFonts(true),
		ui.WithDefaultFont(global),
	)
}
