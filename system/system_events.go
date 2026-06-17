package system

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SystemEventKind identifies a system-level event category.
type SystemEventKind string

const (
	SystemEventDisplayChanged  SystemEventKind = "display_changed"
	SystemEventDPIChanged      SystemEventKind = "dpi_changed"
	SystemEventThemeChanged    SystemEventKind = "theme_changed"
	SystemEventSettingsChanged SystemEventKind = "settings_changed"
	SystemEventPowerChanged    SystemEventKind = "power_changed"
	SystemEventSessionChanged  SystemEventKind = "session_changed"
)

// SystemEvent records a system-level event reported by a platform driver.
type SystemEvent struct {
	Kind   SystemEventKind
	Detail string
	Time   time.Time
}

type systemEventDriver interface {
	subscribeSystemEvents(ctx context.Context, kinds []SystemEventKind) (systemEventHandle, error)
}

type systemEventHandle interface {
	events() <-chan SystemEvent
	close() error
}

// SystemEventSubscription is a handle returned by SubscribeSystemEvents.
type SystemEventSubscription struct {
	mu     sync.Mutex
	handle systemEventHandle
}

// SubscribeSystemEvents subscribes to system-level events.
//
// Passing no kinds subscribes to every event the active driver can report.
func SubscribeSystemEvents(ctx context.Context, kinds ...SystemEventKind) (*SystemEventSubscription, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validateSystemEventKinds(kinds); err != nil {
		return nil, err
	}

	d, supported := currentDriverFor(CapabilitySystemEvents)
	sd, ok := d.(systemEventDriver)
	if !ok || !supported {
		return nil, fmt.Errorf("system: %s: %w", CapabilitySystemEvents, ErrUnsupported)
	}

	handle, err := sd.subscribeSystemEvents(ctx, cloneSystemEventKinds(kinds))
	if err != nil {
		return nil, err
	}
	if handle == nil {
		return nil, fmt.Errorf("system: %s: %w", CapabilitySystemEvents, ErrUnavailable)
	}
	return &SystemEventSubscription{handle: handle}, nil
}

// Events returns the subscription event channel.
func (s *SystemEventSubscription) Events() <-chan SystemEvent {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	handle := s.handle
	s.mu.Unlock()
	if handle == nil {
		return nil
	}
	return handle.events()
}

// Close closes the subscription and releases platform resources owned by it.
func (s *SystemEventSubscription) Close() error {
	if s == nil {
		return ErrClosed
	}

	s.mu.Lock()
	handle := s.handle
	s.handle = nil
	s.mu.Unlock()
	if handle == nil {
		return ErrClosed
	}
	return handle.close()
}

func cloneSystemEventKinds(kinds []SystemEventKind) []SystemEventKind {
	if len(kinds) == 0 {
		return nil
	}
	cloned := make([]SystemEventKind, len(kinds))
	copy(cloned, kinds)
	return cloned
}

func validateSystemEventKinds(kinds []SystemEventKind) error {
	for _, kind := range kinds {
		if kind == "" {
			continue
		}
		if !validSystemEventKind(kind) {
			return fmt.Errorf("system: %s: event kind %q: %w", CapabilitySystemEvents, kind, ErrUnsupported)
		}
	}
	return nil
}

func validSystemEventKind(kind SystemEventKind) bool {
	switch kind {
	case SystemEventDisplayChanged,
		SystemEventDPIChanged,
		SystemEventThemeChanged,
		SystemEventSettingsChanged,
		SystemEventPowerChanged,
		SystemEventSessionChanged:
		return true
	default:
		return false
	}
}
