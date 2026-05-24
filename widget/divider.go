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
	vertical   bool
	thickness  float32
	color      color.NRGBA
	hasColor   bool
	length     float32
	margin     style.Insets
	decoration style.Decoration
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

// DividerDecoration 通过 Decoration 统一设置分割线颜色、外边距和圆角。
func DividerDecoration(d style.Decoration) DividerOption {
	return func(cfg *dividerConfig) {
		cfg.decoration = d
	}
}

func (d *dividerWidget) Layout(ctx *internal.Context) layout.Dimensions {
	col := d.config.decoration.ResolveBg(ctx.Theme().SurfaceMuted)
	if d.config.hasColor {
		col = d.config.color
	}
	margin := d.config.decoration.ResolveMargin(d.config.margin)
	radiusDp := d.config.decoration.ResolveRad(0)

	size := ctx.LayoutInset(toInternalInsets(margin), func(contentCtx *internal.Context) image.Point {
		thickness := contentCtx.Gtx.Dp(safeDp(d.config.thickness))
		if thickness < 1 {
			thickness = 1
		}
		radius := contentCtx.Gtx.Dp(safeDp(radiusDp))

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
			if radius > 0 {
				paint.FillShape(contentCtx.Gtx.Ops, col, clip.UniformRRect(image.Rectangle{Max: size}, clampRRectRadiusPx(size, radius)).Op(contentCtx.Gtx.Ops))
			} else {
				paint.FillShape(contentCtx.Gtx.Ops, col, clip.Rect(image.Rectangle{Max: size}).Op())
			}
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
		if radius > 0 {
			paint.FillShape(contentCtx.Gtx.Ops, col, clip.UniformRRect(image.Rectangle{Max: size}, clampRRectRadiusPx(size, radius)).Op(contentCtx.Gtx.Ops))
		} else {
			paint.FillShape(contentCtx.Gtx.Ops, col, clip.Rect(image.Rectangle{Max: size}).Op())
		}
		return size
	})
	return layout.Dimensions{Size: size}
}
