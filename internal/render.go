package internal

import (
	"image"
	"image/color"
	"strconv"
	"strings"

	"fluxui/theme"

	"gioui.org/f32"
	gioFont "gioui.org/font"
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
	Content   string
	Size      float32
	Color     color.NRGBA
	Alignment Alignment
	Font      theme.FontSpec
}

// SurfaceSpec 描述容器样式。
type SurfaceSpec struct {
	Background color.NRGBA
	Radius     float32
	Padding    Insets
}

// ButtonSpec 描述按钮样式。
type ButtonSpec struct {
	Background color.NRGBA
	Foreground color.NRGBA
	Radius     float32
	Padding    Insets
	Disabled   bool
}

// InputSpec 描述输入框样式。
type InputSpec struct {
	Background  color.NRGBA
	Foreground  color.NRGBA
	Border      color.NRGBA
	Radius      float32
	Padding     Insets
	TextSize    float32
	Placeholder string
	Password    bool
	MaxLen      int
	SingleLine  bool
	Font        theme.FontSpec
}

// CheckboxSpec 描述复选框样式。
type CheckboxSpec struct {
	Size     float32
	Color    color.NRGBA
	Disabled bool
}

// RadioSpec 描述单选框样式。
type RadioSpec struct {
	Size     float32
	Color    color.NRGBA
	Disabled bool
}

// SwitchSpec 描述开关样式。
type SwitchSpec struct {
	Width      float32
	Height     float32
	TrackColor color.NRGBA
	ThumbColor color.NRGBA
	Disabled   bool
}

// SliderSpec 描述滑块样式。
type SliderSpec struct {
	Width         float32
	TrackColor    color.NRGBA
	ThumbColor    color.NRGBA
	ProgressColor color.NRGBA
	Disabled      bool
}

// LayoutText 渲染文本。
func (c *Context) LayoutText(spec TextSpec) image.Point {
	size := spec.Size
	if size <= 0 {
		size = c.Theme().TextSize
	}
	font := spec.Font
	if strings.TrimSpace(font.Family) == "" {
		font = c.Font()
	}
	font = font.Normalize()
	label := material.Label(c.MaterialTheme(), unit.Sp(size), spec.Content)
	label.Font = gioFont.Font{
		Typeface: gioFont.Typeface(font.Family),
		Style:    toGioFontStyle(font.Style),
		Weight:   gioFont.Weight(font.Weight),
	}
	label.Color = spec.Color
	label.Alignment = toTextAlignment(spec.Alignment)
	dims := label.Layout(c.Gtx)
	return dims.Size
}

// LayoutInset 应用内边距。
func (c *Context) LayoutInset(insets Insets, child func(*Context) image.Point) image.Point {
	dims := gioLayout.Inset{
		Top:    unit.Dp(insets.Top),
		Right:  unit.Dp(insets.Right),
		Bottom: unit.Dp(insets.Bottom),
		Left:   unit.Dp(insets.Left),
	}.Layout(c.Gtx, func(gtx gioLayout.Context) gioLayout.Dimensions {
		next := c.sameScope(gtx)
		return gioLayout.Dimensions{Size: child(next)}
	})
	return dims.Size
}

// LayoutSurface 绘制带背景的容器。
func (c *Context) LayoutSurface(spec SurfaceSpec, child func(*Context) image.Point) image.Point {
	dims := gioLayout.Background{}.Layout(c.Gtx,
		func(gtx gioLayout.Context) gioLayout.Dimensions {
			fillRoundedRect(gtx, gtx.Constraints.Min, spec.Background, spec.Radius)
			return gioLayout.Dimensions{Size: gtx.Constraints.Min}
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

// LayoutButton 绘制按钮并注册点击区域。
func (c *Context) LayoutButton(clickable *ClickableState, spec ButtonSpec, child func(*Context) image.Point) image.Point {
	gtx := c.Gtx
	// 按钮默认不继承父级的最小高度，避免在 Expanded/Stack 场景被意外拉满。
	gtx.Constraints.Min.Y = 0
	if gtx.Constraints.Min.Y > gtx.Constraints.Max.Y {
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	}
	if spec.Disabled {
		gtx = gtx.Disabled()
	}

	style := material.ButtonLayout(c.MaterialTheme(), clickable.raw())
	style.Background = spec.Background
	style.CornerRadius = unit.Dp(spec.Radius)

	dims := style.Layout(gtx, func(gtx gioLayout.Context) gioLayout.Dimensions {
		next := c.sameScope(gtx).WithForeground(spec.Foreground)
		size := next.LayoutInset(spec.Padding, func(content *Context) image.Point {
			centered := gioLayout.Center.Layout(content.Gtx, func(gtx gioLayout.Context) gioLayout.Dimensions {
				return gioLayout.Dimensions{Size: child(content.sameScope(gtx))}
			})
			return centered.Size
		})
		return gioLayout.Dimensions{Size: size}
	})

	return dims.Size
}

// LayoutClickArea 注册无样式点击区域，不附带任何视觉反馈。
func (c *Context) LayoutClickArea(clickable *ClickableState, child func(*Context) image.Point) image.Point {
	if child == nil {
		return image.Point{}
	}
	if clickable == nil {
		return child(c.sameScope(c.Gtx))
	}

	dims := clickable.raw().Layout(c.Gtx, func(gtx gioLayout.Context) gioLayout.Dimensions {
		next := c.sameScope(gtx)
		return gioLayout.Dimensions{Size: child(next)}
	})
	return dims.Size
}

// LayoutInput 绘制输入框。
func (c *Context) LayoutInput(editor *widget.Editor, spec InputSpec) image.Point {
	gtx := c.Gtx

	editor.SingleLine = spec.SingleLine

	minSize := gtx.Dp(unit.Dp(36))
	if gtx.Constraints.Min.Y < minSize {
		gtx.Constraints.Min.Y = minSize
	}
	if gtx.Constraints.Min.X < minSize {
		gtx.Constraints.Min.X = minSize
	}

	textColor := spec.Foreground
	if editor.Len() == 0 {
		textColor = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
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
				ed.Color = textColor
				ed.TextSize = unit.Sp(size)
				return ed.Layout(gtx)
			})
		},
	)

	return dims.Size
}

// LayoutCheckbox 绘制复选框。
func (c *Context) LayoutCheckbox(clickable *ClickableState, checked bool, spec CheckboxSpec) image.Point {
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
			onColor = c.Theme().Primary
		}
		fillColor := c.Theme().Surface
		borderColor := c.Theme().SurfaceMuted
		if spec.Disabled {
			onColor = c.Theme().Disabled
			borderColor = c.Theme().Disabled
		}
		if checked {
			fillColor = onColor
			borderColor = onColor
		}

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
		if checked {
			mark := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			if spec.Disabled {
				mark = c.Theme().Surface
			}
			drawCheckMark(gtx, size, mark)
		}
		return gioLayout.Dimensions{Size: rect.Max}
	}

	if clickable == nil {
		return draw(baseCtx).Size
	}
	return clickable.raw().Layout(baseCtx, draw).Size
}

// LayoutRadio 绘制单选框。
func (c *Context) LayoutRadio(clickable *ClickableState, checked bool, spec RadioSpec) image.Point {
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
		radius := size / 2
		onColor := spec.Color
		if onColor.A == 0 {
			onColor = c.Theme().Primary
		}
		borderColor := c.Theme().SurfaceMuted
		if spec.Disabled {
			onColor = c.Theme().Disabled
			borderColor = c.Theme().Disabled
		}
		if checked {
			borderColor = onColor
		}

		bg := c.Theme().Surface
		paint.FillShape(gtx.Ops, bg, clip.UniformRRect(rect, radius).Op(gtx.Ops))
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

		if checked {
			dotSize := int(float32(size) * 0.42)
			if dotSize < 4 {
				dotSize = 4
			}
			if dotSize > size-6 {
				dotSize = size - 6
			}
			if dotSize > 0 {
				inset := (size - dotSize + 1) / 2
				dotRect := image.Rectangle{
					Min: image.Point{X: inset, Y: inset},
					Max: image.Point{X: inset + dotSize, Y: inset + dotSize},
				}
				paint.FillShape(gtx.Ops, onColor, clip.UniformRRect(dotRect, dotSize/2).Op(gtx.Ops))
			}
		}

		return gioLayout.Dimensions{Size: rect.Max}
	}

	if clickable == nil {
		return draw(baseCtx).Size
	}
	return clickable.raw().Layout(baseCtx, draw).Size
}

// LayoutSwitch 绘制开关。
func (c *Context) LayoutSwitch(clickable *ClickableState, checked bool, spec SwitchSpec) image.Point {
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
		if spec.Disabled {
			trackColor = c.Theme().Disabled
			thumbColor = c.Theme().Surface
		}

		rr := height / 2
		trackRect := image.Rectangle{Max: image.Point{X: width, Y: height}}
		paint.FillShape(gtx.Ops, trackColor, clip.UniformRRect(trackRect, rr).Op(gtx.Ops))

		thumbPadding := 2
		thumbSize := height - thumbPadding*2
		if thumbSize < 2 {
			thumbSize = 2
		}
		thumbOffset := thumbPadding
		if checked {
			thumbOffset = width - thumbSize - thumbPadding
		}
		thumbRect := image.Rectangle{
			Min: image.Point{X: thumbOffset, Y: thumbPadding},
			Max: image.Point{X: thumbOffset + thumbSize, Y: thumbPadding + thumbSize},
		}
		thumbRR := thumbSize / 2
		paint.FillShape(gtx.Ops, thumbColor, clip.UniformRRect(thumbRect, thumbRR).Op(gtx.Ops))

		return gioLayout.Dimensions{Size: image.Point{X: width, Y: height}}
	}

	if clickable == nil {
		return draw(baseCtx).Size
	}
	return clickable.raw().Layout(baseCtx, draw).Size
}

// LayoutSlider 绘制滑块。
func (c *Context) LayoutSlider(slider *widget.Float, spec SliderSpec) image.Point {
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
	_ = slider.Layout(interactiveCtx, gioLayout.Horizontal, unit.Dp(10))

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

	progressColor := spec.ProgressColor
	thumbColor := spec.ThumbColor
	trackColor := spec.TrackColor
	if spec.Disabled {
		progressColor = c.Theme().Disabled
		thumbColor = c.Theme().Disabled
	}

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

// LayoutFlex 执行 Flex 布局。
func (c *Context) LayoutFlex(axis Axis, children ...FlexChild) image.Point {
	flexChildren := make([]gioLayout.FlexChild, 0, len(children))
	for index, child := range children {
		idx := index
		layoutChild := func(gtx gioLayout.Context) gioLayout.Dimensions {
			next := c.childWithGtx(gtx, "flex-"+strconv.Itoa(idx))
			return gioLayout.Dimensions{Size: child.Layout(next)}
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
	stackChildren := make([]gioLayout.StackChild, 0, len(children))
	for index, child := range children {
		idx := index
		layoutChild := func(gtx gioLayout.Context) gioLayout.Dimensions {
			next := c.childWithGtx(gtx, "stack-"+strconv.Itoa(idx))
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
	if size.X <= 0 || size.Y <= 0 || background.A == 0 {
		return
	}
	rr := gtx.Dp(unit.Dp(radius))
	defer clip.UniformRRect(image.Rectangle{Max: size}, rr).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, background)
}

func drawBorder(gtx gioLayout.Context, size image.Point, borderColor color.NRGBA, radius float32) {
	if size.X <= 0 || size.Y <= 0 || borderColor.A == 0 {
		return
	}
	width := gtx.Dp(unit.Dp(1))
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

	rr := gtx.Dp(unit.Dp(radius))
	paint.FillShape(gtx.Ops, borderColor, clip.Stroke{
		Path:  clip.UniformRRect(rect, rr).Path(gtx.Ops),
		Width: float32(width),
	}.Op())
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
