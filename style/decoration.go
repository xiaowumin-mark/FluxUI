package style

import "image/color"

// Decoration describes optional visual surface properties (background,
// padding, radius) that may be applied to any widget. Nil pointer fields
// mean "use the widget's default", which typically resolves to Theme
// values or per-widget hardcoded defaults.
//
// Use the Bg/Pad/Rad constructors in ui/ or the With* methods to build
// one in a readable, chainable style:
//
//	ui.Bg(primary).Pad(ui.All(12)).Rad(8)
type Decoration struct {
	Background *color.NRGBA
	Padding    *Insets
	Radius     *float32
}

// WithBg sets the background colour.
func (d Decoration) WithBg(c color.NRGBA) Decoration {
	d.Background = &c
	return d
}

// WithPad sets the padding.
func (d Decoration) WithPad(p Insets) Decoration {
	d.Padding = &p
	return d
}

// WithRad sets the corner radius.
func (d Decoration) WithRad(r float32) Decoration {
	d.Radius = &r
	return d
}

// Merge returns a new Decoration where non-nil fields of other override
// the receiver. Useful for layering a base decoration with overrides.
func (d Decoration) Merge(other Decoration) Decoration {
	if other.Background != nil {
		d.Background = other.Background
	}
	if other.Padding != nil {
		d.Padding = other.Padding
	}
	if other.Radius != nil {
		d.Radius = other.Radius
	}
	return d
}

// ResolveBg returns d.Background if set, otherwise defaultBg.
func (d Decoration) ResolveBg(defaultBg color.NRGBA) color.NRGBA {
	if d.Background != nil {
		return *d.Background
	}
	return defaultBg
}

// ResolvePad returns d.Padding if set, otherwise defaultPad.
func (d Decoration) ResolvePad(defaultPad Insets) Insets {
	if d.Padding != nil {
		return *d.Padding
	}
	return defaultPad
}

// ResolveRad returns d.Radius if set, otherwise defaultRad.
func (d Decoration) ResolveRad(defaultRad float32) float32 {
	if d.Radius != nil {
		return *d.Radius
	}
	return defaultRad
}
