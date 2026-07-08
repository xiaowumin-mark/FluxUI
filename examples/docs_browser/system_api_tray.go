package main

import (
	"fmt"
	"os"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsSystemTraySection(th *ui.Theme) ui.Element {
	return ui.ComponentElement(func(sectionCtx *ui.Context) ui.Element {
		status := ui.UseState(sectionCtx, "托盘使用默认应用图标启动。")
		menuLog := ui.UseState(sectionCtx, "托盘菜单回调将显示在此处。")
		tray := ui.UseState[*system.Tray](sectionCtx, nil)
		menuVersion := ui.UseState(sectionCtx, 1)
		disabled := !system.Supports(system.CapabilityTray)

		traySummary := "托盘尚未创建。"
		if current := tray.Value(); current != nil {
			traySummary = fmt.Sprintf("托盘 可见=%v 已关闭=%v", current.Visible(), current.Closed())
		}

		createTray := func(ctx *ui.Context) string {
			current := tray.Value()
			if current != nil && !current.Closed() {
				return "托盘已存在。"
			}

			trayOptions := []system.TrayOption{
				system.TrayTooltip("FluxUI 文档浏览器"),
				system.TrayMenuItems(docsSystemTrayMenuItems(menuLog, tray, "static")...),
				system.TrayMenuProvider(func() system.TrayMenu {
					return docsSystemTrayMenuItems(menuLog, tray, fmt.Sprintf("dynamic v%d", menuVersion.Value()))
				}),
				system.TrayOnClick(func(ev system.TrayEvent) {
					menuLog.Set("托盘已点击")
				}),
				system.TrayOnDoubleClick(func(ev system.TrayEvent) {
					menuLog.Set("托盘已双击")
				}),
			}
			if iconPath := docsSystemNotificationIconPath(); iconPath != "" {
				trayOptions = append(trayOptions, system.TrayIcon(iconPath))
				if data, err := os.ReadFile(iconPath); err == nil {
					trayOptions = append(trayOptions, system.TrayIconBytes(data))
				}
			}

			created, err := system.NewTray(trayOptions...)
			if err != nil {
				return "托盘创建失败：" + err.Error()
			}
			tray.Set(created)
			return "托盘已创建。"
		}

		createResourceTray := func(ctx *ui.Context) string {
			current := tray.Value()
			if current != nil && !current.Closed() {
				return "托盘已存在。"
			}

			created, err := system.NewTray(
				system.TrayIconResource(1),
				system.TrayTooltip("FluxUI 文档浏览器资源图标"),
				system.TrayMenuItems(docsSystemTrayMenuItems(menuLog, tray, "resource")...),
				system.TrayOnClick(func(ev system.TrayEvent) {
					menuLog.Set("资源托盘已点击")
				}),
			)
			if err != nil {
				return "资源托盘创建失败：" + err.Error()
			}
			tray.Set(created)
			return "资源托盘已创建。"
		}

		trayAction := func(label string, fn func(*system.Tray) error) ui.Element {
			return ui.ExpandedElement(docsSystemRunAsyncButton(label, status, disabled, func(ctx *ui.Context) string {
				current := tray.Value()
				if current == nil {
					return "托盘尚未创建。"
				}
				if err := fn(current); err != nil {
					return label + " 失败：" + err.Error()
				}
				return label + " 成功。"
			}))
		}

		return docsSystemSection("Tray API", ui.ColumnElement(
			ui.TextElement(traySummary, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("创建托盘", status, disabled, createTray)),
				ui.HSpacerElement(8),
				trayAction("显示托盘", func(current *system.Tray) error { return current.Show() }),
				ui.HSpacerElement(8),
				trayAction("隐藏托盘", func(current *system.Tray) error { return current.Hide() }),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("创建资源托盘", status, disabled, createResourceTray)),
				ui.HSpacerElement(8),
				trayAction("关闭托盘", func(current *system.Tray) error { return current.Close() }),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("关闭所有托盘", status, disabled, func(ctx *ui.Context) string {
					if err := system.CloseTrays(); err != nil {
						return "关闭所有托盘失败：" + err.Error()
					}
					tray.Set(nil)
					return "已关闭所有托盘。"
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				trayAction("设置工具提示", func(current *system.Tray) error {
					next := menuVersion.Value() + 1
					menuVersion.Set(next)
					return current.SetTooltip(fmt.Sprintf("FluxUI 文档浏览器 v%d", next))
				}),
				ui.HSpacerElement(8),
				trayAction("设置静态菜单", func(current *system.Tray) error {
					return current.SetMenu(docsSystemTrayMenuItems(menuLog, tray, "updated static"))
				}),
				ui.HSpacerElement(8),
				trayAction("设置动态菜单", func(current *system.Tray) error {
					return current.SetMenuProvider(func() system.TrayMenu {
						return docsSystemTrayMenuItems(menuLog, tray, fmt.Sprintf("provider v%d", menuVersion.Value()))
					})
				}),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				trayAction("设置图标路径", func(current *system.Tray) error {
					path := docsSystemNotificationIconPath()
					if path == "" {
						return fmt.Errorf("no .ico path is available in examples")
					}
					return current.SetIcon(path)
				}),
				ui.HSpacerElement(8),
				trayAction("设置图标字节", func(current *system.Tray) error {
					path := docsSystemNotificationIconPath()
					if path == "" {
						return fmt.Errorf("no .ico path is available in examples")
					}
					data, err := os.ReadFile(path)
					if err != nil {
						return err
					}
					return current.SetIconBytes(data)
				}),
				ui.HSpacerElement(8),
				trayAction("设置图标资源", func(current *system.Tray) error {
					return current.SetIconResource(1)
				}),
			),
			ui.VSpacerElement(8),
			ui.TextElement("菜单日志："+menuLog.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(4),
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}

func docsSystemTrayMenuItems(menuLog docsStringState, tray interface {
	Value() *system.Tray
}, label string) system.TrayMenu {
	return system.TrayMenu{
		system.TrayMenuAction("show", "显示托盘 ("+label+")", func(ev system.TrayEvent) {
			menuLog.Set("托盘菜单：显示 " + label)
			if tr := tray.Value(); tr != nil {
				_ = tr.Show()
			}
		}),
		{
			ID:       "mode",
			Label:    "模式",
			Children: docsSystemTrayModeMenu(menuLog, label),
		},
		{ID: "checked", Label: "已选中项", Checked: true},
		{ID: "disabled", Label: "禁用项", Disabled: true},
		system.TrayMenuAction("hide", "隐藏托盘", func(ev system.TrayEvent) {
			menuLog.Set("托盘菜单：隐藏 " + label)
			if tr := tray.Value(); tr != nil {
				_ = tr.Hide()
			}
		}),
		system.TrayMenuSeparator(),
		system.TrayMenuAction("close", "关闭托盘", func(ev system.TrayEvent) {
			menuLog.Set("托盘菜单：关闭 " + label)
			if tr := tray.Value(); tr != nil {
				_ = tr.Close()
			}
		}),
	}
}

func docsSystemTrayModeMenu(menuLog docsStringState, label string) system.TrayMenu {
	return system.TrayMenu{
		{ID: "compact", Label: "紧凑", Checked: true, OnClick: func(ev system.TrayEvent) {
			menuLog.Set("托盘子菜单：紧凑 " + label)
		}},
		{ID: "expanded", Label: "展开", Default: true, OnClick: func(ev system.TrayEvent) {
			menuLog.Set("托盘子菜单：展开 " + label)
		}},
	}
}
