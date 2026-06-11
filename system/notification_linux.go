//go:build linux

package system

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func unixNotificationProviders() []unixNotificationProvider {
	return []unixNotificationProvider{
		{
			name:             "notify-send",
			commands:         []string{"notify-send"},
			supportsGroup:    true,
			supportsCancel:   true,
			build:            linuxNotifySendCommand,
			buildCancelGroup: linuxNotifySendCancelCommand,
		},
		{
			name:     "kdialog",
			commands: []string{"kdialog"},
			build:    linuxKdialogNotificationCommand,
		},
	}
}

func linuxNotifySendCommand(opts notificationOptions, replaceID string) (unixNotificationCommand, error) {
	title := opts.title
	if title == "" {
		title = "FluxUI"
	}

	args := []string{"--app-name", "FluxUI"}
	if urgency := linuxNotifySendUrgency(opts.kind); urgency != "" {
		args = append(args, "--urgency", urgency)
	}
	if opts.icon != "" {
		args = append(args, "--icon", opts.icon)
	}
	if opts.timeout > 0 {
		args = append(args, "--expire-time", linuxNotifySendTimeout(opts.timeout))
	}
	if opts.group != "" {
		args = append(args, "--print-id")
		if replaceID != "" {
			args = append(args, "--replace-id", replaceID)
		}
	}
	args = append(args, title)
	if opts.body != "" {
		args = append(args, opts.body)
	}

	return unixNotificationCommand{
		name: "notify-send",
		args: args,
		parse: func(output []byte, err error) (unixNotificationResult, error) {
			return linuxNotifySendParseResult(output, err)
		},
	}, nil
}

func linuxNotifySendCancelCommand(id string) (unixNotificationCommand, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return unixNotificationCommand{}, fmt.Errorf("system: %s: notification group id is empty: %w", CapabilityNotification, ErrUnavailable)
	}
	return unixNotificationCommand{
		name: "gdbus",
		args: []string{
			"call",
			"--session",
			"--dest", "org.freedesktop.Notifications",
			"--object-path", "/org/freedesktop/Notifications",
			"--method", "org.freedesktop.Notifications.CloseNotification",
			"uint32", id,
		},
		parse: func(output []byte, err error) (unixNotificationResult, error) {
			if err != nil {
				return unixNotificationResult{}, unixNotificationCommandError("gdbus", err, output)
			}
			return unixNotificationResult{}, nil
		},
	}, nil
}

func linuxNotifySendUrgency(kind NotificationKind) string {
	switch kind {
	case NotificationError:
		return "critical"
	case NotificationWarning:
		return "normal"
	default:
		return "normal"
	}
}

func linuxNotifySendTimeout(timeout time.Duration) string {
	if timeout < time.Second {
		return "1000"
	}
	if timeout > time.Minute {
		timeout = time.Minute
	}
	return fmt.Sprintf("%d", timeout/time.Millisecond)
}

func linuxNotifySendParseResult(output []byte, err error) (unixNotificationResult, error) {
	if err != nil {
		return unixNotificationResult{}, unixNotificationCommandError("notify-send", err, output)
	}
	id := strings.TrimSpace(string(output))
	return unixNotificationResult{groupID: id}, nil
}

func linuxKdialogNotificationCommand(opts notificationOptions, _ string) (unixNotificationCommand, error) {
	if opts.group != "" {
		return unixNotificationCommand{}, fmt.Errorf("system: %s: notification groups are unsupported by kdialog: %w", CapabilityNotification, ErrUnsupported)
	}
	title := opts.title
	if title == "" {
		title = "FluxUI"
	}
	body := opts.body
	if body == "" {
		body = title
	}
	timeout := "10"
	if opts.timeout > 0 {
		seconds := int(opts.timeout / time.Second)
		if seconds < 1 {
			seconds = 1
		}
		if seconds > 60 {
			seconds = 60
		}
		timeout = fmt.Sprintf("%d", seconds)
	}

	return unixNotificationCommand{
		name: "kdialog",
		args: []string{"--title", title, "--passivepopup", body, timeout},
		parse: func(output []byte, err error) (unixNotificationResult, error) {
			if err != nil {
				return unixNotificationResult{}, unixNotificationCommandError("kdialog", err, output)
			}
			return unixNotificationResult{}, nil
		},
	}, nil
}

func linuxNotificationCloseCommandAvailable() bool {
	_, err := exec.LookPath("gdbus")
	return err == nil
}
