package system

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

type clipboardDriver interface {
	readClipboardText(ctx context.Context) (string, error)
	writeClipboardText(ctx context.Context, text string) error
}

type clipboardFilesDriver interface {
	readClipboardFiles(ctx context.Context) ([]string, error)
	writeClipboardFiles(ctx context.Context, paths []string) error
}

type clipboardImageDriver interface {
	readClipboardImagePNG(ctx context.Context) ([]byte, error)
	writeClipboardImagePNG(ctx context.Context, data []byte) error
}

// ReadClipboardText reads Unicode text from the system clipboard.
func ReadClipboardText(ctx context.Context) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	cd, ok := d.(clipboardDriver)
	if !ok || d == nil || !d.capabilities().Supports(CapabilityClipboard) {
		return "", fmt.Errorf("system: %s: %w", CapabilityClipboard, ErrUnsupported)
	}
	return cd.readClipboardText(ctx)
}

// WriteClipboardText writes Unicode text to the system clipboard.
func WriteClipboardText(ctx context.Context, text string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	cd, ok := d.(clipboardDriver)
	if !ok || d == nil || !d.capabilities().Supports(CapabilityClipboard) {
		return fmt.Errorf("system: %s: %w", CapabilityClipboard, ErrUnsupported)
	}
	return cd.writeClipboardText(ctx, text)
}

// ReadClipboardFiles reads a file-drop list from the system clipboard.
//
// The returned paths are absolute when they can be resolved. A clipboard that
// does not currently contain a file list returns an empty slice and nil error on
// supported platforms.
func ReadClipboardFiles(ctx context.Context) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	fd, ok := d.(clipboardFilesDriver)
	if !ok || d == nil || !d.capabilities().Supports(CapabilityClipboard) {
		return nil, fmt.Errorf("system: %s files: %w", CapabilityClipboard, ErrUnsupported)
	}
	paths, err := fd.readClipboardFiles(ctx)
	if err != nil {
		return nil, err
	}
	return normalizeClipboardReadFiles(paths), nil
}

// WriteClipboardFiles writes a file-drop list to the system clipboard.
func WriteClipboardFiles(ctx context.Context, paths []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	normalized, err := normalizeClipboardWriteFiles(paths)
	if err != nil {
		return err
	}

	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	fd, ok := d.(clipboardFilesDriver)
	if !ok || d == nil || !d.capabilities().Supports(CapabilityClipboard) {
		return fmt.Errorf("system: %s files: %w", CapabilityClipboard, ErrUnsupported)
	}
	return fd.writeClipboardFiles(ctx, normalized)
}

// ReadClipboardImagePNG reads the current clipboard image as PNG-encoded bytes.
//
// A clipboard without an image returns nil bytes and nil error on supported
// platforms.
func ReadClipboardImagePNG(ctx context.Context) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	id, ok := d.(clipboardImageDriver)
	if !ok || d == nil || !d.capabilities().Supports(CapabilityClipboard) {
		return nil, fmt.Errorf("system: %s image: %w", CapabilityClipboard, ErrUnsupported)
	}
	data, err := id.readClipboardImagePNG(ctx)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return append([]byte(nil), data...), nil
}

// WriteClipboardImagePNG writes PNG-encoded image bytes to the system clipboard.
func WriteClipboardImagePNG(ctx context.Context, data []byte) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateClipboardImagePNG(data); err != nil {
		return err
	}

	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	id, ok := d.(clipboardImageDriver)
	if !ok || d == nil || !d.capabilities().Supports(CapabilityClipboard) {
		return fmt.Errorf("system: %s image: %w", CapabilityClipboard, ErrUnsupported)
	}
	return id.writeClipboardImagePNG(ctx, append([]byte(nil), data...))
}

func normalizeClipboardReadFiles(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		normalized = append(normalized, path)
	}
	return normalized
}

func validateClipboardImagePNG(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("system: %s image: empty PNG data: %w", CapabilityClipboard, ErrInvalidTarget)
	}
	if _, err := png.DecodeConfig(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("system: %s image: invalid PNG data: %w: %w", CapabilityClipboard, ErrInvalidTarget, err)
	}
	return nil
}

func normalizeClipboardWriteFiles(paths []string) ([]string, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("system: %s files: empty path list", CapabilityClipboard)
	}
	normalized := make([]string, 0, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			return nil, fmt.Errorf("system: %s files: empty path", CapabilityClipboard)
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("system: %s files: path %q is invalid: %w", CapabilityClipboard, path, err)
		}
		if _, err := os.Stat(abs); err != nil {
			return nil, fmt.Errorf("system: %s files: path %q is unavailable: %w: %w", CapabilityClipboard, abs, ErrUnavailable, err)
		}
		normalized = append(normalized, abs)
	}
	return normalized, nil
}
