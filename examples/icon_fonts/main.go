package main

import (
	"fmt"
	"image/color"

	"github.com/xiaowumin-mark/FluxUI/icons/md3"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func App(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	count := ui.UseState(ctx, 0)

	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.Surface).WithPad(ui.All(20)),
		ui.ColumnElement(
			ui.TextElement("Icon Fonts", ui.TextType(th.Types.HeadlineSmall), ui.TextColor(th.Colors.OnSurface)),
			ui.SpacerElement(0, 8),
			ui.TextElement(
				fmt.Sprintf("registered icon fonts: %d, default: %s", len(ui.RegisteredIconFonts()), defaultIconFontLabel()),
				ui.TextType(th.Types.BodyMedium),
				ui.TextColor(th.Colors.OnSurfaceVariant),
			),
			ui.SpacerElement(0, 18),
			ui.RowElement(
				iconTile("Home", "home", th.Colors.Primary),
				ui.SpacerElement(12, 0),
				iconTile("Search", "search", th.Colors.Secondary),
				ui.SpacerElement(12, 0),
				iconTile("Settings", "settings", th.Colors.Tertiary),
			),
			ui.SpacerElement(0, 18),
			ui.RowElement(
				ui.FilledIconButtonElement(
					ui.IconElement("favorite", ui.IconUseFont(md3.ID)),
					ui.IconButtonSelected(true),
				),
				ui.SpacerElement(12, 0),
				ui.FilledTonalIconButtonElement(ui.IconElement("notifications")),
				ui.SpacerElement(12, 0),
				ui.OutlinedIconButtonElement(ui.IconElement("mail")),
			),
			ui.SpacerElement(0, 18),
			ui.ExtendedFloatingActionButtonElement(
				ui.IconElement("add"),
				ui.TextElement(fmt.Sprintf("Create %d", count.Value())),
				ui.FloatingActionButtonOnClick(func(ctx *ui.Context) {
					count.Set(count.Value() + 1)
				}),
			),
		),
	)
}

func iconTile(label, name string, col color.NRGBA) ui.Element {
	return ui.FilledCardElement(
		ui.ContainerDecorationElement(
			ui.Pad(ui.All(16)),
			ui.ColumnElement(
				ui.IconElement(name, ui.IconSize(32), ui.IconColor(col)),
				ui.SpacerElement(0, 8),
				ui.TextElement(label, ui.TextSize(13)),
			),
		),
	)
}

func defaultIconFontLabel() string {
	font, ok := ui.DefaultIconFont()
	if !ok {
		return "(none)"
	}
	return font.ID + " / " + font.Family
}

func main() {
	_ = ui.RunElement(App,
		ui.Title("FluxUI Icon Fonts"),
		ui.Size(720, 480),
	)
}
