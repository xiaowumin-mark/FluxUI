<!-- fluxui-doc-meta
{
  "id": "clipboard_shell_api",
  "title": "Clipboard / Shell API 剪贴板与系统打开",
  "category": "使用指南",
  "order": 131,
  "summary": "Clipboard / Shell API 提供系统剪贴板文本/文件列表/图片读写、默认程序打开、文件管理器定位和细分错误能力。",
  "example": { "id": "system_api_basic" },
  "apis": [
    "system.ReadClipboardText(ctx context.Context) (string, error)",
    "system.WriteClipboardText(ctx context.Context, text string) error",
    "system.ReadClipboardFiles(ctx context.Context) ([]string, error)",
    "system.WriteClipboardFiles(ctx context.Context, paths []string) error",
    "system.ReadClipboardImagePNG(ctx context.Context) ([]byte, error)",
    "system.WriteClipboardImagePNG(ctx context.Context, data []byte) error",
    "system.OpenURL(ctx context.Context, target string) error",
    "system.OpenPath(ctx context.Context, path string) error",
    "system.RevealPath(ctx context.Context, path string) error",
    "system.ErrInvalidTarget",
    "system.ErrTargetNotFound",
    "system.ErrNoDefaultHandler",
    "system.ErrAccessDenied",
    "system.CapabilityClipboard",
    "system.CapabilityShell"
  ]
}
-->

# Clipboard / Shell API 剪贴板与系统打开

Clipboard / Shell 属于 System API 的 Phase G 能力。它们不参与布局，也不会 fallback 到自绘 UI；不可用时会返回明确错误。

## 剪贴板文本

```go
text, err := system.ReadClipboardText(ctx)
if err != nil {
    return err
}

err = system.WriteClipboardText(ctx, "Copied from FluxUI")
```

Windows v1 使用 `Get-Clipboard` / `Set-Clipboard` 管理系统文本剪贴板：

- `ReadClipboardText` 读取当前文本剪贴板；剪贴板没有文本格式时返回空字符串。
- `WriteClipboardText` 写入 Unicode 文本，并替换当前剪贴板内容。
- 剪贴板被其他进程短暂占用时会重试一小段时间；仍不可用时返回包装后的 `ErrUnavailable`。

Windows v1 还支持 Explorer 文件拖放列表剪贴板：

```go
files, err := system.ReadClipboardFiles(ctx)
err = system.WriteClipboardFiles(ctx, []string{"C:\\Users\\me\\report.pdf"})
```

- `ReadClipboardFiles` 读取当前文件列表；剪贴板没有文件列表时返回空切片。
- `WriteClipboardFiles` 会确认每个路径存在并写入文件拖放列表，效果等同于在 Explorer 中复制文件路径列表。
- macOS / Linux 当前不承诺文件列表剪贴板，调用返回 `ErrUnsupported`。

Windows v1 还支持 PNG 图片剪贴板：

```go
pngData, err := system.ReadClipboardImagePNG(ctx)
err = system.WriteClipboardImagePNG(ctx, pngData)
```

- `ReadClipboardImagePNG` 将当前剪贴板图片转换为 PNG bytes；剪贴板没有图片时返回 `nil, nil`。
- `WriteClipboardImagePNG` 要求输入是有效 PNG bytes，并将其写入系统图片剪贴板。
- macOS / Linux 当前不承诺图片剪贴板，调用返回 `ErrUnsupported`。

macOS / Linux v1 提供文本剪贴板最小 driver：

- macOS 使用 `pbpaste` / `pbcopy`。
- Linux 按可用性依次选择 `wl-paste` / `wl-copy`、`xclip` 或 `xsel`。
- 缺少这些系统命令、桌面会话不可用或系统拒绝访问剪贴板时返回包装后的 `ErrUnavailable`。

调用前可以检查：

```go
if !system.Availability(system.CapabilityClipboard).Available() {
    return
}
```

## 系统打开

```go
_ = system.OpenURL(ctx, "https://example.com")
_ = system.OpenPath(ctx, "C:\\Users\\me\\report.pdf")
_ = system.RevealPath(ctx, "C:\\Users\\me\\report.pdf")
```

Windows v1 使用 `ShellExecuteW`：

- `OpenURL` 用系统默认 handler 打开带 scheme 的 URL；`file:` URL 会先还原成本地路径并确认目标存在。
- `OpenPath` 用默认程序打开文件或目录。
- `RevealPath` 用 Explorer 定位文件或目录。
- `OpenURL(file://...)` / `OpenPath` / `RevealPath` 会在提交给系统前确认目标路径存在；目标已删除时直接返回包装后的 `ErrUnavailable`，避免 Explorer 弹出“位置不可用”的系统错误框。

macOS / Linux v1 提供最小 shell driver：

- macOS 使用 `open` 打开 URL、文件或目录，`RevealPath` 使用 `open -R` 在 Finder 中定位。
- Linux 使用 `xdg-open` 打开 URL、文件或目录；`RevealPath` 会打开目标目录，文件路径会打开其父目录。
- 与 Windows 一样，macOS / Linux 会在 `OpenPath`、`RevealPath` 和 `OpenURL(file://...)` 提交给系统命令前确认本地目标存在；目标缺失时返回包装后的 `ErrUnavailable` + `ErrTargetNotFound`，权限拒绝时返回包装后的 `ErrUnavailable` + `ErrAccessDenied`。

Shell 调用只负责把请求提交给操作系统，不等待外部程序退出。目标不存在、默认程序缺失、系统策略禁止打开等情况会返回包装后的 `ErrUnavailable`。

为了让业务层可以做更细分的提示，Shell 相关错误还会尽量包装以下错误：

- `ErrInvalidTarget`: URL 或路径为空、缺少 URL scheme，或 `file:` URL 解析失败。
- `ErrTargetNotFound`: 本地目标不存在。
- `ErrNoDefaultHandler`: 操作系统没有为目标注册默认 handler。
- `ErrAccessDenied`: 系统策略或权限拒绝打开请求。

这些细分错误通常会与 `ErrUnavailable` 同时包装，旧代码继续用 `system.IsUnavailable(err)` 判断仍然有效；需要细分时使用 `IsInvalidTarget`、`IsTargetNotFound`、`IsNoDefaultHandler` 和 `IsAccessDenied`。

## 平台策略

Windows、macOS 和 Linux driver 当前都声明 `CapabilityClipboard` 和 `CapabilityShell`。Clipboard 文本在三类平台可用；文件列表和图片剪贴板当前只在 Windows driver 支持；本地路径和 `file:` URL 的缺失目标分类已经覆盖 Windows、macOS 和 Linux。Linux 文件管理器更精确定位、macOS/Linux 图片/文件列表剪贴板，以及默认 handler 启动后更细的桌面环境错误分类仍可继续扩展。
