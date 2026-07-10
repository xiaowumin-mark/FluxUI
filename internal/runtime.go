package internal

import (
	"fmt"
	"sync"

	theme "github.com/xiaowumin-mark/FluxUI/theme"

	giofont "gioui.org/font"
	"gioui.org/font/gofont"
	gioText "gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// Runtime 持有跨 frame 的稳定数据。
type Runtime struct {
	mu          sync.Mutex
	memory      map[MemoryKey]any
	activeMem   map[MemoryKey]struct{}
	trackingMem bool
	theme       *theme.Theme
	material    *material.Theme
	invalidate  func()
	windowCtrl  WindowController
	effects     map[MemoryKey]*effectSlot
	activeFx    map[MemoryKey]struct{}
	pendingFx   []func()

	hookCounts               map[string]int
	prevHookCounts           map[string]int
	hookCountIDs             map[PathID]int
	prevHookCountIDs         map[PathID]int
	windowDragAreaActive     bool
	nativeWindowActionRouter bool
	nativeWindowActions      []NativeWindowActionRegion

	// hookCounts and prevHookCounts enforce React's "Rules of Hooks":
	// hooks must always be called in the same count and order every frame.
	// BeginFrame snapshots the previous frame's counts into prevHookCounts;
	// EndFrame panics if any path rendered a different number of hooks —
	// this means hooks were called conditionally (inside if/for/switch).
	// These fields are NOT related to click counting or user-event tracking.
	hookStore  *HookStore
	perf       runtimePerfState
	interact   runtimeInteractionState
	render     runtimeRenderCache
	events     runtimeEventState
	pathIDs    map[pathLookupKey]PathID
	pathDebug  map[PathID]*pathDebugEntry
	nextPathID PathID
}

type effectSlot struct {
	initialized    bool
	hasDeps        bool
	deps           []any
	cleanup        func()
	pending        bool
	pendingHasDeps bool
	pendingDeps    []any
	pendingSetup   EffectSetup
}

// EffectSetup defines post-frame side effects with an optional cleanup.
type EffectSetup func() func()

// NewRuntime 创建运行时。
func NewRuntime(th *theme.Theme) *Runtime {
	if th == nil {
		th = theme.Default()
	}

	mt := material.NewTheme()
	shaper, err := th.BuildShaper()
	if err != nil || shaper == nil {
		shaper = gioText.NewShaper(gioText.WithCollection(gofont.Collection()))
	}
	mt.Shaper = shaper
	mt.Fg = th.TextColor
	mt.Bg = th.Surface
	mt.ContrastBg = th.Primary
	mt.ContrastFg = th.TextOnPrimary
	mt.TextSize = unit.Sp(th.TextSize)
	mt.Face = giofont.Typeface(th.DefaultFont.Normalize().Family)

	return &Runtime{
		memory:           make(map[MemoryKey]any),
		activeMem:        make(map[MemoryKey]struct{}),
		theme:            th,
		material:         mt,
		effects:          make(map[MemoryKey]*effectSlot),
		activeFx:         make(map[MemoryKey]struct{}),
		hookCounts:       make(map[string]int),
		prevHookCounts:   make(map[string]int),
		hookCountIDs:     make(map[PathID]int),
		prevHookCountIDs: make(map[PathID]int),
		hookStore:        NewHookStore(),
		pathDebug:        map[PathID]*pathDebugEntry{rootPathID: &pathDebugEntry{readable: "root"}},
		nextPathID:       rootPathID + 1,
	}
}

// Theme 返回当前主题。
func (r *Runtime) Theme() *theme.Theme {
	return r.theme
}

// MaterialTheme 返回内部使用的 Gio 主题。
func (r *Runtime) MaterialTheme() *material.Theme {
	return r.material
}

// SetInvalidator 绑定窗口重绘函数。
func (r *Runtime) SetInvalidator(fn func()) {
	r.mu.Lock()
	r.invalidate = fn
	r.mu.Unlock()
}

// RequestRedraw 请求窗口重绘。可从任意 goroutine 安全调用。
func (r *Runtime) RequestRedraw() {
	r.RequestRedrawReason("RequestRedraw")
}

func (r *Runtime) requestRedraw() {
	if r == nil {
		return
	}
	r.mu.Lock()
	fn := r.invalidate
	r.mu.Unlock()
	if fn != nil {
		fn()
	}
}

// SetWindowController 绑定当前窗口控制器。
func (r *Runtime) SetWindowController(controller WindowController) {
	r.windowCtrl = controller
}

// WindowController 返回当前窗口控制器。
func (r *Runtime) WindowController() WindowController {
	return r.windowCtrl
}

// BeginFrame resets per-frame bookkeeping for the next render pass.
// It snapshots the current frame's hookCounts into prevHookCounts
// (consumed by EndFrame to detect conditional hook calls) and clears
// hookCounts, active effects, and pending side-effects for the new frame.
func (r *Runtime) BeginFrame() {
	if r == nil {
		return
	}
	r.beginInteractionFrame()
	r.beginPerfFrame()
	r.beginRenderCacheFrame()
	r.beginEventFrame()
	if r.hookStore != nil {
		r.hookStore.BeginFrame()
	}
	clear(r.activeMem)
	r.trackingMem = true
	r.hookCounts, r.prevHookCounts = r.prevHookCounts, r.hookCounts
	clear(r.hookCounts)
	r.hookCountIDs, r.prevHookCountIDs = r.prevHookCountIDs, r.hookCountIDs
	clear(r.hookCountIDs)
	clear(r.activeFx)
	for _, slot := range r.effects {
		if slot == nil || !slot.pending {
			continue
		}
		slot.pending = false
		slot.pendingHasDeps = false
		slot.pendingDeps = nil
		slot.pendingSetup = nil
	}
	clear(r.pendingFx)
	r.pendingFx = r.pendingFx[:0]
	r.windowDragAreaActive = false
	r.nativeWindowActionRouter = false
	r.nativeWindowActions = r.nativeWindowActions[:0]
}

// EndFrame runs queued effects, cleans up unmounted effects, and validates
// hook consistency. The hook-count check enforces React's "Rules of Hooks":
// if any context path rendered a different number of hooks compared to the
// previous frame, it means hooks were called conditionally (inside if/for/switch),
// which breaks hook state identity and causes subtle bugs.
func (r *Runtime) EndFrame() {
	if r == nil {
		return
	}
	r.endEventFrame()
	if r.hookStore != nil {
		r.hookStore.EndFrame()
	}

	for path, count := range r.hookCounts {
		if prev, ok := r.prevHookCounts[path]; ok && prev != count {
			panic(fmt.Sprintf(
				"FluxUI: path %q 本帧渲染了 %d 个 hook，但上一帧为 %d —— "+
					"hooks 不得在条件语句(if/for/switch)中调用，调用数量和顺序必须每帧一致。\n"+
					"       path %q rendered %d hooks this frame but %d last frame — "+
					"hooks must not be called inside if/for/switch or any conditional block",
				path, count, prev, path, count, prev,
			))
		}
	}
	for pathID, count := range r.hookCountIDs {
		if prev, ok := r.prevHookCountIDs[pathID]; ok && prev != count {
			path := r.DebugPath(pathID)
			panic(fmt.Sprintf(
				"FluxUI: path %q 本帧渲染了 %d 个 hook，但上一帧为 %d -- "+
					"hooks 不得在条件语句(if/for/switch)中调用，调用数量和顺序必须每帧一致。\n"+
					"       path %q rendered %d hooks this frame but %d last frame -- "+
					"hooks must not be called inside if/for/switch or any conditional block",
				path, count, prev, path, count, prev,
			))
		}
	}

	for key, slot := range r.effects {
		if _, ok := r.activeFx[key]; ok {
			continue
		}
		if slot != nil && slot.cleanup != nil {
			slot.cleanup()
			slot.cleanup = nil
		}
		delete(r.effects, key)
	}

	for _, run := range r.pendingFx {
		if run != nil {
			run()
		}
	}
	clear(r.pendingFx)
	r.pendingFx = r.pendingFx[:0]

	r.sweepInactiveMemory()
	r.endRenderCacheFrame()
	r.trackingMem = false
	r.endInteractionFrame()
	r.endPerfFrame()
}

// Dispose releases runtime resources and effect cleanups.
func (r *Runtime) Dispose() {
	if r == nil {
		return
	}
	if r.hookStore != nil {
		r.hookStore.Dispose()
	}
	for key, slot := range r.effects {
		if slot != nil && slot.cleanup != nil {
			slot.cleanup()
			slot.cleanup = nil
		}
		delete(r.effects, key)
	}
	clear(r.memory)
	clear(r.activeMem)
	r.disposeRenderCache()
	r.trackingMem = false
	r.pendingFx = nil
	clear(r.activeFx)
}

// HookStore returns the experimental component hook slot store.
func (r *Runtime) HookStore() *HookStore {
	if r == nil {
		return nil
	}
	return r.hookStore
}

// UseEffect registers a post-frame side effect bound to a stable key.
// hasDeps=false means "run every frame".
// hasDeps=true means "run on mount and whenever deps change".
func (r *Runtime) UseEffect(key string, hasDeps bool, deps []any, setup EffectSetup) {
	r.UseEffectKey(memoryKeyString(key), hasDeps, deps, setup)
}

func (r *Runtime) UseEffectKey(key MemoryKey, hasDeps bool, deps []any, setup EffectSetup) {
	if r == nil || !key.valid() || setup == nil {
		return
	}
	slot, ok := r.effects[key]
	if !ok || slot == nil {
		slot = &effectSlot{}
		r.effects[key] = slot
	}

	r.activeFx[key] = struct{}{}

	nextDeps := CloneDeps(deps)
	if slot.pending {
		slot.pendingHasDeps = hasDeps
		slot.pendingDeps = nextDeps
		slot.pendingSetup = setup
		return
	}
	shouldRun := shouldRunEffect(slot, hasDeps, nextDeps)
	if !shouldRun {
		return
	}

	slot.pending = true
	slot.pendingHasDeps = hasDeps
	slot.pendingDeps = nextDeps
	slot.pendingSetup = setup
	r.pendingFx = append(r.pendingFx, func() {
		pendingHasDeps := slot.pendingHasDeps
		pendingDeps := slot.pendingDeps
		pendingSetup := slot.pendingSetup
		slot.pending = false
		slot.pendingHasDeps = false
		slot.pendingDeps = nil
		slot.pendingSetup = nil
		if !shouldRunEffect(slot, pendingHasDeps, pendingDeps) {
			return
		}
		if slot.cleanup != nil {
			slot.cleanup()
			slot.cleanup = nil
		}
		slot.initialized = true
		slot.hasDeps = pendingHasDeps
		slot.deps = pendingDeps
		slot.cleanup = pendingSetup()
	})
}

// RecordHookCount stores the number of hooks rendered at the given context path
// in the current frame. EndFrame later compares this against prevHookCounts to
// enforce React's "Rules of Hooks": any path that changes count between frames
// indicates hooks were called conditionally (inside if/for/switch), which breaks
// hook state identity.
func (r *Runtime) RecordHookCount(path string, count int) {
	if r == nil {
		return
	}
	r.hookCounts[path] = count
}

func (r *Runtime) RecordHookCountID(path PathID, count int) {
	if r == nil {
		return
	}
	r.hookCountIDs[normalizePathID(path)] = count
}

// RegisterWindowDragArea marks the current frame as containing at least one
// system window move region.
func (r *Runtime) RegisterWindowDragArea() {
	if r == nil {
		return
	}
	r.windowDragAreaActive = true
}

// WindowDragAreaActive reports whether the current frame registered a window
// move region.
func (r *Runtime) WindowDragAreaActive() bool {
	return r != nil && r.windowDragAreaActive
}

func (r *Runtime) remember(key string, factory func() any) any {
	return r.rememberKey(memoryKeyString(key), factory)
}

func (r *Runtime) memoryValue(key string) (any, bool) {
	return r.memoryValueKey(memoryKeyString(key))
}

func (r *Runtime) forgetMemory(key string) {
	r.forgetMemoryKey(memoryKeyString(key))
}

func (r *Runtime) rememberKey(key MemoryKey, factory func() any) any {
	if r == nil || !key.valid() {
		if factory == nil {
			return nil
		}
		return factory()
	}
	r.markMemoryActive(key)
	if value, ok := r.memory[key]; ok {
		return value
	}
	value := factory()
	r.memory[key] = value
	return value
}

func (r *Runtime) memoryValueKey(key MemoryKey) (any, bool) {
	if r == nil || !key.valid() {
		return nil, false
	}
	value, ok := r.memory[key]
	if ok {
		r.markMemoryActive(key)
	}
	return value, ok
}

func (r *Runtime) forgetMemoryKey(key MemoryKey) {
	if r == nil || !key.valid() {
		return
	}
	delete(r.memory, key)
	delete(r.activeMem, key)
}

func (r *Runtime) markMemoryActive(key MemoryKey) {
	if r == nil || !r.trackingMem || !key.valid() {
		return
	}
	r.activeMem[key] = struct{}{}
}

func (r *Runtime) sweepInactiveMemory() {
	if r == nil || !r.trackingMem {
		return
	}
	for key := range r.memory {
		if _, ok := r.activeMem[key]; !ok {
			delete(r.memory, key)
		}
	}
}

func shouldRunEffect(slot *effectSlot, hasDeps bool, nextDeps []any) bool {
	if slot == nil || !slot.initialized {
		return true
	}
	if !hasDeps {
		return true
	}
	if !slot.hasDeps {
		return true
	}
	return !DepsEqual(slot.deps, nextDeps)
}

func depsEqual(a, b []any) bool {
	return DepsEqual(a, b)
}

func cloneDeps(deps []any) []any {
	return CloneDeps(deps)
}
