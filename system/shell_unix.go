//go:build darwin || linux

package system

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type unixShellCommand struct {
	name string
	args []string
}

func (unixDriver) openURL(ctx context.Context, target string) error {
	if path, ok, err := unixShellResolveFileURL(target); ok || err != nil {
		if err != nil {
			return err
		}
		return runUnixShellOpen(ctx, unixShellOpenPathCommand(path))
	}
	return runUnixShellOpen(ctx, unixShellOpenURLCommand(target))
}

func (unixDriver) openPath(ctx context.Context, path string) error {
	abs, err := unixShellResolveExistingPath(path, "open")
	if err != nil {
		return err
	}
	return runUnixShellOpen(ctx, unixShellOpenPathCommand(abs))
}

func (unixDriver) revealPath(ctx context.Context, path string) error {
	abs, err := unixShellResolveExistingPath(path, "reveal")
	if err != nil {
		return err
	}
	return runUnixShellOpen(ctx, unixShellRevealPathCommand(abs))
}

func runUnixShellOpen(ctx context.Context, command unixShellCommand) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if command.name == "" {
		return fmt.Errorf("system: %s: shell command is empty: %w", CapabilityShell, ErrUnavailable)
	}

	cmd := exec.CommandContext(ctx, command.name, command.args...)
	if err := cmd.Start(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return fmt.Errorf("system: %s: start %s: %w: %v", CapabilityShell, command.name, ErrUnavailable, err)
	}
	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("system: %s: release %s: %w: %v", CapabilityShell, command.name, ErrUnavailable, err)
	}
	return nil
}

func unixShellResolveExistingPath(path, operation string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("system: shell %s path %q is invalid: %w: %w", operation, path, ErrInvalidTarget, err)
	}
	if _, err := os.Stat(abs); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("system: shell %s path %q was not found: %w: %w: %w", operation, abs, ErrUnavailable, ErrTargetNotFound, err)
		}
		if os.IsPermission(err) {
			return "", fmt.Errorf("system: shell %s path %q is denied: %w: %w: %w", operation, abs, ErrUnavailable, ErrAccessDenied, err)
		}
		return "", fmt.Errorf("system: shell %s path %q is unavailable: %w: %w", operation, abs, ErrUnavailable, err)
	}
	return abs, nil
}

func unixShellResolveFileURL(target string) (string, bool, error) {
	parsed, err := url.Parse(target)
	if err != nil || !strings.EqualFold(parsed.Scheme, "file") {
		return "", false, nil
	}
	path := parsed.Opaque
	if path == "" {
		path, err = url.PathUnescape(parsed.EscapedPath())
		if err != nil {
			return "", true, fmt.Errorf("system: shell open URL %q is invalid: %w: %w", target, ErrInvalidTarget, err)
		}
	}
	if path == "" {
		return "", true, fmt.Errorf("system: shell open URL %q has no file path: %w", target, ErrInvalidTarget)
	}
	if parsed.Host != "" && !strings.EqualFold(parsed.Host, "localhost") {
		path = "//" + parsed.Host + path
	}
	resolved, err := unixShellResolveExistingPath(filepath.FromSlash(path), "open URL")
	return resolved, true, err
}
