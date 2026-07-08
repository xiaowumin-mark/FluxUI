package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type docsFloat32State interface {
	Value() float32
	Set(float32)
}

func docsPressableDemo(clickCount docsIntState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
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
				ui.TextElement(fmt.Sprintf("点击次数：%d", clickCount.Value()), ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
				ui.VSpacerElement(8),
				ui.PressableElement(
					ui.FillWidthElement(
						ui.ContainerDecorationElement(
							ui.Bg(th.Colors.SecondaryContainer).
								WithPad(ui.All(14)).
								WithRad(8).
								WithHover(ui.Bg(th.Colors.PrimaryContainer)).
								WithPressed(ui.Bg(th.Colors.TertiaryContainer)),
							ui.TextElement(fmt.Sprintf("PressableElement 悬停=%t 按下=%t", hovered.Value(), pressed.Value()), ui.TextColor(th.Colors.OnSecondaryContainer)),
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
							ui.Bg(th.Colors.SurfaceContainerLow).
								WithPad(ui.Symmetric(6, 10)).
								WithRad(8).
								WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant}),
							ui.TextElement("ClickArea 兼容性", ui.TextSize(12), ui.TextColor(th.Colors.OnSurface)),
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
		ui.TextElement("默认文本"),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("大号文本", ui.TextSize(20))),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("主色文本", ui.TextColor(th.Primary))),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("标题缩放与显式行高", ui.TextType(th.Types.TitleMedium), ui.TextLineHeight(26))),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("作用域字体族", ui.TextFont(ui.FontFamily("Segoe UI")))),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("居中半粗体文本", ui.TextAlign(ui.AlignCenter), ui.TextFontWeight(ui.FontWeightSemiBold))),
	)
}

func docsDemoControlButton(label string, onClick func(*ui.Context)) ui.Element {
	return ui.OutlinedButtonElement(
		ui.TextElement(label, ui.TextSize(12)),
		ui.ButtonPadding(ui.Symmetric(5, 10)),
		ui.OnClick(onClick),
	)
}
