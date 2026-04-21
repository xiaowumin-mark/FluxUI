package main

import (
	"image/color"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)
		popupOpen := ui.State[bool](ctx)
		name := ui.State[string](ctx)

		return ui.Stack(
			ui.FillWidth(
				ui.Container(
					ui.Style{Background: th.Surface, Padding: ui.All(20)},
					ui.Column(
						ui.Text("Popup 弹窗示例", ui.TextSize(22)),
						ui.VSpacer(12),
						ui.Text("Popup 提供一个纯净的弹窗容器，内部内容完全由你定义。"),
						ui.VSpacer(16),
						ui.Button(
							ui.Text("打开 Popup"),
							ui.OnClick(func(ctx *ui.Context) {
								popupOpen.Set(true)
							}),
						),
						ui.VSpacer(8),
						func() ui.Widget {
							if name.Value() != "" {
								return ui.Text("你输入了: "+name.Value(), ui.TextSize(14))
							}
							return ui.Spacer(0, 0)
						}(),
					),
				),
			),
			ui.Popup(
				popupOpen.Value(),
				ui.Column(
					ui.Text("自定义弹窗", ui.TextSize(18)),
					ui.VSpacer(8),
					ui.Text("弹窗内容完全由你控制，可以放任意组件。", ui.TextSize(13)),
					ui.VSpacer(12),
					ui.TextField(name.Value(), ui.InputPlaceholder("请输入姓名"),
						ui.InputOnChange(func(ctx *ui.Context, v string) {
							name.Set(v)
						}),
					),
					ui.VSpacer(12),
					ui.Row(
						ui.Button(
							ui.Text("关闭"),
							ui.ButtonBackground(color.NRGBA{R: 200, G: 200, B: 200, A: 255}),
							ui.OnClick(func(ctx *ui.Context) {
								popupOpen.Set(false)
							}),
						),
						ui.HSpacer(8),
						ui.Button(
							ui.Text("确认"),
							ui.OnClick(func(ctx *ui.Context) {
								popupOpen.Set(false)
							}),
						),
					),
				),
				ui.PopupWidth(360),
				ui.PopupPadding(ui.All(20)),
				ui.PopupRadius(16),
				ui.PopupMaskClosable(true),
				ui.PopupOnOpenChange(func(ctx *ui.Context, open bool) {
					popupOpen.Set(open)
				}),
			),
		)
	}, ui.Title("Popup Demo"), ui.Size(500, 400))
}
