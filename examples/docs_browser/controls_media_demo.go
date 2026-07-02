package main

import (
	"fmt"

	"github.com/xiaowumin-mark/FluxUI/icons/md3"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsImageDemo() ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		ref := ui.UseRef(ctx, ui.NewButtonRef())
		clicks := ui.UseState(ctx, 0)
		if ref.Current == nil {
			ref.Current = ui.NewButtonRef()
		}
		src := ui.ImageSource{Path: "examples/assets/sample.png", Label: "sample.png"}
		return ui.ColumnElement(
			ui.RowElement(
				ui.ImageElement(
					src,
					ui.ImageWidth(150),
					ui.ImageHeight(90),
					ui.ImageFitMode(ui.ImageFitContain),
					ui.ImageRadius(8),
					ui.ImageBackground(ui.NRGBA(241, 245, 249, 255)),
					ui.ImageDecoration(ui.BorderDeco(1, ui.NRGBA(203, 213, 225, 255))),
					ui.ImageAttachRef(ref.Current),
					ui.ImageOnClick(func(ctx *ui.Context) {
						clicks.Set(clicks.Value() + 1)
					}),
				),
				ui.PaddingElement(
					ui.Insets{Left: 10},
					ui.ImageElement(
						src,
						ui.ImageWidth(150),
						ui.ImageHeight(90),
						ui.ImageFitMode(ui.ImageFitCover),
						ui.ImageRadius(8),
					),
				),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				docsDemoControlButton("ImageRef.Click", func(ctx *ui.Context) {
					clicks.Set(clicks.Value() + 1)
					ref.Current.Click()
				}),
				ui.HSpacerElement(8),
				ui.TextElement(fmt.Sprintf("image clicks = %d", clicks.Value()), ui.TextSize(12)),
			),
		)
	})
}

func docsIconDemo() ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		ref := ui.UseRef(ctx, ui.NewButtonRef())
		clicks := ui.UseState(ctx, 0)
		if ref.Current == nil {
			ref.Current = ui.NewButtonRef()
		}
		return ui.RowElement(
			ui.IconElement("home", ui.IconSize(20), ui.IconColor(ui.NRGBA(30, 136, 229, 255)), ui.IconAttachRef(ref.Current), ui.IconOnClick(func(ctx *ui.Context) {
				clicks.Set(clicks.Value() + 1)
			})),
			ui.PaddingElement(ui.Insets{Left: 12}, ui.IconElement("search", ui.IconSize(20), ui.IconColor(ui.NRGBA(67, 160, 71, 255)))),
			ui.PaddingElement(ui.Insets{Left: 12}, ui.IconElement("settings", ui.IconSize(20), ui.IconColor(ui.NRGBA(245, 124, 0, 255)))),
			ui.PaddingElement(ui.Insets{Left: 12}, docsDemoControlButton(fmt.Sprintf("clicks %d", clicks.Value()), func(ctx *ui.Context) {
				clicks.Set(clicks.Value() + 1)
				ref.Current.Click()
			})),
		)
	})
}

func docsIconFontsDemo() ui.Element {
	return ui.ColumnElement(
		ui.TextElement(fmt.Sprintf("default icon font: %s", docsDefaultIconFontLabel()), ui.TextSize(12), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
		ui.VSpacerElement(10),
		ui.RowElement(
			ui.IconElement("home", ui.IconSize(24), ui.IconColor(ui.NRGBA(30, 136, 229, 255))),
			ui.HSpacerElement(12),
			ui.IconElement("search", ui.IconSize(24), ui.IconColor(ui.NRGBA(67, 160, 71, 255))),
			ui.HSpacerElement(12),
			ui.IconElement("settings", ui.IconSize(24), ui.IconColor(ui.NRGBA(245, 124, 0, 255))),
			ui.HSpacerElement(12),
			ui.IconElement("favorite", ui.IconUseFont(md3.ID), ui.IconSize(24), ui.IconColor(ui.NRGBA(220, 38, 38, 255))),
		),
		ui.VSpacerElement(10),
		ui.RowElement(
			ui.FilledIconButtonElement(ui.IconElement("add")),
			ui.HSpacerElement(8),
			ui.FilledTonalIconButtonElement(ui.IconElement("notifications")),
			ui.HSpacerElement(8),
			ui.OutlinedIconButtonElement(ui.IconElement("mail")),
		),
	)
}

func docsDefaultIconFontLabel() string {
	font, ok := ui.DefaultIconFont()
	if !ok {
		return "(none)"
	}
	return font.ID + " / " + font.Family
}

func docsListItemDemo(selected docsStringState) ui.Element {
	return ui.FixedWidthElement(
		380,
		ui.ColumnElement(
			ui.ListItemElementWithSlots(
				ui.TextElement("Inbox"),
				ui.TextElement("12 unread messages"),
				ui.IconElement("info"),
				ui.TextElement("12"),
				ui.ListItemSelected(selected.Value() == "inbox"),
				ui.ListItemMinHeight(64),
				ui.ListItemDecoration(ui.Bg(ui.NRGBA(248, 250, 252, 255)).WithRad(10)),
				ui.ListItemOnClick(func(ctx *ui.Context) { selected.Set("inbox") }),
			),
			ui.ListItemElementWithSlots(
				ui.TextElement("Archive"),
				ui.TextElement("Older conversations"),
				ui.IconElement("archive"),
				nil,
				ui.ListItemSelected(selected.Value() == "archive"),
				ui.ListItemOnClick(func(ctx *ui.Context) { selected.Set("archive") }),
			),
			ui.ListItemElement("Compact item", ui.ListItemOnClick(func(ctx *ui.Context) { selected.Set("compact") })),
			ui.ListItemElementWithSlots(
				ui.TextElement("Disabled"),
				ui.TextElement("Unavailable item"),
				ui.IconElement("delete"),
				nil,
				ui.ListItemDisabled(true),
			),
		),
	)
}

func docsIconButtonDemo(selected docsBoolState) ui.Element {
	return ui.RowElement(
		ui.PaddingElement(ui.Insets{Right: 10}, ui.IconButtonElement(ui.IconElement("search"), ui.IconButtonSelected(selected.Value()), ui.IconButtonOnClick(func(ctx *ui.Context) {
			selected.Set(!selected.Value())
		}))),
		ui.PaddingElement(ui.Insets{Right: 10}, ui.FilledIconButtonElement(ui.IconElement("favorite"), ui.IconButtonSelected(true), ui.IconButtonSize(42))),
		ui.PaddingElement(ui.Insets{Right: 10}, ui.FilledTonalIconButtonElement(ui.IconElement("tune"), ui.IconButtonBackground(ui.NRGBA(219, 234, 254, 255)), ui.IconButtonForeground(ui.NRGBA(30, 64, 175, 255)))),
		ui.PaddingElement(ui.Insets{Right: 10}, ui.OutlinedIconButtonElement(ui.IconElement("radio_button_unchecked"), ui.IconButtonDecoration(ui.BorderDeco(1, ui.NRGBA(148, 163, 184, 255))))),
		ui.OutlinedIconButtonElement(ui.IconElement("delete"), ui.IconButtonDisabled(true)),
	)
}

func docsFloatingActionButtonDemo(count docsIntState) ui.Element {
	click := ui.FloatingActionButtonOnClick(func(ctx *ui.Context) {
		count.Set(count.Value() + 1)
	})
	return ui.ColumnElement(
		ui.RowElement(
			ui.PaddingElement(ui.Insets{Right: 12}, ui.SmallFloatingActionButtonElement(ui.IconElement("add"), click)),
			ui.PaddingElement(ui.Insets{Right: 12}, ui.FloatingActionButtonElement(ui.IconElement("add"), click)),
			ui.PaddingElement(ui.Insets{Right: 12}, ui.LargeFloatingActionButtonElement(ui.IconElement("add"), click, ui.FloatingActionButtonBackground(ui.NRGBA(37, 99, 235, 255)), ui.FloatingActionButtonForeground(ui.NRGBA(255, 255, 255, 255)))),
			ui.PaddingElement(ui.Insets{Right: 12}, ui.ExtendedFloatingActionButtonElement(ui.IconElement("add"), ui.TextElement("Create"), click, ui.FloatingActionButtonDecoration(ui.Bg(ui.NRGBA(220, 252, 231, 255)).WithRad(16)))),
			ui.FloatingActionButtonElement(ui.IconElement("close"), ui.FloatingActionButtonDisabled(true)),
		),
		ui.PaddingElement(ui.Insets{Top: 12}, ui.TextElement(fmt.Sprintf("FAB clicks: %d", count.Value()), ui.TextSize(13))),
	)
}

func docsBadgeDemo() ui.Element {
	return ui.RowElement(
		ui.PaddingElement(
			ui.Insets{Right: 24},
			ui.BadgeElement(
				ui.IconButtonElement(ui.IconElement("mail")),
				"3",
				ui.BadgeBackground(ui.NRGBA(220, 38, 38, 255)),
				ui.BadgeForeground(ui.NRGBA(255, 255, 255, 255)),
				ui.BadgeOffset(3, -3),
			),
		),
		ui.PaddingElement(
			ui.Insets{Right: 24},
			ui.BadgeElement(
				ui.IconButtonElement(ui.IconElement("notifications")),
				"",
				ui.BadgeVisible(true),
				ui.BadgeDecoration(ui.Bg(ui.NRGBA(37, 99, 235, 255)).WithRad(999)),
			),
		),
		ui.BadgeElement(ui.TextElement("Hidden"), "0", ui.BadgeVisible(false)),
	)
}

func docsChipDemo(selected docsBoolState) ui.Element {
	return ui.ColumnElement(
		ui.TextElement("Assist chips", ui.TextSize(13)),
		ui.VSpacerElement(6),
		ui.RowElement(
			ui.PaddingElement(
				ui.Insets{Right: 8},
				ui.AssistChipElement("Assist chip"),
			),
			ui.PaddingElement(
				ui.Insets{Right: 8},
				ui.AssistChipElement("Assist chip with icon", ui.ChipLeading(ui.Icon("info", ui.IconSize(18)))),
			),
			ui.PaddingElement(
				ui.Insets{Right: 8},
				ui.AssistChipElement("Elevated assist", ui.ChipElevated(true), ui.ChipLeading(ui.Icon("open_in_new", ui.IconSize(18)))),
			),
			ui.AssistChipElement("Soft-disabled assist", ui.ChipSoftDisabled(true)),
		),
		ui.VSpacerElement(10),
		ui.TextElement("Filter chips", ui.TextSize(13)),
		ui.VSpacerElement(6),
		ui.RowElement(
			ui.PaddingElement(
				ui.Insets{Right: 8},
				ui.FilterChipElement(
					"Filter chip",
					ui.ChipSelected(selected.Value()),
					ui.ChipOnClick(func(ctx *ui.Context) {
						selected.Set(!selected.Value())
					}),
				),
			),
			ui.PaddingElement(ui.Insets{Right: 8}, ui.FilterChipElement("Filter chip with icon", ui.ChipLeading(ui.Icon("label", ui.IconSize(18))))),
			ui.PaddingElement(ui.Insets{Right: 8}, ui.FilterChipElement("Removable filter chip", ui.ChipRemovable(true))),
			ui.FilterChipElement("Soft-disabled filter chip", ui.ChipSoftDisabled(true), ui.ChipRemovable(true)),
		),
		ui.VSpacerElement(10),
		ui.TextElement("Input chips", ui.TextSize(13)),
		ui.VSpacerElement(6),
		ui.RowElement(
			ui.PaddingElement(ui.Insets{Right: 8}, ui.InputChipElement("Input chip", ui.ChipOnRemove(func(ctx *ui.Context) {}))),
			ui.InputChipElement("Soft-disabled input chip", ui.ChipSoftDisabled(true)),
		),
		ui.VSpacerElement(10),
		ui.TextElement("Suggestion chips", ui.TextSize(13)),
		ui.VSpacerElement(6),
		ui.RowElement(
			ui.PaddingElement(ui.Insets{Right: 8}, ui.SuggestionChipElement("Suggestion chip")),
			ui.PaddingElement(ui.Insets{Right: 8}, ui.SuggestionChipElement("Suggestion chip with icon", ui.ChipLeading(ui.Icon("auto_awesome", ui.IconSize(18))))),
			ui.PaddingElement(ui.Insets{Right: 8}, ui.SuggestionChipElement("Elevated suggestion", ui.ChipElevated(true))),
			ui.SuggestionChipElement("Soft-disabled suggestion", ui.ChipSoftDisabled(true)),
		),
		ui.VSpacerElement(8),
		ui.FilterChipElementWithSlots(
			ui.TextElement("Styled slot filter", ui.TextSize(12), ui.TextColor(ui.NRGBA(30, 64, 175, 255))),
			ui.ChipLeading(ui.Icon("search", ui.IconSize(18))),
			ui.ChipSelected(true),
			ui.ChipBackground(ui.NRGBA(219, 234, 254, 255)),
			ui.ChipForeground(ui.NRGBA(30, 64, 175, 255)),
			ui.ChipDecoration(ui.Bg(ui.NRGBA(219, 234, 254, 255)).WithRad(999).WithBorder(ui.Border{Width: 1, Color: ui.NRGBA(147, 197, 253, 255)})),
		),
	)
}

func docsSearchBarDemo(value docsStringState, th *ui.Theme) ui.Element {
	return ui.FixedWidthElement(
		460,
		ui.ColumnElement(
			ui.SearchBarElement(
				value.Value(),
				ui.SearchBarPlaceholder("Search docs"),
				ui.SearchBarLeading(ui.Icon("search", ui.IconSize(18))),
				ui.SearchBarTrailing(ui.Icon("close", ui.IconSize(16))),
				ui.SearchBarInputOptions(ui.InputSingleLine(true), ui.InputMaxLen(40)),
				ui.SearchBarDecoration(ui.Bg(th.Colors.SurfaceContainerLow).WithRad(999).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant})),
				ui.SearchBarOnChange(func(ctx *ui.Context, next string) {
					value.Set(next)
				}),
			),
			ui.PaddingElement(
				ui.Insets{Top: 10},
				ui.TextElement("value = "+value.Value(), ui.TextSize(13), ui.TextColor(th.Colors.OnSurfaceVariant)),
			),
			ui.PaddingElement(
				ui.Insets{Top: 8},
				ui.SearchBarElement("disabled", ui.SearchBarDisabled(true)),
			),
		),
	)
}
