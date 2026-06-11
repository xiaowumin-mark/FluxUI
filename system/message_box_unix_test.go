//go:build darwin || linux

package system

import (
	"context"
	"errors"
	"testing"
)

func TestUnixMessageBoxChecksContextBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := (unixDriver{}).showMessageBox(ctx, defaultMessageBoxOptions()); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestUnixMessageBoxProviderSelection(t *testing.T) {
	candidates := unixMessageBoxProviders()
	if len(candidates) == 0 {
		t.Fatal("expected message box candidates")
	}

	available := map[string]bool{}
	for _, name := range candidates[0].commandNames() {
		available[name] = true
	}
	provider, err := unixMessageBoxProviderWithLookPath(func(name string) (string, error) {
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

	if _, err := unixMessageBoxProviderWithLookPath(func(string) (string, error) {
		return "", errors.New("missing")
	}); err == nil {
		t.Fatal("expected missing providers to fail")
	}
}

func TestUnixMessageBoxRejectsRichOptions(t *testing.T) {
	opts := defaultMessageBoxOptions()
	opts.details = "details"

	if _, err := (unixDriver{}).showDetailedMessageBox(context.Background(), opts); !IsUnsupported(err) {
		t.Fatalf("expected rich options to be unsupported, got %v", err)
	}
}

func TestUnixMessageBoxButtonMappings(t *testing.T) {
	label, ok := unixMessageBoxLabelForResult(MessageBoxYesNoCancel, MessageBoxResultNo)
	if !ok || label != "No" {
		t.Fatalf("unexpected No label: %q %v", label, ok)
	}
	if result, ok := unixMessageBoxResultForLabel("Retry"); !ok || result != MessageBoxResultRetry {
		t.Fatalf("unexpected Retry result: %q %v", result, ok)
	}
	if _, ok := unixMessageBoxCancelLabel(MessageBoxYesNo); ok {
		t.Fatal("yes/no should not expose a cancel button")
	}
	if label, ok := unixMessageBoxCancelLabel(MessageBoxRetryCancel); !ok || label != "Cancel" {
		t.Fatalf("unexpected retry/cancel cancel label: %q %v", label, ok)
	}
}
