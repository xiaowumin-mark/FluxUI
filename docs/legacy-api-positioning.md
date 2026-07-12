# Legacy API 定位

## 目标

本文记录 Phase 7.1 的 legacy API 定位。legacy 在当前阶段表示“兼容入口”，不表示立即删除或停止维护；当前弃用状态、替代项和版本窗口以 [`docs/deprecation-ledger.md`](deprecation-ledger.md) 为准。

## 结论

- `Widget` / `Run` / `App` / `Window` / `RunMulti` 已标记为 Deprecated，但继续作为稳定兼容入口。
- `RunElement` / `Element` / `Component` 是新项目推荐入口。
- `FromWidget` 是新旧 API 混用的长期桥接能力，细则见 `docs/escape-hatch-strategy.md`。
- 旧 Router API 继续保留，React-style Router hooks 和 `RouterElement` 是新写法。
- deprecation 文案和版本节奏见 `docs/deprecation-and-versioning.md`；新增标记必须同步更新 ledger、API snapshot 与 CHANGELOG。

## 推荐使用边界

| 场景 | 推荐入口 | 说明 |
| --- | --- | --- |
| 新项目 root | `RunElement` | 优先进入 React-style Element runtime。 |
| 旧项目 root | `Run` / `App` / `Window` / `RunMulti` | 保持稳定，不要求立即迁移。 |
| 新组件 | `Element` / `Component` | 优先使用新 hooks 和 reconciler 管理。 |
| 旧组件 | `Widget` | 继续作为兼容层和未迁移控件入口。 |
| 新旧混用 | `FromWidget` / Element wrappers | 渐进迁移，不需要一次性重写。 |
| Router | `UseNavigate` / `UseLocation` / `UseParams` / `RouterElement` | 新代码优先使用。 |
| Router 旧代码 | `Router` / `Route` / `Navigate` / `UseParams` | 稳定兼容，不强制替换。 |

## 非目标

- 不删除旧 API。
- 不修改 `Run` / `App` / `Window` / `RunMulti` 签名。
- 不把所有 examples 立即迁移到 `RunElement`。
- 不在小版本中删除已弃用的兼容入口。

## 迁移建议

- 新页面优先用 `Element` / `Component` 编写。
- 旧 widget 可以先通过 `FromWidget` 接入 Element 树。
- 旧项目可保持 `Run` / `Widget` root，再逐步引入 Element wrapper。
- Router 新代码优先使用 hooks，旧 `Router` 调用路径继续可用。

## Escape Hatch

- `FromWidget` 长期保留，用于 legacy widget 进入 Element 树。
- `RenderElement` 主要用于运行时、测试和兼容层，不作为业务层首选主路径。
- 新旧 state 生命周期保持分离，legacy widget state 不迁入 HookSlot。
