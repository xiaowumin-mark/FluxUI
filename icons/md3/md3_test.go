package md3

import (
	"bytes"
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

func TestEmbeddedWOFF2IsDecodedForShaper(t *testing.T) {
	if !bytes.HasPrefix(materialSymbolsOutlined, []byte("wOF2")) {
		t.Fatal("expected embedded Material Symbols source to be WOFF2")
	}
	font := Font()
	if len(font.Faces) == 0 {
		t.Fatal("expected font faces")
	}
	data := font.Faces[0].Data()
	if bytes.HasPrefix(data, []byte("wOF2")) || bytes.HasPrefix(data, []byte("wOFF")) {
		t.Fatal("expected registered font face data to be decoded SFNT, got compressed WOFF data")
	}
	if len(data) == 0 {
		t.Fatal("expected decoded font face data")
	}
}

func TestFontResolvesMaterialSymbolNamesToCodepoints(t *testing.T) {
	font := Font()
	cases := map[string]string{
		"home":     "\ue9b2",
		"search":   "\uef7a",
		"add":      "\ue145",
		"settings": "\ue8b8",
	}
	for name, want := range cases {
		got, ok := font.ResolveIconText(name)
		if !ok {
			t.Fatalf("ResolveIconText(%q) did not resolve", name)
		}
		if got != want {
			t.Fatalf("ResolveIconText(%q) = %q (%U), want %q (%U)", name, got, []rune(got), want, []rune(want))
		}
	}
	if got, ok := font.ResolveIconText("not_a_real_symbol"); ok {
		t.Fatalf("unexpected resolution for unknown symbol: %q", got)
	}
}
