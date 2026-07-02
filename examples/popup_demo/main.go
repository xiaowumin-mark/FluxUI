package main

import (
	"image/color"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func App(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	popupOpen := ui.UseState(ctx, false)
	name := ui.UseState(ctx, "")

	return ui.StackElement(
		ui.ContainerDecorationElement(
			ui.Bg(th.Surface).WithPad(ui.All(20)),
			ui.ColumnElement(
				ui.TextElement("Popup 弹窗示例", ui.TextSize(22)),
				ui.SpacerElement(0, 12),
				ui.TextElement("Popup 提供一个纯净的弹窗容器，内部内容完全由你定义。"),
				ui.SpacerElement(0, 16),
				ui.ButtonElement(
					ui.TextElement("打开 Popup"),
					ui.OnClick(func(ctx *ui.Context) {
						popupOpen.Set(true)
					}),
				),
				ui.SpacerElement(0, 8),
				func() ui.Element {
					if name.Value() != "" {
						return ui.TextElement("你输入了: "+name.Value(), ui.TextSize(14))
					}
					return nil
				}(),
			),
		),
		ui.PopupElement(
			popupOpen.Value(),
			ui.ColumnElement(
				ui.TextElement("自定义弹窗", ui.TextSize(18)),
				ui.SpacerElement(0, 8),
				ui.TextElement("弹窗内容完全由你控制，可以放任意组件。", ui.TextSize(13)),
				ui.SpacerElement(0, 12),
				ui.TextFieldElement(name.Value(), ui.InputPlaceholder("请输入姓名"),
					ui.InputOnChange(func(ctx *ui.Context, v string) {
						name.Set(v)
					}),
				),
				ui.SpacerElement(0, 12),
				ui.RowElement(
					ui.ButtonElement(
						ui.TextElement("关闭"),
						ui.ButtonBackground(color.NRGBA{R: 200, G: 200, B: 200, A: 255}),
						ui.OnClick(func(ctx *ui.Context) {
							popupOpen.Set(false)
						}),
					),
					ui.SpacerElement(8, 0),
					ui.ButtonElement(
						ui.TextElement("确认"),
						ui.OnClick(func(ctx *ui.Context) {
							popupOpen.Set(false)
						}),
					),
				),
			),
			ui.PopupWidth(360),
			ui.PopupMaskClosable(true),
			ui.PopupOnOpenChange(func(ctx *ui.Context, open bool) {
				popupOpen.Set(open)
			}),
		),
	)
}

func main() {
	_ = ui.RunElement(App, ui.Title("Popup Demo"), ui.Size(500, 400))
}
