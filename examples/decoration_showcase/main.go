package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/xiaowumin-mark/FluxUI/state"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.RunElement(App, ui.Title("FluxUI Decoration 体验"), ui.Size(720, 1700))
}

var (
	blue   = ui.NRGBA(33, 133, 209, 255)
	green  = ui.NRGBA(40, 167, 69, 255)
	orange = ui.NRGBA(255, 193, 7, 255)
	purple = ui.NRGBA(155, 89, 182, 255)
	gray   = ui.NRGBA(100, 116, 139, 255)
	surf   = ui.NRGBA(245, 247, 250, 255)
	panel  = ui.NRGBA(255, 255, 255, 255)
	muted  = ui.NRGBA(226, 232, 240, 255)
)

func App(ctx *ui.Context) ui.Element {
	radVal := ui.UseState(ctx, float32(8))

	return ui.ContainerDecorationElement(
		ui.Bg(surf).WithPad(ui.All(16)),
		ui.ScrollViewElement(
			ui.ColumnElement(
				sectionTitle("Section 1 — 新旧 API 对比"),
				comparisonSection(),
				ui.SpacerElement(0, 16),

				sectionTitle("Section 2 — Decoration 效果变体"),
				varietySection(),
				ui.SpacerElement(0, 16),

				sectionTitle("Section 3 — Merge 组合"),
				mergeSection(),
				ui.SpacerElement(0, 16),

				sectionTitle("Section 4 — 响应式 Decoration"),
				responsiveSection(radVal),
				ui.SpacerElement(0, 24),

				sectionTitle("Section 5 — Margin 外边距"),
				marginSection(),
				ui.SpacerElement(0, 16),

				sectionTitle("Section 6 — Border 边框"),
				borderSection(),
				ui.SpacerElement(0, 16),

				sectionTitle("Section 7 — Container 新旧对比"),
				containerComparisonSection(),
				ui.SpacerElement(0, 24),

				sectionTitle("Section 8 — Opacity 不透明度"),
				opacitySection(),
				ui.SpacerElement(0, 16),

				sectionTitle("Section 9 — LinearGradient 渐变背景"),
				gradientSection(),
				ui.SpacerElement(0, 16),

				sectionTitle("Section 10 — Circle Clip 圆形裁切"),
				circleClipSection(),
				ui.SpacerElement(0, 24),

				sectionTitle("Section 11 — BoxShadow 盒阴影"),
				shadowSection(),
				ui.SpacerElement(0, 24),
			),
		),
	)
}

func sectionTitle(text string) ui.Element {
	return ui.PaddingElement(
		ui.Symmetric(0, 12),
		ui.TextElement(text, ui.TextSize(18), ui.TextColor(ui.NRGBA(15, 23, 42, 255))),
	)
}

func subtitle(text string) ui.Element {
	return ui.PaddingElement(
		ui.Insets{Top: 2, Bottom: 8},
		ui.TextElement(text, ui.TextSize(13), ui.TextColor(gray)),
	)
}

func smallTag(label string, bg color.NRGBA) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(bg).WithPad(ui.Symmetric(3, 8)).WithRad(4),
		ui.TextElement(label, ui.TextSize(10), ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
	)
}

// ─── Section 1 ───────────────────────────────────────────────

func comparisonSection() ui.Element {
	deco := ui.Bg(blue).WithPad(ui.All(12)).WithRad(8)

	return ui.ColumnElement(
		subtitle("同一种视觉配置，旧 API 每个组件要重复三行，新 API 一次定义、多处复用。"),

		ui.RowElement(
			smallTag("旧写法", ui.NRGBA(150, 150, 150, 255)),
			ui.SpacerElement(8, 0),
			smallTag("新写法", blue),
			ui.SpacerElement(8, 0),
			ui.TextElement("← 同一套 Decoration", ui.TextSize(12), ui.TextColor(gray)),
		),

		ui.RowElement(
			oldButton(),
			ui.SpacerElement(12, 0),
			ui.ButtonElement(
				ui.TextElement("统一按钮"),
				ui.ButtonDecoration(deco),
			),
		),

		ui.RowElement(
			oldInput(),
			ui.SpacerElement(12, 0),
			ui.TextFieldElement(
				"统一输入框",
				ui.InputDecoration(deco),
			),
		),

		ui.RowElement(
			oldCard(),
			ui.SpacerElement(12, 0),
			ui.CardElement(
				ui.PaddingElement(ui.All(8), ui.TextElement("统一卡片", ui.TextSize(14))),
				ui.CardDecoration(deco),
			),
		),
	)
}

func oldButton() ui.Element {
	return ui.ButtonElement(
		ui.TextElement("旧写法按钮"),
		ui.ButtonBackground(blue),
		ui.ButtonPadding(ui.All(12)),
		ui.ButtonRadius(8),
	)
}

func oldInput() ui.Element {
	return ui.TextFieldElement(
		"旧写法输入框",
		ui.InputBackground(blue),
		ui.InputPadding(ui.All(12)),
		ui.InputRadius(8),
	)
}

func oldCard() ui.Element {
	return ui.CardElement(
		ui.PaddingElement(ui.All(8), ui.TextElement("旧写法卡片", ui.TextSize(14))),
		ui.CardBackground(blue),
		ui.CardPadding(ui.All(12)),
		ui.CardRadius(8),
	)
}

// ─── Section 2 ───────────────────────────────────────────────

func varietySection() ui.Element {
	bold := ui.NRGBA(255, 255, 255, 255)

	return ui.ColumnElement(
		subtitle("同一组件用不同 Decoration 实现多种外观变体。"),

		ui.RowElement(
			smallTag("Button 变体", purple),
			ui.SpacerElement(8, 0),
			ui.ButtonElement(
				ui.TextElement("直角"),
				ui.ButtonDecoration(ui.Bg(blue).WithPad(ui.Symmetric(8, 14)).WithRad(0)),
			),
			ui.SpacerElement(8, 0),
			ui.ButtonElement(
				ui.TextElement("圆角"),
				ui.ButtonDecoration(ui.Bg(green).WithPad(ui.Symmetric(8, 14)).WithRad(8)),
			),
			ui.SpacerElement(8, 0),
			ui.ButtonElement(
				ui.TextElement("胶囊"),
				ui.ButtonDecoration(ui.Bg(purple).WithPad(ui.Symmetric(8, 20)).WithRad(20)),
			),
		),

		ui.RowElement(
			smallTag("Card 变体", orange),
			ui.SpacerElement(8, 0),
			ui.CardElement(
				ui.PaddingElement(ui.All(8), ui.TextElement("直角", ui.TextSize(13))),
				ui.CardDecoration(ui.Bg(panel).WithPad(ui.All(10)).WithRad(0)),
			),
			ui.SpacerElement(8, 0),
			ui.CardElement(
				ui.PaddingElement(ui.All(8),
					ui.ColumnElement(
						ui.TextElement("彩色", ui.TextSize(14), ui.TextColor(bold)),
						ui.TextElement("rad=12", ui.TextSize(11), ui.TextColor(ui.NRGBA(255, 255, 255, 180))),
					),
				),
				ui.CardDecoration(ui.Bg(orange).WithPad(ui.All(10)).WithRad(12)),
			),
			ui.SpacerElement(8, 0),
			ui.CardElement(
				ui.PaddingElement(ui.All(8), ui.TextElement("大圆角", ui.TextSize(13))),
				ui.CardDecoration(ui.Bg(panel).WithPad(ui.All(10)).WithRad(24)),
			),
		),

		ui.RowElement(
			smallTag("Input 变体", gray),
			ui.SpacerElement(8, 0),
			ui.TextFieldElement("直角输入",
				ui.InputDecoration(ui.Bg(panel).WithPad(ui.Symmetric(8, 12)).WithRad(0)),
			),
			ui.SpacerElement(8, 0),
			ui.TextFieldElement("圆角输入",
				ui.InputDecoration(ui.Bg(panel).WithPad(ui.Symmetric(8, 12)).WithRad(8)),
			),
			ui.SpacerElement(8, 0),
			ui.TextFieldElement("大圆角输入",
				ui.InputDecoration(ui.Bg(panel).WithPad(ui.Symmetric(8, 14)).WithRad(16)),
			),
		),
	)
}

// ─── Section 3 ───────────────────────────────────────────────

func mergeSection() ui.Element {
	base := ui.Bg(blue).WithPad(ui.All(12)).WithRad(8)
	bold := ui.NRGBA(255, 255, 255, 255)
	dark := ui.NRGBA(60, 60, 60, 255)

	return ui.ColumnElement(
		subtitle("定义 base 装饰，用 Merge 局部覆盖个别属性。"),

		ui.RowElement(
			smallTag("base", blue),
			ui.SpacerElement(8, 0),
			ui.ButtonElement(
				ui.TextElement("默认"),
				ui.ButtonDecoration(base),
			),
			ui.SpacerElement(12, 0),

			smallTag("Merge(Rad(0))", dark),
			ui.SpacerElement(8, 0),
			ui.ButtonElement(
				ui.TextElement("直角"),
				ui.ButtonDecoration(base.Merge(ui.Rad(0))),
			),
			ui.SpacerElement(12, 0),

			smallTag("Merge(Pad(20))", dark),
			ui.SpacerElement(8, 0),
			ui.ButtonElement(
				ui.TextElement("大间距"),
				ui.ButtonDecoration(base.Merge(ui.Pad(ui.All(20)))),
			),
		),

		ui.RowElement(
			smallTag("Card base", blue),
			ui.SpacerElement(8, 0),
			ui.CardElement(
				ui.PaddingElement(ui.All(8), ui.TextElement("base", ui.TextSize(13), ui.TextColor(bold))),
				ui.CardDecoration(base),
			),
			ui.SpacerElement(12, 0),

			smallTag("Merge(Bg(green))", dark),
			ui.SpacerElement(8, 0),
			ui.CardElement(
				ui.PaddingElement(ui.All(8), ui.TextElement("覆盖颜色", ui.TextSize(13), ui.TextColor(bold))),
				ui.CardDecoration(base.Merge(ui.Bg(green))),
			),
		),
	)
}

// ─── Section 4 ───────────────────────────────────────────────

func responsiveSection(radVal *state.State[float32]) ui.Element {
	return ui.ColumnElement(
		subtitle("UseState 驱动 Decoration，拖动滑块实时改变按钮圆角。"),

		ui.RowElement(
			ui.TextElement(fmt.Sprintf("圆角值: %.0f", radVal.Value()), ui.TextSize(14)),
		),

		ui.SliderElement(radVal.Value(),
			ui.SliderMin(0),
			ui.SliderMax(32),
			ui.SliderOnChange(func(ctx *ui.Context, v float32) {
				radVal.Set(v)
			}),
		),

		ui.RowElement(
			ui.ButtonElement(
				ui.TextElement(fmt.Sprintf("R=%.0f", radVal.Value())),
				ui.ButtonDecoration(ui.Bg(blue).WithPad(ui.Symmetric(8, 14)).WithRad(radVal.Value())),
			),
			ui.SpacerElement(12, 0),
			ui.ButtonElement(
				ui.TextElement("R=32"),
				ui.ButtonDecoration(ui.Bg(blue).WithPad(ui.Symmetric(8, 14)).WithRad(32)),
			),
		),

		ui.SpacerElement(0, 12),
	)
}

// ─── Section 5 ───────────────────────────────────────────────

func marginSection() ui.Element {
	return ui.ColumnElement(
		subtitle("通过 Decoration.Margin 控制外边距，支持 All / Only / LeftRight / TopBottom。"),

		ui.RowElement(
			smallTag("WithMargin(All(4))", blue),
			ui.SpacerElement(8, 0),
			smallTag("WithMargin(All(12))", green),
			ui.SpacerElement(8, 0),
			smallTag("WithMargin(All(20))", orange),
		),

		ui.ContainerDecorationElement(
			ui.Bg(muted).WithPad(ui.All(8)).WithRad(8),
			ui.RowElement(
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(12)).WithRad(6).WithMargin(ui.All(4)),
					ui.TextElement("M=4", ui.TextSize(12)),
				),
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(12)).WithRad(6).WithMargin(ui.All(12)),
					ui.TextElement("M=12", ui.TextSize(12)),
				),
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(12)).WithRad(6).WithMargin(ui.All(20)),
					ui.TextElement("M=20", ui.TextSize(12)),
				),
			),
		),

		ui.SpacerElement(0, 8),

		ui.RowElement(
			smallTag("ui.LeftRight(24)", blue),
			ui.SpacerElement(8, 0),
			smallTag("ui.TopBottom(12)", green),
			ui.SpacerElement(8, 0),
			smallTag("ui.Only(0,0,16,8)", orange),
		),

		ui.ContainerDecorationElement(
			ui.Bg(muted).WithPad(ui.All(8)).WithRad(8),
			ui.RowElement(
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(12)).WithRad(6).WithMargin(ui.LeftRight(24)),
					ui.TextElement("水平间距", ui.TextSize(12)),
				),
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(12)).WithRad(6).WithMargin(ui.TopBottom(12)),
					ui.TextElement("垂直间距", ui.TextSize(12)),
				),
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(12)).WithRad(6).WithMargin(ui.Only(0, 0, 16, 8)),
					ui.TextElement("偏置", ui.TextSize(12)),
				),
			),
		),
	)
}

// ─── Section 6 ───────────────────────────────────────────────

func borderSection() ui.Element {
	return ui.ColumnElement(
		subtitle("通过 ui.BorderDeco(width, color) 创建边框装饰，可独立使用或 Merge 到已有 Decoration。"),

		ui.RowElement(
			smallTag("无边框", gray),
			ui.SpacerElement(8, 0),
			smallTag("1px 灰边", gray),
			ui.SpacerElement(8, 0),
			smallTag("2px 蓝边", blue),
			ui.SpacerElement(8, 0),
			smallTag("4px 紫边", purple),
		),

		ui.ContainerDecorationElement(
			ui.Bg(muted).WithPad(ui.All(8)).WithRad(8),
			ui.RowElement(
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(12)).WithRad(6),
					ui.TextElement("无边框", ui.TextSize(12)),
				),
				ui.SpacerElement(12, 0),
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(12)).WithRad(6).Merge(ui.BorderDeco(1, gray)),
					ui.TextElement("细边框", ui.TextSize(12)),
				),
				ui.SpacerElement(12, 0),
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(12)).WithRad(6).Merge(ui.BorderDeco(2, blue)),
					ui.TextElement("蓝边框", ui.TextSize(12)),
				),
				ui.SpacerElement(12, 0),
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(12)).WithRad(6).Merge(ui.BorderDeco(4, purple)),
					ui.TextElement("粗边框", ui.TextSize(12)),
				),
			),
		),

		ui.SpacerElement(0, 8),

		ui.RowElement(
			smallTag("白边+深色底", purple),
			ui.SpacerElement(8, 0),
			smallTag("绿边+大圆角", green),
		),

		ui.ContainerDecorationElement(
			ui.Bg(muted).WithPad(ui.All(8)).WithRad(8),
			ui.RowElement(
				ui.ContainerDecorationElement(
					ui.Bg(blue).WithPad(ui.All(14)).WithRad(8).Merge(ui.BorderDeco(3, ui.NRGBA(255, 255, 255, 200))),
					ui.TextElement("白边深底", ui.TextSize(13), ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
				),
				ui.SpacerElement(12, 0),
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(14)).WithRad(20).Merge(ui.BorderDeco(3, green)),
					ui.TextElement("大圆角绿边", ui.TextSize(13)),
				),
			),
		),

		ui.SpacerElement(0, 8),

		ui.RowElement(
			smallTag("Card + BorderDeco", blue),
			ui.SpacerElement(8, 0),
			ui.TextElement("← Card 也可直接通过 Decoration.Border 加边框", ui.TextSize(12), ui.TextColor(gray)),
		),

		ui.ContainerDecorationElement(
			ui.Bg(muted).WithPad(ui.All(8)).WithRad(8),
			ui.RowElement(
				ui.CardElement(
					ui.PaddingElement(ui.All(10), ui.TextElement("有边框的卡片", ui.TextSize(13), ui.TextColor(blue))),
					ui.CardDecoration(ui.Bg(panel).WithPad(ui.All(10)).WithRad(8).Merge(ui.BorderDeco(2, blue))),
				),
				ui.SpacerElement(12, 0),
				ui.CardElement(
					ui.PaddingElement(ui.All(10),
						ui.ColumnElement(
							ui.TextElement("彩色边框卡片", ui.TextSize(14), ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
							ui.TextElement("Card + BorderDeco", ui.TextSize(11), ui.TextColor(ui.NRGBA(255, 255, 255, 180))),
						),
					),
					ui.CardDecoration(ui.Bg(purple).WithPad(ui.All(10)).WithRad(12).Merge(ui.BorderDeco(3, ui.NRGBA(255, 255, 255, 180)))),
				),
			),
		),
	)
}

// ─── Section 7 ───────────────────────────────────────────────

func containerComparisonSection() ui.Element {
	return ui.ColumnElement(
		subtitle("旧 Container(Style{...}) → 新 ContainerDecoration(Bg(...).Pad(...).Rad(...))，新增 Margin/Border 支持。"),

		ui.RowElement(
			smallTag("旧 API: Container()", gray),
			ui.SpacerElement(8, 0),
			smallTag("新 API: ContainerDecoration()", blue),
		),

		ui.RowElement(
			ui.ContainerDecorationElement(
				ui.Bg(ui.NRGBA(230, 230, 230, 255)).WithPad(ui.All(4)).WithRad(6),
				ui.ColumnElement(
					ui.TextElement("旧 Container", ui.TextSize(11), ui.TextColor(gray)),
					ui.ContainerElement(
						ui.Style{Background: panel, Padding: ui.All(12), Radius: 8},
						ui.TextElement("Content", ui.TextSize(13)),
					),
				),
			),
			ui.SpacerElement(12, 0),
			ui.ContainerDecorationElement(
				ui.Bg(ui.NRGBA(33, 133, 209, 40)).WithPad(ui.All(4)).WithRad(6),
				ui.ColumnElement(
					ui.TextElement("新 ContainerDecoration", ui.TextSize(11), ui.TextColor(blue)),
					ui.ContainerDecorationElement(
						ui.Bg(panel).WithPad(ui.All(12)).WithRad(8),
						ui.TextElement("Content", ui.TextSize(13)),
					),
				),
			),
		),

		ui.SpacerElement(0, 8),

		subtitle("新 API 额外支持 Margin + Border，一行链式组合："),

		ui.ContainerDecorationElement(
			ui.Bg(ui.NRGBA(33, 133, 209, 25)).WithPad(ui.All(4)).WithRad(8),
			ui.RowElement(
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(12)).WithRad(8),
					ui.TextElement("无扩展", ui.TextSize(12)),
				),
				ui.SpacerElement(12, 0),
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(12)).WithRad(8).WithMargin(ui.Symmetric(0, 12)).Merge(ui.BorderDeco(2, blue)),
					ui.TextElement("+Margin +Border", ui.TextSize(12)),
				),
			),
		),
	)
}

// ─── Section 8 ───────────────────────────────────────────────

func opacitySection() ui.Element {
	return ui.ColumnElement(
		subtitle("通过 Decoration.Opacity 控制容器整体透明度，实现毛玻璃叠加效果。"),

		ui.RowElement(
			smallTag("Opacity(1.0)", blue),
			ui.SpacerElement(8, 0),
			smallTag("Opacity(0.7)", blue),
			ui.SpacerElement(8, 0),
			smallTag("Opacity(0.4)", blue),
			ui.SpacerElement(8, 0),
			smallTag("Opacity(0.15)", blue),
		),

		ui.ContainerDecorationElement(
			ui.Bg(muted).WithPad(ui.All(12)).WithRad(8),
			ui.RowElement(
				ui.ContainerDecorationElement(
					ui.Bg(blue).WithPad(ui.All(14)).WithRad(8),
					ui.TextElement("1.0", ui.TextSize(13), ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
				),
				ui.SpacerElement(10, 0),
				ui.ContainerDecorationElement(
					ui.Bg(blue).WithPad(ui.All(14)).WithRad(8).WithOpacity(0.7),
					ui.TextElement("0.7", ui.TextSize(13), ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
				),
				ui.SpacerElement(10, 0),
				ui.ContainerDecorationElement(
					ui.Bg(blue).WithPad(ui.All(14)).WithRad(8).WithOpacity(0.4),
					ui.TextElement("0.4", ui.TextSize(13), ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
				),
				ui.SpacerElement(10, 0),
				ui.ContainerDecorationElement(
					ui.Bg(blue).WithPad(ui.All(14)).WithRad(8).WithOpacity(0.15),
					ui.TextElement("0.15", ui.TextSize(13), ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
				),
			),
		),

		ui.SpacerElement(0, 8),

		subtitle("Opacity 叠加在 Border + 基础色之上，创造半透明面板："),

		ui.ContainerDecorationElement(
			ui.Bg(muted).WithPad(ui.All(12)).WithRad(8),
			ui.RowElement(
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(14)).WithRad(8).Merge(ui.BorderDeco(2, blue)).WithOpacity(0.5),
					ui.TextElement("半透明白卡", ui.TextSize(13)),
				),
				ui.SpacerElement(12, 0),
				ui.ContainerDecorationElement(
					ui.Bg(purple).WithPad(ui.All(14)).WithRad(8).WithOpacity(0.6),
					ui.TextElement("半透明紫底", ui.TextSize(13), ui.TextColor(ui.NRGBA(255, 255, 255, 200))),
				),
			),
		),
	)
}

// ─── Section 9 ───────────────────────────────────────────────

func gradientSection() ui.Element {
	bold := ui.NRGBA(255, 255, 255, 255)
	cardW := 140
	cardH := 80

	return ui.ColumnElement(
		subtitle("通过 ContainerDecoration + LinearGradient 实现渐变背景。Gradient 设置后 Background 忽略。"),

		ui.RowElement(
			smallTag("垂直渐变 ↓", blue),
			ui.SpacerElement(8, 0),
			smallTag("水平渐变 →", green),
			ui.SpacerElement(8, 0),
			smallTag("对角 ↘", orange),
			ui.SpacerElement(8, 0),
			smallTag("对角线 ↗", purple),
		),

		ui.ContainerDecorationElement(
			ui.Bg(muted).WithPad(ui.All(8)).WithRad(8),
			ui.RowElement(
				ui.ContainerDecorationElement(
					ui.Decoration{}.
						WithGradient(ui.LinearGradient{
							Start: image.Pt(0, 0),
							End:   image.Pt(0, cardH),
							From:  ui.NRGBA(33, 133, 209, 255),
							To:    ui.NRGBA(15, 60, 120, 255),
						}).
						WithPad(ui.All(14)).WithRad(8),
					ui.ContainerDecorationElement(
						ui.Bg(panel).WithPad(ui.Symmetric(4, 10)).WithRad(4).WithOpacity(0.15),
						ui.ColumnElement(
							ui.TextElement("Top→Bottom", ui.TextSize(12), ui.TextColor(bold)),
							ui.TextElement("蓝→深蓝", ui.TextSize(11), ui.TextColor(ui.NRGBA(255, 255, 255, 180))),
						),
					),
				),
				ui.SpacerElement(10, 0),
				ui.ContainerDecorationElement(
					ui.Decoration{}.
						WithGradient(ui.LinearGradient{
							Start: image.Pt(0, 0),
							End:   image.Pt(cardW, 0),
							From:  ui.NRGBA(40, 167, 69, 255),
							To:    ui.NRGBA(20, 120, 40, 255),
						}).
						WithPad(ui.All(14)).WithRad(8),
					ui.ContainerDecorationElement(
						ui.Bg(panel).WithPad(ui.Symmetric(4, 10)).WithRad(4).WithOpacity(0.15),
						ui.ColumnElement(
							ui.TextElement("Left→Right", ui.TextSize(12), ui.TextColor(bold)),
							ui.TextElement("绿→深绿", ui.TextSize(11), ui.TextColor(ui.NRGBA(255, 255, 255, 180))),
						),
					),
				),
				ui.SpacerElement(10, 0),
				ui.ContainerDecorationElement(
					ui.Decoration{}.
						WithGradient(ui.LinearGradient{
							Start: image.Pt(0, 0),
							End:   image.Pt(cardW, cardH),
							From:  ui.NRGBA(255, 193, 7, 255),
							To:    ui.NRGBA(255, 120, 0, 255),
						}).
						WithPad(ui.All(14)).WithRad(8),
					ui.ContainerDecorationElement(
						ui.Bg(panel).WithPad(ui.Symmetric(4, 10)).WithRad(4).WithOpacity(0.15),
						ui.ColumnElement(
							ui.TextElement("TL→BR", ui.TextSize(12), ui.TextColor(bold)),
							ui.TextElement("黄→橙", ui.TextSize(11), ui.TextColor(ui.NRGBA(255, 255, 255, 180))),
						),
					),
				),
				ui.SpacerElement(10, 0),
				ui.ContainerDecorationElement(
					ui.Decoration{}.
						WithGradient(ui.LinearGradient{
							Start: image.Pt(cardW, 0),
							End:   image.Pt(0, cardH),
							From:  ui.NRGBA(155, 89, 182, 255),
							To:    ui.NRGBA(80, 40, 120, 255),
						}).
						WithPad(ui.All(14)).WithRad(8),
					ui.ContainerDecorationElement(
						ui.Bg(panel).WithPad(ui.Symmetric(4, 10)).WithRad(4).WithOpacity(0.15),
						ui.ColumnElement(
							ui.TextElement("TR→BL", ui.TextSize(12), ui.TextColor(bold)),
							ui.TextElement("紫→深紫", ui.TextSize(11), ui.TextColor(ui.NRGBA(255, 255, 255, 180))),
						),
					),
				),
			),
		),

		ui.SpacerElement(0, 8),

		subtitle("渐变 + 边框 + 圆角的组合："),

		ui.ContainerDecorationElement(
			ui.Bg(muted).WithPad(ui.All(8)).WithRad(8),
			ui.RowElement(
				ui.ContainerDecorationElement(
					ui.Decoration{}.
						WithGradient(ui.LinearGradient{
							Start: image.Pt(0, 0), End: image.Pt(0, 60),
							From: blue, To: green,
						}).
						WithPad(ui.All(14)).WithRad(12).Merge(ui.BorderDeco(2, ui.NRGBA(255, 255, 255, 180))),
					ui.TextElement("渐变+边框+圆角", ui.TextSize(13), ui.TextColor(bold)),
				),
			),
		),
	)
}

// ─── Section 10 ───────────────────────────────────────────────

func circleClipSection() ui.Element {
	bold := ui.NRGBA(255, 255, 255, 255)

	return ui.ColumnElement(
		subtitle("通过 Decoration.CircleClip 启用圆形裁切，适用于头像、徽章、FAB 按钮。"),

		ui.RowElement(
			smallTag("纯色圆形", blue),
			ui.SpacerElement(8, 0),
			smallTag("渐变圆形", green),
			ui.SpacerElement(8, 0),
			smallTag("圆形+边框", purple),
			ui.SpacerElement(8, 0),
			smallTag("渐变+边框", orange),
		),

		ui.ContainerDecorationElement(
			ui.Bg(muted).WithPad(ui.All(12)).WithRad(8),
			ui.RowElement(
				ui.ContainerDecorationElement(
					ui.Bg(blue).WithPad(ui.All(22)).WithCircleClip(),
					ui.TextElement("A", ui.TextSize(18), ui.TextColor(bold)),
				),
				ui.SpacerElement(12, 0),
				ui.ContainerDecorationElement(
					ui.Decoration{}.
						WithGradient(ui.LinearGradient{
							Start: image.Pt(0, 0),
							End:   image.Pt(0, 60),
							From:  green,
							To:    ui.NRGBA(20, 100, 40, 255),
						}).
						WithPad(ui.All(22)).WithCircleClip(),
					ui.TextElement("G", ui.TextSize(18), ui.TextColor(bold)),
				),
				ui.SpacerElement(12, 0),
				ui.ContainerDecorationElement(
					ui.Bg(purple).WithPad(ui.All(22)).Merge(ui.BorderDeco(3, ui.NRGBA(255, 255, 255, 200))).WithCircleClip(),
					ui.TextElement("P", ui.TextSize(18), ui.TextColor(bold)),
				),
				ui.SpacerElement(12, 0),
				ui.ContainerDecorationElement(
					ui.Decoration{}.
						WithGradient(ui.LinearGradient{
							Start: image.Pt(0, 0),
							End:   image.Pt(60, 60),
							From:  orange,
							To:    ui.NRGBA(255, 80, 0, 255),
						}).
						WithPad(ui.All(22)).Merge(ui.BorderDeco(3, ui.NRGBA(255, 255, 255, 200))).WithCircleClip(),
					ui.TextElement("O", ui.TextSize(18), ui.TextColor(bold)),
				),
			),
		),

		ui.SpacerElement(0, 8),

		subtitle("半透明圆形，叠加在有底色的背景上："),

		ui.ContainerDecorationElement(
			ui.Bg(muted).WithPad(ui.All(12)).WithRad(8),
			ui.RowElement(
				ui.ContainerDecorationElement(
					ui.Bg(blue).WithPad(ui.All(22)).WithOpacity(0.6).WithCircleClip(),
					ui.TextElement("60%", ui.TextSize(13), ui.TextColor(bold)),
				),
				ui.SpacerElement(12, 0),
				ui.ContainerDecorationElement(
					ui.Bg(green).WithPad(ui.All(26)).WithOpacity(0.5).WithCircleClip(),
					ui.TextElement("50%", ui.TextSize(14), ui.TextColor(bold)),
				),
				ui.SpacerElement(12, 0),
				ui.ContainerDecorationElement(
					ui.Bg(purple).WithPad(ui.All(30)).WithOpacity(0.35).WithCircleClip(),
					ui.TextElement("35%", ui.TextSize(16), ui.TextColor(bold)),
				),
			),
		),
	)
}

// ─── Section 11 ───────────────────────────────────────────────

func shadowSection() ui.Element {
	return ui.ColumnElement(
		subtitle("通过 BoxShadow 实现 Material Design 高度阴影。支持自定义偏移/模糊/颜色或 Elevation 预设。"),

		ui.RowElement(
			smallTag("Elevation(1)", blue),
			ui.SpacerElement(8, 0),
			smallTag("Elevation(2)", blue),
			ui.SpacerElement(8, 0),
			smallTag("Elevation(3)", blue),
			ui.SpacerElement(8, 0),
			smallTag("Elevation(4)", blue),
			ui.SpacerElement(8, 0),
			smallTag("Elevation(5)", blue),
		),

		ui.ContainerDecorationElement(
			ui.Bg(muted).WithPad(ui.All(16)).WithRad(8),
			ui.RowElement(
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(14)).WithRad(8).Merge(ui.Elevation(1)),
					ui.TextElement("Level 1", ui.TextSize(12)),
				),
				ui.SpacerElement(14, 0),
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(14)).WithRad(8).Merge(ui.Elevation(2)),
					ui.TextElement("Level 2", ui.TextSize(12)),
				),
				ui.SpacerElement(14, 0),
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(14)).WithRad(8).Merge(ui.Elevation(3)),
					ui.TextElement("Level 3", ui.TextSize(12)),
				),
				ui.SpacerElement(14, 0),
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(14)).WithRad(8).Merge(ui.Elevation(4)),
					ui.TextElement("Level 4", ui.TextSize(12)),
				),
				ui.SpacerElement(14, 0),
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(14)).WithRad(8).Merge(ui.Elevation(5)),
					ui.TextElement("Level 5", ui.TextSize(12)),
				),
			),
		),

		ui.SpacerElement(0, 8),

		subtitle("自定义 Shadow 偏移/模糊/颜色："),

		ui.RowElement(
			smallTag("↑6 Blur10", blue),
			ui.SpacerElement(8, 0),
			smallTag("→8 Blur16", green),
			ui.SpacerElement(8, 0),
			smallTag("↘10 Blur20", orange),
			ui.SpacerElement(8, 0),
			smallTag("←8 Blur8 蓝色", purple),
		),

		ui.ContainerDecorationElement(
			ui.Bg(muted).WithPad(ui.All(16)).WithRad(8),
			ui.RowElement(
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(14)).WithRad(8).Merge(
						ui.Shadow(0, 6, 10, ui.NRGBA(0, 0, 0, 60))),
					ui.TextElement("底部阴影", ui.TextSize(12)),
				),
				ui.SpacerElement(14, 0),
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(14)).WithRad(8).Merge(
						ui.Shadow(8, 0, 16, ui.NRGBA(0, 0, 0, 55))),
					ui.TextElement("右向阴影", ui.TextSize(12)),
				),
				ui.SpacerElement(14, 0),
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(14)).WithRad(8).Merge(
						ui.Shadow(10, 10, 20, ui.NRGBA(0, 0, 0, 60))),
					ui.TextElement("右下阴影", ui.TextSize(12)),
				),
				ui.SpacerElement(14, 0),
				ui.ContainerDecorationElement(
					ui.Bg(panel).WithPad(ui.All(14)).WithRad(8).Merge(
						ui.Shadow(-8, 0, 8, ui.NRGBA(40, 100, 200, 70))),
					ui.TextElement("彩色阴影", ui.TextSize(12)),
				),
			),
		),

		ui.SpacerElement(0, 8),

		subtitle("阴影 + 渐变 + 圆形裁切的组合："),

		ui.ContainerDecorationElement(
			ui.Bg(muted).WithPad(ui.All(16)).WithRad(8),
			ui.RowElement(
				ui.ContainerDecorationElement(
					ui.Decoration{}.
						WithGradient(ui.LinearGradient{
							Start: image.Pt(0, 0),
							End:   image.Pt(0, 60),
							From:  blue,
							To:    ui.NRGBA(15, 60, 120, 255),
						}).
						WithPad(ui.All(14)).WithRad(12).
						Merge(ui.BorderDeco(2, ui.NRGBA(255, 255, 255, 180))).
						Merge(ui.Elevation(3)),
					ui.TextElement("渐变+边框+阴影", ui.TextSize(12), ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
				),
				ui.SpacerElement(14, 0),
				ui.ContainerDecorationElement(
					ui.Bg(green).WithPad(ui.All(22)).WithCircleClip().Merge(ui.Elevation(3)),
					ui.TextElement("⚪", ui.TextSize(18), ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
				),
				ui.SpacerElement(14, 0),
				ui.ContainerDecorationElement(
					ui.Decoration{}.
						WithGradient(ui.LinearGradient{
							Start: image.Pt(0, 0),
							End:   image.Pt(60, 60),
							From:  orange,
							To:    ui.NRGBA(255, 80, 0, 255),
						}).
						WithPad(ui.All(22)).WithCircleClip().
						Merge(ui.BorderDeco(3, ui.NRGBA(255, 255, 255, 200))).
						Merge(ui.Elevation(3)),
					ui.TextElement("◉", ui.TextSize(18), ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
				),
			),
		),

		ui.SpacerElement(0, 8),

		subtitle("CardShadow（旧 API）复活 — 通过 Decoration.Shadow 底层统一渲染："),

		ui.ContainerDecorationElement(
			ui.Bg(muted).WithPad(ui.All(16)).WithRad(8),
			ui.RowElement(
				ui.CardElement(
					ui.PaddingElement(ui.All(10),
						ui.TextElement("CardShadow(3)", ui.TextSize(13)),
					),
					ui.CardDecoration(ui.Bg(panel).WithPad(ui.All(10)).WithRad(8)),
					ui.CardShadow(3),
				),
				ui.SpacerElement(14, 0),
				ui.CardElement(
					ui.PaddingElement(ui.All(10),
						ui.ColumnElement(
							ui.TextElement("CardShadow(5)", ui.TextSize(14), ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
						),
					),
					ui.CardDecoration(ui.Bg(purple).WithPad(ui.All(10)).WithRad(12)),
					ui.CardShadow(5),
				),
			),
		),
	)
}
