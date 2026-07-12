//go:build windows

package app

import (
	"image"
	"testing"
	"time"

	"github.com/xiaowumin-mark/FluxUI/internal"

	"gioui.org/io/system"
	"gioui.org/op"
	"gioui.org/op/clip"
)

func TestNativeWindowTargetRejectsReusedHandleGeneration(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{Alive: true})
	const reusedHWND = uintptr(0x46575702)
	stale := entry.setNativeWindowHandle(reusedHWND)
	entry.invalidateNativeWindowHandle()
	current := entry.setNativeWindowHandle(reusedHWND)
	if stale.generation == current.generation {
		t.Fatal("reused HWND should have a new lifecycle generation")
	}

	if stale.valid() {
		t.Fatal("stale HWND generation should be invalid")
	}
	if !current.valid() {
		t.Fatal("current HWND generation should remain valid")
	}
}

func TestNativeWindowTargetIsNotPublishedUntilReady(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{Alive: true})
	target := entry.stageNativeWindowHandle(0x46575705)
	if snapshot := entry.nativeWindowTargetSnapshot(); snapshot.handle != 0 {
		t.Fatalf("unready HWND was published: %#x", snapshot.handle)
	}
	if handle, ok := (WindowHandle{id: entry.id}).NativeHandle(); ok || handle != 0 {
		t.Fatalf("NativeHandle before ready = (%#x, %v)", handle, ok)
	}
	if !entry.markNativeWindowTargetReady(target) {
		t.Fatal("failed to mark current HWND generation ready")
	}
	if snapshot := entry.nativeWindowTargetSnapshot(); snapshot != target {
		t.Fatalf("ready target = %#v, want %#v", snapshot, target)
	}
}

func TestNativeWindowOperationRejectsCallbackReentry(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{Alive: true})
	entry.nativeCallbackDepth.Store(1)
	ran := false
	if entry.withNativeWindowOperation(func() bool {
		ran = true
		return true
	}) {
		t.Fatal("native mutation should fail during a native callback")
	}
	if ran {
		t.Fatal("native mutation ran during a native callback")
	}
	if (WindowHandle{id: entry.id}).Maximize() {
		t.Fatal("Gio window mutation should fail during a native callback")
	}
}

func TestNativeWindowOperationSerializesStateCommit(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{Alive: true})

	firstStarted := make(chan struct{})
	releaseFirstCommit := make(chan struct{})
	firstDone := make(chan bool, 1)
	go func() {
		firstDone <- entry.withNativeWindowOperation(func() bool {
			close(firstStarted)
			<-releaseFirstCommit
			entry.updateState(func(state *WindowState) { state.X = 1 })
			return true
		})
	}()
	<-firstStarted

	secondCalling := make(chan struct{})
	secondStarted := make(chan struct{})
	secondDone := make(chan bool, 1)
	go func() {
		close(secondCalling)
		secondDone <- entry.withNativeWindowOperation(func() bool {
			close(secondStarted)
			if entry.snapshot().X != 1 {
				return false
			}
			entry.updateState(func(state *WindowState) { state.X = 2 })
			return true
		})
	}()
	<-secondCalling

	select {
	case <-secondStarted:
		t.Fatal("next operation ran before the prior state commit")
	case <-time.After(25 * time.Millisecond):
	}
	close(releaseFirstCommit)
	if !<-firstDone || !<-secondDone {
		t.Fatal("expected both serialized native operations to succeed")
	}
	if state := entry.snapshot(); state.X != 2 {
		t.Fatalf("serialized state = %d, want 2", state.X)
	}
}

func TestNativeCloseHookBindingRejectsReusedHWND(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{Alive: true})
	const reusedHWND = uintptr(0x46575703)
	stale := entry.setNativeWindowHandle(reusedHWND)
	nativeCloseHookMu.Lock()
	nativeCloseHookEntries[reusedHWND] = nativeCloseHookBinding{target: stale, oldProc: 0x1234}
	nativeCloseHookMu.Unlock()
	defer forgetNativeWindowCloseHookTarget(stale)

	entry.invalidateNativeWindowHandle()
	current := entry.setNativeWindowHandle(reusedHWND)
	if target, ok := nativeCloseHookTarget(reusedHWND); ok || target.entry != nil {
		t.Fatal("stale close-hook binding must not resolve after HWND reuse")
	}
	if binding, ok := nativeCloseHookBindingForHandle(reusedHWND); !ok || binding.oldProc != 0x1234 {
		t.Fatal("installed hook must retain its old WNDPROC after target invalidation")
	}
	if !current.valid() {
		t.Fatal("current reused HWND target should remain valid")
	}
}

func TestForgetNativeCloseHookTargetDoesNotClearNewGeneration(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{Alive: true})
	oldTarget := entry.setNativeWindowHandle(0x46575706)
	newTarget := entry.setNativeWindowHandle(0x46575707)
	entry.mu.Lock()
	entry.nativeCloseHookHandle = newTarget.handle
	entry.nativeCloseHookGeneration = newTarget.generation
	entry.nativeCloseHooked = true
	entry.nativeCloseOldProc = 0x5678
	entry.mu.Unlock()

	nativeCloseHookMu.Lock()
	nativeCloseHookEntries[oldTarget.handle] = nativeCloseHookBinding{target: oldTarget, oldProc: 0x1234}
	nativeCloseHookEntries[newTarget.handle] = nativeCloseHookBinding{target: newTarget, oldProc: 0x5678}
	nativeCloseHookMu.Unlock()
	defer forgetNativeWindowCloseHookTarget(newTarget)

	forgetNativeWindowCloseHookTarget(oldTarget)
	entry.mu.RLock()
	hooked := entry.nativeCloseHooked &&
		entry.nativeCloseHookHandle == newTarget.handle &&
		entry.nativeCloseHookGeneration == newTarget.generation &&
		entry.nativeCloseOldProc == 0x5678
	entry.mu.RUnlock()
	if !hooked {
		t.Fatal("forgetting an old HWND generation cleared the current hook")
	}
	if binding, ok := nativeCloseHookBindingForHandle(newTarget.handle); !ok || binding.target != newTarget {
		t.Fatal("forgetting an old HWND generation cleared the current binding")
	}
}

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
		if request.Window.Maximize() {
			t.Fatal("window mutation must fail instead of deadlocking inside WM_CLOSE")
		}
		return false
	}) {
		t.Fatal("expected close requested handler to be set")
	}

	const fakeHWND = uintptr(0x46575549)
	target := entry.setNativeWindowHandle(fakeHWND)
	nativeCloseHookMu.Lock()
	nativeCloseHookEntries[fakeHWND] = nativeCloseHookBinding{target: target}
	nativeCloseHookMu.Unlock()
	defer forgetNativeWindowCloseHookTarget(target)

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
	target := entry.setNativeWindowHandle(fakeHWND)
	nativeCloseHookMu.Lock()
	nativeCloseHookEntries[fakeHWND] = nativeCloseHookBinding{target: target}
	nativeCloseHookMu.Unlock()
	defer forgetNativeWindowCloseHookTarget(target)

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

func TestNativeMaximizeButtonSystemCommandDefersWithoutCallbackReentry(t *testing.T) {
	tests := []struct {
		name  string
		state WindowState
		want  uintptr
		ok    bool
	}{
		{
			name:  "maximize",
			state: WindowState{Resizable: true},
			want:  nativeSCMaximize,
			ok:    true,
		},
		{
			name:  "restore",
			state: WindowState{Resizable: true, Maximized: true},
			want:  nativeSCRestore,
			ok:    true,
		},
		{
			name:  "non-resizable",
			state: WindowState{Resizable: false},
			ok:    false,
		},
		{
			name:  "fullscreen",
			state: WindowState{Resizable: true, Fullscreen: true},
			ok:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := nativeMaximizeButtonSystemCommand(test.state)
			if ok != test.ok || got != test.want {
				t.Fatalf("native maximize command = (%#x, %v), want (%#x, %v)", got, ok, test.want, test.ok)
			}
		})
	}
}

func TestNativeMaximizeButtonPostsSystemCommandDuringNativeCallback(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{Resizable: true, Alive: true})
	entry.nativeCallbackDepth.Store(1)
	defer entry.nativeCallbackDepth.Store(0)

	var got struct {
		hwnd   uintptr
		msg    uintptr
		wparam uintptr
		lparam uintptr
	}
	posted := nativePostMaximizeButtonCommandWith(entry, 0x46575577, func(hwnd, msg, wparam, lparam uintptr) bool {
		got.hwnd = hwnd
		got.msg = msg
		got.wparam = wparam
		got.lparam = lparam
		return true
	})
	if !posted {
		t.Fatal("expected native maximize callback to post a deferred system command")
	}
	if got.hwnd != 0x46575577 || got.msg != nativeWMSysCommand || got.wparam != nativeSCMaximize || got.lparam != 0 {
		t.Fatalf("posted command = %#v, want hwnd=%#x WM_SYSCOMMAND/SC_MAXIMIZE", got, 0x46575577)
	}
	if state := entry.snapshot(); state.Maximized {
		t.Fatalf("posting during native callback mutated state synchronously: %#v", state)
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
