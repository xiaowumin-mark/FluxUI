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
	selectValue := ui.UseState(ctx, "medium")
	menuOpen := ui.UseState(ctx, false)
	menuValue := ui.UseState(ctx, "copy")
	fabCount := ui.UseState(ctx, 0)
	railValue := ui.UseState(ctx, "home")
	drawerValue := ui.UseState(ctx, "inbox")
	chipSelected := ui.UseState(ctx, true)
	searchValue := ui.UseState(ctx, "")
	progressValue := ui.UseState(ctx, float32(56))
	snackbarMessage := ui.UseState(ctx, "")
	snackbarSerial := ui.UseState(ctx, 0)
	snackbarActionCount := ui.UseState(ctx, 0)

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

				sectionTitle("Dynamic Color"),
				ui.RowElement(
					seedThemePanel("Blue seed", color.NRGBA{R: 0, G: 87, B: 217, A: 255}, false),
					ui.SpacerElement(14, 0),
					seedThemePanel("Green seed", color.NRGBA{R: 27, G: 128, B: 77, A: 255}, false),
					ui.SpacerElement(14, 0),
					seedThemePanel("Orange dark", color.NRGBA{R: 224, G: 110, B: 0, A: 255}, true),
					ui.SpacerElement(14, 0),
					seedThemePanel("Purple dark", color.NRGBA{R: 103, G: 80, B: 164, A: 255}, true),
				),
				gap(),

				sectionTitle("Type Scale"),
				typeScalePanel(),
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

				sectionTitle("Menus and List Items"),
				ui.RowElement(
					ui.FixedWidthElement(230, ui.SelectElement(
						selectValue.Value(),
						[]ui.SelectOptionItem[string]{
							{Label: "Low priority", Value: "low"},
							{Label: "Medium priority", Value: "medium"},
							{Label: "High priority", Value: "high"},
						},
						ui.SelectOnChange[string](func(ctx *ui.Context, value string) { selectValue.Set(value) }),
					)),
					ui.SpacerElement(14, 0),
					ui.FixedWidthElement(230, ui.DropdownMenuElement(
						menuOpen.Value(),
						ui.ContainerDecorationElement(
							ui.Bg(th.Colors.Surface).WithPad(ui.Symmetric(10, 16)).WithRad(th.Shapes.ExtraSmall).WithBorder(ui.Border{Width: 1, Color: th.Colors.Outline}),
							ui.TextElement("Menu", ui.TextColor(th.Colors.OnSurface)),
						),
						[]ui.MenuItem{
							{Key: "copy", Label: "Copy"},
							{Key: "share", Label: "Share"},
							{Key: "archive", Label: "Archive"},
						},
						ui.DropdownMenuSelectedKey(menuValue.Value()),
						ui.DropdownMenuOnOpenChange(func(ctx *ui.Context, open bool) { menuOpen.Set(open) }),
						ui.DropdownMenuOnSelect(func(ctx *ui.Context, key string) {
							menuValue.Set(key)
							menuOpen.Set(false)
						}),
					)),
					ui.SpacerElement(14, 0),
					ui.FixedWidthElement(260, ui.ColumnElement(
						ui.ListItemElementWithSlots(ui.TextElement("Inbox"), ui.TextElement("12 unread messages"), ui.IconElement("I"), ui.TextElement("12"), ui.ListItemSelected(true)),
						ui.ListItemElementWithSlots(ui.TextElement("Archive"), ui.TextElement("Older conversations"), ui.IconElement("A"), nil),
					)),
				),
				gap(),

				sectionTitle("Icon Buttons and FABs"),
				ui.RowElement(
					padded(ui.IconButtonElement(ui.IconElement("S"), ui.IconButtonSelected(true))),
					padded(ui.FilledIconButtonElement(ui.IconElement("F"), ui.IconButtonSelected(true))),
					padded(ui.FilledTonalIconButtonElement(ui.IconElement("T"))),
					padded(ui.OutlinedIconButtonElement(ui.IconElement("O"))),
					ui.SpacerElement(16, 0),
					padded(ui.SmallFloatingActionButtonElement(ui.IconElement("+"), ui.FloatingActionButtonOnClick(func(ctx *ui.Context) { fabCount.Set(fabCount.Value() + 1) }))),
					padded(ui.FloatingActionButtonElement(ui.IconElement("+"), ui.FloatingActionButtonOnClick(func(ctx *ui.Context) { fabCount.Set(fabCount.Value() + 1) }))),
					padded(ui.ExtendedFloatingActionButtonElement(ui.IconElement("+"), ui.TextElement(fmt.Sprintf("Create %d", fabCount.Value())), ui.FloatingActionButtonOnClick(func(ctx *ui.Context) { fabCount.Set(fabCount.Value() + 1) }))),
				),
				gap(),

				sectionTitle("Chips and Badges"),
				ui.RowElement(
					padded(ui.AssistChipElement("Assist", ui.ChipLeading(ui.Icon("i", ui.IconSize(16))))),
					padded(ui.FilterChipElement(
						"Filter",
						ui.ChipSelected(chipSelected.Value()),
						ui.ChipOnClick(func(ctx *ui.Context) { chipSelected.Set(!chipSelected.Value()) }),
					)),
					padded(ui.InputChipElement("Input", ui.ChipTrailing(ui.Icon("x", ui.IconSize(14))))),
					padded(ui.SuggestionChipElement("Suggestion")),
					ui.SpacerElement(18, 0),
					padded(ui.BadgeElement(ui.IconButtonElement(ui.IconElement("M")), "3")),
					padded(ui.BadgeElement(ui.IconButtonElement(ui.IconElement("N")), "", ui.BadgeVisible(true))),
				),
				gap(),

				sectionTitle("Search and Progress"),
				ui.RowElement(
					ui.FixedWidthElement(
						340,
						ui.SearchBarElement(
							searchValue.Value(),
							ui.SearchBarPlaceholder("Search components"),
							ui.SearchBarLeading(ui.Icon("S", ui.IconSize(18))),
							ui.SearchBarOnChange(func(ctx *ui.Context, value string) { searchValue.Set(value) }),
						),
					),
					ui.SpacerElement(18, 0),
					ui.FixedWidthElement(
						260,
						ui.ColumnElement(
							ui.SliderElement(
								progressValue.Value(),
								ui.SliderMin(0),
								ui.SliderMax(100),
								ui.SliderOnChange(func(ctx *ui.Context, value float32) { progressValue.Set(value) }),
							),
							ui.PaddingElement(ui.Insets{Top: 10}, ui.LinearProgressIndicatorElement(progressValue.Value())),
						),
					),
					ui.SpacerElement(18, 0),
					ui.CircularProgressIndicatorElement(progressValue.Value(), ui.ProgressSize(72), ui.ProgressLabelVisible(true)),
				),
				gap(),

				sectionTitle("Navigation"),
				ui.AppBarElementWithSlots(
					ui.TextElement("Top App Bar", ui.TextType(th.Types.TitleLarge), ui.TextColor(th.Colors.OnSurface)),
					ui.TextElement("Menu", ui.TextType(th.Types.LabelLarge), ui.TextColor(th.Colors.OnSurface)),
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

				sectionTitle("Navigation Rail and Drawer"),
				ui.FixedHeightElement(260, ui.RowElement(
					ui.NavigationRailElement(
						railValue.Value(),
						[]ui.ElementNavItem{
							{Key: "home", Label: "Home", Icon: ui.IconElement("H")},
							{Key: "search", Label: "Search", Icon: ui.IconElement("S")},
							{Key: "settings", Label: "Settings", Icon: ui.IconElement("G")},
						},
						ui.NavigationRailOnChange(func(ctx *ui.Context, key string) { railValue.Set(key) }),
					),
					ui.SpacerElement(14, 0),
					ui.NavigationDrawerElement(
						drawerValue.Value(),
						[]ui.ElementNavItem{
							{Key: "inbox", Label: "Inbox", Icon: ui.IconElement("I")},
							{Key: "sent", Label: "Sent", Icon: ui.IconElement("S")},
							{Key: "drafts", Label: "Drafts", Icon: ui.IconElement("D")},
						},
						ui.NavigationDrawerWidth(280),
						ui.NavigationDrawerOnChange(func(ctx *ui.Context, key string) { drawerValue.Set(key) }),
					),
					ui.ExpandedElement(ui.CenterElement(ui.TextElement("Rail: "+railValue.Value()+" / Drawer: "+drawerValue.Value(), ui.TextType(th.Types.BodyMedium)))),
				)),
				gap(),

				sectionTitle("Snackbar and Tooltip"),
				ui.FixedHeightElement(180, ui.StackElement(
					ui.ContainerDecorationElement(
						ui.Bg(th.Colors.SurfaceContainerLow).WithRad(th.Shapes.Medium).WithPad(ui.All(16)),
						ui.RowElement(
							padded(ui.TooltipElement("Tooltip text", ui.FilledTonalButtonElement(ui.TextElement("Hover me")))),
							padded(ui.FilledButtonElement(
								ui.TextElement("Show snackbar"),
								ui.OnClick(func(ctx *ui.Context) {
									snackbarSerial.Set(snackbarSerial.Value() + 1)
									snackbarMessage.Set("Draft archived")
								}),
							)),
							ui.PaddingElement(
								ui.Insets{Top: 10},
								ui.TextElement(fmt.Sprintf("Undo clicks: %d", snackbarActionCount.Value()), ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
							),
						),
					),
					func() ui.Element {
						if snackbarMessage.Value() == "" {
							return nil
						}
						return ui.Key(
							fmt.Sprintf("showcase-snackbar-%d", snackbarSerial.Value()),
							ui.SnackbarElement(
								snackbarMessage.Value(),
								ui.SnackbarAction("Undo", func(ctx *ui.Context) {
									snackbarActionCount.Set(snackbarActionCount.Value() + 1)
									snackbarMessage.Set("")
								}),
								ui.ToastDuration(0),
							),
						)
					}(),
				)),
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
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		return ui.PaddingElement(
			ui.Insets{Top: 12, Bottom: 8},
			ui.TextElement(label, ui.TextType(th.Types.TitleMedium), ui.TextColor(th.Colors.OnSurface)),
		)
	})
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
				ui.TextElement(name, ui.TextType(th.Types.TitleMedium), ui.TextColor(th.Colors.OnSurface)),
				ui.SpacerElement(0, 8),
				colorRow("Primary", th.Colors.Primary, th.Colors.OnPrimary),
				colorRow("SecondaryContainer", th.Colors.SecondaryContainer, th.Colors.OnSecondaryContainer),
				colorRow("SurfaceContainer", th.Colors.SurfaceContainer, th.Colors.OnSurface),
				colorRow("ErrorContainer", th.Colors.ErrorContainer, th.Colors.OnErrorContainer),
				ui.SpacerElement(0, 8),
				ui.TextElement(fmt.Sprintf("Shape: xs %.0f / md %.0f / xl %.0f", th.Shapes.ExtraSmall, th.Shapes.Medium, th.Shapes.ExtraLarge), ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
			),
		)
	})
}

func seedThemePanel(name string, seed color.NRGBA, dark bool) ui.Element {
	th := ui.LightThemeFromSeed(seed)
	if dark {
		th = ui.DarkThemeFromSeed(seed)
	}
	return ui.FixedWidthElement(190, ui.ThemeProviderElement(th, dynamicColorPanel(name)))
}

func dynamicColorPanel(name string) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		return ui.ContainerDecorationElement(
			ui.Bg(th.Colors.SurfaceContainer).WithPad(ui.All(14)).WithRad(th.Shapes.Medium),
			ui.ColumnElement(
				ui.TextElement(name, ui.TextType(th.Types.TitleMedium), ui.TextColor(th.Colors.OnSurface)),
				ui.SpacerElement(0, 8),
				colorRow("Primary", th.Colors.Primary, th.Colors.OnPrimary),
				colorRow("Secondary", th.Colors.Secondary, th.Colors.OnSecondary),
				colorRow("TertiaryContainer", th.Colors.TertiaryContainer, th.Colors.OnTertiaryContainer),
				colorRow("Surface", th.Colors.Surface, th.Colors.OnSurface),
			),
		)
	})
}

func typeScalePanel() ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		return ui.ContainerDecorationElement(
			ui.Bg(th.Colors.SurfaceContainer).WithPad(ui.All(14)).WithRad(th.Shapes.Medium),
			ui.ColumnElement(
				typeSample("Display Large", th.Types.DisplayLarge, "Display"),
				typeSample("Headline Medium", th.Types.HeadlineMedium, "Headline"),
				typeSample("Title Large", th.Types.TitleLarge, "Title"),
				typeSample("Body Large", th.Types.BodyLarge, "Body text uses readable line height."),
				typeSample("Body Medium", th.Types.BodyMedium, "Body medium supports dense product screens."),
				typeSample("Label Large", th.Types.LabelLarge, "Label"),
				typeSample("Label Small", th.Types.LabelSmall, "Small label"),
			),
		)
	})
}

func typeSample(name string, textStyle ui.TextStyle, sample string) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		return ui.PaddingElement(
			ui.Insets{Bottom: 10},
			ui.RowElement(
				ui.FixedWidthElement(
					170,
					ui.TextElement(
						fmt.Sprintf("%s  %.0f/%.0f", name, textStyle.Size, textStyle.LineHeight),
						ui.TextType(th.Types.LabelMedium),
						ui.TextColor(th.Colors.OnSurfaceVariant),
					),
				),
				ui.ExpandedElement(ui.TextElement(sample, ui.TextType(textStyle), ui.TextColor(th.Colors.OnSurface))),
			),
		)
	})
}

func colorRow(label string, bg, fg color.NRGBA) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		return ui.PaddingElement(
			ui.Insets{Bottom: 6},
			ui.ContainerDecorationElement(
				ui.Bg(bg).WithPad(ui.Symmetric(7, 10)).WithRad(6),
				ui.TextElement(label, ui.TextColor(fg), ui.TextType(th.Types.LabelMedium)),
			),
		)
	})
}

func cardSample(title string, factory func(ui.Element, ...ui.CardOption) ui.Element) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		return ui.FixedWidthElement(210, factory(
			ui.ColumnElement(
				ui.TextElement(title, ui.TextType(th.Types.TitleMedium)),
				ui.SpacerElement(0, 6),
				ui.TextElement("MD3 card variant", ui.TextType(th.Types.BodyMedium)),
			),
		))
	})
}
