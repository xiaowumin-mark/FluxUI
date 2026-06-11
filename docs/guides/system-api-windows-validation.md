<!-- fluxui-doc-meta
{
  "id": "system_api_windows_validation",
  "title": "System API Windows 验收",
  "category": "使用指南",
  "order": 129,
  "summary": "记录 System API 在 Windows 上的自动化验证、人工验收清单和待补点验结果。",
  "example": { "id": "system_api_basic" },
  "apis": [
    "go test ./...",
    "go vet ./...",
    "go run ./examples/system_validation",
    "system.OpenFileDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.ShowMessageBox(ctx context.Context, opts ...system.MessageBoxOption) (system.MessageBoxResult, error)",
    "system.Notify(ctx context.Context, opts ...system.NotificationOption) error",
    "system.NewTray(opts ...system.TrayOption) (*system.Tray, error)",
    "system.SubscribeSystemEvents(ctx context.Context, kinds ...system.SystemEventKind) (*system.SystemEventSubscription, error)",
    "system.ReadClipboardImagePNG(ctx context.Context) ([]byte, error)",
    "system.WriteClipboardImagePNG(ctx context.Context, data []byte) error",
    "system.IsTargetNotFound(err error) bool",
    "system.RegisterToastShortcut(ctx context.Context, appID, shortcutName, executable string, opts ...system.ToastShortcutOption) error",
    "system.ToastShortcutActivatorCLSID(clsid string) system.ToastShortcutOption",
    "system.RegisterToastActivator(ctx context.Context, clsid, command string) error",
    "system.StartToastActivator(ctx context.Context, clsid string, fn func(system.ToastActivationEvent)) (*system.ToastActivator, error)",
    "system.RunToastActivator(ctx context.Context, clsid string, fn func(system.ToastActivationEvent)) error"
  ]
}
-->

# System API Windows 验收

本文档记录 System API 的 Windows 验收状态。当前自动化验证已覆盖公共 API、driver 分发、错误语义和 Windows 结构构造；人工 GUI 点验仍需在本机运行示例完成，未执行项不能视为已通过。

## 自动化验证

最近一次本地自动化验证：

```sh
go test ./...
go vet ./...
go run ./examples/system_validation
go run ./examples/system_validation -dialogs
go run ./examples/system_validation -messagebox-sources
go run ./examples/system_validation -owner-modal
go run ./examples/system_validation -notify
go run ./examples/system_validation -toast
go run ./examples/system_validation -toast-shortcut -toast-app-id com.example.FluxUI
go run ./examples/system_validation -toast-shortcut -toast-app-id com.example.FluxUI -toast-shortcut-name "FluxUI Validation Activator" -toast-activator-clsid "{01234567-89AB-CDEF-0123-456789ABCDEF}"
go run ./examples/system_validation -toast-activator -toast-activator-clsid "{01234567-89AB-CDEF-0123-456789ABCDEF}"
go run ./examples/system_validation -tray
go run ./examples/system_validation -clipboard
go run ./examples/system_validation -clipboard-files
go run ./examples/system_validation -clipboard-image
go run ./examples/system_validation -shell
go run ./examples/system_validation -shell-errors
go run ./examples/system_validation -single-instance
go run ./examples/system_validation -events 5s
```

结果：已通过。

补充检查：

- docs meta JSON 校验：已通过。
- `git diff --check`：无空白错误，仅 Windows LF/CRLF 提示。
- Windows File Dialog helper 测试：已覆盖显示后 context cancel watcher 的关闭/停止行为、过滤器规范化和 HRESULT helper。
- Windows Window helper 测试：已覆盖 `WM_CLOSE` 进入 native close hook 后触发 `WindowEventCloseRequested`，handler 返回 false 时取消系统关闭路径。
- Windows Toast helper 测试：已覆盖 Toast XML、protocol activation、listener PowerShell 脚本、event line 解析，以及 click/action/dismiss 到 `NotificationEvent` 的回调分发；Toast shortcut helper 测试已覆盖 AppUserModelID shortcut 名称归一化、参数 quote、图标 `path,index` 解析、`ToastActivatorCLSID` property key/CLSID 解析和缺失目标错误分类；Toast COM activator 测试已覆盖 `INotificationActivationCallback` IID、user input 解析、class factory 创建 callback、`Activate` 分发、`StartToastActivator` 生命周期和 `CoCreateInstance` 到 callback 的本进程 COM 链路。
- 2026-06-09 Windows 本机命令行验证：`system_validation` 默认探测确认 window/file_dialog/message_box/notification/tray/system_events/clipboard/shell 均 available。
- 2026-06-09 Windows 本机 dialog 验证：`-dialogs` 已弹出原生 MessageBox 和 File Dialog，并通过 context 自动关闭；MessageBox 的 `TaskDialog` / `MessageBoxW` 回退路径已补齐显示后 context error 返回。
- 2026-06-09 Windows 本机 MessageBox source 验证：`-messagebox-sources` 已在带 common-controls v6 manifest 的 `system_validation` 中强制走 `TaskDialogIndirect`，并自动验证 Cancel、Escape 和右上角关闭来源分别返回 `cancel`、`escape` 和 `close`。
- 2026-06-09 Windows 本机 owner modal 验证：`-owner-modal` 已创建临时 Win32 owner 窗口，验证 owner-bound MessageBox 和 File Dialog 打开期间会禁用 owner，并通过 context 自动关闭。
- 2026-06-09 Windows 本机 notification 验证：`-notify` 已提交同组替换和 `CancelNotificationGroup`；balloon probe available，Toast probe available 且 `SupportsDurableActivation=true`。
- 2026-06-09 Windows 本机 Toast 验证：`-toast` 已提交带 action button 的显式 Toast，并通过 `CancelNotificationGroup` 清理 group。
- 2026-06-09 Windows 本机 Toast shortcut 验证：`-toast-shortcut -toast-app-id com.example.FluxUI` 已创建并清理当前用户开始菜单 AppUserModelID shortcut；修正 `IID_IPropertyStore` 后可成功写入 shortcut 属性。`-toast-shortcut -toast-app-id com.example.FluxUI -toast-shortcut-name "FluxUI Validation Activator" -toast-activator-clsid "{01234567-89AB-CDEF-0123-456789ABCDEF}"` 已通过，可创建并清理带 `AppUserModel.ToastActivatorCLSID` 前置属性的 shortcut。
- 2026-06-09 Windows 本机 Toast activator 验证：`-toast-activator -toast-activator-clsid "{01234567-89AB-CDEF-0123-456789ABCDEF}"` 已创建并清理当前用户 HKCU `Software\Classes\CLSID\{...}\LocalServer32`；`-toast-shortcut -toast-activator` 联合验证已通过，Toast backend probe 报告 durable=true。
- 2026-06-09 Windows 本机 tray 验证：`-tray` 已完成创建、显示、内存图标更新、临时 `.ico` 路径图标更新、动态菜单 provider、隐藏、状态查询和关闭；`-tray -tray-resource-id 2` 已验证进程资源图标更新。`system_validation` 的 `.syso` 同时嵌入 common-controls v6 manifest 和验证图标，manifest 使用资源 ID 1，图标 group resource 使用 ID 2。
- 2026-06-09 Windows 本机 clipboard/shell/single-instance 验证：`-clipboard` 已完成文本写入、读取和恢复，并修复空文本恢复时 `Set-Clipboard` 拒绝 null 的路径；`-clipboard-files` 已完成文件列表写入和读取；`-clipboard-image` 已完成 PNG 图片写入、读取和恢复；`-shell` 已提交稳定工作区 file URL、工作区目录打开和 Explorer 定位验证文件请求，不再使用运行后会删除的临时目录；`-shell-errors` 已验证空 URL 命中 `ErrInvalidTarget`、缺失路径同时命中 `ErrUnavailable` 和 `ErrTargetNotFound`；`-single-instance` 已验证二次启动 payload/args 转发。Windows Shell driver 也会在 `OpenURL(file://...)` / `OpenPath` / `RevealPath` 提交系统打开前检查目标存在，已删除路径会返回 `ErrUnavailable` + `ErrTargetNotFound` 而不是触发 Explorer “位置不可用”弹窗。
- 2026-06-09 Windows 本机回归复跑：`-dialogs`、`-messagebox-sources`、`-owner-modal`、`-notify`、`-toast`、`-toast-shortcut -toast-activator`、`-toast-activator`、`-tray`、`-tray -tray-resource-id 2`、`-events 5s`、`-clipboard`、`-clipboard-files`、`-clipboard-image`、`-shell`、`-shell-errors`、`-single-instance` 均通过；组合验证使用 `system_validation` 的 60 秒默认总超时，避免长链路系统调用耗尽共享 context；`-events 5s` 本次未观察到实际系统事件。
- Windows Tray helper 测试：已覆盖 `TaskbarCreated` 恢复路径，visible 且未关闭的托盘图标会在任务栏重建消息后尝试 `NIM_ADD`，隐藏或已关闭托盘不会恢复。
- 2026-06-09 Windows 本机 system events 验证：`-events 5s` 可正常订阅并超时退出；本次窗口期内未观察到实际系统事件。
- 2026-06-09 非 Windows 编译：`GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go test -c ./system`、`./widget`、`./examples/system_validation` 与 `GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go test -c ./system`、`./widget`、`./examples/system_validation` 均通过，验证 `system`、`widget` 和命令行验证入口可交叉编译；macOS/Linux `CapabilityClipboard` / `CapabilityFileDialog` / `CapabilityMessageBox` / `CapabilityNotification` / `CapabilitySystemEvents` / `CapabilityTray` / `CapabilityShell` 最小 driver 可编译，其中 macOS Tray 使用 `osascript` + AppKit `NSStatusItem` 命令型 driver，Linux Tray 使用 `yad --notification` 命令型 driver，其它未实现平台仍保持 `ErrUnsupported`。

## 人工点验记录

| 能力 | 点验项 | 当前记录 |
| --- | --- | --- |
| Window close request | 点击标题栏关闭、Alt+F4、系统菜单关闭时触发 `WindowEventCloseRequested`，handler 返回 false 可取消关闭 | 2026-06-09 Windows helper 测试通过：`nativeWindowProc(WM_CLOSE)` 会触发 close-request handler 并可取消关闭；`examples/window_showcase` 已提供关闭前“是否保存”视觉路径；真实标题栏/Alt+F4 仍需人工点击确认 |
| Window native controls | `RequestFocus`、`SetPosition`、`SetResizable(false/true)`、`SetDecorated(false/true)` 行为符合预期 | 2026-06-09 公共层测试通过：窗口状态、订阅事件、关闭后操作、无 native handle 降级路径已覆盖；Windows native helper 路径由 `examples/window_showcase` 人工入口覆盖，真实拖拽/焦点视觉仍需人工确认 |
| Window constraints | 设置 `MaxSize` 后最大化不可用，全屏仍能占满显示器 | 2026-06-09 `go test ./app` 覆盖 `MaxSize` 阻止 `Maximize()`、全屏配置保留请求约束且不会丢失状态；真实标题栏最大化按钮禁用和全屏视觉仍需人工确认 |
| Window scale event | 拖动窗口到不同缩放显示器或调整显示缩放后，`WindowEventScaleChanged` 能报告窗口级 `Scale` / `DPI` | 2026-06-09 `go test ./app` 覆盖 frame metric 变化产生 `WindowEventScaleChanged`，并更新 `Scale` / `TextScale` / `DPI`；真实跨显示器拖动仍需人工确认 |
| Window hide/show | `Hide()` 后窗口隐藏，托盘或保留的 `WindowHandle.Show()` 可恢复 | 2026-06-09 公共层已覆盖隐藏内存策略和隐藏后渲染暂停状态；`examples/window_showcase` / `examples/system_showcase` 提供隐藏主窗口后从托盘恢复入口；真实视觉路径仍需人工确认 |
| Owner modal | `ui.OpenFileDialog` / `ui.ShowMessageBox` 自动绑定当前 HWND，弹窗打开时主窗口不能正常切回交互焦点 | 2026-06-09 Windows 本机：命令行 `system_validation -owner-modal` 通过，owner-bound MessageBox 和 File Dialog 会禁用 owner；`ui` wrapper 的自动 HWND 注入仍由 `system_showcase` 视觉路径确认 |
| File Dialog | 打开单文件、多文件、保存、选择目录；取消返回 `Cancelled=true`；中文/空格/长路径可返回 | 2026-06-09 公共层和 Windows helper 测试覆盖过滤器、默认扩展名、自动补扩展名、记忆目录、默认目录结构化错误和 context 自动关闭；完整用户选择路径仍需在 `examples/system_showcase` 人工点选 |
| File Dialog cancel | dialog 显示后取消 context，会尝试关闭原生 Common Item Dialog 并返回 context 错误 | 2026-06-09 Windows 本机：命令行 `system_validation -dialogs` 通过，原生 File Dialog 可被 context 自动关闭并返回 context error |
| MessageBox | OK、OKCancel、YesNo、YesNoCancel、RetryCancel 返回值正确；默认按钮生效 | 2026-06-09 `go test ./system` 覆盖 Windows TaskDialog 标准按钮集和默认按钮映射；真实用户点击每个按钮仍需人工点选 |
| MessageBox rich | `ShowMessageBoxDetailed` 的 details、footer、verification、自定义按钮和 command links 显示正确 | 2026-06-09 Windows TaskDialog helper 测试覆盖 details、footer、verification、自定义按钮和 command links 配置；真实视觉显示仍需人工确认 |
| MessageBox close source | `TaskDialogIndirect` 路径下点击 Cancel、按 Escape、点击右上角关闭按钮分别返回 `MessageBoxResultCancel`、`MessageBoxResultEscape`、`MessageBoxResultClose`；`TaskDialog` / `MessageBoxW` 回退路径仍按 `IDCANCEL` 记录 | 2026-06-09 Windows 本机：命令行 `system_validation -messagebox-sources` 通过，已自动驱动 `TaskDialogIndirect` 的 Cancel、Escape 和 WM_CLOSE 路径；真实人工键盘/鼠标点击仍可用 `examples/system_showcase` 复核 |
| MessageBox cancel context | `TaskDialogIndirect` 消息框显示后取消 context，会尝试关闭原生 dialog 并返回 context 错误；`TaskDialog` / `MessageBoxW` 回退路径通过线程 dialog 枚举做 best-effort 关闭 | 2026-06-09 Windows 本机：命令行 `system_validation -dialogs` 通过；`TaskDialog` / `MessageBoxW` 回退路径已补齐返回后的 context error 检查 |
| Notification balloon | 托盘气泡显示中文标题/正文；点击主体触发 `NotificationOnClick`；关闭/超时尽量触发 dismiss | 2026-06-09 命令行 `system_validation -notify` 通过，已提交中文可见通知、同组替换和主动取消；用户点击主体和关闭/超时回调仍需人工确认 |
| Notification group | 相同 `NotificationGroup` 的托盘气泡和 Toast 会在新通知显示前清理旧通知；`CancelNotificationGroup` 可清理当前组 | 2026-06-09 Windows 本机：命令行 `system_validation -notify` 通过，已提交同组替换和主动取消 group；视觉点击仍待人工确认 |
| Notification Toast | 显式 Toast backend 可显示 Toast；action buttons 可见；AppUserModelID 不可用时返回 `ErrUnavailable` | 2026-06-09 Windows 本机：命令行 `system_validation -toast` 通过，已提交带 action button 的 Toast 并取消 group；action 点击仍待人工确认 |
| Notification Toast probe | `ProbeNotificationBackend(ctx, NotificationBackendToast, NotificationAppID(...))` 能报告 Toast backend 当前状态；Windows Toast 可用时 `SupportsDurableActivation=true` | 2026-06-09 Windows 本机：命令行探测通过，Toast available，actions/click/dismiss/actionCallback/protocol/durable 均为 true |
| Notification Toast shortcut | `RegisterToastShortcut` 能创建当前用户开始菜单 `.lnk`，写入 AppUserModelID，可选写入 `ToastActivatorCLSID`，并在卸载/清理时删除 | 2026-06-09 Windows 本机：`system_validation -toast-shortcut -toast-app-id com.example.FluxUI` 通过；追加 `-toast-shortcut-name "FluxUI Validation Activator" -toast-activator-clsid "{01234567-89AB-CDEF-0123-456789ABCDEF}"` 也通过 |
| Notification Toast activator | `RegisterToastActivator` 能写入当前用户 COM LocalServer32；`StartToastActivator` / `RunToastActivator` 可注册本进程 COM class object 并分发 `ToastActivationEvent` | 2026-06-09 Windows 本机：`system_validation -toast-activator -toast-activator-clsid "{01234567-89AB-CDEF-0123-456789ABCDEF}"` 通过；单元测试覆盖 `CoCreateInstance` 到 `Activate` 分发；真实通知中心点击到具体应用启动模式仍需人工确认 |
| Notification Toast callbacks | helper 监听进程存活期间，Toast 主体点击、action 点击和 dismiss 可触发 `NotificationOnClick` / `NotificationOnAction` / `NotificationOnDismiss` | 2026-06-09 Windows helper 测试通过：event line 解析和 click/action/dismiss 回调分发已覆盖；真实用户点击 Toast 仍待人工确认 |
| Notification Toast protocol | 配置 `NotificationLaunchURI` / `NotificationAction.URI` 后，Toast 主体或协议 action 可通过已注册 URI scheme 被 Windows 打开；foreground action 仍触发 Go callback | 2026-06-09 Windows Toast XML/helper 测试覆盖 protocol launch URI、protocol action 和 foreground action 分流；真实已注册 URI scheme 激活仍需人工确认 |
| Tray | 创建、显示、隐藏、关闭托盘图标；关闭后操作返回 `ErrClosed` | 2026-06-09 Windows 本机：命令行 `system_validation -tray` 通过，已覆盖 show/hide/state/close；菜单点击仍待人工确认 |
| Tray menu | 右键菜单支持禁用、勾选、默认项、子菜单和菜单打开前 provider 刷新 | 2026-06-09 Windows 本机：命令行 `system_validation -tray` 通过，已覆盖动态 provider、默认项、子菜单、勾选/禁用构造和刷新；右键视觉点选仍待人工确认 |
| Tray icon sources | `TrayIcon`、`TrayIconBytes`、`TrayIconResource` 都能加载图标 | 2026-06-09 Windows 本机：命令行 `system_validation -tray` 已验证 `TrayIconBytes` 和临时 `.ico` 路径 `TrayIcon`；`system_validation -tray -tray-resource-id 2` 已验证 `TrayIconResource` |
| Explorer restart | Explorer/任务栏重启后 visible 托盘图标可尝试恢复 | 2026-06-09 Windows helper 测试通过：`TaskbarCreated` 恢复路径会对 visible 且未关闭托盘尝试 `NIM_ADD`；真实 Explorer 重启视觉恢复仍待人工确认 |
| App lifecycle | `ui.RunElement` / `RunElementMulti` 退出后自动调用 `system.CloseTrays()` 清理托盘 | 2026-06-09 代码路径已在 `app.runSpecs` 窗口全部结束后调用 `system.CloseTrays()`；`go test ./system` 覆盖 `CloseTrays` 会关闭注册托盘；真实应用退出视觉清理仍需人工确认 |
| System events | display/theme/settings/power/session 事件能通过 `SubscribeSystemEvents` 收到；DPI 事件按 Windows 消息可达性记录 | 2026-06-09 Windows 本机：命令行 `system_validation -events 5s` 通过，订阅可创建并超时退出；本次未触发实际系统事件，事件种类仍待人工触发点验 |
| Clipboard text | `ReadClipboardText` / `WriteClipboardText` 可读写 Unicode 文本，并在剪贴板短暂占用时重试 | 2026-06-09 Windows 本机：命令行 `system_validation -clipboard` 通过，已写入测试文本、读取确认并恢复原文本；非文本剪贴板格式恢复不在该自动验证范围内 |
| Clipboard files/image | `ReadClipboardFiles` / `WriteClipboardFiles` 可读写 Explorer 文件列表；`ReadClipboardImagePNG` / `WriteClipboardImagePNG` 可读写 PNG 图片剪贴板 | 2026-06-09 Windows helper 测试覆盖文件列表和图片 PowerShell 脚本、输出解析和 PNG 公共层校验；`system_validation -clipboard-files` 和 `system_validation -clipboard-image` 本机通过，PNG 图片写入/读取/恢复路径已回归 |
| Shell open | `OpenURL` / `OpenPath` / `RevealPath` 可把打开请求提交给系统默认程序或 Explorer | 2026-06-09 Windows 本机：命令行 `system_validation -shell` 通过，已提交稳定工作区 file URL、工作区目录打开和 Explorer 定位验证文件请求；`OpenURL(file://...)`、`OpenPath` 和 `RevealPath` 对缺失本地目标会返回 `ErrUnavailable`，不再触发 Explorer “位置不可用”弹窗；外部程序视觉打开仍需人工确认 |
| Shell errors | 空 URL/路径、缺失本地目标、无默认 handler、权限拒绝等错误可被细分 helper 识别 | 2026-06-09 公共层和 Windows helper 测试覆盖 `ErrInvalidTarget`、`ErrTargetNotFound`、`ErrNoDefaultHandler` 和 `ErrAccessDenied`；`system_validation -shell-errors` 本机通过，空 URL 和缺失本地路径分类已回归 |

## 点验建议

使用 `examples/system_showcase` 和 `examples/window_showcase` 做人工点验。涉及 Toast 时，应准备一个可用的 AppUserModelID；未打包桌面应用通常需要安装器或开始菜单快捷方式注册 AppUserModelID，否则 Toast 服务可能拒绝显示并返回 `ErrUnavailable`。

命令行验证入口：

```sh
go run ./examples/system_validation
go run ./examples/system_validation -dialogs
go run ./examples/system_validation -messagebox-sources
go run ./examples/system_validation -owner-modal
go run ./examples/system_validation -notify
go run ./examples/system_validation -toast
go run ./examples/system_validation -toast-shortcut -toast-app-id com.example.FluxUI
go run ./examples/system_validation -tray
go run ./examples/system_validation -tray -tray-resource-id 2
go run ./examples/system_validation -clipboard
go run ./examples/system_validation -shell
go run ./examples/system_validation -events 30s
```

示例入口映射：

- `examples/window_showcase` 主窗口的系统关闭请求会被 `OnCloseRequested` 先取消，再异步打开“是否保存”系统消息框；选择 Yes 或 No 后才会调用 `WindowHandle.Close()`，可用于验证标题栏关闭、Alt+F4 和系统菜单关闭。
- `examples/window_showcase` 覆盖持续置顶、隐藏/显示、隐藏内存策略、尺寸限制、最大化禁用、全屏绕过尺寸限制、窗口事件和状态快照。
- `examples/system_showcase` 的 File Dialog 区覆盖单文件、多文件、保存、目录选择、记忆目录、保存默认扩展名和显示后 context 自动取消。
- `examples/system_showcase` 的 MessageBox 区覆盖标准按钮集、默认按钮、富 TaskDialog、自定义 command links、verification 状态和显示后 context 自动取消。
- `examples/system_showcase` 的 Notification 区覆盖托盘气泡、同组替换、主动取消 group、Toast backend 探测、Toast protocol activation、action callback 和 group replace。
- `examples/system_showcase` 的 Tray 区覆盖创建、显示、隐藏、主窗口隐藏、状态查询、内存 `.ico` 图标、进程资源图标、动态右键菜单、默认项、子菜单、动态勾选/禁用和关闭清理。
- `examples/system_showcase` 的 System Events 区覆盖 30 秒限时订阅和最近事件显示。
- `examples/system_validation` 覆盖 PA-F/G 核心能力探测，并通过命令行参数触发真实 notification、tray、clipboard、shell 和 system events 验证，适合作为人工点验前后的快速回归入口。

人工点验完成后，把“待执行”更新为具体日期、环境和结果，例如：

```text
2026-06-09 Windows 11 23H2: 自动化通过。备注：Toast COM LocalServer 注册和本进程 `INotificationActivationCallback` 分发已落地；真实通知中心点击到具体应用启动模式仍需按应用命令行人工点验。
```

不要把未执行的 GUI 行为写成通过。
