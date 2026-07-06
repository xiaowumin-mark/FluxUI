# A6.3 横向滚动策略审查
> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 3：滚动、手势和嵌套交互审查。

- 状态：Done
- 日期：2026-07-06
- 负责人：Codex
- 关注：Layout、Event
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "### A6\\.|A6\\.3|横向滚动|horizontal scroll|shift-wheel|touchpad|tabs|chips|table|code block" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md`
  - `rg -n "type ScrollView|func .*ScrollView|AxisHorizontal|Horizontal|ScrollX|ScrollY|Shift|Modifiers|WheelEvent|processWheelEvents|applyWheelDefault" event internal widget ui examples docs -g "*.go"`
  - `rg -n "Tabs|Tab|Chip|chips|table|code block|CodeBlock|Markdown|ScrollView|Horizontal" examples docs widget ui -g "*.go"`
  - `rg -n "Shift|ModShift|shift|Scroll:|DeltaX|DeltaY|ScrollHorizontal|horizontal wheel|vertical wheel" -g "*_test.go" widget event examples/docs_browser`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `event/input.go`
  - `widget/list_grid.go`
  - `widget/list_grid_test.go`
  - `widget/tabs_dialog_toast.go`
  - `examples/docs_browser/category_filter.go`
  - `examples/docs_browser/markdown_renderer.go`
  - `examples/docs_browser/controls_overlay_demo.go`
- 关联能力：
  - 横向 `ScrollView` 的 `ScrollX` 过滤策略
  - 纵向 wheel 不触发横向滚动的边界
  - touchpad 横向 delta 入口
  - shift-wheel 未定义策略识别
  - docs browser tabs、chips、table、code block 横向滚动入口

## 事实结论

1. A6.3 的目标要求横向滚动必须由横向 delta 或明确的 shift-wheel 策略触发，不能误用纵向滚动；docs browser 的代码块、表格、tabs、chips 不应阻挡页面纵向滚动。证据：`docs/project-audit-roadmap.md:206`、`docs/project-audit-task-breakdown.md:320`。
2. `WheelEventFromGio` 直接把 Gio `pointer.Event.Scroll.X` 映射为 `DeltaX`，把 `Scroll.Y` 映射为 `DeltaY`，并保留 `Modifiers`；事件本身没有把 `Shift+DeltaY` 转换为 `DeltaX` 的逻辑。证据：`event/input.go:249`、`event/input.go:341`。
3. `ScrollView` 默认纵向，只有设置 `ScrollHorizontal(true)` 且通常配合 `ScrollVertical(false)` 后才成为横向容器。证据：`widget/list_grid.go:56`、`widget/list_grid.go:77`。
4. 横向 `ScrollView` 的 Gio filter 只开放 `ScrollX`，并把 `ScrollY` 范围锁为 0；纵向 `ScrollView` 则反向处理。证据：`widget/list_grid.go:295`。
5. 横向默认滚动只读取 `wheel.DeltaX`；普通纵向 wheel 产生的 `DeltaY` 不会被横向 `ScrollView` 当作 offset 使用。证据：`widget/list_grid.go:321`。
6. 当前没有 shift-wheel 策略实现：虽然 `WheelEvent.Modifiers.Shift` 会被记录，但 `processWheelEvents` 和 `applyWheelDefault` 没有检查 `Shift`，也没有将 `Shift+DeltaY` 转为横向 delta。证据：`event/input.go:163`、`widget/list_grid.go:295`、`widget/list_grid.go:321`。
7. touchpad 或支持横向滚动的输入设备只要后端提供非零 `Scroll.X`，就会进入横向 `ScrollView` 的 `ScrollX` filter，并作为 `DeltaX` 改变横向 offset；这是当前唯一明确的横向 wheel 策略。证据：`event/input.go:259`、`widget/list_grid.go:297`、`widget/list_grid_test.go:279`。
8. 已有自动测试覆盖普通纵向 wheel 不会滚动横向 `ScrollView`，同时横向 delta 会滚动横向 `ScrollView`。证据：`widget/list_grid_test.go:224`。
9. 已有自动测试覆盖纵向 wheel 落在横向内层 `ScrollView` 上时，外层纵向 `ScrollView` 会滚动，内层横向 `ScrollView` 不滚动。证据：`widget/list_grid_test.go:343`。
10. `TabsScrollable(true)` 会把 tab row 包进 `ScrollView(row, ScrollHorizontal(true), ScrollVertical(false), ScrollBarVisible(false))`；因此 tabs 的横向滚动策略与普通横向 `ScrollView` 一致。证据：`widget/tabs_dialog_toast.go:295`。
11. docs browser 的分类 chips row 使用 `ScrollViewElement(RowElement(...), ScrollHorizontal(true), ScrollVertical(false), ScrollBarVisible(false))`；普通纵向 wheel 不应被该 chips row 横向消费。证据：`examples/docs_browser/category_filter.go:68`。
12. Markdown 代码块内部使用横向 `ScrollViewElement`，设置 `ScrollHorizontal(true)` 和 `ScrollVertical(false)`；长代码依赖横向 delta 或滚动条，不支持纵向 wheel 横向滚动。证据：`examples/docs_browser/markdown_renderer.go:193`。
13. Markdown 表格外层也使用横向 `ScrollViewElement`，设置 `ScrollHorizontal(true)` 和 `ScrollVertical(false)`；列宽由 `markdownTableColumnWidths` 限制，避免无限宽直接撑破文档页。证据：`examples/docs_browser/markdown_renderer.go:260`、`examples/docs_browser/markdown_renderer.go:268`。
14. docs browser tabs 示例包含 `TabsScrollable(true)`，并固定在 520dp 宽内容区域内，适合作为横向 tabs 手动验收入口。证据：`examples/docs_browser/controls_overlay_demo.go:103`、`examples/docs_browser/controls_overlay_demo.go:118`。
15. 测试检索未发现 `Shift+Wheel` 或 `ModShift` 与 `ScrollHorizontal` 组合的专项测试；现有 modifier 测试只验证 pointer/wheel event 的 modifier 映射。证据：`widget/pointer_area_test.go:81`、`widget/pointer_area_test.go:127`。

### 横向滚动策略矩阵

| 场景 | 横向 delta | shift-wheel | touchpad | 纵向 wheel | 结论 |
| --- | --- | --- | --- | --- | --- |
| 横向 `ScrollView` | 支持，读取 `DeltaX` | 未定义/未实现 | 支持后端提供的 `Scroll.X` | 不消费，不转横向 | 符合“不误触发横滚” |
| `TabsScrollable(true)` | 支持，经内部横向 `ScrollView` | 未定义/未实现 | 支持 `Scroll.X` | 不消费，外层可继续纵向滚动 | 与 ScrollView 一致 |
| docs 分类 chips | 支持，经横向 `ScrollViewElement` | 未定义/未实现 | 支持 `Scroll.X` | 不消费，外层可继续纵向滚动 | 与 ScrollView 一致 |
| Markdown code block | 支持，经横向 `ScrollViewElement` | 未定义/未实现 | 支持 `Scroll.X` | 不消费，外层文档滚动应继续 | 手动验收入口 |
| Markdown table | 支持，经横向 `ScrollViewElement` | 未定义/未实现 | 支持 `Scroll.X` | 不消费，外层文档滚动应继续 | 手动验收入口 |

## 风险

| 风险 | 等级 | 说明 | 后续关联 |
| --- | --- | --- | --- |
| shift-wheel 体验未定义 | 中 | 当前不会把 `Shift+DeltaY` 转为横向滚动。若用户期望桌面常见的 shift-wheel 横滚，需要明确产品策略并新增实现和测试；在此之前不应在文档中承诺支持。 | A11 文档、A13 收敛修复 |
| touchpad 双轴 delta 的主次轴策略较粗 | 中 | 当前横向容器只消费 `DeltaX`，纵向容器只消费 `DeltaY`；没有“主轴优先”“阈值”“剩余 delta”模型。双轴手势在嵌套场景下依赖 Gio 路由和轴向 filter，体验需要手动验收。 | A6.2、A13 嵌套滚动 |
| 横向内容触边后剩余 delta 未建模 | 中 | 横向 `ScrollView` 到左/右边界后，不记录已消费量和剩余 `DeltaX`，同轴父级传递或边界回退行为没有明确模型。 | A6.2、A13 |
| docs browser 主滚动是 `ListView` | 中 | 右侧文档页的纵向主体不走 `ScrollView` wheel default action；横向代码块/表格与文档纵向滚动的实际组合还需要 smoke 手动确认。 | A11 docs browser 验收 |
| tabs/chips 没有专项横向 wheel 测试 | 低 | 它们复用 `ScrollView` 的已测策略，但没有组件级测试直接覆盖 tabs/chips 在文档页内的纵向 wheel 透传。 | A12 测试补齐 |

## 验收

- 已明确横向 delta 策略：当前只认后端提供的 `Scroll.X`/`DeltaX`，并由横向 `ScrollView` 改变横向 offset。
- 已明确 shift-wheel 策略：当前未定义、未实现、未测试；不能作为已支持能力记录。
- 已明确 touchpad 策略：只要后端提供横向 delta 即支持；双轴手势没有额外主轴仲裁或剩余 delta 模型。
- 已明确纵向 wheel 策略：普通 `DeltaY` 不会触发横向滚动，已有自动测试证明横向 `ScrollView` 不消费纵向 wheel。
- 已覆盖输入组件：tabs、chips、Markdown table、Markdown code block 都通过横向 `ScrollView` 或其 Element 包装实现横向滚动。
- 已标出手动验收入口：docs browser 的 `controls_overlay` tabs 示例、顶部分类 chips、文档代码块和 Markdown 表格。

## 后续依赖

- A11 docs browser 文档审查应把“横向滚动依赖横向 delta/触摸板/滚动条，暂不承诺 shift-wheel”写清楚。
- A12 测试审查应补齐 tabs/chips/code block/table 在嵌套文档页中的 smoke 或组件级 wheel 测试。
- 若 A13 决定支持 shift-wheel，需要在 `ScrollView` 里定义转换规则、优先级、cancelable 行为和测试矩阵。
- 若后续实现 nested scroll chaining，需要同时处理横向触边后的剩余 `DeltaX` 和双轴 touchpad 的剩余 delta。
