//go:build darwin

package system

import (
	"fmt"
	"strings"
)

func unixNotificationProviders() []unixNotificationProvider {
	return []unixNotificationProvider{
		{
			name:     "osascript",
			commands: []string{"osascript"},
			build:    darwinNotificationCommand,
		},
	}
}

func darwinNotificationCommand(opts notificationOptions, _ string) (unixNotificationCommand, error) {
	title := opts.title
	if title == "" {
		title = "FluxUI"
	}

	parts := []string{
		"display notification " + darwinAppleScriptQuote(opts.body),
		"with title " + darwinAppleScriptQuote(title),
	}
	if opts.kind != "" && opts.kind != NotificationInfo {
		parts = append(parts, "subtitle "+darwinAppleScriptQuote(string(opts.kind)))
	}

	return unixNotificationCommand{
		name: "osascript",
		args: []string{"-e", strings.Join(parts, " ")},
		parse: func(output []byte, err error) (unixNotificationResult, error) {
			if err != nil {
				return unixNotificationResult{}, unixNotificationCommandError("osascript", err, output)
			}
			return unixNotificationResult{}, nil
		},
	}, nil
}

func darwinNotificationCancelCommand(string) (unixNotificationCommand, error) {
	return unixNotificationCommand{}, fmt.Errorf("system: %s: notification group cancel is unsupported on macOS command driver: %w", CapabilityNotification, ErrUnsupported)
}
