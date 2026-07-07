# A13.3 滚动和 hit-test 冗余审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A13.3 滚动和 hit-test 冗余审查。

## 事实结论

A13.3 的目标是审查 ScrollView、ListView、Grid、Select menu、tabs horizontal scroll 中 offset、wheel、axis、hit refresh 的重复逻辑，并为后续围绕一个底层滚动模型规划修复提供边界。当前代码没有一个统一的 FluxUI 滚动内核，而是并行存在两套模型：`ScrollView` 自己维护像素级主轴 offset、wheel tag、scrollbar、content/viewport 测量；`ListView` 和 `GridView` 直接使用 Gio `layout.List` 的 `Position`、虚拟化和输入处理。

### 现有滚动模型矩阵

| 入口 | offset 来源 | wheel 来源 | axis 策略 | hit refresh / 命中边界 | 结论 |
| --- | --- | --- | --- | --- | --- |
| `ScrollView` | `scrollState.list.Position.Offset`，但实际按单个 child 的主轴像素 offset 解释；`ScrollRef.ScrollToOffset` 也写入此 offset | 自建 `wheelTag`，`processWheelEvents` 读取 Gio `pointer.Scroll` 后派发 FluxUI `WheelEvent`，未被 prevent 时执行 default scroll | `resolveAxis(vertical, horizontal)`，仅 horizontal 且非 vertical 时为横向，否则默认纵向 | child 在 `WithPositionOffset(-offset)` 下布局，viewport clip 限制绘制；wheel tag 用 `pointer.PassOp` 注册到 viewport，不替代子控件 hit area | 功能最完整，但与 Gio `layout.List` 语义不同，适合作为未来统一模型的候选外壳 |
| `ListView` | Gio `layout.List.Position`，包含 `First`、`Offset`、`Count`、`Length`、`BeforeEnd` | Gio `layout.List.Layout` 内建处理滚动输入，FluxUI 层没有显式 wheel default action | `ListAxis` 转为 Gio axis；虚拟化和非虚拟化路径分别处理 | 虚拟化仅布局可见 item；viewport 通过 `WithViewport` 传给 item；命中依赖 Gio 每帧 ops/filter 重建 | 性能和虚拟化能力强，但与 `ScrollView` 的 `ScrollRef`、`ScrollOnChange`、preventDefault 规则不共享 |
| `GridView` | Gio `layout.List.Position` 按行滚动，`viewportMaj` 用于 reach-end fallback | Gio `layout.List.Layout` 内建处理 | 固定纵向 list，横向只用于每行 `Row` / column 布局 | 每行内构建 cell，viewport 传入 row/cell；命中依赖每帧布局重建 | 与 `ListView` 高度重复，可共享虚拟化 reach-end 和 viewport 记录逻辑 |
| `Grid` | 无滚动 offset；静态 Row/Column 网格 | 无 wheel 处理 | 静态列数或 `minItemWidth` 决定列数 | 命中只来自子控件实际布局区域 | 不应纳入滚动内核，但可共享 grid row/cell 构建逻辑 |
| Select menu | Popup 本身用 `op.Defer` 和 anchor offset 绘制；选项区域内部使用 `ListView(..., ListVirtualized(true))` | 继承 `ListView` 的 Gio list 输入；outside press 由 `md3DismissOnOutsidePress` 的 protected rect 处理 | Select popup 纵向选项列表；宽度跟 field/popup clamp | option row 用 `event.UseClickable` + ripple 注册命中；popup 移动后 protected rect 需要和当前帧位置一致 | 同时依赖 overlay placement、ListView、Clickable，最需要统一“滚动后命中刷新”测试 |
| Tabs horizontal scroll | `TabsScrollable` 外层套 `ScrollView(..., ScrollHorizontal(true), ScrollVertical(false))` | 继承 `ScrollView` 自建 wheel target；当前横向只接收 `DeltaX` | tab row 自己计算每个 tab rect；外层 ScrollView 决定横向 viewport | tab item hit area 来自 row 内 child 布局；外层 ScrollView clip/offset 改变视觉位置 | 这是 `ScrollView` 横向策略的主要消费方；shift-wheel/touchpad 横向策略应在底层统一 |
| PointerArea wheel | 自身不维护 offset | `PointerArea` 注册全轴 `pointer.Scroll` 并派发 FluxUI `WheelEvent`，没有 default scroll | `ScrollX` / `ScrollY` 均开放 | hit area 等于 child layout size，可配置 pass-through | 低层 wheel listener 与 ScrollView wheel default action 相互独立，后续需要明确 listener 与 scrollable 的优先级 |

### 重复逻辑清单

| 编号 | 重复点 | 当前位置 | 可收敛内容 | 保留差异 |
| --- | --- | --- | --- | --- |
| SH-01 | offset clamp / max offset | `scrollViewClampOffset`、`scrollViewMaxOffset`，Gio `layout.List.Position` 内部也有等价边界 | 抽象为 `scrollPosition` 的主轴 offset、content size、viewport size、clamp 规则 | `ListView` 还需要保留 item index + offset 的虚拟化 position，不能强行降级成单像素 offset |
| SH-02 | viewport 记录和传递 | `ScrollView.layoutContent`、`ListView.Layout`、`GridView.Layout` 都计算 viewport 并传给子树 | 统一 `scrollViewportScope(ctx, viewport, offset)`，集中 `RecordViewport` / `RecordVirtualizedItems` | `ScrollView` 对 child 设置大上限并整体平移；`ListView/GridView` 只布局可见 item |
| SH-03 | wheel axis filter | `ScrollView.processWheelEvents` 对横/纵轴分别设置 `ScrollX/ScrollY`；`PointerArea` 全轴开放；`ListView` 交给 Gio | 建立统一 wheel policy：主轴、交叉轴、shift-wheel、touchpad 双轴、剩余 delta、preventDefault | PointerArea 是原始事件目标，不能自动滚动；ListView 的 Gio 内建输入可能需要封装或替换才能完全统一 |
| SH-04 | wheel default action 和可取消性 | `ScrollView` 先 `DispatchWheelEvent`，返回 false 时不滚动；`ListView/GridView` 没有 FluxUI 层 default action gate | 将 scrollable 的 default action 统一放在 FluxUI event 之后，明确 `PreventDefault` 生效点 | 在替换 Gio list 输入前，ListView/GridView 只能记录为缺口，不能声称已可取消 |
| SH-05 | scrollbar 与 position 比例 | `drawScrollBar`、`viewportFromListPosition`、`scrollViewScrollByFraction` | 统一 viewport fraction 计算和 scrollbar 拖动到 offset/position 的转换 | `ListView/GridView` 当前没有公开 FluxUI scrollbar；是否启用是组件策略 |
| SH-06 | reach-end 判定 | `ListView.dispatchReachEnd`、`GridView.dispatchReachEnd` 逻辑同形 | 抽 `dispatchReachEnd(pos, total, viewportMaj, state)` | Grid 按 rowCount 判定，List 按 item count 判定 |
| SH-07 | 横向滚动封装 | Tabs scrollable、chips row、docs code/table 等都通过 ScrollView 或组合实现横向区域 | 统一 horizontal scroll wrapper：轴过滤、最大宽度、bar 可见、wheel strategy、hit refresh 测试 | Tabs row 仍需保留 indicator rect 和 fullWidth/equalWidth 策略 |
| SH-08 | 滚动后 hit refresh 假设 | `ScrollView` 用 `WithPositionOffset(-offset)`，`ListView/GridView` 依赖 Gio list ops，Select/Tabs 继承各自父滚动 | 建立显式回归：scroll 后不移动鼠标，当前坐标点击/hover 应命中新目标；诊断记录 hover target 切换 | 当前代码没有 FluxUI 通用 hit-test cache，也没有主动 synthetic pointer move；短期只能以测试锁定 Gio 行为 |
| SH-09 | overlay 内滚动 protected rect | Select popup 的 fieldRect/popupRect 与 ListView option hit area 分别计算 | 下拉类 overlay 复用 region 计算，并在滚动/重排后从当前帧 rect 更新 outside press 保护区 | Dialog/Popup mask close 不是同一模型，不能和 Select protected rect 混合 |

### 底层滚动模型候选边界

| 层级 | 建议职责 | 应覆盖 | 不应覆盖 |
| --- | --- | --- | --- |
| `scrollModel` | 主轴、content size、viewport size、pixel offset、clamp、fraction、ScrollRef command、redraw reason | `ScrollView`、横向 tabs/chips/code/table、未来非虚拟滚动容器 | List/Grid 的 item builder、tab indicator、Select option action |
| `virtualScrollModel` | Gio `layout.List.Position` 或等价虚拟化 position、visible range、reach-end、viewport diagnostics | `ListView`、`GridView`、Select/Menu 内部长列表 | 普通 ScrollView 的整块 child 测量 |
| `wheelPolicy` | 主轴/交叉轴过滤、DeltaX/DeltaY/shift-wheel、preventDefault、剩余 delta 与父级传递 | ScrollView、ListView/GridView、PointerArea wheel listener 协调 | 业务组件的 onChange/onClick 语义 |
| `hitRefreshContract` | scroll 后下一帧 ops/filter 重建、hover target 诊断、无 pointer move 点击当前目标的测试口径 | ScrollView、ListView、GridView、Select menu、Tabs horizontal scroll | 自己实现完整 hit-test engine |
| `scrollableOverlayRegion` | overlay 内滚动区域、trigger/popup protected rect、deferred 绘制位置更新 | Select、DropdownMenu、Menu submenu | Dialog/Popup 遮罩关闭 |

## 风险

- `ScrollView` 已经实现 FluxUI wheel event + preventDefault + pixel offset；`ListView/GridView` 依赖 Gio `layout.List` 内建滚动，缺少同等的 FluxUI default action gate。后续如果只修一边，会继续出现 wheel 可取消性不一致。
- 横向滚动策略集中在 `ScrollView`，但 `ListView` 也支持 horizontal axis。当前没有统一处理 shift-wheel、双轴 touchpad、纵向 delta 是否转横向的策略入口，Tabs/chips/code/table 的行为容易漂移。
- `ScrollView` 的 `ScrollOnChange` 使用 `first + offset/1024` 这样的兼容输出，而 `ListView/GridView` 没有同一套 offset callback；后续统一模型时需要先冻结旧 API 兼容输出，避免破坏 A1.4/A6.1 的承诺。
- 滚动后 hit refresh 仍主要依赖 Gio 每帧 ops/filter 重建；FluxUI 没有显式 hit refresh API，也没有 synthetic pointer move。Select/Menu 等 overlay 内部滚动场景仍缺直接回归测试。
- Select menu 同时跨越 overlay protected rect、deferred 绘制、ListView 虚拟化、Clickable option row；若底层滚动模型和 overlay region 分开修，容易产生 outside click 正确但 option hit 过期，或 option hit 正确但 outside click 误关的组合问题。

## 验收

- 已覆盖 ScrollView、ListView、Grid/GridView、Select menu、Tabs horizontal scroll、PointerArea wheel 的 offset、wheel、axis、hit refresh 相关路径。
- 已确认当前存在两套滚动实现：`ScrollView` 的 FluxUI pixel-offset 模型与 `ListView/GridView` 的 Gio virtual list position 模型。
- 已输出 SH-01 到 SH-09 的 offset、wheel、axis、hit refresh 重复逻辑清单。
- 已给出 `scrollModel`、`virtualScrollModel`、`wheelPolicy`、`hitRefreshContract`、`scrollableOverlayRegion` 五个后续收敛候选边界。
- 本轮只记录审查结果，不修改 widget/event/runtime 行为。

## 后续依赖

- A6.1/A6.2/A6.3：统一滚动模型前应复用既有 offset、wheel 父子传递、横向滚动策略审查结论，尤其是旧 `ScrollOnChange` 兼容输出。
- A6.4：需要把 scroll 后无 pointer move 的 hover/click 命中刷新补成自动或手动回归测试，覆盖 Select/Menu/ListView/Tabs。
- A10.4：集合控件族回归应作为 `scrollModel`、`virtualScrollModel`、`wheelPolicy` 收敛后的主入口。
- A12.1/A12.2：高频 wheel/scroll 和大组件树 benchmark 应分别区分 `ScrollView` 自建 wheel path 与 `ListView/GridView` Gio list path。
- A13.4：最终修复排序应把 wheel policy 和 hit refresh contract 放在高优先级，因为它们影响 ScrollView、ListView、GridView、Select menu、Tabs 和 docs browser 横向区域。
