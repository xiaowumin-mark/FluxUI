package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func App(ctx *ui.Context) ui.Element {
	count := ui.UseState(ctx, 0)
	th := ui.UseTheme(ctx)

	var label string
	if count.Value() == 0 {
		label = "当前值: 0"
	} else {
		label = fmt.Sprintf("当前值: %d", count.Value())
	}

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(20)),
		ui.ColumnElement(
			ui.TextElement("FluxUI Counter", ui.TextSize(24)),
			ui.PaddingElement(
				ui.All(4),
				ui.TextElement(label, ui.TextSize(18)),
			),
			ui.PaddingElement(
				ui.All(4),
				ui.ButtonElement(
					ui.TextElement("+1"),
					ui.OnClick(func(ctx *ui.Context) {
						count.Set(count.Value() + 1)
					}),
				),
			),
		),
	)
}

func main() {
	_ = ui.RunElement(App, ui.Title("FluxUI Counter"), ui.Size(420, 220))
}
