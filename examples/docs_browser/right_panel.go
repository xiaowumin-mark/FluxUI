package main

import (
	"fmt"
	"strings"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type docsDemoBuilder func(*widgetDoc) ui.Element

func docsRightPanelContent(
	doc *widgetDoc,
	th *ui.Theme,
	examplePopupOpen docsBoolState,
	apiCopyStatus docsStringState,
	markdownContent []ui.Element,
	buildDemo docsDemoBuilder,
) []ui.Element {
	_ = buildDemo
	if doc == nil {
		return docsRightPanelEmptyState(th)
	}
	if th == nil {
		th = docsBrowserTheme(defaultDocsThemeSeed, false)
	}
	if buildDemo == nil {
		buildDemo = func(*widgetDoc) ui.Element {
			return ui.TextElement("No example configured")
		}
	}

	exampleID := doc.Meta.Example.ID
	if exampleID == "" {
		exampleID = doc.Meta.ID
	}

	content := []ui.Element{
		ui.TextElement(doc.Meta.Title, ui.TextSize(26), ui.TextColor(th.Colors.OnSurface)),
		ui.PaddingElement(
			ui.Insets{Top: 6},
			ui.TextElement(
				fmt.Sprintf("组件ID: %s  |  分类: %s", doc.Meta.ID, doc.Meta.Category),
				ui.TextSize(12),
				ui.TextColor(th.Colors.OnSurfaceVariant),
			),
		),
	}

	if strings.TrimSpace(doc.Meta.Summary) != "" {
		content = append(content,
			ui.PaddingElement(
				ui.Insets{Top: 10},
				ui.TextElement(doc.Meta.Summary, ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
			),
		)
	}

	content = append(content,
		ui.PaddingElement(
			ui.Insets{Top: 16},
			docsExampleSectionHeader(doc.Meta.Title, exampleID, examplePopupOpen, th),
		),
	)

	if len(doc.Meta.APIs) > 0 {
		content = append(content,
			ui.PaddingElement(
				ui.Insets{Top: 14},
				docsAPIIndexSection(doc.Meta.APIs, apiCopyStatus, th),
			),
		)
	}

	content = append(content,
		ui.PaddingElement(
			ui.Insets{Top: 14},
			ui.TextElement("文档正文", ui.TextSize(17), ui.TextColor(th.Colors.OnSurface)),
		),
	)
	if len(markdownContent) == 0 {
		markdownContent = []ui.Element{ui.TextElement("No markdown content", ui.TextSize(13), ui.TextColor(th.Colors.OnSurfaceVariant))}
	}
	for _, item := range markdownContent {
		content = append(content, ui.PaddingElement(ui.Insets{Top: 8}, item))
	}

	return content
}

func docsRightPanelList(content []ui.Element) ui.Element {
	if len(content) == 0 {
		return ui.SpacerElement(0, 0)
	}
	return ui.ListViewElement(
		len(content),
		func(ctx *ui.Context, index int) ui.Element {
			if index < 0 || index >= len(content) {
				return ui.SpacerElement(0, 0)
			}
			return content[index]
		},
		ui.ListVirtualized(true),
	)
}

func docsRightPanelEmptyState(th *ui.Theme) []ui.Element {
	if th == nil {
		th = docsBrowserTheme(defaultDocsThemeSeed, false)
	}
	return []ui.Element{
		ui.ContainerDecorationElement(
			ui.Bg(th.Colors.SurfaceContainerLow).
				WithPad(ui.All(18)).
				WithRad(12).
				WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
			ui.ColumnElement(
				ui.TextElement("No matching docs", ui.TextSize(20), ui.TextColor(th.Colors.OnSurface)),
				ui.VSpacerElement(8),
				ui.TextElement("Try another keyword or category filter.", ui.TextSize(13), ui.TextColor(th.Colors.OnSurfaceVariant)),
			),
		),
	}
}
