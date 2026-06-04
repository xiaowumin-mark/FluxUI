package system

import (
	"context"
	"fmt"
	"time"
)

// NotificationKind identifies the semantic style of a system notification.
type NotificationKind string

const (
	NotificationInfo    NotificationKind = "info"
	NotificationSuccess NotificationKind = "success"
	NotificationWarning NotificationKind = "warning"
	NotificationError   NotificationKind = "error"
)

// NotificationEventKind identifies a notification lifecycle event.
type NotificationEventKind string

const (
	NotificationEventClicked   NotificationEventKind = "clicked"
	NotificationEventDismissed NotificationEventKind = "dismissed"
	NotificationEventAction    NotificationEventKind = "action"
)

// NotificationEvent records a system notification event when a driver can report one.
type NotificationEvent struct {
	Kind   NotificationEventKind
	Group  string
	Action string
}

// NotificationOption configures system notifications.
type NotificationOption func(*notificationOptions)

type notificationOptions struct {
	title   string
	body    string
	kind    NotificationKind
	icon    string
	group   string
	timeout time.Duration
	onClick func(NotificationEvent)
}

type notificationDriver interface {
	notify(ctx context.Context, opts notificationOptions) error
}

func defaultNotificationOptions() notificationOptions {
	return notificationOptions{
		kind: NotificationInfo,
	}
}

// NotificationTitle sets the system notification title.
func NotificationTitle(value string) NotificationOption {
	return func(opts *notificationOptions) {
		opts.title = value
	}
}

// NotificationBody sets the system notification body text.
func NotificationBody(value string) NotificationOption {
	return func(opts *notificationOptions) {
		opts.body = value
	}
}

// NotificationKindStyle sets the system notification semantic style.
func NotificationKindStyle(kind NotificationKind) NotificationOption {
	return func(opts *notificationOptions) {
		opts.kind = kind
	}
}

// NotificationIcon sets the system notification icon path.
func NotificationIcon(path string) NotificationOption {
	return func(opts *notificationOptions) {
		opts.icon = path
	}
}

// NotificationGroup sets a grouping key for related notifications.
func NotificationGroup(group string) NotificationOption {
	return func(opts *notificationOptions) {
		opts.group = group
	}
}

// NotificationTimeout sets the requested notification display duration.
func NotificationTimeout(timeout time.Duration) NotificationOption {
	return func(opts *notificationOptions) {
		opts.timeout = timeout
	}
}

// NotificationOnClick registers a click callback for drivers that can report notification events.
//
// Drivers that cannot observe system notification events will not synthesize callbacks.
func NotificationOnClick(fn func(NotificationEvent)) NotificationOption {
	return func(opts *notificationOptions) {
		opts.onClick = fn
	}
}

// Notify sends a non-blocking system notification.
func Notify(ctx context.Context, options ...NotificationOption) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	opts := defaultNotificationOptions()
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}

	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	nd, ok := d.(notificationDriver)
	if !ok || d == nil || !d.capabilities().Supports(CapabilityNotification) {
		return fmt.Errorf("system: %s: %w", CapabilityNotification, ErrUnsupported)
	}
	return nd.notify(ctx, opts)
}
