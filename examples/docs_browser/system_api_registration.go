package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

const (
	docsSystemRegistrationScheme      = "fluxui-docs-demo"
	docsSystemRegistrationExtension   = ".fluxuidoc"
	docsSystemRegistrationProgID      = "FluxUI.DocsBrowserDemo"
	docsSystemRegistrationStartupName = "FluxUIDocsBrowserDemo"
	docsSystemToastAppID              = "com.example.FluxUI.DocsBrowser"
	docsSystemToastShortcutName       = "FluxUI Docs Browser Demo"
	docsSystemToastActivatorCLSID     = "{2A90C22D-37D8-4B8D-92E4-91D0E1E7B6A1}"
)

type docsSystemRegistrationDemoState struct {
	ProtocolHandler       bool
	FileAssociation       bool
	StartupTask           bool
	ToastShortcut         bool
	ToastActivator        bool
	ToastActivatorRunning bool
}

type docsSystemRegistrationTargets struct {
	Executable             string
	ProtocolCommand        string
	FileAssociationCommand string
	StartupCommand         string
	ToastActivatorCommand  string
	Icon                   string
}

func docsSystemRegistrationSection(th *ui.Theme) ui.Element {
	return ui.ComponentElement(func(sectionCtx *ui.Context) ui.Element {
		status := ui.UseState(sectionCtx, "注册写入当前用户条目。测试后请使用清理。")
		state := ui.UseState(sectionCtx, docsSystemRegistrationDemoState{})
		events := ui.UseState(sectionCtx, []string{"暂无 Toast 激活。"})
		activator := ui.UseState[*system.ToastActivator](sectionCtx, nil)
		registrationDisabled := !system.Supports(system.CapabilitySystemRegistration)

		currentActivator := activator.Value()
		ui.UseEffectWithDeps(sectionCtx, []any{currentActivator}, func() func() {
			handle := currentActivator
			if handle == nil {
				return nil
			}
			return func() {
				_ = handle.Close()
			}
		})

		targets, targetErr := docsSystemRegistrationTargetsForCurrentExe()
		if targetErr != nil {
			registrationDisabled = true
		}
		current := state.Value()
		displayState := current
		displayState.ToastActivatorRunning = currentActivator != nil

		targetText := "可执行文件不可用：" + docsSystemErrorText(targetErr)
		if targetErr == nil {
			targetText = "可执行文件：" + targets.Executable
		}

		return docsSystemSection("System Registration API", ui.ColumnElement(
			ui.TextElement(targetText, ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(6),
			ui.TextElement(formatDocsSystemRegistrationState(displayState), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("注册协议", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.RegisterProtocolHandler(context.Background(), docsSystemRegistrationScheme, targets.ProtocolCommand,
						system.RegistrationDisplayName("FluxUI Docs Browser Demo"),
						system.RegistrationIcon(targets.Icon),
					); err != nil {
						return "注册协议失败：" + err.Error()
					}
					next := state.Value()
					next.ProtocolHandler = true
					state.Set(next)
					return "已注册协议处理器：" + docsSystemRegistrationScheme + "://"
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("注册文件类型", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.RegisterFileAssociation(context.Background(), docsSystemRegistrationExtension, docsSystemRegistrationProgID, targets.FileAssociationCommand,
						system.RegistrationDisplayName("FluxUI Docs Browser Demo Document"),
						system.RegistrationIcon(targets.Icon),
					); err != nil {
						return "注册文件类型失败：" + err.Error()
					}
					next := state.Value()
					next.FileAssociation = true
					state.Set(next)
					return "已注册文件关联：" + docsSystemRegistrationExtension
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("注册自启动", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.RegisterStartupTask(context.Background(), docsSystemRegistrationStartupName, targets.StartupCommand); err != nil {
						return "注册自启动失败：" + err.Error()
					}
					next := state.Value()
					next.StartupTask = true
					state.Set(next)
					return "已注册当前用户自启动任务。"
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("注销协议", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.UnregisterProtocolHandler(context.Background(), docsSystemRegistrationScheme); err != nil {
						return "注销协议失败：" + err.Error()
					}
					next := state.Value()
					next.ProtocolHandler = false
					state.Set(next)
					return "协议处理器已移除。"
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("注销文件类型", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.UnregisterFileAssociation(context.Background(), docsSystemRegistrationExtension, docsSystemRegistrationProgID); err != nil {
						return "注销文件类型失败：" + err.Error()
					}
					next := state.Value()
					next.FileAssociation = false
					state.Set(next)
					return "文件关联已移除。"
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("注销自启动", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.UnregisterStartupTask(context.Background(), docsSystemRegistrationStartupName); err != nil {
						return "注销自启动失败：" + err.Error()
					}
					next := state.Value()
					next.StartupTask = false
					state.Set(next)
					return "自启动任务已移除。"
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("注册 Toast 快捷方式", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.RegisterToastShortcut(context.Background(), docsSystemToastAppID, docsSystemToastShortcutName, targets.Executable,
						system.ToastShortcutArguments("--fluxui-docs-startup"),
						system.ToastShortcutIcon(targets.Icon),
						system.ToastShortcutActivatorCLSID(docsSystemToastActivatorCLSID),
					); err != nil {
						return "注册 Toast 快捷方式失败：" + err.Error()
					}
					next := state.Value()
					next.ToastShortcut = true
					state.Set(next)
					return "Toast 快捷方式已注册，AppUserModelID：" + docsSystemToastAppID + "。"
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("注册 Toast 激活器", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.RegisterToastActivator(context.Background(), docsSystemToastActivatorCLSID, targets.ToastActivatorCommand); err != nil {
						return "注册 Toast 激活器失败：" + err.Error()
					}
					next := state.Value()
					next.ToastActivator = true
					state.Set(next)
					return "Toast 激活器 LocalServer 已注册。"
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("启动 Toast 激活器", status, !system.Supports(system.CapabilityNotification) || currentActivator != nil, func(ctx *ui.Context) string {
					started, err := system.StartToastActivator(context.Background(), docsSystemToastActivatorCLSID, func(event system.ToastActivationEvent) {
						events.Set(docsSystemPrependLog(events.Value(), formatDocsSystemToastActivationEvent(event), 8))
					})
					if err != nil {
						return "启动 Toast 激活器失败：" + err.Error()
					}
					activator.Set(started)
					next := state.Value()
					next.ToastActivatorRunning = true
					state.Set(next)
					return "Toast 激活器已在此进程中启动。"
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("注销 Toast 快捷方式", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.UnregisterToastShortcut(context.Background(), docsSystemToastShortcutName); err != nil {
						return "注销 Toast 快捷方式失败：" + err.Error()
					}
					next := state.Value()
					next.ToastShortcut = false
					state.Set(next)
					return "Toast 快捷方式已移除。"
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("注销 Toast 激活器", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.UnregisterToastActivator(context.Background(), docsSystemToastActivatorCLSID); err != nil {
						return "注销 Toast 激活器失败：" + err.Error()
					}
					next := state.Value()
					next.ToastActivator = false
					state.Set(next)
					return "Toast 激活器注册已移除。"
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("停止 Toast 激活器", status, currentActivator == nil, func(ctx *ui.Context) string {
					current := activator.Value()
					if current == nil {
						return "Toast 激活器未运行。"
					}
					if err := current.Close(); err != nil {
						return "停止 Toast 激活器失败：" + err.Error()
					}
					activator.Set(nil)
					next := state.Value()
					next.ToastActivatorRunning = false
					state.Set(next)
					return "Toast 激活器已停止。"
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("清理全部", status, registrationDisabled && currentActivator == nil, func(ctx *ui.Context) string {
					return cleanupDocsSystemRegistrationDemo(state, activator)
				})),
			),
			ui.VSpacerElement(8),
			docsSystemLogPanel("Toast 激活事件", events.Value(), th, 98),
			ui.VSpacerElement(8),
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}

func docsSystemRegistrationTargetsForCurrentExe() (docsSystemRegistrationTargets, error) {
	return cachedDocsSystemRegistrationTargetsForCurrentExe()
}

func buildDocsSystemRegistrationTargetsForCurrentExe() (docsSystemRegistrationTargets, error) {
	exe, err := os.Executable()
	if err != nil {
		return docsSystemRegistrationTargets{}, err
	}
	if exe == "" {
		return docsSystemRegistrationTargets{}, fmt.Errorf("executable path is empty")
	}
	if abs, err := filepath.Abs(exe); err == nil {
		exe = abs
	}
	quoted := quoteDocsSystemCommandArg(exe)
	return docsSystemRegistrationTargets{
		Executable:             exe,
		ProtocolCommand:        quoted + ` --fluxui-docs-open "%1"`,
		FileAssociationCommand: quoted + ` --fluxui-docs-open "%1"`,
		StartupCommand:         quoted + " --fluxui-docs-startup",
		ToastActivatorCommand:  quoted + " --fluxui-docs-toast-activator",
		Icon:                   exe + ",0",
	}, nil
}

func quoteDocsSystemCommandArg(arg string) string {
	if arg == "" {
		return `""`
	}
	if !strings.ContainsAny(arg, " \t\r\n\"") {
		return arg
	}
	return `"` + strings.ReplaceAll(arg, `"`, `\"`) + `"`
}

func formatDocsSystemRegistrationState(state docsSystemRegistrationDemoState) string {
	return fmt.Sprintf(
		"protocol=%v fileType=%v startup=%v toastShortcut=%v toastActivator=%v activatorRunning=%v",
		state.ProtocolHandler,
		state.FileAssociation,
		state.StartupTask,
		state.ToastShortcut,
		state.ToastActivator,
		state.ToastActivatorRunning,
	)
}

func formatDocsSystemToastActivationEvent(event system.ToastActivationEvent) string {
	inputs := make([]string, 0, len(event.UserInput))
	for key, value := range event.UserInput {
		inputs = append(inputs, key+"="+value)
	}
	if len(inputs) == 0 {
		inputs = append(inputs, "(none)")
	}
	return fmt.Sprintf("%s appID=%s args=%s input=%s", time.Now().Format("15:04:05"), event.AppID, event.Arguments, strings.Join(inputs, ", "))
}

func cleanupDocsSystemRegistrationDemo(state interface {
	Value() docsSystemRegistrationDemoState
	Set(docsSystemRegistrationDemoState)
}, activator interface {
	Value() *system.ToastActivator
	Set(*system.ToastActivator)
}) string {
	var failures []string
	ctx := context.Background()

	if current := activator.Value(); current != nil {
		if err := current.Close(); err != nil {
			failures = append(failures, "stop Toast activator: "+err.Error())
		}
		activator.Set(nil)
	}
	if err := system.UnregisterProtocolHandler(ctx, docsSystemRegistrationScheme); err != nil {
		failures = append(failures, "protocol: "+err.Error())
	}
	if err := system.UnregisterFileAssociation(ctx, docsSystemRegistrationExtension, docsSystemRegistrationProgID); err != nil {
		failures = append(failures, "file type: "+err.Error())
	}
	if err := system.UnregisterStartupTask(ctx, docsSystemRegistrationStartupName); err != nil {
		failures = append(failures, "startup: "+err.Error())
	}
	if err := system.UnregisterToastShortcut(ctx, docsSystemToastShortcutName); err != nil {
		failures = append(failures, "Toast shortcut: "+err.Error())
	}
	if err := system.UnregisterToastActivator(ctx, docsSystemToastActivatorCLSID); err != nil {
		failures = append(failures, "Toast activator: "+err.Error())
	}

	state.Set(docsSystemRegistrationDemoState{})
	if len(failures) > 0 {
		return "清理完成但存在错误：" + strings.Join(failures, "; ")
	}
	return "所有演示注册条目已移除。"
}

func docsSystemErrorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
