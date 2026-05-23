package main

import (
	"fmt"
	"image/color"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

const (
	listItemCount = 50_000
	gridItemCount = 100_000
	gridColumns   = 4
)

func App(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	activeTab := ui.UseState(ctx, "list")

	tabs := ui.TabsElement(
		activeTab.Value(),
		[]ui.TabItem{
			{Key: "list", Label: fmt.Sprintf("ListView (%d)", listItemCount)},
			{Key: "grid", Label: fmt.Sprintf("GridView (%d)", gridItemCount)},
		},
		ui.TabsOnChange(func(ctx *ui.Context, key string) { activeTab.Set(key) }),
	)

	var content ui.Element
	switch activeTab.Value() {
	case "grid":
		content = gridDemo(th)
	default:
		content = listDemo(th)
	}

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(16)),
		ui.ColumnElement(
			ui.TextElement("Virtual Scroll Demo", ui.TextSize(22)),
			ui.SpacerElement(0, 4),
			ui.TextElement("大数据量虚拟滚动，只渲染可见区域", ui.TextSize(13), ui.TextColor(th.SurfaceMuted)),
			ui.SpacerElement(0, 12),
			tabs,
			ui.SpacerElement(0, 12),
			ui.ExpandedElement(content),
		),
	)
}

func listDemo(th *ui.Theme) ui.Element {
	return ui.ListViewElement(
		listItemCount,
		func(ctx *ui.Context, index int) ui.Element {
			bg := color.NRGBA{R: 240, G: 242, B: 245, A: 255}
			if index%2 == 1 {
				bg = color.NRGBA{R: 250, G: 250, B: 252, A: 255}
			}
			return ui.ContainerDecorationElement(
				ui.Bg(bg).WithPad(ui.Symmetric(8, 12)).WithRad(6),
				ui.RowElement(
					ui.ContainerDecorationElement(
						ui.Bg(color.NRGBA{R: uint8(59 + index%60), G: 130, B: 246, A: 255}).WithPad(ui.Symmetric(4, 8)).WithRad(4),
						ui.TextElement(fmt.Sprintf("#%d", index), ui.TextSize(11),
							ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
					),
					ui.SpacerElement(12, 0),
					ui.ColumnElement(
						ui.TextElement(fmt.Sprintf("Item %d", index), ui.TextSize(14)),
						ui.TextElement(fmt.Sprintf("This is list item number %d of %d", index, listItemCount),
							ui.TextSize(11), ui.TextColor(th.SurfaceMuted)),
					),
				),
			)
		},
		ui.ListItemSpacing(2),
	)
}

func gridDemo(th *ui.Theme) ui.Element {
	return ui.GridViewElement(
		gridItemCount,
		gridColumns,
		func(ctx *ui.Context, index int) ui.Element {
			r := uint8((index*7)%200 + 55)
			g := uint8((index*13)%180 + 75)
			b := uint8((index*3)%160 + 95)
			bg := color.NRGBA{R: r, G: g, B: b, A: 255}

			return ui.ContainerDecorationElement(
				ui.Bg(bg).WithPad(ui.All(10)).WithRad(8),
				ui.ColumnElement(
					ui.TextElement(fmt.Sprintf("#%d", index), ui.TextSize(13),
						ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
					ui.SpacerElement(0, 2),
					ui.TextElement(fmt.Sprintf("Cell %d", index), ui.TextSize(11),
						ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 200})),
				),
			)
		},
		ui.GridGap(4, 4),
	)
}

func main() {
	_ = ui.RunElement(App, ui.Title("Virtual Scroll Demo"), ui.Size(800, 700))
}
