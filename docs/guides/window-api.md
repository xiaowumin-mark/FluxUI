<!-- fluxui-doc-meta
{
  "id": "window_api",
  "title": "Window API 窗口能力",
  "category": "使用指南",
  "order": 123,
  "summary": "Window API 提供 FluxUI 的窗口创建、窗口控制、状态查询和尺寸约束能力。",
  "apis": [
    "WindowElement(root Component, opts ...AppOption) WindowSpec",
    "RunElementMulti(windows ...WindowSpec) error",
    "ListWindows() []WindowHandle",
    "GetWindow(id WindowID) (WindowHandle, bool)",
    "(*WindowHandle).NativeHandle() (uintptr, bool)",
    "CurrentWindowNativeHandle(ctx *Context) (uintptr, bool)",
    "(*WindowHandle).State() (WindowState, bool)",
    "(*WindowHandle).PollEvents() []WindowEvent",
    "(*WindowHandle).SetMinSize(width, height int) bool",
    "(*WindowHandle).SetMaxSize(width, height int) bool",
    "MinSize(width, height int) AppOption",
    "MaxSize(width, height int) AppOption",
    "WindowSetMinSize(ctx *Context, width, height int) bool",
    "WindowSetMaxSize(ctx *Context, width, height int) bool"
  ]
}
-->

# Window API 窗口能力

Window API 是 FluxUI 系统层能力的第一批基础设施。它负责窗口创建、多窗口启动、当前窗口控制、运行中窗口查询和窗口状态快照。

当前窗口能力仍由 `app` runtime 承载，`ui` 层提供 React-style Element API 和当前窗口便利函数。文件选择器、系统消息框和托盘等后续能力会复用这里的窗口句柄和状态模型。

完整示例见 `examples/window_showcase`，它展示了多窗口启动、状态快照、事件拉取、标题/尺寸修改、尺寸约束和常用窗口动作。

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
- `WindowMinimize(ctx)`
- `WindowMaximize(ctx)`
- `WindowRestore(ctx)`
- `WindowFullscreen(ctx)`
- `WindowRaise(ctx)`
- `WindowCenter(ctx)`
- `WindowSetTitle(ctx, title)`
- `WindowSetSize(ctx, width, height)`
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
- `Close()`
- `Minimize()`
- `Maximize()`
- `Restore()`
- `Fullscreen()`
- `Raise()`
- `Center()`
- `SetTitle(title)`
- `SetSize(width, height)`
- `SetMinSize(width, height)`
- `SetMaxSize(width, height)`
- `Invalidate()`

`NativeHandle()` 返回平台原生窗口句柄。Windows 下该值是 `HWND`，主要用于把 `FileDialogOwner` / `MessageBoxOwner` 绑定到当前 FluxUI 窗口；非 Windows 平台、窗口尚未收到原生 view event 或窗口已关闭时返回 `false`。

在组件事件回调中，也可以使用 `CurrentWindowNativeHandle(ctx)` 查询当前窗口的 native handle。普通文件选择器和消息框调用不需要手动查询；`ui.OpenFileDialog`、`ui.SaveFileDialog`、`ui.PickFolderDialog` 和 `ui.ShowMessageBox` 会自动绑定当前窗口 owner。

窗口关闭后，`State()` 和 `NativeHandle()` 返回 false，控制方法返回 false。

## WindowEvent

`PollEvents()` 返回并清空窗口积累的事件：

```go
events := handle.PollEvents()
for _, event := range events {
    switch event.Kind {
    case ui.WindowEventSizeChanged:
        fmt.Println("size:", event.State.Width, event.State.Height)
    case ui.WindowEventFocusChanged:
        fmt.Println("focused:", event.State.Focused)
    case ui.WindowEventStateChanged:
        fmt.Println("state changed:", event.State.Title)
    case ui.WindowEventClosed:
        fmt.Println("closed:", event.State.ID)
    }
}
```

第一版事件模型是轻量 pull model，不提供跨线程回调总线。这样可以避免 Win32/Gio 回调线程直接修改 UI 状态。当前事件类型包括：

- `WindowEventSizeChanged`
- `WindowEventFocusChanged`
- `WindowEventStateChanged`
- `WindowEventClosed`

事件来自 FluxUI runtime 状态同步，不是独立的 OS 事件订阅系统。后续如果需要 close request、display changed、DPI changed、theme changed，会在确认 Gio/平台能力后再扩展。

## WindowState

`WindowState` 是 FluxUI runtime 维护的窗口状态快照：

```go
type WindowState struct {
    ID         WindowID
    Title      string
    Width      int
    Height     int
    MinWidth   int
    MinHeight  int
    MaxWidth   int
    MaxHeight  int
    Minimized  bool
    Maximized  bool
    Fullscreen bool
    Focused    bool
    Decorated  bool
    Alive      bool
}
```

尺寸字段使用 dp。FluxUI 会在自身发出窗口命令时同步状态，并在收到 Gio `ConfigEvent` / `FrameEvent` 时尽量校正快照。它不是直接 OS 查询结果；平台可能忽略或调整窗口 option。

## B1-B5 当前结论

- B1 现有窗口 API 已集中梳理：窗口创建、多窗口、当前窗口 helper 和 `WindowHandle` 继续兼容。
- B2 已新增 `WindowState` 和 `WindowHandle.State()`。
- B3 已新增 `MinSize` / `MaxSize` app option，以及 `WindowHandle.SetMinSize` / `SetMaxSize` 和当前窗口 helper。
- B4 已调研 Gio v0.9：没有公开的 always-on-top 或 request focus 窗口 API，因此当前不暴露 `SetAlwaysOnTop` / `RequestFocus`。
- B5 已调研 Gio v0.9：没有可在 `DestroyEvent` 前稳定拦截 OS close 的公共 API，因此当前不实现关闭拦截。需要确认关闭前提示时，请使用应用内退出按钮配合 `Dialog`，不要承诺拦截系统关闭按钮。
- B6 已新增轻量窗口事件模型：`WindowHandle.PollEvents()`、`WindowEvent` 和 size/focus/state/closed 事件。
- B7 已完成 native owner 接入：Windows 下通过 Gio `Win32ViewEvent` 捕获 HWND，并由 `WindowHandle.NativeHandle()` 暴露为 `uintptr`；非 Windows 平台返回 false。
- B8 已新增本指南作为窗口 API 集中入口。
- B9 已新增 `examples/window_showcase`。

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
- 最小化、最大化、还原、全屏后状态快照正确。
- `PollEvents()` 能收到并清空 size/focus/state/closed 事件。
- 多窗口 `ListWindows` 和 `GetWindow` 不串窗口。
- Windows 下 `NativeHandle()` 返回非 0 HWND，关闭后返回 false。
- 关闭后 `WindowHandle` 不再可操作。
