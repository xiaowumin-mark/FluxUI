package md3

import (
	_ "embed"

	"github.com/xiaowumin-mark/FluxUI/theme"
)

const (
	// ID is the registered icon font ID for Material Design 3 symbols.
	ID = "md3"
	// Family is the font family used by Material Symbols Outlined.
	Family = "Material Symbols Outlined"
)

//go:embed MaterialSymbolsOutlined.woff2
var materialSymbolsOutlined []byte

var font theme.IconFont

func init() {
	var err error
	font, err = theme.LoadIconFontFromBytes(
		ID,
		"MaterialSymbolsOutlined.woff2",
		materialSymbolsOutlined,
		theme.IconFontFamilyName(Family),
		theme.IconFontDefault(true),
	)
	if err != nil {
		panic(err)
	}
	theme.RegisterIconFont(font)
}

// Font returns the embedded Material Design 3 icon font.
func Font() theme.IconFont {
	out := font
	out.Faces = append([]theme.FontFace(nil), font.Faces...)
	return out
}
