# A4.1 Decoration 合并规则审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 2：布局、渲染和样式稳定性。

- 状态：Done
- 日期：2026-07-06 19:05:27 +08:00
- 负责人：Codex
- 关注：Style、Widget
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A4\\.1|Decoration|decoration|合并" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md docs/audits/project-audit-baseline.md`
  - `rg -n "type Decoration|Decoration|Merge|merge|Decorat|With.*Decoration|Default.*Decoration|Disabled|Hover|Focus" style widget ui internal`
  - `rg -n "withDefaultStates|resolveDecorationState|componentDecoration|hasAnyDecoration|hasDecorationVisual|\\.Merge\\(|ResolveBg|ResolvePad|ResolveRad|Disabled == nil|Disabled\\.Background" widget style ui -g "*.go"`
  - `rg -n "Decoration|resolveDecorationState|withDefaultStates|Disabled.*Decoration|Hover|Focused|ButtonDecoration|InputDecoration|TabsTabDecoration|SwitchDecoration|CheckboxDecoration" widget style ui internal -g "*_test.go"`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `style/decoration.go`
  - `widget/utils.go`
  - `widget/container.go`
  - `widget/button.go`
  - `widget/input.go`
  - `widget/checkbox.go`
  - `widget/switch.go`
  - `widget/selection.go`
  - `widget/tabs_dialog_toast.go`
  - `widget/material3_components.go`
- 关联能力：
  - Decoration 字段级覆盖规则
  - 交互态 Decoration 显式接管边界
  - Material 3 默认样式 fallback
  - disabled/hover/focus 视觉保护
  - 状态 decoration 动画前清理
  - 用户覆盖风险识别

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| 当前分支 | `main`，相对 `origin/main` ahead 17 |
| `git status --short --branch --untracked-files=all` | 仅输出分支行，无脏文件清单 |
| 判断 | A4.1 执行前工作区干净；本任务只新增 `audit-a4.1-decoration-merge-rules.md` 并更新索引，不修改 style/widget 源码。 |

## 合并规则总览

```text
style.Decoration.Merge(base, other)
  -> other 的非 nil 指针字段覆盖 base
  -> CircleClip 只能从 false 合并为 true
  -> Hover/Pressed/Focused/Disabled 指针字段本身也是整体覆盖

resolveDecorationState(d, hovered, pressed, disabled)
  -> disabled 优先，其次 pressed，其次 hovered
  -> 命中状态时返回 d.Merge(*stateDecoration)
  -> 未命中或状态指针 nil 时返回 d

withDefaultStates(d, hover, pressed, disabled)
  -> 仅当 d.Hover / d.Pressed / d.Disabled 为 nil 时补默认状态
  -> 用户显式提供对应状态后，该状态由用户接管

stripStateDecoration(d)
  -> 动画和绘制前移除 Hover/Pressed/Focused/Disabled 指针
  -> 避免状态子树作为普通 visual decoration 泄漏到渲染层
```

## 默认样式与用户样式矩阵

| 范围 | 默认样式来源 | 用户覆盖方式 | 状态合并规则 | disabled/hover/focus 表达 | 结论 |
| --- | --- | --- | --- | --- | --- |
| `style.Decoration` | 零值表示“使用组件自身默认值”。 | `WithBg`、`WithPad`、`WithRad`、`WithBorder` 等设置指针字段。 | `Merge` 只用 `other` 的非 nil 字段覆盖；未设置字段继续保留 base/default。 | 状态字段是 `*Decoration`，一旦设置会作为该状态的覆盖片段参与 `Merge`。 | 字段级覆盖清晰；风险在状态指针是整体接管语义。 |
| 通用状态解析 | 无组件默认参与。 | 组件把自身 decoration 传入 `resolveDecorationState`。 | 优先级为 disabled > pressed > hovered；当前未处理 focused。 | focused 通常由组件单独处理，或 input 手动合并 `Focused`。 | 状态优先级明确；focused 不是通用解析的一部分。 |
| `ContainerDecoration` | 默认 surface 背景来自 `ctx.Theme().Surface`，其他字段走 `Resolve*` fallback。 | 用户传入完整 `Decoration`，可含 Hover/Pressed/Disabled。 | interactive 模式直接按 disabled/pressed/hover 做 `Merge`，之后 `stripStateDecoration` 再绘制。 | 通用容器没有自动 Material hover/focus/disabled 表达，只有用户显式提供状态时才有状态视觉。 | 这是底层 escape-style API，用户接管是预期语义。 |
| `Button` | `resolveButtonDefaults`、主题颜色、密度 padding、variant border/shadow。 | `ButtonDecoration` 的字段通过 `ResolveBg/ResolvePad/ResolveRad` 覆盖默认；独立的 background/foreground/padding/radius option 仍参与默认值选择。 | 先 `resolveDecorationState`，再按组件逻辑叠加 state layer、disabled、focus。 | disabled 背景只有在 `Decoration.Disabled.Background` 未设置时才用 Material disabled container；前景 disabled 和 focus indicator 独立于 decoration。 | 用户只改普通字段不会丢 disabled/focus；显式 `Disabled.Background` 会接管 disabled 背景。 |
| `Input` | `resolveInputDefaults`、variant border/background、focus/error/disabled 颜色。 | `InputDecoration` 覆盖 bg、border、radius、padding；`Focused` 和 `Disabled` 可作为状态片段。 | disabled 走 `resolveDecorationState`；focused 且有 `Focused` 时手动 `Merge`。 | focused border、error border、disabled fg/border 在 decoration 之后仍由组件修正；仅当 `Decoration.Disabled.Border` 存在时接管 disabled border。 | focus/disabled 主表达有保护；显式状态 border 可接管局部视觉。 |
| `Checkbox` / `RadioGroup` / `Switch` | 控件本体颜色、state layer、ripple、selection progress。 | `*Decoration` 用于外层装饰和交互态。 | `withDefaultStates` 在用户未提供状态指针时补 hover/pressed/disabled decoration。 | 控件本体 disabled 色、state layer/ripple 独立存在；外层 decoration 的默认状态不会因普通字段覆盖丢失。 | 用户普通 decoration 安全；显式 Hover/Pressed/Disabled 表示接管外层状态。 |
| `Select` field | select variant、focused/open/error/disabled 颜色和 md3ActionSurface。 | `SelectDecoration` 通过 `ResolveBg/ResolvePad/ResolveRad` 覆盖 field 表面。 | field 未使用 decoration 状态指针；popup 面板 `SelectMenuDecoration` 只走 `Resolve*`。 | focused/open/error/disabled 由组件逻辑控制；menu item hover/focus/ripple 独立。 | 用户 decoration 不会吞掉 select 的核心交互态。 |
| `Tabs` | item 文本色、indicator、state layer、ripple、focus indicator。 | `TabsDecoration` 控制外层 row；`TabsTabDecoration` 控制单项 bg/pad/rad/border。 | 单项未使用 Hover/Pressed/Disabled 指针；tabBg 来自 state layer 后再交给 `ResolveBg`。 | disabled 文本色、state layer、ripple、focus indicator 独立于 tab decoration。 | 用户可覆盖 tab 背景和尺寸字段，但不会删除 focus/ripple/indicator 逻辑。 |
| Dialog / Popup surface | `materialDialogSurfaceDecoration` 以默认 bg/pad/radius/shadow 为基础。 | `DialogDecoration` / `PopupDecoration` 用 `Resolve*` 替换默认字段，border/shadow 显式接管。 | 这些 surface 不处理 hover/pressed/focused 状态。 | 不是交互控件主状态来源；动画与遮罩逻辑在 overlay 层。 | 用户 decoration 是 surface 定制，不承诺交互态合并。 |
| Menu / Dropdown panel | SurfaceContainer、padding、shape、elevation。 | `MenuDecoration` / `DropdownMenuDecoration` 控制面板 surface。 | 面板 decoration 不处理状态；item 状态由 item row 逻辑生成。 | menu item hover/selected/disabled 不依赖 panel decoration。 | 面板自定义不会吞掉 item 交互态。 |

## 事实结论

- `style.Decoration.Merge` 是字段级“非 nil 覆盖”规则；普通字段未设置时不会抹掉组件默认值，适合用户只改背景、padding、radius、border 等局部样式。
- `Hover`、`Pressed`、`Focused`、`Disabled` 也是 `Decoration` 的指针字段，因此在 `Merge` 层面是整体状态指针覆盖；状态内部再通过 `d.Merge(*state)` 做字段级合并。
- `resolveDecorationState` 当前只统一处理 disabled、pressed、hover，优先级为 disabled > pressed > hovered；focused 由组件单独处理，典型例子是 `Button` 的 focus indicator 和 `Input` 的 focused decoration 手动合并。
- `withDefaultStates` 为 checkbox、radio、switch 这类选择控件补默认 hover/pressed/disabled 外层装饰，但只在用户没有提供对应状态指针时补；用户显式提供状态指针表示接管该状态。
- `Button`、`Input`、`Select`、`Tabs` 等核心组件把必要的 disabled、hover、focus 表达拆在组件逻辑中处理，用户普通 decoration 覆盖不会无意删除这些表达。
- `ContainerDecoration` 是低层通用装饰容器，不自动提供 Material 交互态；如果用户给它绑定交互回调，需要自己提供 Hover/Pressed/Disabled decoration 或接受无额外状态视觉。
- `stripStateDecoration` 在绘制/动画前清除状态指针，避免状态子 decoration 被当作普通视觉属性继续下传。

## 风险

| 风险 | 等级 | 说明 |
| --- | --- | --- |
| 状态指针整体接管语义容易被误解 | 中 | 用户写 `WithHover(style.Decoration{}.WithBg(...))` 时只接管 hover 背景字段，但 `Hover` 指针本身会阻止 `withDefaultStates` 补默认 hover；如果组件依赖默认 hover 外层 decoration，这属于显式覆盖。 |
| focused 未纳入通用 `resolveDecorationState` | 中 | 组件需要自行处理 focus。Button/Input/Tabs 当前有独立 focus 表达，但新组件若只调用 `resolveDecorationState` 可能遗漏 focused decoration。 |
| ContainerDecoration 不提供默认交互态 | 中 | 通用容器绑定 onClick/onHover 后，视觉状态完全取决于用户 decoration；这符合底层 API，但应在文档中明确，避免误以为它自动套 Material state layer。 |
| 缺少状态合并单元测试 | 中 | 当前测试覆盖交互布局、动画和部分 Material defaults，但没有直接断言 `Merge`、`resolveDecorationState`、`withDefaultStates` 在状态指针和 disabled/focus 下的合并结果。 |
| 显式 disabled 局部字段会接管组件保护字段 | 低 | Button 的 disabled 背景、Input 的 disabled border 在用户提供对应字段时会被用户值替换；这是可定制能力，但可能导致可访问性对比度不足。 |

## 验收

| 验收项 | 结果 | 证据 |
| --- | --- | --- |
| 输出默认样式与用户样式合并规则 | 通过 | 已形成“字段级覆盖、状态解析、默认状态补齐、组件独立状态表达”四类规则。 |
| 用户普通自定义不丢 necessary disabled/hover/focus 表达 | 通过 | Button/Input/Select/Tabs 的关键 disabled/focus/ripple/state layer 多数独立于 decoration 普通字段。 |
| 显式状态覆盖边界明确 | 通过 | 记录 `Hover/Pressed/Focused/Disabled` 指针为接管语义，`withDefaultStates` 只在 nil 时补默认。 |
| 低层容器 API 边界明确 | 通过 | `ContainerDecoration` 被标记为用户自主管理状态视觉的底层装饰容器。 |
| 未修改运行时代码 | 通过 | 本任务只新增审计子文件并更新索引。 |

## 后续依赖

- A4.2 交互态视觉来源审查应继续把 focused 未纳入 `resolveDecorationState` 的事实纳入状态来源表，避免新组件遗漏 focus。
- A4.3 ripple 和 state layer 审查应继续确认 Button、Tabs、Select item、选择控件的 state layer/ripple 与 decoration 背景叠加顺序。
- 建议补 `style.Decoration.Merge`、`resolveDecorationState`、`withDefaultStates` 的单元测试，覆盖 nil 状态、显式状态、disabled 优先级、focused 手动合并边界。
- 文档层应补一句：普通 decoration 字段是局部覆盖；状态 decoration 指针是显式接管该状态，组件只在状态指针缺省时补默认状态。
