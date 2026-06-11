package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsButtonDemo(buttonCount docsIntState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		ref := ui.UseRef(ctx, ui.NewButtonRef())
		hovered := ui.UseState(ctx, false)
		if ref.Current == nil {
			ref.Current = ui.NewButtonRef()
		}

		return ui.FixedWidthElement(
			680,
			ui.ColumnElement(
				ui.TextElement("MD3 button variants", ui.TextSize(14), ui.TextColor(ui.NRGBA(15, 23, 42, 255))),
				ui.VSpacerElement(10),
				ui.RowElement(
					ui.PaddingElement(ui.Insets{Right: 8}, ui.FilledButtonElement(
						ui.TextElement("Filled"),
						ui.ButtonAttachRef(ref.Current),
						ui.OnHover(func(ctx *ui.Context, value bool) {
							hovered.Set(value)
						}),
						docsButtonCountClick(buttonCount),
					)),
					ui.PaddingElement(ui.Insets{Right: 8}, ui.FilledTonalButtonElement(ui.TextElement("Tonal"), docsButtonCountClick(buttonCount))),
					ui.PaddingElement(ui.Insets{Right: 8}, ui.OutlinedButtonElement(ui.TextElement("Outlined"), docsButtonCountClick(buttonCount))),
					ui.PaddingElement(ui.Insets{Right: 8}, ui.TextButtonElement(ui.TextElement("Text"), docsButtonCountClick(buttonCount))),
					ui.ElevatedButtonElement(ui.TextElement("Elevated"), docsButtonCountClick(buttonCount)),
				),
				ui.VSpacerElement(10),
				ui.RowElement(
					ui.PaddingElement(
						ui.Insets{Right: 8},
						ui.ButtonElement(
							ui.TextElement("Custom"),
							ui.ButtonRadius(12),
							ui.ButtonBackground(ui.NRGBA(15, 23, 42, 255)),
							ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
							ui.ButtonDecoration(
								ui.Bg(ui.NRGBA(15, 23, 42, 255)).
									WithPad(ui.Symmetric(8, 14)).
									WithRad(12),
							),
							docsButtonCountClick(buttonCount),
						),
					),
					ui.PaddingElement(ui.Insets{Right: 8}, ui.ButtonElement(ui.TextElement("Disabled"), ui.Disabled(true))),
					docsDemoControlButton("ButtonRef.Click", func(ctx *ui.Context) {
						buttonCount.Set(buttonCount.Value() + 1)
						ref.Current.Click()
					}),
					ui.ExpandedElement(ui.SpacerElement(0, 0)),
					ui.TextElement(fmt.Sprintf("clicks = %d hovered = %t", buttonCount.Value(), hovered.Value()), ui.TextSize(13), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
				),
			),
		)
	})
}

func docsButtonCountClick(buttonCount docsIntState) ui.ButtonOption {
	return ui.OnClick(func(ctx *ui.Context) {
		buttonCount.Set(buttonCount.Value() + 1)
	})
}

func docsTextFieldDemo(inputValue docsStringState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		ref := ui.UseRef(ctx, ui.NewInputRef())
		focusState := ui.UseState(ctx, "blurred")
		if ref.Current == nil {
			ref.Current = ui.NewInputRef()
		}

		return ui.FixedWidthElement(
			680,
			ui.ColumnElement(
				ui.TextElement("Controlled input variants", ui.TextSize(14), ui.TextColor(ui.NRGBA(15, 23, 42, 255))),
				ui.VSpacerElement(10),
				ui.RowElement(
					ui.ExpandedElement(
						ui.OutlinedTextFieldElement(
							inputValue.Value(),
							ui.InputPlaceholder("Outlined"),
							ui.InputSingleLine(true),
							ui.InputMaxLen(32),
							ui.InputPadding(ui.Symmetric(8, 12)),
							ui.InputRadius(10),
							ui.InputBorder(ui.NRGBA(148, 163, 184, 255)),
							ui.InputBorderFocus(ui.NRGBA(37, 99, 235, 255)),
							ui.InputBackground(ui.NRGBA(255, 255, 255, 255)),
							ui.InputForeground(ui.NRGBA(15, 23, 42, 255)),
							ui.InputTextSize(13),
							ui.InputFontFamily("Segoe UI"),
							ui.InputDecoration(ui.Focused(ui.BorderDeco(2, ui.NRGBA(37, 99, 235, 255)))),
							ui.InputAttachRef(ref.Current),
							ui.InputOnFocus(func(ctx *ui.Context, focused bool) {
								if focused {
									focusState.Set("focused")
								} else {
									focusState.Set("blurred")
								}
							}),
							ui.InputOnChange(func(ctx *ui.Context, value string) {
								inputValue.Set(value)
							}),
						),
					),
					ui.HSpacerElement(10),
					ui.ExpandedElement(
						ui.FilledTextFieldElement(
							inputValue.Value(),
							ui.InputPlaceholder("Filled"),
							ui.InputSingleLine(true),
							ui.InputMaxLen(32),
							ui.InputOnChange(func(ctx *ui.Context, value string) {
								inputValue.Set(value)
							}),
						),
					),
				),
				ui.VSpacerElement(10),
				ui.RowElement(
					docsDemoControlButton("SetText", func(ctx *ui.Context) {
						inputValue.Set("FluxUI")
						ref.Current.SetText("FluxUI")
					}),
					ui.HSpacerElement(8),
					docsDemoControlButton("Append", func(ctx *ui.Context) {
						inputValue.Set(inputValue.Value() + "!")
						ref.Current.Append("!")
					}),
					ui.HSpacerElement(8),
					docsDemoControlButton("Clear", func(ctx *ui.Context) {
						inputValue.Set("")
						ref.Current.Clear()
					}),
					ui.HSpacerElement(8),
					docsDemoControlButton("Focus", func(ctx *ui.Context) {
						focusState.Set("focused")
						ref.Current.Focus()
					}),
					ui.HSpacerElement(8),
					docsDemoControlButton("Blur", func(ctx *ui.Context) {
						focusState.Set("blurred")
						ref.Current.Blur()
					}),
				),
				ui.VSpacerElement(10),
				ui.RowElement(
					ui.ExpandedElement(
						ui.TextFieldElement(
							"secret",
							ui.InputPlaceholder("Password"),
							ui.InputPassword(true),
							ui.InputSingleLine(true),
						),
					),
					ui.HSpacerElement(10),
					ui.ExpandedElement(
						ui.TextFieldElement(
							"disabled",
							ui.InputDisabled(true),
						),
					),
				),
				ui.VSpacerElement(8),
				ui.TextElement("value = "+inputValue.Value()+" | "+focusState.Value(), ui.TextSize(13), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
			),
		)
	})
}

func docsCardDemo(buttonCount docsIntState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		ref := ui.UseRef(ctx, ui.NewButtonRef())
		if ref.Current == nil {
			ref.Current = ui.NewButtonRef()
		}
		click := ui.CardOnClick(func(ctx *ui.Context) {
			buttonCount.Set(buttonCount.Value() + 1)
		})
		return ui.FixedWidthElement(
			720,
			ui.ColumnElement(
				ui.TextElement("MD3 card variants", ui.TextSize(14), ui.TextColor(ui.NRGBA(15, 23, 42, 255))),
				ui.VSpacerElement(10),
				ui.RowElement(
					ui.ExpandedElement(docsCardVariant("Filled", "Default surface card.", ui.FilledCardElement, click, ui.CardAttachRef(ref.Current))),
					ui.HSpacerElement(10),
					ui.ExpandedElement(docsCardVariant("Elevated", "Adds tonal elevation.", ui.ElevatedCardElement, click, ui.CardShadow(2))),
					ui.HSpacerElement(10),
					ui.ExpandedElement(docsCardVariant("Outlined", "Uses outline boundary.", ui.OutlinedCardElement, click, ui.CardBorder(ui.NRGBA(37, 99, 235, 255), 1))),
				),
				ui.VSpacerElement(10),
				ui.RowElement(
					ui.ExpandedElement(
						docsCardVariant(
							"Custom options",
							"Padding, radius, background, border, shadow, decoration.",
							ui.CardElement,
							click,
							ui.CardPadding(ui.All(14)),
							ui.CardRadius(16),
							ui.CardBackground(ui.NRGBA(240, 253, 244, 255)),
							ui.CardBorder(ui.NRGBA(22, 163, 74, 255), 1),
							ui.CardShadow(1),
							ui.CardDecoration(ui.HoverBg(ui.NRGBA(220, 252, 231, 255))),
						),
					),
					ui.HSpacerElement(10),
					docsDemoControlButton("CardRef.Click", func(ctx *ui.Context) {
						buttonCount.Set(buttonCount.Value() + 1)
						ref.Current.Click()
					}),
				),
				ui.VSpacerElement(8),
				ui.TextElement(fmt.Sprintf("card clicks = %d", buttonCount.Value()), ui.TextSize(13), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
			),
		)
	})
}

func docsCardVariant(label string, body string, card func(ui.Element, ...ui.CardOption) ui.Element, opts ...ui.CardOption) ui.Element {
	return card(
		ui.ColumnElement(
			ui.TextElement(label, ui.TextSize(13)),
			ui.VSpacerElement(6),
			ui.TextElement(body, ui.TextSize(12), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
		),
		opts...,
	)
}
