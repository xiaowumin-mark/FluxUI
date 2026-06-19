//go:build windows

package app

import (
	"image"
	"testing"

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
