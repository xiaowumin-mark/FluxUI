<!-- fluxui-doc-meta
{
  "id": "material3_design_plan",
  "title": "Material Design 3 默认样式计划",
  "category": "使用指南",
  "order": 120,
  "summary": "记录 FluxUI 默认组件样式向 Material Design 3 对齐的设计规范、差距分析、实施步骤和验收标准。",
  "example": { "id": "material3_showcase" },
  "apis": [
    "type ColorScheme = theme.ColorScheme",
    "type ShapeScale = theme.ShapeScale",
    "type TypeScale = theme.TypeScale",
    "NewTheme(cs ColorScheme) *Theme",
    "FilledButton(child Widget, opts ...ButtonOption) Widget",
    "FilledTonalButton(child Widget, opts ...ButtonOption) Widget",
    "OutlinedButton(child Widget, opts ...ButtonOption) Widget",
    "TextButton(child Widget, opts ...ButtonOption) Widget",
    "ElevatedButton(child Widget, opts ...ButtonOption) Widget",
    "OutlinedTextField(value string, opts ...InputOption) Widget",
    "FilledTextField(value string, opts ...InputOption) Widget",
    "FilledCard(child Widget, opts ...CardOption) Widget",
    "ElevatedCard(child Widget, opts ...CardOption) Widget",
    "OutlinedCard(child Widget, opts ...CardOption) Widget"
  ]
}
-->

# Material Design 3 默认样式计划

本文档是 FluxUI 默认组件样式向 Material Design 3（MD3）对齐的执行入口。目标不是给现有组件临时换色，而是建立一套可复用的 MD3 token 系统，让主题、默认组件、文档示例和后续 Element API 都从同一套设计语言派生。

## 目标

- 默认 Light/Dark 主题接近 MD3 视觉语言。
- 组件默认样式使用语义 token，而不是在各组件里分散硬编码颜色、圆角、阴影和 padding。
- Legacy `Widget` API 与 React-style `Element` API 共享同一套主题结果。
- 保留现有 API 兼容性，新 MD3 能力优先作为扩展字段、新组件变体和推荐默认值进入。
- 增加 Material 3 showcase 示例，作为人工视觉回归基准。

## 官方依据

MD3 主题系统的核心是 color scheme、typography 和 shapes。组件通过这些 token 派生默认样式，并在状态、层级、强调程度上选择不同语义角色。Android Compose 的 Material 3 文档也以 `MaterialTheme.colorScheme`、`MaterialTheme.typography`、`MaterialTheme.shapes` 作为主题入口，并说明 color roles 会根据组件状态、重要程度和强调程度使用。

参考资料：

- Android Material 3 theme overview: https://developer.android.google.cn/develop/ui/compose/designsystems/material3
- Material 3 color roles: https://m3.material.io/styles/color/roles
- Material 3 typography: https://m3.material.io/styles/typography
- Material 3 shape: https://m3.material.io/styles/shape
- Material 3 elevation: https://m3.material.io/styles/elevation
- Material 3 interaction states: https://m3.material.io/foundations/interaction/states
- Material Web components and tokens: https://material-web.dev/

## 当前状态

截至 2026-05-28，当前工作区已经开始落地 MD3 基础设施：

| Phase | 状态 | 当前结果 |
| --- | --- | --- |
| Phase 1 Theme Foundation | 已完成 | `ColorScheme` 已扩展 MD3 roles，`Theme` 已加入 `Shapes` 和 `Types`，Light/Dark 默认值已切到 MD3 baseline，并保留兼容字段同步。 |
| Phase 2 Shared Visual Helpers | 已完成 | 已新增 state layer、disabled color、tonal elevation、elevation shadow、颜色混合 helper。 |
| Phase 3 First Component Pass | 已完成 | Button/TextField/Card 已提供 MD3 变体，旧 `Button`/`TextField`/`Card` 映射到稳定默认变体。 |
| Phase 4 Selection And Navigation | 已完成 | Checkbox、Radio、Switch、Slider、AppBar、BottomNavigation、Tabs、Dialog、Popup、Toast 已统一使用 MD3 color/type/shape/elevation 默认值。 |
| Phase 5 Docs And Showcase | 已完成 | 已新增 `examples/material3_showcase` 和 `docs/guides/material3.md`，并更新 theme/style/widgets 文档。 |
| Phase 6 API Cleanup | 已完成 | 文档已记录 legacy API 与 MD3 variant API 的对应关系；旧单属性 option 保留为兼容覆盖入口。 |

## 设计原则

### Token First

默认样式必须从 token 派生。组件不应直接写死品牌色、灰阶、圆角和状态透明度；必要时只保留局部 option 作为 escape hatch。

优先 token：

- Color roles
- Type scale
- Shape scale
- Elevation levels
- State layer opacities
- Component defaults

### Compatibility First

现有字段不能直接删除。`Theme.Primary`、`Theme.Surface`、`Theme.TextColor`、`Theme.TextOnPrimary`、`Theme.Disabled`、`Theme.TextSize` 继续作为兼容字段，由新 token 同步生成。

### Variant First

MD3 中 Button、TextField、Card 不是单一默认样式。FluxUI 应提供清晰的组件变体，而不是把所有行为塞进一个组件加大量 option。

推荐映射：

- `Button` 保持兼容，默认映射到 MD3 Filled Button。
- 新增 `FilledButton`、`FilledTonalButton`、`OutlinedButton`、`TextButton`、`ElevatedButton`。
- `TextField` 保持兼容，新增 `OutlinedTextField`、`FilledTextField`。
- `Card` 默认映射到 Filled Card，新增 `ElevatedCard`、`OutlinedCard`。

### State Layer 一致

hover、focus、pressed、dragged 状态必须使用统一透明度。初始值：

| State | Opacity |
| --- | ---: |
| Hover | 0.08 |
| Focus | 0.12 |
| Pressed | 0.12 |
| Dragged | 0.16 |

组件只决定使用哪个 `onColor`，不各自发明混色比例。

## 目标 Token 设计

### ColorScheme

`theme.ColorScheme` 应覆盖 MD3 常用 roles，同时保留 FluxUI 已有业务色。

目标字段：

```go
type ColorScheme struct {
    Primary            color.NRGBA
    OnPrimary          color.NRGBA
    PrimaryContainer   color.NRGBA
    OnPrimaryContainer color.NRGBA

    Secondary            color.NRGBA
    OnSecondary          color.NRGBA
    SecondaryContainer   color.NRGBA
    OnSecondaryContainer color.NRGBA

    Tertiary            color.NRGBA
    OnTertiary          color.NRGBA
    TertiaryContainer   color.NRGBA
    OnTertiaryContainer color.NRGBA

    Error            color.NRGBA
    OnError          color.NRGBA
    ErrorContainer   color.NRGBA
    OnErrorContainer color.NRGBA

    Background   color.NRGBA
    OnBackground color.NRGBA

    Surface                 color.NRGBA
    OnSurface               color.NRGBA
    SurfaceVariant          color.NRGBA
    OnSurfaceVariant        color.NRGBA
    SurfaceContainerLowest  color.NRGBA
    SurfaceContainerLow     color.NRGBA
    SurfaceContainer        color.NRGBA
    SurfaceContainerHigh    color.NRGBA
    SurfaceContainerHighest color.NRGBA

    Outline        color.NRGBA
    OutlineVariant color.NRGBA

    InverseSurface   color.NRGBA
    InverseOnSurface color.NRGBA
    InversePrimary   color.NRGBA

    Scrim  color.NRGBA
    Shadow color.NRGBA

    Success   color.NRGBA
    OnSuccess color.NRGBA
    Warning   color.NRGBA
    OnWarning color.NRGBA
    Disabled  color.NRGBA

    // Compatibility alias. Prefer SurfaceVariant / SurfaceContainer* in new code.
    SurfaceMuted color.NRGBA
}
```

落地要求：

- `LightColors()` 和 `DarkColors()` 填完整字段。
- `New(cs)` 对自定义旧色板做 normalize，避免新增字段为零值。
- `New(cs)` 同步兼容字段：
  - `Theme.Primary = cs.Primary`
  - `Theme.Surface = cs.Surface`
  - `Theme.SurfaceMuted = cs.SurfaceVariant`
  - `Theme.TextColor = cs.OnSurface`
  - `Theme.TextOnPrimary = cs.OnPrimary`
  - `Theme.Disabled = cs.Disabled`

### ShapeScale

新增 `theme.ShapeScale`：

```go
type ShapeScale struct {
    None       float32
    ExtraSmall float32
    Small      float32
    Medium     float32
    Large      float32
    ExtraLarge float32
    Full       float32
}
```

默认值：

| Token | dp |
| --- | ---: |
| None | 0 |
| ExtraSmall | 4 |
| Small | 8 |
| Medium | 12 |
| Large | 16 |
| ExtraLarge | 28 |
| Full | 999 |

组件映射：

| Component | Shape |
| --- | --- |
| Button | `Full` |
| TextField | `ExtraSmall` 或 `Small` |
| Card | `Medium` |
| Dialog | `ExtraLarge` |
| Popup/Menu | `ExtraSmall` 或 `Small` |
| Checkbox | `ExtraSmall` |
| Switch thumb/track | `Full` |

### TypeScale

新增 `theme.TypeScale`。第一版先表达 size、line height、weight；letter spacing 后续扩展。

```go
type TextStyle struct {
    Size       float32
    LineHeight float32
    Weight     FontWeight
}

type TypeScale struct {
    DisplayLarge  TextStyle
    DisplayMedium TextStyle
    DisplaySmall  TextStyle
    HeadlineLarge TextStyle
    HeadlineMedium TextStyle
    HeadlineSmall TextStyle
    TitleLarge    TextStyle
    TitleMedium   TextStyle
    TitleSmall    TextStyle
    BodyLarge     TextStyle
    BodyMedium    TextStyle
    BodySmall     TextStyle
    LabelLarge    TextStyle
    LabelMedium   TextStyle
    LabelSmall    TextStyle
}
```

默认字号和行高：

| Token | Size / LineHeight |
| --- | --- |
| DisplayLarge | 57 / 64 |
| DisplayMedium | 45 / 52 |
| DisplaySmall | 36 / 44 |
| HeadlineLarge | 32 / 40 |
| HeadlineMedium | 28 / 36 |
| HeadlineSmall | 24 / 32 |
| TitleLarge | 22 / 28 |
| TitleMedium | 16 / 24 |
| TitleSmall | 14 / 20 |
| BodyLarge | 16 / 24 |
| BodyMedium | 14 / 20 |
| BodySmall | 12 / 16 |
| LabelLarge | 14 / 20 |
| LabelMedium | 12 / 16 |
| LabelSmall | 11 / 16 |

首批组件映射：

| Component | Text token |
| --- | --- |
| Button text | `LabelLarge` |
| TextField input | `BodyLarge` |
| TextField helper/error | `BodySmall` |
| AppBar title | `TitleLarge` |
| Card title | `TitleMedium` |
| Card body | `BodyMedium` |
| Tabs label | `TitleSmall` 或 `LabelLarge` |
| Navigation label | `LabelMedium` |

### Elevation

MD3 elevation 优先用 tonal elevation 区分 surface 层级，再辅以轻量 shadow。公共 helper：

```go
func SurfaceAtElevation(cs theme.ColorScheme, level int) color.NRGBA
func ElevationShadow(cs theme.ColorScheme, level int) style.BoxShadow
```

初始 tonal overlay 比例：

| Level | Primary overlay |
| --- | ---: |
| 0 | 0 |
| 1 | 0.05 |
| 2 | 0.08 |
| 3 | 0.11 |
| 4 | 0.12 |
| 5 | 0.14 |

组件映射：

| Component | Elevation |
| --- | --- |
| AppBar / NavigationBar | level 2 或 `SurfaceContainer` |
| Filled Card | level 0 |
| Elevated Card | level 1 |
| Dialog | level 3 |
| Popup/Menu | level 2 |
| Toast/Snackbar | `InverseSurface`，不走普通 surface elevation |

## 组件落地规范

### Button

MD3 目标：

| Variant | Container | Content | Border | Elevation |
| --- | --- | --- | --- | --- |
| Filled | `Primary` | `OnPrimary` | none | 0 |
| Filled Tonal | `SecondaryContainer` | `OnSecondaryContainer` | none | 0 |
| Outlined | transparent | `Primary` | `Outline` | 0 |
| Text | transparent | `Primary` | none | 0 |
| Elevated | `SurfaceContainerLow` | `Primary` | none | 1 |

落地步骤：

1. 保留 `Button(...)`，内部映射到 Filled Button。
2. 新增 `FilledButton`、`FilledTonalButton`、`OutlinedButton`、`TextButton`、`ElevatedButton`。
3. 默认 radius 使用 `Theme.Shapes.Full`。
4. 默认文字使用 `Theme.Types.LabelLarge`。
5. hover/pressed 使用 `style.StateLayer`。
6. disabled 使用 `style.DisabledContainer` 和 `style.DisabledContent`。
7. `ButtonDecoration` 继续作为覆盖入口。
8. 补齐 Outlined Button 的 border 和 Elevated Button 的 shadow。

### TextField

MD3 目标：

| Variant | Container | Border / Indicator | Text | Placeholder |
| --- | --- | --- | --- | --- |
| Filled | `SurfaceContainerHighest` | focus 使用 `Primary` | `OnSurface` / `BodyLarge` | `OnSurfaceVariant` |
| Outlined | `Surface` | normal `Outline`，focus `Primary` | `OnSurface` / `BodyLarge` | `OnSurfaceVariant` |

落地步骤：

1. 新增内部 `inputVariant`。
2. 保留 `TextField(...)` 兼容行为，并明确默认映射策略。
3. 新增 `FilledTextField` 和 `OutlinedTextField`。
4. focus、hover、disabled、error 分支统一读取 MD3 color roles。
5. 在设计好 API 后再加入 `InputError(...)`、`InputSupportingText(...)`，避免破坏现有简洁用法。
6. 给 variant 默认值和兼容行为补测试。

### Card

MD3 目标：

| Variant | Container | Border | Shape | Elevation |
| --- | --- | --- | --- | --- |
| Filled | `SurfaceContainerHighest` | none | `Medium` | 0 |
| Elevated | `SurfaceContainerLow` | none | `Medium` | 1 |
| Outlined | `Surface` | `OutlineVariant` | `Medium` | 0 |

落地步骤：

1. `Card` 默认映射 Filled Card。
2. 新增 `FilledCard`、`ElevatedCard`、`OutlinedCard`。
3. `CardDecoration` 继续覆盖 container、padding、radius、border、shadow。
4. Elevated Card 使用 `SurfaceAtElevation` 和 `ElevationShadow`。
5. Outlined Card 使用 `OutlineVariant`，避免与交互控件 `Outline` 过度竞争。

### Selection Controls

Checkbox、Radio、Switch、Slider 统一规则：

- active: `Primary`
- on active: `OnPrimary`
- inactive track/border: `Outline` / `SurfaceVariant`
- disabled container: `OnSurface` 12%
- disabled content: `OnSurface` 38%
- hover/pressed/focus: state layer

落地步骤：

1. 先替换颜色和状态 helper，不重写布局。
2. 给 checked/unchecked、enabled/disabled、hover/pressed 分支补覆盖。
3. 确认状态切换不会改变 layout 尺寸。

### Navigation And Overlays

目标映射：

| Component | Container | Content | Shape / Elevation |
| --- | --- | --- | --- |
| AppBar | `Surface` 或 `SurfaceContainer` | title `TitleLarge` | level 0/2 |
| BottomNavigation | `SurfaceContainer` | selected `Primary` 或 `OnSecondaryContainer` | indicator `SecondaryContainer` |
| Tabs | `Surface` | selected `Primary` | indicator `Primary` |
| Dialog | `SurfaceContainerHigh` | `OnSurface` | `ExtraLarge`, level 3 |
| Popup/Menu | `SurfaceContainer` | `OnSurface` | `ExtraSmall/Small`, level 2 |
| Toast/Snackbar | `InverseSurface` | `InverseOnSurface` | small radius |

落地步骤：

1. AppBar、BottomNavigation、Tabs 先统一色彩和文字 token。
2. Dialog、Popup、Toast 统一 surface、shape、elevation。
3. 保留现有 option 覆盖能力，新增默认值只在未显式设置时生效。

## 分阶段实施计划

### Phase 1: Theme Foundation

Scope:

- 扩展 `theme.ColorScheme` 到 MD3 color roles。
- 新增 `theme.ShapeScale`、`theme.TypeScale`。
- `Theme` 新增 `Shapes ShapeScale`、`Types TypeScale`。
- `LightColors()` / `DarkColors()` 填完整默认值。
- `Theme.New` normalize 自定义旧色板并同步兼容字段。

Tests:

- 默认 Light/Dark 不返回零值关键 token。
- 兼容字段与新 roles 同步。
- 自定义旧 `ColorScheme` 缺字段时有合理 fallback。

### Phase 2: Shared Visual Helpers

Scope:

- 新增 tonal elevation helper。
- 新增 state layer helper。
- 新增 disabled color helper。
- 统一 `MixNRGBA`，避免 widget 重复实现混色。

Tests:

- overlay opacity clamp。
- disabled 输出 alpha 符合 12% / 38%。
- tonal elevation level 0..5 输出稳定。

### Phase 3: First Component Pass

Scope:

- Button: Filled/FilledTonal/Outlined/Text/Elevated。
- TextField: Filled/Outlined。
- Card: Filled/Elevated/Outlined。

Rules:

- 旧 API 继续可用。
- 新变体优先放在 `widget`，再由 `ui` re-export。
- Element wrapper 后续跟进，不在第一批强行扩所有 wrapper。

Tests:

- 构造函数和 option 不破坏现有测试。
- 对关键默认值写单元测试，避免后续回退。

### Phase 4: Selection And Navigation Pass

Scope:

- Checkbox、Radio、Switch、Slider 状态色统一。
- AppBar、BottomNavigation、Tabs 使用 MD3 type/color/shape。
- Dialog、Popup、Toast 使用 MD3 surface/elevation/inverse roles。

Tests:

- 状态切换不改变 layout 尺寸。
- disabled/hover/pressed 分支有覆盖。

### Phase 5: Docs And Showcase

Scope:

- 新增 `examples/material3_showcase`。
- 展示 Light/Dark、color roles、type scale、shape scale。
- 展示 Button/TextField/Card/Selection/Navigation/Overlay 的默认状态和变体。
- 更新 `docs/theme`、`docs/style`、`docs/widgets` 中涉及默认样式的页面。

Acceptance:

- showcase 可以运行。
- 每次默认样式调整后，都能通过 showcase 做人工视觉回归。

### Phase 6: API Cleanup

Scope:

- 为旧的单属性 style options 标注推荐替代方案，但不立即废弃。
- 记录 legacy API 与 MD3 variant API 的对应关系。
- 确认 Element API 后续如何暴露相同变体。

## 文件级落点

| Area | Files |
| --- | --- |
| Theme roles | `theme/colors.go`, `theme/theme.go` |
| Typography | `theme/font.go`, `theme/typescale.go` |
| Shape scale | `theme/shape.go` |
| Elevation/state helpers | `style/material3.go`, `style/decoration.go` |
| Button | `widget/button.go`, `ui/ui.go`, `ui/extended_types.go` |
| TextField | `widget/input.go`, `ui/ui.go`, `ui/extended_types.go` |
| Card/media | `widget/media_card.go`, `ui/extended_types.go` |
| Selection | `widget/checkbox.go`, `widget/switch.go`, `widget/slider.go`, `widget/selection.go` |
| Navigation | `widget/navigation.go`, `widget/tabs_dialog_toast.go` |
| Docs | `docs/theme/*.md`, `docs/style/*.md`, `docs/widgets/*.md`, this file |
| Showcase | `examples/material3_showcase/main.go` |

## 验收标准

每个阶段完成必须满足：

- `go test ./...`
- `go vet ./...`
- 如果涉及 goroutine/state：`go test -race ./...`
- 新增或更新相关单元测试。
- showcase 可以运行并展示该阶段涉及组件。
- 文档说明旧 API 与新 MD3 行为的关系。

## 非目标

- 不在第一阶段实现 Material You 动态取色。
- 不立即移除现有 `Theme` 兼容字段。
- 不一次性重写所有组件绘制逻辑。
- 不把 MD3 作为唯一可用风格；未来仍允许自定义 theme 覆盖。
