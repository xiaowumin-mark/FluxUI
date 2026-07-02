package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsLeftPanel(
	docs []widgetDoc,
	filteredDocs []widgetDoc,
	currentDoc *widgetDoc,
	selectedDocID docsStringState,
	searchKeyword docsStringState,
	categoryFilter docsStringState,
	themeSeed docsStringState,
	themeDark docsBoolState,
	th *ui.Theme,
	docsSource string,
	onlineLoading bool,
	loadErr error,
) ui.Element {
	leftMenuList := docsLeftMenuList(filteredDocs, currentDoc, selectedDocID, th)

	sourceLabel := map[string]string{
		"online": "online",
		"local":  "local",
	}[docsSource]
	if sourceLabel == "" {
		sourceLabel = docsSource
	}
	docCountText := fmt.Sprintf("Loaded %d docs (%s) · showing %d", len(docs), sourceLabel, len(filteredDocs))
	if onlineLoading {
		docCountText = "Local docs are unavailable; loading online docs..."
	}

	return ui.FixedWidthElement(
		300,
		ui.ContainerElement(
			ui.Style{
				Background: th.Colors.SurfaceContainerLow,
				Padding:    ui.All(12),
			},
			ui.ColumnElement(
				ui.TextElement("FluxUI Docs", ui.TextSize(18), ui.TextColor(th.Colors.OnSurface)),
				ui.PaddingElement(
					ui.Insets{Top: 8},
					ui.TextElement(
						docCountText,
						ui.TextSize(12),
						ui.TextColor(th.Colors.OnSurfaceVariant),
					),
				),
				ui.PaddingElement(
					ui.Insets{Top: 10},
					ui.SearchBarElement(
						searchKeyword.Value(),
						ui.SearchBarPlaceholder("Search docs / API"),
						ui.SearchBarInputOptions(ui.InputSingleLine(true)),
						ui.SearchBarOnChange(func(ctx *ui.Context, value string) {
							searchKeyword.Set(value)
						}),
					),
				),
				ui.PaddingElement(
					ui.Insets{Top: 8},
					docsAPISearchSummary(
						filterDocsByCategory(docs, categoryFilter.Value()),
						searchKeyword.Value(),
						selectedDocID,
						th,
					),
				),
				ui.PaddingElement(
					ui.Insets{Top: 10},
					docsCategoryFilterBar(docs, categoryFilter, th),
				),
				ui.PaddingElement(
					ui.Insets{Top: 10},
					docsThemeControls(themeSeed, themeDark, th),
				),
				ui.PaddingElement(
					ui.Insets{Top: 10},
					ui.DividerElement(ui.DividerColor(th.Colors.OutlineVariant)),
				),
				ui.PaddingElement(
					ui.Insets{Top: 10},
					ui.ExpandedElement(
						leftMenuList,
					),
				),
				docsLoadStatusMessage(onlineLoading, loadErr, th),
			),
		),
	)
}

func docsLeftMenuList(
	filteredDocs []widgetDoc,
	currentDoc *widgetDoc,
	selectedDocID docsStringState,
	th *ui.Theme,
) ui.Element {
	menuEntries := buildMenuEntries(filteredDocs)
	if len(menuEntries) == 0 {
		return ui.SpacerElement(0, 0)
	}
	return ui.ListViewElement(
		len(menuEntries),
		func(ctx *ui.Context, index int) ui.Element {
			if index < 0 || index >= len(menuEntries) {
				return ui.SpacerElement(0, 0)
			}
			entry := menuEntries[index]
			if entry.IsCategory {
				return ui.PaddingElement(
					ui.Insets{Top: 10, Bottom: 4},
					ui.TextElement(entry.Category, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
				)
			}
			doc := entry.Doc
			if doc == nil {
				return ui.SpacerElement(0, 0)
			}

			selected := currentDoc != nil && doc.Meta.ID == currentDoc.Meta.ID
			bg := ui.NRGBA(0, 0, 0, 0)
			textColor := th.Colors.OnSurface
			if selected {
				bg = th.Colors.SecondaryContainer
				textColor = th.Colors.OnSecondaryContainer
			}

			return ui.PaddingElement(
				ui.Insets{Bottom: 6},
				ui.FillWidthElement(
					ui.ButtonElement(
						ui.FillWidthElement(
							ui.ContainerElement(
								ui.Style{
									Background: bg,
									Padding:    ui.Symmetric(8, 10),
									Radius:     6,
								},
								ui.TextElement(doc.Meta.Title, ui.TextSize(13), ui.TextColor(textColor)),
							),
						),
						ui.ButtonBackground(ui.NRGBA(0, 0, 0, 0)),
						ui.ButtonPadding(ui.All(0)),
						ui.OnClick(func(ctx *ui.Context) {
							selectedDocID.Set(doc.Meta.ID)
						}),
					),
				),
			)
		},
		ui.ListVirtualized(true),
	)
}

func docsLoadStatusMessage(onlineLoading bool, loadErr error, th *ui.Theme) ui.Element {
	if th == nil {
		th = docsBrowserTheme(defaultDocsThemeSeed, false)
	}
	if onlineLoading {
		msg := "Failed to read local docs; loading GitHub online docs..."
		if loadErr != nil {
			msg += " Reason: " + loadErr.Error()
		}
		return ui.PaddingElement(
			ui.Insets{Top: 8},
			ui.TextElement(
				msg,
				ui.TextSize(11),
				ui.TextColor(th.Colors.Tertiary),
			),
		)
	}
	if loadErr == nil {
		return ui.SpacerElement(0, 0)
	}
	return ui.PaddingElement(
		ui.Insets{Top: 8},
		ui.TextElement(
			"Docs load warning: "+loadErr.Error(),
			ui.TextSize(11),
			ui.TextColor(th.Colors.Error),
		),
	)
}
