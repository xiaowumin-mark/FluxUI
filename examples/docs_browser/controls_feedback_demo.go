package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type docsFloat32State interface {
	Value() float32
	Set(float32)
}

func docsPressableDemo(clickCount docsIntState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		pressableRef := ui.UseRef(ctx, ui.NewPressableRef())
		clickAreaRef := ui.UseRef(ctx, ui.NewClickAreaRef())
		hovered := ui.UseState(ctx, false)
		pressed := ui.UseState(ctx, false)
		if pressableRef.Current == nil {
			pressableRef.Current = ui.NewPressableRef()
		}
		if clickAreaRef.Current == nil {
			clickAreaRef.Current = ui.NewClickAreaRef()
		}

		return ui.FixedWidthElement(
			520,
			ui.ColumnElement(
				ui.TextElement(fmt.Sprintf("Click count: %d", clickCount.Value()), ui.TextSize(13)),
				ui.VSpacerElement(8),
				ui.PressableElement(
					ui.FillWidthElement(
						ui.ContainerDecorationElement(
							ui.Bg(ui.NRGBA(227, 242, 253, 255)).
								WithPad(ui.All(14)).
								WithRad(8).
								WithHover(ui.Bg(ui.NRGBA(219, 234, 254, 255))).
								WithPressed(ui.Bg(ui.NRGBA(191, 219, 254, 255))),
							ui.TextElement(fmt.Sprintf("PressableElement hovered=%t pressed=%t", hovered.Value(), pressed.Value())),
							ui.OnDecoHover(func(ctx *ui.Context, value bool) {
								hovered.Set(value)
							}),
							ui.OnDecoPressed(func(ctx *ui.Context, value bool) {
								pressed.Set(value)
							}),
						),
					),
					func(ctx *ui.Context) {
						clickCount.Set(clickCount.Value() + 1)
					},
					ui.PressableAttachRef(pressableRef.Current),
				),
				ui.VSpacerElement(8),
				ui.RowElement(
					docsDemoControlButton("PressableRef.Click", func(ctx *ui.Context) {
						clickCount.Set(clickCount.Value() + 1)
						pressableRef.Current.Click()
					}),
					ui.HSpacerElement(8),
					ui.ClickAreaElement(
						ui.ContainerDecorationElement(
							ui.Bg(ui.NRGBA(248, 250, 252, 255)).
								WithPad(ui.Symmetric(6, 10)).
								WithRad(8).
								WithBorder(ui.Border{Width: 1, Color: ui.NRGBA(203, 213, 225, 255)}),
							ui.TextElement("ClickArea compatibility", ui.TextSize(12)),
						),
						func(ctx *ui.Context) {
							clickCount.Set(clickCount.Value() + 1)
						},
						ui.ClickAreaAttachRef(clickAreaRef.Current),
					),
				),
			),
		)
	})
}

func docsTextDemo(th *ui.Theme) ui.Element {
	return ui.ColumnElement(
		ui.TextElement("Default text"),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("Large text", ui.TextSize(20))),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("Primary color text", ui.TextColor(th.Primary))),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("Title scale with explicit line height", ui.TextType(th.Types.TitleMedium), ui.TextLineHeight(26))),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("Scoped font family", ui.TextFont(ui.FontFamily("Segoe UI")))),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("Centered semibold text", ui.TextAlign(ui.AlignCenter), ui.TextFontWeight(ui.FontWeightSemiBold))),
	)
}

func docsDemoControlButton(label string, onClick func(*ui.Context)) ui.Element {
	return ui.OutlinedButtonElement(
		ui.TextElement(label, ui.TextSize(12)),
		ui.ButtonPadding(ui.Symmetric(5, 10)),
		ui.OnClick(onClick),
	)
}
