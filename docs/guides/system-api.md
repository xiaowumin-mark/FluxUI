<!-- fluxui-doc-meta
{
  "id": "system_api",
  "title": "System API 系统能力",
  "category": "使用指南",
  "order": 122,
  "summary": "System API 提供窗口、文件选择、系统弹窗、通知、托盘、剪贴板、Shell 和拖放探测等操作系统能力入口。",
  "example": { "id": "system_api_basic" },
  "apis": [
    "system.Capabilities() system.CapabilitySet",
    "system.Supports(cap system.Capability) bool",
    "system.Probe(cap system.Capability) system.CapabilityAvailability",
    "system.Availability(cap system.Capability) system.CapabilityAvailability",
    "system.OpenFileDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.OpenFilesDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.SaveFileDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.PickFolderDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.FileDialogOwner(owner uintptr) system.FileDialogOption",
    "system.FileDialogDefaultExtension(value string) system.FileDialogOption",
    "system.FileDialogRememberDir(key string) system.FileDialogOption",
    "system.ShowMessageBox(ctx context.Context, opts ...system.MessageBoxOption) (system.MessageBoxResult, error)",
    "system.ShowMessageBoxDetailed(ctx context.Context, opts ...system.MessageBoxOption) (system.MessageBoxDetailedResult, error)",
    "system.ShowMessageBoxAsync(ctx context.Context, opts ...system.MessageBoxOption) <-chan system.MessageBoxResponse",
    "system.ShowMessageBoxDetailedAsync(ctx context.Context, opts ...system.MessageBoxOption) <-chan system.MessageBoxDetailedResponse",
    "system.MessageBoxTitle(value string) system.MessageBoxOption",
    "system.MessageBoxText(value string) system.MessageBoxOption",
    "system.MessageBoxStyle(kind system.MessageBoxKind) system.MessageBoxOption",
    "system.MessageBoxButtonSet(buttons system.MessageBoxButtons) system.MessageBoxOption",
    "system.MessageBoxCustomButtons(buttons ...system.MessageBoxButton) system.MessageBoxOption",
    "system.MessageBoxVerification(label string, checked bool) system.MessageBoxOption",
    "system.MessageBoxOwner(owner uintptr) system.MessageBoxOption",
    "system.Notify(ctx context.Context, opts ...system.NotificationOption) error",
    "system.NotificationTitle(value string) system.NotificationOption",
    "system.NotificationBody(value string) system.NotificationOption",
    "system.NotificationKindStyle(kind system.NotificationKind) system.NotificationOption",
    "system.NotificationBackendPath(backend system.NotificationBackend) system.NotificationOption",
    "system.NotificationAppID(appID string) system.NotificationOption",
    "system.NotificationLaunchURI(uri string) system.NotificationOption",
    "system.NotificationActions(actions ...system.NotificationAction) system.NotificationOption",
    "system.NotificationOnClick(fn func(system.NotificationEvent)) system.NotificationOption",
    "system.NotificationOnDismiss(fn func(system.NotificationEvent)) system.NotificationOption",
    "system.NotificationOnAction(fn func(system.NotificationEvent)) system.NotificationOption",
    "system.ProbeNotificationBackend(ctx context.Context, backend system.NotificationBackend, opts ...system.NotificationOption) system.NotificationBackendProbe",
    "system.CancelNotificationGroup(ctx context.Context, group string) error",
    "system.SubscribeSystemEvents(ctx context.Context, kinds ...system.SystemEventKind) (*system.SystemEventSubscription, error)",
    "(*system.SystemEventSubscription).Events() <-chan system.SystemEvent",
    "(*system.SystemEventSubscription).Close() error",
    "system.ReadClipboardText(ctx context.Context) (string, error)",
    "system.WriteClipboardText(ctx context.Context, text string) error",
    "system.ReadClipboardFiles(ctx context.Context) ([]string, error)",
    "system.WriteClipboardFiles(ctx context.Context, paths []string) error",
    "system.ReadClipboardImagePNG(ctx context.Context) ([]byte, error)",
    "system.WriteClipboardImagePNG(ctx context.Context, data []byte) error",
    "system.OpenURL(ctx context.Context, target string) error",
    "system.OpenPath(ctx context.Context, path string) error",
    "system.RevealPath(ctx context.Context, path string) error",
    "system.ProbeDragAndDrop(ctx context.Context) system.DragAndDropProbe",
    "system.AcquireSingleInstance(ctx context.Context, opts ...system.SingleInstanceOption) (*system.SingleInstance, error)",
    "system.RegisterProtocolHandler(ctx context.Context, scheme, command string, opts ...system.RegistrationOption) error",
    "system.RegisterFileAssociation(ctx context.Context, extension, progID, command string, opts ...system.RegistrationOption) error",
    "system.RegisterStartupTask(ctx context.Context, name, command string) error",
    "system.RegisterToastShortcut(ctx context.Context, appID, shortcutName, executable string, opts ...system.ToastShortcutOption) error",
    "system.ToastShortcutActivatorCLSID(clsid string) system.ToastShortcutOption",
    "system.RegisterToastActivator(ctx context.Context, clsid, command string) error",
    "system.UnregisterToastActivator(ctx context.Context, clsid string) error",
    "system.StartToastActivator(ctx context.Context, clsid string, fn func(system.ToastActivationEvent)) (*system.ToastActivator, error)",
    "system.RunToastActivator(ctx context.Context, clsid string, fn func(system.ToastActivationEvent)) error",
    "system.UnregisterToastShortcut(ctx context.Context, shortcutName string) error",
    "system.RegisterGlobalShortcut(ctx context.Context, spec system.GlobalShortcutSpec, fn func(system.GlobalShortcutEvent)) (*system.GlobalShortcut, error)",
    "system.NewTray(opts ...system.TrayOption) (*system.Tray, error)",
    "system.TrayIcon(path string) system.TrayOption",
    "system.TrayIconBytes(data []byte) system.TrayOption",
    "system.TrayIconResource(id uint16) system.TrayOption",
    "system.TrayTooltip(text string) system.TrayOption",
    "system.TrayMenuItems(items ...system.TrayMenuItem) system.TrayOption",
    "system.TrayMenuProvider(fn func() system.TrayMenu) system.TrayOption",
    "system.TrayMenuAction(id, label string, onClick func(system.TrayEvent)) system.TrayMenuItem",
    "(*system.Tray).Show() error",
    "(*system.Tray).Hide() error",
    "(*system.Tray).Visible() bool",
    "(*system.Tray).Closed() bool",
    "(*system.Tray).Close() error",
    "system.CloseTrays() error",
    "system.ErrUnsupported",
    "system.ErrUnavailable",
    "system.ErrClosed",
    "system.ErrAlreadyRunning",
    "system.ErrInvalidTarget",
    "system.ErrTargetNotFound",
    "system.ErrNoDefaultHandler",
    "system.ErrAccessDenied",
    "system.IsUnsupported(err error) bool",
    "system.IsUnavailable(err error) bool",
    "system.IsClosed(err error) bool",
    "system.IsAlreadyRunning(err error) bool",
    "system.IsInvalidTarget(err error) bool",
    "system.IsTargetNotFound(err error) bool",
    "system.IsNoDefaultHandler(err error) bool",
    "system.IsAccessDenied(err error) bool",
    "WindowHandle.NativeHandle() (uintptr, bool)",
    "WindowHandle.SetAlwaysOnTop(always bool) bool",
    "WindowHandle.SetHiddenMemoryPolicy(policy WindowHiddenMemoryPolicy) bool",
    "WindowHandle.SubscribeEvents(kinds ...WindowEventKind) (*WindowEventSubscription, bool)",
    "WindowHandle.RequestFocus() bool",
    "WindowHandle.SetPosition(x, y int) bool",
    "WindowHandle.SetResizable(resizable bool) bool",
    "WindowHandle.SetDecorated(decorated bool) bool",
    "WindowHandle.SetCloseRequestedHandler(fn func(WindowCloseRequest) bool) bool",
    "WindowHandle.Show() bool",
    "WindowHandle.Hide() bool",
    "ui.HiddenMemoryPolicy(policy WindowHiddenMemoryPolicy) AppOption",
    "ui.OpenFileDialog(ctx *ui.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "ui.ShowMessageBox(ctx *ui.Context, opts ...system.MessageBoxOption) (system.MessageBoxResult, error)"
  ]
}
-->

# System API 系统能力

`system` 包是 FluxUI 渲染层之外的操作系统能力边界。它用于承载窗口、文件选择、系统弹窗、系统通知、托盘、剪贴板、shell open 和拖放能力探测等能力。

当前阶段已经完成能力枚举、能力检测、错误语义、平台 driver 边界，以及 Windows 窗口、File Dialog、MessageBox、Notification、Tray、系统事件订阅、文本剪贴板、Shell open 和 Drag & Drop probe 的第一版实现；macOS / Linux 已补 File Dialog、Clipboard 文本、Shell open、MessageBox、Notification、System Events、Tray 和 Drag & Drop probe 最小 driver，其中 macOS Tray 使用 `osascript` + AppKit `NSStatusItem`，Linux Tray 使用 `yad --notification`。窗口能力包括 native owner、关闭请求拦截、持续置顶、显示/隐藏、隐藏内存策略和全屏尺寸约束处理；Windows 通知默认使用 `Shell_NotifyIconW` 托盘气泡路径，支持同组替换/取消，显式 Toast backend 支持 action buttons 和短生命周期实时 click/action/dismiss 回调；macOS / Linux 通知通过 `osascript`、`notify-send` 或 `kdialog` 提交最小系统通知；托盘已支持长期图标、bytes/resource 图标、右键菜单、显示隐藏、关闭清理和 Explorer 重启恢复。

## 能力检测

调用系统能力前，优先通过 `system.Supports` 判断当前平台是否支持：

```go
if !system.Supports(system.CapabilityFileDialog) {
    return
}
```

也可以一次性读取当前平台能力表：

```go
caps := system.Capabilities()
if caps.Supports(system.CapabilityWindow) {
    // 当前平台 driver 声明支持窗口能力。
}
```

`Capabilities()` 返回的是副本，业务代码修改返回值不会影响 FluxUI 内部能力表。

需要区分“不支持”和“当前不可用”时，使用 `system.Probe`：

```go
availability := system.Probe(system.CapabilityTray)
if !availability.Supported() {
    return
}
if !availability.Available() {
    // 平台支持，但当前系统资源暂时不可用。
    return
}
```

Windows driver 会对 File Dialog COM 创建、MessageBox native 入口、通知/托盘/系统事件隐藏窗口等做轻量真实探测；macOS / Linux driver 会对 File Dialog、Clipboard、Shell、MessageBox、Notification 和 System Events 所需平台命令或系统路径做探测。探测失败会返回 `CapabilityStatusUnavailable` 并携带可被 `system.IsUnavailable` 识别的错误。其他没有更深层 driver 检查的平台或能力，会把已声明能力视为 available；具体调用仍可能因为系统服务、权限或 native 资源失败返回 `ErrUnavailable`。

## 能力列表

当前公共能力枚举包括：

- `CapabilityWindow`: 窗口控制和窗口状态。
- `CapabilityFileDialog`: 系统原生文件选择和保存对话框。
- `CapabilityMessageBox`: 系统原生消息框。
- `CapabilityNotification`: 系统通知。
- `CapabilityTray`: 任务栏托盘。
- `CapabilitySystemEvents`: 系统显示、主题、设置、电源、会话等事件订阅。
- `CapabilityClipboard`: 系统剪贴板。
- `CapabilityShell`: 通过系统默认程序打开 URL、文件或目录。
- `CapabilityDragAndDrop`: Gio transfer 拖放能力探测。

Windows driver 当前声明 `CapabilityWindow`、`CapabilityFileDialog`、`CapabilityMessageBox`、`CapabilityNotification`、`CapabilityTray`、`CapabilitySystemEvents`、`CapabilityClipboard`、`CapabilityShell` 和 `CapabilityDragAndDrop`。Notification 使用临时托盘图标提交系统气泡；Tray 使用长期托盘图标和独立隐藏消息窗口接收图标、菜单事件；SystemEvents 使用独立隐藏消息窗口接收可到达的 Windows 系统消息；Clipboard 通过 Windows PowerShell Clipboard cmdlets 读写系统文本剪贴板；Shell 使用 `ShellExecuteW` 和 Explorer 定位路径；Drag & Drop probe 描述 FluxUI 基于 Gio `io/transfer` 暴露的拖入和应用内拖拽源/接收能力；跨应用拖出只按 `SupportsExternalDragOut` 单独报告，当前默认不承诺。如果 Windows shell、通知区或图标加载不可用，相关 API 会返回包装后的 `ErrUnavailable`。

macOS 和 Linux 当前都声明 `CapabilityClipboard`、`CapabilityFileDialog`、`CapabilityMessageBox`、`CapabilityNotification`、`CapabilitySystemEvents`、`CapabilityTray`、`CapabilityShell` 和 `CapabilityDragAndDrop`。macOS / Linux 使用平台命令或系统路径提交剪贴板、文件选择、系统消息框、系统通知、系统事件订阅和打开请求；macOS Tray 使用 `osascript` + AppKit `NSStatusItem`，Linux Tray 使用 `yad --notification`。Drag & Drop probe 保持与 widget/ui 层 transfer API 对齐，不额外创建原生拖放服务。移动端和 Web 后续可在不改变公共 API 的前提下补充各自 driver。

## 窗口能力

窗口生命周期和窗口控制由 `app` runtime 承载，并通过 `ui` 层暴露当前窗口 helper。`WindowRaise` / `WindowHandle.Raise()` 是一次性前置请求；需要持续置顶时使用 `WindowSetAlwaysOnTop(ctx, true)` 或 `WindowHandle.SetAlwaysOnTop(true)`，取消时传 `false`。

Windows 下 `SetAlwaysOnTop`、`Show`、`Hide`、`RequestFocus`、`SetPosition` 和 `SetResizable` 使用 Gio 捕获到的 native `HWND` 实现，并异步排队执行，避免在 Gio 帧处理期间同步阻塞窗口线程。`SetDecorated` 使用 Gio `Decorated` option。窗口尚未完成 native view 初始化、窗口已关闭或非 Windows 平台不可用时，native 相关操作返回 `false`。隐藏当前唯一窗口后，应用应保留 `WindowHandle`，通过托盘、计时器、另一个窗口或后台事件调用 `Show()`。

默认隐藏内存策略是 `WindowHiddenMemoryReleaseTransient`：隐藏后暂停该窗口的 FluxUI root layout 和 redraw invalidation，清空临时 `op.Ops` 并异步触发 Go 内存回收。`WindowHiddenMemoryKeepRenderingState` 可用于保留隐藏窗口的渲染状态。当前不直接释放 Gio 内部 GPU context；这需要 Gio 公开释放/重建渲染后端的 API。

全屏会临时移除 Gio 层当前窗口的 `MinSize` / `MaxSize` 限制，避免设置最大尺寸后全屏不能占满显示器；退出全屏或回到普通窗口模式时会重新应用应用请求的尺寸约束。设置完整 `MaxSize(width, height)` 后，最大化能力会被视为不可用，API 最大化调用返回 `false`，Windows 标题栏最大化按钮和系统菜单最大化项会同步禁用。

关闭请求拦截通过 `WindowHandle.SetCloseRequestedHandler(fn)` 或 `ui.OnCloseRequested(fn)` 配置。Windows 下 FluxUI 会在捕获 `HWND` 后安装 `WM_CLOSE` hook，handler 返回 `false` 可取消标题栏关闭、Alt+F4 和系统菜单关闭；每次请求会产生 `WindowEventCloseRequested`。非 Windows 平台当前不承诺可取消系统关闭。

关闭前确认应先在 handler 里快速返回 `false`，再异步打开 `system.ShowMessageBox` 或应用内确认 UI；用户确认后再调用 `WindowHandle.Close()`。`examples/window_showcase` 已按这个模式演示“关闭前是否保存”。

窗口事件既可以用 `WindowHandle.PollEvents()` 拉取，也可以用 `WindowHandle.SubscribeEvents(kinds...)` 订阅。订阅 channel 是 best-effort 固定缓冲；窗口关闭并从 registry 移除后会自动关闭。窗口级 DPI/scale 变化通过 `WindowEventScaleChanged` 和 `WindowState.Scale` / `DPI` 报告；显示器、系统主题、电源和会话变化属于 `system.SubscribeSystemEvents` 的系统级事件。

## 文件选择

Windows 平台已经提供原生 Common Item Dialog 封装；macOS / Linux 提供最小原生命令 driver：

```go
result, err := system.OpenFileDialog(ctx,
    system.FileDialogTitle("选择图片"),
    system.FileDialogDefaultDir("C:\\Users\\me\\Pictures"),
    system.FileDialogFilters(
        system.FileFilter{Name: "Images", Patterns: []string{"png", ".jpg", "*.webp"}},
        system.FileFilter{Name: "All files", Patterns: []string{"*.*"}},
    ),
)
if err != nil {
    if system.IsUnsupported(err) {
        return
    }
    // 处理真实错误。
}
if result.Cancelled {
    return
}
path := result.Paths[0]
```

公共入口包括：

- `OpenFileDialog`: 选择单个文件。
- `OpenFilesDialog`: 选择多个文件。
- `SaveFileDialog`: 选择保存目标。
- `PickFolderDialog`: 选择目录。

文件选择选项包括：

- `FileDialogTitle(value)`: 设置原生对话框标题。
- `FileDialogDefaultDir(value)`: 设置初始目录。
- `FileDialogDefaultName(value)`: 设置初始文件名。
- `FileDialogDefaultExtension(value)`: 设置默认扩展名，主要用于保存文件。
- `FileDialogFilters(filters...)`: 设置扩展名过滤器，`png`、`.png` 和 `*.png` 都会映射为 `*.png`。
- `FileDialogOwner(owner)`: 设置原生 owner 窗口句柄。Windows 下解释为 `HWND`，通常由 `ui.OpenFileDialog` 自动注入；传 0 表示无 owner。macOS / Linux 最小命令 driver 暂不支持非 0 owner。
- `FileDialogAllowCreateDirs(allow)`: 控制保存时是否允许系统提示创建目录。
- `FileDialogAllowMissingPath(allow)`: 控制打开/保存/选目录时是否强制路径已存在。
- `FileDialogOverwritePrompt(prompt)`: 控制保存到已有文件时是否显示覆盖确认。
- `FileDialogRememberDir(key)`: 记住指定 key 上一次成功选择的目录，并在下一次没有显式默认目录时自动使用。

取消选择返回 `FileDialogResult{Cancelled: true}` 且 `err == nil`。成功选择时 `Paths` 返回绝对路径，多选结果保持系统返回顺序。

`ui.OpenFileDialog`、`ui.OpenFilesDialog`、`ui.SaveFileDialog` 和 `ui.PickFolderDialog` 会从当前 `*ui.Context` 自动取得 `WindowHandle.NativeHandle()` 并注入 owner；直接调用 `system` 包时才需要显式传 `FileDialogOwner(owner)`。

macOS 使用 `osascript choose file/folder/file name`；Linux 优先使用 `zenity`，不可用时尝试 `kdialog`。这些最小 driver 支持打开文件、多选、保存文件、选择目录、取消结果、默认目录/文件名和基础过滤器；非 0 owner 返回 `ErrUnsupported`，缺少平台命令时返回 `ErrUnavailable`。

`examples/system_showcase` 提供 File Dialog 的人工验收入口，示例会根据 `system.Supports(system.CapabilityFileDialog)` 禁用按钮。Windows 下通过 `ui` wrapper 自动把当前 FluxUI 窗口的 native owner 传给系统对话框；macOS / Linux 最小 driver 暂不支持 owner。示例同时覆盖记忆目录、保存默认扩展名和显示后 context 自动取消。

第一版限制：

- Windows 下可通过 `WindowHandle.NativeHandle()` 获取 HWND；`ui` wrapper 会自动处理这一步。
- `system` 包没有 `*ui.Context`，因此不会自动推导当前窗口；直接调用 `system` 包且需要 modal owner 时由调用方显式传入 owner。
- 未传 `FileDialogOwner` 或 owner 为 0 时使用无 owner 的系统对话框。
- macOS / Linux 最小 driver 不支持 owner modal；更完整的 owner/modal、portal 和 toolkit 行为后续可用 AppKit、xdg-desktop-portal 或 GTK/Qt driver 继续补。
- 调用会阻塞到用户确认或取消，应从事件处理逻辑或 goroutine 调用，不要放在布局函数中。
- `context.Context` 会在打开前检查；Windows 下原生文件对话框显示后也会监听取消，并调用 `IFileDialog.Close` 尝试关闭系统对话框；macOS / Linux 通过 `exec.CommandContext` 终止平台命令，具体关闭行为取决于平台命令。

## 系统消息框

Windows 平台已经提供原生消息框封装，优先使用较新的 `TaskDialog`；macOS / Linux 提供最小原生命令 driver：

```go
result, err := system.ShowMessageBox(ctx,
    system.MessageBoxTitle("保存更改"),
    system.MessageBoxText("关闭前是否保存当前文档？"),
    system.MessageBoxStyle(system.MessageBoxQuestion),
    system.MessageBoxButtonSet(system.MessageBoxYesNoCancel),
    system.MessageBoxDefaultButton(system.MessageBoxResultCancel),
)
if err != nil {
    if system.IsUnsupported(err) {
        return
    }
    // 处理真实错误。
}
if result == system.MessageBoxResultCancel {
    return
}
```

消息框选项包括：

- `MessageBoxTitle(value)`: 设置系统消息框标题。
- `MessageBoxText(value)`: 设置正文。
- `MessageBoxStyle(kind)`: 设置样式，支持 info、warning、error、question。
- `MessageBoxButtonSet(buttons)`: 设置按钮集，支持 OK、OKCancel、YesNo、YesNoCancel、RetryCancel。
- `MessageBoxDefaultButton(result)`: 设置默认按钮，必须属于当前按钮集。
- `MessageBoxOwner(owner)`: 设置原生 owner 窗口句柄。Windows 下解释为 `HWND`，通常由 `ui.ShowMessageBox` 自动注入；传 0 表示无 owner。macOS / Linux 最小命令 driver 暂不支持非 0 owner。

Windows `TaskDialogIndirect` 路径会记录真实按钮点击和关闭来源：点击 Cancel 按钮返回 `MessageBoxResultCancel`；按 Escape 返回 `MessageBoxResultEscape`；右上角关闭按钮或 `WM_CLOSE` 路径返回 `MessageBoxResultClose`。传统 `TaskDialog` / `MessageBoxW` 回退路径没有可用的 dialog callback，仍会把 `IDCANCEL` 映射为 `MessageBoxResultCancel`。

`ui.ShowMessageBox` 会从当前 `*ui.Context` 自动取得 `WindowHandle.NativeHandle()` 并注入 owner；直接调用 `system.ShowMessageBox` 时才需要显式传 `MessageBoxOwner(owner)`。`examples/system_showcase` 已改用 `ui.ShowMessageBoxContext` 自动传 owner，并提供富消息框和显示后 context 自动取消入口。带 owner 的 Windows 消息框关闭前，主窗口不能正常切回交互焦点。

富 TaskDialog 能力使用 `TaskDialogIndirect`：`MessageBoxDetails`、`MessageBoxFooter`、`MessageBoxVerification`、`MessageBoxCustomButtons`、`MessageBoxDefaultButtonID`、`MessageBoxCommandLinks` 等选项可以提供详细信息、复选框、自定义按钮和 command links。需要读取复选框状态或自定义按钮 ID 时，使用 `ShowMessageBoxDetailed` / `ui.ShowMessageBoxDetailedContext`，返回 `MessageBoxDetailedResult`。如果使用了富选项而 `TaskDialogIndirect` 不可用，FluxUI 不会静默退回到会丢能力的 `MessageBoxW`。

macOS 使用 `osascript display dialog`；Linux 优先使用 `zenity`，不可用时尝试 `kdialog`。这些最小 driver 支持标题、正文、info/warning/error/question 样式和 OK、OKCancel、YesNo、YesNoCancel、RetryCancel 标准按钮集；富 TaskDialog 选项和 owner modal 返回 `ErrUnsupported`。缺少平台命令时返回 `ErrUnavailable`。

不想手动创建 goroutine 时，可以使用 `system.ShowMessageBoxAsync`、`system.ShowMessageBoxDetailedAsync` 或对应 `ui` wrapper，通过 response channel 接收结果。

第一版限制：

- Windows 下可通过 `WindowHandle.NativeHandle()` 获取 HWND；`ui` wrapper 会自动处理这一步。
- `system` 包没有 `*ui.Context`，因此不会自动推导当前窗口；直接调用 `system` 包且需要 modal owner 时由调用方显式传入 owner。
- 未传 `MessageBoxOwner` 或 owner 为 0 时使用无 owner 的系统消息框。
- Windows 优先使用 `TaskDialog`。如果应用没有启用 common controls v6，系统可能不提供 Task Dialog，此时会回退到传统 `MessageBoxW`。
- macOS / Linux 最小 driver 不支持 owner modal、verification、自定义按钮、command links 或 Escape/关闭来源区分；这些子能力后续可用 `NSAlert`、xdg portal 或桌面 toolkit driver 继续补。
- 调用会阻塞到用户选择按钮，应从事件处理逻辑或 goroutine 调用，不要放在布局函数中。
- `context.Context` 会在打开前检查；Windows `TaskDialogIndirect` 路径在消息框显示后也会监听取消，并向原生 dialog 投递关闭请求后优先返回 context 错误。
- `TaskDialog` / `MessageBoxW` 兼容回退路径会在调用期间锁定当前 OS 线程，context 取消后通过枚举该线程上的原生 dialog 做 best-effort 关闭；如果系统没有暴露可匹配窗口，则仍可能等待用户手动关闭。

## 系统通知

Notification API 提供非阻塞系统通知入口：

```go
err := system.Notify(ctx,
    system.NotificationTitle("导出完成"),
    system.NotificationBody("文件已经保存到输出目录。"),
    system.NotificationKindStyle(system.NotificationSuccess),
    system.NotificationGroup("exports"),
)
if err != nil {
    if system.IsUnsupported(err) {
        return
    }
    if system.IsUnavailable(err) {
        // Windows shell、通知区或图标加载当前不可用。
        return
    }
    // 处理真实错误。
}
```

通知选项包括：

- `NotificationTitle(value)`: 设置通知标题。
- `NotificationBody(value)`: 设置通知正文。
- `NotificationKindStyle(kind)`: 设置样式，支持 info、success、warning、error。
- `NotificationIcon(path)`: 设置图标路径。
- `NotificationGroup(group)`: 设置分组键。
- `NotificationBackendPath(backend)`: 选择 backend，支持 auto、balloon 和 toast。
- `NotificationAppID(appID)`: 设置 Windows Toast AppUserModelID，默认 `FluxUI`。
- `NotificationLaunchURI(uri)`: 设置 Toast 主体点击时使用的协议 URI，Windows 下映射为 protocol activation。
- `NotificationActions(actions...)`: 设置 Toast action buttons；配置 actions 后 Windows 会要求 Toast 路径，`NotificationAction.URI` 可让单个 action 使用协议 URI 激活。
- `NotificationTimeout(timeout)`: 设置期望展示时长，平台可能忽略。
- `NotificationOnClick(fn)`: 注册点击回调；Windows 托盘气泡收到 `NIN_BALLOONUSERCLICK` 时会触发，Toast helper 监听进程收到主体 Activated 时也会触发。
- `NotificationOnDismiss(fn)`: 注册关闭/超时回调；Windows 托盘气泡收到 `NIN_BALLOONHIDE` 或 `NIN_BALLOONTIMEOUT` 时会触发，Toast helper 监听进程收到 Dismissed 时也会触发。
- `NotificationOnAction(fn)`: 注册 action 回调；Windows Toast helper 监听进程收到 action Activated 时会触发并填充 `NotificationEvent.Action`。

Windows 下，同一个 `NotificationGroup(group)` 的新通知会在显示前清理前一个仍在显示或等待清理的托盘气泡和 Toast；Toast 路径也会设置 Toast `Group` / `Tag`。也可以调用 `CancelNotificationGroup(ctx, group)` 主动取消当前组通知。取消一个不存在的组是 no-op。

Windows 默认使用托盘气泡通知；显式 `NotificationBackendToast` 或配置 `NotificationActions` 时会构造 Toast XML 并调用 Windows Toast API。Toast 对未打包桌面应用通常要求可用的 AppUserModelID，通常由安装器或开始菜单快捷方式注册。可用性和能力边界可通过 `ProbeNotificationBackend(ctx, NotificationBackendToast, NotificationAppID(...))` 探测；Windows Toast backend 现在会报告 `SupportsDurableActivation=true`，表示 FluxUI 已提供当前进程 COM LocalServer 注册和 `INotificationActivationCallback` 分发 API。当前实现也会在配置回调时保留短生命周期 Toast 监听进程，用于接收当前 Toast 对象的实时 click/action/dismiss 事件；也支持通过 `NotificationLaunchURI` / `NotificationAction.URI` 走协议激活，让通知中心延迟点击由系统打开已注册 URI。

macOS 使用 `osascript display notification` 提交最小通知；Linux 优先使用 `notify-send`，不可用时尝试 `kdialog --passivepopup`。macOS/Linux v1 不支持 action buttons、protocol activation 或回调；Linux `notify-send` provider 在可用时支持 group replace，主动取消 group 需要 `gdbus` close command。缺少平台命令或系统 notification daemon 拒绝请求时返回包装后的 `ErrUnavailable`。

系统通知不会 fallback 到 FluxUI 自绘 `Toast`。如果需要应用内降级提示，应在业务层显式调用自绘 Toast 或 Snackbar。

## 系统事件

System Events API 提供系统级事件订阅：

```go
sub, err := system.SubscribeSystemEvents(ctx,
    system.SystemEventDisplayChanged,
    system.SystemEventThemeChanged,
)
if err != nil {
    return err
}
defer sub.Close()

go func() {
    for event := range sub.Events() {
        // 根据 event.Kind / event.Detail 更新应用状态。
    }
}()
```

事件类型包括：

- `SystemEventDisplayChanged`
- `SystemEventDPIChanged`
- `SystemEventThemeChanged`
- `SystemEventSettingsChanged`
- `SystemEventPowerChanged`
- `SystemEventSessionChanged`

不传 kind 时表示订阅 driver 能报告的全部事件。Windows driver 使用隐藏消息窗口接收 `WM_DISPLAYCHANGE`、`WM_THEMECHANGED`、`WM_SETTINGCHANGE`、`WM_POWERBROADCAST`、`WM_WTSSESSION_CHANGE` 和可到达的 `WM_DPICHANGED`。macOS / Linux 最小 driver 使用状态快照轮询：macOS 读取 `ioreg`、`defaults`、`pmset` 和 console session；Linux 读取 `/sys/class/drm`、`/sys/class/power_supply`，并在可用时使用 `gsettings` 和 `loginctl`。事件 channel 有固定缓冲；如果调用方长期不消费，driver 可以丢弃后续事件，避免阻塞 Win32 消息线程或 Unix 轮询 goroutine。

`examples/system_showcase` 提供 System Events 的人工验收入口，覆盖 30 秒限时订阅和最近事件显示。

## 托盘

Tray API 提供长期系统托盘图标：

```go
tray, err := system.NewTray(
    system.TrayTooltip("FluxUI"),
    system.TrayMenuItems(
        system.TrayMenuAction("show", "显示窗口", func(event system.TrayEvent) {
            // 显示主窗口。
        }),
        system.TrayMenuSeparator(),
        system.TrayMenuAction("quit", "退出", func(event system.TrayEvent) {
            // 关闭托盘并退出应用。
        }),
    ),
)
if err != nil {
    if system.IsUnsupported(err) {
        return
    }
    if system.IsUnavailable(err) {
        return
    }
    return
}
defer tray.Close()
_ = tray.Show()
```

`NewTray` 只创建托盘句柄，不自动显示图标。`Show()` 显示图标，`Hide()` 隐藏图标但保留句柄，`Close()` 释放 native 资源；关闭后继续操作会返回包装后的 `ErrClosed`。

托盘选项包括：

- `TrayIcon(path)`: 设置图标路径。
- `TrayIconBytes(data)`: 从内存中的 `.ico` bytes 设置图标。
- `TrayIconResource(id)`: 从当前进程资源中的 icon id 设置图标。
- `TrayTooltip(text)`: 设置 tooltip。
- `TrayMenuItems(items...)`: 设置右键菜单。
- `TrayMenuProvider(fn)`: 每次打开菜单前动态生成右键菜单。
- `TrayOnClick(fn)`: 注册左键点击回调。
- `TrayOnDoubleClick(fn)`: 注册左键双击回调。
- `TrayMenuAction(id, label, fn)`: 创建可点击菜单项。
- `TrayMenuSeparator()`: 创建菜单分隔线。

Windows driver 使用独立隐藏消息窗口接收托盘事件，不复用 Gio 主事件循环。右键菜单支持普通项、禁用项、勾选项、默认项、分隔线和子菜单；菜单点击事件会携带 `TrayEventMenuItem` 和 `ItemID`。Explorer 或任务栏重启后，仍处于 visible 状态的托盘图标会尝试自动恢复。

macOS v1 使用 `osascript` 运行 AppKit `NSStatusItem` 脚本。它支持显示、隐藏、关闭、tooltip、路径/bytes 图标、菜单/子菜单、禁用项、勾选项、无菜单左键点击和菜单项回调；缺少 `osascript`、桌面会话不允许 AppKit 状态栏项或脚本启动后立即退出时 `Probe(CapabilityTray)` / `NewTray` / `Show()` 返回 `ErrUnavailable`。macOS 命令型 driver 不支持 `TrayIconResource`，默认项视觉状态、每次菜单打开前实时刷新和双击事件不作为 v1 承诺。

Linux v1 使用 `yad --notification`。它支持显示、隐藏、关闭、tooltip、路径/bytes 图标、左键点击和菜单项回调；缺少 `yad` 或托盘 host 时 `Probe(CapabilityTray)` / `NewTray` 返回 `ErrUnavailable`。Linux 命令型 driver 不支持 `TrayIconResource`，子菜单会扁平化，`Checked` / `Default` 视觉状态和双击事件不作为 v1 承诺。

通过 `ui.RunElement` / `ui.RunElementMulti` 启动的应用会在 app lifecycle 正常退出时调用 `system.CloseTrays()`，统一清理仍然注册中的托盘。

`examples/system_showcase` 提供 Tray 的人工验收入口，覆盖创建、显示、隐藏、主窗口隐藏、右键菜单和关闭清理。

## 剪贴板与 Shell

Clipboard v1 提供文本读写，并在 Windows 下补充文件列表和 PNG 图片剪贴板：

```go
text, err := system.ReadClipboardText(ctx)
if err != nil {
    return err
}
err = system.WriteClipboardText(ctx, text+"\nupdated")

files, err := system.ReadClipboardFiles(ctx)
err = system.WriteClipboardFiles(ctx, []string{"C:\\Users\\me\\report.pdf"})

pngData, err := system.ReadClipboardImagePNG(ctx)
err = system.WriteClipboardImagePNG(ctx, pngData)
```

Windows driver 通过 `Get-Clipboard` / `Set-Clipboard` 读写系统文本剪贴板，通过 Windows Forms Clipboard 的 `GetFileDropList` / `SetFileDropList` 读写 Explorer 文件拖放列表，并通过 `ContainsImage` / `GetImage` / `SetImage` 读写图片剪贴板。`WriteClipboardFiles` 会在写入前确认路径存在并规范化为绝对路径；`ReadClipboardFiles` 在剪贴板没有文件列表时返回空切片。`ReadClipboardImagePNG` 会把当前图片转换为 PNG bytes，剪贴板没有图片时返回 `nil, nil`；`WriteClipboardImagePNG` 要求输入是有效 PNG bytes。macOS 使用 `pbpaste` / `pbcopy`；Linux 按可用性选择 `wl-clipboard`、`xclip` 或 `xsel`。macOS/Linux 当前只承诺文本，文件列表和图片剪贴板返回 `ErrUnsupported`。命令缺失、系统服务不可用或权限不足时返回包装后的 `ErrUnavailable`。

Shell v1 提供默认程序打开和文件管理器定位：

```go
_ = system.OpenURL(ctx, "https://example.com")
_ = system.OpenPath(ctx, "C:\\Users\\me\\report.pdf")
_ = system.RevealPath(ctx, "C:\\Users\\me\\report.pdf")
```

`OpenURL` 要求传入带 scheme 的 URL。`OpenPath` 用系统默认程序打开文件或目录；`RevealPath` 在系统文件管理器中定位目标。Windows driver 使用 `ShellExecuteW` / Explorer；macOS 使用 `open` / `open -R`；Linux 使用 `xdg-open`，文件 reveal 会打开父目录。Windows、macOS 和 Linux 都会在打开 `file:` URL、文件或定位路径前确认目标存在，目标已删除时直接返回包装后的 `ErrUnavailable` + `ErrTargetNotFound`，Windows 可避免 Explorer 弹出“位置不可用”的系统错误框。缺少平台命令或系统拒绝打开时返回包装后的 `ErrUnavailable`，本地路径权限拒绝会尽量同时包装 `ErrAccessDenied`。这些调用会交给操作系统启动外部程序，不会等待外部程序结束，也不会自动 fallback 到 FluxUI 自绘 UI。

## 单实例、系统注册与全局快捷键

单实例 API 用于让二次启动把参数转发给已经运行的主实例：

```go
instance, err := system.AcquireSingleInstance(ctx,
    system.SingleInstanceID("com.example.myapp"),
    system.SingleInstanceOnSecondLaunch(func(event system.SingleInstanceEvent) {
        // event.Args / event.Payload 来自二次启动。
    }),
)
if system.IsAlreadyRunning(err) {
    return nil // 当前进程是第二实例，已转发给主实例。
}
defer instance.Close()
```

Windows/macOS/Linux v1 使用本机 loopback coordination channel。它不需要安装器或管理员权限，适合普通桌面应用防重复启动、协议 URI 二次启动转发和文件关联打开转发。应用应传稳定的 reverse-DNS 风格 ID；未传时会从可执行文件名派生 best-effort ID。

Windows 系统注册 API 覆盖当前用户范围的 protocol handler、file association 和 startup task：

```go
_ = system.RegisterProtocolHandler(ctx, "fluxui", `"C:\Program Files\FluxUI\fluxui.exe" "%1"`)
_ = system.RegisterFileAssociation(ctx, ".flux", "FluxUI.Document", `"C:\Program Files\FluxUI\fluxui.exe" "%1"`)
_ = system.RegisterStartupTask(ctx, "FluxUI", `"C:\Program Files\FluxUI\fluxui.exe" --tray`)
_ = system.RegisterToastShortcut(ctx,
    "com.example.FluxUI",
    "FluxUI",
    `C:\Program Files\FluxUI\fluxui.exe`,
    system.ToastShortcutArguments("--tray"),
    system.ToastShortcutIcon(`C:\Program Files\FluxUI\fluxui.exe,0`),
    system.ToastShortcutActivatorCLSID("{01234567-89AB-CDEF-0123-456789ABCDEF}"),
)
_ = system.RegisterToastActivator(ctx,
    "{01234567-89AB-CDEF-0123-456789ABCDEF}",
    `"C:\Program Files\FluxUI\fluxui.exe" --toast-activator`,
)
```

这些 API 写入当前用户范围，不需要管理员权限：protocol/file association/startup 写入 HKCU，Toast shortcut 写入当前用户开始菜单 Programs 目录并设置 AppUserModelID，Toast activator 写入 HKCU `Software\Classes\CLSID\{...}\LocalServer32`。未传 `%1` 时 protocol/file association command 会自动追加 `" %1"`。macOS/Linux 当前不声明 `CapabilitySystemRegistration`，相关调用返回 `ErrUnsupported`。`RegisterToastShortcut` 解决的是未打包桌面应用显示 Toast 所需的 AppUserModelID shortcut 注册；`ToastShortcutActivatorCLSID` 写入 `AppUserModel.ToastActivatorCLSID`；`RegisterToastActivator` 注册启动命令；应用在该启动模式中调用 `StartToastActivator` 或 `RunToastActivator` 接收 `ToastActivationEvent`。

Windows 全局快捷键通过 `RegisterHotKey` 和隐藏消息窗口实现：

```go
shortcut, err := system.RegisterGlobalShortcut(ctx, system.GlobalShortcutSpec{
    ID:        "show",
    Key:       "F9",
    Modifiers: system.GlobalShortcutControl | system.GlobalShortcutAlt,
}, func(event system.GlobalShortcutEvent) {
    // 显示主窗口或打开命令面板。
})
defer shortcut.Close()
```

全局快捷键会占用系统级组合键；冲突、权限或桌面会话限制会返回 `ErrUnavailable`。macOS/Linux 当前不声明 `CapabilityGlobalShortcut`。

## 拖放能力探测

Drag & Drop 的实际收发由 `widget.DropTarget` / `ui.DropTargetElement` 和 `widget.DragSource` / `ui.DragSourceElement` 承担，因为拖放区域参与布局和输入命中测试。`system` 层只提供当前 driver 的能力探测：

```go
probe := system.ProbeDragAndDrop(ctx)
if !probe.Available() {
    return
}
if probe.SupportsDropTarget {
    // 可以启用 DropTarget / DropTargetElement。
}
if probe.SupportsDragSource {
    // 可以启用 DragSource / DragSourceElement。
}
```

`DragAndDropProbe` 会报告 drop target、drag source、文本、文件、自定义 MIME、外部拖入和应用级 operation 支持情况。`SupportsExternalDragOut` 只有在后端明确支持跨应用拖出时才会为 `true`；当前默认 probe 不承诺外部拖出。`SupportedOperations` 中的 copy/move/link 只表示 FluxUI 事件里可表达的应用级语义；当前 Gio `io/transfer` 不暴露 OS-level copy/move/link 协商字段，因此 FluxUI 不承诺能控制 Explorer、Finder、文件管理器或其它外部应用最终执行的拖放动作。

人工验证入口：

```sh
go run ./examples/system_validation -drag-drop
go run ./examples/drag_drop_showcase
```

`examples/system_validation -drag-drop` 只打印 probe，不打开 GUI；`examples/drag_drop_showcase` 用于手动拖入文本、文件和自定义 MIME payload。

## 错误语义

系统能力统一使用基础错误：

```go
system.ErrUnsupported
system.ErrUnavailable
system.ErrClosed
system.ErrAlreadyRunning
system.ErrInvalidTarget
system.ErrTargetNotFound
system.ErrNoDefaultHandler
system.ErrAccessDenied
```

- `ErrUnsupported`: 当前平台或当前 driver 不支持该能力。
- `ErrUnavailable`: 能力理论上受支持，但当前环境暂时不可用，例如系统服务缺失、权限不足或依赖的托盘未创建。
- `ErrClosed`: 长生命周期系统资源已经关闭，例如 `Tray.Close()` 后继续操作旧托盘句柄。
- `ErrAlreadyRunning`: `AcquireSingleInstance` 发现主实例已存在，并且当前启动数据已经转发给主实例。
- `ErrInvalidTarget`: URL、路径或图片数据等输入格式无效。
- `ErrTargetNotFound`: 本地打开目标不存在。
- `ErrNoDefaultHandler`: 系统没有为目标注册默认 handler。
- `ErrAccessDenied`: 系统策略或权限拒绝操作。

判断错误时使用 helper，而不是直接比较错误值：

```go
if system.IsUnsupported(err) {
    // 当前平台不支持，可以禁用入口或显示降级说明。
}

if system.IsUnavailable(err) {
    // 平台支持，但当前环境不可用，可以提示用户检查配置。
}

if system.IsClosed(err) {
    // 旧资源已经关闭，应丢弃句柄或重新创建。
}

if system.IsAlreadyRunning(err) {
    // 当前进程是第二实例，通常可以直接退出。
}

if system.IsTargetNotFound(err) {
    // 打开的本地路径不存在，可以提示用户重新选择。
}
```

这些 helper 支持 `fmt.Errorf("...: %w", err)` 这类包装错误。

## 分层约定

- `system` 包提供公共类型、能力检测、错误语义和后续系统函数。
- `app` 包继续负责窗口事件循环和窗口生命周期。
- `ui` 包后续只提供少量基于当前 `*ui.Context` 的便利封装。
- `widget` 包不承载文件选择、系统通知和托盘能力，因为这些能力不参与布局。

当前通过 `system` 包暴露底层系统能力；`ui.OpenFileDialog(ctx, ...)` 和 `ui.ShowMessageBox(ctx, ...)` 这类便利 API 只应从事件处理逻辑或 goroutine 调用，不要放进布局函数。

## 平台策略

FluxUI 的系统层采用 Windows first 策略：

- Windows 优先落地真实 driver。
- 非 Windows 平台必须可编译。
- 未实现平台返回明确的 `ErrUnsupported`。
- 不使用隐藏 fallback。系统弹窗、通知或托盘不可用时，不会自动退化为自绘 `Dialog` 或 `Toast`。

下一步建议继续补充 macOS/Linux 图片/文件列表剪贴板、Shell 默认 handler/桌面环境启动后的更细错误分类和更多平台原生 driver；Drag & Drop 接收和 FluxUI 应用内拖拽源 v1 已通过 `DropTarget` / `DragSource` 提供，真实跨应用拖入仍需按平台人工点验，跨应用拖出待后端明确支持后再打开 `SupportsExternalDragOut`。
