//go:build visual

package main

import (
	"image"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"
	ui "github.com/xiaowumin-mark/FluxUI/ui"

	"gioui.org/gpu/headless"
	"gioui.org/io/input"
	gioLayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestDocsBrowserAppFirstFrameLayouts(t *testing.T) {
	docs, err := loadWidgetDocs()
	if err != nil {
		t.Fatalf("load docs: %v", err)
	}

	runtime := internal.NewRuntime(ui.NewTheme(ui.LightColors()))
	defer runtime.Dispose()
	root := ui.VisualRootBuilder(func(ctx *ui.Context) ui.Element {
		return docsBrowserApp(ctx, &docsRuntimeState{
			Docs:   docs,
			Source: "local",
		})
	})

	var ops op.Ops
	gtx := gioLayout.Context{
		Constraints: gioLayout.Exact(image.Pt(1360, 880)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Ops:         &ops,
	}

	runtime.BeginFrame()
	ctx := runtime.Frame(gtx)
	widget := root(ctx.Scope("build"))
	if widget == nil {
		t.Fatal("expected docs browser widget")
	}
	widget.Layout(ctx.Scope("layout"))
	runtime.EndFrame()
}

func TestDocsBrowserAppFirstFrameScreenshotSmoke(t *testing.T) {
	docs, err := loadWidgetDocs()
	if err != nil {
		t.Fatalf("load docs: %v", err)
	}

	viewport := image.Pt(1360, 880)
	img := renderDocsBrowserScreenshot(t, docs, viewport)
	assertDocsBrowserScreenshotSmoke(t, "docs browser first frame", img, viewport)
}

func renderDocsBrowserScreenshot(t *testing.T, docs []widgetDoc, viewport image.Point) *image.RGBA {
	t.Helper()

	window, err := headless.NewWindow(viewport.X, viewport.Y)
	if err != nil {
		t.Fatalf("create headless window: %v", err)
	}
	defer window.Release()

	runtime := internal.NewRuntime(ui.NewTheme(ui.LightColors()))
	defer runtime.Dispose()
	runtime.SetInvalidator(func() {})

	root := ui.VisualRootBuilder(func(ctx *ui.Context) ui.Element {
		return docsBrowserApp(ctx, &docsRuntimeState{
			Docs:   docs,
			Source: "local",
		})
	})

	var ops op.Ops
	var router input.Router
	baseTime := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	for frame := 0; frame < 3; frame++ {
		ops.Reset()
		gtx := gioLayout.Context{
			Constraints: gioLayout.Exact(viewport),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Now:         baseTime.Add(time.Duration(frame) * 16 * time.Millisecond),
			Source:      router.Source(),
			Ops:         &ops,
		}

		runtime.BeginFrame()
		ctx := runtime.Frame(gtx)
		if widget := root(ctx.Scope("build")); widget != nil {
			widget.Layout(ctx.Scope("layout"))
		}
		runtime.EndFrame()

		router.Frame(&ops)
		if err := window.Frame(&ops); err != nil {
			t.Fatalf("render headless frame %d: %v", frame, err)
		}
	}

	img := image.NewRGBA(image.Rectangle{Max: viewport})
	if err := window.Screenshot(img); err != nil {
		t.Fatalf("capture screenshot: %v", err)
	}
	return img
}

type docsBrowserScreenshotStats struct {
	nonZeroAlpha    int
	opaque          int
	quantizedColors int
	luminanceMin    int
	luminanceMax    int
}

func assertDocsBrowserScreenshotSmoke(t *testing.T, name string, img *image.RGBA, viewport image.Point) docsBrowserScreenshotStats {
	t.Helper()
	if img == nil {
		t.Fatalf("%s: screenshot image is nil", name)
	}
	bounds := img.Bounds()
	if bounds.Dx() != viewport.X || bounds.Dy() != viewport.Y {
		t.Fatalf("%s: screenshot bounds = %v, want %dx%d", name, bounds, viewport.X, viewport.Y)
	}

	stats := analyzeDocsBrowserScreenshot(img)
	totalPixels := viewport.X * viewport.Y
	if stats.nonZeroAlpha < totalPixels*95/100 {
		t.Fatalf("%s: only %d/%d pixels have alpha", name, stats.nonZeroAlpha, totalPixels)
	}
	if stats.opaque < totalPixels*90/100 {
		t.Fatalf("%s: only %d/%d pixels are opaque", name, stats.opaque, totalPixels)
	}
	if stats.quantizedColors < 24 {
		t.Fatalf("%s: only %d quantized colors", name, stats.quantizedColors)
	}
	if stats.luminanceMax-stats.luminanceMin < 24 {
		t.Fatalf("%s: luminance range %d-%d is too narrow", name, stats.luminanceMin, stats.luminanceMax)
	}
	return stats
}

func analyzeDocsBrowserScreenshot(img *image.RGBA) docsBrowserScreenshotStats {
	bounds := img.Bounds()
	colors := make(map[uint32]struct{}, 128)
	stats := docsBrowserScreenshotStats{luminanceMin: 255}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		offset := img.PixOffset(bounds.Min.X, y)
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r := img.Pix[offset]
			g := img.Pix[offset+1]
			b := img.Pix[offset+2]
			a := img.Pix[offset+3]
			offset += 4

			if a > 0 {
				stats.nonZeroAlpha++
			}
			if a == 255 {
				stats.opaque++
			}
			luminance := (int(r)*299 + int(g)*587 + int(b)*114) / 1000
			if luminance < stats.luminanceMin {
				stats.luminanceMin = luminance
			}
			if luminance > stats.luminanceMax {
				stats.luminanceMax = luminance
			}
			key := uint32(r>>4)<<12 | uint32(g>>4)<<8 | uint32(b>>4)<<4 | uint32(a>>4)
			colors[key] = struct{}{}
		}
	}
	stats.quantizedColors = len(colors)
	return stats
}
