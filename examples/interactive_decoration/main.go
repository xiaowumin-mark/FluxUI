package main

import (
	"image/color"
	"strings"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.RunElement(App, ui.Title("交互状态示例"), ui.Size(560, 760))
}

func App(ctx *ui.Context) ui.Element {
	currentBg := ui.UseState(ctx, "无事件")
	eventLog := ui.UseState(ctx, []string{})

	addEvent := func(e string) {
		currentBg.Set(e)
		log := eventLog.Value()
		log = append(log, e)
		if len(log) > 8 {
			log = log[len(log)-8:]
		}
		eventLog.Set(log)
	}

	hoverEntered := func(c color.NRGBA) ui.Decoration {
		return ui.Bg(c).WithPad(ui.All(14)).WithRad(10)
	}

	return ui.FromWidget(ui.ContainerDecoration(
		ui.Bg(color.NRGBA{R: 248, G: 250, B: 252, A: 255}).WithPad(ui.All(16)),
		ui.Column(
			ui.Text("FluxUI 交互状态 (React API)", ui.TextSize(20), ui.TextAlign(ui.AlignCenter)),
			ui.VSpacer(2),
			ui.Text("事件: "+currentBg.Value(), ui.TextSize(13), ui.TextColor(color.NRGBA{R: 148, G: 163, B: 184, A: 255})),

			ui.VSpacer(8),
			ui.Text("Hover 背景变化", ui.TextSize(15)),

			ui.Row(
				ui.Padding(ui.All(6),
					ui.ContainerDecoration(
						ui.Bg(color.NRGBA{R: 49, G: 107, B: 255, A: 255}).WithHover(hoverEntered(color.NRGBA{R: 30, G: 64, B: 175, A: 255})).WithPad(ui.All(14)).WithRad(10),
						ui.Text("Hover me", ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
						ui.OnDecoHoverEnter(func(ctx *ui.Context) { addEvent("进入蓝色") }),
						ui.OnDecoHoverLeave(func(ctx *ui.Context) { addEvent("离开蓝色") }),
					),
				),
				ui.Padding(ui.All(6),
					ui.ContainerDecoration(
						ui.Bg(color.NRGBA{R: 40, G: 167, B: 69, A: 255}).WithHover(hoverEntered(color.NRGBA{R: 20, G: 130, B: 45, A: 255})).WithPad(ui.All(14)).WithRad(10),
						ui.Text("Hover me", ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
						ui.OnDecoHover(func(ctx *ui.Context, hovering bool) {
							if hovering {
								addEvent("悬停绿色")
							} else {
								addEvent("离开绿色")
							}
						}),
					),
				),
			),

			ui.VSpacer(8),
			ui.Text("Pressed 反馈", ui.TextSize(15)),

			ui.Row(
				ui.Padding(ui.All(6),
					ui.ContainerDecoration(
						ui.Bg(color.NRGBA{R: 220, G: 53, B: 69, A: 255}).WithPressed(ui.Bg(color.NRGBA{R: 170, G: 30, B: 45, A: 255}).WithPad(ui.All(14)).WithRad(10)).WithPad(ui.All(14)).WithRad(10),
						ui.Text("Press me", ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
						ui.OnDecoPressed(func(ctx *ui.Context, pressed bool) {
							if pressed {
								addEvent("按下红色")
							} else {
								addEvent("释放红色")
							}
						}),
					),
				),
				ui.Padding(ui.All(6),
					ui.ContainerDecoration(
						ui.Bg(color.NRGBA{R: 255, G: 152, B: 0, A: 255}).WithPressed(ui.Bg(color.NRGBA{R: 200, G: 110, B: 0, A: 255}).WithPad(ui.All(14)).WithRad(10)).WithPad(ui.All(14)).WithRad(10),
						ui.Text("Press me", ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
					),
				),
			),

			ui.VSpacer(8),
			ui.Text("Click 事件", ui.TextSize(15)),

			ui.Row(
				ui.Padding(ui.All(6),
					ui.ContainerDecoration(
						ui.Bg(color.NRGBA{R: 33, G: 150, B: 243, A: 255}).WithHover(hoverEntered(color.NRGBA{R: 20, G: 110, B: 190, A: 255})).WithPad(ui.All(14)).WithRad(10),
						ui.Text("Click me", ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
						ui.OnDecoClick(func(ctx *ui.Context) { addEvent("点击了蓝色") }),
					),
				),
				ui.Padding(ui.All(6),
					ui.ContainerDecoration(
						ui.Bg(color.NRGBA{R: 76, G: 175, B: 80, A: 255}).WithHover(hoverEntered(color.NRGBA{R: 50, G: 130, B: 55, A: 255})).WithPad(ui.All(14)).WithRad(10),
						ui.Text("Click me", ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
						ui.OnDecoClick(func(ctx *ui.Context) { addEvent("点击了绿色") }),
					),
				),
			),

			ui.VSpacer(12),
			ui.Text("装饰卡片 (Hover + Click)", ui.TextSize(15)),

			ui.Row(
				ui.Padding(ui.All(6),
					ui.ContainerDecoration(
						ui.Bg(color.NRGBA{R: 255, G: 255, B: 255, A: 255}).WithHover(ui.Bg(color.NRGBA{R: 230, G: 240, B: 255, A: 255}).WithPad(ui.All(12)).WithRad(8)).WithPad(ui.All(12)).WithRad(8).
							WithBorder(ui.Border{Width: 1, Color: color.NRGBA{R: 203, G: 213, B: 225, A: 255}}).Merge(ui.Elevation(1)),
						ui.Column(
							ui.Text("卡片 1", ui.TextSize(14)),
							ui.Text("Hover+Click", ui.TextSize(11), ui.TextColor(color.NRGBA{R: 148, G: 163, B: 184, A: 255})),
						),
						ui.OnDecoClick(func(ctx *ui.Context) { addEvent("卡片1 clicked") }),
						ui.OnDecoHoverEnter(func(ctx *ui.Context) { addEvent("卡片1 hover") }),
					),
				),
				ui.Padding(ui.All(6),
					ui.ContainerDecoration(
						ui.Bg(color.NRGBA{R: 255, G: 255, B: 255, A: 255}).WithHover(ui.Bg(color.NRGBA{R: 230, G: 255, B: 235, A: 255}).WithPad(ui.All(12)).WithRad(8)).WithPad(ui.All(12)).WithRad(8).
							WithBorder(ui.Border{Width: 1, Color: color.NRGBA{R: 203, G: 213, B: 225, A: 255}}).Merge(ui.Elevation(1)),
						ui.Column(
							ui.Text("卡片 2", ui.TextSize(14)),
							ui.Text("Hover+Click", ui.TextSize(11), ui.TextColor(color.NRGBA{R: 148, G: 163, B: 184, A: 255})),
						),
						ui.OnDecoClick(func(ctx *ui.Context) { addEvent("卡片2 clicked") }),
						ui.OnDecoHoverEnter(func(ctx *ui.Context) { addEvent("卡片2 hover") }),
					),
				),
			),

			ui.VSpacer(8),
			ui.Text("Disabled 状态", ui.TextSize(15)),

			ui.Row(
				ui.Padding(ui.All(6),
					ui.ContainerDecoration(
						ui.Bg(color.NRGBA{R: 158, G: 158, B: 158, A: 255}).WithPad(ui.All(14)).WithRad(10),
						ui.Text("已禁用", ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
						ui.ContainerDecorationDisabled(true),
					),
				),
				ui.Padding(ui.All(6),
					ui.ContainerDecoration(
						ui.Bg(color.NRGBA{R: 100, G: 100, B: 100, A: 255}).WithPad(ui.All(14)).WithRad(10),
						ui.Text("纯视觉(无交互)", ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
					),
				),
			),

			ui.VSpacer(12),
			ui.Text("便捷构造器: HoverBg / PressedBg", ui.TextSize(15)),

			ui.Row(
				ui.Padding(ui.All(6),
					ui.ContainerDecoration(
						ui.Bg(color.NRGBA{R: 103, G: 58, B: 183, A: 255}).WithPad(ui.All(12)).WithRad(8).
							WithHover(ui.HoverBg(color.NRGBA{R: 80, G: 40, B: 150, A: 255})),
						ui.Text("HoverBg", ui.TextSize(13), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
					),
				),
				ui.Padding(ui.All(6),
					ui.ContainerDecoration(
						ui.Bg(color.NRGBA{R: 0, G: 150, B: 136, A: 255}).WithPad(ui.All(12)).WithRad(8).
							WithPressed(ui.PressedBg(color.NRGBA{R: 0, G: 110, B: 100, A: 255})),
						ui.Text("PressedBg", ui.TextSize(13), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
					),
				),
			),

			ui.VSpacer(12),
			ui.Text("事件日志:", ui.TextSize(13)),
			ui.ContainerDecoration(
				ui.Bg(color.NRGBA{R: 255, G: 255, B: 255, A: 255}).WithPad(ui.All(8)).WithRad(6).
					WithBorder(ui.Border{Width: 1, Color: color.NRGBA{R: 203, G: 213, B: 225, A: 255}}),
				ui.Text(strings.Join(eventLog.Value(), " → "), ui.TextSize(11), ui.TextColor(color.NRGBA{R: 71, G: 85, B: 105, A: 255})),
			),

			ui.VSpacer(16),
			ui.Text("Phase 7 · React API (RunElement)", ui.TextSize(12), ui.TextColor(color.NRGBA{R: 148, G: 163, B: 184, A: 255})),
		),
	))
}
