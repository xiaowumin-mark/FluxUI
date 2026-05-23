package main

import (
	"image/color"

	statepkg "github.com/xiaowumin-mark/FluxUI/state"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

var (
	lightTheme  = ui.NewTheme(ui.LightColors())
	darkTheme   = ui.NewTheme(ui.DarkColors())
	redTheme    = ui.NewTheme(redColors())
	greenTheme  = ui.NewTheme(greenColors())
	orangeTheme = ui.NewTheme(orangeColors())
)

func redColors() ui.ColorScheme {
	return ui.ColorScheme{
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
}

func greenColors() ui.ColorScheme {
	return ui.ColorScheme{
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
}

func orangeColors() ui.ColorScheme {
	return ui.ColorScheme{
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
}

func resolveTheme(key string) *ui.Theme {
	switch key {
	case "dark":
		return darkTheme
	case "red":
		return redTheme
	case "green":
		return greenTheme
	case "orange":
		return orangeTheme
	default:
		return lightTheme
	}
}

func App(ctx *ui.Context) ui.Element {
	selectedTheme := ui.UseState(ctx, "light")
	currentTheme := resolveTheme(selectedTheme.Value())
	cs := currentTheme.Colors

	return ui.ContainerDecorationElement(
		ui.Bg(cs.Background).WithPad(ui.All(16)),
		ui.ColumnElement(
			ui.PaddingElement(ui.All(8), ui.TextElement("FluxUI 主题系统", ui.TextSize(22), ui.TextColor(cs.OnBackground), ui.TextAlign(ui.AlignCenter))),
			ui.PaddingElement(ui.All(2), ui.TextElement("当前: "+selectedTheme.Value(), ui.TextSize(13), ui.TextColor(cs.SurfaceMuted))),

			ui.PaddingElement(ui.TopBottom(4), ui.TextElement("主题预设", ui.TextSize(15), ui.TextColor(cs.OnSurface))),
			ui.RowElement(themeButton("浅色", "light", selectedTheme, cs), themeButton("深色", "dark", selectedTheme, cs)),
			ui.RowElement(themeButton("红色", "red", selectedTheme, cs), themeButton("绿色", "green", selectedTheme, cs), themeButton("橙色", "orange", selectedTheme, cs)),

			ui.PaddingElement(ui.TopBottom(6), ui.TextElement("语义色板 ColorScheme", ui.TextSize(15), ui.TextColor(cs.OnSurface))),
			ui.RowElement(colorSwatch("Primary", cs.Primary, cs.OnPrimary), colorSwatch("Secondary", cs.Secondary, cs.OnSecondary), colorSwatch("Surface", cs.Surface, cs.OnSurface)),
			ui.RowElement(colorSwatch("Muted", cs.SurfaceMuted, cs.OnSurface), colorSwatch("Bg", cs.Background, cs.OnBackground), colorSwatch("Outline", cs.Outline, cs.OnBackground)),
			ui.RowElement(colorSwatch("Error", cs.Error, cs.OnError), colorSwatch("Success", cs.Success, cs.OnSuccess), colorSwatch("Warning", cs.Warning, cs.OnWarning)),
			ui.RowElement(colorSwatch("Disabled", cs.Disabled, cs.Surface)),

			ui.PaddingElement(ui.TopBottom(6), ui.TextElement("组件效果 (跟随主题色)", ui.TextSize(15), ui.TextColor(cs.OnSurface))),
			ui.RowElement(
				ui.PaddingElement(ui.All(4), ui.ButtonElement(ui.TextElement("主按钮"))),
				ui.PaddingElement(ui.All(4), ui.ButtonElement(ui.TextElement("禁用按钮"), ui.Disabled(true))),
			),
			ui.CheckboxElement("Checkbox 跟随 Primary", true),
			ui.PaddingElement(ui.All(4), ui.SwitchElement(true)),
			ui.PaddingElement(ui.All(2), ui.TextElement("Switch / Checkbox 自动使用主题 Primary 色", ui.TextSize(12), ui.TextColor(cs.SurfaceMuted))),

			ui.PaddingElement(ui.TopBottom(6), ui.TextElement("语义色场景", ui.TextSize(15), ui.TextColor(cs.OnSurface))),
			ui.ContainerDecorationElement(ui.Bg(cs.Primary).WithPad(ui.All(12)).WithRad(10),
				ui.TextElement("Primary + OnPrimary: 主容器", ui.TextSize(14), ui.TextColor(cs.OnPrimary)),
			),
			ui.VSpacerElement(2),
			ui.ContainerDecorationElement(ui.Bg(cs.Success).WithPad(ui.All(10)).WithRad(8),
				ui.TextElement("成功状态: Colors.Success + OnSuccess", ui.TextSize(13), ui.TextColor(cs.OnSuccess)),
			),
			ui.VSpacerElement(2),
			ui.ContainerDecorationElement(ui.Bg(cs.Warning).WithPad(ui.All(10)).WithRad(8),
				ui.TextElement("警告状态: Colors.Warning + OnWarning", ui.TextSize(13), ui.TextColor(cs.OnWarning)),
			),
			ui.VSpacerElement(2),
			ui.ContainerDecorationElement(ui.Bg(cs.Error).WithPad(ui.All(10)).WithRad(8),
				ui.TextElement("错误状态: Colors.Error + OnError", ui.TextSize(13), ui.TextColor(cs.OnError)),
			),

			ui.PaddingElement(ui.TopBottom(6), ui.TextElement("向后兼容 (扁平字段)", ui.TextSize(15), ui.TextColor(cs.OnSurface))),
			ui.RowElement(
				ui.PaddingElement(ui.All(4), ui.ContainerDecorationElement(ui.Bg(currentTheme.Primary).WithPad(ui.All(8)).WithRad(6),
					ui.TextElement("Primary", ui.TextSize(11), ui.TextColor(currentTheme.TextOnPrimary)),
				)),
				ui.PaddingElement(ui.All(4), ui.ContainerDecorationElement(ui.Bg(currentTheme.Surface).WithPad(ui.All(8)).WithRad(6),
					ui.TextElement("Surface", ui.TextSize(11), ui.TextColor(currentTheme.TextColor)),
				)),
				ui.PaddingElement(ui.All(4), ui.ContainerDecorationElement(ui.Bg(currentTheme.SurfaceMuted).WithPad(ui.All(8)).WithRad(6),
					ui.TextElement("Muted", ui.TextSize(11), ui.TextColor(currentTheme.TextColor)),
				)),
				ui.PaddingElement(ui.All(4), ui.ContainerDecorationElement(ui.Bg(currentTheme.Disabled).WithPad(ui.All(8)).WithRad(6),
					ui.TextElement("Disabled", ui.TextSize(11)),
				)),
			),
			ui.PaddingElement(ui.All(2), ui.TextElement("Theme.Primary / .TextColor 等旧字段仍然可用", ui.TextSize(12), ui.TextColor(cs.SurfaceMuted))),

			ui.PaddingElement(ui.All(4), ui.TextElement("v0.1.0 主题系统", ui.TextSize(12), ui.TextColor(cs.SurfaceMuted))),
		),
	)
}

func themeButton(label, key string, selected *statepkg.State[string], cs ui.ColorScheme) ui.Element {
	isSelected := selected.Value() == key
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
	return ui.PaddingElement(
		ui.All(4),
		ui.ButtonElement(
			ui.TextElement(label, ui.TextColor(tc)),
			ui.ButtonDecoration(deco),
			ui.OnClick(func(ctx *ui.Context) {
				selected.Set(key)
			}),
		),
	)
}

func colorSwatch(label string, c, onC color.NRGBA) ui.Element {
	return ui.PaddingElement(
		ui.All(4),
		ui.ContainerDecorationElement(
			ui.Bg(c).WithPad(ui.Symmetric(6, 8)).WithRad(6),
			ui.TextElement(label, ui.TextSize(10), ui.TextColor(onC)),
		),
	)
}

func main() {
	_ = ui.RunElement(App, ui.Title("FluxUI 主题系统"), ui.Size(560, 820))
}
