package system

import "sync"

type driver interface {
	capabilities() CapabilitySet
}

var (
	driverMu     sync.RWMutex
	activeDriver driver = newPlatformDriver()
)

func setDriver(d driver) {
	if d == nil {
		d = newPlatformDriver()
	}

	driverMu.Lock()
	activeDriver = d
	driverMu.Unlock()
}

func currentCapabilities() CapabilitySet {
	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	if d == nil {
		return CapabilitySet{}
	}
	return d.capabilities().clone()
}
