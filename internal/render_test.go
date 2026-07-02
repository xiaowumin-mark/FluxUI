package internal

import (
	"image"
	"image/color"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/theme"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestClampRoundedRadiusPxCapsFullRadiusToHalfShortestSide(t *testing.T) {
	got := clampRoundedRadiusPx(image.Pt(120, 40), 999)
	if got != 20 {
		t.Fatalf("clampRoundedRadiusPx() = %d, want 20", got)
	}
}

func TestResolveSliderVisualColorsKeepsStateLayerSeparate(t *testing.T) {
	cs := theme.LightColors()
	spec := SliderSpec{
		TrackColor:    cs.SurfaceVariant,
		ProgressColor: cs.Primary,
		ThumbColor:    cs.Primary,
		Dragged:       true,
		Pressed:       true,
		Hovered:       true,
	}

	track, _, thumb := resolveSliderVisualColors(spec, cs)
	if track != cs.SurfaceVariant {
		t.Fatalf("dragged track = %#v, want %#v", track, cs.SurfaceVariant)
	}
	if thumb != cs.Primary {
		t.Fatalf("dragged thumb = %#v, want %#v", thumb, cs.Primary)
	}
}

func TestDrawFocusIndicatorHandlesEmptyInputs(t *testing.T) {
	_, ctx := newTestContext()
	ctx.DrawFocusIndicator(image.Point{}, FocusIndicatorSpec{Color: theme.LightColors().Primary})
	ctx.DrawFocusIndicator(image.Pt(80, 32), FocusIndicatorSpec{})
}

func TestLayoutTextUsesStaticCacheAcrossFrames(t *testing.T) {
	rt := NewRuntime(nil)
	rt.SetPerfDiagnostics(PerfDiagnostics{Enabled: true, MeasureDurations: true})
	spec := TextSpec{
		Content:   "static text cache",
		Size:      14,
		Color:     theme.LightColors().OnSurface,
		Alignment: AlignStart,
		Font:      theme.DefaultFontSpec(),
		FontReady: true,
	}

	first := layoutRenderCacheTestFrame(rt, image.Pt(320, 120), func(ctx *Context) {
		_ = ctx.LayoutText(spec)
	})
	if first.Cache.TextMisses != 1 || first.Cache.TextHits != 0 || first.Text.Count != 1 {
		t.Fatalf("first text frame stats = %+v, want one miss and one text layout", first)
	}

	second := layoutRenderCacheTestFrame(rt, image.Pt(320, 120), func(ctx *Context) {
		_ = ctx.LayoutText(spec)
	})
	if second.Cache.TextHits != 1 || second.Cache.TextMisses != 0 {
		t.Fatalf("second text frame stats = %+v, want one text cache hit", second)
	}
	if second.Text.Count != 0 {
		t.Fatalf("cache hit should skip text section, got count=%d stats=%+v", second.Text.Count, second)
	}
}

func TestLayoutSurfaceUsesStaticPaintCacheAcrossFrames(t *testing.T) {
	rt := NewRuntime(nil)
	rt.SetPerfDiagnostics(PerfDiagnostics{Enabled: true, MeasureDurations: true})
	spec := SurfaceSpec{
		Background:  color.NRGBA{R: 24, G: 96, B: 160, A: 255},
		Radius:      8,
		BorderColor: color.NRGBA{R: 4, G: 24, B: 48, A: 255},
		BorderWidth: 1,
	}

	first := layoutRenderCacheTestFrame(rt, image.Pt(120, 44), func(ctx *Context) {
		_ = ctx.LayoutSurface(spec, func(*Context) image.Point {
			return image.Pt(120, 44)
		})
	})
	if first.Cache.StaticPaintMisses != 1 || first.Cache.StaticPaintHits != 0 || first.Draw.Count != 1 {
		t.Fatalf("first surface frame stats = %+v, want one static paint miss and one draw", first)
	}

	second := layoutRenderCacheTestFrame(rt, image.Pt(120, 44), func(ctx *Context) {
		_ = ctx.LayoutSurface(spec, func(*Context) image.Point {
			return image.Pt(120, 44)
		})
	})
	if second.Cache.StaticPaintHits != 1 || second.Cache.StaticPaintMisses != 0 {
		t.Fatalf("second surface frame stats = %+v, want one static paint cache hit", second)
	}
	if second.Draw.Count != 0 {
		t.Fatalf("cache hit should skip draw section, got count=%d stats=%+v", second.Draw.Count, second)
	}
}

func TestLayoutStaticSubtreeCacheSkipsChildAcrossFrames(t *testing.T) {
	rt := NewRuntime(nil)
	rt.SetPerfDiagnostics(PerfDiagnostics{Enabled: true, MeasureDurations: true})
	var calls int
	deps := HashStaticDeps("panel", 1)

	first := layoutRenderCacheTestFrame(rt, image.Pt(320, 120), func(ctx *Context) {
		size := ctx.LayoutStaticSubtree(deps, func(childCtx *Context) image.Point {
			calls++
			return childCtx.LayoutText(TextSpec{
				Content:   "static subtree",
				Size:      14,
				Color:     theme.LightColors().OnSurface,
				Font:      theme.DefaultFontSpec(),
				FontReady: true,
			})
		})
		if size.X <= 0 || size.Y <= 0 {
			t.Fatalf("expected non-empty static subtree size, got %v", size)
		}
	})
	if first.Cache.StaticTreeMisses != 1 || first.Cache.StaticTreeHits != 0 || calls != 1 {
		t.Fatalf("first static subtree stats=%+v calls=%d, want one miss and one child call", first, calls)
	}

	second := layoutRenderCacheTestFrame(rt, image.Pt(320, 120), func(ctx *Context) {
		_ = ctx.LayoutStaticSubtree(deps, func(childCtx *Context) image.Point {
			calls++
			return childCtx.LayoutText(TextSpec{
				Content:   "static subtree",
				Size:      14,
				Color:     theme.LightColors().OnSurface,
				Font:      theme.DefaultFontSpec(),
				FontReady: true,
			})
		})
	})
	if second.Cache.StaticTreeHits != 1 || second.Cache.StaticTreeMisses != 0 {
		t.Fatalf("second static subtree stats=%+v, want one hit", second)
	}
	if calls != 1 {
		t.Fatalf("cache hit should skip child layout, calls=%d", calls)
	}
	if second.Text.Count != 0 {
		t.Fatalf("cache hit should skip nested text layout, text count=%d stats=%+v", second.Text.Count, second)
	}
}

func TestSwitchThumbRectMatchesMaterialGeometry(t *testing.T) {
	var ops op.Ops
	gtx := gioLayout.Context{
		Ops:    &ops,
		Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
	}

	cases := []struct {
		name             string
		selectedProgress float32
		positionProgress float32
		pressedProgress  float32
		want             image.Rectangle
	}{
		{name: "unchecked", want: image.Rect(8, 8, 24, 24)},
		{name: "unchecked pressed", pressedProgress: 1, want: image.Rect(2, 2, 30, 30)},
		{name: "checked", selectedProgress: 1, positionProgress: 1, want: image.Rect(24, 4, 48, 28)},
		{name: "checked pressed", selectedProgress: 1, positionProgress: 1, pressedProgress: 1, want: image.Rect(22, 2, 50, 30)},
	}

	for _, tc := range cases {
		got := switchThumbRect(gtx, 52, 32, tc.selectedProgress, tc.positionProgress, tc.pressedProgress, false)
		if got != tc.want {
			t.Fatalf("%s switch thumb rect = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func layoutRenderCacheTestFrame(rt *Runtime, size image.Point, layout func(*Context)) FrameStats {
	var ops op.Ops
	gtx := gioLayout.Context{
		Ops:         &ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: gioLayout.Exact(size),
	}
	rt.BeginFrame()
	ctx := NewContext(gtx, rt)
	if layout != nil {
		layout(ctx)
	}
	rt.EndFrame()
	return rt.LastFrameStats()
}
