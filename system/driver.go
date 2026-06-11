package system

import "sync"

type driver interface {
	capabilities() CapabilitySet
}

type capabilityProbeDriver interface {
	probeCapability(cap Capability) CapabilityAvailability
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

func currentAvailability(cap Capability) CapabilityAvailability {
	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	if d == nil {
		return defaultCapabilityAvailability(nil, cap)
	}
	if pd, ok := d.(capabilityProbeDriver); ok {
		return normalizeCapabilityAvailability(cap, pd.probeCapability(cap))
	}
	return defaultCapabilityAvailability(d.capabilities(), cap)
}
