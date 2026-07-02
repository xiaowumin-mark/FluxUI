package main

import (
	"fmt"
	"time"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsMenuDemo(menuOpen docsBoolState, menuValue docsStringState, th *ui.Theme) ui.Element {
	items := []ui.MenuItem{
		{Key: "copy", Label: "Copy", Leading: ui.Icon("content_copy")},
		{Key: "share", Label: "Share", Leading: ui.Icon("share")},
		{Key: "archive", Label: "Archive", Leading: ui.Icon("archive"), Children: []ui.MenuItem{
			{Key: "archive-today", Label: "Today"},
			{Key: "archive-week", Label: "This week"},
			{Key: "archive-custom", Label: "Custom date", Leading: ui.Icon("calendar_month")},
		}},
		{Divider: true},
		{Key: "delete", Label: "Delete", Disabled: true},
	}
	return ui.FixedWidthElement(
		520,
		ui.RowElement(
			ui.ExpandedElement(
				ui.DropdownMenuElement(
					menuOpen.Value(),
					ui.ContainerDecorationElement(
						ui.Bg(th.Colors.Surface).WithPad(ui.Symmetric(10, 16)).WithRad(th.Shapes.ExtraSmall).WithBorder(ui.Border{Width: 1, Color: th.Colors.Outline}),
						ui.TextElement("Open menu", ui.TextColor(th.Colors.OnSurface)),
					),
					items,
					ui.DropdownMenuSelectedKey(menuValue.Value()),
					ui.DropdownMenuWidth(220),
					ui.DropdownMenuMaxHeight(240),
					ui.DropdownMenuXOffset(4),
					ui.DropdownMenuYOffset(2),
					ui.DropdownMenuAnchorCorner(ui.MenuCornerEndStart),
					ui.DropdownMenuMenuCorner(ui.MenuCornerStartStart),
					ui.DropdownMenuDefaultFocusOf(ui.MenuDefaultFocusFirstItem),
					ui.DropdownMenuTypeaheadDelay(500*time.Millisecond),
					ui.DropdownMenuNoNavigationWrap(false),
					ui.DropdownMenuDecoration(ui.Bg(th.Colors.SurfaceContainer).WithRad(10).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant})),
					ui.DropdownMenuOnOpenChange(func(ctx *ui.Context, open bool) {
						menuOpen.Set(open)
					}),
					ui.DropdownMenuOnSelect(func(ctx *ui.Context, key string) {
						menuValue.Set(key)
						menuOpen.Set(false)
					}),
				),
			),
			ui.HSpacerElement(12),
			ui.ExpandedElement(
				ui.MenuElement(
					items[:3],
					ui.MenuSelectedKey(menuValue.Value()),
					ui.MenuWidth(190),
					ui.MenuMaxHeight(160),
					ui.MenuQuick(true),
					ui.MenuDecoration(ui.Bg(th.Colors.SurfaceContainerLow).WithRad(10)),
					ui.MenuOnSelect(func(ctx *ui.Context, key string) {
						menuValue.Set(key)
					}),
				),
			),
		),
	)
}

func docsTabsDemo(value docsStringState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		ref := ui.UseRef(ctx, ui.NewTabsRef())
		if ref.Current == nil {
			ref.Current = ui.NewTabsRef()
		}
		scrollingActive := ui.UseState(ctx, "scroll-keyboard-0")
		items := []ui.TabItem{
			{Key: "overview", Label: "Keyboard", Icon: ui.Icon("keyboard")},
			{Key: "api", Label: "Guitar", Icon: ui.Icon("tune")},
			{Key: "example", Label: "Drums", Icon: ui.Icon("graphic_eq")},
			{Key: "notes", Label: "Bass", Icon: ui.Icon("speaker")},
			{Key: "saxophone", Label: "Saxophone", Icon: ui.Icon("nightlife"), Disabled: true},
		}
		scrollingBase := []ui.TabItem{
			{Key: "keyboard", Label: "Keyboard", Icon: ui.Icon("keyboard")},
			{Key: "guitar", Label: "Guitar", Icon: ui.Icon("tune")},
			{Key: "drums", Label: "Drums", Icon: ui.Icon("graphic_eq")},
			{Key: "bass", Label: "Bass", Icon: ui.Icon("speaker")},
			{Key: "saxophone", Label: "Saxophone", Icon: ui.Icon("nightlife")},
		}
		scrollingItems := make([]ui.TabItem, 0, len(scrollingBase)*3)
		for i := 0; i < 3; i++ {
			for _, item := range scrollingBase {
				scrollingItems = append(scrollingItems, ui.TabItem{
					Key:   fmt.Sprintf("scroll-%s-%d", item.Key, i),
					Label: item.Label,
					Icon:  item.Icon,
				})
			}
		}
		return ui.FixedWidthElement(
			520,
			ui.ColumnElement(
				ui.TextElement("Primary full width", ui.TextSize(13)),
				ui.TabsElement(
					value.Value(),
					items,
					ui.TabsFullWidth(true),
					ui.TabsAttachRef(ref.Current),
					ui.TabsOnChange(func(ctx *ui.Context, key string) {
						value.Set(key)
					}),
				),
				ui.VSpacerElement(10),
				ui.TextElement("Scrolling tabs", ui.TextSize(13)),
				ui.TabsElement(
					scrollingActive.Value(),
					scrollingItems,
					ui.TabsScrollable(true),
					ui.TabsOnChange(func(ctx *ui.Context, key string) {
						scrollingActive.Set(key)
					}),
				),
				ui.VSpacerElement(10),
				ui.TextElement("Secondary full width", ui.TextSize(13)),
				ui.TabsElement(
					value.Value(),
					[]ui.TabItem{
						{Key: "overview", Label: "Travel", Icon: ui.Icon("flight")},
						{Key: "api", Label: "Hotel", Icon: ui.Icon("hotel")},
						{Key: "example", Label: "Activities", Icon: ui.Icon("hiking")},
						{Key: "notes", Label: "Food", Icon: ui.Icon("restaurant")},
					},
					ui.TabsSecondary(true),
					ui.TabsInlineIcon(true),
					ui.TabsFullWidth(true),
					ui.TabsIndicatorColor(ui.NRGBA(37, 99, 235, 255)),
					ui.TabsTextColor(ui.NRGBA(71, 85, 105, 255)),
					ui.TabsActiveTextColor(ui.NRGBA(30, 64, 175, 255)),
					ui.TabsOnChange(func(ctx *ui.Context, key string) {
						value.Set(key)
					}),
				),
				ui.PaddingElement(
					ui.Insets{Top: 8},
					ui.RowElement(
						docsDemoControlButton("Set API", func(ctx *ui.Context) {
							value.Set("api")
							ref.Current.SetActive("api")
						}),
						ui.HSpacerElement(8),
						ui.TextElement("Current tab: "+value.Value(), ui.TextSize(13)),
					),
				),
			),
		)
	})
}

func docsDialogDemo(open docsBoolState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		ref := ui.UseRef(ctx, ui.NewDialogRef())
		if ref.Current == nil {
			ref.Current = ui.NewDialogRef()
		}
		return ui.StackElement(
			ui.FillWidthElement(
				ui.ColumnElement(
					ui.RowElement(
						docsDemoControlButton("Open dialog", func(ctx *ui.Context) {
							open.Set(true)
							ref.Current.Open()
						}),
						ui.HSpacerElement(8),
						docsDemoControlButton("Toggle ref", func(ctx *ui.Context) {
							open.Set(!open.Value())
							ref.Current.Toggle()
						}),
					),
					ui.PaddingElement(
						ui.Insets{Top: 8},
						ui.TextElement("Dialog supports mask close, custom action text, ref commands, decoration, and mask color.", ui.TextSize(13)),
					),
				),
			),
			ui.DialogElement(
				open.Value(),
				ui.TextElement("This dialog uses Material Web style slots: icon, headline, content, actions, scrim, and modal lifecycle callbacks."),
				ui.DialogIconElement(ui.IconElement("info")),
				ui.DialogHeadlineElement(ui.TextElement("Dialog example")),
				ui.DialogWidth(420),
				ui.DialogRadius(28),
				ui.DialogMaskClosable(true),
				ui.DialogTypeOf(ui.DialogTypeAlert),
				ui.DialogGlobalOverlay(true),
				ui.DialogMaskColor(ui.NRGBA(15, 23, 42, 82)),
				ui.DialogActions(
					ui.TextButton(ui.Text("Dismiss"), ui.OnClick(func(ctx *ui.Context) {
						open.Set(false)
						ref.Current.Close()
					})),
					ui.TextButton(ui.Text("Apply"), ui.OnClick(func(ctx *ui.Context) {
						open.Set(false)
						ref.Current.Close()
					})),
				),
				ui.DialogAttachRef(ref.Current),
				ui.DialogOnOpenChange(func(ctx *ui.Context, next bool) {
					open.Set(next)
				}),
				ui.DialogOnCancelOnly(func(ctx *ui.Context) {
					open.Set(false)
					ref.Current.Close()
				}),
			),
		)
	})
}

func docsPopupDemo(open docsBoolState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		ref := ui.UseRef(ctx, ui.NewPopupRef())
		if ref.Current == nil {
			ref.Current = ui.NewPopupRef()
		}
		return ui.StackElement(
			ui.FillWidthElement(
				ui.ColumnElement(
					ui.RowElement(
						docsDemoControlButton("Open Popup", func(ctx *ui.Context) {
							open.Set(true)
							ref.Current.Open()
						}),
						ui.HSpacerElement(8),
						docsDemoControlButton("Toggle ref", func(ctx *ui.Context) {
							open.Set(!open.Value())
							ref.Current.Toggle()
						}),
					),
					ui.PaddingElement(
						ui.Insets{Top: 8},
						ui.TextElement("Popup content is fully custom and can use ref commands.", ui.TextSize(13)),
					),
				),
			),
			ui.PopupElement(
				open.Value(),
				ui.ColumnElement(
					ui.TextElement("Custom popup content", ui.TextSize(16)),
					ui.PaddingElement(
						ui.Insets{Top: 8},
						ui.TextElement("Any Element tree can be placed here.", ui.TextSize(13)),
					),
					ui.PaddingElement(
						ui.Insets{Top: 12},
						docsDemoControlButton("Close", func(ctx *ui.Context) {
							open.Set(false)
							ref.Current.Close()
						}),
					),
				),
				ui.PopupWidth(320),
				ui.PopupMaskClosable(true),
				ui.PopupMaskColor(ui.NRGBA(0, 0, 0, 82)),
				ui.PopupAttachRef(ref.Current),
				ui.PopupOnOpenChange(func(ctx *ui.Context, next bool) {
					open.Set(next)
				}),
			),
		)
	})
}

func docsToastDemo(message docsStringState) ui.Element {
	var layers []ui.Element
	layers = append(layers,
		ui.FillWidthElement(
			ui.RowElement(
				docsDemoControlButton("Show success Toast", func(ctx *ui.Context) {
					message.Set("Saved changes")
				}),
				ui.HSpacerElement(8),
				ui.TextElement("Toast uses duration, position, action, decoration, and close callback.", ui.TextSize(12)),
			),
		),
	)
	if message.Value() != "" {
		layers = append(layers,
			ui.ToastElement(
				message.Value(),
				ui.ToastTypeOf(ui.ToastSuccess),
				ui.ToastPositionOf(ui.ToastBottom),
				ui.ToastDuration(1600*time.Millisecond),
				ui.ToastTextColor(ui.NRGBA(255, 255, 255, 255)),
				ui.ToastDecoration(ui.Bg(ui.NRGBA(22, 101, 52, 245)).WithRad(10)),
				ui.ToastAction("Dismiss", func(ctx *ui.Context) {
					message.Set("")
				}),
				ui.ToastOnClose(func(ctx *ui.Context) {
					message.Set("")
				}),
			),
		)
	}
	return ui.StackElement(layers...)
}

func docsSnackbarDemo(serial docsIntState, message docsStringState, actionCount docsIntState, th *ui.Theme) ui.Element {
	layers := []ui.Element{
		ui.FixedHeightElement(
			160,
			ui.ContainerDecorationElement(
				ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.All(16)).WithRad(th.Shapes.Medium),
				ui.ColumnElement(
					ui.FilledButtonElement(
						ui.TextElement("Show snackbar"),
						ui.OnClick(func(ctx *ui.Context) {
							serial.Set(serial.Value() + 1)
							message.Set("Draft archived")
						}),
					),
					ui.PaddingElement(
						ui.Insets{Top: 10},
						ui.TextElement(fmt.Sprintf("Action clicks: %d", actionCount.Value()), ui.TextSize(13), ui.TextColor(th.Colors.OnSurfaceVariant)),
					),
				),
			),
		),
	}
	if message.Value() != "" {
		layers = append(layers,
			ui.Key(
				fmt.Sprintf("snackbar-%d", serial.Value()),
				ui.SnackbarElement(
					message.Value(),
					ui.SnackbarAction("Undo", func(ctx *ui.Context) {
						actionCount.Set(actionCount.Value() + 1)
						message.Set("")
					}),
					ui.ToastDuration(0),
				),
			),
		)
	}
	return ui.StackElement(layers...)
}

func docsTooltipDemo() ui.Element {
	return ui.RowElement(
		ui.TooltipElement(
			"Tooltip text",
			ui.FilledTonalButtonElement(ui.TextElement("Hover me")),
			ui.TooltipOffset(8),
			ui.TooltipTextColor(ui.NRGBA(255, 255, 255, 255)),
			ui.TooltipDecoration(ui.Bg(ui.NRGBA(15, 23, 42, 245)).WithPad(ui.Symmetric(6, 10)).WithRad(8)),
		),
		ui.HSpacerElement(12),
		ui.TooltipElement(
			"Disabled tooltip",
			ui.OutlinedButtonElement(ui.TextElement("Disabled")),
			ui.TooltipDisabled(true),
		),
	)
}
