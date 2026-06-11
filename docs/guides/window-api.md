<!-- fluxui-doc-meta
{
  "id": "window_api",
  "title": "Window API 窗口能力",
  "category": "使用指南",
  "order": 123,
  "summary": "Window API 提供 FluxUI 的窗口创建、窗口控制、状态查询和尺寸约束能力。",
  "example": { "id": "system_api_basic" },
  "apis": [
    "WindowElement(root Component, opts ...AppOption) WindowSpec",
    "RunElementMulti(windows ...WindowSpec) error",
    "ListWindows() []WindowHandle",
    "GetWindow(id WindowID) (WindowHandle, bool)",
    "CurrentWindowID(ctx *Context) WindowID",
    "(*WindowHandle).NativeHandle() (uintptr, bool)",
    "CurrentWindowNativeHandle(ctx *Context) (uintptr, bool)",
    "(*WindowHandle).State() (WindowState, bool)",
    "(*WindowHandle).PollEvents() []WindowEvent",
    "(*WindowHandle).SubscribeEvents(kinds ...WindowEventKind) (*WindowEventSubscription, bool)",
    "WindowEventScaleChanged",
    "(*WindowEventSubscription).Events() <-chan WindowEvent",
    "(*WindowEventSubscription).Close() bool",
    "(*WindowHandle).SetCloseRequestedHandler(fn func(WindowCloseRequest) bool) bool",
    "(*WindowHandle).Show() bool",
    "(*WindowHandle).Hide() bool",
    "(*WindowHandle).RequestFocus() bool",
    "(*WindowHandle).SetAlwaysOnTop(always bool) bool",
    "(*WindowHandle).SetHiddenMemoryPolicy(policy WindowHiddenMemoryPolicy) bool",
    "(*WindowHandle).SetPosition(x, y int) bool",
    "(*WindowHandle).SetResizable(resizable bool) bool",
    "(*WindowHandle).SetDecorated(decorated bool) bool",
    "(*WindowHandle).SetMinSize(width, height int) bool",
    "(*WindowHandle).SetMaxSize(width, height int) bool",
    "HiddenMemoryPolicy(policy WindowHiddenMemoryPolicy) AppOption",
    "OnCloseRequested(fn func(WindowCloseRequest) bool) AppOption",
    "Decorated(enabled bool) AppOption",
    "Resizable(enabled bool) AppOption",
    "MinSize(width, height int) AppOption",
    "MaxSize(width, height int) AppOption",
    "WindowShow(ctx *Context) bool",
    "WindowHide(ctx *Context) bool",
    "WindowRequestFocus(ctx *Context) bool",
    "WindowSetAlwaysOnTop(ctx *Context, always bool) bool",
    "WindowSetHiddenMemoryPolicy(ctx *Context, policy WindowHiddenMemoryPolicy) bool",
    "WindowSetPosition(ctx *Context, x, y int) bool",
    "WindowSetResizable(ctx *Context, resizable bool) bool",
    "WindowSetDecorated(ctx *Context, decorated bool) bool",
    "WindowSetMinSize(ctx *Context, width, height int) bool",
    "WindowSetMaxSize(ctx *Context, width, height int) bool"
  ]
}
-->

# Window API 窗口能力

Window API 是 FluxUI 系统层能力的第一批基础设施。它负责窗口创建、多窗口启动、当前窗口控制、运行中窗口查询和窗口状态快照。

当前窗口能力仍由 `app` runtime 承载，`ui` 层提供 React-style Element API 和当前窗口便利函数。文件选择器、系统消息框和托盘等后续能力会复用这里的窗口句柄和状态模型。

完整示例见 `examples/window_showcase`，它展示了多窗口启动、状态快照、事件订阅、关闭请求拦截、关闭前“是否保存”系统消息框、标题/尺寸修改、尺寸约束、持续置顶、显示/隐藏、隐藏内存策略和常用窗口动作；`PollEvents()` 仍保留为兼容入口。

## 创建窗口

单窗口应用继续使用 `RunElement`：

```go
ui.RunElement(App, ui.Title("FluxUI"), ui.Size(640, 480))
```

多窗口应用使用 `WindowElement` 和 `RunElementMulti`：

```go
ui.RunElementMulti(
    ui.WindowElement(MainWindow, ui.Title("Main"), ui.Size(800, 600)),
    ui.WindowElement(ToolsWindow, ui.Title("Tools"), ui.Size(360, 480)),
)
```

`WindowElement` 会把 root component 和 app options 打包成 `WindowSpec`，由 `RunElementMulti` 启动。

## 初始尺寸约束

窗口初始尺寸使用 `Size`，最小和最大尺寸使用 `MinSize` / `MaxSize`：

```go
ui.RunElement(
    App,
    ui.Title("Editor"),
    ui.Size(900, 640),
    ui.MinSize(640, 420),
    ui.MaxSize(1600, 1000),
)
```

尺寸单位是 dp。非法尺寸不会在运行中应用；`MinSize` / `MaxSize` 作为 app option 时只记录配置，底层平台是否完全遵守仍由 Gio 和操作系统决定。

设置完整 `MaxSize(width, height)` 后，FluxUI 会把最大化视为不可用：`WindowMaximize(ctx)` / `WindowHandle.Maximize()` 返回 `false`，Windows 标题栏的最大化按钮和系统菜单最大化项也会同步禁用。全屏不受该限制影响，进入全屏时仍会临时清除尺寸约束以占满显示器。

## 当前窗口控制

组件事件回调中可以通过当前 `*ui.Context` 控制当前窗口：

```go
ui.FilledButtonElement(
    ui.TextElement("全屏"),
    ui.OnClick(func(ctx *ui.Context) {
        ui.WindowFullscreen(ctx)
    }),
)
```

当前窗口 helper 包括：

- `CurrentWindowID(ctx)`
- `WindowClose(ctx)`
- `WindowShow(ctx)`
- `WindowHide(ctx)`
- `WindowMinimize(ctx)`
- `WindowMaximize(ctx)`
- `WindowRestore(ctx)`
- `WindowFullscreen(ctx)`
- `WindowRaise(ctx)`
- `WindowRequestFocus(ctx)`
- `WindowSetAlwaysOnTop(ctx, always)`
- `WindowSetHiddenMemoryPolicy(ctx, policy)`
- `WindowCenter(ctx)`
- `WindowSetTitle(ctx, title)`
- `WindowSetPosition(ctx, x, y)`
- `WindowSetSize(ctx, width, height)`
- `WindowSetResizable(ctx, resizable)`
- `WindowSetDecorated(ctx, decorated)`
- `WindowSetMinSize(ctx, width, height)`
- `WindowSetMaxSize(ctx, width, height)`
- `WindowInvalidate(ctx)`
- `WindowIsAlive(ctx)`

这些函数没有当前窗口控制器时返回 false。

## WindowHandle

`WindowHandle` 表示运行中的窗口句柄。可以通过 `ListWindows` 或 `GetWindow` 获取：

```go
for _, handle := range ui.ListWindows() {
    state, ok := handle.State()
    if !ok {
        continue
    }
    fmt.Println(state.ID, state.Title)
}
```

`WindowHandle` 支持：

- `ID()`
- `IsAlive()`
- `NativeHandle()`
- `State()`
- `PollEvents()`
- `SubscribeEvents(kinds...)`
- `Close()`
- `SetCloseRequestedHandler(fn)`
- `Show()`
- `Hide()`
- `Minimize()`
- `Maximize()`
- `Restore()`
- `Fullscreen()`
- `Raise()`
- `RequestFocus()`
- `SetAlwaysOnTop(always)`
- `SetHiddenMemoryPolicy(policy)`
- `Center()`
- `SetTitle(title)`
- `SetPosition(x, y)`
- `SetSize(width, height)`
- `SetResizable(resizable)`
- `SetDecorated(decorated)`
- `SetMinSize(width, height)`
- `SetMaxSize(width, height)`
- `Invalidate()`

`NativeHandle()` 返回平台原生窗口句柄。Windows 下该值是 `HWND`，主要用于把 `FileDialogOwner` / `MessageBoxOwner` 绑定到当前 FluxUI 窗口；非 Windows 平台、窗口尚未收到原生 view event 或窗口已关闭时返回 `false`。

在组件事件回调中，也可以使用 `CurrentWindowNativeHandle(ctx)` 查询当前窗口的 native handle。普通文件选择器和消息框调用不需要手动查询；`ui.OpenFileDialog`、`ui.SaveFileDialog`、`ui.PickFolderDialog` 和 `ui.ShowMessageBox` 会自动绑定当前窗口 owner。

`Raise()` / `WindowRaise(ctx)` 是一次性前置请求，等价于让系统把窗口带到前面一次；它不是持续置顶。需要窗口保持在其他普通窗口之上时，使用 `SetAlwaysOnTop(true)` / `WindowSetAlwaysOnTop(ctx, true)`，取消时传 `false`。Windows 下持续置顶依赖 `HWND`，窗口尚未完成 native view 初始化或非 Windows 平台会返回 `false`。Windows native 操作会异步排队执行，避免在 Gio 帧处理期间同步阻塞窗口线程。

`SetCloseRequestedHandler(fn)` 和 `OnCloseRequested(fn)` 用于拦截用户发起的关闭请求。handler 返回 `true` 表示允许关闭，返回 `false` 表示取消关闭；每次关闭请求都会写入 `WindowEventCloseRequested` 事件。Windows 下 FluxUI 会在拿到 `HWND` 后安装 `WM_CLOSE` hook，因此标题栏关闭按钮、Alt+F4 和系统菜单关闭可以被取消。非 Windows 平台目前没有 native close-request hook，`SetCloseRequestedHandler` 仍会记录 handler，但不承诺能阻止系统关闭。

关闭确认建议采用“先取消、再异步弹出确认 UI”的方式：handler 里快速返回 `false`，在 goroutine 或业务状态中打开确认框；用户确认后再允许关闭或走应用自己的退出路径。不要在 Win32 close hook 回调里长时间阻塞。`examples/window_showcase` 演示了通过 `system.ShowMessageBox` 异步弹出“是否保存”，选择 Yes/No 后再调用 `WindowHandle.Close()`。

`Show()` / `Hide()` 和 `WindowShow(ctx)` / `WindowHide(ctx)` 用于显示或隐藏窗口。隐藏当前唯一窗口后，用户无法再通过该窗口触发显示操作；请保留 `WindowHandle` 并通过计时器、托盘、另一个窗口或后台事件调用 `Show()`。Windows 下该能力依赖 `HWND`，不可用时返回 `false`；返回 `true` 表示操作已排队，不代表系统已经完成显示状态切换。

`RequestFocus()` / `WindowRequestFocus(ctx)` 请求当前窗口获得键盘焦点。Windows 下优先通过 native `SetForegroundWindow` 排队执行；如果 native handle 尚不可用，则退回到 Gio `ActionRaise` 请求。

`SetPosition(x, y)` / `WindowSetPosition(ctx, x, y)` 使用 dp 设置窗口左上角位置。Windows 下会按当前窗口 metric 转换为像素并调用 `SetWindowPos`；窗口尚未完成 native view 初始化时返回 `false`。

`SetResizable(resizable)` / `WindowSetResizable(ctx, resizable)` 控制 Windows 系统边框是否允许拖拽调整大小；当前依赖 `HWND`，非 Windows 或 native handle 不可用时返回 `false`。`Resizable(enabled)` 可设置初始意图，Windows native handle 捕获后会同步应用。

`SetDecorated(decorated)` / `WindowSetDecorated(ctx, decorated)` 控制是否显示系统装饰边框。该能力通过 Gio `Decorated` option 接入；`Decorated(enabled)` 可设置窗口初始装饰状态。

默认隐藏内存策略是 `WindowHiddenMemoryReleaseTransient`：隐藏窗口后 FluxUI 会暂停该窗口的 root layout 和 redraw invalidation，隐藏帧会清空当前 `op.Ops` 并异步触发 Go 内存回收，尽量释放公开层可安全释放的临时渲染内存。显示窗口时会恢复渲染并触发一次重绘。如果应用需要隐藏期间继续保留渲染状态，可以使用 `HiddenMemoryPolicy(WindowHiddenMemoryKeepRenderingState)` 或 `SetHiddenMemoryPolicy(WindowHiddenMemoryKeepRenderingState)`。

当前策略不会直接调用 Gio 内部 `destroyGPU()`，因此不承诺释放 D3D/OpenGL/Vulkan device、swapchain 或 Gio GPU cache。真正释放这些资源需要 Gio 公开 API 或维护定制 backend。

进入全屏时，FluxUI 会临时清除 Gio 层当前窗口的 `MinSize` / `MaxSize` 限制，让全屏尺寸可以占满显示器；退出全屏、设置普通窗口尺寸、最小化或最大化时，会把原先记录的尺寸约束重新应用回窗口。`WindowState.MinWidth` / `MaxWidth` 等字段保存的是应用请求的约束，不会因为全屏期间底层配置临时清空而丢失。

窗口关闭后，`State()` 和 `NativeHandle()` 返回 false，控制方法返回 false。

## WindowEvent

`PollEvents()` 返回并清空窗口积累的事件：

```go
events := handle.PollEvents()
for _, event := range events {
    switch event.Kind {
    case ui.WindowEventSizeChanged:
        fmt.Println("size:", event.State.Width, event.State.Height)
    case ui.WindowEventScaleChanged:
        fmt.Println("scale:", event.State.Scale, "dpi:", event.State.DPI)
    case ui.WindowEventFocusChanged:
        fmt.Println("focused:", event.State.Focused)
    case ui.WindowEventStateChanged:
        fmt.Println("state changed:", event.State.Title)
    case ui.WindowEventCloseRequested:
        fmt.Println("close requested:", event.State.ID)
    case ui.WindowEventClosed:
        fmt.Println("closed:", event.State.ID)
    }
}
```

窗口事件模型现在同时支持 `PollEvents()` 拉取和 `SubscribeEvents()` 订阅。订阅 channel 仍是 best-effort 投递，不会阻塞 Win32/Gio 事件线程；需要兼容旧调用点时继续使用 `PollEvents()`。当前事件类型包括：

- `WindowEventSizeChanged`
- `WindowEventScaleChanged`
- `WindowEventFocusChanged`
- `WindowEventStateChanged`
- `WindowEventCloseRequested`
- `WindowEventClosed`

事件来自 FluxUI runtime 状态同步，不是独立的 OS 事件订阅系统。`WindowEventScaleChanged` 来自当前窗口 Gio metric 的 `PxPerDp` / `PxPerSp` 变化，适合处理 per-window DPI/scale；显示器、系统主题、电源和会话变化仍属于 `system.SubscribeSystemEvents` 的系统事件订阅能力。

不想每帧轮询时，可以使用 `SubscribeEvents`：

```go
sub, ok := handle.SubscribeEvents(ui.WindowEventCloseRequested, ui.WindowEventClosed)
if ok {
    defer sub.Close()
    go func() {
        for event := range sub.Events() {
            // 根据 event.Kind 更新应用状态。
        }
    }()
}
```

不传事件类型表示订阅全部窗口事件。订阅 channel 是固定缓冲、best-effort 投递；如果调用方长期不消费，后续事件可能被丢弃，避免阻塞窗口事件循环。窗口关闭并从 registry 移除后，订阅 channel 会自动关闭。

## WindowState

`WindowState` 是 FluxUI runtime 维护的窗口状态快照：

```go
type WindowState struct {
    ID         WindowID
    Title      string
    X          int
    Y          int
    Width      int
    Height     int
    Scale      float32
    TextScale  float32
    DPI        int
    MinWidth   int
    MinHeight  int
    MaxWidth   int
    MaxHeight  int
    Visible            bool
    AlwaysOnTop        bool
    RenderSuspended    bool
    HiddenMemoryPolicy WindowHiddenMemoryPolicy
    Minimized  bool
    Maximized  bool
    Fullscreen bool
    Focused    bool
    Decorated  bool
    Resizable  bool
    Alive      bool
}
```

尺寸字段使用 dp。`Scale` 是当前窗口 `PxPerDp`，`TextScale` 是 `PxPerSp`，`DPI` 是按 `Scale * 96` 计算的近似 DPI。FluxUI 会在自身发出窗口命令时同步状态，并在收到 Gio `ConfigEvent` / `FrameEvent` 时尽量校正快照。它不是直接 OS 查询结果；平台可能忽略或调整窗口 option。

## B1-B5 当前结论

- B1 现有窗口 API 已集中梳理：窗口创建、多窗口、当前窗口 helper 和 `WindowHandle` 继续兼容。
- B2 已新增 `WindowState` 和 `WindowHandle.State()`。
- B3 已新增 `MinSize` / `MaxSize` app option，以及 `WindowHandle.SetMinSize` / `SetMaxSize` 和当前窗口 helper。
- B4 已补充 Windows 持续置顶能力：`WindowHandle.SetAlwaysOnTop(always)` 和 `WindowSetAlwaysOnTop(ctx, always)` 通过 native `HWND` 实现；`Raise` 保留为一次性前置请求。
- B5 已完成关闭请求拦截：Windows 下通过 native `WM_CLOSE` hook 支持标题栏关闭、Alt+F4 和系统菜单关闭的 allow/cancel；非 Windows 平台暂不承诺可取消系统关闭。
- B6 已新增窗口事件模型：`WindowHandle.PollEvents()`、`WindowHandle.SubscribeEvents()`、`WindowEvent` 和 size/scale/focus/state/close_requested/closed 事件。
- B7 已完成 native owner 接入：Windows 下通过 Gio `Win32ViewEvent` 捕获 HWND，并由 `WindowHandle.NativeHandle()` 暴露为 `uintptr`；非 Windows 平台返回 false。
- B8 已新增本指南作为窗口 API 集中入口。
- B9 已新增 `examples/window_showcase`，并覆盖窗口事件订阅、关闭请求拦截、关闭前“是否保存”系统消息框、持续置顶、一次性前置、隐藏后延迟显示、隐藏内存策略和全屏尺寸约束行为。
- B10 已补充 `Show` / `Hide`：Windows 下通过 native `HWND` 显示或隐藏窗口；非 Windows 平台当前返回 false。
- B11 已新增隐藏内存策略：默认 `WindowHiddenMemoryReleaseTransient` 会在隐藏后暂停 FluxUI 渲染、跳过隐藏窗口的 redraw invalidation、清空临时 `op.Ops` 并异步触发 Go 内存回收；`WindowHiddenMemoryKeepRenderingState` 可保留旧行为。
- B12 已修正最大尺寸限制下的最大化行为：设置完整 `MaxSize(width, height)` 后，API 最大化调用返回 false，Windows 原生标题栏最大化按钮和系统菜单最大化项同步禁用；全屏仍保留可用。
- F+7 第一批已完成：新增 `RequestFocus`、`SetPosition`、`SetResizable`、`SetDecorated`、`Decorated` 和 `Resizable`。Windows 下焦点、位置和可调整大小通过 native HWND 实现；装饰边框通过 Gio option 实现。
- F+8 已新增系统事件订阅入口：`system.SubscribeSystemEvents` 可订阅 display/theme/settings/power/session/DPI 事件，Windows 通过隐藏消息窗口接收可到达的系统消息；窗口级 `WindowEventScaleChanged` 用于补 per-window DPI/scale 精度。

## Native owner

文件选择器和系统消息框应绑定当前窗口 owner，避免原生弹窗成为无主窗口。Windows 下 FluxUI runtime 会在收到 Gio `Win32ViewEvent` 时记录 HWND，并通过 `WindowHandle.NativeHandle()` 提供给 `system` owner option。

如果直接调用 `system` 包，可以显式传 owner：

```go
owner, ok := handle.NativeHandle()
if ok {
    result, err := system.ShowMessageBox(ctx, system.MessageBoxOwner(owner))
    _ = result
    _ = err
}
```

普通 UI 代码更推荐使用 `ui` wrapper，让 FluxUI 自动处理 owner：

```go
result, err := ui.ShowMessageBox(ctx, system.MessageBoxTitle("FluxUI"))
_ = result
_ = err
```

当前结论是：

- 公共 API 只暴露 `uintptr`，不引入 HWND、NSWindow、X11 Window 等平台专属类型。
- Windows 下 `NativeHandle()` 返回 HWND，可用于 `FileDialogOwner` 和 `MessageBoxOwner`。
- `ui` 层的文件选择和消息框 wrapper 会自动注入当前窗口 owner。
- 非 Windows 平台、窗口尚未完成 native view 初始化或窗口已关闭时，`NativeHandle()` 返回 false。
- 未传 owner 或 owner 为 0 时，系统弹窗仍可显示，但不承诺严格 modal 到 FluxUI 主窗口。

## 验收

窗口 API 相关改动至少运行：

```sh
go test ./app ./internal ./ui
go vet ./app ./internal ./ui
```

涉及窗口运行时、状态同步或尺寸约束时，建议在 Windows 本地人工检查：

- `SetTitle` / `SetSize` 后 `State()` 同步。
- `SetMinSize` / `SetMaxSize` 对系统窗口尺寸约束生效。
- 最小化、最大化、还原、全屏后状态快照正确；设置最大尺寸后，全屏仍可占满显示器。
- Windows 下 `SetAlwaysOnTop(true/false)` 能切换持续置顶，`Raise()` 只做一次性前置。
- Windows 下 `Hide()` 可隐藏窗口，随后用保留的 `WindowHandle.Show()` 可显示回来。
- Windows 下 `RequestFocus()`、`SetPosition()`、`SetResizable(false/true)` 和 `SetDecorated(false/true)` 行为符合预期。
- 默认隐藏内存策略下，`Hide()` 后 `WindowState.RenderSuspended` 为 true，`Show()` 后恢复为 false。
- `PollEvents()` 能收到并清空 size/focus/state/close_requested/closed 事件。
- 多窗口 `ListWindows` 和 `GetWindow` 不串窗口。
- Windows 下 `NativeHandle()` 返回非 0 HWND，关闭后返回 false。
- 关闭后 `WindowHandle` 不再可操作。
