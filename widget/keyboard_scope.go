package widget

import (
	fluxevent "github.com/xiaowumin-mark/FluxUI/event"
	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"

	"gioui.org/io/key"
)

type KeyboardScopeOption func(*keyboardScopeConfig)

type keyboardScopeConfig struct {
	disabled  bool
	focusable bool
	tabIndex  int
	autoFocus bool
	listeners []keyboardScopeListener
	shortcuts []keyboardScopeShortcut
}

type keyboardScopeListener struct {
	eventType fluxevent.Type
	key       fluxevent.KeyboardHandler
	focus     fluxevent.FocusHandler
	options   []fluxevent.ListenerOption
}

type keyboardScopeShortcut struct {
	spec    fluxevent.Shortcut
	handler fluxevent.KeyboardHandler
	options []fluxevent.ListenerOption
}

type keyboardScopeWidget struct {
	child  Widget
	config keyboardScopeConfig
}

type keyboardScopeState struct {
	autoFocused bool
}

// KeyboardScope creates a component-tree keyboard event scope around child.
func KeyboardScope(child Widget, opts ...KeyboardScopeOption) Widget {
	cfg := keyboardScopeConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return &keyboardScopeWidget{child: child, config: cfg}
}

func KeyboardScopeDisabled(disabled bool) KeyboardScopeOption {
	return func(cfg *keyboardScopeConfig) {
		cfg.disabled = disabled
	}
}

func KeyboardScopeFocusable(focusable bool) KeyboardScopeOption {
	return func(cfg *keyboardScopeConfig) {
		cfg.focusable = focusable
	}
}

func KeyboardScopeTabIndex(index int) KeyboardScopeOption {
	return func(cfg *keyboardScopeConfig) {
		cfg.tabIndex = index
		cfg.focusable = true
	}
}

func KeyboardScopeAutoFocus(autoFocus bool) KeyboardScopeOption {
	return func(cfg *keyboardScopeConfig) {
		cfg.autoFocus = autoFocus
		cfg.focusable = true
	}
}

func KeyOn(eventType fluxevent.Type, fn fluxevent.KeyboardHandler, opts ...fluxevent.ListenerOption) KeyboardScopeOption {
	return func(cfg *keyboardScopeConfig) {
		cfg.listeners = append(cfg.listeners, keyboardScopeListener{
			eventType: eventType,
			key:       fn,
			options:   append([]fluxevent.ListenerOption(nil), opts...),
		})
	}
}

func KeyOnDown(fn fluxevent.KeyboardHandler, opts ...fluxevent.ListenerOption) KeyboardScopeOption {
	return KeyOn(fluxevent.KeyDown, fn, opts...)
}

func KeyOnUp(fn fluxevent.KeyboardHandler, opts ...fluxevent.ListenerOption) KeyboardScopeOption {
	return KeyOn(fluxevent.KeyUp, fn, opts...)
}

func FocusOn(eventType fluxevent.Type, fn fluxevent.FocusHandler, opts ...fluxevent.ListenerOption) KeyboardScopeOption {
	return func(cfg *keyboardScopeConfig) {
		cfg.listeners = append(cfg.listeners, keyboardScopeListener{
			eventType: eventType,
			focus:     fn,
			options:   append([]fluxevent.ListenerOption(nil), opts...),
		})
	}
}

func FocusOnFocus(fn fluxevent.FocusHandler, opts ...fluxevent.ListenerOption) KeyboardScopeOption {
	return FocusOn(fluxevent.Focus, fn, opts...)
}

func FocusOnBlur(fn fluxevent.FocusHandler, opts ...fluxevent.ListenerOption) KeyboardScopeOption {
	return FocusOn(fluxevent.Blur, fn, opts...)
}

func FocusOnIn(fn fluxevent.FocusHandler, opts ...fluxevent.ListenerOption) KeyboardScopeOption {
	return FocusOn(fluxevent.FocusIn, fn, opts...)
}

func FocusOnOut(fn fluxevent.FocusHandler, opts ...fluxevent.ListenerOption) KeyboardScopeOption {
	return FocusOn(fluxevent.FocusOut, fn, opts...)
}

func ShortcutOn(spec fluxevent.Shortcut, fn fluxevent.KeyboardHandler, opts ...fluxevent.ListenerOption) KeyboardScopeOption {
	return func(cfg *keyboardScopeConfig) {
		cfg.shortcuts = append(cfg.shortcuts, keyboardScopeShortcut{
			spec:    spec,
			handler: fn,
			options: append([]fluxevent.ListenerOption(nil), opts...),
		})
	}
}

func (w *keyboardScopeWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if w.child == nil {
		return layout.Dimensions{}
	}
	if w.config.disabled {
		return w.child.Layout(ctx.Child(0))
	}

	scopeCtx := ctx.Scope("keyboard-scope")
	registerKeyboardScopeListeners(scopeCtx, w.config)
	if w.config.focusable {
		fluxevent.RegisterFocusTarget(scopeCtx, fluxevent.FocusTabIndex(w.config.tabIndex))
	}
	if w.config.autoFocus && w.config.focusable {
		state := keyboardScopeStateFor(scopeCtx)
		if !state.autoFocused {
			fluxevent.RequestFocus(scopeCtx)
			state.autoFocused = true
		}
	}

	childDims := w.child.Layout(scopeCtx.Child(0))
	processKeyboardScopeEvents(scopeCtx)
	return childDims
}

func keyboardScopeStateFor(ctx *internal.Context) *keyboardScopeState {
	value := ctx.Memo("keyboard-scope", func() any {
		return &keyboardScopeState{}
	})
	state, ok := value.(*keyboardScopeState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUI/widget: keyboard scope state type mismatch")
	}
	return state
}

func registerKeyboardScopeListeners(ctx *internal.Context, cfg keyboardScopeConfig) {
	for _, listener := range cfg.listeners {
		switch {
		case listener.key != nil:
			fluxevent.OnKeyboard(ctx, listener.eventType, listener.key, listener.options...)
		case listener.focus != nil:
			fluxevent.OnFocus(ctx, listener.eventType, listener.focus, listener.options...)
		}
	}
	for _, shortcut := range cfg.shortcuts {
		fluxevent.OnShortcut(ctx, shortcut.spec, shortcut.handler, shortcut.options...)
	}
}

func processKeyboardScopeEvents(ctx *internal.Context) {
	if ctx == nil || ctx.Runtime() == nil {
		return
	}
	focused := ctx.Runtime().FocusedTarget()
	if focused == 0 || !ctx.Runtime().EventPathContains(focused, ctx.PathID()) {
		return
	}
	for {
		raw, ok := ctx.Gtx.Event(key.Filter{})
		if !ok {
			break
		}
		if keyEvent, ok := raw.(key.Event); ok {
			ev := fluxevent.KeyboardEventFromGio(ctx, keyEvent)
			fluxevent.DispatchKeyboardEvent(ctx, focused, &ev)
		}
	}
}
