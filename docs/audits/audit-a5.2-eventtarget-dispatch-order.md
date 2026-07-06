# A5.2 EventTarget 分发顺序审查
> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 3：事件系统审查。

- 状态：Done
- 日期：2026-07-06 20:09:58 +08:00
- 负责人：Codex
- 关注：Event、Runtime
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A5\\.2|EventTarget|分发顺序|capture|bubble|preventDefault|passive|once|default action" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md docs/audits/project-audit-baseline.md`
  - `rg -n "type .*Event|Dispatch|dispatch|EventTarget|AddEventListener|Listener|Capture|Bubble|Default|PreventDefault|StopPropagation|stop|passive|once|target path|DefaultAction|RunDefault|default action" event internal widget -g "*.go"`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `internal/events.go`
  - `event/event.go`
  - `event/input.go`
  - `event/keyboard.go`
  - `event/text.go`
  - `event/drag.go`
  - `event/custom.go`
  - `event/dispatcher.go`
  - `widget/event_defaults.go`
  - `widget/list_grid.go`
  - `widget/input.go`
  - `event/event_test.go`
  - `event/keyboard_test.go`
  - `event/text_test.go`
  - `widget/input_event_test.go`
- 关联能力：
  - EventTarget target-to-root path 构建
  - capture、target、bubble 三阶段分发
  - listener priority、注册顺序、once、passive 语义
  - stopPropagation、stopImmediatePropagation、preventDefault 语义
  - keyboard default action、widget default action 分层
  - portal/boundary 对事件路径的改写

## 分发顺序图

### 通用 `Runtime.DispatchEvent`

```text
event.DispatchEvent / typed Dispatch*
  -> Runtime.DispatchEvent(ctx, target, event)
     -> normalize target
     -> RegisterEventTarget(target, eventParentFor(target))
     -> fill event.Target when empty
     -> fill event.Time when empty
     -> path = eventPath(target)              # target -> parent -> ... -> root
     -> capture phase                         # root -> ... -> parent
        for i := len(path)-1; i >= 1; i--
          dispatchEventListeners(current, Capture, capture=true)
     -> target phase
        dispatchEventListeners(target, Target, capture=true)
        if !ImmediatePropagationStopped:
          dispatchEventListeners(target, Target, capture=false)
     -> bubble phase when event.Bubbles
        for _, current := range path[1:]      # parent -> ... -> root
          dispatchEventListeners(current, Bubble, capture=false)
     -> allowed = !(event.Cancelable && event.DefaultPrevented)
     -> record diagnostics
     -> reset event.Phase/Event.CurrentTarget
     -> return allowed
```

结论：当前通用 dispatcher 采用 browser-style 三阶段模型，但 target 阶段会先执行 target capture listener，再执行 target bubble/non-capture listener；这点是已实现语义，应作为后续兼容约束。

### listener 调用规则

| 规则 | 实现位置 | 当前行为 |
| --- | --- | --- |
| listener 存储 | `Runtime.RegisterEventListener` | 每帧注册到 `events.listeners[target]`，注册时写入递增 `seq`。 |
| 同 target 同 phase 顺序 | `eventListenerLess` | `Priority` 高者先执行；同 `Priority` 按 `seq` 先注册先执行。 |
| 类型过滤 | `dispatchEventListeners` | 只调用 `listener.Type == event.Type` 的 listener。 |
| capture 过滤 | `dispatchEventListeners` | 只调用 `listener.Options.Capture == capture` 的 listener。 |
| once | `dispatchEventListeners` | handler 返回后标记 `removed`，本 target 本次循环后 `pruneRemovedEventListeners`。 |
| passive | `dispatchEventListeners` | handler 执行期间设置 `event.currentPassiveListener`；`PreventDefault` 在 passive listener 中返回 `false` 且不设置 `DefaultPrevented`。 |
| stopPropagation | `Event.StopPropagation` + `DispatchEvent` phase loop | 阻止后续 target；不跳过当前 target 已进入的同一 listener 循环，除非同时 immediate stop。 |
| stopImmediatePropagation | `Event.StopImmediatePropagation` + `dispatchEventListeners` | 设置 propagation stopped，并跳过当前 target/phase 的后续 listener。 |
| preventDefault | `Event.PreventDefault` | 只有 `Cancelable && !currentPassiveListener` 时成功；成功后 `DispatchEvent` 返回 `false`。 |

### target path、portal、boundary

| 机制 | 路径影响 | 证据 | 结论 |
| --- | --- | --- | --- |
| 普通 target | `eventPath(target)` 产生 `target -> parent -> ... -> root` | `TestDispatchOrderCaptureTargetBubble` 断言 path 为 child、parent、root | 普通路径稳定且有测试覆盖。 |
| portal | `RegisterEventPortal(ctx, owner)` 将逻辑 event parent 改为 owner | `TestPortalAndBoundaryEventPathRules` 断言 portal 内路径进入 owner 再到 root | overlay/portal 的事件路径可脱离 layout parent。 |
| boundary stop | `RegisterBoundary` 默认 `BoundaryStop` | 同测试断言 boundary 内路径停止在 boundary | 用于 Dialog 等 overlay 阻断外层传播。 |
| boundary redirect | `BoundaryRedirectTo(target)` | 同测试断言路径进入 redirect target 再到 root | 可将 boundary 之后的传播接回指定 owner。 |

### default action 顺序

```text
通用 pointer/input/drag/custom event:
  typed defaults -> Runtime.DispatchEvent -> caller sees allowed/defaultPrevented
  default action is NOT run by Runtime.DispatchEvent

keyboard event:
  Runtime.DispatchKeyboardEvent
    -> applyKeyboardDefaults
    -> Runtime.DispatchEvent
    -> if keydown && allowed:
         dispatchShortcuts
         if not defaultPrevented:
           runKeyboardDefault(Tab focus move, Enter/Space focus activation)

widget click/input/wheel examples:
  widget creates event
  -> Dispatch*Event
  -> if allowed:
       run widget-local default action or legacy callback
```

| default action 类型 | 执行位置 | 可取消性 | 当前结论 |
| --- | --- | --- | --- |
| `keydown` Tab focus move | `Runtime.runKeyboardDefault` | `keydown` listener 可 `PreventDefault` | runtime 统一执行，有测试覆盖。 |
| `keydown` Enter/Space focus activation | `Runtime.runKeyboardDefault` -> `activateFocusTarget` | `keydown` listener 可 `PreventDefault` | runtime 统一执行，有测试覆盖。 |
| local shortcut | `Runtime.DispatchKeyboardEvent` 在普通 `DispatchEvent` 后、keyboard default 前执行 | shortcut listener 也可 `PreventDefault` 阻止后续 keyboard default | 属于键盘 default-action 管线的一部分，但不是通用 Event default action。 |
| legacy click callback | `event.Dispatcher.DispatchClickEvent`、`widget.dispatchClickDefault` | `click` listener 可 `PreventDefault` | 分散在 event/widget 兼容层。 |
| ScrollView wheel scroll | `widget/list_grid.go` 的 `DispatchWheelEvent` 后调用 `applyWheelDefault` | `wheel` listener 可 `PreventDefault` | 分散在 widget，未进入 runtime default action registry。 |
| TextField beforeinput mutation | `widget/input.go` 的 `commitUserInput` / `commitProgrammaticInput` | `beforeinput` 可取消；用户输入取消后回滚 editor 文本 | 分散在 widget，测试覆盖 beforeinput 回滚。 |
| Drag/drop 接受语义 | `event/drag.go` 设置 cancelable；drop/dragover caller 根据 allowed 决策 | 多数 drag event 可取消，`dragleave/dragend` 不可取消 | cancelable 默认值在 typed dispatch，后续动作在 widget/source-target 逻辑中。 |

## 事实结论

1. `internal.Event` 已包含 `Target`、`CurrentTarget`、`Phase`、`Bubbles`、`Cancelable`、`DefaultPrevented`、`Trusted`、`Detail` 和内部 path/stop/passive 状态，具备 browser-style event object 的基础形态。
2. `Runtime.DispatchEvent` 是唯一通用 capture/target/bubble dispatcher；`event.DispatchEvent` 和各 typed dispatch 最终都进入它。
3. capture 顺序为 root 到 target parent；target 阶段先执行 capture listener，再执行 non-capture listener；bubble 顺序为 target parent 到 root。
4. `ComposedPath()` 返回 target-to-root 的 path 副本；测试已覆盖普通 path、portal path、boundary stop path、boundary redirect path。
5. 每帧 `beginEventFrame` 会清空 `targets`、`listeners`、`focusTargets`、`shortcuts` 并重置 `nextSeq`；listener 顺序是帧内注册顺序，不跨帧保留。
6. listener 顺序由 `Priority` 降序和同优先级 `seq` 升序决定；源码有明确实现，但当前未发现针对 `Priority` 的专门测试。
7. `Once()` listener 在首次调用后标记移除，并在当前 target listener 循环后裁剪；已有测试覆盖同一帧内第二次 dispatch 不再调用 once listener。
8. `Passive()` listener 期间 `PreventDefault()` 返回 `false`，不会设置 `DefaultPrevented`；已有测试覆盖。
9. `StopPropagation()` 阻止后续 target/phase；`StopImmediatePropagation()` 额外阻止同 target/phase 的后续 listener；已有测试覆盖。
10. `PreventDefault()` 只改变 cancelable event；`DispatchEvent` 返回值只表达“是否允许 default action”，不自动执行通用 default action。
11. runtime 统一 default action 目前只覆盖键盘：`DispatchKeyboardEvent` 在普通事件分发后执行 shortcut，再执行 `Tab` focus move 和 `Enter/Space` focus activation。
12. click、wheel、input、drag/drop 等 default action 仍分散在调用方 widget 或兼容 dispatcher 中。
13. typed dispatch 会覆盖或补齐部分默认字段：pointer/wheel/keyboard/input/composition/drag/custom/activation 的 `Bubbles`、`Cancelable`、`Detail` 规则不完全一致，应按类型矩阵理解。

## 风险

| 风险 | 级别 | 说明 | 建议归属 |
| --- | --- | --- | --- |
| default action 未集中注册 | M | `Runtime.DispatchEvent` 只返回 allowed；click、wheel、beforeinput、drag/drop 的默认动作由各 widget 自行在 dispatch 后判断，后续新增组件可能漏掉 `PreventDefault` gate。 | A5.5 default action 可取消性矩阵 |
| target capture listener 顺序需文档化 | S | target 阶段先 capture 后 non-capture，接近 DOM 语义但项目内文档不足；后续测试或文档应固定该兼容语义。 | Event API 文档 |
| `Priority` 缺少测试 | S | 源码按 priority 降序排序，但现有 `event/event_test.go` 未见专门覆盖；后续改排序可能无测试提醒。 | Event test |
| `Once()` 是帧内 listener 生命周期 | S | listener registry 每帧重建；`Once()` 只能约束当前帧已注册 listener 的多次 dispatch，不能表达跨帧“一次性订阅”。 | Event API 文档 |
| passive 只保护 `PreventDefault` | S | passive listener 仍可调用 stop/stopImmediate；若期望 passive 只影响 default action，应明确写入文档。 | Event API 文档 |
| shortcut 与普通 keydown 的 stop 共享同一 event 状态 | M | `dispatchShortcuts` 在普通 dispatch 后执行，并会受 `ImmediatePropagationStopped/PropagationStopped` 影响；这与 DOM 全局 shortcut 模型不同，需要作为 FluxUI 语义固定。 | A5.3/A5.5 |
| portal/boundary 改写传播路径 | M | overlay 事件 parent 可能不同于 layout parent；事件审查和 hit-area 审查必须共同验证 target path 与视觉层级。 | A3.4、A5 后续 |

## 验收

- 已能解释 listener 顺序：
  - 不同 target：capture root->parent，target，bubble parent->root。
  - 同 target/phase：`Priority` 降序，同 priority 按注册 `seq` 升序。
  - target 阶段：capture listener 先于 non-capture listener。
- 已能解释 `once`：
  - handler 调用后标记 removed；
  - 本 target listener 循环结束后裁剪；
  - 每帧 listener registry 重建，因此语义是当前帧注册项内的一次性调用。
- 已能解释 `passive`：
  - passive listener 中 `PreventDefault()` 返回 `false`；
  - 不设置 `DefaultPrevented`；
  - 不阻止 stop/stopImmediate。
- 已能解释 `stop`：
  - `StopPropagation()` 阻止后续 target/phase；
  - `StopImmediatePropagation()` 阻止当前 target/phase 的后续 listener，并停止继续传播。
- 已能解释 `preventDefault`：
  - 仅 cancelable 且非 passive listener 中成功；
  - 成功后 `DispatchEvent` 返回 `false`；
  - 是否执行 default action 由 `DispatchKeyboardEvent` 或 widget 调用方检查 allowed。
- 已能解释 default action 顺序：
  - 通用 dispatcher 不执行 default action；
  - keyboard default action 在普通 dispatch 和 shortcut 之后执行；
  - widget-local default action 在 typed dispatch 返回 allowed 后执行。

## 后续依赖

- A5.3 pointer capture / hover / enter-leave 审查应复用本文件的 target path 和 stop 语义。
- A5.4 keyboard/focus 审查应复用 `DispatchKeyboardEvent -> dispatchShortcuts -> runKeyboardDefault` 顺序。
- A5.5 default action 可取消性矩阵应把 runtime default action 与 widget-local default action 分列。
- A7 文本输入审查应复用 beforeinput 可取消和 widget 回滚语义。
- A9 文档/API 审查应补充 `Once` 帧内语义、`Priority` 顺序、target capture 顺序和 passive 限制范围。
