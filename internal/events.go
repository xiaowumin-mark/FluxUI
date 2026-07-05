package internal

import (
	"sort"
	"time"
)

// EventType identifies a FluxUI event kind.
type EventType string

const (
	EventTypeFocus    EventType = "focus"
	EventTypeBlur     EventType = "blur"
	EventTypeFocusIn  EventType = "focusin"
	EventTypeFocusOut EventType = "focusout"
	EventTypeKeyDown  EventType = "keydown"
	EventTypeKeyUp    EventType = "keyup"
	EventTypeActivate EventType = "activate"

	EventTypeBeforeInput       EventType = "beforeinput"
	EventTypeInput             EventType = "input"
	EventTypeChange            EventType = "change"
	EventTypeSubmit            EventType = "submit"
	EventTypeCompositionStart  EventType = "compositionstart"
	EventTypeCompositionUpdate EventType = "compositionupdate"
	EventTypeCompositionEnd    EventType = "compositionend"
)

// EventPhase identifies the current dispatch phase.
type EventPhase int

const (
	EventPhaseNone EventPhase = iota
	EventPhaseCapture
	EventPhaseTarget
	EventPhaseBubble
)

// Modifiers records the keyboard modifiers active for an input event.
type Modifiers struct {
	Ctrl     bool
	Shift    bool
	Alt      bool
	Meta     bool
	Shortcut bool
}

// KeyLocation follows the DOM KeyboardEvent location constants.
type KeyLocation int

const (
	KeyLocationStandard KeyLocation = iota
	KeyLocationLeft
	KeyLocationRight
	KeyLocationNumpad
)

// FocusDirection identifies tab-order focus movement.
type FocusDirection int

const (
	FocusForward FocusDirection = iota
	FocusBackward
)

// Event is the base event object used by the FluxUI event dispatcher.
type Event struct {
	Type             EventType
	Target           PathID
	CurrentTarget    PathID
	Phase            EventPhase
	Time             time.Time
	Bubbles          bool
	Cancelable       bool
	DefaultPrevented bool
	Trusted          bool
	Detail           any

	path                        []PathID
	propagationStopped          bool
	immediatePropagationStopped bool
	currentPassiveListener      bool
}

// StopPropagation prevents the event from reaching further targets.
func (e *Event) StopPropagation() {
	if e == nil {
		return
	}
	e.propagationStopped = true
}

// StopImmediatePropagation prevents later listeners and further propagation.
func (e *Event) StopImmediatePropagation() {
	if e == nil {
		return
	}
	e.immediatePropagationStopped = true
	e.propagationStopped = true
}

// PreventDefault cancels the event default action when the event is cancelable.
func (e *Event) PreventDefault() bool {
	if e == nil || !e.Cancelable || e.currentPassiveListener {
		return false
	}
	e.DefaultPrevented = true
	return true
}

// PropagationStopped reports whether StopPropagation has been called.
func (e *Event) PropagationStopped() bool {
	return e != nil && e.propagationStopped
}

// ImmediatePropagationStopped reports whether StopImmediatePropagation was called.
func (e *Event) ImmediatePropagationStopped() bool {
	return e != nil && e.immediatePropagationStopped
}

// ComposedPath returns the target-to-root dispatch path.
func (e *Event) ComposedPath() []PathID {
	if e == nil || len(e.path) == 0 {
		return nil
	}
	out := make([]PathID, len(e.path))
	copy(out, e.path)
	return out
}

// EventListenerOptions configures an event listener.
type EventListenerOptions struct {
	Capture  bool
	Once     bool
	Passive  bool
	Priority int
}

// EventHandler handles a FluxUI event.
type EventHandler func(ctx *Context, event *Event)

// FocusEvent is the browser-style focus event payload.
type FocusEvent struct {
	Event
	RelatedTarget PathID
}

// KeyboardEvent is the browser-style keyboard event payload.
type KeyboardEvent struct {
	Event
	Key         string
	Code        string
	Location    KeyLocation
	Repeat      bool
	Modifiers   Modifiers
	IsComposing bool
	Native      any
}

// KeyboardHandler handles a typed keyboard event.
type KeyboardHandler func(ctx *Context, event *KeyboardEvent)

// FocusActivation is a cancelable keyboard default action for focused targets.
type FocusActivation func(ctx *Context)

// FocusTargetOptions configures a focusable target in the current frame.
type FocusTargetOptions struct {
	TabIndex int
	Disabled bool
	Hidden   bool
	Activate FocusActivation
}

// Shortcut describes a local component-tree shortcut.
type Shortcut struct {
	Key            string
	Code           string
	Modifiers      Modifiers
	ExactModifiers bool
	Scope          PathID
}

// EventBoundaryMode controls how propagation continues after a boundary target.
type EventBoundaryMode int

const (
	EventBoundaryNone EventBoundaryMode = iota
	EventBoundaryStop
	EventBoundaryRedirect
)

// EventBoundaryPolicy describes propagation behavior at a target boundary.
type EventBoundaryPolicy struct {
	Mode     EventBoundaryMode
	Redirect PathID
}

// EventTargetOptions configures advanced event target behavior.
type EventTargetOptions struct {
	Boundary EventBoundaryPolicy
}

type runtimeEventState struct {
	targets         map[PathID]eventTarget
	listeners       map[PathID][]*eventListener
	pointerCaptures map[uint64]PathID
	focusTargets    map[PathID]*focusTarget
	focusOrder      []PathID
	focusTarget     PathID
	focusContext    *Context
	nextFocusOrder  int
	keyDown         map[string]struct{}
	shortcuts       map[PathID][]*shortcutListener
	nextSeq         uint64
}

type eventTarget struct {
	Parent   PathID
	Boundary EventBoundaryPolicy
}

type eventListener struct {
	Type    EventType
	Handler EventHandler
	Options EventListenerOptions
	Context *Context
	seq     uint64
	removed bool
}

type eventListenerDispatchStats struct {
	calls    int64
	duration time.Duration
}

func (s *eventListenerDispatchStats) add(other eventListenerDispatchStats) {
	s.calls += other.calls
	s.duration += other.duration
}

type focusTarget struct {
	Target   PathID
	Parent   PathID
	TabIndex int
	Disabled bool
	Hidden   bool
	Activate FocusActivation
	Context  *Context
	order    int
}

func (t *focusTarget) focusable() bool {
	return t != nil && !t.Disabled && !t.Hidden
}

func (t *focusTarget) tabStop() bool {
	return t.focusable() && t.TabIndex >= 0
}

type shortcutListener struct {
	Spec    Shortcut
	Handler KeyboardHandler
	Options EventListenerOptions
	Context *Context
	seq     uint64
	removed bool
}

func (r *Runtime) beginEventFrame() {
	if r == nil {
		return
	}
	if r.events.targets == nil {
		r.events.targets = make(map[PathID]eventTarget)
	}
	if r.events.listeners == nil {
		r.events.listeners = make(map[PathID][]*eventListener)
	}
	if r.events.pointerCaptures == nil {
		r.events.pointerCaptures = make(map[uint64]PathID)
	}
	if r.events.focusTargets == nil {
		r.events.focusTargets = make(map[PathID]*focusTarget)
	}
	if r.events.keyDown == nil {
		r.events.keyDown = make(map[string]struct{})
	}
	if r.events.shortcuts == nil {
		r.events.shortcuts = make(map[PathID][]*shortcutListener)
	}
	clear(r.events.targets)
	clear(r.events.listeners)
	clear(r.events.focusTargets)
	clear(r.events.shortcuts)
	r.events.focusOrder = r.events.focusOrder[:0]
	r.events.nextFocusOrder = 0
	r.events.nextSeq = 0
	r.RegisterEventTarget(rootPathID, 0)
}

func (r *Runtime) endEventFrame() {
	if r == nil || r.events.focusTarget == 0 {
		return
	}
	target := normalizeOptionalPathID(r.events.focusTarget)
	entry := r.events.focusTargets[target]
	if entry != nil && entry.focusable() {
		return
	}
	r.changeFocus(r.events.focusContext, target, 0)
}

// RegisterEventTarget records an active target in the current frame.
func (r *Runtime) RegisterEventTarget(target, parent PathID) {
	r.registerEventTarget(target, parent, EventTargetOptions{}, false)
}

// RegisterEventTargetOptions records an active target with advanced event path
// options. It may update an already-registered target in the current frame.
func (r *Runtime) RegisterEventTargetOptions(target, parent PathID, opts EventTargetOptions) {
	r.registerEventTarget(target, parent, opts, true)
}

func (r *Runtime) registerEventTarget(target, parent PathID, opts EventTargetOptions, update bool) {
	if r == nil {
		return
	}
	if r.events.targets == nil {
		r.events.targets = make(map[PathID]eventTarget)
	}
	target = normalizePathID(target)
	if target == rootPathID {
		parent = 0
	} else {
		parent = normalizePathID(parent)
	}
	if entry, ok := r.events.targets[target]; ok {
		if update {
			entry.Parent = parent
			if opts.Boundary.Mode != EventBoundaryNone || opts.Boundary.Redirect != 0 {
				entry.Boundary = normalizeEventBoundary(opts.Boundary)
			}
			r.events.targets[target] = entry
		}
		return
	}
	r.events.targets[target] = eventTarget{
		Parent:   parent,
		Boundary: normalizeEventBoundary(opts.Boundary),
	}
}

// RegisterEventBoundary marks ctx as an event boundary for the current frame.
func (r *Runtime) RegisterEventBoundary(ctx *Context, policy EventBoundaryPolicy) {
	if r == nil || ctx == nil {
		return
	}
	target := normalizePathID(ctx.pathID)
	r.RegisterEventTargetOptions(target, r.eventParentFor(target), EventTargetOptions{
		Boundary: policy,
	})
}

// RegisterEventPortal rewires ctx's logical event parent to owner. This is
// intended for overlay/portal roots whose layout parent differs from their
// component owner.
func (r *Runtime) RegisterEventPortal(ctx *Context, owner PathID) {
	if r == nil || ctx == nil {
		return
	}
	target := normalizePathID(ctx.pathID)
	owner = normalizeOptionalPathID(owner)
	if owner == 0 {
		owner = rootPathID
	}
	r.RegisterEventTargetOptions(target, owner, EventTargetOptions{})
}

// RegisterEventListener records a listener for the current frame target.
func (r *Runtime) RegisterEventListener(ctx *Context, eventType EventType, handler EventHandler, opts EventListenerOptions) {
	if r == nil || ctx == nil || eventType == "" || handler == nil {
		return
	}
	if r.events.listeners == nil {
		r.events.listeners = make(map[PathID][]*eventListener)
	}
	target := normalizePathID(ctx.pathID)
	r.RegisterEventTarget(target, r.eventParentFor(target))
	r.events.nextSeq++
	r.events.listeners[target] = append(r.events.listeners[target], &eventListener{
		Type:    eventType,
		Handler: handler,
		Options: opts,
		Context: ctx,
		seq:     r.events.nextSeq,
	})
	sort.SliceStable(r.events.listeners[target], func(i, j int) bool {
		return eventListenerLess(r.events.listeners[target][i], r.events.listeners[target][j])
	})
}

// RegisterFocusTarget records a focusable target in the current frame.
func (r *Runtime) RegisterFocusTarget(ctx *Context, opts FocusTargetOptions) {
	if r == nil || ctx == nil {
		return
	}
	target := normalizePathID(ctx.pathID)
	r.RegisterEventTarget(target, r.eventParentFor(target))
	entry := &focusTarget{
		Target:   target,
		Parent:   r.eventParentFor(target),
		TabIndex: opts.TabIndex,
		Disabled: opts.Disabled,
		Hidden:   opts.Hidden,
		Activate: opts.Activate,
		Context:  ctx,
		order:    r.events.nextFocusOrder,
	}
	r.events.nextFocusOrder++
	_, existed := r.events.focusTargets[target]
	r.events.focusTargets[target] = entry
	if !existed && entry.tabStop() {
		r.events.focusOrder = append(r.events.focusOrder, target)
	}
	if r.events.focusTarget == target && entry.focusable() {
		r.events.focusContext = ctx
	}
}

// FocusedTarget returns the current component-tree focus target.
func (r *Runtime) FocusedTarget() PathID {
	if r == nil {
		return 0
	}
	return normalizeOptionalPathID(r.events.focusTarget)
}

// Focused reports whether target is the current focus target.
func (r *Runtime) Focused(target PathID) bool {
	if r == nil || target == 0 {
		return false
	}
	return normalizePathID(target) == normalizeOptionalPathID(r.events.focusTarget)
}

// RequestFocus moves component-tree focus to target.
func (r *Runtime) RequestFocus(ctx *Context, target PathID) bool {
	if r == nil {
		return false
	}
	if target == 0 {
		return r.BlurFocus(ctx, 0)
	}
	target = normalizePathID(target)
	entry := r.events.focusTargets[target]
	if entry == nil || !entry.focusable() {
		return false
	}
	if ctx == nil {
		ctx = entry.Context
	}
	return r.changeFocus(ctx, normalizeOptionalPathID(r.events.focusTarget), target)
}

// BlurFocus clears focus when target owns focus. Passing target 0 clears
// whichever target is currently focused.
func (r *Runtime) BlurFocus(ctx *Context, target PathID) bool {
	if r == nil {
		return false
	}
	current := normalizeOptionalPathID(r.events.focusTarget)
	if current == 0 {
		return true
	}
	target = normalizeOptionalPathID(target)
	if target != 0 && target != current {
		return false
	}
	return r.changeFocus(ctx, current, 0)
}

// MoveFocus moves focus through the current frame's tab order.
func (r *Runtime) MoveFocus(ctx *Context, direction FocusDirection) bool {
	if r == nil {
		return false
	}
	order := r.sortedFocusOrder()
	if len(order) == 0 {
		return false
	}
	current := normalizeOptionalPathID(r.events.focusTarget)
	index := -1
	for i, target := range order {
		if target == current {
			index = i
			break
		}
	}
	var next int
	if direction == FocusBackward {
		if index < 0 {
			next = len(order) - 1
		} else {
			next = (index - 1 + len(order)) % len(order)
		}
	} else {
		if index < 0 {
			next = 0
		} else {
			next = (index + 1) % len(order)
		}
	}
	return r.RequestFocus(ctx, order[next])
}

// EventPathContains reports whether ancestor is in target's current event path.
func (r *Runtime) EventPathContains(target, ancestor PathID) bool {
	if r == nil {
		return false
	}
	ancestor = normalizePathID(ancestor)
	for _, id := range r.eventPath(target) {
		if id == ancestor {
			return true
		}
	}
	return false
}

// RegisterShortcut records a local shortcut listener for the current frame.
func (r *Runtime) RegisterShortcut(ctx *Context, spec Shortcut, handler KeyboardHandler, opts EventListenerOptions) {
	if r == nil || ctx == nil || handler == nil {
		return
	}
	if spec.Key == "" && spec.Code == "" {
		return
	}
	if r.events.shortcuts == nil {
		r.events.shortcuts = make(map[PathID][]*shortcutListener)
	}
	target := normalizePathID(ctx.pathID)
	r.RegisterEventTarget(target, r.eventParentFor(target))
	if spec.Scope != 0 {
		spec.Scope = normalizePathID(spec.Scope)
	}
	r.events.nextSeq++
	r.events.shortcuts[target] = append(r.events.shortcuts[target], &shortcutListener{
		Spec:    spec,
		Handler: handler,
		Options: opts,
		Context: ctx,
		seq:     r.events.nextSeq,
	})
}

// SetPointerCapture records the target that should receive subsequent events
// for pointerID until release or cancel.
func (r *Runtime) SetPointerCapture(pointerID uint64, target PathID) {
	if r == nil || target == 0 {
		return
	}
	if r.events.pointerCaptures == nil {
		r.events.pointerCaptures = make(map[uint64]PathID)
	}
	target = normalizePathID(target)
	r.RegisterEventTarget(target, r.eventParentFor(target))
	r.events.pointerCaptures[pointerID] = target
}

// ReleasePointerCapture releases pointerID when target owns the capture. Passing
// target 0 releases the capture unconditionally.
func (r *Runtime) ReleasePointerCapture(pointerID uint64, target PathID) bool {
	if r == nil || r.events.pointerCaptures == nil {
		return false
	}
	owner, ok := r.events.pointerCaptures[pointerID]
	if !ok {
		return false
	}
	if target == 0 {
		delete(r.events.pointerCaptures, pointerID)
		return true
	}
	target = normalizePathID(target)
	if target != 0 && owner != target {
		return false
	}
	delete(r.events.pointerCaptures, pointerID)
	return true
}

// PointerCaptureTarget returns the current capture target for pointerID.
func (r *Runtime) PointerCaptureTarget(pointerID uint64) (PathID, bool) {
	if r == nil || r.events.pointerCaptures == nil {
		return 0, false
	}
	target, ok := r.events.pointerCaptures[pointerID]
	return target, ok
}

// HasPointerCapture reports whether target owns pointerID.
func (r *Runtime) HasPointerCapture(pointerID uint64, target PathID) bool {
	owner, ok := r.PointerCaptureTarget(pointerID)
	return ok && owner == normalizePathID(target)
}

// DispatchKeyboardEvent dispatches a typed keyboard event to the focused target
// and runs cancelable local shortcut and default-action handling for keydown.
func (r *Runtime) DispatchKeyboardEvent(ctx *Context, target PathID, event *KeyboardEvent) bool {
	if event == nil {
		return true
	}
	if r == nil {
		return !(event.Cancelable && event.DefaultPrevented)
	}
	target = normalizeOptionalPathID(target)
	if target == 0 {
		target = normalizeOptionalPathID(r.events.focusTarget)
	}
	if target == 0 {
		target = rootPathID
	}
	r.applyKeyboardDefaults(ctx, event)
	event.Event.Detail = event
	allowed := r.DispatchEvent(ctx, target, &event.Event)
	if event.Type == EventTypeKeyDown && allowed {
		r.dispatchShortcuts(ctx, target, event)
		if !(event.Cancelable && event.DefaultPrevented) {
			r.runKeyboardDefault(ctx, target, event)
		}
	}
	return !(event.Cancelable && event.DefaultPrevented)
}

// DispatchEvent dispatches an event to a target and returns false when the
// event was cancelable and default-prevented.
func (r *Runtime) DispatchEvent(ctx *Context, target PathID, event *Event) bool {
	if event == nil {
		return true
	}
	if r == nil {
		return !(event.Cancelable && event.DefaultPrevented)
	}
	target = normalizePathID(target)
	r.RegisterEventTarget(target, r.eventParentFor(target))
	if event.Type == "" {
		return true
	}
	if event.Target == 0 {
		event.Target = target
	} else {
		event.Target = normalizePathID(event.Target)
	}
	if event.Time.IsZero() {
		if ctx != nil && !ctx.Now().IsZero() {
			event.Time = ctx.Now()
		} else {
			event.Time = time.Now()
		}
	}
	path := r.eventPath(target)
	event.path = path
	if r.eventDiagnosticsEnabled() {
		r.ObserveEventDispatch(event.Type)
	}

	var listenerStats eventListenerDispatchStats
	for i := len(path) - 1; i >= 1; i-- {
		if event.propagationStopped {
			break
		}
		current := path[i]
		listenerStats.add(r.dispatchEventListeners(current, EventPhaseCapture, event, true))
	}

	if !event.propagationStopped {
		event.Phase = EventPhaseTarget
		event.CurrentTarget = target
		listenerStats.add(r.dispatchEventListeners(target, EventPhaseTarget, event, true))
		if !event.immediatePropagationStopped {
			listenerStats.add(r.dispatchEventListeners(target, EventPhaseTarget, event, false))
		}
	}

	if event.Bubbles && !event.propagationStopped {
		for _, current := range path[1:] {
			if event.propagationStopped {
				break
			}
			listenerStats.add(r.dispatchEventListeners(current, EventPhaseBubble, event, false))
		}
	}

	allowed := !(event.Cancelable && event.DefaultPrevented)
	r.recordEventDispatch(event, path, listenerStats.calls, listenerStats.duration, allowed)
	event.Phase = EventPhaseNone
	event.CurrentTarget = 0
	return allowed
}

func (r *Runtime) dispatchEventListeners(target PathID, phase EventPhase, event *Event, capture bool) eventListenerDispatchStats {
	var stats eventListenerDispatchStats
	target = normalizePathID(target)
	listeners := r.events.listeners[target]
	if len(listeners) == 0 {
		return stats
	}
	event.Phase = phase
	event.CurrentTarget = target
	measureDuration := r.eventListenerDurationEnabled()
	for _, listener := range listeners {
		if listener == nil || listener.removed || listener.Handler == nil || listener.Type != event.Type || listener.Options.Capture != capture {
			continue
		}
		if event.immediatePropagationStopped {
			break
		}
		previousPassive := event.currentPassiveListener
		event.currentPassiveListener = listener.Options.Passive
		var started time.Time
		if measureDuration {
			started = time.Now()
		}
		listener.Handler(listener.Context, event)
		if measureDuration {
			stats.duration += time.Since(started)
		}
		stats.calls++
		event.currentPassiveListener = previousPassive
		if listener.Options.Once {
			listener.removed = true
		}
	}
	r.pruneRemovedEventListeners(target)
	return stats
}

func eventListenerLess(a, b *eventListener) bool {
	if a == nil || b == nil {
		return b != nil
	}
	if a.Options.Priority == b.Options.Priority {
		return a.seq < b.seq
	}
	return a.Options.Priority > b.Options.Priority
}

func (r *Runtime) pruneRemovedEventListeners(target PathID) {
	listeners := r.events.listeners[normalizePathID(target)]
	if len(listeners) == 0 {
		return
	}
	next := listeners[:0]
	for _, listener := range listeners {
		if listener != nil && !listener.removed {
			next = append(next, listener)
		}
	}
	if len(next) == 0 {
		delete(r.events.listeners, normalizePathID(target))
		return
	}
	r.events.listeners[normalizePathID(target)] = next
}

func (r *Runtime) pruneRemovedShortcutListeners(target PathID) {
	listeners := r.events.shortcuts[normalizePathID(target)]
	if len(listeners) == 0 {
		return
	}
	next := listeners[:0]
	for _, listener := range listeners {
		if listener != nil && !listener.removed {
			next = append(next, listener)
		}
	}
	if len(next) == 0 {
		delete(r.events.shortcuts, normalizePathID(target))
		return
	}
	r.events.shortcuts[normalizePathID(target)] = next
}

func (r *Runtime) sortedFocusOrder() []PathID {
	if r == nil || len(r.events.focusOrder) == 0 {
		return nil
	}
	entries := make([]*focusTarget, 0, len(r.events.focusOrder))
	for _, target := range r.events.focusOrder {
		entry := r.events.focusTargets[normalizePathID(target)]
		if entry != nil && entry.tabStop() {
			entries = append(entries, entry)
		}
	}
	if len(entries) == 0 {
		return nil
	}
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		aPositive := a.TabIndex > 0
		bPositive := b.TabIndex > 0
		if aPositive != bPositive {
			return aPositive
		}
		if aPositive && a.TabIndex != b.TabIndex {
			return a.TabIndex < b.TabIndex
		}
		return a.order < b.order
	})
	out := make([]PathID, len(entries))
	for i, entry := range entries {
		out[i] = entry.Target
	}
	return out
}

func (r *Runtime) changeFocus(ctx *Context, oldTarget, newTarget PathID) bool {
	oldTarget = normalizeOptionalPathID(oldTarget)
	newTarget = normalizeOptionalPathID(newTarget)
	if oldTarget == newTarget {
		if newTarget != 0 {
			r.events.focusContext = r.focusContextFor(newTarget, ctx)
		}
		return true
	}
	r.events.focusTarget = newTarget
	r.events.focusContext = r.focusContextFor(newTarget, ctx)

	if oldTarget != 0 {
		r.dispatchFocusEvent(ctx, oldTarget, EventTypeBlur, false, newTarget)
		r.dispatchFocusEvent(ctx, oldTarget, EventTypeFocusOut, true, newTarget)
	}
	if newTarget != 0 {
		r.dispatchFocusEvent(ctx, newTarget, EventTypeFocus, false, oldTarget)
		r.dispatchFocusEvent(ctx, newTarget, EventTypeFocusIn, true, oldTarget)
	}
	return true
}

func (r *Runtime) dispatchFocusEvent(ctx *Context, target PathID, eventType EventType, bubbles bool, related PathID) {
	if r == nil || target == 0 {
		return
	}
	eventCtx := r.focusContextFor(target, ctx)
	ev := &FocusEvent{
		Event: Event{
			Type:       eventType,
			Target:     target,
			Bubbles:    bubbles,
			Trusted:    true,
			Cancelable: false,
		},
		RelatedTarget: normalizeOptionalPathID(related),
	}
	ev.Event.Detail = ev
	r.DispatchEvent(eventCtx, target, &ev.Event)
}

func (r *Runtime) focusContextFor(target PathID, fallback *Context) *Context {
	if r == nil {
		return fallback
	}
	target = normalizeOptionalPathID(target)
	if target == 0 {
		return fallback
	}
	if entry := r.events.focusTargets[target]; entry != nil && entry.Context != nil {
		return entry.Context
	}
	if target == normalizeOptionalPathID(r.events.focusTarget) && r.events.focusContext != nil {
		return r.events.focusContext
	}
	return fallback
}

func (r *Runtime) applyKeyboardDefaults(ctx *Context, event *KeyboardEvent) {
	if event.Type == "" {
		event.Type = EventTypeKeyDown
	}
	event.Bubbles = true
	event.Cancelable = true
	if event.Time.IsZero() {
		if ctx != nil && !ctx.Now().IsZero() {
			event.Time = ctx.Now()
		} else {
			event.Time = time.Now()
		}
	}
	r.recordKeyState(event)
}

func (r *Runtime) recordKeyState(event *KeyboardEvent) {
	if r == nil || event == nil {
		return
	}
	if r.events.keyDown == nil {
		r.events.keyDown = make(map[string]struct{})
	}
	key := keyboardIdentity(event)
	if key == "" {
		return
	}
	switch event.Type {
	case EventTypeKeyDown:
		if _, ok := r.events.keyDown[key]; ok {
			event.Repeat = true
		}
		r.events.keyDown[key] = struct{}{}
	case EventTypeKeyUp:
		delete(r.events.keyDown, key)
	}
}

func keyboardIdentity(event *KeyboardEvent) string {
	if event == nil {
		return ""
	}
	if event.Code != "" {
		return event.Code
	}
	return event.Key
}

type shortcutMatch struct {
	listener *shortcutListener
	target   PathID
	depth    int
}

func (r *Runtime) dispatchShortcuts(ctx *Context, target PathID, event *KeyboardEvent) {
	if r == nil || event == nil || len(r.events.shortcuts) == 0 {
		return
	}
	path := r.eventPath(target)
	matches := make([]shortcutMatch, 0)
	for listenerTarget, listeners := range r.events.shortcuts {
		for _, listener := range listeners {
			if listener == nil || listener.removed || listener.Handler == nil {
				continue
			}
			scope := listener.Spec.Scope
			if scope == 0 {
				scope = normalizePathID(listenerTarget)
			} else {
				scope = normalizePathID(scope)
			}
			depth := pathIndex(path, scope)
			if depth < 0 || !shortcutMatches(listener.Spec, event) {
				continue
			}
			matches = append(matches, shortcutMatch{listener: listener, target: listenerTarget, depth: depth})
		}
	}
	if len(matches) == 0 {
		return
	}
	sort.SliceStable(matches, func(i, j int) bool {
		a, b := matches[i], matches[j]
		if a.depth != b.depth {
			return a.depth < b.depth
		}
		if a.listener.Options.Priority != b.listener.Options.Priority {
			return a.listener.Options.Priority > b.listener.Options.Priority
		}
		return a.listener.seq < b.listener.seq
	})
	prune := make(map[PathID]struct{})
	for _, match := range matches {
		listener := match.listener
		if listener == nil || listener.removed || listener.Handler == nil {
			continue
		}
		if event.ImmediatePropagationStopped() {
			break
		}
		previousPassive := event.currentPassiveListener
		event.currentPassiveListener = listener.Options.Passive
		listener.Handler(listener.Context, event)
		event.currentPassiveListener = previousPassive
		if listener.Options.Once {
			listener.removed = true
			prune[match.target] = struct{}{}
		}
		if event.PropagationStopped() {
			break
		}
	}
	for target := range prune {
		r.pruneRemovedShortcutListeners(target)
	}
	_ = ctx
}

func shortcutMatches(spec Shortcut, event *KeyboardEvent) bool {
	if event == nil {
		return false
	}
	if spec.Key != "" && spec.Key != event.Key {
		return false
	}
	if spec.Code != "" && spec.Code != event.Code {
		return false
	}
	if !modifiersContain(event.Modifiers, spec.Modifiers) {
		return false
	}
	if spec.ExactModifiers && !modifiersEqual(event.Modifiers, spec.Modifiers) {
		return false
	}
	return true
}

func modifiersContain(have, required Modifiers) bool {
	if required.Ctrl && !have.Ctrl {
		return false
	}
	if required.Shift && !have.Shift {
		return false
	}
	if required.Alt && !have.Alt {
		return false
	}
	if required.Meta && !have.Meta {
		return false
	}
	if required.Shortcut && !have.Shortcut {
		return false
	}
	return true
}

func modifiersEqual(a, b Modifiers) bool {
	return a.Ctrl == b.Ctrl &&
		a.Shift == b.Shift &&
		a.Alt == b.Alt &&
		a.Meta == b.Meta &&
		a.Shortcut == b.Shortcut
}

func (r *Runtime) runKeyboardDefault(ctx *Context, target PathID, event *KeyboardEvent) {
	if r == nil || event == nil || event.Type != EventTypeKeyDown {
		return
	}
	switch event.Key {
	case "Tab":
		if event.Modifiers.Shift {
			r.MoveFocus(ctx, FocusBackward)
		} else {
			r.MoveFocus(ctx, FocusForward)
		}
	case "Enter", "\n", "Space", " ":
		r.activateFocusTarget(ctx, target)
	}
}

func (r *Runtime) activateFocusTarget(ctx *Context, target PathID) {
	if r == nil {
		return
	}
	target = normalizeOptionalPathID(target)
	if target == 0 {
		target = normalizeOptionalPathID(r.events.focusTarget)
	}
	entry := r.events.focusTargets[target]
	if entry == nil || !entry.focusable() || entry.Activate == nil {
		return
	}
	if entry.Context != nil {
		ctx = entry.Context
	}
	entry.Activate(ctx)
}

func (r *Runtime) eventPath(target PathID) []PathID {
	target = normalizePathID(target)
	path := make([]PathID, 0, 8)
	path = append(path, target)
	for current := target; current != rootPathID; {
		if len(path) >= 256 {
			break
		}
		parent := r.eventParentFor(current)
		if entry, ok := r.events.targets[normalizePathID(current)]; ok {
			switch entry.Boundary.Mode {
			case EventBoundaryStop:
				return path
			case EventBoundaryRedirect:
				redirect := normalizeOptionalPathID(entry.Boundary.Redirect)
				if redirect != 0 {
					parent = redirect
				} else {
					parent = rootPathID
				}
			}
		}
		if parent == 0 {
			parent = rootPathID
		}
		parent = normalizePathID(parent)
		if pathIndex(path, parent) >= 0 {
			break
		}
		path = append(path, parent)
		current = parent
	}
	return path
}

func normalizeEventBoundary(policy EventBoundaryPolicy) EventBoundaryPolicy {
	switch policy.Mode {
	case EventBoundaryStop:
		policy.Redirect = 0
	case EventBoundaryRedirect:
		policy.Redirect = normalizeOptionalPathID(policy.Redirect)
	default:
		policy = EventBoundaryPolicy{}
	}
	return policy
}

func (r *Runtime) eventParentFor(target PathID) PathID {
	if r == nil {
		return 0
	}
	target = normalizePathID(target)
	if target == rootPathID {
		return 0
	}
	if r.events.targets != nil {
		if entry, ok := r.events.targets[target]; ok {
			return entry.Parent
		}
	}
	r.initPathTable()
	if entry := r.pathDebug[target]; entry != nil {
		return entry.parent
	}
	return rootPathID
}

func pathIndex(path []PathID, target PathID) int {
	target = normalizePathID(target)
	for i, id := range path {
		if normalizePathID(id) == target {
			return i
		}
	}
	return -1
}

func normalizeOptionalPathID(id PathID) PathID {
	if id == 0 {
		return 0
	}
	return normalizePathID(id)
}
