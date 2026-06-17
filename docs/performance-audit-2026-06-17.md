# FluxUI 性能审查报告 - 2026-06-17

## 范围

本次审查覆盖核心渲染/运行时路径和文档浏览器示例：

- 核心：`internal`、`ui`、`widget`、`state`、`system`、`app`。
- 示例：`examples/docs_browser`，重点为 `system_api_basic`。

## 已确认问题

### P0: `system_api_basic` 在渲染路径中执行系统能力探测

`examples/docs_browser/system_api_demo.go` 的能力卡片在每次构建时调用：

- `system.Capabilities()`
- `system.Availability(cap)`

`Availability` 是真实系统探测。Windows 下部分探测会初始化 COM、创建隐藏窗口，Clipboard 还会启动 PowerShell；macOS/Linux 下部分探测会查找或执行平台命令。把这些调用放在 UI 渲染路径中会阻塞 Gio frame loop。表现上会出现帧率急剧下降但 CPU 占用很低，因为线程更多是在等待系统调用/外部进程，而不是忙算。

修复策略：

- 首帧只渲染轻量占位状态。
- 挂载后在 goroutine 中执行真实 probe。
- probe 结果写入状态后触发一次重绘。
- 后续 frame 使用缓存结果，不再在渲染路径中调用 `system.Availability`。

### P1: `system.Supports` 每次调用都会拷贝 capability map

`system.Supports(cap)` 当前实现为 `Capabilities().Supports(cap)`，而 `Capabilities()` 为了保护外部不可变语义会 clone map。文档示例和业务 UI 常在每帧多处调用 `Supports` 控制按钮禁用状态，这会造成不必要分配。

修复策略：

- 在 system driver 设置时缓存 capability set。
- `Capabilities()` 继续返回副本，保持公共 API 语义。
- `Supports()` 直接读取缓存，不再分配。
- File Dialog、MessageBox、Notification、Tray、Clipboard、Shell 等 system API 入口内部也复用缓存能力判断，不再调用 driver 重新构造 capability map。

### P1: `Runtime.BeginFrame` 每帧重新分配 hook count map

`internal.Runtime.BeginFrame` 每帧通过 `make(map[string]int)` 创建新的 `hookCounts`，只为了保留上一帧用于 hook 规则检查。这是高频路径上的稳定分配。

修复策略：

- 初始化 `hookCounts` 和 `prevHookCounts` 两个 map。
- 每帧交换两个 map 并清空新的 `hookCounts`。
- 保持上一帧 hook 数量检查语义不变。

### P2: reconciler 子节点调和在无 key 场景也创建 key map

`ui.reconcileChildren` 每次都会创建 `byKey` map，即使绝大多数静态布局没有 key。docs_browser 大量使用 `RowElement`/`ColumnElement`，该分配会在每帧放大。

修复策略：

- 只有当前 children 中存在 keyed element 时才构建 keyed lookup map。
- 复用 `parent.Children` 的 slice backing，缩小每帧 slice 分配。
- 清理缩短后的尾部引用，避免旧节点被 backing array 保留。

## 观察到但本次不展开的风险

- `internal.Context` 通过字符串拼接维护 tree path，深层树会产生大量短字符串；这和 legacy path-based state 兼容性绑定较深，建议后续用结构化 path/slot id 单独设计。
- `RenderWithChildren`/`HostChildren` 在复杂 Element 树中仍有 per-frame slice 分配。可继续降低，但需要更系统地调整 composite element 接口。
- `Probe` 语义是 point-in-time availability，不建议在核心层全局缓存真实 probe 结果；缓存应由调用方或专门的带 TTL API 明确表达。

## 本次修复清单

- 核心运行时复用 hook count map。
- system layer 缓存 driver capability set，并让 `Supports` 零分配读取。
- reconciler 对无 keyed children 的路径避免 map 分配并复用 child slice。
- docs_browser 的 System API capability grid 改为挂载后异步 probe。
- docs_browser 的 System API 路径查询结果改为 `sync.Once` 缓存，避免长页面每帧重复执行 `os.Stat`、`os.Executable` 和路径解析。
- 保留 System API 交互按钮的异步执行模式，避免把文件对话框、消息框、通知等真实系统调用放进渲染路径。
