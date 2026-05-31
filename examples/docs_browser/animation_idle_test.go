//go:build visual

package main

import (
	"fmt"
	"image"
	"image/color"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"
	ui "github.com/xiaowumin-mark/FluxUI/ui"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestAnimationDemoSettlesWithoutIdleRedraw(t *testing.T) {
	runtime := internal.NewRuntime(ui.NewTheme(ui.LightColors()))
	defer runtime.Dispose()

	root := ui.VisualRootBuilder(animationIdleDemoRoot)
	baseTime := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)

	var ops op.Ops
	for frame := 0; frame < 30; frame++ {
		redraws := 0
		runtime.SetInvalidator(func() {
			redraws++
		})

		ops.Reset()
		gtx := gioLayout.Context{
			Constraints: gioLayout.Exact(image.Pt(620, 340)),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         baseTime.Add(time.Duration(frame) * 16 * time.Millisecond),
			Ops:         &ops,
		}

		runtime.BeginFrame()
		ctx := runtime.Frame(gtx)
		if widget := root(ctx.Scope("build")); widget != nil {
			widget.Layout(ctx.Scope("tree"))
		}
		runtime.EndFrame()

		if frame > 0 && redraws != 0 {
			t.Fatalf("frame %d requested %d redraw(s) while animation demo target was stable", frame, redraws)
		}
	}
}

func animationIdleDemoRoot(ctx *ui.Context) ui.Element {
	const (
		trackWidth  = float32(156)
		squareSize  = float32(28)
		trackHeight = float32(36)
		target      = float32(0)
	)

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

	return ui.ColumnElement(rows...)
}
