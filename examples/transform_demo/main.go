package main

import (
	"image/color"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.RunElement(App, ui.Title("变换示例"), ui.Size(620, 780))
}

func App(ctx *ui.Context) ui.Element {
	cardPad := ui.Pad(ui.Symmetric(6, 10)).WithRad(8)

	rotCard := func(label string, deg float32) ui.Element {
		return ui.PaddingElement(ui.All(6),
			ui.ContainerDecorationElement(
				ui.Bg(color.NRGBA{R: 66, G: 133, B: 244, A: 255}).Merge(cardPad).
					WithTransform(ui.Transform2D{RotateDeg: deg, ScaleX: 1, ScaleY: 1, Origin: ui.TransformCenter}).
					Merge(ui.Elevation(1)),
				ui.TextElement(label, ui.TextSize(13), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255}), ui.TextAlign(ui.AlignCenter)),
			),
		)
	}

	scaleCard := func(label string, sx, sy float32) ui.Element {
		return ui.PaddingElement(ui.All(6),
			ui.ContainerDecorationElement(
				ui.Bg(color.NRGBA{R: 40, G: 167, B: 69, A: 255}).Merge(cardPad).
					WithTransform(ui.Transform2D{ScaleX: sx, ScaleY: sy, Origin: ui.TransformCenter}),
				ui.TextElement(label, ui.TextSize(13), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255}), ui.TextAlign(ui.AlignCenter)),
			),
		)
	}

	offsetCard := func(label string, tx, ty float32) ui.Element {
		return ui.PaddingElement(ui.All(6),
			ui.ContainerDecorationElement(
				ui.Bg(color.NRGBA{R: 220, G: 53, B: 69, A: 255}).Merge(cardPad).
					WithTransform(ui.Transform2D{ScaleX: 1, ScaleY: 1, TranslateX: tx, TranslateY: ty, Origin: ui.TransformCenter}),
				ui.TextElement(label, ui.TextSize(13), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255}), ui.TextAlign(ui.AlignCenter)),
			),
		)
	}

	originCard := func(label string, origin ui.TransformOrigin) ui.Element {
		return ui.PaddingElement(ui.All(6),
			ui.ContainerDecorationElement(
				ui.Bg(color.NRGBA{R: 255, G: 152, B: 0, A: 255}).Merge(cardPad).
					WithTransform(ui.Transform2D{RotateDeg: -12, ScaleX: 1, ScaleY: 1, Origin: origin}),
				ui.TextElement(label, ui.TextSize(13), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255}), ui.TextAlign(ui.AlignCenter)),
			),
		)
	}

	return ui.ContainerDecorationElement(
		ui.Bg(color.NRGBA{R: 248, G: 250, B: 252, A: 255}).WithPad(ui.All(16)),
		ui.ColumnElement(
			ui.TextElement("FluxUI 2D 变换", ui.TextSize(20), ui.TextAlign(ui.AlignCenter)),
			ui.VSpacerElement(8),

			ui.TextElement("旋转 (Rotate)", ui.TextSize(14)),
			ui.RowElement(rotCard("0°", 0), rotCard("15°", 15), rotCard("345°", 345), rotCard("-10°", -10)),
			ui.PaddingElement(ui.TopBottom(2), ui.TextElement("围绕中心旋转，不改变占位空间", ui.TextSize(11), ui.TextColor(color.NRGBA{R: 148, G: 163, B: 184, A: 255}))),

			ui.VSpacerElement(8),
			ui.TextElement("缩放 (Scale)", ui.TextSize(14)),
			ui.RowElement(scaleCard("1:1", 1, 1), scaleCard("1.2×", 1.2, 1.2), scaleCard("0.8×", 0.8, 0.8), scaleCard("1.3×0.7", 1.3, 0.7)),
			ui.PaddingElement(ui.TopBottom(2), ui.TextElement("ScaleX/Y 控制各轴缩放", ui.TextSize(11), ui.TextColor(color.NRGBA{R: 148, G: 163, B: 184, A: 255}))),

			ui.VSpacerElement(8),
			ui.TextElement("平移 (Translate)", ui.TextSize(14)),
			ui.RowElement(offsetCard("原处", 0, 0), offsetCard("→8 ↓4", 8, 4), offsetCard("←6", -6, 0), offsetCard("↑10", 0, -10)),
			ui.PaddingElement(ui.TopBottom(2), ui.TextElement("TranslateX/Y dp 偏移", ui.TextSize(11), ui.TextColor(color.NRGBA{R: 148, G: 163, B: 184, A: 255}))),

			ui.VSpacerElement(8),
			ui.TextElement("原点对比 (Origin)", ui.TextSize(14)),
			ui.RowElement(
				originCard("TopLeft", ui.TransformTopLeft),
				originCard("Center", ui.TransformCenter),
				originCard("BottomRight", ui.TransformBottomRight),
			),
			ui.PaddingElement(ui.TopBottom(2), ui.TextElement("同角度旋转，不同原点效果不同", ui.TextSize(11), ui.TextColor(color.NRGBA{R: 148, G: 163, B: 184, A: 255}))),

			ui.VSpacerElement(8),
			ui.TextElement("组合变换", ui.TextSize(14)),
			ui.RowElement(
				ui.PaddingElement(ui.All(6),
					ui.ContainerDecorationElement(
						ui.Bg(color.NRGBA{R: 103, G: 58, B: 183, A: 255}).Merge(cardPad).
							WithTransform(ui.Transform2D{RotateDeg: 8, ScaleX: 1.15, ScaleY: 1.15, Origin: ui.TransformCenter}).
							Merge(ui.Elevation(1)),
						ui.TextElement("8°+1.15×", ui.TextSize(13), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255}), ui.TextAlign(ui.AlignCenter)),
					),
				),
				ui.PaddingElement(ui.All(6),
					ui.ContainerDecorationElement(
						ui.Bg(color.NRGBA{R: 0, G: 150, B: 136, A: 255}).Merge(cardPad).
							WithTransform(ui.Transform2D{RotateDeg: -5, ScaleX: 0.9, ScaleY: 0.9, TranslateX: 4, TranslateY: 2, Origin: ui.TransformCenter}),
						ui.TextElement("-5°+0.9+移", ui.TextSize(13), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255}), ui.TextAlign(ui.AlignCenter)),
					),
				),
			),

			ui.VSpacerElement(10),
			ui.TextElement("交互: Hover 放大", ui.TextSize(14)),
			ui.RowElement(
				ui.PaddingElement(ui.All(6),
					ui.ContainerDecorationElement(
						ui.Bg(color.NRGBA{R: 49, G: 107, B: 255, A: 255}).Merge(cardPad).Merge(ui.Elevation(1)).
							WithHover(
								ui.Bg(color.NRGBA{R: 30, G: 80, B: 200, A: 255}).Merge(cardPad).Merge(ui.Elevation(2)).
									WithTransform(ui.Transform2D{ScaleX: 1.1, ScaleY: 1.1, Origin: ui.TransformCenter}),
							),
						ui.TextElement("Hover 放大", ui.TextSize(13), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255}), ui.TextAlign(ui.AlignCenter)),
					),
				),
				ui.PaddingElement(ui.All(6),
					ui.ContainerDecorationElement(
						ui.Bg(color.NRGBA{R: 40, G: 167, B: 69, A: 255}).Merge(cardPad).Merge(ui.Elevation(1)).
							WithHover(
								ui.Bg(color.NRGBA{R: 20, G: 130, B: 45, A: 255}).Merge(cardPad).Merge(ui.Elevation(2)).
									WithTransform(ui.Transform2D{RotateDeg: 5, ScaleX: 1.08, ScaleY: 1.08, Origin: ui.TransformCenter}),
							),
						ui.TextElement("Hover 旋+放", ui.TextSize(13), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255}), ui.TextAlign(ui.AlignCenter)),
					),
				),
			),

			ui.VSpacerElement(10),
			ui.TextElement("便捷构造器", ui.TextSize(14)),
			ui.RowElement(
				ui.PaddingElement(ui.All(6),
					ui.ContainerDecorationElement(
						ui.Bg(color.NRGBA{R: 76, G: 175, B: 80, A: 255}).Merge(cardPad).
							Merge(ui.Rotate(10)),
						ui.TextElement("Rotate(10)", ui.TextSize(12), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
					),
				),
				ui.PaddingElement(ui.All(6),
					ui.ContainerDecorationElement(
						ui.Bg(color.NRGBA{R: 244, G: 67, B: 54, A: 255}).Merge(cardPad).
							Merge(ui.ScaleDeco(1.2, 0.85)),
						ui.TextElement("ScaleDeco", ui.TextSize(12), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
					),
				),
				ui.PaddingElement(ui.All(6),
					ui.ContainerDecorationElement(
						ui.Bg(color.NRGBA{R: 255, G: 152, B: 0, A: 255}).Merge(cardPad).
							Merge(ui.TranslateDeco(6, -4)),
						ui.TextElement("Translate", ui.TextSize(12), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
					),
				),
			),

			ui.VSpacerElement(10),
			ui.TextElement("圆形 + 旋转", ui.TextSize(14)),
			ui.RowElement(
				ui.PaddingElement(ui.All(6),
					ui.ContainerDecorationElement(
						ui.Bg(color.NRGBA{R: 66, G: 133, B: 244, A: 255}).WithPad(ui.All(28)).WithCircleClip().
							WithTransform(ui.Transform2D{RotateDeg: 30, ScaleX: 1, ScaleY: 1, Origin: ui.TransformCenter}),
						ui.TextElement("R", ui.TextSize(18), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
					),
				),
				ui.PaddingElement(ui.All(6),
					ui.ContainerDecorationElement(
						ui.Bg(color.NRGBA{R: 40, G: 167, B: 69, A: 255}).WithPad(ui.All(28)).WithCircleClip().
							WithTransform(ui.Transform2D{RotateDeg: -20, ScaleX: 1, ScaleY: 1, Origin: ui.TransformCenter}),
						ui.TextElement("G", ui.TextSize(18), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
					),
				),
				ui.PaddingElement(ui.All(6),
					ui.ContainerDecorationElement(
						ui.Bg(color.NRGBA{R: 220, G: 53, B: 69, A: 255}).WithPad(ui.All(28)).WithCircleClip().
							WithTransform(ui.Transform2D{RotateDeg: 10, ScaleX: 1, ScaleY: 1, Origin: ui.TransformCenter}),
						ui.TextElement("B", ui.TextSize(18), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
					),
				),
			),
			ui.PaddingElement(ui.TopBottom(2), ui.TextElement("CircleClip + Transform 旋转头像", ui.TextSize(11), ui.TextColor(color.NRGBA{R: 148, G: 163, B: 184, A: 255}))),

			ui.VSpacerElement(16),
			ui.TextElement("Phase 9 · 2D 变换 (RunElement)", ui.TextSize(11), ui.TextColor(color.NRGBA{R: 148, G: 163, B: 184, A: 255})),
		),
	)
}
