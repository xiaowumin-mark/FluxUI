# A1.2 public API 所有权清单

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 1：包边界和 runtime 基础。

- 状态：Done
- 日期：2026-07-06 14:42:48 +08:00
- 负责人：Codex
- 关注：Runtime、Event、Widget
- 输入命令：
  - `gopls go_package_api` for `internal`、`event`、`ui`、`widget`、`layout`、`style`、`theme`、`system`
  - `go doc github.com/xiaowumin-mark/FluxUI/internal`
  - `go doc github.com/xiaowumin-mark/FluxUI/event`
  - `go doc github.com/xiaowumin-mark/FluxUI/ui`
  - `go doc github.com/xiaowumin-mark/FluxUI/widget`
  - `go doc github.com/xiaowumin-mark/FluxUI/layout`
  - `go doc github.com/xiaowumin-mark/FluxUI/style`
  - `go doc github.com/xiaowumin-mark/FluxUI/theme`
  - `go doc github.com/xiaowumin-mark/FluxUI/system`
- 输入文件：
  - `internal/*.go`
  - `event/*.go`
  - `ui/*.go`
  - `widget/*.go`
  - `layout/*.go`
  - `style/*.go`
  - `theme/*.go`
  - `system/*.go`
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
- 关联能力：
  - public API 所有权
  - 旧 API 兼容边界
  - runtime/event/widget 语义归属

## 所有权规则

- `internal` 导出符号只在 Go `internal` 可见边界内公开；语义由 runtime 基础层维护，不应被外部用户直接视为稳定公共 API。
- `event` 维护事件语义、事件类型转换、listener 选项、focus/keyboard/pointer/input/drag 事件门面；其中别名到 `internal` 的基础类型应保持行为与 runtime 分发一致。
- `widget` 维护组件实现语义、组件 option、Ref、默认行为、状态更新和组合控件行为。
- `layout` 维护基础布局尺寸语义，不维护 widget 默认行为。
- `style` 维护样式数据结构、装饰、状态层、阴影、图片填充、transform 和交互动画参数。
- `theme` 维护颜色、字体、typography、density、shape、icon font 和 interaction quality 语义。
- `system` 维护平台能力、剪贴板、文件对话框、系统通知、消息框、注册、tray、single instance 和系统事件语义。
- `ui` 维护 Element/Component/Hook/Router/RunElement 等公开门面语义；对 `widget`、`event`、`style`、`theme`、`system`、`app`、`router`、`state`、`anim` 的 re-export 不改变原始语义所有权。

## 按包归属清单

| 包 | 公开 API 范围 | 语义维护者 |
| --- | --- | --- |
| `internal` | `Runtime`、`Context`、`PathID`、`MemoryKey`、`WindowID`、`WindowController`、window command methods、frame lifecycle、persistent memory、provider、hook store、diagnostics、perf stats、render/cache helpers、runtime event registry、focus/shortcut/pointer capture primitives、`ClickableState`、foundation specs such as `ButtonSpec`/`InputSpec`/`SurfaceSpec` | `internal`。这是 runtime 基础层；其他包可以适配它，但不应重新定义 runtime 生命周期、path identity、memory 或 dispatch 规则。 |
| `event` | `Type`、`Event`、`TargetID`、`Phase`、`ListenerOption`、`On*` listener helpers、`Dispatch*` helpers、`PointerEvent`、`WheelEvent`、`KeyboardEvent`、`InputEvent`、`CompositionEvent`、`DragEvent`、`ActivationEvent`、focus target/boundary/portal/shortcut helpers、Gio event conversion helpers | `event`。浏览器式事件 API、事件字段映射、listener 选项、默认可取消语义的 public facade 归 `event`；底层 path/dispatch 状态仍由 `internal` 执行。 |
| `widget` | `Widget` interface and widget constructors, component option types, Ref types, controlled/uncontrolled component options, layout wrapper widgets, Material 3 widgets, overlay widgets, list/grid/scroll widgets, pointer/keyboard/event boundary widgets, drag/drop widgets, media/image/icon/text helpers | `widget`。组件默认行为、内部状态、Ref 命令消费、callback 触发条件、组件级 layout/render/event 组合语义归 `widget`。 |
| `layout` | `Axis`、`Dimensions`、`FlexChild`、`StackChild`、`Flex`、`Stack`、`Rigid`、`Flexed`、`Stacked` | `layout`。只维护基础约束输入/输出、axis、flex/stack 尺寸语义。 |
| `style` | `Style`、`Decoration`、`Border`、`Insets`、`CornerShape(s)`、`BoxShadow`、`ShadowLayer`、`ImageFill`、`ImageFillFit`、`LinearGradient`、`Transform2D`、`TransformOrigin`、state layer/elevation/image/color/interaction easing helpers | `style`。只维护样式数据和渲染参数语义，不维护 widget 状态或 runtime 生命周期。 |
| `theme` | `Theme`、`ThemeOption`、`ColorScheme`、`ColorOption`、`TextStyle`、`TypeScale`、`DensityScale`、`ShapeScale`、`FontFace`、`FontSpec`、`IconFont`、`InteractionQuality` and font/icon/color discovery/registration helpers | `theme`。维护设计 token、字体、颜色、density、shape 和主题质量等级语义。 |
| `system` | `Capability`/probe APIs, standard system errors, clipboard, file dialog, drag/drop probe, global shortcut, notification, message box, registration, shell open/reveal, single instance, system events, Toast activator, tray APIs | `system`。维护平台能力和 OS 边界语义；不维护 FluxUI widget 或 runtime 交互规则。 |
| `ui` | `Context` alias, `Component`/`Element`/`Widget` facade, Element builders, hooks, router element APIs, app/window run APIs, system convenience wrappers, and re-exported widget/event/style/theme/system/app/router/state/anim APIs | `ui` for Element/hooks/router/run facade only. Re-exported aliases and forwarding constructors keep semantic ownership in their source package. |

## `internal` 公开 API 所有权

| API 组 | 符号 | 所有权结论 |
| --- | --- | --- |
| Runtime/context/window | `Runtime`、`NewRuntime`、`Context`、`NewContext`、`Frame`、`WindowID`、`WindowController`、`WindowHiddenMemoryPolicy`、`Context.Window*` methods | `internal` 维护 frame、window command 和 runtime 生命周期语义。A2 必须以这些 API 为事实来源。 |
| Path/memory/provider | `PathID`、`MemoryKey`、`Context.NextKey`、`NextMemoryKey`、`ScopeMemoryKey`、`Persistent*`、`Memo`、`ProviderKey`、`ProviderKeyFor`、`WithProvider*`、`Provider*Value` | `internal` 维护组件身份、路径内存和 provider 作用域语义。 |
| Hooks | `HookKind`、`HookSlot`、`ComponentIdentity`、`ComponentInstance`、`HookStore`、`NewHookStore`、`ShouldRunHookEffect`、`DepsEqual`、`CloneDeps`、`EffectSetup` | `internal` 维护 hook slot 生命周期；`ui` hooks 只是公开 facade。 |
| Events/focus/keyboard | `EventType`、`Event`、`EventPhase`、`EventHandler`、`EventListenerOptions`、`FocusEvent`、`KeyboardEvent`、`Modifiers`、`Shortcut`、`FocusTargetOptions`、`Runtime.Register*`、`Dispatch*`、`RequestFocus`/`MoveFocus` primitives | `internal` 维护真实注册表和 dispatch 执行；`event` 维护 public event facade。 |
| Interaction/perf/diagnostics | `InteractionChangeKind`、`InteractionFrameStats`、`PerfDiagnostics`、`PerfSection`、`FrameStats`、`FrameSectionStats`、`EventDiagnosticsStats`、`VirtualizationStats`、`RenderCacheStats`、`FormatFrameStats` | `internal` 维护 runtime 诊断字段和统计口径。 |
| Render/style bridge | `RippleSpec`、`ShadowSpec`、`SurfaceSpec`、`TextSpec`、`ButtonSpec`、`CheckboxSpec`、`SwitchSpec`、`InputSpec`、`SliderSpec`、`FocusIndicatorSpec`、`DrawCheckMark`、`MixNRGBA` | 目前由 `internal` 维护，但与 A1.1 的 `internal -> style/theme` 风险相关；A4 需确认这些 spec 是否属于 runtime 还是 style/widget。 |
| Foundation layout specs | `Axis`、`Alignment`、`Insets`、`FlexChild`、`StackChild` | 当前由 `internal` 暴露基础表示；`layout`/`widget` 负责公开布局组合语义。 |

## `event` 公开 API 所有权

| API 组 | 符号 | 所有权结论 |
| --- | --- | --- |
| Event identity and phases | `Type`、`TargetID`、`Event`、`Phase`、`BoundaryMode`、`BoundaryPolicy`、`ListenerOptions` aliases and constants | `event` 维护 public 命名、兼容和 facade；实际 path/phase 状态来自 `internal`。 |
| Listener registration | `On`、`OnActivate`、`OnFocus`、`OnKeyboard`、`OnKeyDown`、`OnKeyUp`、`OnShortcut`、`OnPointer`、`OnWheel`、`OnInput`、`OnComposition`、`OnDrag`、`Capture`、`Once`、`Passive`、`Priority` | `event` 维护事件注册 API 和 listener option 语义。 |
| Dispatch helpers | `Dispatch`、`DispatchEvent`、`DispatchCustomEvent`、`DispatchKeyboardEvent`、`DispatchPointerEvent`、`DispatchWheelEvent`、`DispatchInputEvent`、`DispatchCompositionEvent`、`DispatchDragEvent`、`DispatchActivationEvent` | `event` 维护 public dispatch facade；runtime dispatch correctness 归 `internal`。 |
| Focus/boundary/portal | `RegisterFocusTarget`、`RequestFocus`、`BlurFocus`、`MoveFocus`、`Focused`、`FocusedTarget`、`FocusManagerFor`、`RegisterBoundary`、`RegisterPortal`、`Focus*` options | `event` 维护 public focus/boundary API；focus registry 归 `internal`。 |
| Pointer/wheel/keyboard/text/drag payloads | `PointerEvent`、`WheelEvent`、`KeyboardEvent`、`InputEvent`、`CompositionEvent`、`DragEvent`、`ActivationEvent`、`Button(s)`、`Modifiers`、`Shortcut`、conversion helpers from Gio | `event` 维护字段映射和 public payload shape。 |
| Legacy compatibility | `Clickable`、`UseClickable`、`ClickHandler`、`HoverHandler` | `event` 维护旧 click/hover bridge 的 public 兼容语义；组件默认行为归 `widget`。 |

## `widget` 公开 API 所有权

| API 组 | 符号 | 所有权结论 |
| --- | --- | --- |
| Core widget model | `Widget`、`Text`、`Static`、`RenderElement` bridge consumers | `widget` 维护 widget contract 和 component-to-widget rendering behavior。 |
| Basic interaction | `Button` variants、`IconButton` variants、`ClickArea`、`Pressable`、`Card` variants、`FloatingActionButton` variants and their option/ref types | `widget` 维护 click/hover/pressed/disabled/loading/ripple/default action and Ref semantics。 |
| Form controls | `Checkbox`、`Switch`、`RadioGroup`、`Select` variants、`TextField`/`FilledTextField`/`OutlinedTextField`、`SearchBar` and option/ref types | `widget` 维护 controlled value、internal state、OnChange、input/focus/validation/default behavior。 |
| Layout and container widgets | `Row`、`Column`、`Stack`、`Center`、`Padding`、`Fixed*`、`Fill*`、`Flexed`、`Container`、`ContainerDecoration`、`Divider`、`Spacer` | `widget` 维护 widget-facing layout composition; raw constraints rules remain shared with `layout`/`internal`。 |
| Scroll/list/grid/navigation | `ScrollView`、`ListView`、`GridView`/`Grid`、`Tabs`、`NavigationDrawer`、`NavigationRail`、`BottomNavigation` and refs/options | `widget` 维护 offset、virtualization hooks, OnChange and interactive state semantics。 |
| Overlay/feedback | `Dialog`、`Popup`、`DropdownMenu`、`Menu`、`Tooltip`、`Toast`、`Snackbar` and option/ref types | `widget` 维护 overlay mount, outside click, focus, animation and callback semantics。 |
| Event/focus wrappers | `PointerArea`、`KeyboardScope`、`EventBoundary`、`EventPortal` and option types | `widget` 维护 component-level registration boundaries; event dispatch semantics remain in `event/internal`。 |
| Drag/drop | `DragSource`、`DropTarget`、`DragPayload`、`DragOperation`、`DragSourceEvent`、`DropEvent` and option types | `widget` 维护 widget drag/drop lifecycle; platform capability probing remains in `system`。 |
| Media/progress/text/icon | `Image`、`Icon`、`Progress*`、`LoadingIndicator`、`Text` and option/source types | `widget` 维护 visual component behavior; styling primitives remain in `style/theme`。 |

## `layout`、`style`、`theme`、`system` 所有权摘要

| 包 | 公开 API | 所有权结论 |
| --- | --- | --- |
| `layout` | `Axis`、`Dimensions`、`FlexChild`、`StackChild`、`Flex`、`Stack`、`Rigid`、`Flexed`、`Stacked` | `layout` owns low-level layout measurement semantics and should not define widget-specific behavior。 |
| `style` | `Style`、`Decoration`、`Insets`、`Border`、`CornerShape(s)`、`BoxShadow`、`ImageFill`、`LinearGradient`、`Transform2D`、state/elevation/color/image/easing helpers | `style` owns visual data and transformation semantics, not event state or runtime lifecycle。 |
| `theme` | `Theme`、`ColorScheme`、`DensityScale`、`ShapeScale`、`TypeScale`、`TextStyle`、`FontFace`、`FontSpec`、`IconFont`、`InteractionQuality` and constructors/options | `theme` owns design token and font/color discovery semantics。 |
| `system` | capability/error APIs, clipboard, file dialogs, drag/drop probe, global shortcut, notification, message box, registration, shell, single instance, system events, toast activator, tray | `system` owns OS/platform integration semantics and must stay independent from widget/runtime behavior。 |

## `ui` facade 所有权矩阵

| `ui` API 来源 | Examples from `go doc ui` | Semantic owner |
| --- | --- | --- |
| Native `ui` Element facade | `Component`、`Element`、`ElementRootBuilder`、`RenderElement`、`ElementKey`、Element builder functions and `RunElement*` | `ui` |
| Hooks/context facade | `Context` alias、`ContextKey`、`NewContextKey`、`UseState`、`UseMemo`、`UseCallback`、`UseEffect*`、`UseMount`、`UseLifecycle`、`UseInterval`、`UseContext`、`UseRef`、`UseAnimatedValue` | `ui` owns facade semantics; state storage and runtime slots are owned by `state/internal/anim` as applicable。 |
| App/window facade | `Run`、`RunMulti`、`App`、`Window`、`WindowElement`、`Window*` functions/options/types | `app` and `internal` own window/runtime execution; `ui` owns public convenience surface。 |
| Widget re-exports | `Widget`、`Button`、`TextField`、`Dialog`、`Popup`、`ScrollView`、`Tabs`、`Select`、`DragSource`、all widget option/ref aliases | `widget` |
| Event re-exports | `Event`、`EventType`/`Type`、`TargetID`、`OnEvent`/`OnActivate`/`OnDrag`、`Dispatch*`、`BoundaryOption`、`Shortcut`、event payload aliases | `event` |
| Style re-exports | `Decoration`、`Style`、`Insets`、`Border`、`CornerShape`、`Bg`、`Pad`、`Shadow`、`TransformDeco`、image/color helpers | `style` |
| Theme re-exports | `Theme`、`ColorScheme`、`TextStyle`、`FontSpec`、`IconFont`、`DensityScale`、theme/color/font constructors and constants | `theme` |
| System wrappers | `OpenFileDialog*`、`SaveFileDialog*`、`ShowMessageBox*`、`CurrentWindowNativeHandle` and system option/result aliases | `system` owns platform semantics; `ui` owns context-aware wrapper convenience。 |
| Router/state/anim re-exports | `RouterElement`、`RouteElement`、`Navigate*`、`UseRoute`、`UseParams`、`State`、`AsyncHandle`、`Animate`、easing/options | `router`、`state`、`anim` respectively; `ui` owns integration surface。 |

## 风险

- `ui` exposes a very broad facade. Without the ownership matrix above, later fixes may accidentally change `widget`/`event`/`theme` semantics through `ui` wrappers instead of editing the owning package。
- `internal` exports many symbols that are only module-internal by Go rules. They should not be treated as end-user public API, but they are still cross-package contracts for `event`、`layout`、`widget`、`ui` and tests。
- A1.1 identified `internal -> style/theme`; A1.2 confirms several `internal` exported specs are style-like. Ownership of those specs remains a follow-up question for A2/A4, not a fix in this task。
- `event` aliases many `internal` types. Changes to internal event fields can become public event API changes through aliases, so compatibility must be checked in both packages。

## 事实结论

- Every exported API surfaced by `go doc` can be assigned to an owning package by the table above。
- `ui` is primarily a public facade and re-export layer; most of its widget/event/style/theme/system symbols should not be considered semantically owned by `ui`。
- Runtime lifecycle, path identity, memory, provider, event registry, focus registry and diagnostics are owned by `internal`。
- Event payload and listener facade semantics are owned by `event`; component default actions and Ref behavior are owned by `widget`。

## 验收

- 已按包记录公开 API 所有权清单。
- 已明确 `ui` re-export 不改变原始包的语义所有权。
- 已标注 `internal` exported API 的特殊边界：模块内部可见，但不是面向最终用户的稳定公共 API。
- 后续修改任一 public API 时，可以从本清单判断应由哪个包维护语义。

## 后续依赖

- A1.3 escape hatch 边界审查应重点查看 `Context.Gtx`、Gio 原始 event、`system` wrapper 和 `ui` facade 的越界入口。
- A1.4 旧 API 兼容矩阵应基于 `event`/`widget`/`ui` 的所有权划分，尤其是 `OnClick`、`OnHover`、`InputOnChange`、`ScrollOnChange`。
- A2/A4 需要继续处理 `internal` 中 style/theme-like exported specs 的所有权归属风险。
