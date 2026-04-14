package layout

import (
	"image"

	"fluxui/internal"
)

// Constraints 是 FluxUI 的约束抽象。
type Constraints struct {
	Min image.Point
	Max image.Point
}

// FromContext 读取当前上下文约束。
func FromContext(ctx *internal.Context) Constraints {
	return Constraints{
		Min: ctx.MinConstraints(),
		Max: ctx.MaxConstraints(),
	}
}
