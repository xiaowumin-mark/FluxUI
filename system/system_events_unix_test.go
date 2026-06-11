//go:build darwin || linux

package system

import (
	"context"
	"errors"
	"testing"
)

func TestUnixSystemEventSourcesForKinds(t *testing.T) {
	if got := unixSystemEventSourcesForKinds(nil); len(got) == 0 {
		t.Fatal("expected default unix system event sources")
	}

	got := unixSystemEventSourcesForKinds([]SystemEventKind{SystemEventThemeChanged})
	if len(got) == 0 {
		t.Fatal("expected theme source")
	}
	for _, source := range got {
		if source.kind != SystemEventThemeChanged {
			t.Fatalf("unexpected source kind %q", source.kind)
		}
	}

	if got := unixSystemEventSourcesForKinds([]SystemEventKind{"not-real"}); len(got) != 0 {
		t.Fatalf("unexpected source for unknown event kind: %#v", got)
	}
}

func TestSelectUnixSystemEventSources(t *testing.T) {
	sources, err := selectUnixSystemEventSources([]unixSystemEventSource{
		{
			kind: SystemEventThemeChanged,
			name: "theme",
			read: func(context.Context) (string, error) {
				return "dark", nil
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected source selection error: %v", err)
	}
	if len(sources) != 1 || sources[0].name != "theme" {
		t.Fatalf("unexpected sources: %#v", sources)
	}

	_, err = selectUnixSystemEventSources([]unixSystemEventSource{
		{
			kind:  SystemEventPowerChanged,
			name:  "power",
			probe: func() error { return errors.New("missing") },
			read: func(context.Context) (string, error) {
				return "battery", nil
			},
		},
	})
	if err == nil || IsUnsupported(err) {
		t.Fatalf("expected unavailable source error, got %v", err)
	}

	_, err = selectUnixSystemEventSources(nil)
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported empty-source error, got %v", err)
	}
}
