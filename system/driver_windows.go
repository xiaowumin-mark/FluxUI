//go:build windows

package system

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"golang.org/x/sys/windows"
)

type windowsDriver struct{}

func newPlatformDriver() driver {
	return windowsDriver{}
}

func (windowsDriver) capabilities() CapabilitySet {
	return CapabilitySet{
		CapabilityWindow:             true,
		CapabilityFileDialog:         true,
		CapabilityMessageBox:         true,
		CapabilityNotification:       true,
		CapabilityTray:               true,
		CapabilitySystemEvents:       true,
		CapabilityClipboard:          true,
		CapabilityShell:              true,
		CapabilitySingleInstance:     true,
		CapabilitySystemRegistration: true,
		CapabilityGlobalShortcut:     true,
		CapabilityDragAndDrop:        true,
	}
}

func (d windowsDriver) probeCapability(cap Capability) CapabilityAvailability {
	caps := d.capabilities()
	if !caps.Supports(cap) {
		return defaultCapabilityAvailability(caps, cap)
	}

	switch cap {
	case CapabilityWindow:
		return windowsCapabilityAvailable(cap)
	case CapabilityFileDialog:
		return windowsProbeFileDialogCapability(cap)
	case CapabilityMessageBox:
		return windowsProbeMessageBoxCapability(cap)
	case CapabilityNotification:
		return windowsProbeNotificationCapability(cap)
	case CapabilityTray:
		return windowsProbeTrayCapability(cap)
	case CapabilitySystemEvents:
		return windowsProbeSystemEventsCapability(cap)
	case CapabilityClipboard:
		return windowsProbeClipboardCapability(cap)
	case CapabilityShell:
		return windowsProbeShellCapability(cap)
	case CapabilitySingleInstance:
		return windowsCapabilityAvailable(cap)
	case CapabilitySystemRegistration:
		return windowsCapabilityAvailable(cap)
	case CapabilityGlobalShortcut:
		return windowsProbeGlobalShortcutCapability(cap)
	case CapabilityDragAndDrop:
		return windowsCapabilityAvailable(cap)
	default:
		return defaultCapabilityAvailability(caps, cap)
	}
}

func (windowsDriver) acquireSingleInstance(ctx context.Context, opts singleInstanceOptions) (singleInstanceHandle, error) {
	return acquireLoopbackSingleInstance(ctx, opts)
}

func (windowsDriver) probeDragAndDrop(ctx context.Context) DragAndDropProbe {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return defaultDragAndDropProbe(CapabilityStatusUnavailable, err)
	}
	return defaultDragAndDropProbe(CapabilityStatusAvailable, nil)
}

func windowsCapabilityAvailable(cap Capability) CapabilityAvailability {
	return CapabilityAvailability{
		Capability: cap,
		Status:     CapabilityStatusAvailable,
	}
}

func windowsCapabilityUnavailable(cap Capability, err error) CapabilityAvailability {
	if err == nil {
		err = ErrUnavailable
	}
	wrapped := err
	if !IsUnavailable(err) {
		wrapped = fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return CapabilityAvailability{
		Capability: cap,
		Status:     CapabilityStatusUnavailable,
		Err:        fmt.Errorf("system: probe %s: %w", cap, wrapped),
	}
}

func windowsProbeFileDialogCapability(cap Capability) CapabilityAvailability {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	initialized, err := initCOMApartment()
	if err != nil {
		return windowsCapabilityUnavailable(cap, err)
	}
	if initialized {
		defer windows.CoUninitialize()
	}

	dialog, err := newWindowsFileDialog(FileDialogOpenFile)
	if err != nil {
		return windowsCapabilityUnavailable(cap, err)
	}
	dialog.Release()
	return windowsCapabilityAvailable(cap)
}

func windowsProbeMessageBoxCapability(cap Capability) CapabilityAvailability {
	if err := procTaskDialogIndirect.Find(); err == nil {
		return windowsCapabilityAvailable(cap)
	}
	if err := procMessageBox.Find(); err == nil {
		return windowsCapabilityAvailable(cap)
	}
	return windowsCapabilityUnavailable(cap, ErrUnavailable)
}

func windowsProbeNotificationCapability(cap Capability) CapabilityAvailability {
	if err := procShellNotifyIcon.Find(); err != nil {
		return windowsCapabilityUnavailable(cap, err)
	}
	if err := procLoadIcon.Find(); err != nil {
		return windowsCapabilityUnavailable(cap, err)
	}
	if _, err := windowsNotificationState.window(); err != nil {
		return windowsCapabilityUnavailable(cap, err)
	}
	return windowsCapabilityAvailable(cap)
}

func (windowsDriver) probeNotificationBackend(ctx context.Context, backend NotificationBackend, opts notificationOptions) NotificationBackendProbe {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return windowsNotificationBackendUnavailable(backend, err)
	}
	if backend == NotificationBackendAuto {
		if len(opts.actions) > 0 {
			backend = NotificationBackendToast
		} else {
			backend = NotificationBackendBalloon
		}
	}

	switch backend {
	case NotificationBackendBalloon:
		availability := windowsProbeNotificationCapability(CapabilityNotification)
		if availability.Status != CapabilityStatusAvailable {
			return windowsNotificationBackendUnavailable(backend, availability.Err)
		}
		return NotificationBackendProbe{
			Backend:                 backend,
			Status:                  CapabilityStatusAvailable,
			SupportsClickCallback:   true,
			SupportsDismissCallback: true,
		}
	case NotificationBackendToast:
		if err := validateWindowsToastOptions(opts); err != nil {
			return windowsNotificationBackendUnavailable(backend, err)
		}
		if err := windowsProbeToastBackend(ctx, windowsToastAppID(opts.appID)); err != nil {
			return windowsNotificationBackendUnavailable(backend, err)
		}
		return NotificationBackendProbe{
			Backend:                    backend,
			Status:                     CapabilityStatusAvailable,
			SupportsActionButtons:      true,
			SupportsClickCallback:      true,
			SupportsDismissCallback:    true,
			SupportsActionCallback:     true,
			SupportsProtocolActivation: true,
			SupportsDurableActivation:  true,
		}
	default:
		return NotificationBackendProbe{
			Backend: backend,
			Status:  CapabilityStatusUnsupported,
			Err:     fmt.Errorf("system: unknown notification backend %q: %w", backend, ErrUnsupported),
		}
	}
}

func windowsNotificationBackendUnavailable(backend NotificationBackend, err error) NotificationBackendProbe {
	if err == nil {
		err = ErrUnavailable
	}
	if !IsUnavailable(err) && !IsUnsupported(err) {
		err = fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return NotificationBackendProbe{
		Backend: backend,
		Status:  CapabilityStatusUnavailable,
		Err:     fmt.Errorf("system: probe notification backend %s: %w", backend, err),
	}
}

func windowsProbeToastBackend(ctx context.Context, appID string) error {
	script := windowsToastProbePowerShellScript(appID)
	cmd := exec.CommandContext(ctx,
		"powershell.exe",
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy",
		"Bypass",
		"-EncodedCommand",
		powershellEncodedCommand(script),
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return fmt.Errorf("toast backend unavailable: %w: %s", ErrUnavailable, strings.TrimSpace(string(output)))
	}
	return nil
}

func windowsProbeTrayCapability(cap Capability) CapabilityAvailability {
	if err := procShellNotifyIcon.Find(); err != nil {
		return windowsCapabilityUnavailable(cap, err)
	}
	if err := procCreatePopupMenu.Find(); err != nil {
		return windowsCapabilityUnavailable(cap, err)
	}
	if _, err := windowsTrayState.window(); err != nil {
		return windowsCapabilityUnavailable(cap, err)
	}
	return windowsCapabilityAvailable(cap)
}

func windowsProbeSystemEventsCapability(cap Capability) CapabilityAvailability {
	if _, err := windowsSystemEventState.window(); err != nil {
		return windowsCapabilityUnavailable(cap, err)
	}
	return windowsCapabilityAvailable(cap)
}

func windowsProbeClipboardCapability(cap Capability) CapabilityAvailability {
	cmd := exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy",
		"Bypass",
		"-EncodedCommand",
		powershellEncodedCommand(windowsProbeClipboardPowerShellScript()),
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return windowsCapabilityUnavailable(cap, fmt.Errorf("clipboard PowerShell cmdlets unavailable: %w: %s", err, strings.TrimSpace(string(output))))
	}
	return windowsCapabilityAvailable(cap)
}

func windowsProbeShellCapability(cap Capability) CapabilityAvailability {
	if err := procShellExecute.Find(); err != nil {
		return windowsCapabilityUnavailable(cap, err)
	}
	return windowsCapabilityAvailable(cap)
}
