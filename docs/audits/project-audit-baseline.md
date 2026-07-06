<!-- fluxui-doc-meta
{
  "id": "project_audit_baseline",
  "title": "项目逻辑审查基线记录",
  "category": "工程审查",
  "order": 47,
  "summary": "记录 FluxUI 大项目逻辑审查 Batch 0~1 的基线事实，按任务拆分到独立子文件。",
  "example": { "id": "event_system_basic" },
  "apis": [
    "internal.Context",
    "internal.Runtime",
    "ui.Element",
    "widget.Widget",
    "event.Event"
  ]
}
-->

# FluxUI 项目逻辑审查基线记录

本文件是审查基线的索引入口，各任务基线记录已拆分到独立子文件。

查看本文涉及的审查路线图和任务拆分，请参见：
- `docs/project-audit-roadmap.md`
- `docs/project-audit-task-breakdown.md`

---

## Batch 0：基线冻结

| 任务 | 文件 | 状态 | 日期 |
| --- | --- | --- | --- |
| A0.1 工作区状态冻结 | [audit-a0.1-workspace-freeze.md](audit-a0.1-workspace-freeze.md) | Done | 2026-07-06 |
| A0.2 版本和依赖快照 | [audit-a0.2-version-deps.md](audit-a0.2-version-deps.md) | Done | 2026-07-06 |
| A0.3 核心测试基线 | [audit-a0.3-core-tests.md](audit-a0.3-core-tests.md) | Done | 2026-07-06 |
| A0.4 示例 smoke 基线 | [audit-a0.4-example-smoke.md](audit-a0.4-example-smoke.md) | Done | 2026-07-06 |

## Batch 1：包边界和 runtime 基础

| 任务 | 文件 | 状态 | 日期 |
| --- | --- | --- | --- |
| A1.1 包依赖方向图 | [audit-a1.1-dependency-graph.md](audit-a1.1-dependency-graph.md) | Done | 2026-07-06 |
| A1.2 public API 所有权清单 | [audit-a1.2-public-api-ownership.md](audit-a1.2-public-api-ownership.md) | Done | 2026-07-06 |
| A1.3 escape hatch 边界审查 | [audit-a1.3-escape-hatch.md](audit-a1.3-escape-hatch.md) | Done | 2026-07-06 |
| A1.4 旧 API 兼容边界 | [audit-a1.4-legacy-api-compatibility.md](audit-a1.4-legacy-api-compatibility.md) | Done | 2026-07-06 |
| A2.1 Frame 生命周期图 | [audit-a2.1-frame-lifecycle.md](audit-a2.1-frame-lifecycle.md) | Done | 2026-07-06 |
| A2.2 PathID 和状态保存审查 | [audit-a2.2-pathid-state-retention.md](audit-a2.2-pathid-state-retention.md) | Done | 2026-07-06 |

---

## 关联能力总览

- 工作区状态冻结
- 审查产物记录
- Go 工具链复现
- Gio 运行时依赖复现
- 核心依赖版本冻结
- 核心包测试基线
- 既有失败识别
- 后续审查回归参照
- 示例 smoke 编译基线
- 包依赖方向
- runtime 所有权边界
- 反向依赖识别
- public API 所有权
- 旧 API 兼容边界
- runtime/event/widget 语义归属
- Gio context escape hatch
- Gio 原始事件桥接
- Widget 与 Element 兼容桥
- 事件系统迁移边界
- Runtime frame begin/end
- Context 根作用域和子作用域
- 每帧 event registry 重建
- 跨帧 PathID 和状态保存
- runtime frame 清扫边界
- PathID 身份规则
- Element keyed/unkeyed 状态保留
- Widget 列表位置状态风险
- Portal event parent 与 state path 边界
- Ref attach 命令目标边界
