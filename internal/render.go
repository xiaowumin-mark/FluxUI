package internal

import (
	"image"
	"image/color"
	"strconv"

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
}

// CheckboxSpec 描述复选框样式。
type CheckboxSpec struct {
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
	label := material.Label(c.MaterialTheme(), unit.Sp(size), spec.Content)
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

// LayoutInput 绘制输入框。
func (c *Context) LayoutInput(editor *widget.Editor, spec InputSpec) image.Point {
	gtx := c.Gtx

	editor.SingleLine = spec.SingleLine

	minSize := gtx.Dp(unit.Dp(40))
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
				ed := material.Editor(c.MaterialTheme(), editor, spec.Placeholder)
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
	if spec.Disabled {
		baseCtx = baseCtx.Disabled()
	}

	draw := func(gtx gioLayout.Context) gioLayout.Dimensions {
		size := gtx.Dp(unit.Dp(spec.Size))
		if size <= 0 {
			size = gtx.Dp(unit.Dp(20))
		}

		radius := size / 5
		if radius < 2 {
			radius = 2
		}

		rect := image.Rectangle{Max: image.Point{X: size, Y: size}}
		offColor := color.NRGBA{R: 220, G: 220, B: 220, A: 255}
		onColor := spec.Color
		if spec.Disabled {
			onColor = c.Theme().Disabled
		}

		paint.FillShape(gtx.Ops, offColor, clip.UniformRRect(rect, radius).Op(gtx.Ops))
		if checked {
			paint.FillShape(gtx.Ops, onColor, clip.UniformRRect(rect, radius).Op(gtx.Ops))
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

		thumbOffset := height / 4
		if checked {
			thumbOffset = width - height + height/4
		}

		rr := height / 2
		trackRect := image.Rectangle{Max: image.Point{X: width, Y: height}}
		paint.FillShape(gtx.Ops, trackColor, clip.UniformRRect(trackRect, rr).Op(gtx.Ops))

		thumbSize := height - 4
		thumbRect := image.Rectangle{
			Min: image.Point{X: thumbOffset, Y: 2},
			Max: image.Point{X: thumbOffset + thumbSize, Y: height - 2},
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

	trackWidth := width - thumbSize
	progressWidth := int(float32(trackWidth) * progress)
	centerY := thumbSize / 2
	rr := trackHeight / 2

	progressColor := spec.ProgressColor
	thumbColor := spec.ThumbColor
	trackColor := spec.TrackColor
	if spec.Disabled {
		progressColor = c.Theme().Disabled
		thumbColor = c.Theme().Disabled
	}

	if progressWidth > 0 {
		progressRect := image.Rectangle{
			Min: image.Point{X: 0, Y: centerY - trackHeight/2},
			Max: image.Point{X: progressWidth, Y: centerY + trackHeight/2},
		}
		paint.FillShape(gtx.Ops, progressColor, clip.UniformRRect(progressRect, rr).Op(gtx.Ops))
	}

	if progressWidth < trackWidth {
		remainingRect := image.Rectangle{
			Min: image.Point{X: progressWidth, Y: centerY - trackHeight/2},
			Max: image.Point{X: trackWidth, Y: centerY + trackHeight/2},
		}
		paint.FillShape(gtx.Ops, trackColor, clip.UniformRRect(remainingRect, rr).Op(gtx.Ops))
	}

	thumbRect := image.Rectangle{
		Min: image.Point{X: progressWidth, Y: 0},
		Max: image.Point{X: progressWidth + thumbSize, Y: thumbSize},
	}
	thumbRR := thumbSize / 2
	paint.FillShape(gtx.Ops, thumbColor, clip.UniformRRect(thumbRect, thumbRR).Op(gtx.Ops))

	return image.Point{X: width, Y: thumbSize}
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

func toLayoutAxis(axis Axis) gioLayout.Axis {
	if axis == Vertical {
		return gioLayout.Vertical
	}
	return gioLayout.Horizontal
}
