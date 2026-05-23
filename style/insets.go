package style

// Insets 描述四个方向的边距。
type Insets struct {
	Top    float32
	Right  float32
	Bottom float32
	Left   float32
}

// All 为四边设置相同值。
func All(v float32) Insets {
	return Insets{
		Top:    v,
		Right:  v,
		Bottom: v,
		Left:   v,
	}
}

// Symmetric 设置垂直与水平边距。
func Symmetric(vertical, horizontal float32) Insets {
	return Insets{
		Top:    vertical,
		Right:  horizontal,
		Bottom: vertical,
		Left:   horizontal,
	}
}

// Only 为四个边分别设置不同值。
func Only(top, right, bottom, left float32) Insets {
	return Insets{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
}

// Horizontal 设置左右水平边距。
func Horizontal(v float32) Insets {
	return Insets{Right: v, Left: v}
}

// Vertical 设置上下垂直边距。
func Vertical(v float32) Insets {
	return Insets{Top: v, Bottom: v}
}

// IsZero 返回是否为零边距。
func (i Insets) IsZero() bool {
	return i.Top == 0 && i.Right == 0 && i.Bottom == 0 && i.Left == 0
}
