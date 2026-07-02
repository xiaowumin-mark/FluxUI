package state

import "github.com/xiaowumin-mark/FluxUI/internal"

func nextKey(ctx *internal.Context) internal.MemoryKey {
	return ctx.NextMemoryKey("state")
}
