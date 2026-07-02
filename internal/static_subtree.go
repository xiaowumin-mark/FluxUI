package internal

import (
	"image"

	"gioui.org/op"
)

func (c *Context) LayoutStaticSubtree(depsHash uint64, child func(*Context) image.Point) image.Point {
	if c == nil || c.runtime == nil || child == nil {
		if child == nil {
			return image.Point{}
		}
		return child(c)
	}
	key := c.staticSubtreeCacheKey(depsHash)
	if entry, ok := c.runtime.lookupStaticSubtreeCache(key); ok {
		c.runtime.RecordStaticSubtreeCache(true)
		entry.call.Add(c.Gtx.Ops)
		return entry.size
	}
	c.runtime.RecordStaticSubtreeCache(false)

	recordOps := new(op.Ops)
	recordGtx := c.Gtx
	recordGtx.Ops = recordOps
	recordCtx := c.sameScope(recordGtx)
	macro := op.Record(recordOps)
	size := child(recordCtx)
	call := macro.Stop()
	c.runtime.storeStaticSubtreeCache(key, recordOps, call, size)
	call.Add(c.Gtx.Ops)
	return size
}
