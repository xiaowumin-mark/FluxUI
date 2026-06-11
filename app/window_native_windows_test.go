//go:build windows

package app

import "testing"

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
