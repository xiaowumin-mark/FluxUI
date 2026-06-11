//go:build darwin

package system

func unixPlatformCapabilities() CapabilitySet {
	return CapabilitySet{
		CapabilityClipboard:      true,
		CapabilityFileDialog:     true,
		CapabilityMessageBox:     true,
		CapabilityNotification:   true,
		CapabilitySystemEvents:   true,
		CapabilityTray:           true,
		CapabilityShell:          true,
		CapabilitySingleInstance: true,
		CapabilityDragAndDrop:    true,
	}
}
