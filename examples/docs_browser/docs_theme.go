package main

import (
	"image/color"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

const defaultDocsThemeSeed = "flux"

type docsStringState interface {
	Value() string
	Set(string)
}

type docsBoolState interface {
	Value() bool
	Set(bool)
}

type docsThemeSeedOption struct {
	Label string
	Value string
	Color color.NRGBA
}

var docsThemeSeedOptions = []docsThemeSeedOption{
	{Label: "Flux", Value: "flux", Color: color.NRGBA{R: 0x38, G: 0x6a, B: 0xe8, A: 255}},
	{Label: "Mint", Value: "mint", Color: color.NRGBA{R: 0x00, G: 0x8f, B: 0x68, A: 255}},
	{Label: "Rose", Value: "rose", Color: color.NRGBA{R: 0xc4, G: 0x2b, B: 0x61, A: 255}},
	{Label: "Amber", Value: "amber", Color: color.NRGBA{R: 0xb7, G: 0x64, B: 0x00, A: 255}},
}

func docsBrowserTheme(seedName string, dark bool) *ui.Theme {
	seed := docsThemeSeed(defaultDocsThemeSeed)
	if selected := docsThemeSeed(seedName); selected.A != 0 {
		seed = selected
	}
	if dark {
		return ui.DarkThemeFromSeed(seed)
	}
	return ui.LightThemeFromSeed(seed)
}

func docsThemeControls(seed docsStringState, dark docsBoolState, th *ui.Theme) ui.Element {
	options := make([]ui.SelectOptionItem[string], 0, len(docsThemeSeedOptions))
	for _, item := range docsThemeSeedOptions {
		options = append(options, ui.SelectOptionItem[string]{
			Label: item.Label,
			Value: item.Value,
		})
	}

	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.SurfaceContainer).WithPad(ui.All(10)).WithRad(8),
		ui.ColumnElement(
			ui.TextElement("Theme", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.SelectElement[string](
				seed.Value(),
				options,
				ui.SelectOnChange[string](func(ctx *ui.Context, value string) {
					seed.Set(value)
				}),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.TextElement("Dark", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
				ui.HSpacerElement(8),
				ui.SwitchElement(
					dark.Value(),
					ui.SwitchOnChange(func(ctx *ui.Context, checked bool) {
						dark.Set(checked)
					}),
				),
			),
			ui.VSpacerElement(8),
			docsThemeSwatches(seed, th),
		),
	)
}

func docsThemeSwatches(seed docsStringState, th *ui.Theme) ui.Element {
	items := make([]ui.Element, 0, len(docsThemeSeedOptions))
	for _, item := range docsThemeSeedOptions {
		selected := item.Value == seed.Value()
		border := ui.Border{Width: 1, Color: th.Colors.OutlineVariant}
		if selected {
			border = ui.Border{Width: 2, Color: th.Colors.Primary}
		}
		value := item.Value
		items = append(items,
			ui.PaddingElement(
				ui.Insets{Right: 6},
				ui.TooltipElement(
					item.Label,
					ui.ButtonElement(
						ui.ContainerDecorationElement(
							ui.Bg(item.Color).WithPad(ui.All(0)).WithRad(7).WithBorder(border),
							ui.SpacerElement(22, 22),
						),
						ui.ButtonPadding(ui.All(0)),
						ui.ButtonBackground(color.NRGBA{}),
						ui.OnClick(func(ctx *ui.Context) {
							seed.Set(value)
						}),
					),
				),
			),
		)
	}
	return ui.RowElement(items...)
}

func docsThemeSeed(name string) color.NRGBA {
	for _, item := range docsThemeSeedOptions {
		if item.Value == name {
			return item.Color
		}
	}
	return color.NRGBA{}
}
