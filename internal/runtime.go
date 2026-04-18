package internal

import (
	"reflect"

	theme "github.com/xiaowumin-mark/FluxUI/theme"

	giofont "gioui.org/font"
	"gioui.org/font/gofont"
	gioText "gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// Runtime 持有跨 frame 的稳定数据。
type Runtime struct {
	memory     map[string]any
	theme      *theme.Theme
	material   *material.Theme
	invalidate func()
	windowCtrl WindowController
	effects    map[string]*effectSlot
	activeFx   map[string]struct{}
	pendingFx  []func()
}

type effectSlot struct {
	initialized bool
	hasDeps     bool
	deps        []any
	cleanup     func()
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
		memory:   make(map[string]any),
		theme:    th,
		material: mt,
		effects:  make(map[string]*effectSlot),
		activeFx: make(map[string]struct{}),
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
	r.invalidate = fn
}

// RequestRedraw 请求窗口重绘。
func (r *Runtime) RequestRedraw() {
	if r.invalidate != nil {
		r.invalidate()
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

// BeginFrame resets per-frame hook bookkeeping.
func (r *Runtime) BeginFrame() {
	if r == nil {
		return
	}
	clear(r.activeFx)
	r.pendingFx = r.pendingFx[:0]
}

// EndFrame runs queued effects and cleans up unmounted effects.
func (r *Runtime) EndFrame() {
	if r == nil {
		return
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
	r.pendingFx = r.pendingFx[:0]
}

// Dispose releases runtime resources and effect cleanups.
func (r *Runtime) Dispose() {
	if r == nil {
		return
	}
	for key, slot := range r.effects {
		if slot != nil && slot.cleanup != nil {
			slot.cleanup()
			slot.cleanup = nil
		}
		delete(r.effects, key)
	}
	r.pendingFx = nil
	clear(r.activeFx)
}

// UseEffect registers a post-frame side effect bound to a stable key.
// hasDeps=false means "run every frame".
// hasDeps=true means "run on mount and whenever deps change".
func (r *Runtime) UseEffect(key string, hasDeps bool, deps []any, setup EffectSetup) {
	if r == nil || key == "" || setup == nil {
		return
	}
	slot, ok := r.effects[key]
	if !ok || slot == nil {
		slot = &effectSlot{}
		r.effects[key] = slot
	}

	r.activeFx[key] = struct{}{}

	nextDeps := cloneDeps(deps)
	shouldRun := shouldRunEffect(slot, hasDeps, nextDeps)
	if !shouldRun {
		return
	}

	r.pendingFx = append(r.pendingFx, func() {
		if slot.cleanup != nil {
			slot.cleanup()
			slot.cleanup = nil
		}
		slot.initialized = true
		slot.hasDeps = hasDeps
		slot.deps = nextDeps
		slot.cleanup = setup()
	})
}

func (r *Runtime) remember(key string, factory func() any) any {
	if value, ok := r.memory[key]; ok {
		return value
	}
	value := factory()
	r.memory[key] = value
	return value
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
	return !depsEqual(slot.deps, nextDeps)
}

func depsEqual(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !reflect.DeepEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func cloneDeps(deps []any) []any {
	if len(deps) == 0 {
		return nil
	}
	out := make([]any, len(deps))
	copy(out, deps)
	return out
}
