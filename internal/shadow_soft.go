package internal

import (
	"image"
	"image/color"
	"math"
	"sync"

	"gioui.org/op/paint"
)

type softShadowCacheKey struct {
	w, h       int
	radius     int
	blur       int
	spread     int
	offX, offY int
	color      color.NRGBA
	circle     bool
}

type softShadowCacheEntry struct {
	op         paint.ImageOp
	padX, padY int
}

var softShadowCache = struct {
	sync.Mutex
	entries map[softShadowCacheKey]softShadowCacheEntry
}{
	entries: make(map[softShadowCacheKey]softShadowCacheEntry),
}

func softShadowEntry(size image.Point, radius, blur, spread, offX, offY int, col color.NRGBA, circle bool) softShadowCacheEntry {
	if size.X <= 0 || size.Y <= 0 || blur <= 0 || col.A == 0 {
		return softShadowCacheEntry{}
	}
	if blur > 96 {
		blur = 96
	}
	if spread < 0 {
		spread = 0
	}
	if radius < 0 {
		radius = 0
	}
	key := softShadowCacheKey{
		w: size.X, h: size.Y, radius: radius, blur: blur, spread: spread,
		offX: offX, offY: offY, color: col, circle: circle,
	}

	softShadowCache.Lock()
	if entry, ok := softShadowCache.entries[key]; ok {
		softShadowCache.Unlock()
		return entry
	}
	softShadowCache.Unlock()

	entry := buildSoftShadowEntry(size, radius, blur, spread, offX, offY, col, circle)

	softShadowCache.Lock()
	if len(softShadowCache.entries) > 192 {
		softShadowCache.entries = make(map[softShadowCacheKey]softShadowCacheEntry)
	}
	softShadowCache.entries[key] = entry
	softShadowCache.Unlock()
	return entry
}

func buildSoftShadowEntry(size image.Point, radius, blur, spread, offX, offY int, col color.NRGBA, circle bool) softShadowCacheEntry {
	padX := blur*2 + spread + absInt(offX) + 2
	padY := blur*2 + spread + absInt(offY) + 2
	if padX < 2 {
		padX = 2
	}
	if padY < 2 {
		padY = 2
	}
	w := size.X + padX*2
	h := size.Y + padY*2
	if w <= 0 || h <= 0 {
		return softShadowCacheEntry{}
	}

	alpha := make([]uint8, w*h)
	shape := image.Rect(padX+offX-spread, padY+offY-spread, padX+offX+size.X+spread, padY+offY+size.Y+spread)
	if circle {
		drawEllipseMask(alpha, w, h, shape, col.A)
	} else {
		drawRoundedRectMask(alpha, w, h, shape, radius+spread, col.A)
	}
	boxBlurAlpha(alpha, w, h, blur)

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		row := y * img.Stride
		for x := 0; x < w; x++ {
			a := alpha[y*w+x]
			if a == 0 {
				continue
			}
			i := row + x*4
			img.Pix[i+0] = premul(col.R, a)
			img.Pix[i+1] = premul(col.G, a)
			img.Pix[i+2] = premul(col.B, a)
			img.Pix[i+3] = a
		}
	}
	op := paint.NewImageOp(img)
	op.Filter = paint.FilterLinear
	return softShadowCacheEntry{op: op, padX: padX, padY: padY}
}

func drawRoundedRectMask(alpha []uint8, w, h int, rect image.Rectangle, radius int, a uint8) {
	if rect.Dx() <= 0 || rect.Dy() <= 0 || a == 0 {
		return
	}
	limit := rect.Dx()
	if rect.Dy() < limit {
		limit = rect.Dy()
	}
	if radius > limit/2 {
		radius = limit / 2
	}
	if radius <= 0 {
		fillMaskRect(alpha, w, h, rect, a)
		return
	}
	r := float64(radius)
	r2 := r * r
	for y := maxInt(rect.Min.Y, 0); y < minInt(rect.Max.Y, h); y++ {
		for x := maxInt(rect.Min.X, 0); x < minInt(rect.Max.X, w); x++ {
			cx := x
			if x < rect.Min.X+radius {
				cx = rect.Min.X + radius
			} else if x >= rect.Max.X-radius {
				cx = rect.Max.X - radius - 1
			}
			cy := y
			if y < rect.Min.Y+radius {
				cy = rect.Min.Y + radius
			} else if y >= rect.Max.Y-radius {
				cy = rect.Max.Y - radius - 1
			}
			dx := float64(x-cx) + 0.5
			dy := float64(y-cy) + 0.5
			if dx*dx+dy*dy <= r2 {
				alpha[y*w+x] = a
			}
		}
	}
}

func drawEllipseMask(alpha []uint8, w, h int, rect image.Rectangle, a uint8) {
	if rect.Dx() <= 0 || rect.Dy() <= 0 || a == 0 {
		return
	}
	cx := float64(rect.Min.X+rect.Max.X) / 2
	cy := float64(rect.Min.Y+rect.Max.Y) / 2
	rx := float64(rect.Dx()) / 2
	ry := float64(rect.Dy()) / 2
	if rx <= 0 || ry <= 0 {
		return
	}
	for y := maxInt(rect.Min.Y, 0); y < minInt(rect.Max.Y, h); y++ {
		for x := maxInt(rect.Min.X, 0); x < minInt(rect.Max.X, w); x++ {
			dx := (float64(x) + 0.5 - cx) / rx
			dy := (float64(y) + 0.5 - cy) / ry
			if dx*dx+dy*dy <= 1 {
				alpha[y*w+x] = a
			}
		}
	}
}

func fillMaskRect(alpha []uint8, w, h int, rect image.Rectangle, a uint8) {
	for y := maxInt(rect.Min.Y, 0); y < minInt(rect.Max.Y, h); y++ {
		for x := maxInt(rect.Min.X, 0); x < minInt(rect.Max.X, w); x++ {
			alpha[y*w+x] = a
		}
	}
}

func boxBlurAlpha(alpha []uint8, w, h, radius int) {
	if radius <= 0 || w <= 0 || h <= 0 {
		return
	}
	passes := 3
	boxRadius := int(math.Ceil(float64(radius) / 2.0))
	if boxRadius < 1 {
		boxRadius = 1
	}
	tmp := make([]uint8, len(alpha))
	for i := 0; i < passes; i++ {
		blurHorizontal(alpha, tmp, w, h, boxRadius)
		blurVertical(tmp, alpha, w, h, boxRadius)
	}
}

func blurHorizontal(src, dst []uint8, w, h, r int) {
	window := r*2 + 1
	for y := 0; y < h; y++ {
		sum := 0
		row := y * w
		for x := -r; x <= r; x++ {
			sum += int(src[row+clampInt(x, 0, w-1)])
		}
		for x := 0; x < w; x++ {
			dst[row+x] = uint8(sum / window)
			remove := clampInt(x-r, 0, w-1)
			add := clampInt(x+r+1, 0, w-1)
			sum += int(src[row+add]) - int(src[row+remove])
		}
	}
}

func blurVertical(src, dst []uint8, w, h, r int) {
	window := r*2 + 1
	for x := 0; x < w; x++ {
		sum := 0
		for y := -r; y <= r; y++ {
			sum += int(src[clampInt(y, 0, h-1)*w+x])
		}
		for y := 0; y < h; y++ {
			dst[y*w+x] = uint8(sum / window)
			remove := clampInt(y-r, 0, h-1)
			add := clampInt(y+r+1, 0, h-1)
			sum += int(src[add*w+x]) - int(src[remove*w+x])
		}
	}
}

func premul(c, a uint8) uint8 {
	return uint8((uint16(c)*uint16(a) + 127) / 255)
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
