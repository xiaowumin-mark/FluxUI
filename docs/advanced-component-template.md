<!-- fluxui-doc-meta
{
  "id": "advanced_component_template",
  "title": "高级组件交付模板",
  "category": "工程路线图",
  "order": 50,
  "summary": "新高级组件在进入公开 API 前使用的 Docs Browser、行为、视觉和性能模板。",
  "example": { "id": "advanced_component_template" },
  "apis": [
    "ui.ComponentElement(component Component) Element",
    "ui.UseState[T any](ctx *Context, initial T) *state.State[T]",
    "ui.SelectElement[T comparable](value T, options []SelectOptionItem[T], opts ...SelectOption[T]) Element"
  ]
}
-->

# 高级组件交付模板

此页是 R0 的可构建 Docs Browser 模板。实际组件进入 Beta 前，复制下列四份模板并替换示例中的 Select 试点控件：

- `examples/docs_browser/advanced_component_template.go`：受控 value、`OnChange`、宿主重置和可观测状态的可构建示例。
- `docs/templates/advanced-component-behavior-test.md`：正常、边界、重排、卸载、Ref、keyboard 和 overlay 行为断言。
- `docs/templates/advanced-component-visual-matrix.md`：Light/Dark、窄/宽、DPI、hover/focus/disabled/error/loading/长文本状态矩阵。
- `docs/templates/advanced-component-benchmark.md`：固定 Go 版本、输入规模、alloc、帧预算和 benchstat 比较记录。

## 使用要求

1. 为真实组件新建带 `fluxui-doc-meta` 的组件文档，并注册唯一 `example.id`。
2. 保持业务值受控；示例中的内部状态只能表示交互草稿、展示状态或演示按钮状态。
3. 显式说明稳定 identity、焦点、可见视口、字段状态和 anchored Overlay 对该组件的影响。
4. 在行为测试和视觉矩阵中填入实际结果；没有自动化覆盖的原生能力必须写明 smoke 步骤。
5. 新 API、弃用变更或平台要求同步更新 `api/ui.snapshot`、`docs/deprecation-ledger.md` 和 `docs/production-governance.md`。

模板演示仅复用已有 Select 试点来验证 Docs Browser 接线；它不宣称 Select 支持搜索、typeahead 或任何 R1 能力。
