//go:build linux

package system

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func unixSystemEventSources() []unixSystemEventSource {
	return []unixSystemEventSource{
		{
			kind:  SystemEventDisplayChanged,
			name:  "linux.display",
			probe: linuxDisplaySystemEventProbe,
			read:  linuxDisplaySystemEventSnapshot,
		},
		{
			kind:  SystemEventDPIChanged,
			name:  "linux.display",
			probe: linuxDisplaySystemEventProbe,
			read:  linuxDisplaySystemEventSnapshot,
		},
		{
			kind:  SystemEventThemeChanged,
			name:  "gsettings.theme",
			probe: unixSystemEventCommandProbe("gsettings"),
			read:  linuxThemeSystemEventSnapshot,
		},
		{
			kind:  SystemEventSettingsChanged,
			name:  "gsettings.settings",
			probe: unixSystemEventCommandProbe("gsettings"),
			read:  linuxSettingsSystemEventSnapshot,
		},
		{
			kind:  SystemEventPowerChanged,
			name:  "linux.power",
			probe: linuxPowerSystemEventProbe,
			read:  linuxPowerSystemEventSnapshot,
		},
		{
			kind:  SystemEventSessionChanged,
			name:  "loginctl.session",
			probe: linuxSessionSystemEventProbe,
			read:  linuxSessionSystemEventSnapshot,
		},
	}
}

func linuxDisplaySystemEventProbe() error {
	if linuxDirectoryHasEntries("/sys/class/drm") {
		return nil
	}
	if _, err := exec.LookPath("xrandr"); err == nil {
		return nil
	}
	if _, err := exec.LookPath("wlr-randr"); err == nil {
		return nil
	}
	return fmt.Errorf("display event sources unavailable")
}

func linuxDisplaySystemEventSnapshot(ctx context.Context) (string, error) {
	if linuxDirectoryHasEntries("/sys/class/drm") {
		if snapshot, err := linuxDirectorySnapshot("/sys/class/drm", []string{"status", "enabled", "modes", "dpms"}); err == nil {
			return snapshot, nil
		}
	}
	if _, err := exec.LookPath("xrandr"); err == nil {
		return runUnixSystemEventCommand(ctx, "xrandr", "--query")
	}
	return runUnixSystemEventCommand(ctx, "wlr-randr")
}

func linuxThemeSystemEventSnapshot(ctx context.Context) (string, error) {
	return linuxGSettingsSnapshot(ctx, []string{
		"org.gnome.desktop.interface color-scheme",
		"org.gnome.desktop.interface gtk-theme",
	})
}

func linuxSettingsSystemEventSnapshot(ctx context.Context) (string, error) {
	return linuxGSettingsSnapshot(ctx, []string{
		"org.gnome.desktop.interface icon-theme",
		"org.gnome.desktop.interface cursor-theme",
		"org.gnome.desktop.interface font-name",
	})
}

func linuxPowerSystemEventProbe() error {
	if linuxDirectoryHasEntries("/sys/class/power_supply") {
		return nil
	}
	if _, err := exec.LookPath("upower"); err == nil {
		return nil
	}
	return fmt.Errorf("power event sources unavailable")
}

func linuxPowerSystemEventSnapshot(ctx context.Context) (string, error) {
	if linuxDirectoryHasEntries("/sys/class/power_supply") {
		if snapshot, err := linuxDirectorySnapshot("/sys/class/power_supply", []string{"type", "status", "capacity", "online"}); err == nil {
			return snapshot, nil
		}
	}
	return runUnixSystemEventCommand(ctx, "upower", "-d")
}

func linuxSessionSystemEventProbe() error {
	if strings.TrimSpace(os.Getenv("XDG_SESSION_ID")) == "" {
		return fmt.Errorf("XDG_SESSION_ID is empty")
	}
	_, err := exec.LookPath("loginctl")
	return err
}

func linuxSessionSystemEventSnapshot(ctx context.Context) (string, error) {
	sessionID := strings.TrimSpace(os.Getenv("XDG_SESSION_ID"))
	if sessionID == "" {
		return "", fmt.Errorf("XDG_SESSION_ID is empty")
	}
	return runUnixSystemEventCommand(ctx, "loginctl", "show-session", sessionID, "--property=Active", "--property=State", "--property=Type", "--property=Desktop")
}

func linuxGSettingsSnapshot(ctx context.Context, keys []string) (string, error) {
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		args := strings.Split(key, " ")
		value, err := runUnixSystemEventCommand(ctx, "gsettings", append([]string{"get"}, args...)...)
		if err != nil {
			if ctx != nil {
				if ctxErr := ctx.Err(); ctxErr != nil {
					return "", ctxErr
				}
			}
			continue
		}
		parts = append(parts, key+"="+strings.TrimSpace(value))
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("gsettings keys unavailable")
	}
	return strings.Join(parts, "\n"), nil
}

func linuxDirectoryHasEntries(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.Name() != "" {
			return true
		}
	}
	return false
}

func linuxDirectorySnapshot(root string, fields []string) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	lines := make([]string, 0, len(entries)*len(fields))
	for _, entry := range entries {
		if entry.Name() == "" {
			continue
		}
		name := entry.Name()
		for _, field := range fields {
			value, err := os.ReadFile(filepath.Join(root, name, field))
			if err != nil {
				continue
			}
			lines = append(lines, name+"."+field+"="+strings.TrimSpace(string(value)))
		}
	}
	if len(lines) == 0 {
		return "", fmt.Errorf("%s has no readable state fields", root)
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n"), nil
}
