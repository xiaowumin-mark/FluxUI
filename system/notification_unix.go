//go:build darwin || linux

package system

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

type unixNotificationCommand struct {
	name  string
	args  []string
	parse func([]byte, error) (unixNotificationResult, error)
}

type unixNotificationResult struct {
	groupID string
}

type unixNotificationProvider struct {
	name             string
	commands         []string
	supportsGroup    bool
	supportsCancel   bool
	build            func(notificationOptions, string) (unixNotificationCommand, error)
	buildCancelGroup func(string) (unixNotificationCommand, error)
}

type unixNotificationGroupEntry struct {
	provider string
	id       string
}

var unixNotificationGroups = struct {
	sync.Mutex
	entries map[string]unixNotificationGroupEntry
}{
	entries: make(map[string]unixNotificationGroupEntry),
}

func (unixDriver) notify(ctx context.Context, opts notificationOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateUnixNotificationOptions(opts); err != nil {
		return err
	}

	provider, err := selectUnixNotificationProvider()
	if err != nil {
		return fmt.Errorf("system: %s: notify: %w: %v", CapabilityNotification, ErrUnavailable, err)
	}
	replaceID := ""
	if opts.group != "" {
		if !provider.supportsGroup {
			return fmt.Errorf("system: %s: notification groups are unsupported by %s: %w", CapabilityNotification, provider.name, ErrUnsupported)
		}
		replaceID = unixNotificationReplaceID(opts.group, provider.name)
	}

	command, err := provider.build(opts, replaceID)
	if err != nil {
		return err
	}
	result, err := runUnixNotification(ctx, provider, command)
	if err != nil {
		return err
	}
	if opts.group != "" && result.groupID != "" {
		unixNotificationRememberGroup(opts.group, provider.name, result.groupID)
	}
	return nil
}

func (unixDriver) cancelNotificationGroup(group string) error {
	group = strings.TrimSpace(group)
	if group == "" {
		return fmt.Errorf("system: notification group is empty")
	}

	entry, ok := unixNotificationForgetGroup(group)
	if !ok {
		return nil
	}
	provider, ok := unixNotificationProviderByName(entry.provider)
	if !ok || provider.buildCancelGroup == nil || !provider.supportsCancel {
		return fmt.Errorf("system: %s: cancel notification group with %s: %w", CapabilityNotification, entry.provider, ErrUnsupported)
	}
	command, err := provider.buildCancelGroup(entry.id)
	if err != nil {
		return err
	}
	_, err = runUnixNotification(context.Background(), provider, command)
	return err
}

func (unixDriver) probeNotificationBackend(ctx context.Context, backend NotificationBackend, opts notificationOptions) NotificationBackendProbe {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return NotificationBackendProbe{
			Backend: unixNotificationProbeBackend(backend),
			Status:  CapabilityStatusUnavailable,
			Err:     err,
		}
	}
	if backend == NotificationBackendToast {
		return NotificationBackendProbe{
			Backend: NotificationBackendToast,
			Status:  CapabilityStatusUnsupported,
			Err:     fmt.Errorf("system: %s: toast backend is unsupported on this platform: %w", CapabilityNotification, ErrUnsupported),
		}
	}
	if err := validateUnixNotificationOptions(opts); err != nil {
		status := CapabilityStatusUnavailable
		if IsUnsupported(err) {
			status = CapabilityStatusUnsupported
		}
		return NotificationBackendProbe{
			Backend: unixNotificationProbeBackend(backend),
			Status:  status,
			Err:     err,
		}
	}
	provider, err := selectUnixNotificationProvider()
	if err != nil {
		return NotificationBackendProbe{
			Backend: unixNotificationProbeBackend(backend),
			Status:  CapabilityStatusUnavailable,
			Err:     fmt.Errorf("system: probe %s backend: %w: %v", CapabilityNotification, ErrUnavailable, err),
		}
	}
	if opts.group != "" && !provider.supportsGroup {
		return NotificationBackendProbe{
			Backend: unixNotificationProbeBackend(backend),
			Status:  CapabilityStatusUnsupported,
			Err:     fmt.Errorf("system: %s: notification groups are unsupported by %s: %w", CapabilityNotification, provider.name, ErrUnsupported),
		}
	}
	return NotificationBackendProbe{
		Backend:                    unixNotificationProbeBackend(backend),
		Status:                     CapabilityStatusAvailable,
		SupportsClickCallback:      false,
		SupportsDismissCallback:    false,
		SupportsActionCallback:     false,
		SupportsActionButtons:      false,
		SupportsDurableActivation:  false,
		SupportsProtocolActivation: false,
		Err:                        nil,
	}
}

func runUnixNotification(ctx context.Context, provider unixNotificationProvider, command unixNotificationCommand) (unixNotificationResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return unixNotificationResult{}, err
	}
	if command.name == "" {
		return unixNotificationResult{}, fmt.Errorf("system: %s: %s command is empty: %w", CapabilityNotification, provider.name, ErrUnavailable)
	}

	cmd := exec.CommandContext(ctx, command.name, command.args...)
	output, err := cmd.CombinedOutput()
	if ctxErr := ctx.Err(); ctxErr != nil {
		return unixNotificationResult{}, ctxErr
	}
	if command.parse == nil {
		if err != nil {
			return unixNotificationResult{}, unixNotificationCommandError(provider.name, err, output)
		}
		return unixNotificationResult{}, nil
	}
	return command.parse(output, err)
}

func selectUnixNotificationProvider() (unixNotificationProvider, error) {
	return unixNotificationProviderWithLookPath(exec.LookPath)
}

func unixNotificationProviderWithLookPath(lookPath unixLookPathFunc) (unixNotificationProvider, error) {
	candidates := unixNotificationProviders()
	var missing []string
	for _, candidate := range candidates {
		ok := true
		for _, name := range candidate.commandNames() {
			if _, err := lookPath(name); err != nil {
				missing = append(missing, name)
				ok = false
				break
			}
		}
		if ok {
			return candidate, nil
		}
	}
	if len(missing) == 0 {
		return unixNotificationProvider{}, fmt.Errorf("no notification providers configured")
	}
	return unixNotificationProvider{}, fmt.Errorf("notification commands unavailable: %s", strings.Join(uniqueStrings(missing), ", "))
}

func unixNotificationProviderByName(name string) (unixNotificationProvider, bool) {
	for _, provider := range unixNotificationProviders() {
		if provider.name == name {
			return provider, true
		}
	}
	return unixNotificationProvider{}, false
}

func (p unixNotificationProvider) commandNames() []string {
	return p.commands
}

func validateUnixNotificationOptions(opts notificationOptions) error {
	switch opts.backend {
	case NotificationBackendAuto, NotificationBackendBalloon:
	case NotificationBackendToast:
		return fmt.Errorf("system: %s: toast backend is unsupported on this platform: %w", CapabilityNotification, ErrUnsupported)
	default:
		return fmt.Errorf("system: unknown notification backend %q", opts.backend)
	}
	switch opts.kind {
	case "", NotificationInfo, NotificationSuccess, NotificationWarning, NotificationError:
	default:
		return fmt.Errorf("system: unknown notification kind %q", opts.kind)
	}
	if len(opts.actions) > 0 {
		return fmt.Errorf("system: %s: notification actions are unsupported on this platform: %w", CapabilityNotification, ErrUnsupported)
	}
	if opts.launchURI != "" {
		return fmt.Errorf("system: %s: protocol activation is unsupported on this platform: %w", CapabilityNotification, ErrUnsupported)
	}
	return nil
}

func unixNotificationProbeBackend(backend NotificationBackend) NotificationBackend {
	if backend == NotificationBackendToast {
		return NotificationBackendToast
	}
	return NotificationBackendBalloon
}

func unixNotificationReplaceID(group string, provider string) string {
	unixNotificationGroups.Lock()
	entry := unixNotificationGroups.entries[group]
	unixNotificationGroups.Unlock()
	if entry.provider != provider {
		return ""
	}
	return entry.id
}

func unixNotificationRememberGroup(group, provider, id string) {
	if group == "" || provider == "" || id == "" {
		return
	}
	unixNotificationGroups.Lock()
	unixNotificationGroups.entries[group] = unixNotificationGroupEntry{provider: provider, id: id}
	unixNotificationGroups.Unlock()
}

func unixNotificationForgetGroup(group string) (unixNotificationGroupEntry, bool) {
	unixNotificationGroups.Lock()
	entry, ok := unixNotificationGroups.entries[group]
	delete(unixNotificationGroups.entries, group)
	unixNotificationGroups.Unlock()
	return entry, ok
}

func unixNotificationCommandError(provider string, err error, output []byte) error {
	detail := strings.TrimSpace(string(output))
	if detail != "" {
		return fmt.Errorf("system: %s: run %s: %w: %v: %s", CapabilityNotification, provider, ErrUnavailable, err, detail)
	}
	return fmt.Errorf("system: %s: run %s: %w: %v", CapabilityNotification, provider, ErrUnavailable, err)
}
