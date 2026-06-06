package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type dialogStatus struct {
	Busy    bool
	Message string
	Paths   []string
	Result  string
}

type dialogStatusSetter interface {
	Set(dialogStatus)
}

type trayState interface {
	Value() *system.Tray
	Set(*system.Tray)
}

func app(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	status := ui.UseState(ctx, dialogStatus{
		Message: "选择文件、打开系统消息框或发送系统通知。",
	})

	fileDialogSupported := system.Supports(system.CapabilityFileDialog)
	messageBoxSupported := system.Supports(system.CapabilityMessageBox)
	notificationSupported := system.Supports(system.CapabilityNotification)
	traySupported := system.Supports(system.CapabilityTray)
	currentID := ui.CurrentWindowID(ctx)
	handle, _ := ui.GetWindow(currentID)
	trayRef := ui.UseState[*system.Tray](ctx, nil)

	state := status.Value()
	fileDialogDisabled := state.Busy || !fileDialogSupported
	messageBoxDisabled := state.Busy || !messageBoxSupported
	notificationDisabled := state.Busy || !notificationSupported
	trayDisabled := state.Busy || !traySupported

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(16)),
		ui.ScrollViewElement(ui.ColumnElement(
			// 滚动、
			ui.TextElement("System API Showcase", ui.TextSize(22)),
			ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement(capabilityText(fileDialogSupported, messageBoxSupported, notificationSupported, traySupported), ui.TextSize(13), ui.TextColor(th.SurfaceMuted))),
			ui.VSpacerElement(16),
			ui.TextElement("File Dialog", ui.TextSize(16)),
			ui.VSpacerElement(8),
			dialogButton("打开文件", fileDialogDisabled, status, handle, func(ctx *ui.Context, callCtx context.Context) (system.FileDialogResult, error) {
				return ui.OpenFileDialogContext(ctx, callCtx,
					system.FileDialogTitle("打开文件"),
					system.FileDialogFilters(
						system.FileFilter{Name: "Text", Patterns: []string{"txt", "md", "json"}},
						system.FileFilter{Name: "Images", Patterns: []string{"png", "jpg", "webp"}},
						system.FileFilter{Name: "All files", Patterns: []string{"*.*"}},
					),
				)
			}),
			ui.VSpacerElement(8),
			dialogButton("打开多个文件", fileDialogDisabled, status, handle, func(ctx *ui.Context, callCtx context.Context) (system.FileDialogResult, error) {
				return ui.OpenFilesDialogContext(ctx, callCtx,
					system.FileDialogTitle("打开多个文件"),
					system.FileDialogFilters(system.FileFilter{Name: "All files", Patterns: []string{"*.*"}}),
				)
			}),
			ui.VSpacerElement(8),
			dialogButton("保存文件", fileDialogDisabled, status, handle, func(ctx *ui.Context, callCtx context.Context) (system.FileDialogResult, error) {
				return ui.SaveFileDialogContext(ctx, callCtx,
					system.FileDialogTitle("保存文件"),
					system.FileDialogDefaultName("fluxui-output.txt"),
					system.FileDialogFilters(
						system.FileFilter{Name: "Text", Patterns: []string{"txt"}},
						system.FileFilter{Name: "All files", Patterns: []string{"*.*"}},
					),
				)
			}),
			ui.VSpacerElement(8),
			dialogButton("选择目录", fileDialogDisabled, status, handle, func(ctx *ui.Context, callCtx context.Context) (system.FileDialogResult, error) {
				return ui.PickFolderDialogContext(ctx, callCtx,
					system.FileDialogTitle("选择目录"),
				)
			}),
			ui.VSpacerElement(16),
			ui.TextElement("MessageBox", ui.TextSize(16)),
			ui.VSpacerElement(8),
			messageBoxButton("信息", messageBoxDisabled, status, handle, func(ctx *ui.Context, callCtx context.Context) (system.MessageBoxResult, error) {
				return ui.ShowMessageBoxContext(ctx, callCtx,
					system.MessageBoxTitle("FluxUI"),
					system.MessageBoxText("这是一条系统信息消息。"),
					system.MessageBoxStyle(system.MessageBoxInfo),
					system.MessageBoxButtonSet(system.MessageBoxOK),
				)
			}),
			ui.VSpacerElement(8),
			messageBoxButton("确认保存", messageBoxDisabled, status, handle, func(ctx *ui.Context, callCtx context.Context) (system.MessageBoxResult, error) {
				return ui.ShowMessageBoxContext(ctx, callCtx,
					system.MessageBoxTitle("保存更改"),
					system.MessageBoxText("关闭前是否保存当前文档？"),
					system.MessageBoxStyle(system.MessageBoxQuestion),
					system.MessageBoxButtonSet(system.MessageBoxYesNoCancel),
					system.MessageBoxDefaultButton(system.MessageBoxResultCancel),
				)
			}),
			ui.VSpacerElement(8),
			messageBoxButton("重试操作", messageBoxDisabled, status, handle, func(ctx *ui.Context, callCtx context.Context) (system.MessageBoxResult, error) {
				return ui.ShowMessageBoxContext(ctx, callCtx,
					system.MessageBoxTitle("操作失败"),
					system.MessageBoxText("网络请求失败，是否重试？"),
					system.MessageBoxStyle(system.MessageBoxWarning),
					system.MessageBoxButtonSet(system.MessageBoxRetryCancel),
					system.MessageBoxDefaultButton(system.MessageBoxResultRetry),
				)
			}),
			ui.VSpacerElement(8),
			messageBoxButton("错误", messageBoxDisabled, status, handle, func(ctx *ui.Context, callCtx context.Context) (system.MessageBoxResult, error) {
				return ui.ShowMessageBoxContext(ctx, callCtx,
					system.MessageBoxTitle("错误"),
					system.MessageBoxText("示例错误消息。"),
					system.MessageBoxStyle(system.MessageBoxError),
					system.MessageBoxButtonSet(system.MessageBoxOKCancel),
				)
			}),
			ui.VSpacerElement(16),
			ui.TextElement("Notification", ui.TextSize(16)),
			ui.VSpacerElement(8),
			notificationButton("发送通知", notificationDisabled, status, handle),
			ui.VSpacerElement(16),
			ui.TextElement("Tray", ui.TextSize(16)),
			ui.VSpacerElement(8),
			trayCreateButton("创建托盘", trayDisabled, status, handle, trayRef),
			ui.VSpacerElement(8),
			trayActionButton("显示托盘", trayDisabled || trayRef.Value() == nil, status, handle, func() error {
				tray := trayRef.Value()
				if tray == nil {
					return system.ErrClosed
				}
				return tray.Show()
			}),
			ui.VSpacerElement(8),
			trayActionButton("隐藏托盘", trayDisabled || trayRef.Value() == nil, status, handle, func() error {
				tray := trayRef.Value()
				if tray == nil {
					return system.ErrClosed
				}
				return tray.Hide()
			}),
			ui.VSpacerElement(8),
			trayActionButton("隐藏主窗口", trayDisabled || trayRef.Value() == nil, status, handle, func() error {
				if !handle.Hide() {
					return fmt.Errorf("hide window unavailable")
				}
				return nil
			}),
			ui.VSpacerElement(8),
			trayCloseButton("关闭托盘", trayDisabled || trayRef.Value() == nil, status, handle, trayRef),
			ui.VSpacerElement(18),
			resultPanel(th, state),
		)),
	)
}

func dialogButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle, run func(*ui.Context, context.Context) (system.FileDialogResult, error)) ui.Element {
	return ui.FillWidthElement(ui.OutlinedButtonElement(
		ui.TextElement(label),
		ui.Disabled(disabled),
		ui.OnClick(func(ctx *ui.Context) {
			status.Set(dialogStatus{Busy: true, Message: label + "..."})
			go func(uiCtx *ui.Context) {
				result, err := run(uiCtx, context.Background())
				status.Set(formatDialogResult(label, result, err))
				handle.Invalidate()
			}(ctx)
		}),
	))
}

func messageBoxButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle, run func(*ui.Context, context.Context) (system.MessageBoxResult, error)) ui.Element {
	return ui.FillWidthElement(ui.OutlinedButtonElement(
		ui.TextElement(label),
		ui.Disabled(disabled),
		ui.OnClick(func(ctx *ui.Context) {
			status.Set(dialogStatus{Busy: true, Message: label + "..."})
			go func(uiCtx *ui.Context) {
				result, err := run(uiCtx, context.Background())
				status.Set(formatMessageBoxResult(label, result, err))
				handle.Invalidate()
			}(ctx)
		}),
	))
}

func notificationButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle) ui.Element {
	return ui.FillWidthElement(ui.OutlinedButtonElement(
		ui.TextElement(label),
		ui.Disabled(disabled),
		ui.OnClick(func(ctx *ui.Context) {
			status.Set(dialogStatus{Busy: true, Message: label + "..."})
			go func() {
				err := system.Notify(context.Background(),
					system.NotificationTitle("FluxUI"),
					system.NotificationBody("后台任务已经完成。"),
					system.NotificationKindStyle(system.NotificationSuccess),
					system.NotificationGroup("system-showcase"),
					system.NotificationOnClick(func(event system.NotificationEvent) {
						status.Set(dialogStatus{
							Message: "通知事件已返回。",
							Result:  string(event.Kind),
						})
						handle.Invalidate()
					}),
				)
				status.Set(formatNotificationResult(label, err))
				handle.Invalidate()
			}()
		}),
	))
}

func trayCreateButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle, trayRef trayState) ui.Element {
	return ui.FillWidthElement(ui.OutlinedButtonElement(
		ui.TextElement(label),
		ui.Disabled(disabled),
		ui.OnClick(func(ctx *ui.Context) {
			status.Set(dialogStatus{Busy: true, Message: label + "..."})
			go func() {
				if current := trayRef.Value(); current != nil {
					_ = current.Close()
					trayRef.Set(nil)
				}

				var tray *system.Tray
				tray, err := system.NewTray(
					system.TrayTooltip("FluxUI System Showcase"),
					system.TrayOnClick(func(event system.TrayEvent) {
						_ = handle.Show()
						status.Set(dialogStatus{
							Message: "托盘图标已点击，主窗口已显示。",
							Result:  string(event.Kind),
						})
						handle.Invalidate()
					}),
					system.TrayOnDoubleClick(func(event system.TrayEvent) {
						_ = handle.Show()
						status.Set(dialogStatus{
							Message: "托盘图标已双击，主窗口已显示。",
							Result:  string(event.Kind),
						})
						handle.Invalidate()
					}),
					system.TrayMenuItems(
						system.TrayMenuAction("show-window", "显示主窗口", func(event system.TrayEvent) {
							_ = handle.Show()
							status.Set(dialogStatus{Message: "托盘菜单已显示主窗口。", Result: event.ItemID})
							handle.Invalidate()
						}),
						system.TrayMenuAction("hide-window", "隐藏主窗口", func(event system.TrayEvent) {
							_ = handle.Hide()
							status.Set(dialogStatus{Message: "托盘菜单已隐藏主窗口。", Result: event.ItemID})
							handle.Invalidate()
						}),
						system.TrayMenuSeparator(),
						system.TrayMenuItem{ID: "disabled", Label: "禁用菜单项", Disabled: true},
						system.TrayMenuItem{
							ID:      "checked",
							Label:   "已选中菜单项",
							Checked: true,
							OnClick: func(event system.TrayEvent) {
								status.Set(dialogStatus{Message: "托盘菜单选中项已点击。", Result: event.ItemID})
								handle.Invalidate()
							},
						},
						system.TrayMenuSeparator(),
						system.TrayMenuAction("close-tray", "关闭托盘", func(event system.TrayEvent) {
							err := system.ErrClosed
							if tray != nil {
								err = tray.Close()
							}
							trayRef.Set(nil)
							status.Set(formatTrayResult("关闭托盘", err))
							handle.Invalidate()
						}),
					),
				)
				if err != nil {
					status.Set(formatTrayResult(label, err))
					handle.Invalidate()
					return
				}
				if err := tray.Show(); err != nil {
					_ = tray.Close()
					status.Set(formatTrayResult(label, err))
					handle.Invalidate()
					return
				}
				trayRef.Set(tray)
				status.Set(dialogStatus{
					Message: label + " 已创建并显示。",
					Result:  "右键托盘图标可打开菜单。",
				})
				handle.Invalidate()
			}()
		}),
	))
}

func trayActionButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle, run func() error) ui.Element {
	return ui.FillWidthElement(ui.OutlinedButtonElement(
		ui.TextElement(label),
		ui.Disabled(disabled),
		ui.OnClick(func(ctx *ui.Context) {
			status.Set(dialogStatus{Busy: true, Message: label + "..."})
			go func() {
				status.Set(formatTrayResult(label, run()))
				handle.Invalidate()
			}()
		}),
	))
}

func trayCloseButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle, trayRef trayState) ui.Element {
	return ui.FillWidthElement(ui.OutlinedButtonElement(
		ui.TextElement(label),
		ui.Disabled(disabled),
		ui.OnClick(func(ctx *ui.Context) {
			status.Set(dialogStatus{Busy: true, Message: label + "..."})
			go func() {
				err := system.ErrClosed
				if tray := trayRef.Value(); tray != nil {
					err = tray.Close()
				}
				trayRef.Set(nil)
				status.Set(formatTrayResult(label, err))
				handle.Invalidate()
			}()
		}),
	))
}

func resultPanel(th *ui.Theme, state dialogStatus) ui.Element {
	body := state.Message
	if state.Result != "" {
		body += "\n\n结果: " + state.Result
	}
	if len(state.Paths) > 0 {
		body += "\n\n" + strings.Join(state.Paths, "\n")
	}
	if state.Busy {
		body += "\n\n等待系统操作返回。"
	}

	return ui.ContainerDecorationElement(
		ui.Bg(th.SurfaceMuted).WithPad(ui.All(12)).WithRad(6),
		ui.ColumnElement(
			ui.TextElement("结果", ui.TextSize(16)),
			ui.PaddingElement(ui.Insets{Top: 8}, ui.TextElement(body, ui.TextSize(13), ui.TextColor(th.TextColor))),
		),
	)
}

func capabilityText(fileDialogSupported, messageBoxSupported, notificationSupported, traySupported bool) string {
	parts := []string{}
	if fileDialogSupported {
		parts = append(parts, "FileDialog 可用")
	} else {
		parts = append(parts, "FileDialog 不可用")
	}
	if messageBoxSupported {
		parts = append(parts, "MessageBox 可用")
	} else {
		parts = append(parts, "MessageBox 不可用")
	}
	if notificationSupported {
		parts = append(parts, "Notification 可用")
	} else {
		parts = append(parts, "Notification 不可用")
	}
	if traySupported {
		parts = append(parts, "Tray 可用")
	} else {
		parts = append(parts, "Tray 不可用")
	}
	return strings.Join(parts, "，") + "。"
}

func formatDialogResult(action string, result system.FileDialogResult, err error) dialogStatus {
	if err != nil {
		if system.IsUnsupported(err) {
			return dialogStatus{Message: action + " 不受当前平台支持。"}
		}
		return dialogStatus{Message: fmt.Sprintf("%s 失败: %v", action, err)}
	}
	if result.Cancelled {
		return dialogStatus{Message: action + " 已取消。"}
	}
	if len(result.Paths) == 0 {
		return dialogStatus{Message: action + " 未返回路径。"}
	}
	return dialogStatus{
		Message: fmt.Sprintf("%s 返回 %d 个路径。", action, len(result.Paths)),
		Paths:   append([]string(nil), result.Paths...),
	}
}

func formatMessageBoxResult(action string, result system.MessageBoxResult, err error) dialogStatus {
	if err != nil {
		if system.IsUnsupported(err) {
			return dialogStatus{Message: action + " 不受当前平台支持。"}
		}
		return dialogStatus{Message: fmt.Sprintf("%s 失败: %v", action, err)}
	}
	return dialogStatus{
		Message: action + " 已返回。",
		Result:  string(result),
	}
}

func formatNotificationResult(action string, err error) dialogStatus {
	if err != nil {
		if system.IsUnsupported(err) {
			return dialogStatus{Message: action + " 不受当前平台支持。"}
		}
		if system.IsUnavailable(err) {
			return dialogStatus{Message: action + " 当前没有可用的系统通知展示路径。"}
		}
		return dialogStatus{Message: fmt.Sprintf("%s 失败: %v", action, err)}
	}
	return dialogStatus{Message: action + " 已提交。"}
}

func formatTrayResult(action string, err error) dialogStatus {
	if err != nil {
		if system.IsUnsupported(err) {
			return dialogStatus{Message: action + " 不受当前平台支持。"}
		}
		if system.IsUnavailable(err) {
			return dialogStatus{Message: action + " 当前没有可用的托盘展示路径。"}
		}
		if system.IsClosed(err) {
			return dialogStatus{Message: action + " 已关闭。"}
		}
		return dialogStatus{Message: fmt.Sprintf("%s 失败: %v", action, err)}
	}
	return dialogStatus{Message: action + " 已完成。"}
}

func main() {
	_ = ui.RunElement(app, ui.Title("System API Showcase"), ui.Size(560, 920), ui.MinSize(420, 760))
}
