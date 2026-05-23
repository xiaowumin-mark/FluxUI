package main

import (
	"fmt"
	"image/color"
	"time"

	anim "github.com/xiaowumin-mark/FluxUI/anim"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.RunElement(App, ui.Title("Animation Showcase"), ui.Size(750, 620))
}

func App(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	tab := ui.UseState(ctx, "easing")

	items := []ui.TabItem{
		{Key: "easing", Label: "缓动对比"},
		{Key: "animValue", Label: "值动画"},
		{Key: "animDeco", Label: "装饰动画"},
		{Key: "hover", Label: "悬停动画"},
		{Key: "pulse", Label: "脉冲动画"},
	}

	var content ui.Element
	switch tab.Value() {
	case "animValue":
		content = ui.ComponentElement(animValueSection)
	case "animDeco":
		content = ui.ComponentElement(animDecoSection)
	case "hover":
		content = ui.ComponentElement(hoverSection)
	case "pulse":
		content = ui.ComponentElement(pulseSection)
	default:
		content = ui.ComponentElement(easingSection)
	}

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(16)),
		ui.ColumnElement(
			ui.TextElement("Animation Showcase", ui.TextSize(22)),
			ui.VSpacerElement(8),
			ui.TabsElement(tab.Value(), items, ui.TabsOnChange(func(ctx *ui.Context, key string) { tab.Set(key) }), ui.TabsScrollable(true)),
			ui.VSpacerElement(12),
			content,
		),
	)
}

func sectionTitle(title, desc string) ui.Element {
	return ui.ColumnElement(
		ui.TextElement(title, ui.TextSize(16)),
		ui.VSpacerElement(2),
		ui.TextElement(desc, ui.TextSize(12), ui.TextColor(color.NRGBA{R: 100, G: 116, B: 139, A: 255})),
	)
}

func blue() color.NRGBA   { return color.NRGBA{R: 59, G: 130, B: 246, A: 255} }
func red() color.NRGBA    { return color.NRGBA{R: 220, G: 38, B: 38, A: 255} }
func green() color.NRGBA  { return color.NRGBA{R: 22, G: 163, B: 74, A: 255} }
func orange() color.NRGBA { return color.NRGBA{R: 234, G: 88, B: 12, A: 255} }
func white() color.NRGBA  { return color.NRGBA{R: 255, G: 255, B: 255, A: 255} }
func gray() color.NRGBA   { return color.NRGBA{R: 226, G: 232, B: 240, A: 255} }

var barColors = []color.NRGBA{blue(), orange(), green(), color.NRGBA{R: 168, G: 85, B: 247, A: 255}, red(), color.NRGBA{R: 236, G: 72, B: 153, A: 255}}

var easingNames = []string{"Linear", "EaseOut", "EaseInOut", "EaseOutBack", "EaseOutBounce", "EaseOutElastic"}
var easingFuncs = []anim.Easing{anim.Linear, anim.EaseOut, anim.EaseInOut, anim.EaseOutBack, anim.EaseOutBounce, anim.EaseOutElastic}

func easingSection(ctx *ui.Context) ui.Element {
	playing := ui.UseState(ctx, false)
	target := float32(0)
	if playing.Value() {
		target = 1
	}

	dur := 800 * time.Millisecond

	progresses := make([]float32, len(easingFuncs))
	children := make([]ui.Element, 0, len(easingFuncs)+4)

	for i, fn := range easingFuncs {
		easing := fn
		progresses[i] = ui.UseAnimatedValue(ctx, target, dur, easing)
	}

	children = append(children,
		sectionTitle("缓动函数对比", "6 种缓动曲线同步播放，观察运动差异"),
		ui.VSpacerElement(8),
		ui.RowElement(
			ui.ButtonElement(ui.TextElement("播放"), ui.OnClick(func(c *ui.Context) { playing.Set(true) })),
			ui.HSpacerElement(8),
			ui.ButtonElement(ui.TextElement("重置"), ui.OnClick(func(c *ui.Context) { playing.Set(false) })),
		),
		ui.VSpacerElement(12),
	)

	for i := range easingFuncs {
		progress := progresses[i]
		barColor := barColors[i%len(barColors)]
		label := easingNames[i]

		barWidth := progress * 400
		if barWidth < 2 {
			barWidth = 2
		}

		children = append(children,
			ui.TextElement(fmt.Sprintf("%s: %.0f%%", label, progress*100), ui.TextSize(12)),
			ui.VSpacerElement(2),
			ui.FixedSizeElement(400, 12,
				ui.StackElement(
					ui.ContainerDecorationElement(ui.Bg(gray()).WithRad(4), ui.SpacerElement(400, 12)),
					ui.FixedWidthElement(barWidth,
						ui.ContainerDecorationElement(ui.Bg(barColor).WithRad(4), ui.SpacerElement(0, 12)),
					),
				),
			),
			ui.VSpacerElement(6),
		)
	}

	return ui.ColumnElement(children...)
}

func animValueSection(ctx *ui.Context) ui.Element {
	targetSize := ui.UseState(ctx, float32(100))
	easingIdx := ui.UseState(ctx, 0)

	idx := easingIdx.Value() % len(easingFuncs)
	easingChoice := easingFuncs[idx]
	easingName := easingNames[idx]

	dur := 600 * time.Millisecond
	animated := ui.UseAnimatedValue(ctx, targetSize.Value(), dur, easingChoice)

	return ui.ColumnElement(
		sectionTitle("UseAnimatedValue", fmt.Sprintf("泛型值动画 — 当前缓动: %s, 目标宽高: %.0f", easingName, targetSize.Value())),
		ui.VSpacerElement(12),
		ui.RowElement(
			ui.ButtonElement(ui.TextElement("100"), ui.OnClick(func(c *ui.Context) { targetSize.Set(100) })),
			ui.HSpacerElement(6),
			ui.ButtonElement(ui.TextElement("200"), ui.OnClick(func(c *ui.Context) { targetSize.Set(200) })),
			ui.HSpacerElement(6),
			ui.ButtonElement(ui.TextElement("300"), ui.OnClick(func(c *ui.Context) { targetSize.Set(300) })),
			ui.HSpacerElement(6),
			ui.ButtonElement(ui.TextElement("换缓动"), ui.OnClick(func(c *ui.Context) { easingIdx.Set(easingIdx.Value() + 1) })),
		),
		ui.VSpacerElement(16),
		ui.FixedHeightElement(340,
			ui.FillWidthElement(
				ui.ContainerDecorationElement(
					ui.Bg(blue()).WithRad(8).Merge(ui.Elevation(2)),
					ui.CenterElement(
						ui.FixedSizeElement(animated, animated,
							ui.ContainerDecorationElement(
								ui.Bg(orange()).WithRad(8),
								ui.CenterElement(
									ui.ColumnElement(
										ui.TextElement(fmt.Sprintf("%.0fpx", animated), ui.TextSize(16), ui.TextColor(white())),
										ui.VSpacerElement(4),
										ui.TextElement(easingName, ui.TextSize(11), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 180})),
									),
								),
							),
						),
					),
				),
			),
		),
	)
}

func animDecoSection(ctx *ui.Context) ui.Element {
	state := ui.UseState(ctx, 0)

	var targetDeco ui.Decoration
	if state.Value()%2 == 0 {
		targetDeco = ui.Bg(blue()).WithRad(8).WithOpacity(1.0).
			Merge(ui.TransformDeco(0, 1.0, 1.0, 0, 0, ui.TransformCenter)).
			Merge(ui.Elevation(1))
	} else {
		targetDeco = ui.Bg(orange()).WithRad(40).WithOpacity(0.8).
			Merge(ui.TransformDeco(5, 1.1, 1.1, 0, 0, ui.TransformCenter)).
			Merge(ui.Elevation(4))
	}

	dur := 600 * time.Millisecond
	animDeco := ui.UseAnimatedDecoration(ctx, targetDeco, dur, anim.EaseInOutCubic)

	return ui.ColumnElement(
		sectionTitle("UseAnimatedDecoration", "点击卡片同时动画背景色、圆角、缩放、透明度、阴影"),
		ui.VSpacerElement(12),
		ui.CenterElement(
			ui.ContainerDecorationElement(
				animDeco.WithPad(ui.Symmetric(32, 48)),
				ui.ColumnElement(
					ui.TextElement("点击切换", ui.TextSize(16), ui.TextColor(white()), ui.TextAlign(ui.AlignCenter)),
					ui.VSpacerElement(6),
					ui.TextElement(fmt.Sprintf("第 %d 次", state.Value()+1), ui.TextSize(13), ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 180}), ui.TextAlign(ui.AlignCenter)),
				),
				ui.OnDecoClick(func(c *ui.Context) { state.Set(state.Value() + 1) }),
			),
		),
	)
}

func hoverSection(ctx *ui.Context) ui.Element {
	hovered := ui.UseState(ctx, false)

	var targetDeco ui.Decoration
	if hovered.Value() {
		targetDeco = ui.Bg(orange()).WithRad(16).Merge(ui.Elevation(4)).
			Merge(ui.TransformDeco(2, 1.08, 1.08, 0, 0, ui.TransformCenter))
	} else {
		targetDeco = ui.Bg(blue()).WithRad(12).Merge(ui.Elevation(1)).
			Merge(ui.TransformDeco(0, 1, 1, 0, 0, ui.TransformCenter))
	}

	dur := 300 * time.Millisecond
	animDeco := ui.UseAnimatedDecoration(ctx, targetDeco, dur, anim.CubicBezier(0.25, 0.1, 0.25, 1.0))

	return ui.ColumnElement(
		sectionTitle("悬停动画", "鼠标悬停时用 CubicBezier 缓动平滑过渡背景色、阴影、缩放"),
		ui.VSpacerElement(12),
		ui.CenterElement(
			ui.ContainerDecorationElement(
				animDeco.WithPad(ui.Symmetric(32, 48)),
				ui.TextElement("悬停我", ui.TextSize(16), ui.TextColor(white()), ui.TextAlign(ui.AlignCenter)),
				ui.OnDecoHoverEnter(func(c *ui.Context) { hovered.Set(true) }),
				ui.OnDecoHoverLeave(func(c *ui.Context) { hovered.Set(false) }),
			),
		),
	)
}

func pulseSection(ctx *ui.Context) ui.Element {
	active := ui.UseState(ctx, false)

	target := float32(0)
	if active.Value() {
		target = 1
	}

	dur := 800 * time.Millisecond
	pulse := ui.UseAnimatedValue(ctx, target, dur, anim.EaseInOutSine)

	if active.Value() && pulse >= 0.99 {
		active.Set(false)
	} else if !active.Value() && pulse <= 0.01 {
		active.Set(true)
	}

	size := pulse*80 + 40
	radius := size / 2
	alpha := uint8(pulse*155 + 100)

	return ui.ColumnElement(
		sectionTitle("脉冲动画", "自动循环 0→1→0，大小和透明度联动"),
		ui.VSpacerElement(12),
		ui.ButtonElement(ui.TextElement("开始/停止"), ui.OnClick(func(c *ui.Context) { active.Set(!active.Value()) })),
		ui.VSpacerElement(20),
		ui.CenterElement(
			ui.FixedSizeElement(size, size,
				ui.ContainerDecorationElement(
					ui.Bg(color.NRGBA{R: 59, G: 130, B: 246, A: alpha}).WithRad(radius).WithBorder(ui.Border{Width: 2, Color: color.NRGBA{R: 59, G: 130, B: 246, A: uint8(pulse*100 + 20)}}),
					ui.CenterElement(
						ui.TextElement(fmt.Sprintf("%.0f", size), ui.TextSize(14), ui.TextColor(white()), ui.TextAlign(ui.AlignCenter)),
					),
				),
			),
		),
	)
}
