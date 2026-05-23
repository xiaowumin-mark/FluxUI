<!-- fluxui-doc-meta
{
  "id": "radio_group",
  "title": "RadioGroup 单选组",
  "category": "输入交互",
  "order": 260,
  "summary": "RadioGroup 用于多选一场景。",
  "example": { "id": "radio_group_basic" },
  "apis": [
    "RadioGroup(value string, items []RadioItem, opts ...RadioGroupOption) Widget",
    "RadioGroupElement(value string, items []RadioItem, opts ...RadioGroupOption) Element",
    "RadioGroupDirection(axis Axis) RadioGroupOption",
    "RadioGroupDisabled(disabled bool) RadioGroupOption",
    "RadioGroupOnChange(fn func(ctx *Context, value string)) RadioGroupOption",
    "RadioGroupSize(size float32) RadioGroupOption",
    "RadioGroupColor(col color.NRGBA) RadioGroupOption",
    "NewRadioGroupRef() *RadioGroupRef",
    "RadioGroupAttachRef(ref *RadioGroupRef) RadioGroupOption",
    "(*RadioGroupRef).SetValue(value string)"
  ]
}
-->

# RadioGroup 单选组

## 组件说明
RadioGroup 用于“多个选项中只能选一个”的场景，例如排序模式、视图模式选择。

## 使用方法
- 当前值通过 `value` 传入。
- 所有候选项放在 `[]RadioItem` 中。
- 变化回调用 `RadioGroupOnChange`。
- 若需外部主动切换选项，使用 `RadioGroupAttachRef` + `SetValue`。

## React-style Element

- `RadioGroupElement` 已可在 `RunElement` root 下直接使用。
- `value` 仍由调用方状态驱动；选项点击和 `RadioGroupRef` 命令队列仍由底层 radio group widget 管理。

## 使用示例
```go
mode := ui.State[string](ctx)
ui.RadioGroup(
    mode.Value(),
    []ui.RadioItem{
        {Label: "布局", Value: "layout"},
        {Label: "输入", Value: "input"},
    },
    ui.RadioGroupOnChange(func(ctx *ui.Context, value string) {
        mode.Set(value)
    }),
)
```

### React-style Element

```go
func ModePicker(ctx *ui.Context) ui.Element {
    mode := ui.UseState(ctx, "layout")
    return ui.RadioGroupElement(
        mode.Value(),
        []ui.RadioItem{{Label: "布局", Value: "layout"}, {Label: "输入", Value: "input"}},
        ui.RadioGroupOnChange(func(ctx *ui.Context, value string) {
            mode.Set(value)
        }),
    )
}
```
