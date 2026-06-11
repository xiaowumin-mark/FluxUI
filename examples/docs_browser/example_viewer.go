package main

import (
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsExampleSectionHeader(title string, exampleID string, popupOpen docsBoolState, th *ui.Theme) ui.Element {
	return ui.RowElement(
		ui.TextElement("组件示例", ui.TextSize(17), ui.TextColor(th.Colors.OnSurface)),
		ui.HSpacerElement(10),
		ui.ContainerDecorationElement(
			ui.Bg(th.Colors.SecondaryContainer).WithPad(ui.Symmetric(3, 8)).WithRad(6),
			ui.TextElement(exampleID, ui.TextSize(11), ui.TextColor(th.Colors.OnSecondaryContainer)),
		),
		ui.ExpandedElement(ui.SpacerElement(0, 0)),
		ui.OutlinedButtonElement(
			ui.TextElement("弹窗查看", ui.TextSize(12)),
			ui.ButtonPadding(ui.Symmetric(5, 10)),
			ui.OnClick(func(ctx *ui.Context) {
				popupOpen.Set(true)
			}),
		),
	)
}

func docsInlineExampleFrame(demoHeight float32, demoViewport ui.Element, th *ui.Theme) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.SurfaceContainerLow).
			WithPad(ui.All(12)).
			WithRad(10).
			WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
		ui.FixedHeightElement(
			demoHeight,
			demoViewport,
		),
	)
}

func docsExamplePopup(open bool, title string, exampleID string, demo ui.Element, popupOpen docsBoolState, th *ui.Theme) ui.Element {
	return ui.PopupElement(
		open,
		ui.ContainerDecorationElement(
			ui.Bg(th.Colors.Surface).WithPad(ui.All(16)).WithRad(16),
			ui.ColumnElement(
				ui.RowElement(
					ui.ColumnElement(
						ui.TextElement(title, ui.TextSize(20), ui.TextColor(th.Colors.OnSurface)),
						ui.VSpacerElement(4),
						ui.TextElement("Example ID: "+exampleID, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
					),
					ui.ExpandedElement(ui.SpacerElement(0, 0)),
					ui.TextButtonElement(
						ui.TextElement("关闭", ui.TextSize(12)),
						ui.ButtonPadding(ui.Symmetric(5, 10)),
						ui.OnClick(func(ctx *ui.Context) {
							popupOpen.Set(false)
						}),
					),
				),
				ui.VSpacerElement(12),
				ui.ContainerDecorationElement(
					ui.Bg(th.Colors.SurfaceContainerLow).
						WithPad(ui.All(12)).
						WithRad(12).
						WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
					ui.FixedHeightElement(
						520,
						ui.ScrollViewElement(
							ui.FillElement(demo),
							ui.ScrollVertical(true),
						),
					),
				),
			),
		),
		ui.PopupWidth(900),
		ui.PopupPadding(ui.All(0)),
		ui.PopupRadius(16),
		ui.PopupBackground(th.Colors.Surface),
		ui.PopupMaskAlpha(112),
		ui.PopupMaskClosable(true),
		ui.PopupOnOpenChange(func(ctx *ui.Context, open bool) {
			popupOpen.Set(open)
		}),
	)
}
