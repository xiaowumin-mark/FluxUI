//go:build darwin

package system

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const darwinTrayProviderCommand = "osascript"

type darwinTrayHandle struct {
	mu             sync.Mutex
	opts           trayOptions
	tempDir        string
	eventPath      string
	iconDataPath   string
	cmd            *exec.Cmd
	cmdDone        chan error
	stopListener   chan struct{}
	eventOffset    int64
	eventRemainder string
	closed         bool
	visible        bool
}

func unixProbeTrayCapability(cap Capability) CapabilityAvailability {
	if _, err := exec.LookPath(darwinTrayProviderCommand); err != nil {
		return CapabilityAvailability{
			Capability: cap,
			Status:     CapabilityStatusUnavailable,
			Err:        fmt.Errorf("system: probe %s: %w: %v", cap, ErrUnavailable, err),
		}
	}
	return CapabilityAvailability{
		Capability: cap,
		Status:     CapabilityStatusAvailable,
	}
}

func (unixDriver) newTray(opts trayOptions) (trayHandle, error) {
	if _, err := exec.LookPath(darwinTrayProviderCommand); err != nil {
		return nil, fmt.Errorf("system: %s: %s unavailable: %w: %v", CapabilityTray, darwinTrayProviderCommand, ErrUnavailable, err)
	}
	if opts.iconResource != 0 {
		return nil, fmt.Errorf("system: %s: resource icons are unsupported by macOS tray command driver: %w", CapabilityTray, ErrUnsupported)
	}

	tempDir, err := os.MkdirTemp("", "fluxui-tray-*")
	if err != nil {
		return nil, fmt.Errorf("system: %s: create temp dir: %w", CapabilityTray, err)
	}
	handle := &darwinTrayHandle{
		opts:         opts,
		tempDir:      tempDir,
		eventPath:    filepath.Join(tempDir, "events.log"),
		stopListener: make(chan struct{}),
	}
	if err := os.WriteFile(handle.eventPath, nil, 0o600); err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, fmt.Errorf("system: %s: create event log: %w", CapabilityTray, err)
	}
	if len(opts.iconData) > 0 {
		if err := handle.writeIconDataLocked(opts.iconData); err != nil {
			_ = os.RemoveAll(tempDir)
			return nil, err
		}
	}
	go handle.listen()
	return handle, nil
}

func (h *darwinTrayHandle) setIcon(path string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureOpenLocked(); err != nil {
		return err
	}
	h.opts.icon = path
	h.opts.iconData = nil
	h.opts.iconResource = 0
	h.iconDataPath = ""
	return h.restartIfVisibleLocked()
}

func (h *darwinTrayHandle) setIconData(data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureOpenLocked(); err != nil {
		return err
	}
	if err := h.writeIconDataLocked(data); err != nil {
		return err
	}
	h.opts.icon = ""
	h.opts.iconData = append([]byte(nil), data...)
	h.opts.iconResource = 0
	return h.restartIfVisibleLocked()
}

func (h *darwinTrayHandle) setIconResource(id uint16) error {
	return fmt.Errorf("system: %s: resource icons are unsupported by macOS tray command driver: %w", CapabilityTray, ErrUnsupported)
}

func (h *darwinTrayHandle) setTooltip(text string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureOpenLocked(); err != nil {
		return err
	}
	h.opts.tooltip = text
	return h.restartIfVisibleLocked()
}

func (h *darwinTrayHandle) setMenu(menu TrayMenu) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureOpenLocked(); err != nil {
		return err
	}
	h.opts.menu = cloneTrayMenu(menu)
	h.opts.menuProvider = nil
	return h.restartIfVisibleLocked()
}

func (h *darwinTrayHandle) setMenuProvider(fn func() TrayMenu) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureOpenLocked(); err != nil {
		return err
	}
	h.opts.menuProvider = fn
	return h.restartIfVisibleLocked()
}

func (h *darwinTrayHandle) show() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureOpenLocked(); err != nil {
		return err
	}
	if h.visible && h.processRunningLocked() {
		return nil
	}
	if err := h.startLocked(); err != nil {
		return err
	}
	h.visible = true
	return nil
}

func (h *darwinTrayHandle) hide() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureOpenLocked(); err != nil {
		return err
	}
	h.stopProcessLocked()
	h.visible = false
	return nil
}

func (h *darwinTrayHandle) close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return trayClosedError()
	}
	h.closed = true
	h.visible = false
	h.stopProcessLocked()
	close(h.stopListener)
	return os.RemoveAll(h.tempDir)
}

func (h *darwinTrayHandle) ensureOpenLocked() error {
	if h == nil || h.closed {
		return trayClosedError()
	}
	return nil
}

func (h *darwinTrayHandle) restartIfVisibleLocked() error {
	if !h.visible {
		return nil
	}
	h.stopProcessLocked()
	return h.startLocked()
}

func (h *darwinTrayHandle) processRunningLocked() bool {
	if h.cmd == nil || h.cmdDone == nil {
		h.cmd = nil
		h.cmdDone = nil
		return false
	}
	select {
	case <-h.cmdDone:
		h.cmd = nil
		h.cmdDone = nil
		h.visible = false
		return false
	default:
		return true
	}
}

func (h *darwinTrayHandle) startLocked() error {
	scriptPath, err := h.writeScriptLocked()
	if err != nil {
		return err
	}
	cmd := exec.Command(darwinTrayProviderCommand, scriptPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("system: %s: start %s: %w: %v", CapabilityTray, darwinTrayProviderCommand, ErrUnavailable, err)
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		return fmt.Errorf("system: %s: %s exited during startup: %w: %v", CapabilityTray, darwinTrayProviderCommand, ErrUnavailable, err)
	case <-time.After(250 * time.Millisecond):
		h.cmd = cmd
		h.cmdDone = done
		return nil
	}
}

func (h *darwinTrayHandle) stopProcessLocked() {
	if h.cmd == nil {
		h.cmdDone = nil
		return
	}
	if h.cmd.Process != nil {
		_ = h.cmd.Process.Kill()
	}
	if h.cmdDone != nil {
		select {
		case <-h.cmdDone:
		case <-time.After(2 * time.Second):
		}
	}
	h.cmd = nil
	h.cmdDone = nil
}

func (h *darwinTrayHandle) writeScriptLocked() (string, error) {
	path := filepath.Join(h.tempDir, "tray.applescript")
	script := darwinTrayScript(h.opts, h.eventPath, h.iconPathLocked())
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		return "", fmt.Errorf("system: %s: write AppleScript: %w", CapabilityTray, err)
	}
	return path, nil
}

func (h *darwinTrayHandle) iconPathLocked() string {
	if h.iconDataPath != "" {
		return h.iconDataPath
	}
	return h.opts.icon
}

func (h *darwinTrayHandle) writeIconDataLocked(data []byte) error {
	if len(data) == 0 {
		h.iconDataPath = ""
		return nil
	}
	path := filepath.Join(h.tempDir, "icon.ico")
	if err := os.WriteFile(path, append([]byte(nil), data...), 0o600); err != nil {
		return fmt.Errorf("system: %s: write icon data: %w", CapabilityTray, err)
	}
	h.iconDataPath = path
	return nil
}

func (h *darwinTrayHandle) listen() {
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-h.stopListener:
			return
		case <-ticker.C:
			h.pollEvents()
		}
	}
}

func (h *darwinTrayHandle) pollEvents() {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	path := h.eventPath
	offset := h.eventOffset
	remainder := h.eventRemainder
	h.mu.Unlock()

	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return
	}
	data, err := io.ReadAll(file)
	if err != nil || len(data) == 0 {
		return
	}
	text := remainder + string(data)
	parts := strings.Split(text, "\n")
	nextRemainder := parts[len(parts)-1]
	lines := parts[:len(parts)-1]

	h.mu.Lock()
	h.eventOffset = offset + int64(len(data))
	h.eventRemainder = nextRemainder
	h.mu.Unlock()

	for _, line := range lines {
		h.dispatchLine(line)
	}
}

func (h *darwinTrayHandle) dispatchLine(line string) {
	line = strings.TrimSpace(line)
	switch {
	case line == "click":
		h.mu.Lock()
		onClick := h.opts.onClick
		h.mu.Unlock()
		if onClick != nil {
			onClick(TrayEvent{Kind: TrayEventClicked})
		}
	case strings.HasPrefix(line, "menu:"):
		h.dispatchMenuItem(strings.TrimPrefix(line, "menu:"))
	}
}

func (h *darwinTrayHandle) dispatchMenuItem(id string) {
	h.mu.Lock()
	menu := cloneTrayMenu(h.currentMenuLocked())
	h.mu.Unlock()
	if item, ok := findTrayMenuItem(menu, id); ok && item.OnClick != nil && !item.Disabled {
		item.OnClick(TrayEvent{Kind: TrayEventMenuItem, ItemID: id})
	}
}

func (h *darwinTrayHandle) currentMenuLocked() TrayMenu {
	if h.opts.menuProvider != nil {
		return cloneTrayMenu(h.opts.menuProvider())
	}
	return cloneTrayMenu(h.opts.menu)
}

func darwinTrayScript(opts trayOptions, eventPath string, iconPath string) string {
	tooltip := strings.TrimSpace(opts.tooltip)
	if tooltip == "" {
		tooltip = "FluxUI"
	}
	menu := opts.menu
	if opts.menuProvider != nil {
		menu = opts.menuProvider()
	}

	var b strings.Builder
	b.WriteString("use framework \"AppKit\"\n")
	b.WriteString("use scripting additions\n\n")
	b.WriteString("property statusItem : missing value\n")
	b.WriteString("property eventPath : ")
	b.WriteString(darwinAppleScriptQuote(eventPath))
	b.WriteString("\n\n")
	b.WriteString("on run\n")
	b.WriteString("\tset appRef to current application's NSApplication's sharedApplication()\n")
	b.WriteString("\tappRef's setActivationPolicy:(current application's NSApplicationActivationPolicyAccessory)\n")
	b.WriteString("\tset statusItem to current application's NSStatusBar's systemStatusBar()'s statusItemWithLength:(current application's NSVariableStatusItemLength)\n")
	b.WriteString("\tset buttonRef to statusItem's button()\n")
	b.WriteString("\tbuttonRef's setToolTip:")
	b.WriteString(darwinAppleScriptQuote(darwinTrayText(tooltip)))
	b.WriteString("\n")
	if iconPath != "" {
		b.WriteString("\tset imageRef to current application's NSImage's alloc()'s initWithContentsOfFile:")
		b.WriteString(darwinAppleScriptQuote(iconPath))
		b.WriteString("\n")
		b.WriteString("\tif imageRef is not missing value then\n")
		b.WriteString("\t\timageRef's setTemplate:true\n")
		b.WriteString("\t\tbuttonRef's setImage:imageRef\n")
		b.WriteString("\telse\n")
		b.WriteString("\t\tbuttonRef's setTitle:\"FluxUI\"\n")
		b.WriteString("\tend if\n")
	} else {
		b.WriteString("\tbuttonRef's setTitle:\"FluxUI\"\n")
	}
	if len(menu) > 0 {
		b.WriteString("\tset menuRef to current application's NSMenu's alloc()'s initWithTitle:\"FluxUI\"\n")
		b.WriteString("\tmenuRef's setAutoenablesItems:false\n")
		next := 1
		darwinTrayAppendMenuScript(&b, "menuRef", menu, &next)
		b.WriteString("\tstatusItem's setMenu:menuRef\n")
	} else if opts.onClick != nil {
		b.WriteString("\tbuttonRef's setTarget:me\n")
		b.WriteString("\tbuttonRef's setAction:\"statusClicked:\"\n")
	}
	b.WriteString("\tappRef's run()\n")
	b.WriteString("end run\n\n")
	b.WriteString("on statusClicked:sender\n")
	b.WriteString("\tmy writeEvent(\"click\")\n")
	b.WriteString("end statusClicked:\n\n")
	b.WriteString("on menuClicked:sender\n")
	b.WriteString("\tset itemID to (sender's representedObject()) as text\n")
	b.WriteString("\tmy writeEvent(\"menu:\" & itemID)\n")
	b.WriteString("end menuClicked:\n\n")
	b.WriteString("on writeEvent(theEvent)\n")
	b.WriteString("\tdo shell script \"printf \" & quoted form of \"%s\" & \" \" & quoted form of (theEvent & linefeed) & \" >> \" & quoted form of eventPath\n")
	b.WriteString("end writeEvent\n")
	return b.String()
}

func darwinTrayAppendMenuScript(b *strings.Builder, menuVar string, menu TrayMenu, next *int) {
	for _, item := range menu {
		if item.Separator {
			b.WriteString("\t")
			b.WriteString(menuVar)
			b.WriteString("'s addItem:(current application's NSMenuItem's separatorItem())\n")
			continue
		}

		index := *next
		*next = *next + 1
		itemVar := fmt.Sprintf("menuItem%d", index)
		label := darwinTrayMenuLabel(item, index)

		if len(item.Children) > 0 {
			submenuVar := fmt.Sprintf("submenu%d", index)
			b.WriteString("\tset ")
			b.WriteString(itemVar)
			b.WriteString(" to current application's NSMenuItem's alloc()'s initWithTitle:")
			b.WriteString(darwinAppleScriptQuote(label))
			b.WriteString(" action:(missing value) keyEquivalent:\"\"\n")
			if item.Disabled {
				b.WriteString("\t")
				b.WriteString(itemVar)
				b.WriteString("'s setEnabled:false\n")
			}
			b.WriteString("\tset ")
			b.WriteString(submenuVar)
			b.WriteString(" to current application's NSMenu's alloc()'s initWithTitle:")
			b.WriteString(darwinAppleScriptQuote(label))
			b.WriteString("\n")
			b.WriteString("\t")
			b.WriteString(submenuVar)
			b.WriteString("'s setAutoenablesItems:false\n")
			darwinTrayAppendMenuScript(b, submenuVar, item.Children, next)
			b.WriteString("\t")
			b.WriteString(itemVar)
			b.WriteString("'s setSubmenu:")
			b.WriteString(submenuVar)
			b.WriteString("\n")
			b.WriteString("\t")
			b.WriteString(menuVar)
			b.WriteString("'s addItem:")
			b.WriteString(itemVar)
			b.WriteString("\n")
			continue
		}

		b.WriteString("\tset ")
		b.WriteString(itemVar)
		b.WriteString(" to current application's NSMenuItem's alloc()'s initWithTitle:")
		b.WriteString(darwinAppleScriptQuote(label))
		b.WriteString(" action:\"menuClicked:\" keyEquivalent:\"\"\n")
		b.WriteString("\t")
		b.WriteString(itemVar)
		b.WriteString("'s setTarget:me\n")
		b.WriteString("\t")
		b.WriteString(itemVar)
		b.WriteString("'s setRepresentedObject:")
		b.WriteString(darwinAppleScriptQuote(item.ID))
		b.WriteString("\n")
		if item.Disabled {
			b.WriteString("\t")
			b.WriteString(itemVar)
			b.WriteString("'s setEnabled:false\n")
		}
		if item.Checked {
			b.WriteString("\t")
			b.WriteString(itemVar)
			b.WriteString("'s setState:1\n")
		}
		b.WriteString("\t")
		b.WriteString(menuVar)
		b.WriteString("'s addItem:")
		b.WriteString(itemVar)
		b.WriteString("\n")
	}
}

func darwinTrayMenuLabel(item TrayMenuItem, fallback int) string {
	label := strings.TrimSpace(item.Label)
	if label == "" {
		label = strings.TrimSpace(item.ID)
	}
	if label == "" {
		label = fmt.Sprintf("Item %d", fallback)
	}
	return darwinTrayText(label)
}

func darwinTrayText(value string) string {
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.TrimSpace(value)
	if value == "" {
		return "FluxUI"
	}
	return value
}
