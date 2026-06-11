package app

import (
	"image"
	"testing"

	gioApp "gioui.org/app"
	"gioui.org/unit"
)

func TestWindowStateSnapshot(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:     "Initial",
		Width:     320,
		Height:    240,
		Visible:   true,
		Decorated: true,
		Alive:     true,
	})

	handle := WindowHandle{id: entry.id}
	state, ok := handle.State()
	if !ok {
		t.Fatal("expected state for live window")
	}
	if state.ID != entry.id || state.Title != "Initial" || state.Width != 320 || state.Height != 240 || !state.Visible || !state.Alive {
		t.Fatalf("unexpected state: %#v", state)
	}
}

func TestWindowHandleStateMutations(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:     "Initial",
		Width:     320,
		Height:    240,
		Decorated: true,
		Alive:     true,
	})

	handle := WindowHandle{id: entry.id}
	if !handle.SetTitle("Updated") {
		t.Fatal("expected SetTitle to succeed")
	}
	if !handle.SetDecorated(false) {
		t.Fatal("expected SetDecorated to succeed")
	}
	if !handle.SetSize(640, 480) {
		t.Fatal("expected SetSize to succeed")
	}
	if !handle.SetMinSize(300, 200) {
		t.Fatal("expected SetMinSize to succeed")
	}
	if !handle.SetMaxSize(900, 700) {
		t.Fatal("expected SetMaxSize to succeed")
	}
	if !handle.Fullscreen() {
		t.Fatal("expected Fullscreen to succeed")
	}

	state, ok := handle.State()
	if !ok {
		t.Fatal("expected state for live window")
	}
	if state.Title != "Updated" || state.Width != 640 || state.Height != 480 || state.Decorated {
		t.Fatalf("unexpected title/size: %#v", state)
	}
	if state.MinWidth != 300 || state.MinHeight != 200 || state.MaxWidth != 900 || state.MaxHeight != 700 {
		t.Fatalf("unexpected size constraints: %#v", state)
	}
	if !state.Fullscreen || state.Minimized || state.Maximized {
		t.Fatalf("unexpected mode flags: %#v", state)
	}
}

func TestWindowMaximizeBlockedByMaxSize(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:     "Initial",
		Width:     320,
		Height:    240,
		MaxWidth:  900,
		MaxHeight: 700,
		Alive:     true,
	})
	handle := WindowHandle{id: entry.id}

	if handle.Maximize() {
		t.Fatal("expected Maximize to fail when max size is set")
	}
	state, ok := handle.State()
	if !ok {
		t.Fatal("expected state for live window")
	}
	if state.Maximized {
		t.Fatalf("max size constrained window should not enter maximized state: %#v", state)
	}
}

func TestWindowSetMaxSizeBlocksFutureMaximize(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:     "Initial",
		Width:     320,
		Height:    240,
		Scale:     1,
		TextScale: 1,
		DPI:       96,
		Decorated: true,
		Alive:     true,
	})
	entry.metric = unit.Metric{PxPerDp: 1, PxPerSp: 1}
	handle := WindowHandle{id: entry.id}

	if !handle.SetMaxSize(900, 700) {
		t.Fatal("expected SetMaxSize to succeed")
	}
	if handle.Maximize() {
		t.Fatal("expected Maximize to fail after SetMaxSize")
	}
	state, ok := handle.State()
	if !ok {
		t.Fatal("expected state for live window")
	}
	if state.MaxWidth != 900 || state.MaxHeight != 700 || state.Maximized {
		t.Fatalf("unexpected constrained state: %#v", state)
	}
}

func TestWindowCloseRequestedHandler(t *testing.T) {
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
		if request.State.Title != "Unsaved" || !request.State.Alive {
			t.Fatalf("unexpected request state: %#v", request.State)
		}
		return false
	}) {
		t.Fatal("expected SetCloseRequestedHandler to succeed")
	}
	if entry.handleCloseRequested() {
		t.Fatal("expected close request to be canceled")
	}
	if calls != 1 {
		t.Fatalf("expected one handler call, got %d", calls)
	}
	events := handle.PollEvents()
	if !hasWindowEvent(events, WindowEventCloseRequested) {
		t.Fatalf("expected close requested event, got %#v", events)
	}

	if !handle.SetCloseRequestedHandler(func(WindowCloseRequest) bool {
		return true
	}) {
		t.Fatal("expected SetCloseRequestedHandler to update handler")
	}
	if !entry.handleCloseRequested() {
		t.Fatal("expected close request to be allowed")
	}
}

func TestWindowEventSubscription(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:     "Initial",
		Width:     320,
		Height:    240,
		Scale:     1,
		TextScale: 1,
		DPI:       96,
		Decorated: true,
		Alive:     true,
	})
	entry.metric = unit.Metric{PxPerDp: 1, PxPerSp: 1}
	handle := WindowHandle{id: entry.id}

	sub, ok := handle.SubscribeEvents(WindowEventStateChanged)
	if !ok {
		t.Fatal("expected SubscribeEvents to succeed")
	}
	if sub.Events() == nil {
		t.Fatal("expected subscription event channel")
	}

	if !handle.SetDecorated(false) {
		t.Fatal("expected SetDecorated to succeed")
	}
	select {
	case event := <-sub.Events():
		if event.Kind != WindowEventStateChanged || event.State.Decorated {
			t.Fatalf("unexpected subscription event: %#v", event)
		}
	default:
		t.Fatal("expected state changed event")
	}

	entry.updateFromFrame(imagePoint(640, 480), unit.Metric{PxPerDp: 1, PxPerSp: 1})
	select {
	case event := <-sub.Events():
		t.Fatalf("size event should be filtered out, got %#v", event)
	default:
	}

	if !sub.Close() {
		t.Fatal("expected subscription Close to succeed")
	}
	if sub.Close() {
		t.Fatal("second subscription Close should fail")
	}
	if _, ok := <-sub.Events(); ok {
		t.Fatal("expected subscription channel to close")
	}
}

func TestWindowScaleChangedEvent(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:     "Initial",
		Width:     320,
		Height:    240,
		Scale:     1,
		TextScale: 1,
		DPI:       96,
		Alive:     true,
	})
	entry.metric = unit.Metric{PxPerDp: 1, PxPerSp: 1}
	handle := WindowHandle{id: entry.id}

	sub, ok := handle.SubscribeEvents(WindowEventScaleChanged)
	if !ok {
		t.Fatal("expected SubscribeEvents to succeed")
	}
	defer sub.Close()

	entry.updateFromFrame(imagePoint(1280, 960), unit.Metric{PxPerDp: 2, PxPerSp: 2})
	select {
	case event := <-sub.Events():
		if event.Kind != WindowEventScaleChanged {
			t.Fatalf("expected scale changed event, got %#v", event)
		}
		if event.State.Scale != 2 || event.State.TextScale != 2 || event.State.DPI != 192 {
			t.Fatalf("unexpected scale state: %#v", event.State)
		}
	default:
		t.Fatal("expected scale changed event")
	}
}

func TestWindowEventSubscriptionClosesOnUnregister(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:  "Initial",
		Width:  320,
		Height: 240,
		Alive:  true,
	})
	sub, ok := (WindowHandle{id: entry.id}).SubscribeEvents()
	if !ok {
		t.Fatal("expected SubscribeEvents to succeed")
	}

	unregisterWindow(entry.id)
	if _, ok := <-sub.Events(); ok {
		t.Fatal("expected subscription channel to close when window unregisters")
	}
}

func TestWindowFullscreenConfigPreservesRequestedConstraints(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:     "Initial",
		Width:     640,
		Height:    480,
		MinWidth:  360,
		MinHeight: 240,
		MaxWidth:  900,
		MaxHeight: 700,
		Alive:     true,
	})
	entry.metric = unit.Metric{PxPerDp: 2, PxPerSp: 2}
	handle := WindowHandle{id: entry.id}

	if !handle.Fullscreen() {
		t.Fatal("expected Fullscreen to succeed")
	}
	entry.updateFromConfig(gioApp.Config{
		Mode:    gioApp.Fullscreen,
		Size:    imagePoint(1920, 1080),
		MinSize: imagePoint(0, 0),
		MaxSize: imagePoint(0, 0),
	})

	state, ok := handle.State()
	if !ok {
		t.Fatal("expected state for live window")
	}
	if !state.Fullscreen {
		t.Fatalf("expected fullscreen state: %#v", state)
	}
	if state.MinWidth != 360 || state.MinHeight != 240 || state.MaxWidth != 900 || state.MaxHeight != 700 {
		t.Fatalf("fullscreen config should not drop requested constraints: %#v", state)
	}
}

func TestNativeWindowControlsRequireNativeHandle(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:  "Initial",
		Width:  320,
		Height: 240,
		Alive:  true,
	})
	handle := WindowHandle{id: entry.id}

	if handle.SetAlwaysOnTop(true) {
		t.Fatal("expected SetAlwaysOnTop without native handle to fail")
	}
	if handle.Hide() {
		t.Fatal("expected Hide without native handle to fail")
	}
	if handle.Show() {
		t.Fatal("expected Show without native handle to fail")
	}
	if handle.SetPosition(20, 40) {
		t.Fatal("expected SetPosition without native handle to fail")
	}
	if handle.SetResizable(false) {
		t.Fatal("expected SetResizable without native handle to fail")
	}
	if !handle.RequestFocus() {
		t.Fatal("expected RequestFocus to fall back to Gio raise request")
	}
}

func TestWindowHiddenMemoryPolicyState(t *testing.T) {
	state := WindowState{
		Visible:            true,
		HiddenMemoryPolicy: WindowHiddenMemoryReleaseTransient,
	}
	applyWindowRenderSuspendedState(&state)
	if state.RenderSuspended {
		t.Fatalf("visible window should not suspend rendering: %#v", state)
	}

	state.Visible = false
	applyWindowRenderSuspendedState(&state)
	if !state.RenderSuspended {
		t.Fatalf("hidden release-transient window should suspend rendering: %#v", state)
	}

	state.HiddenMemoryPolicy = WindowHiddenMemoryKeepRenderingState
	applyWindowRenderSuspendedState(&state)
	if state.RenderSuspended {
		t.Fatalf("keep-rendering policy should not suspend rendering: %#v", state)
	}
}

func TestWindowSetHiddenMemoryPolicy(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:              "Initial",
		Width:              320,
		Height:             240,
		Visible:            false,
		HiddenMemoryPolicy: WindowHiddenMemoryKeepRenderingState,
		Alive:              true,
	})
	handle := WindowHandle{id: entry.id}
	entry.updateState(func(state *WindowState) {
		state.Visible = false
		state.HiddenMemoryPolicy = WindowHiddenMemoryKeepRenderingState
		applyWindowRenderSuspendedState(state)
	})

	if handle.SetHiddenMemoryPolicy(WindowHiddenMemoryPolicy(99)) {
		t.Fatal("expected invalid hidden memory policy to fail")
	}
	if !handle.SetHiddenMemoryPolicy(WindowHiddenMemoryReleaseTransient) {
		t.Fatal("expected SetHiddenMemoryPolicy to succeed")
	}
	state, ok := handle.State()
	if !ok {
		t.Fatal("expected state for live window")
	}
	if state.HiddenMemoryPolicy != WindowHiddenMemoryReleaseTransient || !state.RenderSuspended {
		t.Fatalf("expected release-transient policy to suspend hidden window: %#v", state)
	}

	if !handle.SetHiddenMemoryPolicy(WindowHiddenMemoryKeepRenderingState) {
		t.Fatal("expected SetHiddenMemoryPolicy keep-rendering to succeed")
	}
	state, ok = handle.State()
	if !ok {
		t.Fatal("expected state for live window")
	}
	if state.HiddenMemoryPolicy != WindowHiddenMemoryKeepRenderingState || state.RenderSuspended {
		t.Fatalf("expected keep-rendering policy to resume hidden window rendering state: %#v", state)
	}
}

func TestWindowNativeHandle(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:  "Initial",
		Width:  320,
		Height: 240,
		Alive:  true,
	})
	handle := WindowHandle{id: entry.id}

	if native, ok := handle.NativeHandle(); ok || native != 0 {
		t.Fatalf("expected missing native handle, got %d ok=%v", native, ok)
	}

	entry.mu.Lock()
	entry.nativeHandle = 12345
	entry.mu.Unlock()

	if native, ok := handle.NativeHandle(); !ok || native != 12345 {
		t.Fatalf("expected native handle 12345, got %d ok=%v", native, ok)
	}
	if native, ok := (WindowHandle{id: entry.id + 1000}).NativeHandle(); ok || native != 0 {
		t.Fatalf("expected missing window native handle to fail, got %d ok=%v", native, ok)
	}

	entry.alive.Store(false)
	if native, ok := handle.NativeHandle(); ok || native != 0 {
		t.Fatalf("expected closed window native handle to fail, got %d ok=%v", native, ok)
	}
}

func TestWindowStateConfigAndFrameSync(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:  "Initial",
		Width:  320,
		Height: 240,
		Alive:  true,
	})
	entry.metric = unit.Metric{PxPerDp: 2, PxPerSp: 2}

	entry.updateFromConfig(gioApp.Config{
		Title:     "Configured",
		Size:      imagePoint(1000, 700),
		MinSize:   imagePoint(300, 200),
		MaxSize:   imagePoint(1400, 1000),
		Mode:      gioApp.Maximized,
		Focused:   true,
		Decorated: true,
	})
	entry.updateFromFrame(imagePoint(1200, 800), unit.Metric{PxPerDp: 2, PxPerSp: 2})

	state, ok := (WindowHandle{id: entry.id}).State()
	if !ok {
		t.Fatal("expected state for live window")
	}
	if state.Title != "Configured" || state.Width != 600 || state.Height != 400 {
		t.Fatalf("unexpected synced state: %#v", state)
	}
	if state.MinWidth != 150 || state.MinHeight != 100 || state.MaxWidth != 700 || state.MaxHeight != 500 {
		t.Fatalf("unexpected synced constraints: %#v", state)
	}
	if !state.Maximized || !state.Focused || !state.Decorated {
		t.Fatalf("unexpected config flags: %#v", state)
	}

	events := (WindowHandle{id: entry.id}).PollEvents()
	if !hasWindowEvent(events, WindowEventSizeChanged) {
		t.Fatalf("expected size changed event, got %#v", events)
	}
	if !hasWindowEvent(events, WindowEventScaleChanged) {
		t.Fatalf("expected scale changed event, got %#v", events)
	}
	if !hasWindowEvent(events, WindowEventFocusChanged) {
		t.Fatalf("expected focus changed event, got %#v", events)
	}
	if !hasWindowEvent(events, WindowEventStateChanged) {
		t.Fatalf("expected state changed event, got %#v", events)
	}
	if more := (WindowHandle{id: entry.id}).PollEvents(); len(more) != 0 {
		t.Fatalf("expected events to drain, got %#v", more)
	}
}

func TestWindowInvalidInputsAndClosedState(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:  "Initial",
		Width:  320,
		Height: 240,
		Alive:  true,
	})
	handle := WindowHandle{id: entry.id}

	if handle.SetSize(0, 100) {
		t.Fatal("expected invalid SetSize to fail")
	}
	if handle.SetMinSize(0, 100) {
		t.Fatal("expected invalid SetMinSize to fail")
	}
	if handle.SetMaxSize(100, 0) {
		t.Fatal("expected invalid SetMaxSize to fail")
	}
	if handle.SetHiddenMemoryPolicy(WindowHiddenMemoryPolicy(99)) {
		t.Fatal("expected invalid SetHiddenMemoryPolicy to fail")
	}
	if handle.SetAlwaysOnTop(true) {
		t.Fatal("expected SetAlwaysOnTop without native handle to fail")
	}
	if handle.Hide() {
		t.Fatal("expected Hide without native handle to fail")
	}
	if handle.Show() {
		t.Fatal("expected Show without native handle to fail")
	}

	entry.alive.Store(false)
	entry.pushEvent(WindowEventClosed, func(state *WindowState) {
		state.Alive = false
		state.Visible = false
	})
	if _, ok := handle.State(); ok {
		t.Fatal("expected closed window state lookup to fail")
	}
	if handle.SetTitle("Closed") {
		t.Fatal("expected closed window mutation to fail")
	}
	events := handle.PollEvents()
	if !hasWindowEvent(events, WindowEventClosed) {
		t.Fatalf("expected closed event, got %#v", events)
	}
}

func testRegisterWindow(t *testing.T, state WindowState) *windowEntry {
	t.Helper()

	entry := &windowEntry{
		id:    nextWindowID(),
		win:   new(gioApp.Window),
		state: state,
	}
	entry.state.ID = entry.id
	entry.state.Alive = true
	entry.state.Visible = true
	if !validWindowHiddenMemoryPolicy(entry.state.HiddenMemoryPolicy) {
		entry.state.HiddenMemoryPolicy = WindowHiddenMemoryReleaseTransient
	}
	applyWindowRenderSuspendedState(&entry.state)
	entry.alive.Store(true)
	registerWindow(entry)

	t.Cleanup(func() {
		entry.alive.Store(false)
		unregisterWindow(entry.id)
	})
	return entry
}

func imagePoint(x, y int) image.Point {
	return image.Point{X: x, Y: y}
}

func hasWindowEvent(events []WindowEvent, kind WindowEventKind) bool {
	for _, event := range events {
		if event.Kind == kind {
			return true
		}
	}
	return false
}
