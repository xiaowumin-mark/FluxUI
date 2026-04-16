package widget

import (
	"fmt"
	"image"
	"image/color"

	"fluxui/internal"
	"fluxui/layout"
	"fluxui/style"

	"gioui.org/op/clip"
	"gioui.org/op/paint"
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

		box := Container(
			style.Style{
				Background: track,
				Padding:    style.All(8),
				Radius:     float32(sizePx / 2),
			},
			Column(
				Text(fmt.Sprintf("%.0f%%", progress*100), TextColor(fill), TextAlign(AlignCenter)),
			),
		)

		return (&fixedSizeWidget{
			width:  float32(sizePx),
			height: float32(sizePx),
			child:  box,
		}).Layout(ctx.Child(0))
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

		paint.FillShape(contentCtx.Gtx.Ops, track, clip.Rect(image.Rectangle{Max: total}).Op())

		fillW := int(float32(total.X) * progress)
		if fillW < 0 {
			fillW = 0
		}
		if fillW > total.X {
			fillW = total.X
		}
		if fillW > 0 {
			fillRect := image.Rectangle{Max: image.Point{X: fillW, Y: total.Y}}
			paint.FillShape(contentCtx.Gtx.Ops, fill, clip.Rect(fillRect).Op())
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
