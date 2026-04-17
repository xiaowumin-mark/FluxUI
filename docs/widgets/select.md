<!-- fluxui-doc-meta
{
  "id": "select",
  "title": "Select 下拉选择",
  "category": "输入交互",
  "order": 270,
  "summary": "Select 用于枚举值选择。",
  "example": { "id": "select_basic" },
  "apis": [
    "Select[T comparable](value T, options []SelectOptionItem[T], opts ...SelectOption[T]) Widget",
    "SelectPlaceholder[T comparable](text string) SelectOption[T]",
    "SelectDisabled[T comparable](disabled bool) SelectOption[T]",
    "SelectSearchable[T comparable](searchable bool) SelectOption[T]",
    "SelectMaxHeight[T comparable](height float32) SelectOption[T]",
    "SelectOnChange[T comparable](fn func(ctx *Context, value T)) SelectOption[T]",
    "SelectOnOpenChange[T comparable](fn func(ctx *Context, opened bool)) SelectOption[T]"
  ]
}
-->

# Select 下拉选择

## 组件说明
Select 适用于中等数量的枚举值选择，支持占位文本、展开状态回调和面板高度控制。

## 使用方法
- 组件是受控模式，`value` 由状态驱动。
- 候选项建议直接定义成固定 `[]SelectOptionItem[T]`。

## 使用示例
```go
level := ui.State[string](ctx)
ui.Select(
    level.Value(),
    []ui.SelectOptionItem[string]{
        {Label: "低优先级", Value: "low"},
        {Label: "中优先级", Value: "medium"},
        {Label: "高优先级", Value: "high"},
    },
    ui.SelectPlaceholder[string]("请选择优先级"),
    ui.SelectOnChange[string](func(ctx *ui.Context, value string) {
        level.Set(value)
    }),
)
```
