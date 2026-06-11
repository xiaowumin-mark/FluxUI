<!-- fluxui-doc-meta
{
  "id": "notification_api",
  "title": "Notification API 系统通知",
  "category": "使用指南",
  "order": 126,
  "summary": "Notification API 提供系统通知的公共入口、能力检测、错误语义和事件边界。",
  "example": { "id": "system_api_basic" },
  "apis": [
    "system.Notify(ctx context.Context, opts ...system.NotificationOption) error",
    "system.NotificationTitle(value string) system.NotificationOption",
    "system.NotificationBody(value string) system.NotificationOption",
    "system.NotificationKindStyle(kind system.NotificationKind) system.NotificationOption",
    "system.NotificationIcon(path string) system.NotificationOption",
    "system.NotificationGroup(group string) system.NotificationOption",
    "system.NotificationBackendPath(backend system.NotificationBackend) system.NotificationOption",
    "system.NotificationAppID(appID string) system.NotificationOption",
    "system.NotificationLaunchURI(uri string) system.NotificationOption",
    "system.NotificationActions(actions ...system.NotificationAction) system.NotificationOption",
    "system.NotificationTimeout(timeout time.Duration) system.NotificationOption",
    "system.NotificationOnClick(fn func(system.NotificationEvent)) system.NotificationOption",
    "system.NotificationOnDismiss(fn func(system.NotificationEvent)) system.NotificationOption",
    "system.NotificationOnAction(fn func(system.NotificationEvent)) system.NotificationOption",
    "system.ProbeNotificationBackend(ctx context.Context, backend system.NotificationBackend, opts ...system.NotificationOption) system.NotificationBackendProbe",
    "system.CancelNotificationGroup(ctx context.Context, group string) error",
    "system.RegisterToastShortcut(ctx context.Context, appID, shortcutName, executable string, opts ...system.ToastShortcutOption) error",
    "system.ToastShortcutActivatorCLSID(clsid string) system.ToastShortcutOption",
    "system.RegisterToastActivator(ctx context.Context, clsid, command string) error",
    "system.UnregisterToastActivator(ctx context.Context, clsid string) error",
    "system.StartToastActivator(ctx context.Context, clsid string, fn func(system.ToastActivationEvent)) (*system.ToastActivator, error)",
    "system.RunToastActivator(ctx context.Context, clsid string, fn func(system.ToastActivationEvent)) error",
    "system.UnregisterToastShortcut(ctx context.Context, shortcutName string) error"
  ]
}
-->

# Notification API 系统通知

Notification API 是 `system` 包里的非阻塞系统消息能力，用于后台任务完成、错误提醒和长任务通知。当前 Windows 默认使用 `Shell_NotifyIconW` 提交托盘气泡通知；显式请求 Toast 或配置 action buttons 时，会走 Windows Toast 请求路径。macOS / Linux 已提供最小命令 driver：macOS 使用 `osascript display notification`，Linux 优先使用 `notify-send`，不可用时尝试 `kdialog --passivepopup`。

## 当前状态

Windows driver 当前声明 `CapabilityNotification`。默认 `Notify` 会创建进程级隐藏消息窗口，并通过 `Shell_NotifyIconW` 临时添加托盘图标后提交系统气泡通知；请求 Toast 时会构造 Toast XML 并调用 Windows WinRT Toast API。macOS / Linux driver 也声明 `CapabilityNotification`，但只承诺最小展示路径；缺少 `osascript`、`notify-send` 或 `kdialog` 等平台命令时返回包装后的 `ErrUnavailable`。

调用方仍应同时处理能力缺失和临时不可用：

```go
if !system.Supports(system.CapabilityNotification) {
    return nil
}

err := system.Notify(ctx,
    system.NotificationTitle("导出完成"),
    system.NotificationBody("文件已经保存到输出目录。"),
    system.NotificationKindStyle(system.NotificationSuccess),
)
if err != nil {
    if system.IsUnavailable(err) {
        // Windows shell、通知区、图标加载或 macOS/Linux 平台命令当前不可用。
        return nil
    }
    return err
}
```

如果需要区分 backend 级别能力，使用 `ProbeNotificationBackend`。它比 `Probe(CapabilityNotification)` 更具体：Windows 上托盘气泡可用，不代表当前 AppUserModelID 对 Toast 也可用；macOS / Linux 上 `Auto` 和 `Balloon` 映射到命令型系统通知，`Toast` 返回 `ErrUnsupported`。

```go
probe := system.ProbeNotificationBackend(ctx,
    system.NotificationBackendToast,
    system.NotificationAppID("com.example.FluxUI"),
)
if !probe.Available() {
    // Toast 当前不可用，可回退到 balloon 或应用内消息。
}
if !probe.SupportsDurableActivation {
    // 当前平台没有 durable activation；可使用 protocol activation。
}
```

## 基本用法

```go
err := system.Notify(ctx,
    system.NotificationTitle("后台任务完成"),
    system.NotificationBody("报告已经生成。"),
    system.NotificationKindStyle(system.NotificationInfo),
    system.NotificationGroup("reports"),
    system.NotificationTimeout(5*time.Second),
)
```

`Notify` 是非阻塞 API。Windows driver 会在提交给 shell 后返回；气泡显示时长由系统通知设置、专注助手和用户交互共同决定。

## 样式

`NotificationKindStyle` 控制通知的语义：

- `NotificationInfo`
- `NotificationSuccess`
- `NotificationWarning`
- `NotificationError`

不同平台可以用图标、强调色或系统通知模板表达这些语义。平台不支持某种视觉差异时，可以保持相同展示，但不应改变错误语义。

## 选项

- `NotificationTitle(value)`: 设置通知标题。
- `NotificationBody(value)`: 设置通知正文。
- `NotificationKindStyle(kind)`: 设置通知语义，默认 `NotificationInfo`。
- `NotificationIcon(path)`: 设置图标路径，平台 driver 可按系统要求解释。
- `NotificationGroup(group)`: 设置通知分组键，用于后续合并、替换或事件归类。
- `NotificationBackendPath(backend)`: 选择通知 backend，支持 `NotificationBackendAuto`、`NotificationBackendBalloon`、`NotificationBackendToast`。
- `NotificationAppID(appID)`: 设置 Windows Toast 使用的 AppUserModelID，默认 `FluxUI`。
- `NotificationLaunchURI(uri)`: 设置 Toast 主体点击时使用的协议 URI。Windows 下映射为 `activationType="protocol"`，调用方需要自行注册 URI scheme。
- `NotificationActions(actions...)`: 设置 Toast action buttons。配置 action 后 Windows 会要求 Toast 路径；`NotificationAction.URI` 可让单个 action 使用协议 URI 激活。
- `NotificationTimeout(timeout)`: 设置期望展示时长；系统可能按平台策略忽略。
- `NotificationOnClick(fn)`: 注册点击回调承载点，只有 driver 能稳定收到系统事件时才会调用。
- `NotificationOnDismiss(fn)`: 注册关闭/超时回调承载点，只有 driver 能稳定收到系统事件时才会调用。
- `NotificationOnAction(fn)`: 注册 action 回调承载点，只有 driver 能稳定收到 action activation 时才会调用。

Windows 下，同一个 `NotificationGroup(group)` 的新通知会在显示前清理前一个仍在显示或等待清理的托盘气泡和 Toast；Toast 路径也会设置 Toast `Group` / `Tag`，让系统通知历史按同组记录处理。若旧 Toast 正在通过短生命周期 helper 监听 click/action/dismiss，替换或取消该 group 时会同时停止旧 helper 进程。Linux `notify-send` provider 会在可用时使用 `--print-id` / `--replace-id` 做同组替换，并在主动取消时通过 `gdbus` 调用 freedesktop notification close；如果当前 provider 不支持 group 或 close 命令不可用，会返回 `ErrUnsupported` 或 `ErrUnavailable`。macOS `osascript` 和 Linux `kdialog` v1 不支持 group。需要主动取消时调用：

```go
_ = system.CancelNotificationGroup(ctx, "reports")
```

取消一个不存在的 group 是 no-op；空 group 会返回错误。

## Toast 与 Action Buttons

Windows Toast 示例：

```go
err := system.Notify(ctx,
    system.NotificationBackendPath(system.NotificationBackendToast),
    system.NotificationAppID("com.example.FluxUI"),
    system.NotificationTitle("导出完成"),
    system.NotificationBody("报告已经生成。"),
    system.NotificationGroup("reports"),
    system.NotificationLaunchURI("fluxui://notifications/reports"),
    system.NotificationActions(
        system.NotificationAction{ID: "open", Label: "打开"},
        system.NotificationAction{Label: "打开文件夹", URI: "fluxui://notifications/reports/open-folder"},
        system.NotificationAction{ID: "dismiss", Label: "忽略"},
    ),
)
```

未显式选择 backend 时，如果配置了 `NotificationActions`，Windows 会自动使用 Toast，因为托盘气泡不支持 action buttons。Toast 对未打包桌面应用通常要求可用的 AppUserModelID，通常由安装器或开始菜单快捷方式注册；如果系统拒绝 AppID 或 Toast 服务不可用，`Notify` 会返回包装后的 `ErrUnavailable`。当前实现能显示 Toast 和 action buttons；如果配置了 `NotificationOnClick`、`NotificationOnDismiss` 或 `NotificationOnAction`，Windows driver 会保留一个短生命周期 Toast 监听进程，在它存活期间把 Activated / Dismissed 事件转回 Go 回调。可用性和能力边界可通过 `ProbeNotificationBackend(ctx, NotificationBackendToast, NotificationAppID(...))` 预先探测。

未打包 Windows 桌面应用可用 `RegisterToastShortcut` 为当前用户创建或更新开始菜单快捷方式，并写入 Toast 使用的 AppUserModelID：

```go
err := system.RegisterToastShortcut(ctx,
    "com.example.FluxUI",
    "FluxUI",
    `C:\Program Files\FluxUI\fluxui.exe`,
    system.ToastShortcutArguments("--tray"),
    system.ToastShortcutIcon(`C:\Program Files\FluxUI\fluxui.exe,0`),
    system.ToastShortcutActivatorCLSID("{01234567-89AB-CDEF-0123-456789ABCDEF}"),
)
```

`shortcutName` 可以省略 `.lnk` 后缀；`ToastShortcutIcon` 支持普通图标路径或 `path,index` 资源格式；`ToastShortcutActivatorCLSID` 可把 `AppUserModel.ToastActivatorCLSID` 写入 shortcut。卸载或清理时调用 `UnregisterToastShortcut(ctx, "FluxUI")`。这个 API 只负责 shortcut 属性注册。

要让通知中心延迟点击、进程外启动或 helper 退出后的 Toast activation 回到应用，需要注册并运行同一个 CLSID 的 COM LocalServer：

```go
const toastActivatorCLSID = "{01234567-89AB-CDEF-0123-456789ABCDEF}"

_ = system.RegisterToastActivator(ctx,
    toastActivatorCLSID,
    `"C:\Program Files\FluxUI\fluxui.exe" --toast-activator`,
)

err := system.RunToastActivator(ctx, toastActivatorCLSID, func(event system.ToastActivationEvent) {
    // event.AppID / event.Arguments / event.UserInput 来自 Windows Toast activation。
})
```

`RegisterToastActivator` 写入当前用户 HKCU `Software\Classes\CLSID\{...}\LocalServer32`，不需要管理员权限。应用通常在安装或首次启动时注册 shortcut 和 LocalServer32，在 `--toast-activator` 启动模式中调用 `RunToastActivator`。`StartToastActivator` 适合应用主进程已经运行时注册 class object；它返回的 `ToastActivator` 需要在退出时 `Close()`。

`NotificationLaunchURI` 和 `NotificationAction.URI` 使用 Windows Toast protocol activation。只要系统中已注册对应 URI scheme，Toast 进入通知中心后或 helper 监听进程退出后，用户点击主体或对应 action 仍可由 Windows 打开该 URI。协议激活不会自动调用当前进程里的 `NotificationOnClick` / `NotificationOnAction`；应用需要在 URI handler 或二次启动参数里接收并转发。

## 事件边界

Notification API 预留了三类事件：

- `NotificationEventClicked`: 用户点击通知主体。
- `NotificationEventDismissed`: 用户关闭或系统移除通知。
- `NotificationEventAction`: 用户点击通知动作按钮。

Windows 托盘气泡会在收到 `NIN_BALLOONUSERCLICK` 时触发 `NotificationOnClick`，并传入 `NotificationEventClicked`；收到 `NIN_BALLOONHIDE` 或 `NIN_BALLOONTIMEOUT` 时会触发 `NotificationOnDismiss`，并传入 `NotificationEventDismissed`。Toast 路径在配置任一回调时，会通过 PowerShell 监听当前 Toast 对象的 Activated / Dismissed 事件：主体点击触发 `NotificationOnClick`，foreground action button 触发 `NotificationOnAction` 并填充 `NotificationEvent.Action`，dismissed 触发 `NotificationOnDismiss`。该监听覆盖当前 helper 进程存活期间的实时事件；协议 URI action 会交给系统 URI handler，不进入 Go 回调。通知中心延迟点击、helper 退出后仍要回到应用时，使用 `RegisterToastActivator` + `RunToastActivator` 接收 `ToastActivationEvent`。系统不保证每一种关闭、超时或通知中心行为都会回调，调用方不能把 callback 当作必达回调。

## Windows 展示路径策略

Toast 路径已提供请求、action button XML 构造、协议 URI 激活、短生命周期实时事件监听、AppUserModelID shortcut 注册、ToastActivatorCLSID shortcut 属性写入、COM LocalServer 注册、当前进程 `INotificationActivationCallback` 分发和 backend 级探测。协议 URI 激活可以覆盖通知中心延迟点击后的 OS 级启动/打开；COM activator 可把 durable activation 转回应用进程。`ProbeNotificationBackend` 中的 `SupportsDurableActivation` 在 Windows Toast backend 可用时为 true。托盘气泡路径仍是默认路径；长期 Tray 图标、菜单和退出清理由 Tray API 单独承载。

## macOS / Linux 展示路径策略

macOS v1 通过 `osascript display notification` 提交 title/body，并把 warning/error 等语义作为 subtitle 提示；不支持 action buttons、protocol activation、click/dismiss/action callback、group replace 或 group cancel。Linux v1 优先使用 `notify-send`，支持 title/body/icon/timeout/urgency，并在支持相关参数时为 group 记录 notification id；`kdialog --passivepopup` 作为基础 fallback，只支持简单 title/body/timeout。命令型 driver 不会伪造 callback，所有 callback capability 在 probe 中保持 false。

## 错误语义

- 当前平台不支持通知能力、请求了当前 driver 不支持的 backend 或子能力：返回包装后的 `ErrUnsupported`。
- 平台 driver 声明通知能力，但 Windows shell、通知区、隐藏窗口、图标加载或 macOS/Linux 平台命令当前不可用：返回包装后的 `ErrUnavailable`。
- `context.Context` 已取消：提交前直接返回 context 错误。

判断错误时使用 helper：

```go
if system.IsUnsupported(err) {
    // 禁用入口或改用应用内 Toast。
}

if system.IsUnavailable(err) {
    // 平台能力存在，但当前系统通知服务不可用。
}
```

Notification 不会自动 fallback 到 FluxUI 自绘 `Toast`。如果业务希望降级到应用内消息，应在业务层显式处理。

## 示例

`examples/system_showcase` 提供 Notification 的人工验收入口。示例会根据 `system.Supports(system.CapabilityNotification)` 和 `ProbeNotificationBackend` 显示能力状态：Windows 点击后应显示系统托盘气泡通知；macOS / Linux 在平台命令可用时应提交系统通知。示例还提供 Toast protocol activation 和 Toast backend 探测按钮；若系统通知区、Toast AppID 或平台命令不可用，示例会展示 `ErrUnavailable` 对应的不可用状态。

## 验收

Phase E v1 自动化验收覆盖：

- option 默认值和 option 透传。
- driver 分发和能力缺失。
- context 已取消时不调用 driver。
- Windows notification data 结构、图标标志、超时映射、click/dismiss callback option、Toast protocol activation XML、Toast group cancel script 和同组取消/替换语义。
- macOS / Linux notification provider 选择、命令构造、unsupported 子能力、context 已取消和交叉编译。

Windows 本地人工验收需要确认中文标题、正文显示，以及点击通知主体时可触发 `NotificationOnClick`。系统专注助手或通知设置可能会抑制气泡显示。

Toast 本地人工验收需要额外确认：`examples/system_validation -toast-shortcut -toast-app-id com.example.FluxUI` 可以创建并清理当前用户开始菜单 AppUserModelID shortcut；追加 `-toast-activator -toast-activator-clsid {01234567-89AB-CDEF-0123-456789ABCDEF}` 可以创建并清理 HKCU COM LocalServer32 注册；显式 Toast backend 可以显示；action buttons 可见；helper 监听进程存活期间主体点击、foreground action 点击和 dismiss 能触发对应回调；配置 `NotificationLaunchURI` / `NotificationAction.URI` 后，已注册的 URI scheme 能在通知中心延迟点击时被 Windows 打开。完整通知中心点击到应用业务流程仍需按具体应用启动参数和 `RunToastActivator` 模式做人工点验。
