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
		th := ui.UseTheme(ctx)
		ref := ui.UseRef(ctx, ui.NewSliderRef())
		if ref.Current == nil {
			ref.Current = ui.NewSliderRef()
		}
		rangeStart := ui.UseState(ctx, float32(18))
		rangeEnd := ui.UseState(ctx, float32(48))
		tickValue := ui.UseState(ctx, float32(20))
		tickRangeStart := ui.UseState(ctx, float32(20))
		tickRangeEnd := ui.UseState(ctx, float32(40))
		customStart := ui.UseState(ctx, float32(15))
		customEnd := ui.UseState(ctx, float32(92))
		return ui.FixedWidthElement(
			480,
			ui.ColumnElement(
				ui.TextElement("Single point sliders", ui.TextType(th.Types.TitleMedium), ui.TextColor(th.Colors.OnSurface)),
				ui.VSpacerElement(8),
				ui.TextElement("Continuous", ui.TextType(th.Types.BodyMedium)),
				ui.SliderElement(
					value.Value(),
					ui.SliderMin(0),
					ui.SliderMax(100),
					ui.SliderStep(0),
					ui.SliderAttachRef(ref.Current),
					ui.SliderOnChange(func(ctx *ui.Context, next float32) {
						value.Set(next)
					}),
				),
				ui.VSpacerElement(10),
				ui.TextElement("Labeled", ui.TextType(th.Types.BodyMedium)),
				ui.SliderElement(value.Value(), ui.SliderLabeled(true), ui.SliderOnChange(func(ctx *ui.Context, next float32) { value.Set(next) })),
				ui.VSpacerElement(10),
				ui.TextElement("Tick marks", ui.TextType(th.Types.BodyMedium)),
				ui.SliderElement(tickValue.Value(), ui.SliderStep(10), ui.SliderTicks(true), ui.SliderOnChange(func(ctx *ui.Context, next float32) { tickValue.Set(next) })),
				ui.VSpacerElement(18),
				ui.TextElement("Range sliders", ui.TextType(th.Types.TitleMedium), ui.TextColor(th.Colors.OnSurface)),
				ui.VSpacerElement(8),
				ui.TextElement("Range", ui.TextType(th.Types.BodyMedium)),
				ui.RangeSliderElement(rangeStart.Value(), rangeEnd.Value(), ui.SliderOnRangeChange(func(ctx *ui.Context, start, end float32) {
					rangeStart.Set(start)
					rangeEnd.Set(end)
				})),
				ui.VSpacerElement(10),
				ui.TextElement("Labeled", ui.TextType(th.Types.BodyMedium)),
				ui.RangeSliderElement(rangeStart.Value(), rangeEnd.Value(), ui.SliderLabeled(true), ui.SliderOnRangeChange(func(ctx *ui.Context, start, end float32) {
					rangeStart.Set(start)
					rangeEnd.Set(end)
				})),
				ui.VSpacerElement(10),
				ui.TextElement("Tick marks", ui.TextType(th.Types.BodyMedium)),
				ui.RangeSliderElement(tickRangeStart.Value(), tickRangeEnd.Value(), ui.SliderStep(10), ui.SliderTicks(true), ui.SliderOnRangeChange(func(ctx *ui.Context, start, end float32) {
					tickRangeStart.Set(start)
					tickRangeEnd.Set(end)
				})),
				ui.VSpacerElement(18),
				ui.TextElement("Custom styling", ui.TextType(th.Types.TitleMedium), ui.TextColor(th.Colors.OnSurface)),
				ui.VSpacerElement(8),
				ui.TextElement("Custom styles", ui.TextType(th.Types.BodyMedium)),
				ui.RangeSliderElement(
					customStart.Value(),
					customEnd.Value(),
					ui.SliderStep(10),
					ui.SliderTicks(true),
					ui.SliderTrackColor(ui.NRGBA(181, 204, 204, 255)),
					ui.SliderProgressColor(ui.NRGBA(55, 80, 80, 255)),
					ui.SliderThumbColor(ui.NRGBA(47, 70, 70, 255)),
					ui.SliderDecoration(ui.Pad(ui.Symmetric(2, 0))),
					ui.SliderOnRangeChange(func(ctx *ui.Context, start, end float32) {
						customStart.Set(start)
						customEnd.Set(end)
					}),
				),
				ui.VSpacerElement(10),
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
				ui.RowElement(ui.SliderElement(40, ui.SliderDisabled(true), ui.SliderWidth(240)), ui.HSpacerElement(12), ui.TextElement("disabled", ui.TextType(th.Types.BodySmall))),
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
			{Label: "Apple", Value: "apple", Leading: ui.Icon("nutrition")},
			{Label: "Apricot", Value: "apricot", Leading: ui.Icon("eco")},
			{Label: "Tomato", Value: "tomato", Disabled: true},
			{Label: "Cucumber", Value: "cucumber"},
		}
		return ui.FixedWidthElement(
			480,
			ui.ColumnElement(
				ui.OutlinedSelectElement(
					value.Value(),
					options,
					ui.SelectLabel[string]("Outlined select"),
					ui.SelectPlaceholder[string]("Choose fruit"),
					ui.SelectSupportingText[string]("Material 3 outlined field"),
					ui.SelectLeading[string](ui.Icon("restaurant")),
					ui.SelectSearchable[string](true),
					ui.SelectMaxHeight[string](180),
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
				ui.VSpacerElement(12),
				ui.FilledSelectElement(
					"apricot",
					options,
					ui.SelectLabel[string]("Filled select"),
					ui.SelectSupportingText[string]("Filled variant with active indicator"),
					ui.SelectMaxHeight[string](180),
				),
				ui.VSpacerElement(12),
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
					docsDemoControlButton("Set apricot", func(ctx *ui.Context) {
						value.Set("apricot")
						ref.Current.SetValue("apricot")
					}),
					ui.HSpacerElement(8),
					ui.TextElement(openState.Value(), ui.TextSize(12)),
				),
				ui.VSpacerElement(8),
				ui.RowElement(
					ui.ExpandedElement(ui.SelectElement("apple", options, ui.SelectDisabled[string](true), ui.SelectLabel[string]("Disabled"))),
					ui.HSpacerElement(12),
					ui.ExpandedElement(ui.SelectElement("", options, ui.SelectLabel[string]("Error"), ui.SelectError[string](true), ui.SelectErrorText[string]("Choose a value"), ui.SelectRequired[string](true))),
				),
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
