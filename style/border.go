package style

import "image/color"

// Border 描述边框样式。
type Border struct {
	Width float32
	Color color.NRGBA
}

// IsZero 返回边框是否不可见。
func (b Border) IsZero() bool {
	return b.Width <= 0 || b.Color.A == 0
}
