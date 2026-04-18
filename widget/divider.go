package widget

import (
	"image"
	"image/color"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/style"

	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// DividerOption 定义分割线配置。
type DividerOption func(*dividerConfig)

type dividerConfig struct {
	vertical  bool
	thickness float32
	color     color.NRGBA
	hasColor  bool
	length    float32
	margin    style.Insets
}

type dividerWidget struct {
	config dividerConfig
}

// Divider 创建分割线组件。
func Divider(opts ...DividerOption) Widget {
	cfg := dividerConfig{
		thickness: 1,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &dividerWidget{config: cfg}
}

// DividerVertical 设置为垂直分割线。
func DividerVertical(vertical bool) DividerOption {
	return func(cfg *dividerConfig) {
		cfg.vertical = vertical
	}
}

// DividerThickness 设置线宽。
func DividerThickness(thickness float32) DividerOption {
	return func(cfg *dividerConfig) {
		cfg.thickness = thickness
	}
}

// DividerColor 设置颜色。
func DividerColor(col color.NRGBA) DividerOption {
	return func(cfg *dividerConfig) {
		cfg.color = col
		cfg.hasColor = true
	}
}

// DividerLength 设置长度。
func DividerLength(length float32) DividerOption {
	return func(cfg *dividerConfig) {
		cfg.length = length
	}
}

// DividerMargin 设置外边距。
func DividerMargin(insets style.Insets) DividerOption {
	return func(cfg *dividerConfig) {
		cfg.margin = insets
	}
}

func (d *dividerWidget) Layout(ctx *internal.Context) layout.Dimensions {
	col := ctx.Theme().SurfaceMuted
	if d.config.hasColor {
		col = d.config.color
	}

	size := ctx.LayoutInset(insetsToInternal(d.config.margin), func(contentCtx *internal.Context) image.Point {
		thickness := contentCtx.Gtx.Dp(safeDp(d.config.thickness))
		if thickness < 1 {
			thickness = 1
		}

		length := 0
		if d.config.length > 0 {
			length = contentCtx.Gtx.Dp(safeDp(d.config.length))
		}

		if d.config.vertical {
			if length <= 0 {
				length = contentCtx.Gtx.Constraints.Max.Y
			}
			if length <= 0 {
				length = contentCtx.Gtx.Constraints.Min.Y
			}
			size := contentCtx.Gtx.Constraints.Constrain(image.Point{
				X: thickness,
				Y: length,
			})
			if size.X <= 0 || size.Y <= 0 {
				return image.Point{}
			}
			paint.FillShape(contentCtx.Gtx.Ops, col, clip.Rect(image.Rectangle{Max: size}).Op())
			return size
		}

		if length <= 0 {
			length = contentCtx.Gtx.Constraints.Max.X
		}
		if length <= 0 {
			length = contentCtx.Gtx.Constraints.Min.X
		}
		size := contentCtx.Gtx.Constraints.Constrain(image.Point{
			X: length,
			Y: thickness,
		})
		if size.X <= 0 || size.Y <= 0 {
			return image.Point{}
		}
		paint.FillShape(contentCtx.Gtx.Ops, col, clip.Rect(image.Rectangle{Max: size}).Op())
		return size
	})
	return layout.Dimensions{Size: size}
}

func insetsToInternal(insets style.Insets) internal.Insets {
	return internal.Insets{
		Top:    insets.Top,
		Right:  insets.Right,
		Bottom: insets.Bottom,
		Left:   insets.Left,
	}
}
