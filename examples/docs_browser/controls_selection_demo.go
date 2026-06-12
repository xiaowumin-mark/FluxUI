package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsCheckboxDemo(checked docsBoolState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		ref := ui.UseRef(ctx, ui.NewCheckboxRef())
		if ref.Current == nil {
			ref.Current = ui.NewCheckboxRef()
		}
		return ui.FixedWidthElement(
			430,
			ui.ColumnElement(
				ui.CheckboxElement(
					"Enable feature",
					checked.Value(),
					ui.CheckboxSize(22),
					ui.CheckboxColor(ui.NRGBA(37, 99, 235, 255)),
					ui.CheckboxDecoration(ui.Bg(ui.NRGBA(248, 250, 252, 255)).WithRad(8)),
					ui.CheckboxAttachRef(ref.Current),
					ui.CheckboxOnChange(func(ctx *ui.Context, value bool) {
						checked.Set(value)
					}),
				),
				ui.VSpacerElement(8),
				ui.RowElement(
					docsDemoControlButton("Toggle ref", func(ctx *ui.Context) {
						checked.Set(!checked.Value())
						ref.Current.Toggle()
					}),
					ui.HSpacerElement(8),
					docsDemoControlButton("Set checked", func(ctx *ui.Context) {
						checked.Set(true)
						ref.Current.SetChecked(true)
					}),
				),
				ui.VSpacerElement(8),
				ui.CheckboxElement("Disabled option", true, ui.CheckboxDisabled(true)),
			),
		)
	})
}

func docsSwitchDemo(checked docsBoolState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		ref := ui.UseRef(ctx, ui.NewSwitchRef())
		if ref.Current == nil {
			ref.Current = ui.NewSwitchRef()
		}
		return ui.FixedWidthElement(
			430,
			ui.ColumnElement(
				ui.RowElement(
					ui.SwitchElement(
						checked.Value(),
						ui.SwitchWidth(56),
						ui.SwitchHeight(32),
						ui.SwitchColor(ui.NRGBA(37, 99, 235, 255)),
						ui.SwitchTrackColor(ui.NRGBA(191, 219, 254, 255)),
						ui.SwitchThumbColor(ui.NRGBA(255, 255, 255, 255)),
						ui.SwitchDecoration(ui.Bg(ui.NRGBA(248, 250, 252, 255)).WithPad(ui.All(3)).WithRad(999)),
						ui.SwitchAttachRef(ref.Current),
						ui.SwitchOnChange(func(ctx *ui.Context, value bool) {
							checked.Set(value)
						}),
					),
					ui.PaddingElement(
						ui.Insets{Left: 10, Top: 5},
						ui.TextElement(fmt.Sprintf("State: %v", checked.Value()), ui.TextSize(13)),
					),
				),
				ui.VSpacerElement(8),
				ui.RowElement(
					docsDemoControlButton("Toggle ref", func(ctx *ui.Context) {
						checked.Set(!checked.Value())
						ref.Current.Toggle()
					}),
					ui.HSpacerElement(8),
					docsDemoControlButton("Off", func(ctx *ui.Context) {
						checked.Set(false)
						ref.Current.SetChecked(false)
					}),
				),
				ui.VSpacerElement(8),
				ui.SwitchElement(false, ui.SwitchDisabled(true)),
			),
		)
	})
}

func docsSliderDemo(value docsFloat32State) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		ref := ui.UseRef(ctx, ui.NewSliderRef())
		if ref.Current == nil {
			ref.Current = ui.NewSliderRef()
		}
		return ui.FixedWidthElement(
			480,
			ui.ColumnElement(
				ui.SliderElement(
					value.Value(),
					ui.SliderMin(0),
					ui.SliderMax(100),
					ui.SliderStep(5),
					ui.SliderWidth(360),
					ui.SliderTrackColor(ui.NRGBA(226, 232, 240, 255)),
					ui.SliderProgressColor(ui.NRGBA(37, 99, 235, 255)),
					ui.SliderThumbColor(ui.NRGBA(29, 78, 216, 255)),
					ui.SliderDecoration(ui.Bg(ui.NRGBA(248, 250, 252, 255)).WithPad(ui.Symmetric(8, 10)).WithRad(12)),
					ui.SliderAttachRef(ref.Current),
					ui.SliderOnChange(func(ctx *ui.Context, next float32) {
						value.Set(next)
					}),
				),
				ui.VSpacerElement(8),
				ui.RowElement(
					docsDemoControlButton("Step +10", func(ctx *ui.Context) {
						next := minFloat32(100, value.Value()+10)
						value.Set(next)
						ref.Current.StepBy(10)
					}),
					ui.HSpacerElement(8),
					docsDemoControlButton("Set 50", func(ctx *ui.Context) {
						value.Set(50)
						ref.Current.SetValue(50)
					}),
					ui.HSpacerElement(8),
					ui.TextElement(fmt.Sprintf("value = %.1f", value.Value()), ui.TextSize(13)),
				),
				ui.VSpacerElement(8),
				ui.SliderElement(40, ui.SliderDisabled(true), ui.SliderWidth(240)),
			),
		)
	})
}

func docsRadioGroupDemo(value docsStringState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		ref := ui.UseRef(ctx, ui.NewRadioGroupRef())
		if ref.Current == nil {
			ref.Current = ui.NewRadioGroupRef()
		}
		items := []ui.RadioItem{
			{Label: "Layout", Value: "layout"},
			{Label: "Input", Value: "input"},
			{Label: "Feedback", Value: "feedback"},
		}
		return ui.ColumnElement(
			ui.RadioGroupElement(
				value.Value(),
				items,
				ui.RadioGroupDirection(ui.Horizontal),
				ui.RadioGroupSize(20),
				ui.RadioGroupColor(ui.NRGBA(37, 99, 235, 255)),
				ui.RadioGroupDecoration(ui.Bg(ui.NRGBA(248, 250, 252, 255)).WithPad(ui.All(8)).WithRad(10)),
				ui.RadioGroupAttachRef(ref.Current),
				ui.RadioGroupOnChange(func(ctx *ui.Context, next string) {
					value.Set(next)
				}),
			),
			ui.VSpacerElement(8),
			ui.RowElement(
				docsDemoControlButton("Set input", func(ctx *ui.Context) {
					value.Set("input")
					ref.Current.SetValue("input")
				}),
				ui.HSpacerElement(8),
				ui.RadioGroupElement("layout", items[:2], ui.RadioGroupDisabled(true)),
			),
		)
	})
}

func docsSelectDemo(value docsStringState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		ref := ui.UseRef(ctx, ui.NewSelectRef[string]())
		openState := ui.UseState(ctx, "closed")
		if ref.Current == nil {
			ref.Current = ui.NewSelectRef[string]()
		}
		options := []ui.SelectOptionItem[string]{
			{Label: "Low priority", Value: "low"},
			{Label: "Medium priority", Value: "medium"},
			{Label: "High priority", Value: "high"},
		}
		return ui.FixedWidthElement(
			480,
			ui.ColumnElement(
				ui.SelectElement(
					value.Value(),
					options,
					ui.SelectPlaceholder[string]("Choose priority"),
					ui.SelectSearchable[string](true),
					ui.SelectMaxHeight[string](180),
					ui.SelectDecoration[string](ui.Bg(ui.NRGBA(248, 250, 252, 255)).WithRad(10)),
					ui.SelectAttachRef[string](ref.Current),
					ui.SelectOnOpenChange[string](func(ctx *ui.Context, open bool) {
						if open {
							openState.Set("open")
						} else {
							openState.Set("closed")
						}
					}),
					ui.SelectOnChange[string](func(ctx *ui.Context, next string) {
						value.Set(next)
					}),
				),
				ui.VSpacerElement(8),
				ui.RowElement(
					docsDemoControlButton("Open", func(ctx *ui.Context) {
						openState.Set("open")
						ref.Current.Open()
					}),
					ui.HSpacerElement(8),
					docsDemoControlButton("Toggle", func(ctx *ui.Context) {
						ref.Current.Toggle()
					}),
					ui.HSpacerElement(8),
					docsDemoControlButton("Set high", func(ctx *ui.Context) {
						value.Set("high")
						ref.Current.SetValue("high")
					}),
					ui.HSpacerElement(8),
					ui.TextElement(openState.Value(), ui.TextSize(12)),
				),
				ui.VSpacerElement(8),
				ui.SelectElement("low", options, ui.SelectDisabled[string](true)),
			),
		)
	})
}

func minFloat32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
