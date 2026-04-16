package main

import (
	"fmt"
	"image/color"
	"time"

	"fluxui/ui"
)

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)
		mode := ui.State[int](ctx) // 1: 展开, -1: 收起, 0: 静止(默认收起)

		var progress float32
		switch mode.Value() {
		case 1:
			progress = ui.Animate(
				ui.From(0),
				ui.To(1),
				ui.Duration(550*time.Millisecond),
				ui.Ease(ui.EaseInOut),
			).Value(ctx)
		case -1:
			progress = ui.Animate(
				ui.From(1),
				ui.To(0),
				ui.Duration(550*time.Millisecond),
				ui.Ease(ui.EaseInOut),
			).Value(ctx)
		default:
			progress = 0
		}

		cardPadding := lerpFloat(10, 26, progress)
		cardRadius := lerpFloat(8, 50, progress)
		titleSize := lerpFloat(16, 28, progress)
		bodySize := lerpFloat(12, 16, progress)

		cardColor := lerpColor(th.SurfaceMuted, th.Primary, progress)
		titleColor := lerpColor(th.TextColor, th.TextOnPrimary, progress)
		bodyColor := lerpColor(th.TextColor, th.TextOnPrimary, progress)

		status := "当前状态: 收起"
		if progress >= 0.99 {
			status = "当前状态: 展开"
		}

		return ui.Container(
			ui.Style{
				Background: th.Surface,
				Padding:    ui.All(20),
			},
			ui.Column(
				ui.Padding(
					ui.All(8),
					ui.Text("动画示例", ui.TextSize(24), ui.TextAlign(ui.AlignCenter)),
				),
				ui.Padding(
					ui.All(8),
					ui.Text("使用 Animate + frame tick 驱动布局与颜色过渡", ui.TextSize(14), ui.TextColor(th.SurfaceMuted)),
				),
				ui.Row(
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("展开"),
							ui.OnClick(func(ctx *ui.Context) {
								mode.Set(1)
							}),
						),
					),
					ui.Padding(
						ui.All(4),
						ui.Button(
							ui.Text("收起"),
							ui.OnClick(func(ctx *ui.Context) {
								mode.Set(-1)
							}),
						),
					),
				),
				ui.Padding(
					ui.All(8),
					ui.Text(
						fmt.Sprintf("%s | 进度: %.2f", status, progress),
						ui.TextSize(13),
						ui.TextColor(th.SurfaceMuted),
					),
				),
				ui.Container(
					ui.Style{
						Background: cardColor,
						Padding:    ui.All(cardPadding),
						Radius:     cardRadius,
					},
					ui.Column(
						ui.Text("FluxUI 动画卡片", ui.TextSize(titleSize), ui.TextColor(titleColor)),
						ui.Padding(
							ui.All(4),
							ui.Text(
								"尺寸、圆角、颜色与字号由同一个动画进度统一控制。",
								ui.TextSize(bodySize),
								ui.TextColor(bodyColor),
							),
						),
						ui.Padding(
							ui.All(4),
							ui.Text(
								fmt.Sprintf("padding=%.1f, radius=%.1f", cardPadding, cardRadius),
								ui.TextSize(12),
								ui.TextColor(bodyColor),
							),
						),
					),
				),
			),
		)
	}, ui.Title("动画示例"), ui.Size(560, 420))
}

func lerpFloat(from, to, t float32) float32 {
	t = clamp01(t)
	return from + (to-from)*t
}

func lerpColor(from, to color.NRGBA, t float32) color.NRGBA {
	t = clamp01(t)
	return color.NRGBA{
		R: uint8(float32(from.R) + (float32(to.R)-float32(from.R))*t),
		G: uint8(float32(from.G) + (float32(to.G)-float32(from.G))*t),
		B: uint8(float32(from.B) + (float32(to.B)-float32(from.B))*t),
		A: uint8(float32(from.A) + (float32(to.A)-float32(from.A))*t),
	}
}

func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
