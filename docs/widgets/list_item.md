<!-- fluxui-doc-meta
{
  "id": "list_item",
  "title": "ListItem 列表项",
  "category": "基础显示",
  "order": 155,
  "summary": "ListItem 提供 MD3 单行/双行列表项，支持 leading、trailing、selected、disabled 和 ripple 状态。",
  "example": { "id": "list_item_basic" },
  "apis": [
    "ListItem(headline string, opts ...ListItemOption) Widget",
    "ListItemWithSlots(headline, supporting, leading, trailing Widget, opts ...ListItemOption) Widget",
    "ListItemElement(headline string, opts ...ListItemOption) Element",
    "ListItemElementWithSlots(headline, supporting, leading, trailing Element, opts ...ListItemOption) Element",
    "ListItemSelected(selected bool) ListItemOption",
    "ListItemDisabled(disabled bool) ListItemOption",
    "ListItemOnClick(fn func(ctx *Context)) ListItemOption",
    "ListItemDecoration(d Decoration) ListItemOption"
  ]
}
-->

# ListItem 列表项

ListItem 是用于列表、菜单式设置页和导航内容的基础行组件。默认一行高度为 56dp，带 supporting text 时为 72dp。

```go
ui.ListItemElementWithSlots(
    ui.TextElement("Inbox"),
    ui.TextElement("12 unread messages"),
    ui.IconElement("I"),
    ui.TextElement("12"),
    ui.ListItemSelected(true),
)
```
