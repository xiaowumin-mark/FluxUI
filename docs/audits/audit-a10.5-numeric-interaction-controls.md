# A10.5 数值交互控件族审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A10.5 数值交互控件族审查。

## 事实结论

审查范围覆盖 `Slider`、`RangeSlider`、`ProgressBar`、`LinearProgressIndicator`、`CircularProgress`、`CircularProgressIndicator` 和 `LoadingIndicator`。可交互数值输入集中在 `widget/slider.go`；Progress 族集中在 `widget/progress.go`，当前是只读进度展示，不注册 pointer、keyboard、focus 或 Ref 命令。

| 控件/族 | pointer drag | capture | keyboard step | value clamp | 竖向父滚动 / 横向拖拽边界 |
| --- | --- | --- | --- | --- | --- |
| `Slider` | 在 `trackRect` 内注册 Gio pointer target，处理 `Press/Move/Drag/Release/Cancel/Enter/Leave`。press 后选择 end handle，按 X 坐标更新 progress；move/drag 期间要求 primary button 或 Gio `Drag`。 | `PointerDown` 默认行为未被取消时调用 runtime `SetPointerCapture(pointerID, ctx.PathID())`；`PointerUp/PointerCancel` 释放 capture。后续 pointer event 分发先查 runtime capture target。 | 当前没有 `RegisterFocusTarget`、`FocusActivate` 或 keydown listener；键盘 step 只存在命令式 `SliderRef.StepBy(delta)`，不是用户键盘默认行为。 | `normalizeSliderConfig` 和 `applySliderStep` 对输入 value、Ref set/step、最终拖拽值都执行 min/max clamp；`step <= 0` 时只 clamp 不量化。 | Slider 不注册 wheel/scroll filter；竖向 wheel 应由父滚动容器处理。横向拖拽只来自 pointer press/move/drag，不消费 wheel delta。 |
| `RangeSlider` | 复用 `sliderWidget`，press 时按当前位置选择最近 handle；start/end handle 分别更新 progress。重叠 handle 且首次拖动可 flip 到另一端。 | 与 `Slider` 相同，capture owner 是当前 slider PathID，不区分 start/end handle。 | 与 `Slider` 相同，没有键盘 step；`SliderRef.StepBy/SetValue` 只作用于 `valueEnd`，没有 start handle Ref 命令。 | `valueStart/valueEnd` 归一化后保证 `start <= end`；拖拽时 start 不能越过 end，end 不能低于 start，重叠 flip 是例外入口。 | 与 `Slider` 相同，不注册 wheel；横向值变化来自 pointer X 轴。 |
| `ProgressBar` / `LinearProgressIndicator` | 无。只绘制 linear progress 和 buffer。 | 无。 | 无。 | `progressRatio(value, min, max)` 把展示 progress clamp 到 0..1；buffer 也 clamp，并且 buffer 小于 progress 时提升到 progress。 | 不注册输入事件，不应影响父滚动或拖拽。 |
| `CircularProgress` / `CircularProgressIndicator` / `LoadingIndicator` | 无。只绘制 circular arc 或 indeterminate animation。 | 无。 | 无。 | determinate 使用 `progressRatio` clamp 到 0..1；indeterminate 忽略外部 value，按 frame time 计算动画 progress。 | 不注册输入事件；indeterminate 会请求 `animation.progress` redraw，但不改变父滚动事件边界。 |

Slider 交互链路可以归纳为：

1. 每帧先从外部 value/options 建立 `cfg`，再消费 `SliderRef` 命令，命令结果只写入本帧局部 cfg。
2. 未 pressed 时，内部 `sliderState.start/end` 从 cfg 重建；pressed 时保留拖拽 progress，避免外部 value 立即抢占正在拖动的视觉状态。
3. `registerSliderPointer` 在 `trackRect` clip 内注册 Gio event tag 和 pointer cursor，disabled 时直接跳过。
4. `processSliderPointerEvents` 把 Gio pointer 转成 FluxUI pointer event，先走 capture/target/bubble 分发；若 `PointerDown` 默认被取消，则不会启动 slider 拖拽和 capture。
5. 拖拽位置只读取 X 轴，经过 `sliderXToProgress`、`applySliderStep` 和 range 边界规则更新内部 progress。
6. 布局末尾把 progress 还原为 stepped value；非 disabled 且与 cfg value 差异超过 epsilon 时触发 `SliderOnChange` 或 `SliderOnRangeChange`。

## 风险

- `Slider` 当前没有 focus target 和键盘默认 step；A10.5 的 keyboard step 验收只能确认“未实现用户键盘 step”，不能承诺 Arrow/Home/End 等按键行为。
- `RangeSlider` 没有直接测试覆盖；start/end handle 选择、重叠 flip、越界 clamp 和 `SliderOnRangeChange` 需要补回归测试。
- `SliderRef` 对 `RangeSlider` 只控制 end value；如果后续文档或 API 承诺双端命令，需要新增 start/end 明确命令或独立 Ref。
- Slider pointer capture 只有在 `PointerDown` 未被 `PreventDefault` 取消时生效；外部 listener 取消 pointerdown 会阻止拖拽，这是可取消默认行为的一部分，需要文档化。
- Slider 不注册 wheel 是父滚动友好的边界；如果未来为了横向滚轮调值而增加 wheel handler，必须同时设计 delta 方向、preventDefault、父子滚动剩余 delta 和可取消性。
- Progress 族是只读 visual indicator；不应把它纳入输入控件或 Ref/OnChange 语义。indeterminate progress 会持续请求动画 redraw，需要在 perf 审查中继续保留基线。

## 验收

- 已建立 `Slider`、`RangeSlider`、Progress 族的 pointer drag、capture、keyboard step、value clamp 和父滚动边界矩阵。
- 已确认 `Slider` / `RangeSlider` 横向拖拽来自 pointer press/move/drag，不消费 wheel；竖向父滚动不会被 slider 自身 wheel listener 截停。
- 已确认 Slider pointer capture 的生效条件：pointerdown 先经过 FluxUI event 分发，未被取消才设置 capture；release/cancel 时释放。
- 已确认 value clamp 覆盖外部 value、Ref 命令、拖拽 progress 和最终 OnChange 比较值。
- 已确认 Progress 族无输入事件、无 capture、无 keyboard step、无 Ref/OnChange，只按 value 或 animation time 绘制。
- 已标出键盘 step 缺口、RangeSlider 测试缺口和 RangeSlider Ref 只控制 end value 的边界。

## 后续依赖

- A5.5 default action 可取消性矩阵：Slider pointerdown 被取消时不应启动默认拖拽和 capture。
- A6.2 / A6.3：Slider 不注册 wheel 是父级滚动 pass-through 的当前依据；未来新增 wheel 调值必须更新父子滚动策略。
- A7.3：若新增 Slider 键盘行为，需要定义 Arrow/Page/Home/End 默认行为和可取消规则。
- A9.1 / A9.2 / A9.3：SliderRef、controlled value、OnChange 触发条件仍是数值控件回归依据。
- A12.3：组件回归矩阵应补 `RangeSlider` 双 handle、重叠 flip、取消 pointerdown、disabled、step clamp、Progress determinate/indeterminate 绘制测试。
