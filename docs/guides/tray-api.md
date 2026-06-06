<!-- fluxui-doc-meta
{
  "id": "tray_api",
  "title": "Tray API 托盘",
  "category": "使用指南",
  "order": 127,
  "summary": "Tray API 提供长期系统托盘图标、菜单、显示隐藏和事件回调能力。",
  "apis": [
    "system.NewTray(opts ...system.TrayOption) (*system.Tray, error)",
    "system.TrayIcon(path string) system.TrayOption",
    "system.TrayTooltip(text string) system.TrayOption",
    "system.TrayMenuItems(items ...system.TrayMenuItem) system.TrayOption",
    "system.TrayOnClick(fn func(system.TrayEvent)) system.TrayOption",
    "system.TrayOnDoubleClick(fn func(system.TrayEvent)) system.TrayOption",
    "system.TrayMenuAction(id, label string, onClick func(system.TrayEvent)) system.TrayMenuItem",
    "system.TrayMenuSeparator() system.TrayMenuItem",
    "(*system.Tray).SetIcon(path string) error",
    "(*system.Tray).SetTooltip(text string) error",
    "(*system.Tray).SetMenu(menu system.TrayMenu) error",
    "(*system.Tray).Show() error",
    "(*system.Tray).Hide() error",
    "(*system.Tray).Close() error"
  ]
}
-->

# Tray API 托盘

Tray API 是 `system` 包里的长期系统托盘能力，用于后台驻留应用、隐藏主窗口后的恢复入口和常驻菜单。当前 Windows driver 已实现 `Shell_NotifyIconW` 长期图标、右键菜单、点击/双击事件、显示隐藏、关闭清理和 Explorer 重启后的图标恢复。非 Windows 平台保持可编译并返回 `ErrUnsupported`。

## 能力检测

调用前先检查能力：

```go
if !system.Supports(system.CapabilityTray) {
    return
}
```

未实现平台也可以直接调用 `system.NewTray`，返回的错误可通过 `system.IsUnsupported(err)` 判断。

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

`NewTray` 只创建托盘句柄，不自动显示图标。调用 `Show()` 后图标才会进入系统通知区。`Hide()` 只隐藏图标并保留句柄，后续可再次 `Show()`；`Close()` 会释放系统图标和 native 资源，关闭后继续操作会返回包装后的 `ErrClosed`。

## 图标和 Tooltip

- `TrayIcon(path)`: 设置图标路径。Windows 下传空字符串时使用默认应用图标；传入文件路径时由系统按图标资源加载。
- `TrayTooltip(text)`: 设置鼠标悬停提示。Windows 默认 tooltip 是 `FluxUI`。
- `SetIcon(path)`: 运行中更新图标；托盘可见时会同步 `Shell_NotifyIconW(NIM_MODIFY)`。
- `SetTooltip(text)`: 运行中更新 tooltip。

图标加载失败、Windows shell 或通知区当前不可用时，driver 会返回普通错误或包装后的 `ErrUnavailable`。

## 菜单

托盘菜单使用有序列表：

```go
menu := system.TrayMenu{
    system.TrayMenuAction("open", "打开", onOpen),
    system.TrayMenuItem{ID: "sync", Label: "正在同步", Disabled: true},
    system.TrayMenuItem{ID: "autostart", Label: "开机启动", Checked: true, OnClick: onToggle},
    system.TrayMenuSeparator(),
    system.TrayMenuAction("quit", "退出", onQuit),
}
_ = tray.SetMenu(menu)
```

Windows 右键菜单支持普通项、禁用项、勾选项和分隔线。菜单项被点击时，callback 收到 `TrayEvent{Kind: TrayEventMenuItem, ItemID: item.ID}`。

## 事件边界

事件类型包括：

- `TrayEventClicked`: 左键点击托盘图标。
- `TrayEventDoubleClick`: 左键双击托盘图标。
- `TrayEventMenuItem`: 点击右键菜单项。

公共层会把点击、双击和菜单项 callback 包装成异步派发，避免 Win32 消息回调线程直接执行业务逻辑。callback 中不要直接做布局工作；需要更新 FluxUI UI 时，使用 goroutine-safe 的 state setter、窗口句柄或应用自己的事件队列，然后触发窗口 redraw。

## Windows 实现说明

Windows driver 使用独立隐藏消息窗口和专用 message pump 接收托盘事件，不复用 Gio 主事件循环。长期托盘图标通过 `Shell_NotifyIconW` 执行 add、modify 和 delete；右键菜单通过 `CreatePopupMenu` / `TrackPopupMenu` 显示，并把 command ID 映射回 `TrayMenuItem.ID`。

driver 会注册 `TaskbarCreated` 消息。Explorer 或任务栏重启后，仍处于 visible 状态的托盘图标会尝试自动重新添加。

## 错误语义

- 当前平台不支持托盘能力：返回包装后的 `ErrUnsupported`。
- 平台支持但隐藏窗口、shell、通知区或图标加载不可用：返回包装后的 `ErrUnavailable` 或具体系统错误。
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
