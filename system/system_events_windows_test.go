//go:build windows

package system

import (
	"sync"
	"testing"
)

func TestWindowsSystemEventFromMessage(t *testing.T) {
	tests := []struct {
		name       string
		msg        uint32
		wParam     uintptr
		lParam     uintptr
		wantKind   SystemEventKind
		wantDetail string
		wantOK     bool
	}{
		{
			name:       "display",
			msg:        wmDisplayChange,
			lParam:     uintptr(1080)<<16 | 1920,
			wantKind:   SystemEventDisplayChanged,
			wantDetail: "1920x1080",
			wantOK:     true,
		},
		{
			name:       "dpi",
			msg:        wmDPIChanged,
			wParam:     144,
			wantKind:   SystemEventDPIChanged,
			wantDetail: "144",
			wantOK:     true,
		},
		{
			name:     "theme",
			msg:      wmThemeChanged,
			wantKind: SystemEventThemeChanged,
			wantOK:   true,
		},
		{
			name:       "power suspend",
			msg:        wmPowerBroadcast,
			wParam:     pbtAPMSuspend,
			wantKind:   SystemEventPowerChanged,
			wantDetail: "suspend",
			wantOK:     true,
		},
		{
			name:       "session",
			msg:        wmWTSSession,
			wParam:     7,
			wantKind:   SystemEventSessionChanged,
			wantDetail: "7",
			wantOK:     true,
		},
		{name: "unknown", msg: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, ok := windowsSystemEventFromMessage(tt.msg, tt.wParam, tt.lParam)
			if ok != tt.wantOK {
				t.Fatalf("expected ok %v, got %v", tt.wantOK, ok)
			}
			if !tt.wantOK {
				return
			}
			if event.Kind != tt.wantKind || event.Detail != tt.wantDetail {
				t.Fatalf("unexpected event: %#v", event)
			}
		})
	}
}

func TestSystemEventFilter(t *testing.T) {
	all := &windowsSystemEventSubscription{}
	if !all.accepts(SystemEventDisplayChanged) {
		t.Fatal("empty filter should accept all event kinds")
	}

	filtered := &windowsSystemEventSubscription{
		filter: systemEventFilter([]SystemEventKind{SystemEventThemeChanged}),
	}
	if !filtered.accepts(SystemEventThemeChanged) {
		t.Fatal("expected filtered subscription to accept theme events")
	}
	if filtered.accepts(SystemEventDisplayChanged) {
		t.Fatal("filtered subscription should reject display events")
	}
}

func TestWindowsSystemEventBroadcastConcurrentWithRemove(t *testing.T) {
	const broadcasters = 8
	state := newWindowsSystemEventState()
	sub := state.add(nil)
	event := SystemEvent{Kind: SystemEventThemeChanged}

	start := make(chan struct{})
	var wg sync.WaitGroup
	for range broadcasters {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range 1_000 {
				state.broadcast(event)
			}
		}()
	}

	close(start)
	state.remove(sub.id)
	wg.Wait()

	for range sub.ch {
	}
}
