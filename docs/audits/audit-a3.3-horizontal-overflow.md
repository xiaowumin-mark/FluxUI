# A3.3 横向内容溢出审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 2：布局、渲染和样式稳定性。

- 状态：Done
- 日期：2026-07-06 18:47:56 +08:00
- 负责人：Codex
- 关注：Layout、Docs/Test
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `gopls go_search Tabs`
  - `gopls go_search Chip`
  - `rg -n "docs browser|DocsBrowser|code block|CodeBlock|table|Table|tabs|Tabs|chip|Chip|menu row|MenuRow|overflow|horizontal|ScrollHorizontal|ScrollView|ListView|GridView|Fill|Fixed" .`
  - `rg -n "renderMarkdown|code|table|ScrollHorizontal|ScrollViewElement|markdownTable|CodeBlock|pre|horizontal|Max|FillWidth|FixedWidth|Table" examples/docs_browser`
  - `rg -n "TabsScrollable|tabsRowWidget|fullWidth|scrollable|ScrollHorizontal|FixedWidth|Min|Max|chipWidget|newChip|Menu|DropdownMenu|menu row|Row\(" widget examples/docs_browser examples/component_lab -g "*.go"`
  - `rg -n "Horizontal|overflow|TableColumnWidths|ScrollHorizontal|TabsScrollable|Chip|Menu|finite|wide|width" widget examples/docs_browser examples/component_lab layout ui internal -g "*_test.go"`
  - `go test ./layout ./internal ./ui ./widget ./examples/docs_browser`
  - `git diff --check`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `docs/audits/audit-a3.1-basic-layout-constraints.md`
  - `docs/audits/audit-a3.2-scroll-container-layout.md`
  - `examples/docs_browser/markdown_renderer.go`
  - `examples/docs_browser/category_filter.go`
  - `examples/docs_browser/controls_overlay_demo.go`
  - `examples/docs_browser/controls_media_demo.go`
  - `examples/docs_browser/docs_metadata_test.go`
  - `widget/flex.go`
  - `widget/tabs_dialog_toast.go`
  - `widget/material3_components.go`
  - `widget/selection.go`
  - `internal/render.go`
- 关联能力：
  - docs browser 代码块横向滚动
  - docs browser 表格有限列宽
  - chips row 横向滚动
  - Tabs scrollable/fullWidth 宽度策略
  - Menu/Select popup 宽度 clamp
  - Row/Flex 非保护型横向布局边界

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| 当前分支 | `main`，相对 `origin/main` ahead 15 |
| `git status --short --branch --untracked-files=all` | 仅输出分支行，无脏文件清单 |
| 判断 | A3.3 执行前工作区干净；本任务只新增 `audit-a3.3-horizontal-overflow.md` 并更新索引，不修改 runtime/widget/layout 源码。 |

## 横向溢出策略总览

```text
普通 Row/Flex
  -> 不提供全局最大宽保护
  -> rigid 子项按自身内容测量
  -> Expanded/Flexed 才使用剩余主轴空间

docs browser code block
  -> Text(code)
  -> ScrollView(horizontal=true, vertical=false)
  -> 只让代码内容横向滚动，页面纵向滚动留给外层文档列表

docs browser table
  -> 每列宽度按文本估算
  -> clamp 到 88..280
  -> Row(cell FixedWidth...) 再包 horizontal ScrollView

docs browser category chips
  -> Row(chips...)
  -> horizontal ScrollView
  -> scrollbar hidden

Tabs
  -> non-scrollable: tab 按父宽等分
  -> scrollable: tab row 包 horizontal ScrollView

Menu/Select popup
  -> popup width 优先配置宽度或 trigger/field 宽度
  -> clamp 到父 Constraints.Max.X
  -> row 内 label 使用 Expanded/Flexed
```

## 横向内容矩阵

| 输入范围 | 当前策略 | 最大宽度/滚动边界 | 证据 | 结论 |
| --- | --- | --- | --- | --- |
| docs browser 代码块 | `markdownCodeBlock` 把 `Text(code)` 放进 `ScrollViewElement`，设置 `ScrollHorizontal(true)` 与 `ScrollVertical(false)`。 | 横向 content 由代码文本自然宽度决定，但 viewport 被 ScrollView 输出尺寸 clamp；纵向不在代码块内部滚动。 | `examples/docs_browser/markdown_renderer.go:181-199` | 合格。长代码不会撑宽文档页，而是在代码块内部横向滚动。 |
| docs browser 表格 | `markdownTable` 为每个 cell 使用 `FixedWidthElement(width)`，行用 `RowElement`，整体包横向 `ScrollViewElement`。 | `markdownTableCellWidth` 把每列宽度限制在 `88..280`；总表宽等于有限列宽累加，不继承 ScrollView 的 `1_000_000` 测量上限。 | `examples/docs_browser/markdown_renderer.go:203-265`、`examples/docs_browser/markdown_renderer.go:268-306` | 合格。表格可以横向滚动，但单列不会无限放大。 |
| docs browser 分类 chips | 分类过滤条使用 `RowElement(items...)` 加水平 ScrollView，并隐藏 scrollbar。 | chip 总宽可以大于视口，但只能在该条内部水平滚动。 | `examples/docs_browser/category_filter.go:59-73` | 合格。chips 不会撑宽左/右文档布局。 |
| Tabs | `TabsScrollable(true)` 时外层包横向 ScrollView；非 scrollable 时 `equalWidth` 按父 `Max.X / len(children)` 固定 tab 宽。 | scrollable tab 的最小宽可随 label 增长；非 scrollable tab 被父宽等分并 clamp。 | `widget/tabs_dialog_toast.go:301-307`、`widget/tabs_dialog_toast.go:312-357`、`widget/tabs_dialog_toast.go:520-545` | 合格但有边界风险：极长 label 在 scrollable Tabs 中会扩大横向 content，应由调用方控制 label 或接受横向滚动。 |
| Chip 单体 | Chip 使用 `middleRow` 组合 leading/label/trailing，没有内置最大宽或文本截断策略。 | 单个 chip 的宽度取决于 label 和槽位内容；它不是防溢出容器。 | `widget/material3_components.go:2145-2197` | 合格但依赖父级。chip row 必须使用水平 ScrollView、Wrap 类布局或外层固定宽。 |
| docs browser chip 示例 | docs browser 的分类 chips 已包水平 ScrollView；component/media demos 多在固定宽 panel 中展示 chip。 | 示例入口避免裸 chip row 撑开主文档。 | `examples/docs_browser/category_filter.go:68-73`、`examples/docs_browser/controls_media_demo.go` | 合格。手动验收入口覆盖 chips 但缺少专门溢出断言。 |
| Menu row | Menu root 无显式 width 时 `expandWidth`，有 `MenuWidth` 时固定宽；row 内 label 使用 `Expanded(Text(...))`，leading/trailing 固定宽。 | 静态 Menu 在有限父约束下填满父宽；DropdownMenu popup 宽度 clamp 到父 `Max.X`。 | `widget/material3_components.go:515-520`、`widget/material3_components.go:586-614`、`widget/material3_components.go:858-865` | 合格。Dropdown/Menu popup 不应因长 item label 突破 viewport；静态 Menu 不宜放进横向 ScrollView 后不设宽度。 |
| Select menu row | Select popup 宽度来自 field 或配置宽度，并 clamp 到父 `Max.X`；option row label 使用 Gio `Flexed(1)`。 | popup constraints 使用 `{X: popupW, Y: MaxHeightPx}`；正文区 `bodyMaxW` 至少 1，避免负宽。 | `widget/selection.go:611-630`、`widget/selection.go:693-713`、`widget/selection.go:863-917` | 合格。Select 下拉不会因长文本扩大到无限宽。 |
| 普通 Row/Flex | `Row` 只是 `internal.Context.LayoutFlex(Horizontal)` 的包装；底层使用 Gio Flex。 | Row 本身不提供最大宽度策略，rigid 子项会按父约束和自身测量累加；在横向 ScrollView 中父主轴上限是 `1_000_000`。 | `widget/flex.go:22-85`、`internal/render.go:1464-1489`、`docs/audits/audit-a3.2-scroll-container-layout.md` | 需要调用方负责。横向内容不能只依赖 Row，必须显式使用 ScrollView、FixedWidth、Expanded 或业务宽度上限。 |

## 最大宽度策略清单

| 场景 | 推荐策略 | 当前状态 |
| --- | --- | --- |
| 文档代码块 | 外层有限宽容器，代码文本内部水平滚动。 | 已实现。 |
| Markdown 表格 | 列宽有限上限，整体水平滚动。 | 已实现，列宽 clamp 到 `88..280`。 |
| 分类/筛选 chips | row 外层水平 ScrollView；必要时隐藏 scrollbar。 | docs browser 已实现。 |
| 可滚动 Tabs | `TabsScrollable(true)` 时允许横向滚动；非滚动 tabs 由父宽等分。 | 已实现。 |
| Menu/Dropdown/Select popup | 配置宽度或锚点宽度，最终 clamp 到 viewport/father `Max.X`。 | DropdownMenu 和 Select 已实现；静态 Menu 依赖父约束或 `MenuWidth`。 |
| 普通业务 Row | Row 不是溢出保护层；长内容需要 `FixedWidth`、`Expanded`、水平 ScrollView 或专用 wrap。 | 需要文档和示例继续提示。 |

## 现有测试覆盖

| 能力 | 测试 | 结论 |
| --- | --- | --- |
| Markdown 代码块渲染 | `TestRenderMarkdownDocumentBuildsCodeBlocks` | 覆盖代码块会被构造成渲染元素，但没有断言其外层是水平 ScrollView。 |
| Markdown 表格渲染 | `TestRenderMarkdownDocumentBuildsTables` | 覆盖表格识别和渲染元素生成。 |
| 表格列宽有限 | `TestMarkdownTableColumnWidthsStayFinite` | 明确断言列宽在 `88..280`，且总宽不会继承 `1_000_000` 滚动测量上限。 |
| docs browser 代表性 API | `TestDocsBrowserDemosUseRepresentativeReactAPIs` | 覆盖 `ScrollHorizontal`、`TabsScrollable`、`MenuWidth`、`DropdownMenuWidth`、Chip options 等示例入口存在。 |
| ScrollView 横向 wheel | `widget.TestHorizontalScrollViewRequiresHorizontalWheelDelta` 等 | A3.2 已覆盖横向 ScrollView 轴向和命中基础行为。 |
| Menu/Select 大列表 | `widget.TestLargeMenuAndSelectUseVirtualizedOptionLists` | 覆盖大 options 虚拟化，不直接覆盖长 label 横向溢出。 |

## 风险

| 风险 | 等级 | 说明 |
| --- | --- | --- |
| Chip 无内置最大宽 | 中 | 单个 Chip 和 chip row 不会自动截断或换行；只要调用方没有水平 ScrollView/固定宽/Wrap，长 label 仍可能撑大横向内容。docs browser 分类 chips 已处理。 |
| scrollable Tabs 极长 label 可形成超宽 content | 中 | `materialTabMinWidth` 会按 rune 数估算 label 宽度，未设单 tab 最大宽；这是横向滚动语义允许的内容宽，但业务若传入异常长 label 会扩大滚动范围。 |
| 静态 Menu 放入横向 ScrollView 时可能填满 `1_000_000` | 中 | `Menu` 未配置宽度时走 `expandWidth`。在普通有限父约束下安全；在横向 ScrollView 主轴测量中，父 `Max.X` 是大上限，可能得到过宽 menu。 |
| 代码块只处理横向滚动 | 低 | 代码块内部 `ScrollVertical(false)`，多行代码的纵向滚动交给外层文档页面；这符合 docs browser 页面模型，但嵌入固定高度区域时需要额外包纵向滚动。 |
| 缺少直接结构断言 | 低 | 现有测试覆盖表格列宽和示例 API 存在，但没有对 markdown code block/chips/tabs/menu 的布局树结构做专门断言。 |

## 事实结论

- docs browser 的代码块和表格都已经把横向内容隔离到内部水平 ScrollView，不会直接撑宽文档主列。
- Markdown 表格有明确列宽策略：按文本估算后 clamp 到 `88..280`，避免在 ScrollView 主轴 `1_000_000` 测量上限下无限扩张。
- docs browser 分类 chips 已使用水平 ScrollView，属于后续手动验收横向 chips 的入口。
- Tabs 的横向策略清晰：非 scrollable 等分父宽，scrollable 包横向 ScrollView；scrollable 模式接受内容宽超过视口。
- Chip 本身不是最大宽度组件；它的安全边界来自父容器。
- DropdownMenu 和 Select popup 宽度会 clamp 到父 `Constraints.Max.X`；Menu/Select row 的 label 使用 Expanded/Flexed，避免 leading、label、trailing 裸累加突破 popup 宽度。
- 普通 Row/Flex 不负责横向溢出保护；横向长内容必须由调用点显式选择水平滚动、固定宽、弹性宽或其他宽度策略。

## 验收

- 已覆盖 docs browser 代码块、表格、tabs、chips、menu row 的横向溢出行为。
- 已列出代码块、表格、chips、Tabs、Menu/Select 的最大宽度策略。
- 已明确不会因为表格单列或 menu popup 继承滚动测量上限而产生异常宽度。
- 已标注 Chip、scrollable Tabs、静态 Menu 在特殊父约束下仍需调用方控制的风险。
- 已记录现有测试覆盖与缺口，后续修复不会把既有横向行为误判为审查引入。
- `go test ./layout ./internal ./ui ./widget ./examples/docs_browser` 通过。
- `git diff --check` 通过；命令提示 `docs/audits/project-audit-baseline.md` 下次 Git touch 会 LF -> CRLF，这是行尾属性提示，不是 diff check 错误。

## 后续依赖

- A3.4 hit area 和 layout area 对齐审查需要验证水平 ScrollView 裁剪后，代码块、表格、chips、tabs 的命中区域是否与视觉区域一致。
- A4 样式审查需要确认代码块、表格、chips、tabs 的 hover/pressed/state layer 不改变布局宽度。
- A10.4 滚动集合控件族审查需要继续覆盖 Tabs、chips row、Menu/Select 与 ScrollView 的 wheel 和命中互不污染。
- A11.1 docs browser 组合场景审查需要把代码块、表格、分类 chips、TabsScrollable、MenuWidth/DropdownMenuWidth 纳入手动验收清单。
- 若后续修复横向溢出，优先补充结构/尺寸测试：代码块外层水平 ScrollView、chips row 不扩大文档主列、scrollable Tabs 极长 label 只扩大内部 scroll content、静态 Menu 在横向 ScrollView 中设置明确宽度。
