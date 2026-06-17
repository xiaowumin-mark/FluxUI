package system

// Capability identifies one class of operating-system functionality exposed by
// FluxUI's system layer.
type Capability string

const (
	CapabilityWindow             Capability = "window"
	CapabilityFileDialog         Capability = "file_dialog"
	CapabilityMessageBox         Capability = "message_box"
	CapabilityNotification       Capability = "notification"
	CapabilityTray               Capability = "tray"
	CapabilitySystemEvents       Capability = "system_events"
	CapabilityClipboard          Capability = "clipboard"
	CapabilityShell              Capability = "shell"
	CapabilitySingleInstance     Capability = "single_instance"
	CapabilitySystemRegistration Capability = "system_registration"
	CapabilityGlobalShortcut     Capability = "global_shortcut"
	CapabilityDragAndDrop        Capability = "drag_and_drop"
)

// CapabilityStatus describes whether a capability is supported and currently usable.
type CapabilityStatus string

const (
	// CapabilityStatusUnsupported means the active driver does not implement the capability.
	CapabilityStatusUnsupported CapabilityStatus = "unsupported"
	// CapabilityStatusUnavailable means the capability exists but is currently unavailable.
	CapabilityStatusUnavailable CapabilityStatus = "unavailable"
	// CapabilityStatusAvailable means the capability is supported and appears usable.
	CapabilityStatusAvailable CapabilityStatus = "available"
)

// CapabilityAvailability is a point-in-time probe result for one system capability.
type CapabilityAvailability struct {
	Capability Capability
	Status     CapabilityStatus
	Err        error
}

// Supported reports whether the capability is implemented by the active driver.
func (a CapabilityAvailability) Supported() bool {
	return a.Status != CapabilityStatusUnsupported
}

// Available reports whether the capability appears usable right now.
func (a CapabilityAvailability) Available() bool {
	return a.Status == CapabilityStatusAvailable
}

// CapabilitySet records the system capabilities supported by the current driver.
type CapabilitySet map[Capability]bool

// Supports reports whether the capability set supports cap.
func (s CapabilitySet) Supports(cap Capability) bool {
	if cap == "" || len(s) == 0 {
		return false
	}
	return s[cap]
}

func (s CapabilitySet) clone() CapabilitySet {
	if len(s) == 0 {
		return CapabilitySet{}
	}
	cloned := make(CapabilitySet, len(s))
	for cap, supported := range s {
		cloned[cap] = supported
	}
	return cloned
}

// Capabilities returns a copy of the active platform driver's capability set.
func Capabilities() CapabilitySet {
	return currentCapabilities()
}

// Supports reports whether the active platform driver supports cap.
func Supports(cap Capability) bool {
	return currentSupports(cap)
}

// Probe returns a point-in-time availability result for one capability.
//
// Probe always distinguishes unsupported capabilities from supported ones. Platform
// drivers may additionally report temporary unavailability, such as a missing shell
// service or native resource failure. When a driver has no deeper probe, supported
// capabilities are treated as available.
func Probe(cap Capability) CapabilityAvailability {
	return currentAvailability(cap)
}

// Availability is an alias for Probe, kept for call sites that read better with
// a noun-style API.
func Availability(cap Capability) CapabilityAvailability {
	return Probe(cap)
}

func defaultCapabilityAvailability(caps CapabilitySet, cap Capability) CapabilityAvailability {
	if cap == "" || !caps.Supports(cap) {
		return CapabilityAvailability{
			Capability: cap,
			Status:     CapabilityStatusUnsupported,
			Err:        ErrUnsupported,
		}
	}
	return CapabilityAvailability{
		Capability: cap,
		Status:     CapabilityStatusAvailable,
	}
}

func normalizeCapabilityAvailability(cap Capability, availability CapabilityAvailability) CapabilityAvailability {
	if availability.Capability == "" {
		availability.Capability = cap
	}
	switch availability.Status {
	case CapabilityStatusAvailable:
		availability.Err = nil
	case CapabilityStatusUnavailable:
		if availability.Err == nil {
			availability.Err = ErrUnavailable
		}
	default:
		availability.Status = CapabilityStatusUnsupported
		if availability.Err == nil {
			availability.Err = ErrUnsupported
		}
	}
	return availability
}
