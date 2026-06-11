//go:build !windows && !darwin && !linux

package system

type unsupportedDriver struct{}

func newPlatformDriver() driver {
	return unsupportedDriver{}
}

func (unsupportedDriver) capabilities() CapabilitySet {
	return CapabilitySet{}
}
