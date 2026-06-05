<!-- fluxui-doc-meta
{
  "id": "system_api",
  "title": "System API 系统能力",
  "category": "使用指南",
  "order": 122,
  "summary": "System API 提供窗口、文件选择、系统弹窗、通知、托盘等操作系统能力的统一入口。",
  "apis": [
    "system.Capabilities() system.CapabilitySet",
    "system.Supports(cap system.Capability) bool",
    "system.OpenFileDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.OpenFilesDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.SaveFileDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.PickFolderDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.FileDialogOwner(owner uintptr) system.FileDialogOption",
    "system.ShowMessageBox(ctx context.Context, opts ...system.MessageBoxOption) (system.MessageBoxResult, error)",
    "system.MessageBoxTitle(value string) system.MessageBoxOption",
    "system.MessageBoxText(value string) system.MessageBoxOption",
    "system.MessageBoxStyle(kind system.MessageBoxKind) system.MessageBoxOption",
    "system.MessageBoxButtonSet(buttons system.MessageBoxButtons) system.MessageBoxOption",
    "system.MessageBoxOwner(owner uintptr) system.MessageBoxOption",
    "system.Notify(ctx context.Context, opts ...system.NotificationOption) error",
    "system.NotificationTitle(value string) system.NotificationOption",
    "system.NotificationBody(value string) system.NotificationOption",
    "system.NotificationKindStyle(kind system.NotificationKind) system.NotificationOption",
    "system.NotificationOnClick(fn func(system.NotificationEvent)) system.NotificationOption",
    "system.ErrUnsupported",
    "system.ErrUnavailable",
    "system.IsUnsupported(err error) bool",
    "system.IsUnavailable(err error) bool",
    "WindowHandle.NativeHandle() (uintptr, bool)",
    "ui.OpenFileDialog(ctx *ui.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "ui.ShowMessageBox(ctx *ui.Context, opts ...system.MessageBoxOption) (system.MessageBoxResult, error)"
  ]
}
-->

# System API 系统能力

`system` 包是 FluxUI 渲染层之外的操作系统能力边界。它用于承载窗口、文件选择、系统弹窗、系统通知、托盘、剪贴板和 shell open 等能力。

当前阶段已经完成能力枚举、能力检测、错误语义、平台 driver 边界，以及 Windows File Dialog、Windows MessageBox 和 Windows Notification 的第一版原生实现。通知当前使用 `Shell_NotifyIconW` 托盘气泡路径。托盘完整生命周期 API 会在后续阶段逐步实现。

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

## 能力列表

当前公共能力枚举包括：

- `CapabilityWindow`: 窗口控制和窗口状态。
- `CapabilityFileDialog`: 系统原生文件选择和保存对话框。
- `CapabilityMessageBox`: 系统原生消息框。
- `CapabilityNotification`: 系统通知。
- `CapabilityTray`: 任务栏托盘。
- `CapabilityClipboard`: 系统剪贴板。
- `CapabilityShell`: 通过系统默认程序打开 URL、文件或目录。

Windows driver 当前声明 `CapabilityWindow`、`CapabilityFileDialog`、`CapabilityMessageBox` 和 `CapabilityNotification`。Notification 使用临时托盘图标提交系统气泡；如果 Windows shell、通知区或图标加载不可用，`Notify` 会返回包装后的 `ErrUnavailable`。

非 Windows 平台当前使用 unsupported driver，能力表为空。后续 macOS、Linux、移动端和 Web 可在不改变公共 API 的前提下补充各自 driver。

## 文件选择

Windows 平台已经提供原生 Common Item Dialog 封装：

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
- `FileDialogFilters(filters...)`: 设置扩展名过滤器，`png`、`.png` 和 `*.png` 都会映射为 `*.png`。
- `FileDialogOwner(owner)`: 设置原生 owner 窗口句柄。Windows 下解释为 `HWND`，通常由 `ui.OpenFileDialog` 自动注入；传 0 表示无 owner。
- `FileDialogAllowCreateDirs(allow)`: 控制保存时是否允许系统提示创建目录。
- `FileDialogAllowMissingPath(allow)`: 控制打开/保存/选目录时是否强制路径已存在。
- `FileDialogOverwritePrompt(prompt)`: 控制保存到已有文件时是否显示覆盖确认。

取消选择返回 `FileDialogResult{Cancelled: true}` 且 `err == nil`。成功选择时 `Paths` 返回绝对路径，多选结果保持系统返回顺序。

`ui.OpenFileDialog`、`ui.OpenFilesDialog`、`ui.SaveFileDialog` 和 `ui.PickFolderDialog` 会从当前 `*ui.Context` 自动取得 `WindowHandle.NativeHandle()` 并注入 owner；直接调用 `system` 包时才需要显式传 `FileDialogOwner(owner)`。

`examples/system_showcase` 提供 File Dialog 的人工验收入口，示例会根据 `system.Supports(system.CapabilityFileDialog)` 禁用按钮，并通过 `ui` wrapper 自动把当前 FluxUI 窗口的 native owner 传给系统对话框。

第一版限制：

- Windows 下可通过 `WindowHandle.NativeHandle()` 获取 HWND；`ui` wrapper 会自动处理这一步。
- `system` 包没有 `*ui.Context`，因此不会自动推导当前窗口；直接调用 `system` 包且需要 modal owner 时由调用方显式传入 owner。
- 未传 `FileDialogOwner` 或 owner 为 0 时使用无 owner 的系统对话框。
- 调用会阻塞到用户确认或取消，应从事件处理逻辑或 goroutine 调用，不要放在布局函数中。
- `context.Context` 会在打开前检查；原生窗口显示后，当前版本不强制关闭系统对话框。

## 系统消息框

Windows 平台已经提供原生消息框封装，优先使用较新的 `TaskDialog`：

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
- `MessageBoxOwner(owner)`: 设置原生 owner 窗口句柄。Windows 下解释为 `HWND`，通常由 `ui.ShowMessageBox` 自动注入；传 0 表示无 owner。

Windows v1 不能稳定区分点击 Cancel、按 Escape 和点击关闭按钮。只要系统返回 `IDCANCEL`，FluxUI 都返回 `MessageBoxResultCancel`。`MessageBoxResultClose` 保留给后续能区分关闭动作的平台 driver。

`ui.ShowMessageBox` 会从当前 `*ui.Context` 自动取得 `WindowHandle.NativeHandle()` 并注入 owner；直接调用 `system.ShowMessageBox` 时才需要显式传 `MessageBoxOwner(owner)`。`examples/system_showcase` 已改用 `ui.ShowMessageBoxContext` 自动传 owner。带 owner 的 Windows 消息框关闭前，主窗口不能正常切回交互焦点。

第一版限制：

- Windows 下可通过 `WindowHandle.NativeHandle()` 获取 HWND；`ui` wrapper 会自动处理这一步。
- `system` 包没有 `*ui.Context`，因此不会自动推导当前窗口；直接调用 `system` 包且需要 modal owner 时由调用方显式传入 owner。
- 未传 `MessageBoxOwner` 或 owner 为 0 时使用无 owner 的系统消息框。
- Windows 优先使用 `TaskDialog`。如果应用没有启用 common controls v6，系统可能不提供 Task Dialog，此时会回退到传统 `MessageBoxW`。
- 调用会阻塞到用户选择按钮，应从事件处理逻辑或 goroutine 调用，不要放在布局函数中。
- `context.Context` 会在打开前检查；原生窗口显示后，当前版本不强制关闭系统消息框。

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
- `NotificationTimeout(timeout)`: 设置期望展示时长，平台可能忽略。
- `NotificationOnClick(fn)`: 注册点击回调；Windows 托盘气泡收到 `NIN_BALLOONUSERCLICK` 时会触发。

当前 Windows v1 不直接实现 Toast Notification。Toast 涉及 AppUserModelID、开始菜单快捷方式、打包/非打包差异和激活回调生命周期，先不作为第一版承诺。当前使用托盘气泡通知；Phase F 会继续补齐完整 Tray 生命周期、菜单和退出清理。

系统通知不会 fallback 到 FluxUI 自绘 `Toast`。如果需要应用内降级提示，应在业务层显式调用自绘 Toast 或 Snackbar。

## 错误语义

系统能力统一使用两个基础错误：

```go
system.ErrUnsupported
system.ErrUnavailable
```

- `ErrUnsupported`: 当前平台或当前 driver 不支持该能力。
- `ErrUnavailable`: 能力理论上受支持，但当前环境暂时不可用，例如系统服务缺失、权限不足或依赖的托盘未创建。

判断错误时使用 helper，而不是直接比较错误值：

```go
if system.IsUnsupported(err) {
    // 当前平台不支持，可以禁用入口或显示降级说明。
}

if system.IsUnavailable(err) {
    // 平台支持，但当前环境不可用，可以提示用户检查配置。
}
```

这两个 helper 支持 `fmt.Errorf("...: %w", err)` 这类包装错误。

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

下一步建议推进完整 Tray API，补齐长期托盘图标、菜单、退出清理和 Explorer 重启恢复策略。
