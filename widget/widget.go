package widget

import (
	"fluxui/internal"
	"fluxui/layout"
)

// Widget 是所有 FluxUI 组件的统一接口。
type Widget interface {
	Layout(ctx *internal.Context) layout.Dimensions
}
