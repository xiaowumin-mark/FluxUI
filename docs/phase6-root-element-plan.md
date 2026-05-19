# Phase 6 Root 支持 Element 规划

## 目标

让 `Element` 成为更稳定的 root 入口，继续保留旧 `Widget` / `Run` 入口作为兼容层，不在本阶段触碰 examples 迁移。

## 小阶段拆分

### Phase 6.1 Root Element 稳定化

目标：梳理 `RunElement` 的稳定性边界，让新项目可以把它视为推荐 root 入口。

交付：

- 明确 `RunElement` 的职责边界。
- 记录 root 级 Element 入口的当前能力和限制。
- 保持旧 `Run` 入口不变。

完成标准：

- `RunElement` 的文档边界清晰。
- 新旧 root 入口共存策略清楚。

进度记录：

- 已确认 `RunElement` 仍通过 root reconciler 统一管理根 function component instance。
- 已确认 root component state / HookSlot 生命周期仍由现有 reconciler + HookStore 负责，不引入新的 root state store。
- 已确认 Phase 6 不修改 Router 兼容层和 examples 迁移策略。

### Phase 6.2 Root API 分工收口

目标：明确 root 级 Element 入口与旧 `Run` / `Widget` 入口的分工，避免后续 API 命名和调用方式混乱。

交付：

- 记录新项目优先 `RunElement` 的建议。
- 记录旧项目继续使用 `Run` / `Widget` 的兼容边界。
- 明确 Phase 6 不修改 Router 兼容层。

完成标准：

- root API 分工有清晰文档。
- 旧入口不受影响。

API 分工：

| API | 定位 | 当前建议 |
| --- | --- | --- |
| `Run(root func(*Context) Widget, opts...)` | 旧 `Widget` root 入口 | 保持稳定，旧项目继续使用。 |
| `App(root func(*Context) Widget, opts...)` | 旧 `Widget` 应用对象入口 | 保持稳定，适合需要手动管理应用对象的旧路径。 |
| `Window(root func(*Context) Widget, opts...)` / `RunMulti` | 多窗口旧 `Widget` 入口 | 保持稳定，Element 多窗口入口暂不在 P6.2 扩展。 |
| `RunElement(root Component, opts...)` | React-style `Element` root 入口 | 新项目优先使用，仍保留实验边界说明。 |

分工结论：

- 不把 `Run` 原地改成接收 `Element`，避免破坏 Go 调用签名和旧代码。
- `RunElement` 是 Element root 的明确入口，内部继续复用 app 运行链路和 root reconciler。
- `Widget` root 与 `Element` root 可以长期共存，迁移时通过 `FromWidget` 和 Element wrapper 渐进过渡。
- 多窗口 Element root、`AppElement`、`WindowElement` 这类扩展暂不在 P6.2 实现，后续可单独规划。
- Router 兼容层不随 root API 分工改变，`ui.Router` 与 `RouterElement` 继续并存。

进度记录：

- 已完成 root API 分工矩阵。
- 已确认不改动 `Run` / `App` / `Window` / `RunMulti` 签名。
- 已确认 `RunElement` 作为新项目推荐入口，但旧入口继续稳定可用。

### Phase 6.3 示例迁移隔离策略

目标：明确 examples 迁移是单独计划，不混入 Phase 6。

交付：

- 写清 examples 后续单独规划原则。
- 标注 Phase 6 不触碰 examples。
- 为后续示例迁移保留单独入口。

完成标准：

- examples 迁移边界写入文档。
- Phase 6 范围不外溢。

示例分类评估：

| 分类 | 示例 | 处理建议 |
| --- | --- | --- |
| canonical | `counter`、`docs_browser`、`router` | 保留，后续优先规划 React-style 对照或迁移。 |
| runtime / feature | `hooks_lifecycle`、`network_request`、`multi_window`、`fonts` | 保留，覆盖生命周期、异步、多窗口和字体能力。 |
| performance / layout | `virtual_scroll`、`horizontal_scroll`、`vscode_layout` | 保留，覆盖性能、大布局和滚动边界。 |
| showcase | `team_workspace`、`advanced_components`、`animation_showcase` | 保留，作为综合能力展示和 smoke demo。 |
| merge candidates | `animation`、`basic_components`、`layout`、`state_management`、`textfield_demo`、`popup_demo` | 暂保留，后续单独计划合并、降级或改造为 React-style 示例。 |
| cleanup candidate | `router_demo` | 空目录，可在单独 cleanup 任务中删除。 |

本阶段统一的书写方式：

- examples 中统一使用 `ui "github.com/xiaowumin-mark/FluxUI/ui"` 显式导入别名。
- 保持 `ui.Run` / `ui.Widget` legacy root 写法不变，避免把示例迁移混入 Phase 6。
- 保留现有示例目录与行为，只做低风险格式统一和边界记录。
- React-style examples 后续单独规划，不在 P6.3 执行。

进度记录：

- 已完成示例保留 / 合并 / 清理候选评估。
- 已统一 examples 的 `ui` 导入别名。
- 已移除 `team_workspace` 源文件开头的 UTF-8 BOM。
- 未迁移示例到 `RunElement`，未删除任何示例目录。

### Phase 6.4 Phase 6 收口与 Phase 7 预备

目标：把 Phase 6 的 root Element 规划收束成清晰结论，并为 Phase 7 legacy API 收敛做准备。

交付：

- 汇总 Phase 6 的 root 入口结论。
- 明确新旧入口的长期共存方式。
- 为 Phase 7 的 legacy API 收敛预留入口。

完成标准：

- Phase 6 完成标准清晰。
- Phase 7 预备路线可直接接续。

Phase 6 收口结论：

- `RunElement` 是 React-style root 入口，继续通过 root reconciler 管理根 function component instance。
- 旧 `Run` / `App` / `Window` / `RunMulti` 入口保持稳定，不在 Phase 6 改签名。
- `Widget` root 与 `Element` root 长期共存，迁移通过 `FromWidget` 与 Element wrappers 渐进完成。
- Router 兼容层不随 root API 分工改变，`ui.Router` 与 `RouterElement` 继续并存。
- examples 已完成低风险书写方式统一，但迁移计划单独处理，不纳入 Phase 6。

Phase 7 预备路线：

1. 明确 legacy `Widget` API 的定位：兼容层而不是立即删除目标。
2. 为 docs 默认示例迁移到 React-style 制定分批策略。
3. 保留 `FromWidget` 作为长期 escape hatch。
4. 明确 deprecation 文案、版本节奏和不破坏旧项目的原则。

P7 入口标准：

- Phase 6 已完成 root 入口收口。
- 新旧 root API 共存边界清晰。
- examples 迁移已经隔离成单独规划。
- Router React 化和 root Element 入口都已有兼容结论。

进度记录：

- 已完成 Phase 6 收口总结。
- 已明确 Phase 7 不会删除旧 API，而是先定位 legacy API 和迁移策略。
- 已为 docs 默认示例迁移、`FromWidget` escape hatch、deprecation 策略预留入口。

## 预期交付

- 稳定推荐新项目优先使用 `RunElement`。
- 旧项目可继续使用 `Run` / `Widget`。
- Phase 6 不改变 Router 兼容层结论。

## 进入标准

- Phase 5 已收口。
- Router React 化的 hooks / route / transition 规则稳定。
- 兼容边界清晰，不需要再扩大 Phase 5 范围。

## 完成标准

- Root Element 入口有清晰文档说明。
- 旧入口保留且不被破坏。
- examples 迁移单独规划，不混入 Phase 6。

## 进度记录

- Phase 6.1：已完成，`RunElement` 仍通过 root reconciler 管理根 function component instance，root state / HookSlot 生命周期保持不变。
- Phase 6.2：已完成，root API 分工矩阵已记录，`RunElement` 与旧 `Run` / `Widget` 入口保持并存。
- Phase 6.3：已完成，示例分类评估和迁移隔离策略已记录，examples 仅做低风险书写方式统一。
- Phase 6.4：已完成，Phase 6 root 入口结论已收口，Phase 7 legacy API 收敛路线已预备。
