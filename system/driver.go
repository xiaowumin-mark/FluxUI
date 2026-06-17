package system

import "sync"

type driver interface {
	capabilities() CapabilitySet
}

type capabilityProbeDriver interface {
	probeCapability(cap Capability) CapabilityAvailability
}

var (
	driverMu           sync.RWMutex
	activeDriver       driver
	activeCapabilities CapabilitySet
)

func init() {
	setDriver(newPlatformDriver())
}

func setDriver(d driver) {
	if d == nil {
		d = newPlatformDriver()
	}
	caps := CapabilitySet{}
	if d != nil {
		caps = d.capabilities().clone()
	}

	driverMu.Lock()
	activeDriver = d
	activeCapabilities = caps
	driverMu.Unlock()
}

func currentCapabilities() CapabilitySet {
	return currentCapabilitiesSnapshot()
}

func currentCapabilitiesSnapshot() CapabilitySet {
	driverMu.RLock()
	caps := activeCapabilities.clone()
	driverMu.RUnlock()
	return caps
}

func currentSupports(cap Capability) bool {
	driverMu.RLock()
	supported := activeCapabilities.Supports(cap)
	driverMu.RUnlock()
	return supported
}

func currentDriverFor(cap Capability) (driver, bool) {
	driverMu.RLock()
	d := activeDriver
	supported := d != nil && activeCapabilities.Supports(cap)
	driverMu.RUnlock()
	return d, supported
}

func currentAvailability(cap Capability) CapabilityAvailability {
	driverMu.RLock()
	d := activeDriver
	caps := activeCapabilities
	driverMu.RUnlock()

	if d == nil {
		return defaultCapabilityAvailability(nil, cap)
	}
	if !caps.Supports(cap) {
		return defaultCapabilityAvailability(caps, cap)
	}
	if pd, ok := d.(capabilityProbeDriver); ok {
		return normalizeCapabilityAvailability(cap, pd.probeCapability(cap))
	}
	return defaultCapabilityAvailability(caps, cap)
}
