<!-- fluxui-doc-meta
{
  "id": "tray_api",
  "title": "Tray API 托盘",
  "category": "使用指南",
  "order": 127,
  "summary": "Tray API 提供长期系统托盘图标、菜单、显示隐藏和事件回调能力。",
  "example": { "id": "system_api_basic" },
  "apis": [
    "system.NewTray(opts ...system.TrayOption) (*system.Tray, error)",
    "system.TrayIcon(path string) system.TrayOption",
    "system.TrayIconBytes(data []byte) system.TrayOption",
    "system.TrayIconResource(id uint16) system.TrayOption",
    "system.TrayTooltip(text string) system.TrayOption",
    "system.TrayMenuItems(items ...system.TrayMenuItem) system.TrayOption",
    "system.TrayMenuProvider(fn func() system.TrayMenu) system.TrayOption",
    "system.TrayOnClick(fn func(system.TrayEvent)) system.TrayOption",
    "system.TrayOnDoubleClick(fn func(system.TrayEvent)) system.TrayOption",
    "system.TrayMenuAction(id, label string, onClick func(system.TrayEvent)) system.TrayMenuItem",
    "system.TrayMenuSeparator() system.TrayMenuItem",
    "(*system.Tray).SetIcon(path string) error",
    "(*system.Tray).SetIconBytes(data []byte) error",
    "(*system.Tray).SetIconResource(id uint16) error",
    "(*system.Tray).SetTooltip(text string) error",
    "(*system.Tray).SetMenu(menu system.TrayMenu) error",
    "(*system.Tray).SetMenuProvider(fn func() system.TrayMenu) error",
    "(*system.Tray).Show() error",
    "(*system.Tray).Hide() error",
    "(*system.Tray).Visible() bool",
    "(*system.Tray).Closed() bool",
    "(*system.Tray).Close() error",
    "system.CloseTrays() error"
  ]
}
-->

# Tray API 托盘

Tray API 是 `system` 包里的长期系统托盘能力，用于后台驻留应用、隐藏主窗口后的恢复入口和常驻菜单。当前 Windows driver 已实现 `Shell_NotifyIconW` 长期图标、右键菜单、点击/双击事件、显示隐藏、关闭清理和 Explorer 重启后的图标恢复。macOS 提供基于 `osascript` + AppKit `NSStatusItem` 的命令型最小 driver；Linux 提供 `yad --notification` 命令型最小 driver。缺少平台命令或桌面托盘 host 不可用时返回包装后的 `ErrUnavailable`。

## 能力检测

调用前先检查能力：

```go
if !system.Supports(system.CapabilityTray) {
    return
}
```

未实现平台也可以直接调用 `system.NewTray`，返回的错误可通过 `system.IsUnsupported(err)` 判断。macOS 和 Linux 已声明 `CapabilityTray`；使用 `Probe(CapabilityTray)` 可以区分 `osascript` / `yad` 缺失导致的 `unavailable`。

## 基本用法

```go
tray, err := system.NewTray(
    system.TrayTooltip("FluxUI"),
    system.TrayOnDoubleClick(func(event system.TrayEvent) {
        // 显示主窗口或打开应用入口。
    }),
    system.TrayMenuItems(
        system.TrayMenuAction("show", "显示窗口", func(event system.TrayEvent) {
            // 显示窗口。
        }),
        system.TrayMenuSeparator(),
        system.TrayMenuAction("quit", "退出", func(event system.TrayEvent) {
            // 关闭托盘并退出应用。
        }),
    ),
)
if err != nil {
    return err
}
defer tray.Close()

if err := tray.Show(); err != nil {
    return err
}
```

`NewTray` 只创建托盘句柄，不自动显示图标。调用 `Show()` 后图标才会进入系统通知区。`Hide()` 只隐藏图标并保留句柄，后续可再次 `Show()`；`Visible()` / `Closed()` 可查询公共层生命周期状态。`Close()` 会释放系统图标和 native 资源，关闭后继续操作会返回包装后的 `ErrClosed`。

通过 `ui.RunElement` / `ui.RunElementMulti` 启动的应用会在 app lifecycle 正常退出时调用 `system.CloseTrays()`，统一清理仍然注册中的托盘。业务代码仍建议在明确退出动作中主动调用 `Close()`。

## 图标和 Tooltip

- `TrayIcon(path)`: 设置图标路径。Windows 下传空字符串时使用默认应用图标；传入文件路径时由系统按图标资源加载。
- `TrayIconBytes(data)`: 从内存中的 `.ico` bytes 设置图标，适合把图标嵌入 Go 代码或从配置加载。
- `TrayIconResource(id)`: 从当前进程资源中的 icon id 设置图标，适合配合 `.syso`/manifest/rsrc 注入的 Windows 资源。
- `TrayTooltip(text)`: 设置鼠标悬停提示。Windows 默认 tooltip 是 `FluxUI`。
- `SetIcon(path)`: 运行中更新图标；托盘可见时会同步 `Shell_NotifyIconW(NIM_MODIFY)`。
- `SetIconBytes(data)`: 运行中用 `.ico` bytes 更新图标。
- `SetIconResource(id)`: 运行中用当前进程资源 icon id 更新图标。
- `SetTooltip(text)`: 运行中更新 tooltip。

图标加载失败、Windows shell 或通知区当前不可用时，driver 会返回普通错误或包装后的 `ErrUnavailable`。

## 菜单

托盘菜单使用有序列表：

```go
menu := system.TrayMenu{
    system.TrayMenuAction("open", "打开", onOpen),
    system.TrayMenuItem{ID: "sync", Label: "正在同步", Disabled: true},
    system.TrayMenuItem{ID: "autostart", Label: "开机启动", Checked: true, OnClick: onToggle},
    system.TrayMenuItem{
        ID: "tools",
        Label: "工具",
        Children: system.TrayMenu{
            system.TrayMenuAction("settings", "设置", onSettings),
        },
    },
    system.TrayMenuSeparator(),
    system.TrayMenuItem{ID: "quit", Label: "退出", Default: true, OnClick: onQuit},
}
_ = tray.SetMenu(menu)
```

Windows 右键菜单支持普通项、禁用项、勾选项、默认项、分隔线和一层或多层子菜单。菜单项被点击时，callback 收到 `TrayEvent{Kind: TrayEventMenuItem, ItemID: item.ID}`。

菜单需要在打开前按当前应用状态动态生成时，使用 `TrayMenuProvider` 或运行中 `SetMenuProvider`：

```go
tray, err := system.NewTray(system.TrayMenuProvider(func() system.TrayMenu {
    return system.TrayMenu{
        system.TrayMenuItem{ID: "sync", Label: "正在同步", Checked: syncing},
        system.TrayMenuAction("quit", "退出", onQuit),
    }
}))
```

provider 会在 driver 需要显示菜单时同步调用，应该快速返回，不要在其中执行阻塞 I/O。

## 事件边界

事件类型包括：

- `TrayEventClicked`: 左键点击托盘图标。
- `TrayEventDoubleClick`: 左键双击托盘图标。
- `TrayEventMenuItem`: 点击右键菜单项。

公共层会把点击、双击和菜单项 callback 包装成异步派发，避免 Win32 消息回调线程直接执行业务逻辑。callback 中不要直接做布局工作；需要更新 FluxUI UI 时，使用 goroutine-safe 的 state setter、窗口句柄或应用自己的事件队列，然后触发窗口 redraw。

## Linux 实现说明

Linux v1 使用 `yad --notification` 启动一个长期托盘进程。`Show()` 启动托盘进程，`Hide()` 终止进程但保留句柄，`Close()` 终止进程并清理临时 FIFO、事件脚本和内存图标文件。左键点击和菜单项点击通过临时 FIFO 回传到 Go callback。`SetIcon` / `SetIconBytes` / `SetTooltip` / `SetMenu` / `SetMenuProvider` 在托盘可见时会重启 `yad` 进程以应用新状态。

Linux 命令型 driver 的限制：

- 需要安装 `yad`，并且当前桌面环境需要有可用的托盘或 StatusNotifier host。
- `TrayIconResource` 不支持，返回 `ErrUnsupported`。
- 子菜单会扁平化为 `父项 / 子项` 文本；禁用项会显示为不可触发的 no-op 项。
- `Checked` / `Default` 视觉状态取决于 `yad` 支持情况，v1 不承诺完全等价 Windows native menu。
- `TrayMenuProvider` 在 `Show()` 或运行中重启托盘进程时求值，不保证每次菜单打开前都实时刷新。
- 双击事件不由 `yad --notification` 稳定提供，v1 不承诺 `TrayOnDoubleClick`。

## macOS 实现说明

macOS v1 使用 `osascript` 运行一段 AppKit AppleScript 创建 `NSStatusItem`。`Show()` 启动长期 `osascript` 进程，`Hide()` 终止该进程但保留句柄，`Close()` 终止进程并清理临时事件日志和内存图标文件。菜单项点击和无菜单左键点击会写入临时事件日志，由 Go 侧轮询后派发 callback。`SetIcon` / `SetIconBytes` / `SetTooltip` / `SetMenu` / `SetMenuProvider` 在托盘可见时会重启 `osascript` 进程以应用新状态。

macOS 命令型 driver 的限制：

- 需要 `osascript`，并且当前会话需要允许 AppleScript 使用 AppKit 创建状态栏项。
- `TrayIconResource` 不支持，返回 `ErrUnsupported`。
- 支持路径/bytes 图标、tooltip、菜单、子菜单、禁用项、勾选项、显示隐藏和关闭清理；默认项视觉状态不作为 v1 承诺。
- `TrayMenuProvider` 在 `Show()` 或运行中重启托盘进程时求值，不保证每次菜单打开前都实时刷新。
- 双击事件不由该命令型 driver 稳定提供，v1 不承诺 `TrayOnDoubleClick`。

## Windows 实现说明

Windows driver 使用独立隐藏消息窗口和专用 message pump 接收托盘事件，不复用 Gio 主事件循环。长期托盘图标通过 `Shell_NotifyIconW` 执行 add、modify 和 delete；右键菜单通过 `CreatePopupMenu` / `TrackPopupMenu` 显示，并把 command ID 映射回 `TrayMenuItem.ID`。

driver 会注册 `TaskbarCreated` 消息。Explorer 或任务栏重启后，仍处于 visible 状态的托盘图标会尝试自动重新添加。

## 错误语义

- 当前平台不支持托盘能力：返回包装后的 `ErrUnsupported`。
- 平台支持但隐藏窗口、shell、通知区、托盘 host、平台命令或图标加载不可用：返回包装后的 `ErrUnavailable` 或具体系统错误。
- 托盘已关闭：返回包装后的 `ErrClosed`。

判断错误时使用 helper：

```go
if system.IsUnsupported(err) {
    // 禁用托盘入口。
}
if system.IsUnavailable(err) {
    // 平台支持，但当前系统托盘不可用。
}
if system.IsClosed(err) {
    // 托盘已经关闭，不应继续使用旧句柄。
}
```

Tray 不会自动 fallback 到 FluxUI 自绘菜单或应用内 Toast。需要降级行为时，应在业务层显式处理。

## 示例和验收

`examples/system_showcase` 提供 Tray 的人工验收入口。示例覆盖创建托盘、显示/隐藏托盘、隐藏主窗口、右键菜单、禁用项、勾选项和关闭清理。

Windows 本地人工验收需要确认：

- 创建后系统通知区出现托盘图标。
- 左键点击和左键双击能触发回调。
- 右键菜单可显示，菜单项点击能返回对应 item ID。
- 隐藏托盘后图标消失，再次显示后恢复。
- 关闭托盘后图标清理，继续操作返回 `ErrClosed`。
- Explorer 重启后 visible 图标可尝试恢复。

Linux 本地人工验收需要确认：

- `Probe(CapabilityTray)` 在安装 `yad` 且托盘 host 可用时返回 available。
- `Show()` 后托盘区出现图标，`Hide()` 后消失，`Close()` 后清理。
- 左键点击和菜单项点击能触发 callback。
- `SetTooltip`、`SetIcon` / `SetIconBytes` 和 `SetMenu` 更新后能重新显示新状态。

macOS 本地人工验收需要确认：

- `Probe(CapabilityTray)` 在 `osascript` 可用且桌面会话允许 AppKit 状态栏项时返回 available。
- `Show()` 后菜单栏出现状态项，`Hide()` 后消失，`Close()` 后清理。
- 左键点击或菜单项点击能触发 callback。
- `SetTooltip`、`SetIcon` / `SetIconBytes` 和 `SetMenu` 更新后能重新显示新状态。
