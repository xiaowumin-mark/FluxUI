//go:build darwin || linux

package system

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestUnixSystemCapabilities(t *testing.T) {
	if !Supports(CapabilityShell) {
		t.Fatal("unix platform driver should expose shell capability")
	}
	if !Supports(CapabilityClipboard) {
		t.Fatal("unix platform driver should expose clipboard capability")
	}
	if !Supports(CapabilityMessageBox) {
		t.Fatal("unix platform driver should expose message box capability")
	}
	if !Supports(CapabilitySingleInstance) {
		t.Fatal("unix platform driver should expose single instance capability")
	}
}

func TestUnixClipboardChecksContextBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := ReadClipboardText(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected read context cancellation, got %v", err)
	}
	if err := WriteClipboardText(ctx, "text"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected write context cancellation, got %v", err)
	}
}

func TestUnixClipboardProviderSelection(t *testing.T) {
	candidates := unixClipboardCandidates()
	if len(candidates) == 0 {
		t.Fatal("expected clipboard candidates")
	}

	available := map[string]bool{}
	for _, name := range candidates[0].commandNames() {
		available[name] = true
	}
	provider, err := unixClipboardProviderWithLookPath(func(name string) (string, error) {
		if available[name] {
			return "/usr/bin/" + name, nil
		}
		return "", errors.New("missing")
	})
	if err != nil {
		t.Fatalf("unexpected provider error: %v", err)
	}
	if provider.name != candidates[0].name {
		t.Fatalf("expected first provider %q, got %q", candidates[0].name, provider.name)
	}

	if _, err := unixClipboardProviderWithLookPath(func(string) (string, error) {
		return "", errors.New("missing")
	}); err == nil {
		t.Fatal("expected missing providers to fail")
	}
}

func TestUnixShellChecksContextBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := OpenURL(ctx, "https://example.com"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestUnixShellCommandBuilders(t *testing.T) {
	if unixShellOpenCommand() == "" {
		t.Fatal("expected shell open command")
	}
	if got := unixShellOpenURLCommand("https://example.com"); got.name == "" || len(got.args) == 0 {
		t.Fatalf("unexpected open URL command: %#v", got)
	}
	if got := unixShellOpenPathCommand("/tmp"); got.name == "" || len(got.args) == 0 {
		t.Fatalf("unexpected open path command: %#v", got)
	}
	if got := unixShellRevealPathCommand("/tmp/example.txt"); got.name == "" || len(got.args) == 0 {
		t.Fatalf("unexpected reveal path command: %#v", got)
	}
}

func TestUnixShellResolveExistingPath(t *testing.T) {
	file := filepath.Join(t.TempDir(), "exists.txt")
	if err := os.WriteFile(file, []byte("ok"), 0600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	resolved, err := unixShellResolveExistingPath(file, "open")
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if resolved == "" || !filepath.IsAbs(resolved) {
		t.Fatalf("expected absolute path, got %q", resolved)
	}

	missing := filepath.Join(t.TempDir(), "missing.txt")
	if _, err := unixShellResolveExistingPath(missing, "open"); !IsUnavailable(err) || !IsTargetNotFound(err) {
		t.Fatalf("expected unavailable target-not-found, got %v", err)
	}
}

func TestUnixShellResolveFileURL(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file url.txt")
	if err := os.WriteFile(file, []byte("ok"), 0600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	fileURL := "file://" + filepath.ToSlash(file)
	resolved, ok, err := unixShellResolveFileURL(fileURL)
	if err != nil || !ok {
		t.Fatalf("unexpected file URL resolve: path=%q ok=%v err=%v", resolved, ok, err)
	}
	if resolved == "" {
		t.Fatal("expected resolved file URL path")
	}

	if _, ok, err := unixShellResolveFileURL("https://example.com"); ok || err != nil {
		t.Fatalf("expected non-file URL to be ignored, ok=%v err=%v", ok, err)
	}
}
