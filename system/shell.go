package system

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type shellDriver interface {
	openURL(ctx context.Context, target string) error
	openPath(ctx context.Context, path string) error
	revealPath(ctx context.Context, path string) error
}

// OpenURL opens a URL with the operating system default handler.
func OpenURL(ctx context.Context, target string) error {
	target = strings.TrimSpace(target)
	if target == "" {
		return fmt.Errorf("system: shell URL is empty: %w", ErrInvalidTarget)
	}
	parsed, err := url.Parse(target)
	if err != nil || parsed.Scheme == "" {
		if err == nil {
			err = fmt.Errorf("missing URL scheme")
		}
		return fmt.Errorf("system: shell URL %q is invalid: %w: %w", target, ErrInvalidTarget, err)
	}
	return withShellDriver(ctx, func(ctx context.Context, driver shellDriver) error {
		return driver.openURL(ctx, target)
	})
}

// OpenPath opens a file or directory with the operating system default handler.
func OpenPath(ctx context.Context, path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("system: shell path is empty: %w", ErrInvalidTarget)
	}
	return withShellDriver(ctx, func(ctx context.Context, driver shellDriver) error {
		return driver.openPath(ctx, path)
	})
}

// RevealPath reveals a file or directory in the operating system file manager.
func RevealPath(ctx context.Context, path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("system: shell path is empty: %w", ErrInvalidTarget)
	}
	return withShellDriver(ctx, func(ctx context.Context, driver shellDriver) error {
		return driver.revealPath(ctx, path)
	})
}

func withShellDriver(ctx context.Context, fn func(context.Context, shellDriver) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	sd, ok := d.(shellDriver)
	if !ok || d == nil || !d.capabilities().Supports(CapabilityShell) {
		return fmt.Errorf("system: %s: %w", CapabilityShell, ErrUnsupported)
	}
	return fn(ctx, sd)
}
