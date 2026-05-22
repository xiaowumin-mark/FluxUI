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
    "FromWidget(w Widget) Element",
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

## React-style 状态

- 当前 `CircularProgress` 仍以 legacy `Widget` 作为稳定实现。
- React-style root 中可先使用 `FromWidget(ui.CircularProgress(...))` 桥接到 Element 树。
- 进度值仍由调用方状态驱动；若使用动画值，应先保留现有 frame-driven animation 语义。
- 本阶段不冻结 `CircularProgressElement` API 名称；后续如引入 Element wrapper，需要先明确 progress value、indeterminate/animation 和 redraw lifecycle 归属。
