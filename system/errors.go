package system

import "errors"

// ErrUnsupported means the current platform or active driver does not support
// the requested system capability.
var ErrUnsupported = errors.New("system: unsupported on this platform")

// ErrUnavailable means the capability is supported in principle but is not
// currently usable in this environment.
var ErrUnavailable = errors.New("system: unavailable")

// ErrClosed means a system resource has already been closed.
var ErrClosed = errors.New("system: closed")

// ErrAlreadyRunning means another process owns the single-instance channel.
var ErrAlreadyRunning = errors.New("system: already running")

// ErrInvalidTarget means a requested system target is malformed or empty.
var ErrInvalidTarget = errors.New("system: invalid target")

// ErrTargetNotFound means a requested file, path, URL target, or registration
// target cannot be found.
var ErrTargetNotFound = errors.New("system: target not found")

// ErrNoDefaultHandler means the operating system has no registered handler for
// a requested shell target.
var ErrNoDefaultHandler = errors.New("system: no default handler")

// ErrAccessDenied means the operating system rejected a system operation because
// of permissions or policy.
var ErrAccessDenied = errors.New("system: access denied")

// IsUnsupported reports whether err indicates an unsupported system capability.
func IsUnsupported(err error) bool {
	return errors.Is(err, ErrUnsupported)
}

// IsUnavailable reports whether err indicates a temporarily unavailable system
// capability.
func IsUnavailable(err error) bool {
	return errors.Is(err, ErrUnavailable)
}

// IsClosed reports whether err indicates an already closed system resource.
func IsClosed(err error) bool {
	return errors.Is(err, ErrClosed)
}

// IsAlreadyRunning reports whether err indicates an existing primary instance.
func IsAlreadyRunning(err error) bool {
	return errors.Is(err, ErrAlreadyRunning)
}

// IsInvalidTarget reports whether err indicates a malformed target.
func IsInvalidTarget(err error) bool {
	return errors.Is(err, ErrInvalidTarget)
}

// IsTargetNotFound reports whether err indicates a missing target.
func IsTargetNotFound(err error) bool {
	return errors.Is(err, ErrTargetNotFound)
}

// IsNoDefaultHandler reports whether err indicates a missing OS handler.
func IsNoDefaultHandler(err error) bool {
	return errors.Is(err, ErrNoDefaultHandler)
}

// IsAccessDenied reports whether err indicates a permissions or policy denial.
func IsAccessDenied(err error) bool {
	return errors.Is(err, ErrAccessDenied)
}
