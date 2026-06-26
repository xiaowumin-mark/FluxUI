package theme

import (
	"errors"
	"strings"
	"sync"
)

// IconFont describes one font family that can render icon ligatures.
type IconFont struct {
	ID      string
	Family  string
	Faces   []FontFace
	Default bool
}

// IconFontOption configures an icon font loaded through helper APIs.
type IconFontOption func(*IconFont)

var (
	iconFontRegistryMu sync.RWMutex
	iconFontRegistry   []IconFont
)

// IconFontFamilyName overrides the family used when rendering icons.
func IconFontFamilyName(family string) IconFontOption {
	return func(font *IconFont) {
		font.Family = strings.TrimSpace(family)
	}
}

// IconFontDefault marks an icon font as the default icon font.
func IconFontDefault(defaulted bool) IconFontOption {
	return func(font *IconFont) {
		font.Default = defaulted
	}
}

// LoadIconFontFromPath loads an icon font from a ttf/otf/ttc/otc file.
func LoadIconFontFromPath(id string, path string, opts ...IconFontOption) (IconFont, error) {
	faces, err := ParseFontFile(path)
	if err != nil {
		return IconFont{}, err
	}
	return newIconFont(id, faces, opts...)
}

// LoadIconFontFromBytes loads an icon font from embedded ttf/otf/ttc/otc data.
func LoadIconFontFromBytes(id string, name string, data []byte, opts ...IconFontOption) (IconFont, error) {
	faces, err := ParseFontBytes(name, data)
	if err != nil {
		return IconFont{}, err
	}
	return newIconFont(id, faces, opts...)
}

// RegisterIconFont adds a process-wide icon font. Side-effect packages use this
// in init so applications only need to import the package.
func RegisterIconFont(font IconFont) {
	font = normalizeIconFont(font)
	if font.ID == "" || font.Family == "" {
		return
	}

	iconFontRegistryMu.Lock()
	defer iconFontRegistryMu.Unlock()
	registerIconFontLocked(&iconFontRegistry, font)
}

// RegisteredIconFonts returns the process-wide registered icon fonts.
func RegisteredIconFonts() []IconFont {
	iconFontRegistryMu.RLock()
	defer iconFontRegistryMu.RUnlock()
	return cloneIconFonts(iconFontRegistry)
}

// LookupIconFont returns a process-wide icon font by ID.
func LookupIconFont(id string) (IconFont, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return IconFont{}, false
	}
	iconFontRegistryMu.RLock()
	defer iconFontRegistryMu.RUnlock()
	return lookupIconFont(iconFontRegistry, id)
}

// DefaultIconFont returns the process-wide default icon font.
func DefaultIconFont() (IconFont, bool) {
	iconFontRegistryMu.RLock()
	defer iconFontRegistryMu.RUnlock()
	if font, ok := explicitDefaultIconFont(iconFontRegistry); ok {
		return font, true
	}
	if len(iconFontRegistry) > 0 {
		return cloneIconFont(iconFontRegistry[0]), true
	}
	return IconFont{}, false
}

// AddIconFonts appends application-scoped icon fonts to the theme.
func (t *Theme) AddIconFonts(fonts ...IconFont) {
	if t == nil || len(fonts) == 0 {
		return
	}
	for _, font := range fonts {
		font = normalizeIconFont(font)
		if font.ID == "" || font.Family == "" {
			continue
		}
		registerIconFontLocked(&t.IconFonts, font)
	}
}

// SetIconFonts replaces application-scoped icon fonts on the theme.
func (t *Theme) SetIconFonts(fonts ...IconFont) {
	if t == nil {
		return
	}
	t.IconFonts = nil
	t.AddIconFonts(fonts...)
}

// SetDefaultIconFont marks one application-scoped icon font as default.
func (t *Theme) SetDefaultIconFont(id string) {
	if t == nil {
		return
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return
	}
	for idx := range t.IconFonts {
		t.IconFonts[idx].Default = iconFontIDEqual(t.IconFonts[idx].ID, id)
	}
}

// ResolveIconFont returns an application-scoped or process-wide icon font by ID.
func (t *Theme) ResolveIconFont(id string) (IconFont, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return IconFont{}, false
	}
	if t != nil {
		if font, ok := lookupIconFont(t.IconFonts, id); ok {
			return font, true
		}
	}
	return LookupIconFont(id)
}

// DefaultIconFont returns the effective default icon font for this theme.
func (t *Theme) DefaultIconFont() (IconFont, bool) {
	if t != nil {
		if font, ok := explicitDefaultIconFont(t.IconFonts); ok {
			return font, true
		}
	}
	if font, ok := DefaultIconFont(); ok {
		return font, true
	}
	if t != nil && len(t.IconFonts) > 0 {
		return cloneIconFont(t.IconFonts[0]), true
	}
	return IconFont{}, false
}

func newIconFont(id string, faces []FontFace, opts ...IconFontOption) (IconFont, error) {
	font := IconFont{
		ID:    strings.TrimSpace(id),
		Faces: append([]FontFace(nil), faces...),
	}
	for _, opt := range opts {
		opt(&font)
	}
	font = normalizeIconFont(font)
	if font.ID == "" {
		return IconFont{}, errors.New("theme: empty icon font id")
	}
	if font.Family == "" {
		return IconFont{}, errors.New("theme: icon font has no family")
	}
	if len(font.Faces) == 0 {
		return IconFont{}, errors.New("theme: icon font has no faces")
	}
	return font, nil
}

func normalizeIconFont(font IconFont) IconFont {
	font.ID = strings.TrimSpace(font.ID)
	font.Family = strings.TrimSpace(font.Family)
	font.Faces = append([]FontFace(nil), font.Faces...)

	if font.Family == "" {
		for _, face := range font.Faces {
			if family := strings.TrimSpace(face.Spec.Family); family != "" {
				font.Family = family
				break
			}
		}
	}
	if font.ID == "" {
		font.ID = defaultIconFontID(font.Family)
	}
	if font.Family != "" {
		for idx := range font.Faces {
			font.Faces[idx] = font.Faces[idx].WithFamily(font.Family)
		}
	}
	return font
}

func defaultIconFontID(family string) string {
	family = strings.ToLower(strings.TrimSpace(family))
	family = strings.ReplaceAll(family, " ", "-")
	return family
}

func registerIconFontLocked(fonts *[]IconFont, font IconFont) {
	if font.Default {
		for idx := range *fonts {
			(*fonts)[idx].Default = false
		}
	}
	for idx := range *fonts {
		if iconFontIDEqual((*fonts)[idx].ID, font.ID) {
			(*fonts)[idx] = cloneIconFont(font)
			return
		}
	}
	*fonts = append(*fonts, cloneIconFont(font))
}

func lookupIconFont(fonts []IconFont, id string) (IconFont, bool) {
	for _, font := range fonts {
		if iconFontIDEqual(font.ID, id) {
			return cloneIconFont(font), true
		}
	}
	return IconFont{}, false
}

func explicitDefaultIconFont(fonts []IconFont) (IconFont, bool) {
	for idx := len(fonts) - 1; idx >= 0; idx-- {
		if fonts[idx].Default {
			return cloneIconFont(fonts[idx]), true
		}
	}
	return IconFont{}, false
}

func cloneIconFonts(fonts []IconFont) []IconFont {
	if len(fonts) == 0 {
		return nil
	}
	out := make([]IconFont, len(fonts))
	for idx, font := range fonts {
		out[idx] = cloneIconFont(font)
	}
	return out
}

func cloneIconFont(font IconFont) IconFont {
	font.Faces = append([]FontFace(nil), font.Faces...)
	return font
}

func iconFontIDEqual(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}
