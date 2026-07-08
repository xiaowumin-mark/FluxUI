package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsScrollViewDemo() ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		lineCount := ui.UseState(ctx, 18)
		offset := ui.UseState(ctx, "x=0.00 y=0.00")
		scrollRef := ui.UseRef(ctx, ui.NewScrollRef())
		if scrollRef.Current == nil {
			scrollRef.Current = ui.NewScrollRef()
		}

		lines := make([]ui.Element, 0, lineCount.Value())
		for i := 1; i <= lineCount.Value(); i++ {
			lines = append(lines,
				ui.PaddingElement(
					ui.Insets{Bottom: 6},
					ui.ContainerDecorationElement(
						ui.Bg(themedRowColor(th, i)).
							WithPad(ui.All(8)).
							WithRad(7),
						ui.TextElement(fmt.Sprintf("日志行 %02d", i), ui.TextSize(12), ui.TextColor(th.Colors.OnSurface)),
					),
				),
			)
		}

		return ui.ColumnElement(
			ui.RowElement(
				ui.TextElement("ScrollViewElement", ui.TextSize(14), ui.TextColor(th.Colors.OnSurface)),
				ui.ExpandedElement(ui.SpacerElement(0, 0)),
				ui.TextElement(offset.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				ui.ExpandedElement(
					ui.FixedHeightElement(
						190,
						ui.ScrollViewElement(
							ui.ColumnElement(lines...),
							ui.ScrollVertical(true),
							ui.ScrollAutoToEndKey(lineCount.Value()),
							ui.ScrollAttachRef(scrollRef.Current),
							ui.ScrollOnChange(func(ctx *ui.Context, x, y float32) {
								offset.Set(fmt.Sprintf("x=%.2f y=%.2f", x, y))
							}),
						),
					),
				),
				ui.HSpacerElement(12),
				ui.FixedWidthElement(
					260,
					ui.ColumnElement(
						docsScrollActionButton("添加行", func(ctx *ui.Context) {
							lineCount.Set(lineCount.Value() + 1)
						}),
						ui.VSpacerElement(8),
						docsScrollActionButton("滚动到顶部", func(ctx *ui.Context) {
							scrollRef.Current.ScrollToTop()
						}),
						ui.VSpacerElement(8),
						docsScrollActionButton("滚动到底部", func(ctx *ui.Context) {
							scrollRef.Current.ScrollToBottom()
						}),
					),
				),
			),
			ui.VSpacerElement(12),
			ui.TextElement("横向内容", ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(6),
			ui.FixedHeightElement(
				78,
				ui.ScrollViewElement(
					docsHorizontalScrollItems(th),
					ui.ScrollVertical(false),
					ui.ScrollHorizontal(true),
					ui.ScrollBarVisible(true),
				),
			),
		)
	})
}

func docsScrollActionButton(label string, onClick func(*ui.Context)) ui.Element {
	return ui.FillWidthElement(
		ui.OutlinedButtonElement(
			ui.TextElement(label, ui.TextSize(12)),
			ui.ButtonPadding(ui.Symmetric(5, 10)),
			ui.OnClick(onClick),
		),
	)
}

func docsHorizontalScrollItems(th *ui.Theme) ui.Element {
	items := make([]ui.Element, 0, 12)
	for i := 1; i <= 12; i++ {
		items = append(items,
			ui.PaddingElement(
				ui.Insets{Right: 8},
				ui.ContainerDecorationElement(
					ui.Bg(th.Colors.SecondaryContainer).
						WithPad(ui.Symmetric(10, 16)).
						WithRad(10),
					ui.TextElement(fmt.Sprintf("列 %02d", i), ui.TextSize(12), ui.TextColor(th.Colors.OnSecondaryContainer)),
				),
			),
		)
	}
	return ui.RowElement(items...)
}

func docsListViewDemo(reachEndCount docsIntState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		return ui.ColumnElement(
			ui.TextElement("ListViewElement：虚拟化纵向列表", ui.TextSize(14), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(8),
			ui.FixedHeightElement(
				190,
				ui.ListViewElement(
					80,
					func(ctx *ui.Context, index int) ui.Element {
						return ui.Key(
							fmt.Sprintf("row-%02d", index),
							ui.ContainerDecorationElement(
								ui.Bg(themedRowColor(th, index)).
									WithPad(ui.Symmetric(8, 10)).
									WithRad(7),
								ui.RowElement(
									ui.TextElement(fmt.Sprintf("列表项 #%02d", index), ui.TextSize(12), ui.TextColor(th.Colors.OnSurface)),
									ui.ExpandedElement(ui.SpacerElement(0, 0)),
									ui.TextElement("稳定键", ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant)),
								),
							),
						)
					},
					ui.ListVirtualized(true),
					ui.ListItemSpacing(6),
					ui.ListPadding(ui.All(8)),
					ui.ListDecoration(
						ui.Bg(th.Colors.SurfaceContainerLow).
							WithPad(ui.All(4)).
							WithRad(10).
							WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
					),
					ui.ListOnReachEnd(func(ctx *ui.Context) {
						reachEndCount.Set(reachEndCount.Value() + 1)
					}),
				),
			),
			ui.VSpacerElement(8),
			ui.TextElement(fmt.Sprintf("到达末尾回调：%d", reachEndCount.Value()), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
			ui.VSpacerElement(12),
			ui.TextElement("ListAxis(Horizontal)", ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(6),
			ui.FixedHeightElement(
				70,
				ui.ListViewElement(
					10,
					func(ctx *ui.Context, index int) ui.Element {
						return ui.ContainerDecorationElement(
							ui.Bg(th.Colors.TertiaryContainer).WithPad(ui.Symmetric(8, 12)).WithRad(10),
							ui.TextElement(fmt.Sprintf("标签 %02d", index+1), ui.TextSize(12), ui.TextColor(th.Colors.OnTertiaryContainer)),
						)
					},
					ui.ListAxis(ui.Horizontal),
					ui.ListItemSpacing(8),
					ui.ListPadding(ui.All(4)),
				),
			),
		)
	})
}
