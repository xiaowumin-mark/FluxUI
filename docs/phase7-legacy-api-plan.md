# Phase 7 Legacy API 收敛规划

## 目标

把旧 `Widget` API 明确定位为 legacy 兼容层，同时让 React-style `Element` API 成为新项目的默认推荐路径。Phase 7 不以删除旧 API 为目标，而是先建立清晰迁移、文档和版本策略。

## 入口标准

- Phase 6 已完成 root Element 入口收口。
- `RunElement` 与旧 `Run` / `Widget` 入口的分工清晰。
- Router React 化和 root Element 入口都已有兼容边界。
- examples 迁移已隔离成单独计划。

## 小阶段拆分

### Phase 7.1 Legacy API 定位

目标：明确旧 `Widget` API 是兼容层，而不是立即删除目标。

交付：

- 标注 legacy API 的推荐使用边界。
- 保持旧项目无需立即迁移。
- 明确哪些 API 仍是稳定兼容入口。

完成标准：

- legacy 定位文档清晰。
- 不破坏现有 `Widget` 用户代码。

定位结论：

- `Widget` / `Run` / `App` / `Window` / `RunMulti` 是稳定 legacy 兼容入口，不在 Phase 7 删除或改签名。
- legacy 不等于 deprecated；是否加入 deprecated 文案留到 Phase 7.4 决定。
- 新项目优先使用 `RunElement` / `Element` / `Component`，旧项目可以继续使用 `Run` / `Widget`。
- legacy widget 仍可以通过 `FromWidget` 进入 Element 树，escape hatch 细则在 Phase 7.2 收口。
- Router 旧 API 继续作为兼容层，`RouterElement` / hooks 是新写法，不强制替换旧写法。

推荐使用边界：

| 场景 | 推荐入口 | 说明 |
| --- | --- | --- |
| 新项目 root | `RunElement` | 进入 React-style runtime 和 reconciler。 |
| 旧项目 root | `Run` / `App` / `Window` / `RunMulti` | 继续稳定运行，不要求立即迁移。 |
| 新组件 API | `Element` / `Component` | 优先使用 `UseState` / `UseEffect` 等新 hooks。 |
| 旧组件 API | `Widget` | 作为兼容层保留，适合旧代码和未迁移控件。 |
| 新旧混用 | `FromWidget` / Element wrappers | 渐进迁移，避免一次性重写。 |
| Router | hooks / `RouterElement` 优先，旧 `Router` 保留 | 新旧路由入口并存。 |

进度记录：

- 已新增 `docs/legacy-api-positioning.md`，集中记录 legacy API 定位与使用边界。
- 已确认 P7.1 不改动旧 API 签名，不添加 deprecated 文案。
- 已确认旧项目无需立即迁移，新项目推荐从 `RunElement` 开始。

### Phase 7.2 Escape Hatch 策略

目标：确认 `FromWidget` 或等价能力长期保留，用于 Element 与 Widget 混用。

交付：

- 记录 `FromWidget` 的长期定位。
- 明确新旧组件混用边界。
- 保留迁移期 fallback 策略。

完成标准：

- `FromWidget` escape hatch 规则清楚。
- 新旧混用不需要一次性迁移。

长期定位：

- `FromWidget` 是长期保留的 Widget -> Element 桥接能力，不是临时迁移 hack。
- Element wrappers 可以继续在内部使用 `FromWidget` 包装 legacy widget，直到对应控件真正 host-state 化。
- `RenderElement` / internal render helper 是 Element -> Widget 的运行时桥接，不建议作为业务代码主路径。
- `FromWidget(nil)` 返回 nil，保持安全 no-op 行为。

混用边界：

| 方向 | 推荐方式 | 边界 |
| --- | --- | --- |
| Widget 子树进入 Element 树 | `FromWidget(widget)` | 适合未迁移 legacy 控件、第三方 Widget、临时适配。 |
| Element 子树进入 Widget root | `RenderElement(element)` 或现有 wrapper | 主要用于运行时/测试/兼容层，不推荐业务层频繁手写。 |
| 新组件调用旧控件 | Element wrapper 内部使用 `FromWidget` | 保持新 API 表面，底层可继续 legacy 实现。 |
| 旧页面逐步迁移 | 保持 `Run` root，局部引入 wrapper | 不要求一次性切换到 `RunElement`。 |
| 新页面保留旧控件 | `RunElement` root + `FromWidget` | 推荐的渐进迁移路径。 |

状态与生命周期规则：

- `FromWidget` 包装的 legacy widget 状态仍由 legacy widget / ref / command queue 持有。
- `FromWidget` 不把 legacy widget state 迁入 component HookSlot。
- `FromWidget` host leaf 只参与 Element identity 和 reconciler 遍历，不改变 Gio 每帧 layout 模型。
- component unmount cleanup 与 legacy widget state cleanup 仍分离。
- 需要 route / list / conditional 稳定复用时，外层应继续使用 Element `Key`。

迁移建议：

- 新 API 表面优先提供 Element wrapper，而不是要求用户手动 `FromWidget`。
- 对暂未迁移的复杂控件，先用 `FromWidget` 暴露兼容入口。
- 等控件具备 host-state 设计后，再逐步减少 wrapper 内部对 legacy widget 的依赖。
- 不在 Phase 7.2 删除或弱化 `FromWidget`。

进度记录：

- 已新增 `docs/escape-hatch-strategy.md`，集中记录 `FromWidget` 长期定位。
- 已确认 `FromWidget` 是长期 bridge，不是短期 hack。
- 已记录 Widget -> Element、Element -> Widget、新旧 wrapper 的混用边界。

### Phase 7.3 Docs 默认示例迁移策略

目标：制定 docs 默认示例迁移到 React-style 的分批计划。

交付：

- 选择第一批适合迁移的 docs/examples。
- 保留 legacy 示例对照或说明。
- 不在此阶段直接大规模重写 examples。

完成标准：

- docs 示例迁移顺序明确。
- 迁移风险和回退路径明确。

迁移原则：

- P7.3 只制定迁移顺序和风险边界，不直接大规模重写 examples。
- docs 默认示例优先迁移为 React-style 写法；legacy 示例保留为对照或兼容说明。
- 每一批迁移都必须保持 `go test ./...` 通过，并尽量保留 docs_browser 的旧示例渲染能力。
- 涉及复杂 host state、refs、滚动、弹层或多窗口的示例暂缓，避免把复杂控件 host-state 化混入 docs 迁移。

分批策略：

| 批次 | 范围 | 代表示例 | 理由 |
| --- | --- | --- | --- |
| Batch 1 | 无状态展示 / 基础布局 | `text_basic`、`spacer_basic`、`divider_basic`、`row_basic`、`column_basic`、`stack_basic`、`center_basic`、`padding_basic`、`container_basic`、`sizing_basic` | 已有 Element wrappers，风险最低。 |
| Batch 2 | 简单交互 | `button_basic`、`click_area_basic`、`checkbox_basic`、`switch_basic` | 已有 Element wrappers，但涉及事件和 legacy widget state。 |
| Batch 3 | Router 新写法 | `router_basic`、`examples/router` 对照版本 | 已有 hooks、`RouterElement`、route identity 和 transition 规则。 |
| Batch 4 | 复杂输入和表单 | `textfield_basic`、`select_basic`、`radio_group_basic`、`slider_basic`、`form_validation` | 需要更清晰的 host state / ref 策略。 |
| Batch 5 | 滚动、列表、网格、弹层、通知 | `scroll_view_basic`、`list_view_basic`、`grid_basic`、`dialog_basic`、`popup_basic`、`toast_basic`、`virtual_scroll` | 涉及 refs、virtual state、overlay 生命周期，最后迁移。 |
| Batch 6 | 综合 examples | `docs_browser`、`team_workspace`、`advanced_components`、`animation_showcase`、`vscode_layout` | 保留为 integration showcase，等基础 docs 稳定后再评估。 |

回退策略：

- 每个 docs 示例迁移前先保留 legacy 代码片段或兼容说明。
- 如果 Element wrapper 行为和 legacy widget 不一致，优先回退到 legacy 示例并记录差异。
- 如果迁移影响 docs_browser 渲染稳定性，先只更新 Markdown 文档，不修改 runtime 示例映射。
- 对 refs / command queue 依赖强的控件，必须等 host-state 边界明确后再迁移。

第一批候选：

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

进度记录：

- 已新增 `docs/docs-example-migration-plan.md`，集中记录 docs/examples 分批迁移策略。
- 已明确 P7.3 不直接大规模重写 examples。
- 已选择第一批低风险迁移候选，并记录回退策略。

### Phase 7.4 Deprecation 与版本节奏

目标：制定不破坏旧项目的 deprecation 文案和版本节奏。

交付：

- 明确 deprecated 文案是否进入代码注释。
- 明确版本节奏和过渡期。
- 明确哪些 API 暂不 deprecated。

完成标准：

- deprecation 策略可执行。
- 旧项目迁移压力可控。

策略结论：

- Phase 7.4 不在代码中添加 Go `Deprecated:` 注释。
- 当前阶段只做文档级 migration guidance，不触发 IDE / pkg.go.dev 的废弃警告。
- `Widget` / `Run` / `App` / `Window` / `RunMulti` 暂不 deprecated，继续作为稳定兼容入口。
- `FromWidget` 明确不 deprecated，长期作为 escape hatch 保留。
- 只有当 React-style API、docs 默认示例和 host-state 复杂控件迁移都稳定后，才重新评估代码级 deprecation。

建议版本节奏：

| 阶段 | 动作 | 说明 |
| --- | --- | --- |
| Current | 文档级推荐 | 新项目推荐 `RunElement`，旧 API 保持稳定。 |
| Next minor | 示例分批迁移 | docs 默认示例逐步切到 React-style，保留 legacy 对照。 |
| Later minor | 兼容层说明强化 | 如果新 API 稳定，可在文档中把 legacy 标记为兼容层。 |
| Future major | 重新评估代码级 deprecation | 只在明确版本窗口、迁移指南和替代 API 后考虑。 |

代码注释规则：

- 暂不添加 `// Deprecated:` 到旧 root API。
- 暂不添加 `// Deprecated:` 到旧 Widget 控件。
- 暂不添加 `// Deprecated:` 到旧 Router API。
- 可在文档中使用“legacy compatibility API”表述。
- 若未来添加 `Deprecated:` 注释，必须同时提供替代 API、迁移示例和版本说明。

旧项目保护原则：

- 不修改旧 API 签名。
- 不删除 legacy examples。
- 不让 docs_browser 因迁移中断旧示例展示。
- 每批迁移都必须可回退到 legacy 示例。
- 用户不需要为了升级小版本被迫迁移 root API。

进度记录：

- 已新增 `docs/deprecation-and-versioning.md`，集中记录 deprecation 与版本节奏。
- 已确认 Phase 7.4 不添加代码级 `Deprecated:` 注释。
- 已确认 legacy API 继续稳定可用，未来 major 才重新评估代码级 deprecation。

## 初始完成标准

- 用户能清楚区分 legacy `Widget` API 和 React-style `Element` API。
- 旧 API 仍可用，且有明确迁移说明。
- 新项目默认推荐路径清晰。
- examples / docs 迁移不会一次性破坏现有材料。

## 进度记录

- Phase 7.1：已完成，legacy API 定位和推荐使用边界已记录，旧 API 不改签名且不立即 deprecated。
- Phase 7.2：已完成，`FromWidget` escape hatch 长期定位、混用边界和生命周期规则已记录。
- Phase 7.3：已完成，docs 默认示例分批迁移策略、第一批候选和回退边界已记录。
- Phase 7.4：已完成，deprecation 文案和版本节奏已记录，当前阶段不添加代码级 deprecated 注释。
