package main

import (
	"fmt"
	"image"
	"image/color"
	"strings"
	"time"

	fluxstate "github.com/xiaowumin-mark/FluxUI/state"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.RunElement(App, ui.Title("FluxUI Component Lab"), ui.Size(1180, 920), ui.MinSize(900, 720))
}

func App(ctx *ui.Context) ui.Element {
	dark := ui.UseState(ctx, false)
	compact := ui.UseState(ctx, false)
	palette := ui.UseState(ctx, "blue")

	buttonCount := ui.UseState(ctx, 0)
	hovered := ui.UseState(ctx, false)
	pressed := ui.UseState(ctx, false)
	textValue := ui.UseState(ctx, "FluxUI")
	passwordValue := ui.UseState(ctx, "secret")
	focusState := ui.UseState(ctx, "blurred")
	checked := ui.UseState(ctx, true)
	switched := ui.UseState(ctx, true)
	sliderValue := ui.UseState(ctx, float32(56))
	radioValue := ui.UseState(ctx, "stable")
	selectValue := ui.UseState(ctx, "medium")
	menuOpen := ui.UseState(ctx, false)
	menuValue := ui.UseState(ctx, "copy")
	tabValue := ui.UseState(ctx, "overview")
	iconSelected := ui.UseState(ctx, true)
	fabCount := ui.UseState(ctx, 0)
	searchValue := ui.UseState(ctx, "")
	bottomNavValue := ui.UseState(ctx, "home")
	railValue := ui.UseState(ctx, "home")
	drawerValue := ui.UseState(ctx, "inbox")
	dialogOpen := ui.UseState(ctx, false)
	popupOpen := ui.UseState(ctx, false)
	toastMessage := ui.UseState(ctx, "")
	toastSerial := ui.UseState(ctx, 0)
	snackbarMessage := ui.UseState(ctx, "")
	snackbarSerial := ui.UseState(ctx, 0)
	snackbarActions := ui.UseState(ctx, 0)
	chipSelected := ui.UseState(ctx, true)
	imageClicks := ui.UseState(ctx, 0)
	pressCount := ui.UseState(ctx, 0)
	decoClicks := ui.UseState(ctx, 0)
	decoHovering := ui.UseState(ctx, false)
	scrollLog := ui.UseState(ctx, "not scrolled")
	listReachCount := ui.UseState(ctx, 0)
	gridReachCount := ui.UseState(ctx, 0)
	dropActive := ui.UseState(ctx, false)
	dropLog := ui.UseState(ctx, []string{"Drop log ready"})
	routerAllowSettings := ui.UseState(ctx, true)
	routerUserID := ui.UseState(ctx, "u1002")
	routerLog := ui.UseState(ctx, "router ready")

	buttonRef := ui.UseRef(ctx, ui.NewButtonRef())
	inputRef := ui.UseRef(ctx, ui.NewInputRef())
	checkboxRef := ui.UseRef(ctx, ui.NewCheckboxRef())
	switchRef := ui.UseRef(ctx, ui.NewSwitchRef())
	sliderRef := ui.UseRef(ctx, ui.NewSliderRef())
	radioRef := ui.UseRef(ctx, ui.NewRadioGroupRef())
	selectRef := ui.UseRef(ctx, ui.NewSelectRef[string]())
	tabsRef := ui.UseRef(ctx, ui.NewTabsRef())
	dialogRef := ui.UseRef(ctx, ui.NewDialogRef())
	popupRef := ui.UseRef(ctx, ui.NewPopupRef())
	scrollRef := ui.UseRef(ctx, ui.NewScrollRef())
	bottomNavRef := ui.UseRef(ctx, ui.NewBottomNavRef())
	pressableRef := ui.UseRef(ctx, ui.NewPressableRef())

	appTheme := componentLabTheme(palette.Value(), dark.Value(), compact.Value())

	return ui.ThemeProviderElement(appTheme, ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		page := ui.ContainerDecorationElement(
			ui.Bg(th.Colors.Surface).WithPad(ui.All(18)),
			ui.ScrollViewElement(
				ui.ColumnElement(
					labHeader(th),
					themeControls(th, dark, compact, palette),
					section("Layout and style primitives", "Row, Column, Stack, Center, Padding, Spacer, Divider, Container, Decoration, sizing, font and nested theme coverage.", layoutStylePanel(th, decoClicks, decoHovering)),
					section("Buttons and generic interaction", "Button variants, Pressable, ClickArea compatibility and command refs.", buttonsPanel(th, buttonCount, hovered, pressed, buttonRef, pressableRef, pressCount)),
					section("Text input and selection", "Controlled fields, checkbox, switch, slider, radio, select, menu and dropdown states.", inputSelectionPanel(th, textValue, passwordValue, focusState, checked, switched, sliderValue, radioValue, selectValue, menuOpen, menuValue, inputRef, checkboxRef, switchRef, sliderRef, radioRef, selectRef)),
					section("Cards, media and chips", "Image, icon, card variants, icon buttons, FABs, badges, chips and search bar.", mediaCardsPanel(th, imageClicks, buttonCount, iconSelected, fabCount, chipSelected, searchValue)),
					section("Navigation and overlays", "AppBar, tabs, bottom navigation, rail, drawer, dialog, popup, toast, snackbar and tooltip.", navigationOverlayPanel(th, tabValue, bottomNavValue, railValue, drawerValue, dialogOpen, popupOpen, toastMessage, toastSerial, snackbarMessage, snackbarSerial, snackbarActions, tabsRef, dialogRef, popupRef, bottomNavRef)),
					section("Progress, scroll, list and grid", "Determinate indicators, ScrollView refs, ListView, Grid and GridView with reach-end callbacks.", progressCollectionPanel(th, sliderValue, scrollLog, listReachCount, gridReachCount, scrollRef)),
					section("Drag, drop and router", "DragSource, DropTarget and RouterElement remain in the same visual test surface.", dragDropRouterPanel(th, dropActive, dropLog, routerAllowSettings, routerUserID, routerLog)),
					ui.VSpacerElement(24),
				),
				ui.ScrollBarVisible(true),
			),
		)

		return ui.StackElement(
			page,
			dialogOverlay(th, dialogOpen, dialogRef),
			popupOverlay(th, popupOpen, popupRef),
			toastOverlay(toastMessage, toastSerial),
			snackbarOverlay(snackbarMessage, snackbarSerial, snackbarActions),
		)
	}))
}

func componentLabTheme(key string, dark bool, compact bool) *ui.Theme {
	var seed ui.ColorScheme
	switch key {
	case "green":
		if dark {
			seed = ui.DarkColorSchemeFromSeed(ui.NRGBA(22, 128, 92, 255))
		} else {
			seed = ui.LightColorSchemeFromSeed(ui.NRGBA(22, 128, 92, 255))
		}
	case "amber":
		if dark {
			seed = ui.DarkColorSchemeFromSeed(ui.NRGBA(220, 132, 0, 255))
		} else {
			seed = ui.LightColorSchemeFromSeed(ui.NRGBA(220, 132, 0, 255))
		}
	default:
		if dark {
			seed = ui.DarkColorSchemeFromSeed(ui.NRGBA(25, 103, 210, 255))
		} else {
			seed = ui.LightColorSchemeFromSeed(ui.NRGBA(25, 103, 210, 255))
		}
	}
	th := ui.NewTheme(seed)
	if compact {
		th.SetDensity(ui.CompactDensityScale())
	}
	return th
}

func labHeader(th *ui.Theme) ui.Element {
	return ui.ContainerDecorationElement(
		ui.LinearGrad(
			image.Point{X: 0, Y: 0},
			image.Point{X: 900, Y: 220},
			th.Colors.Primary,
			th.Colors.Tertiary,
		).WithPad(ui.All(18)).WithRad(8),
		ui.ColumnElement(
			ui.TextElement("FluxUI Component Lab", ui.TextType(th.Types.HeadlineMedium), ui.TextColor(th.Colors.OnPrimary)),
			ui.VSpacerElement(6),
			ui.TextElement("Single example for visual alignment, interaction checks and component debugging.", ui.TextType(th.Types.BodyMedium), ui.TextColor(th.Colors.OnPrimary)),
		),
	)
}

func themeControls(th *ui.Theme, dark *fluxstate.State[bool], compact *fluxstate.State[bool], palette *fluxstate.State[string]) ui.Element {
	paletteOptions := []ui.SelectOptionItem[string]{
		{Label: "Blue seed", Value: "blue"},
		{Label: "Green seed", Value: "green"},
		{Label: "Amber seed", Value: "amber"},
	}
	return ui.PaddingElement(
		ui.Insets{Top: 14, Bottom: 4},
		ui.ContainerDecorationElement(
			ui.Bg(th.Colors.SurfaceContainer).WithPad(ui.All(14)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
			ui.RowElement(
				ui.FixedWidthElement(
					180,
					ui.ColumnElement(
						ui.TextElement("Theme", ui.TextType(th.Types.TitleMedium), ui.TextColor(th.Colors.OnSurface)),
						ui.TextElement("global provider", ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
					),
				),
				ui.HSpacerElement(16),
				ui.FixedWidthElement(
					240,
					ui.SelectElement(
						palette.Value(),
						paletteOptions,
						ui.SelectOnChange[string](func(ctx *ui.Context, value string) { setIfChanged(palette, value) }),
					),
				),
				ui.HSpacerElement(20),
				ui.CheckboxElement("Dark mode", dark.Value(), ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) {
					setIfChanged(dark, checked)
				})),
				ui.HSpacerElement(20),
				ui.RowElement(
					ui.TextElement("Compact", ui.TextType(th.Types.BodyMedium), ui.TextColor(th.Colors.OnSurface)),
					ui.HSpacerElement(8),
					ui.SwitchElement(compact.Value(), ui.SwitchOnChange(func(ctx *ui.Context, checked bool) {
						setIfChanged(compact, checked)
					})),
				),
				ui.ExpandedElement(ui.SpacerElement(0, 0)),
				ui.FixedWidthElement(220, tokenStrip(th)),
			),
		),
	)
}

func tokenStrip(th *ui.Theme) ui.Element {
	return ui.RowElement(
		tokenChip("Primary", th.Colors.Primary, th.Colors.OnPrimary),
		ui.HSpacerElement(6),
		tokenChip("Surface", th.Colors.SurfaceContainerHigh, th.Colors.OnSurface),
	)
}

func tokenChip(label string, bg, fg color.NRGBA) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(bg).WithPad(ui.Symmetric(6, 8)).WithRad(999),
		ui.TextElement(label, ui.TextSize(11), ui.TextColor(fg)),
	)
}

func section(title, summary string, body ui.Element) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		return ui.PaddingElement(
			ui.Insets{Top: 14},
			ui.ContainerDecorationElement(
				ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.All(16)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
				ui.ColumnElement(
					ui.TextElement(title, ui.TextType(th.Types.TitleMedium), ui.TextColor(th.Colors.OnSurface)),
					ui.VSpacerElement(4),
					ui.TextElement(summary, ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
					ui.VSpacerElement(14),
					body,
				),
			),
		)
	})
}

func layoutStylePanel(th *ui.Theme, decoClicks *fluxstate.State[int], decoHovering *fluxstate.State[bool]) ui.Element {
	localDark := ui.NewTheme(ui.DarkColors())
	return ui.ColumnElement(
		ui.RowElement(
			ui.ExpandedElement(surfaceDemo(th, "ContainerDecoration", ui.ContainerDecorationElement(
				ui.Bg(th.Colors.PrimaryContainer).
					WithPad(ui.All(14)).
					WithRad(8).
					WithBorder(ui.Border{Width: 1, Color: th.Colors.Primary}).
					WithHover(ui.Bg(th.Colors.SecondaryContainer)).
					WithPressed(ui.Bg(th.Colors.TertiaryContainer)).
					Merge(ui.Elevation(1)),
				ui.TextElement("hover and click decoration", ui.TextColor(th.Colors.OnPrimaryContainer)),
				ui.OnDecoClick(func(ctx *ui.Context) { decoClicks.Set(decoClicks.Value() + 1) }),
				ui.OnDecoHover(func(ctx *ui.Context, value bool) { setIfChanged(decoHovering, value) }),
			))),
			ui.HSpacerElement(12),
			ui.ExpandedElement(surfaceDemo(th, "Stack and Center", ui.FixedHeightElement(120, ui.FillElement(ui.StackElement(
				ui.FillElement(ui.ContainerDecorationElement(ui.Bg(th.Colors.SurfaceContainerHigh).WithRad(8), ui.SpacerElement(0, 0))),
				ui.PaddingElement(ui.Insets{Left: 12, Top: 12}, ui.ContainerDecorationElement(ui.Bg(th.Colors.Secondary).WithPad(ui.Symmetric(5, 8)).WithRad(6), ui.TextElement("Layer", ui.TextColor(th.Colors.OnSecondary), ui.TextType(th.Types.LabelMedium)))),
				ui.CenterElement(ui.TextElement("Center", ui.TextType(th.Types.TitleMedium), ui.TextColor(th.Colors.OnSurface))),
			))))),
			ui.HSpacerElement(12),
			ui.ExpandedElement(surfaceDemo(th, "Transform", ui.RowElement(
				ui.ContainerDecorationElement(ui.Bg(th.Colors.Warning).WithPad(ui.All(10)).WithRad(8).Merge(ui.Rotate(-5)), ui.TextElement("Rotate", ui.TextColor(th.Colors.OnWarning))),
				ui.HSpacerElement(14),
				ui.ContainerDecorationElement(ui.Bg(th.Colors.Success).WithPad(ui.All(10)).WithRad(8).Merge(ui.ScaleDeco(1.08, 1.08)).Merge(ui.TranslateDeco(4, 2)), ui.TextElement("Scale", ui.TextColor(th.Colors.OnSuccess))),
			))),
		),
		ui.VSpacerElement(12),
		ui.RowElement(
			ui.ExpandedElement(surfaceDemo(th, "Row, Column, Spacer", ui.RowElement(
				pill(th, "A", th.Colors.Primary, th.Colors.OnPrimary),
				ui.HSpacerElement(8),
				pill(th, "B", th.Colors.Secondary, th.Colors.OnSecondary),
				ui.HSpacerElement(8),
				ui.ColumnElement(pill(th, "Top", th.Colors.Tertiary, th.Colors.OnTertiary), ui.VSpacerElement(6), pill(th, "Bottom", th.Colors.Error, th.Colors.OnError)),
			))),
			ui.HSpacerElement(12),
			ui.ExpandedElement(surfaceDemo(th, "Sizing", ui.ColumnElement(
				ui.RowElement(
					ui.FixedWidthElement(120, pill(th, "FixedWidth", th.Colors.Primary, th.Colors.OnPrimary)),
					ui.HSpacerElement(8),
					ui.ExpandedElement(pill(th, "Expanded", th.Colors.Tertiary, th.Colors.OnTertiary)),
				),
				ui.VSpacerElement(8),
				ui.FixedHeightElement(42, ui.FillWidthElement(pill(th, "FixedHeight + FillWidth", th.Colors.Secondary, th.Colors.OnSecondary))),
				ui.VSpacerElement(8),
				ui.RowElement(
					ui.FixedSizeElement(104, 42, ui.ContainerDecorationElement(ui.Bg(th.Colors.PrimaryContainer).WithRad(8), ui.CenterElement(ui.TextElement("FixedSize", ui.TextType(th.Types.LabelMedium), ui.TextColor(th.Colors.OnPrimaryContainer))))),
					ui.HSpacerElement(8),
					ui.FixedHeightElement(42, ui.FillHeightElement(ui.ContainerDecorationElement(ui.Bg(th.Colors.TertiaryContainer).WithPad(ui.LeftRight(10)).WithRad(8), ui.CenterElement(ui.TextElement("FillHeight", ui.TextType(th.Types.LabelMedium), ui.TextColor(th.Colors.OnTertiaryContainer)))))),
				),
				ui.VSpacerElement(8),
				ui.RowElement(
					ui.FlexedElement(1, pill(th, "Flex 1", th.Colors.Success, th.Colors.OnSuccess)),
					ui.HSpacerElement(8),
					ui.FlexedElement(2, pill(th, "Flex 2", th.Colors.Warning, th.Colors.OnWarning)),
				),
			))),
			ui.HSpacerElement(12),
			ui.ExpandedElement(surfaceDemo(th, "Divider and font", ui.ColumnElement(
				ui.TextElement("Top area", ui.TextColor(th.Colors.OnSurface)),
				ui.DividerElement(ui.DividerThickness(1), ui.DividerColor(th.Colors.OutlineVariant), ui.DividerMargin(ui.Insets{Top: 8, Bottom: 8})),
				ui.WithFontElement(ui.FontFamily("Georgia"), ui.TextElement("WithFontElement sample", ui.TextType(th.Types.BodyMedium), ui.TextColor(th.Colors.OnSurface))),
				ui.VSpacerElement(8),
				ui.ContainerElement(ui.Style{Background: th.Colors.SurfaceVariant, Padding: ui.All(8), Radius: 8}, ui.TextElement("ContainerElement compatibility", ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant))),
				ui.VSpacerElement(8),
				ui.ThemeProviderElement(localDark, ui.ContainerDecorationElement(ui.Bg(localDark.Colors.SurfaceContainer).WithPad(ui.All(10)).WithRad(8), ui.TextElement("Nested dark theme", ui.TextColor(localDark.Colors.OnSurface)))),
			))),
		),
		ui.VSpacerElement(10),
		ui.TextElement(fmt.Sprintf("decoration clicks = %d, hovering = %t", decoClicks.Value(), decoHovering.Value()), ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
	)
}

func buttonsPanel(th *ui.Theme, count *fluxstate.State[int], hovered *fluxstate.State[bool], pressed *fluxstate.State[bool], buttonRef *ui.Ref[*ui.ButtonRef], pressableRef *ui.Ref[*ui.PressableRef], pressCount *fluxstate.State[int]) ui.Element {
	click := ui.OnClick(func(ctx *ui.Context) { count.Set(count.Value() + 1) })
	return ui.ColumnElement(
		ui.RowElement(
			wrap(ui.FilledButtonElement(ui.TextElement("Filled"), click, ui.ButtonAttachRef(buttonRef.Current), ui.OnHover(func(ctx *ui.Context, value bool) { setIfChanged(hovered, value) }))),
			wrap(ui.FilledTonalButtonElement(ui.TextElement("Tonal"), click)),
			wrap(ui.OutlinedButtonElement(ui.TextElement("Outlined"), click)),
			wrap(ui.TextButtonElement(ui.TextElement("Text"), click)),
			wrap(ui.ElevatedButtonElement(ui.TextElement("Elevated"), click)),
			wrap(ui.FilledButtonElement(ui.TextElement("Disabled"), ui.Disabled(true))),
		),
		ui.VSpacerElement(10),
		ui.RowElement(
			ui.PressableElement(
				ui.ContainerDecorationElement(
					ui.Bg(th.Colors.SecondaryContainer).
						WithPad(ui.Symmetric(10, 14)).
						WithRad(8).
						WithHover(ui.Bg(th.Colors.TertiaryContainer)).
						WithPressed(ui.Bg(th.Colors.PrimaryContainer)),
					ui.TextElement("Pressable", ui.TextColor(th.Colors.OnSecondaryContainer)),
					ui.OnDecoPressed(func(ctx *ui.Context, value bool) { setIfChanged(pressed, value) }),
				),
				func(ctx *ui.Context) { pressCount.Set(pressCount.Value() + 1) },
				ui.PressableAttachRef(pressableRef.Current),
			),
			ui.HSpacerElement(10),
			ui.ClickAreaElement(
				ui.ContainerDecorationElement(ui.Bg(th.Colors.SurfaceContainerHigh).WithPad(ui.Symmetric(10, 14)).WithRad(8), ui.TextElement("ClickArea", ui.TextColor(th.Colors.OnSurface))),
				func(ctx *ui.Context) { pressCount.Set(pressCount.Value() + 1) },
			),
			ui.HSpacerElement(10),
			ui.ButtonElement(ui.TextElement("ButtonRef.Click"), ui.OnClick(func(ctx *ui.Context) {
				buttonRef.Current.Click()
			})),
			ui.HSpacerElement(10),
			ui.ButtonElement(ui.TextElement("PressableRef.Click"), ui.OnClick(func(ctx *ui.Context) {
				pressableRef.Current.Click()
			})),
			ui.ExpandedElement(ui.SpacerElement(0, 0)),
			ui.TextElement(fmt.Sprintf("button clicks=%d, hover=%t, pressed=%t, pressable clicks=%d", count.Value(), hovered.Value(), pressed.Value(), pressCount.Value()), ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
		),
	)
}

func inputSelectionPanel(
	th *ui.Theme,
	textValue *fluxstate.State[string],
	passwordValue *fluxstate.State[string],
	focusState *fluxstate.State[string],
	checked *fluxstate.State[bool],
	switched *fluxstate.State[bool],
	sliderValue *fluxstate.State[float32],
	radioValue *fluxstate.State[string],
	selectValue *fluxstate.State[string],
	menuOpen *fluxstate.State[bool],
	menuValue *fluxstate.State[string],
	inputRef *ui.Ref[*ui.InputRef],
	checkboxRef *ui.Ref[*ui.CheckboxRef],
	switchRef *ui.Ref[*ui.SwitchRef],
	sliderRef *ui.Ref[*ui.SliderRef],
	radioRef *ui.Ref[*ui.RadioGroupRef],
	selectRef *ui.Ref[*ui.SelectRef[string]],
) ui.Element {
	options := []ui.SelectOptionItem[string]{
		{Label: "Low priority", Value: "low"},
		{Label: "Medium priority", Value: "medium"},
		{Label: "High priority", Value: "high"},
	}
	menuItems := []ui.MenuItem{
		{Key: "copy", Label: "Copy", Leading: ui.Icon("C")},
		{Key: "share", Label: "Share", Leading: ui.Icon("S")},
		{Key: "archive", Label: "Archive", Leading: ui.Icon("A")},
		{Key: "delete", Label: "Delete", Leading: ui.Icon("D"), Disabled: true},
	}
	return ui.ColumnElement(
		ui.RowElement(
			ui.ExpandedElement(ui.ColumnElement(
				ui.OutlinedTextFieldElement(
					textValue.Value(),
					ui.InputPlaceholder("Outlined text"),
					ui.InputSingleLine(true),
					ui.InputMaxLen(40),
					ui.InputAttachRef(inputRef.Current),
					ui.InputOnFocus(func(ctx *ui.Context, focused bool) {
						next := "blurred"
						if focused {
							next = "focused"
						}
						setIfChanged(focusState, next)
					}),
					ui.InputOnChange(func(ctx *ui.Context, value string) { setIfChanged(textValue, value) }),
				),
				ui.VSpacerElement(8),
				ui.FilledTextFieldElement(
					passwordValue.Value(),
					ui.InputPlaceholder("Password"),
					ui.InputPassword(true),
					ui.InputSingleLine(true),
					ui.InputOnChange(func(ctx *ui.Context, value string) { setIfChanged(passwordValue, value) }),
				),
				ui.VSpacerElement(8),
				ui.TextFieldElement("disabled field", ui.InputDisabled(true)),
			)),
			ui.HSpacerElement(12),
			ui.FixedWidthElement(360, ui.ColumnElement(
				ui.RowElement(
					ui.ButtonElement(ui.TextElement("Set text"), ui.OnClick(func(ctx *ui.Context) {
						setIfChanged(textValue, "Visual test")
						inputRef.Current.SetText("Visual test")
					})),
					ui.HSpacerElement(8),
					ui.ButtonElement(ui.TextElement("Focus"), ui.OnClick(func(ctx *ui.Context) {
						setIfChanged(focusState, "focused")
						inputRef.Current.Focus()
					})),
					ui.HSpacerElement(8),
					ui.ButtonElement(ui.TextElement("Clear"), ui.OnClick(func(ctx *ui.Context) {
						setIfChanged(textValue, "")
						inputRef.Current.Clear()
					})),
				),
				ui.VSpacerElement(10),
				ui.TextElement("field value = "+textValue.Value()+" | "+focusState.Value(), ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
			)),
		),
		ui.VSpacerElement(14),
		ui.RowElement(
			ui.FixedWidthElement(270, ui.ColumnElement(
				ui.CheckboxElement("Checkbox", checked.Value(), ui.CheckboxAttachRef(checkboxRef.Current), ui.CheckboxOnChange(func(ctx *ui.Context, value bool) { setIfChanged(checked, value) })),
				ui.VSpacerElement(6),
				ui.RowElement(
					ui.TextElement("Switch", ui.TextType(th.Types.BodyMedium), ui.TextColor(th.Colors.OnSurface)),
					ui.HSpacerElement(8),
					ui.SwitchElement(switched.Value(), ui.SwitchAttachRef(switchRef.Current), ui.SwitchOnChange(func(ctx *ui.Context, value bool) { setIfChanged(switched, value) })),
				),
				ui.VSpacerElement(8),
				ui.RowElement(
					ui.ButtonElement(ui.TextElement("Toggle both"), ui.OnClick(func(ctx *ui.Context) {
						checkboxRef.Current.Toggle()
						switchRef.Current.Toggle()
					})),
				),
			)),
			ui.HSpacerElement(12),
			ui.FixedWidthElement(300, ui.ColumnElement(
				ui.SliderElement(
					sliderValue.Value(),
					ui.SliderMin(0),
					ui.SliderMax(100),
					ui.SliderStep(1),
					ui.SliderWidth(280),
					ui.SliderAttachRef(sliderRef.Current),
					ui.SliderOnChange(func(ctx *ui.Context, value float32) { setIfChanged(sliderValue, value) }),
				),
				ui.VSpacerElement(8),
				ui.RowElement(
					ui.ButtonElement(ui.TextElement("-10"), ui.OnClick(func(ctx *ui.Context) { sliderRef.Current.StepBy(-10) })),
					ui.HSpacerElement(8),
					ui.ButtonElement(ui.TextElement("+10"), ui.OnClick(func(ctx *ui.Context) { sliderRef.Current.StepBy(10) })),
					ui.HSpacerElement(8),
					ui.TextElement(fmt.Sprintf("value %.0f", sliderValue.Value()), ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
				),
			)),
			ui.HSpacerElement(12),
			ui.FixedWidthElement(270, ui.ColumnElement(
				ui.RadioGroupElement(
					radioValue.Value(),
					[]ui.RadioItem{{Label: "Stable", Value: "stable"}, {Label: "Beta", Value: "beta"}, {Label: "Nightly", Value: "nightly"}},
					ui.RadioGroupAttachRef(radioRef.Current),
					ui.RadioGroupDirection(ui.Vertical),
					ui.RadioGroupOnChange(func(ctx *ui.Context, value string) { setIfChanged(radioValue, value) }),
				),
				ui.VSpacerElement(8),
				ui.ButtonElement(ui.TextElement("RadioRef -> beta"), ui.OnClick(func(ctx *ui.Context) {
					setIfChanged(radioValue, "beta")
					radioRef.Current.SetValue("beta")
				})),
			)),
			ui.HSpacerElement(12),
			ui.ExpandedElement(ui.ColumnElement(
				ui.SelectElement(
					selectValue.Value(),
					options,
					ui.SelectAttachRef[string](selectRef.Current),
					ui.SelectOnChange[string](func(ctx *ui.Context, value string) { setIfChanged(selectValue, value) }),
				),
				ui.VSpacerElement(10),
				ui.RowElement(
					ui.ButtonElement(ui.TextElement("SelectRef.Open"), ui.OnClick(func(ctx *ui.Context) { selectRef.Current.Open() })),
					ui.HSpacerElement(8),
					ui.ButtonElement(ui.TextElement("SelectRef -> high"), ui.OnClick(func(ctx *ui.Context) {
						setIfChanged(selectValue, "high")
						selectRef.Current.SetValue("high")
					})),
				),
				ui.VSpacerElement(10),
				ui.RowElement(
					ui.FixedWidthElement(160, ui.DropdownMenuElement(
						menuOpen.Value(),
						ui.ContainerDecorationElement(ui.Bg(th.Colors.Surface).WithPad(ui.Symmetric(9, 12)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: th.Colors.Outline}), ui.TextElement("Dropdown", ui.TextColor(th.Colors.OnSurface))),
						menuItems,
						ui.DropdownMenuSelectedKey(menuValue.Value()),
						ui.DropdownMenuOnOpenChange(func(ctx *ui.Context, open bool) { setIfChanged(menuOpen, open) }),
						ui.DropdownMenuOnSelect(func(ctx *ui.Context, key string) {
							setIfChanged(menuValue, key)
							setIfChanged(menuOpen, false)
						}),
					)),
					ui.HSpacerElement(10),
					ui.FixedWidthElement(180, ui.MenuElement(menuItems, ui.MenuSelectedKey(menuValue.Value()), ui.MenuOnSelect(func(ctx *ui.Context, key string) { setIfChanged(menuValue, key) }))),
				),
			)),
		),
	)
}

func mediaCardsPanel(th *ui.Theme, imageClicks *fluxstate.State[int], cardClicks *fluxstate.State[int], iconSelected *fluxstate.State[bool], fabCount *fluxstate.State[int], chipSelected *fluxstate.State[bool], searchValue *fluxstate.State[string]) ui.Element {
	src := ui.ImageSource{Path: "examples/assets/sample.png", Label: "sample.png"}
	fabClick := ui.FloatingActionButtonOnClick(func(ctx *ui.Context) { fabCount.Set(fabCount.Value() + 1) })
	return ui.ColumnElement(
		ui.RowElement(
			ui.FixedWidthElement(360, ui.ColumnElement(
				ui.RowElement(
					ui.ImageElement(src, ui.ImageWidth(160), ui.ImageHeight(100), ui.ImageFitMode(ui.ImageFitContain), ui.ImageRadius(8), ui.ImageBackground(th.Colors.SurfaceContainerHighest), ui.ImageDecoration(ui.BorderDeco(1, th.Colors.OutlineVariant)), ui.ImageOnClick(func(ctx *ui.Context) { imageClicks.Set(imageClicks.Value() + 1) })),
					ui.HSpacerElement(12),
					ui.ImageElement(src, ui.ImageWidth(160), ui.ImageHeight(100), ui.ImageFitMode(ui.ImageFitCover), ui.ImageRadius(8)),
				),
				ui.VSpacerElement(8),
				ui.RowElement(
					ui.IconElement("H", ui.IconSize(22), ui.IconColor(th.Colors.Primary), ui.IconOnClick(func(ctx *ui.Context) { imageClicks.Set(imageClicks.Value() + 1) })),
					ui.HSpacerElement(12),
					ui.IconElement("S", ui.IconSize(22), ui.IconColor(th.Colors.Secondary)),
					ui.HSpacerElement(12),
					ui.TextElement(fmt.Sprintf("image/icon clicks=%d", imageClicks.Value()), ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
				),
			)),
			ui.HSpacerElement(12),
			ui.ExpandedElement(ui.RowElement(
				cardSample(th, "Base card", ui.CardElement, cardClicks),
				ui.HSpacerElement(10),
				cardSample(th, "Filled card", ui.FilledCardElement, cardClicks),
				ui.HSpacerElement(10),
				cardSample(th, "Elevated card", ui.ElevatedCardElement, cardClicks),
				ui.HSpacerElement(10),
				cardSample(th, "Outlined card", ui.OutlinedCardElement, cardClicks),
			)),
		),
		ui.VSpacerElement(14),
		ui.RowElement(
			ui.FixedWidthElement(360, ui.ColumnElement(
				ui.RowElement(
					wrap(ui.IconButtonElement(ui.IconElement("S"), ui.IconButtonSelected(iconSelected.Value()), ui.IconButtonOnClick(func(ctx *ui.Context) { iconSelected.Set(!iconSelected.Value()) }))),
					wrap(ui.FilledIconButtonElement(ui.IconElement("F"), ui.IconButtonSelected(true))),
					wrap(ui.FilledTonalIconButtonElement(ui.IconElement("T"))),
					wrap(ui.OutlinedIconButtonElement(ui.IconElement("O"))),
					wrap(ui.OutlinedIconButtonElement(ui.IconElement("D"), ui.IconButtonDisabled(true))),
				),
				ui.VSpacerElement(12),
				ui.RowElement(
					wrap(ui.SmallFloatingActionButtonElement(ui.IconElement("+"), fabClick)),
					wrap(ui.FloatingActionButtonElement(ui.IconElement("+"), fabClick)),
					wrap(ui.LargeFloatingActionButtonElement(ui.IconElement("+"), fabClick)),
					wrap(ui.ExtendedFloatingActionButtonElement(ui.IconElement("+"), ui.TextElement(fmt.Sprintf("Create %d", fabCount.Value())), fabClick)),
				),
			)),
			ui.HSpacerElement(12),
			ui.ExpandedElement(ui.ColumnElement(
				ui.RowElement(
					wrap(ui.AssistChipElement("Assist", ui.ChipLeading(ui.Icon("i", ui.IconSize(14))))),
					wrap(ui.FilterChipElement("Filter", ui.ChipSelected(chipSelected.Value()), ui.ChipOnClick(func(ctx *ui.Context) { chipSelected.Set(!chipSelected.Value()) }))),
					wrap(ui.InputChipElement("Input", ui.ChipTrailing(ui.Icon("x", ui.IconSize(14))))),
					wrap(ui.SuggestionChipElement("Suggestion")),
					wrap(ui.AssistChipElement("Disabled", ui.ChipDisabled(true))),
				),
				ui.VSpacerElement(10),
				ui.ChipElementWithSlots(
					ui.RowElement(
						ui.IconElement("S", ui.IconSize(14), ui.IconColor(th.Colors.Primary)),
						ui.HSpacerElement(6),
						ui.TextElement("Slot chip", ui.TextType(th.Types.LabelMedium), ui.TextColor(th.Colors.Primary)),
					),
					ui.ChipBackground(th.Colors.PrimaryContainer),
					ui.ChipForeground(th.Colors.OnPrimaryContainer),
				),
				ui.VSpacerElement(10),
				ui.RowElement(
					ui.BadgeElement(ui.IconButtonElement(ui.IconElement("M")), "3", ui.BadgeBackground(th.Colors.Error), ui.BadgeForeground(th.Colors.OnError)),
					ui.HSpacerElement(24),
					ui.BadgeElement(ui.IconButtonElement(ui.IconElement("N")), "", ui.BadgeVisible(true), ui.BadgeDecoration(ui.Bg(th.Colors.Primary).WithRad(999))),
					ui.HSpacerElement(24),
					ui.BadgeElement(ui.TextElement("Hidden"), "0", ui.BadgeVisible(false)),
				),
				ui.VSpacerElement(12),
				ui.SearchBarElement(
					searchValue.Value(),
					ui.SearchBarPlaceholder("Search components"),
					ui.SearchBarLeading(ui.Icon("S", ui.IconSize(18))),
					ui.SearchBarTrailing(ui.Icon("x", ui.IconSize(16))),
					ui.SearchBarOnChange(func(ctx *ui.Context, value string) { setIfChanged(searchValue, value) }),
				),
			)),
		),
	)
}

func navigationOverlayPanel(
	th *ui.Theme,
	tabValue *fluxstate.State[string],
	bottomNavValue *fluxstate.State[string],
	railValue *fluxstate.State[string],
	drawerValue *fluxstate.State[string],
	dialogOpen *fluxstate.State[bool],
	popupOpen *fluxstate.State[bool],
	toastMessage *fluxstate.State[string],
	toastSerial *fluxstate.State[int],
	snackbarMessage *fluxstate.State[string],
	snackbarSerial *fluxstate.State[int],
	snackbarActions *fluxstate.State[int],
	tabsRef *ui.Ref[*ui.TabsRef],
	dialogRef *ui.Ref[*ui.DialogRef],
	popupRef *ui.Ref[*ui.PopupRef],
	bottomNavRef *ui.Ref[*ui.BottomNavRef],
) ui.Element {
	navItems := []ui.ElementNavItem{
		{Key: "home", Label: "Home", Icon: ui.IconElement("H")},
		{Key: "docs", Label: "Docs", Icon: ui.IconElement("D")},
		{Key: "tools", Label: "Tools", Icon: ui.IconElement("T")},
	}
	return ui.ColumnElement(
		ui.AppBarElementWithSlots(
			ui.TextElement("Component Lab", ui.TextType(th.Types.TitleLarge), ui.TextColor(th.Colors.OnSurface)),
			ui.IconElement("<", ui.IconSize(18)),
			[]ui.Element{
				ui.TextButtonElement(ui.TextElement("Toast"), ui.OnClick(func(ctx *ui.Context) {
					toastSerial.Set(toastSerial.Value() + 1)
					setIfChanged(toastMessage, "Toast event")
				})),
				ui.FilledTonalButtonElement(ui.TextElement("Dialog"), ui.OnClick(func(ctx *ui.Context) {
					setIfChanged(dialogOpen, true)
					dialogRef.Current.Open()
				})),
			},
			ui.AppBarBackground(th.Colors.SurfaceContainer),
			ui.AppBarDecoration(ui.Bg(th.Colors.SurfaceContainer).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant})),
		),
		ui.VSpacerElement(10),
		ui.AppBarElement(
			ui.TextElement("Compact AppBarElement", ui.TextType(th.Types.TitleSmall), ui.TextColor(th.Colors.OnSurface)),
			ui.AppBarLeading(ui.Icon("<", ui.IconSize(16))),
			ui.AppBarActions(ui.IconButton(ui.Icon("A"), ui.IconButtonOnClick(func(ctx *ui.Context) {
				toastSerial.Set(toastSerial.Value() + 1)
				setIfChanged(toastMessage, "AppBar action")
			}))),
			ui.AppBarHeight(48),
			ui.AppBarBackground(th.Colors.SurfaceContainerHigh),
		),
		ui.VSpacerElement(10),
		ui.TabsElement(
			tabValue.Value(),
			[]ui.TabItem{{Key: "overview", Label: "Overview"}, {Key: "components", Label: "Components"}, {Key: "tokens", Label: "Tokens"}, {Key: "events", Label: "Events"}},
			ui.TabsAttachRef(tabsRef.Current),
			ui.TabsScrollable(true),
			ui.TabsOnChange(func(ctx *ui.Context, key string) { setIfChanged(tabValue, key) }),
		),
		ui.VSpacerElement(10),
		ui.RowElement(
			ui.ButtonElement(ui.TextElement("TabsRef -> tokens"), ui.OnClick(func(ctx *ui.Context) {
				setIfChanged(tabValue, "tokens")
				tabsRef.Current.SetActive("tokens")
			})),
			ui.HSpacerElement(8),
			ui.ButtonElement(ui.TextElement("Open popup"), ui.OnClick(func(ctx *ui.Context) {
				setIfChanged(popupOpen, true)
				popupRef.Current.Open()
			})),
			ui.HSpacerElement(8),
			ui.ButtonElement(ui.TextElement("Show snackbar"), ui.OnClick(func(ctx *ui.Context) {
				snackbarSerial.Set(snackbarSerial.Value() + 1)
				setIfChanged(snackbarMessage, "Saved visual state")
			})),
			ui.HSpacerElement(8),
			ui.ButtonElement(ui.TextElement("BottomNavRef -> docs"), ui.OnClick(func(ctx *ui.Context) {
				setIfChanged(bottomNavValue, "docs")
				bottomNavRef.Current.SetActive("docs")
			})),
			ui.HSpacerElement(8),
			ui.TextElement(fmt.Sprintf("snackbar actions=%d", snackbarActions.Value()), ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
		),
		ui.VSpacerElement(12),
		ui.RowElement(
			ui.FixedWidthElement(360, ui.FixedHeightElement(210, ui.ColumnElement(
				ui.ExpandedElement(ui.CenterElement(ui.TextElement("Bottom nav: "+bottomNavValue.Value(), ui.TextType(th.Types.BodyMedium), ui.TextColor(th.Colors.OnSurface)))),
				ui.BottomNavigationElement(
					bottomNavValue.Value(),
					navItems,
					ui.BottomNavAttachRef(bottomNavRef.Current),
					ui.BottomNavAlignmentOf(ui.BottomNavAlignSpaceEvenly),
					ui.BottomNavOnChange(func(ctx *ui.Context, key string) { setIfChanged(bottomNavValue, key) }),
				),
			))),
			ui.HSpacerElement(12),
			ui.FixedWidthElement(320, ui.FixedHeightElement(260, ui.RowElement(
				ui.NavigationRailElement(
					railValue.Value(),
					navItems,
					ui.NavigationRailWidth(88),
					ui.NavigationRailHeader(ui.Text("Menu", ui.TextType(th.Types.LabelMedium))),
					ui.NavigationRailFooter(ui.Text("v1", ui.TextType(th.Types.LabelSmall))),
					ui.NavigationRailOnChange(func(ctx *ui.Context, key string) { setIfChanged(railValue, key) }),
				),
				ui.ExpandedElement(ui.CenterElement(ui.TextElement("Rail: "+railValue.Value(), ui.TextColor(th.Colors.OnSurface)))),
			))),
			ui.HSpacerElement(12),
			ui.ExpandedElement(ui.FixedHeightElement(260, ui.RowElement(
				ui.NavigationDrawerElement(
					drawerValue.Value(),
					[]ui.ElementNavItem{
						{Key: "inbox", Label: "Inbox", Icon: ui.IconElement("I")},
						{Key: "sent", Label: "Sent", Icon: ui.IconElement("S")},
						{Key: "drafts", Label: "Drafts", Icon: ui.IconElement("D")},
					},
					ui.NavigationDrawerWidth(260),
					ui.NavigationDrawerHeader(ui.Text("Mailbox", ui.TextType(th.Types.TitleMedium))),
					ui.NavigationDrawerFooter(ui.Text("3 folders", ui.TextType(th.Types.LabelSmall))),
					ui.NavigationDrawerOnChange(func(ctx *ui.Context, key string) { setIfChanged(drawerValue, key) }),
				),
				ui.ExpandedElement(ui.CenterElement(ui.TextElement("Drawer: "+drawerValue.Value(), ui.TextColor(th.Colors.OnSurface)))),
			))),
		),
		ui.VSpacerElement(10),
		ui.TooltipElement("Tooltip overlay", ui.FilledTonalButtonElement(ui.TextElement("Hover for tooltip"))),
	)
}

func progressCollectionPanel(th *ui.Theme, progress *fluxstate.State[float32], scrollLog *fluxstate.State[string], listReachCount *fluxstate.State[int], gridReachCount *fluxstate.State[int], scrollRef *ui.Ref[*ui.ScrollRef]) ui.Element {
	return ui.ColumnElement(
		ui.RowElement(
			ui.FixedWidthElement(280, ui.ColumnElement(
				ui.ProgressBarElement(progress.Value(), ui.ProgressMin(0), ui.ProgressMax(100), ui.ProgressLabelVisible(true)),
				ui.VSpacerElement(10),
				ui.LinearProgressIndicatorElement(progress.Value(), ui.ProgressMin(0), ui.ProgressMax(100)),
				ui.VSpacerElement(10),
				ui.RowElement(
					ui.CircularProgressElement(progress.Value(), ui.ProgressSize(58), ui.ProgressLabelVisible(true)),
					ui.HSpacerElement(18),
					ui.CircularProgressIndicatorElement(progress.Value(), ui.ProgressSize(58), ui.ProgressLabelVisible(true)),
				),
			)),
			ui.HSpacerElement(16),
			ui.FixedWidthElement(330, ui.ColumnElement(
				ui.RowElement(
					ui.ButtonElement(ui.TextElement("Top"), ui.OnClick(func(ctx *ui.Context) {
						scrollRef.Current.ScrollToStart()
						setIfChanged(scrollLog, "scroll to start")
					})),
					ui.HSpacerElement(8),
					ui.ButtonElement(ui.TextElement("End"), ui.OnClick(func(ctx *ui.Context) {
						scrollRef.Current.ScrollToEnd()
						setIfChanged(scrollLog, "scroll to end")
					})),
					ui.HSpacerElement(8),
					ui.TextElement(scrollLog.Value(), ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
				),
				ui.VSpacerElement(8),
				ui.FixedHeightElement(150, ui.ScrollViewElement(
					ui.ColumnElement(scrollItems(th, 18)...),
					ui.ScrollAttachRef(scrollRef.Current),
					ui.ScrollOnChange(func(ctx *ui.Context, x, y float32) {
						setIfChanged(scrollLog, fmt.Sprintf("scroll y %.2f", y))
					}),
				)),
			)),
			ui.HSpacerElement(16),
			ui.ExpandedElement(ui.ColumnElement(
				ui.FixedHeightElement(150, ui.ListViewElement(28, func(ctx *ui.Context, index int) ui.Element {
					if index%5 == 0 {
						return ui.ListItemElement(fmt.Sprintf("Simple row %02d", index+1))
					}
					return ui.ListItemElementWithSlots(ui.TextElement(fmt.Sprintf("Virtual row %02d", index+1)), ui.TextElement("ListViewElement"), ui.IconElement("L"), nil)
				}, ui.ListItemSpacing(4), ui.ListOnReachEnd(func(ctx *ui.Context) { listReachCount.Set(listReachCount.Value() + 1) }))),
				ui.VSpacerElement(8),
				ui.TextElement(fmt.Sprintf("list reach end=%d", listReachCount.Value()), ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
			)),
		),
		ui.VSpacerElement(14),
		ui.RowElement(
			ui.FixedWidthElement(420, ui.GridElement(
				3,
				gridCell(th, "Grid A", th.Colors.PrimaryContainer, th.Colors.OnPrimaryContainer),
				gridCell(th, "Grid B", th.Colors.SecondaryContainer, th.Colors.OnSecondaryContainer),
				gridCell(th, "Grid C", th.Colors.TertiaryContainer, th.Colors.OnTertiaryContainer),
				gridCell(th, "Grid D", th.Colors.ErrorContainer, th.Colors.OnErrorContainer),
				gridCell(th, "Grid E", th.Colors.Success, th.Colors.OnSuccess),
				gridCell(th, "Grid F", th.Colors.Warning, th.Colors.OnWarning),
			)),
			ui.HSpacerElement(16),
			ui.ExpandedElement(ui.ColumnElement(
				ui.FixedHeightElement(190, ui.GridViewElement(42, 4, func(ctx *ui.Context, index int) ui.Element {
					return gridCell(th, fmt.Sprintf("Cell %02d", index+1), th.Colors.SurfaceContainerHighest, th.Colors.OnSurface)
				}, ui.GridGap(8, 8), ui.GridOnReachEnd(func(ctx *ui.Context) { gridReachCount.Set(gridReachCount.Value() + 1) }))),
				ui.VSpacerElement(8),
				ui.TextElement(fmt.Sprintf("grid reach end=%d", gridReachCount.Value()), ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
			)),
		),
	)
}

func dragDropRouterPanel(th *ui.Theme, dropActive *fluxstate.State[bool], dropLog *fluxstate.State[[]string], allowSettings *fluxstate.State[bool], userID *fluxstate.State[string], routerLog *fluxstate.State[string]) ui.Element {
	appendDrop := func(line string) {
		items := append([]string{}, dropLog.Value()...)
		items = append(items, line)
		if len(items) > 5 {
			items = items[len(items)-5:]
		}
		dropLog.Set(items)
	}
	return ui.RowElement(
		ui.FixedWidthElement(420, ui.ColumnElement(
			ui.RowElement(
				ui.DragSourceElement(
					ui.ContainerDecorationElement(
						ui.Bg(th.Colors.PrimaryContainer).WithPad(ui.All(14)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: th.Colors.Primary}),
						ui.TextElement("Drag text", ui.TextColor(th.Colors.OnPrimaryContainer)),
					),
					ui.DragSourceText("FluxUI drag payload"),
					ui.DragSourcePreview(ui.ContainerDecoration(ui.Bg(th.Colors.Primary).WithPad(ui.Symmetric(8, 12)).WithRad(8), ui.Text("Dragging", ui.TextColor(th.Colors.OnPrimary)))),
					ui.DragSourceOnEvent(func(ctx *ui.Context, event ui.DragSourceEvent) { appendDrop("source: " + string(event.Kind)) }),
				),
				ui.HSpacerElement(12),
				ui.ExpandedElement(ui.DropTargetElement(
					ui.ContainerDecorationElement(
						dropTargetDecoration(th, dropActive.Value()),
						ui.CenterElement(ui.TextElement("Drop here", ui.TextColor(th.Colors.OnSurface))),
					),
					func(ctx *ui.Context, event ui.DropEvent) {
						text := strings.TrimSpace(event.Text)
						if text == "" && len(event.Paths) > 0 {
							text = strings.Join(event.Paths, ", ")
						}
						if text == "" {
							text = event.Type
						}
						appendDrop("drop: " + text)
					},
					ui.DropTargetOnActiveChange(func(ctx *ui.Context, event ui.DropTargetStateEvent) {
						setIfChanged(dropActive, event.Active)
						if event.Active {
							appendDrop("target active")
						}
					}),
				)),
			),
			ui.VSpacerElement(10),
			ui.ContainerDecorationElement(
				ui.Bg(th.Colors.SurfaceContainerHigh).WithPad(ui.All(10)).WithRad(8),
				ui.ColumnElement(textLines(th, dropLog.Value())...),
			),
		)),
		ui.HSpacerElement(16),
		ui.ExpandedElement(ui.FixedHeightElement(300, routerPanel(th, allowSettings, userID, routerLog))),
	)
}

func routerPanel(th *ui.Theme, allowSettings *fluxstate.State[bool], userID *fluxstate.State[string], log *fluxstate.State[string]) ui.Element {
	home := func(ctx *ui.Context) ui.Element {
		navigate := ui.UseNavigate(ctx)
		return ui.ColumnElement(
			ui.TextElement("Router home", ui.TextType(th.Types.TitleSmall), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(8),
			ui.TextFieldElement(userID.Value(), ui.InputPlaceholder("user id"), ui.InputSingleLine(true), ui.InputOnChange(func(ctx *ui.Context, value string) { setIfChanged(userID, value) })),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ButtonElement(ui.TextElement("Detail"), ui.OnClick(func(ctx *ui.Context) {
					id := strings.TrimSpace(userID.Value())
					if id == "" {
						setIfChanged(log, "empty user id")
						return
					}
					navigate("/user/"+id+"?tab=profile", ui.WithNavTransition(ui.TransitionSlideLeft))
				})),
				ui.HSpacerElement(8),
				ui.ButtonElement(ui.TextElement("Settings"), ui.OnClick(func(ctx *ui.Context) {
					ui.Navigate(ctx, "/settings", ui.WithNavTransition(ui.TransitionFade))
				})),
				ui.HSpacerElement(8),
				ui.ButtonElement(ui.TextElement("404"), ui.OnClick(func(ctx *ui.Context) { navigate("/missing") })),
			),
		)
	}
	detail := func(ctx *ui.Context) ui.Element {
		params := ui.UseParams(ctx)
		id := params.Path("id")
		tab := params.Query("tab")
		if tab == "" {
			tab = "overview"
		}
		return ui.ColumnElement(
			ui.TextElement("User detail", ui.TextType(th.Types.TitleSmall), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(6),
			ui.TextElement("id="+id+" tab="+tab, ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ButtonElement(ui.TextElement("Activity"), ui.OnClick(func(ctx *ui.Context) { ui.NavigateReplace(ctx, "/user/"+id+"?tab=activity") })),
				ui.HSpacerElement(8),
				ui.ButtonElement(ui.TextElement("Back"), ui.OnClick(func(ctx *ui.Context) {
					if ui.CanGoBack(ctx) {
						ui.NavigateBack(ctx, ui.WithNavTransition(ui.TransitionSlideRight))
					}
				})),
			),
		)
	}
	settings := func(ctx *ui.Context) ui.Element {
		route := ui.UseRoute(ctx)
		title := "Settings"
		if route != nil && route.Title != "" {
			title = route.Title
		}
		return ui.ColumnElement(
			ui.TextElement(title, ui.TextType(th.Types.TitleSmall), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(6),
			ui.TextElement("guarded route", ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.ButtonElement(ui.TextElement("Back"), ui.OnClick(func(ctx *ui.Context) {
				if ui.CanGoBack(ctx) {
					ui.NavigateBack(ctx)
				}
			})),
		)
	}
	notFound := func(ctx *ui.Context) ui.Element {
		navigate := ui.UseNavigate(ctx)
		return ui.ColumnElement(
			ui.TextElement("Not found: "+ui.CurrentPath(ctx), ui.TextColor(th.Colors.Error)),
			ui.VSpacerElement(8),
			ui.ButtonElement(ui.TextElement("Home"), ui.OnClick(func(ctx *ui.Context) { navigate("/") })),
		)
	}
	router := ui.RouterElement(
		ui.RouteElement("/", home, ui.RouteName("home"), ui.RouteTitle("Home")),
		ui.RouteElement("/user/:id", detail, ui.RouteName("user-detail"), ui.RouteTitle("User Detail"), ui.RouteMeta("section", "users")),
		ui.RouteElement("/settings", settings, ui.RouteName("settings"), ui.RouteTitle("Settings"), ui.RouteBeforeEnter(func(ctx *ui.Context, from, to string) bool {
			if !allowSettings.Value() {
				setIfChanged(log, "blocked: "+from+" -> "+to)
				return false
			}
			return true
		})),
	).With(
		ui.RouterTransition(ui.TransitionSlideLeft),
		ui.RouterTransitionDuration(180*time.Millisecond),
		ui.RouterBeforeEach(func(ctx *ui.Context, from, to string) bool {
			setIfChanged(log, "nav: "+from+" -> "+to)
			return true
		}),
		ui.RouterNotFoundElement(notFound),
	)
	return ui.ColumnElement(
		ui.RowElement(
			ui.CheckboxElement("allow settings", allowSettings.Value(), ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) { setIfChanged(allowSettings, checked) })),
			ui.HSpacerElement(12),
			ui.TextElement(log.Value(), ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
		),
		ui.VSpacerElement(8),
		ui.ExpandedElement(ui.ContainerDecorationElement(ui.Bg(th.Colors.SurfaceContainerHigh).WithPad(ui.All(12)).WithRad(8), router)),
	)
}

func dialogOverlay(th *ui.Theme, open *fluxstate.State[bool], ref *ui.Ref[*ui.DialogRef]) ui.Element {
	if !open.Value() {
		return nil
	}
	return ui.DialogElement(
		open.Value(),
		ui.TextElement("DialogElement is rendered at the root stack so masking and focus can be checked against the full page.", ui.TextColor(th.Colors.OnSurface)),
		ui.DialogTitle("Dialog"),
		ui.DialogWidth(460),
		ui.DialogAttachRef(ref.Current),
		ui.DialogOnOpenChange(func(ctx *ui.Context, value bool) { setIfChanged(open, value) }),
		ui.DialogOnConfirm(func(ctx *ui.Context) {
			setIfChanged(open, false)
			ref.Current.Close()
		}),
		ui.DialogOnCancel(func(ctx *ui.Context) {
			setIfChanged(open, false)
			ref.Current.Close()
		}),
		ui.DialogConfirmText("Apply"),
		ui.DialogCancelText("Close"),
	)
}

func popupOverlay(th *ui.Theme, open *fluxstate.State[bool], ref *ui.Ref[*ui.PopupRef]) ui.Element {
	if !open.Value() {
		return nil
	}
	return ui.PopupElement(
		open.Value(),
		ui.ColumnElement(
			ui.TextElement("PopupElement", ui.TextType(th.Types.TitleMedium), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(8),
			ui.TextElement("Custom overlay content shares the active theme.", ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(12),
			ui.FilledButtonElement(ui.TextElement("Close"), ui.OnClick(func(ctx *ui.Context) {
				setIfChanged(open, false)
				ref.Current.Close()
			})),
		),
		ui.PopupWidth(380),
		ui.PopupPadding(ui.All(18)),
		ui.PopupAttachRef(ref.Current),
		ui.PopupOnOpenChange(func(ctx *ui.Context, value bool) { setIfChanged(open, value) }),
	)
}

func toastOverlay(message *fluxstate.State[string], serial *fluxstate.State[int]) ui.Element {
	if message.Value() == "" {
		return nil
	}
	return ui.Key(
		fmt.Sprintf("toast-%d", serial.Value()),
		ui.ToastElement(
			message.Value(),
			ui.ToastTypeOf(ui.ToastSuccess),
			ui.ToastPositionOf(ui.ToastTop),
			ui.ToastDuration(2200*time.Millisecond),
			ui.ToastOnClose(func(ctx *ui.Context) { setIfChanged(message, "") }),
		),
	)
}

func snackbarOverlay(message *fluxstate.State[string], serial *fluxstate.State[int], actionCount *fluxstate.State[int]) ui.Element {
	if message.Value() == "" {
		return nil
	}
	return ui.Key(
		fmt.Sprintf("snackbar-%d", serial.Value()),
		ui.SnackbarElement(
			message.Value(),
			ui.SnackbarAction("Undo", func(ctx *ui.Context) {
				actionCount.Set(actionCount.Value() + 1)
				setIfChanged(message, "")
			}),
			ui.ToastDuration(0),
		),
	)
}

func surfaceDemo(th *ui.Theme, title string, child ui.Element) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.Surface).WithPad(ui.All(12)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
		ui.ColumnElement(
			ui.TextElement(title, ui.TextType(th.Types.LabelLarge), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(10),
			child,
		),
	)
}

func cardSample(th *ui.Theme, title string, factory func(ui.Element, ...ui.CardOption) ui.Element, clicks *fluxstate.State[int]) ui.Element {
	return ui.ExpandedElement(factory(
		ui.ColumnElement(
			ui.TextElement(title, ui.TextType(th.Types.TitleSmall), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(6),
			ui.TextElement(fmt.Sprintf("clicks=%d", clicks.Value()), ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
		),
		ui.CardOnClick(func(ctx *ui.Context) { clicks.Set(clicks.Value() + 1) }),
	))
}

func pill(th *ui.Theme, label string, bg, fg color.NRGBA) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(bg).WithPad(ui.Symmetric(8, 10)).WithRad(999),
		ui.CenterElement(ui.TextElement(label, ui.TextType(th.Types.LabelMedium), ui.TextColor(fg))),
	)
}

func gridCell(th *ui.Theme, label string, bg, fg color.NRGBA) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(bg).
			WithPad(ui.All(10)).
			WithMargin(ui.All(4)).
			WithRad(8).
			WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
		ui.CenterElement(ui.TextElement(label, ui.TextType(th.Types.LabelMedium), ui.TextColor(fg))),
	)
}

func wrap(el ui.Element) ui.Element {
	return ui.PaddingElement(ui.Insets{Right: 8, Bottom: 8}, el)
}

func scrollItems(th *ui.Theme, count int) []ui.Element {
	items := make([]ui.Element, 0, count)
	for i := 0; i < count; i++ {
		items = append(items, ui.ContainerDecorationElement(
			ui.Bg(th.Colors.SurfaceContainerHigh).WithPad(ui.Symmetric(8, 10)).WithRad(7).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
			ui.TextElement(fmt.Sprintf("Scroll row %02d", i+1), ui.TextColor(th.Colors.OnSurface)),
		))
		if i < count-1 {
			items = append(items, ui.VSpacerElement(6))
		}
	}
	return items
}

func textLines(th *ui.Theme, lines []string) []ui.Element {
	out := make([]ui.Element, 0, len(lines))
	for _, line := range lines {
		out = append(out, ui.TextElement(line, ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)))
	}
	return out
}

func dropTargetDecoration(th *ui.Theme, active bool) ui.Decoration {
	deco := ui.Bg(th.Colors.SurfaceContainerHigh).WithPad(ui.All(18)).WithRad(8).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant})
	if active {
		deco = ui.Bg(th.Colors.PrimaryContainer).WithPad(ui.All(18)).WithRad(8).WithBorder(ui.Border{Width: 2, Color: th.Colors.Primary})
	}
	return deco
}

func setIfChanged[T comparable](state *fluxstate.State[T], value T) {
	if state == nil || state.Value() == value {
		return
	}
	state.Set(value)
}
