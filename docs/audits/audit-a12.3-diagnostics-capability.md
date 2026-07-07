# A12.3 diagnostics 能力审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A12.3 diagnostics 能力审查。

## 事实结论

A12.3 的目标是审查当前 runtime/event/perf diagnostics 是否能支撑定位“谁注册事件、谁取消默认行为、谁触发 redraw”。当前实现已经有可用的 frame 级和事件级聚合入口，但归因能力仍是粗粒度。

### 现有诊断入口

| 能力 | 当前入口 | 已记录字段 | 结论 |
| --- | --- | --- | --- |
| frame diagnostics 开关 | `Runtime.SetPerfDiagnostics` / `PerfDiagnostics` | `Enabled`、`MeasureDurations`、`LogRedrawReasons`、`LogEvents`、`Writer` | 可按 runtime 打开聚合统计、耗时测量和日志输出 |
| frame snapshot | `Runtime.LastFrameStats` / `ui.FormatFrameStats` | frame、duration、interaction、event、virtualization、cache、layout/draw/state/text/input、reason | 可复现每帧粗粒度运行情况 |
| redraw reason | `RecordRedrawReason` / `RequestRedrawReason` / `Context.RequestRedrawReason` / `Context.WindowInvalidate` | `ReasonCounts`、排序后的 `Reasons`、格式化 `reason=` | 能知道 redraw 原因字符串和次数，但不能知道调用点或组件 owner |
| interaction diagnostics | `InteractionFrameStats` / `ObserveInteractionSnapshot` / `ObserveEventDispatch` | pointer/wheel/keyboard/focus 计数、hover/pressed/focus 变化、hover target | 能知道输入类别、变化数量和 hover target，不能定位具体 listener 或 widget API |
| event diagnostics | `EventDiagnosticsStats` / `recordEventDispatch` | dispatch 数、listener 调用数、listener duration、defaultPrevented、propagation stopped、事件分类、last type/target/path | 能看见事件路径和默认行为结果；只有 `LogEvents` 开启时记录 last target/path |
| event log | `PerfDiagnostics.LogEvents` + `Writer` | `event type`、`target`、`path`、`listeners`、`listener_duration`、`default_prevented`、`propagation_stopped`、`default_allowed` | 可输出单次 dispatch 的紧凑日志，适合手动定位路径 |

### event path 与注册表边界

| 项 | 当前规则 | diagnostics 可见性 | 缺口 |
| --- | --- | --- | --- |
| target 注册 | `beginEventFrame` 每帧清空 `targets/listeners/focusTargets/shortcuts`，保留 `pointerCaptures/keyDown/focusTarget` 等跨帧状态 | 只通过 dispatch path 间接可见 | 无 target 总数、注册 owner、注册来源文件/组件名 |
| listener 注册 | `RegisterEventListener` 按当前 `ctx.pathID` 追加 listener，记录 type/options/context/seq，并按 priority/seq 排序 | dispatch 后只有 listener 调用总数和总耗时 | 无 listener 清单、capture/passive/once/priority 分布、具体 handler 归因 |
| event path | `eventPath` 从 target 到 root 构造 path，遇到 `EventBoundaryStop` 截断，遇到 `EventBoundaryRedirect` 改写 parent | `LogEvents` 输出 readable path，`LastFrameStats.Events.LastPath` 只保留最后一次 | 无每次事件历史列表；未记录 boundary/portal 改写原因 |
| default action | `DispatchEvent` 在 listener 后计算 `allowed := !(Cancelable && DefaultPrevented)`；键盘默认行为在 `DispatchKeyboardEvent` 中受 allowed gate 控制 | 记录 `DefaultPrevented` 次数和 `default_allowed` | 无“哪个 listener 调用了 PreventDefault”；passive 失败也不可见 |
| stop propagation | `StopPropagation` / `StopImmediatePropagation` 改变 event 内部标志 | 记录 stopped 次数和 immediate stopped 次数 | 无停止发生的 phase/currentTarget/listener |

### redraw reason 来源

| 来源类别 | 现有调用 | 当前 reason | 可定位程度 |
| --- | --- | --- | --- |
| 显式 runtime 命令 | `RequestRedrawReason(reason)` | 调用者传入字符串，空值归一为 `RequestRedraw` | 只能定位到约定字符串 |
| context/window 命令 | `Context.RequestRedrawReason`、`Context.WindowInvalidate` | 传入字符串或 `WindowInvalidate` | 能区分窗口 invalidate，但没有组件 path |
| state 更新 | `state.State.Set` | `state.Set` | 能知道是 state 更新，但不知道哪个 state key 或组件 |
| interaction 变化 | `ObserveInteractionSnapshot`、`endInteractionFrame` | `pointer.hover_changed`、`pointer.pressed_changed`、`pointer.focus_changed` | hover target 可见；pressed/focus change 没有 target 字段 |
| animation/layout/widget | 依赖各调用点传入 reason | 字符串约定不统一 | 需要调用点自律，缺少枚举或结构化字段 |

### 诊断字段缺口清单

| 缺口编号 | 缺失字段 | 当前影响 | 建议方向 |
| --- | --- | --- | --- |
| DG-01 | `event_targets_registered`、`event_listeners_registered`、`focus_targets_registered` | 只能知道 dispatch 后 listener calls，不能知道注册规模 | 在 frame stats 增加 registry size 计数 |
| DG-02 | listener `target/path/type/options/priority/seq` 快照 | 无法回答“谁注册了事件” | 增加可选 listener registry dump，默认关闭 |
| DG-03 | listener handler 语义名或 owner debug label | path 可读但 handler 不可读 | 支持注册时传入 debug name，或由 widget 层设置组件族 label |
| DG-04 | `prevent_default_target`、`prevent_default_phase`、`prevent_default_listener` | 只能知道发生了取消，不能知道谁取消 | 在 `PreventDefault` 成功时捕获 currentTarget/phase/passive 状态 |
| DG-05 | passive preventDefault failed count | passive listener 调用 `PreventDefault` 失败不可见 | 统计 passive 拒绝次数和最后一次 target |
| DG-06 | stop propagation 的 target/phase | 只能知道停止次数 | 在 stop 方法中捕获当前 phase/currentTarget |
| DG-07 | redraw `owner_path`、`reason_kind`、`source` | reason 是自由字符串，归因不稳定 | 引入结构化 redraw record：reason/source/path/count |
| DG-08 | state redraw 的 state key/path | `state.Set` 过粗 | 在 state 层把 `DebugMemoryKey` 或 PathID 作为可选诊断字段 |
| DG-09 | event path boundary/portal reason | path 已改写但不说明原因 | event log 增加 boundary mode、redirect target、modal stop 标记 |
| DG-10 | per-frame event history ring | `LastPath` 只保留最后一次事件 | 增加小容量 ring buffer，避免高频事件下只剩最后一条 |

## 风险

- 当前 diagnostics 足以判断“一帧是否发生过 redraw、发生过多少事件、最后一次事件走了哪条 path、是否 defaultPrevented”，但不足以直接定位具体 widget、listener 或 state key。
- `ReasonCounts` 以自由字符串为 key，调用点命名不一致会导致同一类 redraw 分裂为多个 reason，或多个调用点共享同一 reason。
- `LogEvents` 记录的是 dispatch 结束后的汇总信息；如果 listener 中途修改 event 后又恢复局部状态，日志无法还原 listener 内部过程。
- `LastFrameStats.Events.LastPath` 只有最后一次事件路径；pointermove/wheel 这类高频输入会覆盖更早、更有价值的 click/key/drop 诊断信息。
- listener duration 是总耗时，不区分 capture/target/bubble 和具体 listener；一旦慢 listener 与大量 listener 混在同一帧，现有字段只能提示有问题，不能直接定位。
- `PreventDefault` 在 passive listener 中返回 false，但现有统计只记录最终 `DefaultPrevented`，无法发现“尝试取消但未生效”的误用。
- 增加更细粒度 diagnostics 会引入额外分配和锁竞争，必须保持默认关闭，并限制 ring buffer/registry dump 的容量。

## 验收

- 已确认 `FrameStats`、`EventDiagnosticsStats`、`InteractionFrameStats`、`ReasonCounts` 和 `FormatFrameStats` 当前可输出 frame、event path、target、defaultPrevented、redraw reason 等基础信息。
- 已确认 event path 由 `eventPath` 生成，并受 boundary stop/redirect 与 portal owner 改写影响；`LogEvents` 可输出 readable target/path。
- 已确认 `beginEventFrame` 每帧重建 event target/listener/focus target/shortcut registry，当前 diagnostics 未暴露注册表规模和注册来源。
- 已确认默认行为取消结果可通过 `DefaultPrevented` / `default_allowed` 观察，但取消者、phase、currentTarget 和 passive 失败缺失。
- 已确认 redraw reason 可统计到字符串和次数，但无法稳定定位 owner path、state key 或调用点。
- 已输出 DG-01 到 DG-10 诊断字段缺口清单，后续可据此设计结构化 diagnostics。
- 本轮只记录审查结果，不修改 runtime/event/perf 行为。

## 后续依赖

- A12.1/A12.2：新增 diagnostics 字段后，应同步纳入高频事件和大组件树 benchmark，确认默认关闭时无可见分配热点，开启时成本可控。
- A5.2/A5.5：event dispatch 顺序和 default action 可取消性若调整，应优先补齐 `prevent_default_*`、stop propagation 和 boundary/portal path 诊断。
- A2.3/A2.4：registry 生命周期和 redraw/invalidation 审查的结论应反向约束 DG-01、DG-07 的字段命名和生命周期。
- A6.4/A8.2：scroll 后 hit refresh 与 portal event path 问题需要 event history ring 和 boundary/portal reason 才能稳定复现。
- A13.x：冗余收敛或 diagnostics API 化时，应把自由字符串 reason 收敛为稳定枚举加可选 debug label。
