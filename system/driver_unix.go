//go:build darwin || linux

package system

import (
	"context"
	"fmt"
	"os/exec"
)

type unixDriver struct{}

func newPlatformDriver() driver {
	return unixDriver{}
}

func (unixDriver) capabilities() CapabilitySet {
	return unixPlatformCapabilities()
}

func (d unixDriver) probeCapability(cap Capability) CapabilityAvailability {
	caps := d.capabilities()
	if !caps.Supports(cap) {
		return defaultCapabilityAvailability(caps, cap)
	}

	switch cap {
	case CapabilityClipboard:
		return unixProbeClipboardCapability(cap)
	case CapabilityFileDialog:
		return unixProbeFileDialogCapability(cap)
	case CapabilityMessageBox:
		return unixProbeMessageBoxCapability(cap)
	case CapabilityNotification:
		return unixProbeNotificationCapability(cap)
	case CapabilitySystemEvents:
		return unixProbeSystemEventsCapability(cap)
	case CapabilityTray:
		return unixProbeTrayCapability(cap)
	case CapabilityShell:
		return unixProbeShellCapability(cap)
	case CapabilitySingleInstance:
		return CapabilityAvailability{
			Capability: cap,
			Status:     CapabilityStatusAvailable,
		}
	case CapabilityDragAndDrop:
		return CapabilityAvailability{
			Capability: cap,
			Status:     CapabilityStatusAvailable,
		}
	default:
		return defaultCapabilityAvailability(caps, cap)
	}
}

func (unixDriver) acquireSingleInstance(ctx context.Context, opts singleInstanceOptions) (singleInstanceHandle, error) {
	return acquireLoopbackSingleInstance(ctx, opts)
}

func (unixDriver) probeDragAndDrop(ctx context.Context) DragAndDropProbe {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return defaultDragAndDropProbe(CapabilityStatusUnavailable, err)
	}
	return defaultDragAndDropProbe(CapabilityStatusAvailable, nil)
}

func unixProbeClipboardCapability(cap Capability) CapabilityAvailability {
	if _, err := selectUnixClipboardProvider(); err != nil {
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

func unixProbeFileDialogCapability(cap Capability) CapabilityAvailability {
	if _, err := selectUnixFileDialogProvider(); err != nil {
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

func unixProbeShellCapability(cap Capability) CapabilityAvailability {
	if _, err := exec.LookPath(unixShellOpenCommand()); err != nil {
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

func unixProbeMessageBoxCapability(cap Capability) CapabilityAvailability {
	if _, err := selectUnixMessageBoxProvider(); err != nil {
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

func unixProbeNotificationCapability(cap Capability) CapabilityAvailability {
	if _, err := selectUnixNotificationProvider(); err != nil {
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

func unixProbeSystemEventsCapability(cap Capability) CapabilityAvailability {
	if _, err := selectUnixSystemEventSources(unixSystemEventSources()); err != nil {
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
