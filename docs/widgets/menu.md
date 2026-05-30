<!-- fluxui-doc-meta
{
  "id": "menu",
  "title": "Menu / Dropdown 菜单",
  "category": "输入交互",
  "order": 275,
  "summary": "Menu 用于展示一组临时操作或选项，DropdownMenu 将触发器与菜单面板组合在一起。",
  "example": { "id": "menu_basic" },
  "apis": [
    "Menu(items []MenuItem, opts ...MenuOption) Widget",
    "MenuElement(items []MenuItem, opts ...MenuOption) Element",
    "DropdownMenu(open bool, trigger Widget, items []MenuItem, opts ...DropdownMenuOption) Widget",
    "DropdownMenuElement(open bool, trigger Element, items []MenuItem, opts ...DropdownMenuOption) Element",
    "type MenuItem struct { Key string; Label string; Leading Widget; Trailing Widget; Disabled bool; Selected bool }",
    "MenuSelectedKey(key string) MenuOption",
    "MenuOnSelect(fn func(ctx *Context, key string)) MenuOption",
    "DropdownMenuOnOpenChange(fn func(ctx *Context, open bool)) DropdownMenuOption"
  ]
}
-->

# Menu / Dropdown 菜单

Menu 使用 MD3 surface container、level 2 elevation、48dp menu item 和 bounded ripple。选中项使用 `SecondaryContainer`，禁用项使用统一 disabled content。

`Menu` 适合已经有外层弹层的场景；`DropdownMenu` 适合把触发器和菜单面板放在一起。

```go
open := ui.UseState(ctx, false)
selected := ui.UseState(ctx, "copy")

return ui.DropdownMenuElement(
    open.Value(),
    ui.ContainerDecorationElement(
        ui.Bg(ui.UseTheme(ctx).Colors.Surface).
            WithPad(ui.Symmetric(10, 16)).
            WithRad(ui.UseTheme(ctx).Shapes.ExtraSmall),
        ui.TextElement("Open menu"),
    ),
    []ui.MenuItem{
        {Key: "copy", Label: "Copy"},
        {Key: "share", Label: "Share"},
        {Key: "delete", Label: "Delete", Disabled: true},
    },
    ui.DropdownMenuSelectedKey(selected.Value()),
    ui.DropdownMenuOnOpenChange(func(ctx *ui.Context, next bool) { open.Set(next) }),
    ui.DropdownMenuOnSelect(func(ctx *ui.Context, key string) { selected.Set(key) }),
)
```
