//go:build windows

package system

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsShellResolveExistingPath(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(file, []byte("ok"), 0600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	got, err := windowsShellResolveExistingPath(file, "open")
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	want, err := filepath.Abs(file)
	if err != nil {
		t.Fatalf("abs temp file: %v", err)
	}
	if got != want {
		t.Fatalf("unexpected resolved path: got %q want %q", got, want)
	}

	_, err = windowsShellResolveExistingPath(filepath.Join(dir, "missing.txt"), "reveal")
	if err == nil {
		t.Fatal("expected missing path to fail")
	}
	if !IsUnavailable(err) {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}
	if !IsTargetNotFound(err) {
		t.Fatalf("expected ErrTargetNotFound, got %v", err)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}
}

func TestWindowsShellResolveFileURL(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(file, []byte("ok"), 0600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	fileURL := "file:///" + strings.TrimPrefix(filepath.ToSlash(file), "/")
	got, ok, err := windowsShellResolveFileURL(fileURL)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if !ok {
		t.Fatal("expected file URL to be handled")
	}
	want, err := filepath.Abs(file)
	if err != nil {
		t.Fatalf("abs temp file: %v", err)
	}
	if got != want {
		t.Fatalf("unexpected resolved file URL: got %q want %q", got, want)
	}

	_, ok, err = windowsShellResolveFileURL("https://example.com")
	if ok || err != nil {
		t.Fatalf("expected non-file URL to be ignored, ok=%v err=%v", ok, err)
	}

	remoteURLs := []string{
		"file://attacker.example/share/secret.txt",
		"file:////attacker.example/share/secret.txt",
		"file:///%5C%5Cattacker.example%5Cshare%5Csecret.txt",
	}
	for _, remoteURL := range remoteURLs {
		_, ok, err = windowsShellResolveFileURL(remoteURL)
		if !ok {
			t.Fatalf("expected remote file URL to be handled and rejected: %q", remoteURL)
		}
		if !errors.Is(err, ErrInvalidTarget) {
			t.Fatalf("expected remote file URL to return ErrInvalidTarget: url=%q err=%v", remoteURL, err)
		}
	}

	missingURL := "file:///" + strings.TrimPrefix(filepath.ToSlash(filepath.Join(dir, "missing.txt")), "/")
	_, ok, err = windowsShellResolveFileURL(missingURL)
	if !ok {
		t.Fatal("expected missing file URL to be handled")
	}
	if !IsUnavailable(err) {
		t.Fatalf("expected missing file URL to return ErrUnavailable, got %v", err)
	}
	if !IsTargetNotFound(err) {
		t.Fatalf("expected missing file URL to return ErrTargetNotFound, got %v", err)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected missing file URL to include os.ErrNotExist, got %v", err)
	}
}

func TestWindowsShellQuoteArgumentEscapesQuotes(t *testing.T) {
	got := windowsShellQuoteArgument(`C:\tmp\a "quoted" file.txt`)
	want := `"C:\tmp\a \"quoted\" file.txt"`
	if got != want {
		t.Fatalf("unexpected shell quote: got %q want %q", got, want)
	}
}

func TestWindowsShellExecuteErrorClassification(t *testing.T) {
	cases := []struct {
		code uintptr
		want error
	}{
		{windowsShellExecuteErrorFileNotFound, ErrTargetNotFound},
		{windowsShellExecuteErrorPathNotFound, ErrTargetNotFound},
		{windowsShellExecuteErrorAccessDenied, ErrAccessDenied},
		{windowsShellExecuteErrorAssociation, ErrNoDefaultHandler},
		{windowsShellExecuteErrorNoAssociation, ErrNoDefaultHandler},
		{windowsShellExecuteErrorBadFormat, ErrInvalidTarget},
		{windowsShellExecuteErrorDDETimeout, ErrUnavailable},
		{uintptr(0), ErrUnavailable},
	}
	for _, tt := range cases {
		if !errors.Is(windowsShellExecuteError(tt.code), tt.want) {
			t.Fatalf("code %d: expected %v, got %v", tt.code, tt.want, windowsShellExecuteError(tt.code))
		}
	}
}
