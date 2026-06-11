package main

import (
	"sort"
	"strings"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

const docsAllCategories = ""

func docsCategories(docs []widgetDoc) []string {
	counts := docsCategoryCounts(docs)
	out := make([]string, 0, 12)
	for category := range counts {
		out = append(out, category)
	}
	sort.Slice(out, func(i, j int) bool {
		leftRank := categoryRank(out[i])
		rightRank := categoryRank(out[j])
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		return out[i] < out[j]
	})
	return out
}

func docsCategoryCounts(docs []widgetDoc) map[string]int {
	counts := make(map[string]int, 12)
	for _, doc := range docs {
		counts[normalizedDocsCategory(doc.Meta.Category)]++
	}
	return counts
}

func normalizedDocsCategory(category string) string {
	category = strings.TrimSpace(category)
	if category == "" {
		return "Uncategorized"
	}
	return category
}

func filterDocsByCategory(docs []widgetDoc, category string) []widgetDoc {
	category = strings.TrimSpace(category)
	if category == docsAllCategories {
		return docs
	}
	out := make([]widgetDoc, 0, len(docs))
	for _, doc := range docs {
		if normalizedDocsCategory(doc.Meta.Category) == category {
			out = append(out, doc)
		}
	}
	return out
}

func docsCategoryFilterBar(docs []widgetDoc, selected docsStringState, th *ui.Theme) ui.Element {
	categories := docsCategories(docs)
	counts := docsCategoryCounts(docs)
	items := make([]ui.Element, 0, len(categories)+1)
	items = append(items, docsCategoryChip(docsCategoryLabel("All", len(docs)), docsAllCategories, selected, th))
	for _, category := range categories {
		items = append(items, ui.HSpacerElement(6))
		items = append(items, docsCategoryChip(docsCategoryLabel(category, counts[category]), category, selected, th))
	}
	return ui.ScrollViewElement(
		ui.RowElement(items...),
		ui.ScrollHorizontal(true),
		ui.ScrollVertical(false),
	)
}

func docsCategoryLabel(label string, count int) string {
	return label + " " + itoa(count)
}

func docsCategoryChip(label string, value string, selected docsStringState, th *ui.Theme) ui.Element {
	active := selected.Value() == value
	return ui.FilterChipElement(
		label,
		ui.ChipSelected(active),
		ui.ChipOnClick(func(ctx *ui.Context) {
			selected.Set(value)
		}),
		func() ui.ChipOption {
			if !active {
				return ui.ChipForeground(th.Colors.OnSurfaceVariant)
			}
			return ui.ChipForeground(th.Colors.OnSecondaryContainer)
		}(),
	)
}
