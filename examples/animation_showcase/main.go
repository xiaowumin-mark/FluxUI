package main

import (
	"fmt"
	"image/color"
	"time"

	"github.com/xiaowumin-mark/FluxUI/anim"
	"github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)
		page := ui.State[string](ctx)

		if page.Value() == "" {
			page.Set("easing")
		}

		tabs := ui.Tabs(
			page.Value(),
			[]ui.TabItem{
				{Key: "easing", Label: "缓动对比"},
				{Key: "pulse", Label: "脉冲呼吸"},
				{Key: "stagger", Label: "交错入场"},
				{Key: "color", Label: "颜色渐变"},
				{Key: "progress", Label: "进度动画"},
			},
			ui.TabsOnChange(func(ctx *ui.Context, key string) {
				page.Set(key)
			}),
			ui.TabsScrollable(true),
		)

		var content ui.Widget
		switch page.Value() {
		case "easing":
			content = easingCompare(ctx.Scope("easing"), th)
		case "pulse":
			content = pulseBreathing(ctx.Scope("pulse"), th)
		case "stagger":
			content = staggerEntrance(ctx.Scope("stagger"), th)
		case "color":
			content = colorTransition(ctx.Scope("color"), th)
		case "progress":
			content = progressAnimation(ctx.Scope("progress"), th)
		default:
			content = ui.Text("选择一个标签页")
		}

		return ui.Container(
			ui.Style{Background: th.Surface, Padding: ui.All(16)},
			ui.Column(
				ui.Text("Animation Showcase", ui.TextSize(22)),
				ui.VSpacer(4),
				ui.Text("全面展示 FluxUI 动画系统的各种用法", ui.TextSize(13), ui.TextColor(th.SurfaceMuted)),
				ui.VSpacer(12),
				tabs,
				ui.VSpacer(12),
				content,
			),
		)
	}, ui.Title("Animation Showcase"), ui.Size(700, 560))
}

// ========== 1. 缓动函数对比 ==========

func easingCompare(ctx *ui.Context, th *ui.Theme) ui.Widget {
	// 0=idle, 1=forward, -1=reverse
	direction := ui.State[int](ctx)

	// 始终调用 3 个动画，用 from/to 控制状态
	var from, to float32
	switch direction.Value() {
	case 1:
		from, to = 0, 1
	case -1:
		from, to = 1, 0
	default:
		from, to = 0, 0 // idle: value 始终为 0
	}

	dur := 1200 * time.Millisecond
	linearP := anim.New(anim.From(from), anim.To(to), anim.Duration(dur), anim.Ease(anim.Linear)).Value(ctx)
	easeOutP := anim.New(anim.From(from), anim.To(to), anim.Duration(dur), anim.Ease(anim.EaseOut)).Value(ctx)
	easeInOutP := anim.New(anim.From(from), anim.To(to), anim.Duration(dur), anim.Ease(anim.EaseInOut)).Value(ctx)

	return ui.Column(
		ui.Text("三种缓动函数同步播放，观察运动曲线差异", ui.TextSize(13), ui.TextColor(th.SurfaceMuted)),
		ui.VSpacer(12),
		ui.Row(
			ui.Button(ui.Text("播放"), ui.OnClick(func(ctx *ui.Context) {
				direction.Set(1)
			})),
			ui.HSpacer(8),
			ui.Button(ui.Text("反向"), ui.OnClick(func(ctx *ui.Context) {
				direction.Set(-1)
			})),
		),
		ui.VSpacer(16),
		easingBar("Linear", linearP, th.Primary),
		ui.VSpacer(8),
		easingBar("EaseOut", easeOutP, color.NRGBA{R: 234, G: 88, B: 12, A: 255}),
		ui.VSpacer(8),
		easingBar("EaseInOut", easeInOutP, color.NRGBA{R: 22, G: 163, B: 74, A: 255}),
	)
}

func easingBar(label string, progress float32, barColor color.NRGBA) ui.Widget {
	width := lerpFloat(0, 400, progress)
	if width < 1 {
		width = 1
	}
	return ui.Column(
		ui.Text(fmt.Sprintf("%s: %.0f%%", label, progress*100), ui.TextSize(12)),
		ui.VSpacer(2),
		ui.FixedSize(400, 12,
			ui.Stack(
				ui.Container(
					ui.Style{Background: color.NRGBA{R: 230, G: 230, B: 230, A: 255}, Radius: 4},
					ui.Spacer(400, 12),
				),
				ui.FixedWidth(width,
					ui.Container(
						ui.Style{Background: barColor, Radius: 4},
						ui.Spacer(0, 12),
					),
				),
			),
		),
	)
}

// ========== 2. 脉冲呼吸动画 ==========

func pulseBreathing(ctx *ui.Context, th *ui.Theme) ui.Widget {
	active := ui.State[bool](ctx)

	// 两个分支都调用一次 Animate().Value(ctx)，hook 计数一致
	var pulse float32
	if active.Value() {
		pulse = anim.New(anim.From(0), anim.To(1), anim.Duration(800*time.Millisecond), anim.Ease(anim.EaseInOut)).Value(ctx)
		if pulse >= 1.0 {
			active.Set(false)
		}
	} else {
		pulse = anim.New(anim.From(1), anim.To(0), anim.Duration(800*time.Millisecond), anim.Ease(anim.EaseInOut)).Value(ctx)
	}

	size := lerpFloat(60, 120, pulse)
	radius := size / 2
	alpha := uint8(lerpFloat(100, 255, pulse))
	circleColor := color.NRGBA{R: th.Primary.R, G: th.Primary.G, B: th.Primary.B, A: alpha}

	return ui.Column(
		ui.Text("点击圆形开始呼吸动画，大小和透明度同步变化", ui.TextSize(13), ui.TextColor(th.SurfaceMuted)),
		ui.VSpacer(20),
		ui.Center(
			ui.FixedSize(size, size,
				ui.Container(
					ui.Style{Background: circleColor, Radius: radius},
					ui.ClickArea(
						ui.Center(
							ui.Text(fmt.Sprintf("%.0f", size), ui.TextSize(14),
								ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})),
						),
						func(ctx *ui.Context) {
							active.Set(!active.Value())
						},
					),
				),
			),
		),
	)
}

// ========== 3. 交错入场动画 ==========

func staggerEntrance(ctx *ui.Context, th *ui.Theme) ui.Widget {
	playing := ui.State[bool](ctx)

	type staggerItem struct {
		label string
		color color.NRGBA
		delay time.Duration
	}

	items := []staggerItem{
		{"组件 A", color.NRGBA{R: 59, G: 130, B: 246, A: 255}, 0},
		{"组件 B", color.NRGBA{R: 234, G: 88, B: 12, A: 255}, 150 * time.Millisecond},
		{"组件 C", color.NRGBA{R: 22, G: 163, B: 74, A: 255}, 300 * time.Millisecond},
		{"组件 D", color.NRGBA{R: 168, G: 85, B: 247, A: 255}, 450 * time.Millisecond},
		{"组件 E", color.NRGBA{R: 236, G: 72, B: 153, A: 255}, 600 * time.Millisecond},
	}

	children := []ui.Widget{
		ui.Text("5 个卡片依次入场，每个延迟 150ms", ui.TextSize(13), ui.TextColor(th.SurfaceMuted)),
		ui.VSpacer(8),
		ui.Button(ui.Text("开始入场"), ui.OnClick(func(ctx *ui.Context) {
			playing.Set(true)
		})),
		ui.VSpacer(12),
	}

	for _, item := range items {
		totalDuration := 500*time.Millisecond + item.delay

		// 始终调用动画，用 to 控制是否播放
		var to float32
		if playing.Value() {
			to = 1
		}
		rawP := anim.New(anim.From(0), anim.To(to), anim.Duration(totalDuration), anim.Ease(anim.EaseOut)).Value(ctx)

		// 延迟效果：在 delay 阶段保持 0
		var progress float32
		if playing.Value() && rawP > 0 {
			delayRatio := float32(item.delay) / float32(totalDuration)
			if rawP <= delayRatio {
				progress = 0
			} else {
				progress = (rawP - delayRatio) / (1 - delayRatio)
				if progress > 1 {
					progress = 1
				}
			}
		}

		offsetY := lerpFloat(30, 0, progress)
		alpha := uint8(lerpFloat(0, 255, progress))

		card := ui.Padding(
			ui.Insets{Top: offsetY, Bottom: 4},
			ui.Container(
				ui.Style{
					Background: color.NRGBA{R: item.color.R, G: item.color.G, B: item.color.B, A: alpha},
					Padding:    ui.Symmetric(10, 16),
					Radius:     8,
				},
				ui.Text(item.label, ui.TextSize(14),
					ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: alpha})),
			),
		)
		children = append(children, card)
	}

	return ui.Column(children...)
}

// ========== 4. 颜色渐变 ==========

func colorTransition(ctx *ui.Context, th *ui.Theme) ui.Widget {
	step := ui.State[int](ctx)

	palette := []color.NRGBA{
		{R: 59, G: 130, B: 246, A: 255},
		{R: 234, G: 88, B: 12, A: 255},
		{R: 22, G: 163, B: 74, A: 255},
		{R: 168, G: 85, B: 247, A: 255},
		{R: 236, G: 72, B: 153, A: 255},
	}

	current := step.Value() % len(palette)
	next := (step.Value() + 1) % len(palette)

	// 用微小偏移让每次 step 变化都触发 track 重置
	offset := float32(step.Value()) * 0.0001
	p := anim.New(
		anim.From(offset), anim.To(1+offset),
		anim.Duration(600*time.Millisecond),
		anim.Ease(anim.EaseInOut),
	).Value(ctx)
	normalP := clamp01((p - offset) / 1.0)

	bg := lerpColor(palette[current], palette[next], normalP)
	textAlpha := uint8(lerpFloat(180, 255, normalP))

	return ui.Column(
		ui.Text("点击卡片切换到下一个颜色，平滑过渡", ui.TextSize(13), ui.TextColor(th.SurfaceMuted)),
		ui.VSpacer(12),
		ui.Text(fmt.Sprintf("颜色 %d → %d  进度: %.0f%%", current+1, next+1, normalP*100),
			ui.TextSize(12), ui.TextColor(th.SurfaceMuted)),
		ui.VSpacer(8),
		ui.FixedHeight(200,
			ui.FillWidth(
				ui.Container(
					ui.Style{Background: bg, Radius: 16, Padding: ui.All(20)},
					ui.ClickArea(
						ui.Center(
							ui.Column(
								ui.Text("点击切换颜色", ui.TextSize(20),
									ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: textAlpha})),
								ui.VSpacer(8),
								ui.Text(fmt.Sprintf("R:%d G:%d B:%d", bg.R, bg.G, bg.B),
									ui.TextSize(13),
									ui.TextColor(color.NRGBA{R: 255, G: 255, B: 255, A: 180})),
							),
						),
						func(ctx *ui.Context) {
							step.Set(step.Value() + 1)
						},
					),
				),
			),
		),
	)
}

// ========== 5. 进度条动画 ==========

func progressAnimation(ctx *ui.Context, th *ui.Theme) ui.Widget {
	// 0=idle, 1=running, 2=done
	phase := ui.State[int](ctx)

	// 始终调用 1 个动画
	var to float32
	if phase.Value() >= 1 {
		to = 1 // running 和 done 都指向 1，done 时动画已完成不会重置
	}

	p := anim.New(anim.From(0), anim.To(to), anim.Duration(3*time.Second), anim.Ease(anim.EaseInOut)).Value(ctx)

	if phase.Value() == 1 && p >= 1.0 {
		phase.Set(2)
	}

	percent := int(p * 100)

	return ui.Column(
		ui.Text("模拟下载进度，同时驱动进度条和百分比文字", ui.TextSize(13), ui.TextColor(th.SurfaceMuted)),
		ui.VSpacer(12),
		ui.Button(
			ui.Text("开始下载"),
			ui.Disabled(phase.Value() != 0),
			ui.OnClick(func(ctx *ui.Context) {
				phase.Set(1)
			}),
		),
		ui.VSpacer(16),
		ui.Text(fmt.Sprintf("下载进度: %d%%", percent), ui.TextSize(16)),
		ui.VSpacer(8),
		ui.ProgressBar(p, ui.ProgressMax(1)),
		ui.VSpacer(16),
		ui.Row(
			ui.CircularProgress(p, ui.ProgressMax(1)),
			ui.HSpacer(16),
			ui.Column(
				ui.Text(fmt.Sprintf("已完成 %d%%", percent), ui.TextSize(14)),
				ui.VSpacer(4),
				func() ui.Widget {
					switch phase.Value() {
					case 2:
						return ui.Text("下载完成!", ui.TextSize(13),
							ui.TextColor(color.NRGBA{R: 22, G: 163, B: 74, A: 255}))
					case 1:
						return ui.Text("下载中...", ui.TextSize(13),
							ui.TextColor(color.NRGBA{R: 2, G: 132, B: 199, A: 255}))
					default:
						return ui.Text("等待开始", ui.TextSize(13), ui.TextColor(th.SurfaceMuted))
					}
				}(),
			),
		),
	)
}

// ========== 工具函数 ==========

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
