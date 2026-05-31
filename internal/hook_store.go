package internal

import (
	"fmt"
	"reflect"
	"strconv"
)

// HookKind identifies the data stored in a component hook slot.
type HookKind string

const (
	HookState     HookKind = "state"
	HookMemo      HookKind = "memo"
	HookRef       HookKind = "ref"
	HookEffect    HookKind = "effect"
	HookAnimValue HookKind = "anim_value"
	HookAnimDeco  HookKind = "anim_deco"
)

// ComponentIdentity describes the stable identity input for a component.
// Keyed components use Key for reuse; unkeyed siblings fall back to Position.
type ComponentIdentity struct {
	ParentID string
	TypeID   string
	Key      string
	Position int
}

// StableID returns the runtime instance id derived from the identity inputs.
func (id ComponentIdentity) StableID() string {
	parent := id.ParentID
	if parent == "" {
		parent = "root"
	}
	typeID := id.TypeID
	if typeID == "" {
		typeID = "component"
	}
	if id.Key != "" {
		return parent + "/" + typeID + "#" + strconv.Quote(id.Key)
	}
	return parent + "/" + typeID + "@" + strconv.Itoa(id.Position)
}

// HookSlot is a component-owned hook cell. Later Phase 2 stages migrate
// UseState, UseMemo, UseRef and UseEffect onto these slots.
type HookSlot struct {
	Kind        HookKind
	Value       any
	Initialized bool
	HasDeps     bool
	Deps        []any
	Cleanup     func()
}

// ComponentInstance owns hook slots for one stable component identity.
type ComponentInstance struct {
	ID       string
	Identity ComponentIdentity

	hooks      []HookSlot
	hookCursor int
}

// BeginHooks resets the per-render hook cursor.
func (i *ComponentInstance) BeginHooks() {
	if i == nil {
		return
	}
	i.hookCursor = 0
}

// NextHook returns the next hook slot for this render, creating it if needed.
func (i *ComponentInstance) NextHook(kind HookKind) *HookSlot {
	if i == nil {
		return nil
	}
	idx := i.hookCursor
	i.hookCursor++
	if idx >= len(i.hooks) {
		i.hooks = append(i.hooks, HookSlot{Kind: kind})
	}
	slot := &i.hooks[idx]
	if slot.Kind == "" {
		slot.Kind = kind
	} else if slot.Kind != kind {
		panic(fmt.Sprintf(
			"FluxUI: component %q rendered hook #%d as %q, but it was %q before -- "+
				"hooks must be called in the same order every render",
			i.ID, idx, kind, slot.Kind,
		))
	}
	return slot
}

// HookCount reports how many hooks were consumed in the current render.
func (i *ComponentInstance) HookCount() int {
	if i == nil {
		return 0
	}
	return i.hookCursor
}

// HookStore tracks component instances and their hook slots across frames.
type HookStore struct {
	instances map[string]*ComponentInstance
	active    map[string]struct{}
	pending   []func()
}

// NewHookStore creates an empty component hook store.
func NewHookStore() *HookStore {
	return &HookStore{
		instances: make(map[string]*ComponentInstance),
		active:    make(map[string]struct{}),
	}
}

// BeginFrame resets active instance tracking for the next render pass.
func (s *HookStore) BeginFrame() {
	if s == nil {
		return
	}
	clear(s.active)
	s.pending = s.pending[:0]
}

// BeginInstance returns the component instance for identity and marks it active.
func (s *HookStore) BeginInstance(identity ComponentIdentity) *ComponentInstance {
	if s == nil {
		return nil
	}
	id := identity.StableID()
	inst := s.instances[id]
	if inst == nil {
		inst = &ComponentInstance{ID: id, Identity: identity}
		s.instances[id] = inst
	}
	s.active[id] = struct{}{}
	inst.BeginHooks()
	return inst
}

// Instance returns a tracked instance by stable id.
func (s *HookStore) Instance(id string) *ComponentInstance {
	if s == nil {
		return nil
	}
	return s.instances[id]
}

// EndFrame unmounts instances that were not rendered this frame.
func (s *HookStore) EndFrame() {
	if s == nil {
		return
	}
	for id, inst := range s.instances {
		if _, ok := s.active[id]; !ok {
			continue
		}
		if inst != nil && inst.hookCursor != len(inst.hooks) {
			panic(fmt.Sprintf(
				"FluxUI: component %q rendered %d hooks this frame, but rendered %d before -- "+
					"hooks must be called in the same order every render",
				inst.ID, inst.hookCursor, len(inst.hooks),
			))
		}
	}
	for id, inst := range s.instances {
		if _, ok := s.active[id]; ok {
			continue
		}
		runHookCleanups(inst)
		delete(s.instances, id)
	}
	for _, run := range s.pending {
		if run != nil {
			run()
		}
	}
	s.pending = s.pending[:0]
}

// Dispose releases all tracked hook instances and runs effect cleanups.
func (s *HookStore) Dispose() {
	if s == nil {
		return
	}
	for id, inst := range s.instances {
		runHookCleanups(inst)
		delete(s.instances, id)
	}
	clear(s.active)
	s.pending = nil
}

// UseEffect registers a component-owned post-frame side effect.
func (s *HookStore) UseEffect(slot *HookSlot, hasDeps bool, deps []any, setup EffectSetup) {
	if s == nil || slot == nil || setup == nil {
		return
	}
	nextDeps := CloneDeps(deps)
	if !ShouldRunHookEffect(slot, hasDeps, nextDeps) {
		return
	}
	s.pending = append(s.pending, func() {
		if slot.Cleanup != nil {
			slot.Cleanup()
			slot.Cleanup = nil
		}
		slot.Initialized = true
		slot.HasDeps = hasDeps
		slot.Deps = nextDeps
		slot.Cleanup = setup()
	})
}

// ShouldRunHookEffect reports whether a HookSlot effect should run this frame.
func ShouldRunHookEffect(slot *HookSlot, hasDeps bool, nextDeps []any) bool {
	if slot == nil || !slot.Initialized {
		return true
	}
	if !hasDeps {
		return true
	}
	if !slot.HasDeps {
		return true
	}
	return !DepsEqual(slot.Deps, nextDeps)
}

// DepsEqual compares hook dependency slices with reflect.DeepEqual per item.
func DepsEqual(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for idx := range a {
		if !reflect.DeepEqual(a[idx], b[idx]) {
			return false
		}
	}
	return true
}

// CloneDeps copies hook dependencies before storing them in a slot.
func CloneDeps(deps []any) []any {
	if len(deps) == 0 {
		return nil
	}
	out := make([]any, len(deps))
	copy(out, deps)
	return out
}

func runHookCleanups(inst *ComponentInstance) {
	if inst == nil {
		return
	}
	for idx := range inst.hooks {
		cleanup := inst.hooks[idx].Cleanup
		if cleanup == nil {
			continue
		}
		inst.hooks[idx].Cleanup = nil
		cleanup()
	}
}
