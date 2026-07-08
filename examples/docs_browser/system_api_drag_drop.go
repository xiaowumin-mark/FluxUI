package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsSystemDragDropProbeSection(th *ui.Theme) ui.Element {
	return ui.ComponentElement(func(sectionCtx *ui.Context) ui.Element {
		probeSerial := ui.UseState(sectionCtx, 0)
		probe := system.ProbeDragAndDrop(context.Background())
		summary := docsSystemDragDropProbeSummary(probe)

		return docsSystemSection("Drag & Drop Probe", ui.ColumnElement(
			ui.TextElement(summary, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.RowElement(
				docsSystemDragProbeCard("放置目标", probe.SupportsDropTarget, th),
				ui.HSpacerElement(8),
				docsSystemDragProbeCard("拖拽源", probe.SupportsDragSource, th),
				ui.HSpacerElement(8),
				docsSystemDragProbeCard("文本", probe.SupportsText, th),
				ui.HSpacerElement(8),
				docsSystemDragProbeCard("文件", probe.SupportsFiles, th),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				docsSystemDragProbeCard("Custom MIME", probe.SupportsCustomMIME, th),
				ui.HSpacerElement(8),
				docsSystemDragProbeCard("外部拖入", probe.SupportsExternalDragIn, th),
				ui.HSpacerElement(8),
				docsSystemDragProbeCard("外部拖出", probe.SupportsExternalDragOut, th),
				ui.HSpacerElement(8),
				docsSystemDragProbeCard("可用", probe.Available(), th),
			),
			ui.VSpacerElement(8),
			ui.TextElement("操作："+docsSystemDragOperationsLabel(probe.SupportedOperations), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(systemActionButton("刷新探测", false, func(ctx *ui.Context) {
					probeSerial.Set(probeSerial.Value() + 1)
				})),
				ui.HSpacerElement(8),
				ui.ExpandedElement(systemActionButton("打开拖拽文档", false, func(ctx *ui.Context) {
					_ = system.OpenURL(context.Background(), "https://github.com/xiaowumin-mark/FluxUI")
				})),
			),
			ui.VSpacerElement(4),
			ui.TextElement(fmt.Sprintf("刷新计数：%d", probeSerial.Value()), ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
		), th)
	})
}

func docsSystemDragProbeCard(label string, supported bool, th *ui.Theme) ui.Element {
	bg := th.Colors.SurfaceContainer
	fg := th.Colors.OnSurfaceVariant
	value := "否"
	if supported {
		bg = th.Colors.PrimaryContainer
		fg = th.Colors.OnPrimaryContainer
		value = "是"
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
				ui.TextElement(value, ui.TextSize(11), ui.TextColor(fg)),
			),
		),
	)
}

func docsSystemDragDropProbeSummary(probe system.DragAndDropProbe) string {
	text := fmt.Sprintf("status=%s supported=%v available=%v", probe.Status, probe.Supported(), probe.Available())
	if probe.Err != nil {
		text += " err=" + probe.Err.Error()
	}
	return text
}

func docsSystemDragOperationsLabel(operations []system.DragDropOperation) string {
	if len(operations) == 0 {
		return "无"
	}
	labels := make([]string, 0, len(operations))
	for _, op := range operations {
		labels = append(labels, string(op))
	}
	return strings.Join(labels, ", ")
}
