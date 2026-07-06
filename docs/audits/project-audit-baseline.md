<!-- fluxui-doc-meta
{
  "id": "project_audit_baseline",
  "title": "项目逻辑审查基线记录",
  "category": "工程审查",
  "order": 47,
  "summary": "记录 FluxUI 大项目逻辑审查 Batch 0~2 的基线事实，按任务拆分到独立子文件。",
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
| A2.3 Runtime registry 审查 | [audit-a2.3-runtime-registry.md](audit-a2.3-runtime-registry.md) | Done | 2026-07-06 |
| A2.4 redraw 和 invalidation 审查 | [audit-a2.4-redraw-invalidation.md](audit-a2.4-redraw-invalidation.md) | Done | 2026-07-06 |

## Batch 2：布局、渲染和样式稳定性

| 任务 | 文件 | 状态 | 日期 |
| --- | --- | --- | --- |
| A3.1 基础布局约束矩阵 | [audit-a3.1-basic-layout-constraints.md](audit-a3.1-basic-layout-constraints.md) | Done | 2026-07-06 |
| A3.2 滚动容器布局审查 | [audit-a3.2-scroll-container-layout.md](audit-a3.2-scroll-container-layout.md) | Done | 2026-07-06 |
| A3.3 横向内容溢出审查 | [audit-a3.3-horizontal-overflow.md](audit-a3.3-horizontal-overflow.md) | Done | 2026-07-06 |
| A3.4 hit area 和 layout area 对齐审查 | [audit-a3.4-hit-layout-area.md](audit-a3.4-hit-layout-area.md) | Done | 2026-07-06 |

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
- Runtime event registry 生命周期
- 每帧 listener/focus/shortcut 重建
- pointer capture 跨帧 stale 风险
- Scroll wheel Gio tag 与 runtime registry 边界
- diagnostics current/last/pending 状态边界
- Runtime invalidator 与 Gio Window.Invalidate 绑定
- frame 内 op.InvalidateCmd 刷新
- redraw reason pending/current 语义
- 用户输入、动画、状态变更、程序命令 redraw 来源区分
- 动画 idle 停止与 redraw reason 基线
- 基础布局 constraints 输入/输出矩阵
- Row/Column Flex 有限和宽松约束语义
- Stack Stacked/Expanded 注释实现差异
- Center Gio Direction 约束边界
- Container/Padding inset 约束链路
- Fixed/Fill sizing 零 Max 和滚动上限风险
- ScrollView 主轴 content/viewport/offset 规则
- ListView/GridView 虚拟化滚动边界
- Grid 静态布局与 GridView 滚动语义区分
- ScrollRef 主轴命令与 offset clamp
- wheel 事件横纵轴过滤和嵌套滚动边界
- docs browser 代码块横向滚动边界
- Markdown 表格有限列宽策略
- chips row 横向 ScrollView 边界
- Tabs scrollable/fullWidth 宽度策略
- Menu/Select popup 宽度 clamp
- 普通 Row/Flex 非保护型横向布局边界
- PointerArea 子尺寸命中注册
- Pressable/ClickArea 无样式点击区域
- ContainerDecoration 视觉区域和 margin 命中边界
- Tabs item 等分/滚动命中边界
- Ripple 不改变 layout/hit size
- 空白区域不触发 hover target
