package app

import (
	"fluxui/internal"
	"fluxui/theme"
	"fluxui/widget"
	"fmt"
	"os"
	"runtime"

	gioApp "gioui.org/app"
	"gioui.org/op"
	"gioui.org/unit"
)

// Builder 定义应用根组件构造函数。
type Builder func(ctx *internal.Context) widget.Widget

// Option 定义应用启动配置。
type Option func(*Application)

// Application 是 Gio window loop 的封装。
type Application struct {
	Title  string
	Width  int
	Height int
	Theme  *theme.Theme
	Root   Builder
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

// Run 启动窗口事件循环。
func (a *Application) Run() error {
	w := new(gioApp.Window)
	w.Option(
		gioApp.Title(a.Title),
		gioApp.Size(unit.Dp(a.Width), unit.Dp(a.Height)),
	)

	runtime := internal.NewRuntime(a.Theme)
	runtime.SetInvalidator(w.Invalidate)

	var ops op.Ops
	for {
		switch evt := w.Event().(type) {
		case gioApp.DestroyEvent:
			return evt.Err
		case gioApp.FrameEvent:
			gtx := gioApp.NewContext(&ops, evt)
			ctx := runtime.Frame(gtx)
			buildCtx := ctx.Scope("build")
			treeCtx := ctx.Scope("tree")

			if a.Root != nil {
				if root := a.Root(buildCtx); root != nil {
					root.Layout(treeCtx.Child(0))
				}
			}

			evt.Frame(gtx.Ops)
		}
	}
}

// Run 直接创建并启动应用。
func Run(root Builder, opts ...Option) error {
	done := make(chan error, 1)
	go func() {
		err := New(root, opts...).Run()
		done <- err

		if runtime.GOOS == "android" || runtime.GOOS == "ios" {
			return
		}
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	gioApp.Main()
	return <-done
}
