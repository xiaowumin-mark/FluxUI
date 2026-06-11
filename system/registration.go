package system

import (
	"context"
	"fmt"
	"strings"
)

type registrationDriver interface {
	registerProtocolHandler(ctx context.Context, scheme, command string, opts registrationOptions) error
	unregisterProtocolHandler(ctx context.Context, scheme string) error
	registerFileAssociation(ctx context.Context, extension, progID, command string, opts registrationOptions) error
	unregisterFileAssociation(ctx context.Context, extension, progID string) error
	registerStartupTask(ctx context.Context, name, command string) error
	unregisterStartupTask(ctx context.Context, name string) error
	registerToastShortcut(ctx context.Context, appID, shortcutName, executable string, opts toastShortcutOptions) error
	unregisterToastShortcut(ctx context.Context, shortcutName string) error
}

type toastActivatorRegistrationDriver interface {
	registerToastActivator(ctx context.Context, clsid, command string) error
	unregisterToastActivator(ctx context.Context, clsid string) error
}

// RegistrationOption configures protocol and file-association registration.
type RegistrationOption func(*registrationOptions)

type registrationOptions struct {
	displayName string
	icon        string
}

// ToastShortcutOption configures Windows Toast shortcut registration.
type ToastShortcutOption func(*toastShortcutOptions)

type toastShortcutOptions struct {
	arguments      []string
	icon           string
	activatorCLSID string
}

// RegistrationDisplayName sets the display name written to the platform
// registration entry.
func RegistrationDisplayName(name string) RegistrationOption {
	return func(opts *registrationOptions) {
		opts.displayName = strings.TrimSpace(name)
	}
}

// RegistrationIcon sets the icon path or icon resource string written to the
// platform registration entry.
func RegistrationIcon(icon string) RegistrationOption {
	return func(opts *registrationOptions) {
		opts.icon = strings.TrimSpace(icon)
	}
}

// ToastShortcutArguments sets command-line arguments stored in a Toast shortcut.
func ToastShortcutArguments(args ...string) ToastShortcutOption {
	return func(opts *toastShortcutOptions) {
		opts.arguments = append([]string(nil), args...)
	}
}

// ToastShortcutIcon sets the shortcut icon path or resource string.
func ToastShortcutIcon(icon string) ToastShortcutOption {
	return func(opts *toastShortcutOptions) {
		opts.icon = strings.TrimSpace(icon)
	}
}

// ToastShortcutActivatorCLSID sets the COM activator CLSID stored on a Windows
// Toast shortcut.
//
// This is one prerequisite for durable Toast activation. The application still
// needs to register and run a matching COM LocalServer that handles activation.
func ToastShortcutActivatorCLSID(clsid string) ToastShortcutOption {
	return func(opts *toastShortcutOptions) {
		opts.activatorCLSID = strings.TrimSpace(clsid)
	}
}

// RegisterProtocolHandler registers a URI scheme handler for the current user.
//
// command is the command line used by the operating system. If it does not
// contain "%1", FluxUI appends " \"%1\"" so the activated URI is delivered to
// the process.
func RegisterProtocolHandler(ctx context.Context, scheme, command string, options ...RegistrationOption) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	scheme, err := normalizeProtocolScheme(scheme)
	if err != nil {
		return err
	}
	command, err = normalizeRegistrationCommand(command, true)
	if err != nil {
		return err
	}
	return withRegistrationDriver(ctx, func(ctx context.Context, driver registrationDriver) error {
		return driver.registerProtocolHandler(ctx, scheme, command, collectRegistrationOptions(options))
	})
}

// UnregisterProtocolHandler removes a URI scheme handler registration for the
// current user.
func UnregisterProtocolHandler(ctx context.Context, scheme string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	scheme, err := normalizeProtocolScheme(scheme)
	if err != nil {
		return err
	}
	return withRegistrationDriver(ctx, func(ctx context.Context, driver registrationDriver) error {
		return driver.unregisterProtocolHandler(ctx, scheme)
	})
}

// RegisterFileAssociation registers a file extension association for the
// current user.
//
// extension may be passed with or without a leading dot. command follows the
// same "%1" placeholder rule as RegisterProtocolHandler.
func RegisterFileAssociation(ctx context.Context, extension, progID, command string, options ...RegistrationOption) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	extension, err := normalizeFileAssociationExtension(extension)
	if err != nil {
		return err
	}
	progID, err = normalizeRegistrationIdentifier("progID", progID)
	if err != nil {
		return err
	}
	command, err = normalizeRegistrationCommand(command, true)
	if err != nil {
		return err
	}
	return withRegistrationDriver(ctx, func(ctx context.Context, driver registrationDriver) error {
		return driver.registerFileAssociation(ctx, extension, progID, command, collectRegistrationOptions(options))
	})
}

// UnregisterFileAssociation removes a file extension association for the current
// user. The extension key is removed only when it still points to progID.
func UnregisterFileAssociation(ctx context.Context, extension, progID string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	extension, err := normalizeFileAssociationExtension(extension)
	if err != nil {
		return err
	}
	progID, err = normalizeRegistrationIdentifier("progID", progID)
	if err != nil {
		return err
	}
	return withRegistrationDriver(ctx, func(ctx context.Context, driver registrationDriver) error {
		return driver.unregisterFileAssociation(ctx, extension, progID)
	})
}

// RegisterStartupTask registers a current-user startup command.
func RegisterStartupTask(ctx context.Context, name, command string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	name, err := normalizeRegistrationIdentifier("startup name", name)
	if err != nil {
		return err
	}
	command, err = normalizeRegistrationCommand(command, false)
	if err != nil {
		return err
	}
	return withRegistrationDriver(ctx, func(ctx context.Context, driver registrationDriver) error {
		return driver.registerStartupTask(ctx, name, command)
	})
}

// UnregisterStartupTask removes a current-user startup command.
func UnregisterStartupTask(ctx context.Context, name string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	name, err := normalizeRegistrationIdentifier("startup name", name)
	if err != nil {
		return err
	}
	return withRegistrationDriver(ctx, func(ctx context.Context, driver registrationDriver) error {
		return driver.unregisterStartupTask(ctx, name)
	})
}

// RegisterToastShortcut creates or updates a current-user Start Menu shortcut
// with a Windows AppUserModelID.
//
// This is the registration Windows Toast needs to associate notifications with
// an unpackaged desktop app. It does not install a COM activator; durable Toast
// callbacks still require an application-specific COM LocalServer or protocol
// activation flow.
func RegisterToastShortcut(ctx context.Context, appID, shortcutName, executable string, options ...ToastShortcutOption) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	appID, err := normalizeRegistrationIdentifier("AppUserModelID", appID)
	if err != nil {
		return err
	}
	shortcutName, err = normalizeToastShortcutName(shortcutName)
	if err != nil {
		return err
	}
	executable, err = normalizeToastShortcutExecutable(executable)
	if err != nil {
		return err
	}
	return withRegistrationDriver(ctx, func(ctx context.Context, driver registrationDriver) error {
		return driver.registerToastShortcut(ctx, appID, shortcutName, executable, collectToastShortcutOptions(options))
	})
}

// UnregisterToastShortcut removes a current-user Start Menu Toast shortcut.
func UnregisterToastShortcut(ctx context.Context, shortcutName string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	shortcutName, err := normalizeToastShortcutName(shortcutName)
	if err != nil {
		return err
	}
	return withRegistrationDriver(ctx, func(ctx context.Context, driver registrationDriver) error {
		return driver.unregisterToastShortcut(ctx, shortcutName)
	})
}

// RegisterToastActivator registers a current-user COM LocalServer command for
// a Windows Toast activator CLSID.
//
// The command should start the application in a mode that calls
// StartToastActivator or RunToastActivator for the same CLSID. This function
// only writes the CLSID registration; RegisterToastShortcut still needs to
// store the same CLSID in the Start Menu shortcut through
// ToastShortcutActivatorCLSID.
func RegisterToastActivator(ctx context.Context, clsid, command string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	clsid, err := normalizeToastActivatorCLSID(clsid)
	if err != nil {
		return err
	}
	command, err = normalizeRegistrationCommand(command, false)
	if err != nil {
		return err
	}
	return withToastActivatorRegistrationDriver(ctx, func(ctx context.Context, driver toastActivatorRegistrationDriver) error {
		return driver.registerToastActivator(ctx, clsid, command)
	})
}

// UnregisterToastActivator removes a current-user COM LocalServer registration
// for a Windows Toast activator CLSID.
func UnregisterToastActivator(ctx context.Context, clsid string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	clsid, err := normalizeToastActivatorCLSID(clsid)
	if err != nil {
		return err
	}
	return withToastActivatorRegistrationDriver(ctx, func(ctx context.Context, driver toastActivatorRegistrationDriver) error {
		return driver.unregisterToastActivator(ctx, clsid)
	})
}

func withRegistrationDriver(ctx context.Context, fn func(context.Context, registrationDriver) error) error {
	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	rd, ok := d.(registrationDriver)
	if !ok || d == nil || !d.capabilities().Supports(CapabilitySystemRegistration) {
		return fmt.Errorf("system: %s: %w", CapabilitySystemRegistration, ErrUnsupported)
	}
	return fn(ctx, rd)
}

func withToastActivatorRegistrationDriver(ctx context.Context, fn func(context.Context, toastActivatorRegistrationDriver) error) error {
	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	rd, ok := d.(toastActivatorRegistrationDriver)
	if !ok || d == nil || !d.capabilities().Supports(CapabilitySystemRegistration) {
		return fmt.Errorf("system: %s: %w", CapabilitySystemRegistration, ErrUnsupported)
	}
	return fn(ctx, rd)
}

func collectRegistrationOptions(options []RegistrationOption) registrationOptions {
	var opts registrationOptions
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}
	return opts
}

func collectToastShortcutOptions(options []ToastShortcutOption) toastShortcutOptions {
	var opts toastShortcutOptions
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}
	opts.arguments = append([]string(nil), opts.arguments...)
	return opts
}

func normalizeProtocolScheme(scheme string) (string, error) {
	scheme = strings.TrimSpace(strings.TrimSuffix(scheme, ":"))
	if scheme == "" {
		return "", fmt.Errorf("system: %s: protocol scheme is empty", CapabilitySystemRegistration)
	}
	for _, r := range scheme {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' {
			continue
		}
		return "", fmt.Errorf("system: %s: protocol scheme %q contains invalid character %q", CapabilitySystemRegistration, scheme, r)
	}
	return strings.ToLower(scheme), nil
}

func normalizeFileAssociationExtension(extension string) (string, error) {
	extension = strings.TrimSpace(extension)
	if extension == "" {
		return "", fmt.Errorf("system: %s: file extension is empty", CapabilitySystemRegistration)
	}
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	if extension == "." || strings.ContainsAny(extension, `\/:*?"<>|`) || strings.Contains(extension, " ") {
		return "", fmt.Errorf("system: %s: file extension %q is invalid", CapabilitySystemRegistration, extension)
	}
	return strings.ToLower(extension), nil
}

func normalizeRegistrationIdentifier(kind, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("system: %s: %s is empty", CapabilitySystemRegistration, kind)
	}
	if strings.ContainsAny(value, `\/:*?"<>|`) {
		return "", fmt.Errorf("system: %s: %s %q is invalid", CapabilitySystemRegistration, kind, value)
	}
	return value, nil
}

func normalizeRegistrationCommand(command string, includeTarget bool) (string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", fmt.Errorf("system: %s: command is empty", CapabilitySystemRegistration)
	}
	if includeTarget && !strings.Contains(command, "%1") {
		command += ` "%1"`
	}
	return command, nil
}

func normalizeToastShortcutName(name string) (string, error) {
	name = strings.TrimSpace(name)
	name = strings.TrimSuffix(name, ".lnk")
	if name == "" {
		return "", fmt.Errorf("system: %s: toast shortcut name is empty", CapabilitySystemRegistration)
	}
	if strings.ContainsAny(name, `\/:*?"<>|`) {
		return "", fmt.Errorf("system: %s: toast shortcut name %q is invalid", CapabilitySystemRegistration, name)
	}
	return name + ".lnk", nil
}

func normalizeToastShortcutExecutable(executable string) (string, error) {
	executable = strings.TrimSpace(executable)
	if executable == "" {
		return "", fmt.Errorf("system: %s: toast shortcut executable is empty: %w", CapabilitySystemRegistration, ErrInvalidTarget)
	}
	return executable, nil
}

func normalizeToastActivatorCLSID(clsid string) (string, error) {
	clsid = strings.TrimSpace(clsid)
	if clsid == "" {
		return "", fmt.Errorf("system: %s: toast activator CLSID is empty: %w", CapabilitySystemRegistration, ErrInvalidTarget)
	}
	return clsid, nil
}
