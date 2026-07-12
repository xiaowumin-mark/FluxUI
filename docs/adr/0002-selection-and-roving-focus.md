# ADR-0002：选择模型与 roving focus 分离

- 状态：Accepted（R0 文档合同）
- 决策日期：2026-07-12
- 适用范围：Select、Combobox、Autocomplete、多选 Select、TagInput、DataGrid、TreeView 与菜单型集合。

## 背景

“已选择的业务值”和“当前可由键盘操作的项”不是同一状态。把二者都保存为 index，会同时引入重排串位、禁用项落焦和 overlay 关闭后焦点丢失。R1 的可搜索选择能力也需要在筛选结果变化时保持一个可解释的 active item。

## 决定

1. 选择是受控业务状态：组件消费宿主提供的 selected key/value 快照，并通过意图回调请求单选、切换、范围选择或清除；组件不在 Layout 中暗改业务选择。
2. roving focus 是内部瞬态交互状态：它保存 `activeKey`，而不是 selected index。选择与 active 可相同，也可以不同。
3. 对一组当前可聚焦项，最多一个项获得 roving tab stop（等价于 tab stop `0`），其余项为 `-1`；没有可聚焦项时不存在 roving tab stop。
4. active item 被删除、禁用或过滤掉时，按视觉顺序寻找下一个可聚焦项；不存在下一个时找上一个；仍不存在则清空 active。禁用项仍在当前模型中，因此直接按当前顺序查找；删除项由 controller 在其上一次有效 reconciliation 中记录的稳定 key 邻接序列恢复，先 following、再 preceding。该私有快照不保存或推断 index；从未在模型中出现过的非空 key 没有可证明的位置，必须清空，而空 active key 可初始化为首个可聚焦项。不得保留指向已经不可交互项的焦点。
5. Arrow/Home/End 只移动 active item；Enter/Space 是否提交选择由具体组件的默认行为定义并可测试。Tab/Shift+Tab 离开 roving group，不把每个项都加入全局 Tab 链。
6. typeahead、分页移动、范围选择和 `aria` 等完整语义不是通用模型的隐式副作用；具体组件必须在文档中声明并测试。

## 合同与不变量

- 选择、active、hover、pressed 和滚动位置分别保存，任何一种都不能由另一种的 index 推断。
- disabled/hidden 项不会成为 active，也不会因 pointer 或 keyboard 被选择；宿主传入的无效 selected 值保持为受控快照，组件只报告可观测的不可用状态，不自行替换业务值。
- 同一帧发生外部 selected 更新、用户输入和 Ref 命令时，外部受控值优先；Ref 不绕过 disabled/controlled 规则；用户意图最多发出一次。
- Overlay 打开时，可按组件合同把实际焦点移入 active item 或保留在 trigger；关闭时由 ADR-0005 的 focus restore 规则处理。不能同时把同一物理焦点注册为两个项。
- 单选/多选、树/表格的范围选择策略由组件层定义；共享模型只处理稳定 key 和有效项遍历。

## 内部测试夹具

在 `widget` 测试中提供能够用 key 构造有序项、注入 keyboard/pointer frame、读取 active 和选择意图的 fixture。最低用例：

| 场景 | 断言 |
| --- | --- |
| Arrow、Home、End | 只移动 activeKey，跳过 disabled/hidden 项，边界不越界。 |
| 单选/多选激活 | Enter/Space 的选择意图符合组件合同且只触发一次。 |
| 重排与筛选 | active/selected 仍按 key 关联；active 不存在时按相邻规则回退。 |
| 删除 active 项 | 优先下一个、再上一个、最后清空，不留下失效 tab stop。 |
| 同帧受控更新/Ref | 外部快照获胜，不产生双重 `OnChange`。 |
| Overlay 开关 | 进入、关闭、重新打开时的 active 与实际 focus 均可断言。 |
| pointer 与 keyboard | 两条输入路径产生同一选择意图，但不会把 hover 当作选择。 |

测试应同时覆盖一个反例：把 active/selected 按 index 保存的实现，在重排后必然失败，从而防止回归到 index 模型。

## 非目标

- 本 ADR 不把现有 `Select` 变为搜索输入，也不定义 Combobox/Autocomplete 的公开 API。
- 本 ADR 不承诺完整读屏支持、跨窗口焦点、全局快捷键或所有平台的 IME 行为。
- 本 ADR 不要求把现有所有静态列表迁移为 roving focus；只有复合、可键盘操作的集合使用它。

## 回滚

roving controller 仅由试点组件通过内部 adapter 使用。出现键盘回归时，回滚该 adapter 和局部测试，不改写 runtime 的全局 focus 管理或已稳定的 KeyboardScope 行为。

## R0 文档完成记录

- 2026-07-12：受控选择与 active focus 的职责、失效回退和 keyboard fixture 已定义。
- 实现完成的最低证据：至少两个复合组件复用 controller，并通过本 ADR 的 pointer、keyboard、重排和同帧冲突用例。
