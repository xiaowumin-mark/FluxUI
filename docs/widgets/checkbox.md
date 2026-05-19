<!-- fluxui-doc-meta
{
  "id": "checkbox",
  "title": "Checkbox 复选框",
  "category": "输入交互",
  "order": 220,
  "summary": "Checkbox 用于布尔选择。",
  "example": { "id": "checkbox_basic" },
  "apis": [
    "Checkbox(label string, checked bool, opts ...CheckboxOption) Widget",
    "CheckboxElement(label string, checked bool, opts ...CheckboxOption) Element",
    "CheckboxOnChange(fn func(ctx *Context, checked bool)) CheckboxOption",
    "CheckboxDisabled(disabled bool) CheckboxOption",
    "CheckboxSize(size float32) CheckboxOption",
    "CheckboxColor(color color.NRGBA) CheckboxOption",
    "NewCheckboxRef() *CheckboxRef",
    "CheckboxAttachRef(ref *CheckboxRef) CheckboxOption",
    "(*CheckboxRef).SetChecked(checked bool)",
    "(*CheckboxRef).Toggle()"
  ]
}
-->

# Checkbox 复选框

## 组件说明
Checkbox 适用于“可多选”或“布尔开关”场景，常见于表单协议、筛选项、功能开关。

## 使用方法
- 通过 `checked` 传入当前值。
- 用 `CheckboxOnChange` 回传新值。
- 批量筛选建议用列表渲染多个 Checkbox。
- 需要外部命令式切换时，使用 `CheckboxAttachRef` + `SetChecked/Toggle`。

## 使用示例

### Legacy Widget
旧 `ui.Checkbox` / `Widget` 写法继续可用：

```go
agree := ui.State[bool](ctx)
ui.Checkbox(
    "同意服务协议",
    agree.Value(),
    ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) {
        agree.Set(checked)
    }),
)
```

### React-style Element
新代码可在 `RunElement` root 下返回 `CheckboxElement`：

```go
func AgreementRow(ctx *ui.Context) ui.Element {
    agree := ui.UseState(ctx, false)
    return ui.RowElement(
        ui.CheckboxElement(
            "同意服务协议",
            agree.Value(),
            ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) {
                agree.Set(checked)
            }),
        ),
        ui.SpacerElement(8, 0),
        ui.TextElement(fmt.Sprintf("协议状态: %v", agree.Value())),
    )
}
```
