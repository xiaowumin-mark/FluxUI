package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

const docsAPISearchPreviewLimit = 5

type docsAPISearchMatch struct {
	DocID    string
	DocTitle string
	API      string
}

func docsAPISearchMatches(docs []widgetDoc, keyword string, limit int) ([]docsAPISearchMatch, int) {
	terms := docsSearchTerms(keyword)
	if len(terms) == 0 {
		return nil, 0
	}
	if limit <= 0 {
		limit = docsAPISearchPreviewLimit
	}

	matches := make([]docsAPISearchMatch, 0, limit)
	total := 0
	for _, doc := range docs {
		for _, api := range doc.Meta.APIs {
			if !docsMatchesSearchTerms(api, terms) {
				continue
			}
			total++
			if len(matches) >= limit {
				continue
			}
			matches = append(matches, docsAPISearchMatch{
				DocID:    doc.Meta.ID,
				DocTitle: doc.Meta.Title,
				API:      api,
			})
		}
	}
	return matches, total
}

func docsAPISearchSummary(docs []widgetDoc, keyword string, selectedDocID docsStringState, th *ui.Theme) ui.Element {
	if th == nil {
		th = docsBrowserTheme(defaultDocsThemeSeed, false)
	}
	if strings.TrimSpace(keyword) == "" {
		return ui.SpacerElement(0, 0)
	}

	matches, total := docsAPISearchMatches(docs, keyword, docsAPISearchPreviewLimit)
	rows := make([]ui.Element, 0, len(matches)+4)
	rows = append(rows,
		ui.RowElement(
			ui.TextElement("API 匹配", ui.TextSize(12), ui.TextColor(th.Colors.OnSurface)),
			ui.ExpandedElement(ui.SpacerElement(0, 0)),
			ui.ContainerDecorationElement(
				ui.Bg(th.Colors.PrimaryContainer).WithPad(ui.Symmetric(2, 7)).WithRad(6),
				ui.TextElement(apiCountLabel(total), ui.TextSize(10), ui.TextColor(th.Colors.OnPrimaryContainer)),
			),
		),
		ui.VSpacerElement(7),
	)

	if total == 0 {
		rows = append(rows,
			ui.TextElement("没有 API 签名匹配此搜索。", ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
		)
		return docsAPISearchSummaryFrame(rows, th)
	}

	for _, match := range matches {
		rows = append(rows, docsAPISearchResultRow(match, selectedDocID, th))
	}
	if total > len(matches) {
		rows = append(rows,
			ui.PaddingElement(
				ui.Insets{Top: 3},
				ui.TextElement(
					fmt.Sprintf("显示 %d 个 API 匹配中的前 %d 个。", total, len(matches)),
					ui.TextSize(10),
					ui.TextColor(th.Colors.OnSurfaceVariant),
				),
			),
		)
	}
	return docsAPISearchSummaryFrame(rows, th)
}

func docsAPISearchSummaryFrame(rows []ui.Element, th *ui.Theme) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.SurfaceContainer).
			WithPad(ui.All(10)).
			WithRad(8).
			WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
		ui.ColumnElement(rows...),
	)
}

func docsAPISearchResultRow(match docsAPISearchMatch, selectedDocID docsStringState, th *ui.Theme) ui.Element {
	title := match.DocTitle
	if title == "" {
		title = match.DocID
	}
	return ui.PaddingElement(
		ui.Insets{Bottom: 5},
		ui.FillWidthElement(
			ui.ButtonElement(
				ui.FillWidthElement(
					ui.ContainerDecorationElement(
						ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.Symmetric(6, 8)).WithRad(6),
						ui.ColumnElement(
							ui.TextElement(docsAPIShortSignature(match.API, 48), ui.TextSize(11), ui.TextColor(th.Colors.OnSurface)),
							ui.VSpacerElement(2),
							ui.TextElement(title, ui.TextSize(10), ui.TextColor(th.Colors.OnSurfaceVariant)),
						),
					),
				),
				ui.ButtonPadding(ui.All(0)),
				ui.ButtonBackground(ui.NRGBA(0, 0, 0, 0)),
				ui.OnClick(func(ctx *ui.Context) {
					if selectedDocID != nil {
						selectedDocID.Set(match.DocID)
					}
				}),
			),
		),
	)
}

func docsAPIShortSignature(api string, limit int) string {
	api = strings.TrimSpace(api)
	if limit <= 0 || utf8.RuneCountInString(api) <= limit {
		return api
	}
	runes := []rune(api)
	if limit <= 3 {
		return string(runes[:limit])
	}
	return string(runes[:limit-3]) + "..."
}
