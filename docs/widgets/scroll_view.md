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
    "ScrollViewElement(child Element, opts ...ScrollOption) Element",
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

## Wheel 和命中规则

- `ScrollView` 是单主轴滚动容器。纵向容器只消费 Gio `Scroll.Y`，横向容器只消费 Gio `Scroll.X`。
- 普通纵向滚轮不会触发横向 `ScrollView`，因此代码块、表格、tabs、chips 等横向区域不会截停父级文档纵向滚动。
- `wheel.PreventDefault()` 只阻止本次 wheel 默认滚动；`ScrollRef`、滚动条拖动和 auto-to-end 等非 wheel 来源仍可更新滚动位置。
- 横向触控板滚动依赖后端提供明确的 `Scroll.X`。当前不会把 `Shift+Scroll.Y` 转换成横向滚动，也不会建模同轴触边后的父级剩余 delta。
- 滚动后下一帧会按最新 offset 重建子树位置和输入区域；不移动鼠标时，当前屏幕坐标点击应命中新的可见目标。

## Lifecycle / React-style Element

- `ScrollViewElement` 已可在 `RunElement` root 下直接使用。
- 滚动偏移、滚动条状态、auto-to-end key 和 `ScrollRef` 命令队列仍属于底层 scroll host state，不迁入 component HookSlot。
- `ScrollRef` 的命令会跨 frame 排队并在布局时消费；调用方仍负责在组件卸载后停止发送无意义命令。

## 使用示例

### React-style Element

```go
func LogPanel(ctx *ui.Context) ui.Element {
    return ui.FixedHeightElement(
        220,
        ui.ScrollViewElement(
            ui.ColumnElement(
                ui.TextElement("长内容 1"),
                ui.TextElement("长内容 2"),
            ),
            ui.ScrollVertical(true),
        ),
    )
}
```

### Legacy Widget
旧 `ui.ScrollView` / `Widget` 写法继续可用：

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
