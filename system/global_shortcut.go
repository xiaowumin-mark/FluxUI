package system

import (
	"context"
	"fmt"
	"strings"
)

// GlobalShortcutModifiers is a bitmask of modifier keys.
type GlobalShortcutModifiers uint32

const (
	GlobalShortcutShift GlobalShortcutModifiers = 1 << iota
	GlobalShortcutControl
	GlobalShortcutAlt
	GlobalShortcutMeta
)

// GlobalShortcutSpec describes one system-wide shortcut.
type GlobalShortcutSpec struct {
	ID        string
	Key       string
	Modifiers GlobalShortcutModifiers
}

// GlobalShortcutEvent is delivered when a registered global shortcut fires.
type GlobalShortcutEvent struct {
	ID        string
	Key       string
	Modifiers GlobalShortcutModifiers
}

type globalShortcutDriver interface {
	registerGlobalShortcut(ctx context.Context, spec GlobalShortcutSpec, fn func(GlobalShortcutEvent)) (globalShortcutHandle, error)
}

type globalShortcutHandle interface {
	events() <-chan GlobalShortcutEvent
	close() error
}

// GlobalShortcut owns one registered system-wide shortcut.
type GlobalShortcut struct {
	handle globalShortcutHandle
}

// RegisterGlobalShortcut registers a system-wide shortcut.
//
// The callback is optional; callers can also read events from
// (*GlobalShortcut).Events().
func RegisterGlobalShortcut(ctx context.Context, spec GlobalShortcutSpec, fn func(GlobalShortcutEvent)) (*GlobalShortcut, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	spec, err := normalizeGlobalShortcutSpec(spec)
	if err != nil {
		return nil, err
	}

	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	gd, ok := d.(globalShortcutDriver)
	if !ok || d == nil || !d.capabilities().Supports(CapabilityGlobalShortcut) {
		return nil, fmt.Errorf("system: %s: %w", CapabilityGlobalShortcut, ErrUnsupported)
	}
	handle, err := gd.registerGlobalShortcut(ctx, spec, fn)
	if err != nil {
		return nil, err
	}
	if handle == nil {
		return nil, fmt.Errorf("system: %s: register: %w", CapabilityGlobalShortcut, ErrUnavailable)
	}
	return &GlobalShortcut{handle: handle}, nil
}

// Events returns the shortcut event channel.
func (s *GlobalShortcut) Events() <-chan GlobalShortcutEvent {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.events()
}

// Close unregisters the global shortcut.
func (s *GlobalShortcut) Close() error {
	if s == nil || s.handle == nil {
		return globalShortcutClosedError()
	}
	return s.handle.close()
}

func normalizeGlobalShortcutSpec(spec GlobalShortcutSpec) (GlobalShortcutSpec, error) {
	spec.ID = strings.TrimSpace(spec.ID)
	spec.Key = strings.ToUpper(strings.TrimSpace(spec.Key))
	if spec.ID == "" {
		return spec, fmt.Errorf("system: %s: id is empty", CapabilityGlobalShortcut)
	}
	if spec.Key == "" {
		return spec, fmt.Errorf("system: %s: key is empty", CapabilityGlobalShortcut)
	}
	if spec.Modifiers == 0 {
		return spec, fmt.Errorf("system: %s: modifiers are empty", CapabilityGlobalShortcut)
	}
	return spec, nil
}

func globalShortcutClosedError() error {
	return fmt.Errorf("system: %s: %w", CapabilityGlobalShortcut, ErrClosed)
}
