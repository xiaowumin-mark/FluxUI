<!-- fluxui-doc-meta
{
  "id": "click_area",
  "title": "ClickArea 兼容点击区",
  "category": "输入交互",
  "order": 250,
  "summary": "ClickArea 是旧名称，继续兼容；新代码推荐使用 Pressable。",
  "example": { "id": "pressable_basic" },
  "apis": [
    "ClickArea(child Widget, onClick func(ctx *Context), opts ...ClickAreaOption) Widget",
    "ClickAreaElement(child Element, onClick func(ctx *Context), opts ...ClickAreaOption) Element",
    "NewClickAreaRef() *ClickAreaRef",
    "ClickAreaAttachRef(ref *ClickAreaRef) ClickAreaOption",
    "(*ClickAreaRef).Click()"
  ]
}
-->

# ClickArea 兼容点击区

## 组件说明

`ClickArea` 是旧版无视觉点击区域 API，当前保留为兼容别名。新代码应优先使用 `Pressable` / `PressableElement`，语义更接近 GUI 中“可按压区域”的概念。

`ClickArea` 不会附带按钮背景、ripple 或固定视觉反馈，适合背景点击关闭、透明热区、自定义整行点击等场景。主操作仍应使用 Button 或对应的 Material 组件。

## 迁移建议

- Widget API：把 `ClickArea(...)` 改为 `Pressable(...)`。
- Element API：把 `ClickAreaElement(...)` 改为 `PressableElement(...)`。
- Ref API：旧的 `ClickAreaRef` 继续可用；新代码可使用 `PressableRef`。

## 使用示例

### 推荐写法

```go
func HotSpot(ctx *ui.Context) ui.Element {
    count := ui.UseState(ctx, 0)
    return ui.PressableElement(
        ui.ContainerDecorationElement(
            ui.Bg(ui.NRGBA(227, 242, 253, 255)).WithPad(ui.All(14)).WithRad(8),
            ui.TextElement(fmt.Sprintf("点击次数: %d", count.Value())),
        ),
        func(ctx *ui.Context) {
            count.Set(count.Value() + 1)
        },
    )
}
```

### 兼容写法

```go
ui.ClickAreaElement(
    ui.TextElement("旧代码仍可运行"),
    func(ctx *ui.Context) {
        // handle click
    },
)
```
