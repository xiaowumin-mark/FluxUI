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

## Host-state / React-style 说明

- 当前仍以 legacy `Widget` + `RadioGroupRef` 作为兼容实现。
- `RadioGroup` 的 Element wrapper 名称与 host-state 边界尚未冻结，不在本批次文档中作为稳定公开 API 推荐。
- 在 Batch 4 期间，文档保持 legacy-first，只保留受控 value、items、direction 和 ref 操作说明，不引入新的 React-style snippet。
- 如果后续引入 `RadioGroupElement`，需要先确认选项 identity、value 更新和 ref 命令如何归属到 host fiber。

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
