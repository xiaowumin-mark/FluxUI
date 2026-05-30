package widget

import (
	"image"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/theme"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestTextTypeAppliesTypeScaleToken(t *testing.T) {
	style := theme.TextStyle{
		Size:       14,
		LineHeight: 20,
		Weight:     theme.FontWeightMedium,
	}
	var cfg textConfig
	TextType(style)(&cfg)

	if cfg.size != style.Size {
		t.Fatalf("TextType size = %v, want %v", cfg.size, style.Size)
	}
	if cfg.lineHeight != style.LineHeight || !cfg.hasLineHeight {
		t.Fatalf("TextType line height = %v has=%v, want %v", cfg.lineHeight, cfg.hasLineHeight, style.LineHeight)
	}
	if cfg.font.Weight != style.Weight || !cfg.hasWeight {
		t.Fatalf("TextType weight = %v has=%v, want %v", cfg.font.Weight, cfg.hasWeight, style.Weight)
	}
}

func TestTextLineHeightAffectsLayout(t *testing.T) {
	short := Text("Line 1\nLine 2", TextSize(12), TextLineHeight(14)).Layout(testTextContext()).Size.Y
	tall := Text("Line 1\nLine 2", TextSize(12), TextLineHeight(40)).Layout(testTextContext()).Size.Y

	if tall <= short {
		t.Fatalf("tall line height produced height %d, want greater than %d", tall, short)
	}
}

func testTextContext() *internal.Context {
	rt := internal.NewRuntime(theme.New(theme.LightColors()))
	var ops op.Ops
	gtx := gioLayout.Context{
		Ops: &ops,
		Constraints: gioLayout.Constraints{
			Max: image.Pt(1000, 1000),
		},
	}
	rt.BeginFrame()
	return internal.NewContext(gtx, rt)
}
