package main

import (
	"image/color"

	theme "github.com/xiaowumin-mark/FluxUI/theme"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	darkTheme := &theme.Theme{
		Primary:       color.NRGBA{R: 66, G: 133, B: 244, A: 255},
		Surface:       color.NRGBA{R: 32, G: 33, B: 36, A: 255},
		SurfaceMuted:  color.NRGBA{R: 48, G: 49, B: 52, A: 255},
		TextColor:     color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		TextOnPrimary: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Disabled:      color.NRGBA{R: 97, G: 97, B: 97, A: 255},
		TextSize:      16,
	}

	redTheme := &theme.Theme{
		Primary:       color.NRGBA{R: 220, G: 53, B: 69, A: 255},
		Surface:       color.NRGBA{R: 253, G: 246, B: 246, A: 255},
		SurfaceMuted:  color.NRGBA{R: 248, G: 215, B: 218, A: 255},
		TextColor:     color.NRGBA{R: 114, G: 28, B: 36, A: 255},
		TextOnPrimary: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Disabled:      color.NRGBA{R: 173, G: 181, B: 189, A: 255},
		TextSize:      16,
	}

	greenTheme := &theme.Theme{
		Primary:       color.NRGBA{R: 40, G: 167, B: 69, A: 255},
		Surface:       color.NRGBA{R: 247, G: 253, B: 250, A: 255},
		SurfaceMuted:  color.NRGBA{R: 209, G: 231, B: 221, A: 255},
		TextColor:     color.NRGBA{R: 21, G: 87, B: 36, A: 255},
		TextOnPrimary: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Disabled:      color.NRGBA{R: 173, G: 181, B: 189, A: 255},
		TextSize:      16,
	}

	orangeTheme := &theme.Theme{
		Primary:       color.NRGBA{R: 255, G: 193, B: 7, A: 255},
		Surface:       color.NRGBA{R: 255, G: 243, B: 224, A: 255},
		SurfaceMuted:  color.NRGBA{R: 255, G: 238, B: 199, A: 255},
		TextColor:     color.NRGBA{R: 133, G: 100, B: 4, A: 255},
		TextOnPrimary: color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		Disabled:      color.NRGBA{R: 173, G: 181, B: 189, A: 255},
		TextSize:      16,
	}

	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)
		currentTheme := ui.State[string](ctx)

		var themeToApply *theme.Theme
		switch currentTheme.Value() {
		case "dark":
			themeToApply = darkTheme
		case "red":
			themeToApply = redTheme
		case "green":
			themeToApply = greenTheme
		case "orange":
			themeToApply = orangeTheme
		default:
			themeToApply = th
		}

		return ui.Container(
			ui.Style{
				Background: themeToApply.Surface,
				Padding:    ui.All(20),
			},
			ui.Column(
				ui.Padding(
					ui.All(8),
					ui.Text("主题示例", ui.TextSize(24), ui.TextAlign(ui.AlignCenter)),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("当前主题: "+currentTheme.Value(), ui.TextSize(16), ui.TextColor(themeToApply.TextColor)),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("点击下方按钮切换主题", ui.TextSize(14), ui.TextColor(themeToApply.SurfaceMuted)),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("默认主题", ui.TextSize(16), ui.TextColor(themeToApply.TextColor)),
				),
				ui.Container(
					ui.Style{
						Background: themeToApply.Primary,
						Padding:    ui.All(20),
						Radius:     12,
					},
					ui.Column(
						ui.Padding(ui.All(4), ui.Text("Primary", ui.TextColor(themeToApply.TextOnPrimary))),
						ui.Padding(ui.All(4), ui.Text("这是主要颜色", ui.TextSize(12), ui.TextColor(themeToApply.TextOnPrimary))),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("选择主题", ui.TextSize(16), ui.TextColor(themeToApply.TextColor)),
				),
				ui.Row(
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("默认"),
							ui.ButtonBackground(th.Primary),
							ui.OnClick(func(ctx *ui.Context) {
								currentTheme.Set("default")
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("深色"),
							ui.ButtonBackground(darkTheme.Primary),
							ui.ButtonForeground(darkTheme.TextOnPrimary),
							ui.OnClick(func(ctx *ui.Context) {
								currentTheme.Set("dark")
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("红色"),
							ui.ButtonBackground(redTheme.Primary),
							ui.ButtonForeground(redTheme.TextOnPrimary),
							ui.OnClick(func(ctx *ui.Context) {
								currentTheme.Set("red")
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("绿色"),
							ui.ButtonBackground(greenTheme.Primary),
							ui.ButtonForeground(greenTheme.TextOnPrimary),
							ui.OnClick(func(ctx *ui.Context) {
								currentTheme.Set("green")
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("橙色"),
							ui.ButtonBackground(orangeTheme.Primary),
							ui.ButtonForeground(orangeTheme.TextOnPrimary),
							ui.OnClick(func(ctx *ui.Context) {
								currentTheme.Set("orange")
							}),
						),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("主题色展示面板", ui.TextSize(16), ui.TextColor(themeToApply.TextColor)),
				),
				ui.Column(
					ui.Row(
						ui.Padding(ui.All(4), ui.Container(
							ui.Style{Background: themeToApply.Primary, Padding: ui.All(12), Radius: 6},
							ui.Text("Primary", ui.TextColor(themeToApply.TextOnPrimary), ui.TextSize(12)),
						)),
						ui.Padding(ui.All(4), ui.Container(
							ui.Style{Background: themeToApply.Surface, Padding: ui.All(12), Radius: 6},
							ui.Text("Surface", ui.TextColor(themeToApply.TextColor), ui.TextSize(12)),
						)),
						ui.Padding(ui.All(4), ui.Container(
							ui.Style{Background: themeToApply.SurfaceMuted, Padding: ui.All(12), Radius: 6},
							ui.Text("Muted", ui.TextColor(themeToApply.TextColor), ui.TextSize(12)),
						)),
					),
					ui.Row(
						ui.Padding(ui.All(4), ui.Container(
							ui.Style{Background: themeToApply.TextColor, Padding: ui.All(12), Radius: 6},
							ui.Text("Text", ui.TextColor(themeToApply.Surface), ui.TextSize(12)),
						)),
						ui.Padding(ui.All(4), ui.Container(
							ui.Style{Background: themeToApply.Disabled, Padding: ui.All(12), Radius: 6},
							ui.Text("Disabled", ui.TextColor(themeToApply.Surface), ui.TextSize(12)),
						)),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Button(
						ui.Text("禁用按钮 (使用当前主题)"),
						ui.Disabled(true),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("动态主题切换示例完成", ui.TextSize(14), ui.TextColor(themeToApply.SurfaceMuted)),
				),
			),
		)
	}, ui.Title("主题示例"), ui.Size(520, 680))
}
