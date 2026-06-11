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
		status := ui.UseState(sectionCtx, "Registration writes current-user entries. Use cleanup after testing.")
		state := ui.UseState(sectionCtx, docsSystemRegistrationDemoState{})
		events := ui.UseState(sectionCtx, []string{"No Toast activations yet."})
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

		targetText := "Executable unavailable: " + docsSystemErrorText(targetErr)
		if targetErr == nil {
			targetText = "Executable: " + targets.Executable
		}

		return docsSystemSection("System Registration API", ui.ColumnElement(
			ui.TextElement(targetText, ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(6),
			ui.TextElement(formatDocsSystemRegistrationState(displayState), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Register protocol", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.RegisterProtocolHandler(context.Background(), docsSystemRegistrationScheme, targets.ProtocolCommand,
						system.RegistrationDisplayName("FluxUI Docs Browser Demo"),
						system.RegistrationIcon(targets.Icon),
					); err != nil {
						return "Register protocol failed: " + err.Error()
					}
					next := state.Value()
					next.ProtocolHandler = true
					state.Set(next)
					return "Registered protocol handler: " + docsSystemRegistrationScheme + "://"
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Register file type", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.RegisterFileAssociation(context.Background(), docsSystemRegistrationExtension, docsSystemRegistrationProgID, targets.FileAssociationCommand,
						system.RegistrationDisplayName("FluxUI Docs Browser Demo Document"),
						system.RegistrationIcon(targets.Icon),
					); err != nil {
						return "Register file type failed: " + err.Error()
					}
					next := state.Value()
					next.FileAssociation = true
					state.Set(next)
					return "Registered file association: " + docsSystemRegistrationExtension
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Register startup", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.RegisterStartupTask(context.Background(), docsSystemRegistrationStartupName, targets.StartupCommand); err != nil {
						return "Register startup failed: " + err.Error()
					}
					next := state.Value()
					next.StartupTask = true
					state.Set(next)
					return "Registered current-user startup task."
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Unregister protocol", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.UnregisterProtocolHandler(context.Background(), docsSystemRegistrationScheme); err != nil {
						return "Unregister protocol failed: " + err.Error()
					}
					next := state.Value()
					next.ProtocolHandler = false
					state.Set(next)
					return "Protocol handler removed."
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Unregister file type", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.UnregisterFileAssociation(context.Background(), docsSystemRegistrationExtension, docsSystemRegistrationProgID); err != nil {
						return "Unregister file type failed: " + err.Error()
					}
					next := state.Value()
					next.FileAssociation = false
					state.Set(next)
					return "File association removed."
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Unregister startup", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.UnregisterStartupTask(context.Background(), docsSystemRegistrationStartupName); err != nil {
						return "Unregister startup failed: " + err.Error()
					}
					next := state.Value()
					next.StartupTask = false
					state.Set(next)
					return "Startup task removed."
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Register Toast shortcut", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.RegisterToastShortcut(context.Background(), docsSystemToastAppID, docsSystemToastShortcutName, targets.Executable,
						system.ToastShortcutArguments("--fluxui-docs-startup"),
						system.ToastShortcutIcon(targets.Icon),
						system.ToastShortcutActivatorCLSID(docsSystemToastActivatorCLSID),
					); err != nil {
						return "Register Toast shortcut failed: " + err.Error()
					}
					next := state.Value()
					next.ToastShortcut = true
					state.Set(next)
					return "Toast shortcut registered for AppUserModelID " + docsSystemToastAppID + "."
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Register Toast activator", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.RegisterToastActivator(context.Background(), docsSystemToastActivatorCLSID, targets.ToastActivatorCommand); err != nil {
						return "Register Toast activator failed: " + err.Error()
					}
					next := state.Value()
					next.ToastActivator = true
					state.Set(next)
					return "Toast activator LocalServer registered."
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Start Toast activator", status, !system.Supports(system.CapabilityNotification) || currentActivator != nil, func(ctx *ui.Context) string {
					started, err := system.StartToastActivator(context.Background(), docsSystemToastActivatorCLSID, func(event system.ToastActivationEvent) {
						events.Set(docsSystemPrependLog(events.Value(), formatDocsSystemToastActivationEvent(event), 8))
					})
					if err != nil {
						return "Start Toast activator failed: " + err.Error()
					}
					activator.Set(started)
					next := state.Value()
					next.ToastActivatorRunning = true
					state.Set(next)
					return "Toast activator started in this process."
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Unregister Toast shortcut", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.UnregisterToastShortcut(context.Background(), docsSystemToastShortcutName); err != nil {
						return "Unregister Toast shortcut failed: " + err.Error()
					}
					next := state.Value()
					next.ToastShortcut = false
					state.Set(next)
					return "Toast shortcut removed."
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Unregister Toast activator", status, registrationDisabled, func(ctx *ui.Context) string {
					if err := system.UnregisterToastActivator(context.Background(), docsSystemToastActivatorCLSID); err != nil {
						return "Unregister Toast activator failed: " + err.Error()
					}
					next := state.Value()
					next.ToastActivator = false
					state.Set(next)
					return "Toast activator registration removed."
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Stop Toast activator", status, currentActivator == nil, func(ctx *ui.Context) string {
					current := activator.Value()
					if current == nil {
						return "Toast activator is not running."
					}
					if err := current.Close(); err != nil {
						return "Stop Toast activator failed: " + err.Error()
					}
					activator.Set(nil)
					next := state.Value()
					next.ToastActivatorRunning = false
					state.Set(next)
					return "Toast activator stopped."
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Cleanup all", status, registrationDisabled && currentActivator == nil, func(ctx *ui.Context) string {
					return cleanupDocsSystemRegistrationDemo(state, activator)
				})),
			),
			ui.VSpacerElement(8),
			docsSystemLogPanel("Toast activation events", events.Value(), th, 98),
			ui.VSpacerElement(8),
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}

func docsSystemRegistrationTargetsForCurrentExe() (docsSystemRegistrationTargets, error) {
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
		return "Cleanup completed with errors: " + strings.Join(failures, "; ")
	}
	return "All demo registration entries removed."
}

func docsSystemErrorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
