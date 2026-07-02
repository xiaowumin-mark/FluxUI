package main

import (
	"fmt"
	"image/color"
	"path/filepath"
	"strings"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type docsDragDropState struct {
	MainActive  bool
	SmallActive bool
	Events      []string
}

type docsDragDropStateHandle interface {
	Value() docsDragDropState
	Set(docsDragDropState)
}

func defaultDocsDragDropState() docsDragDropState {
	return docsDragDropState{
		Events: []string{"ready: drag text, files, JSON, or custom MIME into a target"},
	}
}

func buildDocsDragDropDemo(ctx *ui.Context, state docsDragDropStateHandle, th *ui.Theme) ui.Element {
	_ = ctx
	current := state.Value()

	return ui.ScrollViewElement(
		ui.ColumnElement(
			ui.RowElement(
				ui.ExpandedElement(docsDragSourcePanel(state, th)),
				ui.HSpacerElement(12),
				ui.ExpandedElement(docsDropTargetPanel(state, current, th)),
			),
			ui.VSpacerElement(12),
			docsDragDropEventLog(current.Events, th),
		),
		ui.ScrollVertical(true),
	)
}

func docsDragSourcePanel(state docsDragDropStateHandle, th *ui.Theme) ui.Element {
	sampleFile := docsDragSampleFile()
	return ui.ColumnElement(
		ui.TextElement("Sources", ui.TextSize(15), ui.TextColor(th.Colors.OnSurface)),
		ui.VSpacerElement(8),
		docsDragSourceCard(
			"Text",
			"text/plain + custom preview",
			th,
			ui.NRGBA(239, 246, 255, 255),
			ui.NRGBA(37, 99, 235, 255),
			ui.DragSourceText("FluxUI docs payload"),
			ui.DragSourceOperations(ui.DragOperationCopy),
			ui.DragSourcePreview(docsDragPreview("Text payload")),
			ui.DragSourceOnEvent(func(ctx *ui.Context, event ui.DragSourceEvent) {
				recordDocsDragEvent(state, event)
			}),
		),
		ui.VSpacerElement(8),
		docsDragSourceCard(
			"File URI",
			filepath.Base(sampleFile),
			th,
			ui.NRGBA(240, 253, 244, 255),
			ui.NRGBA(22, 163, 74, 255),
			ui.DragSourceFiles(sampleFile),
			ui.DragSourceOperations(ui.DragOperationCopy, ui.DragOperationLink),
			ui.DragSourceOnEvent(func(ctx *ui.Context, event ui.DragSourceEvent) {
				recordDocsDragEvent(state, event)
			}),
		),
		ui.VSpacerElement(8),
		docsDragSourceCard(
			"JSON",
			"application/json + OnRequest",
			th,
			ui.NRGBA(254, 249, 195, 255),
			ui.NRGBA(202, 138, 4, 255),
			ui.DragSourceData("application/json", []byte(`{"source":"docs_browser","kind":"drag_source"}`)),
			ui.DragSourceOperations(ui.DragOperationCopy, ui.DragOperationMove),
			ui.DragSourceOnRequest(func(ctx *ui.Context, event ui.DragSourceEvent) {
				recordDocsDragRequest(state, event)
			}),
			ui.DragSourceOnEvent(func(ctx *ui.Context, event ui.DragSourceEvent) {
				recordDocsDragEvent(state, event)
			}),
		),
		ui.VSpacerElement(8),
		docsDragSourceCard(
			"Payloads",
			"application/x-fluxui-doc",
			th,
			ui.NRGBA(250, 245, 255, 255),
			ui.NRGBA(147, 51, 234, 255),
			ui.DragSourcePayloads(
				ui.DragPayload{Type: "application/x-fluxui-doc", Data: []byte(`{"doc":"drag_drop","payload":"custom"}`)},
				ui.DragPayload{Type: "text/plain", Data: []byte("custom payload fallback")},
			),
			ui.DragSourceOperations(ui.DragOperationCopy),
			ui.DragSourceOnEvent(func(ctx *ui.Context, event ui.DragSourceEvent) {
				recordDocsDragEvent(state, event)
			}),
		),
		ui.VSpacerElement(8),
		docsDragSourceCard(
			"Disabled",
			"DragSourceDisabled(true)",
			th,
			ui.NRGBA(248, 250, 252, 255),
			ui.NRGBA(100, 116, 139, 255),
			ui.DragSourceText("disabled"),
			ui.DragSourceDisabled(true),
		),
	)
}

func docsDropTargetPanel(state docsDragDropStateHandle, current docsDragDropState, th *ui.Theme) ui.Element {
	return ui.ColumnElement(
		ui.TextElement("Targets", ui.TextSize(15), ui.TextColor(th.Colors.OnSurface)),
		ui.VSpacerElement(8),
		docsDropTargetCard(
			"Accept text, files, JSON, custom MIME",
			current.MainActive,
			th,
			func(ctx *ui.Context, event ui.DropEvent) {
				recordDocsDropEvent(state, "main drop", event)
			},
			ui.DropTargetTypes(
				"text/uri-list",
				"text/plain",
				"text/plain;charset=utf-8",
				"application/json",
				"application/x-fluxui-doc",
			),
			ui.DropTargetOperation(ui.DragOperationCopy),
			ui.DropTargetOnActiveChange(func(ctx *ui.Context, event ui.DropTargetStateEvent) {
				recordDocsTargetActive(state, true, event)
			}),
			ui.DropTargetOnError(func(ctx *ui.Context, event ui.DropEvent) {
				recordDocsDropEvent(state, "main error", event)
			}),
		),
		ui.VSpacerElement(8),
		docsDropTargetCard(
			"Small maxBytes target",
			current.SmallActive,
			th,
			func(ctx *ui.Context, event ui.DropEvent) {
				recordDocsDropEvent(state, "limited drop", event)
			},
			ui.DropTargetTypes("application/json", "application/x-fluxui-doc", "text/plain"),
			ui.DropTargetMaxBytes(16),
			ui.DropTargetOperation(ui.DragOperationMove),
			ui.DropTargetOnActiveChange(func(ctx *ui.Context, event ui.DropTargetStateEvent) {
				recordDocsTargetActive(state, false, event)
			}),
			ui.DropTargetOnError(func(ctx *ui.Context, event ui.DropEvent) {
				recordDocsDropEvent(state, "limited error", event)
			}),
		),
		ui.VSpacerElement(8),
		ui.DropTargetElement(
			docsPanelCard(
				"Disabled target",
				"DropTargetDisabled(true)",
				false,
				th,
				ui.NRGBA(248, 250, 252, 255),
				ui.NRGBA(100, 116, 139, 255),
			),
			func(ctx *ui.Context, event ui.DropEvent) {},
			ui.DropTargetDisabled(true),
		),
	)
}

func docsDragSourceCard(title string, subtitle string, th *ui.Theme, bg color.NRGBA, accent color.NRGBA, opts ...ui.DragSourceOption) ui.Element {
	return ui.DragSourceElement(
		docsPanelCard(title, subtitle, false, th, bg, accent),
		opts...,
	)
}

func docsDropTargetCard(title string, active bool, th *ui.Theme, onDrop func(ctx *ui.Context, event ui.DropEvent), opts ...ui.DropTargetOption) ui.Element {
	bg := ui.NRGBA(255, 255, 255, 255)
	accent := ui.NRGBA(14, 116, 144, 255)
	if active {
		bg = ui.NRGBA(224, 242, 254, 255)
		accent = ui.NRGBA(2, 132, 199, 255)
	}
	return ui.DropTargetElement(
		docsPanelCard(title, "active state + error callback", active, th, bg, accent),
		onDrop,
		opts...,
	)
}

func docsPanelCard(title string, subtitle string, active bool, th *ui.Theme, bg color.NRGBA, accent color.NRGBA) ui.Element {
	titleColor := ui.NRGBA(15, 23, 42, 255)
	border := ui.Border{Width: 1, Color: ui.NRGBA(203, 213, 225, 255)}
	if th != nil {
		bg = th.Colors.SurfaceContainerLow
		titleColor = th.Colors.OnSurface
		border = ui.Border{Width: 1, Color: th.Colors.OutlineVariant}
	}
	if active {
		border = ui.Border{Width: 2, Color: accent}
	}
	return ui.ContainerDecorationElement(
		ui.Bg(bg).
			WithPad(ui.All(12)).
			WithRad(8).
			WithBorder(border),
		ui.FixedHeightElement(
			62,
			ui.ColumnElement(
				ui.TextElement(title, ui.TextSize(13), ui.TextColor(titleColor)),
				ui.PaddingElement(
					ui.Insets{Top: 5},
					ui.TextElement(subtitle, ui.TextSize(11), ui.TextColor(accent)),
				),
			),
		),
	)
}

func docsDragPreview(label string) ui.Widget {
	return ui.Container(
		ui.Style{
			Background: ui.NRGBA(15, 23, 42, 235),
			Padding:    ui.Symmetric(6, 10),
			Radius:     8,
		},
		ui.Text(label, ui.TextSize(12), ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
	)
}

func docsDragDropEventLog(events []string, th *ui.Theme) ui.Element {
	textColor := ui.NRGBA(51, 65, 85, 255)
	bg := ui.NRGBA(241, 245, 249, 255)
	if th != nil {
		textColor = th.TextColor
		bg = th.Colors.SurfaceVariant
	}

	items := make([]ui.Element, 0, len(events)+1)
	items = append(items, ui.TextElement("Events", ui.TextSize(15), ui.TextColor(textColor)))
	if len(events) == 0 {
		events = []string{"ready"}
	}
	for _, event := range events {
		items = append(items,
			ui.PaddingElement(
				ui.Insets{Top: 5},
				ui.TextElement(event, ui.TextSize(11), ui.TextColor(textColor)),
			),
		)
	}

	return ui.ContainerDecorationElement(
		ui.Bg(bg).WithPad(ui.All(10)).WithRad(8),
		ui.FixedHeightElement(112, ui.ScrollViewElement(ui.ColumnElement(items...), ui.ScrollVertical(true))),
	)
}

func recordDocsTargetActive(state docsDragDropStateHandle, mainTarget bool, event ui.DropTargetStateEvent) {
	next := state.Value()
	if mainTarget {
		next.MainActive = event.Active
	} else {
		next.SmallActive = event.Active
	}
	label := "target inactive"
	if event.Active {
		label = "target active: " + strings.Join(event.Types, ", ")
	}
	next.Events = prependDocsDragDropEvent(next.Events, label)
	state.Set(next)
}

func recordDocsDragRequest(state docsDragDropStateHandle, event ui.DragSourceEvent) {
	recordDocsDragDropEvent(state, fmt.Sprintf("request callback type=%s bytes=%d", event.Type, len(event.Data)))
}

func recordDocsDragEvent(state docsDragDropStateHandle, event ui.DragSourceEvent) {
	label := fmt.Sprintf("source %s op=%s", docsDragSourceEventLabel(event.Kind), event.Operation)
	if event.Type != "" {
		label += " type=" + event.Type
	}
	if len(event.Data) > 0 {
		label += fmt.Sprintf(" bytes=%d", len(event.Data))
	}
	if event.Err != nil {
		label += " err=" + event.Err.Error()
	}
	recordDocsDragDropEvent(state, label)
}

func docsDragSourceEventLabel(kind ui.DragSourceEventKind) string {
	switch kind {
	case ui.DragSourceEventStarted:
		return "started"
	case ui.DragSourceEventRequested:
		return "requested"
	case ui.DragSourceEventCompleted:
		return "completed"
	case ui.DragSourceEventCancelled:
		return "cancelled"
	default:
		return string(kind)
	}
}

func recordDocsDropEvent(state docsDragDropStateHandle, prefix string, event ui.DropEvent) {
	next := state.Value()
	next.MainActive = false
	next.SmallActive = false
	next.Events = prependDocsDragDropEvent(next.Events, formatDocsDropEvent(prefix, event))
	state.Set(next)
}

func recordDocsDragDropEvent(state docsDragDropStateHandle, message string) {
	next := state.Value()
	next.Events = prependDocsDragDropEvent(next.Events, message)
	state.Set(next)
}

func formatDocsDropEvent(prefix string, event ui.DropEvent) string {
	label := fmt.Sprintf("%s type=%s op=%s bytes=%d", prefix, event.Type, event.Operation, len(event.Data))
	if len(event.Paths) > 0 {
		label += " paths=" + strings.Join(event.Paths, ", ")
	} else if event.Text != "" {
		label += " text=" + compactDocsPayloadText(event.Text)
	}
	if event.Err != nil {
		label += " err=" + event.Err.Error()
	}
	return label
}

func compactDocsPayloadText(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	if len(text) > 52 {
		return text[:52] + "..."
	}
	return text
}

func prependDocsDragDropEvent(events []string, event string) []string {
	next := append([]string{event}, events...)
	if len(next) > 8 {
		next = next[:8]
	}
	return next
}

func docsDragSampleFile() string {
	return cachedDocsDragSampleFile()
}
