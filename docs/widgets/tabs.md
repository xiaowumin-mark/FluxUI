<!-- fluxui-doc-meta
{
  "id": "tabs",
  "title": "Tabs 标签栏",
  "category": "导航组件",
  "order": 400,
  "summary": "Tabs 用于同级视图切换。",
  "example": { "id": "tabs_basic" },
  "apis": [
    "Tabs(active string, items []TabItem, opts ...TabsOption) Widget",
    "TabsElement(active string, items []TabItem, opts ...TabsOption) Element",
    "TabsOnChange(fn func(ctx *Context, key string)) TabsOption",
    "TabsScrollable(scrollable bool) TabsOption",
    "TabsPrimary(primary bool) TabsOption",
    "TabsSecondary(secondary bool) TabsOption",
    "TabsAutoActivate(autoActivate bool) TabsOption",
    "TabsInlineIcon(inline bool) TabsOption",
    "TabsFullWidth(fullWidth bool) TabsOption",
    "TabsIndicatorColor(col color.NRGBA) TabsOption",
    "TabsTextColor(col color.NRGBA) TabsOption",
    "TabsActiveTextColor(col color.NRGBA) TabsOption",
    "TabsDecoration(d Decoration) TabsOption",
    "TabsTabDecoration(d Decoration) TabsOption",
    "NewTabsRef() *TabsRef",
    "TabsAttachRef(ref *TabsRef) TabsOption",
    "(*TabsRef).SetActive(key string)"
  ]
}
-->

# Tabs 标签栏

## 组件说明
Tabs 用于在同一层级内容中切换不同子页面，适合文档页、设置页、详情页多标签场景。

## MD3 默认样式

- inactive label 使用 `OnSurfaceVariant`。
- active label 和 indicator 使用 `Primary`。
- 标签文字默认使用 `LabelLarge`。
- hover/pressed 使用统一 state layer。
- Primary tabs 使用 64dp icon+label stacked 布局，inline icon 可通过 `TabsInlineIcon(true)` 开启。
- Secondary tabs 通过 `TabsSecondary(true)` 使用 48dp 高度和 full-width indicator。
- `TabItem.Icon` 对应 Material Web 的 `slot="icon"`，`TabItem.Disabled` 禁止点击和 state layer。
- active indicator 使用 row 级共享绘制：切换时从旧 tab 的实际 indicator rect 过渡到新 tab 的实际 rect，模拟 Material Web `translateX(...) scaleX(...)` 的移动和拉伸效果。

## 使用方法
- `active` 标识当前选中标签 key。
- `items` 定义可切换标签集合。
- 通过 `TabsOnChange` 回写状态。
- 需要外部命令式切换标签时，使用 `TabsAttachRef` + `SetActive`。
- 大量标签用 `TabsScrollable(true)`；固定等宽标签用 `TabsFullWidth(true)`。

## 使用示例

### 完整 Material 示例

下面的示例覆盖官方 demo 中常见的 Primary、Secondary、Scrolling、Custom、Primary and Secondary、Dynamic Tabs 组合。

```go
func TabsShowcase(ctx *ui.Context) ui.Element {
    active := ui.UseState(ctx, "keyboard")
    scrollingActive := ui.UseState(ctx, "scroll-keyboard-0")
    travel := ui.UseState(ctx, "travel")
    movie := ui.UseState(ctx, "star-wars")
    ref := ui.UseRef(ctx, ui.NewTabsRef())
    th := ui.UseTheme(ctx)

    instrumentTabs := []ui.TabItem{
        {Key: "keyboard", Label: "Keyboard", Icon: ui.Icon("keyboard")},
        {Key: "guitar", Label: "Guitar", Icon: ui.Icon("tune")},
        {Key: "drums", Label: "Drums", Icon: ui.Icon("graphic_eq")},
        {Key: "bass", Label: "Bass", Icon: ui.Icon("speaker")},
        {Key: "saxophone", Label: "Saxophone", Icon: ui.Icon("nightlife")},
    }
    travelTabs := []ui.TabItem{
        {Key: "travel", Label: "Travel", Icon: ui.Icon("flight")},
        {Key: "hotel", Label: "Hotel", Icon: ui.Icon("hotel")},
        {Key: "activities", Label: "Activities", Icon: ui.Icon("hiking")},
        {Key: "food", Label: "Food", Icon: ui.Icon("restaurant")},
    }
    scrollingTabs := make([]ui.TabItem, 0, 20)
    for i := 0; i < 4; i++ {
        for _, item := range instrumentTabs {
            scrollingTabs = append(scrollingTabs, ui.TabItem{
                Key:   fmt.Sprintf("scroll-%s-%d", item.Key, i),
                Label: item.Label,
                Icon:  item.Icon,
            })
        }
    }

    return ui.ColumnElement(
        ui.TextElement("Primary Tabs", ui.TextType(th.Types.TitleMedium)),
        ui.TabsElement(
            active.Value(),
            instrumentTabs,
            ui.TabsFullWidth(true),
            ui.TabsAttachRef(ref.Current),
            ui.TabsOnChange(func(ctx *ui.Context, key string) { active.Set(key) }),
        ),
        ui.TextElement(active.Value()),

        ui.VSpacerElement(20),
        ui.TextElement("Secondary Tabs", ui.TextType(th.Types.TitleMedium)),
        ui.TabsElement(
            travel.Value(),
            travelTabs,
            ui.TabsSecondary(true),
            ui.TabsInlineIcon(true),
            ui.TabsFullWidth(true),
            ui.TabsOnChange(func(ctx *ui.Context, key string) { travel.Set(key) }),
        ),
        ui.TextElement(travel.Value()),

        ui.VSpacerElement(20),
        ui.TextElement("Scrolling Tabs", ui.TextType(th.Types.TitleMedium)),
        ui.TabsElement(
            scrollingActive.Value(),
            scrollingTabs,
            ui.TabsScrollable(true),
            ui.TabsOnChange(func(ctx *ui.Context, key string) { scrollingActive.Set(key) }),
        ),

        ui.VSpacerElement(20),
        ui.TextElement("Custom Tabs", ui.TextType(th.Types.TitleMedium)),
        ui.TabsElement(
            travel.Value(),
            travelTabs,
            ui.TabsFullWidth(true),
            ui.TabsIndicatorColor(ui.NRGBA(174, 45, 120, 255)),
            ui.TabsActiveTextColor(ui.NRGBA(174, 45, 120, 255)),
            ui.TabsOnChange(func(ctx *ui.Context, key string) { travel.Set(key) }),
        ),

        ui.VSpacerElement(20),
        ui.TextElement("Primary and Secondary Tabs", ui.TextType(th.Types.TitleMedium)),
        ui.TabsElement(
            "movies",
            []ui.TabItem{
                {Key: "movies", Label: "Movies", Icon: ui.Icon("videocam")},
                {Key: "photos", Label: "Photos", Icon: ui.Icon("image")},
                {Key: "music", Label: "Music", Icon: ui.Icon("music_note")},
            },
            ui.TabsFullWidth(true),
        ),
        ui.TabsElement(
            movie.Value(),
            []ui.TabItem{
                {Key: "star-wars", Label: "Star Wars"},
                {Key: "avengers", Label: "Avengers"},
                {Key: "jaws", Label: "Jaws"},
                {Key: "frozen", Label: "Frozen"},
            },
            ui.TabsSecondary(true),
            ui.TabsFullWidth(true),
            ui.TabsOnChange(func(ctx *ui.Context, key string) { movie.Set(key) }),
        ),

        ui.VSpacerElement(20),
        ui.TextElement("Dynamic Tabs", ui.TextType(th.Types.TitleMedium)),
        ui.RowElement(
            ui.IconButtonElement(ui.IconElement("add"), ui.IconButtonOnClick(func(ctx *ui.Context) {
                ref.Current.SetActive("guitar")
            })),
            ui.IconButtonElement(ui.IconElement("chevron_left"), ui.IconButtonOnClick(func(ctx *ui.Context) {
                ref.Current.SetActive("keyboard")
            })),
            ui.IconButtonElement(ui.IconElement("chevron_right"), ui.IconButtonOnClick(func(ctx *ui.Context) {
                ref.Current.SetActive("saxophone")
            })),
        ),
    )
}
```

### React-style Element

```go
func SettingsTabs(ctx *ui.Context) ui.Element {
    active := ui.UseState(ctx, "profile")
    return ui.TabsElement(
        active.Value(),
        []ui.TabItem{
            {Key: "profile", Label: "Profile", Icon: ui.Icon("person")},
            {Key: "team", Label: "Team", Icon: ui.Icon("group")},
            {Key: "billing", Label: "Billing", Icon: ui.Icon("payments"), Disabled: true},
        },
        ui.TabsInlineIcon(true),
        ui.TabsOnChange(func(ctx *ui.Context, key string) {
            active.Set(key)
        }),
    )
}
```

`TabsElement` 已可在 `RunElement` root 下直接使用。`active` key、滚动标签栏状态、`TabsOnChange` 和 `TabsAttachRef` 仍由底层 widget host / 调用方状态管理。切换标签页对应的内容子树建议由调用方用 `Key` 或路由参数显式表达 identity。

### Legacy Widget
旧 `ui.Tabs` / `Widget` 写法继续可用：

```go
active := ui.State[string](ctx)
ui.Tabs(
    active.Value(),
    []ui.TabItem{
        {Key: "overview", Label: "Overview", Icon: ui.Icon("dashboard")},
        {Key: "api", Label: "API", Icon: ui.Icon("code")},
        {Key: "example", Label: "Example", Icon: ui.Icon("play_circle")},
    },
    ui.TabsSecondary(true),
    ui.TabsOnChange(func(ctx *ui.Context, key string) {
        active.Set(key)
    }),
)
```
