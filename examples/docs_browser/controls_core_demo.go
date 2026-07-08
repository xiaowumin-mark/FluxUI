package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsButtonDemo(buttonCount docsIntState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		ref := ui.UseRef(ctx, ui.NewButtonRef())
		hovered := ui.UseState(ctx, false)
		if ref.Current == nil {
			ref.Current = ui.NewButtonRef()
		}

		return ui.FixedWidthElement(
			680,
			ui.ColumnElement(
				ui.TextElement("MD3 按钮变体", ui.TextSize(14), ui.TextColor(th.Colors.OnSurface)),
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
							ui.ButtonBackground(th.Colors.Primary),
							ui.ButtonForeground(th.Colors.OnPrimary),
							ui.ButtonDecoration(
								ui.Bg(th.Colors.Primary).
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
					ui.TextElement(fmt.Sprintf("点击 = %d 悬停 = %t", buttonCount.Value(), hovered.Value()), ui.TextSize(13), ui.TextColor(th.Colors.OnSurfaceVariant)),
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
		th := ui.UseTheme(ctx)
		ref := ui.UseRef(ctx, ui.NewInputRef())
		focusState := ui.UseState(ctx, "blurred")
		if ref.Current == nil {
			ref.Current = ui.NewInputRef()
		}

		return ui.FixedWidthElement(
			680,
			ui.ColumnElement(
				ui.TextElement("受控输入变体", ui.TextSize(14), ui.TextColor(th.Colors.OnSurface)),
				ui.VSpacerElement(10),
				ui.RowElement(
					ui.ExpandedElement(
						ui.OutlinedTextFieldElement(
							inputValue.Value(),
							ui.InputLabel("用户名"),
							ui.InputPlaceholder("name@example.com"),
							ui.InputSingleLine(true),
							ui.InputMaxLen(32),
							ui.InputCounter(true),
							ui.InputLeading(ui.Icon("person")),
							ui.InputTrailing(ui.Icon("mail")),
							ui.InputSupportingText("浮动标签、图标与计数器"),
							ui.InputFontFamily("Segoe UI"),
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
							ui.InputLabel("Filled"),
							ui.InputPrefixText("$"),
							ui.InputSuffixText("USD"),
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
					docsDemoControlButton("设置文本", func(ctx *ui.Context) {
						inputValue.Set("FluxUI")
						ref.Current.SetText("FluxUI")
					}),
					ui.HSpacerElement(8),
					docsDemoControlButton("追加", func(ctx *ui.Context) {
						inputValue.Set(inputValue.Value() + "!")
						ref.Current.Append("!")
					}),
					ui.HSpacerElement(8),
					docsDemoControlButton("清除", func(ctx *ui.Context) {
						inputValue.Set("")
						ref.Current.Clear()
					}),
					ui.HSpacerElement(8),
					docsDemoControlButton("聚焦", func(ctx *ui.Context) {
						focusState.Set("focused")
						ref.Current.Focus()
					}),
					ui.HSpacerElement(8),
					docsDemoControlButton("失焦", func(ctx *ui.Context) {
						focusState.Set("blurred")
						ref.Current.Blur()
					}),
				),
				ui.VSpacerElement(10),
				ui.RowElement(
					ui.ExpandedElement(
						ui.TextFieldElement(
							"secret",
							ui.InputLabel("密码"),
							ui.InputPassword(true),
							ui.InputSingleLine(true),
							ui.InputLeading(ui.Icon("lock")),
							ui.InputTrailing(ui.Icon("visibility")),
						),
					),
					ui.HSpacerElement(10),
					ui.ExpandedElement(
						ui.TextFieldElement(
							"disabled",
							ui.InputLabel("已禁用"),
							ui.InputDisabled(true),
						),
					),
				),
				ui.VSpacerElement(10),
				ui.RowElement(
					ui.ExpandedElement(
						ui.OutlinedTextFieldElement(
							"Draft",
							ui.InputLabel("标题"),
							ui.InputMaxLen(10),
							ui.InputCounter(true),
							ui.InputError(true),
							ui.InputErrorText("使用 10 个字符或更少"),
							ui.InputRequired(true),
						),
					),
					ui.HSpacerElement(10),
					ui.ExpandedElement(
						ui.OutlinedTextFieldElement(
							"Multiline value",
							ui.InputLabel("备注"),
							ui.InputSingleLine(false),
							ui.InputRows(3),
							ui.InputSupportingText("多行文本字段"),
						),
					),
				),
				ui.VSpacerElement(8),
				ui.TextElement("值 = "+inputValue.Value()+" | "+focusState.Value(), ui.TextSize(13), ui.TextColor(th.Colors.OnSurfaceVariant)),
			),
		)
	})
}

func docsCardDemo(buttonCount docsIntState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
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
				ui.TextElement("MD3 卡片变体", ui.TextSize(14), ui.TextColor(th.Colors.OnSurface)),
				ui.VSpacerElement(10),
				ui.RowElement(
					ui.ExpandedElement(docsCardVariant("Filled", "默认 Surface 卡片。", th, ui.FilledCardElement, click, ui.CardAttachRef(ref.Current))),
					ui.HSpacerElement(10),
					ui.ExpandedElement(docsCardVariant("Elevated", "添加暮色海拔。", th, ui.ElevatedCardElement, click, ui.CardShadow(2))),
					ui.HSpacerElement(10),
					ui.ExpandedElement(docsCardVariant("Outlined", "使用轮廓边界。", th, ui.OutlinedCardElement, click, ui.CardBorder(th.Colors.Primary, 1))),
				),
				ui.VSpacerElement(10),
				ui.RowElement(
					ui.ExpandedElement(
						docsCardVariant(
							"自定义选项",
							"内边距、圆角、背景、边框、阴影、装饰。",
							th,
							ui.CardElement,
							click,
							ui.CardPadding(ui.All(14)),
							ui.CardRadius(16),
							ui.CardBackground(th.Colors.PrimaryContainer),
							ui.CardBorder(th.Colors.Primary, 1),
							ui.CardShadow(1),
							ui.CardDecoration(ui.HoverBg(th.Colors.SecondaryContainer)),
						),
					),
					ui.HSpacerElement(10),
					docsDemoControlButton("CardRef.Click", func(ctx *ui.Context) {
						buttonCount.Set(buttonCount.Value() + 1)
						ref.Current.Click()
					}),
				),
				ui.VSpacerElement(8),
				ui.TextElement(fmt.Sprintf("卡片点击 = %d", buttonCount.Value()), ui.TextSize(13), ui.TextColor(th.Colors.OnSurfaceVariant)),
			),
		)
	})
}

func docsCardVariant(label string, body string, th *ui.Theme, card func(ui.Element, ...ui.CardOption) ui.Element, opts ...ui.CardOption) ui.Element {
	return card(
		ui.ColumnElement(
			ui.TextElement(label, ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(6),
			ui.TextElement(body, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		),
		opts...,
	)
}
