//go:build windows

package system

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const swShownormal = 1

const (
	windowsShellExecuteErrorFileNotFound  = uintptr(2)
	windowsShellExecuteErrorPathNotFound  = uintptr(3)
	windowsShellExecuteErrorAccessDenied  = uintptr(5)
	windowsShellExecuteErrorOutOfMemory   = uintptr(8)
	windowsShellExecuteErrorBadFormat     = uintptr(11)
	windowsShellExecuteErrorShare         = uintptr(26)
	windowsShellExecuteErrorAssociation   = uintptr(27)
	windowsShellExecuteErrorDDETimeout    = uintptr(28)
	windowsShellExecuteErrorDDEFail       = uintptr(29)
	windowsShellExecuteErrorDDEBusy       = uintptr(30)
	windowsShellExecuteErrorNoAssociation = uintptr(31)
	windowsShellExecuteErrorDLLNotFound   = uintptr(32)
)

var procShellExecute = shell32.NewProc("ShellExecuteW")

func (windowsDriver) openURL(ctx context.Context, target string) error {
	if err := checkShellContext(ctx); err != nil {
		return err
	}
	if path, ok, err := windowsShellResolveFileURL(target); ok || err != nil {
		if err != nil {
			return err
		}
		return windowsShellExecute("open", path, "", "")
	}
	return windowsShellExecute("open", target, "", "")
}

func (windowsDriver) openPath(ctx context.Context, path string) error {
	if err := checkShellContext(ctx); err != nil {
		return err
	}
	abs, err := windowsShellResolveExistingPath(path, "open")
	if err != nil {
		return err
	}
	return windowsShellExecute("open", abs, "", "")
}

func (windowsDriver) revealPath(ctx context.Context, path string) error {
	if err := checkShellContext(ctx); err != nil {
		return err
	}
	abs, err := windowsShellResolveExistingPath(path, "reveal")
	if err != nil {
		return err
	}
	return windowsShellExecute("open", "explorer.exe", "/select,"+windowsShellQuoteArgument(abs), "")
}

func checkShellContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func windowsShellResolveExistingPath(path, operation string) (string, error) {
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

func windowsShellResolveFileURL(target string) (string, bool, error) {
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
		path = `\\` + parsed.Host + filepath.FromSlash(path)
	} else if strings.HasPrefix(path, "/") && len(path) >= 3 && path[2] == ':' {
		path = path[1:]
	}
	path = filepath.FromSlash(path)
	resolved, err := windowsShellResolveExistingPath(path, "open URL")
	return resolved, true, err
}

func windowsShellExecute(verb, file, parameters, directory string) error {
	verb16, err := windows.UTF16PtrFromString(verb)
	if err != nil {
		return err
	}
	file16, err := windows.UTF16PtrFromString(file)
	if err != nil {
		return err
	}
	var parameters16 *uint16
	if parameters != "" {
		parameters16, err = windows.UTF16PtrFromString(parameters)
		if err != nil {
			return err
		}
	}
	var directory16 *uint16
	if directory != "" {
		directory16, err = windows.UTF16PtrFromString(directory)
		if err != nil {
			return err
		}
	}

	r, _, callErr := procShellExecute.Call(
		0,
		uintptr(unsafe.Pointer(verb16)),
		uintptr(unsafe.Pointer(file16)),
		uintptr(unsafe.Pointer(parameters16)),
		uintptr(unsafe.Pointer(directory16)),
		swShownormal,
	)
	if r > 32 {
		return nil
	}
	classified := windowsShellExecuteError(r)
	if callErr != syscall.Errno(0) {
		return fmt.Errorf("system: %s: shell execute %q: %w: %w: %v", CapabilityShell, file, ErrUnavailable, classified, callErr)
	}
	return fmt.Errorf("system: %s: shell execute %q failed with code %d: %w: %w", CapabilityShell, file, r, ErrUnavailable, classified)
}

func windowsShellQuoteArgument(value string) string {
	value = strings.ReplaceAll(value, `"`, `\"`)
	return `"` + value + `"`
}

func windowsShellExecuteError(code uintptr) error {
	switch code {
	case windowsShellExecuteErrorFileNotFound, windowsShellExecuteErrorPathNotFound:
		return ErrTargetNotFound
	case windowsShellExecuteErrorAccessDenied:
		return ErrAccessDenied
	case windowsShellExecuteErrorAssociation, windowsShellExecuteErrorNoAssociation:
		return ErrNoDefaultHandler
	case windowsShellExecuteErrorBadFormat:
		return ErrInvalidTarget
	case windowsShellExecuteErrorOutOfMemory,
		windowsShellExecuteErrorShare,
		windowsShellExecuteErrorDDETimeout,
		windowsShellExecuteErrorDDEFail,
		windowsShellExecuteErrorDDEBusy,
		windowsShellExecuteErrorDLLNotFound:
		return ErrUnavailable
	default:
		return ErrUnavailable
	}
}
