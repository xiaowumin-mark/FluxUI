package system

import (
	"context"
	"fmt"
	"strings"
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

// NotificationBackend selects the platform notification path.
type NotificationBackend string

const (
	// NotificationBackendAuto lets the driver choose the platform default.
	// Windows uses tray balloons unless actions are present, in which case Toast
	// is required. macOS/Linux command drivers map Auto to their basic system
	// notification path.
	NotificationBackendAuto NotificationBackend = ""
	// NotificationBackendBalloon requests the basic non-Toast notification path.
	NotificationBackendBalloon NotificationBackend = "balloon"
	// NotificationBackendToast requests a Windows Toast notification.
	NotificationBackendToast NotificationBackend = "toast"
)

// NotificationBackendProbe is a point-in-time probe for a specific notification backend.
type NotificationBackendProbe struct {
	Backend                    NotificationBackend
	Status                     CapabilityStatus
	Err                        error
	SupportsActionButtons      bool
	SupportsClickCallback      bool
	SupportsDismissCallback    bool
	SupportsActionCallback     bool
	SupportsProtocolActivation bool
	SupportsDurableActivation  bool
}

// Supported reports whether the backend is implemented by the active driver.
func (p NotificationBackendProbe) Supported() bool {
	return p.Status != CapabilityStatusUnsupported
}

// Available reports whether the backend appears usable right now.
func (p NotificationBackendProbe) Available() bool {
	return p.Status == CapabilityStatusAvailable
}

// NotificationAction describes one action button for drivers that support it.
type NotificationAction struct {
	ID    string
	Label string
	// URI makes the action use platform protocol activation when supported.
	//
	// On Windows Toast this maps to activationType="protocol" and can activate
	// a registered URI handler even after FluxUI's short-lived Toast listener exits.
	URI string
}

// NotificationOption configures system notifications.
type NotificationOption func(*notificationOptions)

type notificationOptions struct {
	title     string
	body      string
	kind      NotificationKind
	icon      string
	group     string
	backend   NotificationBackend
	appID     string
	launchURI string
	actions   []NotificationAction
	timeout   time.Duration
	onClick   func(NotificationEvent)
	onDismiss func(NotificationEvent)
	onAction  func(NotificationEvent)
}

type notificationDriver interface {
	notify(ctx context.Context, opts notificationOptions) error
}

type notificationCancelDriver interface {
	cancelNotificationGroup(group string) error
}

type notificationBackendProbeDriver interface {
	probeNotificationBackend(ctx context.Context, backend NotificationBackend, opts notificationOptions) NotificationBackendProbe
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

// NotificationBackendPath selects a specific notification backend when supported.
func NotificationBackendPath(backend NotificationBackend) NotificationOption {
	return func(opts *notificationOptions) {
		opts.backend = backend
	}
}

// NotificationAppID sets the platform AppUserModelID used by Toast notifications.
//
// Windows Toasts for unpackaged desktop apps typically require an AppUserModelID
// that is registered by an installer or shortcut. The default is "FluxUI".
func NotificationAppID(appID string) NotificationOption {
	return func(opts *notificationOptions) {
		opts.appID = strings.TrimSpace(appID)
	}
}

// NotificationLaunchURI sets the URI opened when the notification body is activated.
//
// On Windows Toast this uses protocol activation. The application or installer must
// register the URI scheme with the operating system for activation to reach the app.
func NotificationLaunchURI(uri string) NotificationOption {
	return func(opts *notificationOptions) {
		opts.launchURI = strings.TrimSpace(uri)
	}
}

// NotificationActions sets action buttons for drivers that support them.
func NotificationActions(actions ...NotificationAction) NotificationOption {
	return func(opts *notificationOptions) {
		opts.actions = cloneNotificationActions(actions)
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

// NotificationOnDismiss registers a dismissal callback for drivers that can report it.
//
// Drivers that cannot observe system notification dismissal will not synthesize callbacks.
func NotificationOnDismiss(fn func(NotificationEvent)) NotificationOption {
	return func(opts *notificationOptions) {
		opts.onDismiss = fn
	}
}

// NotificationOnAction registers an action callback for drivers that can report it.
//
// Drivers that cannot observe action activation will not synthesize callbacks.
func NotificationOnAction(fn func(NotificationEvent)) NotificationOption {
	return func(opts *notificationOptions) {
		opts.onAction = fn
	}
}

// ProbeNotificationBackend returns a point-in-time availability result for one
// notification backend.
//
// This is more specific than Probe(CapabilityNotification): on Windows the
// notification capability can be available through tray balloons even when Toast
// is unavailable because the current AppUserModelID is not registered.
func ProbeNotificationBackend(ctx context.Context, backend NotificationBackend, options ...NotificationOption) NotificationBackendProbe {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return normalizeNotificationBackendProbe(backend, NotificationBackendProbe{
			Backend: backend,
			Status:  CapabilityStatusUnavailable,
			Err:     err,
		})
	}

	opts := defaultNotificationOptions()
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}
	opts.actions = cloneNotificationActions(opts.actions)

	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	if d == nil || !d.capabilities().Supports(CapabilityNotification) {
		return normalizeNotificationBackendProbe(backend, NotificationBackendProbe{
			Backend: backend,
			Status:  CapabilityStatusUnsupported,
			Err:     fmt.Errorf("system: %s: %w", CapabilityNotification, ErrUnsupported),
		})
	}
	if _, ok := d.(notificationDriver); !ok {
		return normalizeNotificationBackendProbe(backend, NotificationBackendProbe{
			Backend: backend,
			Status:  CapabilityStatusUnsupported,
			Err:     fmt.Errorf("system: %s: %w", CapabilityNotification, ErrUnsupported),
		})
	}
	if pd, ok := d.(notificationBackendProbeDriver); ok {
		return normalizeNotificationBackendProbe(backend, pd.probeNotificationBackend(ctx, backend, opts))
	}
	return normalizeNotificationBackendProbe(backend, defaultNotificationBackendProbe(backend))
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
	opts.actions = cloneNotificationActions(opts.actions)

	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	nd, ok := d.(notificationDriver)
	if !ok || d == nil || !d.capabilities().Supports(CapabilityNotification) {
		return fmt.Errorf("system: %s: %w", CapabilityNotification, ErrUnsupported)
	}
	return nd.notify(ctx, opts)
}

// CancelNotificationGroup removes the current notification associated with group when supported.
//
// On Windows this applies to Shell_NotifyIcon balloon notifications and Toasts
// created with NotificationGroup. On Linux it applies to notify-send group IDs
// when the close command is available. Cancelling a missing group is a no-op.
func CancelNotificationGroup(ctx context.Context, group string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if group == "" {
		return fmt.Errorf("system: notification group is empty")
	}

	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	cd, ok := d.(notificationCancelDriver)
	if !ok || d == nil || !d.capabilities().Supports(CapabilityNotification) {
		return fmt.Errorf("system: %s: %w", CapabilityNotification, ErrUnsupported)
	}
	return cd.cancelNotificationGroup(group)
}

func cloneNotificationActions(actions []NotificationAction) []NotificationAction {
	if len(actions) == 0 {
		return nil
	}
	cloned := make([]NotificationAction, len(actions))
	copy(cloned, actions)
	return cloned
}

func defaultNotificationBackendProbe(backend NotificationBackend) NotificationBackendProbe {
	if backend == NotificationBackendAuto {
		backend = NotificationBackendBalloon
	}
	return NotificationBackendProbe{
		Backend: backend,
		Status:  CapabilityStatusAvailable,
	}
}

func normalizeNotificationBackendProbe(backend NotificationBackend, probe NotificationBackendProbe) NotificationBackendProbe {
	if probe.Backend == "" {
		probe.Backend = backend
	}
	if probe.Backend == NotificationBackendAuto {
		probe.Backend = NotificationBackendBalloon
	}
	switch probe.Status {
	case CapabilityStatusAvailable:
		probe.Err = nil
	case CapabilityStatusUnavailable:
		if probe.Err == nil {
			probe.Err = ErrUnavailable
		}
	default:
		probe.Status = CapabilityStatusUnsupported
		if probe.Err == nil {
			probe.Err = ErrUnsupported
		}
	}
	return probe
}
