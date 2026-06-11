//go:build darwin || linux

package system

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	unixSystemEventChannelSize  = 16
	unixSystemEventPollInterval = 2 * time.Second
)

type unixSystemEventSource struct {
	kind  SystemEventKind
	name  string
	probe func() error
	read  func(context.Context) (string, error)
}

type unixSystemEventSubscription struct {
	once   sync.Once
	cancel context.CancelFunc
	ch     chan SystemEvent
}

func (unixDriver) subscribeSystemEvents(ctx context.Context, kinds []SystemEventKind) (systemEventHandle, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	sources, err := selectUnixSystemEventSources(unixSystemEventSourcesForKinds(kinds))
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			return nil, err
		}
		return nil, fmt.Errorf("system: %s: %w: %v", CapabilitySystemEvents, ErrUnavailable, err)
	}

	state := make(map[SystemEventKind]string, len(sources))
	active := make([]unixSystemEventSource, 0, len(sources))
	for _, source := range sources {
		value, err := source.read(ctx)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			continue
		}
		state[source.kind] = value
		active = append(active, source)
	}
	if len(active) == 0 {
		return nil, fmt.Errorf("system: %s: initialize event sources: %w", CapabilitySystemEvents, ErrUnavailable)
	}

	runCtx, cancel := context.WithCancel(ctx)
	sub := &unixSystemEventSubscription{
		cancel: cancel,
		ch:     make(chan SystemEvent, unixSystemEventChannelSize),
	}
	go sub.run(runCtx, active, state)
	return sub, nil
}

func (s *unixSystemEventSubscription) events() <-chan SystemEvent {
	if s == nil {
		return nil
	}
	return s.ch
}

func (s *unixSystemEventSubscription) close() error {
	if s == nil {
		return ErrClosed
	}
	s.once.Do(func() {
		s.cancel()
	})
	return nil
}

func (s *unixSystemEventSubscription) run(ctx context.Context, sources []unixSystemEventSource, state map[SystemEventKind]string) {
	defer close(s.ch)

	ticker := time.NewTicker(unixSystemEventPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.poll(ctx, sources, state)
		}
	}
}

func (s *unixSystemEventSubscription) poll(ctx context.Context, sources []unixSystemEventSource, state map[SystemEventKind]string) {
	for _, source := range sources {
		value, err := source.read(ctx)
		if err != nil {
			continue
		}
		if state[source.kind] == value {
			continue
		}
		state[source.kind] = value
		event := SystemEvent{
			Kind:   source.kind,
			Detail: source.name,
			Time:   time.Now(),
		}
		select {
		case s.ch <- event:
		default:
		}
	}
}

func unixSystemEventSourcesForKinds(kinds []SystemEventKind) []unixSystemEventSource {
	sources := unixSystemEventSources()
	if len(kinds) == 0 {
		return sources
	}
	filter := make(map[SystemEventKind]bool, len(kinds))
	for _, kind := range kinds {
		if kind != "" {
			filter[kind] = true
		}
	}
	selected := make([]unixSystemEventSource, 0, len(sources))
	for _, source := range sources {
		if filter[source.kind] {
			selected = append(selected, source)
		}
	}
	return selected
}

func selectUnixSystemEventSources(sources []unixSystemEventSource) ([]unixSystemEventSource, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("requested system event kinds are unsupported on this platform: %w", ErrUnsupported)
	}

	available := make([]unixSystemEventSource, 0, len(sources))
	missing := make([]string, 0, len(sources))
	for _, source := range sources {
		if source.kind == "" || source.read == nil {
			continue
		}
		if source.probe != nil {
			if err := source.probe(); err != nil {
				missing = append(missing, source.name)
				continue
			}
		}
		available = append(available, source)
	}
	if len(available) == 0 {
		if len(missing) == 0 {
			return nil, fmt.Errorf("no system event sources configured")
		}
		return nil, fmt.Errorf("system event sources unavailable: %s", strings.Join(uniqueStrings(missing), ", "))
	}
	return available, nil
}

func unixSystemEventCommandProbe(command string) func() error {
	return func() error {
		_, err := exec.LookPath(command)
		return err
	}
}

func runUnixSystemEventCommand(ctx context.Context, command string, args ...string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	runCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	output, err := exec.CommandContext(runCtx, command, args...).CombinedOutput()
	if ctxErr := ctx.Err(); ctxErr != nil {
		return "", ctxErr
	}
	if runCtxErr := runCtx.Err(); runCtxErr != nil {
		return "", runCtxErr
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w: %s", command, err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}
