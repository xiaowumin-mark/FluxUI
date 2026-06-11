package main

import (
	"fmt"
	"image/color"
	"path/filepath"
	"strings"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type dragDropState struct {
	Active bool
	Events []string
}

func app(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	state := ui.UseState(ctx, dragDropState{
		Events: []string{"Drop text, files, or MIME payloads into the target."},
	})
	probe := system.ProbeDragAndDrop(nil)

	drop := state.Value()
	targetBg := th.Colors.SurfaceContainer
	if drop.Active {
		targetBg = ui.NRGBA(219, 234, 254, 255)
	}

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(18)),
		ui.ScrollViewElement(ui.ColumnElement(
			ui.TextElement("Drag & Drop Showcase", ui.TextSize(22)),
			ui.VSpacerElement(6),
			ui.TextElement(probeText(probe), ui.TextSize(13), ui.TextColor(th.SurfaceMuted)),
			ui.VSpacerElement(16),
			ui.RowElement(
				ui.ExpandedElement(sourcePanel(state)),
				ui.HSpacerElement(14),
				ui.ExpandedElement(dropPanel(state, targetBg)),
			),
			ui.VSpacerElement(16),
			ui.TextElement("Events", ui.TextSize(16)),
			ui.VSpacerElement(8),
			ui.ContainerDecorationElement(
				ui.Bg(th.Colors.SurfaceVariant).WithPad(ui.All(12)).WithRad(8),
				ui.ColumnElement(eventElements(drop.Events)...),
			),
		)),
	)
}

func sourcePanel(state interface {
	Value() dragDropState
	Set(dragDropState)
}) ui.Element {
	sampleFile, _ := filepath.Abs("docs/README.md")
	return ui.ColumnElement(
		ui.TextElement("Sources", ui.TextSize(16)),
		ui.VSpacerElement(8),
		dragCard(
			"Drag text",
			ui.DragSourceText("FluxUI drag text payload"),
			ui.DragSourceOperations(ui.DragOperationCopy),
			ui.DragSourceOnEvent(func(ctx *ui.Context, event ui.DragSourceEvent) {
				recordDragEvent(state, event)
			}),
		),
		ui.VSpacerElement(10),
		dragCard(
			"Drag file URI",
			ui.DragSourceFiles(sampleFile),
			ui.DragSourceOperations(ui.DragOperationCopy, ui.DragOperationLink),
			ui.DragSourceOnEvent(func(ctx *ui.Context, event ui.DragSourceEvent) {
				recordDragEvent(state, event)
			}),
		),
		ui.VSpacerElement(10),
		dragCard(
			"Drag JSON MIME",
			ui.DragSourceData("application/json", []byte(`{"source":"FluxUI","kind":"demo"}`)),
			ui.DragSourceOperations(ui.DragOperationCopy),
			ui.DragSourceOnEvent(func(ctx *ui.Context, event ui.DragSourceEvent) {
				recordDragEvent(state, event)
			}),
		),
	)
}

func dragCard(label string, opts ...ui.DragSourceOption) ui.Element {
	return ui.DragSourceElement(
		ui.ContainerDecorationElement(
			ui.Bg(ui.NRGBA(245, 247, 250, 255)).WithPad(ui.All(14)).WithRad(8),
			ui.TextElement(label),
		),
		opts...,
	)
}

func dropPanel(state interface {
	Value() dragDropState
	Set(dragDropState)
}, bg color.NRGBA) ui.Element {
	return ui.ColumnElement(
		ui.TextElement("Target", ui.TextSize(16)),
		ui.VSpacerElement(8),
		ui.DropTargetElement(
			ui.ContainerDecorationElement(
				ui.Bg(bg).WithPad(ui.All(18)).WithRad(8),
				ui.FixedHeightElement(180, ui.CenterElement(ui.TextElement("Drop here"))),
			),
			func(ctx *ui.Context, event ui.DropEvent) {
				recordDropEvent(state, event)
			},
			ui.DropTargetTypes("text/uri-list", "text/plain", "text/plain;charset=utf-8", "application/json"),
			ui.DropTargetOperation(ui.DragOperationCopy),
			ui.DropTargetOnActiveChange(func(ctx *ui.Context, event ui.DropTargetStateEvent) {
				next := state.Value()
				next.Active = event.Active
				if event.Active {
					next.Events = prependEvent(next.Events, "target active: "+strings.Join(event.Types, ", "))
				} else {
					next.Events = prependEvent(next.Events, "target inactive")
				}
				state.Set(next)
			}),
			ui.DropTargetOnError(func(ctx *ui.Context, event ui.DropEvent) {
				recordDropEvent(state, event)
			}),
		),
	)
}

func recordDragEvent(state interface {
	Value() dragDropState
	Set(dragDropState)
}, event ui.DragSourceEvent) {
	next := state.Value()
	label := fmt.Sprintf("source %s op=%s", event.Kind, event.Operation)
	if event.Type != "" {
		label += " type=" + event.Type
	}
	if event.Err != nil {
		label += " err=" + event.Err.Error()
	}
	next.Events = prependEvent(next.Events, label)
	state.Set(next)
}

func recordDropEvent(state interface {
	Value() dragDropState
	Set(dragDropState)
}, event ui.DropEvent) {
	next := state.Value()
	next.Active = false
	label := fmt.Sprintf("drop type=%s op=%s bytes=%d", event.Type, event.Operation, len(event.Data))
	if len(event.Paths) > 0 {
		label += " paths=" + strings.Join(event.Paths, ", ")
	} else if event.Text != "" {
		label += " text=" + event.Text
	}
	if event.Err != nil {
		label += " err=" + event.Err.Error()
	}
	next.Events = prependEvent(next.Events, label)
	state.Set(next)
}

func prependEvent(events []string, event string) []string {
	next := append([]string{event}, events...)
	if len(next) > 8 {
		next = next[:8]
	}
	return next
}

func eventElements(events []string) []ui.Element {
	if len(events) == 0 {
		return []ui.Element{ui.TextElement("No events yet")}
	}
	out := make([]ui.Element, 0, len(events))
	for _, event := range events {
		out = append(out, ui.TextElement(event, ui.TextSize(13)))
	}
	return out
}

func probeText(probe system.DragAndDropProbe) string {
	return fmt.Sprintf(
		"status=%s drop=%t source=%t text=%t files=%t custom=%t external-in=%t external-out=%t operations=%v",
		probe.Status,
		probe.SupportsDropTarget,
		probe.SupportsDragSource,
		probe.SupportsText,
		probe.SupportsFiles,
		probe.SupportsCustomMIME,
		probe.SupportsExternalDragIn,
		probe.SupportsExternalDragOut,
		probe.SupportedOperations,
	)
}

func main() {
	_ = ui.RunElement(app, ui.Title("FluxUI Drag & Drop"), ui.Size(760, 520))
}
