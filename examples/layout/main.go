package main

import (
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)

		blue := ui.NRGBA(33, 133, 209, 255)
		green := ui.NRGBA(34, 153, 84, 255)
		orange := ui.NRGBA(243, 156, 18, 255)
		purple := ui.NRGBA(155, 89, 182, 255)

		return ui.Container(
			ui.Style{
				Background: th.Surface,
				Padding:    ui.All(20),
			},
			ui.Column(
				ui.Padding(
					ui.All(8),
					ui.Text("布局示例", ui.TextSize(24), ui.TextAlign(ui.AlignCenter)),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("Column 纵向布局 - 垂直排列子组件", ui.TextSize(16), ui.TextColor(th.TextColor)),
				),
				ui.Container(
					ui.Style{
						Background: th.Primary,
						Padding:    ui.All(10),
						Radius:     8,
					},
					ui.Column(
						ui.Padding(ui.All(4), ui.Text("第一行", ui.TextColor(th.TextOnPrimary))),
						ui.Padding(ui.All(4), ui.Text("第二行", ui.TextColor(th.TextOnPrimary))),
						ui.Padding(ui.All(4), ui.Text("第三行", ui.TextColor(th.TextOnPrimary))),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("Row 横向布局 - 水平排列子组件", ui.TextSize(16), ui.TextColor(th.TextColor)),
				),
				ui.Container(
					ui.Style{
						Background: blue,
						Padding:    ui.All(10),
						Radius:     8,
					},
					ui.Row(
						ui.Padding(ui.All(4), ui.Text("左", ui.TextColor(th.TextOnPrimary))),
						ui.Padding(ui.All(4), ui.Text("中", ui.TextColor(th.TextOnPrimary))),
						ui.Padding(ui.All(4), ui.Text("右", ui.TextColor(th.TextOnPrimary))),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("嵌套布局 - 复杂的组合布局", ui.TextSize(16), ui.TextColor(th.TextColor)),
				),
				ui.Container(
					ui.Style{
						Background: th.SurfaceMuted,
						Padding:    ui.All(10),
						Radius:     8,
					},
					ui.Column(
						ui.Row(
							ui.Padding(
								ui.All(4),
								ui.Container(
									ui.Style{
										Background: th.Primary,
										Padding:    ui.All(8),
										Radius:     4,
									},
									ui.Text("卡片1", ui.TextColor(th.TextOnPrimary)),
								),
							),
							ui.Padding(
								ui.All(4),
								ui.Container(
									ui.Style{
										Background: blue,
										Padding:    ui.All(8),
										Radius:     4,
									},
									ui.Text("卡片2", ui.TextColor(th.TextOnPrimary)),
								),
							),
						),
						ui.Row(
							ui.Padding(
								ui.All(4),
								ui.Container(
									ui.Style{
										Background: green,
										Padding:    ui.All(8),
										Radius:     4,
									},
									ui.Text("卡片3", ui.TextColor(th.TextOnPrimary)),
								),
							),
							ui.Padding(
								ui.All(4),
								ui.Container(
									ui.Style{
										Background: th.Primary,
										Padding:    ui.All(8),
										Radius:     4,
									},
									ui.Text("卡片4", ui.TextColor(th.TextOnPrimary)),
								),
							),
						),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("对称边距示例", ui.TextSize(16), ui.TextColor(th.TextColor)),
				),
				ui.Row(
					ui.Padding(
						ui.Symmetric(16, 8),
						ui.Container(
							ui.Style{
								Background: th.Primary,
								Padding:    ui.All(4),
								Radius:     4,
							},
							ui.Text("上下16 左右8", ui.TextColor(th.TextOnPrimary)),
						),
					),
					ui.Padding(
						ui.Symmetric(8, 16),
						ui.Container(
							ui.Style{
								Background: green,
								Padding:    ui.All(4),
								Radius:     4,
							},
							ui.Text("上下8 左右16", ui.TextColor(th.TextOnPrimary)),
						),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("多颜色卡片网格", ui.TextSize(16), ui.TextColor(th.TextColor)),
				),
				ui.Column(
					ui.Row(
						ui.Padding(ui.All(4), ui.Container(
							ui.Style{Background: th.Primary, Padding: ui.All(12), Radius: 6},
							ui.Text("Primary", ui.TextColor(th.TextOnPrimary), ui.TextSize(14)),
						)),
						ui.Padding(ui.All(4), ui.Container(
							ui.Style{Background: blue, Padding: ui.All(12), Radius: 6},
							ui.Text("Blue", ui.TextColor(th.TextOnPrimary), ui.TextSize(14)),
						)),
						ui.Padding(ui.All(4), ui.Container(
							ui.Style{Background: green, Padding: ui.All(12), Radius: 6},
							ui.Text("Green", ui.TextColor(th.TextOnPrimary), ui.TextSize(14)),
						)),
					),
					ui.Row(
						ui.Padding(ui.All(4), ui.Container(
							ui.Style{Background: orange, Padding: ui.All(12), Radius: 6},
							ui.Text("Orange", ui.TextColor(th.TextOnPrimary), ui.TextSize(14)),
						)),
						ui.Padding(ui.All(4), ui.Container(
							ui.Style{Background: purple, Padding: ui.All(12), Radius: 6},
							ui.Text("Purple", ui.TextColor(th.TextOnPrimary), ui.TextSize(14)),
						)),
						ui.Padding(ui.All(4), ui.Container(
							ui.Style{Background: th.SurfaceMuted, Padding: ui.All(12), Radius: 6},
							ui.Text("Muted", ui.TextColor(th.TextColor), ui.TextSize(14)),
						)),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("布局示例完成", ui.TextSize(14), ui.TextColor(th.SurfaceMuted)),
				),
			),
		)
	}, ui.Title("布局示例"), ui.Size(480, 750))
}
