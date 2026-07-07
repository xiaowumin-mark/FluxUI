# A10.3 文本输入控件族审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A10.3 文本输入控件族审查。

## 事实结论

审查范围覆盖 `Input`、`TextField`、`OutlinedTextField`、`FilledTextField`、`SearchBar`，核心实现集中在 `widget/input.go`，SearchBar 位于 `widget/material3_components.go` 并通过 `Expanded(TextField(...))` 复用输入框。

| 控件 | editor | focus | keyboard | IME / composition | wheel pass-through |
| --- | --- | --- | --- | --- | --- |
| Input / TextField | 每个 path 通过 `inputStateFor` 保存一个 `gioWidget.Editor`，`Layout` 中设置 `SingleLine`、`ReadOnly`、`Submit`、`MaxLen`、`Mask`，由 `material.Editor(...).Layout` 注册 Gio editor ops。 | 使用 Gio `gtx.Focused(editor)` 判断焦点；`InputRef.Focus/Blur` 发送 `key.FocusCmd`；聚焦时注册全窗 outside press blur tag。 | Gio Editor 直接处理文本编辑键；`Submit=true` 时 Gio `SubmitEvent` 转为 FluxUI `submit`。没有额外注册 FluxUI runtime focus target。 | `event/text.go` 定义 `compositionstart/update/end` 与 `IsComposing` 字段，但 `inputWidget` 当前没有从 Gio Editor 显式桥接 composition lifecycle。 | 输入控件自身没有注册 wheel listener 或 pointer scroll filter；outside press blur 使用 `pointer.PassOp` 且只监听 `pointer.Press`，不消费 wheel。多行 Gio Editor 是否内部滚动取决于 Gio Editor 行为，FluxUI 层没有无条件截停父级滚动。 |
| SearchBar | 组合 `TextField`，通过 `SearchBarInputOptions` 透传输入选项。 | 焦点完全继承内部 TextField；外层 `ContainerDecoration` 不注册独立 focus target。 | 键盘编辑与 submit 由内部 TextField 处理；SearchBar 只追加 placeholder、disabled、foreground、padding、decoration、OnChange。 | 继承 TextField 当前的 composition 缺口。 | 外层 SearchBar 容器没有独立 wheel 处理；wheel 语义继承内部 TextField。 |

文本事件链路为：

1. `editor.Update(ctx.Gtx)` 产生 Gio `ChangeEvent` 或 `SubmitEvent`。
2. `ChangeEvent` 触发 `commitUserInput`，先构造 `beforeinput`；若 `PreventDefault` 生效，则回滚到 `state.syncedValue`。
3. 允许后调用 `finishInputMutation`，依次派发 `input`、`change`，最后调用旧 `InputOnChange`。
4. `InputRef.SetText/Append/Clear` 走 `commitProgrammaticInput`，source 为 `programmatic`，也会先派发可取消的 `beforeinput`。
5. `SubmitEvent` 走 `dispatchSubmit`，派发可取消的 `submit`，但当前没有基于 `submit PreventDefault` 执行额外默认行为。

输入来源识别为 best-effort：`inputTextDiff` 计算删除和插入片段，`classifyUserInput` 将多字符或含换行/制表的插入推断为 paste，将历史相邻值推断为 undo/redo。真实 Gio 原生事件没有提供 before-mutation 细节，因此 `beforeinput` 的用户输入取消通过回滚实现。

## 风险

- IME composition 目前只有事件类型和数据结构，没有 Input 层实际桥接；`IsComposing` 对真实 IME 输入不会自动置位，组合中 Enter submit 的边界依赖 Gio Editor 原生行为，FluxUI 层缺少可审计的显式状态。
- Input 不注册 FluxUI runtime focus target，焦点状态来自 Gio Editor；这与 A7.1 中“Input Gio editor focus 与 FluxUI runtime focus 分离”的结论一致，但会影响统一 shortcut/focus path 语义。
- 多行 TextField 由 Gio Editor 自身布局和交互决定；FluxUI 层没有 wheel default action gate，也没有剩余 delta 传递模型。当前审查只能确认没有 FluxUI 级别无条件截停父滚动，不能保证 Gio Editor 内部滚动在所有平台都把未消费 wheel 传给父级。
- `SearchBarInputOptions` 先应用用户传入选项，再追加 SearchBar 默认输入选项；placeholder、disabled、background、foreground、border、padding、decoration 可能覆盖用户意图，事件选项仍可透传。
- `InputOnChange` 被用作 controlled 判定：只有设置 `onChange` 时外部 `value` 变化才会同步覆盖 editor。无 `onChange` 的 TextField 更接近 uncontrolled 初始值语义。

## 验收

- 已建立 Input/TextField/SearchBar 的 editor、focus、keyboard、IME、wheel pass-through 矩阵。
- 已确认 SearchBar 是 TextField 组合包装，不单独实现 editor、focus、keyboard、wheel 语义。
- 已确认输入控件没有 FluxUI 层的 wheel listener、`WheelEvent` dispatch 或 `pointer.Scroll` filter；outside press blur 使用 `pointer.PassOp` 并只监听 press，不会无条件阻挡父级滚动。
- 已标出 IME/composition 真实桥接缺口、Gio Editor focus 与 FluxUI runtime focus 分离、Gio Editor 内部滚动剩余 delta 不可见三个后续风险。

## 后续依赖

- A6.2 / A6.4：若补充嵌套滚动和 scroll 后 hit 刷新测试，输入控件应作为“不会无条件截停父滚动”的回归样例。
- A7.4 / A7.5：composition lifecycle、`IsComposing`、组合中 Enter submit 需要继续沿用既有缺口结论，后续修复应补充 Input 层显式状态。
- A9.2 / A9.3：Input/TextField/SearchBar 的 controlled value 与 `OnChange` 触发规则应保持一致，避免 SearchBar 包装层改变受控语义。
- A10 后续控件族审查：所有包含 TextField 的高阶控件应继承本文件的 editor/focus/keyboard/wheel 边界，除非显式声明新的事件策略。
