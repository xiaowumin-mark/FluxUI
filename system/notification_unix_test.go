//go:build darwin || linux

package system

import (
	"context"
	"errors"
	"testing"
)

func TestUnixNotificationChecksContextBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := (unixDriver{}).notify(ctx, defaultNotificationOptions()); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestUnixNotificationProviderSelection(t *testing.T) {
	candidates := unixNotificationProviders()
	if len(candidates) == 0 {
		t.Fatal("expected notification candidates")
	}

	available := map[string]bool{}
	for _, name := range candidates[0].commandNames() {
		available[name] = true
	}
	provider, err := unixNotificationProviderWithLookPath(func(name string) (string, error) {
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

	if _, err := unixNotificationProviderWithLookPath(func(string) (string, error) {
		return "", errors.New("missing")
	}); err == nil {
		t.Fatal("expected missing providers to fail")
	}
}

func TestUnixNotificationRejectsUnsupportedOptions(t *testing.T) {
	opts := defaultNotificationOptions()
	opts.backend = NotificationBackendToast
	if err := validateUnixNotificationOptions(opts); !IsUnsupported(err) {
		t.Fatalf("expected toast backend to be unsupported, got %v", err)
	}

	opts = defaultNotificationOptions()
	opts.actions = []NotificationAction{{ID: "open", Label: "Open"}}
	if err := validateUnixNotificationOptions(opts); !IsUnsupported(err) {
		t.Fatalf("expected actions to be unsupported, got %v", err)
	}

	opts = defaultNotificationOptions()
	opts.launchURI = "fluxui://open"
	if err := validateUnixNotificationOptions(opts); !IsUnsupported(err) {
		t.Fatalf("expected protocol activation to be unsupported, got %v", err)
	}
}

func TestUnixNotificationCommandBuilders(t *testing.T) {
	opts := defaultNotificationOptions()
	opts.title = "Done"
	opts.body = "Export finished."
	opts.timeout = 0

	for _, provider := range unixNotificationProviders() {
		command, err := provider.build(opts, "")
		if err != nil {
			t.Fatalf("%s build: %v", provider.name, err)
		}
		if command.name == "" || len(command.args) == 0 {
			t.Fatalf("%s returned empty command: %#v", provider.name, command)
		}
	}
}

func TestUnixNotificationBackendProbeRejectsToast(t *testing.T) {
	probe := (unixDriver{}).probeNotificationBackend(context.Background(), NotificationBackendToast, defaultNotificationOptions())
	if probe.Status != CapabilityStatusUnsupported || !IsUnsupported(probe.Err) {
		t.Fatalf("expected unsupported toast probe, got %#v", probe)
	}
}
