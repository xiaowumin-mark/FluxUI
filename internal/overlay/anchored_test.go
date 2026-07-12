package overlay_test

import (
	"image"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal/testkit"
)

func TestAnchoredOverlayFixtureDefinesProtectedRegions(t *testing.T) {
	fixture := testkit.AnchoredOverlayFixture{
		Anchor:  image.Rect(10, 10, 30, 30),
		Content: image.Rect(10, 34, 80, 90),
	}
	region := fixture.Region()
	if got := len(region.ProtectedRects()); got != 2 {
		t.Fatalf("protected region count = %d, want 2", got)
	}
	for _, point := range []image.Point{{X: 10, Y: 10}, {X: 79, Y: 89}} {
		if !region.Contains(point) {
			t.Fatalf("expected protected point %v", point)
		}
	}
	if region.Contains(image.Pt(30, 30)) || region.Contains(image.Pt(100, 100)) {
		t.Fatal("outside point was treated as protected")
	}
}
