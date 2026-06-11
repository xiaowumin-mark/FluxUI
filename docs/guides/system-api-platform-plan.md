<!-- fluxui-doc-meta
{
  "id": "system_api_platform_plan",
  "title": "System API 平台计划",
  "category": "使用指南",
  "order": 130,
  "summary": "规划 macOS 和 Linux 的 System API 最小 driver，实现顺序、平台接口和错误语义。",
  "example": { "id": "system_api_basic" },
  "apis": [
    "system.CapabilityFileDialog",
    "system.CapabilityMessageBox",
    "system.CapabilityNotification",
    "system.CapabilityTray",
    "system.CapabilitySystemEvents",
    "system.CapabilityClipboard",
    "system.CapabilityShell",
    "system.CapabilityDragAndDrop",
    "system.ErrUnsupported",
    "system.ErrUnavailable"
  ]
}
-->

# System API 平台计划

System API 当前采用 Windows first 策略，非 Windows 平台保持可编译；macOS / Linux 已先补 `CapabilityClipboard`、`CapabilityFileDialog`、`CapabilityMessageBox`、`CapabilityNotification`、`CapabilitySystemEvents`、`CapabilityTray`、`CapabilityShell` 和 `CapabilityDragAndDrop` 最小 driver。macOS Tray 使用 `osascript` + AppKit `NSStatusItem` 命令型 driver，Linux Tray 使用 `yad --notification` 命令型 driver，其它未实现平台仍返回 `ErrUnsupported`。Drag & Drop 的实际收发仍由 `widget` / `ui` 基于 Gio `io/transfer` 完成，system driver 只报告 probe。后续 macOS 和 Linux driver 应在不改变公共 API 的前提下逐步补实现，并继续遵守“不静默 fallback”的原则。

## 优先级

建议先补最小可用 desktop driver：

1. File Dialog：打开文件、多选、保存文件、选择目录、取消语义、owner/modal 行为。
2. MessageBox：基础信息/确认框、owner/modal 行为、按钮映射。
3. Notification：系统通知最小展示路径，点击/动作回调按平台能力补。
4. Tray：长期托盘图标、菜单、显示/隐藏、关闭清理。
5. System Events：显示器/DPI/主题/电源/会话事件。

Clipboard、Shell open 和 Drag & Drop probe 的 Windows v1 已在 Phase G 落地；macOS/Linux Clipboard 文本、Shell open、File Dialog、MessageBox、Notification 命令 driver、System Events 轮询 driver、Tray 命令型 driver 和 Drag & Drop probe v1 已补充。后续继续按本规划把 macOS Tray 升级到更完整的 AppKit driver，把 Linux Tray 升级到 StatusNotifier/AppIndicator，把 System Events 升级到原生通知/portal/DBus 事件流，把 Notification 从命令 driver 升级到 UserNotifications / freedesktop DBus 路径，并把 File Dialog / MessageBox 从命令 driver 升级到 owner/modal 能力更完整的原生 toolkit 或 portal driver。

## macOS 最小方案

| 能力 | 推荐接口 | 第一版范围 |
| --- | --- | --- |
| File Dialog | v1 `osascript choose file/folder/file name`；后续 `NSOpenPanel` / `NSSavePanel` | v1 支持打开文件、多选、保存、目录选择、取消返回 `Cancelled=true`；后续补 owner/modal |
| MessageBox | v1 `osascript display dialog`；后续 `NSAlert` | v1 支持 info/warning/error/question 和标准按钮映射；后续补 owner/modal 与更细关闭语义 |
| Notification | v1 `osascript display notification`；后续 UserNotifications.framework | v1 支持 title/body 和基础语义提示；不支持 group、cancel、actions、protocol activation 或 callback；后续补 UserNotifications request、group 与 action callback |
| Tray | v1 `osascript` + AppKit `NSStatusItem`；后续 cgo/AppKit driver | v1 支持路径/bytes 图标、tooltip、菜单/子菜单、显示隐藏、关闭清理、无菜单左键点击和菜单项回调；不支持 resource icon、默认项视觉状态或双击事件 |
| System Events | v1 `ioreg` / `defaults` / `pmset` / `stat` 轮询；后续 `NSWorkspace` notifications、DistributedNotificationCenter、screen notifications | v1 支持 display/DPI/theme/settings/power/session 快照变化；后续补原生事件流 |
| Clipboard | `pbpaste` / `pbcopy` | Unicode 文本读写 |
| Shell open | `open` / `open -R` | URL、文件、目录打开；Finder 定位文件或目录 |
| Drag & Drop probe | Gio `io/transfer` | system 层只报告能力；实际 `DropTarget` / `DragSource` 行为由 Gio 桌面后端决定 |

实现建议：

- 使用 `//go:build darwin` 独立文件，避免影响 Windows driver。
- AppKit 调用必须切到 main thread；命令型 v1 通过长期 `osascript` 进程隔离，后续原生 AppKit driver 需要在 driver 内部处理线程约束，不暴露给调用方。
- owner 句柄第一版可先只支持当前 app modal；如果无法绑定 FluxUI native window，明确返回能力限制或无 owner 行为。
- 未实现的子能力返回包装后的 `ErrUnsupported` 或 `ErrUnavailable`，不要调用自绘 Dialog/Toast 作为 fallback。

## Linux 最小方案

| 能力 | 推荐接口 | 第一版范围 |
| --- | --- | --- |
| File Dialog | v1 `zenity` / `kdialog`；后续 xdg-desktop-portal FileChooser | v1 支持打开文件、多选、保存、目录选择；后续优先支持 Flatpak/Wayland/X11 一致 portal 路径 |
| MessageBox | v1 `zenity` / `kdialog`；后续 xdg-desktop-portal Dialog 或 GTK/Qt fallback | v1 支持基础信息/确认框和标准按钮映射；无可用命令时 `ErrUnavailable` |
| Notification | v1 `notify-send` / `kdialog`；后续 freedesktop.org Desktop Notifications DBus | v1 支持 title/body/icon/timeout/urgency；`notify-send` 可记录 id 做 group replace，cancel 需要 `gdbus`；actions 和 callback 后续按 notification daemon 能力评估 |
| Tray | v1 `yad --notification`；后续 StatusNotifierItem/AppIndicator | v1 支持图标、tooltip、扁平菜单、显示隐藏和关闭清理；后续补原生菜单、checked/default 和双击等完整行为 |
| System Events | v1 `/sys` / `gsettings` / `loginctl` 轮询；后续 DBus、xdg portals、desktop settings | v1 支持 display/DPI/theme/settings/power/session 快照变化；后续补桌面环境原生事件流 |
| Clipboard | `wl-clipboard`、`xclip`、`xsel` | Unicode 文本读写，按可用命令选择 provider |
| Shell open | `xdg-open` | URL、文件、目录打开；文件 reveal 打开父目录 |
| Drag & Drop probe | Gio `io/transfer` | system 层只报告能力；实际 `DropTarget` / `DragSource` 行为由 Gio 桌面后端决定 |

实现建议：

- 使用 `//go:build linux` 独立文件。
- 优先走 DBus/portal，减少对 GTK/Qt 运行时的强依赖。
- portal 不存在、notification daemon 不支持 actions、StatusNotifier host 不存在时返回 `ErrUnavailable`。
- Linux 桌面环境差异大，文档必须记录 GNOME/KDE/Wayland/X11 的测试环境。

## 测试要求

每个平台补 driver 时至少增加：

- 公共层 dispatch 测试：能力声明、unsupported/unavailable、context 已取消。
- 平台构造测试：过滤器、按钮映射、通知 payload、菜单模型。
- 人工验收文档：平台版本、桌面环境、owner/modal 行为、通知/托盘服务可用性。

不允许因为某个平台未实现而让 `go test ./...` 或交叉平台编译失败。未实现能力必须稳定返回 `ErrUnsupported`。
