<!-- fluxui-doc-meta
{
  "id": "click_area",
  "title": "ClickArea 无视觉点击区",
  "category": "输入交互",
  "order": 250,
  "summary": "ClickArea 提供点击能力但不附带默认按钮视觉反馈。",
  "example": { "id": "click_area_basic" },
  "apis": [
    "ClickArea(child Widget, onClick func(ctx *Context), opts ...ClickAreaOption) Widget",
    "NewClickAreaRef() *ClickAreaRef",
    "ClickAreaAttachRef(ref *ClickAreaRef) ClickAreaOption",
    "(*ClickAreaRef).Click()"
  ]
}
-->

# ClickArea 无视觉点击区

## 组件说明
ClickArea 用于“可点击但不希望出现按钮动效”的区域，如背景点击关闭、透明热区等。

## 使用方法
- 将可交互区域内容作为 child。
- 点击行为通过 `onClick` 回调处理。
- 不建议用于主操作按钮（主操作仍应使用 Button 保证反馈一致性）。
- 若要在外部主动触发，可通过 `ClickAreaAttachRef` 绑定后调用 `ref.Click()`。

## 使用示例
```go
ui.ClickArea(
    ui.FillWidth(
        ui.Container(
            ui.Style{Background: ui.NRGBA(227, 242, 253, 255), Padding: ui.All(14)},
            ui.Text("点击这里触发回调"),
        ),
    ),
    func(ctx *ui.Context) {
        // handle click
    },
)
```
