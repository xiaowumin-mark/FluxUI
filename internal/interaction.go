package internal

import (
	"image"
	"sync"

	"gioui.org/f32"
)

// InteractionChangeKind describes the Flux-level semantic change caused by
// input. Pointer moves without any state change are intentionally separate from
// hover/press/focus changes so callers can avoid business redraws.
type InteractionChangeKind string

const (
	InteractionPointerMoveChanged InteractionChangeKind = "pointer.move"
	InteractionHoverChanged       InteractionChangeKind = "pointer.hover_changed"
	InteractionPressedChanged     InteractionChangeKind = "pointer.pressed_changed"
	InteractionFocusChanged       InteractionChangeKind = "pointer.focus_changed"
)

// InteractionFrameStats summarizes Flux interaction changes observed while
// building one frame.
type InteractionFrameStats struct {
	Frame           uint64
	PointerMoves    int
	PointerEvents   int
	WheelEvents     int
	KeyboardEvents  int
	FocusEvents     int
	HoverChanged    int
	PressedChanged  int
	FocusChanged    int
	HoverTarget     string
	HoverTargetPrev string
}

type runtimeInteractionState struct {
	mu           sync.Mutex
	frame        uint64
	last         InteractionFrameStats
	hoverTarget  PathID
	activeHover  PathID
	pointerMoves int
	pointerPos   image.Point
	hasPointer   bool
}

func (r *Runtime) beginInteractionFrame() {
	if r == nil {
		return
	}
	r.interact.mu.Lock()
	r.interact.frame++
	pointerMoves := r.interact.pointerMoves
	r.interact.pointerMoves = 0
	r.interact.last = InteractionFrameStats{
		Frame:        r.interact.frame,
		PointerMoves: pointerMoves,
	}
	r.interact.hasPointer = false
	r.interact.activeHover = 0
	r.interact.mu.Unlock()
}

func (r *Runtime) endInteractionFrame() {
	if r == nil {
		return
	}
	r.interact.mu.Lock()
	prevTarget := r.interact.hoverTarget
	nextTarget := r.interact.activeHover
	hoverChanged := 0
	if prevTarget != nextTarget {
		hoverChanged = 1
		r.interact.hoverTarget = nextTarget
	}
	last := r.interact.last
	last.PointerMoves += r.interact.pointerMoves
	last.HoverChanged = hoverChanged
	last.HoverTargetPrev = r.debugInteractionTarget(prevTarget)
	last.HoverTarget = r.debugInteractionTarget(nextTarget)
	r.interact.last = last
	r.interact.pointerMoves = 0
	r.interact.mu.Unlock()

	if hoverChanged > 0 {
		r.RequestRedrawReason(string(InteractionHoverChanged))
	}
}

// QueuePointerMove records only the latest pointer move for the next Flux
// frame. The caller should still pass the latest event to Gio's router.
func (r *Runtime) QueuePointerMove(pos image.Point) {
	if r == nil {
		return
	}
	r.interact.mu.Lock()
	r.interact.pointerPos = pos
	r.interact.hasPointer = true
	r.interact.mu.Unlock()
}

// QueuePointerMoveF32 records a pointer move using Gio's f32 coordinates.
func (r *Runtime) QueuePointerMoveF32(pos f32.Point) {
	r.QueuePointerMove(image.Pt(int(pos.X+0.5), int(pos.Y+0.5)))
}

// CoalescedPointerMove returns the latest queued pointer move and clears it.
func (r *Runtime) CoalescedPointerMove() (image.Point, bool) {
	if r == nil {
		return image.Point{}, false
	}
	r.interact.mu.Lock()
	pos, ok := r.interact.pointerPos, r.interact.hasPointer
	if ok {
		r.interact.pointerMoves++
	}
	r.interact.hasPointer = false
	r.interact.mu.Unlock()
	return pos, ok
}

// ObserveInteractionSnapshot records Flux-level interaction changes for a
// stable component target.
func (r *Runtime) ObserveInteractionSnapshot(target PathID, previous, current ClickableSnapshot, initialized bool) InteractionFrameStats {
	if r == nil || target == 0 {
		return InteractionFrameStats{}
	}

	r.interact.mu.Lock()
	if current.Hovered {
		r.interact.activeHover = normalizePathID(target)
	}
	if initialized && previous.Pressed != current.Pressed {
		r.interact.last.PressedChanged++
	}
	if initialized && previous.Focused != current.Focused {
		r.interact.last.FocusChanged++
	}
	stats := r.interact.last
	r.interact.mu.Unlock()

	if initialized && previous.Pressed != current.Pressed {
		r.RequestRedrawReason(string(InteractionPressedChanged))
	}
	if initialized && previous.Focused != current.Focused {
		r.RequestRedrawReason(string(InteractionFocusChanged))
	}
	return stats
}

func (r *Runtime) ObserveEventDispatch(eventType EventType) {
	if r == nil || eventType == "" {
		return
	}
	r.interact.mu.Lock()
	switch eventInteractionKind(eventType) {
	case "pointer":
		r.interact.last.PointerEvents++
	case "wheel":
		r.interact.last.WheelEvents++
	case "keyboard":
		r.interact.last.KeyboardEvents++
	case "focus":
		r.interact.last.FocusEvents++
	}
	r.interact.mu.Unlock()
}

func (r *Runtime) LastInteractionStats() InteractionFrameStats {
	if r == nil {
		return InteractionFrameStats{}
	}
	r.interact.mu.Lock()
	stats := r.interact.last
	r.interact.mu.Unlock()
	return stats
}

func (r *Runtime) debugInteractionTarget(id PathID) string {
	if r == nil || id == 0 {
		return ""
	}
	return r.DebugPath(id)
}

func eventInteractionKind(eventType EventType) string {
	switch eventType {
	case "pointerdown", "pointerup", "pointermove", "pointerenter", "pointerleave", "pointerover", "pointerout", "pointercancel", "click", "dblclick", "auxclick", "contextmenu":
		return "pointer"
	case "wheel":
		return "wheel"
	case EventTypeKeyDown, EventTypeKeyUp:
		return "keyboard"
	case EventTypeFocus, EventTypeBlur, EventTypeFocusIn, EventTypeFocusOut:
		return "focus"
	default:
		return ""
	}
}
