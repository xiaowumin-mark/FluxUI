<!-- fluxui-doc-meta
{
  "id": "circular_progress",
  "title": "CircularProgress 环形进度",
  "category": "反馈组件",
  "order": 310,
  "summary": "CircularProgress 用于环形进度展示。",
  "example": { "id": "circular_progress_basic" },
  "apis": [
    "CircularProgress(value float32, opts ...ProgressOption) Widget",
    "CircularProgressElement(value float32, opts ...ProgressOption) Element",
    "ProgressMin(min float32) ProgressOption",
    "ProgressMax(max float32) ProgressOption",
    "ProgressThickness(thickness float32) ProgressOption",
    "ProgressSize(size float32) ProgressOption",
    "ProgressTrackColor(col color.NRGBA) ProgressOption",
    "ProgressFillColor(col color.NRGBA) ProgressOption"
  ]
}
-->

# CircularProgress 环形进度

## 组件说明
CircularProgress 常用于面板、仪表卡片、局部任务状态展示。

## 使用方法
- `ProgressSize` 控制整体尺寸。
- `ProgressThickness` 控制环宽。
- 数值范围仍由 `ProgressMin/ProgressMax` 指定。

## 使用示例
```go
ui.CircularProgress(
    72,
    ui.ProgressMin(0),
    ui.ProgressMax(100),
    ui.ProgressSize(80),
    ui.ProgressThickness(8),
)
```

## React-style Element

`CircularProgressElement` 已可在 `RunElement` root 下直接使用。进度值仍由调用方状态驱动；动画值和不定进度的 frame redraw lifecycle 仍由底层 progress widget 管理。

```go
func CapacityGauge(ctx *ui.Context) ui.Element {
    usage := ui.UseState(ctx, float32(72))
    return ui.CircularProgressElement(
        usage.Value(),
        ui.ProgressMin(0),
        ui.ProgressMax(100),
        ui.ProgressSize(80),
        ui.ProgressThickness(8),
    )
}
```
