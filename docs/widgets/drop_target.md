<!-- fluxui-doc-meta
{
  "id": "drop_target",
  "title": "DropTarget 拖拽接收区",
  "category": "输入交互",
  "order": 250,
  "summary": "DropTarget 在子组件区域接收拖放数据，并从 URI list、file URI 或文本 payload 中解析文件路径。",
  "example": { "id": "drop_target_basic" },
  "apis": [
    "DropTarget(child Widget, onDrop func(ctx *Context, event DropEvent), opts ...DropTargetOption) Widget",
    "DropTargetElement(child Element, onDrop func(ctx *Context, event DropEvent), opts ...DropTargetOption) Element",
    "DropTargetTypes(types ...string) DropTargetOption",
    "DropTargetMaxBytes(maxBytes int64) DropTargetOption",
    "DropTargetOperation(operation DragOperation) DropTargetOption",
    "DropTargetDisabled(disabled bool) DropTargetOption",
    "DropTargetOnActiveChange(fn func(ctx *Context, event DropTargetStateEvent)) DropTargetOption",
    "DropTargetOnError(fn func(ctx *Context, event DropEvent)) DropTargetOption",
    "DropEvent",
    "DropTargetStateEvent"
  ]
}
-->

# DropTarget 拖拽接收区

`DropTarget` 是无视觉样式的拖放接收区。它包住一个 child，并在 child 的布局区域内注册 Gio transfer target。

```go
ui.DropTargetElement(
    ui.ContainerDecorationElement(
        ui.Bg(ui.NRGBA(245, 247, 250, 255)).WithPad(ui.All(16)).WithRad(8),
        ui.TextElement("把文件拖到这里"),
    ),
    func(ctx *ui.Context, event ui.DropEvent) {
        if event.Err != nil {
            return
        }
        for _, path := range event.Paths {
            _ = path
        }
    },
)
```

## DropEvent

`DropEvent` 包含：

- `Type`: 后端提供的 MIME type。
- `Data`: 原始 payload bytes。
- `Text`: 按文本方式转换后的字符串。
- `Paths`: 从 `text/uri-list`、`file:` URI 或绝对路径文本中解析出的本地路径。
- `Operation`: 应用级 copy/move/link 语义。
- `Err`: payload 不可用、读取失败或超过大小限制时的错误。

默认接收常见文本和文件拖放类型：`text/uri-list`、`text/plain;charset=utf-8`、`text/plain;charset=utf8`、`text/plain`、`application/text`。

## 选项

`DropTargetTypes(types...)` 替换允许的 MIME type 列表：

```go
ui.DropTargetTypes("application/json", "text/plain")
```

`DropTargetMaxBytes(maxBytes)` 限制单次读取的 payload 大小，默认 32 MiB。超限时 `DropEvent.Err` 会带出错误。

`DropTargetOperation(operation)` 设置应用内语义。和 `DragSourceOperations` 一样，它不会控制 OS-level 拖放协商，只用于 FluxUI 事件、日志和业务逻辑。

`DropTargetDisabled(true)` 保留 child 布局，但不处理 transfer 事件。

## Active 与错误回调

`DropTargetOnActiveChange` 用于更新拖入高亮状态：

```go
ui.DropTargetElement(
    target,
    onDrop,
    ui.DropTargetOnActiveChange(func(ctx *ui.Context, event ui.DropTargetStateEvent) {
        if event.Active {
            _ = event.Types
        }
    }),
)
```

`DropTargetOnError` 会在读取 payload 失败时触发。`onDrop` 仍会收到同一个 `DropEvent`，这样调用方可以统一记录事件，也可以只在 `onError` 中处理失败提示。

## 文件路径解析

`DropTarget` 会从以下格式解析路径：

- `text/uri-list` 中的 `file:` URI。
- 普通文本中的 `file:` URI。
- 非 `text/uri-list` 文本中的绝对路径。

解析结果会去重。Windows 下去重大小写不敏感。解析出的路径不代表文件一定存在；需要读文件时仍应按业务要求调用 `os.Stat` 或直接打开并处理错误。

## 平台说明

拖放接收依赖 Gio `io/transfer` 和当前桌面后端。FluxUI 提供统一 API、payload 读取、URI/path 解析、大小限制、active/error 回调和 Element 包装；跨应用拖入是否提供文件 URI、纯文本或自定义 MIME，取决于操作系统、桌面环境和源应用。

需要在启动时判断能力时使用：

```go
probe := system.ProbeDragAndDrop(ctx)
if !probe.SupportsDropTarget {
    // 禁用接收区或显示降级 UI。
}
```
