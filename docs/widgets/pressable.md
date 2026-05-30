<!-- fluxui-doc-meta
{
  "id": "pressable",
  "title": "Pressable 通用按压区",
  "category": "输入交互",
  "order": 249,
  "summary": "Pressable 提供无固定视觉样式的点击能力，适合自定义交互区域。",
  "example": { "id": "pressable_basic" },
  "apis": [
    "Pressable(child Widget, onClick func(ctx *Context), opts ...PressableOption) Widget",
    "PressableElement(child Element, onClick func(ctx *Context), opts ...PressableOption) Element",
    "NewPressableRef() *PressableRef",
    "PressableAttachRef(ref *PressableRef) PressableOption",
    "(*PressableRef).Click()"
  ]
}
-->

# Pressable 通用按压区

## 组件说明

`Pressable` 是无固定视觉样式的可点击区域。它只提供命中、点击回调和 ref 触发能力，不会自动绘制按钮背景或 Material ripple。

适合使用 `Pressable` 的场景：

- 自定义卡片、列表行、图片或复杂布局的整体点击。
- 背景点击关闭、透明热区等不需要默认视觉反馈的区域。
- 需要完全自定义 hover/pressed 样式的组件外壳。

## 使用方法

- 将可交互内容作为 child。
- 通过 `onClick` 处理点击。
- 需要外部主动触发时，绑定 `PressableAttachRef(ref)` 后调用 `ref.Click()`。

## 使用示例

```go
func PressableDemo(ctx *ui.Context) ui.Element {
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
