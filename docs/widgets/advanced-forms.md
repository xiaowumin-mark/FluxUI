<!-- fluxui-doc-meta
{
  "id": "advanced_forms",
  "title": "高级表单 Beta",
  "category": "输入交互",
  "order": 280,
  "summary": "受控搜索选择、多选/标签、精确数值和宿主驱动的 Form 校验。",
  "example": { "id": "advanced_forms" },
  "apis": [
    "SearchSelect[T comparable](value T, query string, items []SearchSelectItem[T], opts ...SearchSelectOption[T]) Widget",
    "SearchSelectSelectedKey[T comparable](key string) SearchSelectOption[T]",
    "Combobox[T comparable](query string, items []ComboboxItem[T], opts ...ComboboxOption[T]) Widget",
    "Autocomplete[T comparable](query string, items []AutocompleteItem[T], opts ...AutocompleteOption[T]) Widget",
    "MultiSelect[T comparable](selected []T, options []ChoiceItem[T], opts ...MultiSelectOption[T]) Widget",
    "MultiSelectSelectedKeys[T comparable](keys []string) MultiSelectOption[T]",
    "MultiSelectOnSelectedKeysChange[T comparable](fn func(ctx *Context, selectedKeys []string)) MultiSelectOption[T]",
    "TagPicker(selectedKeys []string, query string, options []TagOptionItem, opts ...TagPickerOption) Widget",
    "TagInput(tags []string, query string, opts ...TagInputOption) Widget",
    "NumericField(value string, opts ...NumericFieldOption) Widget",
    "SpinBox(value string, opts ...NumericFieldOption) Widget",
    "Form(child Widget, opts ...FormOption) Widget",
    "FormField(key string, child Widget, opts ...FormFieldOption) Widget",
    "ValidationSummary(fields []FieldState, opts ...ValidationSummaryOption) Widget",
    "NewComboboxRef[T comparable]() *ComboboxRef[T]",
    "NewMultiSelectRef[T comparable]() *MultiSelectRef[T]",
    "NewNumericFieldRef() *NumericFieldRef",
    "NewFormRef() *FormRef"
  ]
}
-->

# 高级表单 Beta

本页覆盖 R1 的可搜索选择、多选/标签、精确数值与表单呈现 API。它们全部是受控组件：组件只渲染调用方在当前帧提供的同步快照，并通过回调报告用户意图。组件不会在 `Layout` 中请求数据、启动校验或提交 goroutine。

Docs Browser 内嵌演示使用 `advanced_forms`，展示 `SearchSelect`、`Combobox`、`TagPicker`、`NumericField`、`Form` 和可取消的提交意图。完整的可运行表单请参见 [examples/form_validation](../../examples/form_validation)，可通过 `go run ./examples/form_validation` 启动。

## 选择组件

| 组件 | 适用场景 | 不能做什么 |
| --- | --- | --- |
| `Select` | 已有的普通枚举单选。 | 不提供查询输入或真实搜索。旧的 `SelectSearchable`、`SelectQuick`、`SelectTypeaheadDelay` 仍是 Deprecated 的兼容 no-op。 |
| `SearchSelect` / `SearchableSelect` | 有查询框、但业务值只能取自 `items` 的可搜索 Select。`value`、`query` 和 `opened` 都由宿主控制。 | 绝不把任意输入文本提交为业务值。 |
| `Autocomplete` | 提供建议、匹配和选中通知。 | 输入只产生 `OnQueryChange`；不会暗中发起业务查询，也不接受自定义值。 |
| `Combobox` | 可从建议中选值，也可让宿主接受自由文本。 | 默认并不保存自由文本；宿主必须在 `ComboboxOnCustomValue` 中写回自己的状态。 |
| `MultiSelect` | 通用搜索多选器；`[]T` 是兼容值快照，`MultiSelectSelectedKeys` 是稳定 key 选择快照。 | 不会就地改写任一 selected snapshot。 |
| `TagPicker` | 以稳定字符串 key 选择现有标签。 | 不能添加不在 `TagOptionItem` 中的自定义标签。 |
| `TagInput` | 以 `[]string` 管理自由文本标签与查询文本。 | 不会保留重复标签；重复的非空文本不会产生第二个 chip。 |
| `NumericField` / `SpinBox` | 保留原始数值文本，并报告精确十进制解析结果；`SpinBox` 增加加减按钮。 | 不将值降级为 `float64`，也不替宿主提交金额或业务校验。 |

`ComboboxAllowCustomValue(true)` 是 Combobox 的默认设置。关闭它后，Combobox 仍是可编辑查询框，但 `Enter` 不能提交自定义值；需要完全 selection-only 的语义时优先使用 `SearchSelect`。

## 默认值

除构造参数中的业务值外，以下为当前公开 API 的默认快照。`opened`、`pending`、`error` 和所有 selection/value 都不是内部持久业务状态；未显式传入时只是各自的零值。

| 组件 | 默认值 | 说明 |
| --- | --- | --- |
| `SearchSelect` | `opened=false`、`pending=false`、`filter=true`、`placeholder="Search…"`、`maxHeight=280`、不允许自定义值 | 未传 `SearchSelectSelectedKey` 时，组件按 `value` 推导 selected key；当不同 option 共享 `Value` 时必须显式传 key。 |
| `Combobox` | `opened=false`、`pending=false`、`filter=true`、`placeholder="Search…"`、`maxHeight=280`、`allowCustom=true` | 无活动建议时，Enter 可通过 `ComboboxOnCustomValue` 报告非空文本。 |
| `Autocomplete` | 与 Combobox 相同，但 `allowCustom=false` | 只报告建议选择，不提交自由文本。 |
| `MultiSelect` | `query=""`、`opened=false`、`pending/error/disabled=false`、`filter=true`、`placeholder="Search…"`、`maxHeight=280` | 构造参数 `selected []T` 是兼容值快照；需稳定 key 选择时传 `MultiSelectSelectedKeys`（空切片也表示受控空选择）。 |
| `TagPicker` | 与 MultiSelect 的搜索/展示默认值相同 | `selectedKeys` 与 `query` 都是必传受控参数；TagPicker 内部始终以 keys 作为选择权威。 |
| `TagInput` | `pending/error/disabled=false`、`placeholder="Add tag…"`、`maxHeight=280` | 没有 suggestion popup；`TagInputOpened` 保留兼容入口，但不会打开面板。 |
| `NumericField` / `SpinBox` | `integer=false`、`disabled/pending/error/required=false`、无 min/max/step | 未传 `NumericFieldPlaceholder` 时沿用底层 TextField 的 `"Enter text..."`；SpinBox 的缺失、非法或非正 step 按 `1` 处理。 |
| `Form` | `disabled=false`、`pending=false`、无 `OnSubmit` | 没有 handler 时只布局 child，不隐式提交。 |
| `FormField` | `Status=FieldValid`、required/disabled/read-only 均为 `false` | 构造函数传入的 `key` 永远是字段身份。 |
| `ValidationSummary` | title 为 `"Please correct the following fields"`、`emptyText=""`、`disabled=false` | 没有 invalid 字段且未给 `emptyText` 时不渲染内容。 |

## 受控状态和回调

调用方拥有下列快照。回调只报告意图；下一帧是否接受该意图，以调用方重新传入的值为准。

| 状态 | 入口 | 相关回调 | 约定 |
| --- | --- | --- | --- |
| 单选 `value` | `SearchSelect(value, ...)`；普通 `Select(value, ...)` | `SearchSelectOnChange`、`SelectOnChange` | 仅在实际选择既有选项时报告变化。 |
| 单选 `selectedKey` | `SearchSelectSelectedKey`、`ComboboxSelectedKey`、`AutocompleteSelectedKey` | `OnSelect`、`OnActiveChange` | 当多个 option 共享 comparable `Value` 时，以稳定 key 消除选中身份歧义。 |
| 查询 `query` | `SearchSelect`、`Combobox`、`Autocomplete`、`TagPicker`、`TagInput` 的构造参数；`MultiSelectQuery` | `OnQueryChange` | 本地过滤可由 `*FilterOptions` 控制；远程/异步建议由宿主在回调后更新 `items`。 |
| 展开 `opened` | 各搜索选择与 `TagPicker` 的 `*Opened` option | `OnOpenChange` | 打开、关闭、Escape、Tab、外部点击和 Ref 都只请求状态变化。TagInput 没有 popup，其兼容 open/close 命令是 no-op。 |
| 已选项 | `MultiSelect(selected, ...)` / `MultiSelectSelectedKeys(keys)`、`TagPicker(selectedKeys, ...)`、`TagInput(tags, ...)` | `OnChange`、`OnSelectedKeysChange`、`OnToggle`、`OnSelect`、`OnRemove`、`OnAdd` | 回调给出的切片是新快照；宿主负责保存或拒绝它。 |
| 活动项 | 组件内部以 key 保存 roving focus | `OnActiveChange` | 只暴露活动 option key，不把焦点状态误当作业务选择。 |
| `pending` / `error` | 各组件的 `Pending`、`Error` / `ErrorText` option；`FieldState.Status` / `FieldState.Pending` | 无 | 只是宿主提供的呈现状态；error 与 pending 可以并存，不会启动查询、校验或提交。 |

对于可搜索控件，`ChoiceItem.Key` 是 selection 和 roving-focus 身份；`Value` 是选择回调交给宿主的业务值。请给逻辑选项提供非空且稳定的 key，不要从过滤后的索引或排序位置推导它。空 key 或重复 key 会使该结果集进入不可交互的错误状态，而不会降级为按 index 复用状态。`TypeaheadText`（为空时使用 `Label`）是本地大小写不敏感过滤的匹配文本；它不是另一个隐式的业务搜索通道。

当不同选项可能拥有相同的 `Value` 时，必须使用 `MultiSelectSelectedKeys` 与 `MultiSelectOnSelectedKeysChange`；这对 key 提供了完整的受控选择语义。仅使用 `selected []T` 的兼容路径适用于业务 value 本身唯一的集合。

`TagOptionItem.Key` 同时是 TagPicker 的稳定身份和 `selectedKeys` 中的值。过滤、重排或从当前 options 中暂时删除项目时，组件不会重写宿主的 selected snapshot；宿主决定何时移除已不存在的业务值。TagInput 以标签文本作为 chip key，并拒绝重复文本，因此标签的顺序变化不会让焦点或删除目标串位。

### SearchSelect 示例

```go
func FruitPicker(ctx *ui.Context) ui.Element {
    value := ui.UseState(ctx, "apple")
    query := ui.UseState(ctx, "")
    opened := ui.UseState(ctx, false)

    return ui.SearchSelectElement(
        value.Value(),
        query.Value(),
        []ui.SearchSelectItem[string]{
            {Key: "apple", Label: "Apple", Value: "apple"},
            {Key: "pear", Label: "Pear", Value: "pear"},
        },
        ui.SearchSelectOpened[string](opened.Value()),
        ui.SearchSelectOnQueryChange[string](func(ctx *ui.Context, next string) {
            query.Set(next) // 宿主可在这里更新异步建议快照。
        }),
        ui.SearchSelectOnOpenChange[string](func(ctx *ui.Context, next bool) {
            opened.Set(next)
        }),
        ui.SearchSelectOnChange[string](func(ctx *ui.Context, next string) {
            value.Set(next)
        }),
    )
}
```

`Combobox` 与 `Autocomplete` 共享上述 `query`、`opened`、`pending`、`errorText`、`items` 和 `selectedKey` 合同。`ComboboxOnSelect` / `AutocompleteOnSelect` 接收完整 item；Combobox 还可使用 `ComboboxOnCustomValue`。Autocomplete 的输入回调不等同于业务查询提交，应用应自行处理防抖、取消、错误和结果竞争。

## 文本、IME 和提交

SearchSelect、Combobox、Autocomplete、MultiSelect、TagPicker 和 TagInput 都通过现有 Input 实现查询编辑；NumericField 也复用 Input 的文本、历史和 Ref 行为。它们继承 Input 的输入、粘贴、删除、IME composition、撤销/重做、`beforeinput`、`input` 与 `submit` 语义。

- 使用 `*OnBeforeInput` 观察或取消可取消的 `beforeinput`。取消后不会把该编辑作为新的受控 query/tag/numeric raw text 报给组件回调。
- 使用 `*OnInputEvent` 观察 Input 事件；`*OnQueryChange` 和 `NumericFieldOnChange` 是该组件的受控值意图回调。不要在 `*InputOptions` 中试图接管 `InputOnChange`：选择/标签/数值组件自身拥有 value bridge。
- `*OnSubmit` 先观察 Input submit。若事件已被 `PreventDefault`，组件不会执行随后默认的自定义 Combobox 值或 TagInput 加标签动作。
- IME 正在 composition 时沿用 Input 的提交保障；不要通过键盘监听自行绕过 Input submit 或把 composition 中的 `Enter` 当成业务提交。
- NumericField 的 `NumericFieldOnChange` 先收到接受的原始文本，随后同一次变更调用 `NumericFieldOnParsedChange`。后者即使面对不完整或非法的中间输入也会报告 `NumericValue{Text, Valid, Value, Error}`，使宿主不会丢失用户草稿。

TagInput 的 `Enter` 仅对非空、此前不存在的标签报告 `OnChange`、`OnAdd` / `OnTagAdd`，然后请求宿主把 query 清空。Autocomplete 和 SearchSelect 不会因为 Enter 产生自由文本业务值。

## 键盘、Overlay 和焦点

以下表格适用于 SearchSelect、Combobox、Autocomplete、MultiSelect 和 TagPicker；TagInput 使用同一 Input/Overlay 基础，但 Enter 用于添加标签。

| 操作 | 行为 |
| --- | --- |
| 聚焦输入或点击触发区 | 调用 `OnOpenChange(true)`；宿主传回 `opened: true` 后才以受控快照显示面板。 |
| `ArrowDown` / `ArrowUp` | 在 enabled options 间移动按 key 保存的活动项，并调用 `OnActiveChange`；同时请求打开面板。 |
| `Home` / `End` | 移到第一个 / 最后一个 enabled option，并请求打开面板。 |
| `Enter` / `Space` | 有活动 enabled option 时，SearchSelect/Autocomplete/Combobox 选中它；MultiSelect/TagPicker 切换它并保持菜单可用。Combobox 在没有活动项时可走 `OnCustomValue`；TagInput 的 Enter 走添加标签。 |
| `Escape` | 共享 Dropdown overlay 请求 `OnOpenChange(false)`，并把焦点恢复到触发区。 |
| `Tab` / `Shift+Tab` | 请求关闭面板，但不阻止默认 Tab 焦点移动。 |
| 面板外点击 | 请求关闭；若焦点仍在 overlay 内，overlay 清除该焦点而不伪造一个业务选择。 |
| disabled | 不接受用户激活、回调或 Ref 变更；已入队命令在 disabled 帧被消费/丢弃，不能绕过 disabled。 |

选择一个建议会关闭单选面板；组件会通过底层 Input Ref 在下一帧请求编辑器焦点。多选/TagPicker 的切换保持菜单打开，便于连续选择。筛选、删除和重排时活动项通过稳定 key reconciliation；若活动项消失，组件不会把焦点错误地转移到旧索引所代表的另一个选项。

## NumericField 和 SpinBox

`NumericField` 的 `value string` 始终是原始受控文本。`ParseNumericValue(text)` 可让宿主在组件外使用同一精确十进制表示；成功时 `NumericValue.Value` 是规范化的精确十进制文本。

- `NumericFieldMin` 与 `NumericFieldMax` 设置包含边界；`NumericFieldInteger(true)` 要求解析结果是数学整数。
- `NumericFieldStep` 设置 SpinBox 的正精确增量；不存在、非法或非正的 step 回退为 `1`。
- `SpinBox` 的增减按钮遵守 disabled、范围、step 与精确十进制语义；`NumericFieldRef` 则使用同一套原始文本受控 Input 命令，因此宿主仍可观察并处理不完整或非法草稿。
- `NumericFieldError`、`NumericFieldErrorText`、`NumericFieldPending`、`NumericFieldRequired` 是宿主状态与呈现 option；它们不替代业务校验。pending 与 error 同时存在时，错误文本保留，pending 以独立 trailing 指示表达。未提供宿主 error text 时，解析/范围错误仍可作为字段错误显示。

## Ref 命令

Ref 是命令意图，不是绕过受控状态的第二个写入口。命令进入有界队列，在已 attach 的目标下一次 Layout 消费；回调仍是宿主状态更新的唯一途径。disabled 帧不执行这些命令。不要依赖已卸载或重新绑定目标继续保存旧命令。

| 组件 | 精确构造与 attach 入口 | 命令 |
| --- | --- | --- |
| `Select` | `NewSelectRef[T]()` + `SelectAttachRef(ref)` | `SetValue`、`Open`、`Close`、`Toggle`。 |
| `Combobox` | `NewComboboxRef[T]()` + `ComboboxAttachRef(ref)` | `SetQuery`、`Open`、`Close`、`Toggle`、`SelectKey`、`Focus`、`Blur`。 |
| `Autocomplete` | `NewAutocompleteRef[T]()` + `AutocompleteAttachRef(ref)` | 与 Combobox 相同；`SelectKey` 只会选择当前结果中 enabled 的稳定 key。 |
| `SearchSelect` / `SearchableSelect` | `NewSearchSelectRef[T]()` + `SearchSelectAttachRef(ref)` | 与 Combobox 相同，但只能选择 supplied `items`，不会提交自由文本。 |
| `MultiSelect` | `NewMultiSelectRef[T]()` + `MultiSelectAttachRef(ref)` | `SetQuery`、`Open`、`Close`、`Toggle`、`ToggleKey`、`SelectKey`、`RemoveKey`、`Focus`、`Blur`。key 命令对应 `ChoiceItem.Key`。 |
| `TagPicker` | `NewTagPickerRef()` + `TagPickerAttachRef(ref)` | 与 MultiSelect 相同；key 命令对应 `TagOptionItem.Key`。 |
| `TagInput` | `NewTagInputRef()` + `TagInputAttachRef(ref)` | `SetQuery`、`ToggleKey`、`SelectKey`、`RemoveKey`、`Focus`、`Blur`；类型也暴露 `Open` / `Close` / `Toggle`，但 TagInput 没有 popup，不能把这些命令当作面板控制。 |
| `NumericField` / `SpinBox` | `NewNumericFieldRef()` + `NumericFieldAttachRef(ref)` | `SetValue` / `SetText`、`Append`、`Clear`、`Focus`、`Blur`；操作的是原始文本。 |
| `Form` | `NewFormRef()` + `FormAttachRef(ref)` | `Submit()` 发出 submit intention。 |
| `FormField` | `NewFormFieldRef()` + `FormFieldAttachRef(ref)` | `Focus()` 让宿主按稳定 field key 路由摘要焦点。 |

## Form、字段状态和 ValidationSummary

`Form` 只聚合后代 Input 的 submit intention，以及 `FormRef.Submit()` 发出的命令。它不拥有字段值、同步/异步校验器、网络请求或提交任务。

```go
ui.FormElement(
    formBody,
    ui.FormPending(pending),
    ui.FormOnSubmit(func(ctx *ui.Context, event *ui.FormSubmitEvent) {
        if !validSnapshot {
            // 对来自 Input 的提交，这也 PreventDefault 底层 Input submit。
            event.PreventDefault()
            return
        }
        // 宿主决定是否开始自己的异步提交，并在后续帧传回 pending/error。
    }),
)
```

`FormSubmitEvent.FromRef` 区分 `FormRef.Submit()` 和后代 Input 的提交。`PreventDefault` 取消的是这次表单意图（以及存在时的底层 Input submit）；它不会等待、取消或终止宿主已经启动的业务工作。`FormDisabled(true)` 与 `FormPending(true)` 阻止 submit intention，且不会自动禁用任意子控件——需要时把对应的 disabled option 明确传给子控件。

`FormField(key, child, ...)` 以稳定 `key` 包装 label、required、supporting/error/pending text 和 presentation status。构造函数的 key 是身份权威，`FormFieldState` 中同名的 key 不能重定向字段。`FieldState.Pending` / `FormFieldPending` 可与 `FieldInvalid` 共存：error text 保持主消息，pending text 单独显示。FormField 不改变 child 的值；disabled/read-only 也只是 field presentation，子控件如需阻断输入必须接收自己的 option。

`ValidationSummary(fields, ...)` 只显示宿主传入的、带非空 message 的 invalid `FieldState`。激活摘要项会调用 `ValidationSummaryOnFocus(ctx, key)`；宿主可用该 key 调用对应 `FormFieldRef.Focus()`，或完成真实输入焦点与滚动定位。Docs Browser 演示和 `examples/form_validation` 均包含同步取消、宿主 pending/error snapshot 与摘要定位路径。

## 异步职责边界

异步建议、验证、提交、取消和竞争处理都属于宿主。推荐路径是：

1. `OnQueryChange` 只更新 query / 请求标识；宿主自行发起或取消查询。
2. 宿主将最新 `items`、`pending`、`errorText`、`opened` 作为下一帧同步快照传回。
3. 表单校验把结果编码为 `FieldState{Status, ErrorText, Pending, PendingText}`；提交期间传入 `FormPending(true)` 防止重复 intention。
4. 结果过期、组件卸载或业务取消都由宿主处理；组件不创建后台 goroutine，也不猜测数据源。

## 手动 smoke 与视觉矩阵

每次修改本组件族时，至少完成下列矩阵；无法自动化的 GUI 路径应在变更记录中保留实际平台结果。

| 维度 | 最少场景 | 期望结果 |
| --- | --- | --- |
| 键盘与焦点 | Arrow、Home/End、Enter、Space、Escape、Tab/Shift+Tab、disabled、关闭后焦点 | 按上表执行；Tab 不被阻止；Escape 回到触发区；disabled 不产生回调。 |
| 查询与 typeahead | `TypeaheadText`、本地 filter 开/关、宿主异步 suggestions、长标签 | 输入只报告 query；局部过滤不启动业务查询；结果替换后不会按 index 串焦点。 |
| 受控冲突 | 外部 value/query/opened 改变、用户输入、同帧 Ref、overlay 关闭 | 下一帧宿主快照为准；Ref 只报告意图；没有重复 change 或幽灵打开状态。 |
| 多选和标签 | 选项重排、过滤后删除、已选 option 从结果移除、chip 删除 | selected snapshot 不被组件改写；key 保持选择/活动项；TagInput 不重复添加同一文本。 |
| Input / IME | 输入、粘贴、composition、撤销/重做、beforeinput 取消、submit | 保持 Input 事件顺序；取消不会回写受控值；IME Enter 不提前执行业务提交。 |
| Form | 同步错误、宿主 pending/error、`FormRef.Submit`、`PreventDefault`、摘要点击 | 错误摘要只包含 invalid field；取消意图不开始隐式提交；摘要按稳定 key 定位。 |
| 视觉 | Light/Dark、normal/error/required/loading/disabled、长标签、窄宽度、200% DPI | 文字不越界，overlay 宽高受约束，错误/焦点/选中状态仍可辨识。 |
