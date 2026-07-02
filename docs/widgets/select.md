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
    "OutlinedSelect[T comparable](value T, options []SelectOptionItem[T], opts ...SelectOption[T]) Widget",
    "FilledSelect[T comparable](value T, options []SelectOptionItem[T], opts ...SelectOption[T]) Widget",
    "SelectElement[T comparable](value T, options []SelectOptionItem[T], opts ...SelectOption[T]) Element",
    "OutlinedSelectElement[T comparable](value T, options []SelectOptionItem[T], opts ...SelectOption[T]) Element",
    "FilledSelectElement[T comparable](value T, options []SelectOptionItem[T], opts ...SelectOption[T]) Element",
    "SelectPlaceholder[T comparable](text string) SelectOption[T]",
    "SelectLabel[T comparable](text string) SelectOption[T]",
    "SelectSupportingText[T comparable](text string) SelectOption[T]",
    "SelectErrorText[T comparable](text string) SelectOption[T]",
    "SelectError[T comparable](error bool) SelectOption[T]",
    "SelectRequired[T comparable](required bool) SelectOption[T]",
    "SelectNoAsterisk[T comparable](noAsterisk bool) SelectOption[T]",
    "SelectDisabled[T comparable](disabled bool) SelectOption[T]",
    "SelectSearchable[T comparable](searchable bool) SelectOption[T]",
    "SelectMaxHeight[T comparable](height float32) SelectOption[T]",
    "SelectWidth[T comparable](width float32) SelectOption[T]",
    "SelectXOffset[T comparable](offset float32) SelectOption[T]",
    "SelectYOffset[T comparable](offset float32) SelectOption[T]",
    "SelectQuick[T comparable](quick bool) SelectOption[T]",
    "SelectTypeaheadDelay[T comparable](delay time.Duration) SelectOption[T]",
    "SelectFilled[T comparable](filled bool) SelectOption[T]",
    "SelectLeading[T comparable](leading Widget) SelectOption[T]",
    "SelectTrailing[T comparable](trailing Widget) SelectOption[T]",
    "SelectDecoration[T comparable](d Decoration) SelectOption[T]",
    "SelectMenuDecoration[T comparable](d Decoration) SelectOption[T]",
    "SelectOnChange[T comparable](fn func(ctx *Context, value T)) SelectOption[T]",
    "SelectOnOpenChange[T comparable](fn func(ctx *Context, opened bool)) SelectOption[T]",
    "NewSelectRef[T comparable]() *SelectRef[T]",
    "SelectAttachRef[T comparable](ref *SelectRef[T]) SelectOption[T]",
    "(*SelectRef[T]).SetValue(value T)",
    "(*SelectRef[T]).Open()",
    "(*SelectRef[T]).Close()",
    "(*SelectRef[T]).Toggle()"
  ]
}
-->

# Select 下拉选择

## 组件说明
Select 适用于枚举值选择，默认使用 Material 3 outlined select。FluxUI 也提供 `FilledSelect` / `FilledSelectElement` 对齐 Material Web 的 filled select 形态。

## 使用方法
- 组件是受控模式，`value` 由状态驱动。
- 候选项建议直接定义成固定 `[]SelectOptionItem[T]`。
- `SelectOptionItem` 支持 `Disabled`、`Leading`、`Trailing` 和 `TypeaheadText` 字段。
- `SelectLabel`、`SelectSupportingText`、`SelectError`、`SelectErrorText`、`SelectRequired` 对齐官方表单字段 API。
- 展开动画使用高度 reveal：内容先按完整尺寸布局，再通过 mask 展开，避免展开过程中重新布局子项。
- 需要外部打开/关闭或切换值时，使用 `SelectAttachRef` 绑定 `SelectRef[T]`。

## React-style Element

- `SelectElement[T]` 已可在 `RunElement` root 下直接使用。
- `value` 仍由调用方状态驱动；展开状态、候选项命中和 `SelectRef[T]` 命令队列仍由底层 select host state 管理。

## 使用示例

### React-style Element

```go
func PrioritySelect(ctx *ui.Context) ui.Element {
    level := ui.UseState(ctx, "medium")
    return ui.SelectElement(
        level.Value(),
        []ui.SelectOptionItem[string]{
            {Label: "Apple", Value: "apple"},
            {Label: "Apricot", Value: "apricot"},
            {Label: "Tomato", Value: "tomato", Disabled: true},
        },
        ui.SelectLabel[string]("Fruit"),
        ui.SelectSupportingText[string]("Pick one fruit"),
        ui.SelectOnChange[string](func(ctx *ui.Context, value string) {
            level.Set(value)
        }),
    )
}
```

### Legacy Widget
旧 `ui.Select` / `Widget` 写法继续可用：

```go
level := ui.State[string](ctx)
ui.Select(
    level.Value(),
    []ui.SelectOptionItem[string]{
        {Label: "Apple", Value: "apple"},
        {Label: "Apricot", Value: "apricot"},
    },
    ui.SelectPlaceholder[string]("Choose fruit"),
    ui.SelectLeading[string](ui.Icon("restaurant")),
    ui.SelectOnChange[string](func(ctx *ui.Context, value string) {
        level.Set(value)
    }),
)
```
