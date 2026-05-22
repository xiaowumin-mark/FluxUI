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
    "FromWidget(w Widget) Element",
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

## React-style 状态

- 当前 `ProgressBar` 仍以 legacy `Widget` 作为稳定实现。
- React-style root 中可先使用 `FromWidget(ui.ProgressBar(...))` 桥接到 Element 树。
- 进度值仍由调用方状态驱动；`ProgressIndeterminate` 和动画更新不自动迁入 component HookSlot。
- 本阶段不冻结 `ProgressBarElement` API 名称；后续如引入 Element wrapper，需要先明确 progress value、indeterminate/animation 和 redraw lifecycle 归属。
