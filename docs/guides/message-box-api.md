<!-- fluxui-doc-meta
{
  "id": "message_box_api",
  "title": "MessageBox API 系统消息框",
  "category": "使用指南",
  "order": 125,
  "summary": "MessageBox API 提供系统原生信息、警告、错误和确认消息框。",
  "example": { "id": "system_api_basic" },
  "apis": [
    "system.ShowMessageBox(ctx context.Context, opts ...system.MessageBoxOption) (system.MessageBoxResult, error)",
    "system.ShowMessageBoxDetailed(ctx context.Context, opts ...system.MessageBoxOption) (system.MessageBoxDetailedResult, error)",
    "system.ShowMessageBoxAsync(ctx context.Context, opts ...system.MessageBoxOption) <-chan system.MessageBoxResponse",
    "system.ShowMessageBoxDetailedAsync(ctx context.Context, opts ...system.MessageBoxOption) <-chan system.MessageBoxDetailedResponse",
    "system.MessageBoxTitle(value string) system.MessageBoxOption",
    "system.MessageBoxText(value string) system.MessageBoxOption",
    "system.MessageBoxDetails(value string) system.MessageBoxOption",
    "system.MessageBoxFooter(value string) system.MessageBoxOption",
    "system.MessageBoxVerification(label string, checked bool) system.MessageBoxOption",
    "system.MessageBoxStyle(kind system.MessageBoxKind) system.MessageBoxOption",
    "system.MessageBoxButtonSet(buttons system.MessageBoxButtons) system.MessageBoxOption",
    "system.MessageBoxCustomButtons(buttons ...system.MessageBoxButton) system.MessageBoxOption",
    "system.MessageBoxDefaultButton(result system.MessageBoxResult) system.MessageBoxOption",
    "system.MessageBoxDefaultButtonID(id string) system.MessageBoxOption",
    "system.MessageBoxCommandLinks(enabled bool) system.MessageBoxOption",
    "system.MessageBoxOwner(owner uintptr) system.MessageBoxOption",
    "ui.ShowMessageBox(ctx *ui.Context, opts ...system.MessageBoxOption) (system.MessageBoxResult, error)",
    "ui.ShowMessageBoxContext(ctx *ui.Context, callCtx context.Context, opts ...system.MessageBoxOption) (system.MessageBoxResult, error)",
    "ui.ShowMessageBoxDetailed(ctx *ui.Context, opts ...system.MessageBoxOption) (system.MessageBoxDetailedResult, error)",
    "ui.ShowMessageBoxDetailedContext(ctx *ui.Context, callCtx context.Context, opts ...system.MessageBoxOption) (system.MessageBoxDetailedResult, error)",
    "ui.ShowMessageBoxAsync(ctx *ui.Context, opts ...system.MessageBoxOption) <-chan system.MessageBoxResponse",
    "ui.ShowMessageBoxAsyncContext(ctx *ui.Context, callCtx context.Context, opts ...system.MessageBoxOption) <-chan system.MessageBoxResponse",
    "ui.ShowMessageBoxDetailedAsync(ctx *ui.Context, opts ...system.MessageBoxOption) <-chan system.MessageBoxDetailedResponse",
    "ui.ShowMessageBoxDetailedAsyncContext(ctx *ui.Context, callCtx context.Context, opts ...system.MessageBoxOption) <-chan system.MessageBoxDetailedResponse"
  ]
}
-->

# MessageBox API 系统消息框

MessageBox API 是 `system` 包里的阻塞式系统能力，用于调用操作系统原生消息框。当前 Windows 优先使用 `TaskDialogIndirect`，以获得比传统 `MessageBoxW` 更现代的系统样式、默认按钮、富内容和显示后 context 取消；不可用时依次回退到 `TaskDialog` 和 `MessageBoxW`。macOS / Linux 已提供最小原生命令 driver：macOS 使用 `osascript display dialog`，Linux 按可用性选择 `zenity` 或 `kdialog`。

## 能力检测

调用前先检查能力：

```go
if !system.Supports(system.CapabilityMessageBox) {
    return
}
```

未实现平台也可以直接调用入口函数，返回的错误可通过 `system.IsUnsupported(err)` 判断。macOS / Linux 缺少 `osascript`、`zenity` 或 `kdialog` 等平台命令时，能力仍声明为 supported，但 `system.Probe(system.CapabilityMessageBox)` 和实际调用会返回包装后的 `ErrUnavailable`。

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

## 富 TaskDialog

Windows 下 `TaskDialogIndirect` 支持详细信息、footer、复选框、自定义按钮和 command links。需要读取复选框状态或自定义按钮 ID 时，使用 `ShowMessageBoxDetailed`：

```go
result, err := system.ShowMessageBoxDetailed(ctx,
    system.MessageBoxTitle("关闭前保存"),
    system.MessageBoxText("当前文档还有未保存的修改。"),
    system.MessageBoxDetails("如果不保存，最近的编辑内容会丢失。"),
    system.MessageBoxFooter("FluxUI"),
    system.MessageBoxVerification("不再提示", false),
    system.MessageBoxCustomButtons(
        system.MessageBoxButton{ID: "save", Label: "保存\n保存修改后关闭", Result: system.MessageBoxResultYes},
        system.MessageBoxButton{ID: "discard", Label: "不保存", Result: system.MessageBoxResultNo},
        system.MessageBoxButton{ID: "cancel", Label: "取消", Result: system.MessageBoxResultCancel},
    ),
    system.MessageBoxDefaultButtonID("save"),
    system.MessageBoxCommandLinks(true),
)
if err != nil {
    return err
}

switch result.ButtonID {
case "save":
    // 保存并关闭。
case "discard":
    // 不保存直接关闭。
case "cancel":
    // 取消关闭。
}
_ = result.VerificationChecked
```

富能力选项包括：

- `MessageBoxDetails(value)`: 设置可展开详细信息。
- `MessageBoxFooter(value)`: 设置 footer 文本。
- `MessageBoxVerification(label, checked)`: 设置复选框及初始状态。
- `MessageBoxCustomButtons(buttons...)`: 用自定义按钮替代标准按钮集。
- `MessageBoxDefaultButtonID(id)`: 设置默认自定义按钮。
- `MessageBoxCommandLinks(enabled)`: 用 command links 展示自定义按钮。
- `MessageBoxCommandLinksNoIcon(enabled)`: command links 不显示图标。
- `MessageBoxExpandedDetailsByDefault(enabled)`: 默认展开详细信息。
- `MessageBoxExpandDetailsInFooterArea(enabled)`: 在 footer 区域展开详细信息。

这些能力依赖 `TaskDialogIndirect`。如果应用使用了富选项而当前系统不可用，FluxUI 会返回错误，不会退回到会丢失这些能力的 `MessageBoxW`。macOS / Linux 最小 driver 只支持标题、正文、样式和标准按钮集；`details`、footer、verification、自定义按钮、command links、默认自定义按钮和 owner window 都会返回包装后的 `ErrUnsupported`。

## 返回值

`ShowMessageBox` 返回用户选择：

- `MessageBoxResultOK`
- `MessageBoxResultCancel`
- `MessageBoxResultYes`
- `MessageBoxResultNo`
- `MessageBoxResultRetry`
- `MessageBoxResultCustom`
- `MessageBoxResultEscape`
- `MessageBoxResultClose`

Windows `TaskDialogIndirect` 路径会记录真实按钮点击和关闭来源：点击 Cancel 按钮返回 `MessageBoxResultCancel`；按 Escape 返回 `MessageBoxResultEscape`；右上角关闭按钮或 `WM_CLOSE` 路径返回 `MessageBoxResultClose`。传统 `TaskDialog` / `MessageBoxW` 回退路径没有可用的 dialog callback，仍会把 `IDCANCEL` 映射为 `MessageBoxResultCancel`。macOS / Linux 命令 driver 只能映射可见按钮结果，无法稳定区分 Escape 和窗口关闭，关闭类结果按命令语义映射为 `MessageBoxResultCancel` 或对应的二级按钮。

## Owner

`MessageBoxOwner(owner)` 设置原生 owner 窗口句柄。Windows 下解释为 `HWND`，传 0 表示无 owner。macOS / Linux 最小命令 driver 暂不支持 owner modal，传入非 0 owner 会返回 `ErrUnsupported`。

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

带 owner 的 Windows `TaskDialog` / `MessageBoxW` 会作为当前窗口的 modal 消息框显示；消息框关闭前，owner 窗口不能正常切回交互焦点。未传 owner 或 owner 为 0 时使用无 owner 的系统消息框，仍可显示，但不承诺严格置于 FluxUI 窗口前方。`examples/system_showcase` 已按这个方式传入 owner，并提供富消息框和显示后 context 自动取消入口。

## 错误和阻塞边界

- 当前平台不支持，或 macOS / Linux 最小 driver 收到富选项、owner window 等未实现子能力：返回包装后的 `ErrUnsupported`。
- 系统服务不可用、`TaskDialog` / `MessageBoxW` 调用失败，或 macOS / Linux 缺少 `osascript`、`zenity`、`kdialog` 命令：返回普通错误或包装后的 `ErrUnavailable`。
- `context.Context` 已取消：打开前直接返回 context 错误。
- Windows `TaskDialogIndirect` 路径在消息框显示后也会监听 context 取消，取消时向原生 dialog 投递关闭请求并优先返回 context 错误。
- `TaskDialog` / `MessageBoxW` 兼容回退路径会在调用期间锁定当前 OS 线程，context 取消后通过枚举该线程上的原生 dialog 做 best-effort 关闭；如果系统没有暴露可匹配窗口，则仍可能等待用户手动关闭。
- macOS / Linux 命令 driver 使用 `exec.CommandContext` 调用平台命令；显示后的 context 取消依赖操作系统终止对应命令进程，是否能立即关闭已显示 dialog 取决于平台命令。

Windows 实现会优先调用 `TaskDialogIndirect`。如果应用没有启用 common controls v6，系统可能不提供 Task Dialog，此时 FluxUI 会回退到 `TaskDialog` 或传统 `MessageBoxW`，仍保持原生系统弹窗和相同返回值语义。

MessageBox 是阻塞式系统 UI，不要放在布局函数中。建议在点击回调、菜单命令、快捷键处理或独立 goroutine 中调用，再把结果送回应用状态。

如果调用方只需要避免手写 goroutine，可以使用异步入口：

```go
responses := ui.ShowMessageBoxAsyncContext(ctx, context.Background(),
    system.MessageBoxTitle("保存更改"),
    system.MessageBoxText("关闭前是否保存当前文档？"),
    system.MessageBoxStyle(system.MessageBoxQuestion),
    system.MessageBoxButtonSet(system.MessageBoxYesNoCancel),
)

go func() {
    response := <-responses
    if response.Err != nil {
        return
    }
    // 使用 response.Result 更新应用状态。
}()
```

富结果也有对应异步入口：`ShowMessageBoxDetailedAsync` 和 `ui.ShowMessageBoxDetailedAsyncContext`。

系统消息框不会 fallback 到 FluxUI 自绘 `Dialog`。如果需要 fallback，应在业务层显式处理。

## 示例

`examples/system_showcase` 提供 MessageBox 的人工验收入口，覆盖信息、警告、错误、确认、重试、富 TaskDialog 和显示后 context 自动取消。示例会根据 `system.Supports(system.CapabilityMessageBox)` 禁用按钮；Windows 下会将当前 FluxUI 窗口的 native owner 传给系统消息框，macOS / Linux 最小 driver 暂不支持 owner。

## 验收

Windows 本地验收时至少覆盖：

- 信息、警告、错误、确认消息框能显示中文标题和正文。
- OK、OKCancel、YesNo、YesNoCancel、RetryCancel 返回值正确。
- 默认按钮生效。
- `TaskDialogIndirect` 路径下 Cancel、Escape 和右上角关闭按钮分别返回 `MessageBoxResultCancel`、`MessageBoxResultEscape` 和 `MessageBoxResultClose`；`TaskDialog` / `MessageBoxW` 回退路径仍把 `IDCANCEL` 映射为 `MessageBoxResultCancel`。
- 无 owner 时仍能显示。
- 显式 owner 可用时消息框位于 owner 窗口前方，并且关闭前主窗口不能正常切回交互焦点。

macOS / Linux 验收时至少覆盖：

- `Probe(CapabilityMessageBox)` 在命令存在时返回 available，缺少命令时返回 unavailable。
- OK、OKCancel、YesNo、YesNoCancel、RetryCancel 返回值按平台命令正确映射。
- 使用富 TaskDialog 选项或非 0 owner 时返回 `ErrUnsupported`，不会静默降级。
