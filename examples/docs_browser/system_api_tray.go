package main

import (
	"fmt"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsSystemTraySection(th *ui.Theme) ui.Element {
	return ui.ComponentElement(func(sectionCtx *ui.Context) ui.Element {
		status := ui.UseState(sectionCtx, "Tray starts with the default application icon.")
		menuLog := ui.UseState(sectionCtx, "Tray menu callbacks will show up here.")
		tray := ui.UseState[*system.Tray](sectionCtx, nil)
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

			created, err := system.NewTray(
				system.TrayTooltip("FluxUI docs browser"),
				system.TrayMenuProvider(func() system.TrayMenu {
					return system.TrayMenu{
						system.TrayMenuAction("show", "Show tray", func(ev system.TrayEvent) {
							menuLog.Set("Tray menu: show")
							if tr := tray.Value(); tr != nil {
								_ = tr.Show()
							}
						}),
						system.TrayMenuAction("hide", "Hide tray", func(ev system.TrayEvent) {
							menuLog.Set("Tray menu: hide")
							if tr := tray.Value(); tr != nil {
								_ = tr.Hide()
							}
						}),
						system.TrayMenuSeparator(),
						system.TrayMenuAction("close", "Close tray", func(ev system.TrayEvent) {
							menuLog.Set("Tray menu: close")
							if tr := tray.Value(); tr != nil {
								_ = tr.Close()
							}
						}),
					}
				}),
				system.TrayOnClick(func(ev system.TrayEvent) {
					menuLog.Set("Tray clicked")
				}),
				system.TrayOnDoubleClick(func(ev system.TrayEvent) {
					menuLog.Set("Tray double-clicked")
				}),
			)
			if err != nil {
				return "Tray create failed: " + err.Error()
			}
			tray.Set(created)
			return "Tray created."
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
			ui.TextElement("Menu log: "+menuLog.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(4),
			ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}
