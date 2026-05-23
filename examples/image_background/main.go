package main

import (
	"image"
	"image/color"

	ui "github.com/xiaowumin-mark/FluxUI/ui"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

var (
	gradA   = image.NewUniform(color.NRGBA{R: 49, G: 107, B: 255, A: 255})
	gradB   = image.NewUniform(color.NRGBA{R: 40, G: 167, B: 69, A: 255})
	gradC   = image.NewUniform(color.NRGBA{R: 220, G: 53, B: 69, A: 255})
	gradD   = image.NewUniform(color.NRGBA{R: 255, G: 152, B: 0, A: 255})
	pattern = checkerPattern(32)
)

func checkerPattern(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 2*size, 2*size))
	for y := 0; y < 2*size; y++ {
		for x := 0; x < 2*size; x++ {
			c := color.NRGBA{R: 240, G: 240, B: 245, A: 255}
			if (x/size)%2 == (y/size)%2 {
				c = color.NRGBA{R: 220, G: 220, B: 230, A: 255}
			}
			img.Set(x, y, c)
		}
	}
	return img
}

func main() {
	_ = ui.RunElement(App, ui.Title("背景图片示例"), ui.Size(620, 780))
}

func App(ctx *ui.Context) ui.Element {
	selected := ui.UseState(ctx, 0)

	imgCard := func(num int, label string, deco ui.Decoration) ui.Element {
		isSel := selected.Value() == num
		var b ui.Border
		if isSel {
			b = ui.Border{Width: 2, Color: color.NRGBA{R: 49, G: 107, B: 255, A: 255}}
		} else {
			b = ui.Border{Width: 1, Color: color.NRGBA{R: 203, G: 213, B: 225, A: 255}}
		}
		return ui.PaddingElement(
			ui.All(6),
			ui.ContainerDecorationElement(
				deco.WithBorder(b),
				ui.ColumnElement(
					ui.TextElement(label, ui.TextSize(12), ui.TextAlign(ui.AlignCenter)),
					ui.VSpacerElement(4),
				),
				ui.OnDecoClick(func(c *ui.Context) { selected.Set(num) }),
			),
		)
	}

	caption := func(t string) ui.Element {
		return ui.PaddingElement(
			ui.TopBottom(2),
			ui.TextElement(t, ui.TextSize(12), ui.TextColor(color.NRGBA{R: 148, G: 163, B: 184, A: 255})),
		)
	}

	cardSize := ui.Bg(color.NRGBA{}).WithPad(ui.All(10)).WithRad(10)

	return ui.ContainerDecorationElement(
		ui.Bg(color.NRGBA{R: 248, G: 250, B: 252, A: 255}).WithPad(ui.All(16)),
		ui.ColumnElement(
			ui.TextElement("FluxUI 背景图片", ui.TextSize(20), ui.TextAlign(ui.AlignCenter)),
			ui.VSpacerElement(8),

			ui.TextElement("Cover 模式 (保持比例，填满容器)", ui.TextSize(14)),
			ui.RowElement(
				imgCard(1, "Cover·蓝",
					ui.Bg(color.NRGBA{R: 200, G: 210, B: 240, A: 255}).Merge(cardSize).
						WithImage(ui.ImageFill{Src: gradA, Fit: ui.ImageFillCover}).Merge(ui.Elevation(1)),
				),
				imgCard(2, "Cover·红",
					ui.Bg(color.NRGBA{R: 240, G: 200, B: 200, A: 255}).Merge(cardSize).
						WithImage(ui.ImageFill{Src: gradC, Fit: ui.ImageFillCover}).Merge(ui.Elevation(1)),
				),
			),

			ui.VSpacerElement(6),
			ui.TextElement("Contain 模式 (保持比例，完整显示)", ui.TextSize(14)),
			ui.RowElement(
				imgCard(3, "Contain·蓝",
					ui.Bg(color.NRGBA{R: 210, G: 220, B: 240, A: 255}).Merge(cardSize).
						WithImage(ui.ImageFill{Src: gradA, Fit: ui.ImageFillContain}),
				),
				imgCard(4, "Contain·绿",
					ui.Bg(color.NRGBA{R: 210, G: 240, B: 220, A: 255}).Merge(cardSize).
						WithImage(ui.ImageFill{Src: gradB, Fit: ui.ImageFillContain}),
				),
			),

			ui.VSpacerElement(6),
			ui.TextElement("Fill 模式 (拉伸填满)", ui.TextSize(14)),
			ui.RowElement(
				imgCard(5, "Fill·蓝",
					ui.Bg(color.NRGBA{R: 220, G: 230, B: 248, A: 255}).Merge(cardSize).
						WithImage(ui.ImageFill{Src: gradA, Fit: ui.ImageFillFill}),
				),
				imgCard(6, "Fill·橙",
					ui.Bg(color.NRGBA{R: 248, G: 235, B: 220, A: 255}).Merge(cardSize).
						WithImage(ui.ImageFill{Src: gradD, Fit: ui.ImageFillFill}),
				),
			),

			ui.VSpacerElement(6),
			ui.TextElement("None 模式 (原始尺寸)", ui.TextSize(14)),
			ui.RowElement(
				imgCard(7, "None·纹",
					ui.Bg(color.NRGBA{R: 235, G: 235, B: 240, A: 255}).Merge(cardSize).
						WithImage(ui.ImageFill{Src: pattern, Fit: ui.ImageFillNone}),
				),
				imgCard(8, "None·原",
					ui.Bg(color.NRGBA{R: 240, G: 240, B: 245, A: 255}).Merge(cardSize).
						WithImage(ui.ImageFill{Src: pattern, Fit: ui.ImageFillNone}),
				),
			),

			ui.VSpacerElement(8),
			ui.TextElement("圆形裁切 + 图片背景", ui.TextSize(14)),
			ui.RowElement(
				ui.PaddingElement(ui.All(6), ui.ContainerDecorationElement(
					ui.Bg(color.NRGBA{R: 230, G: 238, B: 250, A: 255}).WithPad(ui.All(36)).WithCircleClip().
						WithImage(ui.ImageFill{Src: gradA, Fit: ui.ImageFillCover}),
					nil,
				)),
				ui.PaddingElement(ui.All(6), ui.ContainerDecorationElement(
					ui.Bg(color.NRGBA{R: 240, G: 225, B: 225, A: 255}).WithPad(ui.All(36)).WithCircleClip().
						WithImage(ui.ImageFill{Src: gradC, Fit: ui.ImageFillCover}),
					nil,
				)),
				ui.PaddingElement(ui.All(6), ui.ContainerDecorationElement(
					ui.Bg(color.NRGBA{R: 225, G: 240, B: 230, A: 255}).WithPad(ui.All(36)).WithCircleClip().
						WithImage(ui.ImageFill{Src: gradB, Fit: ui.ImageFillCover}),
					nil,
				)),
			),
			caption("CircleClip + ImageFillCover: 头像型背景"),

			ui.VSpacerElement(8),
			ui.TextElement("交互态: Hover 切换图片", ui.TextSize(14)),
			ui.RowElement(
				ui.PaddingElement(ui.All(6), ui.ContainerDecorationElement(
					ui.Bg(color.NRGBA{R: 225, G: 238, B: 255, A: 255}).WithPad(ui.All(10)).WithRad(10).
						WithImage(ui.ImageFill{Src: gradA, Fit: ui.ImageFillCover}).
						WithHover(
							ui.Bg(color.NRGBA{R: 180, G: 200, B: 240, A: 255}).WithPad(ui.All(10)).WithRad(10).
								WithImage(ui.ImageFill{Src: gradB, Fit: ui.ImageFillCover}),
						),
					ui.ColumnElement(
						ui.TextElement("Hover 换图", ui.TextSize(13), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
						ui.TextElement("悬停变绿", ui.TextSize(10), ui.TextColor(color.NRGBA{R: 220, G: 230, B: 255, A: 255})),
					),
				)),
				ui.PaddingElement(ui.All(6), ui.ContainerDecorationElement(
					ui.Bg(color.NRGBA{R: 248, G: 235, B: 235, A: 255}).WithPad(ui.All(10)).WithRad(10).
						WithImage(ui.ImageFill{Src: gradC, Fit: ui.ImageFillCover}).
						WithPressed(
							ui.Bg(color.NRGBA{R: 180, G: 160, B: 180, A: 255}).WithPad(ui.All(10)).WithRad(10).
								WithImage(ui.ImageFill{Src: gradD, Fit: ui.ImageFillCover}),
						),
					ui.ColumnElement(
						ui.TextElement("Press 换图", ui.TextSize(13), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
						ui.TextElement("按下变橙", ui.TextSize(10), ui.TextColor(color.NRGBA{R: 255, G: 210, B: 210, A: 255})),
					),
				)),
			),
			caption("Hover/Press 态通过 Decoration Merge 继承 Image"),

			ui.VSpacerElement(8),
			ui.TextElement("图片 + 边框", ui.TextSize(14)),
			ui.PaddingElement(ui.All(6), ui.ContainerDecorationElement(
				ui.Bg(color.NRGBA{R: 235, G: 240, B: 248, A: 255}).WithPad(ui.All(10)).WithRad(12).
					WithImage(ui.ImageFill{Src: gradA, Fit: ui.ImageFillContain}).
					WithBorder(ui.Border{Width: 2, Color: color.NRGBA{R: 49, G: 107, B: 255, A: 255}}),
				ui.ColumnElement(
					ui.TextElement("带边框 + 图片背景", ui.TextSize(14), ui.TextAlign(ui.AlignCenter)),
					ui.TextElement("Contain 模式，底色为 Fallback", ui.TextSize(11), ui.TextColor(color.NRGBA{R: 148, G: 163, B: 184, A: 255})),
				),
			)),

			ui.VSpacerElement(16),
			caption("Phase 8 · 背景图片 (RunElement + Element API)"),
		),
	)
}
