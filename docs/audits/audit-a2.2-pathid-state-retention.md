# A2.2 PathID 和状态保存审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 1：包边界和 runtime 基础。

- 状态：Done
- 日期：2026-07-06 17:38:23 +08:00
- 负责人：Codex
- 关注：Runtime、Widget
- 输入命令：
  - `git status --short`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "PathID|PathMemory|ComponentState|State|Ref|Attach|Portal|portal|child|Children|key|Key|Use(State|Ref)|HookStore|path" internal ui widget layout event app examples -g "*.go"`
  - `rg -n "AttachRef|bindInvalidator|drainCommands|UseRef|New.*Ref|Ref\[|\.Current|Attach" ui widget state internal -g "*.go"`
  - `rg -n "RegisterEventPortal|portal|Boundary|owner|eventTarget|parent" internal/events.go event/event.go event/event_test.go`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `internal/path_memory.go`
  - `internal/context.go`
  - `internal/runtime.go`
  - `internal/hook_store.go`
  - `internal/tree.go`
  - `internal/events.go`
  - `state/state.go`
  - `state/key.go`
  - `state/state_test.go`
  - `ui/element.go`
  - `ui/hooks.go`
  - `ui/reconciler.go`
  - `ui/reconciler_test.go`
  - `widget/component_refs.go`
  - `widget/event_boundary.go`
  - `widget/list_grid.go`
  - `widget/list_grid_state_mismatch_test.go`
- 关联能力：
  - PathID 身份规则
  - Runtime memory active 标记和帧尾清扫
  - HookStore component identity
  - Element keyed/unkeyed reconciler
  - Widget ListView/GridView 位置状态
  - Ref attach 命令队列
  - Portal event parent 与状态路径边界

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| 当前分支 | `main` |
| `git status --short` | 无输出 |
| 判断 | A2.2 执行前工作区干净；本任务只新增 `audit-a2.2-pathid-state-retention.md` 并更新索引，不修改 runtime 代码。 |

## 路径身份规则

```text
root PathID = 1

Context.Child(index)
  -> Runtime.childPath(parent PathID, index)
  -> path key: (parent, child, index)
  -> hookIndex reset to 0
  -> RegisterEventTarget(childPath, parent)

Context.Scope(name)
  -> Runtime.scopePath(parent PathID, name)
  -> path key: (parent, scope, name)
  -> hookIndex reset to 0
  -> RegisterEventTarget(scopePath, parent)

Context.NextMemoryKey(namespace)
  -> MemoryKey{Path: current PathID, Namespace: namespace, Slot: current hookIndex}
  -> hookIndex++
  -> Runtime.RecordHookCountID(current PathID, next hook count)

Context.ScopeMemoryKey(namespace)
  -> MemoryKey{Path: current PathID, Namespace: namespace, NoSlot: true}
```

`Runtime.pathIDs`、`pathDebug`、`nextPathID` 挂在 `Runtime` 上，不在 `BeginFrame` 清空。相同 parent+index 或 parent+name 会复用同一个 PathID；不同业务数据只要落到同一结构位置，就会复用同一条路径。

## 状态保存矩阵

| 状态来源 | key 规则 | 保留规则 | 清理规则 | 重排语义 |
| --- | --- | --- | --- | --- |
| `state.Use` / `UseWithInitial` | `ctx.NextMemoryKey("state")` | `Runtime.memory` 按 PathID+namespace+slot 保存 | 本帧未访问的 key 在 `sweepInactiveMemory()` 删除 | 按路径和 slot，不按业务数据 |
| `ctx.Memo(namespace)` | `ctx.NextMemoryKey(namespace)` | `Runtime.memory` 保存 | 本帧未访问删除 | 按路径和 slot |
| `ctx.PersistentKey(key)` | 调用方传入 `MemoryKey` | `Runtime.memory` 保存 | 本帧未访问删除，或显式 `ForgetPersistentKey` 删除 | 取决于调用方 key |
| legacy `ui.UseMemo` / `ui.UseRef` / animation hooks | `ctx.NextMemoryKey("memo"/"ref"/"anim_*")` | `Runtime.memory` 保存 | 本帧未访问删除 | 按路径和 slot |
| component `ui.UseState` / `UseMemo` / `UseRef` | `HookStore` 的 `ComponentIdentity.StableID()` + hook slot | `HookStore.instances` 保存 | 组件实例本帧未 active 时 unmount，执行 cleanup 后删除 | keyed 组件按 key，unkeyed 组件按 position |
| widget 内部状态 | 多数用 `ctx.Memo(...)` | `Runtime.memory` 保存 | 本帧未访问删除 | 按 widget 当前布局路径 |
| Ref 命令 | 外部持有 Ref 对象的 command queue | 不依赖 PathID；组件 Layout 时 attach invalidator 并 drain | 命令被 drain 后清空 | 命令作用于当前 attach 的组件 |
| Event portal parent | `RegisterEventPortal(ctx, owner)` 改写 event parent | 每帧 event registry 重建 | 下帧未注册即消失 | 只影响 event path，不迁移状态 key |

## 列表重排规则

### Element keyed list

`ui.Key(key, child)` 进入 reconciler 身份匹配。`ComponentIdentity.StableID()` 在有 key 时使用 `ParentID + TypeID + #key`，无 key 时使用 `ParentID + TypeID + @position`。

当前测试覆盖：

- `TestReconcilerKeyedChildrenPreserveStateAcrossReorder`：keyed 子组件重排后，`UseState` 状态跟随 key。
- `TestReconcilerKeyedDynamicListInsertDeleteReorderKeepsStateByKey`：插入、删除、重排组合下，保留项状态跟随 key，删除项执行 cleanup。
- `TestReconcilerUnkeyedChildrenUseIndexFallback` 和 `TestReconcilerUnkeyedDynamicListInsertDeleteUsesIndexFallback`：unkeyed 子组件状态按 position 继承，业务项会互相串状态。

结论：Element 层已有稳定 key 入口，后续修复不得破坏 keyed 组件按 key 保留状态的签名和行为。

### Widget ListView/GridView

`widget.ListView` 的 virtualized item 在 `gioLayout.List.Layout` 回调中使用 `childCtx := next.Child(index)`。`widget.GridView` 先用 `rowCtx := next.Scope(strconv.Itoa(rowIndex))`，再用 `rowCtx.Child(i-startIdx)` 生成 cell 路径。业务 item ID 没有进入 PathID。

当前测试 `widget/list_grid_state_mismatch_test.go` 明确记录：

- `TestListViewStateFollowsPositionOnInsertAndDelete`：列表头部插入或删除后，item 内部 `state.Use` 按旧位置继承。
- `TestGridViewStateFollowsPositionOnInsertAndDelete`：网格头部插入或删除后，cell 内状态按旧 slot 继承。

结论：ListView/GridView 当前是按位置保存 item 内部 path state。若后续要求按业务数据保存，需要新增 keyed item API 或在 Element 层使用 `Key(...)` 组织组件状态。

## 条件渲染规则

- 同一路径本帧调用的 `NextMemoryKey` 数量若与上一帧不同，`Runtime.EndFrame()` 会通过 `RecordHookCountID` 检查 panic，提示 hooks 不得条件调用。
- 某个子 scope 整体消失时，本帧不再访问其 memory key；`sweepInactiveMemory()` 会删除该子树状态，不会因为子树消失触发同一路径 hook count mismatch。
- `HookStore.EndFrame()` 会删除本帧未 active 的 component instance，并执行 hook cleanup。

当前测试覆盖：

- `state.TestHookCountMismatchPanics`：同一路径 hook 数减少会 panic。
- `state.TestHookUnmountDoesNotPanic`：子 scope 消失视为 unmount，不误报 hook count mismatch。
- `ui/reconciler_test.go` 的 key/type/provider/fragment 删除测试：component instance 未 active 时执行 cleanup。

结论：条件渲染隐藏子树不会保持该子树内部 path memory 或 component hook state；重新出现时会重新初始化。若要隐藏期间保留状态，状态必须提升到仍然渲染的父路径，或保留对应 keyed component 的 active 渲染入口。

## Portal 状态规则

`widget.EventPortal(child, owner)` 的布局规则是：

```text
portalCtx := ctx.Scope("event-portal")
RegisterPortal(portalCtx, owner)
child.Layout(portalCtx.Child(0))
```

`RegisterEventPortal` 只把 portal root 的 event parent 改为 owner；`eventPath` 因此可以变为 `insidePortal -> portal -> owner -> root`。PathID 和 MemoryKey 仍来自 portal 实际布局上下文 `event-portal/0`，不会因为 owner 改变而自动迁移到 owner 子树。

结论：

- portal 下状态如何保持：同一视觉挂载点、同一 parent PathID、同一 `event-portal` scope、同一 child index 时保持。
- portal owner 改变：事件冒泡 owner 改变，但内部状态 key 不随 owner 迁移。
- portal 被条件移除：本帧未访问的 portal 子树 memory 和 HookStore instance 会在帧尾删除。
- 多个兄弟 portal 若都使用相同 parent 下的 `Scope("event-portal")`，会落到同一 PathID；需要外层额外 `Scope`、Element `Key` 或不同结构位置来避免状态/事件 target 混用。

## Ref attach 规则

Ref 对象由调用方创建并持有，例如 `NewButtonRef()`、`NewInputRef()`、`NewScrollRef()`。组件 Layout 时执行：

- `ref.bindInvalidator(redrawInvalidator(ctx))`
- `ref.drainCommands()`
- 把 drain 出来的命令应用到当前组件内部状态或触发 synthetic/default action

Ref 本身不保存 PathID，也不参与 Runtime memory sweep。它的命令作用目标是“当前帧 attach 这个 Ref 的组件”。因此：

- 同一个 Ref 稳定 attach 到同一 keyed 组件时，命令目标稳定。
- 同一个 Ref 被列表重排后的另一个 unkeyed item 复用时，命令会作用到新的位置目标。
- 同一个 Ref 同帧 attach 到多个组件时，先 drain 的组件会消费命令，后续组件可能拿不到命令；当前审查未看到一 Ref 多 attach 的防护。

## 风险

错配风险清单：

| 场景 | 当前行为 | 风险等级 | 说明 |
| --- | --- | --- | --- |
| unkeyed Element 列表重排 | 状态按 position 继承 | 高 | 已有测试确认，会导致业务项串状态；应要求动态列表使用 `Key`。 |
| Widget ListView/GridView 头部插入/删除 | item 内 `state.Use` 按 index/row+cell 继承 | 高 | 已有 mismatch 测试确认；当前没有业务 key 参数。 |
| 条件调用 hook/state | 同一路径 hook count 变化时 panic | 中 | 能暴露错误，但属于运行时崩溃；文档和示例应强调不要条件调用。 |
| 条件隐藏子树 | 子树状态帧尾清扫 | 中 | 重新显示会重建状态；若预期隐藏保留，需状态提升。 |
| portal owner 改变 | event parent 改变，state path 不变 | 中 | 父组件可能以为状态跟 owner 走，实际跟视觉挂载路径走。 |
| 同 parent 多个 `EventPortal` | 相同 `Scope("event-portal")` 可能复用 PathID | 高 | 兄弟 portal 若同结构层级重复命名 scope，存在状态和 event target 混用风险。 |
| Ref 在 unkeyed 列表中复用 | 命令作用到当前 attach 位置 | 中 | 外部 Ref 语义随 attach 目标变化，不跟业务数据自动绑定。 |
| Ref 同帧多 attach | 第一个 drain 者消费命令 | 中 | 当前未见唯一 attach 校验或诊断。 |
| PathID 表只增不删 | `pathIDs` 长期保留历史路径 | 低 | 可避免 ID 抖动，但大量动态 scope/name 可能导致 path table 增长。 |

## 事实结论

- PathID 的基础身份来源是 `root`、`Child(index)` 和 `Scope(name)`；业务数据不会自动进入 PathID。
- `Runtime.memory`、runtime effects、legacy `UseRef`/`UseMemo`/animation hooks 都依赖 PathID+namespace+slot，并在帧尾删除未访问 key。
- `HookStore` 是 component state 的独立保存层；有 key 时按 key 保留，无 key 时按 position 保留。
- Element keyed list 已有测试确认状态跟 key，unkeyed list 已有测试确认状态跟 position。
- Widget ListView/GridView 当前没有 item key 语义，item 内 path state 会在插入、删除、重排时按位置错配。
- 条件渲染整棵子树会 unmount 并清理状态；同一路径条件调用 hook/state 会触发 hook count mismatch panic。
- Portal 改写的是 event parent，不是状态路径；portal 内状态仍跟实际布局路径保存。
- Ref attach 是外部命令队列绑定到当前 Layout 目标，不是路径状态；Ref 目标稳定性取决于调用方是否把 Ref 稳定绑定到正确实例。

## 验收

- 已给出 PathID 身份规则：root、`Child(index)`、`Scope(name)`、`NextMemoryKey(namespace)`、`ScopeMemoryKey(namespace)`。
- 已解释列表重排：Element keyed list 按 key 保留；Element unkeyed list、Widget ListView/GridView 按位置保留并存在错配风险。
- 已解释条件渲染：同一路径 hook 数变化会 panic；子树消失会 unmount 并清理未访问状态。
- 已解释 portal 下状态保持：状态跟 portal 实际布局路径保持，event parent 可被 owner 改写但不迁移 state key。
- 已列出 Ref attach 与状态路径的边界：Ref 不参与 PathID，命令作用于当前 attach 目标。
- 已标出错配风险清单，后续修复不得把已有 keyed 行为、hook count 检查和旧 Ref 签名破坏掉。

## 后续依赖

- A2.3 Runtime registry 审查需要继续确认 event target、focus target、pointer capture 在 PathID 复用和 portal owner 改写下是否会 stale。
- A4.x 布局审查需要基于 ListView/GridView 的位置状态事实，评估虚拟列表、大列表和网格是否需要 keyed item 入口。
- A5.2 EventTarget 分发顺序审查需要基于 portal event parent 与状态路径分离事实，继续确认 composed path、owner、boundary 语义。
- A8.x overlay/portal 审查需要重点验证多个 overlay/portal 同级挂载时是否会因为固定 `Scope("event-portal")` 产生 PathID 冲突。
- A9.x 组件状态、Ref 和默认行为审查需要继续为每个 Ref 建立状态机，确认一 Ref 多 attach、unkeyed 列表 Ref 复用和 controlled/internal state 的优先级。
