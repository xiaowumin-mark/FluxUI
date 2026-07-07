# A7.4 文本输入事件审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 4：键盘、焦点和可访问性审查。

- 状态：Done
- 日期：2026-07-07
- 负责人：Codex
- 关注：Event、Widget
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "### A7\.|A7\.4|beforeinput|TextField|SearchBar|InputRef" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md docs/audits/project-audit-baseline.md`
  - `rg -n "type .*Input|InputRef|TextField|SearchBar|beforeinput|BeforeInput|InputEvent|OnChange|Submit|SetText|Paste|Undo|Redo|programmatic" widget event internal examples docs -g "*.go" -g "*.md"`
  - `go test ./event ./internal ./widget ./examples/event_system_testbench ./examples/docs_browser`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `event/text.go`
  - `event/text_test.go`
  - `internal/events.go`
  - `widget/input.go`
  - `widget/input_event_test.go`
  - `widget/component_refs.go`
  - `widget/material3_components.go`
  - `ui/ui.go`
  - `ui/extended_types.go`
  - `examples/event_system_testbench/main.go`
- 关联能力：
  - `beforeinput`、`input`、`change`、`submit` typed event 链路
  - `InputRef.SetText/Append/Clear` programmatic source 标记
  - 用户输入、粘贴、删除、撤销/重做 best-effort 推断
  - TextField/SearchBar 对 InputOption 的复用边界
  - composition synthetic dispatch 与真实 IME 生命周期边界

## 事实结论

1. A7.4 的目标是审查 Input、TextField、SearchBar、InputRef 的文本输入事件，并输出 `beforeinput`、`input`、`change`、`submit`、programmatic source 规则；验收要求是用户输入、程序设置、粘贴、删除、撤销重做可区分。证据：`docs/project-audit-task-breakdown.md:362`、`docs/project-audit-task-breakdown.md:365`。
2. `event.InputEvent` 已公开 `Data`、`InputType`、`IsComposing`、`Source`、`Value`、`PreviousValue`、`BestEffort`、`Native` 字段，并定义 `user`、`programmatic`、`paste`、`delete`、`undo`、`redo` 来源常量。证据：`event/text.go:15`、`event/text.go:18`、`event/text.go:42`。
3. 文本事件类型包括 `beforeinput`、`input`、`change`、`submit` 和 composition 三段事件；`beforeinput` 与 `submit` 默认 cancelable，`input`/`change`/composition 默认不可取消。证据：`event/text.go:5`、`event/text.go:124`、`event/text.go:130`。
4. TextField 系列共享 `inputWidget` 实现，`InputOnBeforeInput`、`InputOnInputEvent`、`InputOnSubmit` 会在布局开始注册到当前 PathID 的事件监听器；旧 `InputOnChange` 仍保留为 value callback。证据：`widget/input.go:324`、`widget/input.go:336`、`widget/input.go:342`、`widget/input.go:348`、`widget/input.go:373`。
5. 受控 TextField 初始化和外部 value 同步使用 `editor.SetText`，但这条路径不会发出 `InputEvent`；只有 `InputRef` 命令和 Gio user edit 会进入文本事件链路。证据：`widget/input.go:377`、`widget/input.go:382`、`widget/input.go:390`。
6. `InputRef.SetText/Append/Clear` 在下一帧 drain 命令后进入 `commitProgrammaticInput`，分别标记 `programmaticSetText`、`programmaticAppend`、`programmaticClear`，`Source=programmatic`。`Focus/Blur` 只执行 Gio focus 命令，不产生文本事件。证据：`widget/component_refs.go:79`、`widget/component_refs.go:87`、`widget/component_refs.go:95`、`widget/input.go:405`、`widget/input.go:409`、`widget/input.go:415`。
7. programmatic 文本变更同样先派发 cancelable `beforeinput`；若被取消，不调用 `editor.SetText`，也不会继续派发 `input/change/InputOnChange`。证据：`widget/input.go:572`、`widget/input.go:575`。
8. Gio 用户编辑只从 `gioWidget.ChangeEvent` 后置进入 `commitUserInput`，因此 FluxUI 的 `beforeinput` 是 mutation-after 的 best-effort 回滚：被取消时把 editor 文本设回 `PreviousValue`。证据：`widget/input.go:423`、`widget/input.go:429`、`widget/input.go:585`、`widget/input.go:591`、`widget/input.go:592`。
9. 用户编辑被接受后，事件顺序是 `beforeinput` -> `input` -> `change` -> 旧 `InputOnChange`；`input` 与 `change` 都通过同一个 `finishInputMutation` 派发，旧 callback 最后执行。证据：`widget/input.go:597`、`widget/input.go:600`、`widget/input.go:604`、`widget/input.go:606`、`widget/input.go:608`。
10. `submit` 来自 Gio `SubmitEvent`，只有配置 `InputOnSubmit` 时才启用 `editor.Submit`；提交事件 `InputType=insertLineBreak`、`Source=user`、`Trusted=true`、`BestEffort=false`，但当前实现不根据 `submit` 是否被取消回滚或阻止其他内建行为。证据：`widget/input.go:397`、`widget/input.go:434`、`widget/input.go:613`、`widget/input.go:618`。
11. 用户输入来源区分依赖 `classifyUserInput(previous, value)` 的 diff/history 推断：删除为 `Source=delete`，疑似粘贴为 `Source=paste`，命中本地 valueHistory 前后项时标记 `undo/redo`。证据：`widget/input.go:678`、`widget/input.go:686`、`widget/input.go:689`、`widget/input.go:692`、`widget/input.go:694`。
12. 粘贴判断是启发式：插入内容包含换行、回车、Tab 或长度大于 1 即视为 paste。连续 IME 提交、多字符自动替换或一次性输入可能被标成 paste；单字符粘贴可能仍被标成 user insertText。证据：`widget/input.go:773`、`widget/input.go:774`、`widget/input.go:777`。
13. undo/redo 判断只基于 FluxUI 自维护的 `valueHistory`，不直接读取 Gio editor 的 undo 栈或平台编辑命令类型；如果外部受控 value 重置历史，后续撤销/重做识别会以重置后的历史为准。证据：`widget/input.go:701`、`widget/input.go:710`、`widget/input.go:718`。
14. Composition API 具备 synthetic `OnComposition`/`DispatchCompositionEvent` 入口，但真实 Gio IME 生命周期没有被 Input 自动桥接；`InputEvent.IsComposing` 在现有 Input 路径中未设置为 true。证据：`event/text.go:60`、`event/text.go:89`、`event/text.go:114`、`widget/input.go:627`。
15. SearchBar 不独立实现文本事件。它把 `SearchBarInputOptions` 追加到内部 `TextField`，再追加 SearchBar 自己的 placeholder/disabled/样式选项，并将 `SearchBarOnChange` 转成 `InputOnChange`；因此高级事件要通过 `SearchBarInputOptions(InputOnBeforeInput(...), ...)` 接入。证据：`widget/material3_components.go:2200`、`widget/material3_components.go:2249`、`widget/material3_components.go:2269`、`widget/material3_components.go:2280`、`widget/material3_components.go:2287`。
16. `TextFieldElement`、`OutlinedTextFieldElement`、`FilledTextFieldElement` 和 `SearchBarElement` 都只是 `FromWidget(...)` 包装；事件语义由底层 widget 维护。证据：`ui/extended_types.go:474`、`ui/extended_types.go:475`、`ui/extended_types.go:479`、`ui/extended_types.go:483`、`ui/extended_types.go:746`。
17. 自动测试覆盖了 programmatic `SetText` 事件、`beforeinput` 取消回滚、被接受用户输入的 `input/InputOnChange` 顺序，以及 `submit` 事件；未看到专门覆盖 paste/delete/undo/redo 分类、SearchBar 高级事件转发和真实 IME composition 的测试。证据：`widget/input_event_test.go:20`、`widget/input_event_test.go:65`、`widget/input_event_test.go:143`、`event/text_test.go:9`、`event/text_test.go:42`。

### 文本事件规则矩阵

| 来源/动作 | 入口 | 事件顺序 | Source | InputType | Cancelable | 备注 |
| --- | --- | --- | --- | --- | --- | --- |
| 用户普通输入 | Gio `ChangeEvent` 后置 diff | `beforeinput` -> `input` -> `change` -> `InputOnChange` | `user` | `insertText` 或 `insertReplacementText` | 仅 `beforeinput` 可取消 | `BestEffort=true`，取消靠回滚 |
| 用户删除 | Gio `ChangeEvent` 后置 diff | 同上 | `delete` | `deleteContentBackward` | 仅 `beforeinput` 可取消 | 未区分 forward/backward/deleteByCut |
| 粘贴 | Gio `ChangeEvent` 后置 diff | 同上 | `paste` | `insertFromPaste` | 仅 `beforeinput` 可取消 | 多字符/换行/Tab 启发式 |
| 撤销 | Gio `ChangeEvent` 后置 diff + valueHistory | 同上 | `undo` | `historyUndo` | 仅 `beforeinput` 可取消 | 依赖 FluxUI history 命中 |
| 重做 | Gio `ChangeEvent` 后置 diff + valueHistory | 同上 | `redo` | `historyRedo` | 仅 `beforeinput` 可取消 | 依赖 FluxUI history 命中 |
| `InputRef.SetText` | ref command drain | `beforeinput` -> `input` -> `change` -> `InputOnChange` | `programmatic` | `programmaticSetText` | `beforeinput` 可取消 | `input/change` 标为不可取消 |
| `InputRef.Append` | ref command drain | 同上 | `programmatic` | `programmaticAppend` | `beforeinput` 可取消 | data 为 append 文本 |
| `InputRef.Clear` | ref command drain | 同上 | `programmatic` | `programmaticClear` | `beforeinput` 可取消 | data 为空 |
| Submit | Gio `SubmitEvent` | `submit` | `user` | `insertLineBreak` | `submit` 可取消 | 当前取消只影响 listener 可见状态 |
| 受控 value 同步 | 外部 value 与 `syncedValue` 不同 | 无文本事件 | 无 | 无 | 无 | 视为 props 同步，不是 programmatic source |
| Composition synthetic | `DispatchCompositionEvent` | composition* | 无 `InputSource` | 无 | 不可取消 | 真实 IME 未自动桥接 |

### 组件覆盖矩阵

| 组件/API | 事件能力 | 旧 API 兼容 | 结论 |
| --- | --- | --- | --- |
| `TextField` / `OutlinedTextField` / `FilledTextField` | 完整复用 `inputWidget` 的 beforeinput/input/change/submit/ref 规则 | `InputOnChange`、`InputOnFocus` 保留 | 正式入口 |
| `TextFieldElement` 等 Element wrapper | `FromWidget` 包装，语义不变 | 复用 `InputOption` | 正式入口，但事件归属仍在 widget |
| `SearchBar` | 通过 `SearchBarInputOptions` 转发高级 `InputOption` | `SearchBarOnChange` 转成 `InputOnChange` | 高级事件不是独立 SearchBar option |
| `InputRef` | `SetText/Append/Clear` 产生 programmatic 文本事件；`Focus/Blur` 不产生文本事件 | ref 命令下一帧消费 | 正式入口 |

## 风险

1. `beforeinput` 对用户编辑是后置回滚，不是浏览器式 mutation-before 拦截；对大文本、复杂 IME 或后端 selection 状态，回滚可能带来 caret/selection 与用户预期不一致的风险。
2. paste/delete/undo/redo 的来源识别是 best-effort 推断，不是 Gio 原生命令类型映射；后续文档和测试不能把它描述为平台级精确分类。
3. `submit` 虽然 cancelable，但当前没有默认行为可阻止；如果后续增加表单提交或 newline 行为，需要重新定义 `PreventDefault` 对 submit 的实际效果。
4. 受控 value 同步不产生文本事件，与 `InputRef.SetText` 的 programmatic source 不同；调用方如果用 state prop 强制改值，不会被 `InputOnInputEvent` 观察到。
5. SearchBar 的高级事件入口藏在 `SearchBarInputOptions`，容易被误认为 `SearchBarOnChange` 等价于完整 InputEvent API。
6. 真实 IME composition 尚未自动桥接，`IsComposing` 默认 false；依赖 composition 生命周期的编辑器场景仍只能使用 synthetic 测试入口或等待后端能力补齐。

## 验收

- 已确认 `beforeinput`、`input`、`change`、`submit` 的事件顺序、可取消性和旧 `InputOnChange` 桥接顺序。
- 已区分用户输入、程序设置、删除、粘贴、撤销、重做的来源字段；其中删除/粘贴/撤销/重做明确标为 best-effort。
- 已确认 `InputRef.SetText/Append/Clear` 是 programmatic source；`Focus/Blur` 不是文本输入事件。
- 已确认 TextField/SearchBar/Element wrapper 的语义归属：TextField 维护核心语义，SearchBar 通过内部 TextField 转发。
- 已记录受控 value 同步和真实 IME composition 的边界，避免后续把它们误判为审查引入问题。

## 后续依赖

- A7 后续任务若审查 IME、表单提交或可访问性，需要复用本文件的 `beforeinput` best-effort 和 `IsComposing=false` 基线。
- 后续修复如要提升 paste/delete/undo/redo 精度，应优先查 Gio 是否能暴露原始 edit command/source，而不是继续扩大 diff 启发式。
- 若 SearchBar 需要一等高级事件 option，可在不改变现有 `SearchBarInputOptions` 的前提下增量增加别名 API。
- 自动测试建议补齐 paste/delete/undo/redo 分类、SearchBar 高级事件转发和 submit `PreventDefault` 行为约定。
