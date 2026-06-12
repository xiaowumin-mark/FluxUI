package main

import (
	"fmt"
	"os"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsSystemTraySection(th *ui.Theme) ui.Element {
	return ui.ComponentElement(func(sectionCtx *ui.Context) ui.Element {
		status := ui.UseState(sectionCtx, "Tray starts with the default application icon.")
		menuLog := ui.UseState(sectionCtx, "Tray menu callbacks will show up here.")
		tray := ui.UseState[*system.Tray](sectionCtx, nil)
		menuVersion := ui.UseState(sectionCtx, 1)
		disabled := !system.Supports(system.CapabilityTray)

		traySummary := "Tray not created."
		if current := tray.Value(); current != nil {
			traySummary = fmt.Sprintf("Tray visible=%v closed=%v", current.Visible(), current.Closed())
		}

		createTray := func(ctx *ui.Context) string {
			current := tray.Value()
			if current != nil && !current.Closed() {
				return "Tray already exists."
			}

			trayOptions := []system.TrayOption{
				system.TrayTooltip("FluxUI docs browser"),
				system.TrayMenuItems(docsSystemTrayMenuItems(menuLog, tray, "static")...),
				system.TrayMenuProvider(func() system.TrayMenu {
					return docsSystemTrayMenuItems(menuLog, tray, fmt.Sprintf("dynamic v%d", menuVersion.Value()))
				}),
				system.TrayOnClick(func(ev system.TrayEvent) {
					menuLog.Set("Tray clicked")
				}),
				system.TrayOnDoubleClick(func(ev system.TrayEvent) {
					menuLog.Set("Tray double-clicked")
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
				return "Tray create failed: " + err.Error()
			}
			tray.Set(created)
			return "Tray created."
		}

		createResourceTray := func(ctx *ui.Context) string {
			current := tray.Value()
			if current != nil && !current.Closed() {
				return "Tray already exists."
			}

			created, err := system.NewTray(
				system.TrayIconResource(1),
				system.TrayTooltip("FluxUI docs browser resource icon"),
				system.TrayMenuItems(docsSystemTrayMenuItems(menuLog, tray, "resource")...),
				system.TrayOnClick(func(ev system.TrayEvent) {
					menuLog.Set("Resource tray clicked")
				}),
			)
			if err != nil {
				return "Resource tray create failed: " + err.Error()
			}
			tray.Set(created)
			return "Resource tray created."
		}

		trayAction := func(label string, fn func(*system.Tray) error) ui.Element {
			return ui.ExpandedElement(docsSystemRunAsyncButton(label, status, disabled, func(ctx *ui.Context) string {
				current := tray.Value()
				if current == nil {
					return "Tray is not created yet."
				}
				if err := fn(current); err != nil {
					return label + " failed: " + err.Error()
				}
				return label + " succeeded."
			}))
		}

		return docsSystemSection("Tray API", ui.ColumnElement(
			ui.TextElement(traySummary, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Create tray", status, disabled, createTray)),
				ui.HSpacerElement(8),
				trayAction("Show tray", func(current *system.Tray) error { return current.Show() }),
				ui.HSpacerElement(8),
				trayAction("Hide tray", func(current *system.Tray) error { return current.Hide() }),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(docsSystemRunAsyncButton("Create resource tray", status, disabled, createResourceTray)),
				ui.HSpacerElement(8),
				trayAction("Close tray", func(current *system.Tray) error { return current.Close() }),
				ui.HSpacerElement(8),
				ui.ExpandedElement(docsSystemRunAsyncButton("Close all trays", status, disabled, func(ctx *ui.Context) string {
					if err := system.CloseTrays(); err != nil {
						return "Close all trays failed: " + err.Error()
					}
					tray.Set(nil)
					return "Closed all trays."
				})),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				trayAction("Set tooltip", func(current *system.Tray) error {
					next := menuVersion.Value() + 1
					menuVersion.Set(next)
					return current.SetTooltip(fmt.Sprintf("FluxUI docs browser v%d", next))
				}),
				ui.HSpacerElement(8),
				trayAction("Set static menu", func(current *system.Tray) error {
					return current.SetMenu(docsSystemTrayMenuItems(menuLog, tray, "updated static"))
				}),
				ui.HSpacerElement(8),
				trayAction("Set dynamic menu", func(current *system.Tray) error {
					return current.SetMenuProvider(func() system.TrayMenu {
						return docsSystemTrayMenuItems(menuLog, tray, fmt.Sprintf("provider v%d", menuVersion.Value()))
					})
				}),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				trayAction("Set icon path", func(current *system.Tray) error {
					path := docsSystemNotificationIconPath()
					if path == "" {
						return fmt.Errorf("no .ico path is available in examples")
					}
					return current.SetIcon(path)
				}),
				ui.HSpacerElement(8),
				trayAction("Set icon bytes", func(current *system.Tray) error {
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
				trayAction("Set icon resource", func(current *system.Tray) error {
					return current.SetIconResource(1)
				}),
			),
			ui.VSpacerElement(8),
			ui.TextElement("Menu log: "+menuLog.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(4),
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}

func docsSystemTrayMenuItems(menuLog docsStringState, tray interface {
	Value() *system.Tray
}, label string) system.TrayMenu {
	return system.TrayMenu{
		system.TrayMenuAction("show", "Show tray ("+label+")", func(ev system.TrayEvent) {
			menuLog.Set("Tray menu: show " + label)
			if tr := tray.Value(); tr != nil {
				_ = tr.Show()
			}
		}),
		{
			ID:       "mode",
			Label:    "Mode",
			Children: docsSystemTrayModeMenu(menuLog, label),
		},
		{ID: "checked", Label: "Checked item", Checked: true},
		{ID: "disabled", Label: "Disabled item", Disabled: true},
		system.TrayMenuAction("hide", "Hide tray", func(ev system.TrayEvent) {
			menuLog.Set("Tray menu: hide " + label)
			if tr := tray.Value(); tr != nil {
				_ = tr.Hide()
			}
		}),
		system.TrayMenuSeparator(),
		system.TrayMenuAction("close", "Close tray", func(ev system.TrayEvent) {
			menuLog.Set("Tray menu: close " + label)
			if tr := tray.Value(); tr != nil {
				_ = tr.Close()
			}
		}),
	}
}

func docsSystemTrayModeMenu(menuLog docsStringState, label string) system.TrayMenu {
	return system.TrayMenu{
		{ID: "compact", Label: "Compact", Checked: true, OnClick: func(ev system.TrayEvent) {
			menuLog.Set("Tray submenu: compact " + label)
		}},
		{ID: "expanded", Label: "Expanded", Default: true, OnClick: func(ev system.TrayEvent) {
			menuLog.Set("Tray submenu: expanded " + label)
		}},
	}
}
