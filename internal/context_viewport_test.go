package internal

import (
	"image"
	"testing"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestContextViewportAndPositionScopes(t *testing.T) {
	var ops op.Ops
	constraints := gioLayout.Constraints{Min: image.Pt(20, 10), Max: image.Pt(100, 80)}
	context := NewContext(gioLayout.Context{Ops: &ops, Constraints: constraints}, NewRuntime(nil))

	if context.MinConstraints() != constraints.Min || context.MaxConstraints() != constraints.Max {
		t.Fatalf("constraints = min %v max %v, want min %v max %v", context.MinConstraints(), context.MaxConstraints(), constraints.Min, constraints.Max)
	}
	if viewport, ok := context.Viewport(); !ok || viewport != (image.Rectangle{Max: constraints.Max}) {
		t.Fatalf("initial viewport = %v, %t", viewport, ok)
	}
	if context.Position() != (image.Point{}) {
		t.Fatalf("initial position = %v", context.Position())
	}

	shifted := context.WithPositionOffset(image.Pt(10, 5)).WithPositionOffset(image.Pt(-2, 3))
	if shifted.Position() != image.Pt(8, 8) {
		t.Fatalf("shifted position = %v, want (8,8)", shifted.Position())
	}
	if context.Position() != (image.Point{}) {
		t.Fatalf("original context position changed to %v", context.Position())
	}
	visible := image.Rect(4, 6, 70, 50)
	constrained := shifted.WithViewport(visible)
	if viewport, ok := constrained.Viewport(); !ok || viewport != visible {
		t.Fatalf("constrained viewport = %v, %t", viewport, ok)
	}
	if viewport, ok := context.WithViewport(image.Rectangle{}).Viewport(); ok || viewport != (image.Rectangle{}) {
		t.Fatalf("empty viewport = %v, %t", viewport, ok)
	}
	if viewport, ok := context.WithViewport(image.Rect(10, 10, 10, 20)).Viewport(); ok || viewport != (image.Rectangle{}) {
		t.Fatalf("zero-width viewport = %v, %t", viewport, ok)
	}

	var nilContext *Context
	if nilContext.MinConstraints() != (image.Point{}) || nilContext.MaxConstraints() != (image.Point{}) || nilContext.Position() != (image.Point{}) {
		t.Fatal("nil context geometry helpers returned non-zero values")
	}
	if viewport, ok := nilContext.Viewport(); ok || viewport != (image.Rectangle{}) {
		t.Fatalf("nil context viewport = %v, %t", viewport, ok)
	}
	if nilContext.WithViewport(visible) != nil || nilContext.WithPositionOffset(image.Pt(1, 1)) != nil {
		t.Fatal("nil context scope helpers returned a context")
	}
}
