<!-- fluxui-doc-meta
{
  "id": "material3_roadmap",
  "title": "Material Design 3 长期路线图",
  "category": "使用指南",
  "order": 121,
  "summary": "总结 FluxUI 融合 Material Design 3 的当前进展，并规划长期推进路线、组件覆盖、交互反馈、文档示例和验收机制。",
  "example": { "id": "material3_showcase" },
  "apis": [
    "RunElement(root Component, opts ...AppOption) error",
    "type ColorScheme = theme.ColorScheme",
    "type ShapeScale = theme.ShapeScale",
    "type TypeScale = theme.TypeScale",
    "FilledButtonElement(child Element, opts ...ButtonOption) Element",
    "PressableElement(child Element, onClick func(ctx *Context), opts ...PressableOption) Element"
  ]
}
-->

# Material Design 3 长期路线图

本文档记录 FluxUI 融合 Material Design 3（MD3）的当前进程和长期计划。`docs/material3-design-plan.md` 仍作为首轮默认样式落地的执行文档；本文档作为长期路线图，关注后续如何把 MD3 变成 FluxUI 的默认设计语言、交互基础设施和文档验收体系。

## 当前进展

截至 2026-05-29，FluxUI 已经完成 MD3 融合的基础阶段。

### 已完成的基础设施

- `theme.ColorScheme` 已扩展为接近 MD3 的语义色板，覆盖 primary、secondary、tertiary、error、surface、container、outline、inverse、scrim、shadow 等角色。
- `theme.Theme` 已加入 `Shapes` 和 `Types`，用于承载 MD3 shape scale 与 type scale。
- Light/Dark 默认主题已切换到 MD3 baseline，并继续同步旧字段，保证旧代码不被破坏。
- `style/material3.go` 已提供 state layer、disabled color、tonal elevation、elevation shadow 等共享 helper。
- `internal/ripple.go` 已新增可复用 ripple primitive，基于 Gio clickable history 绘制有界按压扩散动画。
- `internal.LayoutButton` 已从 Gio material button layout 切到 FluxUI 自绘圆角背景、边框、阴影和 ripple，修复 full radius 不能正确呈现胶囊形的问题。

### 已完成的组件覆盖

- Button 已提供 `FilledButton`、`FilledTonalButton`、`OutlinedButton`、`TextButton`、`ElevatedButton`，默认 `Button` 映射到 Filled Button。
- Button 默认圆角使用 `Theme.Shapes.Full`，实际绘制时按组件尺寸 clamp，呈现 MD3 胶囊按钮。
- Button 已接入 ripple，按压反馈使用按钮 foreground 颜色和 MD3 pressed state layer 透明度。
- TextField 已提供 `OutlinedTextField` 和 `FilledTextField`。
- Card 已提供 `FilledCard`、`ElevatedCard`、`OutlinedCard`。
- Checkbox、Radio、Switch、Slider、AppBar、BottomNavigation、Tabs、Dialog、Popup、Toast 已开始使用 MD3 color/type/shape/elevation token。
- `ClickArea` 已保留为兼容别名，新增更符合 GUI 语义的 `Pressable` / `PressableElement` 作为推荐无固定视觉点击区域。

### 已完成的 React-style API 对齐

- `examples/material3_showcase` 已改为推荐的 `RunElement` / `Element` 写法。
- MD3 新组件变体已在 `ui` 层提供 Element wrapper。
- 长期方向是文档、示例和新功能都优先使用 React-style Element API，legacy Widget API 继续兼容但不作为首选展示方式。

### 已完成的文档与验证

- `docs/material3-design-plan.md` 已记录首轮 MD3 默认样式计划。
- `docs/guides/material3.md` 已作为文档浏览器内的 MD3 使用指南入口。
- `examples/material3_showcase` 已作为人工视觉回归基准。
- 当前验证命令已通过：

```sh
go test ./...
go vet ./...
```

## 长期目标

FluxUI 的 MD3 推进不只是“默认颜色更像 Material”。长期目标如下：

- 默认主题即 MD3：用户不配置主题时，组件视觉、状态、圆角、字体、层级和交互反馈应自然符合 MD3。
- Token first：组件默认值只能从 theme/style token 派生，避免在组件里分散硬编码。
- Element first：新文档、新示例、新能力优先通过 React-style Element API 展示。
- Interaction first：hover、focus、pressed、dragged、disabled、ripple、focus ring 形成统一交互系统，并为状态变化提供克制、稳定的组件动画。
- Variant first：MD3 有明确变体的组件，FluxUI 应提供命名清晰的变体 API。
- Compatibility first：旧 API 不突然删除，旧代码继续运行；新 API 和文档引导用户自然迁移。
- Visual regression first：默认样式每次调整都必须能通过 showcase 做人工视觉回归，后续逐步补自动截图回归。

## 设计原则

### Token 不下沉到组件硬编码

组件可以决定自己使用哪个 token，但不应自己发明颜色、圆角、阴影、透明度和字体层级。

正确方向：

- Button 使用 `Primary`、`OnPrimary`、`Shapes.Full`、`Types.LabelLarge`。
- Dialog 使用 `SurfaceContainerHigh`、`OnSurface`、`Shapes.ExtraLarge`、level 3 elevation。
- Navigation 使用 `SurfaceContainer`、`SecondaryContainer`、`Primary`、`Types.LabelMedium`。

需要避免：

- 在组件里写死灰阶。
- 每个组件重复一套 hover/pressed 混色比例。
- 用单个 `Theme.Primary` 推导所有强调、容器、边框和状态。

### MD3 默认，但允许自定义

MD3 是 FluxUI 默认设计语言，不应成为硬锁定风格。用户仍应能通过 `ThemeProviderElement`、`NewTheme`、`ColorScheme`、`Decoration`、组件 option 覆盖视觉。

默认路径应该完整、好看、一致；高级用户仍可覆盖。

### 交互反馈是基础设施

ripple、state layer、focus ring、dragged overlay 不应散落在 Button、Tabs、Navigation 各自实现里。它们应逐步沉淀为内部可复用 primitive：

- `DrawRipple`
- `StateLayer`
- focus indicator helper
- dragged/selected state helper
- disabled content/container helper

组件只传入 clickable、size、shape、color、opacity。

组件状态变化应优先使用轻量动画，而不是瞬间切换视觉结果。hover、pressed、focus、selected、expanded、disabled、loading 等状态都应评估是否需要过渡动画；动画应服务于状态感知，不应让桌面工具界面显得拖沓。

动画和视觉细节可以参考 GitHub/Primer 这类成熟产品界面：hover 背景过渡、按钮按压反馈、菜单展开、焦点描边、列表项选中、loading skeleton、toast 出入场等都可以作为节奏和克制度的参考。MD3 仍是 FluxUI 的主规范，GitHub/Primer 主要作为产品级交互质感参考。

### ClickArea 语义收敛为 Pressable

`ClickArea` 作为框架 API 会让用户感觉像低层命中测试对象，不像一个 GUI 组件。长期推荐改为：

- `Pressable`: 无固定视觉样式的通用可点击区域。
- `PressableElement`: React-style Element API 入口。
- `ClickArea`: 保留为 deprecated 兼容别名。

后续文档和示例应使用 `Pressable`，只在低层兼容章节提到 `ClickArea`。

## 推进阶段

### Phase A: MD3 基础稳定化

目标：把已落地的 MD3 基础设施稳定下来，减少后续回退风险。

任务：

- 为 `ColorScheme` normalize 补充完整测试，确保自定义旧色板不会出现关键 token 零值。
- 为 `ShapeScale`、`TypeScale` 增加默认值测试。
- 为 `SurfaceAtElevation`、`ElevationShadow`、`StateLayer`、disabled helper 增加边界测试。
- 为 `DrawRipple` 增加可测试的参数拆分，至少覆盖 opacity、diameter、fade-out 过期逻辑。
- 检查所有组件是否还存在重复混色逻辑，逐步替换到 `style/material3.go`。
- 检查所有默认圆角是否使用 `clampRoundedRadiusPx` 或等价逻辑，避免大圆角溢出。

验收：

```sh
go test ./theme ./style ./internal ./widget ./ui
go vet ./...
```

### Phase B: 交互反馈统一

目标：让所有可交互组件共享 MD3 状态、ripple 和轻量动画语言。

任务：

- Button 已接入 ripple，继续验证 disabled、outlined、text、elevated 的视觉结果。
- Button、Pressable、Tabs、BottomNavigation 等交互组件应支持 hover/pressed/focus 的平滑过渡，避免背景色、边框、阴影瞬间跳变。
- Tabs item 接入 ripple，ripple clip 到 tab item 区域，indicator 保持在内容之上。
- BottomNavigation item 接入 ripple，selected indicator 与 ripple 层级要清晰。
- Checkbox、Radio、Switch 增加 touch target ripple，注意 ripple 应围绕交互目标而不是只围绕图形本体。
- Slider thumb 增加 pressed/dragged state layer；拖动时使用 `StateLayerDraggedOpacity`。
- Select、Popup trigger、菜单项后续使用 Pressable + ripple。
- ContainerDecoration 若配置 `OnClick`，可选择只提供 state layer，不默认加 ripple，避免普通容器突然出现 Material 反馈。
- 参考 GitHub/Primer 的 hover、focus、menu、toast、loading 等交互样式，提炼适合 FluxUI 的默认动画时长、缓动曲线和视觉克制度。

验收：

- 所有 ripple 不改变 layout 尺寸。
- 所有 ripple 被正确裁剪，不溢出圆角容器。
- disabled 状态不绘制 hover/pressed/ripple。
- hover、focus、pressed、selected 等状态变化有稳定过渡，不出现闪烁、重叠或 layout shift。
- 使用 showcase 人工检查 Button、Tabs、BottomNavigation、Selection controls。

### Phase C: 组件 MD3 覆盖扩展

目标：从基础组件扩展到完整常用组件集。

优先组件：

- Select / Dropdown / Menu
- List item
- IconButton
- FloatingActionButton
- NavigationRail
- NavigationDrawer
- Snackbar action
- Tooltip
- Badge
- Chip
- SearchBar
- Progress indicators

组件策略：

- 有 MD3 明确变体的组件应提供显式变体 API。
- 无明显变体但有明确 token 的组件直接改默认值。
- 不足以稳定设计 API 的组件先进入 experimental 命名空间或文档标注。

示例 API 方向：

```go
FilledIconButtonElement(...)
OutlinedIconButtonElement(...)
SmallFloatingActionButtonElement(...)
AssistChipElement(...)
FilterChipElement(...)
ListItemElement(...)
```

验收：

- 每个新增 MD3 组件至少有一个 docs 页面、一个 docs browser 示例 ID、一个 showcase 分区。
- 默认样式覆盖 enabled、disabled、hover、pressed、selected/focused 等关键状态。

### Phase D: Typography 与 Density 完善

目标：让 FluxUI 的文字层级、间距和密度更接近真实产品使用。

任务：

- `TextElement` 增加使用 type token 的推荐 option，例如 `TextType(th.Types.BodyMedium)` 或更符合本项目 API 风格的封装。
- Button、TextField、Tabs、Navigation、Dialog、Card 默认文字全部从 `Types` 派生。
- 检查组件 padding 是否符合 MD3 的默认触控尺寸与视觉密度。
- 提供 density 配置策略，允许桌面应用使用 compact density，但默认仍保持 MD3 可触控目标。
- 补充 line height 的实际渲染支持和测试。

验收：

- Button、Navigation、TextField、Dialog 的文字 token 有明确文档。
- 组件不会因为字体变大导致裁剪或重叠。
- showcase 中加入 type scale 样张。

### Phase E: Dynamic Color 与品牌主题

目标：支持从品牌色生成完整 MD3 ColorScheme。

任务：

- 设计 `ColorSchemeFromSeed(seed color.NRGBA, opts ...ColorOption) ColorScheme`。
- 评估是否引入 HCT/CAM16 算法实现；如果不引入外部依赖，需要明确近似算法边界。
- 支持 light/dark 两套 scheme。
- 支持 error、success、warning 业务色与 MD3 roles 并存。
- 文档说明品牌主题与默认主题的关系。

验收：

- 给定 seed 能生成完整非零 ColorScheme。
- contrast 基本可读。
- showcase 支持 seed theme preview。

### Phase F: 文档与示例全面迁移

目标：文档成为用户学习 MD3 FluxUI 的主要路径。

任务：

- 所有新文档示例优先使用 `RunElement`、`Element`、`ThemeProviderElement`。
- 更新 `docs/widgets/click_area.md`，标注 `ClickArea` deprecated，推荐 `Pressable`。
- 新增 `docs/widgets/pressable.md`。
- 更新 `docs/widgets/button.md`，说明 Button variant、ripple、shape、disabled、state layer。
- 更新 `docs/widgets/tabs.md`、`bottom_navigation.md`、`checkbox.md`、`switch.md`、`radio_group.md`、`slider.md`，说明 MD3 交互反馈。
- `examples/material3_showcase` 持续扩展为总览示例。
- 旧 Widget 示例逐步迁移或标注 legacy。

验收：

- 文档浏览器能加载所有新增文档。
- 每篇 MD3 组件文档有 doc meta。
- 示例无 mojibake，无旧 API 优先展示。

### Phase G: 视觉回归体系

目标：避免默认样式在后续维护中无意退化。

任务：

- 为 `examples/material3_showcase` 增加固定窗口尺寸截图流程。
- 建立 desktop 和 narrow/mobile-like 两种 viewport。
- 对关键组件增加 pixel smoke check，至少确保非空、无明显裁剪、无异常重叠。
- 增加交互状态截图或短帧检查，覆盖 hover、pressed、focus、selected、expanded、toast enter/exit 等动画关键帧。
- 若 CI 环境支持，加入截图 artifact；不强制在第一版进行像素级 golden 比较。
- 将人工验收清单写入文档。

验收：

- 每次 MD3 默认样式调整都能产出 showcase 截图。
- 截图覆盖 Light/Dark、Button、TextField、Card、Selection、Navigation、Overlay、Ripple 状态。

### Phase H: 版本与兼容策略

目标：明确 MD3 默认化对现有用户的影响。

任务：

- 在 release notes 中标记默认视觉变化。
- 旧 API 不删除，但文档降级为兼容说明。
- 若引入 breaking change，必须配合迁移指南。
- 提供“恢复旧样式”的最小方案，至少能通过自定义 Theme/Decoration 复原主要颜色与圆角。
- 对 deprecated API 设置长期窗口，不做短期删除。

验收：

- `ClickArea`、legacy Widget API、旧 Theme 字段都有清晰说明。
- 用户能从文档判断应该用新 API 还是兼容 API。

## 组件覆盖矩阵

| 组件 | MD3 token | Variant API | Ripple | Element API | 文档 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| Button | 已接入 | 已接入 | 已接入 | 已接入 | 已接入 | 稳定化中 |
| TextField | 已接入 | 已接入 | 不适用 | 已接入 | 已接入 | 稳定化中 |
| Card | 已接入 | 已接入 | 不适用 | 已接入 | 已接入 | 稳定化中 |
| Checkbox | 部分接入 | 不适用 | 待接入 | 已接入 | 已接入 | 下一阶段 |
| Radio | 部分接入 | 不适用 | 待接入 | 已接入 | 已接入 | 下一阶段 |
| Switch | 部分接入 | 不适用 | 待接入 | 已接入 | 已接入 | 下一阶段 |
| Slider | 部分接入 | 不适用 | dragged 待完善 | 已接入 | 已接入 | 下一阶段 |
| Tabs | 部分接入 | 不适用 | 待接入 | 已接入 | 已接入 | 下一阶段 |
| BottomNavigation | 部分接入 | 不适用 | 待接入 | 已接入 | 已接入 | 下一阶段 |
| AppBar | 已接入 | 待扩展 | 可选 | 已接入 | 已接入 | 稳定化中 |
| Dialog | 已接入 | 不适用 | 不适用 | 已接入 | 已接入 | 稳定化中 |
| Popup/Menu | 部分接入 | 待扩展 | 菜单项待接入 | 已接入 | 已接入 | 下一阶段 |
| Toast/Snackbar | 已接入 | 待扩展 action | action 待接入 | 已接入 | 已接入 | 下一阶段 |
| Pressable | 不固定 | 不适用 | 可选待设计 | 已接入 | 待新增 | 新增 |
| Select | 部分接入 | 待设计 | 待接入 | 已接入 | 已有 | 后续扩展 |
| IconButton | 待接入 | 待新增 | 待接入 | 待新增 | 待新增 | 后续扩展 |
| FAB | 待新增 | 待新增 | 待接入 | 待新增 | 待新增 | 后续扩展 |
| Chip | 待新增 | 待新增 | 待接入 | 待新增 | 待新增 | 后续扩展 |
| ListItem | 待新增 | 待设计 | 待接入 | 待新增 | 待新增 | 后续扩展 |

## 下一批推荐任务

短期最值得做的任务：

1. 把 ripple primitive 接入 Tabs 和 BottomNavigation。
2. 把 selection controls 的 touch target ripple 统一起来。
3. 新增 `docs/widgets/pressable.md`，并更新 `docs/widgets/click_area.md` 为兼容说明。
4. 扩展 `examples/material3_showcase`，加入 ripple/pressed 状态说明区域和 Pressable 示例。
5. 梳理组件动画规范，先定义 hover、pressed、focus、selected、menu、toast、loading 的默认时长和缓动曲线，并参考 GitHub/Primer 的产品级交互质感。
6. 给 `internal/ripple.go` 拆出可测计算函数，补单元测试。
7. 检查所有 docs 示例是否仍优先使用 legacy Widget API，逐步迁移到 Element API。
8. 制定截图回归脚本或手动检查清单。

## 验收命令

每次推进 MD3 默认样式、组件绘制或交互反馈时，至少运行：

```sh
go test ./...
go vet ./...
```

涉及并发、hook、runtime、动画状态时追加：

```sh
go test -race ./...
```

涉及示例或文档展示时追加：

```sh
go test ./examples/material3_showcase
```

## 风险与约束

- MD3 默认化会改变旧应用视觉，应通过版本说明和兼容字段降低迁移成本。
- Ripple 动画依赖 clickable history 和 frame invalidation，必须避免无交互时持续触发重绘。
- hover、focus、loading 等动画必须可控，不能让普通业务界面持续重绘或产生明显延迟。
- 大圆角必须按实际尺寸 clamp，否则小组件会出现不正确的圆角。
- Element API 是推荐方向，但 Widget API 仍是底层兼容层，不能因文档迁移破坏旧入口。
- 不应一次性重写所有组件，优先抽共享 primitive，再分批接入组件。

## 长期完成定义

FluxUI 可以认为“全面推进 MD3 设计”达到稳定状态，需要满足：

- 默认 Light/Dark 主题完整、稳定、可定制。
- 常用组件默认视觉与 MD3 一致，并有清晰变体 API。
- 可交互组件共享 state layer、ripple、focus、disabled 规则。
- 文档和示例以 Element API + MD3 默认样式为主。
- showcase 覆盖主要组件和状态，可作为视觉回归入口。
- 旧 API 有清楚的兼容定位和迁移路径。
