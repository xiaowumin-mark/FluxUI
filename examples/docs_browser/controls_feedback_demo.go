package main

import (
	"fmt"
	"time"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type docsFloat32State interface {
	Value() float32
	Set(float32)
}

func docsPressableDemo(clickCount docsIntState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		pressableRef := ui.UseRef(ctx, ui.NewPressableRef())
		clickAreaRef := ui.UseRef(ctx, ui.NewClickAreaRef())
		hovered := ui.UseState(ctx, false)
		pressed := ui.UseState(ctx, false)
		if pressableRef.Current == nil {
			pressableRef.Current = ui.NewPressableRef()
		}
		if clickAreaRef.Current == nil {
			clickAreaRef.Current = ui.NewClickAreaRef()
		}

		return ui.FixedWidthElement(
			520,
			ui.ColumnElement(
				ui.TextElement(fmt.Sprintf("Click count: %d", clickCount.Value()), ui.TextSize(13)),
				ui.VSpacerElement(8),
				ui.PressableElement(
					ui.FillWidthElement(
						ui.ContainerDecorationElement(
							ui.Bg(ui.NRGBA(227, 242, 253, 255)).
								WithPad(ui.All(14)).
								WithRad(8).
								WithHover(ui.Bg(ui.NRGBA(219, 234, 254, 255))).
								WithPressed(ui.Bg(ui.NRGBA(191, 219, 254, 255))),
							ui.TextElement(fmt.Sprintf("PressableElement hovered=%t pressed=%t", hovered.Value(), pressed.Value())),
							ui.OnDecoHover(func(ctx *ui.Context, value bool) {
								hovered.Set(value)
							}),
							ui.OnDecoPressed(func(ctx *ui.Context, value bool) {
								pressed.Set(value)
							}),
						),
					),
					func(ctx *ui.Context) {
						clickCount.Set(clickCount.Value() + 1)
					},
					ui.PressableAttachRef(pressableRef.Current),
				),
				ui.VSpacerElement(8),
				ui.RowElement(
					docsDemoControlButton("PressableRef.Click", func(ctx *ui.Context) {
						clickCount.Set(clickCount.Value() + 1)
						pressableRef.Current.Click()
					}),
					ui.HSpacerElement(8),
					ui.ClickAreaElement(
						ui.ContainerDecorationElement(
							ui.Bg(ui.NRGBA(248, 250, 252, 255)).
								WithPad(ui.Symmetric(6, 10)).
								WithRad(8).
								WithBorder(ui.Border{Width: 1, Color: ui.NRGBA(203, 213, 225, 255)}),
							ui.TextElement("ClickArea compatibility", ui.TextSize(12)),
						),
						func(ctx *ui.Context) {
							clickCount.Set(clickCount.Value() + 1)
						},
						ui.ClickAreaAttachRef(clickAreaRef.Current),
					),
				),
			),
		)
	})
}

func docsTextDemo(th *ui.Theme) ui.Element {
	return ui.ColumnElement(
		ui.TextElement("Default text"),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("Large text", ui.TextSize(20))),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("Primary color text", ui.TextColor(th.Primary))),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("Title scale with explicit line height", ui.TextType(th.Types.TitleMedium), ui.TextLineHeight(26))),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("Scoped font family", ui.TextFont(ui.FontFamily("Segoe UI")))),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("Centered semibold text", ui.TextAlign(ui.AlignCenter), ui.TextFontWeight(ui.FontWeightSemiBold))),
	)
}

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
			ui.IconElement("H", ui.IconSize(20), ui.IconColor(ui.NRGBA(30, 136, 229, 255)), ui.IconAttachRef(ref.Current), ui.IconOnClick(func(ctx *ui.Context) {
				clicks.Set(clicks.Value() + 1)
			})),
			ui.PaddingElement(ui.Insets{Left: 12}, ui.IconElement("S", ui.IconSize(20), ui.IconColor(ui.NRGBA(67, 160, 71, 255)))),
			ui.PaddingElement(ui.Insets{Left: 12}, ui.IconElement("G", ui.IconSize(20), ui.IconColor(ui.NRGBA(245, 124, 0, 255)))),
			ui.PaddingElement(ui.Insets{Left: 12}, docsDemoControlButton(fmt.Sprintf("clicks %d", clicks.Value()), func(ctx *ui.Context) {
				clicks.Set(clicks.Value() + 1)
				ref.Current.Click()
			})),
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

func docsMenuDemo(menuOpen docsBoolState, menuValue docsStringState, th *ui.Theme) ui.Element {
	items := []ui.MenuItem{
		{Key: "copy", Label: "Copy"},
		{Key: "share", Label: "Share"},
		{Key: "archive", Label: "Archive"},
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
					ui.DropdownMenuMaxHeight(180),
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
					ui.MenuDecoration(ui.Bg(th.Colors.SurfaceContainerLow).WithRad(10)),
					ui.MenuOnSelect(func(ctx *ui.Context, key string) {
						menuValue.Set(key)
					}),
				),
			),
		),
	)
}

func docsListItemDemo(selected docsStringState) ui.Element {
	return ui.FixedWidthElement(
		380,
		ui.ColumnElement(
			ui.ListItemElementWithSlots(
				ui.TextElement("Inbox"),
				ui.TextElement("12 unread messages"),
				ui.IconElement("I"),
				ui.TextElement("12"),
				ui.ListItemSelected(selected.Value() == "inbox"),
				ui.ListItemMinHeight(64),
				ui.ListItemDecoration(ui.Bg(ui.NRGBA(248, 250, 252, 255)).WithRad(10)),
				ui.ListItemOnClick(func(ctx *ui.Context) { selected.Set("inbox") }),
			),
			ui.ListItemElementWithSlots(
				ui.TextElement("Archive"),
				ui.TextElement("Older conversations"),
				ui.IconElement("A"),
				nil,
				ui.ListItemSelected(selected.Value() == "archive"),
				ui.ListItemOnClick(func(ctx *ui.Context) { selected.Set("archive") }),
			),
			ui.ListItemElement("Compact item", ui.ListItemOnClick(func(ctx *ui.Context) { selected.Set("compact") })),
			ui.ListItemElementWithSlots(
				ui.TextElement("Disabled"),
				ui.TextElement("Unavailable item"),
				ui.IconElement("D"),
				nil,
				ui.ListItemDisabled(true),
			),
		),
	)
}

func docsIconButtonDemo(selected docsBoolState) ui.Element {
	return ui.RowElement(
		ui.PaddingElement(ui.Insets{Right: 10}, ui.IconButtonElement(ui.IconElement("S"), ui.IconButtonSelected(selected.Value()), ui.IconButtonOnClick(func(ctx *ui.Context) {
			selected.Set(!selected.Value())
		}))),
		ui.PaddingElement(ui.Insets{Right: 10}, ui.FilledIconButtonElement(ui.IconElement("F"), ui.IconButtonSelected(true), ui.IconButtonSize(42))),
		ui.PaddingElement(ui.Insets{Right: 10}, ui.FilledTonalIconButtonElement(ui.IconElement("T"), ui.IconButtonBackground(ui.NRGBA(219, 234, 254, 255)), ui.IconButtonForeground(ui.NRGBA(30, 64, 175, 255)))),
		ui.PaddingElement(ui.Insets{Right: 10}, ui.OutlinedIconButtonElement(ui.IconElement("O"), ui.IconButtonDecoration(ui.BorderDeco(1, ui.NRGBA(148, 163, 184, 255))))),
		ui.OutlinedIconButtonElement(ui.IconElement("D"), ui.IconButtonDisabled(true)),
	)
}

func docsFloatingActionButtonDemo(count docsIntState) ui.Element {
	click := ui.FloatingActionButtonOnClick(func(ctx *ui.Context) {
		count.Set(count.Value() + 1)
	})
	return ui.ColumnElement(
		ui.RowElement(
			ui.PaddingElement(ui.Insets{Right: 12}, ui.SmallFloatingActionButtonElement(ui.IconElement("+"), click)),
			ui.PaddingElement(ui.Insets{Right: 12}, ui.FloatingActionButtonElement(ui.IconElement("+"), click)),
			ui.PaddingElement(ui.Insets{Right: 12}, ui.LargeFloatingActionButtonElement(ui.IconElement("+"), click, ui.FloatingActionButtonBackground(ui.NRGBA(37, 99, 235, 255)), ui.FloatingActionButtonForeground(ui.NRGBA(255, 255, 255, 255)))),
			ui.PaddingElement(ui.Insets{Right: 12}, ui.ExtendedFloatingActionButtonElement(ui.IconElement("+"), ui.TextElement("Create"), click, ui.FloatingActionButtonDecoration(ui.Bg(ui.NRGBA(220, 252, 231, 255)).WithRad(16)))),
			ui.FloatingActionButtonElement(ui.IconElement("x"), ui.FloatingActionButtonDisabled(true)),
		),
		ui.PaddingElement(ui.Insets{Top: 12}, ui.TextElement(fmt.Sprintf("FAB clicks: %d", count.Value()), ui.TextSize(13))),
	)
}

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

func docsTabsDemo(value docsStringState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		ref := ui.UseRef(ctx, ui.NewTabsRef())
		if ref.Current == nil {
			ref.Current = ui.NewTabsRef()
		}
		items := []ui.TabItem{
			{Key: "overview", Label: "Overview"},
			{Key: "api", Label: "API"},
			{Key: "example", Label: "Example"},
			{Key: "notes", Label: "Notes"},
		}
		return ui.FixedWidthElement(
			520,
			ui.ColumnElement(
				ui.TabsElement(
					value.Value(),
					items,
					ui.TabsScrollable(true),
					ui.TabsIndicatorColor(ui.NRGBA(37, 99, 235, 255)),
					ui.TabsTextColor(ui.NRGBA(71, 85, 105, 255)),
					ui.TabsActiveTextColor(ui.NRGBA(30, 64, 175, 255)),
					ui.TabsDecoration(ui.Bg(ui.NRGBA(248, 250, 252, 255)).WithRad(10)),
					ui.TabsTabDecoration(ui.HoverBg(ui.NRGBA(239, 246, 255, 255))),
					ui.TabsAttachRef(ref.Current),
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
				ui.TextElement("This is a docs browser Dialog example."),
				ui.DialogTitle("Dialog example"),
				ui.DialogWidth(340),
				ui.DialogRadius(18),
				ui.DialogMaskClosable(true),
				ui.DialogConfirmText("Apply"),
				ui.DialogCancelText("Dismiss"),
				ui.DialogDecoration(ui.Bg(ui.NRGBA(255, 255, 255, 255)).WithRad(18).Merge(ui.Elevation(4))),
				ui.DialogMaskColor(ui.NRGBA(15, 23, 42, 255)),
				ui.DialogMaskAlpha(90),
				ui.DialogAttachRef(ref.Current),
				ui.DialogOnOpenChange(func(ctx *ui.Context, next bool) {
					open.Set(next)
				}),
				ui.DialogOnCancel(func(ctx *ui.Context) {
					open.Set(false)
					ref.Current.Close()
				}),
				ui.DialogOnConfirm(func(ctx *ui.Context) {
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
				ui.PopupPadding(ui.All(16)),
				ui.PopupRadius(12),
				ui.PopupMaskClosable(true),
				ui.PopupMaskColor(ui.NRGBA(15, 23, 42, 255)),
				ui.PopupMaskAlpha(72),
				ui.PopupDecoration(ui.Bg(ui.NRGBA(255, 255, 255, 255)).WithRad(12).Merge(ui.Elevation(3))),
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

func docsBadgeDemo() ui.Element {
	return ui.RowElement(
		ui.PaddingElement(
			ui.Insets{Right: 24},
			ui.BadgeElement(
				ui.IconButtonElement(ui.IconElement("M")),
				"3",
				ui.BadgeBackground(ui.NRGBA(220, 38, 38, 255)),
				ui.BadgeForeground(ui.NRGBA(255, 255, 255, 255)),
				ui.BadgeOffset(3, -3),
			),
		),
		ui.PaddingElement(
			ui.Insets{Right: 24},
			ui.BadgeElement(
				ui.IconButtonElement(ui.IconElement("N")),
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
		ui.RowElement(
			ui.PaddingElement(
				ui.Insets{Right: 8},
				ui.AssistChipElement("Assist", ui.ChipLeading(ui.Icon("i", ui.IconSize(16)))),
			),
			ui.PaddingElement(
				ui.Insets{Right: 8},
				ui.FilterChipElement(
					"Filter",
					ui.ChipSelected(selected.Value()),
					ui.ChipOnClick(func(ctx *ui.Context) {
						selected.Set(!selected.Value())
					}),
				),
			),
			ui.PaddingElement(
				ui.Insets{Right: 8},
				ui.InputChipElement("Input", ui.ChipTrailing(ui.Icon("x", ui.IconSize(14)))),
			),
			ui.PaddingElement(
				ui.Insets{Right: 8},
				ui.SuggestionChipElement("Suggestion"),
			),
			ui.AssistChipElement("Disabled", ui.ChipDisabled(true)),
		),
		ui.VSpacerElement(8),
		ui.ChipElementWithSlots(
			ui.RowElement(
				ui.IconElement("S", ui.IconSize(14), ui.IconColor(ui.NRGBA(30, 64, 175, 255))),
				ui.HSpacerElement(6),
				ui.TextElement("Styled slots", ui.TextSize(12), ui.TextColor(ui.NRGBA(30, 64, 175, 255))),
			),
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
				ui.SearchBarLeading(ui.Icon("S", ui.IconSize(18))),
				ui.SearchBarTrailing(ui.Icon("x", ui.IconSize(16))),
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

func docsProgressIndicatorsDemo(value docsFloat32State, th *ui.Theme) ui.Element {
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
					value.Value(),
					ui.ProgressMin(0),
					ui.ProgressMax(100),
					ui.ProgressTrackColor(ui.NRGBA(226, 232, 240, 255)),
					ui.ProgressFillColor(th.Primary),
					ui.ProgressDecoration(ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.All(6)).WithRad(8)),
				),
			),
			ui.PaddingElement(
				ui.Insets{Top: 16},
				ui.CircularProgressIndicatorElement(
					value.Value(),
					ui.ProgressMin(0),
					ui.ProgressMax(100),
					ui.ProgressSize(72),
					ui.ProgressFillColor(th.Primary),
					ui.ProgressLabelVisible(true),
				),
			),
			ui.PaddingElement(
				ui.Insets{Top: 16},
				ui.LinearProgressIndicatorElement(0, ui.ProgressIndeterminate(true), ui.ProgressFillColor(th.Colors.Tertiary)),
			),
		),
	)
}

func docsDemoControlButton(label string, onClick func(*ui.Context)) ui.Element {
	return ui.OutlinedButtonElement(
		ui.TextElement(label, ui.TextSize(12)),
		ui.ButtonPadding(ui.Symmetric(5, 10)),
		ui.OnClick(onClick),
	)
}

func minFloat32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
