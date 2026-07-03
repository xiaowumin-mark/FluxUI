package style

// CornerShape describes the shape applied to a rounded corner.
//
// It follows the CSS corner-shape keyword model. A non-round shape is visible
// only when the decoration also has a non-zero radius.
type CornerShape uint8

const (
	CornerRound CornerShape = iota
	CornerSquare
	CornerBevel
	CornerNotch
	CornerScoop
	CornerSquircle
)

// CornerShapes stores per-corner shapes in CSS order:
// top-left, top-right, bottom-right, bottom-left.
type CornerShapes struct {
	TopLeft     CornerShape
	TopRight    CornerShape
	BottomRight CornerShape
	BottomLeft  CornerShape
}

func UniformCornerShape(shape CornerShape) CornerShapes {
	return CornerShapes{
		TopLeft:     shape,
		TopRight:    shape,
		BottomRight: shape,
		BottomLeft:  shape,
	}
}

func (s CornerShapes) IsZero() bool {
	return s.TopLeft == CornerRound &&
		s.TopRight == CornerRound &&
		s.BottomRight == CornerRound &&
		s.BottomLeft == CornerRound
}
