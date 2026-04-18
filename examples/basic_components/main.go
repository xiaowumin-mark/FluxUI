package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)
		blue := ui.NRGBA(33, 133, 209, 255)
		green := ui.NRGBA(40, 167, 69, 255)
		orange := ui.NRGBA(255, 193, 7, 255)
		red := ui.NRGBA(220, 53, 69, 255)
		purple := ui.NRGBA(155, 89, 182, 255)

		return ui.Container(
			ui.Style{
				Background: th.Surface,
				Padding:    ui.All(20),
			},
			ui.Column(
				ui.Padding(
					ui.All(8),
					ui.Text("基础组件示例", ui.TextSize(24), ui.TextAlign(ui.AlignCenter)),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("Button、Text、Container 组件的用法", ui.TextSize(14), ui.TextColor(th.SurfaceMuted)),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("Button 按钮组件", ui.TextSize(16), ui.TextColor(th.TextColor)),
				),
				ui.Row(
					ui.Padding(ui.All(4), ui.Button(ui.Text("默认按钮"))),
					ui.Padding(ui.All(4), ui.Button(ui.Text("点击我"), ui.OnClick(func(ctx *ui.Context) {
						fmt.Println("按钮被点击!")
					}))),
					ui.Padding(ui.All(4), ui.Button(ui.Text("禁用按钮"), ui.Disabled(true))),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("自定义样式按钮", ui.TextSize(16), ui.TextColor(th.TextColor)),
				),
				ui.Row(
					ui.Padding(ui.All(4), ui.Button(
						ui.Text("主要按钮"),
						ui.ButtonBackground(th.Primary),
						ui.ButtonForeground(th.TextOnPrimary),
					)),
					ui.Padding(ui.All(4), ui.Button(
						ui.Text("成功按钮"),
						ui.ButtonBackground(green),
						ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
					)),
					ui.Padding(ui.All(4), ui.Button(
						ui.Text("警告按钮"),
						ui.ButtonBackground(orange),
						ui.ButtonForeground(ui.NRGBA(0, 0, 0, 255)),
					)),
					ui.Padding(ui.All(4), ui.Button(
						ui.Text("危险按钮"),
						ui.ButtonBackground(red),
						ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
					)),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("带内边距和圆角的按钮", ui.TextSize(16), ui.TextColor(th.TextColor)),
				),
				ui.Row(
					ui.Padding(ui.All(4), ui.Button(
						ui.Text("圆角按钮"),
						ui.ButtonRadius(20),
						ui.ButtonPadding(ui.Symmetric(12, 24)),
					)),
					ui.Padding(ui.All(4), ui.Button(
						ui.Text("方形按钮"),
						ui.ButtonRadius(0),
					)),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("多颜色 Text 组件", ui.TextSize(16), ui.TextColor(th.TextColor)),
				),
				ui.Column(
					ui.Padding(ui.All(4), ui.Text("默认文本颜色", ui.TextSize(14))),
					ui.Padding(ui.All(4), ui.Text("蓝色文本", ui.TextSize(14), ui.TextColor(blue))),
					ui.Padding(ui.All(4), ui.Text("绿色文本", ui.TextSize(14), ui.TextColor(green))),
					ui.Padding(ui.All(4), ui.Text("橙色文本", ui.TextSize(14), ui.TextColor(orange))),
					ui.Padding(ui.All(4), ui.Text("红色文本", ui.TextSize(14), ui.TextColor(red))),
					ui.Padding(ui.All(4), ui.Text("紫色文本", ui.TextSize(14), ui.TextColor(purple))),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("Container 容器组件", ui.TextSize(16), ui.TextColor(th.TextColor)),
				),
				ui.Row(
					ui.Padding(
						ui.All(4),
						ui.Container(
							ui.Style{
								Background: th.Primary,
								Padding:    ui.All(16),
								Radius:     8,
							},
							ui.Text("Primary", ui.TextColor(th.TextOnPrimary)),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Container(
							ui.Style{
								Background: blue,
								Padding:    ui.All(16),
								Radius:     8,
							},
							ui.Text("Blue", ui.TextColor(th.TextOnPrimary)),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Container(
							ui.Style{
								Background: green,
								Padding:    ui.All(16),
								Radius:     8,
							},
							ui.Text("Green", ui.TextColor(th.TextOnPrimary)),
						),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("Padding 内边距组件", ui.TextSize(16), ui.TextColor(th.TextColor)),
				),
				ui.Container(
					ui.Style{
						Background: th.SurfaceMuted,
						Padding:    ui.All(4),
						Radius:     4,
					},
					ui.Padding(
						ui.All(16),
						ui.Container(
							ui.Style{
								Background: purple,
								Padding:    ui.All(8),
								Radius:     4,
							},
							ui.Text("嵌套内边距", ui.TextColor(th.TextOnPrimary)),
						),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("基础组件示例完成", ui.TextSize(14), ui.TextColor(th.SurfaceMuted)),
				),
			),
		)
	}, ui.Title("基础组件示例"), ui.Size(480, 820))
}
