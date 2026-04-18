package main

import (
	"fmt"
	"strings"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type FormData struct {
	username    string
	email       string
	password    string
	confirmPass string
}

type ValidationResult struct {
	isValid bool
	errors  map[string]string
}

func validateEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func validatePassword(password string) bool {
	return len(password) >= 8
}

func validateForm(data FormData) ValidationResult {
	result := ValidationResult{
		isValid: true,
		errors:  make(map[string]string),
	}

	if strings.TrimSpace(data.username) == "" {
		result.errors["username"] = "用户名不能为空"
		result.isValid = false
	} else if len(data.username) < 3 {
		result.errors["username"] = "用户名至少需要 3 个字符"
		result.isValid = false
	}

	if strings.TrimSpace(data.email) == "" {
		result.errors["email"] = "邮箱不能为空"
		result.isValid = false
	} else if !validateEmail(data.email) {
		result.errors["email"] = "请输入有效的邮箱地址"
		result.isValid = false
	}

	if strings.TrimSpace(data.password) == "" {
		result.errors["password"] = "密码不能为空"
		result.isValid = false
	} else if !validatePassword(data.password) {
		result.errors["password"] = "密码至少需要 8 个字符"
		result.isValid = false
	}

	if data.confirmPass != data.password {
		result.errors["confirmPass"] = "两次输入的密码不一致"
		result.isValid = false
	}

	return result
}

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)

		username := ui.State[string](ctx)
		email := ui.State[string](ctx)
		password := ui.State[string](ctx)
		confirmPass := ui.State[string](ctx)
		formResult := ui.State[ValidationResult](ctx)
		isSubmitted := ui.State[bool](ctx)

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
					ui.Text("表单验证示例", ui.TextSize(24), ui.TextAlign(ui.AlignCenter)),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("用户注册表单 - 包含完整的验证逻辑", ui.TextSize(14), ui.TextColor(th.SurfaceMuted)),
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
						ui.InputBorder(th.SurfaceMuted),
						ui.InputBorderFocus(th.Primary),
						ui.InputOnChange(func(ctx *ui.Context, value string) {
							username.Set(value)
							isSubmitted.Set(false)
						}),
					),
				),
				ui.Container(
					ui.Style{
						Background: th.SurfaceMuted,
						Padding:    ui.All(8),
						Radius:     4,
					},
					ui.Text("当前输入: "+username.Value(), ui.TextSize(12)),
				),
				ui.Row(
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("设置: Alice"),
							ui.OnClick(func(ctx *ui.Context) {
								username.Set("Alice")
								isSubmitted.Set(false)
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("设置: Bo"),
							ui.OnClick(func(ctx *ui.Context) {
								username.Set("Bo")
								isSubmitted.Set(false)
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("清空"),
							ui.OnClick(func(ctx *ui.Context) {
								username.Set("")
								isSubmitted.Set(false)
							}),
						),
					),
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
						ui.InputBorder(th.SurfaceMuted),
						ui.InputBorderFocus(th.Primary),
						ui.InputOnChange(func(ctx *ui.Context, value string) {
							email.Set(value)
							isSubmitted.Set(false)
						}),
					),
				),
				ui.Container(
					ui.Style{
						Background: th.SurfaceMuted,
						Padding:    ui.All(8),
						Radius:     4,
					},
					ui.Text("当前输入: "+email.Value(), ui.TextSize(12)),
				),
				ui.Row(
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("有效邮箱"),
							ui.OnClick(func(ctx *ui.Context) {
								email.Set("user@example.com")
								isSubmitted.Set(false)
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("无效邮箱"),
							ui.OnClick(func(ctx *ui.Context) {
								email.Set("invalid-email")
								isSubmitted.Set(false)
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("清空"),
							ui.OnClick(func(ctx *ui.Context) {
								email.Set("")
								isSubmitted.Set(false)
							}),
						),
					),
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
						ui.InputPassword(true),
						ui.InputBorder(th.SurfaceMuted),
						ui.InputBorderFocus(th.Primary),
						ui.InputOnChange(func(ctx *ui.Context, value string) {
							password.Set(value)
							isSubmitted.Set(false)
						}),
					),
				),
				ui.Container(
					ui.Style{
						Background: th.SurfaceMuted,
						Padding:    ui.All(8),
						Radius:     4,
					},
					ui.Text("当前输入: "+strings.Repeat("*", len(password.Value())), ui.TextSize(12)),
				),
				ui.Row(
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("有效密码"),
							ui.OnClick(func(ctx *ui.Context) {
								password.Set("password123")
								isSubmitted.Set(false)
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("短密码"),
							ui.OnClick(func(ctx *ui.Context) {
								password.Set("short")
								isSubmitted.Set(false)
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("清空"),
							ui.OnClick(func(ctx *ui.Context) {
								password.Set("")
								isSubmitted.Set(false)
							}),
						),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("确认密码", ui.TextSize(14), ui.TextColor(th.TextColor)),
				),
				ui.Padding(
					ui.All(4),
					ui.TextField(
						confirmPass.Value(),
						ui.InputPlaceholder("请再次输入密码"),
						ui.InputPassword(true),
						ui.InputBorder(th.SurfaceMuted),
						ui.InputBorderFocus(th.Primary),
						ui.InputOnChange(func(ctx *ui.Context, value string) {
							confirmPass.Set(value)
							isSubmitted.Set(false)
						}),
					),
				),
				ui.Container(
					ui.Style{
						Background: th.SurfaceMuted,
						Padding:    ui.All(8),
						Radius:     4,
					},
					ui.Text("当前输入: "+strings.Repeat("*", len(confirmPass.Value())), ui.TextSize(12)),
				),
				ui.Row(
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("匹配密码"),
							ui.OnClick(func(ctx *ui.Context) {
								confirmPass.Set(password.Value())
								isSubmitted.Set(false)
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("不匹配"),
							ui.OnClick(func(ctx *ui.Context) {
								confirmPass.Set("different")
								isSubmitted.Set(false)
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("清空"),
							ui.OnClick(func(ctx *ui.Context) {
								confirmPass.Set("")
								isSubmitted.Set(false)
							}),
						),
					),
				),
				ui.Padding(
					ui.All(12),
					ui.Button(
						ui.Text("提交表单"),
						ui.ButtonBackground(th.Primary),
						ui.ButtonForeground(th.TextOnPrimary),
						ui.ButtonRadius(8),
						ui.OnClick(func(ctx *ui.Context) {
							data := FormData{
								username:    username.Value(),
								email:       email.Value(),
								password:    password.Value(),
								confirmPass: confirmPass.Value(),
							}
							result := validateForm(data)
							formResult.Set(result)
							isSubmitted.Set(true)
						}),
					),
				),
				ui.Padding(
					ui.All(4),
					ui.Button(
						ui.Text("重置表单"),
						ui.ButtonBackground(red),
						ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
						ui.OnClick(func(ctx *ui.Context) {
							username.Set("")
							email.Set("")
							password.Set("")
							confirmPass.Set("")
							formResult.Set(ValidationResult{isValid: true, errors: make(map[string]string)})
							isSubmitted.Set(false)
						}),
					),
				),
				func() ui.Widget {
					if !isSubmitted.Value() {
						return ui.Padding(
							ui.All(8),
							ui.Text("请填写并提交表单", ui.TextSize(14), ui.TextColor(th.SurfaceMuted)),
						)
					}

					result := formResult.Value()
					if result.isValid {
						return ui.Container(
							ui.Style{
								Background: green,
								Padding:    ui.All(12),
								Radius:     8,
							},
							ui.Text("表单验证通过", ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(16)),
						)
					}

					errorWidgets := []ui.Widget{
						ui.Padding(ui.All(2), ui.Text("表单验证失败:", ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(14))),
					}
					for field, errMsg := range result.errors {
						errorWidgets = append(errorWidgets, ui.Padding(
							ui.All(2),
							ui.Text(fmt.Sprintf("- %s: %s", field, errMsg), ui.TextColor(ui.NRGBA(255, 255, 255, 255)), ui.TextSize(12)),
						))
					}

					return ui.Container(
						ui.Style{
							Background: red,
							Padding:    ui.All(12),
							Radius:     8,
						},
						ui.Column(errorWidgets...),
					)
				}(),
				ui.Padding(
					ui.All(8),
					ui.Text("表单验证示例完成", ui.TextSize(14), ui.TextColor(th.SurfaceMuted)),
				),
			),
		)
	}, ui.Title("表单验证示例"), ui.Size(480, 1200))
}
