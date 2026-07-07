# A10.7 拖放控件族审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A10.7 拖放控件族审查。

## 事实结论

审查范围覆盖 `DragSource` 和 `DropTarget`，主实现位于 `widget/drag_source.go`、`widget/drop_target.go`，UI facade 位于 `ui/ui.go` 和 `ui/extended_types.go`。这两个组件的核心能力依赖 Gio `transfer` 与 Gio pointer routing：FluxUI 负责 payload 归一化、事件桥接、旧 callback gate 和 child 尺寸内的输入 tag 注册。

| 维度 | `DragSource` | `DropTarget` | 结论 |
| --- | --- | --- | --- |
| payload | `DragSourcePayloads` 直接替换 payload；`DragSourceData` 添加单 MIME 数据；`DragSourceText` 生成 `text/plain`；`DragSourceFiles` 生成 `text/uri-list` 与 plain text。payload type 会 trim/lower，空 type、空 data、重复 type 被丢弃，data 会复制。 | 默认接受 `text/uri-list`、`text/plain;charset=utf-8`、`text/plain;charset=utf8`、`text/plain`、`application/text`；`DropTargetTypes` 会 trim/lower 去重，空列表回退默认类型。`DropTargetMaxBytes` 默认 32 MiB。 | payload 语义以 MIME type 为主，文件路径是 URI/plain text 的解析结果，不是独立原生文件对象。 |
| operation | `DragSourceOperations` 只记录 app-level 语义；默认 `copy`，支持 `copy/move/link`，非法值丢弃。注释已说明 Gio transfer 不暴露 OS-level operation negotiation。 | `DropTargetOperation` 默认 `copy`；`DropEvent.Operation` 和 `DragEvent.Operation` 使用归一化后的 operation。 | operation 是 FluxUI 事件/回调语义，不代表平台后端已完成协商。 |
| 命中/布局区域 | child 先正常布局；只有 child 尺寸为正、未 disabled、payload 非空时，才在 `clip.Rect(Max: childDims.Size)` 内注册 `gesture.Drag` 和 `transfer.SourceFilter` tag。返回尺寸始终是 child 尺寸。 | child 先正常布局；只有 child 尺寸为正、未 disabled 且存在 `onDrop`/`onError`/`onActiveChange` 时，才在 `clip.Rect(Max: childDims.Size)` 内注册 `transfer.TargetFilter` tag。返回尺寸始终是 child 尺寸。 | 两者不会把整行或整窗注册成拖放区域；hit area 与 child layout size 对齐。 |
| pointer conflict | 使用 `gesture.Drag`，axis 为 `gesture.Both`。pointer `Press` 记录起点，`Drag` 后进入 active，`Release/Cancel` 结束或取消。未发现 FluxUI click default 或 button callback 复用。 | 不直接注册 pointer press/move/drag/click，只注册 Gio transfer target。active 来自 `transfer.InitiateEvent`，离开/取消来自 `transfer.CancelEvent`。 | `DragSource` 会参与 pointer drag 手势竞争；`DropTarget` 主要参与 transfer routing，不主动抢 pointer click。 |
| scroll conflict | 未注册 `pointer.Scroll`，也不分发 FluxUI `WheelEvent`。拖拽预览通过 `op.Defer` 绘制，不改变 layout。 | 未注册 `pointer.Scroll`，也不分发 FluxUI `WheelEvent`。DropTarget 本身不滚动；内部 child 若是 ScrollView/ListView，滚动语义由 child 决定。 | FluxUI 代码层没有无条件截停纵向滚动的来源；拖拽期间 wheel 是否被平台 transfer 后端影响仍属于手动验收项。 |
| drop default | `transfer.RequestEvent` 到来时先执行 `transfer.OfferCmd`，再 dispatch FluxUI `drag` 事件；若事件 allowed 才调用 `DragSourceOnRequest` / `DragSourceOnEvent`。 | `InitiateEvent` 依次 dispatch `dragenter` 和 `dragover`，两者 allowed 才 `setActive(true)`；`DataEvent` 先读取 payload，再 dispatch cancelable `drop`，allowed 才调用 `onError` 和 `onDrop`，最后清理 active。 | `PreventDefault` 能阻止 FluxUI active/onDrop/onError/onEvent 回调，但不能阻止 Gio initiate，也不能阻止 `DropTarget` 在 drop dispatch 前读取 payload。 |
| 事件桥接 | `started` 映射 `dragstart`，`requested` 映射 `drag`，`completed/cancelled` 映射 `dragend`。`DragSourceOnEvent` 只在 event allowed 后调用。 | `dragenter/dragover/dragleave/drop` 都通过 `event.DispatchDragEvent` 分发到当前 `ctx.PathID()`；`DropTargetOnActiveChange` 不是 DOM-style event，而是 state callback。 | 新事件系统是旧 callback 的 gate，但仍存在 native Gio transfer 已发生、FluxUI 只能事后桥接的边界。 |

### payload 和默认行为矩阵

| 场景 | 输入 API | 输出字段 | 默认行为 | 可取消性 |
| --- | --- | --- | --- | --- |
| 文本拖出 | `DragSourceText(text)` | `DragPayload{Type:"text/plain", Data:text}`，事件 `Text=string(Data)` | Gio source request 时提供 `OfferCmd` | `drag` 可取消旧 `onRequest/onEvent`，但 `OfferCmd` 已先执行。 |
| 自定义 MIME 拖出 | `DragSourceData(mime, data)` / `DragSourcePayloads` | type lower-case，data copy | 同上 | 同上。 |
| 文件拖出 | `DragSourceFiles(paths...)` | `text/uri-list` 与 plain text；路径会清理并转 file URI | 同上，平台外部拖出能力依赖 Gio/backend | 同上；operation 不代表 OS negotiation。 |
| 文本/URI drop | `DropTargetTypes(...)` | `DropEvent.Type/Data/Text/Paths/Operation` | 读取 `DataEvent.Open()`，大小受 `DropTargetMaxBytes` 限制 | `drop.PreventDefault()` 阻止 `onDrop/onError`，不阻止读取。 |
| drag over | Gio `transfer.InitiateEvent` | `DragEvent{Type: dragenter/dragover, Active:true}` | allowed 后进入 active 并触发 `DropTargetOnActiveChange(true)` | `dragenter` 或 `dragover` 取消都能阻止 active 进入。 |
| cancel/leave | Gio `transfer.CancelEvent` | `dragleave`，然后 active=false | 清理 active 并触发 `DropTargetOnActiveChange(false)` | `dragleave` 默认不可取消，不阻止清理。 |

### pointer 与 scroll conflict 矩阵

| 场景 | 当前实现 | 对父级纵向滚动的影响 | 审计判断 |
| --- | --- | --- | --- |
| 普通 hover/未拖拽经过 `DragSource` | 只有 child 尺寸内 drag source tag，无 scroll filter | 不应截停 wheel | 符合验收方向。 |
| 在 `DragSource` 上按下并移动 | `gesture.Drag` 使用 `gesture.Both`，进入 active，并绘制 deferred preview | pointer drag 可能和可拖动 child/平台 DnD 竞争；wheel 未显式处理 | 需要手动 smoke 验证拖拽中 touchpad/wheel 行为。 |
| `DropTarget` 包住滚动 child | DropTarget 只注册 transfer target；child 自己处理 scroll | DropTarget 不截停 wheel，滚动由 child 决定 | 符合验收方向。 |
| `DropTarget` 没有 handler | 不注册 target filter | 不参与 drop/scroll | 符合最小影响边界。 |
| disabled 或空 payload | `DragSource`/`DropTarget` 不注册对应 source/target | 不参与拖放，也不改变 child layout | 符合兼容边界。 |

## 风险

- `DragSource` 使用 `gesture.Both`，适合自由拖拽，但没有阈值、轴锁或与可滚动父级的策略说明；拖拽启动时的 pointer capture/scroll gesture 竞争仍需要手动验收。
- `DragSource` 在 `transfer.RequestEvent` 中先执行 `OfferCmd`，再分发 FluxUI `drag` 事件；因此 `PreventDefault` 不能阻止数据 offer，只能阻止后续旧事件回调。
- `DropTarget` 在 dispatch `drop` 前已经读取 `DataEvent.Open()`；`drop.PreventDefault()` 不能避免 payload 读取成本、读取错误或 maxBytes 错误产生。
- `DropTarget` active gate 依赖 `dragenter` 和 `dragover` 两次 dispatch 都 allowed，但当前缺少 widget 级测试直接覆盖 `dragover.PreventDefault()` 阻止 active callback。
- `DropTarget` 的 file drop 来自 `text/uri-list` 或 plain text 解析；外部平台若提供不同 MIME 或 backend 字段，当前只能 best-effort。
- operation 只是 FluxUI app-level 字段，不能承诺 OS/native drag operation negotiation。
- 现有自动测试覆盖正常 DragSource 到 DropTarget transfer、pass-through hit testing、interactive target 和无目标取消；未覆盖 wheel during drag、drop preventDefault gate、maxBytes error gate、外部文件 drop 后端差异。

## 验收

- 已建立 `DragSource` / `DropTarget` 的 payload、pointer conflict、scroll conflict、drop default 矩阵。
- 已确认命中注册区域与 child layout size 对齐：source 和 target 都在 `clip.Rect(Max: childDims.Size)` 内注册 Gio tag，返回尺寸不扩大。
- 已确认两者不注册 `pointer.Scroll` 或 FluxUI wheel listener；默认不会无条件截停父级纵向滚动。
- 已明确 drop/default 可取消边界：`dragenter/dragover` 可阻止 active；`drop` 可阻止旧 `onDrop/onError`；但 `OfferCmd` 和 drop payload 读取不受 `PreventDefault` 控制。
- 已明确 operation、file drop、external drag-in/out 均依赖 Gio/platform 后端能力，不能只从 widget API 推断原生能力。
- 已标出后续手动验收入口：`examples/drag_drop_showcase` 和 docs browser drag/drop demo，重点验证拖拽中滚轮、嵌套 ScrollView 内 drop target、外部文件拖入和 preventDefault gate。

## 后续依赖

- A5.1：Gio transfer 到 FluxUI `DragEvent` / `DropEvent` 字段映射需复用本文的 MIME、operation、external backend 边界。
- A5.5：default action 可取消性矩阵需继续记录 `dragover/drop` gate 与 payload 读取先后顺序。
- A6.2 / A6.4：拖拽期间 wheel、滚动后 hit refresh、嵌套 ScrollView 内 DropTarget 需要作为滚动/命中回归入口。
- A11.4：拖放示例/API 文档应明确 operation 不等于 OS negotiation，`PreventDefault` 不阻止 payload 读取，并补充外部文件拖入 smoke 说明。
- A12.3：建议补充 widget 级测试：`dragover.PreventDefault()` 阻止 active、`drop.PreventDefault()` 阻止 `onDrop/onError`、`DropTargetMaxBytes` 错误路径、拖拽中父级 wheel 不被组件层截停。
