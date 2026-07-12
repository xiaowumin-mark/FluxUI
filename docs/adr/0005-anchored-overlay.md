# ADR-0005：锚定 Overlay 使用统一 placement、关闭与焦点恢复

- 状态：Accepted（R0 文档合同）
- 决策日期：2026-07-12
- 适用范围：Select、Combobox、DatePicker、ContextMenu、Menu、Popup 及其未来 Element 包装层。

## 背景

锚定浮层同时涉及窗口坐标、Portal 事件路径、outside click、Escape、z-order 和 focus restore。各组件分别计算保护区域或关闭时机，会造成打开 click 立即关闭、嵌套菜单错关、滚动后位置失真和焦点回不到 trigger。

## 决定

1. 锚定 Overlay 的内部输入统一为：anchor rect（窗口坐标）、已测量的 popup 尺寸、window viewport、placement preference、当前 overlay stack 和 opener focus。输出为实际 popup rect、side/alignment、可用最大尺寸及关闭/恢复焦点决策。
2. placement 的默认策略是优先 placement，空间不足时翻转，再在可见 viewport 内平移；仍放不下时约束 popup 的可用尺寸并让其自身滚动。不得使用窗口全局尺寸绕过 `Context.Viewport()`。
3. protected region 是 trigger rect 与 popup rect 的并集（必要时含明确的子菜单安全区域）。只有 pointer 事件落在 topmost overlay 的 protected region 之外，才可以触发 outside-close。
4. overlay stack 的 z-order 是显式、确定的。Escape 和 outside-close 只交给 topmost 可关闭 overlay；关闭一个 overlay 不得顺带关闭无关的祖先/兄弟，除非其组件合同明确级联。
5. 打开时记录可恢复的 opener focus target。关闭时仅在该目标仍挂载、可聚焦且不在已关闭子树中时恢复；否则按组件合同退化为 anchor、最近合法 owner 或无焦点。modal overlay 还要按自己的合同限制 Tab 边界。
6. Portal 只改变视觉挂载与事件路径的必要部分：Overlay 保留 owner/state identity，普通文档流不因 popup 而占位；Popup 内部事件仍按 owner path 冒泡，modal boundary 的截断必须显式。

## 合同与不变量

- 打开 overlay 的 pointer event 不能在同一事件周期被 outside-close 当成“外部点击”。
- anchor 或 scroll viewport 改变后，下一 frame 的 popup rect 必须重新计算；旧 rect 不能永久用作命中事实。
- 关闭动作必须尊重 disabled、`PreventDefault`/default-action gate 和组件公开的 mask/escape 选项。
- Overlay 不在 Layout 中执行 I/O、后台任务或业务查询；动画只影响视觉或已声明的位置，不占普通 layout 尺寸。
- 选择列表的 active/selection 和 overlay focus 使用 ADR-0001/0002 的 key 与 roving 合同，不复制一套按 index 的状态。

## 内部测试夹具

测试夹具应能在同一 runtime 中布局 anchor、portal popup 和多层 overlay，注入 pointer/keyboard 事件并读取 stack、rect、关闭原因与 focused target。最低用例：

| 场景 | 断言 |
| --- | --- |
| 下方空间不足 | placement 翻转或平移；popup rect 不越出 window viewport。 |
| 窄 viewport/超高 popup | 计算出的最大尺寸有限，popup 可在自身区域滚动。 |
| 打开 click 与内部 click | 不关闭；trigger/popup 都属于 protected region。 |
| 真正 outside click | 只关闭 topmost 可关闭 overlay，关闭原因可观测。 |
| 多层 overlay + Escape | Escape 逐层关闭 topmost，不误关下层。 |
| 关闭后的 focus | 合法 opener 恢复；已卸载/disabled opener 有确定退化。 |
| 滚动/resize 后 | rect、hit region 和 hover/click 在下一 frame 更新。 |
| Portal 事件路径 | popup 事件保留 owner path；modal boundary 只按合同截断。 |

## 非目标

- 本 ADR 不统一 Dialog 的全部 modal 视觉样式，也不取代系统原生菜单/文件对话框。
- 本 ADR 不承诺跨窗口 overlay、跨显示器全局定位、任意 Docking 或完整辅助技术支持。
- 本 ADR 不把 `SelectSearchable` 变成搜索输入；真实搜索由 R1 Combobox/Autocomplete 合同定义。

## 回滚

placement/stack/focus helper 必须通过每个组件的局部 adapter 接入。若试点回归，可回滚一个 adapter 并保留原有 Popup/Dialog 路径；不把 `widget` 逻辑下沉为 `internal` 对组件的依赖。

## R0 文档完成记录

- 2026-07-12：placement、protected region、topmost 关闭、Portal owner 和 focus restore 合同已冻结。
- 实现完成的最低证据：两个锚定组件复用 helper，并通过 placement、outside click、Escape、滚动重定位和焦点恢复夹具。
