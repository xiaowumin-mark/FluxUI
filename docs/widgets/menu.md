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
    "type MenuItem struct { Key string; Label string; Leading Widget; Trailing Widget; Disabled bool; Selected bool; Divider bool; Children []MenuItem; Type string; Href string; Target string; KeepOpen bool; TypeaheadText string }",
    "type MenuCorner int",
    "type MenuDefaultFocus int",
    "type MenuPositioning int",
    "MenuSelectedKey(key string) MenuOption",
    "MenuOnSelect(fn func(ctx *Context, key string)) MenuOption",
    "MenuWidth(width float32) MenuOption",
    "MenuMaxHeight(height float32) MenuOption",
    "MenuDecoration(d Decoration) MenuOption",
    "MenuQuick(quick bool) MenuOption",
    "MenuHasOverflow(hasOverflow bool) MenuOption",
    "MenuXOffset(offset float32) MenuOption",
    "MenuYOffset(offset float32) MenuOption",
    "MenuAnchorCorner(corner MenuCorner) MenuOption",
    "MenuMenuCorner(corner MenuCorner) MenuOption",
    "MenuDefaultFocusOf(focus MenuDefaultFocus) MenuOption",
    "MenuPositioningOf(positioning MenuPositioning) MenuOption",
    "MenuTypeaheadDelay(delay time.Duration) MenuOption",
    "MenuNoHorizontalFlip(disabled bool) MenuOption",
    "MenuNoVerticalFlip(disabled bool) MenuOption",
    "MenuStayOpenOnOutsideClick(stayOpen bool) MenuOption",
    "MenuStayOpenOnFocusout(stayOpen bool) MenuOption",
    "MenuSkipRestoreFocus(skip bool) MenuOption",
    "MenuNoNavigationWrap(noWrap bool) MenuOption",
    "MenuHoverOpenDelay(delay time.Duration) MenuOption",
    "MenuHoverCloseDelay(delay time.Duration) MenuOption",
    "DropdownMenuSelectedKey(key string) DropdownMenuOption",
    "DropdownMenuOnSelect(fn func(ctx *Context, key string)) DropdownMenuOption",
    "DropdownMenuOnOpenChange(fn func(ctx *Context, open bool)) DropdownMenuOption",
    "DropdownMenuWidth(width float32) DropdownMenuOption",
    "DropdownMenuMaxHeight(height float32) DropdownMenuOption",
    "DropdownMenuDecoration(d Decoration) DropdownMenuOption",
    "DropdownMenuQuick(quick bool) DropdownMenuOption",
    "DropdownMenuHasOverflow(hasOverflow bool) DropdownMenuOption",
    "DropdownMenuXOffset(offset float32) DropdownMenuOption",
    "DropdownMenuYOffset(offset float32) DropdownMenuOption",
    "DropdownMenuAnchorCorner(corner MenuCorner) DropdownMenuOption",
    "DropdownMenuMenuCorner(corner MenuCorner) DropdownMenuOption",
    "DropdownMenuDefaultFocusOf(focus MenuDefaultFocus) DropdownMenuOption",
    "DropdownMenuPositioningOf(positioning MenuPositioning) DropdownMenuOption",
    "DropdownMenuTypeaheadDelay(delay time.Duration) DropdownMenuOption",
    "DropdownMenuNoHorizontalFlip(disabled bool) DropdownMenuOption",
    "DropdownMenuNoVerticalFlip(disabled bool) DropdownMenuOption",
    "DropdownMenuStayOpenOnOutsideClick(stayOpen bool) DropdownMenuOption",
    "DropdownMenuStayOpenOnFocusout(stayOpen bool) DropdownMenuOption",
    "DropdownMenuSkipRestoreFocus(skip bool) DropdownMenuOption",
    "DropdownMenuNoNavigationWrap(noWrap bool) DropdownMenuOption",
    "DropdownMenuHoverOpenDelay(delay time.Duration) DropdownMenuOption",
    "DropdownMenuHoverCloseDelay(delay time.Duration) DropdownMenuOption"
  ]
}
-->

# Menu / Dropdown 菜单

Menu 使用 MD3 surface container、level 2 elevation、48dp menu item 和 bounded ripple。选中项使用 `SecondaryContainer`，禁用项使用统一 disabled content。Dropdown menu 打开时使用高度 reveal 动画：内容先按完整尺寸布局，再通过裁剪高度展开，因此内部 item 不会在动画中重新排版或挤压。

`Menu` 适合已经有外层弹层的场景；`DropdownMenu` 适合把触发器和菜单面板放在一起。

- `MenuItem.Children` 声明子菜单，父项 hover/focus 时在右侧打开。
- `MenuItem.Divider` 渲染分隔线。
- `MenuItem.KeepOpen` 可让叶子项选择后保持 DropdownMenu 打开。
- `MenuItem.Type`、`Href`、`Target`、`TypeaheadText` 对齐 Material Web item API，当前 `Href/Target` 作为结构化数据保留。
- `DropdownMenuQuick(true)` 跳过打开/关闭动画。
- `DropdownMenuXOffset` / `DropdownMenuYOffset` 设置锚点偏移。
- `DropdownMenuAnchorCorner` / `DropdownMenuMenuCorner` 控制锚点和菜单角对齐。
- `DropdownMenuDefaultFocusOf` 和 `DropdownMenuPositioningOf` 对齐 Material Web API；当前用于配置语义保留，后续会继续扩展键盘 focus 和 fixed positioning 细节。

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
        {Key: "copy", Label: "Copy", Leading: ui.Icon("content_copy")},
        {Key: "share", Label: "Share", Leading: ui.Icon("share")},
        {Key: "archive", Label: "Archive", Leading: ui.Icon("archive"), Children: []ui.MenuItem{
            {Key: "archive-today", Label: "Today"},
            {Key: "archive-week", Label: "This week"},
        }},
        {Divider: true},
        {Key: "delete", Label: "Delete", Disabled: true},
    },
    ui.DropdownMenuSelectedKey(selected.Value()),
    ui.DropdownMenuWidth(220),
    ui.DropdownMenuMaxHeight(260),
    ui.DropdownMenuXOffset(4),
    ui.DropdownMenuAnchorCorner(ui.MenuCornerEndStart),
    ui.DropdownMenuMenuCorner(ui.MenuCornerStartStart),
    ui.DropdownMenuDefaultFocusOf(ui.MenuDefaultFocusFirstItem),
    ui.DropdownMenuTypeaheadDelay(500*time.Millisecond),
    ui.DropdownMenuNoNavigationWrap(false),
    ui.DropdownMenuDecoration(ui.Bg(ui.UseTheme(ctx).Colors.SurfaceContainer).WithRad(10)),
    ui.DropdownMenuOnOpenChange(func(ctx *ui.Context, next bool) { open.Set(next) }),
    ui.DropdownMenuOnSelect(func(ctx *ui.Context, key string) { selected.Set(key) }),
)
```

静态菜单可在外层弹层中直接使用 `MenuElement`，并通过宽度、高度和 Decoration 统一控制样式：

```go
ui.MenuElement(
    items,
    ui.MenuSelectedKey(selected.Value()),
    ui.MenuWidth(190),
    ui.MenuMaxHeight(160),
    ui.MenuQuick(true),
    ui.MenuDecoration(ui.Bg(ui.UseTheme(ctx).Colors.SurfaceContainerLow).WithRad(10)),
    ui.MenuOnSelect(func(ctx *ui.Context, key string) { selected.Set(key) }),
)
```
