package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	mainWindow := ui.Window(
		func(ctx *ui.Context) ui.Widget {
			th := ui.UseTheme(ctx)
			count := ui.State[int](ctx)
			titleSeq := ui.State[int](ctx)

			currentID := ui.CurrentWindowID(ctx)
			allWindows := ui.ListWindows()

			return ui.Container(
				ui.Style{
					Background: th.Surface,
					Padding:    ui.All(16),
				},
				ui.Column(
					ui.Text("主窗口", ui.TextSize(22)),
					ui.Padding(
						ui.Insets{Top: 8},
						ui.Text(fmt.Sprintf("当前窗口 ID: %d", currentID), ui.TextSize(13), ui.TextColor(th.SurfaceMuted)),
					),
					ui.Padding(
						ui.Insets{Top: 8},
						ui.Text(fmt.Sprintf("当前存活窗口数: %d", len(allWindows)), ui.TextSize(13), ui.TextColor(th.SurfaceMuted)),
					),
					ui.Padding(
						ui.Insets{Top: 10},
						ui.Text(fmt.Sprintf("计数: %d", count.Value()), ui.TextSize(16)),
					),
					ui.Padding(
						ui.Insets{Top: 10},
						ui.Row(
							ui.Button(
								ui.Text("增加计数"),
								ui.OnClick(func(ctx *ui.Context) {
									count.Set(count.Value() + 1)
								}),
							),
							ui.HSpacer(10),
							ui.Button(
								ui.Text("改当前标题"),
								ui.OnClick(func(ctx *ui.Context) {
									next := titleSeq.Value() + 1
									titleSeq.Set(next)
									ui.WindowSetTitle(ctx, fmt.Sprintf("Main #%d", next))
								}),
							),
						),
					),
					ui.Padding(
						ui.Insets{Top: 10},
						ui.Row(
							ui.Button(
								ui.Text("居中"),
								ui.OnClick(func(ctx *ui.Context) {
									ui.WindowCenter(ctx)
								}),
							),
							ui.HSpacer(10),
							ui.Button(
								ui.Text("置顶"),
								ui.OnClick(func(ctx *ui.Context) {
									ui.WindowRaise(ctx)
								}),
							),
							ui.HSpacer(10),
							ui.Button(
								ui.Text("最小化"),
								ui.OnClick(func(ctx *ui.Context) {
									ui.WindowMinimize(ctx)
								}),
							),
						),
					),
					ui.Padding(
						ui.Insets{Top: 10},
						ui.Button(
							ui.Text("关闭工具窗口"),
							ui.OnClick(func(ctx *ui.Context) {
								for _, h := range ui.ListWindows() {
									if h.ID() == currentID {
										continue
									}
									h.Close()
								}
							}),
						),
					),
				),
			)
		},
		ui.Title("FluxUI Multi Window - Main"),
		ui.Size(560, 360),
	)

	toolWindow := ui.Window(
		func(ctx *ui.Context) ui.Widget {
			th := ui.UseTheme(ctx)
			text := ui.State[string](ctx)
			if text.Value() == "" {
				text.Set("这是一个独立的工具窗口")
			}

			currentID := ui.CurrentWindowID(ctx)
			titleSeq := ui.State[int](ctx)

			return ui.Container(
				ui.Style{
					Background: th.Surface,
					Padding:    ui.All(16),
				},
				ui.Column(
					ui.Text("工具窗口", ui.TextSize(20)),
					ui.Padding(
						ui.Insets{Top: 8},
						ui.Text(fmt.Sprintf("窗口 ID: %d", currentID), ui.TextSize(13), ui.TextColor(th.SurfaceMuted)),
					),
					ui.Padding(
						ui.Insets{Top: 10},
						ui.TextField(
							text.Value(),
							ui.InputOnChange(func(ctx *ui.Context, value string) {
								text.Set(value)
							}),
						),
					),
					ui.Padding(
						ui.Insets{Top: 10},
						ui.Text("内容: "+text.Value(), ui.TextSize(13), ui.TextColor(th.SurfaceMuted)),
					),
					ui.Padding(
						ui.Insets{Top: 12},
						ui.Row(
							ui.Button(
								ui.Text("改标题"),
								ui.OnClick(func(ctx *ui.Context) {
									next := titleSeq.Value() + 1
									titleSeq.Set(next)
									ui.WindowSetTitle(ctx, fmt.Sprintf("Tool #%d", next))
								}),
							),
							ui.HSpacer(10),
							ui.Button(
								ui.Text("全屏"),
								ui.OnClick(func(ctx *ui.Context) {
									ui.WindowFullscreen(ctx)
								}),
							),
							ui.HSpacer(10),
							ui.Button(
								ui.Text("还原"),
								ui.OnClick(func(ctx *ui.Context) {
									ui.WindowRestore(ctx)
								}),
							),
						),
					),
					ui.Padding(
						ui.Insets{Top: 10},
						ui.Button(
							ui.Text("关闭当前窗口"),
							ui.OnClick(func(ctx *ui.Context) {
								ui.WindowClose(ctx)
							}),
						),
					),
				),
			)
		},
		ui.Title("FluxUI Multi Window - Tool"),
		ui.Size(480, 320),
	)

	_ = ui.RunMulti(mainWindow, toolWindow)
}
