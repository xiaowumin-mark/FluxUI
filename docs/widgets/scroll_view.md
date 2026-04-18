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
