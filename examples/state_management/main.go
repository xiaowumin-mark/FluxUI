package main

import (
	"fmt"

	statepkg "github.com/xiaowumin-mark/FluxUI/state"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

var (
	red   = ui.NRGBA(220, 53, 69, 255)
	green = ui.NRGBA(40, 167, 69, 255)
)

func App(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	count := ui.UseState(ctx, 0)
	message := ui.UseState(ctx, "")
	items := ui.UseState(ctx, []string{})

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(20)),
		ui.ColumnElement(
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("状态管理示例", ui.TextSize(24), ui.TextAlign(ui.AlignCenter)),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("基础状态: int 类型计数器", ui.TextSize(16), ui.TextColor(th.TextColor)),
			),
			ui.RowElement(
				ui.PaddingElement(
					ui.All(4),
					ui.ButtonElement(
						ui.TextElement("-1"),
						ui.OnClick(func(ctx *ui.Context) {
							count.Set(count.Value() - 1)
						}),
					),
				),
				ui.PaddingElement(
					ui.All(4),
					ui.ContainerDecorationElement(
						ui.Bg(th.Primary).WithPad(ui.All(12)).WithRad(8),
						ui.TextElement(fmt.Sprintf("计数: %d", count.Value()), ui.TextColor(th.TextOnPrimary)),
					),
				),
				ui.PaddingElement(
					ui.All(4),
					ui.ButtonElement(
						ui.TextElement("+1"),
						ui.OnClick(func(ctx *ui.Context) {
							count.Set(count.Value() + 1)
						}),
					),
				),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.ButtonElement(
					ui.TextElement("重置计数器"),
					ui.ButtonBackground(red),
					ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
					ui.OnClick(func(ctx *ui.Context) {
						count.Set(0)
					}),
				),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("字符串状态: 消息展示", ui.TextSize(16), ui.TextColor(th.TextColor)),
			),
			ui.PaddingElement(
				ui.All(4),
				ui.TextElement(message.Value(), ui.TextSize(14), ui.TextColor(th.SurfaceMuted)),
			),
			ui.RowElement(
				ui.PaddingElement(
					ui.All(4),
					ui.ButtonElement(
						ui.TextElement("设置消息"),
						ui.OnClick(func(ctx *ui.Context) {
							message.Set("你好, FluxUI!")
						}),
					),
				),
				ui.PaddingElement(
					ui.All(4),
					ui.ButtonElement(
						ui.TextElement("清空消息"),
						ui.OnClick(func(ctx *ui.Context) {
							message.Set("")
						}),
					),
				),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("复杂状态: 列表操作", ui.TextSize(16), ui.TextColor(th.TextColor)),
			),
			ui.PaddingElement(
				ui.All(4),
				ui.TextElement(fmt.Sprintf("列表长度: %d", len(items.Value())), ui.TextSize(14)),
			),
			ui.PaddingElement(
				ui.All(4),
				ui.ContainerDecorationElement(
					ui.Bg(th.SurfaceMuted).WithPad(ui.All(8)).WithRad(4),
					ui.ColumnElement(buildItemElements(items)...),
				),
			),
			ui.RowElement(
				ui.PaddingElement(
					ui.All(4),
					ui.ButtonElement(
						ui.TextElement("添加项目"),
						ui.ButtonBackground(green),
						ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
						ui.OnClick(func(ctx *ui.Context) {
							newItems := append(items.Value(), fmt.Sprintf("项目 %d", len(items.Value())+1))
							items.Set(newItems)
						}),
					),
				),
				ui.PaddingElement(
					ui.All(4),
					ui.ButtonElement(
						ui.TextElement("清空列表"),
						ui.ButtonBackground(red),
						ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
						ui.OnClick(func(ctx *ui.Context) {
							items.Set([]string{})
						}),
					),
				),
			),
		),
	)
}

func buildItemElements(items *statepkg.State[[]string]) []ui.Element {
	list := items.Value()
	elements := make([]ui.Element, 0, len(list))
	for i, item := range list {
		currentIndex := i
		elements = append(elements, ui.PaddingElement(
			ui.All(4),
			ui.RowElement(
				ui.PaddingElement(
					ui.All(2),
					ui.TextElement(fmt.Sprintf("[%d] %s", currentIndex, item)),
				),
				ui.PaddingElement(
					ui.All(2),
					ui.ButtonElement(
						ui.TextElement("删除"),
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
	return elements
}

func main() {
	_ = ui.RunElement(App, ui.Title("状态管理示例"), ui.Size(520, 760))
}
