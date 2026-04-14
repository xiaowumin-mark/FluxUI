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

// IsZero 返回是否为零边距。
func (i Insets) IsZero() bool {
	return i.Top == 0 && i.Right == 0 && i.Bottom == 0 && i.Left == 0
}
