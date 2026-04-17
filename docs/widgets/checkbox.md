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
    "CheckboxOnChange(fn func(ctx *Context, checked bool)) CheckboxOption",
    "CheckboxDisabled(disabled bool) CheckboxOption",
    "CheckboxSize(size float32) CheckboxOption",
    "CheckboxColor(color color.NRGBA) CheckboxOption"
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

## 使用示例
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
