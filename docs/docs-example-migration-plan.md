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

- `examples/react_counter` 是最小 React-style counter 示例，作为后续文档 snippet 的 canonical example。
- `examples/counter` 继续保留为 legacy `Run` / `Widget` 对照示例，不在 rollout 初期删除或替换。
- 新文档示例优先参考 `RunElement`、`UseState`、`TextElement`、`ButtonElement`、基础 layout Element wrappers 的组合方式。

## 分批计划

| 批次 | 范围 | 候选 |
| --- | --- | --- |
| Batch 1 | 无状态展示和基础布局 | `text_basic`、`spacer_basic`、`divider_basic`、`row_basic`、`column_basic`、`stack_basic`、`center_basic`、`padding_basic`、`container_basic`、`sizing_basic` |
| Batch 2 | 简单交互 | `button_basic`、`click_area_basic`、`checkbox_basic`、`switch_basic` |
| Batch 3 | Router 新写法 | `router_basic`、`examples/router_element` React-style 对照版本 |
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

- 已新增 `examples/router_element` 作为独立 React-style Router 对照示例。
- 示例覆盖 `RouterElement`、`RouteElement`、`UseNavigate`、`UseLocation`、`UseParams`。
- legacy `examples/router` 保持不变，继续作为旧 Router API 对照。
- 本步骤未修改 docs_browser runtime 示例映射。

### D1.4 兼容性确认

- Batch 1 文档仍然使用 legacy example id：`text_basic`、`spacer_basic`、`divider_basic`、`row_basic`、`column_basic`、`stack_basic`、`center_basic`、`padding_basic`、`container_basic`、`sizing_basic`。
- Router 文档仍然使用 legacy example id：`router_basic`。
- `examples/docs_browser` 继续解析这些 legacy example id，没有新增 runtime 映射变更。
- 新增的 `examples/react_counter` 和 `examples/router_element` 先作为独立对照示例，不接入 docs_browser 旧示例映射。

### Batch 2 准备

- Batch 2 只覆盖简单交互文档：`button.md`、`click_area.md`、`checkbox.md`、`switch.md`。
- `examples/react_counter` 作为 `Button + UseState` 的 canonical 参考，不再新增单独 counter 示例。
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
- `examples/router_element` 作为 Router React-style 对照的 canonical example，`examples/router` 继续保留为 legacy 对照。
- 本批次仍然不修改 `examples/docs_browser` runtime 映射，`router_basic` 继续作为文档浏览器的旧入口。
- 推荐顺序：先补 Router 文档的 React-style API 说明，再整理示例段落和 identity 说明，最后做兼容性检查。

### Batch 3 进度

- `router.md` 已补充 React-style API metadata：`RouterElement`、`RouteElement`、`RouteKey`、`UseNavigate`、`UseLocation`、`UseParams`。
- `router.md` 已补充 React-style 对照示例，并说明默认 route identity 按 pattern 复用、`RouteKey` 可强制 remount。
- `examples/router/main.go` 仍然作为 legacy 对照，`examples/router_element/main.go` 作为 React-style 对照示例。
- 本次更新未修改 `examples/docs_browser` runtime 映射。

### Batch 3 兼容性确认

- `examples/docs_browser` 仍然只解析 `router_basic`。
- Router 文档的 legacy 基础用法和导航示例保持不变。
- `go test ./...` 继续通过，route identity 说明与现有测试行为一致。

### Batch 3 完成

- Router 文档迁移已完成，legacy / React-style 对照路径均已记录。
- `examples/router_element` 作为独立 React-style 对照示例继续保留。
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
