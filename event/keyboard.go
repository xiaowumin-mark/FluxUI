package event

import "fluxui/internal"

// KeyHandler 处理键盘事件。
type KeyHandler func(ctx *internal.Context, key string)
