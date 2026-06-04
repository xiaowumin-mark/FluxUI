package system

// Capability 表示 FluxUI 系统层可以提供的一类操作系统能力。
type Capability string

const (
	// CapabilityWindow 表示窗口控制和窗口状态能力。
	CapabilityWindow Capability = "window"
	// CapabilityFileDialog 表示系统原生文件选择和保存对话框能力。
	CapabilityFileDialog Capability = "file_dialog"
	// CapabilityMessageBox 表示系统原生消息框能力。
	CapabilityMessageBox Capability = "message_box"
	// CapabilityNotification 表示系统通知能力。
	CapabilityNotification Capability = "notification"
	// CapabilityTray 表示任务栏托盘能力。
	CapabilityTray Capability = "tray"
	// CapabilityClipboard 表示系统剪贴板能力。
	CapabilityClipboard Capability = "clipboard"
	// CapabilityShell 表示通过系统默认程序打开 URL、文件或目录的能力。
	CapabilityShell Capability = "shell"
)

// CapabilitySet 记录当前平台 driver 支持的系统能力。
type CapabilitySet map[Capability]bool

// Supports 返回能力集合是否支持指定能力。
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

// Capabilities 返回当前平台 driver 支持的系统能力集合。
//
// 返回值是副本，调用方修改它不会影响系统层内部能力表。
func Capabilities() CapabilitySet {
	return currentCapabilities()
}

// Supports 返回当前平台 driver 是否支持指定系统能力。
func Supports(cap Capability) bool {
	return Capabilities().Supports(cap)
}
