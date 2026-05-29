package internal

import (
	"image"
	"testing"
)

func TestClampRoundedRadiusPxCapsFullRadiusToHalfShortestSide(t *testing.T) {
	got := clampRoundedRadiusPx(image.Pt(120, 40), 999)
	if got != 20 {
		t.Fatalf("clampRoundedRadiusPx() = %d, want 20", got)
	}
}
