package internal

import (
	"image"
	"image/color"
	"testing"
)

func TestSoftShadowMasksBlurAndBounds(t *testing.T) {
	rounded := make([]uint8, 25)
	drawRoundedRectMask(rounded, 5, 5, image.Rect(1, 1, 4, 4), 0, 120)
	if rounded[1+1*5] != 120 || rounded[3+3*5] != 120 || rounded[0] != 0 {
		t.Fatalf("rectangle mask was not filled as expected: %#v", rounded)
	}

	cornered := make([]uint8, 25)
	drawRoundedRectMask(cornered, 5, 5, image.Rect(0, 0, 5, 5), 99, 200)
	if cornered[2+2*5] != 200 || cornered[0] != 0 {
		t.Fatalf("rounded corner mask was not clipped as expected: %#v", cornered)
	}
	drawRoundedRectMask(cornered, 5, 5, image.Rectangle{}, 2, 255)
	drawRoundedRectMask(cornered, 5, 5, image.Rect(0, 0, 2, 2), 2, 0)

	ellipse := make([]uint8, 25)
	drawEllipseMask(ellipse, 5, 5, image.Rect(0, 0, 5, 5), 180)
	if ellipse[2+2*5] != 180 || ellipse[0] != 0 {
		t.Fatalf("ellipse mask was not filled as expected: %#v", ellipse)
	}
	drawEllipseMask(ellipse, 5, 5, image.Rectangle{}, 255)
	drawEllipseMask(ellipse, 5, 5, image.Rect(0, 0, 2, 2), 0)

	clipped := make([]uint8, 6)
	fillMaskRect(clipped, 3, 2, image.Rect(-1, -1, 2, 2), 77)
	if clipped[0] != 77 || clipped[1] != 77 || clipped[3] != 77 || clipped[4] != 77 || clipped[2] != 0 || clipped[5] != 0 {
		t.Fatalf("clipped fill = %#v", clipped)
	}

	horizontal := make([]uint8, 3)
	blurHorizontal([]uint8{0, 255, 0}, horizontal, 3, 1, 1)
	if len(horizontal) != 3 || horizontal[0] != 85 || horizontal[1] != 85 || horizontal[2] != 85 {
		t.Fatalf("horizontal blur = %#v, want [85 85 85]", horizontal)
	}
	vertical := make([]uint8, 3)
	blurVertical([]uint8{0, 255, 0}, vertical, 1, 3, 1)
	if len(vertical) != 3 || vertical[0] != 85 || vertical[1] != 85 || vertical[2] != 85 {
		t.Fatalf("vertical blur = %#v, want [85 85 85]", vertical)
	}

	alpha := make([]uint8, 9)
	alpha[4] = 255
	boxBlurAlpha(alpha, 3, 3, 1)
	nonZero := 0
	for _, value := range alpha {
		if value > 0 {
			nonZero++
		}
	}
	if nonZero == 0 || alpha[4] == 255 {
		t.Fatalf("box blur did not diffuse alpha: %#v", alpha)
	}
	boxBlurAlpha(alpha, 0, 3, 1)
	boxBlurAlpha(alpha, 3, 3, 0)

	if premul(255, 128) != 128 || premul(100, 0) != 0 || absInt(-4) != 4 || absInt(4) != 4 {
		t.Fatal("alpha arithmetic helpers returned unexpected results")
	}
	if minInt(1, 2) != 1 || minInt(2, 1) != 1 || maxInt(1, 2) != 2 || maxInt(2, 1) != 2 {
		t.Fatal("min/max helpers returned unexpected results")
	}
	if clampInt(-1, 0, 3) != 0 || clampInt(4, 0, 3) != 3 || clampInt(2, 0, 3) != 2 {
		t.Fatal("clamp helper returned unexpected results")
	}
}

func TestSoftShadowEntriesNormalizeAndCache(t *testing.T) {
	shadowColor := color.NRGBA{R: 100, G: 80, B: 60, A: 220}
	if entry := softShadowEntry(image.Point{}, 1, 1, 0, 0, 0, shadowColor, false); entry.padX != 0 || entry.padY != 0 {
		t.Fatalf("invalid shadow produced padding: %#v", entry)
	}
	if entry := softShadowEntry(image.Pt(1, 1), 1, 0, 0, 0, 0, shadowColor, false); entry.padX != 0 || entry.padY != 0 {
		t.Fatalf("zero-blur shadow produced padding: %#v", entry)
	}
	if entry := softShadowEntry(image.Pt(1, 1), 1, 1, 0, 0, 0, color.NRGBA{}, false); entry.padX != 0 || entry.padY != 0 {
		t.Fatalf("transparent shadow produced padding: %#v", entry)
	}

	entry := softShadowEntry(image.Pt(4, 3), -1, 97, -1, -2, 3, shadowColor, false)
	if entry.padX != 196 || entry.padY != 197 {
		t.Fatalf("normalized shadow padding = (%d,%d), want (196,197)", entry.padX, entry.padY)
	}
	cached := softShadowEntry(image.Pt(4, 3), -1, 97, -1, -2, 3, shadowColor, false)
	if cached.padX != entry.padX || cached.padY != entry.padY {
		t.Fatalf("cached shadow padding = (%d,%d), want (%d,%d)", cached.padX, cached.padY, entry.padX, entry.padY)
	}
	circle := softShadowEntry(image.Pt(4, 3), 1, 1, 0, 0, 0, shadowColor, true)
	if circle.padX == 0 || circle.padY == 0 {
		t.Fatalf("circle shadow did not produce an entry: %#v", circle)
	}
	if entry := buildSoftShadowEntry(image.Pt(-100, 1), 0, 0, 0, 0, 0, shadowColor, false); entry.padX != 0 || entry.padY != 0 {
		t.Fatalf("invalid built shadow produced padding: %#v", entry)
	}
}
