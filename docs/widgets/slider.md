<!-- fluxui-doc-meta
{
  "id": "slider",
  "title": "Slider 滑块",
  "category": "输入交互",
  "order": 240,
  "summary": "Slider 用于区间连续值输入。",
  "example": { "id": "slider_basic" },
  "apis": [
    "Slider(value float32, opts ...SliderOption) Widget",
    "RangeSlider(start, end float32, opts ...SliderOption) Widget",
    "SliderElement(value float32, opts ...SliderOption) Element",
    "RangeSliderElement(start, end float32, opts ...SliderOption) Element",
    "SliderOnChange(fn func(ctx *Context, value float32)) SliderOption",
    "SliderOnRangeChange(fn func(ctx *Context, start, end float32)) SliderOption",
    "SliderDisabled(disabled bool) SliderOption",
    "SliderMin(min float32) SliderOption",
    "SliderMax(max float32) SliderOption",
    "SliderStep(step float32) SliderOption",
    "SliderWidth(width float32) SliderOption",
    "SliderLabeled(labeled bool) SliderOption",
    "SliderTicks(ticks bool) SliderOption",
    "SliderRange(start, end float32) SliderOption",
    "SliderValueLabel(label string) SliderOption",
    "SliderValueLabels(start, end string) SliderOption",
    "SliderTrackColor(color color.NRGBA) SliderOption",
    "SliderThumbColor(color color.NRGBA) SliderOption",
    "SliderProgressColor(color color.NRGBA) SliderOption",
    "SliderLabelColor(color color.NRGBA) SliderOption",
    "SliderLabelTextColor(color color.NRGBA) SliderOption",
    "SliderDecoration(d Decoration) SliderOption",
    "NewSliderRef() *SliderRef",
    "SliderAttachRef(ref *SliderRef) SliderOption",
    "(*SliderRef).SetValue(value float32)",
    "(*SliderRef).StepBy(delta float32)"
  ]
}
-->

# Slider 滑块

## 组件说明
Slider 适用于音量、进度、阈值等连续数值场景。FluxUI 的 Slider 已按 Material Web 的 `md-slider` 更新，支持 single point、range、labeled、tick marks、disabled 和自定义 token 色。

## MD3 默认样式

- active track、handle、value label 默认使用 `Primary`。
- inactive track 默认使用 `SurfaceContainerHighest`，没有该 token 时回退 `SurfaceVariant`。
- state layer 为 40dp，handle 为 20dp，track 为 4dp，value label 为 28dp，tick marks 为 2dp。
- labeled 模式在 hover / press / drag 时以 Material Web 的 scale/reveal 动画显示 value label。
- disabled 使用 `OnSurface` 的 12% inactive track 和 38% active/handle。

## 使用方法
- 设定取值范围：`SliderMin` + `SliderMax`。
- 离散步进：`SliderStep`，配合 `SliderTicks(true)` 显示 Material tick marks。
- 标签：`SliderLabeled(true)`，可用 `SliderValueLabel` / `SliderValueLabels` 覆盖显示文本。
- 范围选择：使用 `RangeSliderElement(start, end, ...)` 或 `RangeSlider(start, end, ...)`，通过 `SliderOnRangeChange` 接收两个值。
- 与进度条联动时建议统一状态源。
- 外部程序调整值可使用 `SliderAttachRef`，通过 `SetValue/StepBy` 下发命令。

## React-style Element

- `SliderElement` 已可在 `RunElement` root 下直接使用。
- `value` 仍由调用方状态驱动；拖拽过程中的 host state 和 `SliderRef` 命令队列仍由底层 slider widget 管理。

## 使用示例

### React-style Element

```go
func VolumeSlider(ctx *ui.Context) ui.Element {
    value := ui.UseState(ctx, float32(30))
    return ui.SliderElement(
        value.Value(),
        ui.SliderMin(0),
        ui.SliderMax(100),
        ui.SliderLabeled(true),
        ui.SliderOnChange(func(ctx *ui.Context, v float32) {
            value.Set(v)
        }),
    )
}
```

### Material Web parity examples

```go
ui.ColumnElement(
    ui.TextElement("Continuous"),
    ui.SliderElement(value.Value(),
        ui.SliderStep(0),
        ui.SliderOnChange(func(ctx *ui.Context, v float32) { value.Set(v) }),
    ),

    ui.TextElement("Labeled"),
    ui.SliderElement(value.Value(),
        ui.SliderLabeled(true),
        ui.SliderOnChange(func(ctx *ui.Context, v float32) { value.Set(v) }),
    ),

    ui.TextElement("Tick marks"),
    ui.SliderElement(value.Value(),
        ui.SliderStep(10),
        ui.SliderTicks(true),
        ui.SliderOnChange(func(ctx *ui.Context, v float32) { value.Set(v) }),
    ),

    ui.TextElement("Range"),
    ui.RangeSliderElement(start.Value(), end.Value(),
        ui.SliderLabeled(true),
        ui.SliderOnRangeChange(func(ctx *ui.Context, s, e float32) {
            start.Set(s)
            end.Set(e)
        }),
    ),
)
```

### Legacy Widget
旧 `ui.Slider` / `Widget` 写法继续可用：

```go
value := ui.State[float32](ctx)
ui.Slider(
    value.Value(),
    ui.SliderMin(0),
    ui.SliderMax(100),
    ui.SliderOnChange(func(ctx *ui.Context, v float32) {
        value.Set(v)
    }),
)
```
