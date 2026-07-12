# ADR-0004：字段状态是宿主控制的展示快照

- 状态：Accepted（R0 文档合同）
- 决策日期：2026-07-12
- 适用范围：FormField、ValidationSummary、TextField、NumericField、Select/Combobox 等表单控件。

## 背景

字段的 value、校验错误、pending、required、disabled 和用户交互焦点来自不同职责。若 FormField 在 Layout 中校验、发请求或直接改业务值，会破坏 FluxUI 的帧模型，也使受控更新与 Ref 命令无法预测。

## 决定

1. FormField 只提供字段语义和布局：label、required 标记、control 容器、supporting/error text、pending 表达及与 ValidationSummary 的可定位关联。它不拥有业务数据。
2. value、error 列表/文本、pending、disabled、read-only 等均是宿主传入的同步快照。异步校验、提交和网络 I/O 由宿主执行，再在后续 frame 提供新快照。
3. 业务值的优先级固定为：外部受控 value > 允许的 Ref 命令结果 > 用户输入意图。FormField 不从 label、显示文本或旧 index 推导业务值，也不自行调用 `OnChange`。
4. 错误是校验状态而非焦点状态：有 error 时保持可聚焦/可编辑性（除非 disabled/read-only 另有规定），同时展示 invalid 语义和 error text。pending 可以与 error 共存，不能悄悄吞掉错误。
5. required 仅声明要求，不自动把空值变成错误；touched/visited 是可选交互信息，必须由具体 API 明确声明，不能作为隐式网络校验开关。
6. 字段、label、supporting/error 文本应使用同一稳定 field identity 建立语义关联；重排后关联按 key 保持，不能按 index 漂移。

## 合同与不变量

- Layout 中不得验证、提交、读文件、发网络请求或启动 goroutine。
- `OnChange` 只表示真实用户值变化；程序化受控值同步和默认 Ref 命令不伪装为用户输入，除非该组件的公开文档明确例外。
- disabled/loading 是入口阻断，不等价于 `PreventDefault`；disabled 字段不接收修改意图，且 Ref 不得绕过该规则。
- error、pending、supporting text、长 label 与窄宽度只能改变视觉/辅助布局，不得隐式改变父级 constraints 或触发持续 redraw。
- ValidationSummary 只汇总宿主给出的字段快照与定位意图；不扫描组件树猜测业务字段，也不负责提交。

## 内部测试夹具

为字段试点提供可在连续 frame 更新受控快照的 fixture，记录用户意图、Ref 命令、语义 ID 与渲染状态。最低用例：

| 场景 | 断言 |
| --- | --- |
| 外部 value 同帧更新 | 外部值获胜，`OnChange` 不重复、不回写旧值。 |
| error/pending 组合 | 错误不丢失；pending 只有已声明的展示效果。 |
| required 与空值 | required 本身不产生未声明的校验/提交副作用。 |
| disabled/read-only/Ref | 用户输入和 Ref 均遵守各自合同，不绕过入口阻断。 |
| 重排/条件渲染 | label、描述、error 仍关联同一 field key。 |
| 长文本、窄宽度、DPI | 无 panic、无遮挡关键错误文本，视觉矩阵有记录。 |
| 宿主异步快照 | fixture 只注入前后同步状态；证明 Layout 没有发起异步工作。 |

## 非目标

- 本 ADR 不提供表单数据仓库、校验规则语言、网络校验器、自动提交或完整读屏实现。
- 本 ADR 不定义 NumericField 的数值类型，也不替代日期、文件等业务值合同。
- 本 ADR 不要求所有现有 TextField 立即套用 FormField。

## 回滚

FormField/ValidationSummary 首先以局部 widget adapter 实现，字段快照保持宿主拥有。若视觉或回调语义回归，可撤销该 adapter；既有 Input/TextField 的受控值路径和事件顺序不变。

## R0 文档完成记录

- 2026-07-12：字段职责、受控优先级、error/pending 组合和测试夹具边界已冻结。
- 实现完成的最低证据：FormField 与另一字段类组件复用 field-state helper，并通过同帧受控、disabled/Ref、重排与视觉状态测试。
