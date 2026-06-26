package theme

import (
	"errors"
	"strings"

	"gioui.org/font/gofont"
	"gioui.org/font/opentype"
	gioText "gioui.org/text"
)

// BuildShaper 从 Theme 构建文本 Shaper。
func (t *Theme) BuildShaper() (*gioText.Shaper, error) {
	if t == nil {
		return nil, errors.New("theme: nil theme")
	}

	collection, err := t.resolveTextCollection()
	if err != nil {
		return nil, err
	}
	if len(collection) == 0 {
		collection = gofont.Collection()
	}

	opts := make([]gioText.ShaperOption, 0, 2)
	if !t.UseSystemFonts {
		opts = append(opts, gioText.NoSystemFonts())
	}
	if len(collection) > 0 {
		opts = append(opts, gioText.WithCollection(collection))
	}
	return gioText.NewShaper(opts...), nil
}

func (t *Theme) resolveTextCollection() ([]gioText.FontFace, error) {
	faces := t.textFontFaces()
	if len(faces) == 0 {
		return nil, nil
	}

	out := make([]gioText.FontFace, 0, len(faces))
	for _, face := range faces {
		if face.source == nil || len(face.source.data) == 0 {
			continue
		}
		parsed, err := opentype.ParseCollection(face.source.data)
		if err != nil {
			continue
		}
		if face.faceIndex < 0 || face.faceIndex >= len(parsed) {
			continue
		}
		raw := parsed[face.faceIndex]
		spec := face.Spec.Normalize()
		if strings.TrimSpace(spec.Family) != "" {
			raw.Font = toGioFont(spec)
		}
		out = append(out, raw)
	}
	return out, nil
}

func (t *Theme) textFontFaces() []FontFace {
	if t == nil {
		return nil
	}
	total := len(t.Fonts)
	for _, font := range t.IconFonts {
		total += len(font.Faces)
	}
	for _, font := range RegisteredIconFonts() {
		total += len(font.Faces)
	}
	if total == 0 {
		return nil
	}

	out := make([]FontFace, 0, total)
	out = append(out, t.Fonts...)
	for _, font := range t.IconFonts {
		out = append(out, font.Faces...)
	}
	for _, font := range RegisteredIconFonts() {
		out = append(out, font.Faces...)
	}
	return out
}
