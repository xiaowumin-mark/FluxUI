# FluxUI React 风格重构计划

## 背景

FluxUI 当前基于 Gio immediate mode 构建，每个 frame 都会重新执行 root builder 并重新布局组件树。跨帧状态主要依赖 `Runtime.memory`、`Context.path` 和 hook 调用顺序维持。

这套模型适合快速封装 Gio，但随着组件、路由、动态列表和复杂交互增加，会逐渐暴露以下问题：

- 状态身份依赖 `path + hookIndex`，动态列表插入、删除或条件渲染时容易错位。
- `Widget.Layout(ctx)` 同时负责构建、状态读取、事件处理、布局和绘制，职责过重。
- `Runtime.memory map[string]any` 同时保存 state、memo、router holder 和组件内部状态，生命周期清理不够清晰。
- `Context` 同时承载 Gio 上下文、运行时、树路径、hook 计数、主题、字体和窗口控制，后续扩展成本较高。
- Router 使用 `Persistent` 模拟上下文传递，不是真正的 Provider/Consumer 模型。

本计划目标是将 FluxUI 演进为 React-like API/runtime on Gio，而不是完全复制 React DOM 的调度和 commit 模型。

## 总体目标

- 保留 Gio 的逐帧 layout/ops 渲染模型。
- 引入 React 风格 API：`Element`、函数组件、`UseState`、`UseEffect`、`UseMemo`、`UseRef`、`UseContext`。
- 引入稳定组件身份：`type + key + position` 决定组件实例复用。
- 引入 Hook Slot：状态归属于组件实例，而不是全局字符串 path。
- 引入 Context Provider：统一 theme、router、route params、window 等上下文传递。
- 支持最小 Reconciler：同类型同 key 复用，不同类型或 key 变化时 unmount/remount。
- 保留旧 `Widget` API 作为 legacy 和 escape hatch，避免一次性破坏现有示例与用户代码。

## 非目标

- 不实现 React DOM 的完整 Fiber 调度器。
- 不引入并发渲染、Suspense、优先级调度等高级能力。
- 不尝试跳过 Gio 每帧 layout，因为 Gio 的 `layout.Context`、events 和 ops 都是 frame 相关。
- 不在第一阶段删除现有 `ui.Widget`、`widget.Widget`、`ui.Run` 等 API。

## 目标 API 草案

### 基础组件

```go
func Counter(ctx *ui.Context) ui.Element {
	count := ui.UseState(ctx, 0)

	return ui.Column(
		ui.Text(fmt.Sprintf("Count: %d", count.Value())),
		ui.Button(
			ui.Text("+1"),
			ui.OnClick(func(ctx *ui.Context) {
				count.Set(count.Value() + 1)
			}),
		),
	)
}

func main() {
	_ = ui.RunElement(Counter, ui.Title("Counter"))
}
```

### 动态列表

```go
func TodoList(ctx *ui.Context, todos []Todo) ui.Element {
	return ui.Column(
		ui.For(todos, func(todo Todo) ui.Element {
			return ui.Key(todo.ID, TodoItem(todo))
		})...,
	)
}
```

### Router

```go
func App(ctx *ui.Context) ui.Element {
	return ui.Router(
		ui.Route("/", Home),
		ui.Route("/users/:id", UserPage),
		ui.Route("/settings", Settings),
	)
}

func UserPage(ctx *ui.Context) ui.Element {
	params := ui.UseParams(ctx)
	nav := ui.UseNavigate(ctx)

	return ui.Column(
		ui.Text("User: " + params.Get("id")),
		ui.Button(ui.Text("Back"), ui.OnClick(func(ctx *ui.Context) {
			nav.Back()
		})),
	)
}
```

## 新核心抽象

### Element

`Element` 是声明式节点，不直接执行 Gio layout。它描述 UI 树结构、props、key 和 children。

建议方向：

```go
type Element interface {
	element()
}
```

后续可按需要扩展为结构化节点：

```go
type Node struct {
	Type     ElementType
	Props    any
	Key      string
	Children []Element
}
```

### Component

函数组件负责从 props 和 hooks 生成 Element。

```go
type Component func(ctx *Context) Element
type ComponentWithProps[P any] func(ctx *Context, props P) Element
```

### Component Instance

组件实例保存稳定身份和 hooks。

```go
type ComponentInstance struct {
	ID        string
	Type      any
	Key       string
	Hooks     []HookSlot
	Effects   []EffectSlot
	HostState any
}
```

复用规则：

- `same type + same key`：复用实例。
- `different type`：卸载旧实例，挂载新实例。
- `different key`：卸载旧实例，挂载新实例。
- 无 key 的 sibling：按 index fallback，但动态集合必须推荐显式 key。

### Hook Slot

Hooks 不再直接依赖全局字符串 key，而是绑定到当前组件实例。

目标 API：

```go
func UseState[T any](ctx *Context, initial T) *State[T]
func UseMemo[T any](ctx *Context, deps []any, factory func() T) T
func UseRef[T any](ctx *Context, initial T) *Ref[T]
func UseEffect(ctx *Context, deps []any, effect func() func())
func UseContext[T any](ctx *Context, key ContextKey[T]) T
```

### Context Provider

统一上下文传递，逐步替代 `Persistent` holder。

```go
type ContextKey[T any] struct{}

func Provider[T any](key ContextKey[T], value T, child Element) Element
func UseContext[T any](ctx *Context, key ContextKey[T]) T
```

优先迁移对象：

- Router state
- Route params/query
- Theme
- Font
- Foreground
- Window handle

### Host Component

基础 UI 控件作为 Host Component，最终仍调用 Gio layout/ops。

Host Component 分层目标：

- Element 描述 props 和 children。
- HostNode 保存 Gio widget state，例如 editor、clickable、scroll state。
- Renderer 每帧 layout host tree。

优先迁移控件：

1. `Text`、`Container`、`Padding`、`Row`、`Column`、`Stack`、`Center`、`Spacer`、`Divider`。
2. `Button`、`ClickArea`、`Checkbox`、`Switch`、`Slider`。
3. `TextField`、`ScrollView`、`List`、`Grid`、`Select`、`Tabs`。
4. `Dialog`、`Popup`、`Toast`、`Router`。

## 分阶段推进计划

### Phase 0：冻结现状并补行为基线

目标：在重构前记录当前行为，建立安全网。

任务：

- 增加 hook 调用顺序变化测试。
- 增加 sibling 组件状态隔离测试。
- 增加动态列表插入、删除时状态行为测试。
- 增加 route push、replace、back 的状态保留测试。
- 增加 route transition 时 from/to 页面 mount/unmount 行为测试。
- 增加 `TextField` controlled/uncontrolled 行为测试。
- 增加 refs 命令队列行为测试。
- 增加 async 跨 goroutine redraw 测试。

验收标准：

- `go test ./...` 通过。
- 当前旧 API 行为有测试覆盖，后续重构可判断是否破坏兼容。

### Phase 1：新增 React 风格 API facade

目标：先验证新 API 手感，不触碰底层 renderer。

任务：

- 新增 `Element` 类型。
- 新增 `RunElement(root func(ctx *Context) Element, opts ...AppOption) error`。
- 新增 `FromWidget(w Widget) Element` 和必要的 adapter。
- 新增 `UseState(ctx, initial)`，内部可先复用现有 `state.Use[T]`。
- 新增 `Fragment(children ...Element) Element`。
- 新增 `Key(key string, child Element) Element`。
- 提供一个 `examples/react_counter` 或 `examples/v2_counter` 验证 API。

验收标准：

- 新示例能运行。
- 旧 `ui.Run` 和旧 examples 不受影响。
- 文档中明确这是实验 API。

#### Phase 1 小阶段拆分

- **1.1 API 骨架期**：补齐 `Element`、`RunElement`、`UseState(ctx, initial)`、`Key`、`Fragment`。
- **1.2 Element 树拆分期**：让 `Element` 从纯 Widget 包装拆分为独立树结构，至少包含 host、fragment、key 三类节点。
- **1.3 渲染适配期**：建立唯一的 `Element -> Widget` 渲染出口，所有实验 API 统一走同一条转换路径。
- **1.4 身份信息接入期**：让 `Key` 进入树节点身份模型，作为后续 reconciler 的输入。
- **1.5 最小 Reconciler 结构期**：定义组件实例、hook slot、children diff 的基础结构。
- **1.6 基础迁移验证期**：用 counter、keyed list、fragment 示例验证新树结构。
- **1.7 Phase 1 收口期**：整理剩余技术债，明确 Phase 2 入口与迁移规则。

#### Phase 1 完成标准

- `Element` 具备独立树结构，不再只是单纯的 Widget 容器。
- `Key` 已进入树身份模型。
- `RunElement` 可用且稳定。
- `UseState(ctx, initial)` 可用且持久化正常。
- 存在统一的 `Element -> Widget` 渲染出口。
- 新 API 和旧 API 可并行存在，并有测试覆盖。

#### Phase 1.7 收口任务

- **1.7.1 API 定稿**：确认 `Element`、`RenderElement`、`ElementInfo`、`Key`、`Fragment` 的最终公开形态。
- **1.7.2 文档收口**：把实验 API 的定位、用法、限制写清楚，避免用户误把它当作最终稳定 API。
- **1.7.3 迁移边界**：明确哪些能力继续走旧 Widget 路径，哪些能力已经进入新树模型。
- **1.7.4 Phase 2 入口定义**：定义 hook slot、instance、diff 的最小切入点。

#### Phase 1.7 完成标准

- 实验 API 入口稳定，不再频繁改名。
- `Element` 的身份和渲染出口已固定。
- Phase 2 所需的输入点已经明确。

#### Phase 2 小阶段拆分

- **2.1 Hook Slot 设计期**：定义组件实例、hook cursor、slot store 的数据结构。
- **2.2 状态迁移期**：让 `UseState/UseMemo/UseRef` 逐步转向 slot 存储。
- **2.3 Effect 生命周期期**：把 `UseEffect` 的 mount、update、cleanup 生命周期接入实例模型。
- **2.4 Context Provider 期**：把 router/theme/font/window 这些上下文转为 provider/consumer 模型。
- **2.5 Legacy 兼容期**：保留旧 `Context.Persistent` / `Context.Memo` 行为，确保旧 API 不破坏。
- **2.6 验证期**：补 hook 顺序、重挂载、卸载 cleanup、状态复用测试。

#### Phase 2 完成标准

- 新 hook 体系不再依赖字符串 path 作为唯一身份。
- 组件实例具有稳定的 hook slot 生命周期。
- `UseEffect` 能正确处理 mount / cleanup。
- 旧 API 保持兼容，且测试覆盖明确。

### Phase 2：Runtime memory 过渡到 Hook Slot

目标：新 React API 使用组件实例级 hook slot，旧 API 继续兼容。

任务：

- 新增 `HookStore` 或等价结构。
- 新增 `ComponentInstance` 的 hook cursor。
- `UseState/UseMemo/UseRef/UseEffect` 优先使用 hook slot。
- 保留 `ctx.Persistent` 和 `ctx.Memo` 支持 legacy API。
- `Runtime.EndFrame` 同时处理新旧 effect cleanup。
- 记录 active instances，未出现的 instance 触发 unmount cleanup。

验收标准：

- 新 API 的 state 不依赖 `Context.path + hookIndex` 字符串 key。
- unmount 后 effect cleanup 能执行。
- legacy state 行为保持不变。

### Phase 3：实现最小 Reconciler

目标：建立稳定组件身份、key 和 unmount 语义。

任务：

- 新增 Fiber/Instance tree。
- 支持 function component。
- 支持 host component。
- 支持 fragment。
- 支持 keyed children diff。
- 支持 unmount cleanup。
- 定义无 key children 的 index fallback 行为。

验收标准：

- 动态列表插入新元素后，已有 keyed item 的 state 不错位。
- key 改变时组件重新 mount。
- 类型改变时组件重新 mount。
- 每帧仍可正常提交 Gio layout/ops。

#### Phase 3 小阶段拆分

- **3.1 Reconciler 设计期**：定义最小 fiber/instance tree、child identity、mount/update/unmount 流程，不改变 Gio 每帧 layout 模型。
- **3.2 Function Component 接入期**：让函数组件进入 reconciler 路径，组件实例由 `type + key + position` 复用，而不是手动 `BeginInstance`。
- **3.3 Fragment 与 Keyed Children diff 期**：支持 fragment children 遍历、keyed sibling 匹配、无 key sibling 的 index fallback。
- **3.4 Host/Widget Adapter 接入期**：让 host element、provider element、legacy `FromWidget` 能作为 reconciler leaf/host 节点参与布局输出。
- **3.5 Unmount 生命周期期**：把 instance tree 未复用节点的 cleanup 接入 HookStore unmount，覆盖 key/type 改变和 subtree 删除。
- **3.6 动态列表验证期**：补 keyed list 插入、删除、重排时 state 不错位测试，并保留无 key index fallback 行为测试。
- **3.7 Phase 3 收口期**：整理 API 边界、记录暂不支持的 React 能力，明确 Phase 4 Host Component 迁移入口。

#### Phase 3 完成标准

- function component 可由 reconciler 自动创建/复用 `ComponentInstance`。
- keyed sibling 插入、删除、重排不导致已有 item 的 hook state 错位。
- key 改变或 component type 改变会 remount，并触发旧实例 cleanup。
- provider、fragment、host/legacy widget adapter 都能在 reconciler 路径中正常渲染。
- 无 key children 明确保持 index fallback，并有测试记录限制。

#### Phase 3.1 Reconciler 设计

P3.1 只定义最小 reconciler 的职责边界和数据流，不在本阶段改变 Gio 的逐帧 layout/ops 模型。

核心边界：

- Reconciler 只管理 `Element` 树的身份、组件实例复用、hook slot 绑定、provider 作用域和 unmount lifecycle。
- Gio 仍然每帧执行 layout；reconciler 输出仍是当前 legacy `widget.Widget`，由 Gio 在本帧布局和绘制。
- 不引入 React DOM 式 commit/update 队列，不做 host diff patch，不尝试跳过 Gio layout。
- 旧 `Widget`、`Context.Persistent`、`Context.Memo`、legacy state/effect 继续保留，host/adapter 节点作为 reconciler leaf。

最小数据结构：

```go
type Reconciler struct {
	root *FiberNode
}

type FiberNode struct {
	ID       string
	Kind     FiberKind
	TypeID   string
	Key      string
	Position int
	Parent   *FiberNode
	Children []*FiberNode
	Instance *internal.ComponentInstance
	Element  Element
}

type FiberKind string

const (
	FiberComponent FiberKind = "component"
	FiberFragment  FiberKind = "fragment"
	FiberProvider  FiberKind = "provider"
	FiberHost      FiberKind = "host"
)
```

身份规则：

- component fiber 使用 `parent fiber ID + component type + key` 复用；无 key 时使用 sibling position fallback。
- fragment/provider fiber 不创建 hook instance，但会进入 tree，提供 child 边界、key 透传和 provider scope。
- host fiber 默认不创建 hook instance；后续 Phase 4 host component 迁移时可挂载 host state。
- 同 parent 下 keyed child 优先按 key 匹配；无 key child 按 kind/type/position fallback。
- type 或 key 变化表示旧 fiber unmount、新 fiber mount。

每帧流程：

1. `Runtime.BeginFrame()` 重置 HookStore active 标记。
2. `RunElement` 调用 `reconciler.Render(ctx, rootComponent)`。
3. Reconciler 从 root component fiber 开始，按当前 frame 的 `Element` 结果递归 reconcile children。
4. 进入 function component 时，由 reconciler 调用 `HookStore.BeginInstance(identity)`，并用 `ctx.WithComponentInstance(instance)` 执行组件函数。
5. 进入 provider 时，使用 `internal.WithProviderValue` 派生子上下文，只影响 provider 子树。
6. 进入 fragment/key 节点时展开或附加 key 元数据，再继续 reconcile child。
7. 进入 host/legacy widget adapter 时停止递归，返回当前 frame 的 `widget.Widget`。
8. `Runtime.EndFrame()` 继续由 HookStore 清理本帧未 active 的 component instance，并运行 pending effects。

接管当前手动路径：

- 当前 `renderComponent(ctx, "root", 0, "", root)` 是临时桥接点。
- P3.2 将新增 internal reconciler 入口替代这个手动函数：`RunElement` 不再直接调用 `renderComponent`，而是通过 root fiber 管理 root component instance。
- `renderComponent` 可先改为 reconciler 的薄封装，待 function component 子节点全部接入后删除。
- `HookStore.BeginInstance` 保持底层 API，但调用方从 `ui.renderComponent` 迁移到 reconciler。

Unmount 设计：

- Reconciler 负责判断 fiber 是否被复用；未复用的 fiber subtree 标记为 detached。
- component unmount 的实际 cleanup 仍由 HookStore 根据 active instance 做统一处理，避免 duplicated cleanup。
- 后续如 host state 需要显式释放，再为 host fiber 增加 release hook。

P3.1 完成标准：

- Reconciler 的最小职责、数据结构、身份规则和 per-frame 流程已记录。
- 明确 `renderComponent` / `HookStore.BeginInstance` 的接管方案。
- 明确 Gio 渲染模型不变：每帧仍返回 `widget.Widget` 并执行 layout/ops。
- 为 P3.2 function component 接入留下清晰实现入口。

#### Phase 3.7 收口记录

Phase 3 已完成最小 reconciler 的目标，但仍保持实验 API 边界。

已完成的能力：

- `RunElement` 通过 root reconciler 管理 root function component instance。
- `ComponentElement` 允许 function component 出现在 fragment / provider / keyed children 下。
- `Fragment`、`Key`、`Provider`、host / legacy `FromWidget` 都能进入 reconciler 路径。
- keyed sibling 使用 `kind + type + key` 匹配，插入、删除、重排时已有 item state 不错位。
- 无 key sibling 明确保持 index fallback，并有测试记录限制。
- key 改变、component type 改变、provider subtree 删除、fragment child 删除都会让旧 component instance 在 `HookStore.EndFrame()` cleanup。
- Gio 渲染模型不变：reconciler 仍输出 legacy `widget.Widget`，每帧由 Gio 执行 layout/ops。

当前 API 边界：

- `Element` / `RunElement` / `ComponentElement` / `Provider` / `UseContext` 仍属于实验 React-style API。
- `FromWidget` 是 Phase 3 的主要 host adapter；大部分 UI 控件仍是 legacy `Widget`。
- `RenderElement` 与 `renderElementWithContext` 继续作为 legacy adapter 出口保留，供未进入 reconciler 的旧路径使用。
- `renderComponent` 仅作为兼容薄封装保留；新 root 和 nested component 路径应走 reconciler。
- `componentTypeID` 使用函数指针区分 function component；后续如果引入 props/component factory，需重新定义 type identity 策略。

暂不支持的 React 能力：

- 不支持并发渲染、Suspense、优先级调度或异步 commit。
- 不支持 DOM 式 host diff patch；host 节点仍每帧通过 Gio layout 输出。
- 不支持 class component、memo component、portal 或 error boundary。
- 不支持自动推断业务 key；动态 sibling 需要显式 `Key` 才能避免状态错位。

Phase 3 完成标准复核：

- function component 已由 reconciler 自动创建/复用 `ComponentInstance`。
- keyed sibling insert/delete/reorder 已有测试确认 state 按 key 保持。
- key/type 改变已触发 remount，并通过 HookStore cleanup 旧实例。
- provider、fragment、host / legacy widget adapter 都能在 reconciler 路径中渲染。
- 无 key children 的 index fallback 限制已测试并记录。

#### Phase 4 小阶段拆分

- **4.1 Host Component 设计期**：定义 host element / host state / props / event dispatch / ref 的最小模型，明确不破坏 legacy `Widget`。
- **4.2 基础展示组件迁移期**：优先迁移 `Text`、`Spacer`、`Divider` 这类无内部状态组件。
- **4.3 基础布局组件迁移期**：迁移 `Column`、`Row`、`Padding`、`Center`、`Container` 等布局组件，让 children 由 reconciler 管理。
- **4.4 简单交互组件迁移期**：迁移 `Button`、`ClickArea`、`Checkbox`、`Switch` 等需要 host state / event callback 的组件。
- **4.5 Host state 生命周期期**：为需要 Gio widget state 的 host fiber 增加创建、复用、释放规则。
- **4.6 Legacy adapter 兼容期**：确认新 host component 与 `FromWidget` / 旧 `Widget` 混用边界。
- **4.7 Phase 4 收口期**：整理新 host component API、迁移示例，并明确复杂控件和 Router 的后续入口。

#### Phase 4 入口标准

- 保持 Phase 3 reconciler 测试全部通过。
- 每迁移一个 host component，旧 `widget.*` / `ui.FromWidget` 路径仍可用。
- 新 host component 先覆盖无状态展示和布局，不提前迁移复杂输入、滚动、弹层和 router。
- Host state 如需释放，应接入 fiber lifecycle，但不能影响 component HookSlot cleanup。

#### Phase 4.1 Host Component 设计

P4.1 只定义最小 host component 模型，不急于迁移具体控件实现。

核心目标：

- 让 host component 作为 reconciler 的一种 leaf/host 节点存在，而不是继续只依赖 legacy `Widget`。
- 明确 host state 归属 host fiber，不进入 component HookSlot。
- 明确 event dispatch、refs 和 layout 输出仍然走 Gio immediate-mode 机制。
- 保持旧 `Widget` 与 `FromWidget` 兼容，先设计边界，不破坏现有控件。

建议的最小抽象：

```go
type HostNode struct {
	Kind     string
	Props    any
	Key      string
	State    any
	Children []Element
	Ref      any
}

type HostFactory func(props any, children ...Element) Element
type HostStateFactory func(ctx *Context, node *HostNode) any
type HostRef any
```

设计原则：

- host node 只负责保存与 widget 生命周期相关的数据，例如 Gio clickable、editor、scroll state。
- host state 与 component hook slot 分离；component 的 `UseState` 不应该直接依赖 host state。
- host props 以不可变输入为主，更新时通过 reconciler 重新传递，不做 DOM 式 patch。
- host 事件仍由 Gio widget 在 layout 时注册，事件回调可以通过 props 注入。
- ref 只暴露宿主控件的 imperative handle，不反向污染 reconciler identity。

边界说明：

- `Text`、`Spacer`、`Divider` 这类无状态展示组件优先迁移为纯 host component。
- `Column`、`Row`、`Padding`、`Container` 这类布局组件迁移时，children 仍由 reconciler 提供。
- `Button`、`ClickArea`、`Checkbox`、`Switch` 的 host state 可在后续阶段加入。
- `TextField`、`ScrollView`、`List`、`Grid` 这类复杂组件暂不在 P4.1 处理。

P4.1 完成标准：

- host component 的最小模型、边界和未来迁移顺序已记录。
- host state 与 component HookSlot 的职责分离已明确。
- legacy `Widget` / `FromWidget` 兼容边界已明确。
- P4.2 可以直接按该设计开始迁移第一个无状态展示组件。

### Phase 4：迁移基础组件为 Host Component

目标：逐步减少新 API 对旧 `Widget` adapter 的依赖。

迁移顺序：

1. 无状态布局与展示组件。
2. 简单交互组件。
3. 复杂输入和滚动组件。
4. 弹层、通知和路由组件。

每个组件迁移时需要明确：

- props 定义。
- controlled/uncontrolled 行为。
- 内部 host state。
- event dispatch 时机。
- ref 映射规则。
- 对应旧 API 兼容方式。

验收标准：

- 每迁移一个组件，都有示例或测试覆盖。
- 旧 API 可以通过 adapter 或 legacy wrapper 继续工作。

### Phase 5：Router React 化

目标：Router 使用 Provider + hooks，替代当前 `Persistent` holder。

任务：

- 新增 `UseNavigate(ctx)`。
- 新增 `UseLocation(ctx)`。
- 新增 `UseParams(ctx)`。
- 新增 React 风格 `Router/Route` API。
- 定义 route instance 复用规则。
- 明确 transition 期间 from/to route 的 mount 状态。

建议初始规则：

- 默认 route key = route pattern。
- 同一 route pattern 参数变化时复用页面实例。
- 用户可显式设置 route key，以实现参数变化时 remount。
- transition 期间 from route 保持 mounted，动画结束后 unmount。

验收标准：

- push、replace、back 行为稳定。
- params/query 可通过 hooks 获取。
- route transition 不造成 hook 状态错位。

### Phase 6：Root 支持 Element

目标：让新 API 成为完整入口，但旧入口仍可用。

任务：

- 提供稳定的 `RunElement`。
- 可选提供 v2 包中的 `Run`，避免 Go 无返回类型重载问题。
- 文档默认推荐 React 风格写法。
- examples 分批迁移。

验收标准：

- 新项目可以完全使用 Element API 开发。
- 旧项目无需立即迁移。

### Phase 7：收敛 legacy API

目标：新 API 稳定后，逐步把旧 API 定位为兼容层。

任务：

- 标记旧 `Widget` API 为 legacy，但不立即删除。
- 保留 `FromWidget` 或等价 escape hatch。
- 将 docs 默认示例迁移为 React 风格。
- 明确 deprecation 策略和版本节奏。

验收标准：

- 用户有清晰迁移路径。
- 旧 API 与新 API 可在一定范围内混用。

## 风险清单

### 组件身份风险

风险：新旧身份模型并存时，state 可能错位或重复创建。

缓解：新 React API 与 legacy API 分层实现，先不强行混用同一 state store。

### Gio 模型冲突风险

风险：过度追求 React diff，试图跳过 Gio 每帧 layout，导致事件和 ops 不正确。

缓解：明确每帧仍 layout host tree，Reconciler 只管理身份、hooks、context 和 lifecycle。

### API 兼容风险

风险：直接修改 `ui.Run`、`ui.Widget` 会破坏现有 examples 和用户代码。

缓解：先新增 `RunElement` 或实验包，不原地替换。

### Router 生命周期风险

风险：transition 同时渲染 from/to 页面时，mount/unmount 和 hook 状态混乱。

缓解：先定义 route key 和 transition 生命周期规则，再实现。

### 复杂控件迁移风险

风险：`TextField`、scroll、select、dialog、toast 等控件内部状态多，迁移容易引入行为差异。

缓解：基础组件先迁移，复杂组件最后迁移，并补充专门测试。

## 进度跟踪

| 阶段 | 状态 | 负责人 | 备注 |
| --- | --- | --- | --- |
| Phase 0 行为基线测试 | 已完成 | OpenCode | 已补状态、hook、ref、router、列表/网格与动态插入删除错位基线测试 |
| Phase 1 React API facade | 已完成 | OpenCode | 已完成 1.1-1.7，实验 Element API、身份信息、渲染出口和验证样例已落地 |
| Phase 2 Hook Slot Runtime | 已完成 | OpenCode | 已完成 2.1-2.6，HookSlot state/effect/provider/legacy 兼容与总体验证已通过 |
| Phase 3 最小 Reconciler | 已完成 | OpenCode | 已完成 3.1-3.7，最小 reconciler、keyed diff、unmount lifecycle 和动态列表验证已收口 |
| Phase 4 Host Component 迁移 | 进行中 | OpenCode | 4.1-4.3 已完成，4.4 进入简单交互组件迁移 |
| Phase 5 Router React 化 | 未开始 |  |  |
| Phase 6 Element Root 入口 | 未开始 |  |  |
| Phase 7 Legacy API 收敛 | 未开始 |  |  |

### 当前进度记录

- Phase 0 已完成：状态、hook、ref、router、基础列表/网格状态持久性，以及 `ListView` / `GridView` 动态插入删除错位基线测试已补齐。
- Phase 1 已完成：`Element`、`RunElement`、`UseState(ctx, initial)`、`FromWidget`、`Key`、`Fragment`、`RenderElement`、`ElementInfo` 已落地并有测试覆盖。
- Phase 1 收口已完成：实验 API 的边界安全测试已补齐，`go test ./...` 已通过。
- Phase 2 已规划：已补 hook slot、状态迁移、effect 生命周期、provider、legacy 兼容和验证期的小阶段拆分。
- Phase 2.1 已完成：新增组件身份 `ComponentIdentity`、组件实例 `ComponentInstance`、hook 槽 `HookSlot`、hook 游标和 `HookStore`，并接入 `Runtime.BeginFrame` / `Runtime.EndFrame` / `Runtime.Dispose`。
- Phase 2.1 验证已完成：新增 keyed 身份、position fallback、实例复用、slot 持久化、unmount cleanup、dispose cleanup 和 Runtime 生命周期测试，`go test ./...` 已通过。
- Phase 2.2 已完成：`Context` 可绑定当前 `ComponentInstance` 并提供 `NextHookSlot`，实验 `ui.UseState(ctx, initial)` 在组件实例上下文中优先使用 `HookSlot`，无实例时继续回退到 legacy path state。
- Phase 2.2 验证已完成：新增 HookSlot-backed `UseState` 持久化、类型错配 panic、无组件实例 legacy fallback 测试，`go test ./...` 已通过。
- Phase 2.3 已完成：`HookStore` 新增组件实例级 effect 队列、依赖比较、deps clone 和 cleanup 管理；实验 `ui.UseEffect` / `ui.UseEffectWithDeps` / `ui.UseMount` 在组件实例上下文中优先使用 `HookSlot`，无实例时继续回退 legacy effect。
- Phase 2.3 验证已完成：新增 HookSlot-backed effect 每帧执行、deps 不变跳过、deps 变化 cleanup+rerun、unmount cleanup、legacy fallback 测试，`go test ./...` 已通过。
- Phase 2.4 已完成：新增 typed provider 存储与继承，公开实验 `ContextKey[T]`、`Provider`、`UseContext`，并让 provider element 在 context-aware render 路径中为子树覆盖上下文值。
- Phase 2.4 验证已完成：新增 default fallback、provider override、nearest provider wins、sibling isolation、provider 内 legacy state fallback 测试，`go test ./...` 已通过。
- Phase 2.5 已完成：补充 mixed-mode 兼容测试，覆盖 legacy `state.UseWithInitial` 与 HookSlot `ui.UseState` 并存、legacy `state.UseEffect` 与 HookSlot `ui.UseEffect` 并存、Provider 内 legacy `Memo` / `Persistent` 稳定性与 hook index 不受污染。
- Phase 2.5 验证已完成：首次测试暴露出 legacy hook count 必须跨帧一致的约束，已按兼容预期修正测试；`go test ./...` 已通过。
- Phase 2.6 已完成：补充 Phase 2 集成验证，覆盖 Provider + legacy state/effect + HookSlot state/effect 在同一子树中并存、跨帧持久化、deps effect 跳过、unmount cleanup 和 key 独立性。
- Phase 2 已完成：新实验 hook 体系已具备组件实例级 HookSlot 生命周期，`UseState` / `UseEffect` 可优先使用 slot 存储，Provider/UseContext 已接入 Element render 路径，legacy `Persistent` / `Memo` / state / effect 保持兼容。
- Phase 3 已规划：已补 3.1-3.7 小阶段拆分，下一步建议进入 Phase 3.1 Reconciler 设计期。
- Phase 3.1 已完成：明确最小 reconciler 只管理 identity、component instance、hook slot、provider scope 和 unmount lifecycle；Gio 仍每帧 layout，reconciler 输出仍是 legacy `widget.Widget`。
- Phase 3.1 设计结论：后续由 root fiber 接管当前 `renderComponent(ctx, "root", 0, "", root)` 手动路径，`HookStore.BeginInstance` 保持底层 API，但调用方迁移到 reconciler。
- Phase 3.2 已完成：新增最小 `reconciler` / `fiberNode`，`RunElement` 已通过 root reconciler 统一管理根 function component 的 `ComponentInstance` 复用，保留 Gio 每帧布局模型不变。
- Phase 3.2 验证已完成：新增 root component instance 复用、root unmount cleanup、reconciler provider context 三个测试，`go test ./...` 已通过。
- Phase 3.3 已完成：reconciler 已能遍历 fragment / key / provider / nested function component，新增 `ComponentElement` 让函数组件可作为子 Element 出现，keyed sibling 按 key+type 复用，无 key sibling 保持 index fallback。
- Phase 3.3 验证已完成：新增 keyed children reorder 状态保持测试，以及 unkeyed children reorder 状态跟随 index fallback 的限制测试，`go test ./...` 已通过。
- Phase 3.4 已完成：host / provider / legacy `FromWidget` adapter 路径已收口，provider 作为 fiber 边界保留并为单 child 建立子 fiber，host leaf 明确清空 children 和 instance，只通过 `renderElementWithContext` 输出 legacy widget。
- Phase 3.4 验证已完成：新增 provider + keyed nested component + legacy host 混合路径测试，确认 provider scope、nested HookSlot state、host layout 和 fiber leaf 形态正确，`go test ./...` 已通过。
- Phase 3.5 已完成：补齐 key 改变、component type 改变、provider subtree 删除、fragment child 删除时的 unmount cleanup 测试，并让非 component fiber 也获得稳定 fiber ID，保证 nested component 的 `ParentID` 能随父 fiber 身份变化。
- Phase 3.5 修正：`componentTypeID` 现在包含函数指针，避免同签名 function literal 被错误视为同一 component type；旧实例未 active 时继续由 `HookStore.EndFrame()` 统一 cleanup。
- Phase 3.5 验证已完成：`go test ./...` 已通过。
- Phase 3.6 已完成：补充 Element/reconciler 动态列表验证，覆盖 keyed insert/delete/reorder 时 state 按业务 key 保持、删除项 cleanup，以及 unkeyed insert/delete/reorder 继续按 index fallback 产生状态错位的限制基线。
- Phase 3.6 结论：新 Element/reconciler 路径已经修复 keyed 动态列表错位问题；无 key children 与 Phase 0 legacy list/grid baseline 一样仍会按位置复用，必须推荐显式 `Key`。
- Phase 3.6 验证已完成：`go test ./...` 已通过。
- Phase 3.7 已完成：完成 Phase 3 收口，记录 API 边界、暂不支持的 React 能力、无 key children 限制，并复核 Phase 3 完成标准全部满足。
- Phase 4 已规划：已补 4.1-4.7 小阶段拆分和入口标准，下一步进入 4.1 Host Component 设计期，优先设计 host element / host state / props / event dispatch / ref 模型。
- Phase 4.1 已完成：定义 host component 最小模型，明确 host state 归属 host fiber、event dispatch/refs 仍走 Gio immediate-mode 机制，并确认 `Text`、`Spacer`、`Divider`、`Column`、`Row`、`Padding`、`Container` 的迁移顺序与边界。
- Phase 4.1 设计结论：P4.2 先迁移 `Text`、`Spacer`、`Divider` 这类无状态展示组件，再逐步迁移布局组件，不提前处理复杂输入、滚动和 router。
- Phase 4.2 已完成基础适配：新增 `TextElement`、`SpacerElement`、`DividerElement`，它们当前作为 reconciler 兼容 Element 包装旧 `Widget`，用于让展示组件先进入 Element API 表面而不改变底层绘制。
- Phase 4.2 验证已完成：新增测试确认 `TextElement`、`SpacerElement`、`DividerElement` 能稳定包装 legacy widget 且可通过 `RenderElement` 渲染，`go test ./...` 已通过。
- Phase 4.3 已完成基础适配：新增 `ColumnElement`、`RowElement`、`StackElement`、`CenterElement`、`PaddingElement`、`ContainerElement`，布局组件现在也能先以 Element 形式进入 API，底层仍包装 legacy widget。
- Phase 4.3 验证已完成：新增测试确认布局 Element 包装 legacy column/row/stack/center/padding/container widget，`go test ./...` 已通过。
- Phase 4.4 已完成基础适配：新增 `ButtonElement`、`ClickAreaElement`、`CheckboxElement`、`SwitchElement`，简单交互组件先以 Element 形式进入 API，底层仍包装 legacy widget。
- Phase 4.4 验证已完成：新增测试确认交互 Element 可包装 legacy button/click area/checkbox/switch widget，`go test ./...` 已通过。

#### Phase 4.5 Host State 生命周期

P4.5 的目标不是立刻把所有 host 控件改成真正的 host-state fiber，而是先把 host state 的生命周期边界写清楚，避免后续迁移把 widget 内部状态和 component HookSlot 混在一起。

生命周期规则：

- host state 的所有权属于 host fiber，不属于 function component hook slot。
- 同一个 host fiber 在 key / type / sibling position 复用时，应复用其内部 host state。
- host fiber 因 key/type 变化或 subtree 删除而失活时，host state 必须释放。
- host state cleanup 不应触发 component HookStore cleanup，二者分开管理。
- legacy widget 的现有 ref/command queue 机制可以继续作为 host state 的临时实现载体。

推荐的 host state 分类：

- ephemeral state：单帧临时状态，例如布局计算缓存、测量缓存。
- interactive state：可跨帧复用的按钮、输入框、滑块、复选框状态。
- imperative state：`ButtonRef`、`InputRef`、`CheckboxRef`、`SwitchRef`、`ScrollRef` 这类命令型引用。

释放边界：

- component unmount：只清理 component HookSlot 和 effect cleanup。
- host unmount：只清理 host state / ref / command queue。
- shared adapter wrapper：如果 host 只是 legacy widget 的薄包装，则释放规则沿用旧 widget 的生命周期。

设计结论：

- P4.6 需要把这些规则写成兼容矩阵，明确新 host component 与 `FromWidget` / 旧 widget 共享状态时的责任归属。
- 当前阶段仍可以保留 legacy widget 实现，但必须明确未来 host fiber state 的归属接口。

#### Phase 4.6 兼容矩阵草案

| 场景 | 推荐归属 | 说明 |
| --- | --- | --- |
| 新 host component 纯渲染 | host fiber | 例如未来的 host `Text` / `Divider` 直接维护自己的轻量 state。 |
| 新 host component + 事件 | host fiber | 点击、hover、toggle、输入等状态应留在 host state。 |
| 旧 legacy widget 包装 | legacy widget | `FromWidget` 继续沿用原有 widget state 与 refs。 |
| function component 内部状态 | component HookSlot | `UseState` / `UseEffect` 继续只管理组件逻辑状态。 |
| shared imperative ref | host fiber 或 legacy widget | 由实际 state 所在层决定，不跨层共享。 |

兼容准则：

- 新 host component 只能向下兼容 legacy widget，不反向要求 legacy widget 感知 host fiber。
- 当同一 UI 能力同时存在 host component 和 legacy widget 两种实现时，优先保留旧行为，再逐步切换入口。
- `FromWidget` 永远是安全逃生口，不能因为 Phase 4 而收紧为不兼容路径。

P4.6 完成标准：

- host / legacy widget / component 的状态责任矩阵已记录。
- `FromWidget` 与未来 host component 的兼容边界已明确。
- 旧 API 仍是可用的 transition path，而不是一次性弃用目标。

#### Phase 4.7 收口与交接

P4.7 的目标不是继续扩大迁移范围，而是把 Phase 4 的成果收束成清晰的 API 边界、迁移示例和后续路线图，方便下一阶段从基础组件平滑进入复杂控件与 Router。

收口内容：

- 汇总 Phase 4 已迁移的 host component 表面：`Text`、`Spacer`、`Divider`、`Column`、`Row`、`Stack`、`Center`、`Padding`、`Container`、`Button`、`ClickArea`、`Checkbox`、`Switch`。
- 明确这些组件当前仍以 legacy widget wrapper 形式落地，Phase 4 的价值在于统一 Element API 表面，而不是立刻替换底层实现。
- 记录 host state、component HookSlot、legacy widget 三层职责分离，避免后续把复杂控件直接塞进 component state。
- 保留 `FromWidget` 作为兼容出口，确保旧示例和旧调用方式在迁移期内继续可用。
- 为复杂输入、滚动、弹层、通知和 Router 预留入口，不在 Phase 4 收口期强行展开实现。

建议迁移示例：

- 新项目优先使用 `Element` + `Component` + `RunElement` 作为入口，`FromWidget` 仅用于临时适配旧控件。
- 展示组件和布局组件优先采用 Element 表面，以便把树结构统一交给 reconciler 管理。
- 交互组件在保留旧行为的同时，逐步把事件和状态责任标清楚，为后续真正 host state 化做准备。

P4.7 完成标准：

- Phase 4 的边界、迁移顺序和兼容规则已收口到文档。
- 已明确哪些组件只是 wrapper、哪些能力要留到后续复杂控件阶段。
- 后续 Router 迁移路径已提前规划，不再和 Phase 4 混在一起。

## 推荐下一步

优先进入 Phase 4.7：

1. 整理 Phase 4 的 API 边界和迁移示例。
2. 确认 Phase 4 的完成标准已被满足。
3. 为复杂控件、输入、滚动和 router 的后续迁移预留入口。
4. 保留 `Widget` / `FromWidget` 兼容路径，确保现有 examples 不受影响。

### Phase 5：Router React 化（预备）

目标：把现有 Router 从 `Persistent` holder 风格逐步迁移到 Provider + hooks 的 React 风格 API，同时继续保持 push / replace / back 语义稳定。

阶段拆分建议：

1. 路由状态查询 hooks：`UseNavigate`、`UseLocation`、`UseParams`。
2. Router / Route 组件化：把 route 定义变成 Element 树的一部分。
3. route identity 规则：默认 pattern key、参数变化复用、显式 key remount。
4. transition 生命周期：from/to 页面并存、动画结束后统一 unmount。
5. Router 与现有 legacy API 兼容：保留现有 `ui.Router`、`Navigate`、`RouteParams` 包装层。

P5 入口标准：

- Phase 4 已完成收口，复杂控件迁移边界清晰。
- Router 现有行为已通过基线测试覆盖，便于对比新旧 API。
- Provider / HookSlot / reconciler 的基础能力已可支持 route 级状态管理。

P5 初始验收标准：

- push、replace、back 的行为不变。
- params / location 可通过 hooks 稳定读取。
- route transition 期间不会产生 hook 状态错位。

#### Phase 5.1 Router hooks 基础

P5.1 先补齐 Router 的查询 hooks 和 API 表面，不急于把整个 Router 改成完全新的组件树实现。当前目标是让新 hooks 先建立在现有路由状态之上，保证迁移期间旧 API 与新 API 可以并存。

本阶段交付：

- `UseNavigate(ctx)`：返回绑定当前上下文的导航函数，供组件内部直接触发 push / replace / back 风格导航。
- `UseLocation(ctx)`：返回当前路由位置，至少包含完整路径、纯路径和查询参数访问。
- `UseParams(ctx)`：与现有 `RouteParams(ctx)` 保持同源，作为 React 风格的参数读取入口。
- `ui/router.go` 同步暴露这些 hooks，保留 `ui.Navigate` / `ui.NavigateReplace` / `ui.NavigateBack` 作为兼容层。

实现原则：

- 新 hooks 直接读取现有 router state 和 route view，不额外引入新的路由状态系统。
- `UseLocation` 与 `UseParams` 在 route view 内优先返回当前页面的上下文信息，保证 transition 期间页面自身读取到的是当前路由位置。
- 旧的 `CurrentPath` / `RouteParams` / `CanGoBack` / `StackDepth` 继续可用，方便现有示例和文档逐步迁移。

P5.1 完成标准：

- Router hooks 可以在现有 route 页面和普通上下文里稳定工作。
- 路由例子可以逐步从旧查询 API 切到 `UseNavigate` / `UseLocation` / `UseParams`。
- 旧导航 API 不受影响。

#### Phase 5.2 Router / Route 组件化

P5.2 开始把路由定义从纯 legacy `Route` 列表推进到 React 风格组件化表述。当前阶段仍以兼容层为主，不要求彻底替换现有 `ui.Router` 入口，而是先让 `RouterElement` / `RouteElement` 成为新写法的入口。

本阶段目标：

- 新增 `RouterElement(...)`，用 Element 方式声明路由树。
- 新增 `RouteElement(path, component, opts...)`，用组件函数作为页面内容入口。
- 提供 route key 选项，支持显式 remount。
- 保持 route pattern 默认复用，同一路由模式参数变化时尽量复用页面实例。
- 继续保留 legacy `ui.Router` / `ui.Route` / `ui.Navigate` 调用路径，避免一次性切换。

实现约束：

- Route 组件化先作为现有 Router 的上层封装，不重新发明路由状态系统。
- 页面组件的状态仍然走现有 reconciler + HookSlot，route 只是决定组件何时复用、何时 remount。
- transition 生命周期仍沿用现有 Router 实现，先保证 from/to 页面并存规则稳定。

P5.2 完成标准：

- 可以用 Element API 声明路由表，并在页面组件中直接使用 hooks。
- route key / pattern 复用规则已明确并有测试覆盖。
- 旧路由入口继续可用，示例迁移可渐进进行。

#### Phase 5.3 Route identity 规则

P5.3 明确组件化 route 的实例身份规则，避免动态参数变化时出现不必要 remount，同时允许用户通过显式 key 控制 remount。

规则：

- 默认 route identity 基于 route pattern，例如 `/users/:id`。
- 同一 route pattern 下，仅参数或 query 变化时复用页面组件实例。
- 设置 `RouteKey(key)` 后，route identity 额外包含显式 key。
- 显式 key 变化时视为新 route instance，旧页面组件应在帧结束时 unmount cleanup。
- route identity 只影响页面 component HookSlot，不改变 legacy router 的 push / replace / back 栈语义。

实现结论：

- `RouteElement` 页面组件使用 `route:<pattern>` 作为 component type identity。
- 默认 parent identity 为 `route-element:<pattern>`，因此参数变化不会改变 HookSlot 实例。
- 显式 key 会进入 parent identity 与 component key，确保 key 改变时 remount。

P5.3 完成标准：

- 默认 pattern key 复用规则已有测试覆盖。
- 显式 `RouteKey` remount 规则已有测试覆盖。
- 规则已记录，后续 transition 生命周期可以在此基础上继续收口。

#### Phase 5.4 Transition 生命周期

P5.4 明确路由 transition 期间 from/to 页面组件的 mounted 状态和 cleanup 时机，避免动画期间读取 location / params 或 hook state 时出现错位。

规则：

- 无 transition 或 transition 已结束时，只 layout 当前目标页面。
- transition active 时，同时 resolve/layout from 页面和 to 页面。
- from 页面在 transition active 期间保持 mounted，继续拥有自己的 route view、location、params 和 HookSlot state。
- to 页面在 transition active 期间也保持 mounted，并读取目标 route view。
- transition duration 结束后，只保留 to 页面；from 页面不再 active，后续由 HookStore frame cleanup 释放旧 component effect。

验证结论：

- transition 起始帧会同时 layout from/to 页面。
- transition 中间帧会继续同时 layout from/to 页面。
- transition 到达 duration 后只 layout destination 页面。
- from/to 页面各自读取到自己的 `CurrentPath` / route view，不共享错误上下文。

P5.4 完成标准：

- transition active 期间 from/to mounted 规则已有测试覆盖。
- transition 结束后 destination-only layout 规则已有测试覆盖。
- 后续可以进入 legacy API 兼容与示例迁移阶段。

#### Phase 5.5 Legacy API 兼容收口

P5.5 只收口兼容边界，不迁移 examples。目标是确认 Router React 化之后，旧 `ui.Router` / `ui.Route` / `ui.Navigate` / `ui.RouteParams` 仍然是稳定可用的兼容层，便于现有项目继续运行并分阶段迁移。

兼容结论：

- 旧 `ui.Router` 继续作为当前稳定入口，内部仍走现有路由状态与 transition 逻辑。
- `ui.Navigate` / `ui.NavigateReplace` / `ui.NavigateBack` 继续有效，且可与 `UseNavigate` 并存。
- `ui.CurrentPath` / `ui.RouteParams` / `ui.CanGoBack` / `ui.StackDepth` 继续可用，供旧页面和迁移中的页面共存。
- `RouterElement` / `RouteElement` 只是新写法，不要求马上替代旧 API。
- examples 暂不改动，留到单独的示例迁移计划处理。

迁移原则：

- 新代码优先使用 `RouterElement` + hooks，但旧代码不因 P5 而失效。
- 兼容层只做转接，不额外引入新的行为差异。
- 新旧 API 可以同仓并存，避免为了统一接口而打断现有文档和示例。

P5.5 完成标准：

- 旧 Router API 的兼容边界已记录。
- 已明确 P5 阶段不触碰 examples。
- Phase 5 可以正式收口并进入 Phase 6 规划。

#### Phase 5 收口与 Phase 6 规划

P5 收口时把已完成的小阶段合并成一组清晰的结论，随后进入 Phase 6：Root 支持 Element 的实现规划。

P5 完成结论：

- P5.1：`UseNavigate` / `UseLocation` / `UseParams` 已落地。
- P5.2：`RouterElement` / `RouteElement` 已落地。
- P5.3：route identity 规则已固定。
- P5.4：transition 生命周期已固定。
- P5.5：legacy API 兼容边界已收口，examples 暂不迁移。

Phase 6 规划：

1. 让 `RunElement` 成为更稳定的推荐入口，必要时补充更清晰的运行时包装。
2. 梳理 root 级 Element 入口与旧 `Run` 入口的分工，避免 Go API 设计冲突。
3. 为 examples 的分批迁移单独制定计划，不在本阶段混入。

Phase 6 入口标准：

- Phase 5 已完成收口。
- Router 迁移路径和 legacy 边界已稳定。
- 新 API 可以作为完整入口开始扩展，但旧入口仍保持可用。

### Phase 6：Root 支持 Element

目标：让 `Element` 成为稳定 root 入口，同时保留 `Widget` / `Run` 作为兼容层；examples 迁移另行规划，不在本阶段处理。

阶段拆分建议：

1. Root Element 稳定化：明确 `RunElement` 的推荐边界。
2. Root API 分工收口：记录新旧 root 入口的职责划分。
3. 示例迁移隔离策略：将 examples 迁移从 Phase 6 中剥离。
4. Phase 6 收口与 Phase 7 预备：为 legacy API 收敛做准备。

P6 入口标准：

- Phase 5 已完成收口。
- Router React 化的兼容边界稳定。
- examples 迁移被明确排除在本阶段之外。

P6 初始验收标准：

- `RunElement` 的定位清晰。
- 旧 `Run` / `Widget` 入口继续可用。
- examples 迁移有单独计划，不与 root API 规划混合。

P6.1 当前进度：

- `RunElement` 仍使用 root reconciler 管理根 function component instance。
- root component state / HookSlot 生命周期继续由现有 reconciler + HookStore 负责。
- Phase 6 暂不调整 Router 兼容层，也不触碰 examples 迁移。

#### Phase 6.2 Root API 分工收口

P6.2 明确 root 入口的职责划分，不修改现有函数签名。

API 分工：

- `Run` / `App` / `Window` / `RunMulti`：继续作为旧 `Widget` root 入口。
- `RunElement`：作为 React-style `Element` root 入口，新项目优先使用。
- `FromWidget` / Element wrapper：作为新旧 root 和子树之间的渐进迁移手段。

分工结论：

- 不把 `Run` 原地改成接收 `Element`，避免破坏旧代码和 Go API 签名。
- 不在 P6.2 新增 `AppElement` / `WindowElement`，多窗口 Element root 后续单独规划。
- Router 兼容层不随 root API 分工改变。

P6.2 完成标准：

- root API 分工已记录。
- 旧入口不受影响。
- `RunElement` 的推荐定位更清晰。

#### Phase 6.3 示例迁移隔离策略

P6.3 先评估 examples 的保留、合并和清理策略，并统一低风险书写方式；不在本阶段迁移示例到 `RunElement`。

处理结论：

- 必须保留：`docs_browser`、`counter`、`router`。
- 保留 feature/runtime 示例：`hooks_lifecycle`、`network_request`、`multi_window`、`fonts`。
- 保留 performance/layout 示例：`virtual_scroll`、`horizontal_scroll`、`vscode_layout`。
- 保留 showcase 示例：`team_workspace`、`advanced_components`、`animation_showcase`。
- 暂保留但后续可合并：`animation`、`basic_components`、`layout`、`state_management`、`textfield_demo`、`popup_demo`。
- 可清理候选：空目录 `router_demo`，等待单独 cleanup 任务处理。

本阶段统一内容：

- examples 中统一使用显式 `ui` 导入别名。
- 移除 `team_workspace` 开头 UTF-8 BOM。
- 不改变 examples root API，不删除目录，不迁移到 React-style。

P6.3 完成标准：

- examples 分类与迁移隔离策略已记录。
- 低风险书写方式统一完成。
- 后续示例迁移可以作为独立计划展开。

#### Phase 6.4 Phase 6 收口与 Phase 7 预备

P6.4 收束 root Element 入口规划，并为 Phase 7 legacy API 收敛准备边界。

Phase 6 完成结论：

- `RunElement` 是 React-style root 入口，继续由 root reconciler 管理根 function component instance。
- 旧 `Run` / `App` / `Window` / `RunMulti` 入口继续稳定可用。
- `Widget` root 与 `Element` root 长期共存，通过 `FromWidget` 和 Element wrappers 渐进迁移。
- Router 兼容层保持不变，`ui.Router` 与 `RouterElement` 继续并存。
- examples 已做低风险书写方式统一，但迁移另行规划。

Phase 7 预备路线：

- 将旧 `Widget` API 定位为 legacy 兼容层，而不是立即删除对象。
- 保留 `FromWidget` 作为长期 escape hatch。
- 制定 docs 默认示例迁移到 React-style 的分批策略。
- 明确 deprecation 文案、版本节奏和旧项目不破坏原则。

P6.4 完成标准：

- Phase 6 root 入口结论已收口。
- Phase 7 入口标准已明确。
- 后续可以进入 legacy API 收敛规划。

### Phase 7：Legacy API 收敛

目标：把旧 `Widget` API 明确定位为 legacy 兼容层，同时让 React-style `Element` API 成为新项目的默认推荐路径。Phase 7 不以删除旧 API 为目标。

#### Phase 7.1 Legacy API 定位

P7.1 先明确 legacy API 是稳定兼容入口，而不是立即 deprecated 或删除对象。

定位结论：

- `Widget` / `Run` / `App` / `Window` / `RunMulti` 继续作为稳定 legacy 兼容入口。
- legacy 不等于 deprecated；是否添加 deprecated 文案留到 P7.4 决定。
- 新项目优先使用 `RunElement` / `Element` / `Component`。
- 旧项目无需立即迁移，可继续使用 `Run` / `Widget`。
- `FromWidget` 和 Element wrappers 是新旧混用的渐进迁移手段。

P7.1 完成标准：

- legacy API 定位已记录到 `docs/legacy-api-positioning.md`。
- 旧 API 不改签名，不破坏现有用户代码。
- 新项目推荐路径与旧项目兼容路径已经区分清楚。

#### Phase 7.2 Escape Hatch 策略

P7.2 明确 `FromWidget` 的长期定位和新旧 API 混用边界。

策略结论：

- `FromWidget` 是长期保留的 Widget -> Element 桥接能力，不是临时 hack。
- Element wrappers 可以继续在内部用 `FromWidget` 包装 legacy widget。
- `RenderElement` 是 Element -> Widget 的运行时桥接，主要用于运行时、测试和兼容层。
- legacy widget state 仍归 legacy widget / ref / command queue 所有，不迁入 HookSlot。
- 需要稳定复用时，外层继续使用 Element `Key`。

P7.2 完成标准：

- escape hatch 规则已记录到 `docs/escape-hatch-strategy.md`。
- 新旧 API 混用边界清楚。
- `FromWidget` 不删除、不弱化、不 deprecated。

#### Phase 7.3 Docs 默认示例迁移策略

P7.3 制定 docs 默认示例迁移到 React-style 的分批计划，不在本阶段大规模重写 examples。

迁移顺序：

- Batch 1：无状态展示和基础布局，例如 `text`、`spacer`、`divider`、`row`、`column`、`stack`、`center`、`padding`、`container`、`sizing`。
- Batch 2：简单交互，例如 `button`、`click_area`、`checkbox`、`switch`。
- Batch 3：Router 新写法，保留 legacy 与 `RouterElement` / hooks 对照。
- Batch 4：输入和表单，例如 `textfield`、`select`、`radio_group`、`slider`、`form_validation`。
- Batch 5：滚动、列表、网格、弹层、通知。
- Batch 6：综合 showcase，例如 `docs_browser`、`team_workspace`、`advanced_components`、`animation_showcase`、`vscode_layout`。

回退边界：

- Element wrapper 行为不一致时，legacy 示例继续作为主示例。
- docs_browser 渲染不稳定时，只更新 Markdown，不改 runtime 映射。
- refs / command queue 依赖强的控件后置。

P7.3 完成标准：

- docs/examples 分批迁移策略已记录到 `docs/docs-example-migration-plan.md`。
- 第一批低风险候选已明确。
- 本阶段未删除 legacy 示例，也未大规模重写 examples。

#### Phase 7.4 Deprecation 与版本节奏

P7.4 制定 legacy API 的 deprecation 文案和版本节奏。当前结论是只做文档级推荐，不添加代码级 `Deprecated:` 注释。

策略结论：

- 当前阶段不在代码中添加 Go `Deprecated:` 注释。
- `Widget` / `Run` / `App` / `Window` / `RunMulti` 暂不 deprecated。
- 旧 Router API 暂不 deprecated。
- `FromWidget` 不 deprecated，长期作为 escape hatch 保留。
- 未来只有在 major version 窗口、替代 API、迁移指南和示例都稳定后，才重新评估代码级 deprecation。

版本节奏：

- Current：文档级推荐新项目使用 `RunElement`。
- Next minor：docs 默认示例分批迁移到 React-style，保留 legacy 对照。
- Later minor：根据新 API 稳定性强化 legacy compatibility 文档说明。
- Future major：重新评估是否添加代码级 deprecation。

P7.4 完成标准：

- deprecation 与版本节奏已记录到 `docs/deprecation-and-versioning.md`。
- 当前阶段不添加代码级 deprecated 注释。
- 旧项目迁移压力可控。

## React-style docs/examples rollout

完成 runtime refactor 后，后续工作进入独立的 docs/examples rollout epic。该阶段不再扩大 runtime 重构范围，而是用小批量、可回退的文档和示例更新验证新 API 的对外体验。

### D1.2 最小 RunElement counter 示例

D1.2 新增 `examples/react_counter`，作为 React-style `RunElement` 入口的最小 canonical example。

示例范围：

- 使用 `RunElement` 启动 root function component。
- 使用 `UseState(ctx, initial)` 保存组件状态。
- 使用 `ContainerElement` / `ColumnElement` / `PaddingElement` 组织基础布局。
- 使用 `TextElement` / `ButtonElement` 展示基础 host wrapper 写法。
- 保留 `examples/counter` 作为 legacy `Run` / `Widget` 对照示例。

D1.2 完成标准：

- `examples/react_counter` 可编译并纳入 `go test ./...` 包扫描。
- legacy counter 示例不修改、不删除。
- docs migration plan 已记录该示例作为后续 Markdown snippet 的参考入口。

### D1.1 Batch 1 docs React-style snippets

D1.1 为第一批无状态展示和基础布局 Markdown 文档补充 React-style snippet，同时保留 legacy `Widget` 示例。

覆盖文档：

- `docs/widgets/text.md`
- `docs/widgets/spacer.md`
- `docs/widgets/divider.md`
- `docs/widgets/row.md`
- `docs/widgets/column.md`
- `docs/widgets/stack.md`
- `docs/widgets/center.md`
- `docs/widgets/padding.md`
- `docs/widgets/container.md`
- `docs/widgets/sizing.md`

D1.1 边界：

- 只修改 Markdown 文档，不修改 docs_browser runtime 映射。
- legacy 示例继续保留，React-style 示例作为新项目推荐路径。
- `Fixed` / `Fill` / `Expanded` 的原生 Element wrapper 尚未冻结，`sizing` 文档先记录 `FromWidget` 桥接路径。

D1.1 完成标准：

- Batch 1 文档均包含 legacy 与 React-style 对照示例。
- docs migration plan 已记录 Batch 1 进度和限制。
- `go test ./...` 继续通过。

### D1.3 RouterElement 对照示例

D1.3 新增 `examples/router_element`，作为 React-style Router API 的独立对照示例；legacy `examples/router` 保持不变。

示例范围：

- 使用 `RunElement` 启动 React-style root。
- 使用 `RouterElement` 声明路由容器。
- 使用 `RouteElement` 声明 `/`、`/users/:id`、`/settings` 页面。
- 页面组件中使用 `UseNavigate` 做跳转。
- 页面组件中使用 `UseLocation` 读取 path / pathname / query。
- 页面组件中使用 `UseParams` 读取动态路由参数。

D1.3 完成标准：

- `examples/router_element` 可编译并纳入 `go test ./...` 包扫描。
- `examples/router` 不修改，继续作为 legacy Router 对照。
- docs migration plan 已记录 RouterElement 示例进度。

### D1.4 Docs browser 兼容性检查

D1.4 只做兼容性验证，不改 docs_browser 的 runtime 映射。

验证结论：

- Batch 1 文档的 example id 仍然是 legacy 值，docs browser 继续命中现有示例分支。
- Router 文档仍然指向 `router_basic`，未切换到 `examples/router_element`。
- `examples/docs_browser` 的示例解析逻辑无需因本轮 rollout 变更。

D1.4 完成标准：

- legacy example id 保持稳定。
- docs_browser runtime 映射保持不变。
- 新的 React-style 对照示例仅作为独立 example 暴露。

## Batch 2 prep

Batch 2 将从输入交互类文档开始，范围锁定为 `button`、`click_area`、`checkbox`、`switch`。这批文档会继续保留 legacy 示例，并在 Markdown 中增加 React-style 对照写法。

### 推荐任务拆分

- D2.1 Batch 2 scope lock and docs plan prep
- D2.2 Update `button.md` and `click_area.md` with React-style snippets
- D2.3 Update `checkbox.md` and `switch.md` with React-style snippets
- D2.4 Batch 2 compatibility check

### 约束

- 不修改 `examples/docs_browser` runtime 映射。
- 不新增独立交互示例除非后续发现文档对照不足。
- `examples/react_counter` 继续作为 `Button + UseState` 的 canonical example。

### D2.2 Button / ClickArea docs

D2.2 先更新 `docs/widgets/button.md` 和 `docs/widgets/click_area.md`，补充 React-style 对照写法与 Element API metadata。

完成标准：

- 两份文档同时保留 legacy 与 React-style 示例。
- metadata 中补齐 `ButtonElement` / `ClickAreaElement` 条目。
- legacy example id 仍保持 `button_basic` / `click_area_basic`。
- docs_browser runtime 映射不变。

### D2.3 Checkbox / Switch docs

D2.3 更新 `docs/widgets/checkbox.md` 和 `docs/widgets/switch.md`，补充 React-style 对照写法与 Element API metadata。

完成标准：

- 两份文档同时保留 legacy 与 React-style 示例。
- metadata 中补齐 `CheckboxElement` / `SwitchElement` 条目。
- legacy example id 仍保持 `checkbox_basic` / `switch_basic`。
- docs_browser runtime 映射不变。

### D2.4 Batch 2 compatibility check

D2.4 确认 Batch 2 文档更新不会影响 docs_browser 的旧示例映射。

验证结果：

- `examples/docs_browser` 仍然匹配 `button_basic`、`click_area_basic`、`checkbox_basic`、`switch_basic`。
- docs metadata 的 legacy example id 没有变化。
- `go test ./...` 继续通过。

D2.4 完成标准：

- Batch 2 docs 更新后，旧 docs_browser 仍正常工作。
- 没有引入新的 runtime 映射改动。

## Batch 3 prep

Batch 3 将进入 Router 文档迁移，聚焦 `docs/widgets/router.md` 的 React-style 对照写法，而不是更改 router runtime 本身。

### 推荐任务拆分

- D3.1 Batch 3 scope lock and docs plan prep
- D3.2 Update `router.md` with React-style API metadata and snippets
- D3.3 Add route identity and compatibility notes
- D3.4 Batch 3 compatibility check

### D3.2 Router docs

D3.2 更新 `docs/widgets/router.md`，补充 React-style RouterElement 说明与 snippet，同时保留 legacy Router 文档内容。

完成标准：

- metadata 中补齐 `RouterElement` / `RouteElement` / `RouteKey` / `UseNavigate` / `UseLocation` / `UseParams` 条目。
- 文档保留 legacy Router 的基础用法和导航说明。
- `examples/router_element` 被引用为 React-style 对照示例，`examples/router` 保持不变。
- docs_browser runtime 映射不变。

### D3.3 Route identity and compatibility

D3.3 统一收束 Router 文档中的 route identity 和兼容性说明，确保文档描述与现有测试一致。

确认点：

- `RouteElement` 默认按 pattern 复用页面实例。
- `RouteKey` 会触发 remount。
- legacy `Router` / `Navigate` / `RouteParams` 保持稳定。
- `examples/router` 与 `router_basic` 不受影响。

D3.3 完成标准：

- 文档中明确 route identity 规则。
- 文档中明确 legacy / React-style 兼容边界。
- `go test ./...` 继续通过。

### D3.4 Batch 3 compatibility check

D3.4 做最终兼容性检查，确认 Batch 3 的 Router 文档迁移没有影响旧 docs_browser 入口或 legacy Router 文档行为。

验证结论：

- `examples/docs_browser` 仍然只解析 `router_basic`。
- `examples/router` 仍然作为 legacy 对照。
- `examples/router_element` 仍然作为独立 React-style 对照示例。
- `go test ./...` 继续通过。

D3.4 完成标准：

- Batch 3 文档迁移完成后，旧 Router 文档和 docs_browser 行为不变。
- 没有引入新的 runtime 映射改动。

## Batch 4 prep

Batch 4 是第一个明确依赖 host-state / ref 边界的 docs batch，覆盖复杂输入和表单：`textfield`、`select`、`radio_group`、`slider`。`examples/form_validation` 作为整合参考例保留 legacy 写法，不在 host-state 设计未定时强行转写。

### 推荐任务拆分

- D4.1 Batch 4 scope lock and host-state boundary prep
- D4.2 Update `textfield.md` and `slider.md` strategy notes
- D4.3 Update `select.md` and `radio_group.md` strategy notes
- D4.4 Form validation example compatibility note
- D4.5 Batch 4 compatibility check

### D4.2 TextField / Slider strategy notes

D4.2 更新 `docs/widgets/textfield.md` 和 `docs/widgets/slider.md`，补充 Batch 4 host-state 策略说明，保持 legacy-first，不冻结 React-style API 名称。

完成标准：

- 文档明确 `InputRef` / `SliderRef` 仍是当前兼容路径。
- 不冻结 `TextFieldElement` / `SliderElement` 名称。
- `examples/form_validation` 保持 legacy 对照。
- docs_browser runtime 映射不变。

### D4.3 Select / RadioGroup strategy notes

D4.3 更新 `docs/widgets/select.md` 和 `docs/widgets/radio_group.md`，补充 Batch 4 host-state 策略说明，保持 legacy-first，不冻结 React-style API 名称。

完成标准：

- 文档明确 `SelectRef[T]` / `RadioGroupRef` 仍是当前兼容路径。
- 不冻结 `SelectElement` / `RadioGroupElement` 名称。
- 泛型 Select 的 Element API、展开生命周期和 ref 命令归属留待 host-state 设计阶段处理。
- docs_browser runtime 映射不变。

### D4.4 Form validation compatibility note

D4.4 为 `examples/form_validation` 补充兼容性说明，明确它在 Batch 4 阶段继续作为 legacy `Run` / `Widget` 对照示例。

完成标准：

- 示例目录包含说明文档，解释为什么暂不迁移到 `RunElement`。
- 说明中明确复杂输入 Element API 需要等待 host-state / ref 策略冻结。
- 不修改 `examples/form_validation/main.go` 行为。
- `go test ./...` 继续通过。

### D4.5 Batch 4 compatibility check

D4.5 做最终兼容性检查，确认 Batch 4 的复杂输入文档策略说明没有影响旧 docs_browser 入口或示例行为。

验证结论：

- `examples/docs_browser` 仍然解析 `textfield_basic`、`slider_basic`、`radio_group_basic`、`select_basic`。
- Batch 4 没有新增 `TextFieldElement` / `SliderElement` / `SelectElement` / `RadioGroupElement` metadata。
- `examples/form_validation/main.go` 未修改，`examples/form_validation/README.md` 只记录兼容策略。
- `go test ./...` 继续通过。

D4.5 完成标准：

- Batch 4 以 host-state / ref strategy notes 收尾。
- 复杂输入 React-style snippet 迁移延后到 host-state component migration 阶段。
- 没有引入新的 runtime 映射改动。

## Batch 5 prep

Batch 5 是下一个更偏 lifecycle 的 docs batch，覆盖滚动、列表、网格、弹层和通知：`scroll_view`、`list_view`、`grid`、`dialog`、`popup`、`toast`。这一批涉及虚拟滚动、refs、overlay 生命周期和短时通知状态，建议继续采用策略说明优先、legacy-first 的方式推进。

### 推荐任务拆分

- D5.1 Batch 5 scope lock and lifecycle prep
- D5.2 Update `scroll_view.md` and `list_view.md` strategy notes
- D5.3 Update `grid.md` strategy notes
- D5.4 Update `dialog.md` and `popup.md` strategy notes
- D5.5 Update `toast.md` and `examples/virtual_scroll` compatibility note
- D5.6 Batch 5 compatibility check

### 约束

- 不修改 `examples/docs_browser` runtime 映射。
- 不在 Element wrapper 与 overlay lifecycle 未冻结前强行定义这批控件的 React-style API 名称。
- `examples/virtual_scroll` 保持 legacy 对照参考，必要时只补文档说明，不直接改写为新 runtime。
- 如果后续要引入 `ScrollViewElement` / `ListViewElement` / `DialogElement` 等 API，必须先完成 host-state / overlay lifecycle 设计收束。

### 约束

- 不修改 `examples/docs_browser` runtime 映射。
- 不在 Element wrapper 未冻结前强行定义这批控件的 React-style API 名称。
- `examples/form_validation` 保持 legacy 对照参考，必要时只补文档说明，不直接改写为新 runtime。
- 如果后续要引入 `TextFieldElement` / `SelectElement` 等 API，必须先完成 host-state / ref 设计收束。
