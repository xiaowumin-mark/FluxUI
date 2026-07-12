// Package collection provides small, UI-agnostic contracts for keyed
// collections used by advanced components.
package collection

import (
	"fmt"
	"sort"
)

// Key identifies an item independently from its current position.
//
// A collection key must be non-empty and stable for the lifetime of the
// logical item. Callers must not derive it from a list index.
type Key string

// Item is the minimum collection metadata needed by the selection and roving
// focus contracts.
type Item struct {
	Key      Key
	Disabled bool
}

// Model is an immutable, keyed view of a collection. Construct it with New so
// duplicate and empty keys are rejected at the component boundary.
type Model struct {
	items []Item
	index map[Key]int
}

// New creates a validated keyed collection model.
func New(items []Item) (Model, error) {
	model := Model{
		items: append([]Item(nil), items...),
		index: make(map[Key]int, len(items)),
	}
	for index, item := range model.items {
		if item.Key == "" {
			return Model{}, fmt.Errorf("collection item %d has an empty key", index)
		}
		if previous, exists := model.index[item.Key]; exists {
			return Model{}, fmt.Errorf("collection key %q is duplicated at indexes %d and %d", item.Key, previous, index)
		}
		model.index[item.Key] = index
	}
	return model, nil
}

// Len returns the number of items in the model.
func (m Model) Len() int {
	return len(m.items)
}

// Items returns a copy of the model items in visual order.
func (m Model) Items() []Item {
	return append([]Item(nil), m.items...)
}

// Has reports whether key belongs to the collection.
func (m Model) Has(key Key) bool {
	_, ok := m.index[key]
	return ok
}

// Item returns the item metadata for key.
func (m Model) Item(key Key) (Item, bool) {
	index, ok := m.index[key]
	if !ok {
		return Item{}, false
	}
	return m.items[index], true
}

// Index returns the current visual index for key.
func (m Model) Index(key Key) (int, bool) {
	index, ok := m.index[key]
	return index, ok
}

// FirstEnabled returns the first non-disabled item in visual order.
func (m Model) FirstEnabled() (Key, bool) {
	for _, item := range m.items {
		if !item.Disabled {
			return item.Key, true
		}
	}
	return "", false
}

// LastEnabled returns the last non-disabled item in visual order.
func (m Model) LastEnabled() (Key, bool) {
	for index := len(m.items) - 1; index >= 0; index-- {
		if !m.items[index].Disabled {
			return m.items[index].Key, true
		}
	}
	return "", false
}

// NextEnabled returns the next enabled item after key. A positive direction
// moves forward; a negative direction moves backward. When wrap is false, the
// collection boundary is terminal.
func (m Model) NextEnabled(key Key, direction int, wrap bool) (Key, bool) {
	if len(m.items) == 0 || direction == 0 {
		return "", false
	}
	step := 1
	if direction < 0 {
		step = -1
	}
	index, found := m.index[key]
	if !found {
		if step > 0 {
			index = -1
		} else {
			index = len(m.items)
		}
	}

	for visited := 0; visited < len(m.items); visited++ {
		index += step
		if index < 0 || index >= len(m.items) {
			if !wrap {
				return "", false
			}
			if index < 0 {
				index = len(m.items) - 1
			} else {
				index = 0
			}
		}
		if !m.items[index].Disabled {
			return m.items[index].Key, true
		}
	}
	return "", false
}

// Selection stores a set of selected keyed items. Its methods return copies so
// controlled component state cannot be changed through an aliased map.
type Selection struct {
	keys map[Key]struct{}
}

// NewSelection creates a selection. Empty keys are ignored because they are
// never valid collection identities.
func NewSelection(keys ...Key) Selection {
	selection := Selection{keys: make(map[Key]struct{}, len(keys))}
	for _, key := range keys {
		if key != "" {
			selection.keys[key] = struct{}{}
		}
	}
	return selection
}

// Empty reports whether no keys are selected.
func (s Selection) Empty() bool {
	return len(s.keys) == 0
}

// Contains reports whether key is selected.
func (s Selection) Contains(key Key) bool {
	_, ok := s.keys[key]
	return ok
}

// Keys returns selected keys in deterministic lexical order. Use Ordered for
// the collection's current visual order.
func (s Selection) Keys() []Key {
	keys := make([]Key, 0, len(s.keys))
	for key := range s.keys {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(left, right int) bool { return keys[left] < keys[right] })
	return keys
}

// Ordered returns selected keys in the model's current visual order.
func (s Selection) Ordered(model Model) []Key {
	keys := make([]Key, 0, len(s.keys))
	for _, item := range model.items {
		if s.Contains(item.Key) {
			keys = append(keys, item.Key)
		}
	}
	return keys
}

// With returns a selection that also contains key.
func (s Selection) With(key Key) Selection {
	if key == "" {
		return s.copy()
	}
	next := s.copy()
	next.keys[key] = struct{}{}
	return next
}

// Without returns a selection without key.
func (s Selection) Without(key Key) Selection {
	next := s.copy()
	delete(next.keys, key)
	return next
}

// Toggle returns a selection with key added or removed.
func (s Selection) Toggle(key Key) Selection {
	if key == "" {
		return s.copy()
	}
	if s.Contains(key) {
		return s.Without(key)
	}
	return s.With(key)
}

// Reconcile removes keys that no longer exist in model. Disabled items remain
// selected: availability and value validity are separate contracts.
func (s Selection) Reconcile(model Model) Selection {
	next := NewSelection()
	for key := range s.keys {
		if model.Has(key) {
			next.keys[key] = struct{}{}
		}
	}
	return next
}

func (s Selection) copy() Selection {
	next := NewSelection()
	for key := range s.keys {
		next.keys[key] = struct{}{}
	}
	return next
}

// RovingFocus tracks one active item by key. It deliberately stores no index,
// so reordering a collection cannot transfer focus to another logical item.
type RovingFocus struct {
	Key  Key
	Wrap bool

	// neighbors records the last visual neighbors of Key by stable key only.
	// It lets Reconcile preserve the next-then-previous fallback rule when Key
	// has been removed from the new model, where an index is unavailable.
	neighbors *rovingNeighbors
}

type rovingNeighbors struct {
	active    Key
	following []Key
	preceding []Key
}

// Reconcile keeps an enabled active item. When the current active item is
// disabled, it first chooses the next enabled item in the current model, then
// the preceding one. When it was removed, it uses the stable-key neighbors
// captured while it was last present, in the same next-then-previous order.
// An empty active key initializes to the first enabled item; a non-empty key
// that was never observed in a model is cleared because it has no position
// from which to infer a neighbor.
func (r RovingFocus) Reconcile(model Model) RovingFocus {
	item, present := model.Item(r.Key)
	if present && !item.Disabled {
		return r.rememberNeighbors(model)
	}

	if present {
		if key, ok := model.NextEnabled(r.Key, 1, false); ok {
			return r.setActive(model, key)
		}
		if key, ok := model.NextEnabled(r.Key, -1, false); ok {
			return r.setActive(model, key)
		}
	} else {
		if key, ok := r.neighbors.firstEnabled(model, r.Key, true); ok {
			return r.setActive(model, key)
		}
		if key, ok := r.neighbors.firstEnabled(model, r.Key, false); ok {
			return r.setActive(model, key)
		}
		if r.Key == "" {
			if key, ok := model.FirstEnabled(); ok {
				return r.setActive(model, key)
			}
		}
	}

	return RovingFocus{Wrap: r.Wrap}
}

// Move advances the active item. It returns whether the active key changed.
func (r RovingFocus) Move(model Model, direction int) (RovingFocus, bool) {
	next := r.Reconcile(model)
	key, ok := model.NextEnabled(next.Key, direction, next.Wrap)
	if !ok || key == next.Key {
		return next, next.Key != r.Key
	}
	return next.setActive(model, key), true
}

// Home moves to the first enabled item.
func (r RovingFocus) Home(model Model) (RovingFocus, bool) {
	key, ok := model.FirstEnabled()
	if !ok {
		return RovingFocus{Wrap: r.Wrap}, r.Key != ""
	}
	return RovingFocus{Key: key, Wrap: r.Wrap}.rememberNeighbors(model), key != r.Key
}

// End moves to the last enabled item.
func (r RovingFocus) End(model Model) (RovingFocus, bool) {
	key, ok := model.LastEnabled()
	if !ok {
		return RovingFocus{Wrap: r.Wrap}, r.Key != ""
	}
	return RovingFocus{Key: key, Wrap: r.Wrap}.rememberNeighbors(model), key != r.Key
}

func (r RovingFocus) setActive(model Model, key Key) RovingFocus {
	r.Key = key
	return r.rememberNeighbors(model)
}

func (r RovingFocus) rememberNeighbors(model Model) RovingFocus {
	index, ok := model.Index(r.Key)
	if !ok {
		r.neighbors = nil
		return r
	}

	neighbors := &rovingNeighbors{active: r.Key}
	for next := index + 1; next < len(model.items); next++ {
		item := model.items[next]
		if !item.Disabled {
			neighbors.following = append(neighbors.following, item.Key)
		}
	}
	for previous := index - 1; previous >= 0; previous-- {
		item := model.items[previous]
		if !item.Disabled {
			neighbors.preceding = append(neighbors.preceding, item.Key)
		}
	}
	if len(neighbors.following) == 0 && len(neighbors.preceding) == 0 {
		r.neighbors = nil
		return r
	}
	r.neighbors = neighbors
	return r
}

func (n *rovingNeighbors) firstEnabled(model Model, active Key, following bool) (Key, bool) {
	if n == nil || n.active != active {
		return "", false
	}
	keys := n.preceding
	if following {
		keys = n.following
	}
	for _, key := range keys {
		item, ok := model.Item(key)
		if ok && !item.Disabled {
			return key, true
		}
	}
	return "", false
}

// Viewport is a one-dimensional visible region in physical pixels.
type Viewport struct {
	Offset int
	Extent int
}

// Range identifies a half-open set of item indexes.
type Range struct {
	Start int
	End   int
}

// Len returns the number of indexes in the range.
func (r Range) Len() int {
	if r.End <= r.Start {
		return 0
	}
	return r.End - r.Start
}

// VisibleRange computes visible item indexes from monotonically increasing
// item boundaries. boundaries must have one more entry than the item count;
// malformed boundaries return an empty range. overscan is applied in item
// units after visibility is calculated.
func (v Viewport) VisibleRange(boundaries []int, overscan int) Range {
	if len(boundaries) < 2 || v.Extent <= 0 || !monotonic(boundaries) {
		return Range{}
	}
	if v.Offset < 0 {
		v.Offset = 0
	}
	count := len(boundaries) - 1
	endOffset := v.Offset + v.Extent
	start := sort.Search(count, func(index int) bool {
		return boundaries[index+1] > v.Offset
	})
	end := sort.Search(count, func(index int) bool {
		return boundaries[index] >= endOffset
	})
	if start >= count || end <= start {
		return Range{}
	}
	if overscan < 0 {
		overscan = 0
	}
	start -= overscan
	if start < 0 {
		start = 0
	}
	end += overscan
	if end > count {
		end = count
	}
	return Range{Start: start, End: end}
}

// ScrollOffsetFor returns the smallest clamped scroll offset that makes the
// item interval visible in the viewport. contentExtent is the full scrollable
// content size.
func (v Viewport) ScrollOffsetFor(itemStart, itemEnd, contentExtent int) int {
	if itemStart < 0 {
		itemStart = 0
	}
	if itemEnd < itemStart {
		itemEnd = itemStart
	}
	if contentExtent < itemEnd {
		contentExtent = itemEnd
	}
	if v.Offset < 0 {
		v.Offset = 0
	}
	if v.Extent <= 0 {
		return clampOffset(itemStart, contentExtent, 0)
	}
	viewEnd := v.Offset + v.Extent
	offset := v.Offset
	if itemStart < v.Offset {
		offset = itemStart
	} else if itemEnd > viewEnd {
		offset = itemEnd - v.Extent
	}
	return clampOffset(offset, contentExtent, v.Extent)
}

func monotonic(boundaries []int) bool {
	for index := 1; index < len(boundaries); index++ {
		if boundaries[index] < boundaries[index-1] {
			return false
		}
	}
	return true
}

func clampOffset(offset, contentExtent, viewportExtent int) int {
	if offset < 0 {
		return 0
	}
	max := contentExtent - viewportExtent
	if max < 0 {
		max = 0
	}
	if offset > max {
		return max
	}
	return offset
}
