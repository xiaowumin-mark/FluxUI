package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func buildDocsSystemAPIDemo(ctx *ui.Context, status docsStringState, th *ui.Theme) ui.Element {
	_ = ctx
	if th == nil {
		th = docsBrowserTheme(defaultDocsThemeSeed, false)
	}

	return ui.ScrollViewElement(
		ui.ColumnElement(
			ui.TextElement("System API", ui.TextSize(16), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(8),
			ui.TextElement("Probe capabilities first, then try the live sections below.", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(12),
			docsSystemCapabilityGrid(th),
			ui.VSpacerElement(14),
			docsSystemDragDropProbeSection(th),
			ui.VSpacerElement(12),
			docsSystemWindowSection(th),
			ui.VSpacerElement(12),
			docsSystemFileDialogSection(th),
			ui.VSpacerElement(12),
			docsSystemMessageBoxSection(th),
			ui.VSpacerElement(12),
			docsSystemNotificationSection(th),
			ui.VSpacerElement(12),
			docsSystemClipboardShellSection(th),
			ui.VSpacerElement(12),
			docsSystemTraySection(th),
			ui.VSpacerElement(12),
			docsSystemEventsSection(th),
			ui.VSpacerElement(12),
			docsSystemSingleInstanceSection(th),
			ui.VSpacerElement(12),
			docsSystemGlobalShortcutSection(th),
			ui.VSpacerElement(12),
			docsSystemRegistrationSection(th),
			ui.VSpacerElement(14),
			ui.RowElement(
				systemActionButton("Copy probe", false, func(actionCtx *ui.Context) {
					copySystemProbe(status)
				}),
			),
			ui.VSpacerElement(8),
			ui.ContainerDecorationElement(
				ui.Bg(th.Colors.SurfaceContainer).WithPad(ui.All(10)).WithRad(8),
				ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			),
		),
		ui.ScrollVertical(true),
	)
}

func systemCapabilityCard(label string, cap system.Capability, th *ui.Theme) ui.Element {
	caps := system.Capabilities()
	availability := system.Availability(cap)
	supported := caps.Supports(cap) && availability.Supported()
	bg := th.Colors.SurfaceContainer
	fg := th.Colors.OnSurfaceVariant
	stateText := string(availability.Status)
	if supported {
		bg = th.Colors.PrimaryContainer
		fg = th.Colors.OnPrimaryContainer
	}
	return ui.ExpandedElement(
		ui.ContainerDecorationElement(
			ui.Bg(bg).
				WithPad(ui.All(10)).
				WithRad(8).
				WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
			ui.ColumnElement(
				ui.TextElement(label, ui.TextSize(12), ui.TextColor(fg)),
				ui.VSpacerElement(4),
				ui.TextElement(stateText, ui.TextSize(11), ui.TextColor(fg)),
			),
		),
	)
}

func systemActionButton(label string, disabled bool, fn func(ctx *ui.Context)) ui.Element {
	return ui.OutlinedButtonElement(
		ui.TextElement(label, ui.TextSize(12)),
		ui.Disabled(disabled),
		ui.ButtonPadding(ui.Symmetric(6, 10)),
		ui.OnClick(fn),
	)
}

func copySystemProbe(status docsStringState) {
	text := systemProbeSummary()
	if err := system.WriteClipboardText(context.Background(), text); err != nil {
		status.Set("Copy probe failed: " + err.Error())
		return
	}
	status.Set("Copied capability probe to clipboard.")
}

func systemProbeSummary() string {
	caps := []system.Capability{
		system.CapabilityWindow,
		system.CapabilityFileDialog,
		system.CapabilityMessageBox,
		system.CapabilityNotification,
		system.CapabilityTray,
		system.CapabilitySystemEvents,
		system.CapabilityClipboard,
		system.CapabilityShell,
		system.CapabilitySingleInstance,
		system.CapabilitySystemRegistration,
		system.CapabilityGlobalShortcut,
		system.CapabilityDragAndDrop,
	}
	lines := make([]string, 0, len(caps))
	for _, cap := range caps {
		lines = append(lines, fmt.Sprintf("%s=%t", cap, system.Supports(cap)))
	}
	return strings.Join(lines, "\n")
}
