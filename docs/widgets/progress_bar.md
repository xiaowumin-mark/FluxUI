<!-- fluxui-doc-meta
{
  "id": "progress_bar",
  "title": "ProgressBar 线性进度条",
  "category": "反馈组件",
  "order": 300,
  "summary": "ProgressBar 用于展示线性进度。",
  "example": { "id": "progress_bar_basic" },
  "apis": [
    "ProgressBar(value float32, opts ...ProgressOption) Widget",
    "ProgressMin(min float32) ProgressOption",
    "ProgressMax(max float32) ProgressOption",
    "ProgressIndeterminate(indeterminate bool) ProgressOption",
    "ProgressThickness(thickness float32) ProgressOption",
    "ProgressTrackColor(col color.NRGBA) ProgressOption",
    "ProgressFillColor(col color.NRGBA) ProgressOption"
  ]
}
-->

# ProgressBar 线性进度条

## 组件说明
ProgressBar 适用于上传、下载、任务执行等线性进度展示。

## 使用方法
- 通过 `ProgressMin/ProgressMax` 指定范围。
- 通过 `value` 传当前进度值。
- 通过 `ProgressTrackColor/ProgressFillColor` 自定义样式。

## 使用示例
```go
ui.ProgressBar(
    40,
    ui.ProgressMin(0),
    ui.ProgressMax(100),
    ui.ProgressTrackColor(ui.NRGBA(226, 232, 240, 255)),
    ui.ProgressFillColor(ui.NRGBA(30, 136, 229, 255)),
)
```
