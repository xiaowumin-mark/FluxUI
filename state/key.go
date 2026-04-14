package state

import "fluxui/internal"

func nextKey(ctx *internal.Context) string {
	return ctx.NextKey("state")
}
