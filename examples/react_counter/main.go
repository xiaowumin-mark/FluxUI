package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.RunElement(Counter, ui.Title("FluxUI React Counter"), ui.Size(420, 240))
}

func Counter(ctx *ui.Context) ui.Element {
	count := ui.UseState(ctx, 0)
	th := ui.UseTheme(ctx)
	fmt.Println("test")

	return ui.ContainerElement(
		ui.Style{
			Background: th.Surface,
			Padding:    ui.All(20),
		},
		ui.ColumnElement(
			ui.PaddingElement(
				ui.All(4),
				ui.TextElement("FluxUI React-style Counter", ui.TextSize(24)),
			),
			ui.PaddingElement(
				ui.All(4),
				ui.TextElement(fmt.Sprintf("Current value: %d", count.Value()), ui.TextSize(18)),
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
