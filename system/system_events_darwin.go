//go:build darwin

package system

import (
	"context"
	"strings"
)

func unixSystemEventSources() []unixSystemEventSource {
	return []unixSystemEventSource{
		{
			kind:  SystemEventDisplayChanged,
			name:  "ioreg.display",
			probe: unixSystemEventCommandProbe("ioreg"),
			read:  darwinDisplaySystemEventSnapshot,
		},
		{
			kind:  SystemEventDPIChanged,
			name:  "ioreg.display",
			probe: unixSystemEventCommandProbe("ioreg"),
			read:  darwinDisplaySystemEventSnapshot,
		},
		{
			kind:  SystemEventThemeChanged,
			name:  "defaults.theme",
			probe: unixSystemEventCommandProbe("defaults"),
			read:  darwinThemeSystemEventSnapshot,
		},
		{
			kind:  SystemEventSettingsChanged,
			name:  "defaults.settings",
			probe: unixSystemEventCommandProbe("defaults"),
			read:  darwinSettingsSystemEventSnapshot,
		},
		{
			kind:  SystemEventPowerChanged,
			name:  "pmset.power",
			probe: unixSystemEventCommandProbe("pmset"),
			read:  darwinPowerSystemEventSnapshot,
		},
		{
			kind:  SystemEventSessionChanged,
			name:  "stat.console",
			probe: unixSystemEventCommandProbe("stat"),
			read:  darwinSessionSystemEventSnapshot,
		},
	}
}

func darwinDisplaySystemEventSnapshot(ctx context.Context) (string, error) {
	return runUnixSystemEventCommand(ctx, "ioreg", "-r", "-c", "IODisplayConnect", "-k", "IODisplayPrefsKey")
}

func darwinThemeSystemEventSnapshot(ctx context.Context) (string, error) {
	value, err := runUnixSystemEventCommand(ctx, "defaults", "read", "-g", "AppleInterfaceStyle")
	if err != nil {
		if ctx != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return "", ctxErr
			}
		}
		return "light", nil
	}
	if strings.EqualFold(strings.TrimSpace(value), "Dark") {
		return "dark", nil
	}
	return "light", nil
}

func darwinSettingsSystemEventSnapshot(ctx context.Context) (string, error) {
	parts := make([]string, 0, 2)
	for _, key := range []string{"AppleAccentColor", "AppleHighlightColor"} {
		value, err := runUnixSystemEventCommand(ctx, "defaults", "read", "-g", key)
		if err != nil {
			if ctx != nil {
				if ctxErr := ctx.Err(); ctxErr != nil {
					return "", ctxErr
				}
			}
			value = ""
		}
		parts = append(parts, key+"="+strings.TrimSpace(value))
	}
	return strings.Join(parts, "\n"), nil
}

func darwinPowerSystemEventSnapshot(ctx context.Context) (string, error) {
	return runUnixSystemEventCommand(ctx, "pmset", "-g", "ps")
}

func darwinSessionSystemEventSnapshot(ctx context.Context) (string, error) {
	return runUnixSystemEventCommand(ctx, "stat", "-f", "%Su", "/dev/console")
}
