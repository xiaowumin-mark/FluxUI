package system

import (
	"fmt"
	"sync"
)

// TrayEventKind identifies a tray icon or menu event.
type TrayEventKind string

const (
	TrayEventClicked     TrayEventKind = "clicked"
	TrayEventDoubleClick TrayEventKind = "double_click"
	TrayEventMenuItem    TrayEventKind = "menu_item"
)

// TrayEvent records a tray icon or menu event.
type TrayEvent struct {
	Kind   TrayEventKind
	ItemID string
}

// TrayMenu is an ordered list of tray menu items.
type TrayMenu []TrayMenuItem

// TrayMenuItem defines one item in a tray menu.
type TrayMenuItem struct {
	ID        string
	Label     string
	Disabled  bool
	Checked   bool
	Separator bool
	OnClick   func(TrayEvent)
}

// TrayOption configures a tray icon at creation time.
type TrayOption func(*trayOptions)

type trayOptions struct {
	icon          string
	tooltip       string
	menu          TrayMenu
	onClick       func(TrayEvent)
	onDoubleClick func(TrayEvent)
}

type trayDriver interface {
	newTray(opts trayOptions) (trayHandle, error)
}

type trayHandle interface {
	setIcon(path string) error
	setTooltip(text string) error
	setMenu(menu TrayMenu) error
	show() error
	hide() error
	close() error
}

// Tray represents a long-lived system tray icon.
type Tray struct {
	mu     sync.Mutex
	handle trayHandle
	closed bool
}

// TrayIcon sets the tray icon path.
func TrayIcon(path string) TrayOption {
	return func(opts *trayOptions) {
		opts.icon = path
	}
}

// TrayTooltip sets the tray tooltip text.
func TrayTooltip(text string) TrayOption {
	return func(opts *trayOptions) {
		opts.tooltip = text
	}
}

// TrayMenuItems sets the tray menu.
func TrayMenuItems(items ...TrayMenuItem) TrayOption {
	return func(opts *trayOptions) {
		opts.menu = cloneTrayMenu(items)
	}
}

// TrayOnClick registers a left-click callback for drivers that can report it.
func TrayOnClick(fn func(TrayEvent)) TrayOption {
	return func(opts *trayOptions) {
		opts.onClick = fn
	}
}

// TrayOnDoubleClick registers a double-click callback for drivers that can report it.
func TrayOnDoubleClick(fn func(TrayEvent)) TrayOption {
	return func(opts *trayOptions) {
		opts.onDoubleClick = fn
	}
}

// TrayMenuAction creates a clickable tray menu item.
func TrayMenuAction(id, label string, onClick func(TrayEvent)) TrayMenuItem {
	return TrayMenuItem{
		ID:      id,
		Label:   label,
		OnClick: onClick,
	}
}

// TrayMenuSeparator creates a tray menu separator item.
func TrayMenuSeparator() TrayMenuItem {
	return TrayMenuItem{Separator: true}
}

// NewTray creates a system tray icon using the active platform driver.
func NewTray(options ...TrayOption) (*Tray, error) {
	opts := defaultTrayOptions()
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}

	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	td, ok := d.(trayDriver)
	if !ok || d == nil || !d.capabilities().Supports(CapabilityTray) {
		return nil, fmt.Errorf("system: %s: %w", CapabilityTray, ErrUnsupported)
	}

	handle, err := td.newTray(opts.cloneForDriver())
	if err != nil {
		return nil, err
	}
	if handle == nil {
		return nil, fmt.Errorf("system: %s: create: %w", CapabilityTray, ErrUnavailable)
	}
	return &Tray{handle: handle}, nil
}

// SetIcon updates the tray icon path.
func (t *Tray) SetIcon(path string) error {
	return t.withHandle(func(handle trayHandle) error {
		return handle.setIcon(path)
	})
}

// SetTooltip updates the tray tooltip text.
func (t *Tray) SetTooltip(text string) error {
	return t.withHandle(func(handle trayHandle) error {
		return handle.setTooltip(text)
	})
}

// SetMenu updates the tray menu.
func (t *Tray) SetMenu(menu TrayMenu) error {
	menu = cloneTrayMenuForDriver(menu)
	return t.withHandle(func(handle trayHandle) error {
		return handle.setMenu(menu)
	})
}

// Show displays the tray icon.
func (t *Tray) Show() error {
	return t.withHandle(func(handle trayHandle) error {
		return handle.show()
	})
}

// Hide hides the tray icon.
func (t *Tray) Hide() error {
	return t.withHandle(func(handle trayHandle) error {
		return handle.hide()
	})
}

// Close releases the tray icon and associated native resources.
func (t *Tray) Close() error {
	if t == nil {
		return trayClosedError()
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return trayClosedError()
	}
	t.closed = true
	if t.handle == nil {
		return fmt.Errorf("system: %s: close: %w", CapabilityTray, ErrUnavailable)
	}
	return t.handle.close()
}

func (t *Tray) withHandle(fn func(trayHandle) error) error {
	if t == nil {
		return trayClosedError()
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return trayClosedError()
	}
	if t.handle == nil {
		return fmt.Errorf("system: %s: %w", CapabilityTray, ErrUnavailable)
	}
	return fn(t.handle)
}

func defaultTrayOptions() trayOptions {
	return trayOptions{}
}

func (opts trayOptions) cloneForDriver() trayOptions {
	opts.menu = cloneTrayMenuForDriver(opts.menu)
	opts.onClick = asyncTrayEventHandler(opts.onClick)
	opts.onDoubleClick = asyncTrayEventHandler(opts.onDoubleClick)
	return opts
}

func cloneTrayMenuForDriver(menu TrayMenu) TrayMenu {
	cloned := cloneTrayMenu(menu)
	for i := range cloned {
		cloned[i].OnClick = asyncTrayEventHandler(cloned[i].OnClick)
	}
	return cloned
}

func cloneTrayMenu(menu TrayMenu) TrayMenu {
	if len(menu) == 0 {
		return nil
	}
	cloned := make(TrayMenu, len(menu))
	copy(cloned, menu)
	return cloned
}

func asyncTrayEventHandler(fn func(TrayEvent)) func(TrayEvent) {
	if fn == nil {
		return nil
	}
	return func(event TrayEvent) {
		go fn(event)
	}
}

func trayClosedError() error {
	return fmt.Errorf("system: %s: %w", CapabilityTray, ErrClosed)
}
