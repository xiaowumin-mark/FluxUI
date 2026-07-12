# 弃用 ledger

本文件是 FluxUI 当前弃用、兼容 shim 与迁移窗口的权威账本。它与 [`api/ui.snapshot`](../api/ui.snapshot) 同步，面向用户的 [`CHANGELOG.md`](../CHANGELOG.md) 只能摘要，不得给出相互矛盾的状态。

## 维护规则

- 新增或修改 Go `Deprecated:` 注释、兼容 no-op、替代 API 或移除计划时，必须同一 PR 更新本账本、snapshot、相关文档和 CHANGELOG。
- `Deprecated` 表示新代码不应继续采用该入口；它不自动等于已删除、不可编译或下一个小版本会移除。
- “兼容保留”表示可继续使用但可能不是新项目推荐路径；没有明确版本窗口，不得承诺移除日期。
- 任何移除都需要：已发布的弃用窗口、可行替代项、迁移说明、旧路径 compile/runtime 回归和版本评审。

## 当前条目

| 符号/能力 | 状态 | 新代码替代项 | 最早重新评估/移除 | 迁移与兼容说明 |
| --- | --- | --- | --- | --- |
| `Widget`、`App`、`Run`、`Window`、`RunMulti` | Deprecated；兼容保留 | `Element`、`RunElement` 与组件函数 | 仅在未来 major 评审 | 现有项目继续可编译；小版本不得静默破坏 root API。 |
| `style.Style`、`Container`、`ContainerElement` | Deprecated；兼容保留 | `Decoration`、`ContainerDecoration`、`ContainerDecorationElement` | 仅在未来 major 评审 | 新代码使用 Decoration 链；旧代码保持迁移期行为。 |
| `ClickArea`、`ClickAreaElement` | Deprecated；兼容保留 | `Pressable`、`PressableElement` | 仅在未来 major 评审 | 低层旧别名不应出现在新示例的首选路径。 |
| `Animate` 及其旧 Widget 动画 option 便捷包装 | Deprecated；兼容保留 | `UseAnimatedValue`、`UseAnimatedDecoration` 及其直接参数 | 仅在 future major 评审 | 旧动画入口可保留以保护已有 Widget 代码；新组件不得建立在该路径上。 |
| `SelectSearchable[T](bool)` | Deprecated；编译兼容 no-op | 当前无直接替代；固定选项使用 `Select`，真实搜索等待 R1 Combobox/Autocomplete | R1 API 评审后 | 不创建搜索框、不筛选、不处理 query/IME/async suggestions；Docs Browser/API 列表不得宣传为搜索能力。 |
| `SelectQuick[T](bool)`、`SelectTypeaheadDelay[T](time.Duration)` | Deprecated；编译兼容 no-op | 当前无直接替代；遵循普通 Select 的已记录动画与键盘行为 | R1 API 评审后 | 当前 Select 不提供 quick-animation 或 typeahead 合同；新代码不得使用这些预留 option。 |

## 兼容但未弃用的边界

以下入口若无本账本的 `Deprecated` 条目，仍按当前 API snapshot 的兼容承诺处理；它们不应因为“旧”或“legacy”措辞而被擅自标记为待删：Router/Route/Navigate/UseParams、`FromWidget` escape hatch，以及其他 snapshot 中未弃用的导出符号。

## 每次发布的对账

发布 owner 必须确认：

1. `api/ui.snapshot` 中的 deprecation 标记与本表一致。
2. CHANGELOG 使用相同的“Deprecated / 兼容保留 / no-op”措辞，并列出用户可见迁移影响。
3. 每个条目的替代项都能编译、其说明可访问，且旧路径有至少一个 compile/runtime 回归。
4. 没有把预留 option、实验 adapter 或未实现的 native capability 描述为已完成特性。

## R0 文档完成记录

- 2026-07-12：将 root API、装饰/点击、旧动画路径及 `SelectSearchable` 的弃用/兼容状态汇总到单一 ledger。
- `SelectSearchable` 的 R0 决定是兼容 no-op；后续只有 R1 的真实搜索 API 才能改变这一结论。
