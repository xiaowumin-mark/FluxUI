package main

import (
	"context"
	"fmt"
	"image/color"
	"regexp"
	"strings"

	"github.com/xiaowumin-mark/FluxUI/system"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

var markdownLinkPattern = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

func renderMarkdownDocument(content string, th *ui.Theme, copyState docsStringState) []ui.Element {
	if th == nil {
		th = docsBrowserTheme(defaultDocsThemeSeed, false)
	}

	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	widgets := make([]ui.Element, 0, len(lines)+12)
	inCode := false
	codeLang := ""
	codeLines := make([]string, 0, 12)

	flushCode := func() {
		if len(codeLines) == 0 {
			return
		}
		widgets = append(widgets, markdownCodeBlock(codeLang, strings.Join(codeLines, "\n"), th, copyState))
		widgets = append(widgets, ui.VSpacerElement(12))
		codeLang = ""
		codeLines = codeLines[:0]
	}

	for i := 0; i < len(lines); i++ {
		raw := lines[i]
		line := strings.TrimRight(raw, " \t")
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			if inCode {
				inCode = false
				flushCode()
			} else {
				inCode = true
				codeLang = strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
				codeLines = codeLines[:0]
			}
			continue
		}

		if inCode {
			codeLines = append(codeLines, line)
			continue
		}

		if isMarkdownTableStart(lines, i) {
			tableLines := []string{trimmed}
			i++
			for i < len(lines) && isMarkdownTableLine(strings.TrimSpace(lines[i])) {
				tableLines = append(tableLines, strings.TrimSpace(lines[i]))
				i++
			}
			i--
			widgets = append(widgets, markdownTable(tableLines, th))
			widgets = append(widgets, ui.VSpacerElement(12))
			continue
		}

		widgets = append(widgets, renderMarkdownLine(trimmed, line, th)...)
	}

	if inCode {
		flushCode()
	}
	if len(widgets) == 0 {
		return []ui.Element{ui.TextElement("No markdown content", ui.TextSize(13), ui.TextColor(th.Colors.OnSurfaceVariant))}
	}
	return widgets
}

func renderMarkdownLine(trimmed string, line string, th *ui.Theme) []ui.Element {
	if trimmed == "" {
		return []ui.Element{ui.VSpacerElement(8)}
	}
	if trimmed == "---" || trimmed == "***" {
		return []ui.Element{
			ui.PaddingElement(
				ui.Insets{Top: 8, Bottom: 8},
				ui.DividerElement(ui.DividerColor(th.Colors.OutlineVariant)),
			),
		}
	}

	if level, text, ok := markdownHeadingInfo(trimmed); ok {
		return markdownHeading(text, markdownHeadingSize(level), level, th)
	}
	if marker, text, indent, ok := markdownListInfo(line); ok {
		return markdownListItem(marker, text, indent, th)
	}

	switch {
	case strings.HasPrefix(trimmed, "> "):
		return []ui.Element{
			ui.ContainerDecorationElement(
				ui.Bg(th.Colors.SurfaceContainerHigh).
					WithPad(ui.Insets{Top: 8, Right: 10, Bottom: 8, Left: 10}).
					WithRad(8).
					WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
				ui.TextElement(markdownInlineText(strings.TrimSpace(strings.TrimPrefix(trimmed, "> "))), ui.TextSize(13), ui.TextColor(th.Colors.OnSurfaceVariant)),
			),
			ui.VSpacerElement(8),
		}
	default:
		return []ui.Element{
			ui.TextElement(markdownInlineText(line), ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(5),
		}
	}
}

func markdownHeading(text string, size float32, level int, th *ui.Theme) []ui.Element {
	top := float32(8)
	bottom := float32(4)
	if level <= 2 {
		top = 12
		bottom = 6
	}
	return []ui.Element{
		ui.PaddingElement(
			ui.Insets{Top: top, Bottom: bottom},
			ui.TextElement(markdownInlineText(text), ui.TextSize(size), ui.TextColor(th.Colors.OnSurface)),
		),
	}
}

func markdownListItem(marker string, text string, indent int, th *ui.Theme) []ui.Element {
	left := float32(indent * 18)
	return []ui.Element{
		ui.PaddingElement(
			ui.Insets{Left: left},
			ui.RowElement(
				ui.FixedWidthElement(30, ui.TextElement(marker, ui.TextSize(13), ui.TextColor(th.Colors.Primary))),
				ui.ExpandedElement(ui.TextElement(markdownInlineText(text), ui.TextSize(13), ui.TextColor(th.Colors.OnSurface))),
			),
		),
		ui.VSpacerElement(5),
	}
}

func markdownCodeBlock(lang string, code string, th *ui.Theme, copyState docsStringState) ui.Element {
	label := strings.TrimSpace(lang)
	if label == "" {
		label = "code"
	}

	status := ""
	if copyState != nil && copyState.Value() != "" {
		status = copyState.Value()
	}

	headerItems := []ui.Element{
		ui.TextElement(label, ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
		ui.ExpandedElement(ui.SpacerElement(0, 0)),
	}
	if status != "" {
		headerItems = append(headerItems,
			ui.PaddingElement(
				ui.Insets{Right: 8},
				ui.TextElement(status, ui.TextSize(11), ui.TextColor(th.Colors.Primary)),
			),
		)
	}
	headerItems = append(headerItems, markdownCopyButton(label, code, copyState))

	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.SurfaceContainerHigh).
			WithPad(ui.All(0)).
			WithRad(8).
			WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
		ui.ColumnElement(
			ui.ContainerDecorationElement(
				ui.Bg(th.Colors.SurfaceContainer).WithPad(ui.Insets{Top: 6, Right: 8, Bottom: 6, Left: 10}).WithRad(8),
				ui.RowElement(headerItems...),
			),
			ui.ContainerDecorationElement(
				ui.Bg(color.NRGBA{R: 15, G: 23, B: 42, A: 255}).WithPad(ui.All(10)).WithRad(8),
				ui.ScrollViewElement(
					ui.TextElement(code, ui.TextSize(12), ui.TextColor(color.NRGBA{R: 226, G: 232, B: 240, A: 255})),
					ui.ScrollHorizontal(true),
					ui.ScrollVertical(false),
				),
			),
		),
	)
}

func markdownTable(lines []string, th *ui.Theme) ui.Element {
	rows := make([][]string, 0, len(lines))
	for i, line := range lines {
		if i == 1 && isMarkdownTableSeparator(line) {
			continue
		}
		cells := markdownTableCells(line)
		if len(cells) > 0 {
			rows = append(rows, cells)
		}
	}
	if len(rows) == 0 {
		return ui.SpacerElement(0, 0)
	}

	items := make([]ui.Element, 0, len(rows))
	for rowIndex, row := range rows {
		bg := th.Colors.SurfaceContainerLow
		textColor := th.Colors.OnSurface
		if rowIndex == 0 {
			bg = th.Colors.SurfaceContainerHigh
			textColor = th.Colors.OnSurface
		}

		cells := make([]ui.Element, 0, len(row))
		for _, cell := range row {
			cells = append(cells,
				ui.ExpandedElement(
					ui.PaddingElement(
						ui.Insets{Right: 8},
						ui.TextElement(markdownInlineText(cell), ui.TextSize(12), ui.TextColor(textColor)),
					),
				),
			)
		}

		items = append(items,
			ui.ContainerDecorationElement(
				ui.Bg(bg).
					WithPad(ui.Insets{Top: 7, Right: 10, Bottom: 7, Left: 10}).
					WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
				ui.RowElement(cells...),
			),
		)
	}

	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.SurfaceContainerLow).
			WithPad(ui.All(0)).
			WithRad(8).
			WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
		ui.ScrollViewElement(
			ui.ColumnElement(items...),
			ui.ScrollHorizontal(true),
			ui.ScrollVertical(false),
		),
	)
}

func markdownCopyButton(label string, code string, copyState docsStringState) ui.Element {
	return ui.TextButtonElement(
		ui.TextElement("Copy", ui.TextSize(11)),
		ui.ButtonPadding(ui.Symmetric(4, 8)),
		ui.OnClick(func(ctx *ui.Context) {
			if copyState == nil {
				return
			}
			if err := system.WriteClipboardText(context.Background(), code); err != nil {
				copyState.Set("copy failed")
				return
			}
			copyState.Set(fmt.Sprintf("copied %s", label))
		}),
	)
}

func markdownInlineText(text string) string {
	text = markdownLinkPattern.ReplaceAllString(text, `$1 ($2)`)
	text = strings.ReplaceAll(text, "`", "")
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "__", "")
	return strings.TrimSpace(text)
}

func markdownHeadingInfo(text string) (int, string, bool) {
	if !strings.HasPrefix(text, "#") {
		return 0, "", false
	}
	level := 0
	for level < len(text) && level < 6 && text[level] == '#' {
		level++
	}
	if level == 0 || level >= len(text) || text[level] != ' ' {
		return 0, "", false
	}
	return level, strings.TrimSpace(text[level+1:]), true
}

func markdownHeadingSize(level int) float32 {
	switch level {
	case 1:
		return 23
	case 2:
		return 19
	case 3:
		return 16
	case 4:
		return 14
	case 5:
		return 13
	default:
		return 12
	}
}

func markdownListInfo(line string) (string, string, int, bool) {
	trimmed := strings.TrimLeft(line, " \t")
	if trimmed == "" {
		return "", "", 0, false
	}
	indent := markdownListIndentLevel(line)
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		marker := "-"
		text := strings.TrimSpace(trimmed[2:])
		if taskMarker, taskText, ok := markdownTaskInfo(text); ok {
			marker = taskMarker
			text = taskText
		}
		return marker, text, indent, true
	}
	if isMarkdownNumberedList(trimmed) {
		dot := strings.Index(trimmed, ".")
		return trimmed[:dot+1], strings.TrimSpace(trimmed[dot+1:]), indent, true
	}
	return "", "", 0, false
}

func markdownListIndentLevel(line string) int {
	spaces := 0
	for _, r := range line {
		switch r {
		case ' ':
			spaces++
		case '\t':
			spaces += 4
		default:
			return spaces / 2
		}
	}
	return 0
}

func markdownTaskInfo(text string) (string, string, bool) {
	if len(text) < 4 || text[0] != '[' || text[2] != ']' || text[3] != ' ' {
		return "", "", false
	}
	switch text[1] {
	case 'x', 'X':
		return "[x]", strings.TrimSpace(text[4:]), true
	case ' ':
		return "[ ]", strings.TrimSpace(text[4:]), true
	default:
		return "", "", false
	}
}

func isMarkdownNumberedList(text string) bool {
	dot := strings.Index(text, ".")
	if dot <= 0 || dot > 3 || dot+1 >= len(text) || text[dot+1] != ' ' {
		return false
	}
	for _, r := range text[:dot] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isMarkdownTableStart(lines []string, index int) bool {
	if index+1 >= len(lines) {
		return false
	}
	current := strings.TrimSpace(lines[index])
	next := strings.TrimSpace(lines[index+1])
	return isMarkdownTableLine(current) && isMarkdownTableSeparator(next)
}

func isMarkdownTableLine(text string) bool {
	return strings.HasPrefix(text, "|") && strings.HasSuffix(text, "|") && strings.Count(text, "|") >= 2
}

func isMarkdownTableSeparator(text string) bool {
	if !isMarkdownTableLine(text) {
		return false
	}
	for _, cell := range markdownTableCells(text) {
		cell = strings.Trim(cell, " :-")
		if cell != "" {
			return false
		}
	}
	return true
}

func markdownTableCells(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	cells := make([]string, 0, len(parts))
	for _, part := range parts {
		cells = append(cells, strings.TrimSpace(part))
	}
	return cells
}
