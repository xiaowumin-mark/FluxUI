package main

import (
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsProgressBarDemo(value docsFloat32State, th *ui.Theme) ui.Element {
	return ui.ColumnElement(
		ui.SliderElement(
			value.Value(),
			ui.SliderMin(0),
			ui.SliderMax(100),
			ui.SliderOnChange(func(ctx *ui.Context, next float32) {
				value.Set(next)
			}),
		),
		ui.PaddingElement(
			ui.Insets{Top: 8},
			ui.ProgressBarElement(
				value.Value(),
				ui.ProgressMin(0),
				ui.ProgressMax(100),
				ui.ProgressTrackColor(ui.NRGBA(226, 232, 240, 255)),
				ui.ProgressFillColor(th.Primary),
				ui.ProgressDecoration(ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.All(8)).WithRad(8)),
			),
		),
		ui.PaddingElement(
			ui.Insets{Top: 12},
			ui.ProgressBarElement(0, ui.ProgressIndeterminate(true), ui.ProgressTrackColor(ui.NRGBA(226, 232, 240, 255)), ui.ProgressFillColor(th.Colors.Tertiary)),
		),
	)
}

func docsCircularProgressDemo(value docsFloat32State, th *ui.Theme) ui.Element {
	return ui.ColumnElement(
		ui.SliderElement(
			value.Value(),
			ui.SliderMin(0),
			ui.SliderMax(100),
			ui.SliderOnChange(func(ctx *ui.Context, next float32) {
				value.Set(next)
			}),
		),
		ui.PaddingElement(
			ui.Insets{Top: 12},
			ui.CircularProgressElement(
				value.Value(),
				ui.ProgressMin(0),
				ui.ProgressMax(100),
				ui.ProgressSize(80),
				ui.ProgressThickness(8),
				ui.ProgressFillColor(th.Primary),
				ui.ProgressTrackColor(ui.NRGBA(226, 232, 240, 255)),
				ui.ProgressLabelVisible(true),
			),
		),
	)
}

func docsProgressIndicatorsDemo(value docsFloat32State, th *ui.Theme) ui.Element {
	materialValue := value.Value() / 100
	return ui.FixedWidthElement(
		380,
		ui.ColumnElement(
			ui.SliderElement(
				value.Value(),
				ui.SliderMin(0),
				ui.SliderMax(100),
				ui.SliderOnChange(func(ctx *ui.Context, next float32) {
					value.Set(next)
				}),
			),
			ui.PaddingElement(
				ui.Insets{Top: 12},
				ui.LinearProgressIndicatorElement(
					materialValue,
					ui.ProgressBuffer(0.8),
					ui.ProgressTrackHeight(4),
					ui.ProgressIndicatorHeight(4),
					ui.ProgressFillColor(th.Primary),
					ui.ProgressDecoration(ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.All(6)).WithRad(8)),
				),
			),
			ui.PaddingElement(
				ui.Insets{Top: 12},
				ui.LinearProgressIndicatorElement(0, ui.ProgressIndeterminate(true), ui.ProgressFourColor(true)),
			),
			ui.PaddingElement(
				ui.Insets{Top: 16},
				ui.RowElement(
					ui.CircularProgressIndicatorElement(
						materialValue,
						ui.ProgressSize(72),
						ui.ProgressFillColor(th.Primary),
						ui.ProgressLabelVisible(true),
					),
					ui.HSpacerElement(18),
					ui.LoadingIndicatorElement(ui.ProgressFourColor(true), ui.ProgressSize(48)),
				),
			),
			ui.PaddingElement(
				ui.Insets{Top: 16},
				ui.RowElement(
					ui.FilledTonalButtonElement(ui.TextElement("加载中"), ui.ButtonLoading(true)),
					ui.HSpacerElement(12),
					ui.IconButtonElement(ui.IconElement("sync"), ui.IconButtonLoading(true), ui.IconButtonForeground(th.Primary)),
				),
			),
		),
	)
}
