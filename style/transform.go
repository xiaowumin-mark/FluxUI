package style

import (
	"image"
	"math"

	"gioui.org/f32"
)

// TransformOrigin 定义变换原点。
type TransformOrigin int

const (
	TransformCenter      TransformOrigin = iota // 0
	TransformTopLeft                            // 1
	TransformTopRight                           // 2
	TransformBottomLeft                         // 3
	TransformBottomRight                        // 4
)

// Transform2D 描述 2D 仿射变换（旋转 / 缩放 / 平移）。
type Transform2D struct {
	RotateDeg  float32
	ScaleX     float32
	ScaleY     float32
	TranslateX float32
	TranslateY float32
	Origin     TransformOrigin
}

// Affine 基于容器尺寸计算 Gio 仿射矩阵。
func (t Transform2D) Affine(size image.Point) f32.Affine2D {
	sx := t.ScaleX
	if sx == 0 {
		sx = 1
	}
	sy := t.ScaleY
	if sy == 0 {
		sy = 1
	}
	origin := t.originPoint(size)
	rad := t.RotateDeg * float32(math.Pi) / 180

	return f32.AffineId().
		Offset(f32.Pt(t.TranslateX, t.TranslateY)).
		Rotate(origin, rad).
		Scale(origin, f32.Pt(sx, sy))
}

func (t Transform2D) originPoint(size image.Point) f32.Point {
	switch t.Origin {
	case TransformTopLeft:
		return f32.Pt(0, 0)
	case TransformTopRight:
		return f32.Pt(float32(size.X), 0)
	case TransformBottomLeft:
		return f32.Pt(0, float32(size.Y))
	case TransformBottomRight:
		return f32.Pt(float32(size.X), float32(size.Y))
	default:
		return f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	}
}
