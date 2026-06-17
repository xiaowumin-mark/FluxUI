package main

import (
	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type docsSystemCapabilityItem struct {
	Label      string
	Capability system.Capability
}

var docsSystemCapabilities = []docsSystemCapabilityItem{
	{Label: "Window", Capability: system.CapabilityWindow},
	{Label: "File Dialog", Capability: system.CapabilityFileDialog},
	{Label: "MessageBox", Capability: system.CapabilityMessageBox},
	{Label: "Notification", Capability: system.CapabilityNotification},
	{Label: "Tray", Capability: system.CapabilityTray},
	{Label: "System Events", Capability: system.CapabilitySystemEvents},
	{Label: "Clipboard", Capability: system.CapabilityClipboard},
	{Label: "Shell", Capability: system.CapabilityShell},
	{Label: "Single Instance", Capability: system.CapabilitySingleInstance},
	{Label: "System Registration", Capability: system.CapabilitySystemRegistration},
	{Label: "Global Shortcut", Capability: system.CapabilityGlobalShortcut},
	{Label: "Drag & Drop", Capability: system.CapabilityDragAndDrop},
}

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

func docsSystemCapabilityGrid(probe docsSystemProbeState, th *ui.Theme) ui.Element {
	if th == nil {
		th = docsBrowserTheme(defaultDocsThemeSeed, false)
	}
	rows := make([]ui.Element, 0, (len(docsSystemCapabilities)+3)/4*2)
	for start := 0; start < len(docsSystemCapabilities); start += 4 {
		end := start + 4
		if end > len(docsSystemCapabilities) {
			end = len(docsSystemCapabilities)
		}
		cells := make([]ui.Element, 0, (end-start)*2-1)
		for idx, item := range docsSystemCapabilities[start:end] {
			if idx > 0 {
				cells = append(cells, ui.HSpacerElement(8))
			}
			cells = append(cells, systemCapabilityCard(item.Label, item.Capability, probe, th))
		}
		if len(rows) > 0 {
			rows = append(rows, ui.VSpacerElement(8))
		}
		rows = append(rows, ui.RowElement(cells...))
	}
	return ui.ColumnElement(rows...)
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
