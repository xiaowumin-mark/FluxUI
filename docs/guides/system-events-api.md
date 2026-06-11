<!-- fluxui-doc-meta
{
  "id": "system_events_api",
  "title": "System Events API 系统事件",
  "category": "使用指南",
  "order": 128,
  "summary": "System Events API 提供显示器、DPI、主题、设置、电源和会话变化的订阅入口。",
  "example": { "id": "system_api_basic" },
  "apis": [
    "system.SubscribeSystemEvents(ctx context.Context, kinds ...system.SystemEventKind) (*system.SystemEventSubscription, error)",
    "(*system.SystemEventSubscription).Events() <-chan system.SystemEvent",
    "(*system.SystemEventSubscription).Close() error",
    "system.SystemEventDisplayChanged",
    "system.SystemEventDPIChanged",
    "system.SystemEventThemeChanged",
    "system.SystemEventSettingsChanged",
    "system.SystemEventPowerChanged",
    "system.SystemEventSessionChanged"
  ]
}
-->

# System Events API 系统事件

System Events API 是 `system` 包里的系统事件订阅能力，用于监听显示器、DPI、系统主题、设置、电源和会话变化。当前 Windows driver 声明 `CapabilitySystemEvents` 并使用隐藏消息窗口接收系统消息；macOS / Linux 也声明该能力，并提供轮询型最小 driver。缺少平台命令、系统路径或桌面服务时，`Probe(CapabilitySystemEvents)` 或订阅会返回包装后的 `ErrUnavailable`。

## 基本用法

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
        // event.Kind 区分事件类型，event.Detail 提供平台相关细节。
    }
}()
```

不传 `kinds` 时订阅 driver 能报告的全部事件。`Close()` 会关闭订阅 channel；关闭后的订阅不应继续使用。

传入未知 `SystemEventKind` 会在公共层直接返回包装后的 `ErrUnsupported`，不会创建一个永远没有事件的空订阅。需要按平台能力做条件逻辑时，先使用 `system.Supports(system.CapabilitySystemEvents)` 或 `system.Probe(system.CapabilitySystemEvents)`。

## 事件类型

- `SystemEventDisplayChanged`: 显示器分辨率或显示配置变化。Windows `Detail` 形如 `1920x1080`。
- `SystemEventDPIChanged`: DPI 变化。Windows `Detail` 是 DPI 数值字符串；只有底层消息能到达隐藏窗口时才会报告。
- `SystemEventThemeChanged`: 系统主题或强调色相关变化。
- `SystemEventSettingsChanged`: 其他系统设置变化。
- `SystemEventPowerChanged`: 睡眠、恢复或电源设置变化。
- `SystemEventSessionChanged`: 登录会话锁定、解锁、连接或断开等变化，具体 detail 是 Windows 原始事件码。

Windows driver 使用独立隐藏消息窗口接收 `WM_DISPLAYCHANGE`、`WM_THEMECHANGED`、`WM_SETTINGCHANGE`、`WM_POWERBROADCAST`、`WM_WTSSESSION_CHANGE` 和可到达的 `WM_DPICHANGED`。需要精确到单个 FluxUI 窗口的 DPI/scale 时，优先订阅 `WindowEventScaleChanged` 并读取 `WindowState.Scale` / `DPI`。

macOS v1 通过 `ioreg`、`defaults`、`pmset` 和 `stat` 读取显示器、主题、设置、电源和 console session 快照。Linux v1 优先读取 `/sys/class/drm` 与 `/sys/class/power_supply`，并在可用时使用 `gsettings` 和 `loginctl` 读取主题、设置和会话状态。macOS / Linux driver 每隔约 2 秒比较一次快照；状态变化时发送对应 `SystemEvent`。它不是完整的原生事件流，不保证捕获短暂变化或所有桌面环境特有事件。

## 事件边界

事件 channel 有固定缓冲；如果调用方长期不消费，driver 可以丢弃后续事件，避免阻塞 Win32 消息线程或 Unix 轮询 goroutine。事件是系统级提示，不保证每一种平台状态变化都能一一对应到回调。窗口级 DPI/scale 精度由 `WindowEventScaleChanged` 补充。

## 示例

`examples/system_showcase` 提供 System Events 的人工验收入口。点击“监听系统事件 30 秒”会订阅 driver 能报告的全部系统事件，并在结果区显示最近事件；可通过切换显示设置、主题、电源状态或会话状态触发验证。
