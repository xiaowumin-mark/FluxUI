<!-- fluxui-doc-meta
{
  "id": "progress_indicators",
  "title": "ProgressIndicators 进度指示器",
  "category": "反馈组件",
  "order": 312,
  "summary": "ProgressIndicators 提供线性和环形两类 Material 3 进度指示器。",
  "example": { "id": "progress_indicators_basic" },
  "apis": [
    "LinearProgressIndicator(value float32, opts ...ProgressOption) Widget",
    "LinearProgressIndicatorElement(value float32, opts ...ProgressOption) Element",
    "CircularProgressIndicator(value float32, opts ...ProgressOption) Widget",
    "CircularProgressIndicatorElement(value float32, opts ...ProgressOption) Element",
    "ProgressLabelVisible(visible bool) ProgressOption",
    "ProgressIndeterminate(indeterminate bool) ProgressOption"
  ]
}
-->

# ProgressIndicators 进度指示器

FluxUI 保留了已有的 `ProgressBar` 和 `CircularProgress` 命名，同时提供 Material 3 风格的 `LinearProgressIndicator` 和 `CircularProgressIndicator` 别名。新的文档入口统一说明线性和环形两类进度指示器。

## 使用方法

- `LinearProgressIndicatorElement` 用于线性进度条。
- `CircularProgressIndicatorElement` 用于环形进度。
- `ProgressIndeterminate(true)` 可启用不定进度。
- `ProgressLabelVisible(true)` 可在环形进度中心显示百分比文本。

```go
ui.ColumnElement(
    ui.LinearProgressIndicatorElement(value.Value()),
    ui.CircularProgressIndicatorElement(
        value.Value(),
        ui.ProgressSize(72),
        ui.ProgressLabelVisible(true),
    ),
)
```
