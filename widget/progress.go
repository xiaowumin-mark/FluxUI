package widget

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/style"

	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

// ProgressOption 定义进度条配置。
type ProgressOption func(*progressConfig)

type progressConfig struct {
	min           float32
	max           float32
	indeterminate bool
	fourColor     bool
	thickness     float32
	trackHeight   float32
	indicatorH    float32
	trackColor    color.NRGBA
	fillColor     color.NRGBA
	bufferColor   color.NRGBA
	size          float32
	showLabel     bool
	buffer        float32
	hasBuffer     bool
	hasTrackColor bool
	hasFillColor  bool
	hasBufferCol  bool
	decoration    style.Decoration
}

type progressWidget struct {
	value    float32
	circular bool
	config   progressConfig
}

// ProgressBar 创建线性进度条。
func ProgressBar(value float32, opts ...ProgressOption) Widget {
	return newProgress(value, false, false, opts...)
}

func LinearProgressIndicator(value float32, opts ...ProgressOption) Widget {
	return newProgress(value, false, true, opts...)
}

// CircularProgress 创建环形进度。
func CircularProgress(value float32, opts ...ProgressOption) Widget {
	return newProgress(value, true, false, opts...)
}

func CircularProgressIndicator(value float32, opts ...ProgressOption) Widget {
	return newProgress(value, true, true, opts...)
}

// LoadingIndicator 创建 Material 3 loading 指示器。
func LoadingIndicator(opts ...ProgressOption) Widget {
	options := append([]ProgressOption{ProgressIndeterminate(true), ProgressSize(24), ProgressThickness(4)}, opts...)
	return newProgress(0, true, true, options...)
}

func newProgress(value float32, circular bool, materialRange bool, opts ...ProgressOption) Widget {
	cfg := progressConfig{
		min:       0,
		max:       1,
		thickness: 4,
		size:      48,
	}
	if !materialRange {
		cfg.max = 100
		cfg.thickness = 8
		cfg.size = 64
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &progressWidget{
		value:    value,
		circular: circular,
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

func ProgressLoading(loading bool) ProgressOption {
	return ProgressIndeterminate(loading)
}

func ProgressFourColor(fourColor bool) ProgressOption {
	return func(cfg *progressConfig) {
		cfg.fourColor = fourColor
	}
}

// ProgressThickness 设置线宽。
func ProgressThickness(thickness float32) ProgressOption {
	return func(cfg *progressConfig) {
		cfg.thickness = thickness
	}
}

func ProgressTrackHeight(height float32) ProgressOption {
	return func(cfg *progressConfig) {
		cfg.trackHeight = height
	}
}

func ProgressIndicatorHeight(height float32) ProgressOption {
	return func(cfg *progressConfig) {
		cfg.indicatorH = height
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

func ProgressActiveIndicatorColor(col color.NRGBA) ProgressOption {
	return ProgressFillColor(col)
}

func ProgressBuffer(value float32) ProgressOption {
	return func(cfg *progressConfig) {
		cfg.buffer = value
		cfg.hasBuffer = true
	}
}

func ProgressBufferColor(col color.NRGBA) ProgressOption {
	return func(cfg *progressConfig) {
		cfg.bufferColor = col
		cfg.hasBufferCol = true
	}
}

// ProgressSize 设置环形进度尺寸。
func ProgressSize(size float32) ProgressOption {
	return func(cfg *progressConfig) {
		cfg.size = size
	}
}

func ProgressLabelVisible(visible bool) ProgressOption {
	return func(cfg *progressConfig) {
		cfg.showLabel = visible
	}
}

// ProgressDecoration 通过 Decoration 统一设置进度组件外层装饰。
func ProgressDecoration(d style.Decoration) ProgressOption {
	return func(cfg *progressConfig) {
		cfg.decoration = d
	}
}

func (p *progressWidget) Layout(ctx *internal.Context) layout.Dimensions {
	gtx := ctx.Gtx
	// 线性进度条不应继承父级强制最小尺寸，否则在某些 Flex/List 场景会被拉成整行高度。
	gtx.Constraints.Min = image.Point{}
	if gtx.Constraints.Min.X > gtx.Constraints.Max.X {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	if gtx.Constraints.Min.Y > gtx.Constraints.Max.Y {
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	}
	ctx = ctx.Scope("progress")
	ctx.Gtx = gtx

	track := ctx.Theme().SurfaceMuted
	fill := ctx.Theme().Primary
	if p.config.hasTrackColor {
		track = p.config.trackColor
	}
	if p.config.hasFillColor {
		fill = p.config.fillColor
	}
	bufferColor := style.StateLayer(track, fill, 0.18)
	if p.config.hasBufferCol {
		bufferColor = p.config.bufferColor
	}

	progress := progressRatio(p.value, p.config.min, p.config.max)
	animatedProgress := md3AnimateFloat(ctx, "progress-value", progress, style.InteractionLoadingValueDuration, style.InteractionStandardEasing)
	if p.config.indeterminate {
		cycle := style.InteractionLoadingLinearCycle
		if p.circular {
			cycle = style.InteractionLoadingCircularCycle
		}
		if rt := ctx.Runtime(); rt != nil {
			rt.RecordFrameSection(internal.PerfAnimation, 1)
		}
		progress = animProgress(ctx, cycle)
		ctx.RequestFrameRedrawReason("animation.progress")
	} else {
		progress = animatedProgress
	}

	layoutProgress := func(progressCtx *internal.Context) layout.Dimensions {
		return p.layoutProgress(progressCtx, track, fill, bufferColor, progress)
	}

	if hasDecorationVisual(p.config.decoration) {
		return ContainerDecoration(p.config.decoration, layoutWidgetFunc(layoutProgress)).Layout(ctx.Child(0))
	}

	return layoutProgress(ctx.Child(0))
}

func (p *progressWidget) layoutProgress(ctx *internal.Context, track, fill, bufferColor color.NRGBA, progress float32) layout.Dimensions {
	if p.circular {
		return p.layoutCircularProgress(ctx, track, fill, progress)
	}
	return p.layoutLinearProgress(ctx, track, fill, bufferColor, progress)
}

func (p *progressWidget) layoutCircularProgress(ctx *internal.Context, track, fill color.NRGBA, progress float32) layout.Dimensions {
	sizePx := ctx.Gtx.Dp(safeDp(p.config.size))
	if sizePx <= 0 {
		sizePx = ctx.Gtx.Dp(safeDp(48))
	}
	if sizePx < ctx.Gtx.Dp(unit.Dp(24)) {
		sizePx = ctx.Gtx.Dp(unit.Dp(24))
	}
	thickness := ctx.Gtx.Dp(safeDp(p.config.thickness))
	if thickness < 2 {
		thickness = 2
	}
	size := image.Point{X: sizePx, Y: sizePx}
	center := f32.Pt(float32(sizePx)/2, float32(sizePx)/2)
	radius := float32(sizePx-thickness) / 2
	if radius < 1 {
		radius = 1
	}
	drawArcStroke(ctx, center, radius, thickness, -math.Pi/2, math.Pi*2, track, false)

	if p.config.indeterminate {
		t := animProgress(ctx, style.InteractionLoadingCircularCycle)
		fill = progressCycleColor(ctx, fill, p.config.fourColor, t)
		start := -math.Pi/2 + float64(t)*math.Pi*2
		sweep := math.Pi*0.35 + math.Pi*0.65*float64(style.InteractionStandardEasing(t))
		drawArcStroke(ctx, center, radius, thickness, start, sweep, fill, true)
	} else if progress > 0 {
		drawArcStroke(ctx, center, radius, thickness, -math.Pi/2, math.Pi*2*float64(progress), fill, true)
	}

	if p.config.showLabel && !p.config.indeterminate {
		percent := fmt.Sprintf("%.0f%%", progress*100)
		labelWidget := Text(percent, TextType(ctx.Theme().Types.LabelSmall), TextColor(fill))
		labelCtx := ctx.Gtx
		labelCtx.Constraints.Min = image.Point{}
		labelCtx.Constraints.Max = size

		labelMacro := op.Record(ctx.Gtx.Ops)
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
		stack := op.Offset(image.Point{X: labelX, Y: labelY}).Push(ctx.Gtx.Ops)
		labelCall.Add(ctx.Gtx.Ops)
		stack.Pop()
	}

	return layout.Dimensions{Size: size}
}

func (p *progressWidget) layoutLinearProgress(ctx *internal.Context, track, fill, bufferColor color.NRGBA, progress float32) layout.Dimensions {
	trackDp := p.config.thickness
	if p.config.trackHeight > 0 {
		trackDp = p.config.trackHeight
	}
	indicatorDp := p.config.thickness
	if p.config.indicatorH > 0 {
		indicatorDp = p.config.indicatorH
	}
	trackH := ctx.Gtx.Dp(safeDp(trackDp))
	indicatorH := ctx.Gtx.Dp(safeDp(indicatorDp))
	if trackH < 2 {
		trackH = 2
	}
	if indicatorH < 2 {
		indicatorH = 2
	}
	height := trackH
	if indicatorH > height {
		height = indicatorH
	}

	size := ctx.LayoutInset(internal.Insets{}, func(contentCtx *internal.Context) image.Point {
		maxW := contentCtx.Gtx.Constraints.Max.X
		if maxW <= 0 {
			maxW = contentCtx.Gtx.Dp(safeDp(240))
		}
		if maxW < 1 {
			maxW = 1
		}
		total := contentCtx.Gtx.Constraints.Constrain(image.Point{X: maxW, Y: height})
		if total.X <= 0 || total.Y <= 0 {
			return image.Point{}
		}
		trackRect := centeredRect(total.X, trackH, total.Y)
		indicatorRect := centeredRect(total.X, indicatorH, total.Y)

		buffer := float32(0)
		if p.config.hasBuffer {
			buffer = progressRatio(p.config.buffer, p.config.min, p.config.max)
		}
		if buffer < progress {
			buffer = progress
		}
		if p.config.hasBuffer && buffer < 1 {
			fillLinearSegment(contentCtx, trackRect, buffer, 1, track, true)
			fillLinearSegment(contentCtx, trackRect, 0, buffer, bufferColor, false)
		} else {
			paint.FillShape(contentCtx.Gtx.Ops, track, clip.UniformRRect(trackRect, trackRect.Dy()/2).Op(contentCtx.Gtx.Ops))
		}

		if p.config.indeterminate {
			t := animProgress(contentCtx, style.InteractionLoadingLinearCycle)
			fill = progressCycleColor(contentCtx, fill, p.config.fourColor, t)
			start1 := float32(math.Mod(float64(t*1.35), 1.35)) - 0.35
			end1 := start1 + 0.42
			start2 := float32(math.Mod(float64(t*1.35+0.55), 1.35)) - 0.35
			end2 := start2 + 0.28
			fillLinearSegment(contentCtx, indicatorRect, start1, end1, fill, false)
			fillLinearSegment(contentCtx, indicatorRect, start2, end2, fill, false)
		} else {
			if p.config.fourColor {
				fill = progressCycleColor(contentCtx, fill, true, progress)
			}
			fillLinearSegment(contentCtx, indicatorRect, 0, progress, fill, false)
		}

		return total
	})

	return layout.Dimensions{Size: size}
}

func centeredRect(width, height, outerHeight int) image.Rectangle {
	top := (outerHeight - height) / 2
	if top < 0 {
		top = 0
	}
	return image.Rect(0, top, width, top+height)
}

func fillLinearSegment(ctx *internal.Context, base image.Rectangle, start, end float32, col color.NRGBA, dotted bool) {
	start = clampFloat32(start, 0, 1)
	end = clampFloat32(end, 0, 1)
	if end <= start || base.Dx() <= 0 || base.Dy() <= 0 || col.A == 0 {
		return
	}
	x0 := base.Min.X + int(float32(base.Dx())*start)
	x1 := base.Min.X + int(float32(base.Dx())*end)
	if x1 <= x0 {
		return
	}
	if dotted {
		dot := base.Dy()
		gap := dot
		for x := x0; x < x1; x += dot + gap {
			rect := image.Rect(x, base.Min.Y, minInt(x+dot, x1), base.Max.Y)
			paint.FillShape(ctx.Gtx.Ops, col, clip.Ellipse(rect).Op(ctx.Gtx.Ops))
		}
		return
	}
	rect := image.Rect(x0, base.Min.Y, x1, base.Max.Y)
	paint.FillShape(ctx.Gtx.Ops, col, clip.UniformRRect(rect, base.Dy()/2).Op(ctx.Gtx.Ops))
}

func drawArcStroke(ctx *internal.Context, center f32.Point, radius float32, width int, start, sweep float64, col color.NRGBA, roundCaps bool) {
	if radius <= 0 || width <= 0 || col.A == 0 || math.Abs(sweep) < 0.001 {
		return
	}
	steps := int(math.Ceil(math.Abs(sweep) / (math.Pi / 24)))
	if steps < 4 {
		steps = 4
	}
	if steps > 96 {
		steps = 96
	}
	var path clip.Path
	path.Begin(ctx.Gtx.Ops)
	for i := 0; i <= steps; i++ {
		a := start + sweep*float64(i)/float64(steps)
		pt := f32.Pt(center.X+float32(math.Cos(a))*radius, center.Y+float32(math.Sin(a))*radius)
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	stroke := clip.Stroke{Path: path.End(), Width: float32(width)}.Op().Push(ctx.Gtx.Ops)
	paint.ColorOp{Color: col}.Add(ctx.Gtx.Ops)
	paint.PaintOp{}.Add(ctx.Gtx.Ops)
	stroke.Pop()
	if roundCaps {
		r := width / 2
		drawArcCap(ctx, center, radius, start, r, col)
		drawArcCap(ctx, center, radius, start+sweep, r, col)
	}
}

func drawArcCap(ctx *internal.Context, center f32.Point, radius float32, angle float64, r int, col color.NRGBA) {
	if r <= 0 {
		return
	}
	x := int(center.X + float32(math.Cos(angle))*radius)
	y := int(center.Y + float32(math.Sin(angle))*radius)
	paint.FillShape(ctx.Gtx.Ops, col, clip.Ellipse(image.Rect(x-r, y-r, x+r, y+r)).Op(ctx.Gtx.Ops))
}

func progressCycleColor(ctx *internal.Context, fallback color.NRGBA, fourColor bool, t float32) color.NRGBA {
	if !fourColor {
		return fallback
	}
	cs := ctx.Theme().Colors
	colors := [4]color.NRGBA{cs.Primary, cs.Secondary, cs.Tertiary, cs.Error}
	idx := int(clampFloat32(t, 0, 0.999) * 4)
	if idx < 0 || idx >= len(colors) {
		return fallback
	}
	return colors[idx]
}

func progressRatio(value, min, max float32) float32 {
	if max <= min {
		return 0
	}
	return clampFloat32((value-min)/(max-min), 0, 1)
}

func animProgress(ctx *internal.Context, cycle time.Duration) float32 {
	if cycle <= 0 {
		cycle = time.Second
	}
	ms := float32(ctx.Now().UnixNano()%int64(cycle)) / float32(cycle)
	return ms
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
