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
