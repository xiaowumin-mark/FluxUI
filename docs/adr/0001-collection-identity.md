# ADR-0001：集合 identity 使用稳定逻辑 key

- 状态：Accepted（R0 文档合同）
- 决策日期：2026-07-12
- 适用范围：Select/Combobox、ListView、DataGrid、TreeView、TagInput 及其 Element 包装层。

## 背景

高级集合会重排、筛选、插入、删除、条件隐藏或通过 Portal/Overlay 呈现。把选择、展开、编辑草稿、动画或焦点状态绑定到可变 index，会在下一帧把状态错误地交给另一个逻辑项。FluxUI 的跨帧状态本来就以 PathID 为 owner；集合内还需要一个由宿主声明的逻辑 identity。

## 决定

1. 可重排或需要保留状态的集合项必须提供稳定逻辑 key。key 的含义由宿主拥有，不能用当前显示 index、排序位置或临时 label 代替。
2. 集合状态的内部归属为“owner PathID + collection identity + item key + state role”。`state role` 例如 selected、expanded、edit-draft、active-focus；它不是对外 API。
3. key 只需在同一集合 owner 内唯一。相同 key 可在不同集合、不同 owner 或不同 overlay 实例中重复使用；它们不得共享状态。
4. 同一帧内出现重复 key 是调用方错误。开发/诊断路径必须报告冲突；冲突项不得复用另一项的持久状态。实现可以把冲突项降级为一次性状态，但不能静默串位。
5. key 改变视为旧项卸载、新项挂载。旧 key 的局部状态不迁移给新 key；由宿主显式把业务值迁移到新 key。
6. `ui.Key` 可以承担 Element 树 identity，但不自动替代业务集合的 item key。将来若两者可安全复用，必须在 API 合同中显式说明。

## 合同与不变量

- 逻辑状态随 key，而不是随当前 index、绘制顺序或可见范围移动。
- 过滤导致项暂时不可见时，是否保留其业务选择由受控值的宿主决定；组件不得根据 index 猜测替代项。
- 移除项后，任何内部临时状态不得泄漏给随后占用同一 index 的新项。
- Portal 只改变视觉/事件路径，不改变集合项的 owner identity；overlay 内的集合要有自己的 collection identity。
- 组件不得把 key 序列化、跨窗口共享，或把它当作安全标识。

## 内部测试夹具

夹具放在 `widget` 的测试专用文件中，使用真实 `internal.Runtime` 与多帧 frame runner；不能为了测试让 `internal` 依赖 `widget`。每个试点至少覆盖：

| 场景 | 断言 |
| --- | --- |
| 重排 `[a,b,c] → [c,a,b]` | `a` 的选择、草稿或焦点仍属于 `a`。 |
| 前插/删除 | 新的 index 0 不获得原 index 0 的状态；删除项状态不可被相邻项读到。 |
| 过滤后恢复 | 受控业务值不被组件按 index 改写；恢复项按其 key 重新关联。 |
| 条件卸载/重新挂载 | 按组件合同验证应保留或应清理的状态，不发生串位。 |
| 重复 key | 有确定的诊断/失败断言，且不会复用错误项的状态。 |
| Element key 与 item key 不同 | 不因恰好同名而错误共享 owner 或状态 role。 |

建议提供一个只服务测试的 collection fixture：接受有 key 的项序列、在相同 runtime 中连续布局多帧，并暴露按 key 读取的测试探针。夹具本身不得进入公共 `ui` API。

## 非目标

- 本 ADR 不规定公开的 `CollectionKey` 类型、全局 key 注册表或通用集合框架。
- 本 ADR 不定义 DataGrid 的行/列 API、Tree 的懒加载模型，也不承诺跨应用持久化。
- 本 ADR 不把 `ui.Key` 的字符串格式当作用户数据或安全边界。

## 回滚

先以 `widget` 内部 helper 和单组件适配器落地。若试点暴露问题，回滚该适配器与其测试夹具即可，保留既有组件的 PathID 语义；不迁移 `internal` 的依赖方向，不发布尚未验证的 key 类型。

## R0 文档完成记录

- 2026-07-12：稳定 key、owner 边界、冲突处理和多帧测试夹具要求已冻结。
- 实现完成的最低证据：两个试点共享同一内部 helper，且上述重排/删除/冲突断言均通过。
