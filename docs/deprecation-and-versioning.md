# Deprecation 与版本节奏

## 权威关系

当前状态、替代项和迁移窗口以 [`docs/deprecation-ledger.md`](deprecation-ledger.md) 为准。本文件解释版本策略；[`CHANGELOG.md`](../CHANGELOG.md) 只能提供面向发布的摘要，不能与 ledger 或 `api/ui.snapshot` 冲突。

## 策略结论

- `Widget`、`Run`、`App`、`Window`、`RunMulti` 已有 Go `Deprecated:` 标记，但仍是**兼容保留**入口；它们没有在小版本删除的计划。
- 新项目推荐 `RunElement`、`Element` 与函数组件；旧项目可继续使用既有 root API。
- `style.Style`、旧 `Container`、`ClickArea` 和旧 Widget 动画路径同样按 ledger 中的“Deprecated；兼容保留”处理。
- `SelectSearchable`、`SelectQuick`、`SelectTypeaheadDelay` 是 Deprecated 的编译兼容 no-op，不能在示例、文档或 CHANGELOG 中暗示已实现搜索、typeahead 或 quick-animation。
- `FromWidget` 是长期 escape hatch，不 deprecated；旧 Router API 也未 deprecated。

## 版本节奏

| 阶段 | 动作 | 说明 |
| --- | --- | --- |
| Current | ledger 驱动的兼容保留 | Deprecated 入口保持可编译、可运行，并有迁移说明。 |
| Next minor | 文档/示例优先路径迁移 | 新文档默认展示推荐 API，保留 legacy 对照和回退。 |
| Later minor | 兼容回归与提示强化 | 仅在替代 API 经真实示例验证稳定后，再加强迁移提示。 |
| Future major | 逐项重新评估移除 | 只有完成弃用窗口、迁移指南、旧路径回归和版本评审后才可移除。 |

## 新弃用或移除的准入条件

1. 同一 PR 更新 `docs/deprecation-ledger.md`、`api/ui.snapshot`、关联文档和 `CHANGELOG.md`。
2. 写清可编译的替代 API、迁移示例和最早重新评估版本；没有替代项不得承诺移除。
3. 保留旧路径的 compile/runtime 回归，直到 future-major 移除评审完成。
4. 小版本不得静默改变 Widget/Run root、callback 顺序、受控值或 Ref 语义。
5. 预留 option 在具备可验证行为前只能标为 no-op/实验性，不能包装成已完成能力。

## 兼容但未弃用的 API

- `Router`
- `Route`
- `Navigate`
- `UseParams`
- `FromWidget`

更多符号以 `api/ui.snapshot` 和 ledger 为准。
