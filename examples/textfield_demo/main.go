package main

import (
	"fmt"
	"strings"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

var (
	red   = ui.NRGBA(220, 53, 69, 255)
	green = ui.NRGBA(40, 167, 69, 255)
)

func App(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)

	username := ui.UseState(ctx, "")
	email := ui.UseState(ctx, "")
	password := ui.UseState(ctx, "")
	notes := ui.UseState(ctx, "FluxUI TextField")
	isSubscribed := ui.UseState(ctx, false)
	emailError := email.Value() != "" && !strings.Contains(email.Value(), "@")

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(20)),
		ui.ColumnElement(
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("TextField 示例", ui.TextSize(24), ui.TextAlign(ui.AlignCenter)),
			),
			ui.PaddingElement(
				ui.All(4),
				ui.OutlinedTextFieldElement(
					username.Value(),
					ui.InputLabel("Username"),
					ui.InputPlaceholder("name@example.com"),
					ui.InputLeading(ui.Icon("person")),
					ui.InputTrailing(ui.Icon("badge")),
					ui.InputSupportingText("Outlined TextField with floating label."),
					ui.InputMaxLen(24),
					ui.InputCounter(true),
					ui.InputOnChange(func(ctx *ui.Context, value string) {
						username.Set(value)
					}),
				),
			),
			ui.PaddingElement(
				ui.All(4),
				ui.TextElement("当前输入: "+username.Value(), ui.TextSize(12), ui.TextColor(th.SurfaceMuted)),
			),
			ui.PaddingElement(
				ui.All(4),
				ui.FilledTextFieldElement(
					email.Value(),
					ui.InputLabel("Email"),
					ui.InputPlaceholder("user@example.com"),
					ui.InputLeading(ui.Icon("mail")),
					ui.InputError(emailError),
					ui.InputErrorText("Email must contain @"),
					ui.InputSupportingText("Filled TextField with validation."),
					ui.InputOnChange(func(ctx *ui.Context, value string) {
						email.Set(value)
					}),
				),
			),
			ui.PaddingElement(
				ui.All(4),
				ui.TextElement("当前输入: "+email.Value(), ui.TextSize(12), ui.TextColor(th.SurfaceMuted)),
			),
			ui.PaddingElement(
				ui.All(4),
				ui.OutlinedTextFieldElement(
					password.Value(),
					ui.InputLabel("Password"),
					ui.InputLeading(ui.Icon("lock")),
					ui.InputTrailing(ui.Icon("visibility")),
					ui.InputPassword(true),
					ui.InputOnChange(func(ctx *ui.Context, value string) {
						password.Set(value)
					}),
				),
			),
			ui.PaddingElement(
				ui.All(4),
				ui.TextElement("当前输入: "+fmt.Sprintf("%s%d", "***", len(password.Value())), ui.TextSize(12), ui.TextColor(th.SurfaceMuted)),
			),
			ui.PaddingElement(
				ui.All(4),
				ui.OutlinedTextFieldElement(
					notes.Value(),
					ui.InputLabel("Notes"),
					ui.InputSingleLine(false),
					ui.InputRows(4),
					ui.InputSupportingText("Textarea mode keeps the same MD3 styling."),
					ui.InputOnChange(func(ctx *ui.Context, value string) {
						notes.Set(value)
					}),
				),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("复选框", ui.TextSize(14), ui.TextColor(th.TextColor)),
			),
			ui.PaddingElement(
				ui.All(4),
				ui.CheckboxElement(
					"订阅更新邮件",
					isSubscribed.Value(),
					ui.CheckboxColor(th.Primary),
					ui.CheckboxSize(24),
					ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) {
						isSubscribed.Set(checked)
					}),
				),
			),
			ui.PaddingElement(
				ui.All(4),
				ui.TextElement("订阅状态: "+fmt.Sprintf("%v", isSubscribed.Value()), ui.TextSize(12), ui.TextColor(th.SurfaceMuted)),
			),
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("操作", ui.TextSize(14), ui.TextColor(th.TextColor)),
			),
			ui.RowElement(
				ui.PaddingElement(
					ui.All(4),
					ui.ButtonElement(
						ui.TextElement("设置用户名"),
						ui.ButtonBackground(green),
						ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
						ui.OnClick(func(ctx *ui.Context) {
							username.Set("FluxUser123")
						}),
					),
				),
				ui.PaddingElement(
					ui.All(4),
					ui.ButtonElement(
						ui.TextElement("设置邮箱"),
						ui.ButtonBackground(green),
						ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
						ui.OnClick(func(ctx *ui.Context) {
							email.Set("user@example.com")
						}),
					),
				),
				ui.PaddingElement(
					ui.All(4),
					ui.ButtonElement(
						ui.TextElement("设置密码"),
						ui.ButtonBackground(green),
						ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
						ui.OnClick(func(ctx *ui.Context) {
							password.Set("secret123")
						}),
					),
				),
			),
			ui.RowElement(
				ui.PaddingElement(
					ui.All(4),
					ui.ButtonElement(
						ui.TextElement("切换订阅"),
						ui.OnClick(func(ctx *ui.Context) {
							isSubscribed.Set(!isSubscribed.Value())
						}),
					),
				),
				ui.PaddingElement(
					ui.All(4),
					ui.ButtonElement(
						ui.TextElement("清空"),
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
			ui.PaddingElement(
				ui.All(8),
				ui.TextElement("TextField 示例完成", ui.TextSize(14), ui.TextColor(th.SurfaceMuted)),
			),
		),
	)
}

func main() {
	_ = ui.RunElement(App, ui.Title("TextField 示例"), ui.Size(480, 780))
}
