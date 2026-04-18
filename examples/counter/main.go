package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		count := ui.State[int](ctx)
		th := ui.UseTheme(ctx)

		return ui.Container(
			ui.Style{
				Background: th.Surface,
				Padding:    ui.All(20),
			},
			ui.Column(
				ui.Padding(ui.All(4), ui.Text("FluxUI Counter", ui.TextSize(24))),
				ui.Padding(
					ui.All(4),
					ui.Text(fmt.Sprintf("当前值: %d", count.Value()), ui.TextSize(18)),
				),
				ui.Padding(
					ui.All(4),
					ui.Button(
						ui.Text("+1"),
						ui.OnClick(func(ctx *ui.Context) {
							count.Set(count.Value() + 1)
						}),
					),
				),
			),
		)
	}, ui.Title("FluxUI Counter"), ui.Size(420, 220))
}
