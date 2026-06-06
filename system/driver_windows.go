//go:build windows

package system

type windowsDriver struct{}

func newPlatformDriver() driver {
	return windowsDriver{}
}

func (windowsDriver) capabilities() CapabilitySet {
	return CapabilitySet{
		CapabilityWindow:       true,
		CapabilityFileDialog:   true,
		CapabilityMessageBox:   true,
		CapabilityNotification: true,
		CapabilityTray:         true,
	}
}
