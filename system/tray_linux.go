//go:build linux

package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

const linuxTrayProviderCommand = "yad"

type linuxTrayHandle struct {
	mu             sync.Mutex
	opts           trayOptions
	tempDir        string
	fifoPath       string
	fifo           *os.File
	stopListener   chan struct{}
	cmd            *exec.Cmd
	iconDataPath   string
	scriptSequence int
	closed         bool
	visible        bool
}

func unixPlatformCapabilities() CapabilitySet {
	return CapabilitySet{
		CapabilityClipboard:      true,
		CapabilityFileDialog:     true,
		CapabilityMessageBox:     true,
		CapabilityNotification:   true,
		CapabilitySystemEvents:   true,
		CapabilityTray:           true,
		CapabilityShell:          true,
		CapabilitySingleInstance: true,
		CapabilityDragAndDrop:    true,
	}
}

func unixProbeTrayCapability(cap Capability) CapabilityAvailability {
	if _, err := exec.LookPath(linuxTrayProviderCommand); err != nil {
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
	if _, err := exec.LookPath(linuxTrayProviderCommand); err != nil {
		return nil, fmt.Errorf("system: %s: %s unavailable: %w: %v", CapabilityTray, linuxTrayProviderCommand, ErrUnavailable, err)
	}
	if opts.iconResource != 0 {
		return nil, fmt.Errorf("system: %s: resource icons are unsupported by linux tray command driver: %w", CapabilityTray, ErrUnsupported)
	}

	tempDir, err := os.MkdirTemp("", "fluxui-tray-*")
	if err != nil {
		return nil, fmt.Errorf("system: %s: create temp dir: %w", CapabilityTray, err)
	}
	handle := &linuxTrayHandle{
		opts:         opts,
		tempDir:      tempDir,
		fifoPath:     filepath.Join(tempDir, "events.fifo"),
		stopListener: make(chan struct{}),
	}
	if err := syscall.Mkfifo(handle.fifoPath, 0o600); err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, fmt.Errorf("system: %s: create event fifo: %w", CapabilityTray, err)
	}
	if len(opts.iconData) > 0 {
		if err := handle.writeIconDataLocked(opts.iconData); err != nil {
			_ = os.RemoveAll(tempDir)
			return nil, err
		}
	}
	if err := handle.startListenerLocked(); err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, err
	}
	return handle, nil
}

func (h *linuxTrayHandle) setIcon(path string) error {
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

func (h *linuxTrayHandle) setIconData(data []byte) error {
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

func (h *linuxTrayHandle) setIconResource(id uint16) error {
	return fmt.Errorf("system: %s: resource icons are unsupported by linux tray command driver: %w", CapabilityTray, ErrUnsupported)
}

func (h *linuxTrayHandle) setTooltip(text string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureOpenLocked(); err != nil {
		return err
	}
	h.opts.tooltip = text
	return h.restartIfVisibleLocked()
}

func (h *linuxTrayHandle) setMenu(menu TrayMenu) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureOpenLocked(); err != nil {
		return err
	}
	h.opts.menu = cloneTrayMenu(menu)
	h.opts.menuProvider = nil
	return h.restartIfVisibleLocked()
}

func (h *linuxTrayHandle) setMenuProvider(fn func() TrayMenu) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureOpenLocked(); err != nil {
		return err
	}
	h.opts.menuProvider = fn
	return h.restartIfVisibleLocked()
}

func (h *linuxTrayHandle) show() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureOpenLocked(); err != nil {
		return err
	}
	if h.visible && h.cmd != nil {
		return nil
	}
	if err := h.startLocked(); err != nil {
		return err
	}
	h.visible = true
	return nil
}

func (h *linuxTrayHandle) hide() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureOpenLocked(); err != nil {
		return err
	}
	h.stopProcessLocked()
	h.visible = false
	return nil
}

func (h *linuxTrayHandle) close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return trayClosedError()
	}
	h.closed = true
	h.visible = false
	h.stopProcessLocked()
	close(h.stopListener)
	if h.fifo != nil {
		_ = h.fifo.Close()
		h.fifo = nil
	}
	return os.RemoveAll(h.tempDir)
}

func (h *linuxTrayHandle) ensureOpenLocked() error {
	if h == nil || h.closed {
		return trayClosedError()
	}
	return nil
}

func (h *linuxTrayHandle) restartIfVisibleLocked() error {
	if !h.visible {
		return nil
	}
	h.stopProcessLocked()
	return h.startLocked()
}

func (h *linuxTrayHandle) startLocked() error {
	args, err := h.commandArgsLocked()
	if err != nil {
		return err
	}
	cmd := exec.Command(linuxTrayProviderCommand, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("system: %s: start %s: %w: %v", CapabilityTray, linuxTrayProviderCommand, ErrUnavailable, err)
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case err := <-done:
		return fmt.Errorf("system: %s: %s exited during startup: %w: %v", CapabilityTray, linuxTrayProviderCommand, ErrUnavailable, err)
	case <-time.After(150 * time.Millisecond):
		h.cmd = cmd
	}
	return nil
}

func (h *linuxTrayHandle) stopProcessLocked() {
	if h.cmd == nil || h.cmd.Process == nil {
		h.cmd = nil
		return
	}
	_ = h.cmd.Process.Kill()
	h.cmd = nil
}

func (h *linuxTrayHandle) startListenerLocked() error {
	fifo, err := os.OpenFile(h.fifoPath, os.O_RDWR, 0o600)
	if err != nil {
		return fmt.Errorf("system: %s: open event fifo: %w", CapabilityTray, err)
	}
	h.fifo = fifo
	go h.listen(fifo)
	return nil
}

func (h *linuxTrayHandle) listen(fifo *os.File) {
	scanner := bufio.NewScanner(fifo)
	for scanner.Scan() {
		select {
		case <-h.stopListener:
			return
		default:
		}
		h.dispatchLine(scanner.Text())
	}
}

func (h *linuxTrayHandle) dispatchLine(line string) {
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
		id := strings.TrimPrefix(line, "menu:")
		h.dispatchMenuItem(id)
	}
}

func (h *linuxTrayHandle) dispatchMenuItem(id string) {
	h.mu.Lock()
	menu := cloneTrayMenu(h.currentMenuLocked())
	h.mu.Unlock()
	if item, ok := findTrayMenuItem(menu, id); ok && item.OnClick != nil && !item.Disabled {
		item.OnClick(TrayEvent{Kind: TrayEventMenuItem, ItemID: id})
	}
}

func (h *linuxTrayHandle) currentMenuLocked() TrayMenu {
	if h.opts.menuProvider != nil {
		return cloneTrayMenu(h.opts.menuProvider())
	}
	return cloneTrayMenu(h.opts.menu)
}

func (h *linuxTrayHandle) commandArgsLocked() ([]string, error) {
	args := []string{"--notification"}
	args = append(args, "--image="+h.iconPathLocked())
	if h.opts.tooltip != "" {
		args = append(args, "--text="+h.opts.tooltip)
	} else {
		args = append(args, "--text=FluxUI")
	}
	clickScript, err := h.writeEventScriptLocked("click", "click")
	if err != nil {
		return nil, err
	}
	args = append(args, "--command="+clickScript)
	if menu := h.currentMenuLocked(); len(menu) > 0 {
		menuSpec, err := h.menuSpecLocked(menu)
		if err != nil {
			return nil, err
		}
		if menuSpec != "" {
			args = append(args, "--menu="+menuSpec)
		}
	}
	return args, nil
}

func (h *linuxTrayHandle) iconPathLocked() string {
	if h.iconDataPath != "" {
		return h.iconDataPath
	}
	if h.opts.icon != "" {
		return h.opts.icon
	}
	return "application-x-executable"
}

func (h *linuxTrayHandle) menuSpecLocked(menu TrayMenu) (string, error) {
	entries := make([]string, 0, len(menu))
	if err := h.appendMenuEntriesLocked(&entries, "", menu); err != nil {
		return "", err
	}
	return strings.Join(entries, "|"), nil
}

func (h *linuxTrayHandle) appendMenuEntriesLocked(entries *[]string, prefix string, menu TrayMenu) error {
	for _, item := range menu {
		if item.Separator {
			continue
		}
		label := strings.TrimSpace(item.Label)
		if label == "" {
			label = strings.TrimSpace(item.ID)
		}
		if label == "" {
			continue
		}
		if prefix != "" {
			label = prefix + " / " + label
		}
		if len(item.Children) > 0 {
			if err := h.appendMenuEntriesLocked(entries, label, item.Children); err != nil {
				return err
			}
			continue
		}
		command := "true"
		if item.ID != "" && !item.Disabled {
			script, err := h.writeEventScriptLocked("menu_"+item.ID, "menu:"+item.ID)
			if err != nil {
				return err
			}
			command = script
		}
		*entries = append(*entries, linuxTrayMenuEscape(label)+"!"+command)
	}
	return nil
}

func (h *linuxTrayHandle) writeEventScriptLocked(name, event string) (string, error) {
	h.scriptSequence++
	safeName := linuxTraySafeFileName(name)
	path := filepath.Join(h.tempDir, fmt.Sprintf("%03d-%s.sh", h.scriptSequence, safeName))
	content := "#!/bin/sh\nprintf '%s\\n' " + linuxTrayShellQuote(event) + " > " + linuxTrayShellQuote(h.fifoPath) + "\n"
	if err := os.WriteFile(path, []byte(content), 0o700); err != nil {
		return "", fmt.Errorf("system: %s: write event script: %w", CapabilityTray, err)
	}
	return path, nil
}

func (h *linuxTrayHandle) writeIconDataLocked(data []byte) error {
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

func linuxTraySafeFileName(value string) string {
	if value == "" {
		return "event"
	}
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "event"
	}
	return b.String()
}

func linuxTrayMenuEscape(value string) string {
	value = strings.ReplaceAll(value, "|", "/")
	value = strings.ReplaceAll(value, "!", "-")
	return value
}

func linuxTrayShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
