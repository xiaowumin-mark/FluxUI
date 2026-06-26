package app

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"runtime"
	"runtime/debug"
	"sort"
	"sync"
	"sync/atomic"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
	fluxSystem "github.com/xiaowumin-mark/FluxUI/system"
	theme "github.com/xiaowumin-mark/FluxUI/theme"
	widget "github.com/xiaowumin-mark/FluxUI/widget"

	gioApp "gioui.org/app"
	"gioui.org/f32"
	gioInput "gioui.org/io/input"
	gioSystem "gioui.org/io/system"
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

// WindowsFrameMode controls how much native non-client frame Windows should draw.
type WindowsFrameMode int

const (
	WindowsFrameDefault WindowsFrameMode = iota
	WindowsFrameHidden
)

// WindowsCornerPreference controls the Windows 11 DWM corner preference.
type WindowsCornerPreference int

const (
	WindowsCornerDefault WindowsCornerPreference = iota
	WindowsCornerDoNotRound
	WindowsCornerRound
	WindowsCornerRoundSmall
)

// WindowsFrameBorderPolicy controls the Windows 11 DWM window border.
type WindowsFrameBorderPolicy int

const (
	WindowsFrameBorderDefault WindowsFrameBorderPolicy = iota
	WindowsFrameBorderHidden
	WindowsFrameBorderColor
)

// WindowsFrameStyle controls Windows-only native frame adornments.
type WindowsFrameStyle struct {
	Mode         WindowsFrameMode
	Shadow       bool
	Corner       WindowsCornerPreference
	Border       WindowsFrameBorderPolicy
	BorderColor  color.NRGBA
	CaptionColor color.NRGBA
	TextColor    color.NRGBA
}

// WindowsChromeAvailability reports which native Windows chrome features are
// available for the current process and OS build.
type WindowsChromeAvailability struct {
	Supported  bool
	FrameStyle bool
	DragMove   bool
}

// WindowState 是 FluxUI 运行时维护的窗口状态快照。
//
// Width/Height/MinWidth/MinHeight/MaxWidth/MaxHeight 使用 dp 单位。窗口实际被
// 平台调整后，FluxUI 会在收到 frame/config 事件时尽量同步该快照。
type WindowState struct {
	ID                 WindowID
	Title              string
	X                  int
	Y                  int
	Width              int
	Height             int
	Scale              float32
	TextScale          float32
	DPI                int
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
	Resizable          bool
	WindowsFrameStyle  WindowsFrameStyle
	Alive              bool
}

// WindowEventKind 是窗口生命周期或状态变化事件类型。
type WindowEventKind string

const (
	WindowEventSizeChanged    WindowEventKind = "size_changed"
	WindowEventScaleChanged   WindowEventKind = "scale_changed"
	WindowEventFocusChanged   WindowEventKind = "focus_changed"
	WindowEventStateChanged   WindowEventKind = "state_changed"
	WindowEventCloseRequested WindowEventKind = "close_requested"
	WindowEventClosed         WindowEventKind = "closed"
)

// WindowEvent 记录窗口状态变化。State 是事件发生后的窗口状态快照。
type WindowEvent struct {
	Window WindowHandle
	Kind   WindowEventKind
	State  WindowState
}

// WindowEventSubscription receives window events without polling.
type WindowEventSubscription struct {
	window WindowID
	id     uint64
	ch     chan WindowEvent
	closed atomic.Bool
}

// WindowCloseRequest is passed to close-request handlers.
type WindowCloseRequest struct {
	Window WindowHandle
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

// SubscribeEvents subscribes to future window events.
//
// Passing no kinds subscribes to every window event. Delivery is best-effort:
// if the receiver stops draining the channel, newer events may be dropped to
// avoid blocking the window event loop.
func (h WindowHandle) SubscribeEvents(kinds ...WindowEventKind) (*WindowEventSubscription, bool) {
	entry, ok := h.liveEntry()
	if !ok {
		return nil, false
	}
	return entry.addEventSubscriber(kinds...), true
}

// Close 请求关闭窗口。
func (h WindowHandle) Close() bool {
	return h.perform(gioSystem.ActionClose)
}

// SetCloseRequestedHandler controls whether close requests are allowed.
//
// The handler returns true to allow the close request and false to cancel it.
// Windows can intercept native WM_CLOSE requests; platforms without close-request
// interception keep programmatic close behavior but may not stop system close.
func (h WindowHandle) SetCloseRequestedHandler(fn func(WindowCloseRequest) bool) bool {
	return h.apply(func(entry *windowEntry) {
		entry.mu.Lock()
		entry.closeRequested = fn
		entry.mu.Unlock()
		entry.syncNativeCloseHook()
	})
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
	return h.perform(gioSystem.ActionRaise)
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
	return h.perform(gioSystem.ActionCenter)
}

// RequestFocus requests keyboard focus for the window.
func (h WindowHandle) RequestFocus() bool {
	entry, ok := h.liveEntry()
	if !ok {
		return false
	}
	if requestNativeWindowFocus(entry.nativeHandleSnapshot()) {
		return true
	}
	return h.perform(gioSystem.ActionRaise)
}

// SetTitle 更新窗口标题。
func (h WindowHandle) SetTitle(title string) bool {
	title = normalizeTitle(title)
	return h.apply(func(entry *windowEntry) {
		entry.win.Option(gioApp.Title(title))
		entry.updateAndEmit(func(state *WindowState) {
			state.Title = title
		})
	})
}

// SetPosition updates the window position in dp.
func (h WindowHandle) SetPosition(x, y int) bool {
	entry, ok := h.liveEntry()
	if !ok {
		return false
	}
	metric := entry.metricSnapshot()
	if !setNativeWindowPosition(entry.nativeHandleSnapshot(), metric.Dp(unit.Dp(x)), metric.Dp(unit.Dp(y))) {
		return false
	}
	entry.updateAndEmit(func(state *WindowState) {
		state.X = x
		state.Y = y
	})
	return true
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
		entry.updateAndEmit(func(state *WindowState) {
			state.Width = width
			state.Height = height
			applyWindowModeState(state, gioApp.Windowed)
		})
	})
}

// SetResizable controls whether the window can be resized by the OS frame.
func (h WindowHandle) SetResizable(resizable bool) bool {
	entry, ok := h.liveEntry()
	if !ok {
		return false
	}
	if !setNativeWindowResizable(entry.nativeHandleSnapshot(), resizable) {
		return false
	}
	entry.updateAndEmit(func(state *WindowState) {
		state.Resizable = resizable
	})
	entry.syncNativeMaximizeAvailability()
	return true
}

// SetDecorated controls whether the window uses the OS decoration frame.
func (h WindowHandle) SetDecorated(decorated bool) bool {
	return h.apply(func(entry *windowEntry) {
		entry.win.Option(gioApp.Decorated(decorated))
		entry.updateAndEmit(func(state *WindowState) {
			state.Decorated = decorated
		})
	})
}

// SetWindowsFrameStyle applies Windows-only native frame adornments.
func (h WindowHandle) SetWindowsFrameStyle(style WindowsFrameStyle) bool {
	entry, ok := h.liveEntry()
	if !ok {
		return false
	}
	style = normalizeWindowsFrameStyle(style)
	decorated := style.Mode != WindowsFrameHidden
	entry.win.Option(gioApp.Decorated(decorated))
	state := entry.snapshot()
	if !setNativeWindowFrameStyle(entry.nativeHandleSnapshot(), style, state.Resizable, windowMaximizeAvailable(state)) {
		return false
	}
	entry.updateAndEmit(func(state *WindowState) {
		state.WindowsFrameStyle = style
		state.Decorated = decorated
	})
	entry.syncNativeCloseHook()
	entry.syncNativeMaximizeAvailability()
	entry.syncNativeResizable()
	entry.win.Invalidate()
	return true
}

// StartDragMove starts a native window move operation from the current pointer press.
func (h WindowHandle) StartDragMove() bool {
	entry, ok := h.liveEntry()
	if !ok {
		return false
	}
	state := entry.snapshot()
	if state.Fullscreen || state.Minimized {
		return false
	}
	return startNativeWindowDragMove(entry.nativeHandleSnapshot())
}

// ProbeWindowsChrome returns Windows native chrome capability information.
func ProbeWindowsChrome() WindowsChromeAvailability {
	return probeNativeWindowsChrome()
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
		entry.updateAndEmit(func(state *WindowState) {
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
		entry.updateAndEmit(func(state *WindowState) {
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
	entry.updateAndEmit(func(state *WindowState) {
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

func (h WindowHandle) perform(actions gioSystem.Action) bool {
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

// Events returns the subscription event channel.
func (s *WindowEventSubscription) Events() <-chan WindowEvent {
	if s == nil {
		return nil
	}
	return s.ch
}

// Close closes the subscription. It returns false when the subscription is nil
// or was already closed.
func (s *WindowEventSubscription) Close() bool {
	if s == nil || !s.closed.CompareAndSwap(false, true) {
		return false
	}
	entry, ok := lookupWindow(s.window)
	if ok && entry != nil {
		entry.removeEventSubscriber(s.id)
		return true
	}
	return true
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
	Decorated          bool
	Resizable          bool
	WindowsFrameStyle  WindowsFrameStyle
	CloseRequested     func(WindowCloseRequest) bool
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
		Decorated:          true,
		Resizable:          true,
		WindowsFrameStyle: WindowsFrameStyle{
			Mode:   WindowsFrameDefault,
			Shadow: true,
			Corner: WindowsCornerDefault,
		},
		Theme: theme.Default(),
		Root:  root,
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

// Decorated sets whether the initial window uses the OS decoration frame.
func Decorated(enabled bool) Option {
	return func(app *Application) {
		app.Decorated = enabled
	}
}

// Resizable sets whether the initial window can be resized by the OS frame.
func Resizable(enabled bool) Option {
	return func(app *Application) {
		app.Resizable = enabled
	}
}

// WindowsFrame configures Windows-only native frame adornments.
func WindowsFrame(style WindowsFrameStyle) Option {
	return func(app *Application) {
		app.WindowsFrameStyle = normalizeWindowsFrameStyle(style)
		app.Decorated = app.WindowsFrameStyle.Mode != WindowsFrameHidden
	}
}

// OnCloseRequested sets the initial close-request handler.
func OnCloseRequested(fn func(WindowCloseRequest) bool) Option {
	return func(app *Application) {
		app.CloseRequested = fn
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

// WithIconFonts 追加全局图标字体集合。
func WithIconFonts(fonts ...theme.IconFont) Option {
	return func(app *Application) {
		if len(fonts) == 0 {
			return
		}
		if app.Theme == nil {
			app.Theme = theme.Default()
		}
		app.Theme.AddIconFonts(fonts...)
	}
}

// WithDefaultIconFont 设置应用级默认图标字体。
func WithDefaultIconFont(id string) Option {
	return func(app *Application) {
		if app.Theme == nil {
			app.Theme = theme.Default()
		}
		app.Theme.SetDefaultIconFont(id)
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
		gioApp.Decorated(a.Decorated),
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
			Decorated:          a.Decorated,
			Resizable:          a.Resizable,
			WindowsFrameStyle:  normalizeWindowsFrameStyle(a.WindowsFrameStyle),
			Alive:              true,
		},
		closeRequested: a.CloseRequested,
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
				entry.updateNativeActionRouter(nil)
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
			if rt.WindowDragAreaActive() {
				entry.updateNativeActionRouter(gtx.Ops)
			} else {
				entry.updateNativeActionRouter(nil)
			}

			if entry.renderSuspendedSnapshot() {
				entry.updateNativeActionRouter(nil)
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
		if err := fluxSystem.CloseTrays(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
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

func (c *windowController) RequestFocus() bool {
	return c.handle.RequestFocus()
}

func (c *windowController) SetTitle(title string) bool {
	return c.handle.SetTitle(title)
}

func (c *windowController) SetPosition(x, y int) bool {
	return c.handle.SetPosition(x, y)
}

func (c *windowController) SetSize(width, height int) bool {
	return c.handle.SetSize(width, height)
}

func (c *windowController) SetResizable(resizable bool) bool {
	return c.handle.SetResizable(resizable)
}

func (c *windowController) SetDecorated(decorated bool) bool {
	return c.handle.SetDecorated(decorated)
}

func (c *windowController) SetWindowsFrameStyle(style any) bool {
	typed, ok := style.(WindowsFrameStyle)
	return ok && c.handle.SetWindowsFrameStyle(typed)
}

func (c *windowController) StartDragMove() bool {
	return c.handle.StartDragMove()
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
	nativeResizableHandle   uintptr
	nativeResizableSynced   bool
	nativeResizableEnabled  bool
	nativeCloseHookHandle   uintptr
	nativeCloseHooked       bool
	nativeCloseOldProc      uintptr
	nativeActionRouter      *gioInput.Router
	state                   WindowState
	closeRequested          func(WindowCloseRequest) bool
	events                  []WindowEvent
	eventSubscribers        map[uint64]*windowEventSubscriber
	nextEventSubscriberID   uint64
}

type windowEventSubscriber struct {
	ch     chan WindowEvent
	filter map[WindowEventKind]bool
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
	entry := windowRegistry[id]
	delete(windowRegistry, id)
	windowRegistryMu.Unlock()
	if entry != nil {
		entry.closeEventSubscribers()
	}
	forgetNativeWindowCloseHook(entry)
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

func (entry *windowEntry) metricSnapshot() unit.Metric {
	entry.mu.RLock()
	metric := entry.metric
	entry.mu.RUnlock()
	return metric
}

func (entry *windowEntry) updateNativeActionRouter(ops *op.Ops) {
	if entry == nil {
		return
	}
	var router *gioInput.Router
	if ops != nil {
		router = new(gioInput.Router)
		router.Frame(ops)
	}
	entry.mu.Lock()
	entry.nativeActionRouter = router
	entry.mu.Unlock()
	entry.syncNativeCloseHook()
}

func (entry *windowEntry) nativeActionMoveAt(x, y int) bool {
	if entry == nil {
		return false
	}
	entry.mu.RLock()
	router := entry.nativeActionRouter
	state := entry.state
	entry.mu.RUnlock()
	if router == nil || state.Fullscreen || state.Minimized {
		return false
	}
	action, ok := router.ActionAt(f32.Pt(float32(x), float32(y)))
	return ok && action == gioSystem.ActionMove
}

func (entry *windowEntry) maximizeDisabledByConstraints() bool {
	return !windowMaximizeAvailable(entry.snapshot())
}

func (entry *windowEntry) syncNativeMaximizeAvailability() {
	entry.mu.Lock()
	handle := entry.nativeHandle
	if handle == 0 {
		entry.mu.Unlock()
		return
	}
	enabled := windowMaximizeAvailable(entry.state)
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

func (entry *windowEntry) syncNativeResizable() {
	entry.mu.Lock()
	handle := entry.nativeHandle
	if handle == 0 {
		entry.mu.Unlock()
		return
	}
	enabled := entry.state.Resizable
	if entry.nativeResizableSynced &&
		entry.nativeResizableHandle == handle &&
		entry.nativeResizableEnabled == enabled {
		entry.mu.Unlock()
		return
	}
	entry.nativeResizableHandle = handle
	entry.nativeResizableSynced = true
	entry.nativeResizableEnabled = enabled
	entry.mu.Unlock()

	setNativeWindowResizable(handle, enabled)
}

func (entry *windowEntry) syncNativeChrome() {
	if entry == nil {
		return
	}
	entry.mu.RLock()
	handle := entry.nativeHandle
	frame := normalizeWindowsFrameStyle(entry.state.WindowsFrameStyle)
	resizable := entry.state.Resizable
	maximizeEnabled := windowMaximizeAvailable(entry.state)
	entry.mu.RUnlock()
	if handle == 0 {
		return
	}
	setNativeWindowFrameStyle(handle, frame, resizable, maximizeEnabled)
}

func (entry *windowEntry) syncNativeCloseHook() {
	entry.mu.Lock()
	handle := entry.nativeHandle
	enabled := entry.closeRequested != nil || entry.state.WindowsFrameStyle.Mode == WindowsFrameHidden || entry.nativeActionRouter != nil
	hooked := entry.nativeCloseHooked
	if handle == 0 {
		entry.mu.Unlock()
		if hooked {
			uninstallNativeWindowCloseHook(entry)
		}
		return
	}
	if !enabled {
		entry.mu.Unlock()
		if hooked {
			uninstallNativeWindowCloseHook(entry)
		}
		return
	}
	if entry.nativeCloseHooked && entry.nativeCloseHookHandle == handle {
		entry.mu.Unlock()
		return
	}
	entry.mu.Unlock()

	installNativeWindowCloseHook(entry)
}

func (entry *windowEntry) handleCloseRequested() bool {
	entry.mu.Lock()
	handler := entry.closeRequested
	state := entry.state
	state.ID = entry.id
	state.Alive = entry.alive.Load()
	event := WindowEvent{
		Window: WindowHandle{id: entry.id},
		Kind:   WindowEventCloseRequested,
		State:  state,
	}
	entry.events = append(entry.events, event)
	entry.dispatchEventLocked(event)
	entry.mu.Unlock()

	if handler == nil {
		return true
	}
	return handler(WindowCloseRequest{
		Window: WindowHandle{id: entry.id},
		State:  state,
	})
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
		updateScaleFromMetric(state, metric)
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
		event := WindowEvent{
			Window: WindowHandle{id: entry.id},
			Kind:   kind,
			State:  after,
		}
		entry.events = append(entry.events, event)
		entry.dispatchEventLocked(event)
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
	event := WindowEvent{
		Window: WindowHandle{id: entry.id},
		Kind:   kind,
		State:  entry.state,
	}
	entry.events = append(entry.events, event)
	entry.dispatchEventLocked(event)
	entry.mu.Unlock()
}

func (entry *windowEntry) drainEvents() []WindowEvent {
	entry.mu.Lock()
	events := append([]WindowEvent(nil), entry.events...)
	entry.events = nil
	entry.mu.Unlock()
	return events
}

func (entry *windowEntry) addEventSubscriber(kinds ...WindowEventKind) *WindowEventSubscription {
	entry.mu.Lock()
	entry.nextEventSubscriberID++
	id := entry.nextEventSubscriberID
	if entry.eventSubscribers == nil {
		entry.eventSubscribers = make(map[uint64]*windowEventSubscriber)
	}
	ch := make(chan WindowEvent, 16)
	entry.eventSubscribers[id] = &windowEventSubscriber{
		ch:     ch,
		filter: windowEventFilter(kinds),
	}
	entry.mu.Unlock()

	return &WindowEventSubscription{
		window: entry.id,
		id:     id,
		ch:     ch,
	}
}

func (entry *windowEntry) removeEventSubscriber(id uint64) {
	entry.mu.Lock()
	sub := entry.eventSubscribers[id]
	delete(entry.eventSubscribers, id)
	entry.mu.Unlock()
	if sub != nil {
		close(sub.ch)
	}
}

func (entry *windowEntry) closeEventSubscribers() {
	entry.mu.Lock()
	subscribers := entry.eventSubscribers
	entry.eventSubscribers = nil
	entry.mu.Unlock()

	for _, sub := range subscribers {
		close(sub.ch)
	}
}

func (entry *windowEntry) dispatchEventLocked(event WindowEvent) {
	for _, sub := range entry.eventSubscribers {
		if !sub.accepts(event.Kind) {
			continue
		}
		select {
		case sub.ch <- event:
		default:
		}
	}
}

func (sub *windowEventSubscriber) accepts(kind WindowEventKind) bool {
	return sub != nil && (len(sub.filter) == 0 || sub.filter[kind])
}

func windowEventFilter(kinds []WindowEventKind) map[WindowEventKind]bool {
	if len(kinds) == 0 {
		return nil
	}
	filter := make(map[WindowEventKind]bool, len(kinds))
	for _, kind := range kinds {
		if kind != "" {
			filter[kind] = true
		}
	}
	return filter
}

func windowEventKinds(before, after WindowState) []WindowEventKind {
	kinds := make([]WindowEventKind, 0, 3)
	if before.Width != after.Width || before.Height != after.Height {
		kinds = append(kinds, WindowEventSizeChanged)
	}
	if before.Scale != after.Scale || before.TextScale != after.TextScale || before.DPI != after.DPI {
		kinds = append(kinds, WindowEventScaleChanged)
	}
	if before.Focused != after.Focused {
		kinds = append(kinds, WindowEventFocusChanged)
	}
	if before.Minimized != after.Minimized ||
		before.Maximized != after.Maximized ||
		before.Fullscreen != after.Fullscreen ||
		before.Decorated != after.Decorated ||
		before.Resizable != after.Resizable ||
		before.WindowsFrameStyle != after.WindowsFrameStyle ||
		before.Visible != after.Visible ||
		before.AlwaysOnTop != after.AlwaysOnTop ||
		before.RenderSuspended != after.RenderSuspended ||
		before.HiddenMemoryPolicy != after.HiddenMemoryPolicy ||
		before.Title != after.Title ||
		before.X != after.X ||
		before.Y != after.Y ||
		before.Scale != after.Scale ||
		before.TextScale != after.TextScale ||
		before.DPI != after.DPI ||
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

func windowMaximizeAvailable(state WindowState) bool {
	return state.Resizable && !hasWindowMaxSize(state)
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

func validWindowsFrameMode(mode WindowsFrameMode) bool {
	switch mode {
	case WindowsFrameDefault, WindowsFrameHidden:
		return true
	default:
		return false
	}
}

func validWindowsCornerPreference(corner WindowsCornerPreference) bool {
	switch corner {
	case WindowsCornerDefault, WindowsCornerDoNotRound, WindowsCornerRound, WindowsCornerRoundSmall:
		return true
	default:
		return false
	}
}

func validWindowsFrameBorderPolicy(policy WindowsFrameBorderPolicy) bool {
	switch policy {
	case WindowsFrameBorderDefault, WindowsFrameBorderHidden, WindowsFrameBorderColor:
		return true
	default:
		return false
	}
}

func normalizeWindowsFrameStyle(style WindowsFrameStyle) WindowsFrameStyle {
	if !validWindowsFrameMode(style.Mode) {
		style.Mode = WindowsFrameDefault
	}
	if !validWindowsCornerPreference(style.Corner) {
		style.Corner = WindowsCornerDefault
	}
	if !validWindowsFrameBorderPolicy(style.Border) {
		style.Border = WindowsFrameBorderDefault
	}
	return style
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

func updateScaleFromMetric(state *WindowState, metric unit.Metric) {
	if metric.PxPerDp > 0 {
		state.Scale = metric.PxPerDp
		state.DPI = int(math.Round(float64(metric.PxPerDp * 96)))
	}
	if metric.PxPerSp > 0 {
		state.TextScale = metric.PxPerSp
	}
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
