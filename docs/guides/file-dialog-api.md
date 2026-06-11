<!-- fluxui-doc-meta
{
  "id": "file_dialog_api",
  "title": "File Dialog API 文件选择",
  "category": "使用指南",
  "order": 124,
  "summary": "File Dialog API 提供系统原生打开文件、打开多个文件、保存文件和选择目录能力。",
  "example": { "id": "system_api_basic" },
  "apis": [
    "system.OpenFileDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.OpenFilesDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.SaveFileDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.PickFolderDialog(ctx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "system.FileDialogTitle(value string) system.FileDialogOption",
    "system.FileDialogDefaultDir(value string) system.FileDialogOption",
    "system.FileDialogDefaultName(value string) system.FileDialogOption",
    "system.FileDialogDefaultExtension(value string) system.FileDialogOption",
    "system.FileDialogFilters(filters ...system.FileFilter) system.FileDialogOption",
    "system.FileDialogOwner(owner uintptr) system.FileDialogOption",
    "system.FileDialogRememberDir(key string) system.FileDialogOption",
    "system.IsFileDialogErrorKind(err error, kind system.FileDialogErrorKind) bool",
    "ui.OpenFileDialog(ctx *ui.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "ui.OpenFileDialogContext(ctx *ui.Context, callCtx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "ui.OpenFilesDialog(ctx *ui.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "ui.OpenFilesDialogContext(ctx *ui.Context, callCtx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "ui.SaveFileDialog(ctx *ui.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "ui.SaveFileDialogContext(ctx *ui.Context, callCtx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "ui.PickFolderDialog(ctx *ui.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)",
    "ui.PickFolderDialogContext(ctx *ui.Context, callCtx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error)"
  ]
}
-->

# File Dialog API 文件选择

File Dialog API 是 `system` 包里的阻塞式系统能力，用于调用操作系统原生文件选择窗口。当前 Windows 已实现 Common Item Dialog；macOS / Linux 已提供最小原生命令 driver：macOS 使用 `osascript choose file/folder/file name`，Linux 按可用性选择 `zenity` 或 `kdialog`。

## 能力检测

调用前先检查能力：

```go
if !system.Supports(system.CapabilityFileDialog) {
    return
}
```

未实现平台也可以直接调用入口函数，返回的错误可通过 `system.IsUnsupported(err)` 判断。macOS / Linux 缺少 `osascript`、`zenity` 或 `kdialog` 等平台命令时，能力仍声明为 supported，但 `system.Probe(system.CapabilityFileDialog)` 和实际调用会返回包装后的 `ErrUnavailable`。

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

`Patterns` 支持 `png`、`.png`、`*.png` 和 `*.*` 这类写法。Windows、macOS 和 Linux driver 都会把 `png` 和 `.png` 规范化为 `*.png`；macOS 命令 driver 会提取简单扩展名交给 AppleScript `of type`，Linux 命令 driver 会映射为 `zenity` / `kdialog` 的 filter 字符串。

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
    system.FileDialogDefaultExtension("txt"),
    system.FileDialogRememberDir("reports"),
    system.FileDialogOverwritePrompt(true),
    system.FileDialogAllowCreateDirs(true),
)
```

保存对话框默认启用覆盖确认和目录创建提示。可以通过 `FileDialogOverwritePrompt(false)` 或 `FileDialogAllowCreateDirs(false)` 关闭。`FileDialogDefaultExtension("txt")` 会把默认扩展名传给系统保存对话框；传入 `.txt` 或 `*.txt` 也会规范化为 `txt`。保存结果没有扩展名时，FluxUI 公共层会兜底追加默认扩展名；已有扩展名时保持用户输入不变。

## 选择目录

```go
result, err := system.PickFolderDialog(ctx,
    system.FileDialogTitle("选择输出目录"),
    system.FileDialogDefaultDir("C:\\Users\\me\\Desktop"),
)
```

选择目录时会忽略文件名和过滤器选项。

## Owner 和模态行为

`FileDialogOwner(owner)` 设置原生 owner 窗口句柄。Windows 下 `owner` 解释为 `HWND`；传 0 或不传该 option 时，对话框为无 owner。macOS / Linux 最小命令 driver 暂不支持 owner modal，传入非 0 owner 会返回包装后的 `ErrUnsupported`。

普通 FluxUI UI 代码不需要手动处理 HWND。使用 `ui` 层 wrapper 时，FluxUI 会从当前 `*ui.Context` 自动取得 `WindowHandle.NativeHandle()` 并注入 owner：

```go
result, err := ui.OpenFileDialogContext(ctx, context.Background(),
    system.FileDialogTitle("选择文件"),
)
```

同一套 owner 自动绑定也适用于多选、保存和目录选择：

```go
files, err := ui.OpenFilesDialogContext(ctx, context.Background(),
    system.FileDialogTitle("选择导入文件"),
)
save, err := ui.SaveFileDialogContext(ctx, context.Background(),
    system.FileDialogDefaultName("report.txt"),
)
folder, err := ui.PickFolderDialogContext(ctx, context.Background(),
    system.FileDialogTitle("选择输出目录"),
)
_, _, _ = files, save, folder
_ = err
```

如果直接调用 `system.OpenFileDialog` / `system.OpenFilesDialog` / `system.SaveFileDialog` / `system.PickFolderDialog`，则仍需要显式传 `FileDialogOwner(owner)` 才能获得 owner modal 行为。

带 owner 的 Windows Common Item Dialog 会作为当前窗口的 modal 对话框显示；对话框关闭前，owner 窗口不能正常切回交互焦点。`examples/system_showcase` 已按这个方式传入 owner，并提供显示后 context 自动取消入口。

## 选项

- `FileDialogTitle(value)`: 设置系统对话框标题。
- `FileDialogDefaultDir(value)`: 设置初始目录。
- `FileDialogDefaultName(value)`: 设置初始文件名，适用于打开文件和保存文件。
- `FileDialogDefaultExtension(value)`: 设置默认扩展名，主要用于保存文件。
- `FileDialogFilters(filters...)`: 设置文件类型过滤器。
- `FileDialogOwner(owner)`: 设置原生 owner 窗口句柄。Windows 下解释为 `HWND`，通常来自 `WindowHandle.NativeHandle()`；传 0 表示无 owner。macOS / Linux 最小命令 driver 暂不支持非 0 owner。
- `FileDialogAllowCreateDirs(allow)`: 控制保存时是否允许系统提示创建目录，默认 true。
- `FileDialogAllowMissingPath(allow)`: 控制是否允许选择不存在路径，默认 false。
- `FileDialogOverwritePrompt(prompt)`: 控制保存到已有文件时是否显示覆盖确认，默认 true。
- `FileDialogRememberDir(key)`: 记录同一个 key 上一次成功选择的目录；下一次没有显式 `FileDialogDefaultDir` 时会自动使用。

## 错误和取消

- 用户取消：返回 `FileDialogResult{Cancelled: true}`，`err == nil`。
- 当前平台不支持，或 macOS / Linux 最小 driver 收到非 0 owner 等未实现子能力：返回包装后的 `ErrUnsupported`。
- 系统服务、COM、shell item 或平台命令不可用：返回普通错误或包装后的 `ErrUnavailable`。
- `context.Context` 已取消：打开前直接返回 context 错误；Windows 下对话框显示后取消 context 会调用 `IFileDialog.Close` 尝试关闭原生窗口并返回 context 错误；macOS / Linux 命令 driver 使用 `exec.CommandContext`，显示后取消会终止对应命令进程，是否能立即关闭已显示 dialog 取决于平台命令。
- `FileDialogDefaultDir` 指向不存在或不可访问目录时，Windows shell item 创建和 macOS / Linux default-dir 预检都会返回带 `FileDialogErrorDefaultDir` 的 `FileDialogError`；调用方可用 `IsFileDialogErrorKind(err, FileDialogErrorDefaultDir)` 判断并回退到可访问目录或不传默认目录。
- driver 返回空路径列表、空路径或无法规范为绝对路径的结果时，公共层会返回带 `FileDialogErrorPath` 的 `FileDialogError`；调用方可将它视为平台 driver 返回了无效选择结果。

## 示例

`examples/system_showcase` 提供 File Dialog 的人工验收入口，覆盖打开单个文件、打开多个文件、保存文件、选择目录、记忆目录、保存默认扩展名和显示后 context 自动取消。示例会根据 `system.Supports(system.CapabilityFileDialog)` 禁用按钮；Windows 下会将当前 FluxUI 窗口的 native owner 传给系统对话框，macOS / Linux 最小 driver 暂不支持 owner。

## 首版限制

- `ui.OpenFileDialog`、`ui.OpenFilesDialog`、`ui.SaveFileDialog` 和 `ui.PickFolderDialog` 会自动绑定当前窗口 owner。
- `system` 包没有 `*ui.Context`，因此不会自动推导当前窗口；直接调用 `system` 包且需要 modal owner 时，调用方应显式传入 `FileDialogOwner(owner)`。
- 未传 `FileDialogOwner` 或 owner 为 0 时使用无 owner 的 Common Item Dialog。
- macOS / Linux 最小 driver 通过平台命令实现，不支持 owner modal，也不承诺 `zenity` / `kdialog` 的所有桌面环境细节；更完整的 owner/modal、portal 和 toolkit 行为后续应由 `NSOpenPanel` / `NSSavePanel`、xdg-desktop-portal 或 GTK/Qt driver 补齐。
- Windows 下对话框显示后会监听 context 取消并尝试关闭原生窗口；macOS / Linux 显示后取消通过终止平台命令进程实现，具体关闭行为取决于平台命令。
- Windows 取消 watcher 会在 Common Item Dialog 返回后停止并等待退出，避免在释放 COM dialog 后继续调用 `IFileDialog.Close`。
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

macOS / Linux 验收时至少覆盖：

- `Probe(CapabilityFileDialog)` 在命令存在时返回 available，缺少命令时返回 unavailable。
- 打开单文件、多文件、保存文件和选择目录可返回路径，取消返回 `Cancelled=true`。
- 默认目录不存在时返回 `FileDialogErrorDefaultDir`。
- 传入非 0 owner 时返回 `ErrUnsupported`，不会静默忽略 owner modal 语义。
