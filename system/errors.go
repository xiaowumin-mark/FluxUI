package system

import "errors"

// ErrUnsupported 表示当前平台或当前 driver 不支持该系统能力。
var ErrUnsupported = errors.New("system: unsupported on this platform")

// ErrUnavailable 表示能力理论上受支持，但当前环境暂时不可用。
var ErrUnavailable = errors.New("system: unavailable")

// ErrClosed 表示系统资源已经关闭，不能继续操作。
var ErrClosed = errors.New("system: closed")

// IsUnsupported 返回 err 是否表示系统能力不受支持。
func IsUnsupported(err error) bool {
	return errors.Is(err, ErrUnsupported)
}

// IsUnavailable 返回 err 是否表示系统能力当前不可用。
func IsUnavailable(err error) bool {
	return errors.Is(err, ErrUnavailable)
}

// IsClosed 返回 err 是否表示系统资源已经关闭。
func IsClosed(err error) bool {
	return errors.Is(err, ErrClosed)
}
