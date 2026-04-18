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
	if len(t.Fonts) == 0 {
		return nil, nil
	}

	out := make([]gioText.FontFace, 0, len(t.Fonts))
	for _, face := range t.Fonts {
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
