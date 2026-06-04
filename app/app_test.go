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
		Decorated: true,
		Alive:     true,
	})

	handle := WindowHandle{id: entry.id}
	state, ok := handle.State()
	if !ok {
		t.Fatal("expected state for live window")
	}
	if state.ID != entry.id || state.Title != "Initial" || state.Width != 320 || state.Height != 240 || !state.Alive {
		t.Fatalf("unexpected state: %#v", state)
	}
}

func TestWindowHandleStateMutations(t *testing.T) {
	entry := testRegisterWindow(t, WindowState{
		Title:  "Initial",
		Width:  320,
		Height: 240,
		Alive:  true,
	})

	handle := WindowHandle{id: entry.id}
	if !handle.SetTitle("Updated") {
		t.Fatal("expected SetTitle to succeed")
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
	if state.Title != "Updated" || state.Width != 640 || state.Height != 480 {
		t.Fatalf("unexpected title/size: %#v", state)
	}
	if state.MinWidth != 300 || state.MinHeight != 200 || state.MaxWidth != 900 || state.MaxHeight != 700 {
		t.Fatalf("unexpected size constraints: %#v", state)
	}
	if !state.Fullscreen || state.Minimized || state.Maximized {
		t.Fatalf("unexpected mode flags: %#v", state)
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

	entry.alive.Store(false)
	entry.pushEvent(WindowEventClosed, func(state *WindowState) {
		state.Alive = false
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
