package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func App(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	scrollX := ui.UseState(ctx, float32(0))
	scrollY := ui.UseState(ctx, float32(0))

	longItems := make([]ui.Element, 0, 16)
	for i := 0; i < 16; i++ {
		longItems = append(longItems,
			ui.PaddingElement(
				ui.Insets{Right: 10},
				ui.CardElement(
					ui.ContainerDecorationElement(
						ui.Bg(ui.NRGBA(239, 246, 255, 255)).WithPad(ui.Symmetric(10, 14)).WithRad(10),
						ui.TextElement(
							fmt.Sprintf("Card #%02d - This is a horizontal item with long text %d", i+1, i+1),
							ui.TextSize(13),
						),
					),
					ui.CardBorder(ui.NRGBA(203, 213, 225, 255), 1),
				),
			),
		)
	}

	lines := make([]ui.Element, 0, 22)
	for i := 0; i < 22; i++ {
		lines = append(lines,
			ui.PaddingElement(
				ui.Insets{Bottom: 8},
				ui.RowElement(
					ui.TextElement(fmt.Sprintf("Line %02d", i+1), ui.TextSize(13)),
					ui.SpacerElement(20, 0),
					ui.TextElement("This row is intentionally long to force horizontal scrolling.", ui.TextSize(13)),
					ui.SpacerElement(20, 0),
					ui.TextElement("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", ui.TextSize(13), ui.TextColor(ui.NRGBA(37, 99, 235, 255))),
					ui.SpacerElement(20, 0),
					ui.TextElement("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB", ui.TextSize(13), ui.TextColor(ui.NRGBA(22, 163, 74, 255))),
				),
			),
		)
	}

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(16)),
		ui.ColumnElement(
			ui.TextElement("Horizontal Scroll Example", ui.TextSize(22)),
			ui.SpacerElement(0, 6),
			ui.PaddingElement(
				ui.Insets{Top: 6},
				ui.TextElement(
					fmt.Sprintf("Scroll offset x=%.2f, y=%.2f", scrollX.Value(), scrollY.Value()),
					ui.TextSize(13),
					ui.TextColor(ui.NRGBA(71, 85, 105, 255)),
				),
			),
			ui.SpacerElement(0, 12),
			ui.TextElement("Case 1: Single long row (horizontal only)", ui.TextSize(14)),
			ui.SpacerElement(0, 8),
			ui.FixedHeightElement(
				110,
				ui.ScrollViewElement(
					ui.RowElement(longItems...),
					ui.ScrollHorizontal(true),
					ui.ScrollVertical(false),
					ui.ScrollOnChange(func(ctx *ui.Context, x, y float32) {
						scrollX.Set(x)
						scrollY.Set(y)
					}),
				),
			),
			ui.SpacerElement(0, 16),
			ui.TextElement("Case 2: Multi-line long content (horizontal only)", ui.TextSize(14)),
			ui.SpacerElement(0, 8),
			ui.FixedHeightElement(
				280,
				ui.ContainerDecorationElement(
					ui.Bg(ui.NRGBA(248, 250, 252, 255)).WithPad(ui.All(10)).WithRad(10),
					ui.ScrollViewElement(
						ui.ColumnElement(lines...),
						ui.ScrollHorizontal(true),
						ui.ScrollVertical(false),
						ui.ScrollOnChange(func(ctx *ui.Context, x, y float32) {
							scrollX.Set(x)
							scrollY.Set(y)
						}),
					),
				),
			),
			ui.SpacerElement(0, 12),
			ui.TextElement("请拖动底部横向滚动条，观察是否出现拉伸/撕裂形变。", ui.TextSize(12), ui.TextColor(ui.NRGBA(100, 116, 139, 255))),
		),
	)
}

func main() {
	_ = ui.RunElement(App, ui.Title("FluxUI Horizontal Scroll"), ui.Size(1200, 760))
}
