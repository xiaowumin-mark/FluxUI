package main

import (
	"fmt"
	"image"
	"image/color"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsRowDemo() ui.Element {
	return ui.RowElement(
		ui.ContainerElement(
			ui.Style{Background: ui.NRGBA(30, 136, 229, 255), Padding: ui.All(10), Radius: 6},
			ui.TextElement("A", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
		),
		ui.PaddingElement(
			ui.Insets{Left: 8},
			ui.ContainerElement(
				ui.Style{Background: ui.NRGBA(67, 160, 71, 255), Padding: ui.All(10), Radius: 6},
				ui.TextElement("B", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
			),
		),
		ui.PaddingElement(
			ui.Insets{Left: 8},
			ui.ContainerElement(
				ui.Style{Background: ui.NRGBA(245, 124, 0, 255), Padding: ui.All(10), Radius: 6},
				ui.TextElement("C", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
			),
		),
	)
}

func docsColumnDemo() ui.Element {
	return ui.ColumnElement(
		docsColorBlock("First row", ui.NRGBA(30, 136, 229, 255)),
		ui.PaddingElement(ui.Insets{Top: 8}, docsColorBlock("Second row", ui.NRGBA(67, 160, 71, 255))),
		ui.PaddingElement(ui.Insets{Top: 8}, docsColorBlock("Expanded area", ui.NRGBA(245, 124, 0, 255))),
	)
}

func docsStackDemo() ui.Element {
	return ui.FixedHeightElement(
		120,
		ui.FillElement(
			ui.StackElement(
				ui.FillElement(
					ui.ContainerElement(
						ui.Style{Background: ui.NRGBA(234, 239, 245, 255), Radius: 8},
						ui.SpacerElement(0, 0),
					),
				),
				ui.PaddingElement(
					ui.Insets{Left: 12, Top: 12},
					ui.ContainerElement(
						ui.Style{Background: ui.NRGBA(30, 136, 229, 255), Padding: ui.All(6), Radius: 6},
						ui.TextElement("Layer 1", ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(12)),
					),
				),
				ui.CenterElement(
					ui.TextElement("Center Layer", ui.TextColor(ui.NRGBA(15, 23, 42, 255)), ui.TextSize(14)),
				),
			),
		),
	)
}

func docsCenterDemo() ui.Element {
	return ui.FixedHeightElement(
		120,
		ui.FillElement(
			ui.ContainerElement(
				ui.Style{Background: ui.NRGBA(240, 244, 248, 255), Radius: 8},
				ui.CenterElement(ui.TextElement("Centered content", ui.TextSize(14))),
			),
		),
	)
}

func docsContainerDemo() ui.Element {
	return ui.ColumnElement(
		ui.ContainerElement(
			ui.Style{
				Background: ui.NRGBA(30, 136, 229, 255),
				Padding:    ui.All(16),
				Margin:     ui.Only(0, 0, 8, 0),
				Radius:     10,
			},
			ui.TextElement("Container: background + padding + margin + radius", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
		),
		ui.TextElement("Style uses All and Only insets.", ui.TextSize(12), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
	)
}

func docsDecorationDemo(active docsBoolState, th *ui.Theme) ui.Element {
	base := ui.Bg(ui.NRGBA(255, 255, 255, 255)).
		WithPad(ui.All(16)).
		WithMargin(ui.All(8)).
		WithRad(16).
		WithBorder(ui.Border{Width: 1, Color: ui.NRGBA(203, 213, 225, 255)}).
		WithHover(ui.Bg(ui.NRGBA(239, 246, 255, 255))).
		WithPressed(ui.Bg(ui.NRGBA(219, 234, 254, 255))).
		Merge(ui.Focused(ui.BorderDeco(2, th.Primary))).
		Merge(ui.DisabledDeco(ui.Opacity(0.46)))
	if active.Value() {
		base = base.Merge(ui.Elevation(2)).WithBorder(ui.Border{Width: 1, Color: th.Primary})
	}
	return ui.PaddingElement(
		ui.Symmetric(8, 12),
		ui.ClickAreaElement(
			ui.FixedWidthElement(
				360,
				ui.ContainerDecorationElement(
					base,
					ui.ColumnElement(
						ui.TextElement("Decoration card", ui.TextSize(16), ui.TextColor(th.TextColor)),
						ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("Click to toggle border and elevation.", ui.TextSize(12), ui.TextColor(ui.NRGBA(71, 85, 105, 255)))),
						ui.PaddingElement(ui.Insets{Top: 8}, docsDecorationStateDots(th)),
						ui.PaddingElement(ui.Insets{Top: 10}, docsDecorationEventSample(th)),
					),
				),
			),
			func(ctx *ui.Context) { active.Set(!active.Value()) },
		),
	)
}

func docsDecorationEventSample(th *ui.Theme) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.SurfaceContainerLow).
			WithPad(ui.Symmetric(6, 8)).
			WithRad(8).
			WithDisabled(ui.Opacity(0.5)),
		ui.TextElement("Disabled event target", ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
		ui.ContainerDecorationDisabled(true),
		ui.OnDecoClick(func(ctx *ui.Context) {}),
		ui.OnDecoHoverEnter(func(ctx *ui.Context) {}),
		ui.OnDecoHoverLeave(func(ctx *ui.Context) {}),
	)
}

func docsInsetsDemo() ui.Element {
	return ui.ColumnElement(
		ui.ContainerDecorationElement(
			ui.Bg(ui.NRGBA(239, 246, 255, 255)).WithRad(12),
			ui.PaddingElement(
				ui.Symmetric(10, 18),
				ui.TextElement("Symmetric(10, 18): top/bottom 10, left/right 18", ui.TextColor(ui.NRGBA(30, 64, 175, 255))),
			),
		),
		ui.VSpacerElement(8),
		ui.ContainerDecorationElement(
			ui.Bg(ui.NRGBA(240, 253, 244, 255)).WithRad(12),
			ui.PaddingElement(
				ui.LeftRight(20),
				ui.TextElement("LeftRight(20) + TopBottom available for focused spacing", ui.TextColor(ui.NRGBA(22, 101, 52, 255))),
			),
		),
	)
}

func docsBorderDemo(th *ui.Theme) ui.Element {
	return ui.RowElement(
		docsBorderSample("WithBorder", ui.Bg(ui.NRGBA(255, 255, 255, 255)).WithPad(ui.All(14)).WithRad(12).WithBorder(ui.Border{Width: 2, Color: th.Primary}), th.Primary),
		ui.HSpacerElement(12),
		docsBorderSample("BorderDeco", ui.Bg(ui.NRGBA(255, 255, 255, 255)).WithPad(ui.All(14)).WithRad(12).Merge(ui.BorderDeco(2, th.Colors.Tertiary)), th.Colors.Tertiary),
	)
}

func docsShadowDemo() ui.Element {
	return ui.PaddingElement(
		ui.Symmetric(8, 12),
		ui.RowElement(
			ui.FixedWidthElement(
				230,
				ui.ContainerDecorationElement(
					ui.Bg(ui.NRGBA(255, 255, 255, 255)).
						WithPad(ui.All(16)).
						WithMargin(ui.All(8)).
						WithRad(16).
						Merge(ui.Elevation(3)),
					ui.TextElement("Elevation(3) card"),
				),
			),
			ui.FixedWidthElement(
				230,
				ui.ContainerDecorationElement(
					ui.Bg(ui.NRGBA(255, 255, 255, 255)).
						WithPad(ui.All(16)).
						WithMargin(ui.All(8)).
						WithRad(16).
						Merge(ui.Shadow(0, 10, 26, ui.NRGBA(15, 23, 42, 60))),
					ui.TextElement("Custom Shadow"),
				),
			),
		),
	)
}

func docsGradientDemo() ui.Element {
	return ui.ContainerDecorationElement(
		ui.LinearGrad(
			image.Point{X: 0, Y: 0},
			image.Point{X: 260, Y: 120},
			ui.NRGBA(14, 165, 233, 255),
			ui.NRGBA(34, 197, 94, 255),
		).WithPad(ui.All(18)).WithRad(18),
		ui.TextElement("LinearGrad background", ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(16)),
	)
}

func docsTransformDemo() ui.Element {
	return ui.RowElement(
		ui.PaddingElement(
			ui.All(10),
			ui.ContainerDecorationElement(
				ui.Bg(ui.NRGBA(254, 243, 199, 255)).
					WithPad(ui.All(14)).
					WithRad(12).
					Merge(ui.Rotate(-4)),
				ui.TextElement("Rotate(-4)"),
			),
		),
		ui.PaddingElement(
			ui.All(10),
			ui.ContainerDecorationElement(
				ui.Bg(ui.NRGBA(219, 234, 254, 255)).
					WithPad(ui.All(14)).
					WithRad(12).
					Merge(ui.ScaleDeco(1.08, 1.08)).
					Merge(ui.TranslateDeco(6, 2)),
				ui.TextElement("Scale + Translate"),
			),
		),
	)
}

func docsImageFillDemo() ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(ui.NRGBA(15, 23, 42, 255)).WithPad(ui.All(16)).WithRad(14),
		ui.ColumnElement(
			ui.TextElement("ImageBg(img, ImageFillCover)", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
			ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("LoadImage, DecodeImageFile, and DecodeImageURL feed ImageBg after async loading.", ui.TextSize(12), ui.TextColor(ui.NRGBA(203, 213, 225, 255)))),
			ui.PaddingElement(ui.Insets{Top: 8}, ui.TextElement(fmt.Sprintf("fit modes: %d cover / %d contain", ui.ImageFillCover, ui.ImageFillContain), ui.TextSize(11), ui.TextColor(ui.NRGBA(148, 163, 184, 255)))),
		),
	)
}

func docsThemeDemo(dark docsBoolState) ui.Element {
	localTheme := ui.NewTheme(ui.LightColors())
	label := "Local light theme"
	if dark.Value() {
		localTheme = ui.NewTheme(ui.DarkColors())
		label = "Local dark theme"
	}
	return ui.ColumnElement(
		ui.ButtonElement(ui.TextElement("Toggle local theme"), ui.OnClick(func(ctx *ui.Context) {
			dark.Set(!dark.Value())
		})),
		ui.PaddingElement(
			ui.Insets{Top: 8},
			ui.ThemeProviderElement(
				localTheme,
				ui.ContainerDecorationElement(
					ui.Bg(localTheme.Surface).WithPad(ui.All(14)).WithRad(12),
					ui.TextElement(label, ui.TextColor(localTheme.TextColor)),
				),
			),
		),
	)
}

func docsColorSchemeDemo(th *ui.Theme) ui.Element {
	return ui.RowElement(
		ui.ContainerDecorationElement(ui.Bg(th.Primary).WithPad(ui.All(10)).WithRad(8), ui.TextElement("Primary", ui.TextColor(th.TextOnPrimary))),
		ui.PaddingElement(ui.Insets{Left: 8}, ui.ContainerDecorationElement(ui.Bg(th.Colors.Warning).WithPad(ui.All(10)).WithRad(8), ui.TextElement("Warning", ui.TextColor(th.Colors.OnWarning)))),
		ui.PaddingElement(ui.Insets{Left: 8}, ui.ContainerDecorationElement(ui.Bg(th.Colors.Success).WithPad(ui.All(10)).WithRad(8), ui.TextElement("Success", ui.TextColor(th.Colors.OnSuccess)))),
		ui.PaddingElement(ui.Insets{Left: 8}, ui.ContainerDecorationElement(ui.Circle().Merge(ui.Bg(th.Colors.Tertiary)).WithPad(ui.All(10)), ui.TextElement("T", ui.TextColor(th.Colors.OnTertiary)))),
	)
}

func docsGettingStartedDemo(count docsIntState, th *ui.Theme) ui.Element {
	return ui.ColumnElement(
		ui.TextElement("Hello FluxUI", ui.TextSize(18)),
		ui.PaddingElement(ui.Insets{Top: 8}, ui.ButtonElement(ui.TextElement("count +1"), ui.OnClick(func(ctx *ui.Context) {
			count.Set(count.Value() + 1)
		}))),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement(fmt.Sprintf("count = %d", count.Value()), ui.TextColor(th.Primary))),
	)
}

func docsPaddingDemo() ui.Element {
	return ui.ContainerElement(
		ui.Style{
			Background: ui.NRGBA(229, 236, 246, 255),
			Radius:     8,
		},
		ui.PaddingElement(
			ui.All(16),
			ui.TextElement("Padding: this area has 16dp inset"),
		),
	)
}

func docsSpacerDemo() ui.Element {
	return ui.RowElement(
		ui.TextElement("Left"),
		ui.HSpacerElement(24),
		ui.TextElement("Right"),
		ui.HSpacerElement(24),
		ui.ColumnElement(
			ui.TextElement("Top"),
			ui.VSpacerElement(8),
			ui.TextElement("Bottom"),
		),
	)
}

func docsDividerDemo() ui.Element {
	return ui.RowElement(
		ui.FixedWidthElement(
			220,
			ui.ColumnElement(
				ui.TextElement("First section"),
				ui.DividerElement(ui.DividerThickness(1), ui.DividerColor(ui.NRGBA(176, 190, 197, 255)), ui.DividerMargin(ui.Insets{Top: 8, Bottom: 8})),
				ui.TextElement("Second section"),
			),
		),
		ui.HSpacerElement(18),
		ui.FixedHeightElement(
			70,
			ui.RowElement(
				ui.TextElement("Left"),
				ui.DividerElement(ui.DividerVertical(true), ui.DividerThickness(2), ui.DividerLength(56), ui.DividerMargin(ui.LeftRight(10))),
				ui.TextElement("Right"),
			),
		),
	)
}

func docsSizingDemo() ui.Element {
	return ui.ColumnElement(
		ui.RowElement(
			ui.FixedWidthElement(
				110,
				ui.ContainerElement(
					ui.Style{Background: ui.NRGBA(3, 169, 244, 255), Padding: ui.All(8), Radius: 6},
					ui.TextElement("FixedWidth", ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(12)),
				),
			),
			ui.PaddingElement(
				ui.Insets{Left: 8},
				ui.ExpandedElement(
					ui.ContainerElement(
						ui.Style{Background: ui.NRGBA(76, 175, 80, 255), Padding: ui.All(8), Radius: 6},
						ui.TextElement("Expanded / Fill", ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(12)),
					),
				),
			),
		),
		ui.PaddingElement(
			ui.Insets{Top: 8},
			ui.FixedHeightElement(
				48,
				ui.FillWidthElement(
					ui.ContainerElement(
						ui.Style{Background: ui.NRGBA(255, 152, 0, 255), Padding: ui.All(8), Radius: 6},
						ui.TextElement("FixedHeight + FillWidth", ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(12)),
					),
				),
			),
		),
		ui.PaddingElement(
			ui.Insets{Top: 8},
			ui.RowElement(
				ui.FlexedElement(1, docsSizingPill("Flexed 1", ui.NRGBA(99, 102, 241, 255))),
				ui.HSpacerElement(8),
				ui.FlexedElement(2, docsSizingPill("Flexed 2", ui.NRGBA(20, 184, 166, 255))),
			),
		),
	)
}

func docsColorBlock(label string, bg color.NRGBA) ui.Element {
	return ui.ContainerElement(
		ui.Style{Background: bg, Padding: ui.All(8), Radius: 6},
		ui.TextElement(label, ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
	)
}

func docsDecorationStateDots(th *ui.Theme) ui.Element {
	return ui.RowElement(
		docsStateDot("HoverBg", ui.HoverBg(th.Colors.PrimaryContainer)),
		ui.HSpacerElement(8),
		docsStateDot("PressedBg", ui.PressedBg(th.Colors.SecondaryContainer)),
	)
}

func docsStateDot(label string, deco ui.Decoration) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(ui.NRGBA(248, 250, 252, 255)).
			WithPad(ui.Symmetric(5, 8)).
			WithRad(999).
			Merge(deco),
		ui.TextElement(label, ui.TextSize(10), ui.TextColor(ui.NRGBA(51, 65, 85, 255))),
	)
}

func docsBorderSample(label string, deco ui.Decoration, color color.NRGBA) ui.Element {
	return ui.ContainerDecorationElement(
		deco,
		ui.TextElement(label, ui.TextColor(color)),
	)
}

func docsSizingPill(label string, bg color.NRGBA) ui.Element {
	return ui.ContainerElement(
		ui.Style{Background: bg, Padding: ui.All(8), Radius: 999},
		ui.CenterElement(ui.TextElement(label, ui.TextSize(12), ui.TextColor(ui.NRGBA(255, 255, 255, 255)))),
	)
}
