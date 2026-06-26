package md3

import (
	"testing"

	"github.com/xiaowumin-mark/FluxUI/theme"
)

func TestFontRegistersDefaultMD3IconFont(t *testing.T) {
	font, ok := theme.DefaultIconFont()
	if !ok {
		t.Fatal("expected a default icon font")
	}
	if font.ID != ID {
		t.Fatalf("default icon font ID = %q, want %q", font.ID, ID)
	}
	if font.Family != Family {
		t.Fatalf("default icon font family = %q, want %q", font.Family, Family)
	}
	if len(font.Faces) == 0 {
		t.Fatal("expected embedded Material Symbols faces")
	}
}

func TestFontReturnsCopy(t *testing.T) {
	first := Font()
	second := Font()
	if len(first.Faces) == 0 || len(second.Faces) == 0 {
		t.Fatal("expected font faces")
	}
	first.Faces[0] = theme.FontFace{}
	if second.Faces[0] == (theme.FontFace{}) {
		t.Fatal("Font should return an independent copy")
	}
}
