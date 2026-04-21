package main

import (
	"fmt"
	"image/color"

	"github.com/xiaowumin-mark/FluxUI/ui"
)

const (
	listItemCount = 50_000
	gridItemCount = 100_000
	gridColumns   = 4
)

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)
		activeTab := ui.State[string](ctx)
		if activeTab.Value() == "" {
			activeTab.Set("list")
		}

		tabs := ui.Tabs(
			activeTab.Value(),
			[]ui.TabItem{
				{Key: "list", Label: fmt.Sprintf("ListView (%d)", listItemCount)},
				{Key: "grid", Label: fmt.Sprintf("GridView (%d)", gridItemCount)},
			},
			ui.TabsOnChange(func(ctx *ui.Context, key string) {
				activeTab.Set(key)
			}),
		)

		var content ui.Widget
		switch activeTab.Value() {
		case "grid":
			content = gridDemo(ctx.Scope("grid"), th)
		default:
			content = listDemo(ctx.Scope("list"), th)
		}

		return ui.Container(
			ui.Style{Background: th.Surface, Padding: ui.All(16)},
			ui.Column(
				ui.Text("Virtual Scroll Demo", ui.TextSize(22)),
				ui.VSpacer(4),
				ui.Text("大数据量虚拟滚动，只渲染可见区域", ui.TextSize(13), ui.TextColor(th.SurfaceMuted)),
				ui.VSpacer(12),
				tabs,
				ui.VSpacer(12),
				ui.Expanded(content),
			),
		)
	}, ui.Title("Virtual Scroll Demo"), ui.Size(800, 700))
}

func listDemo(ctx *ui.Context, th *ui.Theme) ui.Widget {
	return ui.ListView(
		listItemCount,
		func(ctx *ui.Context, index int) ui.Widget {
			bg := color.NRGBA{R: 240, G: 242, B: 245, A: 255}
			if index%2 == 1 {
				bg = color.NRGBA{R: 250, G: 250, B: 252, A: 255}
			}
			return ui.Container(
				ui.Style{Background: bg, Padding: ui.Symmetric(8, 12), Radius: 6},
				ui.Row(
					ui.Container(
						ui.Style{
							Background: color.NRGBA{R: uint8(59 + index%60), G: 130, B: 246, A: 255},
							Radius:     4,
							Padding:    ui.Symmetric(4, 8),
						},
						ui.Text(fmt.Sprintf("#%d", index), ui.TextSize(11),
							ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
					),
					ui.HSpacer(12),
					ui.Column(
						ui.Text(fmt.Sprintf("Item %d", index), ui.TextSize(14)),
						ui.Text(fmt.Sprintf("This is list item number %d of %d", index, listItemCount),
							ui.TextSize(11), ui.TextColor(th.SurfaceMuted)),
					),
				),
			)
		},
		ui.ListItemSpacing(2),
	)
}

func gridDemo(ctx *ui.Context, th *ui.Theme) ui.Widget {
	return ui.GridView(
		gridItemCount,
		gridColumns,
		func(ctx *ui.Context, index int) ui.Widget {
			r := uint8((index * 7) % 200 + 55)
			g := uint8((index * 13) % 180 + 75)
			b := uint8((index * 3) % 160 + 95)
			bg := color.NRGBA{R: r, G: g, B: b, A: 255}

			return ui.Container(
				ui.Style{Background: bg, Padding: ui.All(10), Radius: 8},
				ui.Column(
					ui.Text(fmt.Sprintf("#%d", index), ui.TextSize(13),
						ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
					ui.VSpacer(2),
					ui.Text(fmt.Sprintf("Cell %d", index), ui.TextSize(11),
						ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 200})),
				),
			)
		},
		ui.GridGap(4, 4),
	)
}
