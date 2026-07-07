# FluxUI 组件编写指南

本文用于约束 `widget`、`ui` 暴露层和示例中的组件开发。新增或修改组件前，先阅读 `CODE_STYLE_GUIDE.md` 和 `FEATURE_INTEGRATION_CHECKLIST.md`。

## 组件职责边界

- `ui` 负责对外声明式 API、option 命名和组合体验。
- `widget` 负责组件行为、布局、事件桥接、Ref、默认行为和样式消费。
- `event` 负责事件类型、分发、路径、默认行为可取消性，不写组件业务语义。
- `internal` 负责 runtime、frame、path、registry、focus、redraw，不依赖具体 widget。
- `layout` 只处理约束和尺寸，不读取 widget 状态。
- `style/theme` 只描述样式和主题，不注册事件、不改变状态。

任何组件修复都不得越过包边界直接修改下层全局规则，除非路线图中明确该修复属于底层能力。

## 新组件最小结构

新增组件应包含以下内容：

- 对外构造函数：`Xxx(child Widget, opts ...XxxOption) Widget` 或 `Xxx(opts ...XxxOption) Widget`。
- option 类型：`type XxxOption func(*xxxConfig)`。
- 内部 widget 结构：只保存配置，不保存跨帧运行状态。
- 状态入口：使用 path 绑定的 state/ref/clickable/focus target，不用包级变量保存组件状态。
- 默认样式：从 theme/style 获取，再与用户 option 合并。
- 事件桥接：新事件先分发，default action 或旧 callback 后执行。
- 文档和示例：组件进入公开 API 后必须有 docs 或 example 入口。

## Option 设计

- option 名称使用组件前缀或明确语义，例如 `ButtonDisabled`、`ScrollHorizontal`、`DialogMaskClosable`。
- option 只设置配置，不执行副作用。
- option 中不得直接请求 redraw、注册事件、读取 Gio event。
- 布尔 option 必须有清楚默认值；默认值应符合最常见使用场景。
- callback option 必须说明触发时机、是否只在真实变化时触发、是否受 `PreventDefault` 影响。
- 受控值 option 和默认值 option 必须区分，例如 `Value`、`DefaultValue`、`OnChange` 的优先级要写清楚。

## 状态和 PathID

- 组件跨帧状态必须绑定当前 path，不能绑定循环 index 以外的隐式顺序。
- 列表、条件渲染、portal、overlay 中的状态保持必须考虑 key 或 owner path。
- 不允许把上一次 layout 的尺寸、命中区域、hover target 当成永久事实；每帧必须可重建。
- Ref attach 必须在当前帧明确绑定目标；未挂载或 disabled 时命令如何处理必须可预测。

## Layout 规则

- 组件必须尊重传入 constraints，不得无条件使用窗口尺寸。
- 视觉扩展、shadow、ripple、state layer 不应改变测量尺寸，除非 API 明确承诺。
- hit area 必须来自真实 layout area 或明确的 touch target 规则，不得意外扩大到整行或整窗。
- overlay 不应占用普通文档流，除非它是明确的 inline popup。
- 横向内容必须有最大宽度或横向滚动策略，不能把无限宽传给普通 Row。

## 事件和默认行为

- 用户输入路径顺序为：原始 Gio event -> FluxUI event -> listener -> default action -> 旧 callback 或状态变化。
- `PreventDefault` 只有在组件读取 dispatch 返回值后才有意义；新增 default action 必须接入 gate。
- `StopPropagation` 只影响事件传播，不应跳过组件内部状态清理。
- `passive` listener 调用 `PreventDefault` 不应生效。
- disabled/loading 这类组件状态属于入口阻断，不等同于 `PreventDefault`。
- keyboard activation 应与 pointer click 复用相同 default action 语义。

## Ref 规则

- Ref 命令只表达外部主动命令，不应绕过组件的 disabled/loading/controlled 规则。
- Ref 命令消费时机必须记录：入队、下一帧布局期、立即生效或丢弃。
- Ref 命令不应触发与用户输入完全相同的回调，除非文档明确。
- 未挂载期间命令是保留、丢弃还是延迟消费，必须在组件文档中说明。

## 样式和动画

- 默认样式与用户样式合并时，不能丢失 disabled、hover、focus、selected、error 表达。
- hover/pressed/focused 状态只能有一个权威来源，优先使用 event/clickable snapshot。
- ripple/state layer 只影响视觉，不扩大 layout 或 hit area。
- 动画由 frame/redraw 驱动，不用 goroutine 驱动 UI 状态。
- 动画结束后必须停止无意义 redraw。

## 文档和示例

- 公开组件必须有 docs 或 example 入口。
- 支持 Ref 的组件必须在文档列出 `NewXxxRef`、attach option 和 Ref 方法。
- 支持键盘或默认行为的组件必须说明 Enter、Space、Escape、Arrow、Tab 的语义。
- 支持滚动、overlay、拖放或 IME 的组件必须提供手动验收步骤。

## 完成前检查

提交组件改动前，至少回答：

- 这个组件改动会不会影响父级滚动、命中区域、focus、keyboard shortcut、overlay、Ref 或旧 callback？
- 这个组件是否保持旧 API 签名和旧 callback 顺序？
- 这个组件是否有自动测试或示例 smoke 入口？
- 如果改动失败，能否按组件局部回滚？
