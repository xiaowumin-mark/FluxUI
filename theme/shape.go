package theme

type ShapeScale struct {
	None       float32
	ExtraSmall float32
	Small      float32
	Medium     float32
	Large      float32
	ExtraLarge float32
	Full       float32
}

func DefaultShapeScale() ShapeScale {
	return ShapeScale{
		None:       0,
		ExtraSmall: 4,
		Small:      8,
		Medium:     12,
		Large:      16,
		ExtraLarge: 28,
		Full:       999,
	}
}
