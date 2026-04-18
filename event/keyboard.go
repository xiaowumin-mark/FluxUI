package event

import "github.com/xiaowumin-mark/FluxUI/internal"

// KeyHandler 处理键盘事件。
type KeyHandler func(ctx *internal.Context, key string)
