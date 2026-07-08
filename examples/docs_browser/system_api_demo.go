package main

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type docsSystemProbeState struct {
	Loading      bool
	CapturedAt   time.Time
	Capabilities system.CapabilitySet
	Availability map[system.Capability]system.CapabilityAvailability
}

func buildDocsSystemAPIDemo(ctx *ui.Context, status docsStringState, th *ui.Theme) ui.Element {
	if th == nil {
		th = docsBrowserTheme(defaultDocsThemeSeed, false)
	}

	return ui.Key("docs-system-api-demo", ui.ComponentElement(func(demoCtx *ui.Context) ui.Element {
		probe := ui.UseState(demoCtx, docsSystemProbeState{Loading: true})
		ui.UseMount(demoCtx, func() func() {
			var cancelled atomic.Bool
			go func() {
				next := docsSystemProbeSnapshot()
				if cancelled.Load() {
					return
				}
				probe.Set(next)
			}()
			return func() {
				cancelled.Store(true)
			}
		})

		return ui.ScrollViewElement(ui.ColumnElement(
			ui.TextElement("System API", ui.TextSize(16), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(8),
			ui.TextElement("首先探测能力，然后尝试下方的实时演示。", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(12),
			docsSystemCapabilityGrid(probe.Value(), th),
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
				systemActionButton("复制探测结果", false, func(actionCtx *ui.Context) {
					copySystemProbe(status, probe.Value())
				}),
			),
			ui.VSpacerElement(8),
			ui.ContainerDecorationElement(
				ui.Bg(th.Colors.SurfaceContainer).WithPad(ui.All(10)).WithRad(8),
				ui.TextElement(status.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			),
		), ui.ScrollVertical(true))
	}))
}

func docsSystemProbeSnapshot() docsSystemProbeState {
	caps := system.Capabilities()
	probes := make(map[system.Capability]system.CapabilityAvailability, len(docsSystemCapabilities))
	for _, item := range docsSystemCapabilities {
		probes[item.Capability] = system.Availability(item.Capability)
	}
	return docsSystemProbeState{
		CapturedAt:   time.Now(),
		Capabilities: caps,
		Availability: probes,
	}
}

func systemCapabilityCard(label string, cap system.Capability, probe docsSystemProbeState, th *ui.Theme) ui.Element {
	availability, ok := probe.Availability[cap]
	if !ok {
		availability = system.CapabilityAvailability{
			Capability: cap,
			Status:     system.CapabilityStatusUnavailable,
		}
	}
	supported := probe.Capabilities.Supports(cap) && availability.Supported()
	bg := th.Colors.SurfaceContainer
	fg := th.Colors.OnSurfaceVariant
	stateText := string(availability.Status)
	if probe.Loading {
		stateText = "探测中"
	}
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

func copySystemProbe(status docsStringState, probe docsSystemProbeState) {
	if probe.Loading || len(probe.Availability) == 0 {
		probe = docsSystemProbeSnapshot()
	}
	text := systemProbeSummary(probe)
	if err := system.WriteClipboardText(context.Background(), text); err != nil {
		status.Set("复制探测结果失败：" + err.Error())
		return
	}
	status.Set("已将能力探测结果复制到剪贴板。")
}

func systemProbeSummary(probe docsSystemProbeState) string {
	lines := make([]string, 0, len(docsSystemCapabilities)+1)
	if !probe.CapturedAt.IsZero() {
		lines = append(lines, "captured_at="+probe.CapturedAt.Format(time.RFC3339))
	}
	for _, item := range docsSystemCapabilities {
		availability := probe.Availability[item.Capability]
		lines = append(lines, fmt.Sprintf("%s supported=%t status=%s", item.Capability, probe.Capabilities.Supports(item.Capability), availability.Status))
	}
	return strings.Join(lines, "\n")
}
