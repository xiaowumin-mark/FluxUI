<!-- fluxui-doc-meta
{
  "id": "drag_source",
  "title": "DragSource 拖拽数据源",
  "category": "输入交互",
  "order": 251,
  "summary": "DragSource 让子组件区域可拖拽，并向 FluxUI DropTarget 提供文本、文件 URI list 或自定义 MIME payload；跨应用拖出只在后端明确支持时可用。",
  "example": { "id": "drag_source_basic" },
  "apis": [
    "DragSource(child Widget, opts ...DragSourceOption) Widget",
    "DragSourceElement(child Element, opts ...DragSourceOption) Element",
    "DragSourceText(text string) DragSourceOption",
    "DragSourceFiles(paths ...string) DragSourceOption",
    "DragSourceData(mime string, data []byte) DragSourceOption",
    "DragSourcePayloads(payloads ...DragPayload) DragSourceOption",
    "DragSourcePreview(preview Widget) DragSourceOption",
    "DragSourceOperations(operations ...DragOperation) DragSourceOption",
    "DragSourceDisabled(disabled bool) DragSourceOption",
    "DragSourceOnEvent(fn func(ctx *Context, event DragSourceEvent)) DragSourceOption",
    "DragSourceOnRequest(fn func(ctx *Context, event DragSourceEvent)) DragSourceOption",
    "DragPayload",
    "DragSourceEvent",
    "DragSourceEventStarted",
    "DragSourceEventRequested",
    "DragSourceEventCompleted",
    "DragSourceEventCancelled",
    "DragOperationCopy",
    "DragOperationMove",
    "DragOperationLink"
  ]
}
-->

# DragSource 拖拽数据源

`DragSource` 包住一个 child，让该区域成为拖拽数据源。目标请求某个 MIME type 时，FluxUI 通过 Gio `io/transfer` 提供对应 payload。

```go
ui.DragSourceElement(
    ui.ContainerDecorationElement(
        ui.Bg(ui.NRGBA(245, 247, 250, 255)).WithPad(ui.All(16)).WithRad(8),
        ui.TextElement("拖出 report.pdf"),
    ),
    ui.DragSourceFiles(`C:\Users\me\report.pdf`),
    ui.DragSourceOperations(ui.DragOperationCopy, ui.DragOperationLink),
)
```

## Payload

- `DragSourceText(text)` 提供 `text/plain;charset=utf-8`、`text/plain` 和 `application/text`。
- `DragSourceFiles(paths...)` 提供 `text/uri-list` 和 `text/plain`。URI list 使用 `file:` URI，适合拖到另一个 FluxUI `DropTarget`；跨应用拖出只在当前后端通过 `SupportsExternalDragOut` 明确报告支持时才应启用。
- `DragSourceData(mime, data)` 增加一个自定义 MIME payload。
- `DragSourcePayloads(payloads...)` 直接替换完整 payload 列表。

FluxUI 会对 MIME type 去空白、转小写并去重。`DragSourceFiles` 会尽量把路径转成绝对路径并去重，但不检查文件是否存在；如果业务要求文件必须存在，请在调用前自行验证。

## 生命周期

`DragSourceOnEvent` 接收更完整的生命周期事件：

- `DragSourceEventStarted`: Gio transfer 后端开始一次拖拽。
- `DragSourceEventRequested`: 目标请求了某个 MIME payload，`Type` 和 `Data` 会带出实际传输内容。
- `DragSourceEventCompleted`: 已向目标提交 payload。
- `DragSourceEventCancelled`: 拖拽被取消或结束但没有完成请求。

`DragSourceOnRequest` 是兼容旧 API 的快捷回调，只在目标请求 payload 时触发。新代码需要区分开始、完成、取消时，应优先使用 `DragSourceOnEvent`。

```go
ui.DragSourceElement(
    item,
    ui.DragSourceText("FluxUI payload"),
    ui.DragSourceOnEvent(func(ctx *ui.Context, event ui.DragSourceEvent) {
        switch event.Kind {
        case ui.DragSourceEventRequested:
            _ = event.Type
        case ui.DragSourceEventCompleted:
            // 更新应用内状态。
        }
    }),
)
```

## Operation 语义

`DragSourceOperations` 声明本次拖拽在应用内代表 copy、move 或 link：

```go
ui.DragSourceOperations(ui.DragOperationCopy, ui.DragOperationMove)
```

当前 Gio `io/transfer` 没有暴露 OS-level copy/move/link 协商字段，所以 FluxUI 不会伪装成能控制系统拖放操作。`Operation` 会写入 `DragSourceEvent`，用于 FluxUI 应用内部逻辑、日志和与 `DropTargetOperation` 的语义配合。跨应用拖到 Explorer、Finder 或文件管理器不是默认承诺；只有 `system.ProbeDragAndDrop(ctx).SupportsExternalDragOut` 为 `true` 时，业务才应把它作为可用能力展示。

## 预览与禁用

默认拖拽预览复用 child。需要自定义预览时使用 `DragSourcePreview`：

```go
ui.DragSourceElement(
    row,
    ui.DragSourceFiles(path),
    ui.DragSourcePreview(ui.Text("report.pdf")),
)
```

`DragSourceDisabled(true)` 会保留 child 布局，但不注册拖拽源，也不会处理 transfer 请求。常用于权限不足、数据未准备好或平台 probe 不可用时禁用入口。

## 平台说明

拖拽传输依赖 Gio `io/transfer` 和当前桌面后端。FluxUI 保证 API 在支持的平台可编译，并负责 payload 规范化、file URI 编码、Element 包装、状态回调和测试覆盖；真实跨应用拖入行为仍需要在 Windows、macOS、Linux 的具体桌面环境人工点验，跨应用拖出只按 `SupportsExternalDragOut` 单独报告。

系统能力探测请使用：

```go
probe := system.ProbeDragAndDrop(ctx)
if !probe.Available() {
    // 禁用拖拽入口或提示用户当前环境不可用。
}
```
