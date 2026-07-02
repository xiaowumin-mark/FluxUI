package internal

import (
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"math"
	"reflect"
	"strconv"

	theme "github.com/xiaowumin-mark/FluxUI/theme"

	"gioui.org/op"
)

type runtimeRenderCache struct {
	text        map[textLayoutCacheKey]*textLayoutCacheEntry
	activeText  map[textLayoutCacheKey]struct{}
	staticPaint map[staticPaintCacheKey]*staticPaintCacheEntry
	activePaint map[staticPaintCacheKey]struct{}
	staticTree  map[staticSubtreeCacheKey]*staticSubtreeCacheEntry
	activeTree  map[staticSubtreeCacheKey]struct{}
}

type textLayoutCacheEntry struct {
	ops  *op.Ops
	call op.CallOp
	size image.Point
}

type staticPaintCacheEntry struct {
	ops  *op.Ops
	call op.CallOp
}

type staticSubtreeCacheEntry struct {
	ops  *op.Ops
	call op.CallOp
	size image.Point
}

type textLayoutCacheKey struct {
	content     string
	fontFamily  string
	fontStyle   theme.FontStyle
	fontWeight  theme.FontWeight
	sizeBits    uint32
	lineBits    uint32
	color       color.NRGBA
	alignment   Alignment
	minX        int
	minY        int
	maxX        int
	maxY        int
	pxPerDpBits uint32
	pxPerSpBits uint32
	locale      string
	direction   uint8
}

type staticPaintCacheKey struct {
	circle       bool
	width        int
	height       int
	radiusBits   uint32
	background   color.NRGBA
	borderColor  color.NRGBA
	borderBits   uint32
	gradient     bool
	gradientFrom color.NRGBA
	gradientTo   color.NRGBA
	gradStartX   uint32
	gradStartY   uint32
	gradEndX     uint32
	gradEndY     uint32
	pxPerDpBits  uint32
}

type staticSubtreeCacheKey struct {
	path        MemoryKey
	depsHash    uint64
	themePtr    uintptr
	foreground  color.NRGBA
	fontFamily  string
	fontStyle   theme.FontStyle
	fontWeight  theme.FontWeight
	minX        int
	minY        int
	maxX        int
	maxY        int
	pxPerDpBits uint32
	pxPerSpBits uint32
	locale      string
	direction   uint8
}

func (r *Runtime) beginRenderCacheFrame() {
	if r == nil {
		return
	}
	r.ensureRenderCache()
	clear(r.render.activeText)
	clear(r.render.activePaint)
	clear(r.render.activeTree)
}

func (r *Runtime) endRenderCacheFrame() {
	if r == nil {
		return
	}
	for key := range r.render.text {
		if _, ok := r.render.activeText[key]; !ok {
			delete(r.render.text, key)
		}
	}
	for key := range r.render.staticPaint {
		if _, ok := r.render.activePaint[key]; !ok {
			delete(r.render.staticPaint, key)
		}
	}
	for key := range r.render.staticTree {
		if _, ok := r.render.activeTree[key]; !ok {
			delete(r.render.staticTree, key)
		}
	}
}

func (r *Runtime) disposeRenderCache() {
	if r == nil {
		return
	}
	clear(r.render.text)
	clear(r.render.activeText)
	clear(r.render.staticPaint)
	clear(r.render.activePaint)
	clear(r.render.staticTree)
	clear(r.render.activeTree)
}

func (r *Runtime) ensureRenderCache() {
	if r.render.text == nil {
		r.render.text = make(map[textLayoutCacheKey]*textLayoutCacheEntry)
	}
	if r.render.activeText == nil {
		r.render.activeText = make(map[textLayoutCacheKey]struct{})
	}
	if r.render.staticPaint == nil {
		r.render.staticPaint = make(map[staticPaintCacheKey]*staticPaintCacheEntry)
	}
	if r.render.activePaint == nil {
		r.render.activePaint = make(map[staticPaintCacheKey]struct{})
	}
	if r.render.staticTree == nil {
		r.render.staticTree = make(map[staticSubtreeCacheKey]*staticSubtreeCacheEntry)
	}
	if r.render.activeTree == nil {
		r.render.activeTree = make(map[staticSubtreeCacheKey]struct{})
	}
}

func (r *Runtime) lookupTextLayoutCache(key textLayoutCacheKey) (*textLayoutCacheEntry, bool) {
	if r == nil {
		return nil, false
	}
	r.ensureRenderCache()
	r.render.activeText[key] = struct{}{}
	entry, ok := r.render.text[key]
	return entry, ok && entry != nil
}

func (r *Runtime) storeTextLayoutCache(key textLayoutCacheKey, ops *op.Ops, call op.CallOp, size image.Point) {
	if r == nil || ops == nil {
		return
	}
	r.ensureRenderCache()
	r.render.activeText[key] = struct{}{}
	r.render.text[key] = &textLayoutCacheEntry{
		ops:  ops,
		call: call,
		size: size,
	}
}

func (r *Runtime) lookupStaticPaintCache(key staticPaintCacheKey) (*staticPaintCacheEntry, bool) {
	if r == nil {
		return nil, false
	}
	r.ensureRenderCache()
	r.render.activePaint[key] = struct{}{}
	entry, ok := r.render.staticPaint[key]
	return entry, ok && entry != nil
}

func (r *Runtime) storeStaticPaintCache(key staticPaintCacheKey, ops *op.Ops, call op.CallOp) {
	if r == nil || ops == nil {
		return
	}
	r.ensureRenderCache()
	r.render.activePaint[key] = struct{}{}
	r.render.staticPaint[key] = &staticPaintCacheEntry{
		ops:  ops,
		call: call,
	}
}

func (r *Runtime) lookupStaticSubtreeCache(key staticSubtreeCacheKey) (*staticSubtreeCacheEntry, bool) {
	if r == nil {
		return nil, false
	}
	r.ensureRenderCache()
	r.render.activeTree[key] = struct{}{}
	entry, ok := r.render.staticTree[key]
	return entry, ok && entry != nil
}

func (r *Runtime) storeStaticSubtreeCache(key staticSubtreeCacheKey, ops *op.Ops, call op.CallOp, size image.Point) {
	if r == nil || ops == nil {
		return
	}
	r.ensureRenderCache()
	r.render.activeTree[key] = struct{}{}
	r.render.staticTree[key] = &staticSubtreeCacheEntry{
		ops:  ops,
		call: call,
		size: size,
	}
}

func (c *Context) textLayoutCacheKey(spec TextSpec, font theme.FontSpec, size float32) textLayoutCacheKey {
	gtx := c.Gtx
	return textLayoutCacheKey{
		content:     spec.Content,
		fontFamily:  font.Family,
		fontStyle:   font.Style,
		fontWeight:  font.Weight,
		sizeBits:    math.Float32bits(size),
		lineBits:    math.Float32bits(spec.LineHeight),
		color:       spec.Color,
		alignment:   spec.Alignment,
		minX:        gtx.Constraints.Min.X,
		minY:        gtx.Constraints.Min.Y,
		maxX:        gtx.Constraints.Max.X,
		maxY:        gtx.Constraints.Max.Y,
		pxPerDpBits: math.Float32bits(float32(gtx.Metric.PxPerDp)),
		pxPerSpBits: math.Float32bits(float32(gtx.Metric.PxPerSp)),
		locale:      gtx.Locale.Language,
		direction:   uint8(gtx.Locale.Direction),
	}
}

func (c *Context) staticSubtreeCacheKey(depsHash uint64) staticSubtreeCacheKey {
	gtx := c.Gtx
	font := c.Font()
	var themePtr uintptr
	if th := c.Theme(); th != nil {
		themePtr = reflect.ValueOf(th).Pointer()
	}
	return staticSubtreeCacheKey{
		path:        c.ScopeMemoryKey("static-subtree"),
		depsHash:    depsHash,
		themePtr:    themePtr,
		foreground:  c.Foreground(),
		fontFamily:  font.Family,
		fontStyle:   font.Style,
		fontWeight:  font.Weight,
		minX:        gtx.Constraints.Min.X,
		minY:        gtx.Constraints.Min.Y,
		maxX:        gtx.Constraints.Max.X,
		maxY:        gtx.Constraints.Max.Y,
		pxPerDpBits: math.Float32bits(float32(gtx.Metric.PxPerDp)),
		pxPerSpBits: math.Float32bits(float32(gtx.Metric.PxPerSp)),
		locale:      gtx.Locale.Language,
		direction:   uint8(gtx.Locale.Direction),
	}
}

func HashStaticDeps(deps ...any) uint64 {
	if len(deps) == 0 {
		return 0
	}
	h := fnv.New64a()
	for _, dep := range deps {
		if dep == nil {
			_, _ = h.Write([]byte("<nil>;"))
			continue
		}
		_, _ = h.Write([]byte(reflect.TypeOf(dep).String()))
		_, _ = h.Write([]byte("="))
		switch v := dep.(type) {
		case string:
			_, _ = h.Write([]byte(v))
		case int:
			_, _ = h.Write([]byte(strconv.Itoa(v)))
		case int64:
			_, _ = h.Write([]byte(strconv.FormatInt(v, 10)))
		case uint64:
			_, _ = h.Write([]byte(strconv.FormatUint(v, 10)))
		case bool:
			_, _ = h.Write([]byte(strconv.FormatBool(v)))
		case float32:
			_, _ = h.Write([]byte(strconv.FormatUint(uint64(math.Float32bits(v)), 16)))
		case float64:
			_, _ = h.Write([]byte(strconv.FormatUint(math.Float64bits(v), 16)))
		default:
			_, _ = fmt.Fprintf(h, "%#v", dep)
		}
		_, _ = h.Write([]byte(";"))
	}
	return h.Sum64()
}

func staticPaintCacheable(spec SurfaceSpec, size image.Point) bool {
	if size.X <= 0 || size.Y <= 0 || spec.HasImage {
		return false
	}
	return spec.HasGradient || spec.Background.A > 0 || (spec.BorderWidth > 0 && spec.BorderColor.A > 0)
}

func staticPaintCacheKeyFor(spec SurfaceSpec, size image.Point, circle bool, pxPerDp float32) staticPaintCacheKey {
	return staticPaintCacheKey{
		circle:       circle,
		width:        size.X,
		height:       size.Y,
		radiusBits:   math.Float32bits(spec.Radius),
		background:   spec.Background,
		borderColor:  spec.BorderColor,
		borderBits:   math.Float32bits(spec.BorderWidth),
		gradient:     spec.HasGradient,
		gradientFrom: spec.GradientFrom,
		gradientTo:   spec.GradientTo,
		gradStartX:   math.Float32bits(spec.GradientStart.X),
		gradStartY:   math.Float32bits(spec.GradientStart.Y),
		gradEndX:     math.Float32bits(spec.GradientEnd.X),
		gradEndY:     math.Float32bits(spec.GradientEnd.Y),
		pxPerDpBits:  math.Float32bits(pxPerDp),
	}
}
