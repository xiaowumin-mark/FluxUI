package main

import (
	"strings"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type formData struct {
	username string
	email    string
	password string
	confirm  string
}

type validationResult struct {
	errors map[string]string
}

func (r validationResult) valid() bool {
	return len(r.errors) == 0
}

func validateForm(data formData) validationResult {
	errors := make(map[string]string)
	if strings.TrimSpace(data.username) == "" {
		errors["username"] = "用户名不能为空"
	} else if len([]rune(data.username)) < 3 {
		errors["username"] = "用户名至少需要 3 个字符"
	}
	if strings.TrimSpace(data.email) == "" {
		errors["email"] = "邮箱不能为空"
	} else if !strings.Contains(data.email, "@") || !strings.Contains(data.email, ".") {
		errors["email"] = "请输入有效的邮箱地址"
	}
	if len(data.password) < 8 {
		errors["password"] = "密码至少需要 8 个字符"
	}
	if data.confirm != data.password {
		errors["confirm"] = "两次输入的密码不一致"
	}
	return validationResult{errors: errors}
}

func fieldState(key, label string, result validationResult, required, pending bool, hostError string) ui.FieldState {
	state := ui.FieldState{
		Key:         key,
		Label:       label,
		Required:    required,
		Pending:     pending,
		PendingText: "宿主正在进行异步校验…",
		Status:      ui.FieldValid,
	}
	if pending {
		state.Status = ui.FieldPending
	}
	if message := result.errors[key]; message != "" {
		state.Status = ui.FieldInvalid
		state.ErrorText = message
	}
	if hostError != "" {
		state.Status = ui.FieldInvalid
		state.ErrorText = hostError
	}
	return state
}

// App is intentionally a host-driven Form example. It demonstrates synchronous
// validation, an externally represented async-pending phase, cancelable submit,
// and ValidationSummary focus routing without a component goroutine or I/O.
func App(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	username := ui.UseState(ctx, "")
	email := ui.UseState(ctx, "")
	password := ui.UseState(ctx, "")
	confirm := ui.UseState(ctx, "")
	result := ui.UseState(ctx, validationResult{errors: map[string]string{}})
	pending := ui.UseState(ctx, false)
	asyncEmailError := ui.UseState(ctx, "")
	cancelNext := ui.UseState(ctx, false)
	status := ui.UseState(ctx, "填写字段后提交。")

	formRef := ui.UseRef(ctx, ui.NewFormRef())
	usernameRef := ui.UseRef(ctx, ui.NewFormFieldRef())
	emailRef := ui.UseRef(ctx, ui.NewFormFieldRef())
	passwordRef := ui.UseRef(ctx, ui.NewFormFieldRef())
	confirmRef := ui.UseRef(ctx, ui.NewFormFieldRef())
	if formRef.Current == nil {
		formRef.Current = ui.NewFormRef()
		usernameRef.Current = ui.NewFormFieldRef()
		emailRef.Current = ui.NewFormFieldRef()
		passwordRef.Current = ui.NewFormFieldRef()
		confirmRef.Current = ui.NewFormFieldRef()
	}

	data := formData{username: username.Value(), email: email.Value(), password: password.Value(), confirm: confirm.Value()}
	fields := []ui.FieldState{
		fieldState("username", "用户名", result.Value(), true, false, ""),
		fieldState("email", "邮箱", result.Value(), true, pending.Value(), asyncEmailError.Value()),
		fieldState("password", "密码", result.Value(), true, false, ""),
		fieldState("confirm", "确认密码", result.Value(), true, false, ""),
	}

	clearResult := func() {
		result.Set(validationResult{errors: map[string]string{}})
		pending.Set(false)
		asyncEmailError.Set("")
		status.Set("字段已修改，等待下一次宿主校验。")
	}
	fieldOptions := func(state ui.FieldState, ref *ui.FormFieldRef) []ui.FormFieldOption {
		return []ui.FormFieldOption{ui.FormFieldState(state), ui.FormFieldAttachRef(ref)}
	}

	form := ui.FormElement(
		ui.ColumnElement(
			ui.ValidationSummaryElement(fields,
				ui.ValidationSummaryTitle("请先修正以下字段"),
				ui.ValidationSummaryEmptyText("当前没有同步校验错误。"),
				ui.ValidationSummaryOnFocus(func(_ *ui.Context, key string) {
					switch key {
					case "username":
						usernameRef.Current.Focus()
					case "email":
						emailRef.Current.Focus()
					case "password":
						passwordRef.Current.Focus()
					case "confirm":
						confirmRef.Current.Focus()
					}
				}),
			),
			ui.VSpacerElement(12),
			ui.FormFieldElement("username",
				ui.TextFieldElement(username.Value(), ui.InputPlaceholder("至少 3 个字符"), ui.InputOnChange(func(_ *ui.Context, value string) {
					username.Set(value)
					clearResult()
				})),
				fieldOptions(fields[0], usernameRef.Current)...,
			),
			ui.VSpacerElement(10),
			ui.FormFieldElement("email",
				ui.TextFieldElement(email.Value(), ui.InputPlaceholder("name@example.com"), ui.InputOnChange(func(_ *ui.Context, value string) {
					email.Set(value)
					clearResult()
				})),
				fieldOptions(fields[1], emailRef.Current)...,
			),
			ui.VSpacerElement(10),
			ui.FormFieldElement("password",
				ui.TextFieldElement(password.Value(), ui.InputPassword(true), ui.InputPlaceholder("至少 8 个字符"), ui.InputOnChange(func(_ *ui.Context, value string) {
					password.Set(value)
					clearResult()
				})),
				fieldOptions(fields[2], passwordRef.Current)...,
			),
			ui.VSpacerElement(10),
			ui.FormFieldElement("confirm",
				ui.TextFieldElement(confirm.Value(), ui.InputPassword(true), ui.InputPlaceholder("再次输入密码"), ui.InputOnChange(func(_ *ui.Context, value string) {
					confirm.Set(value)
					clearResult()
				})),
				fieldOptions(fields[3], confirmRef.Current)...,
			),
			ui.VSpacerElement(14),
			ui.RowElement(
				ui.FilledButtonElement(ui.TextElement("提交"), ui.ButtonLoading(pending.Value()), ui.OnClick(func(_ *ui.Context) {
					formRef.Current.Submit()
				})),
				ui.HSpacerElement(8),
				ui.OutlinedButtonElement(ui.TextElement("下次提交取消"), ui.OnClick(func(_ *ui.Context) {
					cancelNext.Set(true)
					status.Set("下一个 submit intention 将被宿主取消。")
				})),
				ui.HSpacerElement(8),
				ui.TextButtonElement(ui.TextElement("完成异步校验"), ui.OnClick(func(_ *ui.Context) {
					pending.Set(false)
					asyncEmailError.Set("")
					status.Set("宿主异步校验已完成。")
				})),
				ui.HSpacerElement(8),
				ui.OutlinedButtonElement(ui.TextElement("返回宿主异步错误"), ui.OnClick(func(_ *ui.Context) {
					pending.Set(false)
					asyncEmailError.Set("宿主异步校验发现该邮箱已注册。")
					status.Set("宿主已回写异步错误；摘要可定位邮箱字段。")
				})),
			),
			ui.VSpacerElement(10),
			ui.TextElement(status.Value(), ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
		),
		ui.FormPending(pending.Value()),
		ui.FormAttachRef(formRef.Current),
		ui.FormOnSubmit(func(_ *ui.Context, event *ui.FormSubmitEvent) {
			if cancelNext.Value() {
				event.PreventDefault()
				cancelNext.Set(false)
				status.Set("宿主取消了本次 submit intention。")
				return
			}
			next := validateForm(data)
			result.Set(next)
			asyncEmailError.Set("")
			if !next.valid() {
				pending.Set(false)
				event.PreventDefault()
				status.Set("同步校验失败；可从摘要定位字段。")
				return
			}
			// A real host would begin asynchronous work outside Layout, then render
			// a later pending/error snapshot. This demo exposes that transition
			// explicitly so it remains deterministic and goroutine-free.
			pending.Set(true)
			status.Set("同步校验通过；宿主异步校验处于 pending。")
		}),
	)

	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.Surface).WithPad(ui.All(20)),
		ui.FixedWidthElement(520, ui.ColumnElement(
			ui.TextElement("R1 Form：受控校验与提交", ui.TextType(th.Types.HeadlineSmall), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(6),
			ui.TextElement("校验、pending 与提交均由宿主状态驱动。", ui.TextType(th.Types.BodyMedium), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(18),
			form,
		)),
	)
}

func main() {
	_ = ui.RunElement(App, ui.Title("FluxUI Form Validation"), ui.Size(620, 860))
}
