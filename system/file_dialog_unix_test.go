//go:build darwin || linux

package system

import (
	"context"
	"errors"
	"testing"
)

func TestUnixFileDialogChecksContextBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := (unixDriver{}).openFileDialog(ctx, FileDialogOpenFile, defaultFileDialogOptions()); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestUnixFileDialogProviderSelection(t *testing.T) {
	candidates := unixFileDialogProviders()
	if len(candidates) == 0 {
		t.Fatal("expected file dialog candidates")
	}

	available := map[string]bool{}
	for _, name := range candidates[0].commandNames() {
		available[name] = true
	}
	provider, err := unixFileDialogProviderWithLookPath(func(name string) (string, error) {
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

	if _, err := unixFileDialogProviderWithLookPath(func(string) (string, error) {
		return "", errors.New("missing")
	}); err == nil {
		t.Fatal("expected missing providers to fail")
	}
}

func TestUnixFileDialogRejectsOwner(t *testing.T) {
	opts := defaultFileDialogOptions()
	opts.owner = 123

	if _, err := (unixDriver{}).openFileDialog(context.Background(), FileDialogOpenFile, opts); !IsUnsupported(err) {
		t.Fatalf("expected owner to be unsupported, got %v", err)
	}
}

func TestUnixFileDialogDefaultDirErrorIsStructured(t *testing.T) {
	opts := defaultFileDialogOptions()
	opts.defaultDir = "/definitely/missing/fluxui/file-dialog"

	if _, err := (unixDriver{}).openFileDialog(context.Background(), FileDialogOpenFile, opts); !IsFileDialogErrorKind(err, FileDialogErrorDefaultDir) {
		t.Fatalf("expected default dir structured error, got %v", err)
	}
}

func TestUnixFileDialogCommandBuilders(t *testing.T) {
	opts := defaultFileDialogOptions()
	opts.title = "Choose"
	opts.defaultDir = "/tmp"
	opts.defaultName = "demo.txt"
	opts.filters = []FileFilter{{Name: "Text", Patterns: []string{"txt", ".md"}}}

	for _, mode := range []FileDialogMode{FileDialogOpenFile, FileDialogOpenFiles, FileDialogSaveFile, FileDialogPickFolder} {
		providers := unixFileDialogProviders()
		for _, provider := range providers {
			command, err := provider.build(mode, opts)
			if err != nil {
				t.Fatalf("%s build %s: %v", provider.name, mode, err)
			}
			if command.name == "" || len(command.args) == 0 {
				t.Fatalf("%s build %s returned empty command: %#v", provider.name, mode, command)
			}
		}
	}
}

func TestUnixFileDialogOutputPaths(t *testing.T) {
	result := unixFileDialogResultFromOutput([]byte("/tmp/one.txt\n/tmp/two.txt\n"))
	if result.Cancelled || len(result.Paths) != 2 {
		t.Fatalf("unexpected output parse result: %#v", result)
	}

	cancelled := unixFileDialogResultFromOutput([]byte("\n"))
	if !cancelled.Cancelled || len(cancelled.Paths) != 0 {
		t.Fatalf("expected empty output to be cancelled, got %#v", cancelled)
	}
}
