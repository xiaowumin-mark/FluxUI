package theme

import "testing"

func TestIconFontRegistrationUsesLastDefault(t *testing.T) {
	iconFontRegistryMu.Lock()
	oldRegistry := iconFontRegistry
	iconFontRegistry = nil
	iconFontRegistryMu.Unlock()
	defer func() {
		iconFontRegistryMu.Lock()
		iconFontRegistry = oldRegistry
		iconFontRegistryMu.Unlock()
	}()

	RegisterIconFont(IconFont{ID: "a", Family: "Icons A", Default: true})
	RegisterIconFont(IconFont{ID: "b", Family: "Icons B", Default: true})

	font, ok := DefaultIconFont()
	if !ok {
		t.Fatal("expected default icon font")
	}
	if font.ID != "b" || font.Family != "Icons B" {
		t.Fatalf("default icon font = %#v, want b / Icons B", font)
	}

	font, ok = LookupIconFont("a")
	if !ok || font.Default {
		t.Fatalf("expected old default to be replaced, got %#v, %v", font, ok)
	}
}

func TestThemeIconFontsOverrideProcessRegistry(t *testing.T) {
	iconFontRegistryMu.Lock()
	oldRegistry := iconFontRegistry
	iconFontRegistry = nil
	iconFontRegistryMu.Unlock()
	defer func() {
		iconFontRegistryMu.Lock()
		iconFontRegistry = oldRegistry
		iconFontRegistryMu.Unlock()
	}()

	RegisterIconFont(IconFont{ID: "icons", Family: "Global Icons", Default: true})
	th := Default()
	th.AddIconFonts(IconFont{ID: "icons", Family: "Theme Icons", Default: true})

	font, ok := th.ResolveIconFont("icons")
	if !ok || font.Family != "Theme Icons" {
		t.Fatalf("theme ResolveIconFont = %#v, %v; want Theme Icons", font, ok)
	}

	font, ok = th.DefaultIconFont()
	if !ok || font.Family != "Theme Icons" {
		t.Fatalf("theme DefaultIconFont = %#v, %v; want Theme Icons", font, ok)
	}
}

func TestLoadIconFontFromBytesRejectsEmptyData(t *testing.T) {
	if _, err := LoadIconFontFromBytes("bad", "bad.ttf", nil); err == nil {
		t.Fatal("expected empty font data to fail")
	}
}

func TestIconFontCloneCopiesGlyphs(t *testing.T) {
	font := IconFont{
		ID:     "icons",
		Family: "Icons",
		Glyphs: map[string]rune{"home": '\ue9b2'},
	}
	cloned := cloneIconFont(font)
	cloned.Glyphs["home"] = '\ue88a'

	got, ok := font.ResolveIconText("home")
	if !ok || got != "\ue9b2" {
		t.Fatalf("original glyph map changed after clone mutation: got %q, %v", got, ok)
	}
}
