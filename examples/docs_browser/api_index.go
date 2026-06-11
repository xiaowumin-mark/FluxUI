package main

import (
	"context"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsAPIIndexSection(apis []string, copyStatus docsStringState, th *ui.Theme) ui.Element {
	if th == nil {
		th = docsBrowserTheme(defaultDocsThemeSeed, false)
	}
	items := make([]ui.Element, 0, len(apis)+2)
	header := []ui.Element{
		ui.TextElement("API 索引", ui.TextSize(17), ui.TextColor(th.Colors.OnSurface)),
		ui.HSpacerElement(8),
		ui.ContainerDecorationElement(
			ui.Bg(th.Colors.SurfaceContainer).WithPad(ui.Symmetric(3, 8)).WithRad(6),
			ui.TextElement(apiCountLabel(len(apis)), ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
		),
		ui.ExpandedElement(ui.SpacerElement(0, 0)),
		docsCopyAPIsButton(apis, copyStatus),
	}
	if copyStatus != nil && copyStatus.Value() != "" {
		header = append(header,
			ui.HSpacerElement(8),
			ui.TextElement(copyStatus.Value(), ui.TextSize(11), ui.TextColor(th.Colors.Primary)),
		)
	}
	items = append(items, ui.RowElement(header...))
	items = append(items, ui.VSpacerElement(8))
	for _, api := range apis {
		items = append(items, docsAPIIndexRow(api, copyStatus, th))
	}
	return ui.ColumnElement(items...)
}

func docsAPIIndexRow(api string, copyStatus docsStringState, th *ui.Theme) ui.Element {
	return ui.PaddingElement(
		ui.Insets{Bottom: 6},
		ui.ContainerDecorationElement(
			ui.Bg(th.Colors.SurfaceContainerLow).
				WithPad(ui.Insets{Top: 6, Right: 8, Bottom: 6, Left: 10}).
				WithRad(6).
				WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
			ui.RowElement(
				ui.ExpandedElement(ui.TextElement(api, ui.TextSize(12), ui.TextColor(th.Colors.OnSurface))),
				ui.HSpacerElement(8),
				ui.TextButtonElement(
					ui.TextElement("Copy", ui.TextSize(11)),
					ui.ButtonPadding(ui.Symmetric(3, 7)),
					ui.OnClick(func(ctx *ui.Context) {
						copyDocsAPI(api, copyStatus)
					}),
				),
			),
		),
	)
}

func docsCopyAPIsButton(apis []string, copyStatus docsStringState) ui.Element {
	return ui.OutlinedButtonElement(
		ui.TextElement("复制全部", ui.TextSize(12)),
		ui.ButtonPadding(ui.Symmetric(5, 10)),
		ui.OnClick(func(ctx *ui.Context) {
			copyDocsAPIs(apis, copyStatus)
		}),
	)
}

func copyDocsAPI(api string, copyStatus docsStringState) {
	if copyStatus == nil {
		return
	}
	if err := system.WriteClipboardText(context.Background(), api); err != nil {
		copyStatus.Set("复制失败")
		return
	}
	copyStatus.Set("已复制 API")
}

func copyDocsAPIs(apis []string, copyStatus docsStringState) {
	if copyStatus == nil {
		return
	}
	text := ""
	for i, api := range apis {
		if i > 0 {
			text += "\n"
		}
		text += api
	}
	if err := system.WriteClipboardText(context.Background(), text); err != nil {
		copyStatus.Set("复制失败")
		return
	}
	copyStatus.Set("已复制全部 API")
}

func apiCountLabel(count int) string {
	if count == 1 {
		return "1 API"
	}
	return itoa(count) + " APIs"
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for v > 0 {
		pos--
		buf[pos] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[pos:])
}
