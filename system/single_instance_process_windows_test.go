//go:build windows

package system

import (
	"errors"
	"net"
	"os"
	"testing"
)

func TestVerifyWindowsSingleInstancePublishedState(t *testing.T) {
	state := loopbackSingleInstanceState{Address: "127.0.0.1:45678", PID: 42}
	currentIdentity := windowsSingleInstanceImageIdentity{volumeSerialNumber: 1, fileIndexLow: 10}
	unrelatedIdentity := windowsSingleInstanceImageIdentity{volumeSerialNumber: 1, fileIndexLow: 11}

	tests := []struct {
		name            string
		listenerPID     int
		currentElevated bool
		targetElevated  bool
		targetIdentity  windowsSingleInstanceImageIdentity
		want            bool
	}{
		{name: "matching standard process", listenerPID: 42, want: true},
		{name: "listener PID mismatch", listenerPID: 7},
		{name: "missing listener", listenerPID: 0},
		{name: "matching elevated process", listenerPID: 42, currentElevated: true, targetElevated: true, targetIdentity: currentIdentity, want: true},
		{name: "lower integrity listener", listenerPID: 42, currentElevated: true},
		{name: "unrelated elevated image", listenerPID: 42, currentElevated: true, targetElevated: true, targetIdentity: unrelatedIdentity},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := verifyWindowsSingleInstancePublishedState(state, windowsSingleInstanceStateVerifier{
				listenerProcessID: func(string) (int, error) { return test.listenerPID, nil },
				processElevated:   func(int) (bool, error) { return test.targetElevated, nil },
				processImageIdentity: func(pid int) (windowsSingleInstanceImageIdentity, error) {
					if pid == state.PID {
						return test.targetIdentity, nil
					}
					return currentIdentity, nil
				},
				currentProcessID: func() int { return 99 },
				currentProcessElevated: func() bool {
					return test.currentElevated
				},
			})
			if err != nil {
				t.Fatalf("verify state: %v", err)
			}
			if got != test.want {
				t.Fatalf("trusted=%v want %v", got, test.want)
			}
		})
	}
}

func TestVerifyWindowsSingleInstancePublishedStatePropagatesLookupError(t *testing.T) {
	wantErr := errors.New("lookup failed")
	trusted, err := verifyWindowsSingleInstancePublishedState(
		loopbackSingleInstanceState{Address: "127.0.0.1:45678", PID: 42},
		windowsSingleInstanceStateVerifier{
			listenerProcessID: func(string) (int, error) { return 0, wantErr },
			processElevated:   func(int) (bool, error) { return false, nil },
			processImageIdentity: func(int) (windowsSingleInstanceImageIdentity, error) {
				return windowsSingleInstanceImageIdentity{}, nil
			},
			currentProcessID: func() int { return 99 },
			currentProcessElevated: func() bool {
				return false
			},
		},
	)
	if trusted || !errors.Is(err, wantErr) {
		t.Fatalf("trusted=%v err=%v, want lookup error", trusted, err)
	}
}

func TestWindowsSingleInstanceListenerProcessID(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	pid, err := windowsSingleInstanceListenerProcessID(listener.Addr().String())
	if err != nil {
		t.Fatalf("query listener owner: %v", err)
	}
	if pid != os.Getpid() {
		t.Fatalf("listener owner PID=%d want %d", pid, os.Getpid())
	}
}

func TestWindowsSingleInstanceProcessImageIdentityMatchesCurrentExecutable(t *testing.T) {
	first, err := windowsSingleInstanceProcessImageIdentity(os.Getpid())
	if err != nil {
		t.Fatalf("query current image identity: %v", err)
	}
	second, err := windowsSingleInstanceProcessImageIdentity(os.Getpid())
	if err != nil {
		t.Fatalf("query current image identity again: %v", err)
	}
	if first != second {
		t.Fatalf("current image identity changed: first=%#v second=%#v", first, second)
	}
}
