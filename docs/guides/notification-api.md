<!-- fluxui-doc-meta
{
  "id": "notification_api",
  "title": "Notification API 系统通知",
  "category": "使用指南",
  "order": 126,
  "summary": "Notification API 提供系统通知的公共入口、能力检测、错误语义和事件边界。",
  "apis": [
    "system.Notify(ctx context.Context, opts ...system.NotificationOption) error",
    "system.NotificationTitle(value string) system.NotificationOption",
    "system.NotificationBody(value string) system.NotificationOption",
    "system.NotificationKindStyle(kind system.NotificationKind) system.NotificationOption",
    "system.NotificationIcon(path string) system.NotificationOption",
    "system.NotificationGroup(group string) system.NotificationOption",
    "system.NotificationTimeout(timeout time.Duration) system.NotificationOption",
    "system.NotificationOnClick(fn func(system.NotificationEvent)) system.NotificationOption"
  ]
}
-->

# Notification API 系统通知

Notification API 是 `system` 包里的非阻塞系统消息能力，用于后台任务完成、错误提醒和长任务通知。当前 Windows 使用 `Shell_NotifyIconW` 提交托盘气泡通知，不需要 AppUserModelID、安装程序或开始菜单快捷方式。Toast Notification 暂不作为第一版承诺。

## 当前状态

Windows driver 当前声明 `CapabilityNotification`，`Notify` 会创建进程级隐藏消息窗口，并通过 `Shell_NotifyIconW` 临时添加托盘图标后提交系统气泡通知。非 Windows 平台保持可编译并返回 `ErrUnsupported`。

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
        // Windows shell、通知区或图标加载当前不可用。
        return nil
    }
    return err
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
- `NotificationTimeout(timeout)`: 设置期望展示时长；系统可能按平台策略忽略。
- `NotificationOnClick(fn)`: 注册点击回调承载点，只有 driver 能稳定收到系统事件时才会调用。

## 事件边界

Notification API 预留了三类事件：

- `NotificationEventClicked`: 用户点击通知主体。
- `NotificationEventDismissed`: 用户关闭或系统移除通知。
- `NotificationEventAction`: 用户点击通知动作按钮。

Windows 托盘气泡会在收到 `NIN_BALLOONUSERCLICK` 时触发 `NotificationOnClick`，并传入 `NotificationEventClicked`。系统不保证每一种关闭、超时或通知中心行为都会回调，调用方不能把 callback 当作必达回调。

## Windows 展示路径策略

Phase E v1 不直接实现 Toast Notification。原因是 Toast 涉及 AppUserModelID、开始菜单快捷方式、打包与非打包应用差异，以及激活回调生命周期。当前先采用托盘气泡路径；完整 Tray API 会在 Phase F 继续补齐生命周期、菜单和退出清理。

## 错误语义

- 当前平台不支持通知能力：返回包装后的 `ErrUnsupported`。
- 平台 driver 声明通知能力，但 Windows shell、通知区、隐藏窗口或图标加载当前不可用：返回包装后的 `ErrUnavailable`。
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

`examples/system_showcase` 提供 Notification 的人工验收入口。示例会根据 `system.Supports(system.CapabilityNotification)` 显示能力状态：非 Windows 平台禁用入口；Windows 点击后应显示系统托盘气泡通知。若系统通知区不可用，示例会展示 `ErrUnavailable` 对应的不可用状态。

## 验收

Phase E v1 自动化验收覆盖：

- option 默认值和 option 透传。
- driver 分发和能力缺失。
- context 已取消时不调用 driver。
- Windows notification data 结构、图标标志、超时映射和 click callback option。
- 非 Windows 平台返回包装后的 `ErrUnsupported`。

Windows 本地人工验收需要确认中文标题、正文显示，以及点击通知主体时可触发 `NotificationOnClick`。系统专注助手或通知设置可能会抑制气泡显示。
