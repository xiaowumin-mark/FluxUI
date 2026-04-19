package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)

		scrollX := ui.State[float32](ctx)
		scrollY := ui.State[float32](ctx)

		longItems := make([]ui.Widget, 0, 16)
		for i := 0; i < 16; i++ {
			longItems = append(longItems,
				ui.Padding(
					ui.Insets{Right: 10},
					ui.Card(
						ui.Container(
							ui.Style{
								Background: ui.NRGBA(239, 246, 255, 255),
								Padding:    ui.Symmetric(10, 14),
								Radius:     10,
							},
							ui.Text(
								fmt.Sprintf("Card #%02d - This is a horizontal item with long text %d", i+1, i+1),
								ui.TextSize(13),
							),
						),
						ui.CardBorder(ui.NRGBA(203, 213, 225, 255), 1),
					),
				),
			)
		}

		lines := make([]ui.Widget, 0, 22)
		for i := 0; i < 22; i++ {
			lines = append(lines,
				ui.Padding(
					ui.Insets{Bottom: 8},
					ui.Row(
						ui.Text(fmt.Sprintf("Line %02d", i+1), ui.TextSize(13)),
						ui.HSpacer(20),
						ui.Text("This row is intentionally long to force horizontal scrolling.", ui.TextSize(13)),
						ui.HSpacer(20),
						ui.Text("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", ui.TextSize(13), ui.TextColor(ui.NRGBA(37, 99, 235, 255))),
						ui.HSpacer(20),
						ui.Text("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB", ui.TextSize(13), ui.TextColor(ui.NRGBA(22, 163, 74, 255))),
					),
				),
			)
		}

		return ui.Container(
			ui.Style{
				Background: th.Surface,
				Padding:    ui.All(16),
			},
			ui.Column(
				ui.Text("Horizontal Scroll Example", ui.TextSize(22)),
				ui.Padding(
					ui.Insets{Top: 6},
					ui.Text(
						fmt.Sprintf("Scroll offset x=%.2f, y=%.2f", scrollX.Value(), scrollY.Value()),
						ui.TextSize(13),
						ui.TextColor(ui.NRGBA(71, 85, 105, 255)),
					),
				),
				ui.Padding(
					ui.Insets{Top: 12},
					ui.Text("Case 1: Single long row (horizontal only)", ui.TextSize(14)),
				),
				ui.Padding(
					ui.Insets{Top: 8},
					ui.FixedHeight(
						110,
						ui.ScrollView(
							ui.Row(longItems...),
							ui.ScrollHorizontal(true),
							ui.ScrollVertical(false),
							ui.ScrollOnChange(func(ctx *ui.Context, x, y float32) {
								scrollX.Set(x)
								scrollY.Set(y)
							}),
						),
					),
				),
				ui.Padding(
					ui.Insets{Top: 16},
					ui.Text("Case 2: Multi-line long content (horizontal only)", ui.TextSize(14)),
				),
				ui.Padding(
					ui.Insets{Top: 8},
					ui.FixedHeight(
						280,
						ui.Container(
							ui.Style{
								Background: ui.NRGBA(248, 250, 252, 255),
								Padding:    ui.All(10),
								Radius:     10,
							},
							ui.ScrollView(
								ui.Column(lines...),
								ui.ScrollHorizontal(true),
								ui.ScrollVertical(false),
								ui.ScrollOnChange(func(ctx *ui.Context, x, y float32) {
									scrollX.Set(x)
									scrollY.Set(y)
								}),
							),
						),
					),
				),
				ui.Padding(
					ui.Insets{Top: 12},
					ui.Text("请拖动底部横向滚动条，观察是否出现拉伸/撕裂形变。", ui.TextSize(12), ui.TextColor(ui.NRGBA(100, 116, 139, 255))),
				),
			),
		)
	}, ui.Title("FluxUI Horizontal Scroll"), ui.Size(1200, 760))
}
