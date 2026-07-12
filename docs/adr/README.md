# R0 共享组件合同 ADR

本目录保存高级组件 R0 的架构决策记录（ADR）。这些记录定义的是实施前必须保持的内部合同，不等同于已经公开的 API；公开符号只有在至少两个试点组件复用、测试和文档齐全后才能进入 API snapshot。

| ADR | 主题 | 状态 | R0 实施证据 |
| --- | --- | --- | --- |
| [0001](0001-collection-identity.md) | 集合 identity | Accepted（文档） | 两个试点在重排、删除和条件渲染后不串位。 |
| [0002](0002-selection-and-roving-focus.md) | 选择模型与 roving focus | Accepted（文档） | 键盘导航、禁用项跳过和焦点恢复夹具通过。 |
| [0003](0003-visible-viewport.md) | 可见视口与虚拟化边界 | Accepted（文档） | 嵌套滚动、滚动后命中刷新和可见范围夹具通过。 |
| [0004](0004-field-state.md) | 字段状态与校验展示 | Accepted（文档） | 受控状态、错误/pending 和重排夹具通过。 |
| [0005](0005-anchored-overlay.md) | 锚定 Overlay | Accepted（文档） | placement、outside click、Escape 和焦点恢复夹具通过。 |

实现约束：共享 helper 先留在 `widget` 内部或测试专用层；`internal` 不得依赖 `widget` 或 `ui`。每个 helper、适配器和试点组件必须可以独立回滚。

## R0 文档完成记录

- 2026-07-12：五项共享合同、测试夹具要求、非目标和回滚边界已建立。
- 尚未以本索引宣称任一高级组件已稳定；达到各 ADR 的实施证据并由两个试点复用后，才能把相应实现状态升级为完成。
