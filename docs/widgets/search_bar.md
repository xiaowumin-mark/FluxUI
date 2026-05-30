<!-- fluxui-doc-meta
{
  "id": "search_bar",
  "title": "SearchBar 搜索栏",
  "category": "输入交互",
  "order": 278,
  "summary": "SearchBar 提供 Material 3 搜索输入栏，支持前后插槽。",
  "example": { "id": "search_bar_basic" },
  "apis": [
    "SearchBar(value string, opts ...SearchBarOption) Widget",
    "SearchBarElement(value string, opts ...SearchBarOption) Element",
    "SearchBarPlaceholder(text string) SearchBarOption",
    "SearchBarDisabled(disabled bool) SearchBarOption",
    "SearchBarOnChange(fn func(ctx *Context, value string)) SearchBarOption",
    "SearchBarLeading(leading Widget) SearchBarOption",
    "SearchBarTrailing(trailing Widget) SearchBarOption",
    "SearchBarDecoration(d Decoration) SearchBarOption",
    "SearchBarInputOptions(opts ...InputOption) SearchBarOption"
  ]
}
-->

# SearchBar 搜索栏

SearchBar 是 Material 3 风格的搜索输入控件，带有圆角容器、默认搜索图标和受控输入值。它适合文档筛选、列表搜索和页面内搜索入口。

## 使用方法

- `value` 由外部状态驱动。
- `SearchBarOnChange` 在输入变化时回调。
- `SearchBarLeading` 和 `SearchBarTrailing` 可设置前后插槽。
- `SearchBarInputOptions` 可透传底层输入框选项。

```go
ui.SearchBarElement(
    query.Value(),
    ui.SearchBarPlaceholder("搜索"),
    ui.SearchBarLeading(ui.Icon("S")),
    ui.SearchBarOnChange(func(ctx *ui.Context, value string) {
        query.Set(value)
    }),
)
```
