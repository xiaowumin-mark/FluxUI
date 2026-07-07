# A11.1 docs browser 组合场景审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A11.1 docs browser 组合场景审查。

## 事实结论

审查范围覆盖 `examples/docs_browser` 的主应用、左栏、右栏、Markdown 渲染、示例弹窗和搜索 chips。该示例是 Docs/Test 阶段的主要手动验收入口，目标不是验证单个控件的底层语义，而是验证文档浏览器在真实组合场景下的滚动、弹窗、横向区域和搜索交互是否互相干扰。

| 区域 | 代码入口 | 组合内容 | 当前边界 |
| --- | --- | --- | --- |
| 应用根 | `examples/docs_browser/main.go`、`docs_browser_app.go` | `RunElement` 启动 1360x880 窗口；根节点是 `ThemeProviderElement` + 左右 `RowElement` | 不创建独立测试 harness；作为真实 docs browser 手动入口 |
| 左栏 | `left_panel.go` | 300px 固定宽度；标题、加载状态、`SearchBarElement`、API 搜索摘要、分类 chips、主题控制、虚拟化文档列表 | 左栏文档列表是 `ListViewElement(..., ListVirtualized(true))`，应独立滚动而不拖动右侧正文 |
| 搜索 | `search.go`、`api_search.go`、`left_panel.go` | `searchKeyword` 过滤文档和 API 摘要；选中项缺失时回退到过滤结果首项 | 搜索会重建右侧 `doc-scroll-<id>` key，用于切换文档时重置右栏滚动 |
| 搜索 chips / 分类 | `category_filter.go` | 分类 chips 放在水平 `ScrollViewElement(RowElement(...), ScrollHorizontal(true), ScrollVertical(false))` 中 | 分类行横向滚动，不应消费普通纵向页面滚动 |
| 主题控制 | `docs_theme.go` | Select、Switch、色块按钮、Tooltip 组合 | 可作为左栏 overlay、focus、点击命中和局部重绘的组合验收点 |
| 右栏正文 | `right_panel.go` | 标题、示例区 header、API 索引、Markdown 正文项 | 右侧内容由 `ListViewElement(..., ListVirtualized(true))` 承载，是主文档纵向滚动入口 |
| 代码块 | `markdown_renderer.go` | 带 header、Copy 按钮、深色代码 surface、水平 `ScrollViewElement` | 代码内容只开横向滚动，`ScrollVertical(false)`，不应截停右栏纵向滚动 |
| 表格 | `markdown_renderer.go` | Markdown 表格转为固定列宽 `RowElement`，外层水平 `ScrollViewElement` | 列宽按内容在 88~280px clamp，宽表横向滚动，不应撑爆右栏 |
| 示例弹窗 | `example_viewer.go` | 右栏 header 的打开示例按钮控制 `examplePopupOpen`；根右栏 `StackElement(rightList, examplePopup)` 叠加 `PopupElement` | popup 宽 900、mask alpha 112、`PopupMaskClosable(true)`；弹窗内容固定 420px 高度 |
| API 复制 | `api_index.go`、`markdown_renderer.go` | API 索引和代码块 Copy 调用 clipboard system API 并更新 copy 状态 | clipboard 失败应只影响 copy 状态，不应破坏文档滚动或选中状态 |

### 手动验收表

| 编号 | 场景 | 操作步骤 | 预期结果 | 覆盖点 |
| --- | --- | --- | --- | --- |
| DB-01 | 启动和初始布局 | 运行 `go run ./examples/docs_browser`；观察左栏、右栏、默认文档 | 左栏固定宽约 300px；右栏占剩余空间；默认文档可见；无空白主区域 | 应用根、左右布局 |
| DB-02 | 左栏文档列表滚动 | 在左栏文档列表区域滚轮滚动到较后文档，再点击一个文档 | 左栏列表滚动；右栏切换到被点击文档；点击命中不扩散到整行以外异常区域 | 左栏 ListView、文档选择 |
| DB-03 | 右栏正文纵向滚动 | 在右栏正文区域滚轮滚动，经过标题、示例区、API 索引、正文 | 右栏主内容独立纵向滚动；左栏位置不被带动 | 右栏 ListView |
| DB-04 | 切换文档后滚动重置 | 先把右栏滚到中后部，再从左栏点击另一个文档 | 新文档从顶部附近开始；旧文档滚动位置不误套到新文档 | `doc-scroll-<id>` key |
| DB-05 | 搜索文档标题 | 在左栏搜索框输入 `button`、`input` 等关键词 | 左栏结果和 API 摘要收窄；当前文档不匹配时自动选中首个结果；清空搜索后列表恢复 | SearchBar、filterDocs |
| DB-06 | 搜索 API | 搜索一个 API 片段，例如 `OnChange` 或 `ScrollView`，点击 API 摘要结果 | 能跳转到对应文档；搜索框保持单行输入；右栏显示匹配文档 | API 搜索摘要 |
| DB-07 | 分类 chips 横向滚动 | 在分类 chips 行使用触控板横向滚动或拖动横向 wheel；点击不同分类 | 分类行只横向移动；普通纵向滚轮不应误触发横向滚动；文档列表按分类过滤 | chips row、横向 ScrollView |
| DB-08 | 主题控制 | 切换主题 Select、Dark Switch、点击色块并悬停 Tooltip | 主题变化同步影响左右栏；Select/Tooltip overlay 不污染左栏或右栏滚动；色块命中只限按钮区域 | Select、Switch、Tooltip |
| DB-09 | 代码块横向滚动 | 找到含长代码行的文档，在代码块内横向滚动，再在代码块上普通纵向滚轮 | 长行可以横向查看；纵向滚轮应继续让右栏文档滚动或至少不造成异常横向跳动 | 代码块 ScrollHorizontal |
| DB-10 | 代码块复制 | 点击代码块 header 的 Copy 按钮 | copy 状态显示成功或失败；按钮点击不改变当前文档选择，不打开示例弹窗 | clipboard、局部状态 |
| DB-11 | 表格横向滚动 | 找到 Markdown 表格区域，横向滚动宽表 | 表格列宽被 clamp；宽表横向滚动；右栏不出现异常无限宽或整体错位 | 表格最大宽度策略 |
| DB-12 | 打开示例弹窗 | 点击右栏示例 header 的打开示例按钮 | 弹窗覆盖在右栏 Stack 上；宽度约 900；示例内容在 420px 高视口中显示 | Popup、Stack overlay |
| DB-13 | 弹窗内部交互 | 在弹窗示例内点击按钮、输入、选择或滚动示例内容 | 内部交互只影响示例状态；不误触发外部文档选择或关闭 | overlay 事件边界 |
| DB-14 | 弹窗关闭 | 点击关闭按钮、点击遮罩空白处；再次打开同一示例 | 弹窗关闭并可重新打开；遮罩点击按 `PopupMaskClosable(true)` 关闭；内部按钮点击不误关闭 | outside click、mask close |
| DB-15 | 组合压力 | 搜索过滤后切换分类，打开示例弹窗，关闭后滚动右栏和代码块 | 搜索、分类、弹窗、右栏滚动和横向区域状态互不污染；没有卡死、空白或明显布局跳动 | 组合回归 |

## 风险

- `docsLeftPanel` 的加载数量文本目前在源码中出现编码异常片段；这属于文案显示风险，不影响本次组合交互边界结论，但手动验收时应记录实际显示是否可接受。
- `right_panel.go` 和 `example_viewer.go` 中部分中文文案在当前文件读取结果中呈现乱码；如果运行界面也显示乱码，应作为 Docs 质量问题单独修复。
- 代码块和表格只显式开启横向滚动；父级纵向滚动是否完全顺滑依赖 A6.2/A6.3 记录的 wheel 分发边界，当前缺少 docs browser 级别的自动 wheel 回归。
- 分类 chips 横向区域没有独立自动测试；若未来 chips 数量增加或分类名变长，仍需用 DB-07 覆盖横向溢出。
- 示例弹窗使用 `PopupMaskClosable(true)`，遮罩关闭符合当前文档浏览器预期；但内部复杂示例若含嵌套 overlay，需要继续观察 outside click 是否被内层正确保护。
- API/代码复制依赖系统 clipboard；在无剪贴板或权限受限环境中可能显示失败，这应记录为环境能力差异，不应误判为 docs browser 布局失败。

## 验收

- 已建立 docs browser 手动验收表，覆盖左栏、右栏、代码块、表格、示例弹窗、搜索 chips、主题控制和复制状态。
- 已明确左栏文档列表和右栏正文是两个独立纵向滚动入口。
- 已明确代码块、表格、分类 chips 是横向滚动入口，且普通纵向滚轮不应异常转成横向滚动。
- 已明确示例弹窗的打开、内部交互、关闭和遮罩点击验收规则。
- 已明确搜索和分类过滤后的当前文档选择、右栏滚动重置和空结果状态验收规则。
- 已保留系统 clipboard、编码文案和 wheel 自动回归缺口作为后续风险，而未将其误判为本轮审查引入问题。

## 后续依赖

- A3.3：继续沿用代码块、表格、tabs/chips 横向溢出策略，补 docs browser 真实组合验收结果。
- A6.2 / A6.3：wheel 父子传递和横向滚动策略调整后，需要回归 DB-07、DB-09、DB-11。
- A8.3 / A8.4：Popup outside click、focus 和 Escape 规则调整后，需要回归 DB-12 到 DB-14。
- A10.4：集合控件滚动命中刷新调整后，需要回归左栏 ListView、右栏 ListView、分类 chips 和 Markdown 横向区域。
- A12.x：建议把 DB-01 到 DB-15 中可自动化的场景转成 docs browser 视觉/交互回归，至少覆盖启动、搜索过滤、代码块横向滚动、表格横向滚动和弹窗打开关闭。
