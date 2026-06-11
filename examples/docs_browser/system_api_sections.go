package main

import (
	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsSystemSection(title string, body ui.Element, th *ui.Theme) ui.Element {
	if th == nil {
		th = docsBrowserTheme(defaultDocsThemeSeed, false)
	}
	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.SurfaceContainerLow).
			WithPad(ui.All(12)).
			WithRad(10).
			WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
		ui.ColumnElement(
			ui.TextElement(title, ui.TextSize(15), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(8),
			body,
		),
	)
}

func docsSystemRunAsyncButton(label string, status docsStringState, disabled bool, run func(*ui.Context) string) ui.Element {
	return ui.OutlinedButtonElement(
		ui.TextElement(label, ui.TextSize(12)),
		ui.Disabled(disabled),
		ui.ButtonPadding(ui.Symmetric(6, 10)),
		ui.OnClick(func(ctx *ui.Context) {
			if status != nil {
				status.Set(label + "...")
			}
			go func(uiCtx *ui.Context) {
				if status != nil {
					status.Set(run(uiCtx))
				} else {
					_ = run(uiCtx)
				}
			}(ctx)
		}),
	)
}

func docsSystemCapabilityGrid(th *ui.Theme) ui.Element {
	if th == nil {
		th = docsBrowserTheme(defaultDocsThemeSeed, false)
	}
	return ui.ColumnElement(
		ui.RowElement(
			systemCapabilityCard("Window", system.CapabilityWindow, th),
			ui.HSpacerElement(8),
			systemCapabilityCard("File Dialog", system.CapabilityFileDialog, th),
			ui.HSpacerElement(8),
			systemCapabilityCard("MessageBox", system.CapabilityMessageBox, th),
			ui.HSpacerElement(8),
			systemCapabilityCard("Notification", system.CapabilityNotification, th),
		),
		ui.VSpacerElement(8),
		ui.RowElement(
			systemCapabilityCard("Tray", system.CapabilityTray, th),
			ui.HSpacerElement(8),
			systemCapabilityCard("System Events", system.CapabilitySystemEvents, th),
			ui.HSpacerElement(8),
			systemCapabilityCard("Clipboard", system.CapabilityClipboard, th),
			ui.HSpacerElement(8),
			systemCapabilityCard("Shell", system.CapabilityShell, th),
		),
		ui.VSpacerElement(8),
		ui.RowElement(
			systemCapabilityCard("Single Instance", system.CapabilitySingleInstance, th),
			ui.HSpacerElement(8),
			systemCapabilityCard("System Registration", system.CapabilitySystemRegistration, th),
			ui.HSpacerElement(8),
			systemCapabilityCard("Global Shortcut", system.CapabilityGlobalShortcut, th),
			ui.HSpacerElement(8),
			systemCapabilityCard("Drag & Drop", system.CapabilityDragAndDrop, th),
		),
	)
}

func docsSystemPrependLog(lines []string, line string, max int) []string {
	if max <= 0 {
		max = 8
	}
	next := append([]string{line}, lines...)
	if len(next) > max {
		next = next[:max]
	}
	return next
}

func docsSystemLogPanel(title string, lines []string, th *ui.Theme, height float32) ui.Element {
	if th == nil {
		th = docsBrowserTheme(defaultDocsThemeSeed, false)
	}
	if height <= 0 {
		height = 110
	}
	items := make([]ui.Element, 0, len(lines)+1)
	items = append(items, ui.TextElement(title, ui.TextSize(12), ui.TextColor(th.Colors.OnSurface)))
	if len(lines) == 0 {
		lines = []string{"No events yet."}
	}
	for _, line := range lines {
		items = append(items,
			ui.PaddingElement(
				ui.Insets{Top: 4},
				ui.TextElement(line, ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
			),
		)
	}
	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.SurfaceContainer).WithPad(ui.All(10)).WithRad(8),
		ui.FixedHeightElement(
			height,
			ui.ScrollViewElement(ui.ColumnElement(items...), ui.ScrollVertical(true)),
		),
	)
}
