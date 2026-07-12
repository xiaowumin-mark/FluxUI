package testkit

import (
	"image"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/internal/collection"
)

func TestFrameHarnessDrivesStableFrames(t *testing.T) {
	harness := NewFrameHarness(image.Pt(320, 180))
	defer harness.Close()
	start := harness.Now
	frames := 0
	harness.Frame(func(ctx *internal.Context) {
		frames++
		viewport, ok := ctx.Viewport()
		if !ok || viewport.Max != harness.Size {
			t.Fatalf("viewport = %v, ok=%t", viewport, ok)
		}
	})
	if frames != 1 || !harness.Now.Equal(start.Add(16*time.Millisecond)) {
		t.Fatalf("frames=%d now=%v", frames, harness.Now)
	}
}

func TestContractFixturesAndZeroSizedHarness(t *testing.T) {
	model := (CollectionFixture{Items: []collection.Item{{Key: "one"}}}).Model(t)
	if model.Len() != 1 {
		t.Fatalf("fixture model length = %d", model.Len())
	}
	region := (AnchoredOverlayFixture{Anchor: image.Rect(0, 0, 1, 1)}).Region()
	if !region.Contains(image.Point{}) {
		t.Fatal("fixture region did not retain anchor")
	}
	harness := NewFrameHarness(image.Point{})
	defer harness.Close()
	if harness.Size != (image.Point{X: 1, Y: 1}) {
		t.Fatalf("zero-sized harness = %v", harness.Size)
	}
	harness.Frame(nil)
	harness.Close()
	harness.Frame(nil)
}
