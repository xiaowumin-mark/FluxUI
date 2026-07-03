package internal

import (
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"

	fluxstyle "github.com/xiaowumin-mark/FluxUI/style"
	theme "github.com/xiaowumin-mark/FluxUI/theme"

	"gioui.org/f32"
	gioFont "gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	gioText "gioui.org/text"

	gioLayout "gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Axis 表示主轴方向。
type Axis int

const (
	Horizontal Axis = iota
	Vertical
)

// Alignment 表示文本对齐方式。
type Alignment int

const (
	AlignStart Alignment = iota
	AlignCenter
	AlignEnd
)

// Insets 是内部使用的边距结构。
type Insets struct {
	Top    float32
	Right  float32
	Bottom float32
	Left   float32
}

// FlexChild 是内部 Flex 子项。
type FlexChild struct {
	Flexed bool
	Weight float32
	Layout func(*Context) image.Point
}

// StackChild 是内部 Stack 子项。
type StackChild struct {
	Expanded bool
	Layout   func(*Context) image.Point
}

// TextSpec 描述文本绘制参数。
type TextSpec struct {
	Content    string
	Size       float32
	LineHeight float32
	Color      color.NRGBA
	Alignment  Alignment
	Font       theme.FontSpec
	FontReady  bool
}

// SurfaceSpec 描述容器样式。
type SurfaceSpec struct {
	Background    color.NRGBA
	Radius        float32
	CornerShape   fluxstyle.CornerShapes
	Padding       Insets
	BorderColor   color.NRGBA
	BorderWidth   float32
	Opacity       float32
	GradientStart f32.Point
	GradientEnd   f32.Point
	GradientFrom  color.NRGBA
	GradientTo    color.NRGBA
	HasGradient   bool
	CircleClip    bool
	HasShadow     bool
	ShadowOffsetX float32
	ShadowOffsetY float32
	ShadowBlur    float32
	ShadowSpread  float32
	ShadowColor   color.NRGBA
	ShadowLayers  []fluxstyle.ShadowLayer
	HasImage      bool
	ImageOp       paint.ImageOp
	ImageFit      int
}

// ButtonSpec 描述按钮样式。
type ButtonSpec struct {
	Background   color.NRGBA
	Foreground   color.NRGBA
	TextStyle    theme.TextStyle
	Radius       float32
	CornerShape  fluxstyle.CornerShapes
	Padding      Insets
	BorderColor  color.NRGBA
	BorderWidth  float32
	HasShadow    bool
	Shadow       ShadowSpec
	FocusOpacity float32
	Disabled     bool
}

// InputSpec 描述输入框样式。
type InputSpec struct {
	Background       color.NRGBA
	Foreground       color.NRGBA
	PlaceholderColor color.NRGBA
	Border           color.NRGBA
	Radius           float32
	Padding          Insets
	TextSize         float32
	LineHeight       float32
	Placeholder      string
	Password         bool
	MaxLen           int
	SingleLine       bool
	Font             theme.FontSpec
}

// ShadowSpec 描述按钮等轻量组件阴影。
type ShadowSpec struct {
	OffsetX float32
	OffsetY float32
	Blur    float32
	Spread  float32
	Color   color.NRGBA
	Layers  []fluxstyle.ShadowLayer
}

// CheckboxSpec 描述复选框样式。
type CheckboxSpec struct {
	Size            float32
	Color           color.NRGBA
	CheckedProgress float32
	Disabled        bool
	Hovered         bool
	Pressed         bool
}

// RadioSpec 描述单选框样式。
type RadioSpec struct {
	Size            float32
	Color           color.NRGBA
	CheckedProgress float32
	Disabled        bool
	Hovered         bool
	Pressed         bool
}

// SwitchSpec 描述开关样式。
type SwitchSpec struct {
	Width               float32
	Height              float32
	TrackColor          color.NRGBA
	TrackBorderColor    color.NRGBA
	TrackBorderWidth    float32
	ThumbColor          color.NRGBA
	CheckedProgress     float32
	PositionProgress    float32
	PressedProgress     float32
	ThumbIcon           string
	ThumbIconFontFamily string
	ThumbIconColor      color.NRGBA
	Disabled            bool
	Hovered             bool
	Pressed             bool
}

// SliderSpec 描述滑块样式。
type SliderSpec struct {
	Width         float32
	TrackColor    color.NRGBA
	ThumbColor    color.NRGBA
	ProgressColor color.NRGBA
	Disabled      bool
	Hovered       bool
	Pressed       bool
	Dragged       bool
}

// LayoutText 渲染文本。
func (c *Context) LayoutText(spec TextSpec) image.Point {
	size := spec.Size
	if size <= 0 {
		size = c.Theme().TextSize
	}
	font := spec.Font
	if !spec.FontReady && strings.TrimSpace(font.Family) == "" {
		font = c.Font()
	}
	if !spec.FontReady {
		font = font.Normalize()
	}

	var cacheKey textLayoutCacheKey
	if c.runtime != nil {
		cacheKey = c.textLayoutCacheKey(spec, font, size)
		if entry, ok := c.runtime.lookupTextLayoutCache(cacheKey); ok {
			c.runtime.RecordTextCache(true)
			entry.call.Add(c.Gtx.Ops)
			return entry.size
		}
		c.runtime.RecordTextCache(false)
	}

	label := material.Label(c.MaterialTheme(), unit.Sp(size), spec.Content)
	label.Font = gioFont.Font{
		Typeface: gioFont.Typeface(font.Family),
		Style:    toGioFontStyle(font.Style),
		Weight:   gioFont.Weight(font.Weight),
	}
	label.Color = spec.Color
	label.Alignment = toTextAlignment(spec.Alignment)
	if spec.LineHeight > 0 {
		label.LineHeight = unit.Sp(spec.LineHeight)
	}

	if c.runtime != nil {
		if done := c.startFrameSection(PerfText, 1); done != nil {
			defer done()
		}
		recordOps := new(op.Ops)
		recordGtx := c.Gtx
		recordGtx.Ops = recordOps
		macro := op.Record(recordOps)
		dims := label.Layout(recordGtx)
		call := macro.Stop()
		call.Add(c.Gtx.Ops)
		c.runtime.storeTextLayoutCache(cacheKey, recordOps, call, dims.Size)
		return dims.Size
	}

	if done := c.startFrameSection(PerfText, 1); done != nil {
		defer done()
	}
	dims := label.Layout(c.Gtx)
	return dims.Size
}

// LayoutInset 应用内边距。
func (c *Context) LayoutInset(insets Insets, child func(*Context) image.Point) image.Point {
	c.recordFrameSection(PerfLayout, 1)
	dims := gioLayout.Inset{
		Top:    unit.Dp(insets.Top),
		Right:  unit.Dp(insets.Right),
		Bottom: unit.Dp(insets.Bottom),
		Left:   unit.Dp(insets.Left),
	}.Layout(c.Gtx, func(gtx gioLayout.Context) gioLayout.Dimensions {
		offset := image.Point{
			X: gtx.Dp(unit.Dp(insets.Left)),
			Y: gtx.Dp(unit.Dp(insets.Top)),
		}
		next := c.sameScope(gtx).WithPositionOffset(offset)
		return gioLayout.Dimensions{Size: child(next)}
	})
	return dims.Size
}

// LayoutSurface 绘制带背景的容器。
func (c *Context) LayoutSurface(spec SurfaceSpec, child func(*Context) image.Point) image.Point {
	c.recordFrameSection(PerfLayout, 1)
	gtx := c.Gtx
	if spec.Opacity > 0 && spec.Opacity < 1 {
		defer paint.PushOpacity(gtx.Ops, spec.Opacity).Pop()
	}

	dims := gioLayout.Background{}.Layout(gtx,
		func(gtx gioLayout.Context) gioLayout.Dimensions {
			size := gtx.Constraints.Min

			if spec.HasShadow {
				c.drawShadowLayers(gtx, size, spec)
			}

			if spec.CircleClip {
				c.layoutCircleSurface(gtx, size, spec)
			} else {
				c.layoutRoundedSurface(gtx, size, spec)
			}

			return gioLayout.Dimensions{Size: size}
		},
		func(gtx gioLayout.Context) gioLayout.Dimensions {
			next := c.sameScope(gtx)
			return gioLayout.Dimensions{
				Size: next.LayoutInset(spec.Padding, child),
			}
		},
	)
	return dims.Size
}

func (c *Context) LayoutSurfaceLooseContent(spec SurfaceSpec, child func(*Context) image.Point) image.Point {
	c.recordFrameSection(PerfLayout, 1)
	gtx := c.Gtx
	if spec.Opacity > 0 && spec.Opacity < 1 {
		defer paint.PushOpacity(gtx.Ops, spec.Opacity).Pop()
	}

	contentGtx := gtx
	contentGtx.Constraints.Min = image.Point{}
	contentMacro := op.Record(gtx.Ops)
	contentSize := c.sameScope(contentGtx).LayoutInset(spec.Padding, child)
	contentCall := contentMacro.Stop()

	size := gtx.Constraints.Constrain(contentSize)
	if spec.HasShadow {
		c.drawShadowLayers(gtx, size, spec)
	}
	if spec.CircleClip {
		c.layoutCircleSurface(gtx, size, spec)
	} else {
		c.layoutRoundedSurface(gtx, size, spec)
	}

	if size != contentSize {
		offset := image.Point{
			X: (size.X - contentSize.X) / 2,
			Y: (size.Y - contentSize.Y) / 2,
		}
		stack := op.Offset(offset).Push(gtx.Ops)
		contentCall.Add(gtx.Ops)
		stack.Pop()
	} else {
		contentCall.Add(gtx.Ops)
	}

	return size
}

func (c *Context) DrawSurfaceShadow(size image.Point, spec SurfaceSpec) {
	if size.X <= 0 || size.Y <= 0 || !spec.HasShadow || len(surfaceShadowLayers(spec)) == 0 {
		return
	}
	c.drawShadowLayers(c.Gtx, size, spec)
}

func (c *Context) drawShadowLayers(gtx gioLayout.Context, size image.Point, spec SurfaceSpec) {
	if done := c.startFrameSection(PerfDraw, 1); done != nil {
		defer done()
	}
	layers := surfaceShadowLayers(spec)
	if size.X <= 0 || size.Y <= 0 || len(layers) == 0 {
		return
	}
	radius := gtx.Dp(unit.Dp(spec.Radius))
	for _, layer := range layers {
		layerBlur := gtx.Dp(unit.Dp(layer.Blur))
		if layerBlur <= 0 || layer.Color.A == 0 {
			continue
		}
		layerOffX := gtx.Dp(unit.Dp(layer.OffsetX))
		layerOffY := gtx.Dp(unit.Dp(layer.OffsetY))
		spread := gtx.Dp(unit.Dp(layer.Spread))
		entry := softShadowEntry(size, radius, layerBlur, spread, layerOffX, layerOffY, layer.Color, spec.CircleClip)
		if entry.op.Size().X <= 0 || entry.op.Size().Y <= 0 {
			continue
		}
		stack := op.Offset(image.Pt(-entry.padX, -entry.padY)).Push(gtx.Ops)
		clipStack := clip.Rect(image.Rectangle{Max: entry.op.Size()}).Push(gtx.Ops)
		entry.op.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		clipStack.Pop()
		stack.Pop()
	}
}

func surfaceShadowLayers(spec SurfaceSpec) []fluxstyle.ShadowLayer {
	if len(spec.ShadowLayers) > 0 {
		return spec.ShadowLayers
	}
	if spec.ShadowBlur <= 0 || spec.ShadowColor.A == 0 {
		return nil
	}
	return []fluxstyle.ShadowLayer{{
		OffsetX: spec.ShadowOffsetX,
		OffsetY: spec.ShadowOffsetY,
		Blur:    spec.ShadowBlur,
		Spread:  spec.ShadowSpread,
		Color:   spec.ShadowColor,
	}}
}

func (c *Context) layoutRoundedSurface(gtx gioLayout.Context, size image.Point, spec SurfaceSpec) {
	draw := func(gtx gioLayout.Context) {
		rr := clampRoundedRadiusPx(size, gtx.Dp(unit.Dp(spec.Radius)))
		rect := image.Rectangle{Max: size}
		if rr <= 0 || spec.CornerShape.IsZero() {
			defer clip.UniformRRect(rect, rr).Push(gtx.Ops).Pop()

			paintSurfaceFill(gtx, size, spec)

			if spec.BorderWidth > 0 && spec.BorderColor.A > 0 {
				bw := gtx.Dp(unit.Dp(spec.BorderWidth))
				if bw > 0 {
					paint.FillShape(gtx.Ops, spec.BorderColor, clip.Stroke{
						Path:  clip.UniformRRect(rect, rr).Path(gtx.Ops),
						Width: float32(bw),
					}.Op())
				}
			}
		} else {
			path := cornerShapePath(gtx.Ops, rect, rr, spec.CornerShape)
			defer clip.Outline{Path: path}.Op().Push(gtx.Ops).Pop()

			paintSurfaceFill(gtx, size, spec)

			if spec.BorderWidth > 0 && spec.BorderColor.A > 0 {
				bw := gtx.Dp(unit.Dp(spec.BorderWidth))
				if bw <= 0 {
					return
				}
				paint.FillShape(gtx.Ops, spec.BorderColor, clip.Stroke{
					Path:  path,
					Width: float32(bw),
				}.Op())
			}
		}
	}
	if c.layoutCachedStaticPaint(gtx, size, spec, false, draw) {
		return
	}
	if done := c.startFrameSection(PerfDraw, 1); done != nil {
		defer done()
	}
	draw(gtx)
}

func paintSurfaceFill(gtx gioLayout.Context, size image.Point, spec SurfaceSpec) {
	if spec.HasImage {
		if spec.HasGradient {
			grad := paint.LinearGradientOp{
				Stop1:  spec.GradientStart,
				Color1: spec.GradientFrom,
				Stop2:  spec.GradientEnd,
				Color2: spec.GradientTo,
			}
			grad.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
		} else if spec.Background.A > 0 {
			paint.Fill(gtx.Ops, spec.Background)
		}
		renderImage(gtx, size, spec)
		return
	}
	if spec.HasGradient {
		grad := paint.LinearGradientOp{
			Stop1:  spec.GradientStart,
			Color1: spec.GradientFrom,
			Stop2:  spec.GradientEnd,
			Color2: spec.GradientTo,
		}
		grad.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		return
	}
	paint.Fill(gtx.Ops, spec.Background)
}

func cornerShapePath(ops *op.Ops, rect image.Rectangle, radius int, shapes fluxstyle.CornerShapes) clip.PathSpec {
	r := float32(clampRoundedRadiusPx(rect.Size(), radius))
	x0 := float32(rect.Min.X)
	y0 := float32(rect.Min.Y)
	x1 := float32(rect.Max.X)
	y1 := float32(rect.Max.Y)

	var path clip.Path
	path.Begin(ops)
	path.MoveTo(f32.Pt(x0+r, y0))
	path.LineTo(f32.Pt(x1-r, y0))
	cornerPathTR(&path, x1, y0, r, shapes.TopRight)
	path.LineTo(f32.Pt(x1, y1-r))
	cornerPathBR(&path, x1, y1, r, shapes.BottomRight)
	path.LineTo(f32.Pt(x0+r, y1))
	cornerPathBL(&path, x0, y1, r, shapes.BottomLeft)
	path.LineTo(f32.Pt(x0, y0+r))
	cornerPathTL(&path, x0, y0, r, shapes.TopLeft)
	path.Close()
	return path.End()
}

func cornerPathTR(path *clip.Path, x, y, r float32, shape fluxstyle.CornerShape) {
	switch shape {
	case fluxstyle.CornerSquare:
		path.LineTo(f32.Pt(x, y))
		path.LineTo(f32.Pt(x, y+r))
	case fluxstyle.CornerBevel:
		path.LineTo(f32.Pt(x, y+r))
	case fluxstyle.CornerNotch:
		path.LineTo(f32.Pt(x-r, y+r))
		path.LineTo(f32.Pt(x, y+r))
	case fluxstyle.CornerScoop:
		path.CubeTo(f32.Pt(x-r, y+r*cornerCircleK), f32.Pt(x-r*cornerCircleK, y+r), f32.Pt(x, y+r))
	case fluxstyle.CornerSquircle:
		appendSuperellipseCorner(path, x-r, y+r, r, cornerQuadrantTopRight, cssCornerShapeSquircleExponent)
	default:
		path.CubeTo(f32.Pt(x-r*(1-cornerCircleK), y), f32.Pt(x, y+r*(1-cornerCircleK)), f32.Pt(x, y+r))
	}
}

func cornerPathBR(path *clip.Path, x, y, r float32, shape fluxstyle.CornerShape) {
	switch shape {
	case fluxstyle.CornerSquare:
		path.LineTo(f32.Pt(x, y))
		path.LineTo(f32.Pt(x-r, y))
	case fluxstyle.CornerBevel:
		path.LineTo(f32.Pt(x-r, y))
	case fluxstyle.CornerNotch:
		path.LineTo(f32.Pt(x-r, y-r))
		path.LineTo(f32.Pt(x-r, y))
	case fluxstyle.CornerScoop:
		path.CubeTo(f32.Pt(x-r*cornerCircleK, y-r), f32.Pt(x-r, y-r*cornerCircleK), f32.Pt(x-r, y))
	case fluxstyle.CornerSquircle:
		appendSuperellipseCorner(path, x-r, y-r, r, cornerQuadrantBottomRight, cssCornerShapeSquircleExponent)
	default:
		path.CubeTo(f32.Pt(x, y-r*(1-cornerCircleK)), f32.Pt(x-r*(1-cornerCircleK), y), f32.Pt(x-r, y))
	}
}

func cornerPathBL(path *clip.Path, x, y, r float32, shape fluxstyle.CornerShape) {
	switch shape {
	case fluxstyle.CornerSquare:
		path.LineTo(f32.Pt(x, y))
		path.LineTo(f32.Pt(x, y-r))
	case fluxstyle.CornerBevel:
		path.LineTo(f32.Pt(x, y-r))
	case fluxstyle.CornerNotch:
		path.LineTo(f32.Pt(x+r, y-r))
		path.LineTo(f32.Pt(x, y-r))
	case fluxstyle.CornerScoop:
		path.CubeTo(f32.Pt(x+r, y-r*cornerCircleK), f32.Pt(x+r*cornerCircleK, y-r), f32.Pt(x, y-r))
	case fluxstyle.CornerSquircle:
		appendSuperellipseCorner(path, x+r, y-r, r, cornerQuadrantBottomLeft, cssCornerShapeSquircleExponent)
	default:
		path.CubeTo(f32.Pt(x+r*(1-cornerCircleK), y), f32.Pt(x, y-r*(1-cornerCircleK)), f32.Pt(x, y-r))
	}
}

func cornerPathTL(path *clip.Path, x, y, r float32, shape fluxstyle.CornerShape) {
	switch shape {
	case fluxstyle.CornerSquare:
		path.LineTo(f32.Pt(x, y))
		path.LineTo(f32.Pt(x+r, y))
	case fluxstyle.CornerBevel:
		path.LineTo(f32.Pt(x+r, y))
	case fluxstyle.CornerNotch:
		path.LineTo(f32.Pt(x+r, y+r))
		path.LineTo(f32.Pt(x+r, y))
	case fluxstyle.CornerScoop:
		path.CubeTo(f32.Pt(x+r*cornerCircleK, y+r), f32.Pt(x+r, y+r*cornerCircleK), f32.Pt(x+r, y))
	case fluxstyle.CornerSquircle:
		appendSuperellipseCorner(path, x+r, y+r, r, cornerQuadrantTopLeft, cssCornerShapeSquircleExponent)
	default:
		path.CubeTo(f32.Pt(x, y+r*(1-cornerCircleK)), f32.Pt(x+r*(1-cornerCircleK), y), f32.Pt(x+r, y))
	}
}

const cornerCircleK = 0.5522847498307936
const cssCornerShapeSquircleExponent = 4

type cornerQuadrant uint8

const (
	cornerQuadrantTopRight cornerQuadrant = iota
	cornerQuadrantBottomRight
	cornerQuadrantBottomLeft
	cornerQuadrantTopLeft
)

func appendSuperellipseCorner(path *clip.Path, cx, cy, radius float32, quadrant cornerQuadrant, exponent float64) {
	segments := superellipseSegments(radius)
	for i := 1; i <= segments; i++ {
		angle := float64(i) * (math.Pi / 2) / float64(segments)
		x := math.Pow(math.Sin(angle), 2/exponent)
		y := math.Pow(math.Cos(angle), 2/exponent)
		px, py := superellipseCornerPoint(cx, cy, radius, quadrant, float32(x), float32(y))
		path.LineTo(f32.Pt(px, py))
	}
}

func superellipseSegments(radius float32) int {
	switch {
	case radius >= 48:
		return 14
	case radius >= 24:
		return 12
	case radius >= 12:
		return 10
	default:
		return 8
	}
}

func superellipseCornerPoint(cx, cy, radius float32, quadrant cornerQuadrant, x, y float32) (float32, float32) {
	switch quadrant {
	case cornerQuadrantTopRight:
		return cx + radius*x, cy - radius*y
	case cornerQuadrantBottomRight:
		return cx + radius*y, cy + radius*x
	case cornerQuadrantBottomLeft:
		return cx - radius*x, cy + radius*y
	default:
		return cx - radius*y, cy - radius*x
	}
}

func (c *Context) layoutCircleSurface(gtx gioLayout.Context, size image.Point, spec SurfaceSpec) {
	draw := func(gtx gioLayout.Context) {
		dim := size.X
		if size.Y < dim {
			dim = size.Y
		}
		offX := (size.X - dim) / 2
		offY := (size.Y - dim) / 2
		circleRect := image.Rect(offX, offY, offX+dim, offY+dim)
		ellipse := clip.Ellipse(circleRect)
		defer ellipse.Push(gtx.Ops).Pop()

		if spec.HasImage {
			if spec.HasGradient {
				grad := paint.LinearGradientOp{
					Stop1:  spec.GradientStart,
					Color1: spec.GradientFrom,
					Stop2:  spec.GradientEnd,
					Color2: spec.GradientTo,
				}
				grad.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
			} else if spec.Background.A > 0 {
				paint.Fill(gtx.Ops, spec.Background)
			}
			renderImage(gtx, size, spec)
		} else if spec.HasGradient {
			grad := paint.LinearGradientOp{
				Stop1:  spec.GradientStart,
				Color1: spec.GradientFrom,
				Stop2:  spec.GradientEnd,
				Color2: spec.GradientTo,
			}
			grad.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
		} else {
			paint.Fill(gtx.Ops, spec.Background)
		}

		if spec.BorderWidth > 0 && spec.BorderColor.A > 0 {
			bw := gtx.Dp(unit.Dp(spec.BorderWidth))
			if bw > 0 {
				whalf := (bw + 1) / 2
				inner := circleRect
				inner.Min = inner.Min.Add(image.Point{X: whalf, Y: whalf})
				inner.Max = inner.Max.Sub(image.Point{X: whalf, Y: whalf})
				if inner.Dx() > 0 && inner.Dy() > 0 {
					paint.FillShape(gtx.Ops, spec.BorderColor, clip.Stroke{
						Path:  clip.Ellipse(inner).Path(gtx.Ops),
						Width: float32(bw),
					}.Op())
				}
			}
		}
	}
	if c.layoutCachedStaticPaint(gtx, size, spec, true, draw) {
		return
	}
	if done := c.startFrameSection(PerfDraw, 1); done != nil {
		defer done()
	}
	draw(gtx)
}

func (c *Context) layoutCachedStaticPaint(gtx gioLayout.Context, size image.Point, spec SurfaceSpec, circle bool, draw func(gioLayout.Context)) bool {
	if c == nil || c.runtime == nil || draw == nil || !staticPaintCacheable(spec, size) {
		return false
	}
	key := staticPaintCacheKeyFor(spec, size, circle, float32(gtx.Metric.PxPerDp))
	if entry, ok := c.runtime.lookupStaticPaintCache(key); ok {
		c.runtime.RecordStaticPaintCache(true)
		entry.call.Add(gtx.Ops)
		return true
	}

	c.runtime.RecordStaticPaintCache(false)
	if done := c.startFrameSection(PerfDraw, 1); done != nil {
		defer done()
	}
	recordOps := new(op.Ops)
	recordGtx := gtx
	recordGtx.Ops = recordOps
	macro := op.Record(recordOps)
	draw(recordGtx)
	call := macro.Stop()
	call.Add(gtx.Ops)
	c.runtime.storeStaticPaintCache(key, recordOps, call)
	return true
}

func renderImage(gtx gioLayout.Context, size image.Point, spec SurfaceSpec) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	var fit widget.Fit
	switch spec.ImageFit {
	case 1:
		fit = widget.Cover
	case 2:
		fit = widget.Fill
	case 3:
		fit = widget.Unscaled
	default:
		fit = widget.Contain
	}
	imgWidget := widget.Image{
		Src:      spec.ImageOp,
		Fit:      fit,
		Position: gioLayout.Center,
	}
	imgCtx := gtx
	imgCtx.Constraints = gioLayout.Exact(size)
	imgWidget.Layout(imgCtx)
}

// LayoutButton 绘制按钮并注册点击区域。
func (c *Context) LayoutButton(clickable *ClickableState, spec ButtonSpec, child func(*Context) image.Point) image.Point {
	c.recordFrameSection(PerfLayout, 1)
	gtx := c.Gtx
	// 按钮默认不继承父级的最小高度，避免在 Expanded/Stack 场景被意外拉满。
	gtx.Constraints.Min.Y = 0
	if gtx.Constraints.Min.Y > gtx.Constraints.Max.Y {
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	}
	if spec.Disabled {
		gtx = gtx.Disabled()
	}

	recorded := op.Record(gtx.Ops)
	buttonLayout := func(gtx gioLayout.Context) gioLayout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		dims := gioLayout.Background{}.Layout(gtx,
			func(gtx gioLayout.Context) gioLayout.Dimensions {
				fillRoundedRectShape(gtx, gtx.Constraints.Min, spec.Background, spec.Radius, spec.CornerShape)
				if !spec.Disabled && clickable != nil {
					c.sameScope(gtx).DrawRipple(clickable, gtx.Constraints.Min, RippleSpec{
						Color:   spec.Foreground,
						Radius:  spec.Radius,
						Opacity: 0.12,
					})
				}
				return gioLayout.Dimensions{Size: gtx.Constraints.Min}
			},
			func(gtx gioLayout.Context) gioLayout.Dimensions {
				next := c.sameScope(gtx).WithForeground(spec.Foreground)
				if spec.TextStyle.Size > 0 || spec.TextStyle.LineHeight > 0 {
					next = next.WithTextStyle(spec.TextStyle)
				}
				size := next.LayoutInset(spec.Padding, func(content *Context) image.Point {
					centered := gioLayout.Center.Layout(content.Gtx, func(gtx gioLayout.Context) gioLayout.Dimensions {
						return gioLayout.Dimensions{Size: child(content.sameScope(gtx))}
					})
					return centered.Size
				})
				return gioLayout.Dimensions{Size: size}
			},
		)
		return dims
	}
	var dims gioLayout.Dimensions
	if spec.Disabled || clickable == nil {
		dims = buttonLayout(gtx)
	} else {
		inputDone := c.startFrameSection(PerfInput, 1)
		dims = clickable.raw().Layout(gtx, buttonLayout)
		if inputDone != nil {
			inputDone()
		}
	}
	buttonOps := recorded.Stop()

	if spec.HasShadow && !shadowSpecIsZero(spec.Shadow) {
		c.drawShadowLayers(gtx, dims.Size, SurfaceSpec{
			Radius:        spec.Radius,
			CornerShape:   spec.CornerShape,
			HasShadow:     true,
			ShadowOffsetX: spec.Shadow.OffsetX,
			ShadowOffsetY: spec.Shadow.OffsetY,
			ShadowBlur:    spec.Shadow.Blur,
			ShadowSpread:  spec.Shadow.Spread,
			ShadowColor:   spec.Shadow.Color,
			ShadowLayers:  spec.Shadow.Layers,
		})
	}
	buttonOps.Add(gtx.Ops)
	if spec.BorderWidth > 0 && spec.BorderColor.A > 0 {
		drawBorderWidthShape(gtx, dims.Size, spec.BorderColor, spec.Radius, spec.BorderWidth, spec.CornerShape)
	}
	if !spec.Disabled && clickable != nil && spec.FocusOpacity > 0 {
		focus := c.Theme().Colors.Primary
		focus.A = uint8(float32(focus.A)*spec.FocusOpacity + 0.5)
		c.DrawFocusIndicator(dims.Size, FocusIndicatorSpec{
			Color:  focus,
			Radius: spec.Radius,
		})
	}

	return dims.Size
}

func shadowSpecIsZero(spec ShadowSpec) bool {
	if len(spec.Layers) > 0 {
		for _, layer := range spec.Layers {
			if layer.Blur > 0 && layer.Color.A > 0 {
				return false
			}
		}
		return true
	}
	return spec.Blur <= 0 || spec.Color.A == 0
}

type FocusIndicatorSpec struct {
	Color  color.NRGBA
	Radius float32
	Width  float32
	Inset  float32
}

func (c *Context) DrawFocusIndicator(size image.Point, spec FocusIndicatorSpec) {
	if size.X <= 0 || size.Y <= 0 || spec.Color.A == 0 {
		return
	}
	if done := c.startFrameSection(PerfDraw, 1); done != nil {
		defer done()
	}
	width := c.Gtx.Dp(unit.Dp(spec.Width))
	if width <= 0 {
		width = c.Gtx.Dp(unit.Dp(2))
	}
	if width <= 0 {
		width = 1
	}
	inset := c.Gtx.Dp(unit.Dp(spec.Inset))
	if inset < 0 {
		inset = 0
	}
	rect := image.Rectangle{Max: size}.Inset(inset + width/2)
	if rect.Dx() <= 0 || rect.Dy() <= 0 {
		rect = image.Rectangle{Max: size}
	}
	radius := clampRoundedRadiusPx(rect.Size(), c.Gtx.Dp(unit.Dp(spec.Radius)))
	paint.FillShape(c.Gtx.Ops, spec.Color, clip.Stroke{
		Path:  clip.UniformRRect(rect, radius).Path(c.Gtx.Ops),
		Width: float32(width),
	}.Op())
}

// LayoutClickArea 注册无样式点击区域，不附带任何视觉反馈。
func (c *Context) LayoutClickArea(clickable *ClickableState, child func(*Context) image.Point) image.Point {
	c.recordFrameSection(PerfLayout, 1)
	if child == nil {
		return image.Point{}
	}
	if clickable == nil {
		return child(c.sameScope(c.Gtx))
	}

	inputDone := c.startFrameSection(PerfInput, 1)
	dims := clickable.raw().Layout(c.Gtx, func(gtx gioLayout.Context) gioLayout.Dimensions {
		next := c.sameScope(gtx)
		return gioLayout.Dimensions{Size: child(next)}
	})
	if inputDone != nil {
		inputDone()
	}
	return dims.Size
}

// LayoutInput 绘制输入框。
func (c *Context) LayoutInput(editor *widget.Editor, spec InputSpec) image.Point {
	c.recordFrameSection(PerfLayout, 1)
	if done := c.startFrameSection(PerfInput, 1); done != nil {
		defer done()
	}
	gtx := c.Gtx

	editor.SingleLine = spec.SingleLine

	minSize := gtx.Dp(unit.Dp(36))
	if gtx.Constraints.Min.Y < minSize {
		gtx.Constraints.Min.Y = minSize
	}
	if gtx.Constraints.Min.X < minSize {
		gtx.Constraints.Min.X = minSize
	}

	size := spec.TextSize
	if size <= 0 {
		size = c.Theme().TextSize
	}

	dims := gioLayout.Background{}.Layout(gtx,
		func(gtx gioLayout.Context) gioLayout.Dimensions {
			fillRoundedRect(gtx, gtx.Constraints.Min, spec.Background, spec.Radius)
			drawBorder(gtx, gtx.Constraints.Min, spec.Border, spec.Radius)
			return gioLayout.Dimensions{Size: gtx.Constraints.Min}
		},
		func(gtx gioLayout.Context) gioLayout.Dimensions {
			return gioLayout.Inset{
				Top:    unit.Dp(spec.Padding.Top),
				Right:  unit.Dp(spec.Padding.Right),
				Bottom: unit.Dp(spec.Padding.Bottom),
				Left:   unit.Dp(spec.Padding.Left),
			}.Layout(gtx, func(gtx gioLayout.Context) gioLayout.Dimensions {
				font := spec.Font
				if strings.TrimSpace(font.Family) == "" {
					font = c.Font()
				}
				font = font.Normalize()
				ed := material.Editor(c.MaterialTheme(), editor, spec.Placeholder)
				ed.Font = gioFont.Font{
					Typeface: gioFont.Typeface(font.Family),
					Style:    toGioFontStyle(font.Style),
					Weight:   gioFont.Weight(font.Weight),
				}
				ed.Color = spec.Foreground
				ed.HintColor = colorOrNRGBA(spec.PlaceholderColor, color.NRGBA{R: 150, G: 150, B: 150, A: 255})
				ed.TextSize = unit.Sp(size)
				if spec.LineHeight > 0 {
					ed.LineHeight = unit.Sp(spec.LineHeight)
				}
				return layoutInputEditor(gtx, ed, spec.SingleLine)
			})
		},
	)

	return dims.Size
}

func layoutInputEditor(gtx gioLayout.Context, ed material.EditorStyle, singleLine bool) gioLayout.Dimensions {
	if !singleLine {
		return ed.Layout(gtx)
	}

	minY := gtx.Constraints.Min.Y
	editCtx := gtx
	editCtx.Constraints.Min.Y = 0

	macro := op.Record(gtx.Ops)
	dims := ed.Layout(editCtx)
	call := macro.Stop()

	size := dims.Size
	if size.X < gtx.Constraints.Min.X {
		size.X = gtx.Constraints.Min.X
	}
	if size.Y < minY {
		size.Y = minY
	}
	size = gtx.Constraints.Constrain(size)

	offsetY := (size.Y - dims.Size.Y) / 2
	if offsetY < 0 {
		offsetY = 0
	}
	offset := op.Offset(image.Pt(0, offsetY)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	offset.Pop()

	return gioLayout.Dimensions{
		Size:     size,
		Baseline: dims.Baseline + offsetY,
	}
}

// LayoutCheckbox 绘制复选框。
func (c *Context) LayoutCheckbox(clickable *ClickableState, checked bool, spec CheckboxSpec) image.Point {
	c.recordFrameSection(PerfLayout, 1)
	baseCtx := c.Gtx
	baseCtx.Constraints.Min = image.Point{}
	if spec.Disabled {
		baseCtx = baseCtx.Disabled()
	}

	draw := func(gtx gioLayout.Context) gioLayout.Dimensions {
		size := gtx.Dp(unit.Dp(spec.Size))
		if size <= 0 {
			size = gtx.Dp(unit.Dp(20))
		}
		if size < 14 {
			size = 14
		}

		rect := image.Rectangle{Max: image.Point{X: size, Y: size}}
		onColor := spec.Color
		if onColor.A == 0 {
			onColor = c.Theme().Colors.Primary
		}
		cs := c.Theme().Colors
		fillColor := cs.Surface
		borderColor := colorOrNRGBA(cs.Outline, cs.SurfaceVariant)
		if spec.Disabled {
			onColor = fluxstyle.DisabledContent(cs.OnSurface)
			borderColor = fluxstyle.DisabledContent(cs.OnSurface)
			fillColor = fluxstyle.DisabledContainer(cs.OnSurface)
		}
		progress := spec.CheckedProgress
		if checked && progress <= 0 {
			progress = 1
		}
		fillColor = MixNRGBA(onColor, fillColor, progress)
		borderColor = MixNRGBA(onColor, borderColor, progress)

		radius := size / 5
		if radius < 3 {
			radius = 3
		}
		paint.FillShape(gtx.Ops, fillColor, clip.UniformRRect(rect, radius).Op(gtx.Ops))
		strokeWidth := gtx.Dp(unit.Dp(1))
		if strokeWidth < 1 {
			strokeWidth = 1
		}
		whalf := (strokeWidth + 1) / 2
		strokeRect := rect
		strokeRect.Min = strokeRect.Min.Add(image.Point{X: whalf, Y: whalf})
		strokeRect.Max = strokeRect.Max.Sub(image.Point{X: whalf, Y: whalf})
		if strokeRect.Dx() <= 0 || strokeRect.Dy() <= 0 {
			strokeRect = rect
		}
		paint.FillShape(gtx.Ops, borderColor, clip.Stroke{
			Path:  clip.UniformRRect(strokeRect, radius).Path(gtx.Ops),
			Width: float32(strokeWidth),
		}.Op())
		if progress > 0 {
			mark := cs.OnPrimary
			if spec.Disabled {
				mark = cs.Surface
			}
			mark.A = uint8(float32(mark.A)*progress + 0.5)
			drawCheckMark(gtx, size, mark)
		}
		return gioLayout.Dimensions{Size: rect.Max}
	}

	if clickable == nil {
		return draw(baseCtx).Size
	}
	inputDone := c.startFrameSection(PerfInput, 1)
	size := clickable.raw().Layout(baseCtx, draw).Size
	if inputDone != nil {
		inputDone()
	}
	return size
}

// LayoutRadio 绘制单选框。
func (c *Context) LayoutRadio(clickable *ClickableState, checked bool, spec RadioSpec) image.Point {
	c.recordFrameSection(PerfLayout, 1)
	baseCtx := c.Gtx
	baseCtx.Constraints.Min = image.Point{}
	if spec.Disabled {
		baseCtx = baseCtx.Disabled()
	}

	draw := func(gtx gioLayout.Context) gioLayout.Dimensions {
		size := gtx.Dp(unit.Dp(spec.Size))
		if size <= 0 {
			size = gtx.Dp(unit.Dp(20))
		}
		if size < 14 {
			size = 14
		}

		rect := image.Rectangle{Max: image.Point{X: size, Y: size}}
		onColor := spec.Color
		if onColor.A == 0 {
			onColor = c.Theme().Colors.Primary
		}
		cs := c.Theme().Colors
		borderColor := colorOrNRGBA(cs.Outline, cs.SurfaceVariant)
		bg := cs.Surface
		if spec.Disabled {
			onColor = fluxstyle.DisabledContent(cs.OnSurface)
			borderColor = fluxstyle.DisabledContent(cs.OnSurface)
			bg = fluxstyle.DisabledContainer(cs.OnSurface)
		}
		progress := spec.CheckedProgress
		if checked && progress <= 0 {
			progress = 1
		}
		borderColor = MixNRGBA(onColor, borderColor, progress)
		paint.FillShape(gtx.Ops, bg, clip.Ellipse(rect).Op(gtx.Ops))
		strokeWidth := gtx.Dp(unit.Dp(2))
		if strokeWidth < 1 {
			strokeWidth = 1
		}
		whalf := (strokeWidth + 1) / 2
		strokeRect := rect
		strokeRect.Min = strokeRect.Min.Add(image.Point{X: whalf, Y: whalf})
		strokeRect.Max = strokeRect.Max.Sub(image.Point{X: whalf, Y: whalf})
		if strokeRect.Dx() <= 0 || strokeRect.Dy() <= 0 {
			strokeRect = rect
		}
		paint.FillShape(gtx.Ops, borderColor, clip.Stroke{
			Path:  clip.Ellipse(strokeRect).Path(gtx.Ops),
			Width: float32(strokeWidth),
		}.Op())

		if progress > 0 {
			dotSize := int(float32(size)*0.5*progress + 0.5)
			if dotSize < 4 {
				dotSize = 4
			}
			if dotSize > size-6 {
				dotSize = size - 6
			}
			if dotSize > 4 && (size-dotSize)%2 != 0 {
				dotSize--
			}
			if dotSize > 0 {
				inset := (size - dotSize) / 2
				dotRect := image.Rectangle{
					Min: image.Point{X: inset, Y: inset},
					Max: image.Point{X: inset + dotSize, Y: inset + dotSize},
				}
				dotColor := onColor
				dotColor.A = uint8(float32(dotColor.A)*progress + 0.5)
				paint.FillShape(gtx.Ops, dotColor, clip.Ellipse(dotRect).Op(gtx.Ops))
			}
		}

		return gioLayout.Dimensions{Size: rect.Max}
	}

	if clickable == nil {
		return draw(baseCtx).Size
	}
	inputDone := c.startFrameSection(PerfInput, 1)
	size := clickable.raw().Layout(baseCtx, draw).Size
	if inputDone != nil {
		inputDone()
	}
	return size
}

// LayoutSwitch 绘制开关。
func (c *Context) LayoutSwitch(clickable *ClickableState, checked bool, spec SwitchSpec) image.Point {
	c.recordFrameSection(PerfLayout, 1)
	baseCtx := c.Gtx
	if spec.Disabled {
		baseCtx = baseCtx.Disabled()
	}

	draw := func(gtx gioLayout.Context) gioLayout.Dimensions {
		width := gtx.Dp(unit.Dp(spec.Width))
		height := gtx.Dp(unit.Dp(spec.Height))
		if width <= 0 {
			width = gtx.Dp(unit.Dp(50))
		}
		if height <= 0 {
			height = gtx.Dp(unit.Dp(26))
		}
		if width < height {
			width = height
		}

		trackColor := spec.TrackColor
		thumbColor := spec.ThumbColor
		cs := c.Theme().Colors
		if spec.Disabled {
			trackColor = fluxstyle.DisabledContainer(cs.OnSurface)
			thumbColor = fluxstyle.DisabledContent(cs.OnSurface)
		}

		rr := height / 2
		trackRect := image.Rectangle{Max: image.Point{X: width, Y: height}}
		paint.FillShape(gtx.Ops, trackColor, clip.UniformRRect(trackRect, rr).Op(gtx.Ops))

		if spec.TrackBorderWidth > 0 && spec.TrackBorderColor.A > 0 {
			drawBorderWidth(gtx, trackRect.Max, spec.TrackBorderColor, float32(rr), spec.TrackBorderWidth)
		}

		progress := spec.CheckedProgress
		if checked && progress <= 0 {
			progress = 1
		}
		if progress < 0 {
			progress = 0
		}
		if progress > 1 {
			progress = 1
		}

		positionProgress := spec.PositionProgress
		if positionProgress == 0 && checked {
			positionProgress = progress
		}
		thumbRect := switchThumbRect(gtx, width, height, progress, positionProgress, spec.PressedProgress, spec.ThumbIcon != "")
		thumbSize := thumbRect.Dx()
		thumbRR := thumbSize / 2
		paint.FillShape(gtx.Ops, thumbColor, clip.UniformRRect(thumbRect, thumbRR).Op(gtx.Ops))
		c.drawSwitchThumbIcon(gtx, thumbRect, spec)

		return gioLayout.Dimensions{Size: image.Point{X: width, Y: height}}
	}

	if clickable == nil {
		return draw(baseCtx).Size
	}
	inputDone := c.startFrameSection(PerfInput, 1)
	size := clickable.raw().Layout(baseCtx, draw).Size
	if inputDone != nil {
		inputDone()
	}
	return size
}

func switchThumbRect(gtx gioLayout.Context, width, height int, sizeProgress, positionProgress, pressedProgress float32, hasIcon bool) image.Rectangle {
	if sizeProgress < 0 {
		sizeProgress = 0
	}
	if sizeProgress > 1 {
		sizeProgress = 1
	}
	positionProgress = clampSwitchProgress(positionProgress)
	if pressedProgress < 0 {
		pressedProgress = 0
	}
	if pressedProgress > 1 {
		pressedProgress = 1
	}
	offThumbSize := gtx.Dp(unit.Dp(16))
	if hasIcon {
		offThumbSize = gtx.Dp(unit.Dp(24))
	}
	onThumbSize := gtx.Dp(unit.Dp(24))
	pressedThumbSize := gtx.Dp(unit.Dp(28))
	if offThumbSize <= 0 {
		offThumbSize = height / 2
	}
	if onThumbSize <= 0 {
		onThumbSize = height - 8
	}
	baseThumbSize := offThumbSize + int(float32(onThumbSize-offThumbSize)*sizeProgress+0.5)
	thumbSize := baseThumbSize
	if pressedThumbSize > baseThumbSize {
		thumbSize = baseThumbSize + int(float32(pressedThumbSize-baseThumbSize)*pressedProgress+0.5)
	}
	maxThumb := height - gtx.Dp(unit.Dp(4))
	if thumbSize > maxThumb {
		thumbSize = maxThumb
	}
	if thumbSize < 2 {
		thumbSize = 2
	}

	centerStart := gtx.Dp(unit.Dp(16))
	centerEnd := width - centerStart
	if centerEnd < centerStart {
		centerEnd = centerStart
	}
	centerX := centerStart + int(float32(centerEnd-centerStart)*positionProgress+0.5)
	thumbOffset := centerX - thumbSize/2
	minOffset := gtx.Dp(unit.Dp(2))
	maxOffset := width - thumbSize - minOffset
	if maxOffset < minOffset {
		maxOffset = minOffset
	}
	if thumbOffset < minOffset {
		thumbOffset = minOffset
	}
	if thumbOffset > maxOffset {
		thumbOffset = maxOffset
	}

	thumbTop := (height - thumbSize) / 2
	return image.Rectangle{
		Min: image.Point{X: thumbOffset, Y: thumbTop},
		Max: image.Point{X: thumbOffset + thumbSize, Y: thumbTop + thumbSize},
	}
}

func clampSwitchProgress(v float32) float32 {
	if v < -0.15 {
		return -0.15
	}
	if v > 1.15 {
		return 1.15
	}
	return v
}

func (c *Context) drawSwitchThumbIcon(gtx gioLayout.Context, thumbRect image.Rectangle, spec SwitchSpec) {
	if spec.ThumbIcon == "" || spec.ThumbIconFontFamily == "" || spec.ThumbIconColor.A == 0 || thumbRect.Dx() < 18 || thumbRect.Dy() < 18 {
		return
	}
	if done := c.startFrameSection(PerfText, 1); done != nil {
		defer done()
	}
	label := material.Label(c.MaterialTheme(), unit.Sp(16), spec.ThumbIcon)
	label.Font = gioFont.Font{Typeface: gioFont.Typeface(spec.ThumbIconFontFamily)}
	label.Color = spec.ThumbIconColor
	label.Alignment = gioText.Middle

	iconCtx := gtx
	iconCtx.Constraints = gioLayout.Exact(thumbRect.Size())
	stack := op.Offset(thumbRect.Min).Push(gtx.Ops)
	gioLayout.Center.Layout(iconCtx, func(gtx gioLayout.Context) gioLayout.Dimensions {
		return label.Layout(gtx)
	})
	stack.Pop()
}

// LayoutSlider 绘制滑块。
func (c *Context) LayoutSlider(slider *widget.Float, spec SliderSpec) image.Point {
	c.recordFrameSection(PerfLayout, 1)
	if slider == nil {
		return image.Point{}
	}

	gtx := c.Gtx

	width := gtx.Dp(unit.Dp(spec.Width))
	trackHeight := gtx.Dp(unit.Dp(6))
	thumbSize := gtx.Dp(unit.Dp(20))
	if width <= 0 {
		width = gtx.Dp(unit.Dp(200))
	}
	if thumbSize <= 0 {
		thumbSize = gtx.Dp(unit.Dp(20))
	}
	if width < thumbSize {
		width = thumbSize
	}

	interactiveHeight := thumbSize
	minInteractive := gtx.Dp(unit.Dp(24))
	if interactiveHeight < minInteractive {
		interactiveHeight = minInteractive
	}

	interactiveCtx := gtx
	if spec.Disabled {
		interactiveCtx = interactiveCtx.Disabled()
	}
	interactiveCtx.Constraints = gioLayout.Exact(image.Point{X: width, Y: interactiveHeight})
	inputDone := c.startFrameSection(PerfInput, 1)
	_ = slider.Layout(interactiveCtx, gioLayout.Horizontal, unit.Dp(10))
	if inputDone != nil {
		inputDone()
	}
	spec.Dragged = spec.Dragged || slider.Dragging()

	progress := slider.Value
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	thumbTravel := width - thumbSize
	if thumbTravel < 0 {
		thumbTravel = 0
	}
	thumbLeft := int(float32(thumbTravel)*progress + 0.5)
	thumbCenter := thumbLeft + thumbSize/2

	trackStart := thumbSize / 2
	trackEnd := width - thumbSize/2
	if trackEnd < trackStart {
		trackEnd = trackStart
	}
	trackWidth := trackEnd - trackStart
	progressX := trackStart + int(float32(trackWidth)*progress+0.5)
	centerY := thumbSize / 2
	rr := trackHeight / 2

	cs := c.Theme().Colors
	trackColor, progressColor, thumbColor := resolveSliderVisualColors(spec, cs)

	if trackWidth > 0 {
		trackRect := image.Rectangle{
			Min: image.Point{X: trackStart, Y: centerY - trackHeight/2},
			Max: image.Point{X: trackEnd, Y: centerY + trackHeight/2},
		}
		paint.FillShape(gtx.Ops, trackColor, clip.UniformRRect(trackRect, rr).Op(gtx.Ops))
	}

	if progressX > trackStart {
		progressRect := image.Rectangle{
			Min: image.Point{X: trackStart, Y: centerY - trackHeight/2},
			Max: image.Point{X: progressX, Y: centerY + trackHeight/2},
		}
		paint.FillShape(gtx.Ops, progressColor, clip.UniformRRect(progressRect, rr).Op(gtx.Ops))
	}

	thumbRect := image.Rectangle{
		Min: image.Point{X: thumbCenter - thumbSize/2, Y: 0},
		Max: image.Point{X: thumbCenter + thumbSize/2, Y: thumbSize},
	}
	thumbRR := thumbSize / 2
	paint.FillShape(gtx.Ops, thumbColor, clip.UniformRRect(thumbRect, thumbRR).Op(gtx.Ops))

	return image.Point{X: width, Y: thumbSize}
}

func resolveSliderVisualColors(spec SliderSpec, cs theme.ColorScheme) (track, progress, thumb color.NRGBA) {
	track = spec.TrackColor
	progress = spec.ProgressColor
	thumb = spec.ThumbColor
	if spec.Disabled {
		return fluxstyle.DisabledContainer(cs.OnSurface), fluxstyle.DisabledContent(cs.OnSurface), fluxstyle.DisabledContent(cs.OnSurface)
	}
	return track, progress, thumb
}

func drawCheckMark(gtx gioLayout.Context, size int, col color.NRGBA) {
	if size <= 0 || col.A == 0 {
		return
	}
	w := float32(size)
	stroke := w * 0.14
	if stroke < 2 {
		stroke = 2
	}
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(w*0.24, w*0.55))
	path.LineTo(f32.Pt(w*0.43, w*0.73))
	path.LineTo(f32.Pt(w*0.76, w*0.35))
	paint.FillShape(gtx.Ops, col, clip.Stroke{
		Path:  path.End(),
		Width: stroke,
	}.Op())
}

// DrawCheckMark 绘制一个复用的勾选标记。
func DrawCheckMark(gtx gioLayout.Context, size int, col color.NRGBA) {
	drawCheckMark(gtx, size, col)
}

// LayoutFlex 执行 Flex 布局。
func (c *Context) LayoutFlex(axis Axis, children ...FlexChild) image.Point {
	c.recordFrameSection(PerfLayout, 1)
	flexChildren := make([]gioLayout.FlexChild, 0, len(children))
	rigidOffset := 0
	for index, child := range children {
		idx := index
		layoutChild := func(gtx gioLayout.Context) gioLayout.Dimensions {
			next := c.scopeWithGtx(gtx, "flex-"+strconv.Itoa(idx))
			if !child.Flexed {
				next = next.WithPositionOffset(axisMainOffset(axis, rigidOffset))
			}
			size := child.Layout(next)
			if !child.Flexed {
				rigidOffset += axisMainSize(axis, size)
			}
			return gioLayout.Dimensions{Size: size}
		}
		if child.Flexed {
			flexChildren = append(flexChildren, gioLayout.Flexed(child.Weight, layoutChild))
		} else {
			flexChildren = append(flexChildren, gioLayout.Rigid(layoutChild))
		}
	}
	dims := gioLayout.Flex{Axis: toLayoutAxis(axis)}.Layout(c.Gtx, flexChildren...)
	return dims.Size
}

// LayoutStack 执行 Stack 布局。
func (c *Context) LayoutStack(children ...StackChild) image.Point {
	c.recordFrameSection(PerfLayout, 1)
	stackChildren := make([]gioLayout.StackChild, 0, len(children))
	for index, child := range children {
		idx := index
		layoutChild := func(gtx gioLayout.Context) gioLayout.Dimensions {
			next := c.scopeWithGtx(gtx, "stack-"+strconv.Itoa(idx))
			return gioLayout.Dimensions{Size: child.Layout(next)}
		}
		if child.Expanded {
			stackChildren = append(stackChildren, gioLayout.Expanded(layoutChild))
		} else {
			stackChildren = append(stackChildren, gioLayout.Stacked(layoutChild))
		}
	}
	dims := gioLayout.Stack{}.Layout(c.Gtx, stackChildren...)
	return dims.Size
}

func fillRoundedRect(gtx gioLayout.Context, size image.Point, background color.NRGBA, radius float32) {
	fillRoundedRectShape(gtx, size, background, radius, fluxstyle.CornerShapes{})
}

func fillRoundedRectShape(gtx gioLayout.Context, size image.Point, background color.NRGBA, radius float32, cornerShape fluxstyle.CornerShapes) {
	if size.X <= 0 || size.Y <= 0 || background.A == 0 {
		return
	}
	rr := clampRoundedRadiusPx(size, gtx.Dp(unit.Dp(radius)))
	if rr <= 0 || cornerShape.IsZero() {
		defer clip.UniformRRect(image.Rectangle{Max: size}, rr).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, background)
		return
	}
	defer clip.Outline{Path: cornerShapePath(gtx.Ops, image.Rectangle{Max: size}, rr, cornerShape)}.Op().Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, background)
}

func drawBorder(gtx gioLayout.Context, size image.Point, borderColor color.NRGBA, radius float32) {
	drawBorderWidth(gtx, size, borderColor, radius, 1)
}

func drawBorderWidth(gtx gioLayout.Context, size image.Point, borderColor color.NRGBA, radius, borderWidth float32) {
	drawBorderWidthShape(gtx, size, borderColor, radius, borderWidth, fluxstyle.CornerShapes{})
}

func drawBorderWidthShape(gtx gioLayout.Context, size image.Point, borderColor color.NRGBA, radius, borderWidth float32, cornerShape fluxstyle.CornerShapes) {
	if size.X <= 0 || size.Y <= 0 || borderColor.A == 0 {
		return
	}
	width := gtx.Dp(unit.Dp(borderWidth))
	if width <= 0 {
		width = 1
	}
	whalf := (width + 1) / 2

	rect := image.Rectangle{Max: size}
	rect.Min = rect.Min.Add(image.Point{X: whalf, Y: whalf})
	rect.Max = rect.Max.Sub(image.Point{X: whalf, Y: whalf})
	if rect.Dx() <= 0 || rect.Dy() <= 0 {
		return
	}

	rr := clampRoundedRadiusPx(image.Point{X: rect.Dx(), Y: rect.Dy()}, gtx.Dp(unit.Dp(radius)))
	path := clip.UniformRRect(rect, rr).Path(gtx.Ops)
	if rr > 0 && !cornerShape.IsZero() {
		path = cornerShapePath(gtx.Ops, rect, rr, cornerShape)
	}
	paint.FillShape(gtx.Ops, borderColor, clip.Stroke{
		Path:  path,
		Width: float32(width),
	}.Op())
}

func clampRoundedRadiusPx(size image.Point, rr int) int {
	if rr <= 0 || size.X <= 0 || size.Y <= 0 {
		return 0
	}
	limit := size.X
	if size.Y < limit {
		limit = size.Y
	}
	limit /= 2
	if limit < 0 {
		return 0
	}
	if rr > limit {
		return limit
	}
	return rr
}

func axisMainOffset(axis Axis, main int) image.Point {
	if axis == Vertical {
		return image.Point{Y: main}
	}
	return image.Point{X: main}
}

func axisMainSize(axis Axis, size image.Point) int {
	if axis == Vertical {
		return size.Y
	}
	return size.X
}

// MixNRGBA blends fg over bg by amount in [0, 1].
func MixNRGBA(fg, bg color.NRGBA, amount float32) color.NRGBA {
	if amount < 0 {
		amount = 0
	}
	if amount > 1 {
		amount = 1
	}
	inv := 1 - amount
	return color.NRGBA{
		R: uint8(float32(bg.R)*inv + float32(fg.R)*amount + 0.5),
		G: uint8(float32(bg.G)*inv + float32(fg.G)*amount + 0.5),
		B: uint8(float32(bg.B)*inv + float32(fg.B)*amount + 0.5),
		A: uint8(float32(bg.A)*inv + float32(fg.A)*amount + 0.5),
	}
}

func colorOrNRGBA(col, fallback color.NRGBA) color.NRGBA {
	if col.A == 0 {
		return fallback
	}
	return col
}

func toTextAlignment(alignment Alignment) gioText.Alignment {
	switch alignment {
	case AlignCenter:
		return gioText.Middle
	case AlignEnd:
		return gioText.End
	default:
		return gioText.Start
	}
}

func toGioFontStyle(style theme.FontStyle) gioFont.Style {
	switch style {
	case theme.FontStyleItalic:
		return gioFont.Italic
	default:
		return gioFont.Regular
	}
}

func toLayoutAxis(axis Axis) gioLayout.Axis {
	if axis == Vertical {
		return gioLayout.Vertical
	}
	return gioLayout.Horizontal
}
