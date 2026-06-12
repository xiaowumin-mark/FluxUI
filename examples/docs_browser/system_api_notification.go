package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsSystemNotificationSection(th *ui.Theme) ui.Element {
	return ui.ComponentElement(func(sectionCtx *ui.Context) ui.Element {
		status := ui.UseState(sectionCtx, "Notifications can report click, dismiss, and action callbacks when supported.")
		disabled := !system.Supports(system.CapabilityNotification)
		iconPath := docsSystemNotificationIconPath()

		sendNotification := func(ctx *ui.Context, title, body string, group string, kind system.NotificationKind) string {
			err := system.Notify(context.Background(),
				system.NotificationTitle(title),
				system.NotificationBody(body),
				system.NotificationKindStyle(kind),
				system.NotificationGroup(group),
				system.NotificationIcon(iconPath),
				system.NotificationBackendPath(system.NotificationBackendAuto),
				system.NotificationAppID("FluxUI"),
				system.NotificationLaunchURI("https://github.com/xiaowumin-mark/FluxUI"),
				system.NotificationTimeout(4*time.Second),
				system.NotificationActions(
					system.NotificationAction{ID: "open", Label: "Open FluxUI", URI: "https://github.com/xiaowumin-mark/FluxUI"},
					system.NotificationAction{ID: "docs", Label: "Open docs", URI: "https://github.com/xiaowumin-mark/FluxUI/tree/main/docs"},
				),
				system.NotificationOnClick(func(ev system.NotificationEvent) {
					status.Set(formatDocsSystemNotificationEvent("clicked", ev))
				}),
				system.NotificationOnDismiss(func(ev system.NotificationEvent) {
					status.Set(formatDocsSystemNotificationEvent("dismissed", ev))
				}),
				system.NotificationOnAction(func(ev system.NotificationEvent) {
					status.Set(formatDocsSystemNotificationEvent("action", ev))
				}),
			)
			if err != nil {
				return "Notification failed: " + err.Error()
			}
			return fmt.Sprintf("Notification sent: %s", title)
		}

		sendBalloonNotification := func(title, body string, group string, kind system.NotificationKind) string {
			err := system.Notify(context.Background(),
				system.NotificationTitle(title),
				system.NotificationBody(body),
				system.NotificationKindStyle(kind),
				system.NotificationGroup(group),
				system.NotificationIcon(iconPath),
				system.NotificationBackendPath(system.NotificationBackendBalloon),
				system.NotificationTimeout(3*time.Second),
				system.NotificationOnClick(func(ev system.NotificationEvent) {
					status.Set(formatDocsSystemNotificationEvent("clicked", ev))
				}),
				system.NotificationOnDismiss(func(ev system.NotificationEvent) {
					status.Set(formatDocsSystemNotificationEvent("dismissed", ev))
				}),
			)
			if err != nil {
				return "Balloon notification failed: " + err.Error()
			}
			return fmt.Sprintf("Balloon notification sent: %s", title)
		}

		return docsSystemSection("Notification API", ui.ColumnElement(
			ui.TextElement("Icon path: "+docsSystemOptionalPathLabel(iconPath), ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Notify info", status, disabled, func(ctx *ui.Context) string {
					return sendNotification(ctx, "FluxUI Docs", "This is a basic notification.", "docs-browser", system.NotificationInfo)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Replace group", status, disabled, func(ctx *ui.Context) string {
					return sendNotification(ctx, "FluxUI Docs", "This replaces the previous group notification.", "docs-browser", system.NotificationSuccess)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Cancel group", status, disabled, func(ctx *ui.Context) string {
					if err := system.CancelNotificationGroup(context.Background(), "docs-browser"); err != nil {
						return "Cancel notification group failed: " + err.Error()
					}
					return "Notification group canceled."
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Probe Toast backend", status, disabled, func(ctx *ui.Context) string {
					probe := system.ProbeNotificationBackend(context.Background(), system.NotificationBackendToast,
						system.NotificationTitle("FluxUI Docs"),
						system.NotificationBody("Probe backend"),
						system.NotificationAppID("FluxUI"),
					)
					return formatDocsSystemNotificationProbe(probe)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Probe balloon", status, disabled, func(ctx *ui.Context) string {
					probe := system.ProbeNotificationBackend(context.Background(), system.NotificationBackendBalloon,
						system.NotificationTitle("FluxUI Docs"),
						system.NotificationBody("Probe balloon backend"),
						system.NotificationIcon(iconPath),
					)
					return formatDocsSystemNotificationProbe(probe)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Notify warning", status, disabled, func(ctx *ui.Context) string {
					return sendNotification(ctx, "FluxUI Docs", "Warning notifications are supported by the active backend.", "docs-browser-warning", system.NotificationWarning)
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Notify balloon", status, disabled, func(ctx *ui.Context) string {
					return sendBalloonNotification("FluxUI balloon", "This explicitly requests the balloon backend.", "docs-browser-balloon", system.NotificationInfo)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Notify error", status, disabled, func(ctx *ui.Context) string {
					return sendNotification(ctx, "FluxUI Docs", "Error style notification with timeout and actions.", "docs-browser-error", system.NotificationError)
				})),
			),
			ui.VSpacerElement(8),
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}

func docsSystemNotificationIconPath() string {
	candidates := []string{
		filepath.Join("examples", "docs_browser", "docs.ico"),
		filepath.Join("examples", "system_showcase", "system_showcase.ico"),
		filepath.Join("examples", "assets", "sample.ico"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func docsSystemOptionalPathLabel(path string) string {
	if path == "" {
		return "(default platform icon)"
	}
	return path
}

func formatDocsSystemNotificationEvent(action string, ev system.NotificationEvent) string {
	if ev.Action != "" {
		return fmt.Sprintf("Notification %s: group=%q action=%q", action, ev.Group, ev.Action)
	}
	return fmt.Sprintf("Notification %s: group=%q", action, ev.Group)
}

func formatDocsSystemNotificationProbe(probe system.NotificationBackendProbe) string {
	return fmt.Sprintf(
		"Notification backend=%q supported=%v available=%v actions=%v click=%v dismiss=%v action=%v protocol=%v durable=%v",
		probe.Backend,
		probe.Supported(),
		probe.Available(),
		probe.SupportsActionButtons,
		probe.SupportsClickCallback,
		probe.SupportsDismissCallback,
		probe.SupportsActionCallback,
		probe.SupportsProtocolActivation,
		probe.SupportsDurableActivation,
	)
}
