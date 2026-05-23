package main

import (
	"image/color"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	lightTheme := ui.NewTheme(ui.LightColors())
	darkTheme := ui.NewTheme(ui.DarkColors())

	redColors := ui.ColorScheme{
		Primary:      color.NRGBA{R: 220, G: 53, B: 69, A: 255},
		OnPrimary:    color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Secondary:    color.NRGBA{R: 240, G: 128, B: 128, A: 255},
		OnSecondary:  color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Surface:      color.NRGBA{R: 253, G: 246, B: 246, A: 255},
		OnSurface:    color.NRGBA{R: 114, G: 28, B: 36, A: 255},
		SurfaceMuted: color.NRGBA{R: 248, G: 215, B: 218, A: 255},
		Background:   color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		OnBackground: color.NRGBA{R: 114, G: 28, B: 36, A: 255},
		Error:        color.NRGBA{R: 220, G: 53, B: 69, A: 255},
		OnError:      color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Success:      color.NRGBA{R: 40, G: 167, B: 69, A: 255},
		OnSuccess:    color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Warning:      color.NRGBA{R: 255, G: 193, B: 7, A: 255},
		OnWarning:    color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		Disabled:     color.NRGBA{R: 173, G: 181, B: 189, A: 255},
		Outline:      color.NRGBA{R: 248, G: 215, B: 218, A: 255},
	}
	redTheme := ui.NewTheme(redColors)

	greenColors := ui.ColorScheme{
		Primary:      color.NRGBA{R: 40, G: 167, B: 69, A: 255},
		OnPrimary:    color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Secondary:    color.NRGBA{R: 144, G: 238, B: 144, A: 255},
		OnSecondary:  color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		Surface:      color.NRGBA{R: 247, G: 253, B: 250, A: 255},
		OnSurface:    color.NRGBA{R: 21, G: 87, B: 36, A: 255},
		SurfaceMuted: color.NRGBA{R: 209, G: 231, B: 221, A: 255},
		Background:   color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		OnBackground: color.NRGBA{R: 21, G: 87, B: 36, A: 255},
		Error:        color.NRGBA{R: 220, G: 53, B: 69, A: 255},
		OnError:      color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Success:      color.NRGBA{R: 40, G: 167, B: 69, A: 255},
		OnSuccess:    color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Warning:      color.NRGBA{R: 255, G: 193, B: 7, A: 255},
		OnWarning:    color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		Disabled:     color.NRGBA{R: 173, G: 181, B: 189, A: 255},
		Outline:      color.NRGBA{R: 209, G: 231, B: 221, A: 255},
	}
	greenTheme := ui.NewTheme(greenColors)

	orangeColors := ui.ColorScheme{
		Primary:      color.NRGBA{R: 255, G: 152, B: 0, A: 255},
		OnPrimary:    color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Secondary:    color.NRGBA{R: 255, G: 183, B: 77, A: 255},
		OnSecondary:  color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		Surface:      color.NRGBA{R: 255, G: 248, B: 240, A: 255},
		OnSurface:    color.NRGBA{R: 83, G: 49, B: 0, A: 255},
		SurfaceMuted: color.NRGBA{R: 255, G: 224, B: 178, A: 255},
		Background:   color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		OnBackground: color.NRGBA{R: 83, G: 49, B: 0, A: 255},
		Error:        color.NRGBA{R: 220, G: 53, B: 69, A: 255},
		OnError:      color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Success:      color.NRGBA{R: 40, G: 167, B: 69, A: 255},
		OnSuccess:    color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Warning:      color.NRGBA{R: 255, G: 193, B: 7, A: 255},
		OnWarning:    color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		Disabled:     color.NRGBA{R: 173, G: 181, B: 189, A: 255},
		Outline:      color.NRGBA{R: 255, G: 224, B: 178, A: 255},
	}
	orangeTheme := ui.NewTheme(orangeColors)

	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		selectedTheme := ui.State[string](ctx)

		currentTheme := lightTheme
		switch selectedTheme.Value() {
		case "dark":
			currentTheme = darkTheme
		case "red":
			currentTheme = redTheme
		case "green":
			currentTheme = greenTheme
		case "orange":
			currentTheme = orangeTheme
		}

		cs := currentTheme.Colors

		themeButton := func(label, key string) ui.Widget {
			isSelected := selectedTheme.Value() == key
			bg := cs.SurfaceMuted
			tc := cs.OnSurface
			if isSelected {
				bg = cs.Primary
				tc = cs.OnPrimary
			}
			deco := ui.Bg(bg).WithPad(ui.Symmetric(6, 12)).WithRad(6)
			if isSelected {
				deco = deco.WithBorder(ui.Border{Width: 2, Color: cs.OnPrimary})
			} else {
				deco = deco.WithBorder(ui.Border{Width: 1, Color: cs.Outline})
			}
			return ui.Padding(
				ui.All(4),
				ui.Button(
					ui.Text(label, ui.TextColor(tc)),
					ui.ButtonDecoration(deco),
					ui.OnClick(func(ctx *ui.Context) {
						selectedTheme.Set(key)
					}),
				),
			)
		}

		colorSwatch := func(label string, c, onC color.NRGBA) ui.Widget {
			return ui.Padding(
				ui.All(4),
				ui.ContainerDecoration(
					ui.Bg(c).WithPad(ui.Symmetric(6, 8)).WithRad(6),
					ui.Text(label, ui.TextSize(10), ui.TextColor(onC)),
				),
			)
		}

		return ui.ContainerDecoration(
			ui.Bg(cs.Background).WithPad(ui.All(16)),
			ui.Column(
				ui.Padding(ui.All(8), ui.Text("FluxUI 主题系统", ui.TextSize(22), ui.TextColor(cs.OnBackground), ui.TextAlign(ui.AlignCenter))),
				ui.Padding(ui.All(2), ui.Text("当前: "+selectedTheme.Value(), ui.TextSize(13), ui.TextColor(cs.SurfaceMuted))),

				ui.Padding(ui.TopBottom(4), ui.Text("主题预设", ui.TextSize(15), ui.TextColor(cs.OnSurface))),
				ui.Row(themeButton("浅色", "light"), themeButton("深色", "dark")),
				ui.Row(themeButton("红色", "red"), themeButton("绿色", "green"), themeButton("橙色", "orange")),

				ui.Padding(ui.TopBottom(6), ui.Text("语义色板 ColorScheme", ui.TextSize(15), ui.TextColor(cs.OnSurface))),
				ui.Row(colorSwatch("Primary", cs.Primary, cs.OnPrimary), colorSwatch("Secondary", cs.Secondary, cs.OnSecondary), colorSwatch("Surface", cs.Surface, cs.OnSurface)),
				ui.Row(colorSwatch("Muted", cs.SurfaceMuted, cs.OnSurface), colorSwatch("Bg", cs.Background, cs.OnBackground), colorSwatch("Outline", cs.Outline, cs.OnBackground)),
				ui.Row(colorSwatch("Error", cs.Error, cs.OnError), colorSwatch("Success", cs.Success, cs.OnSuccess), colorSwatch("Warning", cs.Warning, cs.OnWarning)),
				ui.Row(colorSwatch("Disabled", cs.Disabled, cs.Surface)),

				ui.Padding(ui.TopBottom(6), ui.Text("组件效果 (跟随主题色)", ui.TextSize(15), ui.TextColor(cs.OnSurface))),
				ui.Row(
					ui.Padding(ui.All(4), ui.Button(ui.Text("主按钮"))),
					ui.Padding(ui.All(4), ui.Button(ui.Text("禁用按钮"), ui.Disabled(true))),
				),
				ui.Checkbox("Checkbox 跟随 Primary", true),
				ui.Padding(ui.All(4), ui.Switch(true)),
				ui.Padding(ui.All(2), ui.Text("Switch / Checkbox 自动使用主题 Primary 色", ui.TextSize(12), ui.TextColor(cs.SurfaceMuted))),

				ui.Padding(ui.TopBottom(6), ui.Text("语义色场景", ui.TextSize(15), ui.TextColor(cs.OnSurface))),
				ui.ContainerDecoration(ui.Bg(cs.Primary).WithPad(ui.All(12)).WithRad(10),
					ui.Text("Primary + OnPrimary: 主容器", ui.TextSize(14), ui.TextColor(cs.OnPrimary)),
				),
				ui.VSpacer(2),
				ui.ContainerDecoration(ui.Bg(cs.Success).WithPad(ui.All(10)).WithRad(8),
					ui.Text("成功状态: Colors.Success + OnSuccess", ui.TextSize(13), ui.TextColor(cs.OnSuccess)),
				),
				ui.VSpacer(2),
				ui.ContainerDecoration(ui.Bg(cs.Warning).WithPad(ui.All(10)).WithRad(8),
					ui.Text("警告状态: Colors.Warning + OnWarning", ui.TextSize(13), ui.TextColor(cs.OnWarning)),
				),
				ui.VSpacer(2),
				ui.ContainerDecoration(ui.Bg(cs.Error).WithPad(ui.All(10)).WithRad(8),
					ui.Text("错误状态: Colors.Error + OnError", ui.TextSize(13), ui.TextColor(cs.OnError)),
				),

				ui.Padding(ui.TopBottom(6), ui.Text("向后兼容 (扁平字段)", ui.TextSize(15), ui.TextColor(cs.OnSurface))),
				ui.Row(
					ui.Padding(ui.All(4), ui.ContainerDecoration(ui.Bg(currentTheme.Primary).WithPad(ui.All(8)).WithRad(6),
						ui.Text("Primary", ui.TextSize(11), ui.TextColor(currentTheme.TextOnPrimary)),
					)),
					ui.Padding(ui.All(4), ui.ContainerDecoration(ui.Bg(currentTheme.Surface).WithPad(ui.All(8)).WithRad(6),
						ui.Text("Surface", ui.TextSize(11), ui.TextColor(currentTheme.TextColor)),
					)),
					ui.Padding(ui.All(4), ui.ContainerDecoration(ui.Bg(currentTheme.SurfaceMuted).WithPad(ui.All(8)).WithRad(6),
						ui.Text("Muted", ui.TextSize(11), ui.TextColor(currentTheme.TextColor)),
					)),
					ui.Padding(ui.All(4), ui.ContainerDecoration(ui.Bg(currentTheme.Disabled).WithPad(ui.All(8)).WithRad(6),
						ui.Text("Disabled", ui.TextSize(11)),
					)),
				),
				ui.Padding(ui.All(2), ui.Text("Theme.Primary / .TextColor 等旧字段仍然可用", ui.TextSize(12), ui.TextColor(cs.SurfaceMuted))),

				ui.Padding(ui.All(4), ui.Text("v0.1.0 主题系统", ui.TextSize(12), ui.TextColor(cs.SurfaceMuted))),
			),
		)
	}, ui.Title("FluxUI 主题系统"), ui.Size(560, 820))
}
