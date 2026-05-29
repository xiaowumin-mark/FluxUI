package main

import (
	"fmt"
	"image/color"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.RunElement(App, ui.Title("FluxUI Material 3 Showcase"), ui.Size(860, 980))
}

func App(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	dark := ui.NewTheme(ui.DarkColors())

	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.Surface).WithPad(ui.All(18)),
		ui.ScrollViewElement(
			ui.ColumnElement(
				sectionTitle("Material 3 Tokens"),
				ui.RowElement(
					tokenPanel("Light"),
					ui.SpacerElement(14, 0),
					ui.ThemeProviderElement(dark, tokenPanel("Dark")),
				),
				gap(),

				sectionTitle("Buttons"),
				ui.RowElement(
					padded(ui.FilledButtonElement(ui.TextElement("Filled"))),
					padded(ui.FilledTonalButtonElement(ui.TextElement("Filled tonal"))),
					padded(ui.OutlinedButtonElement(ui.TextElement("Outlined"))),
					padded(ui.TextButtonElement(ui.TextElement("Text"))),
					padded(ui.ElevatedButtonElement(ui.TextElement("Elevated"))),
					padded(ui.FilledButtonElement(ui.TextElement("Disabled"), ui.Disabled(true))),
				),
				gap(),

				sectionTitle("Text Fields"),
				ui.RowElement(
					ui.FixedWidthElement(260, ui.OutlinedTextFieldElement("Outlined value", ui.InputPlaceholder("Outlined"))),
					ui.SpacerElement(14, 0),
					ui.FixedWidthElement(260, ui.FilledTextFieldElement("Filled value", ui.InputPlaceholder("Filled"))),
					ui.SpacerElement(14, 0),
					ui.FixedWidthElement(220, ui.OutlinedTextFieldElement("", ui.InputPlaceholder("Disabled"), ui.InputDisabled(true))),
				),
				gap(),

				sectionTitle("Cards"),
				ui.RowElement(
					cardSample("Filled", ui.FilledCardElement),
					ui.SpacerElement(14, 0),
					cardSample("Elevated", ui.ElevatedCardElement),
					ui.SpacerElement(14, 0),
					cardSample("Outlined", ui.OutlinedCardElement),
				),
				gap(),

				sectionTitle("Selection Controls"),
				ui.RowElement(
					ui.CheckboxElement("Checkbox", true),
					ui.SpacerElement(24, 0),
					ui.SwitchElement(true),
					ui.SpacerElement(24, 0),
					ui.FixedWidthElement(220, ui.SliderElement(62, ui.SliderMin(0), ui.SliderMax(100))),
					ui.SpacerElement(24, 0),
					ui.RadioGroupElement("b", []ui.RadioItem{
						{Label: "A", Value: "a"},
						{Label: "B", Value: "b"},
					}, ui.RadioGroupDirection(ui.Horizontal)),
				),
				gap(),

				sectionTitle("Navigation"),
				ui.AppBarElementWithSlots(
					ui.TextElement("Top App Bar", ui.TextSize(th.Types.TitleLarge.Size), ui.TextColor(th.Colors.OnSurface)),
					ui.TextElement("Menu", ui.TextSize(14), ui.TextColor(th.Colors.OnSurface)),
					[]ui.Element{ui.TextButtonElement(ui.TextElement("Action"))},
				),
				ui.SpacerElement(0, 8),
				ui.TabsElement("overview", []ui.TabItem{
					{Key: "overview", Label: "Overview"},
					{Key: "components", Label: "Components"},
					{Key: "tokens", Label: "Tokens"},
				}),
				ui.SpacerElement(0, 8),
				ui.BottomNavigationElement("home", []ui.ElementNavItem{
					{Key: "home", Label: "Home", Icon: ui.TextElement("Home")},
					{Key: "search", Label: "Search", Icon: ui.TextElement("Search")},
					{Key: "settings", Label: "Settings", Icon: ui.TextElement("Settings")},
				}),
				gap(),

				sectionTitle("Overlays"),
				ui.FixedHeightElement(220, ui.StackElement(
					ui.ContainerDecorationElement(
						ui.Bg(th.Colors.SurfaceContainerLow).WithRad(th.Shapes.Medium).WithPad(ui.All(16)),
						ui.ColumnElement(
							ui.ToastElement("Inverse snackbar surface", ui.ToastDuration(0)),
							ui.PopupElement(true, ui.TextElement("Popup uses SurfaceContainer + elevation", ui.TextColor(th.Colors.OnSurface)), ui.PopupPadding(ui.All(16)), ui.PopupWidth(340)),
						),
					),
					ui.DialogElement(true,
						ui.TextElement("Dialog uses SurfaceContainerHigh, ExtraLarge shape and level 3 elevation.", ui.TextColor(th.Colors.OnSurface)),
						ui.DialogTitle("Dialog"),
						ui.DialogWidth(420),
					),
				)),
			),
		),
	)
}

func sectionTitle(label string) ui.Element {
	return ui.PaddingElement(
		ui.Insets{Top: 12, Bottom: 8},
		ui.TextElement(label, ui.TextSize(18)),
	)
}

func gap() ui.Element {
	return ui.SpacerElement(0, 12)
}

func padded(el ui.Element) ui.Element {
	return ui.PaddingElement(ui.Insets{Right: 8, Bottom: 8}, el)
}

func tokenPanel(name string) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		return ui.ContainerDecorationElement(
			ui.Bg(th.Colors.SurfaceContainer).WithPad(ui.All(14)).WithRad(th.Shapes.Medium),
			ui.ColumnElement(
				ui.TextElement(name, ui.TextSize(th.Types.TitleMedium.Size), ui.TextColor(th.Colors.OnSurface)),
				ui.SpacerElement(0, 8),
				colorRow("Primary", th.Colors.Primary, th.Colors.OnPrimary),
				colorRow("SecondaryContainer", th.Colors.SecondaryContainer, th.Colors.OnSecondaryContainer),
				colorRow("SurfaceContainer", th.Colors.SurfaceContainer, th.Colors.OnSurface),
				colorRow("ErrorContainer", th.Colors.ErrorContainer, th.Colors.OnErrorContainer),
				ui.SpacerElement(0, 8),
				ui.TextElement(fmt.Sprintf("Shape: xs %.0f / md %.0f / xl %.0f", th.Shapes.ExtraSmall, th.Shapes.Medium, th.Shapes.ExtraLarge), ui.TextSize(th.Types.BodySmall.Size), ui.TextColor(th.Colors.OnSurfaceVariant)),
			),
		)
	})
}

func colorRow(label string, bg, fg color.NRGBA) ui.Element {
	return ui.PaddingElement(
		ui.Insets{Bottom: 6},
		ui.ContainerDecorationElement(
			ui.Bg(bg).WithPad(ui.Symmetric(7, 10)).WithRad(6),
			ui.TextElement(label, ui.TextColor(fg), ui.TextSize(12)),
		),
	)
}

func cardSample(title string, factory func(ui.Element, ...ui.CardOption) ui.Element) ui.Element {
	return ui.FixedWidthElement(210, factory(
		ui.ColumnElement(
			ui.TextElement(title, ui.TextSize(16)),
			ui.SpacerElement(0, 6),
			ui.TextElement("MD3 card variant", ui.TextSize(13)),
		),
	))
}
