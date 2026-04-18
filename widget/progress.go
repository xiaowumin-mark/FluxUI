package widget

import (
	"fmt"
	"image"
	"image/color"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// ProgressOption 定义进度条配置。
type ProgressOption func(*progressConfig)

type progressConfig struct {
	min           float32
	max           float32
	indeterminate bool
	thickness     float32
	trackColor    color.NRGBA
	fillColor     color.NRGBA
	size          float32
	hasTrackColor bool
	hasFillColor  bool
}

type progressWidget struct {
	value    float32
	circular bool
	config   progressConfig
}

// ProgressBar 创建线性进度条。
func ProgressBar(value float32, opts ...ProgressOption) Widget {
	cfg := progressConfig{
		min:       0,
		max:       100,
		thickness: 8,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &progressWidget{
		value:    value,
		circular: false,
		config:   cfg,
	}
}

// CircularProgress 创建环形进度（当前实现为文本/颜色简化版）。
func CircularProgress(value float32, opts ...ProgressOption) Widget {
	cfg := progressConfig{
		min:       0,
		max:       100,
		thickness: 8,
		size:      64,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &progressWidget{
		value:    value,
		circular: true,
		config:   cfg,
	}
}

// ProgressMin 设置最小值。
func ProgressMin(min float32) ProgressOption {
	return func(cfg *progressConfig) {
		cfg.min = min
	}
}

// ProgressMax 设置最大值。
func ProgressMax(max float32) ProgressOption {
	return func(cfg *progressConfig) {
		cfg.max = max
	}
}

// ProgressIndeterminate 设置不定进度模式。
func ProgressIndeterminate(indeterminate bool) ProgressOption {
	return func(cfg *progressConfig) {
		cfg.indeterminate = indeterminate
	}
}

// ProgressThickness 设置线宽。
func ProgressThickness(thickness float32) ProgressOption {
	return func(cfg *progressConfig) {
		cfg.thickness = thickness
	}
}

// ProgressTrackColor 设置轨道颜色。
func ProgressTrackColor(col color.NRGBA) ProgressOption {
	return func(cfg *progressConfig) {
		cfg.trackColor = col
		cfg.hasTrackColor = true
	}
}

// ProgressFillColor 设置进度颜色。
func ProgressFillColor(col color.NRGBA) ProgressOption {
	return func(cfg *progressConfig) {
		cfg.fillColor = col
		cfg.hasFillColor = true
	}
}

// ProgressSize 设置环形进度尺寸。
func ProgressSize(size float32) ProgressOption {
	return func(cfg *progressConfig) {
		cfg.size = size
	}
}

func (p *progressWidget) Layout(ctx *internal.Context) layout.Dimensions {
	track := ctx.Theme().SurfaceMuted
	fill := ctx.Theme().Primary
	if p.config.hasTrackColor {
		track = p.config.trackColor
	}
	if p.config.hasFillColor {
		fill = p.config.fillColor
	}

	progress := progressRatio(p.value, p.config.min, p.config.max)
	if p.config.indeterminate {
		progress = animProgress(ctx)
		ctx.RequestRedraw()
	}

	if p.circular {
		sizePx := ctx.Gtx.Dp(safeDp(p.config.size))
		if sizePx <= 0 {
			sizePx = ctx.Gtx.Dp(safeDp(64))
		}
		if sizePx < ctx.Gtx.Dp(unit.Dp(24)) {
			sizePx = ctx.Gtx.Dp(unit.Dp(24))
		}

		gtx := ctx.Gtx
		drawCtx := gtx
		drawCtx.Constraints = gioLayout.Exact(image.Point{X: sizePx, Y: sizePx})

		trackStyle := material.ProgressCircle(ctx.MaterialTheme(), 1)
		trackStyle.Color = track
		_ = trackStyle.Layout(drawCtx)

		fillStyle := material.ProgressCircle(ctx.MaterialTheme(), progress)
		fillStyle.Color = fill
		_ = fillStyle.Layout(drawCtx)

		percent := fmt.Sprintf("%.0f%%", progress*100)
		labelWidget := Text(percent, TextSize(12), TextColor(fill))
		labelCtx := gtx
		labelCtx.Constraints.Min = image.Point{}
		labelCtx.Constraints.Max = image.Point{X: sizePx, Y: sizePx}

		labelMacro := op.Record(gtx.Ops)
		next := *ctx
		next.Gtx = labelCtx
		labelSize := labelWidget.Layout(&next).Size
		labelCall := labelMacro.Stop()

		labelX := (sizePx - labelSize.X) / 2
		labelY := (sizePx - labelSize.Y) / 2
		if labelX < 0 {
			labelX = 0
		}
		if labelY < 0 {
			labelY = 0
		}
		stack := op.Offset(image.Point{X: labelX, Y: labelY}).Push(gtx.Ops)
		labelCall.Add(gtx.Ops)
		stack.Pop()

		return layout.Dimensions{Size: image.Point{X: sizePx, Y: sizePx}}
	}

	thickness := ctx.Gtx.Dp(safeDp(p.config.thickness))
	if thickness < 2 {
		thickness = 2
	}

	size := ctx.LayoutInset(internal.Insets{}, func(contentCtx *internal.Context) image.Point {
		maxW := contentCtx.Gtx.Constraints.Max.X
		if maxW <= 0 {
			maxW = contentCtx.Gtx.Dp(safeDp(180))
		}
		if maxW < 1 {
			maxW = 1
		}
		total := image.Point{X: maxW, Y: thickness}
		total = contentCtx.Gtx.Constraints.Constrain(total)
		if total.X <= 0 || total.Y <= 0 {
			return image.Point{}
		}

		paint.FillShape(contentCtx.Gtx.Ops, track, clip.UniformRRect(image.Rectangle{Max: total}, total.Y/2).Op(contentCtx.Gtx.Ops))

		fillW := int(float32(total.X) * progress)
		if fillW < 0 {
			fillW = 0
		}
		if fillW > total.X {
			fillW = total.X
		}
		if fillW > 0 {
			fillRect := image.Rectangle{Max: image.Point{X: fillW, Y: total.Y}}
			paint.FillShape(contentCtx.Gtx.Ops, fill, clip.UniformRRect(fillRect, total.Y/2).Op(contentCtx.Gtx.Ops))
		}

		return total
	})

	return layout.Dimensions{Size: size}
}

func progressRatio(value, min, max float32) float32 {
	if max <= min {
		return 0
	}
	return clampFloat32((value-min)/(max-min), 0, 1)
}

func animProgress(ctx *internal.Context) float32 {
	// 循环 1s 的简单不定进度动画。
	ms := float32(ctx.Now().UnixNano()%1_000_000_000) / 1_000_000_000
	return ms
}
