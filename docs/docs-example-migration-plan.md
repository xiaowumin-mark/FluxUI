# Docs 默认示例迁移策略

## 目标

制定 docs 默认示例从 legacy `Widget` 写法迁移到 React-style `Element` 写法的分批计划。本文只规划顺序和风险边界，不直接大规模重写 examples。

## 原则

- 新文档默认示例逐步推荐 `RunElement` / `Element` / `Component`。
- legacy 示例继续保留为兼容说明或对照路径。
- 每批迁移都必须保持 `go test ./...` 通过。
- docs_browser 的旧示例渲染能力不能因为文档迁移被打断。
- 涉及 refs、command queue、overlay、scroll、virtual list 的示例后置。

## Canonical React-style 示例

- `examples/react_workspace` 是当前唯一完整 React-style showcase，覆盖 `RunElement`、`UseState`、hooks、context、router、keyed identity 和 `FromWidget`。
- `examples/counter` 继续保留为 compact legacy `Run` / `Widget` smoke demo。
- 新文档示例优先参考 `RunElement`、`UseState`、`UseMemo`、`UseRef`、`TextElement`、`ButtonElement`、基础 layout Element wrappers 的组合方式。

## 分批计划

| 批次 | 范围 | 候选 |
| --- | --- | --- |
| Batch 1 | 无状态展示和基础布局 | `text_basic`、`spacer_basic`、`divider_basic`、`row_basic`、`column_basic`、`stack_basic`、`center_basic`、`padding_basic`、`container_basic`、`sizing_basic` |
| Batch 2 | 简单交互 | `button_basic`、`click_area_basic`、`checkbox_basic`、`switch_basic` |
| Batch 3 | Router 新写法 | `router_basic`、`examples/react_workspace` React-style 对照版本 |
| Batch 4 | 输入和表单 | `textfield_basic`、`select_basic`、`radio_group_basic`、`slider_basic`、`form_validation` |
| Batch 5 | 滚动、列表、网格、弹层、通知 | `scroll_view_basic`、`list_view_basic`、`grid_basic`、`dialog_basic`、`popup_basic`、`toast_basic`、`virtual_scroll` |
| Batch 6 | 综合 showcase | `docs_browser`、`team_workspace`、`advanced_components`、`animation_showcase`、`vscode_layout` |

## 第一批候选

- `docs/widgets/text.md`
- `docs/widgets/spacer.md`
- `docs/widgets/divider.md`
- `docs/widgets/row.md`
- `docs/widgets/column.md`
- `docs/widgets/stack.md`
- `docs/widgets/center.md`
- `docs/widgets/padding.md`
- `docs/widgets/container.md`
- `docs/widgets/sizing.md`

### Batch 1 进度

- 已为 Batch 1 Markdown 文档补充 React-style `Element` snippet。
- 每个文档继续保留 legacy `Widget` 示例作为兼容对照。
- 本批次未修改 docs_browser runtime 示例映射。
- `sizing` 中 `Fixed` / `Fill` / `Expanded` 的原生 Element wrapper 尚未冻结，文档先展示 `FromWidget` 桥接路径和基础 Element wrapper 替代表达。

### Batch 3 进度

- `examples/react_workspace` 覆盖 `RouterElement`、`RouteElement`、`UseNavigate`、`UseLocation`、`UseParams`。
- legacy `examples/router` 保持不变，继续作为旧 Router API 对照。
- 本步骤未修改 docs_browser runtime 示例映射。

### D1.4 兼容性确认

- Batch 1 文档仍然使用 legacy example id：`text_basic`、`spacer_basic`、`divider_basic`、`row_basic`、`column_basic`、`stack_basic`、`center_basic`、`padding_basic`、`container_basic`、`sizing_basic`。
- Router 文档仍然使用 legacy example id：`router_basic`。
- `examples/docs_browser` 继续解析这些 legacy example id，没有新增 runtime 映射变更。
- React-style 对照能力集中在 `examples/react_workspace`，不接入 docs_browser 旧示例映射。

### Batch 2 准备

- Batch 2 只覆盖简单交互文档：`button.md`、`click_area.md`、`checkbox.md`、`switch.md`。
- `examples/react_workspace` 作为 `Button + UseState` 的 canonical 参考，不再新增单独 counter 示例。
- `button` 和 `click_area` 先写 React-style 对照示例，再补 metadata 中的 Element API 条目。
- `checkbox` 和 `switch` 示例用 `UseState` 展示受控布尔状态，保持 legacy 示例与文档 id 不变。
- 本批次仍然不修改 `examples/docs_browser` runtime 映射。
- 执行顺序建议：`button` / `click_area` 先行，`checkbox` / `switch` 后置。

### Batch 2 进度

- `button.md` 和 `click_area.md` 已补充 React-style 对照示例。
- 两份文档的 metadata 已补齐 `ButtonElement` / `ClickAreaElement` 条目。
- legacy example id 仍保持 `button_basic` 与 `click_area_basic` 不变。
- 本次更新未修改 `examples/docs_browser` runtime 映射。

### Batch 2 进行中

- `checkbox.md` 和 `switch.md` 仍待补充 React-style 对照示例。
- 这两份文档的 legacy example id 继续保持不变：`checkbox_basic`、`switch_basic`。

### Batch 2 兼容性确认

- `examples/docs_browser` 仍然只解析旧 example id：`button_basic`、`click_area_basic`、`checkbox_basic`、`switch_basic`。
- Batch 2 文档新增的 React-style 示例不影响 docs_browser 的现有映射。
- `go test ./...` 继续通过，说明 docs 变更未破坏仓库级验证。

### Batch 3 准备

- Batch 3 只覆盖 Router 文档：`docs/widgets/router.md`。
- `examples/react_workspace` 作为 Router React-style 对照的 canonical example，`examples/router` 继续保留为 legacy 对照。
- 本批次仍然不修改 `examples/docs_browser` runtime 映射，`router_basic` 继续作为文档浏览器的旧入口。
- 推荐顺序：先补 Router 文档的 React-style API 说明，再整理示例段落和 identity 说明，最后做兼容性检查。

### Batch 3 进度

- `router.md` 已补充 React-style API metadata：`RouterElement`、`RouteElement`、`RouteKey`、`UseNavigate`、`UseLocation`、`UseParams`。
- `router.md` 已补充 React-style 对照示例，并说明默认 route identity 按 pattern 复用、`RouteKey` 可强制 remount。
- `examples/router/main.go` 仍然作为 legacy 对照，`examples/react_workspace/main.go` 作为 React-style 对照示例。
- 本次更新未修改 `examples/docs_browser` runtime 映射。

### Batch 3 兼容性确认

- `examples/docs_browser` 仍然只解析 `router_basic`。
- Router 文档的 legacy 基础用法和导航示例保持不变。
- `go test ./...` 继续通过，route identity 说明与现有测试行为一致。

### Batch 3 完成

- Router 文档迁移已完成，legacy / React-style 对照路径均已记录。
- React-style Router 对照路径已集中到 `examples/react_workspace`。
- `examples/router` 和 `router_basic` 保持不变，docs_browser 行为未被改动。

### Batch 4 准备

- Batch 4 只覆盖复杂输入与表单相关文档：`docs/widgets/textfield.md`、`docs/widgets/select.md`、`docs/widgets/radio_group.md`、`docs/widgets/slider.md`。
- `examples/form_validation` 是这一批的整合示例参考，但它目前仍是 legacy `Run` / `Widget` 组合，不应在未完成 host-state 设计前强行改写。
- 目前仓库中尚未冻结这些控件的 Element wrapper（例如 `TextFieldElement`、`SelectElement`、`RadioGroupElement`、`SliderElement`），因此 Batch 4 需要先做 host-state / ref 边界确认，再决定文档里的 React-style 写法。
- 本批次仍然不修改 `examples/docs_browser` runtime 映射；`textfield_basic`、`select_basic`、`radio_group_basic`、`slider_basic` 继续作为旧入口。
- 推荐顺序：先做 host-state / API 命名评审，再处理 `textfield` 和 `slider`，然后处理 `select` 和 `radio_group`，最后给 `examples/form_validation` 补兼容性说明。

### Batch 4 进行中

- `textfield.md` 和 `slider.md` 已补充 host-state / ref strategy notes，继续保持 legacy-first。
- `examples/form_validation` 仍然仅作 legacy 对照参考。
- Batch 4 目前不冻结 `TextFieldElement` / `SliderElement` 等 API 名称。
- `select.md` 和 `radio_group.md` 已补充 host-state / ref strategy notes，继续保持 legacy-first。
- Batch 4 目前不冻结 `SelectElement` / `RadioGroupElement` 等 API 名称。
- `examples/form_validation/README.md` 已补充兼容性说明，明确该示例暂不迁移到 `RunElement`。

### Batch 4 兼容性确认

- `examples/docs_browser` 仍然解析旧 example id：`textfield_basic`、`slider_basic`、`radio_group_basic`、`select_basic`。
- Batch 4 没有新增复杂输入的 Element API metadata，也没有更改 docs_browser runtime 映射。
- `examples/form_validation/main.go` 未修改，新增 README 只用于说明兼容策略。
- `go test ./...` 继续通过，说明 Batch 4 文档和示例说明变更没有破坏仓库级验证。

### Batch 4 完成

- Batch 4 以策略说明和兼容性记录收尾，不做复杂输入 React-style snippet 迁移。
- 后续若要迁移复杂输入，应进入 host-state component migration 工作流，而不是继续仅靠 docs snippet 迁移。

### Batch 5 准备

- Batch 5 覆盖滚动、列表、网格、弹层和通知相关文档：`docs/widgets/scroll_view.md`、`docs/widgets/list_view.md`、`docs/widgets/grid.md`、`docs/widgets/dialog.md`、`docs/widgets/popup.md`、`docs/widgets/toast.md`，以及 `examples/virtual_scroll` 兼容参考。
- 这一批比 Batch 4 更依赖 host-state、虚拟列表和 overlay 生命周期，建议继续保持 legacy-first，先写策略说明，再考虑是否补 React-style 对照写法。
- 目前仓库中尚未冻结这些控件的 Element wrapper（例如 `ScrollViewElement`、`ListViewElement`、`GridElement`、`DialogElement`、`PopupElement`、`ToastElement`），因此 Batch 5 需要先做 lifecycle / host-state 边界确认。
- 本批次仍然不修改 `examples/docs_browser` runtime 映射；`scroll_view_basic`、`list_view_basic`、`grid_basic`、`dialog_basic`、`popup_basic`、`toast_basic` 继续作为旧入口。
- `examples/virtual_scroll` 继续作为 integration/compatibility 参考，不在这个阶段强行迁移为 `RunElement`。
- 推荐顺序：先做 scroll/list lifecycle 评审，再处理 grid，然后处理 dialog/popup，再处理 toast 与 virtual_scroll 兼容说明，最后做兼容性检查。

### Batch 5 D5.1 lifecycle prep

- D5.1 已锁定 Batch 5 范围：`scroll_view`、`list_view`、`grid`、`dialog`、`popup`、`toast`，以及 `examples/virtual_scroll` 的兼容说明。
- Batch 5 继续采用 legacy-first 策略：先补 lifecycle / host-state 策略说明，不在本批次冻结 `ScrollViewElement`、`ListViewElement`、`GridElement`、`DialogElement`、`PopupElement`、`ToastElement` 等 API 名称。
- `ScrollView` / `ListView` 优先处理，因为它们涉及 `ScrollRef` 命令队列、滚动偏移、auto-to-end key、虚拟窗口、触底回调和 item identity。
- `Grid` 单独处理，区分固定 children 的 `Grid` 与动态/虚拟化的 `GridView`，避免提前承诺 keyed item 或 host-state 复用规则。
- `Dialog` / `Popup` 后置处理，重点记录受控 open、mask close、imperative ref、overlay 挂载/卸载和可能的 stacking/focus 边界。
- `Toast` 与 `examples/virtual_scroll` 最后处理，分别记录短时通知 timer/auto-close 生命周期和虚拟滚动示例继续 legacy 对照的原因。

### Batch 5 进行中

- D5.1 已完成 scope lock 和 lifecycle prep 记录。
- D5.2 已完成 `docs/widgets/scroll_view.md` 和 `docs/widgets/list_view.md` 策略说明，继续保持 legacy-first。
- `scroll_view` 已记录 `ScrollRef` 命令队列、滚动偏移、auto-to-end key 和 host-state 归属边界。
- `list_view` 已记录虚拟窗口、触底回调去重、itemBuilder 调用时机和 item identity 边界。
- 本步骤没有新增 `ScrollViewElement` / `ListViewElement` metadata，没有修改 examples 代码或 `examples/docs_browser` runtime 映射。
- `go test ./...` 继续通过。
- D5.3 已完成 `docs/widgets/grid.md` 策略说明，区分固定 children 的 `Grid` 与动态/列表承载的 `GridView`。
- `grid` 已记录 row virtualization、触底回调去重、`GridMinItemWidth` 列数变化和 cell identity 边界。
- 本步骤没有新增 `GridElement` / `GridViewElement` metadata，没有修改 examples 代码或 `examples/docs_browser` runtime 映射。
- `go test ./...` 继续通过。
- D5.4 已完成 `docs/widgets/dialog.md` 和 `docs/widgets/popup.md` overlay 生命周期说明。
- `dialog` / `popup` 已记录受控 open、mask close、`DialogRef` / `PopupRef` 命令队列、overlay mount/unmount、stacking/focus 和事件拦截边界。
- 本步骤没有新增 `DialogElement` / `PopupElement` metadata，没有修改 examples 代码或 `examples/docs_browser` runtime 映射。
- `go test ./...` 继续通过。
- D5.5 已完成 `docs/widgets/toast.md` 生命周期说明和 `examples/virtual_scroll/README.md` 兼容说明。
- `toast` 已记录 timer / auto-close、`ToastOnClose` 清理、redraw 请求、重复 message 和 multiple toast stacking 边界。
- `examples/virtual_scroll` 已明确继续作为 legacy `Run` / `Widget` integration reference，不在本阶段迁移到 `RunElement`。
- 本步骤没有新增 `ToastElement` metadata，没有修改 examples 代码或 `examples/docs_browser` runtime 映射。
- `go test ./...` 继续通过。
- D5.6 已完成 Batch 5 最终兼容性检查。
- 已确认 `examples/docs_browser` 仍使用 `scroll_view_basic`、`list_view_basic`、`grid_basic`、`dialog_basic`、`popup_basic`、`toast_basic` legacy 示例入口。
- 已确认 Batch 5 文档没有新增稳定 Element wrapper metadata 或 React-style snippet；所有 Element 名称仅作为“不冻结 API 名称”的边界说明出现。
- 已确认 `examples/virtual_scroll/main.go` 仍使用 legacy `ui.Run` / `Widget` 路径，未改写为 `RunElement`。
- `go test ./...` 继续通过。

### Batch 5 完成

- Batch 5 的 D5.1-D5.6 已全部完成。
- 本批次只补 strategy / compatibility notes，没有修改 runtime、examples 代码或 docs_browser runtime 映射。
- 本批次没有冻结 `ScrollViewElement`、`ListViewElement`、`GridElement`、`GridViewElement`、`DialogElement`、`PopupElement`、`ToastElement` API 名称。
- 后续如果要为这些控件补 React-style wrapper，需要先完成 host-state、overlay lifecycle、notification cleanup 和 virtual item identity 设计收束。

### Batch 6 准备

- Batch 6 覆盖综合 showcase 示例：`examples/docs_browser`、`examples/team_workspace`、`examples/advanced_components`、`examples/animation_showcase`、`examples/vscode_layout`。
- Batch 6 是 integration showcase planning，不是原地迁移批次；这些示例继续保留 legacy `Run` / `Widget` 路径。
- 本批次不修改 `examples/docs_browser` runtime 映射，不改写现有 showcase 代码，不新增或冻结综合示例专用的 React-style API 名称。
- 如果后续要迁移综合 showcase，优先新增 parallel React-style showcase，再评估是否替换 legacy 示例，避免打断现有集成演示。
- 推荐顺序：先锁定 scope 和风险清单，再处理 `docs_browser`，然后处理 `team_workspace`，再处理 `advanced_components` / `vscode_layout`，最后处理 `animation_showcase` 与兼容性检查。

### Batch 6 D6.1 integration inventory

- D6.1 已锁定 Batch 6 范围：`docs_browser`、`team_workspace`、`advanced_components`、`animation_showcase`、`vscode_layout`。
- `docs_browser` 是文档运行时和 legacy example id host，承担 Markdown 加载、metadata 解析、示例映射和预览职责，后续必须单独设计迁移路径。
- `team_workspace` 是最复杂业务 dashboard showcase，组合复杂输入、筛选状态、滚动/列表、弹层、toast、app bar、bottom navigation 和大量共享状态。
- `advanced_components` 是组件综合展示，组合 tabs、select、image、list、dialog、toast、bottom navigation，适合作为后续 parallel React-style showcase 的候选。
- `vscode_layout` 主要覆盖复杂布局、菜单、文件列表和 textfield 状态，风险低于 `docs_browser` / `team_workspace`，但仍不适合在未规划前原地改写。
- `animation_showcase` 依赖 time-based `anim.New(...).Value(ctx)`、redraw 和 frame lifecycle，迁移前需要 animation lifecycle / effect cleanup 设计收束。
- D6.1 没有修改 examples 代码、没有新增 README、没有修改 docs_browser runtime 映射。
- 后续 D6.2 从 `examples/docs_browser` 兼容说明开始。

### Batch 6 D6.2 docs_browser compatibility note

- `examples/docs_browser/README.md` 已补充兼容说明，明确 docs browser 继续作为 legacy `Run` / `Widget` integration reference。
- 该示例承担 Markdown 加载、metadata 解析、legacy example id 映射、交互式 preview host 和在线 docs fallback，不能作为普通 showcase 原地迁移。
- D6.2 明确不把 `examples/docs_browser` 改写为 `RunElement`，也不修改 runtime 示例映射。
- 后续若要迁移 docs browser，应作为独立 runtime project 设计，先定义 legacy id 保留、React-style preview 并存、local/remote docs loading 和 preview state 稳定性规则。
- `examples/docs_browser/main.go` 未修改，现有 docs_browser 仍用于验证迁移后的 Markdown 文档能被旧浏览器正常承载。
- `go test ./...` 继续通过。
- 后续 D6.3 可以进入 `examples/team_workspace` 兼容说明。

### Batch 6 D6.3 team_workspace compatibility note

- `examples/team_workspace/README.md` 已补充兼容说明，明确 team workspace 继续作为 legacy `Run` / `Widget` integration reference。
- 该示例是 Batch 6 中最复杂的业务 dashboard showcase，组合 shared task state、selected item、search/filter/sort、tabs、responsive layout、多个 scroll regions、virtual lists、dialogs、toasts、app bar、bottom navigation 和 settings controls。
- D6.3 明确不把 `examples/team_workspace` 原地改写为 `RunElement`，也不冻结 dashboard 专用 React-style API 名称。
- 该示例依赖复杂输入和 host-state-heavy 控件：`TextField`、`Select`、`RadioGroup`、`Slider`、`ScrollView`、`ListView`、`Dialog`、`Toast` 等；迁移前需要 complex input、list identity、overlay lifecycle、toast cleanup 和 shared state 策略稳定。
- 后续若要迁移 team workspace，应优先新增 parallel React-style dashboard 示例，验证等价后再讨论是否替换 legacy 示例。
- `examples/team_workspace/main.go` 未修改，现有示例继续作为多控件集成回归参考。
- `go test ./...` 继续通过。
- 后续 D6.4 可以进入 `examples/advanced_components` 和 `examples/vscode_layout` 兼容说明。

### Batch 6 D6.4 advanced_components and vscode_layout compatibility notes

- `examples/advanced_components/README.md` 已补充兼容说明，明确 advanced components 继续作为 legacy `Run` / `Widget` integration reference。
- `examples/advanced_components` 组合 `AppBar`、`Tabs`、`Select`、`Image`、`ListView`、`ScrollView`、`BottomNavigation`、`Dialog` 和 `Toast`，适合作为未来 parallel React-style advanced showcase 候选，但不在本批次原地迁移。
- advanced components 迁移前需要先稳定 select popup state、list reach-end、scroll offset、dialog mount/open state 和 toast cleanup 等 lifecycle / host-state 边界。
- `examples/vscode_layout/README.md` 已补充兼容说明，明确 VSCode layout 继续作为 legacy `Run` / `Widget` integration reference。
- `examples/vscode_layout` 覆盖 IDE-style layout、top menus、nested menu panels、sidebar tools、file list、editor-like `TextField` state、status bar 和 layered click-away handling；风险低于 `docs_browser` / `team_workspace`，但仍不适合未规划原地改写。
- VSCode layout 迁移前需要先稳定 text input host-state、list selection identity、menu overlay lifecycle、command-driven updates 和 click-away event handling。
- D6.4 明确不改写这两个示例的 `main.go`，不改交互行为，不新增或冻结 showcase 专用 React-style API 名称。
- `go test ./...` 继续通过。
- 后续 D6.5 可以进入 `examples/animation_showcase` lifecycle 说明。

### Batch 6 D6.5 animation_showcase lifecycle note

- `examples/animation_showcase/README.md` 已补充 lifecycle 说明，明确 animation showcase 继续作为 legacy `Run` / `Widget` integration reference。
- 该示例覆盖 easing comparison、pulse transition、staggered entrance、color interpolation 和 progress indicators，核心依赖 `anim.New(...).Value(ctx)`、frame time、redraw scheduling 和 legacy `ui.State` target state。
- `examples/animation_showcase` 还依赖 tab-scoped context path，例如 `ctx.Scope("easing")`、`ctx.Scope("pulse")`、`ctx.Scope("stagger")`，用来隔离不同动画 panel 的状态身份。
- D6.5 明确不把该示例原地改写为 `RunElement`，不改动画 track 逻辑，不新增或冻结 animation showcase 专用 React-style API 名称。
- 后续迁移前需要明确 React-style animation lifecycle：animation tracks 如何绑定 component instance、`UseEffect` cleanup 如何停止或重置动画、key/remount 如何影响 in-flight animation、redraw 如何调度。
- `examples/animation_showcase/main.go` 未修改，现有示例继续作为 frame-driven animation 兼容回归参考。
- `go test ./...` 继续通过。
- 后续 D6.6 可以进入 Batch 6 最终兼容性检查。

### Batch 6 D6.6 compatibility check

- D6.6 已完成 Batch 6 最终兼容性检查，确认本批次只补 integration inventory、compatibility notes 和 lifecycle note。
- `examples/docs_browser/main.go`、`examples/team_workspace/main.go`、`examples/advanced_components/main.go`、`examples/animation_showcase/main.go`、`examples/vscode_layout/main.go` 均继续使用 legacy `ui.Run` / `Widget` 路径，未改写为 `RunElement`。
- `examples/docs_browser` runtime 示例映射保持不变，仍包含 legacy entries：`row_basic`、`button_basic`、`textfield_basic`、`slider_basic`、`radio_group_basic`、`select_basic`、`dialog_basic`、`popup_basic`、`toast_basic`、`scroll_view_basic`、`list_view_basic`、`grid_basic`、`router_basic`。
- Batch 6 已新增兼容说明文件：`examples/docs_browser/README.md`、`examples/team_workspace/README.md`、`examples/advanced_components/README.md`、`examples/vscode_layout/README.md`、`examples/animation_showcase/README.md`。
- Batch 6 没有新增或冻结综合 showcase 专用 React-style API 名称，没有修改示例交互行为，没有修改 docs_browser runtime mapping。
- `examples/react_workspace` 是当前 React-style standalone showcase；Batch 6 candidate examples 保持 legacy integration references。
- `go test ./...` 继续通过。

### Batch 6 完成

- Batch 6 的 D6.1-D6.6 已全部完成。
- 本批次以 compatibility/planning notes 收尾，不做综合 showcase 原地迁移。
- 后续如需迁移综合 showcase，应优先新增 parallel React-style showcase，并在 host-state、overlay lifecycle、animation lifecycle、list identity 和 docs_browser runtime 设计收束后再评估替换 legacy 示例。

## 回退策略

- Element wrapper 行为不一致时，保留 legacy 示例为主，React-style 示例标注为实验。
- docs_browser 渲染不稳定时，只更新 Markdown 文档，不改 runtime 示例映射。
- refs / command queue 依赖强的控件后置，等 host-state 策略更清楚后再迁移。
- Router 示例保留 legacy 与 React-style 对照，避免打断现有 Router 文档。

## 非目标

- 不在 P7.3 批量重写 examples。
- 不删除 legacy 示例。
- 不把 docs_browser 立即迁移到 `RunElement`。
- 不在示例迁移阶段顺手改变 widget 行为。
