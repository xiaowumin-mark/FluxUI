# A11.4 示例覆盖矩阵

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A11.4 示例覆盖矩阵。

## 事实结论

审查范围覆盖 `examples/advanced_components`、`examples/drag_drop_showcase`、`examples/form_validation`、`examples/horizontal_scroll` 和 `examples/virtual_scroll`。这五个示例共同承担控件族、滚动、拖放、表单、overlay、媒体、导航和虚拟化的真实示例入口，可作为 A11.1/A11.2/A11.3 之外的专项 smoke 补充。

| 示例 | 入口 | 主要控件族 | 底层能力 | 当前覆盖价值 |
| --- | --- | --- | --- | --- |
| `advanced_components` | `go run ./examples/advanced_components` | AppBar/header、Tabs、Select、Image、Card、ListView、ScrollView、BottomNavigation、Dialog、Toast | 组合页面状态、滚动 offset、列表 reach-end、select popup、图片 fit/radius/click、overlay mount/open/close、toast cleanup、底部导航切换 | 综合型控件组合入口，覆盖导航、overlay、媒体、集合滚动和状态联动 |
| `drag_drop_showcase` | `go run ./examples/drag_drop_showcase` | DragSource、DropTarget、ScrollView、Row/Column、ContainerDecoration | Gio transfer、text/file/custom MIME payload、drop active 反馈、operation、error、`system.ProbeDragAndDrop` 后端能力探测 | 拖放控件族和系统能力探测的真实示例入口 |
| `form_validation` | `go run ./examples/form_validation` | TextField、Button、ContainerDecoration、Row/Column | controlled input、`InputOnChange`、password 模式、程序化状态设置、提交验证、错误/成功反馈 | 表单输入、验证状态和 controlled value 的专项入口 |
| `horizontal_scroll` | `go run ./examples/horizontal_scroll` | ScrollView、Card、Row、Column、FixedHeight、ContainerDecoration | 横向滚动、`ScrollHorizontal(true)`、`ScrollVertical(false)`、长行溢出、`ScrollOnChange` x/y offset、横向滚动条形变观察 | 横向滚动策略和横向内容溢出的专项入口 |
| `virtual_scroll` | `go run ./examples/virtual_scroll` | Tabs、ListView、GridView、ContainerDecoration、Expanded | 大数据量虚拟列表、虚拟网格、tab 切换、可见窗口渲染、列表/网格 item builder、grid gap/columns | 高容量虚拟化和集合控件性能/稳定性入口 |

### 示例到控件族覆盖矩阵

| 控件族/能力 | advanced_components | drag_drop_showcase | form_validation | horizontal_scroll | virtual_scroll |
| --- | --- | --- | --- | --- | --- |
| 基础交互 Button/click | Dialog 打开按钮、图片 click | 间接依赖拖放交互 | 提交、重置、快捷填充按钮 | 无按钮，主要观察滚动 | Tabs 切换 |
| 表单输入 TextField | 不覆盖 | 不覆盖 | 用户名、邮箱、密码、确认密码 | 不覆盖 | 不覆盖 |
| 表单选择 Select | 优先级 Select、max height、onChange | 不覆盖 | 不覆盖 | 不覆盖 | 不覆盖 |
| 导航 Tabs/BottomNavigation | scrollable Tabs、BottomNavigation | 不覆盖 | 不覆盖 | 不覆盖 | Tabs 切换 List/Grid |
| 媒体 Image/Card | Image contain/cover、radius、click；Card | 不覆盖 | 状态结果容器 | Card 横向长内容 | 虚拟 item 容器 |
| Overlay Dialog/Toast | Dialog maskClosable、confirm/cancel、Toast duration/onClose | 不覆盖 | 不覆盖 | 不覆盖 | 不覆盖 |
| ScrollView 纵向滚动 | 页面内容滚动、ScrollOnChange | 外层 ScrollView 包裹拖放页面 | 窗口大高度内容，无显式 ScrollView | 不覆盖纵向 | List/Grid 自身滚动 |
| ScrollView 横向滚动 | 不覆盖 | 不覆盖 | 不覆盖 | 两个 horizontal-only ScrollView case | 不覆盖 |
| ListView 虚拟化 | 120 项列表、reach-end | 不覆盖 | 不覆盖 | 不覆盖 | 50,000 项 ListView |
| GridView 虚拟化 | 不覆盖 | 不覆盖 | 不覆盖 | 不覆盖 | 100,000 项 GridView，4 列 |
| DragSource/DropTarget | 不覆盖 | text/file/custom JSON source、typed target、active/error log | 不覆盖 | 不覆盖 | 不覆盖 |
| 状态联动 | tab/nav/select/dialog/toast/reach/scroll state | active/event log state | input/validation/submitted state | x/y offset state | active tab state |
| 系统能力探测 | 不覆盖 | `system.ProbeDragAndDrop` | 不覆盖 | 不覆盖 | 不覆盖 |

### 核心能力真实入口清单

| 核心能力 | 至少一个真实示例入口 | 验收关注 |
| --- | --- | --- |
| 控件组合页面 | `advanced_components` | 多控件同屏状态切换不会互相污染 |
| 表单 controlled value | `form_validation` | 用户输入和按钮程序化设置都能更新受控值；提交时错误/成功状态正确 |
| 输入密码模式 | `form_validation` | 密码字段显示掩码，验证仍使用真实值 |
| Select popup | `advanced_components` | 下拉展开、选择、关闭和 toast/state 联动正常 |
| Dialog mask close | `advanced_components` | 打开、确认、取消、遮罩关闭和 `OnOpenChange` 一致 |
| Toast 生命周期 | `advanced_components` | Toast 出现、自动关闭、清空状态，不推动布局 |
| 图片 fit/click | `advanced_components` | contain/cover、radius、click 反馈正常，资源路径可用 |
| 纵向滚动 offset | `advanced_components` | `ScrollOnChange` 回传真实 y offset |
| 横向滚动 offset | `horizontal_scroll` | `ScrollOnChange` 回传 x offset，y 保持策略明确 |
| 横向内容溢出 | `horizontal_scroll` | 长 row/多行长内容只产生横向滚动，不拉伸撕裂 |
| 列表 reach-end | `advanced_components` | 120 项 ListView 到底触发 reach-end 计数 |
| 大列表虚拟化 | `virtual_scroll` | 50,000 项 ListView 只渲染可见窗口，滚动稳定 |
| 大网格虚拟化 | `virtual_scroll` | 100,000 项 GridView、4 列、gap 稳定 |
| 拖放 payload | `drag_drop_showcase` | text、file URI、custom JSON MIME 都能在目标侧记录 |
| 拖放后端能力 | `drag_drop_showcase` | `system.ProbeDragAndDrop` 显示平台能力，external drag-out 不被默认承诺 |

### 示例 smoke 建议

| 编号 | 示例 | 操作步骤 | 预期结果 |
| --- | --- | --- | --- |
| EX-01 | `advanced_components` | 启动后切换 tabs、BottomNavigation，打开 Select，选择不同优先级 | 当前 tab/nav/select 状态变化，Toast 出现并自动清理 |
| EX-02 | `advanced_components` | 点击图片、滚动主页内容到 ListView 底部，打开/关闭 Dialog | 图片 click 反馈进入 Toast；滚动 offset 更新；reach-end 计数增加；Dialog 不影响主布局 |
| EX-03 | `drag_drop_showcase` | 拖动文本、文件 URI、JSON MIME 卡片到 DropTarget | target active 视觉反馈出现；drop 日志记录 type、operation、bytes/text/path |
| EX-04 | `drag_drop_showcase` | 查看 probe 状态并尝试外部 drag-in/out | probe 文案明确当前平台能力；external drag-out 只在能力为 true 时作为可用入口 |
| EX-05 | `form_validation` | 分别输入无效用户名、邮箱、短密码、不匹配确认密码并提交 | 表单验证失败，错误列表显示对应字段 |
| EX-06 | `form_validation` | 使用快捷按钮填入有效用户名、邮箱、密码和匹配确认密码并提交 | 表单验证通过，绿色成功反馈显示 |
| EX-07 | `horizontal_scroll` | 拖动第一个横向滚动区到底部滚动条，观察 offset 文本 | x offset 增加，内容横向移动，不出现纵向滚动或内容拉伸 |
| EX-08 | `horizontal_scroll` | 在多行长内容区横向滚动 | 多行内容整体横向移动，y offset 不应被纵向 wheel 污染 |
| EX-09 | `virtual_scroll` | 在 ListView tab 快速滚动大量数据 | item 序号连续、无空白、无重复错乱，滚动保持流畅 |
| EX-10 | `virtual_scroll` | 切换 GridView tab 并滚动 | 网格列数稳定、gap 稳定、颜色/编号随 index 正常变化 |

## 风险

- `advanced_components`、`form_validation`、`virtual_scroll` 的 README 仍称它们是 legacy `Run` / `Widget` 兼容示例，但当前 `main.go` 均使用 `ui.RunElement`。这是文档与实现的漂移风险，后续应统一说明其真实入口状态。
- 五个示例当前都没有独立自动断言；`go test` 只能验证编译，不验证 GUI 行为、滚动 offset、拖放 payload 或 overlay 生命周期。
- `drag_drop_showcase` 的外部拖入/拖出依赖操作系统和 Gio 后端能力；示例通过 probe 降低误判，但人工验收仍需区分平台不可用和框架失败。
- `horizontal_scroll` 专注 horizontal-only case，不覆盖 shift-wheel、touchpad 横向 delta 或嵌套父子滚动剩余 delta。
- `virtual_scroll` 覆盖大数据量列表/网格，但没有记录 reach-end、item identity、动态 columns 调整或滚动后 hit refresh。
- `advanced_components` 同时覆盖 Select、Dialog、Toast、ListView 和 ScrollView，适合发现组合问题，但问题定位需要回到 A6/A8/A10 的专项审查。

## 验收

- 已建立五个示例到控件族和底层能力的覆盖矩阵。
- 已确认每个核心能力至少有一个真实示例入口：表单、选择、overlay、toast、图片、纵向滚动、横向滚动、虚拟列表、虚拟网格和拖放均有对应 example。
- 已输出 EX-01 到 EX-10 的手工 smoke 建议，可用于后续示例回归。
- 已识别 README 与入口实现漂移、拖放平台依赖、横向滚动策略覆盖不足、虚拟滚动缺少自动行为断言等风险。
- 已确认本轮只记录 Docs/Test 覆盖关系，不修改示例代码或运行时行为。

## 后续依赖

- A6.2/A6.3/A6.4：wheel 分发、横向滚动策略、滚动后命中刷新变更后，需要回归 `advanced_components`、`horizontal_scroll` 和 `virtual_scroll`。
- A8.1/A8.3/A8.4：overlay mount、outside click、focus/Escape 变更后，需要回归 `advanced_components` 的 Dialog/Toast/Select。
- A9.2/A9.3：controlled value 和 OnChange 触发条件变更后，需要回归 `form_validation` 与 `advanced_components`。
- A10.2/A10.3/A10.4/A10.6/A10.7：表单选择、文本输入、滚动集合、overlay、拖放控件族变更后，需要回归对应示例入口。
- A12.x：建议把 EX-01、EX-03、EX-05、EX-07、EX-09 中稳定可复现的路径转成自动或半自动 smoke。
- Docs 后续清理：需要统一 `advanced_components`、`form_validation`、`virtual_scroll` README 中 legacy 说明与当前 `RunElement` 实现之间的关系。
