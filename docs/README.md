# FluxUI 文档结构

当前文档已重构为“每个控件一个 Markdown 文件”，位于 `docs/widgets` 目录。

每个控件文档必须包含 `fluxui-doc-meta` 元数据块，格式如下：

```md
<!-- fluxui-doc-meta
{
  "id": "button",
  "title": "Button 按钮",
  "category": "输入交互",
  "order": 200,
  "summary": "按钮用于触发点击动作。",
  "example": { "id": "button_basic" },
  "apis": [
    "Button(child Widget, opts ...ButtonOption) Widget",
    "OnClick(fn func(ctx *Context)) ButtonOption"
  ]
}
-->
```

字段说明：

- `id`: 控件唯一标识（用于菜单选择与路由）
- `title`: 控件标题
- `category`: 控件分类（用于左侧菜单分组）
- `order`: 同分类内排序
- `summary`: 控件摘要
- `example.id`: 示例渲染器 ID（由示例程序映射）
- `apis`: 需要重点展示的 API 列表

示例应用：`examples/docs_browser/main.go`

## 编码与终端显示说明

- 文档文件统一使用 UTF-8 编码保存。
- 在 Windows PowerShell 默认编码下，中文可能显示为“乱码样式”（例如 `鏄`、`鍙`），这通常是终端解码问题，不代表文件内容损坏。
- 建议使用支持 UTF-8 的终端或在读取时显式指定 UTF-8。

## Ref 能力约定

框架中的命令式 Ref 能力用于“外部主动调用组件行为”，例如：

- `ButtonRef.Click()`
- `InputRef.SetText/Append/Clear/Focus/Blur`
- `SliderRef.SetValue/StepBy`
- `DialogRef.Open/Close/Toggle`

文档中若组件支持 Ref，需要在 `apis` 列表中显式列出：

- `NewXxxRef() *XxxRef`
- `XxxAttachRef(ref *XxxRef) XxxOption`
- `(*XxxRef).Method(...)`
