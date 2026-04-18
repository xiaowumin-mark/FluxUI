package widget

import (
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
)

// Widget 是所有 FluxUI 组件的统一接口。
type Widget interface {
	Layout(ctx *internal.Context) layout.Dimensions
}
