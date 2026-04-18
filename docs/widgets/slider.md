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
    "SliderOnChange(fn func(ctx *Context, value float32)) SliderOption",
    "SliderDisabled(disabled bool) SliderOption",
    "SliderMin(min float32) SliderOption",
    "SliderMax(max float32) SliderOption",
    "SliderStep(step float32) SliderOption",
    "SliderWidth(width float32) SliderOption",
    "SliderTrackColor(color color.NRGBA) SliderOption",
    "SliderThumbColor(color color.NRGBA) SliderOption",
    "SliderProgressColor(color color.NRGBA) SliderOption",
    "NewSliderRef() *SliderRef",
    "SliderAttachRef(ref *SliderRef) SliderOption",
    "(*SliderRef).SetValue(value float32)",
    "(*SliderRef).StepBy(delta float32)"
  ]
}
-->

# Slider 滑块

## 组件说明
Slider 适用于音量、进度、阈值等连续数值场景。

## 使用方法
- 设定取值范围：`SliderMin` + `SliderMax`。
- 离散步进：`SliderStep`。
- 与进度条联动时建议统一状态源。
- 外部程序调整值可使用 `SliderAttachRef`，通过 `SetValue/StepBy` 下发命令。

## 使用示例
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
