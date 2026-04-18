package app

import (
	"errors"
	"fmt"
	"os"
	"runtime"
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

// WindowHandle 表示运行中的窗口句柄。
type WindowHandle struct {
	id WindowID
}

// ID 返回窗口 ID。
func (h WindowHandle) ID() WindowID {
	return h.id
}

// IsAlive 返回窗口是否仍在运行。
func (h WindowHandle) IsAlive() bool {
	entry, ok := lookupWindow(h.id)
	return ok && entry != nil && entry.alive.Load()
}

// Close 请求关闭窗口。
func (h WindowHandle) Close() bool {
	return h.perform(system.ActionClose)
}

// Minimize 最小化窗口。
func (h WindowHandle) Minimize() bool {
	return h.applyOption(gioApp.Minimized.Option())
}

// Maximize 最大化窗口。
func (h WindowHandle) Maximize() bool {
	return h.applyOption(gioApp.Maximized.Option())
}

// Restore 还原窗口为普通模式。
func (h WindowHandle) Restore() bool {
	return h.applyOption(gioApp.Windowed.Option())
}

// Fullscreen 切换窗口为全屏模式。
func (h WindowHandle) Fullscreen() bool {
	return h.applyOption(gioApp.Fullscreen.Option())
}

// Raise 请求将窗口置于最前。
func (h WindowHandle) Raise() bool {
	return h.perform(system.ActionRaise)
}

// Center 请求将窗口居中。
func (h WindowHandle) Center() bool {
	return h.perform(system.ActionCenter)
}

// SetTitle 更新窗口标题。
func (h WindowHandle) SetTitle(title string) bool {
	if title == "" {
		title = "FluxUI"
	}
	return h.applyOption(gioApp.Title(title))
}

// SetSize 更新窗口尺寸（单位为 dp）。
func (h WindowHandle) SetSize(width, height int) bool {
	if width <= 0 || height <= 0 {
		return false
	}
	return h.applyOption(gioApp.Size(unit.Dp(width), unit.Dp(height)))
}

// Invalidate 请求窗口重绘。
func (h WindowHandle) Invalidate() bool {
	return h.apply(func(entry *windowEntry) {
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

func (h WindowHandle) perform(actions system.Action) bool {
	if actions == 0 {
		return false
	}
	return h.apply(func(entry *windowEntry) {
		entry.win.Perform(actions)
	})
}

func (h WindowHandle) apply(fn func(entry *windowEntry)) bool {
	if h.id == 0 || fn == nil {
		return false
	}
	entry, ok := lookupWindow(h.id)
	if !ok || entry == nil || !entry.alive.Load() {
		return false
	}
	fn(entry)
	return true
}

// Application 是 Gio window loop 的封装。
type Application struct {
	Title  string
	Width  int
	Height int
	Theme  *theme.Theme
	Root   Builder
}

// WindowSpec 是多窗口运行时的窗口配置。
type WindowSpec struct {
	Root    Builder
	Options []Option
}

// New 创建应用实例。
func New(root Builder, opts ...Option) *Application {
	app := &Application{
		Title:  "FluxUI",
		Width:  420,
		Height: 240,
		Theme:  theme.Default(),
		Root:   root,
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

// WithTheme 覆盖应用主题。
func WithTheme(th *theme.Theme) Option {
	return func(app *Application) {
		if th != nil {
			app.Theme = th
		}
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

	width := a.Width
	height := a.Height
	title := a.Title
	if width <= 0 {
		width = 420
	}
	if height <= 0 {
		height = 240
	}
	if title == "" {
		title = "FluxUI"
	}
	th := a.Theme
	if th == nil {
		th = theme.Default()
	}

	w := new(gioApp.Window)
	w.Option(
		gioApp.Title(title),
		gioApp.Size(unit.Dp(width), unit.Dp(height)),
	)

	windowID := nextWindowID()
	entry := &windowEntry{id: windowID, win: w}
	entry.alive.Store(true)
	registerWindow(entry)
	defer func() {
		entry.alive.Store(false)
		unregisterWindow(windowID)
	}()

	rt := internal.NewRuntime(th)
	rt.SetInvalidator(w.Invalidate)
	rt.SetWindowController(&windowController{
		handle: WindowHandle{id: windowID},
	})
	defer rt.Dispose()

	var ops op.Ops
	for {
		switch evt := w.Event().(type) {
		case gioApp.DestroyEvent:
			entry.alive.Store(false)
			return evt.Err
		case gioApp.FrameEvent:
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

func (c *windowController) Center() bool {
	return c.handle.Center()
}

func (c *windowController) SetTitle(title string) bool {
	return c.handle.SetTitle(title)
}

func (c *windowController) SetSize(width, height int) bool {
	return c.handle.SetSize(width, height)
}

func (c *windowController) Invalidate() bool {
	return c.handle.Invalidate()
}

func (c *windowController) IsAlive() bool {
	return c.handle.IsAlive()
}

type windowEntry struct {
	id    WindowID
	win   *gioApp.Window
	alive atomic.Bool
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
