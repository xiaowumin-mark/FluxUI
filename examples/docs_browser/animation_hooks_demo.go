package main

import (
	"fmt"
	"image/color"
	"time"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type docsIntState interface {
	Value() int
	Set(int)
}

type docsStringSliceState interface {
	Value() []string
	Set([]string)
}

func buildDocsAnimationDemo(demoCtx *ui.Context, active docsBoolState) ui.Element {
	const (
		trackWidth  = float32(156)
		squareSize  = float32(28)
		trackHeight = float32(36)
	)
	target := float32(0)
	if active.Value() {
		target = trackWidth - squareSize
	}

	type curveDemo struct {
		name   string
		short  string
		color  color.NRGBA
		easing ui.Easing
	}

	curves := []curveDemo{
		{name: "EaseOut", short: "EO", color: ui.NRGBA(59, 130, 246, 255), easing: ui.EaseOut},
		{name: "EaseInOut", short: "IO", color: ui.NRGBA(34, 197, 94, 255), easing: ui.EaseInOut},
		{name: "OutBack", short: "BK", color: ui.NRGBA(251, 146, 60, 255), easing: ui.EaseOutBack},
		{name: "Elastic", short: "EL", color: ui.NRGBA(168, 85, 247, 255), easing: easeOutElastic},
		{name: "Bounce", short: "BO", color: ui.NRGBA(239, 68, 68, 255), easing: ui.EaseOutBounce},
		{name: "Linear", short: "LN", color: ui.NRGBA(100, 116, 139, 255), easing: ui.Linear},
	}

	curveCard := func(curve curveDemo, index int) ui.Element {
		return ui.Key(curve.name, ui.ComponentElement(func(squareCtx *ui.Context) ui.Element {
			hovered := ui.UseState(squareCtx, false)
			x := ui.UseAnimatedValue(squareCtx, target, 640*time.Millisecond, curve.easing)
			scale := ui.UseAnimatedValue(squareCtx, hoverScale(hovered.Value()), 150*time.Millisecond, ui.EaseOut)

			square := ui.ContainerDecorationElement(
				ui.Bg(curve.color).
					WithRad(8).
					Merge(ui.Shadow(0, 6, 14, color.NRGBA{R: curve.color.R, G: curve.color.G, B: curve.color.B, A: 56})).
					WithTransform(ui.Transform2D{ScaleX: scale, ScaleY: scale, Origin: ui.TransformCenter}),
				ui.FixedSizeElement(
					28,
					28,
					ui.CenterElement(ui.TextElement(curve.short, ui.TextSize(10), ui.TextColor(ui.NRGBA(255, 255, 255, 255)))),
				),
				ui.OnDecoHoverEnter(func(ctx *ui.Context) {
					hovered.Set(true)
				}),
				ui.OnDecoHoverLeave(func(ctx *ui.Context) {
					hovered.Set(false)
				}),
			)

			track := ui.FixedSizeElement(
				trackWidth,
				trackHeight,
				ui.StackElement(
					ui.ContainerDecorationElement(
						ui.Bg(ui.NRGBA(226, 232, 240, 255)).WithRad(12),
						ui.SpacerElement(trackWidth, trackHeight),
					),
					ui.PaddingElement(
						ui.Insets{Left: x, Top: (trackHeight - squareSize) / 2},
						square,
					),
				),
			)

			return ui.ContainerDecorationElement(
				ui.Bg(ui.NRGBA(248, 250, 252, 255)).
					WithPad(ui.All(8)).
					WithRad(12).
					WithBorder(ui.Border{Width: 1, Color: ui.NRGBA(226, 232, 240, 255)}),
				ui.ColumnElement(
					ui.RowElement(
						ui.FixedWidthElement(86, ui.TextElement(curve.name, ui.TextSize(12), ui.TextColor(ui.NRGBA(71, 85, 105, 255)))),
						ui.TextElement(fmt.Sprintf("#%d", index+1), ui.TextSize(11), ui.TextColor(ui.NRGBA(148, 163, 184, 255))),
					),
					ui.PaddingElement(ui.Insets{Top: 6}, track),
				),
			)
		}))
	}

	rows := make([]ui.Element, 0, 3)
	for i := 0; i < len(curves); i += 2 {
		rowItems := []ui.Element{
			ui.FixedWidthElement(245, curveCard(curves[i], i)),
		}
		if i+1 < len(curves) {
			rowItems = append(rowItems,
				ui.PaddingElement(ui.Insets{Left: 10}, ui.FixedWidthElement(245, curveCard(curves[i+1], i+1))),
			)
		}
		rows = append(rows, ui.PaddingElement(ui.Insets{Bottom: 8}, ui.RowElement(rowItems...)))
	}

	return ui.ColumnElement(
		ui.TextElement("Easing 曲线对比", ui.TextSize(15), ui.TextColor(ui.NRGBA(15, 23, 42, 255))),
		ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("点击触发位移动画；悬浮方块查看 transform hover 缩放。", ui.TextSize(12), ui.TextColor(ui.NRGBA(100, 116, 139, 255)))),
		ui.PaddingElement(ui.Insets{Top: 10}, ui.ColumnElement(rows...)),
		ui.PaddingElement(
			ui.Insets{Top: 8, Bottom: 8},
			ui.FixedWidthElement(
				160,
				ui.ButtonElement(
					ui.TextElement("触发动画"),
					ui.ButtonDecoration(
						ui.Bg(ui.NRGBA(15, 23, 42, 255)).
							WithPad(ui.Symmetric(8, 12)).
							WithRad(10).
							WithHover(ui.Bg(ui.NRGBA(30, 41, 59, 255))).
							WithPressed(ui.Bg(ui.NRGBA(51, 65, 85, 255))),
					),
					ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
					ui.OnClick(func(ctx *ui.Context) {
						active.Set(!active.Value())
					}),
				),
			),
		),
	)
}
