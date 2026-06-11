package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

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

type boolState interface {
	Value() bool
	Set(bool)
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
	systemEventsSupported := system.Supports(system.CapabilitySystemEvents)
	currentID := ui.CurrentWindowID(ctx)
	handle, _ := ui.GetWindow(currentID)
	trayRef := ui.UseState[*system.Tray](ctx, nil)
	trayChecked := ui.UseState(ctx, false)

	state := status.Value()
	fileDialogDisabled := state.Busy || !fileDialogSupported
	messageBoxDisabled := state.Busy || !messageBoxSupported
	notificationDisabled := state.Busy || !notificationSupported
	trayDisabled := state.Busy || !traySupported
	systemEventsDisabled := state.Busy || !systemEventsSupported

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(16)),
		ui.ScrollViewElement(ui.ColumnElement(
			// 滚动、
			ui.TextElement("System API Showcase", ui.TextSize(22)),
			ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement(capabilityText(fileDialogSupported, messageBoxSupported, notificationSupported, traySupported, systemEventsSupported), ui.TextSize(13), ui.TextColor(th.SurfaceMuted))),
			ui.VSpacerElement(16),
			ui.TextElement("File Dialog", ui.TextSize(16)),
			ui.VSpacerElement(8),
			dialogButton("打开文件", fileDialogDisabled, status, handle, func(ctx *ui.Context, callCtx context.Context) (system.FileDialogResult, error) {
				return ui.OpenFileDialogContext(ctx, callCtx,
					system.FileDialogTitle("打开文件"),
					system.FileDialogRememberDir("system-showcase-open"),
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
					system.FileDialogRememberDir("system-showcase-open"),
					system.FileDialogFilters(system.FileFilter{Name: "All files", Patterns: []string{"*.*"}}),
				)
			}),
			ui.VSpacerElement(8),
			dialogButton("保存文件", fileDialogDisabled, status, handle, func(ctx *ui.Context, callCtx context.Context) (system.FileDialogResult, error) {
				return ui.SaveFileDialogContext(ctx, callCtx,
					system.FileDialogTitle("保存文件"),
					system.FileDialogDefaultName("fluxui-output"),
					system.FileDialogDefaultExtension("txt"),
					system.FileDialogRememberDir("system-showcase-save"),
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
					system.FileDialogRememberDir("system-showcase-folder"),
				)
			}),
			ui.VSpacerElement(8),
			dialogCancelButton("自动取消文件框", fileDialogDisabled, status, handle, 1500*time.Millisecond, func(ctx *ui.Context, callCtx context.Context) (system.FileDialogResult, error) {
				return ui.OpenFileDialogContext(ctx, callCtx,
					system.FileDialogTitle("1.5 秒后自动取消"),
					system.FileDialogFilters(system.FileFilter{Name: "All files", Patterns: []string{"*.*"}}),
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
			ui.VSpacerElement(8),
			detailedMessageBoxButton("富消息框", messageBoxDisabled, status, handle, func(ctx *ui.Context, callCtx context.Context) (system.MessageBoxDetailedResult, error) {
				return ui.ShowMessageBoxDetailedContext(ctx, callCtx,
					system.MessageBoxTitle("保存更改"),
					system.MessageBoxText("选择关闭前的处理方式。"),
					system.MessageBoxDetails("这里会显示可展开的详细信息，用于验证 Windows TaskDialogIndirect 的 details 区域。"),
					system.MessageBoxFooter("Footer 文本用于验证富消息框的底部区域。"),
					system.MessageBoxVerification("记住我的选择", false),
					system.MessageBoxCustomButtons(
						system.MessageBoxButton{ID: "save", Label: "保存并关闭\n保存当前内容后关闭窗口。", Result: system.MessageBoxResultCustom},
						system.MessageBoxButton{ID: "discard", Label: "不保存\n放弃更改并关闭窗口。", Result: system.MessageBoxResultCustom},
						system.MessageBoxButton{ID: "cancel", Label: "取消", Result: system.MessageBoxResultCancel},
					),
					system.MessageBoxDefaultButtonID("cancel"),
					system.MessageBoxCommandLinks(true),
				)
			}),
			ui.VSpacerElement(8),
			messageBoxCancelButton("自动取消消息框", messageBoxDisabled, status, handle, 1500*time.Millisecond, func(ctx *ui.Context, callCtx context.Context) (system.MessageBoxResult, error) {
				return ui.ShowMessageBoxContext(ctx, callCtx,
					system.MessageBoxTitle("1.5 秒后自动取消"),
					system.MessageBoxText("如果 context 取消生效，这个消息框会被自动关闭。"),
					system.MessageBoxStyle(system.MessageBoxQuestion),
					system.MessageBoxButtonSet(system.MessageBoxOKCancel),
				)
			}),
			ui.VSpacerElement(16),
			ui.TextElement("Notification", ui.TextSize(16)),
			ui.VSpacerElement(8),
			notificationButton("发送通知", notificationDisabled, status, handle),
			ui.VSpacerElement(8),
			notificationReplaceButton("替换同组通知", notificationDisabled, status, handle),
			ui.VSpacerElement(8),
			notificationCancelGroupButton("取消通知组", notificationDisabled, status, handle),
			ui.VSpacerElement(8),
			notificationProtocolButton("Toast 协议激活", notificationDisabled, status, handle),
			ui.VSpacerElement(8),
			notificationProbeButton("探测 Toast 后端", notificationDisabled, status, handle),
			ui.VSpacerElement(16),
			ui.TextElement("Tray", ui.TextSize(16)),
			ui.VSpacerElement(8),
			trayCreateButton("创建托盘", trayDisabled, status, handle, trayRef, trayChecked),
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
			trayIconBytesButton("托盘内存图标", trayDisabled || trayRef.Value() == nil, status, handle, trayRef),
			ui.VSpacerElement(8),
			trayIconResourceButton("托盘资源图标", trayDisabled || trayRef.Value() == nil, status, handle, trayRef),
			ui.VSpacerElement(8),
			trayStatusButton("托盘状态", trayDisabled || trayRef.Value() == nil, status, handle, trayRef),
			ui.VSpacerElement(8),
			trayCloseButton("关闭托盘", trayDisabled || trayRef.Value() == nil, status, handle, trayRef),
			ui.VSpacerElement(16),
			ui.TextElement("System Events", ui.TextSize(16)),
			ui.VSpacerElement(8),
			systemEventsButton("监听系统事件 30 秒", systemEventsDisabled, status, handle),
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

func dialogCancelButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle, delay time.Duration, run func(*ui.Context, context.Context) (system.FileDialogResult, error)) ui.Element {
	return ui.FillWidthElement(ui.OutlinedButtonElement(
		ui.TextElement(label),
		ui.Disabled(disabled),
		ui.OnClick(func(ctx *ui.Context) {
			status.Set(dialogStatus{Busy: true, Message: fmt.Sprintf("%s... 将在 %.1f 秒后取消。", label, delay.Seconds())})
			go func(uiCtx *ui.Context) {
				callCtx, cancel := context.WithCancel(context.Background())
				timer := time.AfterFunc(delay, cancel)
				result, err := run(uiCtx, callCtx)
				timer.Stop()
				cancel()
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

func detailedMessageBoxButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle, run func(*ui.Context, context.Context) (system.MessageBoxDetailedResult, error)) ui.Element {
	return ui.FillWidthElement(ui.OutlinedButtonElement(
		ui.TextElement(label),
		ui.Disabled(disabled),
		ui.OnClick(func(ctx *ui.Context) {
			status.Set(dialogStatus{Busy: true, Message: label + "..."})
			go func(uiCtx *ui.Context) {
				result, err := run(uiCtx, context.Background())
				status.Set(formatDetailedMessageBoxResult(label, result, err))
				handle.Invalidate()
			}(ctx)
		}),
	))
}

func messageBoxCancelButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle, delay time.Duration, run func(*ui.Context, context.Context) (system.MessageBoxResult, error)) ui.Element {
	return ui.FillWidthElement(ui.OutlinedButtonElement(
		ui.TextElement(label),
		ui.Disabled(disabled),
		ui.OnClick(func(ctx *ui.Context) {
			status.Set(dialogStatus{Busy: true, Message: fmt.Sprintf("%s... 将在 %.1f 秒后取消。", label, delay.Seconds())})
			go func(uiCtx *ui.Context) {
				callCtx, cancel := context.WithCancel(context.Background())
				timer := time.AfterFunc(delay, cancel)
				result, err := run(uiCtx, callCtx)
				timer.Stop()
				cancel()
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

func notificationReplaceButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle) ui.Element {
	return ui.FillWidthElement(ui.OutlinedButtonElement(
		ui.TextElement(label),
		ui.Disabled(disabled),
		ui.OnClick(func(ctx *ui.Context) {
			status.Set(dialogStatus{Busy: true, Message: label + "..."})
			go func() {
				err := system.Notify(context.Background(),
					system.NotificationTitle("FluxUI"),
					system.NotificationBody("这是同组通知的第一条，会被下一条替换。"),
					system.NotificationKindStyle(system.NotificationInfo),
					system.NotificationGroup("system-showcase-replace"),
				)
				if err == nil {
					time.Sleep(500 * time.Millisecond)
					err = system.Notify(context.Background(),
						system.NotificationTitle("FluxUI"),
						system.NotificationBody("同组通知已替换为这一条。"),
						system.NotificationKindStyle(system.NotificationSuccess),
						system.NotificationGroup("system-showcase-replace"),
					)
				}
				status.Set(formatNotificationResult(label, err))
				handle.Invalidate()
			}()
		}),
	))
}

func notificationCancelGroupButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle) ui.Element {
	return ui.FillWidthElement(ui.OutlinedButtonElement(
		ui.TextElement(label),
		ui.Disabled(disabled),
		ui.OnClick(func(ctx *ui.Context) {
			status.Set(dialogStatus{Busy: true, Message: label + "..."})
			go func() {
				err := system.CancelNotificationGroup(context.Background(), "system-showcase")
				status.Set(formatNotificationResult(label, err))
				handle.Invalidate()
			}()
		}),
	))
}

func notificationProtocolButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle) ui.Element {
	return ui.FillWidthElement(ui.OutlinedButtonElement(
		ui.TextElement(label),
		ui.Disabled(disabled),
		ui.OnClick(func(ctx *ui.Context) {
			status.Set(dialogStatus{Busy: true, Message: label + "..."})
			go func() {
				err := system.Notify(context.Background(),
					system.NotificationBackendPath(system.NotificationBackendToast),
					system.NotificationTitle("FluxUI"),
					system.NotificationBody("点击主体或“打开”动作会使用 fluxui:// 协议激活。"),
					system.NotificationKindStyle(system.NotificationInfo),
					system.NotificationGroup("system-showcase-protocol"),
					system.NotificationLaunchURI("fluxui://system-showcase/notification"),
					system.NotificationActions(
						system.NotificationAction{Label: "打开", URI: "fluxui://system-showcase/notification/open"},
						system.NotificationAction{ID: "dismiss", Label: "忽略"},
					),
					system.NotificationOnAction(func(event system.NotificationEvent) {
						status.Set(dialogStatus{
							Message: "Toast foreground action 已返回。",
							Result:  event.Action,
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

func notificationProbeButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle) ui.Element {
	return ui.FillWidthElement(ui.OutlinedButtonElement(
		ui.TextElement(label),
		ui.Disabled(disabled),
		ui.OnClick(func(ctx *ui.Context) {
			status.Set(dialogStatus{Busy: true, Message: label + "..."})
			go func() {
				probe := system.ProbeNotificationBackend(context.Background(),
					system.NotificationBackendToast,
					system.NotificationAppID("FluxUI"),
				)
				status.Set(formatNotificationBackendProbe(label, probe))
				handle.Invalidate()
			}()
		}),
	))
}

func trayCreateButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle, trayRef trayState, checked boolState) ui.Element {
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
					system.TrayMenuProvider(func() system.TrayMenu {
						isChecked := checked.Value()
						return system.TrayMenu{
							system.TrayMenuItem{
								ID:      "show-window",
								Label:   "显示主窗口",
								Default: true,
								OnClick: func(event system.TrayEvent) {
									_ = handle.Show()
									status.Set(dialogStatus{Message: "托盘默认菜单项已显示主窗口。", Result: event.ItemID})
									handle.Invalidate()
								},
							},
							system.TrayMenuAction("hide-window", "隐藏主窗口", func(event system.TrayEvent) {
								_ = handle.Hide()
								status.Set(dialogStatus{Message: "托盘菜单已隐藏主窗口。", Result: event.ItemID})
								handle.Invalidate()
							}),
							system.TrayMenuSeparator(),
							system.TrayMenuItem{ID: "disabled", Label: "动态禁用菜单项", Disabled: !isChecked},
							system.TrayMenuItem{
								ID:      "checked",
								Label:   "动态勾选菜单项",
								Checked: isChecked,
								OnClick: func(event system.TrayEvent) {
									checked.Set(!checked.Value())
									status.Set(dialogStatus{Message: "托盘菜单勾选状态已切换。", Result: event.ItemID})
									handle.Invalidate()
								},
							},
							system.TrayMenuItem{
								ID:    "tools",
								Label: "工具",
								Children: system.TrayMenu{
									system.TrayMenuAction("tool-status", "显示托盘状态", func(event system.TrayEvent) {
										status.Set(formatTrayStateResult("托盘子菜单", tray))
										handle.Invalidate()
									}),
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
						}
					}),
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

func trayIconBytesButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle, trayRef trayState) ui.Element {
	return trayActionButton(label, disabled, status, handle, func() error {
		tray := trayRef.Value()
		if tray == nil {
			return system.ErrClosed
		}
		return tray.SetIconBytes(sampleTrayIconBytes())
	})
}

func trayIconResourceButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle, trayRef trayState) ui.Element {
	return trayActionButton(label, disabled, status, handle, func() error {
		tray := trayRef.Value()
		if tray == nil {
			return system.ErrClosed
		}
		return tray.SetIconResource(1)
	})
}

func trayStatusButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle, trayRef trayState) ui.Element {
	return ui.FillWidthElement(ui.OutlinedButtonElement(
		ui.TextElement(label),
		ui.Disabled(disabled),
		ui.OnClick(func(ctx *ui.Context) {
			tray := trayRef.Value()
			status.Set(formatTrayStateResult(label, tray))
			handle.Invalidate()
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

func systemEventsButton(label string, disabled bool, status dialogStatusSetter, handle ui.WindowHandle) ui.Element {
	return ui.FillWidthElement(ui.OutlinedButtonElement(
		ui.TextElement(label),
		ui.Disabled(disabled),
		ui.OnClick(func(ctx *ui.Context) {
			status.Set(dialogStatus{Busy: true, Message: label + "..."})
			go func() {
				callCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				sub, err := system.SubscribeSystemEvents(callCtx)
				if err != nil {
					status.Set(formatSystemEventsResult(label, nil, err))
					handle.Invalidate()
					return
				}
				defer sub.Close()

				events := []system.SystemEvent{}
				status.Set(dialogStatus{
					Busy:    true,
					Message: label + " 已启动。请切换主题、显示设置、电源状态或会话状态来触发事件。",
					Result:  "等待事件，30 秒后自动停止。",
				})
				handle.Invalidate()

				for {
					select {
					case event, ok := <-sub.Events():
						if !ok {
							status.Set(formatSystemEventsResult(label, events, nil))
							handle.Invalidate()
							return
						}
						events = append(events, event)
						if len(events) > 8 {
							events = events[len(events)-8:]
						}
						status.Set(dialogStatus{
							Busy:    true,
							Message: label + " 正在监听。",
							Result:  formatSystemEvents(events),
						})
						handle.Invalidate()
					case <-callCtx.Done():
						status.Set(formatSystemEventsResult(label, events, nil))
						handle.Invalidate()
						return
					}
				}
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

func capabilityText(fileDialogSupported, messageBoxSupported, notificationSupported, traySupported, systemEventsSupported bool) string {
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
	if systemEventsSupported {
		parts = append(parts, "SystemEvents 可用")
	} else {
		parts = append(parts, "SystemEvents 不可用")
	}
	return strings.Join(parts, "，") + "。"
}

func formatDialogResult(action string, result system.FileDialogResult, err error) dialogStatus {
	if err != nil {
		if system.IsUnsupported(err) {
			return dialogStatus{Message: action + " 不受当前平台支持。"}
		}
		if errors.Is(err, context.Canceled) {
			return dialogStatus{Message: action + " 已由 context 取消。"}
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
		if errors.Is(err, context.Canceled) {
			return dialogStatus{Message: action + " 已由 context 取消。"}
		}
		return dialogStatus{Message: fmt.Sprintf("%s 失败: %v", action, err)}
	}
	return dialogStatus{
		Message: action + " 已返回。",
		Result:  string(result),
	}
}

func formatDetailedMessageBoxResult(action string, result system.MessageBoxDetailedResult, err error) dialogStatus {
	if err != nil {
		if system.IsUnsupported(err) {
			return dialogStatus{Message: action + " 不受当前平台支持。"}
		}
		if errors.Is(err, context.Canceled) {
			return dialogStatus{Message: action + " 已由 context 取消。"}
		}
		return dialogStatus{Message: fmt.Sprintf("%s 失败: %v", action, err)}
	}
	return dialogStatus{
		Message: action + " 已返回。",
		Result: fmt.Sprintf("result=%s buttonID=%s verification=%v",
			result.Result,
			result.ButtonID,
			result.VerificationChecked,
		),
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

func formatNotificationBackendProbe(action string, probe system.NotificationBackendProbe) dialogStatus {
	if probe.Err != nil && probe.Status != system.CapabilityStatusAvailable {
		return dialogStatus{
			Message: fmt.Sprintf("%s: %s", action, probe.Status),
			Result:  probe.Err.Error(),
		}
	}
	return dialogStatus{
		Message: fmt.Sprintf("%s: %s", action, probe.Status),
		Result: fmt.Sprintf("backend=%s actions=%v click=%v dismiss=%v actionCallback=%v protocol=%v durable=%v",
			probe.Backend,
			probe.SupportsActionButtons,
			probe.SupportsClickCallback,
			probe.SupportsDismissCallback,
			probe.SupportsActionCallback,
			probe.SupportsProtocolActivation,
			probe.SupportsDurableActivation,
		),
	}
}

func formatSystemEventsResult(action string, events []system.SystemEvent, err error) dialogStatus {
	if err != nil {
		if system.IsUnsupported(err) {
			return dialogStatus{Message: action + " 不受当前平台支持。"}
		}
		if system.IsUnavailable(err) {
			return dialogStatus{Message: action + " 当前没有可用的系统事件订阅路径。"}
		}
		if errors.Is(err, context.Canceled) {
			return dialogStatus{Message: action + " 已由 context 取消。"}
		}
		return dialogStatus{Message: fmt.Sprintf("%s 失败: %v", action, err)}
	}
	if len(events) == 0 {
		return dialogStatus{Message: action + " 已停止。", Result: "未收到系统事件。"}
	}
	return dialogStatus{
		Message: fmt.Sprintf("%s 已停止，共记录最近 %d 条事件。", action, len(events)),
		Result:  formatSystemEvents(events),
	}
}

func formatSystemEvents(events []system.SystemEvent) string {
	if len(events) == 0 {
		return "暂无系统事件。"
	}
	lines := make([]string, 0, len(events))
	for _, event := range events {
		detail := event.Detail
		if detail == "" {
			detail = "-"
		}
		lines = append(lines, fmt.Sprintf("%s %s %s", event.Time.Format("15:04:05"), event.Kind, detail))
	}
	return strings.Join(lines, "\n")
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

func formatTrayStateResult(action string, tray *system.Tray) dialogStatus {
	if tray == nil {
		return dialogStatus{Message: action + " 没有托盘实例。", Result: "visible=false closed=true"}
	}
	return dialogStatus{
		Message: action + " 已读取。",
		Result:  fmt.Sprintf("visible=%v closed=%v", tray.Visible(), tray.Closed()),
	}
}

func sampleTrayIconBytes() []byte {
	const (
		width      = 16
		height     = 16
		headerSize = 40
		pixelBytes = width * height * 4
		maskBytes  = height * 4
		imageBytes = headerSize + pixelBytes + maskBytes
		imageOff   = 6 + 16
	)

	data := make([]byte, imageOff+imageBytes)
	binary.LittleEndian.PutUint16(data[2:], 1)
	binary.LittleEndian.PutUint16(data[4:], 1)
	data[6] = width
	data[7] = height
	binary.LittleEndian.PutUint16(data[10:], 1)
	binary.LittleEndian.PutUint16(data[12:], 32)
	binary.LittleEndian.PutUint32(data[14:], imageBytes)
	binary.LittleEndian.PutUint32(data[18:], imageOff)

	image := data[imageOff:]
	binary.LittleEndian.PutUint32(image[0:], headerSize)
	binary.LittleEndian.PutUint32(image[4:], width)
	binary.LittleEndian.PutUint32(image[8:], height*2)
	binary.LittleEndian.PutUint16(image[12:], 1)
	binary.LittleEndian.PutUint16(image[14:], 32)
	binary.LittleEndian.PutUint32(image[20:], pixelBytes)
	pixels := image[headerSize : headerSize+pixelBytes]
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := (y*width + x) * 4
			pixels[i+0] = byte(24 + y*7)
			pixels[i+1] = byte(110 + x*6)
			pixels[i+2] = 220
			pixels[i+3] = 255
		}
	}
	return data
}

func main() {
	_ = ui.RunElement(app, ui.Title("System API Showcase"), ui.Size(560, 920), ui.MinSize(420, 760))
}
