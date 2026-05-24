# FluxUI 文档结构

当前文档已重构为“每个主题一个 Markdown 文件”，文档浏览器会扫描以下目录：

- `docs/widgets`: 组件文档
- `docs/style`: 样式系统 API 文档
- `docs/theme`: 主题系统 API 文档
- `docs/guides`: 使用指南与动画文档

## 当前状态入口

- React-style runtime 和 docs/examples rollout 当前状态：`docs/react-style-status.md`
- examples 保留 / 合并 / 清理清单：`docs/examples-inventory.md`
- legacy API 定位：`docs/legacy-api-positioning.md`
- `FromWidget` escape hatch 策略：`docs/escape-hatch-strategy.md`
- deprecation 与版本节奏：`docs/deprecation-and-versioning.md`

历史计划文档仍保留用于追溯，但不再作为当前执行入口：

- `docs/react-style-refactor-plan.md`
- `docs/docs-example-migration-plan.md`
- `docs/phase6-root-element-plan.md`
- `docs/phase7-legacy-api-plan.md`

每篇可被文档浏览器加载的 Markdown 文档都必须包含 `fluxui-doc-meta` 元数据块，格式如下：

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

- `id`: 文档唯一标识（用于菜单选择与路由）
- `title`: 文档标题
- `category`: 文档分类（用于左侧菜单分组）
- `order`: 同分类内排序
- `summary`: 控件摘要
- `example.id`: 示例渲染器 ID（由示例程序映射）
- `apis`: 需要重点展示的 API 列表

示例应用：`examples/docs_browser/main.go`。本地加载失败时，示例应用会从 GitHub `docs/widgets`、`docs/style`、`docs/theme`、`docs/guides` 拉取在线 Markdown 文档。

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

## Router 文档补充

- 路由能力文档：`docs/widgets/router.md`
- 路由独立示例：`examples/router/main.go`
- 文档浏览器示例ID：`router_basic`
