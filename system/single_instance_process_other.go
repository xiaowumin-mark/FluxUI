//go:build !darwin && !linux && !windows

package system

import (
	"context"
	"errors"
)

func lockSingleInstanceState(context.Context, string) (func() error, error) {
	return func() error { return nil }, nil
}

func singleInstanceProcessAlive(int) (bool, error) {
	return false, errors.New("process liveness check is unsupported")
}

func singleInstanceForwardingAllowed() bool {
	return true
}

func singleInstancePublishedStateTrusted(loopbackSingleInstanceState) (bool, error) {
	return true, nil
}
