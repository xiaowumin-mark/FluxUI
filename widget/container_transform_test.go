package widget

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/style"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestDecorationTransformMatrixUsesLaidOutSizeForOrigin(t *testing.T) {
	gtx := gioLayout.Context{}
	matrix := decorationTransformMatrix(gtx, style.Transform2D{
		RotateDeg: 90,
		ScaleX:    1,
		ScaleY:    1,
		Origin:    style.TransformCenter,
	}, image.Pt(100, 40))

	center := matrix.Transform(gioLayout.FPt(image.Pt(50, 20)))
	if !approxFloat32(center.X, 50) || !approxFloat32(center.Y, 20) {
		t.Fatalf("expected center origin to stay fixed at (50,20), got (%v,%v)", center.X, center.Y)
	}

	topLeft := matrix.Transform(gioLayout.FPt(image.Pt(0, 0)))
	if topLeft.X == 0 && topLeft.Y == 0 {
		t.Fatalf("expected rotation around real center, got top-left unchanged")
	}
}

func approxFloat32(got, want float32) bool {
	return math.Abs(float64(got-want)) < 0.001
}

func TestDecorationTransformMatrixConvertsTranslateDpToPx(t *testing.T) {
	gtx := gioLayout.Context{Metric: unit.Metric{PxPerDp: 2}}
	matrix := decorationTransformMatrix(gtx, style.Transform2D{
		ScaleX:     1,
		ScaleY:     1,
		TranslateX: 8,
		TranslateY: -4,
	}, image.Pt(10, 10))

	_, _, ox, _, _, oy := matrix.Elems()
	if ox != 16 || oy != -8 {
		t.Fatalf("expected translate dp converted to (16,-8) px, got (%v,%v)", ox, oy)
	}
}

func TestLayoutSurfaceWithTransformKeepsLayoutSize(t *testing.T) {
	var ops op.Ops
	gtx := gioLayout.Context{
		Ops:         &ops,
		Constraints: gioLayout.Exact(image.Pt(120, 36)),
	}
	ctx := internal.NewContext(gtx, internal.NewRuntime(nil))

	size := layoutSurfaceWithTransform(ctx, internal.SurfaceSpec{}, &style.Transform2D{
		RotateDeg: 12,
		ScaleX:    1.2,
		ScaleY:    1.2,
		Origin:    style.TransformCenter,
	}, func(*internal.Context) image.Point {
		return image.Pt(120, 36)
	})

	if size != (image.Pt(120, 36)) {
		t.Fatalf("expected transform not to affect layout size, got %v", size)
	}
}

func TestInteractiveDecorationTransformKeepsLayoutSize(t *testing.T) {
	var ops op.Ops
	gtx := gioLayout.Context{
		Ops:         &ops,
		Constraints: gioLayout.Exact(image.Pt(120, 36)),
	}
	ctx := internal.NewContext(gtx, internal.NewRuntime(nil))

	w := ContainerDecoration(
		style.Decoration{}.
			WithBg(color.NRGBA{A: 255}).
			WithHover(style.Decoration{}.WithTransform(style.Transform2D{
				ScaleX: 1.2,
				ScaleY: 1.2,
				Origin: style.TransformCenter,
			})),
		layoutWidgetFunc(func(*internal.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(120, 36)}
		}),
	)

	dims := w.Layout(ctx)
	if dims.Size != (image.Pt(120, 36)) {
		t.Fatalf("expected interactive transform not to affect layout size, got %v", dims.Size)
	}
}
