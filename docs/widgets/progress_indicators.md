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
    "CircularProgress(value float32, opts ...ProgressOption) Widget",
    "CircularProgressElement(value float32, opts ...ProgressOption) Element",
    "CircularProgressIndicator(value float32, opts ...ProgressOption) Widget",
    "CircularProgressIndicatorElement(value float32, opts ...ProgressOption) Element",
    "LoadingIndicator(opts ...ProgressOption) Widget",
    "LoadingIndicatorElement(opts ...ProgressOption) Element",
    "ProgressMin(min float32) ProgressOption",
    "ProgressMax(max float32) ProgressOption",
    "ProgressSize(size float32) ProgressOption",
    "ProgressThickness(thickness float32) ProgressOption",
    "ProgressTrackHeight(height float32) ProgressOption",
    "ProgressIndicatorHeight(height float32) ProgressOption",
    "ProgressTrackColor(col color.NRGBA) ProgressOption",
    "ProgressFillColor(col color.NRGBA) ProgressOption",
    "ProgressActiveIndicatorColor(col color.NRGBA) ProgressOption",
    "ProgressBuffer(value float32) ProgressOption",
    "ProgressBufferColor(col color.NRGBA) ProgressOption",
    "ProgressLabelVisible(visible bool) ProgressOption",
    "ProgressIndeterminate(indeterminate bool) ProgressOption",
    "ProgressLoading(loading bool) ProgressOption",
    "ProgressFourColor(fourColor bool) ProgressOption",
    "ButtonLoading(loading bool) ButtonOption",
    "ButtonLoadingIndicator(indicator Widget) ButtonOption",
    "IconButtonLoading(loading bool) IconButtonOption",
    "IconButtonLoadingIndicator(indicator Widget) IconButtonOption",
    "ProgressDecoration(d Decoration) ProgressOption"
  ]
}
-->

# ProgressIndicators 进度指示器

FluxUI 保留了已有的 `ProgressBar` 和 `CircularProgress` 命名，同时提供 Material 3 风格的 `LinearProgressIndicator` 和 `CircularProgressIndicator` 别名。新的文档入口统一说明线性和环形两类进度指示器。

## 使用方法

- `LinearProgressIndicatorElement` / `CircularProgressIndicatorElement` 默认使用 Material Web 的 `value`/`max` 语义，默认范围为 `0..1`。
- 旧 `ProgressBar` / `CircularProgress` 默认范围仍为 `0..100`；需要统一时可显式传 `ProgressMin` / `ProgressMax`。
- `ProgressBuffer` 只影响线性进度，可展示下载缓冲或预加载进度。
- `ProgressIndeterminate(true)` 可启用不定进度。
- `ProgressFourColor(true)` 可启用四颜色 loading 循环。
- `ProgressLabelVisible(true)` 可在环形进度中心显示百分比文本。
- `LoadingIndicatorElement` 是环形不定进度的快捷入口。
- `ButtonLoading(true)` / `IconButtonLoading(true)` 可在组件内部显示 loading，并暂停点击分发。

```go
ui.ColumnElement(
    ui.LinearProgressIndicatorElement(
        0.5,
        ui.ProgressBuffer(0.8),
        ui.ProgressTrackHeight(4),
        ui.ProgressIndicatorHeight(4),
        ui.ProgressFillColor(ui.NRGBA(37, 99, 235, 255)),
    ),
    ui.CircularProgressIndicatorElement(
        0.5,
        ui.ProgressSize(72),
        ui.ProgressThickness(4),
        ui.ProgressLabelVisible(true),
    ),
    ui.RowElement(
        ui.FilledTonalButtonElement(ui.TextElement("Loading"), ui.ButtonLoading(true)),
        ui.IconButtonElement(ui.IconElement("sync"), ui.IconButtonLoading(true)),
    ),
)
```
