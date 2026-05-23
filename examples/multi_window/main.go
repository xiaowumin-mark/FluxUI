package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func mainWindow(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	count := ui.UseState(ctx, 0)
	titleSeq := ui.UseState(ctx, 0)
	mainMaximized := ui.UseState(ctx, false)

	currentID := ui.CurrentWindowID(ctx)
	allWindows := ui.ListWindows()
	currentHandle, currentAlive := ui.GetWindow(currentID)

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(16)),
		ui.ColumnElement(
			ui.TextElement("主窗口", ui.TextSize(22)),
			ui.PaddingElement(ui.Insets{Top: 8}, ui.TextElement(fmt.Sprintf("当前窗口 ID: %d", currentID), ui.TextSize(13), ui.TextColor(th.SurfaceMuted))),
			ui.PaddingElement(ui.Insets{Top: 8}, ui.TextElement(fmt.Sprintf("当前存活窗口数: %d / 当前句柄存活: %v", len(allWindows), currentAlive && currentHandle.IsAlive()), ui.TextSize(13), ui.TextColor(th.SurfaceMuted))),
			ui.PaddingElement(ui.Insets{Top: 10}, ui.TextElement(fmt.Sprintf("计数: %d", count.Value()), ui.TextSize(16))),
			ui.PaddingElement(ui.Insets{Top: 10}, ui.RowElement(
				ui.ButtonElement(ui.TextElement("增加计数"), ui.OnClick(func(ctx *ui.Context) { count.Set(count.Value() + 1) })),
				ui.HSpacerElement(10),
				ui.ButtonElement(ui.TextElement("改当前标题"), ui.OnClick(func(ctx *ui.Context) {
					next := titleSeq.Value() + 1
					titleSeq.Set(next)
					ui.WindowSetTitle(ctx, fmt.Sprintf("Main #%d", next))
				})),
			)),
			ui.PaddingElement(ui.Insets{Top: 10}, ui.RowElement(
				ui.ButtonElement(ui.TextElement("居中"), ui.OnClick(func(ctx *ui.Context) { ui.WindowCenter(ctx) })),
				ui.HSpacerElement(10),
				ui.ButtonElement(ui.TextElement("置顶"), ui.OnClick(func(ctx *ui.Context) { ui.WindowRaise(ctx) })),
				ui.HSpacerElement(10),
				ui.ButtonElement(ui.TextElement("最小化"), ui.OnClick(func(ctx *ui.Context) { ui.WindowMinimize(ctx) })),
			)),
			ui.PaddingElement(ui.Insets{Top: 10}, ui.RowElement(
				ui.ButtonElement(ui.TextElement("最大化/还原"), ui.OnClick(func(ctx *ui.Context) {
					if mainMaximized.Value() {
						ui.WindowRestore(ctx)
					} else {
						ui.WindowMaximize(ctx)
					}
					mainMaximized.Set(!mainMaximized.Value())
				})),
				ui.HSpacerElement(10),
				ui.ButtonElement(ui.TextElement("改尺寸"), ui.OnClick(func(ctx *ui.Context) { ui.WindowSetSize(ctx, 640, 420) })),
				ui.HSpacerElement(10),
				ui.ButtonElement(ui.TextElement("关闭工具窗口"), ui.OnClick(func(ctx *ui.Context) {
					for _, h := range ui.ListWindows() {
						if h.ID() != currentID {
							h.Close()
						}
					}
				})),
			)),
		),
	)
}

func toolWindow(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	text := ui.UseState(ctx, "这是一个独立的工具窗口")
	if text.Value() == "" {
		text.Set("这是一个独立的工具窗口")
	}

	currentID := ui.CurrentWindowID(ctx)
	titleSeq := ui.UseState(ctx, 0)
	aliveText := "否"
	if ui.WindowIsAlive(ctx) {
		aliveText = "是"
	}

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(16)),
		ui.ColumnElement(
			ui.TextElement("工具窗口", ui.TextSize(20)),
			ui.PaddingElement(ui.Insets{Top: 8}, ui.TextElement(fmt.Sprintf("窗口 ID: %d / 存活: %s", currentID, aliveText), ui.TextSize(13), ui.TextColor(th.SurfaceMuted))),
			ui.PaddingElement(ui.Insets{Top: 10}, ui.TextFieldElement(text.Value(), ui.InputOnChange(func(ctx *ui.Context, value string) { text.Set(value) }))),
			ui.PaddingElement(ui.Insets{Top: 10}, ui.TextElement("内容: "+text.Value(), ui.TextSize(13), ui.TextColor(th.SurfaceMuted))),
			ui.PaddingElement(ui.Insets{Top: 12}, ui.RowElement(
				ui.ButtonElement(ui.TextElement("重绘"), ui.OnClick(func(ctx *ui.Context) { ui.WindowInvalidate(ctx) })),
				ui.HSpacerElement(10),
				ui.ButtonElement(ui.TextElement("改标题"), ui.OnClick(func(ctx *ui.Context) {
					next := titleSeq.Value() + 1
					titleSeq.Set(next)
					ui.WindowSetTitle(ctx, fmt.Sprintf("Tool #%d", next))
				})),
				ui.HSpacerElement(10),
				ui.ButtonElement(ui.TextElement("全屏"), ui.OnClick(func(ctx *ui.Context) { ui.WindowFullscreen(ctx) })),
				ui.HSpacerElement(10),
				ui.ButtonElement(ui.TextElement("还原"), ui.OnClick(func(ctx *ui.Context) { ui.WindowRestore(ctx) })),
			)),
			ui.PaddingElement(ui.Insets{Top: 10}, ui.ButtonElement(ui.TextElement("关闭当前窗口"), ui.OnClick(func(ctx *ui.Context) { ui.WindowClose(ctx) }))),
		),
	)
}

func main() {
	_ = ui.RunElementMulti(
		ui.WindowElement(
			mainWindow,
			ui.Title("FluxUI Multi Window - Main"),
			ui.Size(560, 360),
		),
		ui.WindowElement(
			toolWindow,
			ui.Title("FluxUI Multi Window - Tool"),
			ui.Size(480, 320),
		),
	)
}
