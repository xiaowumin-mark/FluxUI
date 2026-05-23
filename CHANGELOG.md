# Changelog

## v0.1.0 (2026-05-23)

### 新增

- React-style 声明式组件：`RunElement`、`Element`、函数组件、`Fragment`、`Key`、`Provider`
- 新增 Hooks：`UseMemo[T]`、`UseRef[T]`、`UseCallback[T]`
- 新增 Context API：`Provider[T]`、`UseContext[T]`
- 新增路由 Hooks：`UseParams`、`UseLocation`、`UseNavigate`
- 新增 Popup 组件（`PopupRef` 独立于 `DialogRef`）
- 新增 `OnHover` 事件支持

### 变更

- `RouteParams` 废弃，统一使用 `UseParams`
- `StateWithInitial` 移除，使用 `UseState(ctx, initial)`
- 字体便捷 API 精简，仅保留 `TextFont(spec)`、`TextFontWeight(weight)`
- `PopupAttachRef` 参数改为 `*PopupRef`（不再使用 `*DialogRef`）
- 移除未使用的 `event/keyboard`、`layout/constraints`、`SelectValue`、`InputValue` 等 API

### 样式系统重构

- **新增 `style.Border` 类型**：`{Width float32, Color color.NRGBA}`
- **`style.Decoration` 扩展**：新增 `Margin *Insets`、`Border *Border` 字段，含对应的 `WithMargin`、`WithBorder`、`ResolveMargin`、`ResolveBorder` 方法以及 `Merge` 覆盖逻辑
- **新增 Insets 构造器**：`Only(top,right,bottom,left)`、`Horizontal(v)` / `Horizontal(v)`（别名 `LeftRight` / `TopBottom`）
- **`style.Style` 标记为 Deprecated**：不再推荐使用，新代码请使用 `Decoration`
- **新增 `ContainerDecoration(d Decoration, child)`**：替代旧版 `Container(Style{}, child)`
- **新增 `ContainerDecorationElement(d Decoration, child)`**：React-style Element 版本
- **新增 `ui` 包全局构造函数**：`Margin(Insets)`、`BorderDeco(width, color)`、`Only(...)`、`LeftRight(v)`、`TopBottom(v)`
- **`internal.SurfaceSpec` 支持 Border**：新增 `BorderColor` 和 `BorderWidth` 渲染支持
- **Card 边框迁移**：移除嵌套 Container 技巧，改用 `Decoration.Border` 单层渲染
- **内部 widget 迁移**：Dialog / Toast / Popup / Select / AppBar / BottomNav / Tabs 全部从 `Container(Style{})` 切换为 `ContainerDecoration`

### 视觉特效

- **新增 `style.LinearGradient` 类型**：`{Start, End image.Point; From, To color.NRGBA}` 双色线性渐变
- **`style.Decoration` 扩展**：新增 `Gradient *LinearGradient`、`Opacity *float32`、`CircleClip bool`
- **新增 `ui` 包构造函数**：`Opacity(v)`、`LinearGrad(start,end,from,to)`、`Circle()`
- **`internal.SurfaceSpec` 扩展**：`Opacity`、`HasGradient`、`GradientStart/End`、`GradientFrom/To`、`CircleClip`
- **`LayoutSurface` 重构**：拆分为 `layoutRoundedSurface` / `layoutCircleSurface`，支持渐变填充 + 圆形裁切 + 透明度
- **示例新增**：Section 8（Opacity）、Section 9（LinearGradient）、Section 10（CircleClip）

### 盒阴影

- **新增 `style.BoxShadow` 类型**：`{OffsetX, OffsetY, Blur float32; Color color.NRGBA}` 自定义阴影
- **新增 `style.ElevationBoxShadow(level int)`**：Material Design 高度预设（1~5）
- **`style.Decoration` 扩展**：+`Shadow *BoxShadow`、`WithShadow`、`ResolveShadow`、`Merge`
- **`internal.SurfaceSpec` 扩展**：+`HasShadow`、`ShadowOffsetX/Y`、`ShadowBlur`、`ShadowColor`
- **`LayoutSurface` 阴影渲染**：`drawShadowLayers` 多层半透明偏移矩形模拟模糊（2~5 层）
- **新增 `ui` 包构造函数**：`Shadow(offX,offY,blur,color)`、`Elevation(level)`
- **`CardShadow` 复活**：通过 `Decoration.Shadow` 统一渲染路径，无需改调用方代码
- **示例新增**：Section 11（BoxShadow 五级 elevation、自定义偏移/颜色、渐变+阴影+圆形组合）

### 主题系统（Phase 6）

- **新增 `theme.ColorScheme` 类型**：17 个语义化色板令牌（Primary/OnPrimary、Secondary/OnSecondary、Surface/OnSurface、SurfaceMuted、Background/OnBackground、Error/OnError、Success/OnSuccess、Warning/OnWarning、Disabled、Outline）
- **新增 `LightColors()` / `DarkColors()`**：生产级浅/深色预设
- **`Theme` 扩展**：新增 `Colors ColorScheme` 字段，旧扁平字段（Primary、Surface、TextColor 等）保持向后兼容
- **新增 `theme.New(cs)`**：从 ColorScheme 创建 Theme，自动同步扁平兼容字段
- **新增 `ThemeProviderElement(th, child)`**：Element 子树局部主题覆盖，通过 `Context.WithTheme` 实现作用域主题
- **`Context` 扩展**：新增 `themeOverride` 字段，`Theme()` 优先返回局部覆盖主题
- **新增 `ui` 包导出**：`ColorScheme`、`NewTheme(cs)`、`LightColors()`、`DarkColors()`
- **示例重写**：`theme_custom` 从手动颜色传递升级为 ColorScheme + 语义色板展示 + 五主题动态切换

### 示例

- 新增：`react_workspace`、`router`、`popup_demo`、`team_workspace`、`virtual_scroll`、`fonts`、`horizontal_scroll`
- 移除冗余示例：`animation`、`react_counter`、`router_element`
