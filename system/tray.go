package system

import (
	"errors"
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
	Default   bool
	Separator bool
	Children  TrayMenu
	OnClick   func(TrayEvent)
}

// TrayOption configures a tray icon at creation time.
type TrayOption func(*trayOptions)

type trayOptions struct {
	icon          string
	iconData      []byte
	iconResource  uint16
	tooltip       string
	menu          TrayMenu
	menuProvider  func() TrayMenu
	onClick       func(TrayEvent)
	onDoubleClick func(TrayEvent)
}

type trayDriver interface {
	newTray(opts trayOptions) (trayHandle, error)
}

type trayHandle interface {
	setIcon(path string) error
	setIconData(data []byte) error
	setIconResource(id uint16) error
	setTooltip(text string) error
	setMenu(menu TrayMenu) error
	setMenuProvider(fn func() TrayMenu) error
	show() error
	hide() error
	close() error
}

// Tray represents a long-lived system tray icon.
type Tray struct {
	mu      sync.Mutex
	handle  trayHandle
	visible bool
	closed  bool
}

var (
	trayRegistryMu sync.Mutex
	trayRegistry   = make(map[*Tray]struct{})
)

// TrayIcon sets the tray icon path.
func TrayIcon(path string) TrayOption {
	return func(opts *trayOptions) {
		opts.icon = path
		opts.iconData = nil
		opts.iconResource = 0
	}
}

// TrayIconBytes sets the tray icon from ICO bytes.
func TrayIconBytes(data []byte) TrayOption {
	return func(opts *trayOptions) {
		opts.icon = ""
		opts.iconData = append([]byte(nil), data...)
		opts.iconResource = 0
	}
}

// TrayIconResource sets the tray icon from a platform resource ID.
//
// On Windows the ID is loaded from the current process module.
func TrayIconResource(id uint16) TrayOption {
	return func(opts *trayOptions) {
		opts.icon = ""
		opts.iconData = nil
		opts.iconResource = id
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

// TrayMenuProvider sets a callback that is evaluated each time a platform driver
// needs to show a tray menu.
func TrayMenuProvider(fn func() TrayMenu) TrayOption {
	return func(opts *trayOptions) {
		opts.menuProvider = fn
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

	d, supported := currentDriverFor(CapabilityTray)
	td, ok := d.(trayDriver)
	if !ok || !supported {
		return nil, fmt.Errorf("system: %s: %w", CapabilityTray, ErrUnsupported)
	}

	handle, err := td.newTray(opts.cloneForDriver())
	if err != nil {
		return nil, err
	}
	if handle == nil {
		return nil, fmt.Errorf("system: %s: create: %w", CapabilityTray, ErrUnavailable)
	}
	tray := &Tray{handle: handle}
	registerTray(tray)
	return tray, nil
}

// SetIcon updates the tray icon path.
func (t *Tray) SetIcon(path string) error {
	return t.withHandle(func(handle trayHandle) error {
		return handle.setIcon(path)
	})
}

// SetIconBytes updates the tray icon from ICO bytes.
func (t *Tray) SetIconBytes(data []byte) error {
	data = append([]byte(nil), data...)
	return t.withHandle(func(handle trayHandle) error {
		return handle.setIconData(data)
	})
}

// SetIconResource updates the tray icon from a platform resource ID.
func (t *Tray) SetIconResource(id uint16) error {
	return t.withHandle(func(handle trayHandle) error {
		return handle.setIconResource(id)
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

// SetMenuProvider updates the dynamic tray menu provider.
func (t *Tray) SetMenuProvider(fn func() TrayMenu) error {
	return t.withHandle(func(handle trayHandle) error {
		return handle.setMenuProvider(asyncTrayMenuProvider(fn))
	})
}

// Show displays the tray icon.
func (t *Tray) Show() error {
	handle, err := t.handleSnapshot()
	if err != nil {
		return err
	}
	if err := handle.show(); err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return trayClosedError()
	}
	t.visible = true
	return nil
}

// Hide hides the tray icon.
func (t *Tray) Hide() error {
	handle, err := t.handleSnapshot()
	if err != nil {
		return err
	}
	if err := handle.hide(); err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return trayClosedError()
	}
	t.visible = false
	return nil
}

// Visible reports whether the tray icon has been shown and not hidden or closed.
func (t *Tray) Visible() bool {
	if t == nil {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return !t.closed && t.visible
}

// Closed reports whether the tray has been closed.
func (t *Tray) Closed() bool {
	if t == nil {
		return true
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.closed
}

// Close releases the tray icon and associated native resources.
func (t *Tray) Close() error {
	if t == nil {
		return trayClosedError()
	}
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return trayClosedError()
	}
	t.closed = true
	t.visible = false
	handle := t.handle
	t.mu.Unlock()

	unregisterTray(t)
	if handle == nil {
		return fmt.Errorf("system: %s: close: %w", CapabilityTray, ErrUnavailable)
	}
	return handle.close()
}

func (t *Tray) withHandle(fn func(trayHandle) error) error {
	handle, err := t.handleSnapshot()
	if err != nil {
		return err
	}
	return fn(handle)
}

func (t *Tray) handleSnapshot() (trayHandle, error) {
	if t == nil {
		return nil, trayClosedError()
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil, trayClosedError()
	}
	if t.handle == nil {
		return nil, fmt.Errorf("system: %s: %w", CapabilityTray, ErrUnavailable)
	}
	return t.handle, nil
}

func defaultTrayOptions() trayOptions {
	return trayOptions{}
}

func (opts trayOptions) cloneForDriver() trayOptions {
	opts.iconData = append([]byte(nil), opts.iconData...)
	opts.menu = cloneTrayMenuForDriver(opts.menu)
	opts.menuProvider = asyncTrayMenuProvider(opts.menuProvider)
	opts.onClick = asyncTrayEventHandler(opts.onClick)
	opts.onDoubleClick = asyncTrayEventHandler(opts.onDoubleClick)
	return opts
}

func cloneTrayMenuForDriver(menu TrayMenu) TrayMenu {
	cloned := cloneTrayMenu(menu)
	for i := range cloned {
		cloned[i].OnClick = asyncTrayEventHandler(cloned[i].OnClick)
		cloned[i].Children = cloneTrayMenuForDriver(cloned[i].Children)
	}
	return cloned
}

func cloneTrayMenu(menu TrayMenu) TrayMenu {
	if len(menu) == 0 {
		return nil
	}
	cloned := make(TrayMenu, len(menu))
	for i, item := range menu {
		cloned[i] = item
		cloned[i].Children = cloneTrayMenu(item.Children)
	}
	return cloned
}

func asyncTrayMenuProvider(fn func() TrayMenu) func() TrayMenu {
	if fn == nil {
		return nil
	}
	return func() TrayMenu {
		return cloneTrayMenuForDriver(fn())
	}
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

func registerTray(tray *Tray) {
	if tray == nil {
		return
	}
	trayRegistryMu.Lock()
	trayRegistry[tray] = struct{}{}
	trayRegistryMu.Unlock()
}

func unregisterTray(tray *Tray) {
	if tray == nil {
		return
	}
	trayRegistryMu.Lock()
	delete(trayRegistry, tray)
	trayRegistryMu.Unlock()
}

// CloseTrays closes all currently registered system trays.
func CloseTrays() error {
	trayRegistryMu.Lock()
	trays := make([]*Tray, 0, len(trayRegistry))
	for tray := range trayRegistry {
		trays = append(trays, tray)
	}
	trayRegistryMu.Unlock()

	errs := make([]error, 0, len(trays))
	for _, tray := range trays {
		if err := tray.Close(); err != nil && !IsClosed(err) {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
