<!-- fluxui-doc-meta
{
  "id": "system_api_roadmap",
  "title": "System API 长期路线图",
  "category": "使用指南",
  "order": 122,
  "summary": "规划 FluxUI 后续系统层 API 封装，包括窗口、文件选择、系统弹窗、系统通知、托盘与多平台占位策略。",
  "example": { "id": "system_api_basic" },
  "apis": [
    "system.OpenFileDialog(ctx context.Context, opts ...FileDialogOption) (FileDialogResult, error)",
    "system.SaveFileDialog(ctx context.Context, opts ...FileDialogOption) (FileDialogResult, error)",
    "system.ShowMessageBox(ctx context.Context, opts ...MessageBoxOption) (MessageBoxResult, error)",
    "system.ShowMessageBoxDetailed(ctx context.Context, opts ...MessageBoxOption) (MessageBoxDetailedResult, error)",
    "system.Notify(ctx context.Context, opts ...NotificationOption) error",
    "system.ProbeNotificationBackend(ctx context.Context, backend NotificationBackend, opts ...NotificationOption) NotificationBackendProbe",
    "system.CancelNotificationGroup(ctx context.Context, group string) error",
    "system.RegisterToastShortcut(ctx context.Context, appID, shortcutName, executable string, opts ...ToastShortcutOption) error",
    "system.ToastShortcutActivatorCLSID(clsid string) ToastShortcutOption",
    "system.RegisterToastActivator(ctx context.Context, clsid, command string) error",
    "system.UnregisterToastActivator(ctx context.Context, clsid string) error",
    "system.StartToastActivator(ctx context.Context, clsid string, fn func(ToastActivationEvent)) (*ToastActivator, error)",
    "system.RunToastActivator(ctx context.Context, clsid string, fn func(ToastActivationEvent)) error",
    "system.SubscribeSystemEvents(ctx context.Context, kinds ...SystemEventKind) (*SystemEventSubscription, error)",
    "system.NewTray(opts ...TrayOption) (*Tray, error)",
    "system.ReadClipboardText(ctx context.Context) (string, error)",
    "system.WriteClipboardText(ctx context.Context, text string) error",
    "system.OpenURL(ctx context.Context, target string) error",
    "system.OpenPath(ctx context.Context, path string) error",
    "system.RevealPath(ctx context.Context, path string) error",
    "system.ProbeDragAndDrop(ctx context.Context) DragAndDropProbe",
    "WindowElement(root Component, opts ...AppOption) WindowSpec",
    "RunElementMulti(windows ...WindowSpec) error",
    "WindowHandle.NativeHandle() (uintptr, bool)",
    "ui.OpenFileDialog(ctx *Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "ui.ShowMessageBox(ctx *Context, opts ...system.MessageBoxOption) (system.MessageBoxResult, error)",
    "WindowHandle.Close/Show/Hide/SetHiddenMemoryPolicy/Minimize/Maximize/Restore/Fullscreen/Raise/SetAlwaysOnTop/Center/SetTitle/SetSize/Invalidate"
  ]
}
-->

# System API 长期路线图

本文档记录 FluxUI 后续系统层 API 的设计方向与分阶段推进计划。当前 FluxUI 在渲染、组件、主题、动画与 React-style Element API 上已经具备较完整的应用构建能力，但桌面应用还缺少一层稳定的系统能力封装：窗口、文件选择、系统弹窗、系统通知、托盘菜单，以及未来多平台可维护的 driver 边界。

短期目标不是一次性实现所有平台，而是先定义稳定的跨平台 API 形状，再优先落地 Windows。其他平台先提供可编译的空实现和明确的 `ErrUnsupported`，后续由开源社区按同一接口补充实现。

## 当前进展

截至 2026-06-09，FluxUI 已经具备一部分窗口级基础能力。

### 已完成的窗口基础设施

- `app.WindowHandle` 已提供 `ID`、`IsAlive`、`NativeHandle`、`Close`、`Show`、`Hide`、`SetHiddenMemoryPolicy`、`Minimize`、`Maximize`、`Restore`、`Fullscreen`、`Raise`、`RequestFocus`、`SetAlwaysOnTop`、`Center`、`SetTitle`、`SetPosition`、`SetSize`、`SetResizable`、`SetDecorated`、`SetWindowsFrameStyle`、`StartDragMove`、`Invalidate`。
- `app.RunMulti` 与 `ui.RunElementMulti` 已支持桌面多窗口启动。
- `ui.WindowElement` 已提供 React-style 多窗口入口。
- `ui.ListWindows` / `ui.GetWindow` 已能查询当前存活窗口。
- `ui.CurrentWindowID(ctx)` 与 `WindowClose`、`WindowShow`、`WindowHide`、`WindowSetHiddenMemoryPolicy`、`WindowMinimize`、`WindowMaximize`、`WindowRestore`、`WindowFullscreen`、`WindowRaise`、`WindowRequestFocus`、`WindowSetAlwaysOnTop`、`WindowCenter`、`WindowSetTitle`、`WindowSetPosition`、`WindowSetSize`、`WindowSetResizable`、`WindowSetDecorated`、`WindowSetWindowsFrameStyle`、`WindowStartDragMove`、`WindowInvalidate`、`WindowIsAlive` 已经把当前窗口控制暴露到 UI 层。
- `internal.WindowController` 已形成当前窗口控制的内部接口，后续可以作为系统 API 与运行时之间的桥。

### 已完成的应用内反馈能力

- `Dialog` / `DialogElement`、`Popup` / `PopupElement`、`Toast` / `ToastElement`、`Snackbar` / `SnackbarElement` 已提供应用内弹层与消息反馈。
- 这些能力是 FluxUI 自绘 UI，不是系统原生弹窗、系统通知或任务栏托盘能力。后续系统 API 不应替代它们，而是补齐“需要操作系统参与”的场景。

### 当前进展与剩余缺口

- 已新增统一的 `system` 包，窗口、文件选择、消息框、通知和托盘已经共享能力检测、错误语义和平台 driver 约定。
- Windows 已封装原生文件选择器：打开文件、保存文件、选择目录、多选、过滤器、默认目录和默认文件名均已具备 v1 API。
- Windows 已封装系统原生弹窗：信息、警告、错误、确认、取消、按钮结果、owner window 和模态行为已具备 v1 API。
- Windows 已提供系统通知路径：默认通过 `Shell_NotifyIconW` 托盘气泡实现；显式 Toast backend、action button XML、AppUserModelID 开始菜单快捷方式注册、ToastActivatorCLSID shortcut 前置属性、COM LocalServer 注册、当前进程 `INotificationActivationCallback` 分发和短生命周期实时 click/action/dismiss 回调已落地。通知中心持久点击到具体应用启动模式仍需要按应用命令行做人工点验。
- Windows 已提供任务栏托盘：长期托盘图标、tooltip、左键/右键事件、菜单、显示隐藏和退出动作已具备 v1 API。
- Windows 已提供系统事件订阅第一版：`SubscribeSystemEvents` 可订阅 display、DPI、theme、settings、power 和 session 事件；具体事件到达仍受 Windows 消息模型限制。
- Windows 已提供 Clipboard / Shell open / Drag & Drop probe 第一版：文本剪贴板读写、默认程序打开 URL/文件/目录、Explorer 定位路径和 Gio transfer 拖放能力探测已具备 v1 API；`file:` URL 和本地路径会在提交 Explorer 前检查目标存在，避免已删除目标触发系统“位置不可用”弹窗。
- macOS / Linux 已补 File Dialog、Clipboard 文本、Shell open、MessageBox、Notification、System Events 和 Tray 最小 driver：macOS 使用 `osascript choose file/folder/file name`、`pbpaste` / `pbcopy`、`open` / `open -R`、`osascript display dialog`、`osascript display notification`，通过 `ioreg` / `defaults` / `pmset` / `stat` 轮询系统事件快照，并用 `osascript` + AppKit `NSStatusItem` 提供命令型最小 Tray；Linux 使用 `zenity` / `kdialog`、`wl-clipboard` / `xclip` / `xsel`、`xdg-open`、`notify-send` / `kdialog --passivepopup`，通过 `/sys` / `gsettings` / `loginctl` 轮询系统事件快照，并用 `yad --notification` 提供命令型最小 Tray；其它未实现平台仍保持可编译并明确返回 `ErrUnsupported`。

## 长期目标

FluxUI 的系统层 API 应成为渲染层之外的稳定桌面能力边界：

- API first：先稳定公共 API 与错误语义，再按平台补 driver。
- Windows first：核心实现优先解决 Windows 桌面应用真实需求。
- Multi-platform compile first：所有平台必须能编译，未实现的平台返回明确的 `ErrUnsupported`。
- Owner aware：文件选择器、消息框等系统 UI 默认绑定当前 FluxUI 窗口，避免无 owner 的浮动弹窗。
- Non-rendering service：系统 API 不应伪装成普通 Widget。它们是应用服务，应该从事件回调、后台任务或 app 生命周期中触发。
- Element friendly：React-style 应用可以自然调用系统能力，但系统能力本身不强行做成 Element。
- No silent fallback：系统弹窗、通知、托盘不可用时必须显式返回错误，不应悄悄退化为自绘 Dialog 或 Toast。
- Community-port friendly：Linux、macOS、移动端和 Web 的实现文件应有清晰的 build tags、driver 接口和测试边界。

## 设计原则

### 系统能力独立成层

建议新增 `system` 包作为公共系统 API 层，`app` 继续负责运行时和窗口事件循环，`ui` 只提供少量便利转发。

推荐分层：

- `system`: 公共类型、能力检测、错误、跨平台函数。
- `system/internal` 或 `internal/system`: driver 注册、平台适配、线程约束、Win32/COM 边界。
- `app`: 窗口生命周期、窗口注册表、当前窗口 owner 映射。
- `ui`: 从 `*ui.Context` 取得当前窗口并调用 `system` 的便利函数。

不建议把文件选择、通知、托盘直接塞进 `widget` 包。它们不参与布局，也不应该跟自绘组件共享生命周期。

### 平台不支持也是稳定行为

所有公共 API 在所有平台都必须存在。非 Windows 平台先提供空实现：

```go
var ErrUnsupported = errors.New("system: unsupported on this platform")

func IsUnsupported(err error) bool {
    return errors.Is(err, ErrUnsupported)
}
```

未实现平台返回 `ErrUnsupported` 或带能力名的包装错误。调用方可以通过 `system.Supports(system.CapabilityFileDialog)` 在展示按钮前决定是否启用。

### 同步 API 与 UI 生命周期隔离

系统 API 可以提供阻塞式 Go 函数，但文档必须明确：

- 不要在组件布局函数中直接调用系统 API。
- 推荐在点击、菜单、快捷键等事件回调中触发。
- 可能阻塞的 native dialog 应支持 `context.Context` 取消。
- 如果 Windows API 要求固定线程或 COM apartment，应由 driver 内部处理，不把线程细节泄漏给用户。

### Owner window 是默认路径

文件选择器和系统弹窗应默认绑定当前窗口。推荐公共 API 允许显式传入 owner：

```go
owner, ok := handle.NativeHandle()
if ok {
    system.OpenFileDialog(ctx, system.FileDialogOwner(owner))
    system.ShowMessageBox(ctx, system.MessageBoxOwner(owner))
}
```

`ui` 层已提供便利函数，自动使用当前窗口 owner：

```go
ui.OpenFileDialog(ctx, opts...)
ui.ShowMessageBox(ctx, opts...)
```

Windows v1 已通过 Gio `Win32ViewEvent` 捕获 HWND，并由 `WindowHandle.NativeHandle()` 暴露给调用方。`system` 包本身没有 `*ui.Context`，不会自动推导当前窗口；`ui` wrapper 会在调用 `system` API 前自动注入当前窗口 owner，直接调用 `system` 包时仍可显式传 owner。

### 自绘反馈与系统反馈各司其职

- `Dialog` / `Popup`: 应用内流程、富内容、表单输入、主题一致的自绘弹层。
- `Toast` / `Snackbar`: 应用内短反馈。
- `system.ShowMessageBox`: 操作系统原生阻塞确认。
- `system.Notify`: 操作系统通知中心或托盘气泡。
- `system.Tray`: 后台驻留应用入口。

系统 API 不应自动调用自绘组件作为 fallback。调用方如果希望 fallback，应在业务层显式处理。

## 推进阶段

### Phase A: 系统 API 边界与能力检测

目标：建立系统层公共 API 的包结构、错误语义、能力检测和平台 driver 边界。

任务：

- 新增 `system` 包，定义 `Capability`、`CapabilitySet`、`ErrUnsupported`、`ErrUnavailable`、`Supports`、`Capabilities`。
- 定义 driver 接口，不在公共 API 中暴露 Win32、COM、HWND 或平台专属类型。
- 新增 `system_unsupported.go`，让非 Windows 平台全部可编译并返回 `ErrUnsupported`。
- 新增 Windows driver skeleton，先只返回能力表，不实现具体功能。
- 明确 `system` 包与 `ui` 包的职责：`system` 提供核心能力，`ui` 提供当前窗口 owner 的便捷封装。
- 为 `ErrUnsupported`、能力检测和非 Windows 空实现增加单元测试。

建议 API：

```go
type Capability string

const (
    CapabilityWindow       Capability = "window"
    CapabilityFileDialog   Capability = "file_dialog"
    CapabilityMessageBox   Capability = "message_box"
    CapabilityNotification Capability = "notification"
    CapabilityTray         Capability = "tray"
    CapabilityClipboard    Capability = "clipboard"
)

func Supports(cap Capability) bool
func Capabilities() CapabilitySet
func IsUnsupported(err error) bool
```

验收：

```sh
go test ./system ./internal/...
go test ./...
go vet ./...
```

### Phase B: Window API 完善

目标：在现有窗口控制基础上补齐桌面应用常用窗口能力，并为 native owner 做准备。

当前进展：

- B1 已完成现有窗口 API 梳理，窗口创建、多窗口、当前窗口 helper 和 `WindowHandle` 继续兼容。
- B2 已新增 `WindowState` 与 `WindowHandle.State()`，作为 FluxUI runtime 维护的窗口状态快照。
- B3 已新增 `MinSize` / `MaxSize` app option，以及 `WindowHandle.SetMinSize` / `SetMaxSize`、`WindowSetMinSize(ctx, ...)` / `WindowSetMaxSize(ctx, ...)`。
- B4 已补充 Windows 持续置顶能力：`WindowHandle.SetAlwaysOnTop(always)` / `WindowSetAlwaysOnTop(ctx, always)` 通过 native `HWND` 异步调用实现；`Raise` 保持为一次性前置请求。
- B5 已完成 Windows 关闭拦截：通过 native close hook 在 `WM_CLOSE` 路径触发 `WindowEventCloseRequested`，`WindowHandle.SetCloseRequestedHandler` / `ui.OnCloseRequested` 的 handler 返回 false 时可取消关闭；非 Windows 平台仍保持可编译并按平台能力降级。
- B6 已新增轻量窗口事件模型，提供 `WindowHandle.PollEvents()`、`WindowEvent` 和 size/focus/state/closed 事件。
- B7 已完成 native owner 接入：Windows 下通过 Gio `Win32ViewEvent` 捕获 HWND，并由 `WindowHandle.NativeHandle()` 暴露为 `uintptr`；非 Windows 平台返回 false。
- B8 已新增 `docs/guides/window-api.md` 作为窗口 API 集中入口。
- B9 已新增 `examples/window_showcase`，覆盖多窗口、状态快照、事件订阅、关闭请求拦截、关闭前“是否保存”系统消息框、持续置顶、显示/隐藏、隐藏内存策略和常用窗口动作；`PollEvents()` 作为兼容入口继续保留。
- B10 已修正全屏尺寸约束：进入全屏时临时清除 Gio 层 `MinSize` / `MaxSize`，退出全屏或回到普通窗口模式时重新应用应用请求的约束。已通过 `go test ./...` 与 `go vet ./...` 验收。
- B11 已新增隐藏内存策略：默认 `WindowHiddenMemoryReleaseTransient` 会在隐藏后暂停 FluxUI 渲染、跳过隐藏窗口 redraw invalidation、清空临时 `op.Ops` 并异步触发 Go 内存回收；`WindowHiddenMemoryKeepRenderingState` 可保留隐藏窗口渲染状态。当前不直接释放 Gio 内部 GPU context。
- B12 已修正最大尺寸限制下的最大化行为：设置完整 `MaxSize(width, height)` 后，`WindowHandle.Maximize()` / `WindowMaximize(ctx)` 返回 `false`，Windows 原生标题栏最大化按钮和系统菜单最大化项同步禁用；全屏仍按 B10 逻辑保留可用。
- Windows chrome 当前保留隐藏 frame、圆角/边框/阴影策略、`WindowDragAreaElement`、`WindowStartDragMove` 和 `ProbeWindowsChrome`；`WindowDragAreaElement` 已改为每帧解析 Gio `ActionMove` 区域并通过 Windows 原生 `WM_NCHITTEST` / `HTCAPTION` 注册拖动区，避免最大化拖拽时用手动 restore/SetWindowPos 模拟；hidden frame 会去掉 `WS_CAPTION`，避免失焦/重绘时露出旧式 Win32 frame。新增 `examples/window_chrome_showcase`，并已集成到 `examples/window_showcase` 与 docs browser Window API 页；示例保持 hidden frame 并自绘现代标题栏/边框，`WindowsFrameDefault` 仅作为兼容恢复 OS frame 的路径。DWM background material、透明背景和 native 背景颜色因自渲染 UI 与 Windows 11 DWM 材质稳定组合成本较高，已转入后续研究项，近期不继续推进。

任务：

- 梳理现有 `WindowHandle`，保持兼容，不破坏 `RunElementMulti`、`WindowElement` 和当前窗口 helper。
- 新增窗口属性查询：标题、尺寸、状态、是否全屏、是否最小化、是否最大化。
- 评估并实现窗口位置、最小尺寸、最大尺寸、是否可调整大小、是否显示系统边框、置顶、可见性、焦点请求等能力。
- 建立窗口事件模型：close requested、focus changed、size changed、scale/DPI changed、theme changed、display changed。
- 设计关闭拦截：允许应用在用户点击关闭按钮时先弹确认，再决定是否关闭。
- 捕获 Gio Windows backend 暴露的 native handle，在内部保存 owner handle，并通过 `WindowHandle.NativeHandle()` 提供给 File Dialog / MessageBox owner option。
- 为窗口 registry、WindowHandle 生命周期、关闭后操作返回 false 等行为补充测试。

建议 API：

```go
type WindowState struct {
    ID         WindowID
    Title      string
    Width      int
    Height     int
    Visible     bool
    AlwaysOnTop bool
    RenderSuspended bool
    HiddenMemoryPolicy WindowHiddenMemoryPolicy
    Minimized  bool
    Maximized  bool
    Fullscreen bool
    Focused    bool
}

func (h WindowHandle) State() (WindowState, bool)
func (h WindowHandle) NativeHandle() (uintptr, bool)
func (h WindowHandle) SetMinSize(width, height int) bool
func (h WindowHandle) SetMaxSize(width, height int) bool
func (h WindowHandle) SetResizable(resizable bool) bool
func (h WindowHandle) SetAlwaysOnTop(always bool) bool
func (h WindowHandle) Show() bool
func (h WindowHandle) Hide() bool
func (h WindowHandle) SetHiddenMemoryPolicy(policy WindowHiddenMemoryPolicy) bool
func (h WindowHandle) RequestFocus() bool
```

验收：

- 旧窗口 API 行为保持不变。
- 多窗口场景下 `ListWindows`、`GetWindow`、当前窗口 helper 不串窗口。
- 关闭后的 `WindowHandle` 操作不 panic，返回 false。
- Windows 下 `WindowHandle.NativeHandle()` 返回 HWND，关闭后返回 false。
- Windows 能正确设置标题、大小、最小化、最大化、还原、全屏、一次性前置、持续置顶、显示、隐藏和居中。

### Phase C: File Dialog 文件选择器

目标：提供 Windows 原生文件选择能力，并定义跨平台一致的文件选择 API。

任务：

- C1 已完成：定义 `FileDialogOption`、`FileFilter`、`FileDialogResult`、`FileDialogMode`。
- C2 已完成：支持打开单个文件、打开多个文件、保存文件、选择目录。
- C3 已完成：支持标题、默认目录、默认文件名、扩展名过滤器、是否允许不存在路径、是否允许创建目录、是否覆盖确认。
- C4 已完成：Windows 实现使用 `IFileOpenDialog` / `IFileSaveDialog` 和 Common Item Dialog。
- C5 已完成：处理 COM 初始化、线程绑定、UTF-16 路径转换、取消返回值、context 打开后取消和错误包装。Windows 下 context 取消会调用 `IFileDialog.Close` 尝试关闭正在显示的 Common Item Dialog；取消 watcher 会在 dialog 返回后停止并等待退出，避免释放 COM 对象后继续关闭，Windows-only helper 测试已覆盖该边界。
- C6 已完成：新增显式 `FileDialogOwner(hwnd)`，Windows 下映射为 `HWND`；FluxUI 可通过 `WindowHandle.NativeHandle()` 取得当前窗口 HWND，未传 owner 时仍保留无 owner fallback。
- C7 已完成并扩展：非 Windows 平台保持可编译；macOS / Linux 已声明 `CapabilityFileDialog` 并提供最小命令 driver，其他未实现平台调用 File Dialog API 返回 `ErrUnsupported`。
- C8 已完成：新增 `examples/system_showcase`，示例只在 `system.Supports(CapabilityFileDialog)` 时启用按钮，并通过 `ui` wrapper 自动将当前窗口 native owner 传给文件对话框；示例已覆盖记忆目录、保存默认扩展名和显示后 context 自动取消。
- C9 已完成：新增 `docs/guides/file-dialog-api.md`，并更新 System API 总入口。
- C10 已完成：补充验收清单，覆盖取消、多选、过滤器、默认目录错误、中文路径、空格路径和长路径。

建议 API：

```go
type FileFilter struct {
    Name     string
    Patterns []string
}

type FileDialogResult struct {
    Paths     []string
    Cancelled bool
}

func OpenFileDialog(ctx context.Context, opts ...FileDialogOption) (FileDialogResult, error)
func OpenFilesDialog(ctx context.Context, opts ...FileDialogOption) (FileDialogResult, error)
func SaveFileDialog(ctx context.Context, opts ...FileDialogOption) (FileDialogResult, error)
func PickFolderDialog(ctx context.Context, opts ...FileDialogOption) (FileDialogResult, error)
func FileDialogOwner(owner uintptr) FileDialogOption
```

验收：

- 取消选择返回 `Cancelled=true` 且 `err == nil`。
- 选择文件返回绝对路径。
- 多选顺序稳定。
- 过滤器能正确显示并限制扩展名。
- 默认目录不存在时返回清晰错误或由 Windows dialog 自身处理，但行为必须测试并文档化。
- Windows 路径包含中文、空格、长路径时不乱码。
- 显式 owner 可用时，文件对话框作为 owner 窗口的 modal 对话框显示。

### Phase D: MessageBox / TaskDialog 系统弹窗

目标：提供系统原生消息框，用于轻量确认、警告和错误提示。

当前状态：D1-D9 已完成，D10 自动化验收已完成，Windows 本地人工点验待执行。Phase D 已沿用 File Dialog 的系统层边界补齐 MessageBox 公共 API、unsupported 行为、Windows `TaskDialogIndirect` 优先实现、`TaskDialog` / `MessageBoxW` 兼容回退、owner modal 行为、关闭/取消语义说明、异步入口、富 TaskDialog 选项、macOS/Linux 最小命令 driver、文档指南和 `system_showcase` 示例。

任务：

- D1 已完成：公共类型设计，定义 `MessageBoxKind`、`MessageBoxButtons`、`MessageBoxResult`、`MessageBoxOption` 和内部 `messageBoxOptions`，覆盖 info、warning、error、question、OK、OKCancel、YesNo、YesNoCancel、RetryCancel。
- D2 已完成：公共入口与 option，新增 `ShowMessageBox(ctx context.Context, opts ...MessageBoxOption) (MessageBoxResult, error)`，以及 `MessageBoxTitle`、`MessageBoxText`、`MessageBoxStyle`、`MessageBoxButtonSet`、`MessageBoxDefaultButton`、`MessageBoxOwner`。
- D3 已完成：driver 边界，新增 message box driver 分发；未支持或未声明 `CapabilityMessageBox` 时返回包装后的 `ErrUnsupported`。
- D4 已完成并扩展：非 Windows 平台保持可编译；macOS / Linux 已声明 `CapabilityMessageBox` 并提供最小命令 driver，其他未实现平台调用 `ShowMessageBox` 返回 `ErrUnsupported`。
- D5 已完成：Windows 优先使用 `TaskDialogIndirect`，实现 icon、button set、默认按钮、owner HWND、返回值到 `MessageBoxResult` 的映射；不可用时依次回退到 `TaskDialog` 和 `MessageBoxW`。使用富选项时不会静默降级到会丢能力的旧实现。
- D6 已完成第一版：取消和关闭语义。Windows `TaskDialogIndirect` 会记录真实按钮点击和关闭来源：点击 Cancel 按钮返回 `MessageBoxResultCancel`，Escape 返回 `MessageBoxResultEscape`，右上角关闭按钮或 `WM_CLOSE` 返回 `MessageBoxResultClose`；`TaskDialog` / `MessageBoxW` 回退路径仍把 `IDCANCEL` 映射为 Cancel。
- D7 已完成：context 与阻塞边界。打开前检查 `context.Context`；Windows `TaskDialogIndirect` 显示后会监听 context 取消并向原生 dialog 投递关闭请求。`TaskDialog` / `MessageBoxW` 回退路径会在调用期间锁定当前 OS 线程，context 取消后通过枚举该线程上的原生 dialog 做 best-effort 关闭。文档强调只能从事件回调、后台任务或 app 生命周期触发，不要放进布局函数。已补 `ShowMessageBoxAsync` 和 `ShowMessageBoxDetailedAsync`。
- D7.1 已完成：富 TaskDialog 选项。新增 `ShowMessageBoxDetailed`、`MessageBoxDetails`、`MessageBoxFooter`、`MessageBoxVerification`、`MessageBoxCustomButtons`、`MessageBoxDefaultButtonID`、`MessageBoxCommandLinks` 等 API，可返回自定义按钮 ID 和复选框状态。
- D8 已完成：单元测试覆盖 option 默认值、driver 分发、能力缺失、context 已取消、按钮/结果映射 helper、错误 helper 包装判断；Windows-only 测试使用 build tag。
- D9 已完成：新增 `docs/guides/message-box-api.md`，更新 `docs/guides/system-api.md` 和 `examples/system_showcase`，示例按 `system.Supports(CapabilityMessageBox)` 启用按钮，并通过 `ui` wrapper 自动将当前窗口 native owner 传给消息框；示例已覆盖富 TaskDialog 和显示后 context 自动取消。
- D10 部分完成：`go test ./...`、`go vet ./...`、`go test ./examples/system_showcase` 已作为自动化验收；Windows 人工验证 info/warning/error/question、OK/Cancel/Yes/No/Retry、富 TaskDialog、context 自动取消、Cancel/Escape/关闭按钮区分、无 owner 与显式 owner modal 行为仍需本地点击确认。

建议 API：

```go
func ShowMessageBox(ctx context.Context, opts ...MessageBoxOption) (MessageBoxResult, error)

func MessageBoxTitle(title string) MessageBoxOption
func MessageBoxText(text string) MessageBoxOption
func MessageBoxStyle(kind MessageBoxKind) MessageBoxOption
func MessageBoxButtonSet(buttons MessageBoxButtons) MessageBoxOption
func MessageBoxDefaultButton(result MessageBoxResult) MessageBoxOption
func MessageBoxOwner(owner uintptr) MessageBoxOption
```

验收：

- 各按钮集返回值正确映射。
- `TaskDialogIndirect` 路径下 Cancel、Escape 和关闭按钮行为可区分且文档化；`TaskDialog` / `MessageBoxW` 回退路径的 `IDCANCEL` 行为文档化。
- owner window 可用时消息框置于当前窗口前方，并在关闭前阻止 owner 窗口正常切回交互焦点。
- 无 owner 时仍可使用，但不承诺置顶。

### Phase E: System Notification 系统通知与消息

目标：提供非阻塞系统消息能力，用于后台任务完成、错误、提醒和托盘气泡。

当前状态：E1-E10 已完成并补充 Windows 展示修复；macOS / Linux Notification 命令型最小 driver 已补充。Phase E v1 已补齐 Notification 公共 API、option 模型、driver 分发、事件 callback 承载点、Windows `Shell_NotifyIconW` 托盘气泡展示路径、Toast 请求路径、Toast backend 级探测、action button XML、协议 URI 激活、AppUserModelID shortcut 注册、ToastActivatorCLSID shortcut 前置属性、COM LocalServer 注册、当前进程 `INotificationActivationCallback` 分发、短生命周期实时 Toast 事件监听、macOS `osascript display notification`、Linux `notify-send` / `kdialog`、文档、示例和自动化验收。通知中心点击后跨进程回到具体应用启动模式仍作为应用打包/命令行集成点验项保留。

任务：

- E1 已完成：公共类型设计，定义 `NotificationKind`、`NotificationOption`、`NotificationEvent`、`notificationOptions`，覆盖 info、success、warning、error、title、body、icon、group、timeout。
- E2 已完成：公共入口与 option，新增 `Notify(ctx context.Context, opts ...NotificationOption) error`，以及 `NotificationTitle`、`NotificationBody`、`NotificationKindStyle`、`NotificationIcon`、`NotificationGroup`、`NotificationTimeout`。
- E3 已完成：driver 边界，新增 notification driver 分发；未支持或未声明 `CapabilityNotification` 时返回包装后的 `ErrUnsupported`。
- E4 已完成并升级：非 Windows 平台可编译；macOS / Linux 已提供命令型最小 notification driver，缺少平台命令时返回 `ErrUnavailable`，Toast、actions、protocol activation 等不支持子能力返回 `ErrUnsupported`；其它平台仍返回 `ErrUnsupported`。
- E5 已完成：Windows v1 展示路径已实现。Windows driver 声明 `CapabilityNotification`，通过隐藏消息窗口和 `Shell_NotifyIconW` 临时托盘图标提交气泡通知；Windows shell、通知区或图标加载不可用时返回包装后的 `ErrUnavailable`。
- E6 已完成：默认展示路径固定为托盘气泡通知；显式请求 Toast 或配置 action buttons 时走 Toast 请求路径。Toast 支持 backend 级探测、foreground action、协议 URI 激活和短生命周期实时回调；AppUserModelID、快捷方式/安装器注册、打包/非打包差异和 COM activation callback 已写入限制文档。长期 Tray API 生命周期已在 Phase F 完成。
- E7 已完成：`NotificationOnClick(fn func(NotificationEvent))` 已作为 option 透传并接入 Windows `NIN_BALLOONUSERCLICK`；系统没有回调时不伪造 clicked/dismissed/action 事件。
- E8 已完成：单元测试覆盖 option 默认值、driver 分发、能力缺失、context 已取消、`ErrUnavailable` 包装、unsupported 平台行为、click callback option 透传，以及 Windows notification data、图标标志和超时映射。
- E9 已完成：新增 `docs/guides/notification-api.md`，更新 `docs/guides/system-api.md`、`docs/README.md` 和 `examples/system_showcase`；示例按 `system.Supports(CapabilityNotification)` 展示能力状态，并在系统通知区不可用时展示 `ErrUnavailable`。
- E10 已完成：已运行 `go test ./...`、`go vet ./...`、`go test ./examples/system_showcase`；Windows 人工验收覆盖中文标题/正文、托盘气泡显示和点击行为。

建议 API：

```go
func Notify(ctx context.Context, opts ...NotificationOption) error

func NotificationTitle(title string) NotificationOption
func NotificationBody(body string) NotificationOption
func NotificationKindStyle(kind NotificationKind) NotificationOption
func NotificationIcon(path string) NotificationOption
func NotificationGroup(group string) NotificationOption
func NotificationTimeout(timeout time.Duration) NotificationOption
func NotificationOnClick(fn func(NotificationEvent)) NotificationOption
```

v1 验收：

- Windows v1 调用 `Notify` 可显示托盘气泡通知；系统通知区不可用时返回包装后的 `ErrUnavailable`。
- 托盘气泡路径不要求安装程序或 AppUserModelID。
- Toast 路径如果系统 AppID/Toast 服务不可用，必须明确返回不支持或不可用，不假装成功。
- 点击回调只在 driver 能稳定收到真实事件时调用；Windows 当前处理 `NIN_BALLOONUSERCLICK`。

### Phase F: Tray 托盘

目标：为后台驻留类桌面应用提供 Windows 托盘图标、菜单和事件。

当前状态：F1-F10 已完成。公共 API、driver 边界、非 Windows/未实现平台错误语义、生命周期状态机、菜单回调规则、Windows 长期托盘图标、右键菜单、退出清理、示例、文档和自动化验收已经落地。Windows driver 当前声明 `CapabilityTray`。

任务：

- F1 已完成：定义 `Tray`、`TrayOption`、`TrayMenu`、`TrayMenuItem`、`TrayEvent` 和图标、tooltip、菜单、点击、双击 option。
- F2 已完成：新增 `trayDriver` 与 `trayHandle` 边界，`NewTray(opts ...TrayOption)` 通过 active driver 和 `CapabilityTray` 分发。
- F3 已完成：未实现平台或未声明能力的 driver 返回包装后的 `ErrUnsupported`，非 Windows 平台保持可编译。
- F4 已完成：公共 `Tray` 外壳提供 `SetIcon`、`SetTooltip`、`SetMenu`、`Show`、`Hide`、`Close` 生命周期状态机；关闭后操作返回 `ErrClosed`。
- F5 已完成：菜单模型支持普通项、分隔线、禁用项和勾选项；点击、双击和菜单项回调在公共层包装为异步派发。
- F6 已完成：新增长期 Tray 专用隐藏消息窗口和独立 message pump，托盘 Win32 回调不复用 Gio 主事件循环。
- F7 已完成：Windows 使用 `Shell_NotifyIconW` 实现长期托盘图标的 add/modify/delete，`SetIcon`、`SetTooltip`、`Show`、`Hide` 和 `Close` 已接入 native driver。
- F8 已完成：Windows 右键菜单已实现，支持菜单项启用/禁用、分隔线、勾选项和 item ID 到 Go callback 的映射。
- F9 已完成：新增 `docs/guides/tray-api.md`，更新 `docs/guides/system-api.md`、`docs/README.md` 和 `examples/system_showcase`。
- F10 已完成：补齐公共层和 Windows 自动化测试，补充 Explorer `TaskbarCreated` 重启恢复策略说明；Windows 人工验收清单已写入文档。

建议 API：

```go
type Tray struct{}

type TrayMenu []TrayMenuItem

type TrayMenuItem struct {
    ID        string
    Label     string
    Disabled  bool
    Checked   bool
    Separator bool
    OnClick   func(TrayEvent)
}

func NewTray(opts ...TrayOption) (*Tray, error)
func TrayIcon(path string) TrayOption
func TrayTooltip(text string) TrayOption
func TrayMenuItems(items ...TrayMenuItem) TrayOption
func TrayOnClick(fn func(TrayEvent)) TrayOption
func TrayOnDoubleClick(fn func(TrayEvent)) TrayOption
func TrayMenuAction(id, label string, onClick func(TrayEvent)) TrayMenuItem
func TrayMenuSeparator() TrayMenuItem
func (t *Tray) SetIcon(path string) error
func (t *Tray) SetTooltip(text string) error
func (t *Tray) SetMenu(menu TrayMenu) error
func (t *Tray) Show() error
func (t *Tray) Hide() error
func (t *Tray) Close() error
```

验收：

- 创建、更新、隐藏、关闭托盘不泄漏资源。
- Explorer 重启后能尝试恢复仍处于 visible 状态的图标。
- 菜单点击回调线程安全，不直接在 Win32 回调里操作 Gio 布局状态。
- 应用退出时自动清理托盘图标。

### Phase F+: System Stabilization 系统能力稳定化

目标：在进入更多桌面系统能力前，先补齐 A-F 暴露出的稳定性、易用性和生命周期缺口。

当前状态：第一批稳定化已完成。已新增能力探测 API、File Dialog 默认扩展名和最近目录记忆、File Dialog 结构化错误、MessageBox 异步入口和显示后 context 取消、Tray 状态查询、动态菜单、子菜单、默认项，以及 app lifecycle 退出时的统一托盘清理。

已完成：

- F+1 已完成：新增 `CapabilityStatus`、`CapabilityAvailability`、`system.Probe(cap)` 和 `system.Availability(cap)`，可区分 unsupported、unavailable、available。Windows driver 已对 File Dialog COM 创建、MessageBox native 入口、通知/托盘/系统事件隐藏窗口做轻量真实探测；没有深度 driver probe 时，已声明能力默认视为 available，具体调用仍可返回 `ErrUnavailable`。
- F+2 已完成：File Dialog 新增 `FileDialogDefaultExtension(value)`、`FileDialogRememberDir(key)`、`FileDialogError` 和 `IsFileDialogErrorKind`，Windows `IFileDialog.SetDefaultExtension` 已接入，默认目录错误会携带 `FileDialogErrorDefaultDir`。
- F+3 已完成：MessageBox 新增 `ShowMessageBoxAsync`、`MessageBoxResponse`，`ui` 层新增自动注入 owner 的异步 wrapper；Windows `TaskDialogIndirect` 路径在显示后监听 context 取消并向原生 dialog 投递关闭请求，`TaskDialog` / `MessageBoxW` fallback 通过锁定调用线程和枚举线程 dialog 做 best-effort 显示后取消。
- F+4 已完成：Tray 新增 `Visible()`、`Closed()`、`TrayMenuProvider`、`SetMenuProvider`、菜单 `Default` 和 `Children` 字段；Windows 右键菜单支持子菜单和默认项。
- F+5 已完成：新增 `system.CloseTrays()`，`app.runSpecs` 在窗口全部结束后统一清理注册中的托盘。

仍待继续：

- F+6 已完成：Window close requested / 关闭拦截。Windows 下通过 native `WM_CLOSE` hook 支持 allow/cancel，并产生 `WindowEventCloseRequested`；非 Windows 平台暂不承诺可取消系统关闭。
- F+7 第一批已完成：Window 新增 `RequestFocus`、`SetPosition`、`SetResizable`、`SetDecorated`，`ui` 层新增当前窗口 helper，app option 新增 `Decorated` / `Resizable`；Windows 下焦点、位置和可调整大小通过 native HWND 实现。
- F+8 已完成第一版：新增 `CapabilitySystemEvents`、`SystemEvent`、`SubscribeSystemEvents` 和 `SystemEventSubscription`。Windows 通过隐藏消息窗口接收 display、DPI、theme/settings、power、session 相关消息；macOS / Linux 已补轮询型最小 driver，按平台命令或系统路径比较 display/DPI/theme/settings/power/session 快照变化。
- F+8.1 已完成：窗口事件从单一 pull model 扩展为 pull + subscription。`WindowHandle.SubscribeEvents(kinds...)` 可订阅 size/scale/focus/state/close_requested/closed 事件，`PollEvents()` 保持兼容；`WindowEventScaleChanged` 与 `WindowState.Scale` / `DPI` 补充 per-window DPI/scale 精度。
- F+9 已完成第一版并扩展：Notification 新增 `NotificationOnDismiss`、`NotificationOnAction`、`NotificationBackendPath`、`NotificationAppID`、`NotificationLaunchURI`、`NotificationActions`、`ProbeNotificationBackend`、`CancelNotificationGroup`。Windows 同组托盘气泡和 Toast 会在新通知显示前清理前一个 group 通知；Toast 请求路径、backend 级探测、action button XML、协议 URI 激活、AppUserModelID 开始菜单快捷方式注册、ToastActivatorCLSID shortcut 前置属性、COM LocalServer 注册、当前进程 `INotificationActivationCallback` 分发和短生命周期实时 click/action/dismiss 回调已落地。macOS / Linux 已补命令型最小 notification driver；macOS 使用 `osascript display notification`，Linux 使用 `notify-send` / `kdialog`，其中 `notify-send` 可记录 id 做 group replace，cancel 需要 `gdbus`。Windows Toast backend 可用时 `NotificationBackendProbe.SupportsDurableActivation=true`；真实通知中心点击到具体应用启动模式仍需按应用命令行人工点验。
- F+10 已完成并扩展：Tray 支持 `TrayIconBytes`、`TrayIconResource`、`SetIconBytes` 和 `SetIconResource`，Windows 可从 `.ico` bytes 或进程资源 id 加载图标；macOS `osascript` + AppKit `NSStatusItem` 命令型 driver 支持路径/bytes 图标、tooltip、菜单/子菜单、显示隐藏、关闭清理和点击/菜单项回调，缺少 `osascript` 或脚本无法创建状态栏项时返回 `ErrUnavailable`，resource icon 和双击事件仍不作为 v1 承诺；Linux `yad --notification` 命令型 driver 支持路径/bytes 图标、tooltip、扁平菜单、显示隐藏、关闭清理和点击/菜单项回调，缺少 `yad` 时返回 `ErrUnavailable`，resource icon 仍返回 `ErrUnsupported`。
- F+11 已完成文档入口：新增 `docs/guides/system-api-windows-validation.md`，区分自动化已通过和 Windows GUI 人工点验待执行项。
- F+12 已完成平台规划：新增 `docs/guides/system-api-platform-plan.md`，规划 macOS/Linux 最小 driver 的接口映射、实现顺序和错误语义。

### Phase G: Clipboard、Shell 与后续系统能力

目标：在核心窗口、文件、弹窗、通知、托盘稳定后，补充常见桌面系统能力。

当前状态：Clipboard / Shell / Drag & Drop probe v1 已完成，macOS / Linux File Dialog 与 MessageBox 命令 driver v1 已补充。Windows driver 已声明 `CapabilityClipboard`、`CapabilityShell` 和 `CapabilityDragAndDrop`，支持 Unicode 文本剪贴板读写、Explorer 文件列表剪贴板、PNG 图片剪贴板、默认程序打开 URL/文件/目录、Explorer 定位路径、Gio transfer 拖放能力探测，以及 `ErrInvalidTarget` / `ErrTargetNotFound` / `ErrNoDefaultHandler` / `ErrAccessDenied` 等更细 Shell 错误分类；macOS / Linux driver 已声明 `CapabilityClipboard`、`CapabilityFileDialog`、`CapabilityMessageBox`、`CapabilityShell` 和 `CapabilityDragAndDrop`，支持通过系统命令读写文本剪贴板、显示基础文件选择和系统消息框、打开 URL/文件/目录并提供最小 reveal 行为，并已在 `OpenPath`、`RevealPath` 和 `OpenURL(file://...)` 提交前覆盖本地目标缺失与权限拒绝分类。Drag & Drop 的真实收发由 `widget.DropTarget` / `widget.DragSource` 基于 Gio `io/transfer` 完成；system 层只负责 probe。后续 macOS/Linux 图片/文件列表剪贴板、默认 handler/桌面环境启动后的更细错误分类和 File Dialog / MessageBox owner/modal 原生 toolkit driver 仍可继续扩展。

候选能力：

- Clipboard：文本读取/写入已完成；Windows 文件列表和 PNG 图片剪贴板已完成；macOS/Linux 文件列表和图片剪贴板后续扩展。
- Shell open：默认程序打开 URL、文件和目录已完成。
- Reveal in file manager：Windows Explorer 定位文件已完成；macOS Finder `open -R` 已完成；Linux v1 打开父目录，后续可按桌面环境补更精确定位。
- Drag and drop：`system.ProbeDragAndDrop` 已提供能力探测；`widget.DropTarget` / `ui.DropTargetElement` 接收 v1 已完成，支持 URI list / 文本 payload、本地路径解析、active/error 回调、大小限制和应用级 operation；`widget.DragSource` / `ui.DragSourceElement` 应用内拖拽源 v1 已完成，支持文本、`text/uri-list` 文件 URI、自定义 MIME payload、生命周期事件、禁用和应用级 operation。跨应用拖出不再默认承诺，只有后端明确支持时才会通过 `SupportsExternalDragOut` 报告。
- Global shortcuts：Windows v1 已完成，后续补 macOS/Linux driver、冲突诊断和示例。
- Single instance：Windows/macOS/Linux loopback v1 已完成，后续补主窗口恢复/置前和 URI/file handoff 示例。
- Power/session events：锁屏、解锁、睡眠、恢复。
- Theme/accent：系统深浅色、强调色变化。

这些能力不应抢在 Phase C-F 前面实现，除非具体应用场景强依赖。

### Phase H: 文档、示例与回归

目标：让系统 API 和现有渲染层一样，有可学习、可验证、可维护的文档体系。

任务：

- 新增 `docs/guides/system-api.md`，作为文档浏览器中的系统 API 总入口。
- 为 Window、File Dialog、MessageBox、Notification、Tray 分别补充独立文档或章节。
- 新增 `examples/system_showcase`，集中展示窗口控制、文件选择、消息框、通知、托盘和系统事件。
- 示例必须根据 `system.Supports` 做能力禁用提示，不能在非 Windows 平台崩溃。
- Windows-only 行为测试放入带 build tag 的测试文件。
- 不把系统 UI 纳入视觉回归，但要提供人工验收清单。

验收：

```sh
go test ./...
go vet ./...
go test ./examples/system_showcase
go run ./examples/system_validation
```

Windows 本地人工验收：

- 打开文件、保存文件、选择目录、多选文件。
- 信息、警告、错误、确认消息框。
- 托盘图标显示、菜单点击、退出清理。
- 通知或托盘气泡显示。
- 多窗口 owner 行为正确。

### Phase I: 社区平台实现

目标：让非 Windows 平台能在统一 API 下逐步补实现，而不是各自发明接口。

平台策略：

- macOS：MessageBox 和 Notification v1 已用 `osascript` 命令实现，文件选择可映射 `NSOpenPanel` / `NSSavePanel`，通知后续可升级到 UserNotifications，托盘可映射 status item。
- macOS：MessageBox / Notification v1 已用 `osascript` 命令实现，Tray v1 已用 `osascript` + AppKit `NSStatusItem` 命令型 driver 实现；后续可升级到更完整的 AppKit driver，并补 owner/modal、UserNotifications 和原生事件流。
- Linux：MessageBox v1 已用 `zenity` / `kdialog` 命令实现，Notification v1 已用 `notify-send` / `kdialog` 命令实现，Tray v1 已用 `yad --notification` 命令实现，文件选择可优先考虑 xdg-desktop-portal，托盘后续可升级到 StatusNotifierItem/AppIndicator，通知后续可升级到 freedesktop DBus notifications。
- Android/iOS：按移动端能力重新评估，不强行照搬桌面托盘。
- Web/js：文件选择和通知需走浏览器权限模型，API 可能只提供受限实现。

社区贡献要求：

- 新平台必须先补能力表和 unsupported 测试，再实现具体能力。
- 每个平台实现必须有最小示例和文档说明。
- 平台专属限制必须写在文档里，不通过隐藏 fallback 掩盖。
- 公共 API 变更必须先更新本路线图或对应设计文档。

## 能力覆盖矩阵

| 能力 | 公共 API | Windows 实现 | 非 Windows 空实现 | ui 便利封装 | 文档 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| 基础窗口控制 | 已有 | 已有 Gio 路径 | 部分可用 | 已有 | 已集中整理 | 稳定化中 |
| 多窗口 | 已有 | 已有 Gio 路径 | 受平台限制 | 已有 | 已集中整理 | 稳定化中 |
| 窗口属性查询 | 已有 `WindowState` / `State()` | 已有 runtime 快照 | 部分可用 | 已有 | 已集中整理 | 稳定化中 |
| 原生窗口句柄 | 已有 `NativeHandle()` | 已有 HWND 捕获 | 返回 false | 已有 | 已集中整理 | Windows v1 已完成 |
| 窗口事件 | 已有 `PollEvents()` / `SubscribeEvents()` | 已有 runtime pull + subscription，含 scale 事件 | 部分可用 | 已有 | 已集中整理 | 稳定化中 |
| 系统能力探测 | 已有 `Probe` / `Availability` | 已有轻量真实探测 | 已有 unsupported 状态 | 不适用 | 已补充 | F+ 已完成第一批 |
| 文件打开 | 已有 | 已有 Common Item Dialog | macOS/Linux 已有最小命令 driver，其它平台返回 `ErrUnsupported` | 已有自动 owner wrapper | 已补充 | Windows v1 已完成；macOS/Linux command v1 已完成 |
| 文件保存 | 已有 | 已有 Common Item Dialog | macOS/Linux 已有最小命令 driver，其它平台返回 `ErrUnsupported` | 已有自动 owner wrapper | 已补充 | Windows v1 已完成；macOS/Linux command v1 已完成 |
| 目录选择 | 已有 | 已有 Common Item Dialog | macOS/Linux 已有最小命令 driver，其它平台返回 `ErrUnsupported` | 已有自动 owner wrapper | 已补充 | Windows v1 已完成；macOS/Linux command v1 已完成 |
| 系统消息框 | 已有 | 已有 `TaskDialog` 优先，`MessageBoxW` 回退 | macOS/Linux 已有最小命令 driver，其它平台返回 `ErrUnsupported` | 已有自动 owner wrapper | 已补充 | Windows v1 待人工点验；macOS/Linux command v1 已完成 |
| 系统通知 | E1-E10 已完成 | 已有 `Shell_NotifyIconW` 托盘气泡和 Toast 请求路径 | macOS/Linux 已有最小命令 driver，其它平台返回 `ErrUnsupported` | 暂缓 | 已补充 | Windows/macOS/Linux v1 已完成 |
| 系统事件订阅 | 已有 `SubscribeSystemEvents` | 已有隐藏消息窗口接收系统消息 | macOS/Linux 已有最小轮询 driver，其它平台返回 `ErrUnsupported` | 不适用 | 已补充 | Windows/macOS/Linux v1 已完成 |
| 托盘 | F1-F10 已完成 | 已有 `Shell_NotifyIconW` 长期图标、bytes/resource 图标和右键菜单 | macOS 已有 `osascript` + AppKit `NSStatusItem` 最小命令 driver；Linux 已有 `yad --notification` 最小命令 driver；其它平台返回 `ErrUnsupported` | 不适用 | 已补充 | Windows v1 已完成；macOS/Linux command v1 已完成 |
| 剪贴板 | 已有文本读写 API | 已有 Windows Clipboard cmdlets 文本路径 | macOS/Linux 已有文本最小 driver，其它平台返回 `ErrUnsupported` | 暂缓 | 已补充 | Windows/macOS/Linux v1 已完成 |
| Shell open | 已有 URL/路径打开与定位 API | 已有 `ShellExecuteW` / Explorer reveal | macOS/Linux 已有最小 driver，其它平台返回 `ErrUnsupported` | 暂缓 | 已补充 | Windows/macOS/Linux v1 已完成 |
| Drag & Drop probe | 已有 `ProbeDragAndDrop` | 已报告 Gio transfer 能力 | macOS/Linux 已报告 Gio transfer 能力，其它平台返回 `ErrUnsupported` | 不适用 | 已补充 | system probe v1 已完成；真实拖放由 widget/ui 层提供 |

## 下一批推荐任务

1. 继续 Phase F+：执行并回填 Windows GUI 人工点验记录；命令行 validation 默认探测、dialog 显示后 context 取消、owner modal 禁用、notification group cancel/replace、Toast action button 提交/取消、Toast shortcut/activator 注册、tray bytes/path/resource icon、Explorer `TaskbarCreated` 恢复路径、native `WM_CLOSE` close-request hook 和 system events 订阅已通过，Toast 实时 callbacks 的用户点击路径和真实通知中心点击到具体应用启动模式仍需点击确认。
2. 继续 Phase F+：补充 Windows 打包/安装器专项文档，明确 common-controls v6 manifest、AppUserModelID shortcut、ToastActivatorCLSID、COM LocalServer 命令行和应用启动路由之间的发布约束。
3. 继续 Phase G / Phase I 扩展：补 macOS/Linux Clipboard 图片/文件列表、Shell 默认 handler/桌面环境启动后的更细错误分类，并继续推进 macOS Tray 原生 AppKit driver、Linux Tray 原生 StatusNotifier/AppIndicator、System Events 原生事件流、Notification 原生 DBus/UserNotifications 升级，以及 File Dialog / MessageBox owner/modal 原生 toolkit 或 portal driver；Drag & Drop 接收和应用内拖拽源 v1 已落地，真实跨应用拖入仍需按平台人工点验，跨应用拖出待后端明确支持后再开放。

## 验收命令

系统 API 公共层每次改动至少运行：

```sh
go test ./...
go vet ./...
go run ./examples/system_validation
go run ./examples/system_validation -dialogs
go run ./examples/system_validation -messagebox-sources
go run ./examples/system_validation -owner-modal
go run ./examples/system_validation -toast
go run ./examples/system_validation -toast-shortcut -toast-app-id com.example.FluxUI -toast-activator-clsid "{01234567-89AB-CDEF-0123-456789ABCDEF}"
go run ./examples/system_validation -toast-activator -toast-activator-clsid "{01234567-89AB-CDEF-0123-456789ABCDEF}"
```

涉及 Windows driver 时追加：

```sh
go test ./system ./internal/...
go test ./examples/system_showcase
go run ./examples/system_validation -dialogs
go run ./examples/system_validation -messagebox-sources
go run ./examples/system_validation -owner-modal
go run ./examples/system_validation -notify
go run ./examples/system_validation -toast
go run ./examples/system_validation -toast-shortcut -toast-app-id com.example.FluxUI -toast-shortcut-name "FluxUI Validation Activator" -toast-activator-clsid "{01234567-89AB-CDEF-0123-456789ABCDEF}"
go run ./examples/system_validation -toast-activator -toast-activator-clsid "{01234567-89AB-CDEF-0123-456789ABCDEF}"
go run ./examples/system_validation -tray
go run ./examples/system_validation -tray -tray-resource-id 2
go run ./examples/system_validation -clipboard
go run ./examples/system_validation -clipboard-files
go run ./examples/system_validation -clipboard-image
go run ./examples/system_validation -shell
go run ./examples/system_validation -shell-errors
go run ./examples/system_validation -drag-drop
go run ./examples/drag_drop_showcase
go run ./examples/system_validation -events 30s
```

涉及非 Windows unsupported 占位时追加最小交叉编译：

```sh
GOOS=linux GOARCH=amd64 go test -c ./system
GOOS=linux GOARCH=amd64 go test -c ./widget
GOOS=linux GOARCH=amd64 go test -c ./examples/system_validation
GOOS=darwin GOARCH=amd64 go test -c ./system
GOOS=darwin GOARCH=amd64 go test -c ./widget
GOOS=darwin GOARCH=amd64 go test -c ./examples/system_validation
```

涉及窗口生命周期、回调、托盘、通知时追加人工验收：

- 多窗口创建、关闭、最小化、最大化、全屏、置顶。
- 系统弹窗 owner 行为。
- 文件路径中文与长路径。
- 托盘图标创建、菜单点击、退出清理。
- Explorer 重启后的托盘图标恢复策略。

## 风险与约束

- Windows owner modal 依赖 Gio `Win32ViewEvent` 提供 HWND；如果后续 Gio 事件语义变化，需要同步调整 `WindowHandle.NativeHandle()`。
- Windows file dialog 依赖 COM apartment，必须把线程初始化、释放和回调边界封装在 driver 内。
- 托盘需要 Win32 消息回调，不能破坏 Gio 主事件循环，也不能在系统回调线程直接修改 UI 状态。
- Windows toast notification 的打包、AppUserModelID、快捷方式和持久激活回调复杂度较高；当前已完成请求路径、action button XML、协议 URI 激活、ToastActivatorCLSID shortcut 前置属性、COM LocalServer 注册、当前进程 `INotificationActivationCallback` 分发和短生命周期实时回调。真实通知中心点击到具体应用启动模式仍依赖应用自己的命令行入口、安装器和人工点验。
- 非 Windows 平台不能因为未实现而编译失败；空实现是长期兼容策略的一部分。
- 系统 API 的阻塞行为必须清晰文档化，避免用户在布局期间调用导致卡帧或死锁。
- 公共 API 一旦发布，后续平台实现会依赖它；命名和错误语义需要比具体 Windows 实现更谨慎。

## 长期完成定义

FluxUI 可以认为系统层 API 达到稳定状态，需要满足：

- `system` 包有清晰的公共 API、能力检测、错误语义和平台 driver 边界。
- Windows 已实现窗口增强、文件选择、系统消息框、托盘和至少一种系统通知路径。
- 非 Windows 平台全部可编译，并对未实现能力返回 `ErrUnsupported`。
- `ui` 层提供少量便利函数，但不把系统服务伪装成 Widget。
- 文档和示例能说明系统 API 与自绘 Dialog/Toast 的边界。
- 多窗口 owner、关闭生命周期、系统回调线程和资源释放都有测试或人工验收清单。
- 后续 macOS、Linux、移动端和 Web 能在不改公共 API 的前提下补 driver。
