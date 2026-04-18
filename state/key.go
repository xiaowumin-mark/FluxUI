package state

import "github.com/xiaowumin-mark/FluxUI/internal"

func nextKey(ctx *internal.Context) string {
	return ctx.NextKey("state")
}
