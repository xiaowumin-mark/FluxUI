# Changelog

## Unreleased

### R1 高级表单 Beta

- 新增受控的 `SearchSelect`、`Combobox`、`Autocomplete`、`MultiSelect`、`TagPicker` 与 `TagInput`；选项焦点和多选交互以稳定 key 建模，异步建议仍完全由宿主提供。
- 新增精确十进制 `NumericField` / `SpinBox`：保留原始文本，使用精确有理数解析与步进，不以 `float64` 表达业务金额。
- 新增 `Form`、`FormField`、`ValidationSummary` 及受控 pending/error、可取消 submit、Ref 和 Docs Browser/示例入口。

### 修复

- Windows 原生/自定义标题栏的最大化按钮现在在窗口过程回调结束后投递系统最大化或还原命令，不再因回调重入保护而失效；pointer 输入路径也会执行同一动作。

### 兼容性与弃用

- 建立 [`docs/deprecation-ledger.md`](docs/deprecation-ledger.md) 作为弃用、兼容 shim 与迁移窗口的权威记录；本 CHANGELOG 仅提供发布摘要。
- `SelectSearchable`、`SelectQuick`、`SelectTypeaheadDelay` 标记为 Deprecated 的编译兼容 no-op。它们从未提供搜索、typeahead 或 quick-animation，不应在新代码或示例中使用；真实搜索/建议能力现由 R1 的 `SearchSelect`（仅选择既有项）、`Combobox`（可由宿主接受自由文本）和 `Autocomplete`（仅建议与匹配，不隐式提交业务查询）提供。
- `Widget`、`Run`、`App`、`Window`、`RunMulti` 仍为 Deprecated **兼容保留**入口：现有项目继续可编译，小版本不会静默移除。

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

### 交互状态（Phase 7）

- **`style.Decoration` 扩展**：新增 `Hover *Decoration`、`Pressed *Decoration`、`Focused *Decoration`、`Disabled *Decoration` 四个交互态字段
- **新增 With* 方法**：`WithHover(d)`、`WithPressed(d)`、`WithFocused(d)`、`WithDisabled(d)`
- **`ContainerDecoration` 集成交互**：内部使用 `event.Clickable` 追踪 `Hovered()` / `Pressed()` / `Clicked()` 状态，按 Pressed > Hover > 默认优先级自动切换装饰
- **新增事件回调**（ContainerDecorationOption）：`OnClick`、`OnHoverEnter`、`OnHoverLeave`、`OnHover`、`OnPressed`
- **`ContainerDecoration` 签名变更为 variadic**：`ContainerDecoration(d, child, opts...)` ← 向后兼容，无 opts 时行为不变
- **`ContainerDecorationElement` 同步支持**：Element 版本同样接受 variadic opts
- **新增 `ui` 包导出**：`ContainerDecorationOption`、`Hover(d)`、`Pressed(d)`、`Focused(d)`、`DisabledDeco(d)`、`HoverBg(c)`、`PressedBg(c)`、`OnDecoClick`、`OnDecoHoverEnter`、`OnDecoHoverLeave`、`OnDecoHover`、`OnDecoPressed`、`ContainerDecorationDisabled`
- **新增示例**：`interactive_decoration` 展示 Hover 背景变化、Pressed 反馈、Click 事件、卡片组合、Disabled 状态

### 背景图片（Phase 8）

- **新增 `style.ImageFill` 类型**：`{Src image.Image, Fit ImageFillFit}` — 背景图片 + 缩放模式
- **新增 `style.ImageFillFit` 枚举**：`ImageFillContain` / `ImageFillCover` / `ImageFillFill` / `ImageFillNone`
- **`style.Decoration` 扩展**：新增 `Image *ImageFill` 字段 + `WithImage` / `ResolveImage` 方法
- **新增图片加载工具**：`LoadImage(src)`（自动检测 URL / 文件路径）、`DecodeImageURL(url)`、`DecodeImageFile(path)`
- **`internal.SurfaceSpec` 扩展**：`HasImage` / `ImageOp` / `ImageFit` 字段
- **渲染集成**：`layoutRoundedSurface` / `layoutCircleSurface` 均支持图片背景，图片叠加在底色/渐变之上，透明区域露出 fallback 底色
- **交互兼容**：`Hover` / `Pressed` 态通过 `Merge` 继承 Image。圆形裁切 + 图片背景正常渲染
- **新增 `ui` 包导出**：`ImageFill`、`ImageFillFit`、`ImageBg(src, fit)`、`LoadImage`、`DecodeImageURL`、`DecodeImageFile`
- **新增示例**：`image_background` — Cover/Contain/Fill/None 四模式 + 圆形裁切 + Hover 换图 + 边框组合

### 废弃标记

- `Widget` 类型标记为 Deprecated：推荐使用 React-style `Element` API (`RunElement`)
- `Run` / `App` / `Window` / `RunMulti` 标记为 Deprecated：新项目使用 `RunElement`，旧入口继续作为兼容路径保留

### 2D 变换（Phase 9）

- **新增 `style.Transform2D` 类型**：`{RotateDeg, ScaleX, ScaleY, TranslateX, TranslateY, Origin TransformOrigin}` — 旋转/缩放/平移
- **新增 `style.TransformOrigin` 枚举**：`TransformCenter` / `TransformTopLeft` / `TransformTopRight` / `TransformBottomLeft` / `TransformBottomRight`
- **`style.Decoration` 扩展**：新增 `Transform *Transform2D` 字段 + `WithTransform` / `ResolveTransform` / `Merge`
- **`internal.SurfaceSpec` 扩展**：+`HasTransform` / `TransformMatrix`
- **渲染集成**：ContainerDecoration 在 `LayoutSurface` 前推入 `op.Affine` 栈，影响阴影/背景/图片/边框/子控件全部渲染。不影响占位空间（与 CSS `transform` 行为一致）
- **便捷构造器**：`Rotate(deg)` / `ScaleDeco(sx,sy)` / `TranslateDeco(tx,ty)` / `TransformDeco(...)`
- **交互兼容**：Hover/Press 态通过 Merge 继承 Transform，可实现 hover 放大等效果
- **新增示例**：`transform_demo` — 旋转/缩放/平移/原点/组合/圆形+旋转/交互 hover 放大

### 示例

- 新增：`react_workspace`、`router`、`popup_demo`、`team_workspace`、`virtual_scroll`、`fonts`、`horizontal_scroll`、`interactive_decoration`、`image_background`、`transform_demo`
- 移除冗余示例：`animation`、`react_counter`、`router_element`
