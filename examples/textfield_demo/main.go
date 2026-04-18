package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)

		username := ui.State[string](ctx)
		email := ui.State[string](ctx)
		password := ui.State[string](ctx)
		isSubscribed := ui.State[bool](ctx)

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
					ui.Text("TextField 示例", ui.TextSize(24), ui.TextAlign(ui.AlignCenter)),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("用户名", ui.TextSize(14), ui.TextColor(th.TextColor)),
				),
				ui.Padding(
					ui.All(4),
					ui.TextField(
						username.Value(),
						ui.InputPlaceholder("请输入用户名"),
						ui.InputPadding(ui.All(12)),
						ui.InputRadius(8),
						ui.InputBorder(th.SurfaceMuted),
						ui.InputBorderFocus(th.Primary),
						ui.InputBackground(ui.NRGBA(255, 255, 255, 255)),
						ui.InputForeground(th.TextColor),
						ui.InputTextSize(16),
						ui.InputOnChange(func(ctx *ui.Context, value string) {
							username.Set(value)
						}),
					),
				),
				ui.Padding(
					ui.All(4),
					ui.Text("当前输入: "+username.Value(), ui.TextSize(12), ui.TextColor(th.SurfaceMuted)),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("邮箱", ui.TextSize(14), ui.TextColor(th.TextColor)),
				),
				ui.Padding(
					ui.All(4),
					ui.TextField(
						email.Value(),
						ui.InputPlaceholder("请输入邮箱"),
						ui.InputPadding(ui.All(12)),
						ui.InputRadius(8),
						ui.InputBorder(th.SurfaceMuted),
						ui.InputBorderFocus(th.Primary),
						ui.InputBackground(ui.NRGBA(255, 255, 255, 255)),
						ui.InputForeground(th.TextColor),
						ui.InputTextSize(16),
						ui.InputOnChange(func(ctx *ui.Context, value string) {
							email.Set(value)
						}),
					),
				),
				ui.Padding(
					ui.All(4),
					ui.Text("当前输入: "+email.Value(), ui.TextSize(12), ui.TextColor(th.SurfaceMuted)),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("密码", ui.TextSize(14), ui.TextColor(th.TextColor)),
				),
				ui.Padding(
					ui.All(4),
					ui.TextField(
						password.Value(),
						ui.InputPlaceholder("请输入密码"),
						ui.InputPadding(ui.All(12)),
						ui.InputRadius(8),
						ui.InputBorder(th.SurfaceMuted),
						ui.InputBorderFocus(th.Primary),
						ui.InputBackground(ui.NRGBA(255, 255, 255, 255)),
						ui.InputForeground(th.TextColor),
						ui.InputTextSize(16),
						ui.InputPassword(true),
						ui.InputOnChange(func(ctx *ui.Context, value string) {
							password.Set(value)
						}),
					),
				),
				ui.Padding(
					ui.All(4),
					ui.Text("当前输入: "+fmt.Sprintf("%s%d", "***", len(password.Value())), ui.TextSize(12), ui.TextColor(th.SurfaceMuted)),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("复选框", ui.TextSize(14), ui.TextColor(th.TextColor)),
				),
				ui.Padding(
					ui.All(4),
					ui.Checkbox(
						"订阅更新邮件",
						isSubscribed.Value(),
						ui.CheckboxColor(th.Primary),
						ui.CheckboxSize(24),
						ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) {
							isSubscribed.Set(checked)
						}),
					),
				),
				ui.Padding(
					ui.All(4),
					ui.Text("订阅状态: "+fmt.Sprintf("%v", isSubscribed.Value()), ui.TextSize(12), ui.TextColor(th.SurfaceMuted)),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("操作", ui.TextSize(14), ui.TextColor(th.TextColor)),
				),
				ui.Row(
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("设置用户名"),
							ui.ButtonBackground(green),
							ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
							ui.OnClick(func(ctx *ui.Context) {
								username.Set("FluxUser123")
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("设置邮箱"),
							ui.ButtonBackground(green),
							ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
							ui.OnClick(func(ctx *ui.Context) {
								email.Set("user@example.com")
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("设置密码"),
							ui.ButtonBackground(green),
							ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
							ui.OnClick(func(ctx *ui.Context) {
								password.Set("secret123")
							}),
						),
					),
				),
				ui.Row(
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("切换订阅"),
							ui.OnClick(func(ctx *ui.Context) {
								isSubscribed.Set(!isSubscribed.Value())
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("清空"),
							ui.ButtonBackground(red),
							ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
							ui.OnClick(func(ctx *ui.Context) {
								username.Set("")
								email.Set("")
								password.Set("")
								isSubscribed.Set(false)
							}),
						),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("TextField 示例完成", ui.TextSize(14), ui.TextColor(th.SurfaceMuted)),
				),
			),
		)
	}, ui.Title("TextField 示例"), ui.Size(480, 780))
}
