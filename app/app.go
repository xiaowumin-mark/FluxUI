package app

import (
	"errors"
	"fmt"
	"image"
	"math"
	"os"
	"runtime"
	"runtime/debug"
	"sort"
	"sync"
	"sync/atomic"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
	theme "github.com/xiaowumin-mark/FluxUI/theme"
	widget "github.com/xiaowumin-mark/FluxUI/widget"

	gioApp "gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/op"
	"gioui.org/unit"
)

// Builder 定义应用根组件构造函数。
type Builder func(ctx *internal.Context) widget.Widget

// Option 定义应用启动配置。
type Option func(*Application)

// WindowID 是窗口唯一标识。
type WindowID uint64

// WindowHiddenMemoryPolicy 定义窗口隐藏后的渲染内存策略。
type WindowHiddenMemoryPolicy int

const (
	// WindowHiddenMemoryReleaseTransient 会在窗口隐藏后暂停 FluxUI 渲染并释放临时内存。
	WindowHiddenMemoryReleaseTransient WindowHiddenMemoryPolicy = iota
	// WindowHiddenMemoryKeepRenderingState 保留隐藏窗口的渲染状态，不主动释放临时内存。
	WindowHiddenMemoryKeepRenderingState
)

// WindowState 是 FluxUI 运行时维护的窗口状态快照。
//
// Width/Height/MinWidth/MinHeight/MaxWidth/MaxHeight 使用 dp 单位。窗口实际被
// 平台调整后，FluxUI 会在收到 frame/config 事件时尽量同步该快照。
type WindowState struct {
	ID                 WindowID
	Title              string
	Width              int
	Height             int
	MinWidth           int
	MinHeight          int
	MaxWidth           int
	MaxHeight          int
	Visible            bool
	AlwaysOnTop        bool
	RenderSuspended    bool
	HiddenMemoryPolicy WindowHiddenMemoryPolicy
	Minimized          bool
	Maximized          bool
	Fullscreen         bool
	Focused            bool
	Decorated          bool
	Alive              bool
}

// WindowEventKind 是窗口生命周期或状态变化事件类型。
type WindowEventKind string

const (
	WindowEventSizeChanged  WindowEventKind = "size_changed"
	WindowEventFocusChanged WindowEventKind = "focus_changed"
	WindowEventStateChanged WindowEventKind = "state_changed"
	WindowEventClosed       WindowEventKind = "closed"
)

// WindowEvent 记录窗口状态变化。State 是事件发生后的窗口状态快照。
type WindowEvent struct {
	Window WindowHandle
	Kind   WindowEventKind
	State  WindowState
}

// WindowHandle 表示运行中的窗口句柄。
type WindowHandle struct {
	id WindowID
}

// ID 返回窗口 ID。
func (h WindowHandle) ID() WindowID {
	return h.id
}

// NativeHandle returns the platform-native window handle when available.
//
// On Windows this is the HWND. Unsupported platforms or windows that have not
// received a native view event return false.
func (h WindowHandle) NativeHandle() (uintptr, bool) {
	if h.id == 0 {
		return 0, false
	}
	entry, ok := lookupWindow(h.id)
	if !ok || entry == nil || !entry.alive.Load() {
		return 0, false
	}
	entry.mu.RLock()
	handle := entry.nativeHandle
	entry.mu.RUnlock()
	return handle, handle != 0
}

// IsAlive 返回窗口是否仍在运行。
func (h WindowHandle) IsAlive() bool {
	entry, ok := lookupWindow(h.id)
	return ok && entry != nil && entry.alive.Load()
}

// State 返回窗口当前状态快照。窗口不存在或已关闭时返回 false。
func (h WindowHandle) State() (WindowState, bool) {
	if h.id == 0 {
		return WindowState{}, false
	}
	entry, ok := lookupWindow(h.id)
	if !ok || entry == nil || !entry.alive.Load() {
		return WindowState{}, false
	}
	return entry.snapshot(), true
}

// PollEvents 返回并清空窗口当前积累的事件。
func (h WindowHandle) PollEvents() []WindowEvent {
	if h.id == 0 {
		return nil
	}
	entry, ok := lookupWindow(h.id)
	if !ok || entry == nil {
		return nil
	}
	return entry.drainEvents()
}

// Close 请求关闭窗口。
func (h WindowHandle) Close() bool {
	return h.perform(system.ActionClose)
}

// Show 显示窗口。
func (h WindowHandle) Show() bool {
	return h.setVisible(true)
}

// Hide 隐藏窗口。
func (h WindowHandle) Hide() bool {
	return h.setVisible(false)
}

// SetHiddenMemoryPolicy 设置窗口隐藏后的渲染内存策略。
func (h WindowHandle) SetHiddenMemoryPolicy(policy WindowHiddenMemoryPolicy) bool {
	if !validWindowHiddenMemoryPolicy(policy) {
		return false
	}
	return h.apply(func(entry *windowEntry) {
		trim := false
		entry.updateAndEmit(func(state *WindowState) {
			before := state.RenderSuspended
			state.HiddenMemoryPolicy = policy
			applyWindowRenderSuspendedState(state)
			trim = !before && state.RenderSuspended
		})
		if trim {
			entry.requestHiddenMemoryTrim()
		}
	})
}

// Minimize 最小化窗口。
func (h WindowHandle) Minimize() bool {
	return h.applyMode(gioApp.Minimized)
}

// Maximize 最大化窗口。
func (h WindowHandle) Maximize() bool {
	return h.applyMode(gioApp.Maximized)
}

// Restore 还原窗口为普通模式。
func (h WindowHandle) Restore() bool {
	return h.applyMode(gioApp.Windowed)
}

// Fullscreen 切换窗口为全屏模式。
func (h WindowHandle) Fullscreen() bool {
	return h.applyMode(gioApp.Fullscreen)
}

// Raise 请求将窗口一次性前置。
func (h WindowHandle) Raise() bool {
	return h.perform(system.ActionRaise)
}

// SetAlwaysOnTop 设置窗口是否持续置顶。
func (h WindowHandle) SetAlwaysOnTop(always bool) bool {
	entry, ok := h.liveEntry()
	if !ok {
		return false
	}
	if !setNativeWindowAlwaysOnTop(entry.nativeHandleSnapshot(), always) {
		return false
	}
	entry.updateAndEmit(func(state *WindowState) {
		state.AlwaysOnTop = always
	})
	entry.win.Invalidate()
	return true
}

// Center 请求将窗口居中。
func (h WindowHandle) Center() bool {
	return h.perform(system.ActionCenter)
}

// SetTitle 更新窗口标题。
func (h WindowHandle) SetTitle(title string) bool {
	title = normalizeTitle(title)
	return h.apply(func(entry *windowEntry) {
		entry.win.Option(gioApp.Title(title))
		entry.updateState(func(state *WindowState) {
			state.Title = title
		})
	})
}

// SetSize 更新窗口尺寸（单位为 dp）。
func (h WindowHandle) SetSize(width, height int) bool {
	if width <= 0 || height <= 0 {
		return false
	}
	return h.apply(func(entry *windowEntry) {
		opts := []gioApp.Option{gioApp.Size(unit.Dp(width), unit.Dp(height))}
		opts = append(opts, entry.constraintOptions()...)
		entry.win.Option(opts...)
		entry.updateState(func(state *WindowState) {
			state.Width = width
			state.Height = height
			applyWindowModeState(state, gioApp.Windowed)
		})
	})
}

// SetMinSize 更新窗口最小尺寸（单位为 dp）。
func (h WindowHandle) SetMinSize(width, height int) bool {
	if width <= 0 || height <= 0 {
		return false
	}
	return h.apply(func(entry *windowEntry) {
		if !entry.snapshot().Fullscreen {
			entry.win.Option(gioApp.MinSize(unit.Dp(width), unit.Dp(height)))
		}
		entry.updateState(func(state *WindowState) {
			state.MinWidth = width
			state.MinHeight = height
		})
	})
}

// SetMaxSize 更新窗口最大尺寸（单位为 dp）。
func (h WindowHandle) SetMaxSize(width, height int) bool {
	if width <= 0 || height <= 0 {
		return false
	}
	return h.apply(func(entry *windowEntry) {
		if !entry.snapshot().Fullscreen {
			entry.win.Option(gioApp.MaxSize(unit.Dp(width), unit.Dp(height)))
		}
		entry.updateState(func(state *WindowState) {
			state.MaxWidth = width
			state.MaxHeight = height
		})
		entry.syncNativeMaximizeAvailability()
	})
}

// Invalidate 请求窗口重绘。
func (h WindowHandle) Invalidate() bool {
	return h.apply(func(entry *windowEntry) {
		if entry.renderSuspendedSnapshot() {
			return
		}
		entry.win.Invalidate()
	})
}

func (h WindowHandle) applyOption(opts ...gioApp.Option) bool {
	if len(opts) == 0 {
		return false
	}
	return h.apply(func(entry *windowEntry) {
		entry.win.Option(opts...)
	})
}

func (h WindowHandle) applyMode(mode gioApp.WindowMode) bool {
	entry, ok := h.liveEntry()
	if !ok {
		return false
	}
	if mode == gioApp.Maximized && entry.maximizeDisabledByConstraints() {
		return false
	}
	entry.win.Option(entry.modeOptions(mode)...)
	entry.updateState(func(state *WindowState) {
		applyWindowModeState(state, mode)
	})
	return true
}

func (h WindowHandle) setVisible(visible bool) bool {
	entry, ok := h.liveEntry()
	if !ok {
		return false
	}
	if !setNativeWindowVisible(entry.nativeHandleSnapshot(), visible) {
		return false
	}
	trim := false
	entry.updateAndEmit(func(state *WindowState) {
		before := state.RenderSuspended
		state.Visible = visible
		applyWindowRenderSuspendedState(state)
		trim = !before && state.RenderSuspended
	})
	if trim {
		entry.requestHiddenMemoryTrim()
	}
	if visible {
		entry.win.Invalidate()
	}
	return true
}

func (h WindowHandle) perform(actions system.Action) bool {
	if actions == 0 {
		return false
	}
	return h.apply(func(entry *windowEntry) {
		entry.win.Perform(actions)
	})
}

func (h WindowHandle) apply(fn func(entry *windowEntry)) bool {
	if fn == nil {
		return false
	}
	entry, ok := h.liveEntry()
	if !ok {
		return false
	}
	fn(entry)
	return true
}

func (h WindowHandle) liveEntry() (*windowEntry, bool) {
	if h.id == 0 {
		return nil, false
	}
	entry, ok := lookupWindow(h.id)
	if !ok || entry == nil || !entry.alive.Load() {
		return nil, false
	}
	return entry, true
}

// Application 是 Gio window loop 的封装。
type Application struct {
	Title              string
	Width              int
	Height             int
	MinWidth           int
	MinHeight          int
	MaxWidth           int
	MaxHeight          int
	HiddenMemoryPolicy WindowHiddenMemoryPolicy
	Theme              *theme.Theme
	Root               Builder
}

// WindowSpec 是多窗口运行时的窗口配置。
type WindowSpec struct {
	Root    Builder
	Options []Option
}

// New 创建应用实例。
func New(root Builder, opts ...Option) *Application {
	app := &Application{
		Title:              "FluxUI",
		Width:              420,
		Height:             240,
		HiddenMemoryPolicy: WindowHiddenMemoryReleaseTransient,
		Theme:              theme.Default(),
		Root:               root,
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

// Window 创建多窗口启动中的单个窗口定义。
func Window(root Builder, opts ...Option) WindowSpec {
	cloned := make([]Option, len(opts))
	copy(cloned, opts)
	return WindowSpec{
		Root:    root,
		Options: cloned,
	}
}

// Title 设置窗口标题。
func Title(value string) Option {
	return func(app *Application) {
		app.Title = value
	}
}

// Size 设置窗口初始尺寸。
func Size(width, height int) Option {
	return func(app *Application) {
		app.Width = width
		app.Height = height
	}
}

// MinSize 设置窗口初始最小尺寸。
func MinSize(width, height int) Option {
	return func(app *Application) {
		app.MinWidth = width
		app.MinHeight = height
	}
}

// MaxSize 设置窗口初始最大尺寸。
func MaxSize(width, height int) Option {
	return func(app *Application) {
		app.MaxWidth = width
		app.MaxHeight = height
	}
}

// HiddenMemoryPolicy 设置窗口隐藏后的渲染内存策略。
func HiddenMemoryPolicy(policy WindowHiddenMemoryPolicy) Option {
	return func(app *Application) {
		if validWindowHiddenMemoryPolicy(policy) {
			app.HiddenMemoryPolicy = policy
		}
	}
}

// WithTheme 覆盖应用主题。
func WithTheme(th *theme.Theme) Option {
	return func(app *Application) {
		if th != nil {
			app.Theme = th
		}
	}
}

// WithDensity 设置主题密度。默认密度保持 MD3 可触控目标，compact 用于更紧凑的桌面 UI。
func WithDensity(density theme.DensityScale) Option {
	return func(app *Application) {
		if app.Theme == nil {
			app.Theme = theme.Default()
		}
		app.Theme.SetDensity(density)
	}
}

// WithFonts 追加全局字体集合。
func WithFonts(faces ...theme.FontFace) Option {
	return func(app *Application) {
		if len(faces) == 0 {
			return
		}
		if app.Theme == nil {
			app.Theme = theme.Default()
		}
		app.Theme.AddFonts(faces...)
	}
}

// WithDefaultFont 设置全局默认字体。
func WithDefaultFont(spec theme.FontSpec) Option {
	return func(app *Application) {
		if app.Theme == nil {
			app.Theme = theme.Default()
		}
		app.Theme.SetDefaultFont(spec)
	}
}

// WithSystemFonts 控制是否启用系统字体回退。
func WithSystemFonts(enabled bool) Option {
	return func(app *Application) {
		if app.Theme == nil {
			app.Theme = theme.Default()
		}
		app.Theme.SetUseSystemFonts(enabled)
	}
}

// ListWindows 返回当前仍然存活的窗口句柄列表。
func ListWindows() []WindowHandle {
	windowRegistryMu.RLock()
	defer windowRegistryMu.RUnlock()

	ids := make([]WindowID, 0, len(windowRegistry))
	for id, entry := range windowRegistry {
		if entry != nil && entry.alive.Load() {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	handles := make([]WindowHandle, 0, len(ids))
	for _, id := range ids {
		handles = append(handles, WindowHandle{id: id})
	}
	return handles
}

// GetWindow 按 ID 查询窗口句柄。
func GetWindow(id WindowID) (WindowHandle, bool) {
	entry, ok := lookupWindow(id)
	if !ok || entry == nil || !entry.alive.Load() {
		return WindowHandle{}, false
	}
	return WindowHandle{id: id}, true
}

// Run 启动窗口事件循环。
func (a *Application) Run() error {
	if a == nil {
		return errors.New("app: nil application")
	}

	width, height := normalizeSize(a.Width, a.Height)
	title := normalizeTitle(a.Title)
	th := a.Theme
	if th == nil {
		th = theme.Default()
	}

	w := new(gioApp.Window)
	opts := []gioApp.Option{
		gioApp.Title(title),
		gioApp.Size(unit.Dp(width), unit.Dp(height)),
	}
	if a.MinWidth > 0 && a.MinHeight > 0 {
		opts = append(opts, gioApp.MinSize(unit.Dp(a.MinWidth), unit.Dp(a.MinHeight)))
	}
	if a.MaxWidth > 0 && a.MaxHeight > 0 {
		opts = append(opts, gioApp.MaxSize(unit.Dp(a.MaxWidth), unit.Dp(a.MaxHeight)))
	}
	w.Option(opts...)

	windowID := nextWindowID()
	entry := &windowEntry{
		id:  windowID,
		win: w,
		state: WindowState{
			ID:                 windowID,
			Title:              title,
			Width:              width,
			Height:             height,
			MinWidth:           maxPositive(a.MinWidth),
			MinHeight:          maxPositive(a.MinHeight),
			MaxWidth:           maxPositive(a.MaxWidth),
			MaxHeight:          maxPositive(a.MaxHeight),
			Visible:            true,
			HiddenMemoryPolicy: normalizeWindowHiddenMemoryPolicy(a.HiddenMemoryPolicy),
			Decorated:          true,
			Alive:              true,
		},
	}
	entry.alive.Store(true)
	registerWindow(entry)
	defer func() {
		entry.alive.Store(false)
		entry.pushEvent(WindowEventClosed, func(state *WindowState) {
			state.Alive = false
			state.Visible = false
		})
		unregisterWindow(windowID)
	}()

	rt := internal.NewRuntime(th)
	rt.SetInvalidator(func() {
		if !entry.renderSuspendedSnapshot() {
			w.Invalidate()
		}
	})
	rt.SetWindowController(&windowController{
		handle: WindowHandle{id: windowID},
	})
	defer rt.Dispose()

	var ops op.Ops
	for {
		switch evt := w.Event().(type) {
		case gioApp.DestroyEvent:
			entry.alive.Store(false)
			entry.pushEvent(WindowEventClosed, func(state *WindowState) {
				state.Alive = false
				state.Visible = false
			})
			return evt.Err
		case gioApp.ConfigEvent:
			entry.updateFromConfig(evt.Config)
		case gioApp.ViewEvent:
			entry.updateNativeHandle(evt)
		case gioApp.FrameEvent:
			entry.updateFromFrame(evt.Size, evt.Metric)
			if entry.renderSuspendedSnapshot() {
				ops.Reset()
				entry.releaseHiddenMemoryIfRequested()
				evt.Frame(&ops)
				continue
			}
			gtx := gioApp.NewContext(&ops, evt)
			rt.BeginFrame()
			ctx := rt.Frame(gtx)
			buildCtx := ctx.Scope("build")
			treeCtx := ctx.Scope("tree")

			if a.Root != nil {
				if root := a.Root(buildCtx); root != nil {
					root.Layout(treeCtx.Child(0))
				}
			}
			rt.EndFrame()

			if entry.renderSuspendedSnapshot() {
				ops.Reset()
				entry.releaseHiddenMemoryIfRequested()
			}
			evt.Frame(gtx.Ops)
		}
	}
}

// Run 直接创建并启动应用。
func Run(root Builder, opts ...Option) error {
	return runSpecs(Window(root, opts...))
}

// RunMulti 同时启动多个窗口（桌面端）。
func RunMulti(windows ...WindowSpec) error {
	if len(windows) == 0 {
		return errors.New("app: RunMulti requires at least one window")
	}

	if !supportsMultiWindow() && len(windows) > 1 {
		return fmt.Errorf("app: multi-window is not supported on %s", runtime.GOOS)
	}
	if !supportsMultiWindow() && len(windows) == 1 {
		first := windows[0]
		return Run(first.Root, first.Options...)
	}

	return runSpecs(windows...)
}

func runSpecs(specs ...WindowSpec) error {
	if len(specs) == 0 {
		return errors.New("app: no windows to run")
	}

	done := make(chan error, len(specs))
	for _, spec := range specs {
		s := spec
		go func() {
			done <- New(s.Root, s.Options...).Run()
		}()
	}

	result := make(chan error, 1)
	go func() {
		var firstErr error
		for i := 0; i < len(specs); i++ {
			if err := <-done; err != nil && firstErr == nil {
				firstErr = err
			}
		}
		result <- firstErr

		if !shouldAutoExit() {
			return
		}
		if firstErr != nil {
			_, _ = fmt.Fprintln(os.Stderr, firstErr)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	gioApp.Main()
	return <-result
}

func supportsMultiWindow() bool {
	switch runtime.GOOS {
	case "android", "ios", "js":
		return false
	default:
		return true
	}
}

func shouldAutoExit() bool {
	switch runtime.GOOS {
	case "android", "ios", "js":
		return false
	default:
		return true
	}
}

type windowController struct {
	handle WindowHandle
}

func (c *windowController) WindowID() internal.WindowID {
	return internal.WindowID(c.handle.ID())
}

func (c *windowController) Close() bool {
	return c.handle.Close()
}

func (c *windowController) Show() bool {
	return c.handle.Show()
}

func (c *windowController) Hide() bool {
	return c.handle.Hide()
}

func (c *windowController) SetHiddenMemoryPolicy(policy internal.WindowHiddenMemoryPolicy) bool {
	return c.handle.SetHiddenMemoryPolicy(WindowHiddenMemoryPolicy(policy))
}

func (c *windowController) Minimize() bool {
	return c.handle.Minimize()
}

func (c *windowController) Maximize() bool {
	return c.handle.Maximize()
}

func (c *windowController) Restore() bool {
	return c.handle.Restore()
}

func (c *windowController) Fullscreen() bool {
	return c.handle.Fullscreen()
}

func (c *windowController) Raise() bool {
	return c.handle.Raise()
}

func (c *windowController) SetAlwaysOnTop(always bool) bool {
	return c.handle.SetAlwaysOnTop(always)
}

func (c *windowController) Center() bool {
	return c.handle.Center()
}

func (c *windowController) SetTitle(title string) bool {
	return c.handle.SetTitle(title)
}

func (c *windowController) SetSize(width, height int) bool {
	return c.handle.SetSize(width, height)
}

func (c *windowController) SetMinSize(width, height int) bool {
	return c.handle.SetMinSize(width, height)
}

func (c *windowController) SetMaxSize(width, height int) bool {
	return c.handle.SetMaxSize(width, height)
}

func (c *windowController) Invalidate() bool {
	return c.handle.Invalidate()
}

func (c *windowController) IsAlive() bool {
	return c.handle.IsAlive()
}

type windowEntry struct {
	id                      WindowID
	win                     *gioApp.Window
	alive                   atomic.Bool
	trimHiddenMemoryPending atomic.Bool
	mu                      sync.RWMutex
	metric                  unit.Metric
	nativeHandle            uintptr
	nativeMaximizeHandle    uintptr
	nativeMaximizeSynced    bool
	nativeMaximizeEnabled   bool
	state                   WindowState
	events                  []WindowEvent
}

var (
	windowRegistryMu sync.RWMutex
	windowRegistry   = make(map[WindowID]*windowEntry)
	windowCounter    atomic.Uint64
)

func nextWindowID() WindowID {
	return WindowID(windowCounter.Add(1))
}

func registerWindow(entry *windowEntry) {
	if entry == nil {
		return
	}
	windowRegistryMu.Lock()
	windowRegistry[entry.id] = entry
	windowRegistryMu.Unlock()
}

func unregisterWindow(id WindowID) {
	windowRegistryMu.Lock()
	delete(windowRegistry, id)
	windowRegistryMu.Unlock()
}

func lookupWindow(id WindowID) (*windowEntry, bool) {
	windowRegistryMu.RLock()
	entry, ok := windowRegistry[id]
	windowRegistryMu.RUnlock()
	return entry, ok
}

func (entry *windowEntry) snapshot() WindowState {
	entry.mu.RLock()
	state := entry.state
	entry.mu.RUnlock()
	state.Alive = entry.alive.Load()
	return state
}

func (entry *windowEntry) nativeHandleSnapshot() uintptr {
	entry.mu.RLock()
	handle := entry.nativeHandle
	entry.mu.RUnlock()
	return handle
}

func (entry *windowEntry) renderSuspendedSnapshot() bool {
	entry.mu.RLock()
	suspended := entry.state.RenderSuspended
	entry.mu.RUnlock()
	return suspended
}

func (entry *windowEntry) maximizeDisabledByConstraints() bool {
	return hasWindowMaxSize(entry.snapshot())
}

func (entry *windowEntry) syncNativeMaximizeAvailability() {
	entry.mu.Lock()
	handle := entry.nativeHandle
	if handle == 0 {
		entry.mu.Unlock()
		return
	}
	enabled := !hasWindowMaxSize(entry.state)
	if enabled && !entry.nativeMaximizeSynced {
		entry.mu.Unlock()
		return
	}
	if entry.nativeMaximizeSynced &&
		entry.nativeMaximizeHandle == handle &&
		entry.nativeMaximizeEnabled == enabled {
		entry.mu.Unlock()
		return
	}
	entry.nativeMaximizeHandle = handle
	entry.nativeMaximizeSynced = true
	entry.nativeMaximizeEnabled = enabled
	entry.mu.Unlock()

	setNativeWindowMaximizeEnabled(handle, enabled)
}

func (entry *windowEntry) requestHiddenMemoryTrim() {
	if entry == nil {
		return
	}
	entry.trimHiddenMemoryPending.Store(true)
}

func (entry *windowEntry) releaseHiddenMemoryIfRequested() {
	if entry == nil || !entry.trimHiddenMemoryPending.Swap(false) {
		return
	}
	go func() {
		runtime.GC()
		debug.FreeOSMemory()
	}()
}

func (entry *windowEntry) updateState(fn func(state *WindowState)) {
	if entry == nil || fn == nil {
		return
	}
	entry.mu.Lock()
	fn(&entry.state)
	entry.state.ID = entry.id
	entry.state.Alive = entry.alive.Load()
	entry.mu.Unlock()
}

func (entry *windowEntry) updateFromConfig(config gioApp.Config) {
	entry.updateAndEmit(func(state *WindowState) {
		wasFullscreen := state.Fullscreen
		if config.Title != "" {
			state.Title = config.Title
		}
		state.Focused = config.Focused
		state.Decorated = config.Decorated
		applyWindowModeState(state, config.Mode)
		updateDpSizeFromMetric(state, entry.metric, config.Size.X, config.Size.Y)
		if shouldSyncConfigConstraints(wasFullscreen, config) {
			updateDpMinSizeFromMetric(state, entry.metric, config.MinSize.X, config.MinSize.Y)
			updateDpMaxSizeFromMetric(state, entry.metric, config.MaxSize.X, config.MaxSize.Y)
		}
	})
}

func (entry *windowEntry) updateFromFrame(size image.Point, metric unit.Metric) {
	entry.updateAndEmit(func(state *WindowState) {
		entry.metric = metric
		updateDpSizeFromMetric(state, metric, size.X, size.Y)
	})
}

func (entry *windowEntry) updateAndEmit(fn func(state *WindowState)) {
	if entry == nil || fn == nil {
		return
	}

	entry.mu.Lock()
	before := entry.state
	fn(&entry.state)
	entry.state.ID = entry.id
	entry.state.Alive = entry.alive.Load()
	after := entry.state
	kinds := windowEventKinds(before, after)
	for _, kind := range kinds {
		entry.events = append(entry.events, WindowEvent{
			Window: WindowHandle{id: entry.id},
			Kind:   kind,
			State:  after,
		})
	}
	entry.mu.Unlock()
}

func (entry *windowEntry) pushEvent(kind WindowEventKind, fn func(state *WindowState)) {
	if entry == nil || kind == "" {
		return
	}
	entry.mu.Lock()
	if kind == WindowEventClosed {
		for _, event := range entry.events {
			if event.Kind == WindowEventClosed {
				entry.mu.Unlock()
				return
			}
		}
	}
	if fn != nil {
		fn(&entry.state)
	}
	entry.state.ID = entry.id
	entry.state.Alive = entry.alive.Load()
	entry.events = append(entry.events, WindowEvent{
		Window: WindowHandle{id: entry.id},
		Kind:   kind,
		State:  entry.state,
	})
	entry.mu.Unlock()
}

func (entry *windowEntry) drainEvents() []WindowEvent {
	entry.mu.Lock()
	events := append([]WindowEvent(nil), entry.events...)
	entry.events = nil
	entry.mu.Unlock()
	return events
}

func windowEventKinds(before, after WindowState) []WindowEventKind {
	kinds := make([]WindowEventKind, 0, 3)
	if before.Width != after.Width || before.Height != after.Height {
		kinds = append(kinds, WindowEventSizeChanged)
	}
	if before.Focused != after.Focused {
		kinds = append(kinds, WindowEventFocusChanged)
	}
	if before.Minimized != after.Minimized ||
		before.Maximized != after.Maximized ||
		before.Fullscreen != after.Fullscreen ||
		before.Decorated != after.Decorated ||
		before.Visible != after.Visible ||
		before.AlwaysOnTop != after.AlwaysOnTop ||
		before.RenderSuspended != after.RenderSuspended ||
		before.HiddenMemoryPolicy != after.HiddenMemoryPolicy ||
		before.Title != after.Title ||
		before.MinWidth != after.MinWidth ||
		before.MinHeight != after.MinHeight ||
		before.MaxWidth != after.MaxWidth ||
		before.MaxHeight != after.MaxHeight ||
		before.Alive != after.Alive {
		kinds = append(kinds, WindowEventStateChanged)
	}
	return kinds
}

func applyWindowModeState(state *WindowState, mode gioApp.WindowMode) {
	state.Minimized = mode == gioApp.Minimized
	state.Maximized = mode == gioApp.Maximized
	state.Fullscreen = mode == gioApp.Fullscreen
}

func hasWindowMaxSize(state WindowState) bool {
	return state.MaxWidth > 0 && state.MaxHeight > 0
}

func applyWindowRenderSuspendedState(state *WindowState) {
	state.RenderSuspended = !state.Visible && state.HiddenMemoryPolicy == WindowHiddenMemoryReleaseTransient
}

func validWindowHiddenMemoryPolicy(policy WindowHiddenMemoryPolicy) bool {
	switch policy {
	case WindowHiddenMemoryReleaseTransient, WindowHiddenMemoryKeepRenderingState:
		return true
	default:
		return false
	}
}

func normalizeWindowHiddenMemoryPolicy(policy WindowHiddenMemoryPolicy) WindowHiddenMemoryPolicy {
	if validWindowHiddenMemoryPolicy(policy) {
		return policy
	}
	return WindowHiddenMemoryReleaseTransient
}

func (entry *windowEntry) modeOptions(mode gioApp.WindowMode) []gioApp.Option {
	if mode == gioApp.Fullscreen {
		return []gioApp.Option{
			clearWindowConstraints(),
			mode.Option(),
		}
	}

	opts := []gioApp.Option{mode.Option()}
	return append(opts, entry.constraintOptions()...)
}

func (entry *windowEntry) constraintOptions() []gioApp.Option {
	state := entry.snapshot()
	opts := make([]gioApp.Option, 0, 2)
	if state.MinWidth > 0 && state.MinHeight > 0 {
		opts = append(opts, gioApp.MinSize(unit.Dp(state.MinWidth), unit.Dp(state.MinHeight)))
	}
	if state.MaxWidth > 0 && state.MaxHeight > 0 {
		opts = append(opts, gioApp.MaxSize(unit.Dp(state.MaxWidth), unit.Dp(state.MaxHeight)))
	}
	return opts
}

func clearWindowConstraints() gioApp.Option {
	return func(_ unit.Metric, config *gioApp.Config) {
		config.MinSize = image.Point{}
		config.MaxSize = image.Point{}
	}
}

func shouldSyncConfigConstraints(wasFullscreen bool, config gioApp.Config) bool {
	if config.Mode == gioApp.Fullscreen {
		return false
	}
	if wasFullscreen && config.MinSize == (image.Point{}) && config.MaxSize == (image.Point{}) {
		return false
	}
	return true
}

func updateDpSizeFromMetric(state *WindowState, metric unit.Metric, widthPx, heightPx int) {
	if widthPx > 0 {
		state.Width = pxToDpInt(metric, widthPx)
	}
	if heightPx > 0 {
		state.Height = pxToDpInt(metric, heightPx)
	}
}

func updateDpMinSizeFromMetric(state *WindowState, metric unit.Metric, widthPx, heightPx int) {
	state.MinWidth = pxToDpIntOrZero(metric, widthPx)
	state.MinHeight = pxToDpIntOrZero(metric, heightPx)
}

func updateDpMaxSizeFromMetric(state *WindowState, metric unit.Metric, widthPx, heightPx int) {
	state.MaxWidth = pxToDpIntOrZero(metric, widthPx)
	state.MaxHeight = pxToDpIntOrZero(metric, heightPx)
}

func pxToDpIntOrZero(metric unit.Metric, px int) int {
	if px <= 0 {
		return 0
	}
	return pxToDpInt(metric, px)
}

func pxToDpInt(metric unit.Metric, px int) int {
	return int(math.Round(float64(metric.PxToDp(px))))
}

func normalizeTitle(title string) string {
	if title == "" {
		return "FluxUI"
	}
	return title
}

func normalizeSize(width, height int) (int, int) {
	if width <= 0 {
		width = 420
	}
	if height <= 0 {
		height = 240
	}
	return width, height
}

func maxPositive(value int) int {
	if value < 0 {
		return 0
	}
	return value
}
