# ADR-0003：可见视口、坐标系与虚拟化边界

- 状态：Accepted（R0 文档合同）
- 决策日期：2026-07-12
- 适用范围：ListView、GridView、未来 DataGrid/TreeView、ScrollView 与锚定 Overlay。

## 背景

“窗口可见区域”“滚动容器的裁剪区域”和“虚拟集合当前构建的项”是不同概念。混用它们会让 overlay 错位、嵌套滚动错误裁剪、滚动后 hover/click 命中停留在旧位置，或使虚拟化工作量退化到与总项数相关。

## 决定

1. 统一区分三层边界：
   - **window viewport**：当前窗口中可用于 placement 的可见矩形，使用窗口坐标；`Context.Viewport()` 是当前承载该信息的内部入口。
   - **scroll viewport**：某个滚动容器实际裁剪其内容的矩形，也使用窗口坐标，且可小于 window viewport。
   - **visible item set/range**：当前帧因 scroll viewport、overscan 和布局策略而需要构建的集合项；它不是窗口矩形，也不是业务选择范围。
2. 任何 viewport 传递都必须显式标明坐标系。子上下文位置偏移后，viewport 仍代表同一窗口坐标空间；如需局部坐标，必须显式转换，不能凭约束尺寸猜测。
3. 嵌套滚动时，子 scroll viewport 是父裁剪区域与子可见区域的交集。Overlay placement 使用 window viewport，不把滚动内容的 clip 误当成整个窗口的可放置边界。
4. 虚拟集合的状态 identity 仍按 ADR-0001 的 key 归属。可见范围变更只能改变本帧布局/绘制项，不能改变 selected、active、编辑草稿或 owner。
5. 滚动位置发生真实变化后，命中、hover、focus target 和需要刷新的可见范围必须在后续 frame 与新几何一致；不得依赖鼠标移动来“修正”旧命中。

## 合同与不变量

- Layout 必须尊重父 constraints；不得无条件读取窗口全局尺寸，也不得在 Layout 执行 I/O 或后台任务。
- 无有效 viewport 的测试/降级路径可以从当前位置和 constraints 构造有限矩形，但必须被标记为降级，不得掩盖真实 viewport 漏传。
- overscan 是性能策略，不是语义可见性。对用户报告“可见项”时必须说明是否包含 overscan。
- `0` 尺寸、窄 viewport、非零 position、缩放和滚动轴变化均不得产生负尺寸或越界 rect。
- 横向、纵向与双轴策略必须显式；子控件不得通过未声明的横向可滚动性吞掉父级 wheel。

## 内部测试夹具

夹具使用固定 constraints、多帧 runtime 和可注入滚动偏移的 fake content。每个实现至少具备以下断言：

| 场景 | 断言 |
| --- | --- |
| 非零位置的局部组件 | viewport、anchor 与 hit rect 都在同一窗口坐标系。 |
| 嵌套 scroll clip | 子 viewport 不越出父 clip；外层仍可接收未消费的轴向滚动。 |
| 滚动后不移动鼠标 | hover/click 命中更新到新视觉位置。 |
| 虚拟化大集合 | 布局/构建量与 visible range 加 overscan 成正比，而非总数。 |
| viewport resize/空 viewport | range 和 placement 有确定退化，不产生 panic 或无限尺寸。 |
| 横向与纵向 | delta、offset、content size、viewport size 与回调只在真实变化时更新。 |

测试专用的 frame runner 可以记录本帧 viewport、构建过的 index/key、命中 rect 和 redraw reason；这些探针不进入公开 API，也不改变默认路径性能。

## 非目标

- 本 ADR 不定义新的通用几何包、窗口管理器或可见性 observer 公共 API。
- 本 ADR 不替代 ScrollView 的业务滚动策略，不规定 DataGrid 的列虚拟化算法。
- 本 ADR 不承诺所有平台上完整的系统窗口 inset、屏幕外 placement 或多显示器策略。

## 回滚

先将 viewport/range helper 放在使用它的 `widget` 边界内，保持 `internal.Context` 只提供中性的 viewport 信息。若试点失败，回滚该集合/overlay adapter 和夹具，不改变已有组件的普通 layout flow。

## R0 文档完成记录

- 2026-07-12：三种 viewport 概念、坐标规则、虚拟化不变量和滚动命中测试要求已明确。
- 实现完成的最低证据：至少一个集合和一个锚定 Overlay 使用同一坐标合同，且通过嵌套滚动与滚动后命中测试。
