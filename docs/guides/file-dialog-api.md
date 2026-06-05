<!-- fluxui-doc-meta
{
  "id": "file_dialog_api",
  "title": "File Dialog API 文件选择",
  "category": "使用指南",
  "order": 124,
  "summary": "File Dialog API 提供系统原生打开文件、打开多个文件、保存文件和选择目录能力。",
  "apis": [
    "system.OpenFileDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.OpenFilesDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.SaveFileDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.PickFolderDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.FileDialogTitle(value string) system.FileDialogOption",
    "system.FileDialogDefaultDir(value string) system.FileDialogOption",
    "system.FileDialogDefaultName(value string) system.FileDialogOption",
    "system.FileDialogFilters(filters ...system.FileFilter) system.FileDialogOption",
    "system.FileDialogOwner(owner uintptr) system.FileDialogOption",
    "ui.OpenFileDialog(ctx *ui.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "ui.OpenFileDialogContext(ctx *ui.Context, callCtx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)"
  ]
}
-->

# File Dialog API 文件选择

File Dialog API 是 `system` 包里的阻塞式系统能力，用于调用操作系统原生文件选择窗口。当前 Windows 已实现，非 Windows 平台保持可编译并返回 `ErrUnsupported`。

## 能力检测

调用前先检查能力：

```go
if !system.Supports(system.CapabilityFileDialog) {
    return
}
```

未实现平台也可以直接调用入口函数，返回的错误可通过 `system.IsUnsupported(err)` 判断。

## 打开单个文件

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
    return err
}
if result.Cancelled {
    return nil
}

path := result.Paths[0]
```

`Patterns` 支持 `png`、`.png`、`*.png` 和 `*.*` 这类写法。Windows driver 会把 `png` 和 `.png` 规范化为 `*.png`。

## 多选文件

```go
result, err := system.OpenFilesDialog(ctx,
    system.FileDialogTitle("选择导入文件"),
    system.FileDialogFilters(system.FileFilter{
        Name:     "Documents",
        Patterns: []string{"pdf", "docx", "txt"},
    }),
)
```

成功时 `result.Paths` 按系统返回顺序保存所有绝对路径。取消选择时 `result.Cancelled == true`，且 `err == nil`。

## 保存文件

```go
result, err := system.SaveFileDialog(ctx,
    system.FileDialogTitle("保存报告"),
    system.FileDialogDefaultName("report.txt"),
    system.FileDialogOverwritePrompt(true),
    system.FileDialogAllowCreateDirs(true),
)
```

保存对话框默认启用覆盖确认和目录创建提示。可以通过 `FileDialogOverwritePrompt(false)` 或 `FileDialogAllowCreateDirs(false)` 关闭。

## 选择目录

```go
result, err := system.PickFolderDialog(ctx,
    system.FileDialogTitle("选择输出目录"),
    system.FileDialogDefaultDir("C:\\Users\\me\\Desktop"),
)
```

选择目录时会忽略文件名和过滤器选项。

## Owner 和模态行为

`FileDialogOwner(owner)` 设置原生 owner 窗口句柄。Windows 下 `owner` 解释为 `HWND`；传 0 或不传该 option 时，对话框为无 owner。

普通 FluxUI UI 代码不需要手动处理 HWND。使用 `ui` 层 wrapper 时，FluxUI 会从当前 `*ui.Context` 自动取得 `WindowHandle.NativeHandle()` 并注入 owner：

```go
result, err := ui.OpenFileDialogContext(ctx, context.Background(),
    system.FileDialogTitle("选择文件"),
)
```

如果直接调用 `system.OpenFileDialog` / `system.OpenFilesDialog` / `system.SaveFileDialog` / `system.PickFolderDialog`，则仍需要显式传 `FileDialogOwner(owner)` 才能获得 owner modal 行为。

带 owner 的 Windows Common Item Dialog 会作为当前窗口的 modal 对话框显示；对话框关闭前，owner 窗口不能正常切回交互焦点。`examples/system_showcase` 已按这个方式传入 owner。

## 选项

- `FileDialogTitle(value)`: 设置系统对话框标题。
- `FileDialogDefaultDir(value)`: 设置初始目录。
- `FileDialogDefaultName(value)`: 设置初始文件名，适用于打开文件和保存文件。
- `FileDialogFilters(filters...)`: 设置文件类型过滤器。
- `FileDialogOwner(owner)`: 设置原生 owner 窗口句柄。Windows 下解释为 `HWND`，通常来自 `WindowHandle.NativeHandle()`；传 0 表示无 owner。
- `FileDialogAllowCreateDirs(allow)`: 控制保存时是否允许系统提示创建目录，默认 true。
- `FileDialogAllowMissingPath(allow)`: 控制是否允许选择不存在路径，默认 false。
- `FileDialogOverwritePrompt(prompt)`: 控制保存到已有文件时是否显示覆盖确认，默认 true。

## 错误和取消

- 用户取消：返回 `FileDialogResult{Cancelled: true}`，`err == nil`。
- 当前平台不支持：返回包装后的 `ErrUnsupported`。
- 系统服务、COM 或 shell item 不可用：返回普通错误或包装后的 `ErrUnavailable`。
- `context.Context` 已取消：打开前直接返回 context 错误。
- `FileDialogDefaultDir` 指向不存在或不可访问目录时，Windows shell item 创建会返回清晰错误；调用方应回退到可访问目录或不传默认目录。

## 示例

`examples/system_showcase` 提供 File Dialog 的人工验收入口，覆盖打开单个文件、打开多个文件、保存文件和选择目录。示例会根据 `system.Supports(system.CapabilityFileDialog)` 禁用按钮，并将当前 FluxUI 窗口的 native owner 传给系统对话框。

## 首版限制

- `ui.OpenFileDialog`、`ui.OpenFilesDialog`、`ui.SaveFileDialog` 和 `ui.PickFolderDialog` 会自动绑定当前窗口 owner。
- `system` 包没有 `*ui.Context`，因此不会自动推导当前窗口；直接调用 `system` 包且需要 modal owner 时，调用方应显式传入 `FileDialogOwner(owner)`。
- 未传 `FileDialogOwner` 或 owner 为 0 时使用无 owner 的 Common Item Dialog。
- 对话框显示后，当前版本不能通过 context 强制关闭原生窗口。
- 文件选择是阻塞调用，不要放在布局函数中。建议在点击回调或独立 goroutine 中调用，再把结果送回应用状态。

## 验收

Windows 本地验收时至少覆盖：

- 打开单个文件并返回绝对路径。
- 打开多个文件并保持选择顺序。
- 保存文件时默认文件名、覆盖确认和过滤器显示正确。
- 选择目录返回目录路径。
- 取消选择返回 `Cancelled=true` 且无错误。
- 不存在的默认目录返回清晰错误。
- 中文路径、空格路径和长路径能正确返回。
- 传入当前窗口 owner 后，对话框关闭前主窗口不能正常切回交互焦点。
