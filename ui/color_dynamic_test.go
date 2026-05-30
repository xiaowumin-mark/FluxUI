package ui

import (
	"image/color"
	"testing"
)

func TestSeedColorHelpersAreExportedThroughUI(t *testing.T) {
	seed := color.NRGBA{R: 0x00, G: 0x57, B: 0xd9, A: 255}
	light := LightColorSchemeFromSeed(seed)
	dark := DarkColorSchemeFromSeed(seed)
	darkViaOption := ColorSchemeFromSeed(seed, WithColorSchemeDark(true))
	th := ThemeFromSeed(seed, WithSecondarySeed(color.NRGBA{R: 0xa6, G: 0x33, B: 0xff, A: 255}))

	if light.Primary == dark.Primary {
		t.Fatalf("ui seed helpers should expose distinct light and dark schemes")
	}
	if darkViaOption != dark {
		t.Fatalf("ui WithColorSchemeDark should route to dark seed scheme")
	}
	if th == nil || th.Colors.Primary.A == 0 || th.Colors.Secondary.A == 0 {
		t.Fatalf("ui ThemeFromSeed should create a populated theme")
	}
}
