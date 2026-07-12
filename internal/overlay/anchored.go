// Package overlay contains UI-agnostic contracts shared by anchored overlays.
package overlay

import "image"

// AnchoredRegion describes the two protected surfaces of an anchored overlay:
// its trigger and its content. A press in either surface is an inside press;
// callers may safely treat all other presses as outside-dismiss candidates.
type AnchoredRegion struct {
	Anchor  image.Rectangle
	Content image.Rectangle
}

// ProtectedRects returns non-empty protected regions in trigger-first order.
func (r AnchoredRegion) ProtectedRects() []image.Rectangle {
	regions := make([]image.Rectangle, 0, 2)
	if !r.Anchor.Empty() {
		regions = append(regions, r.Anchor)
	}
	if !r.Content.Empty() {
		regions = append(regions, r.Content)
	}
	return regions
}

// Contains reports whether point falls inside the trigger or the overlay
// content. Rectangle upper bounds follow image.Rectangle's exclusive rule.
func (r AnchoredRegion) Contains(point image.Point) bool {
	for _, region := range r.ProtectedRects() {
		if point.In(region) {
			return true
		}
	}
	return false
}
