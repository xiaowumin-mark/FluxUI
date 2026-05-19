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
