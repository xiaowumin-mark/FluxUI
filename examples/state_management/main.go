package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)
		count := ui.State[int](ctx)
		message := ui.State[string](ctx)
		items := ui.State[[]string](ctx)

		red := ui.NRGBA(220, 53, 69, 255)
		green := ui.NRGBA(40, 167, 69, 255)

		return ui.Container(
			ui.Style{
				Background: th.Surface,
				Padding:    ui.All(20),
			},
			ui.Column(
				ui.Padding(
					ui.All(8),
					ui.Text("状态管理示例", ui.TextSize(24), ui.TextAlign(ui.AlignCenter)),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("基础状态: int 类型计数器", ui.TextSize(16), ui.TextColor(th.TextColor)),
				),
				ui.Row(
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("-1"),
							ui.OnClick(func(ctx *ui.Context) {
								count.Set(count.Value() - 1)
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Container(
							ui.Style{
								Background: th.Primary,
								Padding:    ui.All(12),
								Radius:     8,
							},
							ui.Text(fmt.Sprintf("计数: %d", count.Value()), ui.TextColor(th.TextOnPrimary)),
						),
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
				ui.Padding(
					ui.All(8),
					ui.Button(
						ui.Text("重置计数器"),
						ui.ButtonBackground(red),
						ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
						ui.OnClick(func(ctx *ui.Context) {
							count.Set(0)
						}),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("字符串状态: 消息展示", ui.TextSize(16), ui.TextColor(th.TextColor)),
				),
				ui.Padding(
					ui.All(4),
					ui.Text(message.Value(), ui.TextSize(14), ui.TextColor(th.SurfaceMuted)),
				),
				ui.Row(
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("设置消息"),
							ui.OnClick(func(ctx *ui.Context) {
								message.Set("你好, FluxUI!")
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("清空消息"),
							ui.OnClick(func(ctx *ui.Context) {
								message.Set("")
							}),
						),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("复杂状态: 列表操作", ui.TextSize(16), ui.TextColor(th.TextColor)),
				),
				ui.Padding(
					ui.All(4),
					ui.Text(fmt.Sprintf("列表长度: %d", len(items.Value())), ui.TextSize(14)),
				),
				ui.Padding(
					ui.All(4),
					ui.Container(
						ui.Style{
							Background: th.SurfaceMuted,
							Padding:    ui.All(8),
							Radius:     4,
						},
						ui.Column(func() []ui.Widget {
							widgets := []ui.Widget{}
							for i, item := range items.Value() {
								currentIndex := i
								widgets = append(widgets, ui.Padding(
									ui.All(4),
									ui.Row(
										ui.Padding(
											ui.All(2),
											ui.Text(fmt.Sprintf("[%d] %s", currentIndex, item)),
										),
										ui.Padding(
											ui.All(2),
											ui.Button(
												ui.Text("删除"),
												ui.ButtonBackground(red),
												ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
												ui.OnClick(func(ctx *ui.Context) {
													newItems := make([]string, len(items.Value()))
													copy(newItems, items.Value())
													newItems = append(newItems[:currentIndex], newItems[currentIndex+1:]...)
													items.Set(newItems)
												}),
											),
										),
									),
								))
							}
							return widgets
						}()...),
					),
				),
				ui.Row(
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("添加项目"),
							ui.ButtonBackground(green),
							ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
							ui.OnClick(func(ctx *ui.Context) {
								newItems := append(items.Value(), fmt.Sprintf("项目 %d", len(items.Value())+1))
								items.Set(newItems)
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("清空列表"),
							ui.ButtonBackground(red),
							ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
							ui.OnClick(func(ctx *ui.Context) {
								items.Set([]string{})
							}),
						),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("状态键值展示", ui.TextSize(16), ui.TextColor(th.TextColor)),
				),
				ui.Padding(
					ui.All(4),
					ui.Text(fmt.Sprintf("count 键: %s", count.Key()), ui.TextSize(12), ui.TextColor(th.SurfaceMuted)),
				),
				ui.Padding(
					ui.All(4),
					ui.Text(fmt.Sprintf("message 键: %s", message.Key()), ui.TextSize(12), ui.TextColor(th.SurfaceMuted)),
				),
				ui.Padding(
					ui.All(4),
					ui.Text(fmt.Sprintf("items 键: %s", items.Key()), ui.TextSize(12), ui.TextColor(th.SurfaceMuted)),
				),
			),
		)
	}, ui.Title("状态管理示例"), ui.Size(520, 760))
}
