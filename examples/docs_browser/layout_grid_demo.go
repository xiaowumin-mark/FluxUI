package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type docsGridShowcaseItem struct {
	ID    string
	Label string
	Tone  string
}

var docsGridShowcaseItems = []docsGridShowcaseItem{
	{ID: "input", Label: "输入", Tone: "primary"},
	{ID: "layout", Label: "布局", Tone: "secondary"},
	{ID: "feedback", Label: "反馈", Tone: "tertiary"},
	{ID: "navigation", Label: "导航", Tone: "primary"},
	{ID: "theme", Label: "Theme", Tone: "secondary"},
	{ID: "system", Label: "System API", Tone: "tertiary"},
	{ID: "drag-drop", Label: "Drag/Drop", Tone: "primary"},
	{ID: "hooks", Label: "Hooks", Tone: "secondary"},
	{ID: "animation", Label: "动画", Tone: "tertiary"},
	{ID: "clipboard", Label: "Clipboard", Tone: "primary"},
	{ID: "tray", Label: "Tray", Tone: "secondary"},
	{ID: "window", Label: "Window", Tone: "tertiary"},
}

func docsGridDemo(th *ui.Theme) ui.Element {
	if th == nil {
		th = docsBrowserTheme(defaultDocsThemeSeed, false)
	}
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		reachEndCount := ui.UseState(ctx, 0)
		staticCells := make([]ui.Element, 0, 6)
		for i := 1; i <= 6; i++ {
			staticCells = append(staticCells, docsGridStaticCell(i, th))
		}

		return ui.ColumnElement(
			ui.TextElement("GridElement：固定子元素", ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(8),
			ui.GridElement(3, staticCells...),
			ui.VSpacerElement(14),
			ui.TextElement("GridViewElement：动态构建器 + 响应式最小宽度", ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
			ui.VSpacerElement(8),
			ui.FixedHeightElement(
				190,
				ui.GridViewElement(
					len(docsGridShowcaseItems),
					3,
					func(ctx *ui.Context, index int) ui.Element {
						item := docsGridShowcaseItems[index]
						return ui.Key(item.ID, docsGridDynamicCell(item, th))
					},
					ui.GridGap(8, 8),
					ui.GridPadding(ui.All(2)),
					ui.GridMinItemWidth(118),
					ui.GridDecoration(
						ui.Bg(th.Colors.SurfaceContainerLow).
							WithPad(ui.All(8)).
							WithRad(10).
							WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
					),
					ui.GridOnReachEnd(func(ctx *ui.Context) {
						reachEndCount.Set(reachEndCount.Value() + 1)
					}),
				),
			),
			ui.VSpacerElement(8),
			ui.TextElement(
				fmt.Sprintf("GridOnReachEnd 回调：%d", reachEndCount.Value()),
				ui.TextSize(12),
				ui.TextColor(th.Colors.OnSurfaceVariant),
			),
		)
	})
}

func docsGridStaticCell(index int, th *ui.Theme) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(th.Colors.PrimaryContainer).
			WithPad(ui.All(10)).
			WithRad(8).
			WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
		ui.CenterElement(
			ui.TextElement(
				fmt.Sprintf("单元格 %d", index),
				ui.TextSize(12),
				ui.TextColor(th.Colors.OnPrimaryContainer),
			),
		),
	)
}

func docsGridDynamicCell(item docsGridShowcaseItem, th *ui.Theme) ui.Element {
	bg := th.Colors.PrimaryContainer
	fg := th.Colors.OnPrimaryContainer
	switch item.Tone {
	case "secondary":
		bg = th.Colors.SecondaryContainer
		fg = th.Colors.OnSecondaryContainer
	case "tertiary":
		bg = th.Colors.TertiaryContainer
		fg = th.Colors.OnTertiaryContainer
	}
	return ui.ContainerDecorationElement(
		ui.Bg(bg).
			WithPad(ui.All(10)).
			WithRad(8),
		ui.CenterElement(
			ui.TextElement(item.Label, ui.TextSize(12), ui.TextColor(fg)),
		),
	)
}
