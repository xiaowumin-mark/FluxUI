<!-- fluxui-doc-meta
{
  "id": "advanced_components_production_roadmap",
  "title": "高级组件补全与生产级演进路线图",
  "category": "工程路线图",
  "order": 49,
  "summary": "在既有 MD3、事件、滚动、Overlay 与回归基线之上，分阶段补全复杂表单、数据密集和桌面工作台组件，并建立生产级质量门。",
  "example": { "id": "material3_showcase" },
  "apis": [
    "ListView(count int, itemBuilder func(ctx *Context, index int) Widget, opts ...ListOption) Widget",
    "DropdownMenu(open bool, trigger Widget, items []MenuItem, opts ...DropdownMenuOption) Widget",
    "KeyboardScope(child Widget, opts ...KeyboardScopeOption) Widget",
    "PointerArea(child Widget, opts ...PointerAreaOption) Widget",
    "VisualRootBuilder(root Component) func(ctx *Context) Widget"
  ]
}
-->

# 高级组件补全与生产级演进路线图

本文档定义 FluxUI 从常规 Material 3 应用组件库演进为可承载复杂业务与桌面工作台的生产级 UI 框架的实施顺序、质量门和发布条件。

它承接已完成的 MD3 基线、项目审查修复和现有示例体系；不重新规划 Button、TextField、Dialog、Menu、基础导航、滚动、事件或 Overlay 的底层重写。新增高级组件必须复用这些已收敛的能力，而不是建立第二套状态、焦点、点击或弹层实现。

## 范围、目标与非目标

### 目标用户与产品能力

首个生产级目标面向桌面 CRUD、管理后台、配置工具和轻量工作台，重点保证以下用户流程可靠：

- 复杂表单：搜索选择、多选、日期时间、数值、校验、文件选择和拖放。
- 数据密集界面：大数据量表格、列操作、分页、树形数据、加载/空/错误状态。
- 工作台界面：可调整分栏、文件/资源树、上下文菜单、工具栏和状态栏。
- 长期维护：可预测的 API、键盘可用性、视觉一致性、性能预算、三平台验证和可审计发布。

### 明确非目标

以下能力拥有独立的数据模型或平台成本，不进入第一个生产级里程碑：

- 富文本/WYSIWYG、完整代码编辑器、电子表格、图表库和 PDF 查看器。
- 网络上传传输层、云端数据源、鉴权或业务查询；组件只消费宿主提供的同步快照与回调。
- VS Code 级完整 Docking、任意窗口浮动和布局编排；首期只做可预测的 SplitPane/ResizablePane。
- 重新实现已存在的 clickable、wheel、focus、portal、ripple、state layer、decoration 或 Ref 队列。

## 当前基线与核心缺口

FluxUI 已具备 MD3 token、主题/密度、Element API、滚动与虚拟 List/Grid、Overlay/focus、输入/IME、拖放、诊断、视觉 smoke 和三平台 CI 基础。常规表单、导航、反馈和展示控件已经可以支撑一般应用。

下一阶段的缺口不是按钮变体，而是以下组件族及其共同契约：

| 领域 | 当前缺口 | 为什么要优先解决 |
| --- | --- | --- |
| 复杂选择与表单 | Combobox、Autocomplete、多选 Select/TagInput、NumericField、FormField/ValidationSummary | `SelectSearchable` 是 Deprecated 的兼容 no-op，不能作为可搜索选择能力对外承诺。 |
| 日期、时间与文件 | Calendar、DatePicker、DateRangePicker、TimePicker、FilePicker/Dropzone UI | 现有 TextField、系统 File Dialog 和 Drag/Drop 可复用，但缺少一致的交互层。 |
| 数据密集 | DataGrid/Table、Pagination、列与行选择、加载/空/错误状态 | Grid/GridView 是布局型网格，并不提供表头、列模型、排序、筛选或行状态。 |
| 信息层级与工作台 | TreeView、TreeTable、SplitPane、Breadcrumb、ContextMenu、Toolbar/MenuBar/StatusBar | 现有示例需要手工用固定宽度 Row/Column 拼出工作台面板。 |
| 完整性体验 | Alert/Banner、Skeleton、EmptyState、Avatar、Stepper、Accordion、ColorPicker | 这些组件耦合较低，适合在核心流程稳定后补齐。 |

## 生产级完成定义

能编译或有一个示例不能视为生产可用。一个高级组件进入稳定公开 API 前，必须同时满足下表要求。

| 维度 | 稳定组件的最低要求 |
| --- | --- |
| API 合同 | Widget 与 Element 两层入口一致；默认值、受控值、回调、disabled/loading、Ref 行为和弃用策略都有文档。 |
| 正确性 | 表驱动测试覆盖正常路径、边界、卸载、列表重排、同帧受控更新/Ref 命令和至少一个反例。 |
| 交互与键盘 | 明确 Tab、Enter、Space、Escape、Arrow、Home/End 行为；Overlay 进入、陷阱、恢复焦点均可测。 |
| 无障碍基础 | 提供 Role、Name、Description、Value/State 语义；Light/Dark 对比度、键盘-only 和 100/125/150/200% DPI 场景纳入验收。当前仅有零散 semantic 输出，不应宣称已达到完整读屏支持。 |
| 性能 | 大数据量下工作量与可见项数量相关；有 benchmark、alloc 与帧预算基线，不允许 idle 持续重绘。 |
| 视觉 | 覆盖 Light/Dark、窄/宽、高低 DPI、hover/focus/disabled/error/loading、长文本与空状态；有截图或可重复人工视觉记录。 |
| 可维护性 | 使用现有包边界和共享 helper；文档、Docs Browser demo、Component Lab/专项示例、迁移说明和局部回滚方案齐全。 |
| 发布 | 公共 API 兼容性可比较，三平台构建可复现，安全扫描、制品校验和版本说明满足当期发布门槛。 |

## 架构约束与共享契约

### 保持现有包边界

- ui 负责声明式 API、Element wrapper 和组合体验。
- widget 负责行为、布局、事件桥接、Ref、默认行为和样式消费。
- event、internal、layout、style、theme 不承载具体业务组件语义。
- 组件跨帧状态绑定 PathID；列表、条件渲染和 Portal 中的状态不得依赖隐式 index。

每个 PR 必须继续执行 COMPONENT_DEVELOPMENT_GUIDE.md 和 FEATURE_INTEGRATION_CHECKLIST.md 的关联性检查。新组件不可为了赶工绕过既有 focus、Overlay、event/default action、scroll hit refresh 或 Ref queue。

### 先冻结的五项共享契约

1. 集合身份与选择

   行、列、选项和树节点必须有稳定 key。选择、展开、编辑和焦点状态按 key 保存，不能按可变 index 保存。抽象应先保持在 widget 内部；至少被两个组件复用后再考虑公开。

2. 复合控件的受控状态

   所有业务值均采用外部 value + OnChange 模式。内部只保存焦点、滚动、临时输入草稿、动画和可见状态。外部值、用户交互和 Ref 命令的优先级必须写入组件文档和测试。

3. 锚定 Overlay

   Combobox、Select、DatePicker、ContextMenu 和菜单必须复用同一套 placement、outside click、Escape、z-order 和 focus restore 规则，不能各自实现保护区域和关闭逻辑。

4. 字段与校验语义

   FormField 负责 label、required、supporting/error text、pending 状态和校验摘要的布局语义；它不拥有业务数据、不在 Layout 发网络请求，也不暗中提交表单。异步校验由宿主更新同步状态快照。

5. 可调整尺寸

   SplitPane 复用 PointerArea/capture 和约束模型，首期只承诺最小/最大尺寸、比例持久化、键盘或指针调整与取消拖拽；Docking 另立项目。

### 不可退让的实现规则

- Layout 中不得执行 I/O、网络、文件读取或启动后台任务。
- Select 只从既有选项中选择；Combobox 允许输入与匹配；两者不可用模糊选项堆叠成不可预测 API。
- 纯日期、时间段和带时区的瞬间是不同值类型；不得用本地午夜 time.Time 冒充纯日期。
- NumericField 保留原始文本、解析结果和校验错误；不得用 float64 直接承担金额语义。
- DataGrid 的排序、筛选、分页、远程加载和编辑提交由宿主受控，组件不猜测数据源。
- 新 Ref 只有在确有脱离声明树的命令时才添加，且必须复用现有有界队列和卸载语义。

## 发布列车与依赖关系

路线图按验收里程碑推进，不按未经确认的人力承诺日期。一个阶段未达到退出条件，不进入下一稳定列车。

| 列车 | 定位 | 主要交付 | 退出条件 |
| --- | --- | --- | --- |
| R0 | 基线与契约 | 共享合同、质量门、API/平台策略 | 两个试点组件已复用共享 helper，关键自动化门已接入。 |
| R1 | 高级表单 Beta | Combobox、Autocomplete、多选、数值、FormField | 键盘/IME/Overlay/受控状态完整回归，可在真实表单示例中使用。 |
| R2 | 日期、文件与反馈 Beta | Calendar、Date/Time、FilePicker/Dropzone、状态反馈 | 日期边界、平台能力和文件拖放有跨平台验收。 |
| R3 | 业务数据 Beta | DataGrid MVP、Pagination、列/行选择 | 大列表基准、横纵滚动、排序/筛选/选择稳定。 |
| R4 | 工作台 RC | DataGrid 规模化、TreeView、SplitPane、ContextMenu 等 | 树、表、分栏和焦点链在工作台示例中稳定。 |
| R5 | v1.0 发布 | 完整性组件、兼容/安全/观测收口 | 全部生产门通过，发布制品和支持矩阵可审计。 |

依赖关系如下：

    共享契约、Identity、Focus、Overlay、质量门
    ├─ 高级表单 ── 日期/文件
    ├─ DataGrid MVP ── DataGrid 规模化 ── Tree/TreeTable/工作台
    └─ 状态反馈与文档回归

## R0：基线、组件合同与工程门

### 范围

- 为集合 identity、selection model、roving focus、可见视口、字段状态和锚定 Overlay 编写设计记录与内部测试夹具。
- 决定 SelectSearchable 的处理：在 R1 落地真实搜索能力，或在落地前明确它不承诺该功能；禁止继续把预留 option 当成已完成特性。
- 建立公共 API snapshot、支持矩阵和弃用记录的单一事实来源；统一 deprecation 文档与 CHANGELOG 的表述。
- 为高级组件补齐 Docs Browser 的可构建示例模板、行为测试模板、视觉状态矩阵和 benchmark 模板。
- 建立发布工程基础：lint、race、覆盖率防回退、依赖/漏洞扫描、制品校验和 release checklist。

### 质量门

- CI 保持现有 Linux/Windows/macOS build、test、vet 和视觉 artifact；新增 gofmt 校验、golangci-lint、核心包 race 测试和覆盖率报告。
- 覆盖率先采用防回退规则；新增或修改代码的有效行为分支目标为至少 80% 自动覆盖。核心包总线先达到 60%，在 v1.0 前提高到 70%，不以 examples 的零覆盖率稀释结果。
- 将现有 benchmark 纳入固定 Go 版本和 benchstat 对比。热点回归超过 10% 或超过已声明的 8ms/16ms 帧预算时必须阻断或经明确批准。
- 定义 Go、Gio、Windows/macOS/Linux、Wayland/X11 的支持矩阵，以及原生能力的 mock 与实际 smoke 边界。

### 退出条件

- 两个试点组件复用同一身份/焦点/Overlay 或字段 helper，证明抽象不是过早的大而全框架。
- 每个 helper 都有可局部回滚的提交边界，internal 不依赖 widget 或 ui。
- 新 PR 模板要求记录关联审查、自动测试、手动 smoke、未覆盖风险和回滚策略。

### 完成状态（2026-07-12）

**完成（工程基线）**。R0 的共享合同、内部夹具、文档/示例模板、API 与弃用治理、CI/release 门均已接入；这不等同于某个 release candidate 已完成 Windows、macOS、Linux/X11、Linux/Wayland 的实际 GUI smoke。后者必须按 `docs/release-checklist.md` 为每个候选版本单独留下证据。

### 完成日志（2026-07-12）

#### 合同与试点

- 已在 `docs/adr/0001-collection-identity.md` 至 `0005-anchored-overlay.md` 冻结集合 identity、选择/roving focus、可见视口、字段状态和 anchored Overlay 的内部合同、非目标、测试夹具与局部回滚边界。
- 新增 `internal/collection`、`internal/fieldstate`、`internal/overlay` 与 `internal/testkit`；它们只依赖 Go/internal 层，不反向依赖 `widget` 或 `ui`。相关单测覆盖 key/选择、roving active、可见范围与 scroll-into-view、字段消息优先级、protected region 和确定性 frame fixture。
- `Select` 与 `DropdownMenu` 已共同复用 `internal/overlay.AnchoredRegion`，作为两个锚定 Overlay 试点的最小共享抽象；组件适配器和 helper 均按目录/文件保持可独立回滚。合并时应按 helper、适配器、文档/CI 分组提交，不把它们与无关功能混合。

#### API、兼容与文档

- R0 决定：`SelectSearchable`、`SelectQuick`、`SelectTypeaheadDelay` 均为 Deprecated 的编译兼容 no-op；它们不提供搜索、typeahead 或 quick-animation。真实搜索能力留给 R1 的 Combobox/Autocomplete。实现注释、Docs Browser、Select 文档、CHANGELOG 和 `docs/deprecation-ledger.md` 已使用同一表述；历史 `SelectQuick` 动画宣称亦已标注为被该决定取代。
- `api/ui.snapshot` 与 `tools/api-snapshot` 构成公共 `ui` API 的可比较事实来源；`docs/production-governance.md` 索引 API、支持矩阵、ledger、质量门和 release checklist 的权威关系。
- 已增加 Docs Browser 可构建模板页与注册示例，以及行为测试、视觉状态矩阵和 benchmark 模板；后续高级组件必须复用这些交付入口，不能把预留 option 或 mock 能力宣传为已完成能力。

#### 发布工程

- CI 保留 Linux/Windows/macOS 的 build、test、vet 和视觉 artifact，并新增 gofmt、API snapshot、golangci-lint 防回退、核心 race、核心覆盖率、依赖清单和 `govulncheck`。PR benchmark job 固定 Go `1.25.12`，对 base/candidate 做五次采样、`benchstat` 比较，并以 `ci/benchmark-gate.json` 阻断 >10% 热点回归或超过 8 ms 帧预算（`performance-approved` 标签是显式例外）。
- Release workflow 生成三平台命名制品、汇总并验证 SHA-256；`docs/release-checklist.md` 记录 API、平台、原生 smoke、供应链、checksum 与回滚签字项。`docs/production-governance.md` 还明确了 Go `1.25.12`、Gio `v0.9.0`、Windows/macOS/Linux/X11/Wayland 的支持矩阵以及 mock/headless 与实际 native smoke 的边界。
- 核心覆盖率基线已升至 **61.33%**（13,234 个语句中的 8,116 个已覆盖；examples 不计入分母），超过 R0 的 60% 首道目标；后续由 `tools/coveragecheck` 防止回退，PR 的新增可执行行门为 80%。各 helper/适配器新增或改动的有效分支仍须在具体 PR 中记录其 80% 覆盖证据。

#### 本地验证记录

- 使用 Go `1.25.12`（Windows/amd64）通过：`go test ./...`、`go vet ./...`、核心包 `go test -race -count=1`、`go run ./tools/gofmtcheck`、`go run ./tools/api-snapshot -check`、`golangci-lint run`、`go mod verify`、`govulncheck ./...`、`actionlint .github/workflows/ci.yml .github/workflows/release.yml`。
- `go test -tags visual ./examples/material3_showcase -run TestMaterial3ShowcaseScreenshots -count=1` 已生成并验证现有视觉入口。
- 本地固定 1 s benchmark 已覆盖 widget 与 `internal/perf`；`benchstat` 自比较和 `tools/benchgate` 通过，受管热点均未超过 8 ms。该自比较只验证本地采集/解析/预算路径；PR CI 的 base/candidate 五次采样才是回归判定证据。

#### 已知边界与回滚

- 本地验证不替代 GitHub 的 Linux/macOS matrix，也不替代 Windows/macOS/Linux-X11/Linux-Wayland 的实际窗口、输入、IME/GPU/compositor 或 native capability smoke；未运行项必须在 release record 中列为风险并获得批准或阻断。
- 若共享合同或锚定适配器造成回归，可单独回滚 `internal/collection`、`internal/fieldstate`、`internal/overlay`、`internal/testkit` 或 Select/Dropdown 的适配文件；公开 API 兼容 shim 暂保留，不能通过重新宣传 no-op option 来回滚。

## R1：高级表单 Beta

### 组件范围

- 真实可搜索的 Select、Autocomplete、Combobox。
- 多选 Select、TagInput/TagPicker。
- NumericField/SpinBox。
- Form、FormField、ValidationSummary。

### API 与行为合同

- Select 只选择选项；Combobox 可输入自由文本或由宿主决定是否允许自定义值；Autocomplete 只提供建议与匹配，不隐式提交业务查询。
- value、query、opened、selected、pending、error 均有清晰受控语义。异步建议、校验和提交由宿主驱动，不在 Layout 或组件 goroutine 内执行。
- 输入、粘贴、IME composition、撤销/重做、beforeinput、submit、Escape 和 Ref 命令的顺序遵循现有 Input 语义。
- 多选与 TagInput 使用稳定 option key；删除选项、过滤结果变化和列表重排不得丢失选择或焦点。

### 验收

- 覆盖 Arrow、Home/End、Enter、Space、Escape、Tab/Shift+Tab、typeahead、disabled 和焦点恢复。
- 覆盖外部值改变、用户输入、同帧 Ref 命令、overlay 关闭、选项重排/删除的组合。
- Form 示例覆盖同步校验、宿主异步 pending/error、提交取消与 ValidationSummary 定位。
- Light/Dark、错误、required、loading、长标签、窄宽度和 200% DPI 进入视觉矩阵。

## R2：日期、文件与状态反馈 Beta

### 组件范围

- Calendar、DatePicker、DateRangePicker、TimePicker。
- FilePicker、Dropzone UI；复用系统 File Dialog 与 DragSource/DropTarget，但不实现网络上传器。
- Alert/Banner、EmptyState、Skeleton、Avatar。

### API 与行为合同

- 日期组件先明确 CalendarDate、时间段与时区瞬间的值类型、locale、首日、禁用日期和范围规则；不要让业务方从显示文本反解析日期。
- FilePicker 明确平台 capability、取消、选择数量、文件元数据和错误语义；外部拖放与原生文件对话框是两条可观测输入路径。
- Skeleton、EmptyState 与 Alert 不改变父级布局的隐式约束，不以持续动画造成 idle redraw。

### 验收

- 日历覆盖跨月、跨年、闰年、范围边界、禁用日、键盘网格导航、Escape、DST/时区约束和长 locale 文案。
- 文件组件覆盖取消、不可用平台、文本/文件 URI/custom MIME drop、大小/类型错误和 disabled 状态。
- 每个状态反馈组件覆盖 loading、empty、error、long text、窄宽度和 screen-scale 场景。

## R3：业务数据 Beta

### DataGrid MVP 范围

- 显式 RowKey、Column.ID、列定义、单元格 renderer、表头和主体同步滚动。
- 受控排序、筛选、分页回调；单选/多选、全选和范围选择。
- Loading、Empty、Error 状态，以及与 Pagination 的组合示例。
- 行虚拟化、横向滚动、列宽最小值和基础列对齐。

### API 与数据合同

- DataGrid 不使用反射猜测字段，不管理远程拉取，不在 Layout 改写数据。
- 排序、筛选、分页均将意图回调给宿主；组件只渲染当前同步快照。
- 选择、编辑预备状态和焦点按 RowKey/Column.ID 保存；删除/重排后不得出现 index 串位。
- 当前 GridView 继续作为卡片型布局组件，不能被包装成隐式 DataGrid 基础。

### 验收

- 50,000 行基准中，每帧构建、布局和绘制量与可见行加 overscan 成正比，而非与总行数成正比。
- 验证表头/主体横向滚动一致、滚动后不移动鼠标的 hover/click 命中刷新、排序/筛选/分页后焦点与选择保持。
- 表格示例包含宽列、长文本、空数据、加载、错误、嵌套滚动和键盘-only 使用路径。

## R4：工作台 RC

### 组件范围

- DataGrid 规模化：固定表头/列、列宽拖拽、列重排、受控行内编辑、复制粘贴和导出 hook。
- TreeView、TreeTable、Breadcrumb、Accordion/Expander。
- SplitPane/ResizablePane、ContextMenu、Toolbar、MenuBar、StatusBar。

### 范围控制

- TreeView 使用显式 node key、受控 expanded/selected 状态、可见节点扁平化和按需加载意图回调；懒加载数据仍由宿主提供。
- SplitPane 只提供分栏、比例、最小/最大尺寸、折叠与取消拖拽。完整 Docking、浮动窗口和标签页编排不在本列车。
- DataGrid 先完成受控编辑事务，再评估批量编辑；冻结列、嵌套滚动和命中区域必须单独测试。

### 验收

- Tree 覆盖 Left/Right/Up/Down、Home/End、展开/收起、类型搜索、禁用节点、重排、卸载和千级节点虚拟化。
- SplitPane 覆盖窗口 resize、最小/最大限制、pointer capture、取消拖拽、比例持久化和键盘操作。
- 工作台示例使用 Tree + DataGrid + SplitPane + ContextMenu，覆盖焦点切换、拖放、长时间运行和恢复布局。

## R5：v1.0 收口与发布

### 组件与文档收口

- 补齐 Stepper、ColorPicker 等低耦合组件；仅在真实需求验证后增加 Carousel、SideSheet 或其他产品化变体。
- 为每个稳定组件提供 docs/widgets 文档、Docs Browser demo、Component Lab 区块、API 快速索引、键盘表、Ref/受控值说明和手动 smoke。
- 清理命名别名的可发现性问题，明确 Pressable/ClickArea、Toast/Snackbar、Progress 相关 API 的推荐名称与兼容定位。

### 生产发布门

- 行为回归：go test ./...、go vet ./...；涉及 runtime、并发、动画、Hook 或 Ref 时执行核心包 go test -race。
- 视觉回归：保留现有 smoke，并升级为受控 Linux 渲染环境中的 golden/perceptual diff；失败截图保留 artifact，更新基线须显式审批。
- 性能回归：DataGrid、Tree、large Select、1,000 visible inputs、scroll+hover、cold/warm text cache、idle redraw 都有固定 benchmark 和预算。
- 兼容性：导出 API snapshot、旧 API compile/runtime 回归、迁移说明和弃用窗口齐全；不在小版本中静默破坏 Widget/Run 入口。
- 平台与供应链：tag 驱动发布、三平台可复现制品、checksum/provenance、SBOM、govulncheck、依赖扫描、许可证检查、SECURITY.md 和漏洞响应流程。Windows 签名、macOS notarization 按发布渠道启用。
- 稳定性与观测：在 PerfDiagnostics 之上提供结构化日志、panic/crash hook、可选指标 sink、隐私与脱敏规则；执行启动/关闭、窗口恢复、文件对话框、拖放、内存、goroutine 和文件句柄的 soak 验收。

## 横向质量门与 PR 准入

每个阶段中的组件 PR 都必须满足以下最小清单：

1. 只修改一个共享能力或一个组件族，并说明包边界与回滚方式。
2. 记录受控值、Ref、disabled/loading、默认行为、PreventDefault、焦点、键盘、滚动、Overlay 和平台能力的影响。
3. 覆盖 pointer 与 keyboard 两条输入路径；复合组件再覆盖重排、卸载、Portal、滚动命中刷新和同帧状态冲突。
4. 提交自动测试、Docs Browser 可构建 demo、专项 example 或 Component Lab 入口，以及无法自动化场景的明确 smoke 步骤。
5. 通过格式、vet、lint、受影响包测试、必要 race、视觉和 benchmark 门；不以手工看起来正常替代行为断言。
6. 新增公开符号时同步 API snapshot、文档、迁移说明和 changelog；破坏性变化必须先经过版本与弃用评审。

## 风险与待决策项

| 风险/决策 | 处理原则 |
| --- | --- |
| 过早抽象集合框架 | 先用两个组件验证 helper，再公开；不以单一 DataGrid 的需求污染所有列表。 |
| DataGrid 范围膨胀 | 先交付受控、虚拟化的 MVP；冻结列、编辑、复制粘贴分列车推进。 |
| 日期与 locale | 先冻结值类型、时区、首日、格式化职责和范围规则，再公开 API。 |
| 完整读屏支持 | 先建立 semantic Role/Name/State 能力矩阵；在平台验证前不夸大可访问性承诺。 |
| 平台差异 | 系统文件、拖放、窗口和字体能力都返回显式 capability；mock 不替代真实平台 smoke。 |
| CI 成本 | benchmark、golden diff 和三平台原生 smoke 分层执行；主分支必须保留最小阻断门。 |
| 发布安全 | 在 v1.0 前完成签名、SBOM、漏洞扫描与 SECURITY 流程；不能把安全工作留到首个公开发布后。 |

## 后续工作的前五项

1. [R0 完成] 建立集合 identity、受控选择、锚定 Overlay、字段状态和支持矩阵的 ADR/治理记录；日期值类型留给 R2 的公开 API 设计。
2. [R0 完成] 将格式、lint、核心包 race、覆盖率报告、API snapshot 和漏洞扫描接入 CI，并确立防回退阈值。
3. [R0 完成] 将 `SelectSearchable` 等预留 option 收敛为 Deprecated 兼容 no-op；R1 再以 Combobox/Autocomplete 试点验证真实搜索、共享 Overlay 与 roving focus。
4. [R1] 使用已经落地的高级表单/DataGrid benchmark、行为和视觉模板，避免先写大组件、后补不可测基础设施。
5. [R4] 以一个工作台示例作为集成验收目标，而不是分别宣布 Tree、Table、SplitPane 已完成。

## 状态维护规则

本路线图中的每个列车只使用 Planned、In progress、Blocked、Done 四种状态。更新状态时必须同时记录：

- 关联 issue/PR、实现范围和已复用的共享契约。
- 自动测试、视觉/性能结果、手动平台 smoke 和未覆盖风险。
- API 或兼容性变化、文档/示例入口与局部回滚方案。

只有达到对应退出条件并完成上述记录，组件或列车才能标记为 Done。
