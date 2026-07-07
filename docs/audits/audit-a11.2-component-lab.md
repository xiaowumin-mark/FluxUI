# A11.2 component_lab 审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A11.2 component_lab 审查。

## 事实结论

审查范围覆盖 `examples/component_lab` 的主入口、单页 section 结构、主题控制、布局样式面板、基础交互、输入选择、媒体卡片、导航 overlay、集合滚动、拖放和 router 面板。`component_lab` 是 Docs/Test 与 Style 共同使用的视觉和交互 smoke 入口，适合在运行时人工观察样式、cursor、hover、pressed、disabled、overlay 和布局异常。

| 页面/区域 | 代码入口 | 覆盖内容 | smoke 价值 |
| --- | --- | --- | --- |
| 应用根 | `examples/component_lab/main.go` | `RunElement(App, Size(1180, 920), MinSize(900, 720))`；根部 `ThemeProviderElement` + `StackElement` | 固定大窗口 + 最小窗口可作为视觉基线；overlay 与主内容同根叠放 |
| 顶层列表 | `App` section switch | 10 个顶层 section，由 `ListViewElement(componentLabSectionCount, ..., ListVirtualized(true))` 承载 | 验证大页面滚动、顶层虚拟化和 section 间布局间距 |
| 主题控制 | `themeControls` | palette Select、Dark checkbox、Compact switch、token strip | 验证主题色、暗色、密度切换后全页面样式同步 |
| Layout and style primitives | `layoutStylePanel` | ContainerDecoration、Stack、Center、Transform、Row/Column/Spacer、Sizing、Divider、WithFont、legacy Container、nested theme | 样式、hover、pressed、transform、嵌套主题和旧 API 兼容的主要视觉入口 |
| Buttons and generic interaction | `buttonsPanel` | Filled/Tonal/Outlined/Text/Elevated/Disabled Button、Pressable、ClickArea、ButtonRef、PressableRef | hover、pressed、disabled、ripple、旧点击区域和 Ref 命令 smoke |
| Text input and selection | `inputSelectionPanel` | TextField、Checkbox、Switch、Slider、RangeSlider、RadioGroup、Select、DropdownMenu、Menu | focus、error、disabled、select/menu overlay、slider cursor、输入布局 smoke |
| Cards, media and chips | `mediaCardsPanel` | Image、Icon、Card variants、IconButton、FAB、Badge、Assist/Filter/Input/Suggestion chips、SearchBar | 卡片 surface、媒体尺寸、chip 行、badge、search 输入和 disabled/selected 样式入口 |
| Navigation and overlays | `navigationOverlayPanel` | AppBar、Tabs、scrollable Tabs、BottomNavigation、NavigationRail、NavigationDrawer、Dialog、Popup、Toast、Snackbar、Tooltip | overlay、z-order、tabs indicator、hover tooltip、导航布局和 toast/snackbar smoke |
| Progress, scroll, list and grid | `progressCollectionPanel` | Linear/Circular progress、loading Button/IconButton、ScrollView、ScrollRef、ListView、Grid、GridView | 进度动画、加载态、嵌套滚动、reach-end、网格尺寸 smoke |
| Drag, drop and router | `dragDropRouterPanel` | DragSource、DropTarget、drag preview、Drop log、RouterElement、route guard、transition | 拖放 active 样式、payload 日志、router 页面切换和动画 smoke |
| 自动 smoke | `perf_smoke_test.go`、`idle_redraw_test.go` | perf diagnostics、顶层虚拟化、文本缓存、输入统计；visual build tag 下空闲重绘 | component_lab 已承担性能与 idle redraw 的自动基线，但不替代视觉人工检查 |

### 样式、cursor、hover、布局异常清单

| 类别 | 观察入口 | 当前结论 | 风险/异常 |
| --- | --- | --- | --- |
| 样式主题 | palette Select、Dark checkbox、Compact switch | `componentLabTheme` 通过 seed 与 `CompactDensityScale` 重建主题，所有 section 走 `ThemeProviderElement` | 需手动确认 dark/compact 切换后没有文字颜色、padding 或高度异常 |
| 样式状态 | Button variants、Input error/disabled、Select error/disabled、chips selected/soft-disabled、IconButton selected/loading | 页面覆盖 selected、disabled、error、loading、pressed、hover 等状态 | 这些状态分散在多个控件，当前没有视觉截图 golden |
| ContainerDecoration hover | Layout 面板 `hover and click decoration` | 使用 `WithHover`、`WithPressed`、`OnDecoHover`、`OnDecoClick`，底部文字显示 hovering/clicks | 可观察 hover/pressed 是否只影响本容器，不应造成整行变色或 layout 抖动 |
| Button hover | Buttons 面板 Filled 按钮 | 绑定 `OnHover` 并显示 `hover=true/false` | 可观察离开按钮后 hover 是否清理；不应停留为 true |
| Pressable pressed | Buttons 面板 Pressable | `OnDecoPressed` 写入 `pressed` 状态，PressableRef 可程序点击 | 可观察按下态和程序点击是否互相污染 |
| cursor | Input/Selection 面板 Slider 与 RangeSlider | A4.4 已确认显式 pointer cursor 来源主要是 Slider，且应 clip 到 track/state-layer | component_lab 是手动观察入口；当前没有 component_lab 级 cursor 自动断言 |
| disabled cursor | Disabled Button、disabled TextField、disabled Select、disabled menu item、disabled IconButton、soft-disabled chips | 这些组件用于观察 disabled 视觉和交互禁用 | A4.4 显示普通交互控件不写 cursor；后续若新增 cursor API 需重跑本页 |
| 布局固定宽 | 多处 `FixedWidthElement`、`FixedHeightElement`、`ExpandedElement` 组合 | 页面在 1180x920 下覆盖多列布局；最小窗口为 900x720 | 多个 Row 内固定宽组合较多，900px 最小宽度仍需人工确认是否出现裁切或内容互相挤压 |
| 顶层滚动 | 根 section `ListViewElement(..., ListVirtualized(true))` | 顶层 section 虚拟化已有测试 `TestComponentLabAppVirtualizesTopLevelSections` | 视觉上需确认滚动到后段时 overlay 和 nested ListView/GridView 仍可用 |
| 嵌套滚动 | `ScrollView`、`ListView`、`GridView` 面板 | 覆盖 ScrollRef Top/End、List reach-end、Grid reach-end | 父页面和子滚动区域的 wheel 传递仍依赖 A6.2/A6.4 的事件边界 |
| overlay | Dialog/Popup/Toast/Snackbar/Tooltip | overlay 放在根 `StackElement`，Dialog/Popup 可由按钮和 Ref 打开 | 需观察 overlay 不改变普通布局，关闭后 cursor/hover/focus 不残留 |
| transform | Layout 面板 Rotate/Scale/Translate | transform 与正常 Row 并列展示 | 需观察 transform 不造成周边内容不可点击或错位 |
| 资源 | Image 使用 `examples/assets/sample.png` | 图片区域覆盖 contain/cover、radius、border、click | 若资产缺失或路径相对运行目录变化，会影响 media smoke |

### 手动 smoke 表

| 编号 | 场景 | 操作步骤 | 预期结果 | 覆盖点 |
| --- | --- | --- | --- | --- |
| CL-01 | 启动 | 运行 `go run ./examples/component_lab` | 窗口标题为 FluxUI Component Lab；首屏 header、主题控制和前几个 section 正常显示 | app 入口、主题、初始布局 |
| CL-02 | 主题切换 | 依次切换 Blue/Green/Amber、Dark mode、Compact | 全页面颜色和密度同步变化；文字保持可读；section 边框、surface、token strip 不错色 | style/theme |
| CL-03 | 顶层滚动 | 从顶部滚到 Drag/drop and router，再滚回顶部 | section 间距稳定；虚拟化不出现空白、重复或错乱内容 | ListView 顶层虚拟化 |
| CL-04 | ContainerDecoration hover | 悬停并点击 Layout 面板的 hover/click decoration | 容器本身变色；计数增加；离开后 `hovering=false`；周边 Row 不被整行染色 | hover/pressed/命中区域 |
| CL-05 | Transform | 观察 Rotate、Scale、Translate 区域并点击周边控件 | transform 只影响视觉；不挤压或覆盖相邻内容的正常交互 | style transform |
| CL-06 | Button hover/click | 悬停 Filled 按钮、点击各 button、点击 disabled button | hover 状态正确清理；点击计数只在可用按钮增加；disabled 不触发 | button state |
| CL-07 | Pressable/ClickArea | 按下 Pressable、点击 ClickArea、触发 PressableRef.Click | pressed 状态按下释放可见；程序点击只增加计数，不制造持续 pressed | generic interaction |
| CL-08 | 输入焦点 | 聚焦 Username，点击 Set text、Focus、Clear | focus 文本状态和输入值同步；布局不跳动；counter/supporting text 保持对齐 | TextField/focus |
| CL-09 | Slider cursor | 在 Slider 与 RangeSlider track 内外移动鼠标并拖动 | track 附近显示 pointer，离开后恢复默认；拖动不截停父页面无关纵向滚动 | cursor/slider |
| CL-10 | Select/Menu | 打开 Select、DropdownMenu 和 Menu 子菜单 | popup/menu 宽高受控；disabled 项不可选；关闭后 hover/cursor 不残留 | overlay/menu |
| CL-11 | Cards/media/chips | 点击图片、卡片、IconButton、FAB、chips remove/restore、SearchBar 输入 | 计数/selected/visible 状态正确；chips 行不撑爆 section | media/chips/search |
| CL-12 | Tabs/navigation | 切换 full-width tabs、scrolling tabs、secondary tabs、BottomNav/Rail/Drawer | indicator 和 selected 状态稳定；scrolling tabs 可横向查看；导航区域不影响其它 section | navigation/layout |
| CL-13 | Dialog/Popup | 点击 Dialog、Open popup、关闭按钮或遮罩/动作按钮 | overlay 在根层显示；关闭后主页面布局不变化；Ref 与状态保持一致 | overlay/z-order |
| CL-14 | Toast/Snackbar/Tooltip | 触发 Toast、Show snackbar、悬停 Tooltip 按钮 | toast/snackbar 叠加在页面上，不推动布局；tooltip 只随 hover 显示 | transient overlay |
| CL-15 | Scroll/List/Grid | 点击 ScrollRef Top/End，滚动内嵌 ScrollView/ListView/GridView | scroll log、reach-end 计数更新；父页面和子区域滚动边界清楚 | collection/wheel |
| CL-16 | Drag/drop/router | 拖拽 Drag text 到 Drop here；切换 router Detail/Settings/404/Back | drop active 样式和日志更新；router guard/transition 可见；页面不产生布局闪烁 | drag/drop/router |
| CL-17 | 最小尺寸 | 将窗口缩到接近 900x720，重复 CL-02、CL-06、CL-13 | 内容仍可通过滚动访问；关键按钮文字不重叠；overlay 仍居中或可见 | responsive smoke |

## 风险

- `component_lab` 覆盖面很大，但目前没有截图 golden 或自动视觉 diff；样式退化仍主要依赖人工观察。
- 顶层 section 是虚拟化 ListView，自动测试验证了可裁剪 offscreen section；但这也意味着某些 section 只有滚动到可见时才布局，手动 smoke 必须覆盖全页。
- 多个 Row 使用固定宽和横向并列布局，在接近 900px 最小宽度时可能出现视觉挤压；当前审查只记录入口，不调整布局。
- cursor 主要通过 Slider 观察；A4.4 已确认 Slider track clip，但 component_lab 自身还没有自动 cursor smoke。
- Tooltip、Select、DropdownMenu、Dialog、Popup、Toast、Snackbar 同页共存，适合暴露 overlay 互扰；但本轮未执行真实 GUI 点击，只建立验收基线。
- Drag/drop 依赖系统和 Gio 后端能力，某些环境可能无法完成真实外部 payload；可先用内部 DragSource 到 DropTarget 的路径做 smoke。
- 图片资源依赖 `examples/assets/sample.png` 和运行目录；若从非仓库根目录运行，media smoke 可能误报资源问题。

## 验收

- 已建立 component_lab 主要页面清单，覆盖主题、样式、cursor、hover、布局、overlay、集合、拖放和 router。
- 已输出样式、cursor、hover、布局异常清单，明确当前已知风险与观察入口。
- 已建立 CL-01 到 CL-17 手动 smoke 表，可作为视觉和交互回归入口。
- 已确认 `component_lab` 已有自动基线：`TestPerfScenario`、`TestComponentLabAppVirtualizesTopLevelSections` 和 visual build tag 下的 idle redraw 测试。
- 已确认本轮只记录审计事实与验收入口，不修改 runtime/widget/style 示例实现。

## 后续依赖

- A4.1/A4.2/A4.3/A4.4：样式合并、状态来源、ripple/state layer、cursor 策略变更后，需要回归 CL-02、CL-04、CL-06、CL-09、CL-11。
- A4.5：动画和布局隔离变更后，需要回归 progress、tabs indicator、Dialog/Popup、Toast/Snackbar 和 router transition。
- A6.2/A6.4：wheel 父子滚动和滚动后命中刷新变更后，需要回归 CL-03、CL-10、CL-15。
- A8.3/A8.4：outside click、focus、Escape 规则变更后，需要回归 CL-10、CL-13、CL-14。
- A10.x：任一控件族修复后，应在 component_lab 对应 section 执行 smoke，确认视觉状态与交互边界未回退。
- A12.x：建议将 CL-01、CL-02、CL-04、CL-09、CL-13、CL-15 中稳定可复现的步骤转为自动视觉或事件 smoke。
