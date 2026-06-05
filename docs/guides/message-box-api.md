<!-- fluxui-doc-meta
{
  "id": "message_box_api",
  "title": "MessageBox API 系统消息框",
  "category": "使用指南",
  "order": 125,
  "summary": "MessageBox API 提供系统原生信息、警告、错误和确认消息框。",
  "apis": [
    "system.ShowMessageBox(ctx context.Context, opts ...system.MessageBoxOption) (system.MessageBoxResult, error)",
    "system.MessageBoxTitle(value string) system.MessageBoxOption",
    "system.MessageBoxText(value string) system.MessageBoxOption",
    "system.MessageBoxStyle(kind system.MessageBoxKind) system.MessageBoxOption",
    "system.MessageBoxButtonSet(buttons system.MessageBoxButtons) system.MessageBoxOption",
    "system.MessageBoxDefaultButton(result system.MessageBoxResult) system.MessageBoxOption",
    "system.MessageBoxOwner(owner uintptr) system.MessageBoxOption",
    "ui.ShowMessageBox(ctx *ui.Context, opts ...system.MessageBoxOption) (system.MessageBoxResult, error)",
    "ui.ShowMessageBoxContext(ctx *ui.Context, callCtx context.Context, opts ...system.MessageBoxOption) (system.MessageBoxResult, error)"
  ]
}
-->

# MessageBox API 系统消息框

MessageBox API 是 `system` 包里的阻塞式系统能力，用于调用操作系统原生消息框。当前 Windows 优先使用 `TaskDialog`，以获得比传统 `MessageBoxW` 更现代的系统样式；如果目标进程没有可用的 common controls v6，则回退到 `MessageBoxW`。非 Windows 平台保持可编译并返回 `ErrUnsupported`。

## 能力检测

调用前先检查能力：

```go
if !system.Supports(system.CapabilityMessageBox) {
    return
}
```

未实现平台也可以直接调用入口函数，返回的错误可通过 `system.IsUnsupported(err)` 判断。

## 基本用法

```go
result, err := system.ShowMessageBox(ctx,
    system.MessageBoxTitle("保存更改"),
    system.MessageBoxText("关闭前是否保存当前文档？"),
    system.MessageBoxStyle(system.MessageBoxQuestion),
    system.MessageBoxButtonSet(system.MessageBoxYesNoCancel),
    system.MessageBoxDefaultButton(system.MessageBoxResultCancel),
)
if err != nil {
    return err
}

switch result {
case system.MessageBoxResultYes:
    // 保存并继续。
case system.MessageBoxResultNo:
    // 不保存并继续。
case system.MessageBoxResultCancel:
    // 取消操作。
}
```

## 样式

`MessageBoxStyle` 控制系统消息框的语义和图标：

- `MessageBoxInfo`
- `MessageBoxWarning`
- `MessageBoxError`
- `MessageBoxQuestion`

默认样式是 `MessageBoxInfo`。

## 按钮集

`MessageBoxButtonSet` 控制可见按钮：

- `MessageBoxOK`
- `MessageBoxOKCancel`
- `MessageBoxYesNo`
- `MessageBoxYesNoCancel`
- `MessageBoxRetryCancel`

默认按钮集是 `MessageBoxOK`。`MessageBoxDefaultButton` 必须指向当前按钮集里真实存在的按钮，否则返回错误。

## 返回值

`ShowMessageBox` 返回用户选择：

- `MessageBoxResultOK`
- `MessageBoxResultCancel`
- `MessageBoxResultYes`
- `MessageBoxResultNo`
- `MessageBoxResultRetry`
- `MessageBoxResultClose`

Windows v1 不能稳定区分点击 Cancel、按 Escape 和点击关闭按钮。只要系统返回 `IDCANCEL`，FluxUI 都返回 `MessageBoxResultCancel`。`MessageBoxResultClose` 保留给后续能区分关闭动作的平台 driver。

## Owner

`MessageBoxOwner(owner)` 设置原生 owner 窗口句柄。Windows 下解释为 `HWND`，传 0 表示无 owner。

普通 FluxUI UI 代码不需要手动处理 HWND。使用 `ui` 层 wrapper 时，FluxUI 会从当前 `*ui.Context` 自动取得 `WindowHandle.NativeHandle()` 并注入 owner：

```go
result, err := ui.ShowMessageBoxContext(ctx, context.Background(),
    system.MessageBoxTitle("保存更改"),
    system.MessageBoxText("关闭前是否保存当前文档？"),
    system.MessageBoxStyle(system.MessageBoxQuestion),
    system.MessageBoxButtonSet(system.MessageBoxYesNoCancel),
)
```

如果直接调用 `system.ShowMessageBox`，则仍需要显式传 `MessageBoxOwner(owner)` 才能获得 owner modal 行为。

带 owner 的 Windows `TaskDialog` / `MessageBoxW` 会作为当前窗口的 modal 消息框显示；消息框关闭前，owner 窗口不能正常切回交互焦点。未传 owner 或 owner 为 0 时使用无 owner 的系统消息框，仍可显示，但不承诺严格置于 FluxUI 窗口前方。`examples/system_showcase` 已按这个方式传入 owner。

## 错误和阻塞边界

- 当前平台不支持：返回包装后的 `ErrUnsupported`。
- 系统服务不可用、`TaskDialog` 或 `MessageBoxW` 调用失败：返回普通错误或包装后的 `ErrUnavailable`。
- `context.Context` 已取消：打开前直接返回 context 错误。
- 原生消息框显示后，当前版本不能通过 context 强制关闭窗口。

Windows 实现会优先调用 `TaskDialog`。如果应用没有启用 common controls v6，系统可能不提供 Task Dialog，此时 FluxUI 会回退到传统 `MessageBoxW`，仍保持原生系统弹窗和相同返回值语义。

MessageBox 是阻塞式系统 UI，不要放在布局函数中。建议在点击回调、菜单命令、快捷键处理或独立 goroutine 中调用，再把结果送回应用状态。

系统消息框不会 fallback 到 FluxUI 自绘 `Dialog`。如果需要 fallback，应在业务层显式处理。

## 示例

`examples/system_showcase` 提供 MessageBox 的人工验收入口，覆盖信息、警告、错误、确认和重试消息框。示例会根据 `system.Supports(system.CapabilityMessageBox)` 禁用按钮，并将当前 FluxUI 窗口的 native owner 传给系统消息框。

## 验收

Windows 本地验收时至少覆盖：

- 信息、警告、错误、确认消息框能显示中文标题和正文。
- OK、OKCancel、YesNo、YesNoCancel、RetryCancel 返回值正确。
- 默认按钮生效。
- Escape 和关闭按钮在带 Cancel 的按钮集上返回 `MessageBoxResultCancel`。
- 无 owner 时仍能显示。
- 显式 owner 可用时消息框位于 owner 窗口前方，并且关闭前主窗口不能正常切回交互焦点。
