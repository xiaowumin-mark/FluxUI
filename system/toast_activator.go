package system

import (
	"context"
	"fmt"
	"sync"
)

// ToastActivationEvent is delivered when a Windows Toast COM activator receives
// a durable notification activation.
type ToastActivationEvent struct {
	AppID     string
	Arguments string
	UserInput map[string]string
}

// ToastActivator is a running Toast COM activator registration for the current
// process.
type ToastActivator struct {
	handle toastActivatorHandle
}

type toastActivatorHandle interface {
	Close() error
	Done() <-chan struct{}
}

type toastActivatorDriver interface {
	startToastActivator(ctx context.Context, clsid string, fn func(ToastActivationEvent)) (toastActivatorHandle, error)
}

// StartToastActivator registers a COM class object in the current process so
// Windows Toast durable activation can call fn.
//
// On Windows, clsid must match the CLSID registered by RegisterToastActivator
// and written to the Start Menu shortcut with ToastShortcutActivatorCLSID. Other
// platforms return ErrUnsupported.
func StartToastActivator(ctx context.Context, clsid string, fn func(ToastActivationEvent)) (*ToastActivator, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	clsid, err := normalizeToastActivatorCLSID(clsid)
	if err != nil {
		return nil, err
	}
	if fn == nil {
		return nil, fmt.Errorf("system: %s: toast activator callback is nil", CapabilityNotification)
	}

	d, supported := currentDriverFor(CapabilityNotification)
	td, ok := d.(toastActivatorDriver)
	if !ok || !supported {
		return nil, fmt.Errorf("system: %s: %w", CapabilityNotification, ErrUnsupported)
	}
	handle, err := td.startToastActivator(ctx, clsid, fn)
	if err != nil {
		return nil, err
	}
	return &ToastActivator{handle: handle}, nil
}

// RunToastActivator starts a Toast activator and blocks until ctx is cancelled
// or the activator stops.
func RunToastActivator(ctx context.Context, clsid string, fn func(ToastActivationEvent)) error {
	activator, err := StartToastActivator(ctx, clsid, fn)
	if err != nil {
		return err
	}
	defer activator.Close()
	<-activator.Done()
	if ctx != nil {
		return ctx.Err()
	}
	return nil
}

// Close stops the current-process Toast activator.
func (a *ToastActivator) Close() error {
	if a == nil || a.handle == nil {
		return nil
	}
	return a.handle.Close()
}

// Done returns a channel closed when the activator has stopped.
func (a *ToastActivator) Done() <-chan struct{} {
	if a == nil || a.handle == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return a.handle.Done()
}

type simpleToastActivatorHandle struct {
	once   sync.Once
	cancel context.CancelFunc
	done   <-chan struct{}
}

func (h *simpleToastActivatorHandle) Close() error {
	h.once.Do(func() {
		if h.cancel != nil {
			h.cancel()
		}
		if h.done != nil {
			<-h.done
		}
	})
	return nil
}

func (h *simpleToastActivatorHandle) Done() <-chan struct{} {
	return h.done
}
