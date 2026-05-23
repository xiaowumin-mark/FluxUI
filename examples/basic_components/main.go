package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

var (
	blue   = ui.NRGBA(33, 133, 209, 255)
	green  = ui.NRGBA(40, 167, 69, 255)
	orange = ui.NRGBA(255, 193, 7, 255)
	red    = ui.NRGBA(220, 53, 69, 255)
	purple = ui.NRGBA(155, 89, 182, 255)
)

func App(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	hovering := ui.UseState(ctx, false)

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(20)),
		ui.ColumnElement(
			ui.TextElement("基础组件示例", ui.TextSize(24), ui.TextAlign(ui.AlignCenter)),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("Button、Text、Container 组件的用法", ui.TextSize(14), ui.TextColor(th.SurfaceMuted)),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("Button 按钮组件", ui.TextSize(16), ui.TextColor(th.TextColor)),
			),
			ui.RowElement(
				ui.PaddingElement(ui.All(4), ui.ButtonElement(ui.TextElement("默认按钮"))),
				ui.PaddingElement(ui.All(4), ui.ButtonElement(
					ui.TextElement("点击我"),
					ui.OnClick(func(ctx *ui.Context) {
						fmt.Println("按钮被点击!")
					}),
					ui.OnHover(func(ctx *ui.Context, value bool) {
						hovering.Set(value)
					}),
				)),
				ui.PaddingElement(ui.All(4), ui.ButtonElement(
					ui.TextElement("禁用按钮"),
					ui.Disabled(true),
				)),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement(fmt.Sprintf("点击我按钮悬浮状态: %v", hovering.Value()), ui.TextSize(13), ui.TextColor(th.SurfaceMuted)),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("自定义样式按钮", ui.TextSize(16), ui.TextColor(th.TextColor)),
			),
			ui.RowElement(
				ui.PaddingElement(ui.All(4), ui.ButtonElement(
					ui.TextElement("主要按钮"),
					ui.ButtonBackground(th.Primary),
					ui.ButtonForeground(th.TextOnPrimary),
				)),
				ui.PaddingElement(ui.All(4), ui.ButtonElement(
					ui.TextElement("成功按钮"),
					ui.ButtonBackground(green),
					ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
				)),
				ui.PaddingElement(ui.All(4), ui.ButtonElement(
					ui.TextElement("警告按钮"),
					ui.ButtonBackground(orange),
					ui.ButtonForeground(ui.NRGBA(0, 0, 0, 255)),
				)),
				ui.PaddingElement(ui.All(4), ui.ButtonElement(
					ui.TextElement("危险按钮"),
					ui.ButtonBackground(red),
					ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
				)),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("带内边距和圆角的按钮", ui.TextSize(16), ui.TextColor(th.TextColor)),
			),
			ui.RowElement(
				ui.PaddingElement(ui.All(4), ui.ButtonElement(
					ui.TextElement("圆角按钮"),
					ui.ButtonRadius(20),
					ui.ButtonPadding(ui.Symmetric(12, 24)),
				)),
				ui.PaddingElement(ui.All(4), ui.ButtonElement(
					ui.TextElement("方形按钮"),
					ui.ButtonRadius(0),
				)),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("多颜色 Text 组件", ui.TextSize(16), ui.TextColor(th.TextColor)),
			),
			ui.ColumnElement(
				ui.PaddingElement(ui.All(4), ui.TextElement("默认文本颜色", ui.TextSize(14))),
				ui.PaddingElement(ui.All(4), ui.TextElement("蓝色文本", ui.TextSize(14), ui.TextColor(blue))),
				ui.PaddingElement(ui.All(4), ui.TextElement("绿色文本", ui.TextSize(14), ui.TextColor(green))),
				ui.PaddingElement(ui.All(4), ui.TextElement("橙色文本", ui.TextSize(14), ui.TextColor(orange))),
				ui.PaddingElement(ui.All(4), ui.TextElement("红色文本", ui.TextSize(14), ui.TextColor(red))),
				ui.PaddingElement(ui.All(4), ui.TextElement("紫色文本", ui.TextSize(14), ui.TextColor(purple))),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("Container 容器组件", ui.TextSize(16), ui.TextColor(th.TextColor)),
			),
			ui.RowElement(
				ui.PaddingElement(
					ui.All(4),
					ui.ContainerDecorationElement(
						ui.Bg(th.Primary).WithPad(ui.All(16)).WithRad(8),
						ui.TextElement("Primary", ui.TextColor(th.TextOnPrimary)),
					),
				),
				ui.PaddingElement(
					ui.All(4),
					ui.ContainerDecorationElement(
						ui.Bg(blue).WithPad(ui.All(16)).WithRad(8),
						ui.TextElement("Blue", ui.TextColor(th.TextOnPrimary)),
					),
				),
				ui.PaddingElement(
					ui.All(4),
					ui.ContainerDecorationElement(
						ui.Bg(green).WithPad(ui.All(16)).WithRad(8),
						ui.TextElement("Green", ui.TextColor(th.TextOnPrimary)),
					),
				),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("Padding 内边距组件", ui.TextSize(16), ui.TextColor(th.TextColor)),
			),
			ui.ContainerDecorationElement(
				ui.Bg(th.SurfaceMuted).WithPad(ui.All(4)).WithRad(4),
				ui.PaddingElement(
					ui.All(16),
					ui.ContainerDecorationElement(
						ui.Bg(purple).WithPad(ui.All(8)).WithRad(4),
						ui.TextElement("嵌套内边距", ui.TextColor(th.TextOnPrimary)),
					),
				),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("基础组件示例完成", ui.TextSize(14), ui.TextColor(th.SurfaceMuted)),
			),
		),
	)
}

func main() {
	_ = ui.RunElement(App, ui.Title("基础组件示例"), ui.Size(480, 820))
}
