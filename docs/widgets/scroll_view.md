<!-- fluxui-doc-meta
{
  "id": "scroll_view",
  "title": "ScrollView 滚动容器",
  "category": "布局系统",
  "order": 100,
  "summary": "ScrollView 为内容提供滚动能力。",
  "example": { "id": "scroll_view_basic" },
  "apis": [
    "ScrollView(child Widget, opts ...ScrollOption) Widget",
    "ScrollVertical(vertical bool) ScrollOption",
    "ScrollHorizontal(horizontal bool) ScrollOption",
    "ScrollBarVisible(visible bool) ScrollOption",
    "ScrollOnChange(fn func(ctx *Context, x, y float32)) ScrollOption",
    "NewScrollRef() *ScrollRef",
    "ScrollAttachRef(ref *ScrollRef) ScrollOption",
    "ScrollAutoToEnd(enabled bool) ScrollOption",
    "ScrollAutoToEndKey(key any) ScrollOption"
  ]
}
-->

# ScrollView 滚动容器

## 组件说明
ScrollView 用于承载超出可视区的内容。常见于文档区、表单区、详情页正文区。

## 使用方法
- 垂直滚动内容用 `ScrollVertical(true)`。
- 水平滚动内容用 `ScrollHorizontal(true)`。
- 滚动偏移回调可用来做吸顶或联动效果。
- 外部主动控制滚动位置时，使用 `ScrollAttachRef` 绑定 `ScrollRef`，命令会在后续 frame 中由 ScrollView 消费。
- 消息流、日志流等需要自动贴底的场景建议配合 `ScrollAutoToEndKey`，只在数据版本变化时触发滚到底部。

## Lifecycle / React-style 说明

- 当前仍以 legacy `Widget` + `ScrollRef` 作为兼容实现。
- `ScrollView` 的滚动偏移、滚动条状态、auto-to-end key 和 ref 命令队列都属于 host state，不应混入 component HookSlot。
- `ScrollRef` 的命令会跨 frame 排队并在布局时消费；后续若引入 Element wrapper，需要明确 ref 绑定、命令释放和 host unmount 的处理规则。
- `ScrollAutoToEnd` / `ScrollAutoToEndKey` 与内容增长、用户手动滚动之间存在生命周期语义，本批次不冻结 `ScrollViewElement` API 名称。
- 在 Batch 5 期间，文档保持 legacy-first，只记录滚动状态和 ref 边界，不新增 React-style snippet，也不修改 docs_browser 示例映射。

## 使用示例
```go
ref := ui.NewScrollRef()

ui.FixedHeight(
    220,
    ui.ScrollView(
        ui.Column(
            ui.Text("长内容 1"),
            ui.Text("长内容 2"),
        ),
        ui.ScrollVertical(true),
        ui.ScrollAttachRef(ref),
    ),
)

// 外部主动控制
ref.ScrollToBottom()
ref.ScrollToTop()
ref.ScrollToOffset(120)
ref.ScrollBy(2.0)
```
