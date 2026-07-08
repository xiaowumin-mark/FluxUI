//go:build windows

package app

import (
	"image"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"

	"gioui.org/io/system"
	"gioui.org/op"
	"gioui.org/op/clip"
)

func TestNativeWindowProcCloseRequestCanCancelWMClose(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:  "Unsaved",
		Width:  320,
		Height: 240,
		Alive:  true,
	})
	handle := WindowHandle{id: entry.id}
	calls := 0
	if !handle.SetCloseRequestedHandler(func(request WindowCloseRequest) bool {
		calls++
		if request.Window.ID() != entry.id {
			t.Fatalf("unexpected request window: %#v", request.Window)
		}
		return false
	}) {
		t.Fatal("expected close requested handler to be set")
	}

	const fakeHWND = uintptr(0x46575549)
	nativeCloseHookMu.Lock()
	nativeCloseHookEntries[fakeHWND] = entry
	nativeCloseHookMu.Unlock()
	defer forgetNativeWindowCloseHookHandle(fakeHWND, entry)

	if result := nativeWindowProc(fakeHWND, nativeWMClose, 0, 0); result != 0 {
		t.Fatalf("expected canceled WM_CLOSE to return 0, got %d", result)
	}
	if calls != 1 {
		t.Fatalf("expected one close request callback, got %d", calls)
	}
	if !hasWindowEvent(handle.PollEvents(), WindowEventCloseRequested) {
		t.Fatal("expected WM_CLOSE to emit close requested event")
	}
}

func TestNativeWindowProcHiddenFrameSuppressesNonClientPaint(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title: "Custom chrome",
		WindowsFrameStyle: WindowsFrameStyle{
			Mode:   WindowsFrameHidden,
			Shadow: true,
			Corner: WindowsCornerRound,
		},
		Alive: true,
	})

	const fakeHWND = uintptr(0x46575550)
	nativeCloseHookMu.Lock()
	nativeCloseHookEntries[fakeHWND] = entry
	nativeCloseHookMu.Unlock()
	defer forgetNativeWindowCloseHookHandle(fakeHWND, entry)

	if result := nativeWindowProc(fakeHWND, nativeWMNCPaint, 0, 0); result != 0 {
		t.Fatalf("expected hidden frame WM_NCPAINT to return 0, got %d", result)
	}
	if result := nativeWindowProc(fakeHWND, nativeWMNCActivate, 0, 0); result != 1 {
		t.Fatalf("expected hidden frame WM_NCACTIVATE to return 1, got %d", result)
	}
	if nativeSuppressHiddenFrameNonClient(nativeWMNCCalcSize) {
		t.Fatal("hidden frame should leave WM_NCCALCSIZE to Gio's borderless window logic")
	}
}

func TestNativeHiddenFrameDoesNotSuppressCaptionInput(t *testing.T) {
	if nativeSuppressHiddenFrameNonClient(nativeWMNCLButtonDown) {
		t.Fatal("hidden frame should not suppress native caption button down")
	}
	if nativeSuppressHiddenFrameNonClient(nativeWMNCLDblClk) {
		t.Fatal("hidden frame should not suppress native caption double click")
	}
}

func TestNativeActionRouterDetectsWindowMoveRegion(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:     "Drag region",
		Resizable: true,
		Alive:     true,
	})

	var ops op.Ops
	stack := clip.Rect(image.Rect(0, 0, 160, 40)).Push(&ops)
	system.ActionInputOp(system.ActionMove).Add(&ops)
	stack.Pop()

	entry.updateNativeActionRouter(&ops)
	if !entry.nativeActionMoveAt(80, 20) {
		t.Fatal("expected native action router to detect ActionMove inside the drag region")
	}
	if entry.nativeActionMoveAt(220, 20) {
		t.Fatal("did not expect ActionMove outside the drag region")
	}

	entry.updateState(func(state *WindowState) {
		state.Fullscreen = true
	})
	if entry.nativeActionMoveAt(80, 20) {
		t.Fatal("fullscreen windows should not expose custom drag regions as caption")
	}

	entry.updateNativeActionRouter(nil)
	if entry.nativeActionMoveAt(80, 20) {
		t.Fatal("cleared native action router should not report drag regions")
	}
}

func TestNativeWindowActionRegionDetectsMaximizeButton(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:     "Maximize region",
		Resizable: true,
		Alive:     true,
	})
	entry.updateNativeWindowActionRegions([]internal.NativeWindowActionRegion{
		{
			Action: internal.NativeWindowActionMaximizeButton,
			Rect:   image.Rect(160, 0, 206, 32),
		},
	})

	if !entry.nativeMaximizeButtonAt(180, 16) {
		t.Fatal("expected native action region to detect maximize button")
	}
	if entry.nativeMaximizeButtonAt(140, 16) {
		t.Fatal("did not expect maximize button outside native action region")
	}

	entry.updateState(func(state *WindowState) {
		state.Fullscreen = true
	})
	if entry.nativeMaximizeButtonAt(180, 16) {
		t.Fatal("fullscreen windows should not expose maximize button regions")
	}

	entry.updateState(func(state *WindowState) {
		state.Fullscreen = false
		state.Resizable = false
	})
	if entry.nativeMaximizeButtonAt(180, 16) {
		t.Fatal("non-resizable windows should not expose maximize button regions")
	}

	entry.updateState(func(state *WindowState) {
		state.Resizable = true
		state.MaxWidth = 640
		state.MaxHeight = 480
	})
	if entry.nativeMaximizeButtonAt(180, 16) {
		t.Fatal("max-size constrained windows should not expose maximize button regions")
	}

	entry.updateState(func(state *WindowState) {
		state.MaxWidth = 0
		state.MaxHeight = 0
	})
	entry.updateNativeWindowActionRegions(nil)
	if entry.nativeMaximizeButtonAt(180, 16) {
		t.Fatal("cleared native action regions should not report maximize buttons")
	}
}

func TestNativeActionRouterDetectsWindowMaximizeButton(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:     "Maximize action",
		Resizable: true,
		Alive:     true,
	})

	var ops op.Ops
	stack := clip.Rect(image.Rect(160, 0, 206, 32)).Push(&ops)
	system.ActionInputOp(system.ActionMaximize).Add(&ops)
	stack.Pop()

	entry.updateNativeActionRouter(&ops)
	if !entry.nativeMaximizeButtonAt(180, 16) {
		t.Fatal("expected native action router to detect ActionMaximize inside the maximize button")
	}
	if entry.nativeMaximizeButtonAt(140, 16) {
		t.Fatal("did not expect ActionMaximize outside the maximize button")
	}

	entry.updateState(func(state *WindowState) {
		state.Resizable = false
	})
	if entry.nativeMaximizeButtonAt(180, 16) {
		t.Fatal("non-resizable windows should not expose ActionMaximize as maximize button")
	}

	entry.updateNativeActionRouter(nil)
}

func TestNativeClientInputMessageMapsNonClientPointerForGio(t *testing.T) {
	tests := []struct {
		name         string
		msg          uintptr
		wparam       uintptr
		wantMsg      uintptr
		wantWParam   uintptr
		clientCoords bool
	}{
		{
			name:       "pointer update",
			msg:        nativeWMNCPointerUpdate,
			wparam:     42,
			wantMsg:    nativeWMPointerUpdate,
			wantWParam: 42,
		},
		{
			name:       "pointer down",
			msg:        nativeWMNCPointerDown,
			wparam:     43,
			wantMsg:    nativeWMPointerDown,
			wantWParam: 43,
		},
		{
			name:       "pointer up",
			msg:        nativeWMNCPointerUp,
			wparam:     44,
			wantMsg:    nativeWMPointerUp,
			wantWParam: 44,
		},
		{
			name:         "legacy mouse move",
			msg:          nativeWMNCMouseMove,
			wparam:       nativeHTMaxButton,
			wantMsg:      nativeWMMouseMove,
			clientCoords: true,
		},
		{
			name:         "legacy mouse down",
			msg:          nativeWMNCLButtonDown,
			wparam:       nativeHTMaxButton,
			wantMsg:      nativeWMLButtonDown,
			wantWParam:   nativeMKLButton,
			clientCoords: true,
		},
		{
			name:         "legacy mouse up",
			msg:          nativeWMNCLButtonUp,
			wparam:       nativeHTMaxButton,
			wantMsg:      nativeWMLButtonUp,
			clientCoords: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotMsg, gotWParam, gotClientCoords, ok := nativeClientInputMessage(test.msg, test.wparam)
			if !ok {
				t.Fatal("expected non-client input message to map to Gio input")
			}
			if gotMsg != test.wantMsg || gotWParam != test.wantWParam || gotClientCoords != test.clientCoords {
				t.Fatalf("mapped message = (%#x, %#x, %v), want (%#x, %#x, %v)",
					gotMsg, gotWParam, gotClientCoords, test.wantMsg, test.wantWParam, test.clientCoords)
			}
		})
	}
}

func TestNativeMaximizeInputStateSuppressesLegacyFallback(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:     "Maximize input",
		Resizable: true,
		Alive:     true,
	})

	if nativeMaximizePointerDown(entry) || nativeMaximizeMouseDown(entry) {
		t.Fatal("new windows should not start with maximize input pressed")
	}

	nativeSetMaximizeMouseDown(entry, true)
	if !nativeMaximizeMouseDown(entry) {
		t.Fatal("expected legacy maximize mouse fallback state to be set")
	}

	nativeSetMaximizePointerDown(entry, true)
	if !nativeMaximizePointerDown(entry) {
		t.Fatal("expected forwarded pointer state to be set")
	}

	activateFallback := nativeMaximizeMouseDown(entry) && !nativeMaximizePointerDown(entry)
	if activateFallback {
		t.Fatal("legacy fallback should be suppressed while pointer input owns the click")
	}

	nativeSetMaximizePointerDown(entry, false)
	nativeSetMaximizeMouseDown(entry, false)
	if nativeMaximizePointerDown(entry) || nativeMaximizeMouseDown(entry) {
		t.Fatal("maximize input states should clear")
	}
}

func TestNativeMaximizeButtonActivationTogglesWindowMode(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:     "Maximize activate",
		Resizable: true,
		Alive:     true,
	})

	if !entry.activateNativeMaximizeButton() {
		t.Fatal("expected native maximize activation to maximize")
	}
	if state := entry.snapshot(); !state.Maximized {
		t.Fatalf("expected maximized state after native activation: %#v", state)
	}

	if !entry.activateNativeMaximizeButton() {
		t.Fatal("expected native maximize activation to restore")
	}
	if state := entry.snapshot(); state.Maximized {
		t.Fatalf("expected restored state after second native activation: %#v", state)
	}

	entry.updateState(func(state *WindowState) {
		state.Resizable = false
	})
	if entry.activateNativeMaximizeButton() {
		t.Fatal("non-resizable windows should not activate native maximize")
	}
}

func TestNativeWindowChromeStyleDefaultKeepsSystemSemantics(t *testing.T) {
	current := uintptr(nativeWSVisible | nativeWSDisabled | nativeWSMinimize | nativeWSMaximize)
	style := nativeWindowChromeStyle(current, WindowsFrameDefault, true, true)
	for _, bit := range []uintptr{
		nativeWSVisible,
		nativeWSDisabled,
		nativeWSMinimize,
		nativeWSMaximize,
		nativeWSCaption,
		nativeWSSysMenu,
		nativeWSSizeBox,
		nativeWSMinimizeBox,
		nativeWSMaximizeBox,
	} {
		if style&bit != bit {
			t.Fatalf("expected style %#x to include bit %#x", style, bit)
		}
	}

	style = nativeWindowChromeStyle(current, WindowsFrameDefault, false, false)
	if style&nativeWSSizeBox != 0 {
		t.Fatalf("non-resizable style should not include WS_SIZEBOX: %#x", style)
	}
	if style&nativeWSMaximizeBox != 0 {
		t.Fatalf("non-maximizable style should not include WS_MAXIMIZEBOX: %#x", style)
	}
	if style&nativeWSCaption != nativeWSCaption || style&nativeWSSysMenu != nativeWSSysMenu {
		t.Fatalf("default chrome style should preserve caption/system menu semantics: %#x", style)
	}
}

func TestNativeWindowChromeStyleHiddenRemovesCaption(t *testing.T) {
	current := uintptr(nativeWSVisible | nativeWSDisabled | nativeWSMinimize | nativeWSMaximize | nativeWSCaption)
	style := nativeWindowChromeStyle(current, WindowsFrameHidden, true, true)
	for _, bit := range []uintptr{
		nativeWSVisible,
		nativeWSDisabled,
		nativeWSMinimize,
		nativeWSMaximize,
		nativeWSSysMenu,
		nativeWSSizeBox,
		nativeWSMinimizeBox,
		nativeWSMaximizeBox,
	} {
		if style&bit != bit {
			t.Fatalf("expected hidden style %#x to include bit %#x", style, bit)
		}
	}
	if style&nativeWSCaption != 0 {
		t.Fatalf("hidden chrome style should remove WS_CAPTION to avoid native frame paint: %#x", style)
	}
}

func TestNativeFrameMargins(t *testing.T) {
	margins := nativeFrameMargins(WindowsFrameStyle{Mode: WindowsFrameHidden, Shadow: true})
	if margins.Left != -1 || margins.Right != -1 || margins.Top != -1 || margins.Bottom != -1 {
		t.Fatalf("hidden shadow margins = %#v, want all -1", margins)
	}

	margins = nativeFrameMargins(WindowsFrameStyle{Mode: WindowsFrameHidden, Shadow: false})
	if margins != (nativeMargins{}) {
		t.Fatalf("hidden no-shadow margins = %#v, want zero", margins)
	}

	margins = nativeFrameMargins(WindowsFrameStyle{Mode: WindowsFrameDefault, Shadow: true})
	if margins != (nativeMargins{}) {
		t.Fatalf("default margins = %#v, want zero", margins)
	}
}
