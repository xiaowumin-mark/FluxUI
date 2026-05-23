package main

import (
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

var (
	blue   = ui.NRGBA(33, 133, 209, 255)
	green  = ui.NRGBA(34, 153, 84, 255)
	orange = ui.NRGBA(243, 156, 18, 255)
	purple = ui.NRGBA(155, 89, 182, 255)
)

func App(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(20)),
		ui.ColumnElement(
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("布局示例", ui.TextSize(24), ui.TextAlign(ui.AlignCenter)),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("Column 纵向布局 - 垂直排列子组件", ui.TextSize(16), ui.TextColor(th.TextColor)),
			),
			ui.ContainerDecorationElement(
				ui.Bg(th.Primary).WithPad(ui.All(10)).WithRad(8),
				ui.ColumnElement(
					ui.PaddingElement(ui.All(4), ui.TextElement("第一行", ui.TextColor(th.TextOnPrimary))),
					ui.PaddingElement(ui.All(4), ui.TextElement("第二行", ui.TextColor(th.TextOnPrimary))),
					ui.PaddingElement(ui.All(4), ui.TextElement("第三行", ui.TextColor(th.TextOnPrimary))),
				),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("Row 横向布局 - 水平排列子组件", ui.TextSize(16), ui.TextColor(th.TextColor)),
			),
			ui.ContainerDecorationElement(
				ui.Bg(blue).WithPad(ui.All(10)).WithRad(8),
				ui.RowElement(
					ui.PaddingElement(ui.All(4), ui.TextElement("左", ui.TextColor(th.TextOnPrimary))),
					ui.PaddingElement(ui.All(4), ui.TextElement("中", ui.TextColor(th.TextOnPrimary))),
					ui.PaddingElement(ui.All(4), ui.TextElement("右", ui.TextColor(th.TextOnPrimary))),
				),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("嵌套布局 - 复杂的组合布局", ui.TextSize(16), ui.TextColor(th.TextColor)),
			),
			ui.ContainerDecorationElement(
				ui.Bg(th.SurfaceMuted).WithPad(ui.All(10)).WithRad(8),
				ui.ColumnElement(
					ui.RowElement(
						ui.PaddingElement(
							ui.All(4),
							ui.ContainerDecorationElement(
								ui.Bg(th.Primary).WithPad(ui.All(8)).WithRad(4),
								ui.TextElement("卡片1", ui.TextColor(th.TextOnPrimary)),
							),
						),
						ui.PaddingElement(
							ui.All(4),
							ui.ContainerDecorationElement(
								ui.Bg(blue).WithPad(ui.All(8)).WithRad(4),
								ui.TextElement("卡片2", ui.TextColor(th.TextOnPrimary)),
							),
						),
					),
					ui.RowElement(
						ui.PaddingElement(
							ui.All(4),
							ui.ContainerDecorationElement(
								ui.Bg(green).WithPad(ui.All(8)).WithRad(4),
								ui.TextElement("卡片3", ui.TextColor(th.TextOnPrimary)),
							),
						),
						ui.PaddingElement(
							ui.All(4),
							ui.ContainerDecorationElement(
								ui.Bg(th.Primary).WithPad(ui.All(8)).WithRad(4),
								ui.TextElement("卡片4", ui.TextColor(th.TextOnPrimary)),
							),
						),
					),
				),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("对称边距示例", ui.TextSize(16), ui.TextColor(th.TextColor)),
			),
			ui.RowElement(
				ui.PaddingElement(
					ui.Symmetric(16, 8),
					ui.ContainerDecorationElement(
						ui.Bg(th.Primary).WithPad(ui.All(4)).WithRad(4),
						ui.TextElement("上下16 左右8", ui.TextColor(th.TextOnPrimary)),
					),
				),
				ui.PaddingElement(
					ui.Symmetric(8, 16),
					ui.ContainerDecorationElement(
						ui.Bg(green).WithPad(ui.All(4)).WithRad(4),
						ui.TextElement("上下8 左右16", ui.TextColor(th.TextOnPrimary)),
					),
				),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("多颜色卡片网格", ui.TextSize(16), ui.TextColor(th.TextColor)),
			),
			ui.ColumnElement(
				ui.RowElement(
					ui.PaddingElement(ui.All(4), ui.ContainerDecorationElement(
						ui.Bg(th.Primary).WithPad(ui.All(12)).WithRad(6),
						ui.TextElement("Primary", ui.TextColor(th.TextOnPrimary), ui.TextSize(14)),
					)),
					ui.PaddingElement(ui.All(4), ui.ContainerDecorationElement(
						ui.Bg(blue).WithPad(ui.All(12)).WithRad(6),
						ui.TextElement("Blue", ui.TextColor(th.TextOnPrimary), ui.TextSize(14)),
					)),
					ui.PaddingElement(ui.All(4), ui.ContainerDecorationElement(
						ui.Bg(green).WithPad(ui.All(12)).WithRad(6),
						ui.TextElement("Green", ui.TextColor(th.TextOnPrimary), ui.TextSize(14)),
					)),
				),
				ui.RowElement(
					ui.PaddingElement(ui.All(4), ui.ContainerDecorationElement(
						ui.Bg(orange).WithPad(ui.All(12)).WithRad(6),
						ui.TextElement("Orange", ui.TextColor(th.TextOnPrimary), ui.TextSize(14)),
					)),
					ui.PaddingElement(ui.All(4), ui.ContainerDecorationElement(
						ui.Bg(purple).WithPad(ui.All(12)).WithRad(6),
						ui.TextElement("Purple", ui.TextColor(th.TextOnPrimary), ui.TextSize(14)),
					)),
					ui.PaddingElement(ui.All(4), ui.ContainerDecorationElement(
						ui.Bg(th.SurfaceMuted).WithPad(ui.All(12)).WithRad(6),
						ui.TextElement("Muted", ui.TextColor(th.TextColor), ui.TextSize(14)),
					)),
				),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("布局示例完成", ui.TextSize(14), ui.TextColor(th.SurfaceMuted)),
			),
		),
	)
}

func main() {
	_ = ui.RunElement(App, ui.Title("布局示例"), ui.Size(480, 750))
}
