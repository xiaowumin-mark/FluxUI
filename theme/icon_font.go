package theme

import (
	"encoding/binary"
	"errors"
	"strings"
	"sync"

	sfntfont "github.com/tdewolff/font"
)

// IconFont describes one font family that can render icon ligatures.
type IconFont struct {
	ID      string
	Family  string
	Faces   []FontFace
	Glyphs  map[string]rune
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

// LoadIconFontFromPath loads an icon font from a ttf/otf/ttc/otc/woff/woff2 file.
func LoadIconFontFromPath(id string, path string, opts ...IconFontOption) (IconFont, error) {
	faces, err := ParseFontFile(path)
	if err != nil {
		return IconFont{}, err
	}
	return newIconFont(id, faces, opts...)
}

// LoadIconFontFromBytes loads an icon font from embedded ttf/otf/ttc/otc/woff/woff2 data.
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

// ResolveIconText returns the text that should be shaped for an icon name.
// Fonts with glyph-name cmap entries, such as Material Symbols WOFF2, can
// render by codepoint even when they do not include ligature substitutions.
func (font IconFont) ResolveIconText(name string) (string, bool) {
	name = strings.TrimSpace(name)
	if name == "" || len(font.Glyphs) == 0 {
		return "", false
	}
	if r, ok := font.Glyphs[name]; ok && r != 0 {
		return string(r), true
	}
	return "", false
}

func newIconFont(id string, faces []FontFace, opts ...IconFontOption) (IconFont, error) {
	font := IconFont{
		ID:     strings.TrimSpace(id),
		Faces:  append([]FontFace(nil), faces...),
		Glyphs: iconGlyphsFromFaces(faces),
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
	font.Glyphs = cloneIconGlyphs(font.Glyphs)

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
	font.Glyphs = cloneIconGlyphs(font.Glyphs)
	return font
}

func iconFontIDEqual(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

func iconGlyphsFromFaces(faces []FontFace) map[string]rune {
	var out map[string]rune
	for _, face := range faces {
		data := face.data()
		if len(data) == 0 {
			continue
		}
		parsed, err := sfntfont.ParseSFNT(data, face.FaceIndex())
		if err != nil {
			continue
		}
		unicodeByGlyphID := cmapUnicodeByGlyphID(parsed.Tables["cmap"])
		for glyphID := uint16(1); glyphID < parsed.NumGlyphs(); glyphID++ {
			name := strings.TrimSpace(parsed.GlyphName(glyphID))
			if name == "" || name == ".notdef" {
				continue
			}
			r, ok := unicodeByGlyphID[glyphID]
			if !ok || r == 0 {
				continue
			}
			if out == nil {
				out = make(map[string]rune, 256)
			}
			out[name] = r
		}
	}
	return out
}

func cloneIconGlyphs(glyphs map[string]rune) map[string]rune {
	if len(glyphs) == 0 {
		return nil
	}
	out := make(map[string]rune, len(glyphs))
	for name, r := range glyphs {
		out[name] = r
	}
	return out
}

func cmapUnicodeByGlyphID(cmap []byte) map[uint16]rune {
	if len(cmap) < 4 {
		return nil
	}
	numTables := int(binary.BigEndian.Uint16(cmap[2:4]))
	out := make(map[uint16]rune, 256)
	for i := 0; i < numTables; i++ {
		record := 4 + i*8
		if record+8 > len(cmap) {
			break
		}
		offset := int(binary.BigEndian.Uint32(cmap[record+4 : record+8]))
		if offset < 0 || offset+2 > len(cmap) {
			continue
		}
		switch binary.BigEndian.Uint16(cmap[offset : offset+2]) {
		case 4:
			parseCmapFormat4(out, cmap[offset:])
		case 12, 13:
			parseCmapFormat12Or13(out, cmap[offset:])
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func parseCmapFormat4(out map[uint16]rune, table []byte) {
	if len(table) < 16 {
		return
	}
	length := int(binary.BigEndian.Uint16(table[2:4]))
	if length > len(table) {
		length = len(table)
	}
	table = table[:length]
	segCount := int(binary.BigEndian.Uint16(table[6:8]) / 2)
	endCodes := 14
	startCodes := endCodes + segCount*2 + 2
	idDeltas := startCodes + segCount*2
	idRangeOffsets := idDeltas + segCount*2
	if idRangeOffsets+segCount*2 > len(table) {
		return
	}
	for i := 0; i < segCount; i++ {
		start := binary.BigEndian.Uint16(table[startCodes+i*2 : startCodes+i*2+2])
		end := binary.BigEndian.Uint16(table[endCodes+i*2 : endCodes+i*2+2])
		delta := binary.BigEndian.Uint16(table[idDeltas+i*2 : idDeltas+i*2+2])
		rangeOffsetPos := idRangeOffsets + i*2
		rangeOffset := binary.BigEndian.Uint16(table[rangeOffsetPos : rangeOffsetPos+2])
		if start == 0xffff && end == 0xffff {
			continue
		}
		for c := start; c <= end; c++ {
			var glyphID uint16
			if rangeOffset == 0 {
				glyphID = c + delta
			} else {
				pos := rangeOffsetPos + int(rangeOffset) + int(c-start)*2
				if pos+2 > len(table) {
					continue
				}
				glyphID = binary.BigEndian.Uint16(table[pos : pos+2])
				if glyphID != 0 {
					glyphID += delta
				}
			}
			if glyphID != 0 {
				out[glyphID] = rune(c)
			}
			if c == 0xffff {
				break
			}
		}
	}
}

func parseCmapFormat12Or13(out map[uint16]rune, table []byte) {
	if len(table) < 16 {
		return
	}
	length := int(binary.BigEndian.Uint32(table[4:8]))
	if length > len(table) {
		length = len(table)
	}
	table = table[:length]
	format := binary.BigEndian.Uint16(table[0:2])
	numGroups := int(binary.BigEndian.Uint32(table[12:16]))
	for i := 0; i < numGroups; i++ {
		group := 16 + i*12
		if group+12 > len(table) {
			break
		}
		startChar := binary.BigEndian.Uint32(table[group : group+4])
		endChar := binary.BigEndian.Uint32(table[group+4 : group+8])
		startGlyph := binary.BigEndian.Uint32(table[group+8 : group+12])
		if endChar < startChar || endChar-startChar > 100000 {
			continue
		}
		for c := startChar; c <= endChar; c++ {
			glyphID := startGlyph
			if format == 12 {
				glyphID += c - startChar
			}
			if glyphID > 0 && glyphID <= 0xffff && c <= 0x10ffff {
				out[uint16(glyphID)] = rune(c)
			}
		}
	}
}
