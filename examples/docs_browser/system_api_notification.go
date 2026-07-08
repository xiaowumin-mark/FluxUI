package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsSystemNotificationSection(th *ui.Theme) ui.Element {
	return ui.ComponentElement(func(sectionCtx *ui.Context) ui.Element {
		status := ui.UseState(sectionCtx, "通知可在支持时报告点击、关闭和操作回调。")
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
					system.NotificationAction{ID: "open", Label: "打开 FluxUI", URI: "https://github.com/xiaowumin-mark/FluxUI"},
					system.NotificationAction{ID: "docs", Label: "打开文档", URI: "https://github.com/xiaowumin-mark/FluxUI/tree/main/docs"},
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
				return "通知失败：" + err.Error()
			}
			return fmt.Sprintf("通知已发送：%s", title)
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
				return "Balloon 通知失败：" + err.Error()
			}
			return fmt.Sprintf("Balloon 通知已发送：%s", title)
		}

		return docsSystemSection("Notification API", ui.ColumnElement(
			ui.TextElement("图标路径："+docsSystemOptionalPathLabel(iconPath), ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("通知 信息", status, disabled, func(ctx *ui.Context) string {
					return sendNotification(ctx, "FluxUI 文档", "这是一条基本通知。", "docs-browser", system.NotificationInfo)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("替换分组", status, disabled, func(ctx *ui.Context) string {
					return sendNotification(ctx, "FluxUI 文档", "这将替换之前的分组通知。", "docs-browser", system.NotificationSuccess)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("取消分组", status, disabled, func(ctx *ui.Context) string {
					if err := system.CancelNotificationGroup(context.Background(), "docs-browser"); err != nil {
						return "取消通知分组失败：" + err.Error()
					}
					return "通知分组已取消。"
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("探测 Toast 后端", status, disabled, func(ctx *ui.Context) string {
					probe := system.ProbeNotificationBackend(context.Background(), system.NotificationBackendToast,
						system.NotificationTitle("FluxUI 文档"),
						system.NotificationBody("探测后端"),
						system.NotificationAppID("FluxUI"),
					)
					return formatDocsSystemNotificationProbe(probe)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("探测 balloon", status, disabled, func(ctx *ui.Context) string {
					probe := system.ProbeNotificationBackend(context.Background(), system.NotificationBackendBalloon,
						system.NotificationTitle("FluxUI 文档"),
						system.NotificationBody("探测 balloon 后端"),
						system.NotificationIcon(iconPath),
					)
					return formatDocsSystemNotificationProbe(probe)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("通知 警告", status, disabled, func(ctx *ui.Context) string {
					return sendNotification(ctx, "FluxUI 文档", "当前后端支持警告通知。", "docs-browser-warning", system.NotificationWarning)
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("通知 balloon", status, disabled, func(ctx *ui.Context) string {
					return sendBalloonNotification("FluxUI balloon", "这显式请求 balloon 后端。", "docs-browser-balloon", system.NotificationInfo)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("通知 错误", status, disabled, func(ctx *ui.Context) string {
					return sendNotification(ctx, "FluxUI 文档", "带超时和操作按钮的错误样式通知。", "docs-browser-error", system.NotificationError)
				})),
			),
			ui.VSpacerElement(8),
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}

func docsSystemNotificationIconPath() string {
	return cachedDocsSystemNotificationIconPath()
}

func docsSystemOptionalPathLabel(path string) string {
	if path == "" {
		return "(默认平台图标)"
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
